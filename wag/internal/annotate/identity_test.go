package annotate

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
	"gopkg.in/yaml.v2"
)

// The identity contract's Go checks (docs/annotation-records.md, D092): the
// built-in dataset is format 2 with a case record, its gold passes them, and
// each planted defect is refused by its rule.

func builtin(t *testing.T) (*Dataset, *Collection) {
	t.Helper()
	ds, err := Builtin()
	if err != nil || len(ds) == 0 {
		t.Fatal(err)
	}
	d := ds[0]
	c, err := LoadCollection(d.WaltonYAML())
	if err != nil {
		t.Fatal(err)
	}
	return d, c
}

// clone is a deep copy of a set of annotations (by JSON, as they are read).
func clone(t *testing.T, set map[string]Annotation) map[string]Annotation {
	t.Helper()
	b, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]Annotation
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func inst(a Annotation, i int) map[string]any { return a["instances"].([]any)[i].(map[string]any) }

func actualOf(a Annotation) map[string]any {
	return a["target"].(map[string]any)["actual"].(map[string]any)
}

// renameIn replaces an atom in an instance's bindings and formulas.
func renameIn(in map[string]any, old, new string) {
	sub := func(f string) string {
		return atomArg.ReplaceAllStringFunc(f, func(x string) string {
			if x == old {
				return new
			}
			return x
		})
	}
	for k, v := range in["bindings"].(map[string]any) {
		if v == old {
			in["bindings"].(map[string]any)[k] = new
		}
	}
	for _, y := range in["premises"].([]any) {
		p := y.(map[string]any)
		p["formula"] = sub(p["formula"].(string))
	}
	cm := in["conclusion"].(map[string]any)
	cm["formula"] = sub(cm["formula"].(string))
}

func TestTheBuiltinDatasetIsFormat2WithACase(t *testing.T) {
	d, c := builtin(t)
	if d.Format != 2 || d.FormatName() != Format2 || d.Case == nil || d.Case.ID != "kialo-13292" || d.thesisID() != "13292.1" {
		t.Fatalf("format %d, case %+v, thesis %q", d.Format, d.Case, d.thesisID())
	}
	for _, id := range d.IDs() {
		if p := c.CheckIn(d.Gold[id], d, d.Gold, 0); len(p) > 0 {
			t.Errorf("%s: %v", id, p)
		}
	}
	if !strings.Contains(d.AnnotatorSystem(), "## case.json") {
		t.Error("the annotators are not given the case record")
	}
}

