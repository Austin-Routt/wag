package server

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/kialo"
	"github.com/yourorg/wag/internal/language"
	"github.com/yourorg/wag/internal/manifest"
	"github.com/yourorg/wag/internal/nascq"
)

// Export is a game as /export writes it and /api/games/load reads it back.
// Only the fields needed to replay the game are read; the rest of the export
// (the state, the events, the notes) is for people.
type Export struct {
	Title   string          `json:"title"`
	Source  string          `json:"source"`
	Seats   []Seat          `json:"seats"`
	Config  ConfigView      `json:"config"`
	Request *NewGameRequest `json:"request,omitempty"`
	History []game.Move     `json:"history"`
	Import  *kialo.Report   `json:"import"`
	Notes   []Note          `json:"notes,omitempty"`
	// DataHash identifies the scheme data the game was played with (Data.Hash);
	// "" in exports made before 2026-10-13, and in a legacy record's own.
	DataHash string `json:"data_sha256,omitempty"`
	// A legacy record's export (one loaded from an export that named no
	// data): the data it was replayed with, and, if it was continued, how
	// many of its moves came before and the data the later ones were played
	// with.
	LegacyReplayedWith string `json:"legacy_replayed_with_data_sha256,omitempty"`
	LegacyMoves        *int   `json:"legacy_moves,omitempty"`
	ContinuedWith      string `json:"continued_with_data_sha256,omitempty"`
	// Pins are the inputs the game was played with, by content (E1.1);
	// RecordSet names its record set (docs/e1-evidence.md). A game with a
	// record set is replayed only with this server's pins equal to its own
	// and its record set read again (loadEvidence).
	Pins      *manifest.Pins `json:"pins,omitempty"`
	RecordSet *RecordSet     `json:"record_set,omitempty"`
}

// requestOf is the request that starts the exported game: the one recorded
// with it or, for exports that predate that, one made from its settings.
func (e *Export) requestOf() *NewGameRequest {
	if e.Request != nil {
		r := *e.Request
		r.Seats = e.Seats
		return &r
	}
	c := e.Config
	rounds, within, concede := c.MaxRounds, c.AnswerWithin, c.ConcedeOnPass
	r := &NewGameRequest{Title: e.Title, Mode: c.Mode, Seats: e.Seats, Scoring: c.Scoring, MaxRounds: &rounds,
		AnswerWithin: &within, ConcedeOnPass: &concede, Standard: c.Standard,
		AllowOther: c.AllowOther, Moderated: c.Moderated, RequireEvidence: c.RequireEvidence} // AnyPremise: the default (an export without a request predates it)
	if c.TargetScore > 0 {
		t := c.TargetScore
		r.TargetScore = &t
	}
	return r
}

// load starts a game from an export and replays its moves, so that a game
// someone played can be looked at, continued, or turned into a test. Clocks
// are not replayed: a move that timed out is replayed as a pass. A game
// played with other scheme data than this server's, as its fingerprint
// says, is refused (an *OtherDataError), unless inspect asks to open it
// read-only: its moves are then replayed with this server's data and it
// takes no new move. Caller holds s.mu.
func (s *Server) load(e *Export, inspect bool) (*Session, error) {
	return s.loadAs(e, inspect, nil, 0)
}

