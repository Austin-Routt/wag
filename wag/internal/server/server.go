// Package server is WAG's local web server: a JSON API over game sessions
// and the browser interface (embedded in the binary), so that a game can be
// played and inspected in a browser on any machine (cmd/wag-server).
//
// The Go engine stays the single source of truth: the browser only renders
// the State it is sent and posts moves, which the engine validates and
// evaluates exactly as in the CLI and the tests. Several browser tabs can
// show the same game (for example the players and a moderator); they poll
// for changes.
package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/kialo"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

//go:embed web
var webFiles embed.FS

// Data is the game data: the Carneades scheme file and the NASCQ files.
type Data struct {
	YAML, Schemes, CQs, Edges []byte
	Contexts                  []byte // context presets (contexts.json); none if empty
}

// Server holds the games in memory.
type Server struct {
	mu       sync.Mutex
	data     Data
	kialoDir string
	claude   *Claude // nil: no Claude seats
	contexts []game.Context
	sessions map[string]*Session
	next     int
	meta     []byte
}

// Seat is a player and who plays it.
type Seat struct {
	Name string `json:"name"`
	Kind string `json:"kind"` // human; llm (any language model, through the prompt and a pasted reply); random, greedy (rule bots); claude (the Claude API)
}

func (s Seat) aiKind() string {
	if s.Kind == "random" || s.Kind == "greedy" || s.Kind == "claude" {
		return s.Kind
	}
	return ""
}

// SetClaude offers Claude seats, played through the Claude API.
func (s *Server) SetClaude(c *Claude) { s.claude = c; s.meta = nil }

// Session is one game.
type Session struct {
	ID      string
	Title   string
	Source  string
	GS      *game.GameState
	Seats   []Seat
	agents  []game.Agent
	Version int // bumped on every change, so pollers can skip unchanged states
	Created time.Time
	Import  *kialo.Report
	Request *NewGameRequest // what started the game (nil for an imported discussion)
	Seed    int64           // the AI seats' seed
	Notes   []Note          // testers' remarks
	Events  []Event         // every request that tried to change the game
	LLM     []LLMRecord     // what Claude seats replied, move by move
	// thinking: a Claude seat is waiting for the model; the game takes no
	// other move meanwhile.
	thinking bool
}

// New makes a server for the data; kialoDir, if not empty, is a directory
// of captured Kialo discussions (<dir>/<id>/graph.json) offered for import.
func New(d Data, kialoDir string) (*Server, error) {
	s := &Server{data: d, kialoDir: kialoDir, sessions: map[string]*Session{}}
	if _, err := s.newState(game.DefaultGameConfig(), []string{"x"}); err != nil {
		return nil, fmt.Errorf("the game data does not load: %v", err)
	}
	if len(d.Contexts) > 0 {
		cs, err := game.LoadContexts(d.Contexts)
		if err != nil {
			return nil, err
		}
		s.contexts = cs
	}
	return s, nil
}

func (s *Server) newState(cfg game.GameConfig, names []string) (*game.GameState, error) {
	ad, err := carneades.NewAdapterFromYAML(s.data.YAML)
	if err != nil {
		return nil, err
	}
	n, err := nascq.LoadFromBytes(s.data.Schemes, s.data.CQs, s.data.Edges)
	if err != nil {
		return nil, err
	}
	return game.NewGame(cfg, ad, n, names), nil
}

// Handler serves the API under /api/ and the interface elsewhere.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/meta", s.handleMeta)
	mux.HandleFunc("/api/suggest", s.handleSuggest)
	mux.HandleFunc("/api/games", s.handleGames)
	mux.HandleFunc("/api/games/", s.handleGame)
	mux.HandleFunc("/api/load", s.handleLoad)
	mux.HandleFunc("/api/kialo", s.handleKialoList)
	mux.HandleFunc("/api/kialo/import", s.handleKialoImport)
	sub, _ := fs.Sub(webFiles, "web")
	mux.Handle("/", http.FileServer(http.FS(sub)))
	return mux
}

// ---------------------------------------------------------------------------
// Metadata: the schemes, questions and links of the NASCQ
// ---------------------------------------------------------------------------