// Each planted defect is refused by its rule, and nothing else is said.
func TestFormat2Refusals(t *testing.T) {
	d, c := builtin(t)
	withCase := func(change func(*Case)) *Dataset {
		cs := *d.Case
		cs.canon, cs.actorOf, cs.actors, cs.contexts = map[string]string{}, map[string]string{}, map[string]bool{}, map[string]bool{}
		for k, v := range d.Case.canon {
			cs.canon[k] = v
		}
		for k, v := range d.Case.actorOf {
			cs.actorOf[k] = v
		}
		for k, v := range d.Case.actors {
			cs.actors[k] = v
		}
		for k, v := range d.Case.contexts {
			cs.contexts[k] = v
		}
		change(&cs)
		dd := *d
		dd.Case = &cs
		return &dd
	}
	cases := []struct {
		name, claim string
		plant       func(set map[string]Annotation) *Dataset
		want        string // "" : no problem
	}{
		{"no id", "13292.12", func(s map[string]Annotation) *Dataset { delete(inst(s["13292.12"], 2), "id"); return nil },
			"R12 an instance has the id null, not 13292.12#<n>"},
		{"another claim's id", "13292.12", func(s map[string]Annotation) *Dataset { inst(s["13292.12"], 2)["id"] = "13292.9#9"; return nil },
			"R12 an instance has the id \"13292.9#9\""},
		{"a leading zero", "13292.12", func(s map[string]Annotation) *Dataset { inst(s["13292.12"], 2)["id"] = "13292.12#03"; return nil },
			"R12 an instance has the id \"13292.12#03\""},
		{"twice", "13292.12", func(s map[string]Annotation) *Dataset { inst(s["13292.12"], 2)["id"] = "13292.12#2"; return nil },
			"R12 two instances have the id 13292.12#2"},
		{"retired", "13292.12", func(s map[string]Annotation) *Dataset { s["13292.12"]["retired"] = []any{"13292.12#3"}; return nil },
			"R12 the instance 13292.12#3 has a retired id"},
		{"an instance that is not there", "13292.12", func(s map[string]Annotation) *Dataset { actualOf(s["13292.12"])["instance"] = "13292.9#7"; return nil },
			"R13 its target names the instance 13292.9#7, which is not an instance of 13292.9"},
		{"an instance of another claim", "13292.12", func(s map[string]Annotation) *Dataset { actualOf(s["13292.12"])["instance"] = "13292.7#1"; return nil },
			"R13 its target names the instance 13292.7#1, which is not an instance of 13292.9"},
		{"a formula of another instance", "13292.12", func(s map[string]Annotation) *Dataset { actualOf(s["13292.12"])["instance"] = "13292.9#2"; return nil },
			"R13 its target names brings_about(tax_meat,tax_revenue_reduced), which is not a premise of 13292.9#2"},
		{"the wrong conclusion", "13292.9", func(s map[string]Annotation) *Dataset {
			actualOf(s["13292.9"])["formula"] = "¬should_be_performed(tax_meat)"
			return nil
		}, "R13 its target names ¬should_be_performed(tax_meat), which is not a conclusion of 13292.1#1"},
		{"a secondary instance's own formula", "13292.12", func(s map[string]Annotation) *Dataset {
			a := actualOf(s["13292.12"])
			a["instance"], a["formula"] = "13292.9#3", "causes(meat_industry_harmed,tax_revenue_reduced)"
			return nil
		}, ""},
		{"no instance named", "13292.12", func(s map[string]Annotation) *Dataset { delete(actualOf(s["13292.12"]), "instance"); return nil },
			"R16 its target names no instance"},
		{"unresolved", "13292.12", func(s map[string]Annotation) *Dataset {
			actualOf(s["13292.12"])["instance"] = map[string]any{"unresolved": "the step or the consequence argument"}
			return nil
		}, ""},
		{"an inference with a formula", "13292.12", func(s map[string]Annotation) *Dataset { actualOf(s["13292.12"])["part"] = "inference"; return nil },
			"R16 its target is the inference, which has no formula"},
		{"another Kialo parent", "13292.180", func(s map[string]Annotation) *Dataset {
			s["13292.180"]["target"].(map[string]any)["other_kialo_parents"].([]any)[0].(map[string]any)["instance"] = "13292.121#2"
			return nil
		}, "R13 its target in 13292.121 names the instance 13292.121#2"},
		{"re_parented recorded", "13292.7", func(s map[string]Annotation) *Dataset {
			s["13292.7"]["target"].(map[string]any)["re_parented"] = false
			return nil
		}, "R16 it records re_parented"},
		{"no polarity", "13292.7", func(s map[string]Annotation) *Dataset {
			s["13292.7"]["role"].(map[string]any)["polarity"] = nil
			return nil
		},
			"R16 its polarity is null"},
		{"no actual target", "13292.7", func(s map[string]Annotation) *Dataset {
			s["13292.7"]["target"].(map[string]any)["actual"] = nil
			return nil
		},
			"R16 its actual target is null"},
		{"a target for the thesis", "13292.1", func(s map[string]Annotation) *Dataset {
			s["13292.1"]["target"].(map[string]any)["actual"] = actualOf(s["13292.7"])
			return nil
		}, "R16 the thesis has no target"},
		{"no primary flag", "13292.7", func(s map[string]Annotation) *Dataset { delete(inst(s["13292.7"], 0), "primary"); return nil },
			"R16 the instance 13292.7#1 says nothing of primary"},
		{"a binding left null", "13292.7", func(s map[string]Annotation) *Dataset {
			inst(s["13292.7"], 0)["bindings"].(map[string]any)["G"] = nil
			return nil
		}, "R16 13292.7#1 binds G to null"},
		{"unstated outside a fragment", "13292.7", func(s map[string]Annotation) *Dataset {
			inst(s["13292.7"], 0)["bindings"].(map[string]any)["G"] = map[string]any{"unstated": "not said"}
			return nil
		}, `R16 13292.7#1 binds G to {"unstated":"not said"}`},
		{"an agent who is not an actor", "13292.6", func(s map[string]Annotation) *Dataset {
			renameIn(inst(s["13292.6"], 0), "governments", "meat_tax_proponents")
			return nil
		}, "R14 13292.6#1 binds the agent P to meat_tax_proponents, which is not an actor of the case"},
		{"an action the case lacks", "13292.123", func(s map[string]Annotation) *Dataset {
			renameIn(inst(s["13292.123"], 0), "abolish_meat_subsidies", "switch_to_other_protein")
			return nil
		}, "R14 13292.123#1 binds the action A2 to switch_to_other_protein, which is not an action of the case"},
		{"another actor's action", "13292.123", func(s map[string]Annotation) *Dataset {
			renameIn(inst(s["13292.123"], 0), "abolish_meat_subsidies", "switch_to_other_protein")
			return withCase(func(cs *Case) {
				cs.actors["people"] = true
				cs.canon["switch_to_other_protein"], cs.actorOf["switch_to_other_protein"] = "switch_to_other_protein", "people"
			})
		}, "R14 13292.123#1 binds A2 to switch_to_other_protein, which is people's action, not governments's"},
		{"an alias is its action", "13292.123", func(s map[string]Annotation) *Dataset {
			renameIn(inst(s["13292.123"], 1), "abolish_meat_subsidies", "meat_subsidies")
			return withCase(func(cs *Case) { cs.canon["meat_subsidies"] = "abolish_meat_subsidies" })
		}, ""},
		{"a context the case lacks", "13292.7", func(s map[string]Annotation) *Dataset {
			inst(s["13292.7"], 0)["context"] = "denmark_2005"
			return nil
		}, `R15 the instance 13292.7#1 is in the context "denmark_2005"`},
		{"an answer in another context", "13292.104", func(s map[string]Annotation) *Dataset {
			inst(s["13292.104"], 0)["context"] = "denmark_2005"
			return withCase(func(cs *Case) { cs.contexts["denmark_2005"] = true })
		}, `R15 its primary instance is in the context "denmark_2005", its target 13292.108#1 in "present"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := clone(t, d.Gold)
			dd := tc.plant(set)
			if dd == nil {
				dd = d
			}
			got := c.CheckIn(set[tc.claim], dd, set, 0)
			if tc.want == "" {
				if len(got) > 0 {
					t.Fatalf("%v", got)
				}
				return
			}
			if len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), tc.want) {
				t.Fatalf("got %v, want %q", got, tc.want)
			}
			rule := strings.Fields(tc.want)[0]
			for _, p := range got {
				if !strings.HasPrefix(p, rule+" ") && !(rule == "R12" && strings.HasPrefix(p, "R16 the instance")) {
					t.Errorf("another problem than %s: %s", rule, p)
				}
			}
		})
	}
}

// Without the annotator's other annotations (Check, not CheckIn), an id that
// names another claim's instance is still refused: by its form.
func TestAnInstanceOfAnotherClaimWithoutTheSet(t *testing.T) {
	d, c := builtin(t)
	set := clone(t, d.Gold)
	actualOf(set["13292.12"])["instance"] = "13292.7#1"
	p := c.Check(set["13292.12"], d)
	if strings.Join(p, "\n") != "R13 its target names the instance 13292.7#1, which is not an instance of 13292.9" {
		t.Fatalf("%v", p)
	}
	actualOf(set["13292.12"])["instance"] = "13292.9#7"
	if p := c.Check(set["13292.12"], d); len(p) > 0 {
		t.Fatalf("without the set, an id of the right claim cannot be looked up, so nothing is said: %v", p)
	}
}

// A format-1 annotation is checked by the rules it was made under: in a run
// made before the identity contract, the kept format-1 gold passes; checked
// as format 2 it has no ids.
func TestFormat1IsCheckedAsMade(t *testing.T) {
	d, c := builtin(t)
	b, err := os.ReadFile("../../../docs/kialo-batch2/gold-format1.json")
	if err != nil {
		t.Skip("the docs are not here:", err)
	}
	format, old, err := parseAnnotationFile(b)
	if err != nil || format != 1 {
		t.Fatalf("%d %v", format, err)
	}
	for _, id := range d.IDs() {
		if p := c.CheckIn(old[id], d, old, 1); len(p) > 0 {
			t.Errorf("%s as format 1: %v", id, p)
		}
	}
	if p := c.CheckIn(old["13292.12"], d, old, 2); !strings.Contains(strings.Join(p, "\n"), "R12 an instance has the id null") {
		t.Errorf("as format 2: %v", p)
	}
	// A run that records no format is format 1, and its file says none.
	r := &Run{Settings: Settings{}}
	if r.format() != 1 {
		t.Error("a run without a format is not format 1")
	}
	r.Settings.Format = Format2
	if r.format() != 2 {
		t.Error("a format-2 run is not format 2")
	}
}

// The agreement fields read each annotator's targets in its own set: the
// format-1 gold and the format-2 gold, the same codings but for D039's
// eleven causal steps, agree on every field but where D039 changed them.
func TestAgreementFieldsAcrossFormats(t *testing.T) {
	d, _ := builtin(t)
	b, err := os.ReadFile("../../../docs/kialo-batch2/gold-format1.json")
	if err != nil {
		t.Skip("the docs are not here:", err)
	}
	_, old, err := parseAnnotationFile(b)
	if err != nil {
		t.Fatal(err)
	}
	p := Agree("format 1", "format 2", old, d.Gold)
	want := map[string]float64{"scheme": 0.94, "instances": 0.88, "target_instance": 0.96, "target_formula": 0.96}
	for _, f := range Fields {
		fa := p.Fields[f]
		w, ok := want[f]
		if !ok {
			w = 1
		}
		if fa.Agreement != w || fa.Unresolved != 0 || fa.N != 50 {
			t.Errorf("%s: %+v, want agreement %v over 50", f, fa, w)
		}
	}
	if v := FieldIn(d.Gold["13292.12"], "target_instance", d.Gold); v != "negative_consequences:¬should_be_performed(tax_meat)" {
		t.Errorf("13292.12's target instance: %q", v)
	}
	if v := FieldIn(d.Gold["13292.45"], "target_formula", d.Gold); v != "may_occur(domestic_meat_industry_pushed_out)" {
		t.Errorf("13292.45's target formula: %q", v)
	}
	if v := FieldIn(old["13292.45"], "target_formula", old); v != "will_occur(domestic_meat_industry_pushed_out)" {
		t.Errorf("13292.45's format-1 target formula: %q", v)
	}
	if v := FieldOf(d.Gold["13292.12"], "target_instance"); v != Unresolved {
		t.Errorf("without its set, a target instance is %q, not unresolved", v)
	}
}

// An unresolved value is left out of its field's agreement and counted.
func TestAgreementLeavesUnresolvedOut(t *testing.T) {
	d, _ := builtin(t)
	a := clone(t, d.Gold)
	actualOf(a["13292.12"])["instance"] = map[string]any{"unresolved": "which inference"}
	inst(a["13292.7"], 0)["bindings"].(map[string]any)["G"] = map[string]any{"unresolved": "which harm"}
	p := Agree("a", "gold", a, d.Gold)
	for f, want := range map[string]int{"target_instance": 1, "target_formula": 1, "bindings": 1, "scheme": 0} {
		if fa := p.Fields[f]; fa.Unresolved != want || fa.N != 50-want || fa.Agreement != 1 {
			t.Errorf("%s: %+v, want %d unresolved and full agreement on the rest", f, fa, want)
		}
	}
	q := AgreeAssigned("a", "gold", a, d.Gold, d.IDs())
	if q.AllAssignedAgreement["bindings"] != 49.0/50 {
		t.Errorf("all-assigned agreement counts an unresolved value as agreeing: %v", q.AllAssignedAgreement["bindings"])
	}
	if dis := Disagrees([]Annotation{a["13292.12"], d.Gold["13292.12"]}, a, d.Gold); strings.Join(dis, ",") != "target_instance,target_formula" {
		t.Errorf("an unresolved target and a resolved one differ: %v", dis)
	}
}

// Disagreement now covers the exact target: two annotations that agree on
// the eight coarse fields but attack different premises differ.
func TestDisagreesOnTheExactTarget(t *testing.T) {
	d, _ := builtin(t)
	a := clone(t, d.Gold)
	actualOf(a["13292.33"])["formula"] = "bad(food_less_affordable_for_poor)"
	if dis := Disagrees([]Annotation{a["13292.33"], d.Gold["13292.33"]}, a, d.Gold); strings.Join(dis, ",") != "target_formula" {
		t.Errorf("%v", dis)
	}
}

// Aliases read as their actions in agreement.
func TestCanonicalReadsAliasesAsTheirActions(t *testing.T) {
	d, _ := builtin(t)
	cs := *d.Case
	cs.canon = map[string]string{}
	for k, v := range d.Case.canon {
		cs.canon[k] = v
	}
	cs.canon["meat_tax"] = "tax_meat"
	a := clone(t, d.Gold)
	renameIn(inst(a["13292.7"], 0), "tax_meat", "meat_tax")
	if FieldIn(a["13292.7"], "bindings", a) == FieldIn(d.Gold["13292.7"], "bindings", d.Gold) {
		t.Fatal("the plant did not change the bindings")
	}
	ca := cs.CanonicalSet(a)
	for _, f := range Fields {
		if x, y := FieldIn(ca["13292.7"], f, ca), FieldIn(d.Gold["13292.7"], f, d.Gold); x != y {
			t.Errorf("%s: %q, %q", f, x, y)
		}
	}
	// A predicate's name is not an atom: an alias that is also a predicate
	// name leaves the predicate alone.
	cs.canon["brings_about"] = "tax_meat"
	if got := cs.canonFormula("brings_about(brings_about,G)"); got != "brings_about(tax_meat,G)" {
		t.Errorf("%q", got)
	}
}

// The case record's own checks.
func TestCaseRecordChecks(t *testing.T) {
	b, err := data.Annotation.ReadFile("annotation/kialo-13292-batch2/case.json")
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct{ old, new, want string }{
		"an alias that is an action": {`"tax_meat", "actor": "governments", "aliases": []`, `"tax_meat", "actor": "governments", "aliases": ["raise_import_standards"]`, "cannot be an alias of tax_meat"},
		"an actor the case lacks":    {`"raise_import_standards", "actor": "governments"`, `"raise_import_standards", "actor": "people"`, `performed by "people"`},
		"no default context":         {`"default_context": "present"`, `"default_context": "past"`, `its default context "past"`},
		"the thesis's action":        {`"action": "tax_meat"}`, `"action": "tax_fish"}`, `the thesis's action "tax_fish"`},
		"another format":             {`"wag-case/1"`, `"wag-case/2"`, "not wag-case/1"},
	} {
		t.Run(name, func(t *testing.T) {
			s := string(b)
			if strings.Count(s, tc.old) != 1 {
				t.Fatalf("%q is not in case.json once", tc.old)
			}
			if _, err := parseCase([]byte(strings.Replace(s, tc.old, tc.new, 1))); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%v, want %q", err, tc.want)
			}
		})
	}
}