// loadAs loads a game as load does, for a resume (resume.go): decisions are
// the game's recorded decisions, for which its computer players draw again
// what they drew (redraw), and seed, if not 0, is the game's seed as its
// trace records it.
func (s *Server) loadAs(e *Export, inspect bool, decisions []Trace, seed int64) (*Session, error) {
	if len(e.Seats) == 0 {
		return nil, fmt.Errorf("the export has no seats")
	}
	req := e.requestOf()
	// An export that names no scheme data is a legacy record: its moves were
	// played with data no one recorded. One continued after it was loaded
	// names the data its later moves were played with, and where they begin.
	legacy := e.DataHash == ""
	legacyMoves, continued := len(e.History), false
	switch {
	case !legacy && (e.LegacyMoves != nil || e.ContinuedWith != ""):
		return nil, fmt.Errorf("the export names its scheme data (%s) and also a legacy record's continuation", e.DataHash)
	case legacy && e.LegacyMoves != nil:
		if *e.LegacyMoves < 0 || *e.LegacyMoves > len(e.History) {
			return nil, fmt.Errorf("the export's legacy_moves (%d) is not between 0 and its %d moves", *e.LegacyMoves, len(e.History))
		}
		legacyMoves = *e.LegacyMoves
		continued = legacyMoves < len(e.History)
		if continued && e.ContinuedWith == "" {
			return nil, fmt.Errorf("the export continues a legacy record after its move %d but names no data for the moves after it", legacyMoves)
		}
	case legacy && e.ContinuedWith != "":
		return nil, fmt.Errorf("the export names the data a legacy record was continued with, but not where (legacy_moves)")
	}
	playedWith := e.DataHash
	if continued {
		playedWith = e.ContinuedWith
	}
	var gs *game.GameState
	var rep *kialo.Report
	switch {
	case strings.HasPrefix(e.Source, "kialo:"):
		if req.Evidence != "" || e.RecordSet != nil {
			return nil, fmt.Errorf("a game imported from Kialo takes no record set")
		}
		id := strings.TrimPrefix(e.Source, "kialo:")
		if s.kialoDir == "" || strings.ContainsAny(id, `/\.`) {
			return nil, fmt.Errorf("the game was imported from Kialo discussion %s: start the server with --kialo-dir to load it", id)
		}
		b, err := os.ReadFile(filepath.Join(s.kialoDir, id, "graph.json"))
		if err != nil {
			return nil, err
		}
		depth, keys := 0, kialo.KeysByText // an export without keys was imported with claims merged by text
		var profile *game.Profile          // none: imported before step C5.2
		if e.Import != nil {
			depth, profile = e.Import.Depth, e.Import.Profile
			if e.Import.Keys != "" {
				keys = e.Import.Keys
			}
		}
		if gs, rep, err = s.kialoGame(b, depth, keys, profile); err != nil {
			return nil, namelessName(err)
		}
	default:
		// An export that records the scheme data (since 13 October 2026)
		// records the weights in effect too, the context preset's and the
		// game's own: the game is replayed with those, as it was played,
		// whatever the preset says now (contexts are not in the data's
		// fingerprint). An older export has only the request's context.
		r := *req
		recorded := e.DataHash != "" || e.LegacyReplayedWith != ""
		if recorded {
			r.Context, r.Weights = "", nil
		}
		cfg, err := r.config(s.contexts)
		if errors.Is(err, game.ErrReservedClaim) {
			return nil, fmt.Errorf("%w. WAG accepted such a claim until 2026-11-28 and refuses it since: this build cannot load the game; a WAG build of commit f38d189 or earlier can, "+
				"with the scheme data the game was played with (--data, if they are not that build's own), and the export stays readable as a record", err)
		}
		if err != nil {
			return nil, namelessName(err)
		}
		if recorded {
			if err := game.CheckWeights(e.Config.Weights); err != nil {
				return nil, fmt.Errorf("the export's weights: %v", err)
			}
			cfg.Context, cfg.Weights = e.Config.Context, maps.Clone(e.Config.Weights)
		}
		// The checks of a new game (docs/e1-evidence.md, E1.7, B4): an
		// inquiry of two, without the settings that make a winner or a
		// loser; a record set only in an inquiry of people and models that
		// is not moderated.
		if cfg.Cooperative {
			if err := game.CheckInquiry(cfg, len(e.Seats)); err != nil {
				return nil, err
			}
		}
		if req.Evidence != "" || e.RecordSet != nil {
			if err := checkEvidenceGame(cfg, e.Seats); err != nil {
				return nil, err
			}
		}
		cfg.TimeLimit, cfg.MoveTimeLimit = 0, 0
		var names []string
		for _, seat := range e.Seats {
			names = append(names, seat.Name)
		}
		if gs, err = s.newState(cfg, names); err != nil {
			return nil, err
		}
		if req.Evidence != "" || e.RecordSet != nil {
			if gs.Evidence, err = s.loadEvidence(e, req, gs, inspect); err != nil {
				return nil, err
			}
		}
	}
	same := s.data.Matches(playedWith)
	if seed == 0 {
		seed = req.Seed
	}
	var rd *redraw
	if decisions != nil {
		rd = newRedraw(gs, e.Seats, seed, decisions)
	}
	for i, m := range e.History {
		rd.draw(gs, i)
		if err := replay(gs, m); err != nil {
			if !same && !errors.Is(err, errBridgeRetired) && !errors.Is(err, errNameless) && !errors.Is(err, errExceptionForm) {
				return nil, fmt.Errorf("move %d (%s): %v. The game was played with other scheme data (%s; this server's is %s): "+
					"start the server with --data on the data of that version to replay it exactly", i+1, m.PlayerID, err, orUnknown(playedWith), s.data.Hash())
			}
			// The bound by the exception's form is the data's: data without a
			// question that refuses Other bounds nothing (A9.13c; the check of
			// F1, F2 and F4, W2), so this build replays such a game on its data.
			if !same && errors.Is(err, errExceptionForm) {
				return nil, fmt.Errorf("move %d (%s): %w. The game was played with other scheme data (%s; this server's is %s): "+
					"this build replays it with --data on the data of that version", i+1, m.PlayerID, err, orUnknown(playedWith), s.data.Hash())
			}
			return nil, fmt.Errorf("move %d (%s): %w", i+1, m.PlayerID, err)
		}
	}
	rd.draw(gs, len(e.History))
	// Legal moves do not establish that the same data produced them: a game
	// known to be played with other data is not registered as if it had
	// been played with this, unless it is opened read-only. A legacy record
	// names no data for its first moves; a continued one names the data of
	// the others.
	other := playedWith != "" && !same
	if other && !inspect {
		return nil, &OtherDataError{PlayedWith: playedWith, Here: s.data.Hash()}
	}
	title := e.Title
	if title == "" {
		title = "WAG — " + gs.Config.Mode
	}
	source := "loaded"
	if strings.HasPrefix(e.Source, "kialo:") {
		// A later export needs this origin to reconstruct the imported graph
		// before replaying its moves. "loaded" alone loses that distinction.
		source = e.Source
	}
	ss := s.add(gs, e.Seats, title, source, seed)
	ss.Request, ss.Import, ss.Notes = req, rep, append([]Note(nil), e.Notes...)
	if rd != nil {
		ss.agents = rd.agents
	}
	if legacy {
		// It stays a legacy record: replayed with this data, but not taken
		// for a game played with it. Moves played from now on are played
		// with this data, which its export names for them.
		ss.dataHash, ss.legacyReplayedWith, ss.legacyMoves, ss.continuedWith = "", s.data.Hash(), legacyMoves, s.data.Hash()
		if continued && other {
			ss.continuedWith = e.ContinuedWith
		}
	}
	if other {
		// Its export keeps the data it was played with, and the history as
		// played.
		ss.dataHash, ss.inspectedWith = e.DataHash, s.data.Hash()
		ss.playedHistory = append([]game.Move(nil), e.History...)
	}
	return ss, nil
}

