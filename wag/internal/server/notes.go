package server

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/game"
)

// Note is a tester's remark on a game: on a move of the history, an
// argument or a statement, or on the game as a whole. Notes travel with the
// export, so the person who reads it sees what the player saw and thought.
type Note struct {
	ID     int       `json:"id"`
	Target string    `json:"target,omitempty"` // "move:3", an argument ID, a statement ID, or "" for the game
	Kind   string    `json:"kind"`             // wrong, confusing, unreadable, good, idea, note
	Text   string    `json:"text"`
	By     string    `json:"by,omitempty"`
	Round  int       `json:"round"`
	At     time.Time `json:"at"`
}

var noteKinds = map[string]bool{"wrong": true, "confusing": true, "unreadable": true, "good": true, "idea": true, "note": true}

// Event is one request that tried to change the game, whether or not it
// did: what a tester attempted is as useful as what was played.
type Event struct {
	At     time.Time `json:"at"`
	What   string    `json:"what"` // move, ai, moderate, decide, reply, note, load
	Player string    `json:"player,omitempty"`
	Action string    `json:"action,omitempty"` // a short form of the move
	OK     bool      `json:"ok"`
	Error  string    `json:"error,omitempty"`
}

// event records a request and its outcome. Caller holds the server lock.
func (ss *Session) event(what, player, action string, res *game.TurnResult, err error) {
	e := Event{At: time.Now(), What: what, Player: player, Action: action, OK: err == nil}
	switch {
	case err != nil:
		e.Error = err.Error()
	case res != nil && !res.Accepted && !res.Abstained && !res.Pending:
		e.OK, e.Error = false, res.Error
	}
	ss.Events = append(ss.Events, e)
	if len(ss.Events) > 2000 {
		ss.Events = ss.Events[len(ss.Events)-2000:]
	}
}

// addNote adds a tester's note. Caller holds the server lock.
func (ss *Session) addNote(target, kind, text, by string) (Note, error) {
	text, kind = strings.TrimSpace(text), strings.TrimSpace(kind)
	if kind == "" {
		kind = "note"
	}
	if !noteKinds[kind] {
		return Note{}, fmt.Errorf("unknown kind of note %q (wrong, confusing, unreadable, good, idea, note)", kind)
	}
	if text == "" && kind == "note" {
		return Note{}, fmt.Errorf("the note is empty")
	}
	n := Note{ID: len(ss.Notes) + 1, Target: strings.TrimSpace(target), Kind: kind, Text: text, By: strings.TrimSpace(by),
		Round: ss.GS.CurrentTurn, At: time.Now()}
	ss.Notes = append(ss.Notes, n)
	return n, nil
}

// buildInfo says which build of WAG made an export or a trace: the commit,
// whether the working tree had changes, and the model SDKs' versions.
func buildInfo() map[string]string {
	out := map[string]string{}
	if bi, ok := debug.ReadBuildInfo(); ok {
		out["go"] = bi.GoVersion
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				out["commit"] = s.Value
			case "vcs.time":
				out["commit_time"] = s.Value
			case "vcs.modified":
				out["modified"] = s.Value
			case "GOOS", "GOARCH":
				out[strings.ToLower(s.Key)] = s.Value
			}
		}
		for _, d := range bi.Deps {
			switch d.Path {
			case "google.golang.org/genai", "github.com/anthropics/anthropic-sdk-go":
				out[d.Path] = d.Version
			}
		}
	}
	return out
}