// SchemeMeta is a scheme with everything the interface shows about it.
type SchemeMeta struct {
	*nascq.ASNode
	VariableOrder []string       `json:"variable_order"`
	Asked         []AskMeta      `json:"asked"`               // the CQs asked of it
	Backs         []string       `json:"backs"`               // CQs it backs by an explicit edge
	Carneades     *CarneadesMeta `json:"carneades,omitempty"` // its formulas
	Optional      bool           `json:"optional,omitempty"`  // only in games that allow it (Other)
}

// AskMeta is a CQ asked of a scheme, with its locus.
type AskMeta struct {
	CQ      string `json:"cq"`
	Locus   string `json:"locus"`
	Pattern string `json:"pattern,omitempty"`
}

// CarneadesMeta is a Carneades scheme's formulas.
type CarneadesMeta struct {
	ID          string   `json:"id"`
	Title       string   `json:"title,omitempty"`
	Conclusions []string `json:"conclusions"`
	Premises    []string `json:"premises"`
	Assumptions []string `json:"assumptions,omitempty"`
	Exceptions  []string `json:"exceptions,omitempty"`
	// The same formulas as the board reads them, variables as {X}.
	Reads *SchemeReads `json:"reads,omitempty"`
}

// SchemeReads is how a scheme's formulas read (game.FormulaText).
type SchemeReads struct {
	Conclusion  string   `json:"conclusion"`
	Premises    []string `json:"premises"`
	Assumptions []string `json:"assumptions,omitempty"`
	Exceptions  []string `json:"exceptions,omitempty"`
}

func readsOf(lang caes.Language, t *caes.Scheme) *SchemeReads {
	all := func(fs []string) []string {
		var out []string
		for _, f := range fs {
			out = append(out, game.FormulaText(lang, f))
		}
		return out
	}
	r := &SchemeReads{Premises: all(t.Premises), Assumptions: all(t.Assumptions), Exceptions: all(t.Exceptions)}
	if len(t.Conclusions) > 0 {
		r.Conclusion = game.FormulaText(lang, t.Conclusions[0])
	}
	return r
}

// CQMeta is a critical question with its links.
type CQMeta struct {
	*nascq.CQNode
	AskedOf  []AskOf  `json:"asked_of"`
	BackedBy []string `json:"backed_by"` // explicit backup edges; thesis-independent schemes back all
	Optional bool     `json:"optional,omitempty"`
}

