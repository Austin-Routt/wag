package annotate

import (
	"encoding/json"
	"os"
	"testing"
)

// The Go agreement reproduces agreement.py's per-field figures on the two
// batch-2 annotators.
func TestAgreementMatchesPython(t *testing.T) {
	read := func(p string) map[string]Annotation {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Skip(err)
		}
		m, err := parseAnnotations(b)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	a, b := read("../../../docs/kialo-batch2/annotator-A.json"), read("../../../docs/kialo-batch2/annotator-B.json")
	var py struct {
		Fields map[string]struct {
			Agreement float64  `json:"agreement"`
			Kappa     *float64 `json:"kappa"`
		} `json:"fields"`
	}
	raw, _ := os.ReadFile("../../../docs/kialo-batch2/agreement.json")
	if err := json.Unmarshal(raw, &py); err != nil {
		t.Fatal(err)
	}
	p := Agree("A", "B", a, b)
	for _, f := range Fields {
		want, ok := py.Fields[f]
		if !ok {
			continue
		}
		got := p.Fields[f]
		t.Logf("%-13s go %.3f κ %v | py %.3f κ %v", f, got.Agreement, deref(got.Kappa), want.Agreement, deref(want.Kappa))
		if got.Agreement != want.Agreement || (got.Kappa == nil) != (want.Kappa == nil) || got.Kappa != nil && *got.Kappa != *want.Kappa {
			t.Errorf("%s differs from agreement.py", f)
		}
	}
}

func deref(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}