// The typed places: no variable of any walton.yml scheme is both an agent
// and an action.
func TestTypedPlacesAreConsistent(t *testing.T) {
	w, err := data.Annotation.ReadFile("annotation/walton.yml")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Schemes []waltonScheme `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal(w, &f); err != nil {
		t.Fatal(err)
	}
	for _, sc := range f.Schemes {
		kinds := map[string]map[string]bool{}
		for _, fs := range [][]string{sc.Premises, sc.Assumptions, sc.Exceptions, sc.Conclusions} {
			for _, x := range fs {
				name, args := splitFormula(x)
				for kind, table := range map[string]map[string][]int{"agent": typedPlaces.Agent, "action": typedPlaces.Action} {
					for _, i := range table[name+"/"+itoa(len(args))] {
						if kinds[args[i-1]] == nil {
							kinds[args[i-1]] = map[string]bool{}
						}
						kinds[args[i-1]][kind] = true
					}
				}
			}
		}
		for v, k := range kinds {
			if len(k) > 1 {
				t.Errorf("%s: %s is both an agent and an action", sc.ID, v)
			}
		}
	}
	if typedVars(f.Schemes[0]) == nil {
		t.Fatal("typedVars")
	}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

// The prompt lists every earlier instance by its id, so that a model can
// target a secondary one.
func TestThePromptListsEveryInstance(t *testing.T) {
	d, _ := builtin(t)
	it := d.Item("13292.43")
	u := d.AnnotatorPrompt(it, d.Gold, []string{"13292.1", "13292.7", "13292.9"})
	for _, want := range []string{"every instance, by its id", "- 13292.9#2 (cause_to_effect_modal (reconstructed)):", "- 13292.9#1 (negative_consequences (reconstructed), primary):"} {
		if !strings.Contains(u, want) {
			t.Errorf("the prompt lacks %q:\n%s", want, u)
		}
	}
}

// A run's annotation file says its format; agreement.py reads both formats
// as the Go measure does (the same fields, the same figures).
func TestAgreementPyMatchesGoAcrossFormats(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "agreement.json")
	cmd := exec.Command(py, "-I", "-B", "agreement.py", "gold-format1.json", "gold.json", "--json", out)
	cmd.Dir = dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, b)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var py1 struct {
		Fields map[string]struct {
			Agreement  float64  `json:"agreement"`
			Kappa      *float64 `json:"kappa"`
			Unresolved int      `json:"unresolved"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(raw, &py1); err != nil {
		t.Fatal(err)
	}
	d, _ := builtin(t)
	b, _ := os.ReadFile(filepath.Join(dir, "gold-format1.json"))
	_, old, _ := parseAnnotationFile(b)
	p := Agree("1", "2", old, d.Gold)
	for _, f := range Fields {
		w, ok := py1.Fields[f]
		if !ok {
			t.Errorf("agreement.py has no %s", f)
			continue
		}
		g := p.Fields[f]
		if g.Agreement != w.Agreement || g.Unresolved != w.Unresolved || (g.Kappa == nil) != (w.Kappa == nil) || g.Kappa != nil && *g.Kappa != *w.Kappa {
			t.Errorf("%s: go %+v, agreement.py %+v", f, g, w)
		}
	}
}

