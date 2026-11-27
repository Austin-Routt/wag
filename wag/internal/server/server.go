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
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
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

	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/kialo"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/manifest"
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

// Hash identifies the scheme data (schemes, questions, edges): a game can be
// replayed exactly only with the data it was played with. It reads CRLF line
// endings as LF, which the YAML and JSON readers do too, so that a Windows
// checkout fingerprints the same data as any other; the fingerprint of LF
// files is unchanged.
func (d Data) Hash() string { return d.hash(lf) }

func (d Data) hash(norm func([]byte) []byte) string {
	h := sha256.New()
	for _, b := range [][]byte{d.YAML, d.Schemes, d.CQs, d.Edges} {
		h.Write(norm(b))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func lf(b []byte) []byte   { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }
func crlf(b []byte) []byte { return bytes.ReplaceAll(lf(b), []byte("\n"), []byte("\r\n")) }

// Matches tells whether a recorded fingerprint is this data's: its Hash, or
// the fingerprint the same data had in a CRLF checkout, which games saved
// before Hash read line endings as LF (24 November 2026) recorded.
func (d Data) Matches(recorded string) bool {
	return recorded != "" && (recorded == d.Hash() || recorded == d.hash(crlf))
}

// Server holds the games in memory.
type Server struct {
	mu       sync.Mutex
	data     Data
	kialoDir string
	models   map[string]*llm.Client // the model profiles seats can use, by name
	ctx      context.Context        // the server's lifetime: model calls stop when it ends
	// The workbench (workbench.go): the data folder, the keys typed in the
	// dashboard, the players added there (and why one cannot play), the
	// annotation runs.
	store     string
	keys      *keyStore
	players   []llm.Profile
	playerErr map[string]string
	runs      map[string]*annotate.Run
	contexts  []game.Context
	sessions  map[string]*Session
	next      int
	meta      []byte
	// pins are the inputs by content (manifest): the engine, the scheme
	// file and the NASCQ data.
	pins manifest.Pins
	// evidenceDir is the evidence folder SetEvidence gave (evidenceSet);
	// without one, the data folder's evidence/ (evidenceFolder).
	evidenceDir string
	evidenceSet bool
	// stopAfterReply, set by a test, stops a model seat's turn after the
	// provider's answer is recorded and saved, before it is used, as a crash
	// would (docs/g-journal.md).
	stopAfterReply func() bool
}

// Seat is a player and who plays it.
type Seat struct {
	Name string `json:"name"`
	// Kind is human; llm (any language model, through the prompt and a
	// pasted reply); random, greedy (rule bots); model (a language model
	// through an API: the profile named by Model); claude (the profile
	// "claude", as --claude sets it up).
	Kind  string `json:"kind"`
	Model string `json:"model,omitempty"`
}

// aiKind is the kind of player the server moves for: random, greedy, or
// model (a model seat, claude included); "" for a seat a person moves.
func (s Seat) aiKind() string {
	switch s.Kind {
	case "random", "greedy":
		return s.Kind
	case "model", "claude":
		return "model"
	}
	return ""
}

// SetContext ties the model calls to ctx: they stop when it is done.
func (s *Server) SetContext(ctx context.Context) { s.ctx = ctx }

// AddContexts adds context presets (from a file given to the server); one
// with the ID of a built-in preset replaces it.
func (s *Server) AddContexts(cs []game.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contexts, s.meta = game.MergeContexts(s.contexts, cs), nil
}

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
	LLM     []LLMRecord     // what model seats replied, move by move
	Traces  []Trace         // every decision of every seat (traces.go)
	Calls   []Call          // every call of a model seat to its provider (traces.go)
	// order counts the traced events (decisions and calls), which share it.
	order int
	// SavedTo is the folder a game with a model seat is saved in, once a
	// save has succeeded; saveErr is why the last save failed (autosave).
	SavedTo string
	saveErr string
	// folder is the data folder's game folder a resumed game goes on saving
	// into (/api/store/resume); "" for a game saved under its own name.
	// resumed are the game's resumes, which its trace's start records.
	folder  string
	resumed []Resumption
	// build is the build that began a resumed game, which its trace's start
	// keeps (nil: this server's, buildInfo); each resume records its own.
	build map[string]string
	// pins are the server's inputs by content (manifest), which its trace's
	// start and its export record; none for a resumed game whose start
	// recorded none (begun before the pins were recorded).
	pins manifest.Pins
	// contexts are the server's context presets, for the verdicts under
	// each (State.Verdicts).
	contexts []game.Context
	// dataHash identifies the scheme data the game is played with.
	dataHash string
	// inspectedWith is set when a loaded game was played with other scheme
	// data (dataHash) and opened read-only on request: it identifies this
	// server's data, with which its moves were replayed (load). playedHistory
	// is then the history as the export recorded it, which its own export
	// keeps (the replay's moves carry this data's weights and labels).
	inspectedWith string
	playedHistory []game.Move
	// legacyReplayedWith is set when a game was loaded from an export that
	// names no scheme data (one made before 13 October 2026, or a legacy
	// record's own): its moves were replayed with this server's data, which
	// this identifies. It stays a legacy record: its export and its traces
	// name no data for its first legacyMoves moves (dataHash is ""), so that
	// nothing takes it for a game played with this data (the critical
	// review's adverse case 11; docs/c5-game.md); the moves after them were
	// played with continuedWith, which its export names.
	legacyReplayedWith string
	legacyMoves        int
	continuedWith      string
	// thinking: a model seat is waiting for its model; the game takes no
	// other move meanwhile.
	thinking bool
	// keys let other computers act in the game by role (access.go).
	keys gameKeys
}

// New makes a server for the data; kialoDir, if not empty, is a directory
// of captured Kialo discussions (<dir>/<id>/graph.json) offered for import.
func New(d Data, kialoDir string) (*Server, error) {
	s := &Server{data: d, kialoDir: kialoDir, sessions: map[string]*Session{}, models: map[string]*llm.Client{}, ctx: context.Background(),
		keys: &keyStore{keys: map[string]string{}, remembered: map[string]bool{}}, playerErr: map[string]string{}, runs: map[string]*annotate.Run{}}
	if _, err := s.newState(game.DefaultGameConfig(), []string{"x"}); err != nil {
		return nil, fmt.Errorf("the game data does not load: %v", err)
	}
	s.pins = manifest.Of(d.YAML, d.Schemes, d.CQs, d.Edges)
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
	mux.HandleFunc("/api/llm/", s.handleLLM)
	mux.HandleFunc("/api/store", s.handleStore)
	mux.HandleFunc("/api/store/", s.handleStore)
	mux.HandleFunc("/api/annotate/", s.handleAnnotate)
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
		"claude":   s.claudeMeta(),
		"models":   s.modelMeta(),
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
	Title         string   `json:"title"`
	Mode          string   `json:"mode"`
	Seats         []Seat   `json:"seats"`
	Scoring       string   `json:"scoring,omitempty"`
	MaxRounds     *int     `json:"max_rounds,omitempty"`
	TargetScore   *float64 `json:"target_score,omitempty"`
	TimeLimitSec  *int     `json:"time_limit_sec,omitempty"`
	MoveTimeSec   *int     `json:"move_time_sec,omitempty"`
	AnswerWithin  *int     `json:"answer_within,omitempty"`
	Claim         string   `json:"claim,omitempty"`
	Standard      string   `json:"standard,omitempty"`
	ConcedeOnPass *bool    `json:"concede_on_pass,omitempty"`
	AllowOther    bool     `json:"allow_other"`
	AnyPremise    *bool    `json:"any_premise,omitempty"` // default: on
	Context       string   `json:"context,omitempty"`     // a context preset: its scheme weights apply
	// Weights change the scheme weights for this game, on top of the
	// context's (scheme ID, or "*" for every scheme, to a weight in (0, 1]).
	Weights         map[string]float64 `json:"weights,omitempty"`
	Moderated       bool               `json:"moderated"`
	RequireEvidence bool               `json:"require_evidence"`
	Seed            int64              `json:"seed,omitempty"`
	// Profile is the game's profile (D094 (b)): its agent, the actions
	// argued about with their actors and aliases, and the situation
	// compared. Case is a case record (format wag-case/1) to make it from.
	// A new game with neither has the default profile; the request recorded
	// with the game holds the profile itself. An export whose request was
	// recorded before step C5.2 has none: it loads (/api/games/load) and
	// replays without the profile's rules. A trace a model player recorded
	// before then is replayed through a new game, which has a profile, so it
	// replays with the build it was recorded with (docs/llm.md).
	Profile *game.Profile   `json:"profile,omitempty"`
	Case    json.RawMessage `json:"case,omitempty"`
	// Evidence names the game's record set in the server's evidence folder
	// (docs/e1-evidence.md): an inquiry of two people or models only.
	Evidence string `json:"evidence,omitempty"`
}

// withProfile is the request of a new game with its profile: the case
// record's, the one given, or the default.
func (req NewGameRequest) withProfile() (NewGameRequest, error) {
	if strings.TrimSpace(string(req.Case)) == "null" {
		req.Case = nil // as if absent
	}
	if len(req.Case) > 0 {
		if req.Profile != nil {
			return req, fmt.Errorf("give a profile or a case record, not both")
		}
		p, err := game.ProfileFromCase(req.Case)
		if err != nil {
			return req, err
		}
		req.Profile, req.Case = p, nil
	}
	if req.Profile == nil {
		req.Profile = game.DefaultProfile()
	} else if strings.TrimSpace(req.Profile.Situation) == "" {
		p := *req.Profile
		p.Situation = game.DefaultProfile().Situation // the present one (D089)
		req.Profile = &p
	}
	return req, nil
}

func (req *NewGameRequest) config(contexts []game.Context) (game.GameConfig, error) {
	cfg, err := game.ModeConfig(req.Mode)
	if err != nil {
		return cfg, err
	}
	if len(req.Case) > 0 {
		return cfg, fmt.Errorf("the request's case record was not read into its profile")
	}
	if req.Profile != nil {
		if err := req.Profile.Check(); err != nil {
			return cfg, err
		}
		cfg.Profile = req.Profile
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
		if cfg.Profile != nil {
			if err := cfg.Profile.CheckClaim(cfg.Claim); err != nil {
				return cfg, err
			}
		}
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
	if len(req.Weights) > 0 {
		if err := game.CheckWeights(req.Weights); err != nil {
			return cfg, err
		}
		w := map[string]float64{}
		if _, all := req.Weights["*"]; !all { // a "*" of the game's own replaces the context's weights
			for k, v := range cfg.Weights {
				w[k] = v
			}
		}
		for k, v := range req.Weights {
			w[k] = v
		}
		cfg.Weights = w
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
		out := []item{}
		for _, ss := range s.sessions {
			if roleIn(r, ss).none() {
				continue // another computer sees the games it has a key for
			}
			out = append(out, item{ss.ID, ss.Title, ss.GS.Config.Mode, ss.Source, ss.GS.CurrentTurn, ss.GS.Terminated, ss.Created})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
		writeJSON(w, out)
	case http.MethodPost:
		if !ownerOnly(w, r) {
			return
		}
		var req NewGameRequest
		if !readJSON(w, r, &req) {
			return
		}
		req, err := req.withProfile()
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		cfg, err := req.config(s.contexts)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		if len(req.Seats) == 0 {
			req.Seats = []Seat{{Name: "Alice", Kind: "human"}, {Name: "Bob", Kind: "greedy"}}
		}
		for _, seat := range req.Seats {
			switch seat.Kind {
			case "human", "llm", "random", "greedy":
			case "claude", "model":
				if _, err := s.model(seat); err != nil {
					httpError(w, err, http.StatusBadRequest)
					return
				}
			default:
				httpError(w, fmt.Errorf("unknown seat kind %q (human, llm, random, greedy, model, claude)", seat.Kind), http.StatusBadRequest)
				return
			}
		}
		if req.Mode == "solo" {
			req.Seats = req.Seats[:1]
		}
		if err := game.CheckInquiry(cfg, len(req.Seats)); err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		if req.Evidence != "" {
			if err := checkEvidenceGame(cfg, req.Seats); err != nil {
				httpError(w, err, http.StatusBadRequest)
				return
			}
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
		if u := gs.UnknownSchemes(req.Weights); len(u) > 0 {
			httpError(w, fmt.Errorf("weights for unknown schemes: %s", strings.Join(u, ", ")), http.StatusBadRequest)
			return
		}
		title := strings.TrimSpace(req.Title)
		if title == "" {
			title = "WAG — " + cfg.Mode
		}
		if req.Evidence != "" {
			// Read and checked without the server's lock: hashing a record
			// set's payloads would stall every game (docs/e1-evidence.md,
			// E1.7). The new game is no one else's yet.
			s.mu.Lock()
			dir, pins := s.evidenceFolder(), s.pins
			s.mu.Unlock()
			ev, err := gameEvidence(dir, pins, req.Evidence, gs, req.Seats)
			if err != nil {
				httpError(w, err, http.StatusBadRequest)
				return
			}
			gs.Evidence = ev
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
	ss := &Session{ID: id, Title: title, Source: source, GS: gs, Seats: seats, Created: time.Now(), Version: 1, Seed: seed, contexts: s.contexts, dataHash: s.data.Hash(),
		keys: newKeys(seats), pins: s.pins}
	for i, seat := range seats {
		ss.agents = append(ss.agents, newAgent(seat.Kind, gs.Players[i].ID, seed+int64(i)))
	}
	s.sessions[id] = ss
	return ss
}

// newAgent is the computer player of a seat of that kind, nil for a seat
// the server does not choose for (a person, a model).
func newAgent(kind, id string, seed int64) game.Agent {
	switch kind {
	case "random":
		return game.NewRandomAgent(id, seed, 0.1, 0.5)
	case "greedy":
		return game.NewGreedyAgent(id, seed, 2)
	}
	return nil
}

// handleGame serves /api/games/{id}[/{action}].
func (s *Server) handleGame(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/games/"), "/"), "/")
	s.mu.Lock()
	defer s.mu.Unlock()
	ss := s.sessions[parts[0]]
	// Another computer without a key learns nothing, not even whether the
	// game exists (access.go).
	who := roleIn(r, ss)
	if who.none() {
		httpError(w, errNoKey, http.StatusForbidden)
		return
	}
	if ss == nil {
		httpError(w, fmt.Errorf("no game %q", parts[0]), http.StatusNotFound)
		return
	}
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	if err := who.may(ss, action); err != nil {
		httpError(w, err, http.StatusForbidden)
		return
	}
	if ss.thinking && r.Method == http.MethodPost && action != "preview" {
		httpError(w, fmt.Errorf("%s (Claude) is thinking: wait for its move", ss.GS.ActivePlayer().ID), http.StatusConflict)
		return
	}
	if ss.inspectedWith != "" && r.Method == http.MethodPost && action != "preview" && action != "note" && action != "save" {
		httpError(w, fmt.Errorf("the game is open read-only: it was played with other scheme data (%s; this server's is %s)", ss.dataHash, ss.inspectedWith), http.StatusConflict)
		return
	}
	switch action {
	case "":
		if v, err := strconv.Atoi(r.URL.Query().Get("since")); err == nil && v == ss.Version {
			writeJSON(w, map[string]interface{}{"unchanged": true, "version": ss.Version})
			return
		}
		writeJSON(w, ss.stateFor(who))
	case "move":
		var req struct {
			Action json.RawMessage `json:"action"`
			Why    string          `json:"why"` // the player's note on the move, for the trace
		}
		if !readJSON(w, r, &req) {
			return
		}
		action, err := ss.decodeAction(req.Action)
		if err != nil {
			httpError(w, fmt.Errorf("bad request: %v", err), http.StatusBadRequest)
			return
		}
		if ss.Seats[ss.GS.ActivePlayerIdx].aiKind() != "" {
			httpError(w, fmt.Errorf("it is %s's turn, an AI", ss.GS.ActivePlayer().ID), http.StatusConflict)
			return
		}
		player := ss.GS.ActivePlayer().ID
		t := ss.newTrace(ss.Seats[ss.GS.ActivePlayerIdx].Kind)
		t.Why = strings.TrimSpace(req.Why)
		res, err := ss.play(action, t, nil)
		ss.event("move", player, short(action), res, err)
		s.reply(w, ss, who, res, err)
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
		if name := who.name(ss); name != "" {
			req.By = name // a key writes in its holder's name
		}
		n, err := ss.addNote(req.Target, req.Kind, req.Text, req.By)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		ss.event("note", req.By, n.Kind+" "+n.Target, nil, nil)
		ss.Version++
		s.autosave(ss) // notes are part of the record; a failure is in the state
		writeJSON(w, map[string]interface{}{"note": n, "state": ss.stateFor(who)})
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
		t := ss.newTrace(ss.Seats[ss.GS.ActivePlayerIdx].Kind).withPrompt(ss.Prompt())
		t.Reply = req.Reply
		rep, err := parseReply(req.Reply, ss.GS.Evidence != nil)
		if err != nil {
			t.Result.Error = err.Error()
			ss.addTrace(t)
			ss.event("reply", player, "", nil, err)
			s.saveRecord(ss) // the reply is in the record, played or not
			httpError(w, err, http.StatusBadRequest)
			return
		}
		res, err := ss.play(rep.Action, t, rep)
		action := t.Action
		ss.event("reply", player, short(action), res, err)
		if err != nil {
			s.saveRecord(ss)
			httpError(w, err, http.StatusConflict)
			return
		}
		ss.Version++
		s.autosave(ss) // a failure is in the state
		writeJSON(w, map[string]interface{}{"result": res, "action": action, "state": ss.stateFor(who)})
	case "moves":
		target := r.URL.Query().Get("target")
		a := ss.GS.Adapter.Graph.Arguments[target]
		if a == nil || !carneades.IsAttackable(a) {
			writeJSON(w, []MoveOption{})
			return
		}
		writeJSON(w, moves(ss.GS, target))
	case "stmt":
		id := r.URL.Query().Get("id")
		if ss.GS.Adapter.Graph.Statements[id] == nil {
			httpError(w, fmt.Errorf("no statement %q", id), http.StatusNotFound)
			return
		}
		writeJSON(w, stmtMoves(ss.GS, id))
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
		if seat := ss.Seats[ss.GS.ActivePlayerIdx]; seat.aiKind() == "model" {
			c, err := s.model(seat)
			if err != nil {
				httpError(w, err, http.StatusConflict)
				return
			}
			if ss.GS.Terminated || ss.GS.Pending != nil {
				httpError(w, fmt.Errorf("no move to make: the game is over or a move awaits the moderator"), http.StatusConflict)
				return
			}
			// A model's replies are paid for and cannot be had again: it is
			// asked nothing while the game's record cannot be written.
			if ss.saveErr != "" {
				if err := s.autosave(ss); err != nil {
					msg := "the game could not be saved: its model seats wait until it is"
					if who.owner {
						msg += " (" + err.Error() + "; Save now tries again)"
					}
					httpError(w, errors.New(msg), http.StatusConflict)
					return
				}
				ss.Version++
			}
			ss.thinking = true
			ss.Version++ // pollers see the seat thinking
			s.mu.Unlock()
			res, err := s.modelTurn(ss, c)
			s.mu.Lock()
			ss.thinking = false
			if err != nil {
				ss.Version++ // pollers saw the seat thinking
			}
			ss.event("ai", player, c.Profile.Name, res, err)
			s.reply(w, ss, who, res, err)
			return
		}
		res, err := ss.aiMove()
		ss.event("ai", player, "", res, err)
		s.reply(w, ss, who, res, err)
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
		s.reply(w, ss, who, res, err)
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
		s.reply(w, ss, who, res, err)
	case "save":
		if !jsonPost(w, r) {
			return
		}
		if err := s.autosave(ss); err != nil {
			httpError(w, fmt.Errorf("the game could not be saved: %v", err), http.StatusConflict)
			return
		}
		ss.Version++
		writeJSON(w, map[string]interface{}{"state": ss.stateFor(who)})
	case "role":
		writeJSON(w, who.view(ss))
	case "keys":
		writeJSON(w, ss.keysView())
	case "export":
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=wag-game-%s.json", ss.ID))
		if who.owner {
			writeJSON(w, ss.export(s.contexts))
			break
		}
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(ss.export(s.contexts)); err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(s.guestCopy(b.Bytes()))
	case "traces":
		s.writeTraces(w, ss, who.owner)
	default:
		httpError(w, fmt.Errorf("unknown action %q", action), http.StatusNotFound)
	}
}

// reply answers a request that changed the game (or failed to): the result
// and the new state, as the role sees it. Only a change bumps the version
// pollers compare. The record changed even when the request failed (its
// event, a model's replies): it is saved before the error is answered, so
// that a model's paid replies are on the disk before it can be asked again,
// or the game is unsaved and its model seats wait (autosave).
func (s *Server) reply(w http.ResponseWriter, ss *Session, who role, res *game.TurnResult, err error) {
	if err != nil {
		s.saveRecord(ss)
		httpError(w, err, http.StatusConflict)
		return
	}
	ss.Version++
	s.autosave(ss)
	writeJSON(w, map[string]interface{}{"result": res, "state": ss.stateFor(who)})
}

// saveRecord saves a game whose record changed without a change of its state
// (an event, a model's replies): the version is bumped only if the save
// changed what the state says about saving.
func (s *Server) saveRecord(ss *Session) {
	before := ss.saveErr
	s.autosave(ss)
	if ss.saveErr != before {
		ss.Version++
	}
}

// maxAgentTries is how many moves a computer player tries for one move
// before it passes.
const maxAgentTries = 5

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
	seat := ss.Seats[gs.ActivePlayerIdx].Kind
	for try := 0; try < maxAgentTries; try++ {
		t := ss.newTrace(seat)
		t.Attempt = try + 1
		r, err := ss.play(a.ChooseAction(gs.BuildObservation()), t, nil)
		if err != nil {
			return nil, err
		}
		if r.Accepted || r.Pending || r.Terminated {
			return r, nil
		}
	}
	t := ss.newTrace(seat)
	t.Attempt = maxAgentTries + 1
	return ss.play(nascq.ProposedAction{Type: "abstain"}, t, nil)
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
	// The game's own validation: the profile's rules around NASCQ's, so that
	// the preview accepts what a turn accepts (an agent place the game binds).
	va, err := gs.Validate(a)
	if err != nil {
		p.Error = err.Error()
		return p
	}
	tmpl := nascq.ActionTemplate{Type: a.Type, TargetArgID: a.TargetArgID, AnswersCQID: a.AnswersCQID, TargetStmtID: a.TargetStmtID,
		Locus: va.Locus, LocusStmtID: va.LocusStmtID, CaesAttackType: va.CaesAttackType}
	if sc := gs.NASCQ.GetScheme(a.SchemeID); sc != nil {
		tmpl.ConclusionTemplateEN = sc.ConclusionTemplateEN
	}
	p.Effect = gs.Effect(tmpl, va.VariableBindings)
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
	if a.TargetArgID == "" && a.TargetStmtID == "" && p.Conclusion != "" {
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
func (ss *Session) export(contexts []game.Context) map[string]interface{} {
	st := ss.buildState()
	// Where the owner keeps the game is not part of it.
	st.SavedTo, st.SavedRel, st.SaveError, st.Unsaved = "", "", "", false
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
	history := ss.GS.MoveHistory
	if ss.inspectedWith != "" {
		history = ss.playedHistory
	}
	now := time.Now()
	out := map[string]interface{}{
		"format": "wag-export/2", "exported_at": now.UTC(), "build": buildInfo(), "data_sha256": ss.dataHash,
		"title": ss.Title, "source": ss.Source, "seats": ss.Seats, "config": st.Config, "request": req,
		"history": history, "state": st, "other_arguments": mined, "import": ss.Import, "pins": ss.pinsOrNil(),
		"notes": ss.Notes, "events": ss.Events, "llm": ss.LLM, "traces": ss.tracesWithoutPrompts(), "verdicts": ss.verdicts(contexts),
		// The time played so far, which the game clock counts: a resume
		// counts this, not the time since the game was created.
		"played_ns": now.Sub(ss.GS.StartedAt).Nanoseconds(),
	}
	if ev := ss.GS.Evidence; ev != nil {
		// The record set by name and hash, and the game's bundle: the record
		// set's records with the context, actors, targets, bindings and
		// events the game wrote (docs/e1-evidence.md).
		out["record_set"] = ss.recordSet()
		out["evidence"] = ev.Bundle
	}
	if ss.inspectedWith != "" {
		// The history is as it was played, with data_sha256; the state, the
		// verdicts and the other arguments were computed again with this data.
		out["inspected_with_data_sha256"] = ss.inspectedWith
	}
	if ss.legacyReplayedWith != "" {
		// A legacy record: data_sha256 is empty, as in the export it was
		// loaded from; the state was computed with this data. The moves
		// played after its own name the data they were played with.
		out["legacy_replayed_with_data_sha256"] = ss.legacyReplayedWith
		if len(history) > ss.legacyMoves {
			out["legacy_moves"] = ss.legacyMoves
			out["continued_with_data_sha256"] = ss.continuedWith
		}
	}
	return out
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
	if !ownerOnly(w, r) {
		return
	}
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
	if !ownerOnly(w, r) || !jsonPost(w, r) {
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
	gs, rep, err := s.kialoGame(b, depth, kialo.KeysByClaim, game.DefaultProfile())
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	ss := s.add(gs, []Seat{{Name: kialo.ProPlayer, Kind: "human"}, {Name: kialo.ConPlayer, Kind: "human"}}, rep.Title, "kialo:"+rep.Discussion, rand.Int63())
	ss.Import = rep
	st := ss.buildState()
	s.mu.Unlock()
	writeJSON(w, map[string]interface{}{"state": st, "report": rep})
}

// kialoGame builds the game of a Kialo discussion (graph.json), down to depth
// (0: all of it), with its claims keyed as keys says (kialo.KeysByClaim, or
// KeysByText to replay a game imported that way), and with a profile (D094
// (b); nil for a game imported before step C5.2, replayed as it was played).
func (s *Server) kialoGame(b []byte, depth int, keys string, profile *game.Profile) (*game.GameState, *kialo.Report, error) {
	d, err := kialo.Parse(b)
	if err != nil {
		return nil, nil, err
	}
	rels, rep := d.Plan(depth)
	rep.Keys = keys
	ad, err := carneades.NewAdapterFromYAML(s.data.YAML)
	if err != nil {
		return nil, nil, err
	}
	n, err := nascq.LoadFromBytes(s.data.Schemes, s.data.CQs, s.data.Edges)
	if err != nil {
		return nil, nil, err
	}
	cfg := kialo.Config()
	cfg.Profile, rep.Profile = profile, profile
	gs, err := kialo.Build(d, rels, rep, cfg, ad, n)
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

// handleLoad starts a game from an export (/export) and replays its moves;
// ?inspect=1 opens a game played with other scheme data read-only.
func (s *Server) handleLoad(w http.ResponseWriter, r *http.Request) {
	if !ownerOnly(w, r) {
		return
	}
	var e Export
	if !readJSON(w, r, &e) {
		return
	}
	inspect, _ := strconv.ParseBool(r.URL.Query().Get("inspect"))
	s.mu.Lock()
	defer s.mu.Unlock()
	ss, err := s.load(&e, inspect)
	if err != nil {
		var other *OtherDataError
		if errors.As(err, &other) {
			// "inspect": the dashboard offers to open it read-only.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "cannot load the game: " + err.Error(), "inspect": true})
			return
		}
		httpError(w, fmt.Errorf("cannot load the game: %v", err), http.StatusBadRequest)
		return
	}
	detail := fmt.Sprintf("%d moves", len(e.History))
	if ss.inspectedWith != "" {
		detail += fmt.Sprintf("; read-only: played with scheme data %s, replayed with %s", ss.dataHash, ss.inspectedWith)
	}
	if ss.legacyReplayedWith != "" {
		detail += fmt.Sprintf("; a legacy record: the export names no scheme data; replayed with %s", ss.legacyReplayedWith)
	}
	ss.event("load", "", detail, nil, nil)
	writeJSON(w, ss.buildState())
}

func (s *Server) claudeMeta() interface{} {
	c := s.models["claude"]
	if c == nil {
		return nil
	}
	return map[string]string{"model": c.Profile.Model, "effort": c.Profile.Effort}
}
