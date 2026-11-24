package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditV2SaveRejectsTruncatedHTTPBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		fmt.Fprint(w, `{"partial":`)
	}))
	defer srv.Close()
	if err := call(srv.URL, "GET", "/fixture", nil, nil); err == nil {
		t.Fatal("truncated API response accepted")
	}
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	g := game{base: srv.URL, id: "fixture"}
	if _, err := g.save(path, "traces"); err == nil {
		t.Fatal("truncated trace accepted as successfully saved")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("partial artifact published: %v", err)
	}
}

func TestAuditV2SummaryFailureStopsPlay(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("requires Linux /dev/full fault sink")
	}
	dir := t.TempDir()
	if err := os.Symlink("/dev/full", filepath.Join(dir, "summary.jsonl")); err != nil {
		t.Fatal(err)
	}
	err := play(context.Background(), []string{"--model", "a=mock:m", "--model", "b=mock:m", "--pro", "a", "--con", "b", "--out", dir, "--games", "2", "--max-rounds", "2"})
	traces, _ := filepath.Glob(filepath.Join(dir, "*.trace.jsonl"))
	if err == nil || len(traces) != 1 {
		t.Fatalf("summary write failure not surfaced before next game: error=%v completed_trace_files=%d", err, len(traces))
	}
}