// AskOf is a scheme a CQ is asked of, with the locus there.
type AskOf struct {
	Scheme  string `json:"scheme"`
	Locus   string `json:"locus"`
	Pattern string `json:"pattern,omitempty"`
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.meta == nil {
		b, err := s.buildMeta()
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		s.meta = b
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(s.meta)
}

func (s *Server) buildMeta() ([]byte, error) {
	cfg := game.DefaultGameConfig()
	cfg.AllowOther = true
	gs, err := s.newState(cfg, []string{"x"})
	if err != nil {
		return nil, err
	}
	n := gs.NASCQ
	base, _ := nascq.LoadFromBytes(s.data.Schemes, s.data.CQs, s.data.Edges)
	theory := map[string]*caes.Scheme{}
	for _, sc := range gs.Adapter.Graph.Theory.ArgSchemes {
		theory[sc.Id] = sc
	}
	var schemes []SchemeMeta
	for _, id := range nascq.SortedStringKeys(n.Schemes) {
		sc := n.Schemes[id]
		sm := SchemeMeta{ASNode: sc, Asked: []AskMeta{}, Backs: []string{}, Optional: base.GetScheme(id) == nil}
		sm.VariableOrder = game.VariableOrder(nascq.ActionTemplate{Variables: sc.Variables, PremiseTemplatesEN: sc.PremiseTemplatesEN,
			ConclusionTemplateEN: sc.ConclusionTemplateEN})
		for _, q := range n.GetAskCQs(id) {
			l := n.LocusOf(q.ID, id)
			sm.Asked = append(sm.Asked, AskMeta{CQ: q.ID, Locus: l.Kind, Pattern: l.Pattern})
		}
		for _, e := range n.BackupEdges {
			if e.Scheme == id {
				sm.Backs = append(sm.Backs, e.CQ)
			}
		}
		if t := theory[sc.CarneadesID]; t != nil {
			title, _ := t.Metadata["title"].(string)
			sm.Carneades = &CarneadesMeta{ID: t.Id, Title: title, Conclusions: t.Conclusions, Premises: t.Premises,
				Assumptions: t.Assumptions, Exceptions: t.Exceptions, Reads: readsOf(gs.Adapter.Graph.Theory.Language, t)}
		}
		schemes = append(schemes, sm)
	}
	var cqs []CQMeta
	for _, id := range nascq.SortedStringKeys(n.Questions) {
		cm := CQMeta{CQNode: n.Questions[id], AskedOf: []AskOf{}, BackedBy: []string{}, Optional: base.GetCQ(id) == nil}
		for _, sid := range nascq.SortedStringKeys(n.Schemes) {
			for _, q := range n.GetAskCQs(sid) {
				if q.ID == id {
					l := n.LocusOf(id, sid)
					cm.AskedOf = append(cm.AskedOf, AskOf{Scheme: sid, Locus: l.Kind, Pattern: l.Pattern})
				}
			}
		}
		for _, e := range n.BackupEdges {
			if e.CQ == id {
				cm.BackedBy = append(cm.BackedBy, e.Scheme)
			}
		}
		cqs = append(cqs, cm)
	}
	return json.Marshal(map[string]interface{}{
		"modes":    game.Modes(),
		"schemes":  schemes,
		"cqs":      cqs,
		"kialo":    s.kialoDir != "",
		"claude":   claudeMeta(s.claude),
		"contexts": s.contexts,
	})
}

func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	n, err := nascq.LoadFromBytes(s.data.Schemes, s.data.CQs, s.data.Edges)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"suggestions": n.Suggest(req.Text)})
}

// ---------------------------------------------------------------------------
// Games
// ---------------------------------------------------------------------------

// NewGameRequest configures a game: a mode preset, the seats, and the
// settings that change the preset.
type NewGameRequest struct {
	Title           string   `json:"title"`
	Mode            string   `json:"mode"`
	Seats           []Seat   `json:"seats"`
	Scoring         string   `json:"scoring,omitempty"`
	MaxRounds       *int     `json:"max_rounds,omitempty"`
	TargetScore     *float64 `json:"target_score,omitempty"`
	TimeLimitSec    *int     `json:"time_limit_sec,omitempty"`
	MoveTimeSec     *int     `json:"move_time_sec,omitempty"`
	AnswerWithin    *int     `json:"answer_within,omitempty"`
	Claim           string   `json:"claim,omitempty"`
	Standard        string   `json:"standard,omitempty"`
	ConcedeOnPass   *bool    `json:"concede_on_pass,omitempty"`
	AllowOther      bool     `json:"allow_other"`
	AnyPremise      *bool    `json:"any_premise,omitempty"` // default: on
	Context         string   `json:"context,omitempty"`     // a context preset: its scheme weights apply
	Moderated       bool     `json:"moderated"`
	RequireEvidence bool     `json:"require_evidence"`
	Seed            int64    `json:"seed,omitempty"`
}

