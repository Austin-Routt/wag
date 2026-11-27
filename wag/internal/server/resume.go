package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Resuming a game of the data folder (docs/g-journal.md). A game with a model
// seat is saved into its own folder after every change, and every call to
// the provider is saved before its reply is used (modelTurn). A game resumed
// from that folder carries its trace over: the calls and the decisions
// recorded, so that its trace goes on from them, and a model seat's turn in
// progress goes on from its record: a reply no decision used is played, not
// asked for (and paid for) again. The game goes on saving into its folder.

// savedGame is a game's export as autosave writes it: what load reads, and
// what a resume carries over besides.
type savedGame struct {
	Export
	ExportedAt time.Time `json:"exported_at"`
	State      struct {
		ID string `json:"id"`
	} `json:"state"`
	Traces []Trace     `json:"traces"`
	Events []Event     `json:"events"`
	LLM    []LLMRecord `json:"llm"`
}

// traceFile is a trace file read back: its events, as the server wrote them.
type traceFile struct {
	start     GameStart
	calls     []Call
	decisions []Trace
}

// readTrace reads a trace file the server wrote (encodeTraces): a game_start
// event first, then calls and decisions, and a game_end once the game was
// over. Anything else is refused.
func readTrace(b []byte) (*traceFile, error) {
	tf := &traceFile{}
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	n := 0
	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		n++
		var head struct {
			Format string `json:"format"`
			Type   string `json:"type"`
		}
		if err := json.Unmarshal(line, &head); err != nil {
			return nil, fmt.Errorf("the trace's event %d is not one: %v", n, err)
		}
		if head.Format != TraceFormat {
			return nil, fmt.Errorf("the trace's event %d is of the format %q, not %s", n, head.Format, TraceFormat)
		}
		if (n == 1) != (head.Type == "game_start") {
			return nil, fmt.Errorf("the trace's event %d is a %s: a trace begins with the game's start, once", n, head.Type)
		}
		var err error
		switch head.Type {
		case "game_start":
			err = json.Unmarshal(line, &tf.start)
		case "call":
			var c Call
			err = json.Unmarshal(line, &c)
			tf.calls = append(tf.calls, c)
		case "decision":
			var d Trace
			err = json.Unmarshal(line, &d)
			tf.decisions = append(tf.decisions, d)
		case "game_end":
		default:
			err = fmt.Errorf("an event of the unknown type %q", head.Type)
		}
		if err != nil {
			return nil, fmt.Errorf("the trace's event %d: %v", n, err)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading the trace: %v", err)
	}
	if n == 0 {
		return nil, fmt.Errorf("the trace is empty")
	}
	return tf, nil
}

// check tells why the trace is not the export's, or "" if it is, with the
// decisions to keep: the export's, which are the trace's first ones. A save
// writes the trace, then the export, so a save that stopped between the two
// leaves the trace ahead of the export; its later decisions are dropped
// (dropped), and their replies are played again (turnInProgress).
func (tf *traceFile) check(g *savedGame) (kept []Trace, dropped int, why string) {
	st, game := &tf.start, tf.start.Game
	switch {
	case g.State.ID != game:
		return nil, 0, fmt.Sprintf("the trace is of the game %s, the export of the game %s", game, orUnknown(g.State.ID))
	case st.DataSHA256 != g.DataHash:
		return nil, 0, fmt.Sprintf("the trace was played with the scheme data %s, the export with %s", orUnknown(st.DataSHA256), orUnknown(g.DataHash))
	case !reflect.DeepEqual(st.Seats, g.Seats):
		return nil, 0, fmt.Sprintf("the trace's seats %v are not the export's %v", st.Seats, g.Seats)
	case (st.Request == nil) != (g.Request == nil):
		return nil, 0, "one of the trace and the export records the game's request, the other does not"
	case g.Request != nil && g.Request.Seed != st.Seed:
		return nil, 0, fmt.Sprintf("the trace's seed is %d, the export's %d", st.Seed, g.Request.Seed)
	}
	if g.Request != nil {
		// The export's request is the game's, with the seed the game drew.
		a, b := *st.Request, *g.Request
		a.Seed, b.Seed = 0, 0
		if !sameJSON(a, b) {
			return nil, 0, "the trace's request is not the export's"
		}
	}
	for _, c := range tf.calls {
		if c.Game != game {
			return nil, 0, fmt.Sprintf("the trace's call %s is of the game %s", c.ID, c.Game)
		}
		if c.Turn > len(g.History) {
			return nil, 0, fmt.Sprintf("the trace's call %s was made after %d moves, the export has %d", c.ID, c.Turn, len(g.History))
		}
	}
	for _, d := range tf.decisions {
		if d.Game != game {
			return nil, 0, fmt.Sprintf("the trace's decision %s is of the game %s", d.ID, d.Game)
		}
	}
	if len(g.Traces) > len(tf.decisions) {
		return nil, 0, fmt.Sprintf("the export has %d decisions, the trace %d", len(g.Traces), len(tf.decisions))
	}
	for i, x := range g.Traces {
		d := tf.decisions[i]
		d.Prompt = "" // the export leaves the prompts out
		if !sameJSON(x, d) {
			return nil, 0, fmt.Sprintf("the export's decision %s is not the trace's", x.ID)
		}
	}
	kept = tf.decisions[:len(g.Traces)]
	for _, d := range kept {
		if d.Move < 0 {
			continue
		}
		if d.Move >= len(g.History) {
			return nil, 0, fmt.Sprintf("the trace's decision %s made the move %d, which the export does not have", d.ID, d.Move+1)
		}
		m := g.History[d.Move]
		if d.Player != m.PlayerID || !sameJSON(d.Action, m.Action) {
			return nil, 0, fmt.Sprintf("the trace's decision %s made the move %d as %s's, the export has it as %s's, or another move", d.ID, d.Move+1, d.Player, m.PlayerID)
		}
	}
	return kept, len(tf.decisions) - len(kept), ""
}

