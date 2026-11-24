package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The replay compares the decisions it made with the trace's: it reads its
// own trace in full, or says why it cannot, rather than comparing part of
// it (a cut-off body, an error answer, a line that is not JSON).
func TestTheReplaysOwnTraceIsReadInFull(t *testing.T) {
	const decision = `{"type":"decision","player":"pro","turn":1}`
	for _, c := range []struct {
		name, body string
		status     int
		length     string // a Content-Length longer than the body cuts it off
		want       string
	}{
		{"cut off", decision + "\n" + `{"type":"game_end"`, 200, "1000", "reading"},
		{"an error answer", `{"error":"no game x"}`, 404, "", "no game x"},
		{"a line that is not JSON", decision + "\n" + `{"type":"decision",` + "\n", 200, "", "unexpected end of JSON input"},
		{"a decision that is not one", `{"type":"decision","turn":"one"}`, 200, "", "a decision"},
	} {
		t.Run(c.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/games/g1/traces" {
					http.NotFound(w, r)
					return
				}
				if c.length != "" {
					w.Header().Set("Content-Length", c.length)
				}
				w.WriteHeader(c.status)
				fmt.Fprint(w, c.body)
			}))
			defer ts.Close()
			got, err := replayedDecisions(ts.URL, "g1")
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("%d decisions; %v", len(got), err)
			}
		})
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"type":"game_start"}`+"\n"+decision+"\n"+`{"type":"call"}`+"\n"+decision+"\n")
	}))
	defer ts.Close()
	got, err := replayedDecisions(ts.URL, "g1")
	if err != nil || len(got) != 2 || got[0].Player != "pro" {
		t.Fatalf("%+v; %v", got, err)
	}
}
