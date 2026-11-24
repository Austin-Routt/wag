package annotate

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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

// agreement.json is what agreement.py makes, with the README's command: the
// figures the Go measure is compared with above are the script's own.
func TestAgreementJSONIsWhatAgreementPyMakes(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "agreement.json")
	cmd := exec.Command(py, "-I", "-B", "agreement.py", "annotator-A.json", "annotator-B.json",
		"--pilot", "../kialo-pilot/pilot.json", "--families", "../../wag/data/schemes/families.json", "--json", out)
	cmd.Dir = dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, b)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(dir, "agreement.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n"))) {
		t.Fatal("agreement.json is not what agreement.py makes: run the README's command")
	}
}
