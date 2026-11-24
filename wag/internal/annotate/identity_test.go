package annotate

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/llm"
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
		{"an empty actual target", "13292.7", func(s map[string]Annotation) *Dataset {
			s["13292.7"]["target"].(map[string]any)["actual"] = map[string]any{}
			return nil
		},
			"R16 its actual target is {}"},
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
				// A lone instance without its primary flag also has no
				// primary instance (format 2 needs exactly one).
				if !strings.HasPrefix(p, rule+" ") && !(rule == "R12" && strings.HasPrefix(p, "R16 the instance")) &&
					!(tc.name == "no primary flag" && p == "0 primary instances, not 1") {
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
	// target_instance is the targeting decision (the primary instance, or
	// which secondary one): D039 changed the coding of 13292.12's primary,
	// not which of its instances 13292.194 and .45 target.
	want := map[string]float64{"scheme": 0.94, "instances": 0.88, "target_formula": 0.96}
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
	if v := FieldIn(d.Gold["13292.12"], "target_instance", d.Gold); v != "primary" {
		t.Errorf("13292.12's target instance: %q", v)
	}
	// A target at a secondary instance says which.
	a := clone(t, d.Gold)
	act := actualOf(a["13292.12"])
	act["instance"], act["formula"] = "13292.9#3", "causes(meat_industry_harmed,tax_revenue_reduced)"
	if v := FieldIn(a["13292.12"], "target_instance", a); v != "secondary:cause_to_effect_modal:may_occur(tax_revenue_reduced)" {
		t.Errorf("a secondary target: %q", v)
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
	// Only what a field itself leaves open makes it unresolved: the formula
	// 13292.12 names is still compared.
	for f, want := range map[string]int{"target_instance": 1, "target_formula": 0, "bindings": 1, "scheme": 0} {
		if fa := p.Fields[f]; fa.Unresolved != want || fa.N != 50-want || fa.Agreement != 1 {
			t.Errorf("%s: %+v, want %d unresolved and full agreement on the rest", f, fa, want)
		}
	}
	q := AgreeAssigned("a", "gold", a, d.Gold, d.IDs())
	if q.AllAssignedAgreement["bindings"] != 49.0/50 {
		t.Errorf("all-assigned agreement counts an unresolved value as agreeing: %v", q.AllAssignedAgreement["bindings"])
	}
	if dis := Disagrees([]Annotation{a["13292.12"], d.Gold["13292.12"]}, a, d.Gold); strings.Join(dis, ",") != "target_instance" {
		t.Errorf("an unresolved target instance and a resolved one differ: %v", dis)
	}
	// An unresolved formula is unresolved, even of a conclusion target whose
	// instance has one conclusion.
	act := actualOf(a["13292.45"])
	if act["part"] != "conclusion" {
		t.Fatalf("13292.45's target is not a conclusion: %v", act)
	}
	act["formula"] = map[string]any{"unresolved": "which step"}
	if v := FieldIn(a["13292.45"], "target_formula", a); v != Unresolved {
		t.Errorf("an unresolved conclusion formula: %q", v)
	}
}

// A format-1 target names no instance: it is its claim's primary one, also
// where a lone instance carries no primary flag (format 1 allows that).
func TestFormat1TargetOfAnUnflaggedLoneInstance(t *testing.T) {
	b, err := os.ReadFile("../../../docs/kialo-batch2/gold-format1.json")
	if err != nil {
		t.Skip("the docs are not here:", err)
	}
	_, old, err := parseAnnotationFile(b)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(old["13292.108"]["instances"].([]any)); n != 1 {
		t.Fatalf("13292.108 has %d instances", n)
	}
	delete(inst(old["13292.108"], 0), "primary")
	if act := actualOf(old["13292.104"]); act["claim"] != "13292.108" || act["instance"] != nil && !strings.HasPrefix(jsonText(act["instance"]), `"primary`) {
		t.Fatalf("13292.104's format-1 target: %v", act)
	}
	if v := FieldIn(old["13292.104"], "target_instance", old); v != "primary" {
		t.Errorf("%q", v)
	}
	// Format 1's check admits it too (format 2 needs the flag).
	d, c := builtin(t)
	for _, p := range c.CheckIn(old["13292.108"], d, old, 1) {
		if strings.Contains(p, "primary instances") {
			t.Errorf("format 1: %s", p)
		}
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

// agreement.py and the Go measure agree on the target fields' edge cases: an
// unresolved formula of a conclusion target, two different secondary
// instances, an unresolved instance, and (format 1) a target whose lone
// instance carries no primary flag.
func TestAgreementPyMatchesGoOnTargets(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	d, _ := builtin(t)
	tmp := t.TempDir()
	write := func(name, format string, set map[string]Annotation) string {
		var items []Annotation
		for _, id := range d.IDs() {
			items = append(items, set[id])
		}
		head := map[string]any{"annotator": name, "items": items}
		if format != "" {
			head["format"] = format
		}
		p := filepath.Join(tmp, name+".json")
		if err := os.WriteFile(p, []byte(mustJSON(t, head)), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	a, b := clone(t, d.Gold), clone(t, d.Gold)
	actualOf(a["13292.45"])["formula"] = map[string]any{"unresolved": "which step"}
	at, bt := actualOf(a["13292.194"]), actualOf(b["13292.194"])
	at["instance"], at["formula"] = "13292.12#2", "may_occur(imported_meat_relatively_cheaper)"
	bt["instance"], bt["formula"] = "13292.12#3", "¬should_be_performed(tax_meat)"
	actualOf(b["13292.33"])["instance"] = map[string]any{"unresolved": "which inference"}
	raw, err := os.ReadFile(filepath.Join(dir, "gold-format1.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, old, err := parseAnnotationFile(raw)
	if err != nil {
		t.Fatal(err)
	}
	delete(inst(old["13292.108"], 0), "primary")
	pairs := []struct {
		name   string
		x, y   map[string]Annotation
		px, py string
	}{
		{"two format-2 sets", a, b, write("a", Format2, a), write("b", Format2, b)},
		{"format 1, an unflagged lone instance", old, d.Gold, write("old", "", old), filepath.Join(dir, "gold.json")},
	}
	for _, pr := range pairs {
		out := filepath.Join(tmp, "out.json")
		cmd := exec.Command(py, "-I", "-B", "agreement.py", pr.px, pr.py, "--json", out)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v: %s", pr.name, err, b)
		}
		var res struct {
			Fields map[string]struct {
				Agreement  float64  `json:"agreement"`
				Kappa      *float64 `json:"kappa"`
				Unresolved int      `json:"unresolved"`
			} `json:"fields"`
		}
		rb, _ := os.ReadFile(out)
		if err := json.Unmarshal(rb, &res); err != nil {
			t.Fatal(err)
		}
		p := Agree("x", "y", pr.x, pr.y)
		for _, f := range Fields {
			g, w := p.Fields[f], res.Fields[f]
			if g.Agreement != w.Agreement || g.Unresolved != w.Unresolved || (g.Kappa == nil) != (w.Kappa == nil) || g.Kappa != nil && *g.Kappa != *w.Kappa {
				t.Errorf("%s, %s: go %+v, agreement.py %+v", pr.name, f, g, w)
			}
		}
	}
	// The plants are seen: the secondary instances differ, the unresolved
	// values are counted, and the unflagged instance is its claim's primary.
	p := Agree("a", "b", a, b)
	if ti, tf := p.Fields["target_instance"], p.Fields["target_formula"]; ti.Unresolved != 1 || tf.Unresolved != 1 || ti.Agreement != 0.98 { // 48 of 49, rounded
		t.Errorf("target_instance %+v, target_formula %+v", ti, tf)
	}
	if ti := Agree("old", "gold", old, d.Gold).Fields["target_instance"]; ti.Agreement != 1 {
		t.Errorf("format 1, an unflagged lone instance: %+v", ti)
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

// target_instance compares the targeting decision, not the target's own
// coding: a coding difference in 13292.108 does not spread to the eleven
// claims that target its primary instance (the review of the identity
// contract, W3).
func TestTargetInstanceDoesNotSpread(t *testing.T) {
	d, _ := builtin(t)
	a := clone(t, d.Gold)
	cm := inst(a["13292.108"], 0)["conclusion"].(map[string]any)
	cm["formula"] = "should_be_performed(abolish_meat_subsidies)"
	for _, id := range d.IDs() {
		if id == "13292.108" {
			continue
		}
		if x, y := FieldIn(a[id], "target_instance", a), FieldIn(d.Gold[id], "target_instance", d.Gold); x != y {
			t.Errorf("%s: %q, %q", id, x, y)
		}
	}
}

// The review choice reads aliases as their actions, as agreement does (W1).
func TestReviewChoiceReadsAliases(t *testing.T) {
	r, _, _ := soloRun(t, 1)
	d := r.Dataset
	cs := *d.Case
	cs.canon = map[string]string{}
	for k, v := range d.Case.canon {
		cs.canon[k] = v
	}
	cs.canon["meat_tax_case"] = "tax_meat"
	dd := *d
	dd.Case = &cs
	r.Dataset = &dd
	a, b := clone(t, d.Gold), clone(t, d.Gold)
	renameIn(inst(a["13292.7"], 0), "tax_meat", "meat_tax_case")
	r.anns["x"], r.anns["y"] = a, b
	if dis := r.disagrees([]string{"x", "y"}, []Annotation{a["13292.7"], b["13292.7"]}); len(dis) > 0 {
		t.Errorf("an alias makes a disagreement: %v", dis)
	}
	if dis := Disagrees([]Annotation{a["13292.7"], b["13292.7"]}, a, b); len(dis) == 0 {
		t.Error("the plant did not change the bindings")
	}
	// The targets are read in canonical sets too: a target at a secondary
	// instance whose conclusion one annotator writes with an alias.
	a, b = clone(t, d.Gold), clone(t, d.Gold)
	for _, s := range []map[string]Annotation{a, b} {
		act := actualOf(s["13292.45"])
		act["claim"], act["instance"], act["part"], act["formula"] = "13292.12", "13292.12#3", "conclusion", "¬should_be_performed(tax_meat)"
	}
	renameIn(targetInstance(a, "13292.12", "13292.12#3"), "tax_meat", "meat_tax_case")
	r.anns["x"], r.anns["y"] = a, b
	if dis := r.disagrees([]string{"x", "y"}, []Annotation{a["13292.45"], b["13292.45"]}); len(dis) > 0 {
		t.Errorf("an alias in a target makes a disagreement: %v", dis)
	}
	if x, y := FieldIn(a["13292.45"], "target_instance", a), FieldIn(b["13292.45"], "target_instance", b); x == y {
		t.Errorf("the plant did not change the target instance: %q", x)
	}
}

// An instance with two agents binds an action of either of them (R14, as
// validate.py reads it). No scheme of walton.yml has two agent places yet:
// the test makes one.
func TestR14TwoAgents(t *testing.T) {
	d, c := builtin(t)
	c.schemes["two_agents"] = waltonScheme{ID: "two_agents", Variables: []string{"P", "Q", "A", "B", "G"},
		Premises:    []string{"goal_of_agent(P,G)", "goal_of_agent(Q,G)", "acceptable_action(P,A)", "acceptable_action(Q,B)"},
		Conclusions: []string{"should_be_performed(A)"}}
	cs := *d.Case
	cs.actors = map[string]bool{"governments": true, "people": true, "farmers": true}
	cs.canon = map[string]string{"tax_meat": "tax_meat", "eat_less_meat": "eat_less_meat", "raise_cattle": "raise_cattle"}
	cs.actorOf = map[string]string{"tax_meat": "governments", "eat_less_meat": "people", "raise_cattle": "farmers"}
	in := func(b string) map[string]any {
		return map[string]any{"scheme": map[string]any{"walton_id": "two_agents"},
			"bindings": map[string]any{"P": "governments", "Q": "people", "A": "tax_meat", "B": b, "G": "lower_emissions"}}
	}
	if p := c.checkActors(in("eat_less_meat"), "x#1", &cs); len(p) > 0 {
		t.Errorf("an action of the second agent: %v", p)
	}
	if p := c.checkActors(in("raise_cattle"), "x#1", &cs); strings.Join(p, "|") != "R14 x#1 binds B to raise_cattle, which is farmers's action, not governments or people's" {
		t.Errorf("an action of neither agent: %v", p)
	}
}

// One primary instance in format 2, a lone instance too; a fragment whose
// conclusion names an unstated variable has a null conclusion formula.
func TestFormat2OnePrimaryAndFragmentConclusions(t *testing.T) {
	d, c := builtin(t)
	set := clone(t, d.Gold)
	inst(set["13292.7"], 0)["primary"] = false
	if p := c.CheckIn(set["13292.7"], d, set, 0); strings.Join(p, "|") != "0 primary instances, not 1" {
		t.Errorf("a lone instance that is not primary: %v", p)
	}
	set = clone(t, d.Gold)
	frag := inst(set["13292.6"], 0)
	frag["bindings"].(map[string]any)["A"] = map[string]any{"unstated": "the claim states the goal's link only"}
	var keep []any
	for _, x := range frag["premises"].([]any) {
		if !strings.Contains(x.(map[string]any)["formula"].(string), "tax_meat") {
			keep = append(keep, x)
		}
	}
	frag["premises"] = keep
	if p := c.CheckIn(set["13292.6"], d, set, 0); !strings.Contains(strings.Join(p, "|"), "is a fragment whose conclusion names a variable it leaves unstated") {
		t.Errorf("%v", p)
	}
	frag["conclusion"].(map[string]any)["formula"] = nil
	if p := c.CheckIn(set["13292.6"], d, set, 0); len(p) > 0 {
		t.Errorf("with a null conclusion: %v", p)
	}
}

// Format 1 has no such rule: its fragments wrote their conclusion's formula
// (docs/annotation-records.md §5 is format 2's).
func TestFormat1FragmentConclusions(t *testing.T) {
	d, c := builtin(t)
	b, err := os.ReadFile("../../../docs/kialo-batch2/gold-format1.json")
	if err != nil {
		t.Skip("the docs are not here:", err)
	}
	_, old, err := parseAnnotationFile(b)
	if err != nil {
		t.Fatal(err)
	}
	frag := inst(old["13292.6"], 0)
	if fit := frag["scheme"].(map[string]any)["fit"]; fit != "fragment" {
		t.Fatalf("13292.6 is %v", fit)
	}
	delete(frag["bindings"].(map[string]any), "A")
	var keep []any
	for _, x := range frag["premises"].([]any) {
		if !strings.Contains(x.(map[string]any)["formula"].(string), "tax_meat") {
			keep = append(keep, x)
		}
	}
	frag["premises"] = keep
	for _, p := range c.CheckIn(old["13292.6"], d, old, 1) {
		if strings.Contains(p, "is a fragment whose conclusion names a variable it leaves unstated") {
			t.Errorf("format 1: %s", p)
		}
	}
}

// The thesis has no polarity; a case record lists aliases for every action
// and declares each actor once (N1 of the review).
func TestThesisPolarityAndCaseRecordGaps(t *testing.T) {
	d, c := builtin(t)
	set := clone(t, d.Gold)
	set["13292.1"]["role"].(map[string]any)["polarity"] = "for"
	if p := c.CheckIn(set["13292.1"], d, set, 0); strings.Join(p, "|") != "R16 the thesis has no polarity (null)" {
		t.Errorf("%v", p)
	}
	b, err := data.Annotation.ReadFile("annotation/kialo-13292-batch2/case.json")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for name, tc := range map[string]struct{ old, new, want string }{
		"no list of aliases": {`"tax_meat", "actor": "governments", "aliases": []`, `"tax_meat", "actor": "governments"`, "has no list of aliases"},
		"an actor twice":     {`"actors": [`, `"actors": [{"id": "governments", "text": "again"}, `, `already used ("governments")`},
	} {
		if strings.Count(s, tc.old) != 1 {
			t.Fatalf("%s: %q is not in case.json once", name, tc.old)
		}
		if _, err := parseCase([]byte(strings.Replace(s, tc.old, tc.new, 1))); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v, want %q", name, err, tc.want)
		}
	}
}

// A format-2 run on a dataset without a case record tells the annotators so,
// in each claim's prompt: the guidelines bind agents and actions by one (W10
// of the review). A format-1 run, or a dataset with one, says nothing.
func TestADatasetWithoutACaseRecordSaysSo(t *testing.T) {
	items := `{"discussion": {"thesis_claim_id": "1.1"}, "items": [{"kialo_claim_id": "1.1", "text": "A."}]}`
	runOn := func(files map[string]string) *Run {
		d, err := FromDir(dataset(t, files))
		if err != nil {
			t.Fatal(err)
		}
		r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{{Name: "ann", Provider: "mock"}}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	const note = "This dataset has no case record (case.json)"
	// What the annotator is sent, on a format-2 dataset without one.
	d, err := FromDir(dataset(t, map[string]string{"items.json": items}))
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var sent []llm.Request
	p := llm.Profile{Name: "ann", Provider: "mock", Params: llm.Params{Model: "m1"}}
	c := llm.WithProvider(p, &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		mu.Lock()
		sent = append(sent, req)
		mu.Unlock()
		return &llm.Response{Text: `{"kialo_claim_id": "1.1"}`, Model: "m1"}, nil
	}})
	c.Sleep = llm.NoWait
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{p}}, map[string]*llm.Client{"ann": c})
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Annotate(context.Background(), nil) // its reply is not an annotation: only the request matters
	mu.Lock()
	defer mu.Unlock()
	if r.format() != 2 || len(sent) == 0 || len(sent[0].Messages) == 0 || !strings.Contains(sent[0].Messages[0].Content, note) {
		t.Fatalf("format %d: the annotator is not told the dataset has no case record: %+v", r.format(), sent)
	}
	if strings.Contains(sent[0].System, note) {
		t.Error("the note is in the system prompt, which the run pins")
	}
	old := runOn(map[string]string{"items.json": items, "gold.json": `{"annotator": "gold", "items": [{"kialo_claim_id": "1.1", "instances": [{"primary": true}]}]}`})
	if p := old.annotatorPrompt(old.Dataset.Item("1.1"), nil, nil); old.format() != 1 || strings.Contains(p, note) {
		t.Errorf("a format-%d run says the dataset has no case record", old.format())
	}
	b, _, ids := soloRun(t, 1)
	if p := b.annotatorPrompt(b.Dataset.Item(ids[0]), nil, nil); strings.Contains(p, note) {
		t.Error("the built-in dataset has a case record, but its prompt says it has none")
	}
}

// The case record is checked against the dataset when it is read: its
// thesis is a claim, its atoms are the lexicon's (W10 of the review).
func TestTheCaseRecordAgainstTheDataset(t *testing.T) {
	items := `{"discussion": {"thesis_claim_id": "1.1"}, "items": [{"kialo_claim_id": "1.1", "text": "A."}]}`
	lexicon := `{"tax_meat": "governments put a tax on meat", "governments": "governments"}`
	caseOf := func(thesis, action, alias string) string {
		return `{"format": "wag-case/1", "id": "c", "thesis": {"claim": "` + thesis + `", "action": "` + action + `"},
 "default_context": "now", "contexts": [{"id": "now", "text": "now"}], "actors": [{"id": "governments", "text": "g"}],
 "actions": [{"id": "` + action + `", "actor": "governments", "aliases": [` + alias + `]}]}`
	}
	for _, c := range []struct{ name, cas, want string }{
		{"sound", caseOf("1.1", "tax_meat", ""), ""},
		{"a thesis that is not a claim", caseOf("9.9", "tax_meat", ""), `its thesis "9.9" is not a claim of the dataset`},
		{"an action outside the lexicon", caseOf("1.1", "tax_fish", ""), "the action tax_fish is not an atom of the lexicon"},
		{"an alias outside the lexicon", caseOf("1.1", "tax_meat", `"meat_tax"`), "the alias meat_tax is not an atom of the lexicon"},
		{"an actor outside the lexicon", strings.ReplaceAll(caseOf("1.1", "tax_meat", ""), `"governments"`, `"farmers"`), "the actor farmers is not an atom of the lexicon"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := FromDir(dataset(t, map[string]string{"items.json": items, "lexicon.json": lexicon, "case.json": c.cas}))
			if c.want == "" && err != nil || c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)) {
				t.Fatalf("%v, want %q", err, c.want)
			}
		})
	}
}