// redraw makes a resumed game's computer players draw again what they drew
// before it stopped (docs/g-journal.md): each computer player, seeded as the
// game's, chooses again for each recorded decision of its seat that chose a
// move (aiMove's tries, not its pass), on the board it chose on, so that it
// goes on as in a game that did not stop.
type redraw struct {
	agents []game.Agent
	at     map[int][]int // by the number of moves made: the seats that chose then, in order
}

func newRedraw(gs *game.GameState, seats []Seat, seed int64, decisions []Trace) *redraw {
	rd := &redraw{at: map[int][]int{}}
	seat := map[string]int{}
	for i, st := range seats {
		rd.agents = append(rd.agents, newAgent(st.Kind, gs.Players[i].ID, seed+int64(i)))
		seat[gs.Players[i].ID] = i
	}
	for _, d := range decisions {
		if i, ok := seat[d.Player]; ok && rd.agents[i] != nil && d.Attempt <= maxAgentTries {
			rd.at[d.Turn] = append(rd.at[d.Turn], i)
		}
	}
	return rd
}

// draw has the seats that chose when n moves were made choose again.
func (rd *redraw) draw(gs *game.GameState, n int) {
	if rd == nil {
		return
	}
	for _, i := range rd.at[n] {
		rd.agents[i].ChooseAction(gs.BuildObservation())
	}
}

