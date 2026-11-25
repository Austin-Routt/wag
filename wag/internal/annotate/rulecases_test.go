package annotate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/yourorg/wag/data"
)

// plain turns what yaml.v2 decodes (map[interface{}]interface{}) into the
// JSON-shaped values an Annotation holds.
func plain(v any) any {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]any{}
		for k, y := range x {
			m[fmt.Sprint(k)] = plain(y)
		}
		return m
	case map[string]interface{}:
		m := map[string]any{}
		for k, y := range x {
			m[k] = plain(y)
		}
		return m
	case []interface{}:
		out := make([]any, len(x))
		for i, y := range x {
			out[i] = plain(y)
		}
		return out
	}
	return v
}

// shape is what a case and the gold must agree on: each instance's scheme,
// primacy, bindings, formulas and conclusion; the role; the target.
func shape(a Annotation) string {
	get := func(m any, k string) any {
		if mm, ok := m.(map[string]any); ok {
			return mm[k]
		}
		return nil
	}
	var b strings.Builder
	insts, _ := a["instances"].([]any)
	for _, x := range insts {
		var binds, forms []string
		if bm, ok := get(x, "bindings").(map[string]any); ok {
			for k, v := range bm {
				binds = append(binds, k+"="+fmt.Sprint(v))
			}
		}
		ps, _ := get(x, "premises").([]any)
		for _, p := range ps {
			forms = append(forms, fmt.Sprint(get(p, "role"))+" "+compact(fmt.Sprint(get(p, "formula"))))
		}
		sort.Strings(binds)
		sort.Strings(forms)
		fmt.Fprintf(&b, "instance primary=%v scheme=%v bindings=%v formulas=%v conclusion=%s\n", get(x, "primary"),
			get(get(x, "scheme"), "walton_id"), binds, forms, compact(fmt.Sprint(get(get(x, "conclusion"), "formula"))))
	}
	r := a["role"]
	fmt.Fprintf(&b, "role kind=%v cq=%v polarity=%v\n", get(r, "kind"), get(r, "cq"), get(r, "polarity"))
	// The formula of a premise or an exception target; a conclusion target's
	// is its instance's conclusion, which format 2 names and a case's coding
	// need not. Instance ids, which a case's coding has not, are not compared.
	t := get(a["target"], "actual")
	f := ""
	if p := get(t, "part"); p == "premise" || p == "exception" {
		f = compact(fmt.Sprint(get(t, "formula")))
	}
	fmt.Fprintf(&b, "target claim=%v part=%v formula=%s\n", get(t, "claim"), get(t, "part"), f)
	return b.String()
}

// flat is a text with its runs of white space made one space, YAML's
// doubled quote undone and its case folded, so that a phrase is found
// whatever the line breaks.
func flat(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.Join(strings.Fields(s), " "), "''", "'"))
}

