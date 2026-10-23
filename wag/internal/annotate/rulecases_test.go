package annotate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	t := get(a["target"], "actual")
	fmt.Fprintf(&b, "target claim=%v part=%v formula=%s\n", get(t, "claim"), get(t, "part"), compact(fmt.Sprint(get(t, "formula"))))
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
// walton.yml; a case that is a claim of the batch is coded as the gold codes
// it; and no retired wording of a rule is left in the guidelines, the cases,
// the catalogue's entries or the candidates registry.
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
			ID     string         `yaml:"id"`
			Rules  []string       `yaml:"rules"`
			Kialo  string         `yaml:"kialo"`
			Claim  string         `yaml:"claim"`
			Target string         `yaml:"target"`
			Coding map[string]any `yaml:"coding"`
			Why    string         `yaml:"why"`
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
	ds, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	gold := ds[0].Gold
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
