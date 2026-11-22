package server

import (
	"strings"
	"testing"
)

// Sources are text. A null source is no source; any other value that is not
// text is refused by name, and a malformed last move is never replaced by an
// earlier object of the reply.
func TestReplySourcesMustBeText(t *testing.T) {
	move := func(sources string) string {
		return `{"type":"support","scheme":"expert_opinion","sources":` + sources + `}`
	}
	for _, ok := range []struct {
		sources string
		want    []string
	}{
		{`["Smith 2020", "https://example.org/a"]`, []string{"Smith 2020", "https://example.org/a"}},
		{`"Smith 2020"`, []string{"Smith 2020"}},
		{`[null]`, nil},
		{`[null, " Smith 2020 "]`, []string{"Smith 2020"}},
		{`null`, nil},
		{`[]`, nil},
	} {
		a, err := ParseReply(move(ok.sources))
		if err != nil {
			t.Errorf("%s: %v", ok.sources, err)
			continue
		}
		if strings.Join(a.Evidence, "|") != strings.Join(ok.want, "|") {
			t.Errorf("%s: evidence %q, want %q", ok.sources, a.Evidence, ok.want)
		}
	}
	for _, bad := range []string{`[false]`, `[123]`, `[{"url":"https://example.org"}]`, `5`, `{"url":"x"}`} {
		a, err := ParseReply(move(bad))
		if err == nil || !strings.Contains(err.Error(), `"sources"`) {
			t.Errorf("%s: got %v, %v; want an error naming sources", bad, a.Evidence, err)
		}
	}
	// An example in the reasoning, then the real (malformed) move.
	reply := `For instance {"action": {"type": "abstain"}} would pass. My move:
{"action": {"type": "support", "scheme": "expert_opinion", "evidence": [true]}}`
	if a, err := ParseReply(reply); err == nil {
		t.Fatalf("played the earlier example %q instead of reporting the malformed move", a.Type)
	}
	// The evidence key name is reported as written.
	if _, err := ParseReply(`{"type":"support","evidence":[1]}`); err == nil || !strings.Contains(err.Error(), `"evidence"`) {
		t.Errorf("error does not name evidence: %v", err)
	}
}