// OtherDataError refuses to load a game played with other scheme data than
// the server's; /api/load?inspect=1 opens it read-only.
type OtherDataError struct{ PlayedWith, Here string }

func (e *OtherDataError) Error() string {
	return fmt.Sprintf("the game was played with other scheme data (%s; this server's is %s): start the server with --data on the data of that version "+
		"to replay it exactly, or open it read-only to look at it with this server's data", e.PlayedWith, e.Here)
}

// errBridgeRetired marks a recorded move that a free-text bridge let through.
// Since D075's rule (2026-11-22) a bridge no longer stands in for an answer's
// conclusion, with any data, so other data (--data) cannot replay it either.
var errBridgeRetired = errors.New("it was played through a free-text bridge, which no longer stands in for an answer's conclusion (D075, since 2026-11-22)")

// errNameless marks a game recorded with a value or a name that names
// nothing (no letter or digit: blank, punctuation, invisible characters): a
// move's value, an actor, an action or an alias of its profile or of a
// declaration. WAG took such a value as a name until 2026-11-26, and
// refuses it since step C5.2's item F, on 2026-11-27 (S2 for a variable, P1
// for an action or a declaration, the profile's check), with any data.
var errNameless = errors.New("the game records a value or a name that names nothing (no letter or digit), which WAG accepted until 2026-11-26 and refuses since 2026-11-27 (S2, P1, the profile's check: docs/game-rules.md)")

// errExceptionForm marks a recorded move that argued or presumed a statement
// of a protected exception's form with a scheme that does not back its
// question, before any bridge on the board had the statement as its
// exception. WAG accepted it until 2026-11-28 (commit 6916369 and before),
// when S5 looked only at the bridges on the board, and refuses it since, by
// the exception's form (D099 (4); F1 of the external review of exchange 3),
// with any data. S5's refusal says so (nascq.FormOnly).
var errExceptionForm = errors.New("it argued or presumed a statement of a protected exception's form before any bridge had it as its exception, " +
	"which WAG accepted until 2026-11-28 and refuses since, by the exception's form (S5, D099 (4): docs/game-rules.md)")

// namelessRemedy says what replays a game errNameless refuses.
func namelessRemedy(what string) error {
	return fmt.Errorf("%w: %s. This build cannot replay the game; a WAG build of 2026-11-26 or earlier can, with the scheme data the game was played with "+
		"(--data, if they are not that build's own), and the export stays readable as a record", errNameless, what)
}

// namelessName names the remedy for a profile whose names name nothing.
func namelessName(err error) error {
	if errors.Is(err, game.ErrNamesNothing) {
		return namelessRemedy(err.Error())
	}
	return err
}