// A new run records the format it is made in, its annotation files say it,
// and the gold's codings it receives pass its check.
func TestARunRecordsItsFormat(t *testing.T) {
	r, _, items := soloRun(t, 2)
	if r.Settings.Format != Format2 || r.format() != 2 {
		t.Fatalf("a run on the built-in dataset is %q", r.Settings.Format)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(r.annotationsPath("solo"))
	if err != nil {
		t.Fatal(err)
	}
	format, anns, err := parseAnnotationFile(b)
	if err != nil || format != 2 || len(anns) != 2 {
		t.Fatalf("its annotation file: format %d, %d annotations, %v", format, len(anns), err)
	}
	rb, err := os.ReadFile(filepath.Join(r.Dir, "run.json"))
	if err != nil || !strings.Contains(string(rb), `"format": "wag-annotation/2"`) {
		t.Fatalf("run.json does not record the format: %v", err)
	}
	var head struct {
		Case string `json:"case"`
	}
	if err := json.Unmarshal(b, &head); err != nil || head.Case != "kialo-13292" {
		t.Fatalf("its annotation file names the case %q, not the dataset's: %v", head.Case, err)
	}
	for _, id := range items {
		v, err := r.Item(id)
		if err != nil {
			t.Fatal(err)
		}
		if p := v.Problems["solo"]; len(p) > 0 {
			t.Errorf("%s: %v", id, p)
		}
	}
}

// A dataset's format: its manifest's; with none, its gold's (format 2
// without a gold); a gold in another format than the manifest names is
// refused.
func TestADatasetsFormat(t *testing.T) {
	items := `{"discussion": {"thesis_claim_id": "1.1"}, "items": [{"kialo_claim_id": "1.1", "text": "A."}]}`
	gold1 := `{"annotator": "gold", "items": [{"kialo_claim_id": "1.1", "instances": [{"primary": true}]}]}`
	gold2 := `{"format": "wag-annotation/2", "annotator": "gold", "items": [{"kialo_claim_id": "1.1", "instances": [{"id": "1.1#1", "primary": true}]}]}`
	for _, c := range []struct {
		name    string
		files   map[string]string
		format  int
		problem string
	}{
		{"no manifest, no gold", map[string]string{"items.json": items}, 2, ""},
		{"no manifest, a format-1 gold", map[string]string{"items.json": items, "gold.json": gold1}, 1, ""},
		{"no manifest, a format-2 gold", map[string]string{"items.json": items, "gold.json": gold2}, 2, ""},
		{"format 2 named, a format-1 gold", map[string]string{"items.json": items, "gold.json": gold1, "dataset.json": `{"format": "wag-annotation/2"}`}, 0,
			"gold.json is in format 1, but the dataset's guidelines ask for format 2"},
		{"format 1 named, a format-2 gold", map[string]string{"items.json": items, "gold.json": gold2, "dataset.json": `{"format": "wag-annotation/1"}`}, 0,
			"gold.json is in format 2, but the dataset's guidelines ask for format 1"},
		{"format 1 named, no gold", map[string]string{"items.json": items, "dataset.json": `{"format": "wag-annotation/1"}`}, 1, ""},
		{"an unknown format", map[string]string{"items.json": items, "dataset.json": `{"format": "wag-annotation/3"}`}, 0, `format "wag-annotation/3" is not`},
	} {
		t.Run(c.name, func(t *testing.T) {
			d, err := FromDir(dataset(t, c.files))
			if c.problem != "" {
				if err == nil || !strings.Contains(err.Error(), c.problem) {
					t.Fatalf("%v, want %q", err, c.problem)
				}
				return
			}
			if err != nil || d.Format != c.format {
				t.Fatalf("format %v, %v; want %d", d, err, c.format)
			}
		})
	}
}

// Two unresolved values are not agreement either: all-assigned agreement
// counts only resolved values alike.
func TestTwoUnresolvedValuesDoNotAgree(t *testing.T) {
	d, _ := builtin(t)
	a := clone(t, d.Gold)
	actualOf(a["13292.12"])["instance"] = map[string]any{"unresolved": "which inference"}
	q := AgreeAssigned("a", "a again", a, clone(t, a), d.IDs())
	if got := q.AllAssignedAgreement["target_instance"]; got != 49.0/50 {
		t.Errorf("all-assigned agreement on target_instance: %v, want 49/50", got)
	}
	if fa := q.Fields["target_instance"]; fa.Unresolved != 1 || fa.N != 49 || fa.Agreement != 1 {
		t.Errorf("%+v", fa)
	}
}

// agreement.py and the Go measure agree when an annotator uses an alias the
// case lists (read as its action with --case, and as another atom without)
// and leaves a fragment's variable unstated.
func TestAgreementPyMatchesGoWithAliases(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	d, _ := builtin(t)
	a := clone(t, d.Gold)
	renameIn(inst(a["13292.7"], 0), "tax_meat", "meat_tax")
	inst(a["13292.6"], 0)["bindings"].(map[string]any)["S1"] = map[string]any{"unstated": "the claim states the goal's link only"}
	tmp := t.TempDir()
	write := func(name string, v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(tmp, name)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	var items []Annotation
	for _, id := range d.IDs() {
		items = append(items, a[id])
	}
	pa := write("a.json", map[string]any{"format": Format2, "annotator": "a", "items": items})
	raw, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cj map[string]any
	if err := json.Unmarshal(raw, &cj); err != nil {
		t.Fatal(err)
	}
	cj["actions"].([]any)[0].(map[string]any)["aliases"] = []any{"meat_tax"}
	pc := write("case.json", cj)
	cs, err := parseCase([]byte(mustJSON(t, cj)))
	if err != nil {
		t.Fatal(err)
	}
	for _, withCase := range []bool{false, true} {
		args := []string{"-I", "-B", "agreement.py", pa, "gold.json", "--json", filepath.Join(tmp, "out.json")}
		ga, gb := a, d.Gold
		if withCase {
			args = append(args, "--case", pc)
			ga, gb = cs.CanonicalSet(a), cs.CanonicalSet(d.Gold)
		}
		cmd := exec.Command(py, args...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, b)
		}
		var out struct {
			Fields map[string]struct {
				Agreement  float64  `json:"agreement"`
				Kappa      *float64 `json:"kappa"`
				Unresolved int      `json:"unresolved"`
			} `json:"fields"`
		}
		b, _ := os.ReadFile(filepath.Join(tmp, "out.json"))
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}
		p := Agree("a", "gold", ga, gb)
		for _, f := range Fields {
			g, w := p.Fields[f], out.Fields[f]
			if g.Agreement != w.Agreement || g.Unresolved != w.Unresolved || (g.Kappa == nil) != (w.Kappa == nil) || g.Kappa != nil && *g.Kappa != *w.Kappa {
				t.Errorf("case %v, %s: go %+v, agreement.py %+v", withCase, f, g, w)
			}
		}
		// The alias changes 13292.7's bindings unless the case reads it as its action.
		if want := map[bool]float64{false: 0.96, true: 0.98}[withCase]; p.Fields["bindings"].Agreement != want {
			t.Errorf("case %v: bindings agreement %v, want %v", withCase, p.Fields["bindings"].Agreement, want)
		}
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A run's annotation file is a format-2 file validate.py reads as the
// batch-2 README says (the gold's codings, copied by a stand-in annotator,
// so that any complaint is the file's, not the coding's).
func TestARunsFilePassesValidatePy(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	r, _, _ := soloRun(t, len(ds[0].IDs())-1)
	// The thesis too: validate.py checks a whole file of the batch's claims.
	r.Settings.Items = ds[0].IDs()
	r.anns["solo"]["13292.1"] = ds[0].Gold["13292.1"]
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	r.mu.Lock()
	err = r.saveAnnotations("solo")
	r.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(py, "-I", "-B", "validate.py", r.annotationsPath("solo"))
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
}