// The worked cases of the guidelines' rules on opposite consequences and
// weighing (docs/kialo-batch2/rule-cases.yml, guidelines §3.1): every rule
// has a case and every case names rules that exist; every coding follows
// walton.yml, and every formula that uses one of its predicates (an other
// instance's too) gives it as many arguments as walton.yml does; a case that
// is a claim of the batch is coded as the gold codes it; and no retired
// wording of a rule is left in the guidelines, the cases, the catalogue's
// entries or the candidates registry.
func TestRuleCases(t *testing.T) {
	const docs = "../../../docs/"
	raw, err := os.ReadFile(docs + "kialo-batch2/rule-cases.yml")
	if err != nil {
		t.Skip("the docs are not here:", err)
	}
	var f struct {
		Retired []struct {
			Phrase string `yaml:"phrase"`
			Now    string `yaml:"now"`
			Why    string `yaml:"why"`
		} `yaml:"retired"`
		Cases []struct {
			ID    string   `yaml:"id"`
			Rules []string `yaml:"rules"`
			Kialo string   `yaml:"kialo"`
			// OutsideBatch marks a claim of the batch's discussion that is
			// not one of its 50, so that a mistyped id is not taken for one.
			OutsideBatch bool           `yaml:"outside_batch"`
			Claim        string         `yaml:"claim"`
			Target       string         `yaml:"target"`
			Coding       map[string]any `yaml:"coding"`
			Why          string         `yaml:"why"`
		} `yaml:"cases"`
	}
	if err := yaml.UnmarshalStrict(raw, &f); err != nil {
		t.Fatal(err)
	}
	guide, err := os.ReadFile(docs + "kialo-batch2/guidelines.md")
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string]int{}
	for _, m := range regexp.MustCompile(`\*\*(OC\d+)\.`).FindAllStringSubmatch(string(guide), -1) {
		rules[m[1]] = 0
	}
	if len(rules) == 0 {
		t.Fatal("no rules OC1, OC2, ... in guidelines.md")
	}
	w, err := data.Annotation.ReadFile("annotation/walton.yml")
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadCollection(w)
	if err != nil {
		t.Fatal(err)
	}
	ar, err := arities(w)
	if err != nil {
		t.Fatal(err)
	}
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	gold := ds[0].Gold
	discussion := strings.SplitN(ds[0].IDs()[0], ".", 2)[0] + "."
	seen := map[string]bool{}
	for _, cs := range f.Cases {
		if cs.ID == "" || seen[cs.ID] {
			t.Errorf("case %q: no id, or one already used", cs.ID)
		}
		seen[cs.ID] = true
		if len(cs.Rules) == 0 || cs.Claim == "" || cs.Why == "" || cs.Coding == nil {
			t.Errorf("%s: a case needs rules, a claim, a coding and why", cs.ID)
		}
		for _, r := range cs.Rules {
			if _, ok := rules[r]; !ok {
				t.Errorf("%s: %s is not a rule of guidelines §3.1", cs.ID, r)
				continue
			}
			rules[r]++
		}
		a, _ := plain(cs.Coding).(map[string]any)
		for _, p := range c.Check(a, nil) {
			t.Errorf("%s: %s", cs.ID, p)
		}
		// The check reads no formula of other, so a case's other instance
		// concluding an exception is held to walton.yml's arities here.
		for _, p := range arityFaults(a, ar) {
			t.Errorf("%s: %s", cs.ID, p)
		}
		_, inGold := gold[cs.Kialo]
		switch {
		case inGold && cs.OutsideBatch:
			t.Errorf("%s: %s is a claim of the batch, not outside it", cs.ID, cs.Kialo)
		case !inGold && !cs.OutsideBatch && strings.HasPrefix(cs.Kialo, discussion):
			t.Errorf("%s: %s is of the batch's discussion but not in the gold: mark it outside_batch, or fix the id", cs.ID, cs.Kialo)
		}
		if g, ok := gold[cs.Kialo]; ok {
			if got, want := shape(a), shape(g); got != want {
				t.Errorf("%s: the case is not coded as gold.json codes %s:\ncase:\n%sgold:\n%s", cs.ID, cs.Kialo, got, want)
			}
		}
	}
	for r, n := range rules {
		if n == 0 {
			t.Errorf("rule %s has no worked case", r)
		}
	}
	// Retired wording: the cases file is searched without its own list.
	texts := map[string]string{
		"guidelines.md":  string(guide),
		"rule-cases.yml": string(raw[strings.Index(string(raw), "\ncases:"):]),
		"candidates.yml": "",
	}
	if b, err := os.ReadFile(docs + "candidates.yml"); err == nil {
		texts["candidates.yml"] = string(b)
	}
	entries, _ := filepath.Glob("../../data/catalogue/schemes/*.yml")
	if len(entries) == 0 {
		t.Fatal("no catalogue entries")
	}
	for _, p := range entries {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		texts[filepath.Base(p)] = string(b)
	}
	for _, r := range f.Retired {
		if r.Phrase == "" || r.Now == "" || r.Why == "" {
			t.Errorf("retired %q: it needs a phrase, what says it now, and why", r.Phrase)
		}
		for name, s := range texts {
			if strings.Contains(flat(s), flat(r.Phrase)) {
				t.Errorf("%s still says %q, a retired wording: now %s", name, r.Phrase, r.Now)
			}
		}
	}
}

// arities are walton.yml's predicates, from its language templates and its
// schemes' formulas, with the numbers of arguments each takes.
func arities(w []byte) (map[string]map[int]bool, error) {
	var f struct {
		Schemes  []waltonScheme    `yaml:"argument_schemes"`
		Language map[string]string `yaml:"language"`
	}
	if err := yaml.Unmarshal(w, &f); err != nil {
		return nil, err
	}
	out := map[string]map[int]bool{}
	put := func(p string, n int) {
		if out[p] == nil {
			out[p] = map[int]bool{}
		}
		out[p][n] = true
	}
	for k := range f.Language {
		if i := strings.LastIndex(k, "/"); i > 0 {
			if n, err := strconv.Atoi(k[i+1:]); err == nil {
				put(k[:i], n)
			}
		}
	}
	for _, s := range f.Schemes {
		for _, fs := range [][]string{s.Premises, s.Assumptions, s.Exceptions, s.Conclusions} {
			for _, x := range fs {
				for _, t := range terms(x, true) {
					put(t.name, t.n)
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("walton.yml has no predicates")
	}
	return out, nil
}

type term struct {
	name string
	n    int
}

// terms are the compound terms of a formula, nested ones included, with
// their numbers of arguments; at the top level (top) a bare atom is a
// proposition and is listed with none. A leading negation is dropped.
func terms(f string, top bool) []term {
	f = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(f), "¬"))
	i := strings.IndexByte(f, '(')
	if i < 0 || !strings.HasSuffix(f, ")") {
		if top && f != "" {
			return []term{{f, 0}}
		}
		return nil
	}
	var args []string
	depth, from := 0, i+1
	for k := i + 1; k < len(f)-1; k++ {
		switch f[k] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				args = append(args, f[from:k])
				from = k + 1
			}
		}
	}
	args = append(args, f[from:len(f)-1])
	out := []term{{strings.TrimSpace(f[:i]), len(args)}}
	for _, a := range args {
		out = append(out, terms(a, false)...)
	}
	return out
}