func (req *NewGameRequest) config(contexts []game.Context) (game.GameConfig, error) {
	cfg, err := game.ModeConfig(req.Mode)
	if err != nil {
		return cfg, err
	}
	if req.Scoring != "" {
		if req.Scoring != game.ScoringMove && req.Scoring != game.ScoringLive {
			return cfg, fmt.Errorf("unknown scoring %q", req.Scoring)
		}
		cfg.Scoring.Model = req.Scoring
	}
	if req.MaxRounds != nil {
		cfg.MaxTurns = *req.MaxRounds
	}
	if req.TargetScore != nil {
		cfg.TargetScore = *req.TargetScore
	}
	if req.TimeLimitSec != nil {
		cfg.TimeLimit = time.Duration(*req.TimeLimitSec) * time.Second
	}
	if req.MoveTimeSec != nil {
		cfg.MoveTimeLimit = time.Duration(*req.MoveTimeSec) * time.Second
	}
	if req.AnswerWithin != nil {
		cfg.AnswerWithin = *req.AnswerWithin
	}
	if req.ConcedeOnPass != nil {
		cfg.ConcedeOnPass = *req.ConcedeOnPass
	}
	if strings.TrimSpace(req.Claim) != "" {
		cfg.Objective, cfg.Claim = game.ObjectiveClaim, strings.TrimSpace(req.Claim)
	}
	switch strings.ToUpper(req.Standard) {
	case "", "PE":
	case "CCE":
		cfg.ClaimStandard = caes.CCE
	case "BRD":
		cfg.ClaimStandard = caes.BRD
	default:
		return cfg, fmt.Errorf("unknown proof standard %q", req.Standard)
	}
	cfg.AllowOther, cfg.Moderated, cfg.RequireEvidence = req.AllowOther, req.Moderated, req.RequireEvidence
	if req.AnyPremise != nil {
		cfg.AnyPremise = *req.AnyPremise
	}
	if req.Context != "" {
		found := false
		for _, c := range contexts {
			if c.ID == req.Context {
				cfg.Context, cfg.Weights, found = c.ID, c.Weights, true
			}
		}
		if !found {
			return cfg, fmt.Errorf("unknown context %q", req.Context)
		}
	}
	return cfg, nil
}

