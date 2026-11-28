package annotate

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Case TS7 (docs/d086-first-gold.md §6; F6 of the external review of
// exchange 3): the Go measure gives target_instance and target_instance_v2
// the values agreement.py gives, on the fixtures of TS1-TS6 (the batch's two
// annotators are TestAgreementMatchesPython's).
func TestStructuralTargetMatchesPython(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	goal := func(id, g string, primary bool, by string, reversed, spaced, unresolved bool) map[string]any {
		first := "goal_of_agent(g," + g + ")"
		if spaced {
			first = "goal_of_agent( g , " + g + " )"
		}
		var second any = "would_bring_about(tax," + g + ")"
		if unresolved {
			second = map[string]any{"unresolved": "which effect"}
		}
		ps := []any{map[string]any{"role": "premise", "formula": first}, map[string]any{"role": "premise", "formula": second},
			map[string]any{"role": "exception", "formula": "side_effect(tax)"}}
		if reversed {
			ps = []any{ps[2], ps[1], ps[0]}
		}
		in := map[string]any{"id": id, "primary": primary, "scheme": map[string]any{"walton_id": "practical_reasoning2"},
			"premises": ps, "conclusion": map[string]any{"formula": "should_be_performed(tax)"}}
		if by != "" {
			in["by"] = by
		}
		return in
	}
	claim := func(ins ...any) map[string]Annotation {
		return map[string]Annotation{"t.1": {"kialo_claim_id": "t.1", "instances": ins}}
	}
	item := func(instance any, formula any) Annotation {
		return Annotation{"kialo_claim_id": "c.5", "target": map[string]any{"actual": map[string]any{
			"claim": "t.1", "instance": instance, "part": "conclusion", "formula": formula}}}
	}
	both := claim(goal("t.1#1", "env", true, "", false, false, false), goal("t.1#2", "env", false, "", false, false, false),
		goal("t.1#3", "health", false, "", false, false, false), goal("t.1#4", "env", false, "the discussion", false, false, false))
	type kase struct {
		S    map[string]Annotation `json:"S"`
		Item Annotation            `json:"item"`
	}
	cases := []kase{
		{both, item("t.1#2", "should_be_performed(tax)")}, // TS1, TS5
		{both, item("t.1#3", "should_be_performed(tax)")}, // TS1
		{claim(goal("t.1#1", "env", true, "", false, false, false), goal("t.1#2", "env", false, "", true, true, false)),
			item("t.1#2", "should_be_performed(tax)")}, // TS2
		{both, item("t.1#1", "should_be_performed(tax)")}, // TS3
		{claim(goal("t.1#1", "env", true, "", false, false, false), goal("t.1#2", "env", false, "", false, false, true)),
			item("t.1#2", "should_be_performed(tax)")}, // TS4
		{both, item("t.1#4", "should_be_performed(tax)")},                                        // TS5
		{both, item("t.1#2", map[string]any{"unresolved": "which formula"})},                     // TS6
		{both, item(map[string]any{"unresolved": "which instance"}, "should_be_performed(tax)")}, // an unresolved instance alone
	}
	dir := t.TempDir()
	b, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(dir, "cases.json")
	if err := os.WriteFile(fixture, b, 0o600); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("../../../docs/kialo-batch2/agreement.py")
	if err != nil {
		t.Fatal(err)
	}
	const read = `import importlib.util, json, sys
spec = importlib.util.spec_from_file_location("agreement", sys.argv[1])
ag = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ag)
print(json.dumps([[ag.field_value(f, c["item"], c["S"]) for f in ("target_instance", "target_instance_v2")]
                  for c in json.load(open(sys.argv[2], encoding="utf-8"))]))`
	cmd := exec.Command(py, "-I", "-B", "-c", read, script, fixture)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	var want [][]string
	if err := json.Unmarshal(out, &want); err != nil || len(want) != len(cases) {
		t.Fatalf("%v: %s", err, out)
	}
	// Read back through JSON, as the measure reads annotation files.
	var got []kase
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for i, c := range got {
		g := []string{FieldIn(c.Item, "target_instance", c.S), FieldIn(c.Item, "target_instance_v2", c.S)}
		if g[0] != want[i][0] || g[1] != want[i][1] {
			t.Errorf("case %d: go %q, agreement.py %q", i, g, want[i])
		}
	}
	// The cases' own outcomes, as TS1-TS6 have them.
	v := func(i int) string { return FieldIn(got[i].Item, "target_instance_v2", got[i].S) }
	v1 := func(i int) string { return FieldIn(got[i].Item, "target_instance", got[i].S) }
	switch {
	case v1(0) != v1(1) || v(0) == v(1):
		t.Errorf("TS1: target_instance %q %q, v2 %q %q", v1(0), v1(1), v(0), v(1))
	case v(0) != v(2):
		t.Errorf("TS2: %q, %q", v(0), v(2))
	case v(3) != "t.1 primary" || v1(3) != "t.1 primary":
		t.Errorf("TS3: %q %q", v1(3), v(3))
	case v(4) != Unresolved || v1(4) == Unresolved:
		t.Errorf("TS4: %q %q", v1(4), v(4))
	case v1(0) != v1(5) || v(0) == v(5):
		t.Errorf("TS5: %q %q", v(0), v(5))
	case v(6) != Unresolved || v1(6) != Unresolved || v(7) != Unresolved || v1(7) != Unresolved:
		t.Errorf("TS6: %q %q %q %q", v1(6), v(6), v1(7), v(7))
	}
}