// arityFaults lists the formulas of a coding, its other instances'
// included, that use a walton.yml predicate with a number of arguments it
// does not take.
func arityFaults(a map[string]any, ar map[string]map[int]bool) []string {
	var out []string
	check := func(where string, f any) {
		s, ok := f.(string)
		if !ok {
			return
		}
		for _, t := range terms(s, true) {
			if ns, known := ar[t.name]; known && !ns[t.n] {
				out = append(out, fmt.Sprintf("%s %s uses %s with %d arguments, which walton.yml does not", where, s, t.name, t.n))
			}
		}
	}
	insts, _ := a["instances"].([]any)
	for _, x := range insts {
		in, _ := x.(map[string]any)
		ps, _ := in["premises"].([]any)
		for _, y := range ps {
			pm, _ := y.(map[string]any)
			check(fmt.Sprint(pm["role"]), pm["formula"])
		}
		if cm, ok := in["conclusion"].(map[string]any); ok {
			check("conclusion", cm["formula"])
		}
	}
	return out
}

// The arity check's own test: its counts are checked against formulas whose
// arities are known, and a fixture walton.yml whose arities come from its
// language keys, so that a check that miscounted (and counted walton.yml's
// formulas the same wrong way) would not pass unseen.
func TestRuleCasesArityCheck(t *testing.T) {
	for f, want := range map[string]string{
		"may_bring_about(tax,g)":           "may_bring_about/2",
		"¬should_be_performed(tax)":        "should_be_performed/1",
		"good_outweighs(a, f(b,c))":        "good_outweighs/2 f/2",
		"should_irrelevant_to_can":         "should_irrelevant_to_can/0",
		"too_slight(tax,meat_price_rise)":  "too_slight/2",
		"fails_to_avert(levy, g(h(x)), y)": "fails_to_avert/3 g/1 h/1",
	} {
		var got []string
		for _, x := range terms(f, true) {
			got = append(got, fmt.Sprintf("%s/%d", x.name, x.n))
		}
		if strings.Join(got, " ") != want {
			t.Errorf("terms(%q) = %v, want %s", f, got, want)
		}
	}
	w := []byte("language:\n  may_bring_about/2: \"The action %s may bring about %s.\"\n" +
		"  too_slight/2: \"In deciding on action %s, %s is too slight a consequence to count.\"\n" +
		"argument_schemes:\n  - id: s\n    variables: [A,G]\n    conclusions:\n      - may_be_worth_performing(A,G)\n" +
		"    premises:\n      - may_bring_about(A,G)\n")
	ar, err := arities(w)
	if err != nil {
		t.Fatal(err)
	}
	if !ar["may_bring_about"][2] || len(ar["may_bring_about"]) != 1 || !ar["may_be_worth_performing"][2] || !ar["too_slight"][2] {
		t.Fatalf("the fixture's arities: %v", ar)
	}
	coding := func(premise, other string) map[string]any {
		return map[string]any{"instances": []any{
			map[string]any{"premises": []any{map[string]any{"role": "premise", "formula": premise}},
				"conclusion": map[string]any{"formula": "may_be_worth_performing(tax,g)"}},
			map[string]any{"premises": []any{map[string]any{"role": "premise", "formula": "other_reason"}},
				"conclusion": map[string]any{"formula": other}},
		}}
	}
	if f := arityFaults(coding("may_bring_about(tax,g)", "¬may_bring_about(tax,g)"), ar); len(f) != 0 {
		t.Errorf("a sound coding has faults: %v", f)
	}
	// too_slight is only in the fixture's language: its arity comes from there.
	for _, bad := range []string{"may_bring_about(tax,g,h)", "¬may_bring_about(tax)", "may_bring_about", "outer(may_be_worth_performing(tax))", "too_slight(tax)"} {
		if f := arityFaults(coding("may_bring_about(tax,g)", bad), ar); len(f) != 1 {
			t.Errorf("%s in another instance's conclusion: %d faults (%v), want 1", bad, len(f), f)
		}
		if f := arityFaults(coding(bad, "may_be_worth_performing(tax,g)"), ar); len(f) != 1 {
			t.Errorf("%s in a premise: %d faults (%v), want 1", bad, len(f), f)
		}
	}
}
