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
	// "" in exports made before 2026-10-13.
	DataHash string `json:"data_sha256,omitempty"`
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
// are not replayed: a move that timed out is replayed as a pass. Caller
// holds s.mu.
func (s *Server) load(e *Export) (*Session, error) {
	if len(e.Seats) == 0 {
		return nil, fmt.Errorf("the export has no seats")
	}
	req := e.requestOf()
	var gs *game.GameState
	var rep *kialo.Report
	switch {
	case strings.HasPrefix(e.Source, "kialo:"):
		id := strings.TrimPrefix(e.Source, "kialo:")
		if s.kialoDir == "" || strings.ContainsAny(id, `/\.`) {
			return nil, fmt.Errorf("the game was imported from Kialo discussion %s: start the server with --kialo-dir to load it", id)
		}
		b, err := os.ReadFile(filepath.Join(s.kialoDir, id, "graph.json"))
		if err != nil {
			return nil, err
		}
		depth, keys := 0, kialo.KeysByText // an export without keys was imported with claims merged by text
		if e.Import != nil {
			depth = e.Import.Depth
			if e.Import.Keys != "" {
				keys = e.Import.Keys
			}
		}
		if gs, rep, err = s.kialoGame(b, depth, keys); err != nil {
			return nil, err
		}
	default:
		cfg, err := req.config(s.contexts)
		if err != nil {
			return nil, err
		}
		// Modern exports contain the effective weights, including the context
		// preset. Contexts may be changed without changing the scheme-data
		// hash; silently replaying against them would change the experiment.
		// Legacy exports without a recorded request/config keep their existing
		// compatibility path; this is not a claim of exact legacy replay.
		if e.Request != nil && e.Config.Mode != "" && !maps.Equal(cfg.Weights, e.Config.Weights) {
			return nil, fmt.Errorf("the exported effective weights differ from the current context/request: restore the recorded context settings before replaying")
		}
		cfg.TimeLimit, cfg.MoveTimeLimit = 0, 0
		var names []string
		for _, seat := range e.Seats {
			names = append(names, seat.Name)
		}
		if gs, err = s.newState(cfg, names); err != nil {
			return nil, err
		}
	}
	for i, m := range e.History {
		if err := replay(gs, m); err != nil {
			if e.DataHash != s.data.Hash() && !errors.Is(err, errBridgeRetired) {
				return nil, fmt.Errorf("move %d (%s): %v. The game was played with other scheme data (%s; this server's is %s): "+
					"start the server with --data on the data of that version to replay it exactly", i+1, m.PlayerID, err, orUnknown(e.DataHash), s.data.Hash())
			}
			return nil, fmt.Errorf("move %d (%s): %v", i+1, m.PlayerID, err)
		}
	}
	// Legal moves do not establish that the same data produced them. Keep
	// the retired-bridge diagnostic above, but never register a successful
	// replay that has a known data mismatch. Older exports lack this hash.
	if e.DataHash != "" && e.DataHash != s.data.Hash() {
		return nil, fmt.Errorf("the game was played with other scheme data (%s; this server's is %s): start the server with --data on the data of that version to replay it exactly", e.DataHash, s.data.Hash())
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
	ss := s.add(gs, e.Seats, title, source, req.Seed)
	ss.Request, ss.Import, ss.Notes = req, rep, append([]Note(nil), e.Notes...)
	return ss, nil
}

// errBridgeRetired marks a recorded move that a free-text bridge let through.
// Since D075's rule (2026-11-22) a bridge no longer stands in for an answer's
// conclusion, with any data, so other data (--data) cannot replay it either.
var errBridgeRetired = errors.New("it was played through a free-text bridge, which no longer stands in for an answer's conclusion (D075, since 2026-11-22)")

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
	res, err := gs.ExecuteTurn(action)
	if err != nil {
		return err
	}
	if res.Pending {
		verdict, note, _ := strings.Cut(m.Moderation, ": ")
		if res, err = gs.Moderate(verdict == "approved", m.Weight, note); err != nil {
			return err
		}
	}
	if played := res.Accepted || res.Abstained; played != (len(m.ArgIDs) > 0 || m.Abstained || m.Resigned) && !m.TimedOut {
		if m.Bridge != "" && !played {
			return fmt.Errorf("%w: %q. This build cannot replay the game; a WAG build from before that date can, and the export stays readable as a record", errBridgeRetired, m.Bridge)
		}
		return fmt.Errorf("it was recorded as %s, but replayed as %s (%s)", outcome(len(m.ArgIDs) > 0 || m.Abstained), outcome(played), res.Error)
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
