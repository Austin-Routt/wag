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

// Cases TS8, TS9 and TS10 (docs/d086-first-gold.md §6; the independent check
// of F3, F5 and F6): the Go measure gives target_instance, target_instance_v2,
// target_formula and instances the values agreement.py gives, on secondary
// targets of every shape the check named (one premise and five, a premise
// with no role, a formula that is a number, a premise that is not an object,
// premises that are not a list), on target claims whose instances are
// malformed (TS10, added with the code: instances that are not a list, an
// entry that is not an object, a later instance whose primary is "true" or 1,
// an empty instance), and on every item of the format-2 gold, read against
// itself.
func TestTS9ParityWithPythonOnEveryShape(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	pm := func(role, formula any) map[string]any {
		p := map[string]any{"formula": formula}
		if role != nil {
			p["role"] = role
		}
		return p
	}
	secondary := func(premises any) map[string]any {
		return map[string]any{"id": "t.1#2", "primary": false, "scheme": map[string]any{"walton_id": "practical_reasoning2"},
			"premises": premises, "conclusion": map[string]any{"formula": "should_be_performed(tax)"}}
	}
	first := map[string]any{"id": "t.1#1", "primary": true, "scheme": map[string]any{"walton_id": "practical_reasoning2"},
		"premises": []any{pm("premise", "goal_of_agent(g,env)")}, "conclusion": map[string]any{"formula": "should_be_performed(tax)"}}
	claim := func(insts any) Annotation { return Annotation{"kialo_claim_id": "t.1", "instances": insts} }
	at := func(id string, ref any, formula any) Annotation {
		return Annotation{"kialo_claim_id": id, "target": map[string]any{"actual": map[string]any{
			"claim": "t.1", "instance": ref, "part": "conclusion", "formula": formula}}}
	}
	type kase struct {
		Name  string                `json:"name"`
		S     map[string]Annotation `json:"S"`
		Items []Annotation          `json:"items"`
	}
	shape := func(name string, t1 Annotation) kase {
		return kase{name, map[string]Annotation{"t.1": t1},
			[]Annotation{t1, at("c.5", "t.1#2", "should_be_performed(tax)"), at("c.6", nil, nil), at("c.7", "t.1#1", nil)}}
	}
	five := []any{pm("premise", "a(x)"), pm("exception", "b(x)"), pm("premise", "c( x )"), pm("assumption", "d(x)"), pm("premise", "e(x)")}
	cases := []kase{
		shape("TS9 one premise", claim([]any{first, secondary([]any{pm("premise", "p(a)")})})),
		shape("TS9 five premises", claim([]any{first, secondary(five)})),
		shape("TS9 a premise with no role", claim([]any{first, secondary([]any{pm(nil, "p(a)"), pm("premise", "q(b)")})})),
		shape("TS9 a formula that is a number", claim([]any{first, secondary([]any{pm("premise", 5), pm("premise", true)})})),
		shape("TS9 a premise that is not an object", claim([]any{first, secondary([]any{"p(a)", 7, nil, pm("premise", "q(b)")})})),
		shape("TS8 premises a text", claim([]any{first, secondary("ab")})),
		shape("TS8 premises an object", claim([]any{first, secondary(map[string]any{"x": 1, "y": 2})})),
		shape("TS8 premises a number", claim([]any{first, secondary(5)})),
		shape("TS8 premises true", claim([]any{first, secondary(true)})),
		shape("TS10 no instances", Annotation{"kialo_claim_id": "t.1"}),
		shape("TS10 instances a number", claim(5)),
		shape("TS10 instances a text", claim("ab")),
		shape("TS10 an entry that is not an object", claim([]any{7, first, secondary([]any{pm("premise", "p(a)")})})),
		shape("TS10 a later primary written true", claim([]any{
			map[string]any{"id": "t.1#1", "primary": false, "conclusion": map[string]any{"formula": "first(x)"}},
			map[string]any{"id": "t.1#2", "primary": "true", "conclusion": map[string]any{"formula": "second(x)"}}})),
		shape("TS10 a later primary written 1", claim([]any{
			map[string]any{"id": "t.1#1", "conclusion": map[string]any{"formula": "first(x)"}},
			map[string]any{"id": "t.1#2", "primary": 1, "conclusion": map[string]any{"formula": "second(x)"}}})),
		shape("TS10 an empty instance", claim([]any{map[string]any{}})),
	}
	// The format-2 gold, every item read against the whole file.
	b, err := os.ReadFile("../../../docs/kialo-batch2/gold.json")
	if err != nil {
		t.Fatal(err)
	}
	var gold struct {
		Items []Annotation `json:"items"`
	}
	if err := json.Unmarshal(b, &gold); err != nil || len(gold.Items) != 50 {
		t.Fatalf("gold: %v, %d items", err, len(gold.Items))
	}
	set := map[string]Annotation{}
	for _, it := range gold.Items {
		set[it["kialo_claim_id"].(string)] = it
	}
	cases = append(cases, kase{"the format-2 gold", set, gold.Items})
	fields := []string{"target_instance", "target_instance_v2", "target_formula", "instances"}
	b, err = json.Marshal(map[string]any{"fields": fields, "cases": cases})
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(t.TempDir(), "cases.json")
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
doc = json.load(open(sys.argv[2], encoding="utf-8"))
print(json.dumps([[[ag.field_value(f, it, c["S"]) for f in doc["fields"]] for it in c["items"]] for c in doc["cases"]]))`
	out, err := exec.Command(py, "-I", "-B", "-c", read, script, fixture).Output()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	var want [][][]string
	if err := json.Unmarshal(out, &want); err != nil || len(want) != len(cases) {
		t.Fatalf("%v: %s", err, out)
	}
	// Read back through JSON, as the measure reads annotation files.
	var doc struct {
		Cases []kase `json:"cases"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	compared := 0
	for i, c := range doc.Cases {
		for j, it := range c.Items {
			for k, f := range fields {
				compared++
				if g := FieldIn(it, f, c.S); g != want[i][j][k] {
					t.Errorf("%s, item %d, %s: go %q, agreement.py %q", c.Name, j, f, g, want[i][j][k])
				}
			}
		}
	}
	if compared != (len(cases)-1)*4*len(fields)+50*len(fields) {
		t.Errorf("%d values compared", compared)
	}
	// TS8: premises that are not a list are none.
	for _, c := range doc.Cases[5:9] {
		if g, w := FieldIn(c.Items[1], "target_instance_v2", c.S), "t.1 secondary:practical_reasoning2:should_be_performed(tax)[]by="; g != w {
			t.Errorf("%s: %q, want %q", c.Name, g, w)
		}
	}
	// TS10: the primary is the first object whose primary is the boolean true, else the first object.
	for _, c := range doc.Cases[13:15] {
		if g := FieldIn(c.Items[2], "target_formula", c.S); g != "first(x)" {
			t.Errorf("%s: the format-1 target reads %q, want the first instance's conclusion", c.Name, g)
		}
	}
}
