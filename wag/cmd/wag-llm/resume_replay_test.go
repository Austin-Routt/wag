package main

import (
	"context"
	"io/fs"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	wgame "github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/server"
)

// G6 (docs/g-journal.md): a game resumed from its folder replays without
// asking a model. A model seat plays a random player; the copy of the
// game's folder taken while the provider answers the seat's second attempt
// (its first reply in that turn refused) is the folder a process stopped
// then leaves. Resumed in a new server and played to its end, the game's
// trace replays: the same requests, the same decisions, the same results,
// the random player's after the resume included.
func TestG6AResumedGameReplays(t *testing.T) {
	const (
		claim  = `{"summary":"State the claim.","action":{"type":"support","scheme_id":"AS:consequences","variable_bindings":{"A":"tax meat","G":"lower emissions"},"declare":[{"id":"tax meat","actor":"decider"}]}}`
		noMove = `I have nothing to add.`
		pass   = `{"summary":"Nothing more.","action":{"type":"abstain"}}`
	)
	d, err := server.ReadData("")
	if err != nil {
		t.Fatal(err)
	}
	newServer := func(store string, answer func(n int32) string) (*httptest.Server, *atomic.Int32) {
		srv, err := server.New(d, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := srv.SetStore(store); err != nil {
			t.Fatal(err)
		}
		var n atomic.Int32
		c := llm.WithProvider(llm.Profile{Name: "m", Provider: "mock", Retries: -1, Params: llm.Params{Model: "m-1"}}, &llm.Mock{Fn: func(llm.Request) (*llm.Response, error) {
			return &llm.Response{Text: answer(n.Add(1)), FinishReason: "stop", Model: "m-1"}, nil
		}})
		c.Sleep = llm.NoWait
		srv.AddModel(c)
		ts := httptest.NewServer(srv.Handler())
		t.Cleanup(ts.Close)
		return ts, &n
	}

	// The game played until the stop: the folder is copied while the
	// provider answers the model's third call, its second turn's second
	// attempt.
	store, stopped := t.TempDir(), t.TempDir()
	copied := false
	ts, _ := newServer(store, func(n int32) string {
		switch n {
		case 1:
			return claim
		case 2:
			return noMove
		case 3:
			if err := os.CopyFS(stopped, os.DirFS(store)); err != nil {
				t.Error(err)
			}
			copied = true
		}
		return pass
	})
	rounds, concede := 4, false
	req := server.NewGameRequest{Mode: "debate", Seats: []server.Seat{{Name: "M", Kind: "model", Model: "m"}, {Name: "R", Kind: "random"}},
		MaxRounds: &rounds, ConcedeOnPass: &concede, AllowOther: true, Seed: 7, Profile: wgame.DefaultProfile()}
	var st stateView
	if err := call(ts.URL, "POST", "/api/games", req, &st); err != nil {
		t.Fatal(err)
	}
	g := &game{base: ts.URL, id: st.ID}
	for i := 0; i < 20 && !copied; i++ {
		if err := call(ts.URL, "POST", "/api/games/"+st.ID+"/ai", map[string]any{}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if !copied {
		t.Fatal("the model was not asked a third time")
	}

	// Resumed from the copy in a new server, and played to its end.
	ts2, asked := newServer(stopped, func(int32) string { return pass })
	dirs, _ := fs.Glob(os.DirFS(stopped), "games/*")
	if len(dirs) != 1 {
		t.Fatalf("the copy holds %v", dirs)
	}
	var st2 stateView
	if err := call(ts2.URL, "POST", "/api/store/resume", map[string]string{"path": dirs[0]}, &st2); err != nil {
		t.Fatal(err)
	}
	g = &game{base: ts2.URL, id: st2.ID}
	if err := g.run(200); err != nil {
		t.Fatal(err)
	}
	if asked.Load() == 0 {
		t.Fatal("the resumed game never asked the provider: nothing after the stop was played live")
	}
	if err := replay(context.Background(), []string{filepath.Join(stopped, filepath.FromSlash(dirs[0]), "trace.jsonl")}); err != nil {
		t.Fatalf("replay: %v", err)
	}
}
