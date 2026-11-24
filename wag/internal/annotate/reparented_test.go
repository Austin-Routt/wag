package annotate

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// re_parented is derived, as the guidelines say the tooling does (§5, item 1):
// true exactly when the actual target claim is not the Kialo parent. The
// annotation's own value is not read, and the Go measure and agreement.py
// derive it alike (the review of 24 November 2026: a missing value was ""
// in Go and false in agreement.py, so the two disagreed).
var reParentedCases = []struct {
	name, json string
	want       bool
}{
	{"another claim, stated false", `{"target": {"kialo_parent": "1.1", "actual": {"claim": "1.2", "part": "conclusion"}, "re_parented": false}}`, true},
	{"the parent, stated true", `{"target": {"kialo_parent": "1.1", "actual": {"claim": "1.1", "part": "premise"}, "re_parented": true}}`, false},
	{"another claim, not stated", `{"target": {"kialo_parent": "1.1", "actual": {"claim": "1.2"}}}`, true},
	{"the parent, not stated", `{"target": {"kialo_parent": "1.1", "actual": {"claim": "1.1"}}}`, false},
	{"the thesis", `{"target": {"kialo_parent": null, "actual": null, "re_parented": false}}`, false},
	{"an empty actual target", `{"target": {"kialo_parent": "1.1", "actual": {}, "re_parented": true}}`, false},
	{"no target", `{"re_parented": true}`, false},
	{"a target that is not an object", `{"target": "1.2"}`, false},
	{"an actual target that is not an object", `{"target": {"kialo_parent": "1.1", "actual": "1.2"}}`, false},
	{"no parent named", `{"target": {"actual": {"claim": "1.2"}}}`, true},
	{"no claim named", `{"target": {"kialo_parent": "1.1", "actual": {"part": "premise"}}}`, true},
	{"an empty claim, no parent", `{"target": {"kialo_parent": null, "actual": {"claim": ""}}}`, false},
	// Only text is read as a claim; anything else is none (finding 4 of the
	// review of the repairs: Go and Python wrote numbers and booleans apart).
	{"a claim that is not text", `{"target": {"kialo_parent": "True", "actual": {"claim": true}}}`, true},
	{"a large number", `{"target": {"kialo_parent": "1000000", "actual": {"claim": 1000000}}}`, true},
	{"a parent that is not text", `{"target": {"kialo_parent": 1.0, "actual": {"claim": "1"}}}`, true},
	{"a list", `{"target": {"kialo_parent": "[1 2]", "actual": {"claim": [1, 2]}}}`, true},
	{"two numbers", `{"target": {"kialo_parent": 13292.1, "actual": {"claim": 13292.1}}}`, false},
}

func TestReParentedIsDerived(t *testing.T) {
	for _, c := range reParentedCases {
		var a Annotation
		if err := json.Unmarshal([]byte(c.json), &a); err != nil {
			t.Fatal(err)
		}
		if got := FieldOf(a, "re_parented"); got != strconv.FormatBool(c.want) {
			t.Errorf("%s: %q, want %v", c.name, got, c.want)
		}
	}
}

func TestReParentedIsDerivedAsAgreementPyDerivesIt(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	var docs []string
	for _, c := range reParentedCases {
		docs = append(docs, c.json)
	}
	in, _ := json.Marshal(docs)
	cmd := exec.Command(py, "-I", "-B", "-c", "import json, sys\nsys.path.insert(0, sys.argv[1])\nimport agreement\n"+
		"print(json.dumps([agreement.re_parented(json.loads(d)) for d in json.load(sys.stdin)]))", dir)
	cmd.Stdin = bytes.NewReader(in)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%v: %s", err, stderr.String())
	}
	var got []bool
	if err := json.Unmarshal(out, &got); err != nil || len(got) != len(reParentedCases) {
		t.Fatalf("%v: %s", err, out)
	}
	for i, c := range reParentedCases {
		if got[i] != c.want {
			t.Errorf("agreement.py, %s: %v, want %v", c.name, got[i], c.want)
		}
	}
}

// The claim's view carries the values compared, so that the dashboard shows
// the derived re-parenting, not the annotation's own field.
func TestTheItemViewCarriesTheValuesCompared(t *testing.T) {
	r, _, items := soloRun(t, 1)
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	v, err := r.Item(items[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range Fields {
		if got, want := v.Compared["solo"][f], FieldOf(v.Annotations["solo"], f); got != want {
			t.Errorf("solo's %s: %q, want %q", f, got, want)
		}
		if got, want := v.GoldCompared[f], FieldOf(v.Gold, f); got != want {
			t.Errorf("the gold's %s: %q, want %q", f, got, want)
		}
	}
	if v.Compared["solo"]["re_parented"] == "" {
		t.Fatalf("%v", v.Compared)
	}
}
