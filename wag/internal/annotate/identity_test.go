package annotate

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
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

// asAnnotator makes a gold's claim an annotator's record: no adjudication,
// and its primary instance first, as <claim>#1, as an annotator numbers it
// (R12); a gold keeps the ids its history gave them (since D008-D074,
// 13292.9's bridge is #4), so the primary and #1 swap ids, and every record
// in the set that names them follows.
func asAnnotator(s map[string]Annotation, claim string) {
	delete(s[claim], "adjudication")
	insts := s[claim]["instances"].([]any)
	first, primary := claim+"#1", ""
	for _, x := range insts {
		if p, _ := x.(map[string]any)["primary"].(bool); p {
			primary = x.(map[string]any)["id"].(string)
		}
	}
	if primary == "" || primary == first {
		return
	}
	swap := func(v any) any {
		switch v {
		case primary:
			return first
		case first:
			return primary
		}
		return v
	}
	for _, a := range s {
		for _, x := range a["instances"].([]any) {
			m := x.(map[string]any)
			m["id"] = swap(m["id"])
		}
		for _, c := range a["challenges"].([]any) {
			m := c.(map[string]any)
			if _, ok := m["instance"]; ok {
				m["instance"] = swap(m["instance"])
			}
		}
		t := a["target"].(map[string]any)
		if act, ok := t["actual"].(map[string]any); ok {
			if _, has := act["instance"]; has {
				act["instance"] = swap(act["instance"])
			}
		}
		if others, ok := t["other_kialo_parents"].([]any); ok {
			for _, o := range others {
				m := o.(map[string]any)
				if _, has := m["instance"]; has {
					m["instance"] = swap(m["instance"])
				}
			}
		}
	}
	sort.SliceStable(insts, func(i, j int) bool {
		pi, _ := insts[i].(map[string]any)["primary"].(bool)
		pj, _ := insts[j].(map[string]any)["primary"].(bool)
		return pi && !pj
	})
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

// atomArg is an atom: a lower-case word, as the lexicon writes one.
var atomArg = regexp.MustCompile(`[a-z_][A-Za-z0-9_]*`)

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
	if f, ok := cm["formula"].(string); ok { // a fragment's unstated conclusion stays null
		cm["formula"] = sub(f)
	}
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
		// 13292.12's target is 13292.9#2's causal premise since D086 round 1;
		// 13292.9#1, the modal argument, does not have it.
		{"a formula of another instance", "13292.12", func(s map[string]Annotation) *Dataset { actualOf(s["13292.12"])["instance"] = "13292.9#1"; return nil },
			"R13 its target names causes(tax_meat,meat_industry_harmed), which is not a premise of 13292.9#1"},
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
			s["13292.180"]["target"].(map[string]any)["other_kialo_parents"].([]any)[0].(map[string]any)["instance"] = "13292.121#3"
			return nil
		}, "R13 its target in 13292.121 names the instance 13292.121#3"},
		{"a secondary Kialo parent's relation", "13292.180", func(s map[string]Annotation) *Dataset {
			s["13292.180"]["target"].(map[string]any)["other_kialo_parents"].([]any)[0].(map[string]any)["relation"] = "con"
			return nil
		}, "R7 its other Kialo parents are [13292.121 con]; items.json has [13292.121 pro]"},
		{"the Kialo parent's relation", "13292.180", func(s map[string]Annotation) *Dataset {
			s["13292.180"]["target"].(map[string]any)["kialo_relation"] = "con"
			return nil
		}, "R7 its Kialo parent is 13292.7 con; items.json has 13292.7 pro"},
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
		// 13292.45 binds the agent and no other item points into it (13292.6,
		// a fragment since D086 round 1, leaves the agent unstated).
		{"an agent who is not an actor", "13292.45", func(s map[string]Annotation) *Dataset {
			renameIn(inst(s["13292.45"], 0), "governments", "meat_tax_proponents")
			return nil
		}, "R14 13292.45#1 binds the agent P to meat_tax_proponents, which is not an actor of the case"},
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
		}, `R15 its primary instance is in the context "denmark_2005", its target 13292.1#2 in "present"`},
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
	d, _ := builtin(t)
	b, err := os.ReadFile("../../../docs/kialo-batch2/gold-format1.json")
	if err != nil {
		t.Skip("the docs are not here:", err)
	}
	// It is checked against the walton.yml it was made with, which format1/
	// keeps beside it (D039: kept as it was), not as walton.yml is since.
	w1, err := os.ReadFile("../../../docs/kialo-batch2/format1/walton.yml")
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadCollection(w1)
	if err != nil {
		t.Fatal(err)
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
// eleven causal steps and proposal D008-D074's recode (2026-11-25), agree on
// every field but where those changed them, where step C5.2's switch
// (2026-11-27) named the game's questions, and where D086's round-1 draft
// (2026-11-27) named causal steps and left fragments' variables unstated.
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
	// target_instance is the targeting decision (the target claim with its
	// primary instance, or with which secondary one): D039 changed the coding of 13292.12's primary,
	// not which of its instances 13292.194 and .45 target. D008-D074 made
	// 13292.174, .121 and .9 a modal argument carried by a bridge (scheme,
	// instances), so the claims at their causal premise or exception
	// (13292.12, .233, .43) target that secondary argument (target_instance),
	// at may_bring_about (target_formula, cq: .12, .233); 13292.25 is at
	// 13292.7's too_slight (role, target_part, target_formula). Since step
	// C5.2 the game asks the questions there, so their cq is the game's: .12
	// and .233 CQ:modal_likelihood, .25 CQ:too_slight, and .127 and .43, at
	// the outweighing exceptions, CQ:bad_outweighs and CQ:good_outweighs (cq).
	// D086's round-1 draft (2026-11-27): 13292.141, .182 and .198 name the
	// causal step they answer (target_instance, target_formula, cq; .182 and
	// .198 at its exception, target_part), as 13292.12 does, whose target
	// differed already; 13292.202, a mitigation, is at the conclusion of the
	// argument that cites the harm it mitigates, 13292.127#2 (target_instance,
	// target_part, target_formula); the fragments 13292.6, .11 and .218 leave
	// unstated the variables of the formulas they do not state (bindings,
	// instances); .182 and .198 conclude the exception (instances). Draft 2:
	// 13292.194, an amended action, is at the conclusion of the argument that
	// cites the harm, 13292.12#3 (target_instance; its formula differed
	// already, will_occur in format 1); 13292.192's primary is its second
	// causal step, whose conclusion reaches the target (bindings). Draft 3
	// (D096): the environmental argument is the discussion's, at the thesis
	// (13292.1#2), so the seven claims that answer it target the thesis
	// (target_claim, target_instance; 13292.13 is no longer re-parented and
	// 13292.6 is, re_parented), and 13292.108 is a fragment backing its
	// circumstances (role, target_part, target_instance, target_formula,
	// bindings, instances). D098 (2026-11-28): 13292.45, an amended action, is
	// at 13292.12#3's conclusion, the argument that cites the harm, where it
	// was at the causal step 13292.12#1 (target_instance). Draft 3's part 4
	// (2026-11-28): 13292.8's relation to 13292.115's premise is unsettled
	// (D102), so its role, question and target part, instance and formula are
	// left out (unresolved, over 49); 13292.104#2 is positive consequences
	// (instances). target_instance_v2 (F6 of the external review of exchange
	// 3) as target_instance: a format-1 file names no secondary instance, so
	// the two versions disagree on the same claims.
	want := map[string]float64{"scheme": 0.88, "instances": 0.68, "target_formula": 0.796, "role": 0.959,
		"target_part": 0.898, "cq": 0.837, "target_instance": 0.653, "target_instance_v2": 0.653, "bindings": 0.9, "target_claim": 0.86, "re_parented": 0.96}
	unsettled := map[string]bool{"role": true, "cq": true, "target_part": true, "target_instance": true, "target_instance_v2": true, "target_formula": true}
	for _, f := range Fields {
		fa := p.Fields[f]
		w, ok := want[f]
		if !ok {
			w = 1
		}
		n, u := 50, 0
		if unsettled[f] {
			n, u = 49, 1
		}
		if fa.Agreement != w || fa.Unresolved != u || fa.N != n {
			t.Errorf("%s: %+v, want agreement %v over %d", f, fa, w, n)
		}
	}
	if v := FieldIn(d.Gold["13292.12"], "target_instance", d.Gold); v != "13292.9 secondary:cause_to_effect_modal:may_occur(meat_industry_harmed)" {
		t.Errorf("13292.12's target instance: %q", v)
	}
	// A target at a secondary instance says which.
	a := clone(t, d.Gold)
	act := actualOf(a["13292.12"])
	act["instance"], act["formula"] = "13292.9#3", "causes(meat_industry_harmed,tax_revenue_reduced)"
	if v := FieldIn(a["13292.12"], "target_instance", a); v != "13292.9 secondary:cause_to_effect_modal:may_occur(tax_revenue_reduced)" {
		t.Errorf("a secondary target: %q", v)
	}
	// D098: at 13292.12#3's conclusion; until 2026-11-28 at 13292.12#1's,
	// may_occur(domestic_meat_industry_pushed_out).
	if v := FieldIn(d.Gold["13292.45"], "target_formula", d.Gold); v != "¬should_be_performed(tax_meat)" {
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
	// 13292.12 names is still compared. 13292.8's relation is unsettled in the
	// gold (D102), on both sides: one more unresolved target instance and
	// formula.
	for f, want := range map[string]int{"target_instance": 2, "target_instance_v2": 2, "target_formula": 1, "bindings": 1, "scheme": 0} {
		if fa := p.Fields[f]; fa.Unresolved != want || fa.N != 50-want || fa.Agreement != 1 {
			t.Errorf("%s: %+v, want %d unresolved and full agreement on the rest", f, fa, want)
		}
	}
	q := AgreeAssigned("a", "gold", a, d.Gold, d.IDs())
	if q.AllAssignedAgreement["bindings"] != 49.0/50 {
		t.Errorf("all-assigned agreement counts an unresolved value as agreeing: %v", q.AllAssignedAgreement["bindings"])
	}
	// The target instance's two versions (target_instance_v2 since F6 of the
	// external review of exchange 3) read one instance: both are unresolved.
	if dis := Disagrees([]Annotation{a["13292.12"], d.Gold["13292.12"]}, a, d.Gold); strings.Join(dis, ",") != "target_instance,target_instance_v2" {
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
	if v := FieldIn(old["13292.104"], "target_instance", old); v != "13292.108 primary" {
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
	d, col := builtin(t)
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
	ca := cs.CanonicalSet(a, col)
	for _, f := range Fields {
		if x, y := FieldIn(ca["13292.7"], f, ca), FieldIn(d.Gold["13292.7"], f, d.Gold); x != y {
			t.Errorf("%s: %q, %q", f, x, y)
		}
	}
	// A predicate's name is not an atom: an alias that is also a predicate
	// name leaves the predicate alone.
	cs.canon["brings_about"] = "tax_meat"
	if got := cs.canonFormula("brings_about(brings_about,G)", nil); got != "brings_about(tax_meat,G)" {
		t.Errorf("%q", got)
	}
}

// An alias is read as its action only where an action stands: in the binding
// of a variable in an action's place of the instance's scheme, and in an
// action's place of a formula (positions.json), nested terms included. A goal
// spelled like an alias keeps its name, so an annotation that puts the action
// in the goal's place does not agree with one that names the goal (the
// external audit of 2026-11-25).
func TestCanonicalReadsAliasesOnlyWhereAnActionStands(t *testing.T) {
	d, col := builtin(t)
	cs := *d.Case
	cs.canon = map[string]string{}
	for k, v := range d.Case.canon {
		cs.canon[k] = v
	}
	cs.canon["meat_tax"] = "tax_meat"
	cs.canon["food_less_affordable_for_poor"] = "tax_meat" // 13292.7's consequence
	for f, want := range map[string]string{
		"brings_about(meat_tax,food_less_affordable_for_poor)":     "brings_about(tax_meat,food_less_affordable_for_poor)",
		"bad(food_less_affordable_for_poor)":                       "bad(food_less_affordable_for_poor)",
		"brings_about(meat_tax,meat_tax)":                          "brings_about(tax_meat,meat_tax)",
		"possible_bring_about(meat_tax,meat_tax)":                  "possible_bring_about(meat_tax,tax_meat)",
		"¬should_be_performed(meat_tax)":                           "¬should_be_performed(tax_meat)",
		"¬ ¬should_be_performed( meat_tax )":                       "¬¬should_be_performed(tax_meat)",
		"bad(brings_about(meat_tax,x))":                            "bad(brings_about(tax_meat,x))",
		"brings_about( tax_meat , food_less_affordable_for_poor )": "brings_about( tax_meat , food_less_affordable_for_poor )",
		"brings_about( meat_tax ,food_less_affordable_for_poor)":   "brings_about(tax_meat,food_less_affordable_for_poor)",
		"meat_tax":                        "meat_tax",
		"should_be_performed(meat_tax,x)": "should_be_performed(meat_tax,x)",
		"¬should_be_performed (meat_tax)": "¬should_be_performed(tax_meat)",
		"bad(brings_about(meat_tax,food_less_affordable_for_poor))": "bad(brings_about(tax_meat,food_less_affordable_for_poor))",
	} {
		if got := cs.canonFormula(f, nil); got != want {
			t.Errorf("%q: %q, want %q", f, got, want)
		}
	}
	a := clone(t, d.Gold)
	b := inst(a["13292.7"], 0)["bindings"].(map[string]any)
	b["A"] = "meat_tax"
	binding := func(x Annotation, v string) any {
		return inst(x, 0)["bindings"].(map[string]any)[v]
	}
	got := cs.Canonical(a["13292.7"], col)
	if binding(got, "A") != "tax_meat" || binding(got, "G") != "food_less_affordable_for_poor" {
		t.Errorf("bindings: %v", inst(got, 0)["bindings"])
	}
	if b["A"] != "meat_tax" {
		t.Errorf("Canonical changed its argument: %v", b)
	}
	// Without the collection, or for a scheme it lacks, no variable is known
	// to stand for an action, so the bindings stay as they are.
	if got := cs.Canonical(a["13292.7"], nil); binding(got, "A") != "meat_tax" {
		t.Errorf("no collection: %v", inst(got, 0)["bindings"])
	}
	inst(a["13292.7"], 0)["scheme"].(map[string]any)["walton_id"] = "other"
	if got := cs.Canonical(a["13292.7"], col); binding(got, "A") != "meat_tax" {
		t.Errorf("other: %v", inst(got, 0)["bindings"])
	}
	// The action in the goal's place is not the goal.
	x := clone(t, d.Gold)
	inst(x["13292.7"], 0)["bindings"].(map[string]any)["G"] = "tax_meat"
	cx, cg := cs.CanonicalSet(x, col), cs.CanonicalSet(d.Gold, col)
	if FieldIn(cx["13292.7"], "bindings", cx) == FieldIn(cg["13292.7"], "bindings", cg) {
		t.Error("an action in a goal's place agrees with the goal")
	}
}

// In an instance's formulas, an alias is read as its action wherever the
// instance's scheme has an action variable, also in places positions.json
// does not type (sunk_costs(A,C)'s A, distress's relieve(B,X)), so that the
// binding and the formulas it fills are read alike; outside an instance, in
// a target's formula, positions.json alone types the places (the independent
// check of ID-k, 2026-11-26).
func TestCanonicalReadsTheInstancesScheme(t *testing.T) {
	d, col := builtin(t)
	cs := *d.Case
	cs.canon = map[string]string{}
	for k, v := range d.Case.canon {
		cs.canon[k] = v
	}
	cs.canon["meat_tax"] = "tax_meat"
	f := func(s string) map[string]any { return map[string]any{"formula": s} }
	a := Annotation{
		"instances": []any{
			map[string]any{
				"scheme":     map[string]any{"walton_id": "sunk_costs"},
				"bindings":   map[string]any{"A": "meat_tax", "C": "costs"},
				"premises":   []any{f("sunk_costs(meat_tax,costs)"), f("too_high_to_waste(costs)"), f("feasible(meat_tax)"), f("future_losses_outweigh(meat_tax)")},
				"conclusion": f("should_be_performed(meat_tax)"),
			},
			map[string]any{
				"scheme":     map[string]any{"walton_id": "distress"},
				"bindings":   map[string]any{"B": "governments", "X": "meat_tax", "A": "farmers"},
				"premises":   []any{f("distress(farmers)"), f("relieve(governments,meat_tax)"), f("possible_bring_about(governments,meat_tax)")},
				"conclusion": f("bring_relief(governments,meat_tax)"),
			},
		},
		"target": map[string]any{"actual": map[string]any{"formula": "sunk_costs(meat_tax,costs)"}},
	}
	got := cs.Canonical(a, col)
	b, _ := json.Marshal(got["instances"])
	if s := string(b); strings.Contains(s, "meat_tax") {
		t.Errorf("an alias is left where the instance's scheme has an action: %s", s)
	}
	if f := got["target"].(map[string]any)["actual"].(map[string]any)["formula"]; f != "sunk_costs(meat_tax,costs)" {
		t.Errorf("a target's formula, outside an instance: %v", f)
	}
	// Without the collection, the scheme is not read: positions.json alone.
	b, _ = json.Marshal(cs.Canonical(a, nil)["instances"])
	if s := string(b); !strings.Contains(s, "sunk_costs(meat_tax,costs)") || strings.Contains(s, "should_be_performed(meat_tax)") {
		t.Errorf("without the collection: %s", s)
	}
}

// R7 in Go as validate.py has it (the independent check of ID-k,
// 2026-11-26): each other parent with its relation, in any order, each as
// often as items.json has it; other_kialo_parents recorded exactly when the
// claim has more than one parent; a claim id and a relation are strings.
func TestR7KialoParents(t *testing.T) {
	r7 := func(c *Collection, d *Dataset, set map[string]Annotation, id string) string {
		var out []string
		for _, p := range c.CheckIn(set[id], d, set, 0) {
			if strings.HasPrefix(p, "R7 ") {
				out = append(out, p)
			}
		}
		return strings.Join(out, "; ")
	}
	d, c := builtin(t)
	if p := r7(c, d, d.Gold, "13292.180"); p != "" {
		t.Fatalf("the gold: %s", p)
	}
	others := func(set map[string]Annotation, id string) []any {
		return set[id]["target"].(map[string]any)["other_kialo_parents"].([]any)
	}
	target := func(set map[string]Annotation, id string) map[string]any {
		return set[id]["target"].(map[string]any)
	}
	for name, tc := range map[string]struct {
		claim, want string
		plant       func(set map[string]Annotation)
	}{
		"a parent twice": {"13292.180", "R7 its other Kialo parents are [13292.121 pro, 13292.121 pro]; items.json has [13292.121 pro]", func(set map[string]Annotation) {
			target(set, "13292.180")["other_kialo_parents"] = append(others(set, "13292.180"), others(set, "13292.180")[0])
		}},
		"a number for the parent": {"13292.180", "R7 its Kialo parent is 13292.7 (not a string) pro", func(set map[string]Annotation) {
			target(set, "13292.180")["kialo_parent"] = 13292.7
		}},
		"a number for another parent": {"13292.180", "[13292.121 (not a string) pro]", func(set map[string]Annotation) {
			others(set, "13292.180")[0].(map[string]any)["claim_id"] = 13292.121
		}},
		"no list for two parents": {"13292.180", "R7 it has 2 Kialo parents, and records no list of other_kialo_parents", func(set map[string]Annotation) {
			delete(target(set, "13292.180"), "other_kialo_parents")
		}},
		"a list for one parent": {"13292.7", "R7 it records other_kialo_parents, but has 1 Kialo parent(s)", func(set map[string]Annotation) {
			target(set, "13292.7")["other_kialo_parents"] = []any{}
		}},
		"a list for the thesis": {"13292.1", "R7 it records other_kialo_parents, but has 0 Kialo parent(s)", func(set map[string]Annotation) {
			target(set, "13292.1")["other_kialo_parents"] = nil
		}},
	} {
		t.Run(name, func(t *testing.T) {
			set := clone(t, d.Gold)
			tc.plant(set)
			if p := r7(c, d, set, tc.claim); !strings.Contains(p, tc.want) {
				t.Errorf("%q, want %q", p, tc.want)
			}
		})
	}
	// Any order: a third parent, as another batch's items.json could have.
	d3, _ := builtin(t)
	for i := range d3.Items {
		if d3.Items[i].ID != "13292.180" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(d3.Items[i].Raw, &raw); err != nil {
			t.Fatal(err)
		}
		raw["kialo_parents"] = append(raw["kialo_parents"].([]any), map[string]any{"claim_id": "13292.115", "relation": "con", "location_id": "13292.999"})
		d3.Items[i].Raw, _ = json.Marshal(raw)
	}
	set := clone(t, d3.Gold)
	target(set, "13292.180")["other_kialo_parents"] = append([]any{map[string]any{"claim_id": "13292.115", "relation": "con"}}, others(set, "13292.180")...)
	if p := r7(c, d3, set, "13292.180"); p != "" {
		t.Errorf("two other parents, listed in another order: %s", p)
	}
	others(set, "13292.180")[0].(map[string]any)["relation"] = "pro"
	if p := r7(c, d3, set, "13292.180"); !strings.Contains(p, "[13292.115 pro, 13292.121 pro]; items.json has [13292.115 con, 13292.121 pro]") {
		t.Errorf("a reversed relation among two other parents: %q", p)
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
		"an alias that is an actor":  {`"tax_meat", "actor": "governments", "aliases": []`, `"tax_meat", "actor": "governments", "aliases": ["governments"]`, `"governments" is an actor, so it cannot be an alias of tax_meat`},
		"an action that is an actor": {`"raise_import_standards", "actor": "governments"`, `"governments", "actor": "governments"`, `"governments" is an actor, so it cannot be an action`},
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
	for _, want := range []string{"every instance, by its id", "- 13292.9#2 (cause_to_effect_modal (reconstructed)):", "- 13292.9#1 (negative_consequences_modal (reconstructed)):", "- 13292.9#4 (bridge_from_possible_harm (reconstructed), primary):"} {
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
	// With 13292.8's unsettled relation (D102), two unresolved values.
	q := AgreeAssigned("a", "a again", a, clone(t, a), d.IDs())
	if got := q.AllAssignedAgreement["target_instance"]; got != 48.0/50 {
		t.Errorf("all-assigned agreement on target_instance: %v, want 48/50", got)
	}
	if fa := q.Fields["target_instance"]; fa.Unresolved != 2 || fa.N != 48 || fa.Agreement != 1 {
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
	actualOf(a["13292.7"])["part"] = map[string]any{"unresolved": "premise or exception"}
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
	// Since D102 (2026-11-27) an unresolved formula (13292.45) or part
	// (13292.7) leaves the relation unsettled, so those two claims' role,
	// question and target part, instance and formula are all left out;
	// 13292.33's unresolved instance leaves out its instance only. The gold's
	// own unsettled relation, 13292.8's (D102, draft 3), is one more of each.
	p := Agree("a", "b", a, b)
	if ti, tf := p.Fields["target_instance"], p.Fields["target_formula"]; ti.Unresolved != 4 || tf.Unresolved != 3 || ti.Agreement != 0.978 { // 45 of 46, rounded
		t.Errorf("target_instance %+v, target_formula %+v", ti, tf)
	}
	if tp := p.Fields["target_part"]; tp.Unresolved != 3 {
		t.Errorf("target_part %+v: the unsettled relations are not left out", tp)
	}
	for _, f := range []string{"role", "cq"} {
		if fa := p.Fields[f]; fa.Unresolved != 3 {
			t.Errorf("%s %+v: the unsettled relations' %s are not left out (D102)", f, fa, f)
		}
	}
	// The unflagged instance reads as its claim's primary: format 1 and
	// format 2 name the same target instance everywhere but where
	// D008-D074's recode moved a target to the modal argument a bridge
	// carries (13292.12, .233, .43), and where D086's round-1 draft named the
	// causal step a claim answers (13292.12, .141, .182, .198) or the argument
	// whose conclusion a mitigation amends (13292.202, at 13292.127#2; in
	// draft 2, 13292.194, at 13292.12#3), and where draft 3 moved the claims
	// that answer the environmental argument to the discussion's instance at
	// the thesis, with 13292.108 at its circumstances (D096).
	moved := map[string]bool{"13292.12": true, "13292.233": true, "13292.43": true, "13292.141": true, "13292.182": true, "13292.198": true, "13292.202": true, "13292.194": true,
		"13292.108": true, "13292.5": true, "13292.147": true, "13292.174": true, "13292.13": true, "13292.6": true, "13292.104": true, "13292.127": true,
		// D098: an amended action at the argument that cites the harm.
		"13292.45": true,
		// D102: a background condition, its relation unsettled.
		"13292.8": true}
	for _, id := range d.IDs() {
		x, y := FieldIn(old[id], "target_instance", old), FieldIn(d.Gold[id], "target_instance", d.Gold)
		if (x != y) != moved[id] {
			t.Errorf("format 1, an unflagged lone instance: %s's target instance %q, format 2 %q", id, x, y)
		}
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
	d, col := builtin(t)
	a := clone(t, d.Gold)
	renameIn(inst(a["13292.7"], 0), "tax_meat", "meat_tax")
	// An annotator who leaves unstated a fragment's variable the gold binds
	// (13292.108 binds S1, the circumstances it states; D096).
	inst(a["13292.108"], 0)["bindings"].(map[string]any)["S1"] = map[string]any{"unstated": "the claim states the circumstances only"}
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
	raw, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cj map[string]any
	if err := json.Unmarshal(raw, &cj); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		aliases []any // tax_meat's aliases in the case record; nil: no case
		goal    bool  // 13292.7 puts the action in its consequence's place
		want    float64
	}{
		// The alias changes 13292.7's bindings unless the case reads it as
		// its action.
		{"no case", nil, false, 0.96},
		{"an alias", []any{"meat_tax"}, false, 0.98},
		// An alias spelled like 13292.7's consequence is read as the action
		// only in an action's place: an annotation that puts the action in
		// the consequence's place still differs from the gold.
		{"an alias spelled like a consequence", []any{"meat_tax", "food_less_affordable_for_poor"}, true, 0.96},
	} {
		x := clone(t, a)
		if tc.goal {
			inst(x["13292.7"], 0)["bindings"].(map[string]any)["G"] = "tax_meat"
		}
		var items []Annotation
		for _, id := range d.IDs() {
			items = append(items, x[id])
		}
		pa := write("a.json", map[string]any{"format": Format2, "annotator": "a", "items": items})
		args := []string{"-I", "-B", "agreement.py", pa, "gold.json", "--json", filepath.Join(tmp, "out.json")}
		ga, gb := x, d.Gold
		if tc.aliases != nil {
			cj["actions"].([]any)[0].(map[string]any)["aliases"] = tc.aliases
			pc := write("case.json", cj)
			cs, err := parseCase([]byte(mustJSON(t, cj)))
			if err != nil {
				t.Fatal(err)
			}
			args = append(args, "--case", pc)
			ga, gb = cs.CanonicalSet(x, col), cs.CanonicalSet(d.Gold, col)
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
				t.Errorf("%s, %s: go %+v, agreement.py %+v", tc.name, f, g, w)
			}
		}
		if p.Fields["bindings"].Agreement != tc.want {
			t.Errorf("%s: bindings agreement %v, want %v", tc.name, p.Fields["bindings"].Agreement, tc.want)
		}
	}
}

// agreement.py reads aliases where Go does, on the same annotations: actions'
// and goals' places in formulas (negations, nested terms and spacing
// included), the bindings of action variables by the instance's scheme, a
// scheme that walton.yml lacks or that is not named, a marker, and strings
// that are not formulas.
func TestCanonicalMatchesAgreementPy(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	_, col := builtin(t)
	raw, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cj map[string]any
	if err := json.Unmarshal(raw, &cj); err != nil {
		t.Fatal(err)
	}
	cj["actions"].([]any)[0].(map[string]any)["aliases"] = []any{"meat_tax", "food_less_affordable_for_poor"}
	cs, err := parseCase([]byte(mustJSON(t, cj)))
	if err != nil {
		t.Fatal(err)
	}
	f := func(s string) map[string]any { return map[string]any{"formula": s} }
	set := map[string]Annotation{
		"x": {
			"instances": []any{
				map[string]any{
					"scheme":   map[string]any{"walton_id": "negative_consequences"},
					"bindings": map[string]any{"A": "meat_tax", "G": "food_less_affordable_for_poor"},
					"premises": []any{
						f("brings_about(meat_tax,food_less_affordable_for_poor)"), f("bad(food_less_affordable_for_poor)"),
						f("possible_bring_about(meat_tax, meat_tax)"), f("bad(brings_about(meat_tax,x))"),
						f("¬ ¬should_be_performed( meat_tax )"), f("brings_about( tax_meat , food_less_affordable_for_poor )"),
						f("should_be_performed(meat_tax,x)"), f("meat_tax"), f("brings_about(meat_tax"),
						f("brings_about( meat_tax ,food_less_affordable_for_poor)"),
						f("should_be_performed(\x1cmeat_tax)"), f("brings_about(\x1cmeat_tax,x)"),
					},
					"conclusion": f("¬should_be_performed(meat_tax)"),
					"notes":      "brings_about(meat_tax,x)",
				},
				map[string]any{"scheme": map[string]any{"walton_id": "other"}, "bindings": map[string]any{"A": "meat_tax"}},
				map[string]any{"bindings": map[string]any{"A": "meat_tax"}},
				map[string]any{
					"scheme":   map[string]any{"walton_id": "practical_reasoning2"},
					"bindings": map[string]any{"A": "meat_tax", "G": "meat_tax", "P": "meat_tax", "S1": map[string]any{"unstated": "not said"}},
				},
				map[string]any{
					"scheme":     map[string]any{"walton_id": "sunk_costs"},
					"bindings":   map[string]any{"A": "meat_tax", "C": "food_less_affordable_for_poor"},
					"premises":   []any{f("sunk_costs(meat_tax,food_less_affordable_for_poor)"), f("feasible( meat_tax )"), f("¬should_be_performed (meat_tax)")},
					"conclusion": f("should_be_performed(meat_tax)"),
				},
				map[string]any{
					"scheme":   map[string]any{"walton_id": "distress"},
					"bindings": map[string]any{"B": "governments", "X": "meat_tax", "A": "meat_tax"},
					"premises": []any{f("relieve(governments,meat_tax)"), f("distress(meat_tax)"), f("bad(brings_about(meat_tax,food_less_affordable_for_poor))")},
				},
			},
			"target": map[string]any{"actual": map[string]any{"claim": "meat_tax", "formula": "too_slight(meat_tax,food_less_affordable_for_poor)"}},
		},
	}
	tmp := t.TempDir()
	in, pc := filepath.Join(tmp, "set.json"), filepath.Join(tmp, "case.json")
	for p, v := range map[string]any{in: set, pc: cj} {
		if err := os.WriteFile(p, []byte(mustJSON(t, v)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	code := "import importlib.util, json, sys, yaml\n" +
		"spec = importlib.util.spec_from_file_location('agreement', 'agreement.py')\n" +
		"m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)\n" +
		"schemes = {s['id']: s for s in yaml.safe_load(open(m.WALTON, encoding='utf-8'))['argument_schemes']}\n" +
		"places = json.load(open(m.POSITIONS, encoding='utf-8'))\n" +
		"S, case = json.load(open(sys.argv[1], encoding='utf-8')), json.load(open(sys.argv[2], encoding='utf-8'))\n" +
		"print(json.dumps(m.canonical(S, case, schemes, places)))\n"
	cmd := exec.Command(py, "-I", "-B", "-c", code, in, pc)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	var got, want any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if err := json.Unmarshal([]byte(mustJSON(t, cs.CanonicalSet(set, col))), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("agreement.py:\n%s\nGo:\n%s", mustJSON(t, got), mustJSON(t, want))
	}
	// What Go gives, so that the comparison is not of two copies of the input.
	x := cs.CanonicalSet(set, col)["x"]
	if b := inst(x, 3)["bindings"].(map[string]any); b["A"] != "tax_meat" || b["G"] != "meat_tax" || b["P"] != "meat_tax" {
		t.Errorf("practical_reasoning2's bindings: %v", b)
	}
	if c := inst(x, 0)["conclusion"].(map[string]any)["formula"]; c != "¬should_be_performed(tax_meat)" {
		t.Errorf("the conclusion: %v", c)
	}
}

// agreement.py refuses the case records Go's parseCase refuses for their
// actions and aliases, so that the two measures never read one case two ways
// (the independent check of ID-k, 2026-11-26: agreement.py took them).
func TestAgreementPyRefusesTheCaseGoRefuses(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatal(err)
	}
	for name, plant := range map[string]func(actions []any){
		"an actor as an alias": func(a []any) { a[0].(map[string]any)["aliases"] = []any{"governments"} },
		"an alias of two actions": func(a []any) {
			a[0].(map[string]any)["aliases"] = []any{"meat_tax"}
			a[1].(map[string]any)["aliases"] = []any{"meat_tax"}
		},
		"an action as an alias": func(a []any) { a[0].(map[string]any)["aliases"] = []any{a[1].(map[string]any)["id"]} },
		"an actor as an action": func(a []any) { a[1].(map[string]any)["id"] = "governments" },
	} {
		t.Run(name, func(t *testing.T) {
			var cj map[string]any
			if err := json.Unmarshal(raw, &cj); err != nil {
				t.Fatal(err)
			}
			plant(cj["actions"].([]any))
			if _, err := parseCase([]byte(mustJSON(t, cj))); err == nil {
				t.Fatal("parseCase takes it")
			}
			pc := filepath.Join(t.TempDir(), "case.json")
			if err := os.WriteFile(pc, []byte(mustJSON(t, cj)), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(py, "-I", "-B", "agreement.py", "gold.json", "gold.json", "--case", pc)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "case: ") {
				t.Errorf("agreement.py takes it: %v %s", err, out)
			}
		})
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
// coding: a coding difference in 13292.108 does not spread to the claims
// that target its primary instance (the review of the identity contract,
// W3; four since D096, at the circumstances it states).
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

// target_instance names the target claim with its inference, so two targets
// of different claims, each the primary inference of its claim, disagree.
// The batch's two annotators read 13292.29 so: A at the thesis, B at
// 13292.115's positive consequences (the independent review of 26 November);
// target_claim counted it, target_instance did not.
func TestTargetInstanceComparesTheClaim(t *testing.T) {
	d, c := builtin(t)
	a, b := clone(t, d.Gold), clone(t, d.Gold)
	act := actualOf(b["13292.29"])
	if act["claim"] != "13292.115" || act["instance"] != "13292.115#1" {
		t.Fatalf("13292.29's gold target: %v", act)
	}
	act["claim"], act["instance"], act["part"], act["formula"] = "13292.1", "13292.1#1", "conclusion", "should_be_performed(tax_meat)"
	if probs := c.CheckIn(b["13292.29"], d, b, 2); len(probs) > 0 {
		t.Fatalf("the other reading is not a valid record: %v", probs)
	}
	x, y := FieldIn(a["13292.29"], "target_instance", a), FieldIn(b["13292.29"], "target_instance", b)
	if x != "13292.115 primary" || y != "13292.1 primary" {
		t.Errorf("the two readings: %q, %q", x, y)
	}
	p := Agree("a", "b", a, b)
	// 13292.8's unsettled relation is left out (D102): 48 of 49.
	if fa := p.Fields["target_instance"]; fa.N != 49 || fa.Agreement != 0.98 {
		t.Errorf("target_instance %+v, want 48 of 49", fa)
	}
	// Both versions of the target instance (F6) name the target claim.
	if dis := Disagrees([]Annotation{a["13292.29"], b["13292.29"]}, a, b); strings.Join(dis, ",") != "target_claim,target_instance,target_instance_v2" {
		t.Errorf("the disagreements: %v", dis)
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
	// 13292.6 leaves the action unstated since D086 round 1 (guidelines §5
	// item 6): a conclusion that names it is refused, a null one is taken.
	set = clone(t, d.Gold)
	frag := inst(set["13292.6"], 0)
	if _, unstated := frag["bindings"].(map[string]any)["A"].(map[string]any); !unstated || frag["conclusion"].(map[string]any)["formula"] != nil {
		t.Fatalf("13292.6#1 binds A or concludes: %v", frag)
	}
	frag["conclusion"].(map[string]any)["formula"] = "should_be_performed(tax_meat)"
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

// Go and agreement.py round alike, ties included: 27/48 is 0.563 in both
// (N-a of the fix check).
func TestRoundingMatchesAgreementPy(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("no python3")
	}
	dir, err := filepath.Abs("../../../docs/kialo-batch2")
	if err != nil {
		t.Fatal(err)
	}
	xs := []float64{27.0 / 48, 0.5625, -0.5625, 0.0005, 0.0015, -0.0015, 0.1235, 2.0 / 3, 1.0 / 3, 0.98, 48.0 / 49, -1.0 / 16, 0.12345, 0.9995, 1, 0}
	code := "import importlib.util, json, sys\n" +
		"spec = importlib.util.spec_from_file_location('agreement', 'agreement.py')\n" +
		"m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)\n" +
		"print(json.dumps([m.round3(x) for x in json.loads(sys.argv[1])]))\n"
	cmd := exec.Command(py, "-I", "-B", "-c", code, mustJSON(t, xs))
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	var got []float64
	if err := json.Unmarshal(out, &got); err != nil || len(got) != len(xs) {
		t.Fatalf("%v: %s", err, out)
	}
	for i, x := range xs {
		if got[i] != round3(x) {
			t.Errorf("%v: agreement.py %v, Go %v", x, got[i], round3(x))
		}
	}
	if round3(27.0/48) != 0.563 {
		t.Errorf("27/48 rounds to %v", round3(27.0/48))
	}
}

// Go reads format 2's fragments, retired ids, other parents and new records
// as validate.py does (W-2, W-4 and N-b of the fix check).
func TestCheckInMatchesValidatePy(t *testing.T) {
	d, c := builtin(t)
	cases := []struct {
		name, claim string
		plant       func(s map[string]Annotation)
		want        string // "" for no problem at all
	}{
		{"a retired id of another claim", "13292.7", func(s map[string]Annotation) { s["13292.7"]["retired"] = []any{"13292.9#7"} },
			"R12 its retired ids are not a list of instance ids of its own"},
		// A fully bound fragment: 13292.108 as draft 2 coded it (fullyBound;
		// since D096 the gold has none).
		{"a fully bound fragment without its conclusion", "13292.108", func(s map[string]Annotation) {
			fullyBound(t, s)["conclusion"].(map[string]any)["formula"] = nil
		}, "practical_reasoning2 concludes <nil>, not "},
		{"a fully bound fragment that omits a premise", "13292.108", func(s map[string]Annotation) {
			in := fullyBound(t, s)
			in["premises"] = in["premises"].([]any)[1:]
		}, "practical_reasoning2: missing "},
		{"a fragment that leaves a variable out of its bindings", "13292.108", func(s map[string]Annotation) {
			in := inst(s["13292.108"], 0)
			bm := in["bindings"].(map[string]any)
			s1 := bm["S1"].(string)
			delete(bm, "S1")
			var keep []any
			for _, p := range in["premises"].([]any) {
				if !strings.Contains(p.(map[string]any)["formula"].(string), s1) {
					keep = append(keep, p)
				}
			}
			in["premises"] = keep
		}, "not the scheme's variables"},
		{"a fragment premise that keeps its unstated variable", "13292.108", func(s map[string]Annotation) {
			in := inst(s["13292.108"], 0)
			bm := in["bindings"].(map[string]any)
			s1 := bm["S1"].(string)
			bm["S1"] = map[string]any{"unstated": "not said"}
			for _, p := range in["premises"].([]any) {
				pm := p.(map[string]any)
				pm["formula"] = strings.ReplaceAll(pm["formula"].(string), s1, "S1")
			}
		}, "is not a formula of the scheme with these bindings"},
		{"an other parent's inference at an instance that is not there", "13292.180", func(s map[string]Annotation) {
			o := s["13292.180"]["target"].(map[string]any)["other_kialo_parents"].([]any)[0].(map[string]any)
			o["part"], o["instance"], o["formula"] = "inference", "13292.121#99", nil
		}, "R13 its target in 13292.121 names the instance 13292.121#99, which is not an instance of 13292.121"},
		// D008: a possible consequence concludes no more than that the action
		// may be unwise; the thesis is reached only through its bridge.
		{"a modal argument that concludes the thesis", "13292.174", func(s map[string]Annotation) {
			inst(s["13292.174"], 0)["conclusion"].(map[string]any)["formula"] = "¬should_be_performed(tax_meat)"
		}, "negative_consequences_modal concludes ¬should_be_performed(tax_meat), not may_be_unwise_to_perform(tax_meat,more_dangerous_meat_production)"},
		{"a gold's primary listed second", "13292.9", func(s map[string]Annotation) {
			insts := s["13292.9"]["instances"].([]any)
			insts[0], insts[1] = insts[1], insts[0]
		}, ""},
		{"an annotator's record as the gold codes it", "13292.9", func(s map[string]Annotation) { asAnnotator(s, "13292.9") }, ""},
		{"an annotator's primary listed second", "13292.9", func(s map[string]Annotation) {
			asAnnotator(s, "13292.9")
			insts := s["13292.9"]["instances"].([]any)
			insts[0], insts[1] = insts[1], insts[0]
		}, "R12 its primary instance is not listed first as 13292.9#1"},
		{"an annotator's primary numbered #2", "13292.7", func(s map[string]Annotation) {
			delete(s["13292.7"], "adjudication")
			inst(s["13292.7"], 0)["id"] = "13292.7#2"
		}, "R12 its primary instance is not listed first as 13292.7#1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := clone(t, d.Gold)
			tc.plant(s)
			p := strings.Join(c.CheckIn(s[tc.claim], d, s, 2), "|")
			if tc.want == "" && p != "" || tc.want != "" && !strings.Contains(p, tc.want) {
				t.Errorf("%q, want %q", p, tc.want)
			}
		})
	}
}

// A reviewer's target at a secondary instance is unresolved for agreement:
// it is numbered as one annotator numbers it (W-6 of the fix check).
func TestReviewerSecondaryTargets(t *testing.T) {
	d, _ := builtin(t)
	adj := clone(t, d.Gold)
	act := actualOf(adj["13292.45"])
	act["instance"], act["formula"] = "13292.12#3", "¬should_be_performed(tax_meat)"
	out := reviewerTargets(adj)
	if !isMarker(actualOf(out["13292.45"])["instance"], "unresolved") {
		t.Errorf("a secondary target: %v", actualOf(out["13292.45"]))
	}
	if actualOf(adj["13292.45"])["instance"] != "13292.12#3" {
		t.Error("the adjudication itself was changed")
	}
	if x, y := actualOf(out["13292.7"])["instance"], actualOf(adj["13292.7"])["instance"]; x != y {
		t.Errorf("a primary target: %v, not %v", x, y)
	}
	// The run's report reads them so.
	items := []string{}
	for _, id := range d.IDs() {
		if id == "13292.12" || id == "13292.45" {
			items = append(items, id)
		}
	}
	gold := func(id string) string { b, _ := json.Marshal(d.Gold[id]); return string(b) }
	annotator := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) { return &llm.Response{Text: gold(lastClaim(req))}, nil }}
	reviewer := &llm.Mock{Fn: func(req llm.Request) (*llm.Response, error) {
		id := lastClaim(req)
		b, _ := json.Marshal(adj[id])
		return &llm.Response{Text: `{"kialo_claim_id": "` + id + `", "fields": {}, "adjudicated": ` + string(b) + `, "confidence": 0.8, "needs_human": false, "notes": ""}`}, nil
	}}
	pa := llm.Profile{Name: "one", Provider: "mock", Params: llm.Params{Model: "m1"}}
	pb := llm.Profile{Name: "two", Provider: "mock", Params: llm.Params{Model: "m2"}}
	pr := llm.Profile{Name: "judge", Provider: "mock", Params: llm.Params{Model: "m3"}}
	r, err := New(filepath.Join(t.TempDir(), "run"), d, Settings{Annotators: []llm.Profile{pa, pb}, Reviewer: &pr, Items: items},
		map[string]*llm.Client{"one": llm.WithProvider(pa, annotator), "two": llm.WithProvider(pb, annotator), "judge": llm.WithProvider(pr, reviewer)})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := r.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	rep, err := r.Report()
	if err != nil {
		t.Fatal(err)
	}
	if rep.Reviewer == nil || rep.Reviewer.VsGold == nil {
		t.Fatalf("no reviewer scores: %+v", rep.Reviewer)
	}
	// 13292.45's secondary target is unresolved; so is 13292.12's, whose
	// target claim the reviewer did not adjudicate (§11). Read as it is, the
	// secondary target would be scored, against the gold's primary one.
	if ti := rep.Reviewer.VsGold.Fields["target_instance"]; ti.Unresolved != 2 || ti.N != 0 {
		t.Errorf("the reviewer's target_instance: %+v", ti)
	}
}

// R18 (guidelines §5 item 6): fit explicit when every premise-role formula is
// explicit; an assumption or exception may be supplied. From the reviewer's
// patch of 2026-11-27 (D086 exchange 1), which found 13292.104#2 explicit with a
// supplied goal premise.
func TestR18ExplicitFit(t *testing.T) {
	d, c := builtin(t)
	for _, tc := range []struct {
		name, fit, ordinary, other string
		want                       bool
	}{
		{"a supplied premise", "explicit", "supplied", "supplied", true},
		{"an inferred premise", "explicit", "inferred", "supplied", true},
		{"a reconstruction may supply premises", "reconstructed", "supplied", "supplied", false},
		{"assumptions and exceptions may be supplied", "explicit", "explicit", "supplied", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			set := clone(t, d.Gold)
			a := set["13292.104"]
			x := inst(a, 1)
			x["scheme"].(map[string]any)["fit"] = tc.fit
			for _, y := range x["premises"].([]any) {
				pm := y.(map[string]any)
				if pm["role"] == "premise" {
					pm["status"] = tc.ordinary
				} else {
					pm["status"] = tc.other
				}
			}
			out := c.CheckIn(a, d, set, 0)
			got := false
			for _, p := range out {
				if strings.HasPrefix(p, "R18 ") {
					got = true
				}
			}
			if got != tc.want {
				t.Fatalf("R18: got %v, want %v: %v", got, tc.want, out)
			}
		})
	}
}

// fullyBound gives 13292.108 the coding it had until D096 (round 1, draft
// 3): the environmental argument, every variable bound, as a fragment, its
// circumstances explicit and the rest supplied. It is the discussion's
// instance at the thesis (13292.1#2) since, and the gold has no fully bound
// fragment left; tests of one build it so. It returns the instance.
func fullyBound(t *testing.T, s map[string]Annotation) map[string]any {
	t.Helper()
	var env map[string]any
	for _, x := range s["13292.1"]["instances"].([]any) {
		if in, _ := x.(map[string]any); in["by"] == "the discussion" {
			env = in
		}
	}
	if env == nil {
		t.Fatal("the thesis has no instance by the discussion")
	}
	var in map[string]any
	b, _ := json.Marshal(env)
	if err := json.Unmarshal(b, &in); err != nil {
		t.Fatal(err)
	}
	delete(in, "by")
	delete(in, "label")
	in["id"], in["primary"] = "13292.108#1", true
	in["scheme"].(map[string]any)["fit"] = "fragment"
	for _, y := range in["premises"].([]any) {
		pm := y.(map[string]any)
		if pm["status"] == "inferred" {
			if pm["cite"] == "13292.108" {
				pm["status"] = "explicit"
			} else {
				pm["status"] = "supplied"
			}
			delete(pm, "cite")
		}
	}
	s["13292.108"]["instances"].([]any)[0] = in
	return in
}