func (s *Server) handleGames(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		defer s.mu.Unlock()
		type item struct {
			ID, Title, Mode, Source string
			Round                   int
			Over                    bool
			Created                 time.Time
		}
		var out []item
		for _, ss := range s.sessions {
			out = append(out, item{ss.ID, ss.Title, ss.GS.Config.Mode, ss.Source, ss.GS.CurrentTurn, ss.GS.Terminated, ss.Created})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
		writeJSON(w, out)
	case http.MethodPost:
		var req NewGameRequest
		if !readJSON(w, r, &req) {
			return
		}
		cfg, err := req.config(s.contexts)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		if len(req.Seats) == 0 {
			req.Seats = []Seat{{"Alice", "human"}, {"Bob", "greedy"}}
		}
		for _, seat := range req.Seats {
			switch seat.Kind {
			case "human", "llm", "random", "greedy":
			case "claude":
				if s.claude == nil {
					httpError(w, fmt.Errorf("this server has no Claude seats: start it with --claude"), http.StatusBadRequest)
					return
				}
			default:
				httpError(w, fmt.Errorf("unknown seat kind %q (human, llm, random, greedy, claude)", seat.Kind), http.StatusBadRequest)
				return
			}
		}
		if req.Mode == "solo" {
			req.Seats = req.Seats[:1]
		}
		var names []string
		for i := range req.Seats {
			req.Seats[i].Name = strings.TrimSpace(req.Seats[i].Name)
			if req.Seats[i].Name == "" {
				req.Seats[i].Name = fmt.Sprintf("Player%d", i+1)
			}
			names = append(names, req.Seats[i].Name)
		}
		gs, err := s.newState(cfg, names)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		title := strings.TrimSpace(req.Title)
		if title == "" {
			title = "WAG — " + cfg.Mode
		}
		s.mu.Lock()
		ss := s.add(gs, req.Seats, title, "new", req.Seed)
		ss.Request = &req
		st := ss.buildState()
		s.mu.Unlock()
		writeJSON(w, st)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// add registers a session. Caller holds s.mu.
func (s *Server) add(gs *game.GameState, seats []Seat, title, source string, seed int64) *Session {
	s.next++
	id := strconv.Itoa(s.next)
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	ss := &Session{ID: id, Title: title, Source: source, GS: gs, Seats: seats, Created: time.Now(), Version: 1, Seed: seed}
	for i, seat := range seats {
		var a game.Agent
		switch seat.Kind {
		case "random":
			a = game.NewRandomAgent(gs.Players[i].ID, seed+int64(i), 0.1, 0.5)
		case "greedy":
			a = game.NewGreedyAgent(gs.Players[i].ID, seed+int64(i), 2)
		}
		ss.agents = append(ss.agents, a)
	}
	s.sessions[id] = ss
	return ss
}

// handleGame serves /api/games/{id}[/{action}].
func (s *Server) handleGame(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/games/"), "/"), "/")
	s.mu.Lock()
	defer s.mu.Unlock()
	ss := s.sessions[parts[0]]
	if ss == nil {
		httpError(w, fmt.Errorf("no game %q", parts[0]), http.StatusNotFound)
		return
	}
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	if ss.thinking && r.Method == http.MethodPost && action != "preview" {
		httpError(w, fmt.Errorf("%s (Claude) is thinking: wait for its move", ss.GS.ActivePlayer().ID), http.StatusConflict)
		return
	}
	switch action {
	case "":
		if v, err := strconv.Atoi(r.URL.Query().Get("since")); err == nil && v == ss.Version {
			writeJSON(w, map[string]interface{}{"unchanged": true, "version": ss.Version})
			return
		}
		writeJSON(w, ss.buildState())
	case "move":
		var req struct {
			Action nascq.ProposedAction `json:"action"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if ss.Seats[ss.GS.ActivePlayerIdx].aiKind() != "" {
			httpError(w, fmt.Errorf("it is %s's turn, an AI", ss.GS.ActivePlayer().ID), http.StatusConflict)
			return
		}
		player := ss.GS.ActivePlayer().ID
		res, err := ss.GS.ExecuteTurn(req.Action)
		ss.event("move", player, short(req.Action), res, err)
		s.reply(w, ss, res, err)
	case "note":
		var req struct {
			Target string `json:"target"`
			Kind   string `json:"kind"`
			Text   string `json:"text"`
			By     string `json:"by"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		n, err := ss.addNote(req.Target, req.Kind, req.Text, req.By)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		ss.event("note", req.By, n.Kind+" "+n.Target, nil, nil)
		ss.Version++
		writeJSON(w, map[string]interface{}{"note": n, "state": ss.buildState()})
	case "prompt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprint(w, ss.Prompt())
	case "reply":
		var req struct {
			Reply string `json:"reply"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if ss.Seats[ss.GS.ActivePlayerIdx].aiKind() != "" {
			httpError(w, fmt.Errorf("it is %s's turn, an AI", ss.GS.ActivePlayer().ID), http.StatusConflict)
			return
		}
		player := ss.GS.ActivePlayer().ID
		action, err := ParseReply(req.Reply)
		if err != nil {
			ss.event("reply", player, "", nil, err)
			httpError(w, err, http.StatusBadRequest)
			return
		}
		res, err := ss.GS.ExecuteTurn(action)
		ss.event("reply", player, short(action), res, err)
		if err != nil {
			httpError(w, err, http.StatusConflict)
			return
		}
		ss.Version++
		writeJSON(w, map[string]interface{}{"result": res, "action": action, "state": ss.buildState()})
	case "moves":
		target := r.URL.Query().Get("target")
		a := ss.GS.Adapter.Graph.Arguments[target]
		if a == nil || !carneades.IsAttackable(a) {
			writeJSON(w, []MoveOption{})
			return
		}
		writeJSON(w, moves(ss.GS, target))
	case "preview":
		var req struct {
			Action nascq.ProposedAction `json:"action"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		writeJSON(w, ss.preview(req.Action))
	case "ai":
		if !jsonPost(w, r) {
			return
		}
		player := ss.GS.ActivePlayer().ID
		if ss.Seats[ss.GS.ActivePlayerIdx].Kind == "claude" {
			if s.claude == nil {
				httpError(w, fmt.Errorf("this server has no Claude seats: start it with --claude"), http.StatusConflict)
				return
			}
			if ss.GS.Terminated || ss.GS.Pending != nil {
				httpError(w, fmt.Errorf("no move to make: the game is over or a move awaits the moderator"), http.StatusConflict)
				return
			}
			ss.thinking = true
			ss.Version++ // pollers see the seat thinking
			s.mu.Unlock()
			res, err := s.claudeTurn(ss)
			s.mu.Lock()
			ss.thinking = false
			ss.event("ai", player, "claude", res, err)
			s.reply(w, ss, res, err)
			return
		}
		res, err := ss.aiMove()
		ss.event("ai", player, "", res, err)
		s.reply(w, ss, res, err)
	case "moderate":
		var req struct {
			Approve bool    `json:"approve"`
			Weight  float64 `json:"weight"`
			Note    string  `json:"note"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		res, err := ss.GS.Moderate(req.Approve, req.Weight, req.Note)
		ss.event("moderate", game.ModeratorID, fmt.Sprintf("approve=%v weight=%v %s", req.Approve, req.Weight, req.Note), res, err)
		s.reply(w, ss, res, err)
	case "decide":
		var req struct {
			Statement string `json:"statement"`
			Decision  string `json:"decision"`
			Note      string `json:"note"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		res, err := ss.GS.Decide(req.Statement, req.Decision, req.Note)
		ss.event("decide", game.ModeratorID, req.Decision+" "+req.Statement, res, err)
		s.reply(w, ss, res, err)
	case "export":
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=wag-game-%s.json", ss.ID))
		writeJSON(w, ss.export())
	default:
		httpError(w, fmt.Errorf("unknown action %q", action), http.StatusNotFound)
	}
}

// reply answers a request that changed the game (or failed to): the result
// and the new state. Only a change bumps the version pollers compare.
func (s *Server) reply(w http.ResponseWriter, ss *Session, res *game.TurnResult, err error) {
	if err != nil {
		httpError(w, err, http.StatusConflict)
		return
	}
	ss.Version++
	writeJSON(w, map[string]interface{}{"result": res, "state": ss.buildState()})
}

// aiMove lets the AI seat to move play once (retrying a rejected move up to
// five times, then passing, as the batch runner does).
func (ss *Session) aiMove() (*game.TurnResult, error) {
	gs := ss.GS
	if gs.Terminated {
		return nil, fmt.Errorf("the game is over")
	}
	if gs.Pending != nil {
		return nil, fmt.Errorf("a move awaits the moderator")
	}
	a := ss.agents[gs.ActivePlayerIdx]
	if a == nil {
		return nil, fmt.Errorf("it is %s's turn, a human", gs.ActivePlayer().ID)
	}
	for try := 0; try < 5; try++ {
		r, err := gs.ExecuteTurn(a.ChooseAction(gs.BuildObservation()))
		if err != nil {
			return nil, err
		}
		if r.Accepted || r.Pending || r.Terminated {
			return r, nil
		}
	}
	return gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"})
}

// Preview is what a move would do, tried on the game and undone.
type Preview struct {
	OK         bool               `json:"ok"`
	Error      string             `json:"error,omitempty"`
	Premises   []string           `json:"premises,omitempty"`
	Conclusion string             `json:"conclusion,omitempty"`
	Effect     string             `json:"effect,omitempty"`
	Deltas     map[string]float64 `json:"deltas,omitempty"`
	Changes    []LabelChange      `json:"changes,omitempty"` // statements whose label the move changes
}

// LabelChange is a statement whose label a move changes.
type LabelChange struct {
	Stmt   string `json:"stmt"`
	Text   string `json:"text"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// preview tries the move (Simulate) and describes it.
func (ss *Session) preview(a nascq.ProposedAction) *Preview {
	gs := ss.GS
	p := &Preview{}
	if gs.Pending != nil {
		p.Error = "a move awaits the moderator"
		return p
	}
	if gs.Terminated {
		p.Error = "the game is over"
		return p
	}
	if _, err := gs.NASCQ.Validate(a, gs.Bridge()); err != nil {
		p.Error = err.Error()
		return p
	}
	tmpl := nascq.ActionTemplate{Type: a.Type, TargetArgID: a.TargetArgID, AnswersCQID: a.AnswersCQID}
	if va, err := gs.NASCQ.Validate(a, gs.Bridge()); err == nil {
		tmpl.Locus, tmpl.LocusStmtID, tmpl.CaesAttackType = va.Locus, va.LocusStmtID, va.CaesAttackType
		if sc := gs.NASCQ.GetScheme(a.SchemeID); sc != nil {
			tmpl.ConclusionTemplateEN = sc.ConclusionTemplateEN
		}
	}
	p.Effect = gs.Effect(tmpl, a.VariableBindings)
	before := gs.Adapter.StatementLabels()
	var after map[string]caes.Label
	texts := map[string]string{} // the changed statements, as the move would leave them
	r, err := gs.SimulateThen(a, func(g *game.GameState) {
		after = g.Adapter.StatementLabels()
		for id, l := range after {
			if b, ok := before[id]; !ok || b != l {
				texts[id] = g.Say(id)
			}
		}
		// The argument as the move builds it, answer variable bound.
		if n := len(g.MoveHistory); n > 0 && a.Type != "question" {
			if m := g.MoveHistory[n-1]; len(m.ArgIDs) > 0 {
				if arg := g.Adapter.Graph.Arguments[m.ArgIDs[0]]; arg != nil && arg.Conclusion != nil {
					for _, pr := range arg.Premises {
						if pr.Stmt != nil {
							p.Premises = append(p.Premises, g.SayIn(arg.Id, pr.Stmt.Id))
						}
					}
					p.Conclusion = g.Say(arg.Conclusion.Id)
				}
			}
		}
	})
	if err != nil {
		p.Error = err.Error()
		return p
	}
	if !r.Accepted {
		p.Error = r.Error
		return p
	}
	p.OK, p.Deltas = true, r.ScoreDeltas
	for _, id := range carneades.SortedKeys(after) {
		b, ok := before[id]
		if !ok || b != after[id] {
			if !ok && after[id] != caes.In {
				continue
			}
			bs := "new"
			if ok {
				bs = labelName(b)
			}
			p.Changes = append(p.Changes, LabelChange{Stmt: id, Before: bs, After: labelName(after[id])})
		}
	}
	for i := range p.Changes {
		p.Changes[i].Text = texts[p.Changes[i].Stmt]
	}
	if a.TargetArgID == "" && p.Conclusion != "" {
		p.Effect = strings.TrimSuffix(p.Conclusion, ".")
	}
	return p
}

// describeAction says what a held move would do.
func (ss *Session) describeAction(a nascq.ProposedAction) string {
	gs := ss.GS
	var desc string
	held := gs.Pending
	gs.Pending = nil // the held move itself is tried
	defer func() { gs.Pending = held }()
	_, _ = gs.SimulateThen(a, func(g *game.GameState) {
		if n := len(g.MoveHistory); n > 0 {
			desc = g.DescribeMove(g.MoveHistory[n-1])
		}
	})
	return desc
}

// export is the whole game for saving: configuration, history with the
// players' own words, and the arguments made with the user-defined scheme
// (material for mining new schemes).
func (ss *Session) export() map[string]interface{} {
	st := ss.buildState()
	var mined []map[string]interface{}
	for _, id := range sortedStrings(keys(st.Arguments)) {
		a := st.Arguments[id]
		if !a.Free {
			continue
		}
		var prem []string
		for _, p := range a.Premises {
			prem = append(prem, st.Statements[p].Text)
		}
		mined = append(mined, map[string]interface{}{"argument": id, "owner": a.Owner, "text": a.Text,
			"premises": prem, "conclusion": st.Statements[a.Conclusion].Text, "kialo": a.Kialo})
	}
	var req *NewGameRequest
	if ss.Request != nil {
		r := *ss.Request
		r.Seed = ss.Seed
		req = &r
	}
	return map[string]interface{}{
		"format": "wag-export/2", "exported_at": time.Now().UTC(), "build": buildInfo(),
		"title": ss.Title, "source": ss.Source, "seats": ss.Seats, "config": st.Config, "request": req,
		"history": ss.GS.MoveHistory, "state": st, "other_arguments": mined, "import": ss.Import,
		"notes": ss.Notes, "events": ss.Events, "llm": ss.LLM,
	}
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// ---------------------------------------------------------------------------
// Kialo
// ---------------------------------------------------------------------------

func (s *Server) handleKialoList(w http.ResponseWriter, r *http.Request) {
	type item struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Size  int64  `json:"size"`
	}
	out := []item{}
	if s.kialoDir != "" {
		entries, _ := os.ReadDir(s.kialoDir)
		for _, e := range entries {
			p := filepath.Join(s.kialoDir, e.Name(), "graph.json")
			fi, err := os.Stat(p)
			if err != nil {
				continue
			}
			title := e.Name()
			if f, err := os.Open(p); err == nil {
				var head struct {
					Title string `json:"displayTitle"`
				}
				// The title is near the start; decode the whole file only if small.
				if fi.Size() < 32<<20 {
					b, _ := io.ReadAll(f)
					json.Unmarshal(b, &head)
				}
				f.Close()
				if head.Title != "" {
					title = head.Title
				}
			}
			out = append(out, item{e.Name(), title, fi.Size()})
		}
	}
	writeJSON(w, out)
}

func (s *Server) handleKialoImport(w http.ResponseWriter, r *http.Request) {
	if !jsonPost(w, r) {
		return
	}
	var b []byte
	var err error
	depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
	if id := r.URL.Query().Get("id"); id != "" {
		if s.kialoDir == "" || strings.ContainsAny(id, `/\.`) {
			httpError(w, fmt.Errorf("no such discussion"), http.StatusBadRequest)
			return
		}
		b, err = os.ReadFile(filepath.Join(s.kialoDir, id, "graph.json"))
	} else {
		b, err = io.ReadAll(io.LimitReader(r.Body, 64<<20))
	}
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	gs, rep, err := s.kialoGame(b, depth)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	ss := s.add(gs, []Seat{{kialo.ProPlayer, "human"}, {kialo.ConPlayer, "human"}}, rep.Title, "kialo:"+rep.Discussion, rand.Int63())
	ss.Import = rep
	st := ss.buildState()
	s.mu.Unlock()
	writeJSON(w, map[string]interface{}{"state": st, "report": rep})
}

// kialoGame builds the game of a Kialo discussion (graph.json), down to depth
// (0: all of it).
func (s *Server) kialoGame(b []byte, depth int) (*game.GameState, *kialo.Report, error) {
	d, err := kialo.Parse(b)
	if err != nil {
		return nil, nil, err
	}
	rels, rep := d.Plan(depth)
	ad, err := carneades.NewAdapterFromYAML(s.data.YAML)
	if err != nil {
		return nil, nil, err
	}
	n, err := nascq.LoadFromBytes(s.data.Schemes, s.data.CQs, s.data.Edges)
	if err != nil {
		return nil, nil, err
	}
	gs, err := kialo.Build(d, rels, rep, kialo.Config(), ad, n)
	if err != nil {
		return nil, nil, err
	}
	return gs, rep, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func readJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if !jsonPost(w, r) {
		return false
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(v); err != nil {
		httpError(w, fmt.Errorf("bad request: %v", err), http.StatusBadRequest)
		return false
	}
	return true
}

// jsonPost admits a request that changes something only as a POST with a
// JSON body. Other web pages cannot send that to this server without a CORS
// preflight, which it never grants, so they cannot play moves in its games.
func jsonPost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		httpError(w, fmt.Errorf("use POST"), http.StatusMethodNotAllowed)
		return false
	}
	if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
		httpError(w, fmt.Errorf("send application/json"), http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, "encoding:", err)
	}
}

func httpError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

// short is a one-line form of an action, for the event log.
func short(a nascq.ProposedAction) string {
	parts := []string{a.Type}
	for _, x := range []string{a.TargetArgID, a.AnswersCQID, a.SchemeID} {
		if x != "" {
			parts = append(parts, x)
		}
	}
	return strings.Join(parts, " ")
}

// handleLoad starts a game from an export (/export) and replays its moves.
func (s *Server) handleLoad(w http.ResponseWriter, r *http.Request) {
	var e Export
	if !readJSON(w, r, &e) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ss, err := s.load(&e)
	if err != nil {
		httpError(w, fmt.Errorf("cannot load the game: %v", err), http.StatusBadRequest)
		return
	}
	ss.event("load", "", fmt.Sprintf("%d moves", len(e.History)), nil, nil)
	writeJSON(w, ss.buildState())
}

func claudeMeta(c *Claude) interface{} {
	if c == nil {
		return nil
	}
	return map[string]string{"model": c.Model, "effort": c.Effort}
}
