package annotate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Fields are the decisions the guidelines keep apart (§2), compared field
// by field.
var Fields = []string{"scheme", "fit", "role", "polarity", "target_claim", "target_part", "re_parented", "evidence"}

// FieldOf returns an annotation's value for a field, as a string ("" if it
// has none).
func FieldOf(a Annotation, field string) string {
	str := func(v any) string {
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
	get := func(m any, k string) any {
		if mm, ok := m.(map[string]any); ok {
			return mm[k]
		}
		return nil
	}
	switch field {
	case "scheme":
		return str(get(primary(a)["scheme"], "walton_id"))
	case "fit":
		return str(get(primary(a)["scheme"], "fit"))
	case "role":
		return str(get(a["role"], "kind"))
	case "polarity":
		return str(get(a["role"], "polarity"))
	case "target_claim":
		return str(get(get(a["target"], "actual"), "claim"))
	case "target_part":
		p := str(get(get(a["target"], "actual"), "part"))
		if i := strings.IndexAny(p, ":("); i > 0 { // "premise: f(x)" → premise
			p = strings.TrimSpace(p[:i])
		}
		return p
	case "re_parented":
		return str(get(a["target"], "re_parented"))
	case "evidence":
		return str(get(a["evidence"], "status"))
	}
	return ""
}

// FieldAgreement is the agreement of two annotators on one field.
type FieldAgreement struct {
	N         int      `json:"n"`
	Agreement float64  `json:"agreement"`
	Kappa     *float64 `json:"kappa"` // Cohen's; nil when undefined (one category only)
}

// Pair is the agreement of two annotators (or an annotator and the gold).
type Pair struct {
	A      string                    `json:"a"`
	B      string                    `json:"b"`
	Items  int                       `json:"items"` // annotated by both
	Fields map[string]FieldAgreement `json:"fields"`
}

// Agree compares two sets of annotations field by field, over the items both
// have.
func Agree(nameA, nameB string, a, b map[string]Annotation) Pair {
	p := Pair{A: nameA, B: nameB, Fields: map[string]FieldAgreement{}}
	var ids []string
	for id := range a {
		if b[id] != nil {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	p.Items = len(ids)
	for _, f := range Fields {
		var xs, ys []string
		for _, id := range ids {
			xs, ys = append(xs, FieldOf(a[id], f)), append(ys, FieldOf(b[id], f))
		}
		p.Fields[f] = kappa(xs, ys)
	}
	return p
}

// kappa is raw agreement and Cohen's kappa of two label lists.
func kappa(xs, ys []string) FieldAgreement {
	n := len(xs)
	out := FieldAgreement{N: n}
	if n == 0 {
		return out
	}
	same := 0
	ca, cb := map[string]int{}, map[string]int{}
	for i := range xs {
		if xs[i] == ys[i] {
			same++
		}
		ca[xs[i]]++
		cb[ys[i]]++
	}
	po := float64(same) / float64(n)
	pe := 0.0
	for k, v := range ca {
		pe += float64(v) / float64(n) * float64(cb[k]) / float64(n)
	}
	out.Agreement = round3(po)
	if pe < 1 {
		k := round3((po - pe) / (1 - pe))
		out.Kappa = &k
	}
	return out
}

func round3(x float64) float64 { return float64(int(x*1000+0.5*sign(x))) / 1000 }

func sign(x float64) float64 {
	if x < 0 {
		return -1
	}
	return 1
}

// Disagrees lists the fields on which the annotations of one claim differ.
func Disagrees(anns []Annotation) []string {
	var out []string
	for _, f := range Fields {
		seen := map[string]bool{}
		for _, a := range anns {
			seen[FieldOf(a, f)] = true
		}
		if len(seen) > 1 {
			out = append(out, f)
		}
	}
	return out
}

// Collection is the scheme collection annotations are checked against
// (walton.yml).
type Collection struct {
	schemes map[string]waltonScheme
}

type waltonScheme struct {
	ID          string   `yaml:"id"`
	Variables   []string `yaml:"variables"`
	Premises    []string `yaml:"premises"`
	Assumptions []string `yaml:"assumptions"`
	Exceptions  []string `yaml:"exceptions"`
	Conclusions []string `yaml:"conclusions"`
}

// LoadCollection reads walton.yml.
func LoadCollection(b []byte) (*Collection, error) {
	var f struct {
		Schemes []waltonScheme `yaml:"argument_schemes"`
	}
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	c := &Collection{schemes: map[string]waltonScheme{}}
	for _, s := range f.Schemes {
		c.schemes[s.ID] = s
	}
	if len(c.schemes) == 0 {
		return nil, fmt.Errorf("no argument_schemes")
	}
	return c, nil
}

var variable = regexp.MustCompile(`\b[A-Z][A-Z0-9]*\b`)

func subst(formula string, b map[string]string) string {
	return variable.ReplaceAllStringFunc(formula, func(v string) string {
		if x, ok := b[v]; ok {
			return x
		}
		return v
	})
}

var (
	fits  = map[string]bool{"explicit": true, "reconstructed": true, "fragment": true, "none": true}
	roles = map[string]bool{"independent_pro": true, "independent_con": true, "rebuttal": true, "undercutter": true, "premise_attack": true,
		"premise_backing": true, "cq_answer": true, "side_effect": true, "alternative_means": true, "action_amendment": true, "clarification": true, "none": true}
	evidence = map[string]bool{"source": true, "definition": true, "example": true, "assertion": true}
)

// Check lists what is wrong with an annotation of a claim of the dataset:
// its shape; its scheme, which must be one of the collection's (or other,
// none); bindings for exactly the scheme's variables; premises,
// assumptions, exceptions and conclusion that are the scheme's formulas
// with the bindings put in; a target that is a claim of the dataset. It is
// the core of docs/kialo-batch2/validate.py (which also checks the
// lexicon and the adjudication rules).
func (c *Collection) Check(a Annotation, d *Dataset) []string {
	var out []string
	add := func(f string, x ...any) { out = append(out, fmt.Sprintf(f, x...)) }
	insts, _ := a["instances"].([]any)
	if len(insts) == 0 {
		return []string{"no instances"}
	}
	nprim := 0
	for _, x := range insts {
		in, _ := x.(map[string]any)
		if in == nil {
			add("an instance is not an object")
			continue
		}
		if p, _ := in["primary"].(bool); p {
			nprim++
		}
		s, _ := in["scheme"].(map[string]any)
		wid, _ := s["walton_id"].(string)
		fit, _ := s["fit"].(string)
		if !fits[fit] {
			add("fit %q is not explicit, reconstructed, fragment or none", fit)
		}
		binds := map[string]string{}
		if bm, ok := in["bindings"].(map[string]any); ok {
			for k, v := range bm {
				if v == nil {
					continue // a fragment's unstated variables are null
				}
				binds[k] = fmt.Sprint(v)
			}
		}
		if wid == "other" || wid == "none" || wid == "" {
			if wid == "" {
				add("an instance has no walton_id")
			}
			continue
		}
		sc, ok := c.schemes[wid]
		if !ok {
			add("%s is not a scheme of walton.yml", wid)
			continue
		}
		if fit != "fragment" {
			want := append([]string(nil), sc.Variables...)
			var got []string
			for k := range binds {
				got = append(got, k)
			}
			sort.Strings(want)
			sort.Strings(got)
			if strings.Join(want, ",") != strings.Join(got, ",") {
				add("%s binds %v, not the scheme's variables %v", wid, got, want)
			}
		}
		expected := map[string]int{}
		for role, fs := range map[string][]string{"premise": sc.Premises, "assumption": sc.Assumptions, "exception": sc.Exceptions} {
			for _, f := range fs {
				expected[role+" "+subst(f, binds)]++
			}
		}
		ps, _ := in["premises"].([]any)
		for _, y := range ps {
			pm, _ := y.(map[string]any)
			k := fmt.Sprintf("%v %v", pm["role"], pm["formula"])
			if expected[k] > 0 {
				expected[k]--
			} else if fit != "fragment" {
				add("%s: %s is not a formula of the scheme with these bindings", wid, k)
			}
		}
		if fit != "fragment" {
			for k, n := range expected {
				if n > 0 {
					add("%s: missing %s", wid, k)
				}
			}
		}
		if len(sc.Conclusions) > 0 {
			cm, _ := in["conclusion"].(map[string]any)
			if got, want := fmt.Sprint(cm["formula"]), subst(sc.Conclusions[0], binds); fit != "fragment" && got != want {
				add("%s concludes %s, not %s", wid, got, want)
			}
		}
	}
	if nprim != 1 && len(insts) > 1 {
		add("%d primary instances, not 1", nprim)
	}
	r, _ := a["role"].(map[string]any)
	if k, _ := r["kind"].(string); !roles[k] {
		add("role %q is not one of the guidelines'", k)
	}
	if p := r["polarity"]; p != nil && p != "for" && p != "against" {
		add("polarity %v is not for or against", p)
	}
	if ev, _ := a["evidence"].(map[string]any); ev != nil {
		if s, _ := ev["status"].(string); !evidence[s] {
			add("evidence %q is not source, definition, example or assertion", s)
		}
	}
	if t := FieldOf(a, "target_claim"); t != "" && d != nil && d.Item(t) == nil {
		add("its target %s is not a claim of the dataset", t)
	}
	return out
}