// namelessRefusal says what the refusal of a recorded move refuses, if it
// refuses a value or a name of the move that names nothing: a variable S2
// lists, or P1 names, whose value names nothing; or a declaration P1 refuses
// for a name that names nothing.
func namelessRefusal(a nascq.ProposedAction, refusal string) (string, bool) {
	switch {
	case strings.HasPrefix(refusal, "S2:"):
		for _, f := range strings.FieldsFunc(refusal, func(r rune) bool { return r == '[' || r == ']' || r == ' ' || r == ';' || r == ':' }) {
			if x, ok := a.VariableBindings[f]; ok && language.NamesNothing(x) {
				return fmt.Sprintf("%s = %q", f, x), true
			}
		}
	case strings.HasPrefix(refusal, "P1:"):
		for _, k := range sortedKeys(a.VariableBindings) {
			if x := a.VariableBindings[k]; language.NamesNothing(x) && strings.HasPrefix(refusal, fmt.Sprintf("P1: %s = ", k)) {
				return fmt.Sprintf("%s = %q", k, x), true
			}
		}
		if strings.Contains(refusal, "names nothing") {
			for _, d := range a.Declare {
				for _, n := range append([]string{d.ID, d.Actor}, d.Aliases...) {
					if language.NamesNothing(n) {
						return fmt.Sprintf("the declaration of %q names %q", d.ID, n), true
					}
				}
			}
		}
	}
	return "", false
}

// replay plays one recorded move again and checks that it went as recorded.
func replay(gs *game.GameState, m game.Move) error {
	if m.PlayerID == game.ModeratorID {
		if m.Action.Type != "decide" {
			return fmt.Errorf("unknown moderator record %q", m.Action.Type)
		}
		_, err := gs.Decide(m.DecisionStmt, m.Decision, m.Action.Text)
		return err
	}
	if p := gs.ActivePlayer(); p == nil || p.ID != m.PlayerID {
		return fmt.Errorf("recorded for %s, but it is the turn of %s", m.PlayerID, gs.ActivePlayer().ID)
	}
	action := m.Action
	if m.TimedOut {
		action = nascq.ProposedAction{Type: "abstain"}
	}
	if m.Evidence != nil {
		gs.ReplayActAt(m.Evidence.Time) // its records are dated as the act was made
	}
	res, err := gs.ExecuteTurn(action)
	if err != nil {
		return err
	}
	if res.Pending {
		verdict, note, _ := strings.Cut(m.Moderation, ": ")
		approve := verdict == "approved"
		// A move the moderator rejected says so. Until 2026-11-27 WAG gave
		// it, as its verdict, the approval of the next held move if that
		// approval came after the game's time was up (the game ended
		// without the move).
		if why, ok := strings.CutPrefix(m.Rejection, "rejected by the moderator"); ok {
			approve, note = false, strings.TrimPrefix(why, ": ")
		}
		if res, err = gs.Moderate(approve, m.Weight, note); err != nil {
			return err
		}
	}
	recorded := len(m.ArgIDs) > 0 || m.Abstained || m.Resigned || m.Evidence != nil
	if played := res.Accepted || res.Abstained; played != recorded && !m.TimedOut {
		if m.Bridge != "" && !played {
			return fmt.Errorf("%w: %q. This build cannot replay the game; a WAG build from before that date can, and the export stays readable as a record", errBridgeRetired, m.Bridge)
		}
		if what, ok := namelessRefusal(m.Action, res.Error); ok && !played {
			return namelessRemedy(what)
		}
		if !played && strings.HasPrefix(res.Error, "S5:") && strings.Contains(res.Error, nascq.FormOnly) {
			return fmt.Errorf("%w: %s. This build cannot replay the game; a WAG build of commit 6916369 or earlier can, with the scheme data the game was played with "+
				"(--data, if they are not that build's own), and the export stays readable as a record",
				errExceptionForm, strings.TrimSuffix(res.Error, nascq.FormOnly))
		}
		return fmt.Errorf("it was recorded as %s, but replayed as %s (%s)", outcome(recorded), outcome(played), res.Error)
	}
	if m.Evidence != nil {
		// A replayed act makes or names the revisions recorded
		// (docs/e1-evidence.md, choice 6).
		if got := gs.MoveHistory[len(gs.MoveHistory)-1].Evidence; got == nil || *got != *m.Evidence {
			return fmt.Errorf("the evidence act was recorded as %+v, but replayed as %+v", *m.Evidence, got)
		}
	}
	return nil
}

func outcome(played bool) string {
	if played {
		return "played"
	}
	return "rejected"
}

func orUnknown(s string) string {
	if s == "" {
		return "not recorded"
	}
	return s
}