// sameJSON: a and b are written alike.
func sameJSON(a, b any) bool {
	x, err1 := json.Marshal(a)
	y, err2 := json.Marshal(b)
	return err1 == nil && err2 == nil && bytes.Equal(x, y)
}

// startDiffers names what of the game's start, as this server would write
// it for the session, is not as the trace's start has it: "" if nothing.
func (s *Server) startDiffers(ss *Session, recorded *GameStart) string {
	var got, want map[string]json.RawMessage
	a, err1 := json.Marshal(s.gameStart(ss))
	b, err2 := json.Marshal(recorded)
	if err1 != nil || err2 != nil || json.Unmarshal(a, &got) != nil || json.Unmarshal(b, &want) != nil {
		return "the start (it cannot be compared)"
	}
	var diff []string
	for k, v := range got {
		if !bytes.Equal(v, want[k]) {
			diff = append(diff, k)
		}
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			diff = append(diff, k)
		}
	}
	sort.Strings(diff)
	return strings.Join(diff, ", ")
}

// resume opens the game saved in dir (a folder of the data folder's games):
// its export, replayed as a load replays one, and its trace, which must be
// the export's, started as this server would start the game. The game keeps
// its id, its start and its clocks, and goes on saving into dir. Caller
// holds s.mu.
func (s *Server) resume(dir string) (*Session, error) {
	eb, err := os.ReadFile(filepath.Join(dir, "export.json"))
	if err != nil {
		return nil, fmt.Errorf("the game's folder has no export: %v", err)
	}
	var g savedGame
	if err := json.Unmarshal(eb, &g); err != nil {
		return nil, fmt.Errorf("the game's export does not read: %v", err)
	}
	tb, err := os.ReadFile(filepath.Join(dir, "trace.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("the game's folder has no trace: %v", err)
	}
	tf, err := readTrace(tb)
	if err != nil {
		return nil, err
	}
	kept, dropped, why := tf.check(&g)
	if why != "" {
		return nil, fmt.Errorf("the trace is not the export's: %s", why)
	}
	id := tf.start.Game
	if s.sessions[id] != nil {
		return nil, fmt.Errorf("the game %s is open in this server: go on playing it here", id)
	}
	ss, err := s.loadAs(&g.Export, false, kept, tf.start.Seed)
	if err != nil {
		return nil, err
	}
	delete(s.sessions, ss.ID) // registered under its own id below, if it resumes
	ss.ID, ss.Created, ss.Source, ss.Request = id, tf.start.Time, g.Source, tf.start.Request
	ss.resumed = tf.start.Resumed
	// Its clocks, which a load drops: the move clock starts again for the
	// turn in progress; the game clock counts the time played before the
	// stop, from the start to the export's last save.
	gs := ss.GS
	gs.Config.TimeLimit, gs.Config.MoveTimeLimit = tf.start.Config.TimeLimit, tf.start.Config.MoveTimeLimit
	played := g.ExportedAt.Sub(tf.start.Time)
	if played < 0 || g.ExportedAt.IsZero() {
		played = 0
	}
	now := time.Now()
	gs.StartedAt, gs.TurnStartedAt = now.Add(-played), now
	if diff := s.startDiffers(ss, &tf.start); diff != "" {
		return nil, fmt.Errorf("this server would not go on with the game as its trace began it: its %s differ. Resume it on a server set up as "+
			"the game was (the same WAG build, model profiles and context presets), or load its export to go on with it as a new record", diff)
	}
	s.sessions[id] = ss
	if n, err := strconv.Atoi(id); err == nil && n > s.next {
		s.next = n
	}
	ss.folder, ss.Calls, ss.Traces = dir, tf.calls, kept
	ss.order = 0
	for _, c := range ss.Calls {
		ss.order = max(ss.order, c.Order)
	}
	for _, d := range ss.Traces {
		ss.order = max(ss.order, d.Order)
	}
	ss.Events, ss.LLM = g.Events, g.LLM
	ss.resumed = append(ss.resumed, Resumption{Time: now.UTC(), After: ss.order, Dropped: dropped})
	detail := fmt.Sprintf("%d moves; %d calls and %d decisions carried over", len(g.History), len(ss.Calls), len(kept))
	if dropped > 0 {
		detail += fmt.Sprintf("; %d decisions dropped, recorded after the export's last save", dropped)
	}
	ss.event("resume", "", detail, nil, nil)
	s.autosave(ss) // a failure is shown in the state, and the model seats wait
	return ss, nil
}
