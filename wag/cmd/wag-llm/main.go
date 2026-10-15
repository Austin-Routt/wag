// wag-llm — language models play WAG, headless, for testing, debugging and
// collecting data.
//
// Usage:
//
//	wag-llm providers
//	wag-llm models PROVIDER [--base-url URL] [--api-key-env VAR]
//	wag-llm ping --model SPEC
//	wag-llm play --model SPEC… --pro SEAT --con SEAT --out DIR [options]
//	wag-llm replay TRACE.jsonl [--data DIR]
//	wag-llm report DIR
//	wag-llm annotate --model SPEC… [--reviewer NAME] --out DIR [--dataset ID|DIR] [--limit N]
//
// A SPEC is a model profile on the command line,
// [NAME=]PROVIDER:MODEL[,option=value…], for example
// gemini:gemini-2.5-flash,temperature=0.7,min_interval=6s; --models FILE
// reads profiles from a JSON file (docs/llm.md). A SEAT is a profile's name,
// greedy or random (the rule bots).
//
// Keys are read from environment variables (GEMINI_API_KEY for Gemini;
// docs/llm.md has the others), never from flags or files.
//
// play runs games through the same server code as the browser and writes,
// for each game, its trace file (wag-trace/2: every call with the whole
// conversation, every decision, the start and the end) and its export, and
// a line of summary.jsonl. replay plays a trace file again with the
// recorded answers, and says whether the game sent the same requests and
// made the same decisions. report sums up summary.jsonl by model and side.
// annotate has models annotate a dataset's claims (the built-in Kialo batch,
// or a folder), and a reviewer model review their disagreements; the run's
// folder holds everything. The dashboard does the same (wag-server, Annotate).
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	var err error
	switch os.Args[1] {
	case "providers":
		for _, n := range llm.ProviderNames() {
			fmt.Printf("%-18s %s\n", n, llm.Providers[n])
		}
	case "models":
		err = models(ctx, os.Args[2:])
	case "ping":
		err = ping(ctx, os.Args[2:])
	case "play":
		err = play(ctx, os.Args[2:])
	case "replay":
		err = replay(ctx, os.Args[2:])
	case "report":
		err = report(os.Args[2:])
	case "annotate":
		err = annotateCmd(ctx, os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "wag-llm:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: wag-llm providers | models PROVIDER | ping --model SPEC | play … | replay TRACE.jsonl | report DIR | annotate …
(go doc ./cmd/wag-llm, or docs/llm.md)`)
	os.Exit(2)
}

func models(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("models", flag.ExitOnError)
	base := fs.String("base-url", "", "the endpoint of an OpenAI-compatible server")
	keyEnv := fs.String("api-key-env", "", "the environment variable that holds the key")
	if len(args) == 0 {
		return fmt.Errorf("models PROVIDER (%s)", strings.Join(llm.ProviderNames(), ", "))
	}
	_ = fs.Parse(args[1:])
	p := llm.Profile{Name: args[0], Provider: args[0], BaseURL: *base, APIKeyEnv: *keyEnv}
	var list []map[string]any
	var err error
	switch p.Provider {
	case "gemini":
		list, err = llm.GeminiModels(ctx, p)
	case "anthropic", "command", "mock":
		return fmt.Errorf("listing %s models is not supported", p.Provider)
	default:
		list, err = llm.OpenAIModels(ctx, p)
	}
	for _, m := range list {
		b, _ := json.Marshal(m)
		fmt.Println(string(b))
	}
	return err
}

func profilesFlags(fs *flag.FlagSet) (*llm.Specs, *string) {
	var specs llm.Specs
	fs.Var(&specs, "model", "a model profile: [NAME=]PROVIDER:MODEL[,option=value…] (repeatable)")
	file := fs.String("models", "", "a JSON file of model profiles")
	return &specs, file
}

func ping(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("ping", flag.ExitOnError)
	specs, file := profilesFlags(fs)
	_ = fs.Parse(args)
	ps, err := llm.ProfilesFrom(*file, *specs)
	if err != nil {
		return err
	}
	if len(ps) == 0 {
		return fmt.Errorf("ping --model SPEC")
	}
	for _, p := range ps {
		if n := p.Notice(); n != "" {
			fmt.Println("Note:", n)
		}
		c, err := llm.NewClient(p)
		if err != nil {
			return err
		}
		req := c.Request("Answer in one word.", []llm.Message{{Role: "user", Content: "Reply with the single word: ready"}})
		resp, err := c.Complete(ctx, req, func(a llm.Attempt) {
			b, _ := json.Marshal(map[string]any{"profile": a.Profile, "try": a.N, "latency_ms": a.Latency.Milliseconds(), "error": a.Error, "response": a.Response})
			fmt.Println(string(b))
		})
		if err != nil {
			return err
		}
		fmt.Printf("%s: %q from %s\n", p.Name, strings.TrimSpace(resp.Text), resp.Model)
	}
	return nil
}

// game drives one game of a server through its HTTP API, as the browser
// does, until it is over or maxMoves moves were asked for.
type game struct {
	base string
	id   string
}

func call(base, method, path string, body, out any) error {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, base+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s %s: %s", method, path, strings.TrimSpace(string(b)))
	}
	if out != nil {
		return json.Unmarshal(b, out)
	}
	return nil
}

type stateView struct {
	ID       string `json:"id"`
	Over     bool   `json:"over"`
	ActiveAI string `json:"active_ai"`
	Active   string `json:"active"`
	Pending  any    `json:"pending"`
}

// run plays the game until it is over; the error is why it stopped early.
func (g *game) run(maxMoves int) error {
	for n := 0; ; n++ {
		var st stateView
		if err := call(g.base, "GET", "/api/games/"+g.id, nil, &st); err != nil {
			return err
		}
		if st.Over {
			return nil
		}
		if n >= maxMoves {
			return fmt.Errorf("stopped after %d moves (--max-moves)", maxMoves)
		}
		if st.ActiveAI == "" {
			return fmt.Errorf("%s is a human seat", st.Active)
		}
		if err := call(g.base, "POST", "/api/games/"+g.id+"/ai", map[string]any{}, nil); err != nil {
			return err
		}
	}
}

func (g *game) save(path string, what string) ([]byte, error) {
	resp, err := http.Get(g.base + "/api/games/" + g.id + "/" + what)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%s: %s", what, b)
	}
	return b, os.WriteFile(path, b, 0o644)
}

// seat makes the seat of a SEAT argument.
func seat(name, who string, profiles map[string]bool) (server.Seat, error) {
	switch who {
	case "greedy", "random":
		return server.Seat{Name: name, Kind: who}, nil
	}
	if !profiles[who] {
		return server.Seat{}, fmt.Errorf("no profile %q: give it with --model or --models, or use greedy or random", who)
	}
	return server.Seat{Name: name, Kind: "model", Model: who}, nil
}

func play(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	specs, file := profilesFlags(fs)
	pro := fs.String("pro", "", "who plays pro: a profile's name, greedy or random")
	con := fs.String("con", "", "who plays con")
	games := fs.Int("games", 1, "how many games")
	swap := fs.Bool("swap", false, "swap the sides every other game (the seat names stay Pro and Con)")
	mode := fs.String("mode", "debate", "the game mode")
	claim := fs.String("claim", "", "the claim, as a statement (e.g. should_be_performed(tax_meat)); without it, Pro's first argument states it")
	title := fs.String("title", "", "the game's title (the topic, for the players)")
	context_ := fs.String("context", "", "a context preset (its scheme weights)")
	standard := fs.String("standard", "", "the claim's proof standard: PE, CCE or BRD")
	rounds := fs.Int("max-rounds", 10, "the game ends after this many rounds")
	maxMoves := fs.Int("max-moves", 80, "stop a game after this many moves (a guard on cost)")
	seed := fs.Int64("seed", 1, "the seed of the rule bots (game n uses seed+n)")
	out := fs.String("out", "", "the directory for the trace files, exports and summary.jsonl")
	dataDir := fs.String("data", "", "the game data directory (default: built in)")
	_ = fs.Parse(args)
	if *pro == "" || *con == "" || *out == "" {
		return fmt.Errorf("play needs --pro, --con and --out")
	}
	ps, err := llm.ProfilesFrom(*file, *specs)
	if err != nil {
		return err
	}
	d, err := server.ReadData(*dataDir)
	if err != nil {
		return err
	}
	srv, err := server.New(d, "")
	if err != nil {
		return err
	}
	srv.SetContext(ctx)
	have := map[string]bool{}
	for _, p := range ps {
		c, err := llm.NewClient(p)
		if err != nil {
			return fmt.Errorf("profile %s: %v", p.Name, err)
		}
		srv.AddModel(c)
		have[p.Name] = true
		if n := p.Notice(); n != "" && (p.Name == *pro || p.Name == *con) {
			fmt.Fprintln(os.Stderr, "Note:", n)
		}
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	sum, err := os.OpenFile(filepath.Join(*out, "summary.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer sum.Close()
	run := time.Now().UTC().Format("20060102-150405")
	for n := 0; n < *games && ctx.Err() == nil; n++ {
		a, b := *pro, *con
		if *swap && n%2 == 1 {
			a, b = b, a
		}
		sp, err := seat("Pro", a, have)
		if err != nil {
			return err
		}
		sc, err := seat("Con", b, have)
		if err != nil {
			return err
		}
		r := *rounds
		req := server.NewGameRequest{Title: *title, Mode: *mode, Seats: []server.Seat{sp, sc}, Claim: *claim, Standard: *standard,
			Context: *context_, MaxRounds: &r, AllowOther: true, Seed: *seed + int64(n)}
		var st stateView
		if err := call(ts.URL, "POST", "/api/games", req, &st); err != nil {
			return err
		}
		g := &game{base: ts.URL, id: st.ID}
		start := time.Now()
		runErr := g.run(*maxMoves)
		name := fmt.Sprintf("%s-%03d", run, n+1)
		trace, err := g.save(filepath.Join(*out, name+".trace.jsonl"), "traces")
		if err != nil {
			return err
		}
		if _, err := g.save(filepath.Join(*out, name+".export.json"), "export"); err != nil {
			return err
		}
		s := summarize(trace)
		s["file"], s["pro"], s["con"], s["seconds"] = name+".trace.jsonl", a, b, time.Since(start).Seconds()
		if runErr != nil {
			s["stopped"] = runErr.Error()
		}
		line, _ := json.Marshal(s)
		fmt.Fprintln(sum, string(line))
		fmt.Printf("game %d: %s (pro) vs %s (con): %v, %v moves, %v decisions (%v rejected), %v calls (%v failed)%s\n", n+1, a, b,
			s["winner_side"], s["moves"], s["decisions"], s["rejected"], s["calls"], s["failed_calls"], stoppedNote(runErr))
	}
	return ctx.Err()
}

func stoppedNote(err error) string {
	if err == nil {
		return ""
	}
	return "; stopped: " + err.Error()
}

// readEvents splits a trace file into its events.
func readEvents(b []byte) ([]map[string]json.RawMessage, error) {
	var out []map[string]json.RawMessage
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 1<<20), 256<<20)
	for sc.Scan() {
		if len(bytes.TrimSpace(sc.Bytes())) == 0 {
			continue
		}
		var e map[string]json.RawMessage
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, sc.Err()
}

func str(e map[string]json.RawMessage, k string) string {
	var s string
	_ = json.Unmarshal(e[k], &s)
	return s
}

// summarize measures a game from its trace file: what happened, how often
// the models' moves were legal, how well they predicted the labels, what
// the calls cost.
func summarize(trace []byte) map[string]any {
	evs, _ := readEvents(trace)
	s := map[string]any{"decisions": 0, "rejected": 0, "unparsed": 0, "calls": 0, "failed_calls": 0, "moves": 0}
	inc := func(k string, n int) { s[k] = s[k].(int) + n }
	tokens := map[string]int{}
	var latency time.Duration
	expOK, expN, altLegal, altN := 0, 0, 0, 0
	byPlayer := map[string]map[string]int{}
	failedKinds := map[string]int{}
	for _, e := range evs {
		switch str(e, "type") {
		case "game_start":
			s["game"], s["data_sha256"], s["protocol"] = str(e, "game"), str(e, "data_sha256"), str(e, "protocol")
		case "call":
			inc("calls", 1)
			var a llm.Attempt
			_ = json.Unmarshal(mustJSON(e), &a)
			latency += a.Latency
			if a.Error != nil {
				inc("failed_calls", 1)
				failedKinds[a.Error.Kind]++
			}
			if r := a.Response; r != nil {
				for k, v := range map[string]*int{"input": r.Usage.InputTokens, "output": r.Usage.OutputTokens, "reasoning": r.Usage.ReasoningTokens} {
					if v != nil {
						tokens[k] += *v
					}
				}
			}
		case "decision":
			var t server.Trace
			_ = json.Unmarshal(mustJSON(e), &t)
			inc("decisions", 1)
			p := byPlayer[t.Player]
			if p == nil {
				p = map[string]int{}
				byPlayer[t.Player] = p
			}
			p["decisions"]++
			if !t.Result.Accepted && !t.Result.Abstained && !t.Result.Pending {
				inc("rejected", 1)
				p["rejected"]++
				if t.Reply != "" && t.Action.Type == "" {
					inc("unparsed", 1)
				}
			}
			for _, c := range t.ExpectCheck {
				expN++
				if c.OK {
					expOK++
				}
			}
			for _, a := range t.AlternativeOutcomes {
				altN++
				if a.Legal {
					altLegal++
				}
			}
		case "game_end":
			var o struct {
				Moves   int `json:"moves"`
				Outcome struct {
					Reason      string             `json:"reason"`
					Winner      string             `json:"winner"`
					WinningSide string             `json:"winning_side"`
					Verdict     string             `json:"verdict"`
					Scores      map[string]float64 `json:"scores"`
				} `json:"outcome"`
			}
			_ = json.Unmarshal(mustJSON(e), &o)
			s["moves"], s["reason"], s["winner"], s["winner_side"], s["verdict"], s["scores"] = o.Moves, o.Outcome.Reason, o.Outcome.Winner, o.Outcome.WinningSide, o.Outcome.Verdict, o.Outcome.Scores
		}
	}
	s["tokens"], s["call_seconds"], s["by_player"] = tokens, latency.Seconds(), byPlayer
	if len(failedKinds) > 0 {
		s["failed_kinds"] = failedKinds
	}
	if expN > 0 {
		s["expect_accuracy"], s["expect_checked"] = float64(expOK)/float64(expN), expN
	}
	if altN > 0 {
		s["alternatives_legal"] = float64(altLegal) / float64(altN)
	}
	return s
}

func mustJSON(e map[string]json.RawMessage) []byte { b, _ := json.Marshal(e); return b }

func replay(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	dataDir := fs.String("data", "", "the game data directory (default: built in)")
	if len(args) == 0 {
		return fmt.Errorf("replay TRACE.jsonl")
	}
	_ = fs.Parse(args[1:])
	b, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	evs, err := readEvents(b)
	if err != nil {
		return err
	}
	if len(evs) == 0 || str(evs[0], "type") != "game_start" {
		return fmt.Errorf("%s does not start with a game_start event (format %s)", args[0], server.TraceFormat)
	}
	var start server.GameStart
	if err := json.Unmarshal(mustJSON(evs[0]), &start); err != nil {
		return err
	}
	if start.Request == nil {
		return fmt.Errorf("the game was not started from a request (an imported or loaded game): it cannot be replayed")
	}
	d, err := server.ReadData(*dataDir)
	if err != nil {
		return err
	}
	if h := d.Hash(); h != start.DataSHA256 {
		fmt.Fprintf(os.Stderr, "Note: the game was played with scheme data %s; this data is %s (use --data for that version).\n", start.DataSHA256, h)
	}
	srv, err := server.New(d, "")
	if err != nil {
		return err
	}
	srv.SetContext(ctx)
	// Each model seat gets the calls recorded for its player, in order.
	profiles := map[string]llm.Profile{}
	for _, p := range start.Profiles {
		profiles[p.Name] = p
	}
	calls := map[string][]llm.Attempt{}
	var decisions []server.Trace
	for _, e := range evs[1:] {
		switch str(e, "type") {
		case "call":
			var c server.Call
			if err := json.Unmarshal(mustJSON(e), &c); err != nil {
				return err
			}
			calls[c.Player] = append(calls[c.Player], c.Attempt)
		case "decision":
			var t server.Trace
			if err := json.Unmarshal(mustJSON(e), &t); err != nil {
				return err
			}
			decisions = append(decisions, t)
		}
	}
	req := *start.Request
	req.Seed = start.Seed
	req.Seats = append([]server.Seat(nil), start.Seats...)
	replays := map[string]*llm.Replay{}
	for i, s := range req.Seats {
		if s.Kind != "model" && s.Kind != "claude" {
			continue
		}
		name := s.Model
		if s.Kind == "claude" {
			name = "claude"
		}
		p := profiles[name]
		p.Name, p.MinInterval = "replay:"+s.Name, 0
		rp := &llm.Replay{Provider: p.Provider, Recorded: calls[s.Name]}
		replays[s.Name] = rp
		c := llm.WithProvider(p, rp)
		c.Sleep = llm.NoWait
		srv.AddModel(c)
		req.Seats[i] = server.Seat{Name: s.Name, Kind: "model", Model: p.Name}
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	var st stateView
	if err := call(ts.URL, "POST", "/api/games", req, &st); err != nil {
		return err
	}
	g := &game{base: ts.URL, id: st.ID}
	runErr := g.run(10 * (len(decisions) + 1))
	resp, err := http.Get(ts.URL + "/api/games/" + st.ID + "/traces")
	if err != nil {
		return err
	}
	nb, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	nevs, _ := readEvents(nb)
	var again []server.Trace
	for _, e := range nevs {
		if str(e, "type") == "decision" {
			var t server.Trace
			_ = json.Unmarshal(mustJSON(e), &t)
			again = append(again, t)
		}
	}
	var problems []string
	for name, rp := range replays {
		for _, m := range rp.Mismatches {
			problems = append(problems, name+": "+m)
		}
		if n := rp.Left(); n > 0 {
			problems = append(problems, fmt.Sprintf("%s: %d recorded calls were not made", name, n))
		}
	}
	if len(again) != len(decisions) {
		problems = append(problems, fmt.Sprintf("%d decisions, not %d", len(again), len(decisions)))
	}
	for i := 0; i < len(again) && i < len(decisions); i++ {
		if d := decisionDiff(decisions[i], again[i]); d != "" {
			problems = append(problems, fmt.Sprintf("decision %d (%s): %s", i, decisions[i].Player, d))
		}
	}
	if runErr != nil {
		problems = append(problems, "the replay stopped: "+runErr.Error())
	}
	sort.Strings(problems)
	fmt.Printf("%d decisions, %d calls replayed.\n", len(again), len(evs)-len(decisions)-1-countType(evs, "game_end"))
	if len(problems) > 0 {
		fmt.Println(strings.Join(problems, "\n"))
		return fmt.Errorf("the replay differs from the trace in %d ways", len(problems))
	}
	fmt.Println("The replay matches the trace: the same requests, the same decisions, the same results.")
	return nil
}

func countType(evs []map[string]json.RawMessage, t string) int {
	n := 0
	for _, e := range evs {
		if str(e, "type") == t {
			n++
		}
	}
	return n
}

// decisionDiff compares what a decision did, not when.
func decisionDiff(a, b server.Trace) string {
	key := func(t server.Trace) string {
		x, _ := json.Marshal(map[string]any{"player": t.Player, "turn": t.Turn, "attempt": t.Attempt, "prompt": t.PromptSHA256, "reply": t.Reply,
			"action": t.Action, "result": t.Result, "claim_before": t.ClaimBefore, "claim_after": t.ClaimAfter, "effects": t.Effects,
			"expect": t.ExpectCheck, "alternatives": t.AlternativeOutcomes, "move": t.Move})
		return string(x)
	}
	ka, kb := key(a), key(b)
	if ka == kb {
		return ""
	}
	switch {
	case a.PromptSHA256 != b.PromptSHA256:
		return "the prompt differs"
	case a.Reply != b.Reply:
		return "the reply differs"
	}
	return "the result differs: " + ka + " / " + kb
}

// report sums up summary.jsonl by profile and side.
func report(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("report DIR")
	}
	b, err := os.ReadFile(filepath.Join(args[0], "summary.jsonl"))
	if err != nil {
		return err
	}
	type agg struct{ games, wins, decisions, rejected int }
	rows := map[string]*agg{}
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for sc.Scan() {
		var s struct {
			Pro, Con   string
			WinnerSide string                    `json:"winner_side"`
			ByPlayer   map[string]map[string]int `json:"by_player"`
		}
		if json.Unmarshal(sc.Bytes(), &s) != nil {
			continue
		}
		for side, who := range map[string]string{"pro": s.Pro, "con": s.Con} {
			k := who + " as " + side
			r := rows[k]
			if r == nil {
				r = &agg{}
				rows[k] = r
			}
			r.games++
			if s.WinnerSide == side {
				r.wins++
			}
			p := s.ByPlayer[map[string]string{"pro": "Pro", "con": "Con"}[side]]
			r.decisions += p["decisions"]
			r.rejected += p["rejected"]
		}
	}
	keys := make([]string, 0, len(rows))
	for k := range rows {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("%-40s %6s %6s %10s %10s\n", "seat", "games", "wins", "decisions", "rejected")
	for _, k := range keys {
		r := rows[k]
		fmt.Printf("%-40s %6d %6d %10d %9.0f%%\n", k, r.games, r.wins, r.decisions, 100*float64(r.rejected)/float64(max(r.decisions, 1)))
	}
	return nil
}
