package annotate

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Fields are the decisions the guidelines keep apart (§2), compared field
// by field: the eight of format 1, then, since the identity contract
// (docs/annotation-records.md §11, D092), the question answered, the exact
// target (the instance it names, read in the annotator's own file, and the
// formula), the primary instance's bindings, and every instance.
var Fields = []string{"scheme", "fit", "role", "polarity", "target_claim", "target_part", "re_parented", "evidence",
	"cq", "target_instance", "target_formula", "bindings", "instances"}

// Unresolved is the value of a field that cannot be compared: an annotator's
// {"unresolved": why}, or a target its own file does not hold. It is left
// out of the field's agreement and counted apart, neither agreement nor
// disagreement. An item whose target formula or part is unresolved has no
// settled relation to its target (D102): its role, its question and its
// target's part, instance and formula are Unresolved.
const Unresolved = "?"

// FieldOf returns an annotation's value for a field, as a string ("" if it
// has none). It reads the target instance without the annotator's other
// annotations, so target_instance (and a format-1 conclusion target's
// formula) is Unresolved: FieldIn reads them.
func FieldOf(a Annotation, field string) string { return FieldIn(a, field, nil) }

// FieldIn is FieldOf with the annotator's own annotations (set), in which a
// target's instance is read: ids are the annotator's own, so the instance
// is compared by what it names, its scheme and its conclusion.
func FieldIn(a Annotation, field string, set map[string]Annotation) string {
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
	actual, _ := get(a["target"], "actual").(map[string]any)
	// D102: a target whose formula or part is unresolved leaves the item's
	// relation unsettled, so its role, question and target part, instance
	// and formula are not compared (agreement.py's RELATION_FIELDS). An
	// unresolved instance alone leaves only target_instance out, below.
	if relationUnsettled(actual) {
		switch field {
		case "role", "cq", "target_part", "target_instance", "target_formula":
			return Unresolved
		}
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
		pv := get(get(a["target"], "actual"), "part")
		if isMarker(pv, "unresolved") {
			return Unresolved
		}
		p := str(pv)
		if i := strings.IndexAny(p, ":("); i > 0 { // "premise: f(x)" → premise
			p = strings.TrimSpace(p[:i])
		}
		return p
	case "re_parented":
		// Derived, as the guidelines say the tooling computes it (§5, item
		// 1): the actual target claim is not the Kialo parent. The
		// annotation's own value is not read; agreement.py's re_parented
		// derives it alike, reading only text as a claim.
		t, _ := a["target"].(map[string]any)
		if len(actual) == 0 {
			return "false"
		}
		claim, _ := actual["claim"].(string)
		parent, _ := t["kialo_parent"].(string)
		return strconv.FormatBool(claim != parent)
	case "evidence":
		return str(get(a["evidence"], "status"))
	case "cq":
		cq, _ := get(a["role"], "cq").(string)
		return cq
	case "bindings":
		bm, _ := primary(a)["bindings"].(map[string]any)
		keys := make([]string, 0, len(bm))
		for k := range bm {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			v := bm[k]
			switch {
			case isMarker(v, "unstated"):
				parts[i] = k + "=(unstated)"
			case isMarker(v, "unresolved"):
				return Unresolved
			default:
				parts[i] = k + "=" + str(v)
			}
		}
		return strings.Join(parts, ";")
	case "instances":
		insts, _ := a["instances"].([]any)
		sigs := make([]string, 0, len(insts))
		for _, x := range insts {
			in, _ := x.(map[string]any)
			sigs = append(sigs, signature(in))
		}
		sort.Strings(sigs)
		return strings.Join(sigs, "|")
	case "target_instance", "target_formula":
		if len(actual) == 0 {
			return ""
		}
		ref, f := actual["instance"], actual["formula"]
		claim, _ := actual["claim"].(string)
		if field == "target_formula" {
			// Only what the formula itself leaves open makes it unresolved
			// (docs/annotation-records.md §11).
			if isMarker(f, "unresolved") {
				return Unresolved
			}
			if s, ok := f.(string); ok && s != "" {
				return compact(s)
			}
			switch str(actual["part"]) {
			case "inference":
				return ""
			case "conclusion": // a format-1 conclusion target names no formula: its instance's conclusion
				if in := targetInstance(set, claim, ref); in != nil && !isMarker(ref, "unresolved") {
					cm, _ := in["conclusion"].(map[string]any)
					if s, ok := cm["formula"].(string); ok {
						return compact(s)
					}
				}
			}
			return Unresolved
		}
		// Which inference of which claim: the target claim with its primary
		// inference, or with which secondary one, so that two targets of
		// different claims disagree. The targeting decision, not the target's
		// own coding: how the target claim is coded is its own fields'
		// business.
		if isMarker(ref, "unresolved") {
			return Unresolved
		}
		in := targetInstance(set, claim, ref)
		if in == nil {
			return Unresolved
		}
		// A format-1 target names no instance (or names it in words that
		// begin "primary"): it is its claim's primary instance.
		if id, ok := ref.(string); ref == nil || ok && (id == "" || strings.HasPrefix(id, "primary")) {
			return claim + " primary"
		}
		if p, _ := in["primary"].(bool); p {
			return claim + " primary"
		}
		return claim + " secondary:" + signature(in)
	}
	return ""
}

// signature is what an instance names: its scheme and its conclusion.
func signature(in map[string]any) string {
	s, _ := in["scheme"].(map[string]any)
	cm, _ := in["conclusion"].(map[string]any)
	concl := ""
	if f, ok := cm["formula"].(string); ok {
		concl = compact(f)
	}
	wid, _ := s["walton_id"].(string)
	return wid + ":" + concl
}

// targetInstance is the instance a target names, in the annotator's own
// annotation of its claim: the instance with that id (format 2), or the
// claim's primary instance (format 1, whose targets name none, or name it in
// words beginning "primary"); nil if the set does not hold it, or the target
// names another instance in words (format 1 has no ids for them).
func targetInstance(set map[string]Annotation, claim string, ref any) map[string]any {
	t := set[claim]
	if t == nil {
		return nil
	}
	id, ok := ref.(string)
	if ref == nil || ok && (id == "" || strings.HasPrefix(id, "primary")) {
		return primary(t)
	}
	if !ok || !instanceID.MatchString(id) {
		return nil
	}
	insts, _ := t["instances"].([]any)
	for _, x := range insts {
		if in, _ := x.(map[string]any); in != nil && in["id"] == id {
			return in
		}
	}
	return nil
}

// relationUnsettled tells whether a target's formula or part is
// {"unresolved": why} (D102).
func relationUnsettled(actual map[string]any) bool {
	return isMarker(actual["formula"], "unresolved") || isMarker(actual["part"], "unresolved")
}

// isMarker tells whether a value is {"unresolved": why} or {"unstated": why}
// (docs/annotation-records.md §8): one key, a reason that is not blank.
func isMarker(v any, kind string) bool {
	m, ok := v.(map[string]any)
	if !ok || len(m) != 1 {
		return false
	}
	why, ok := m[kind].(string)
	return ok && strings.TrimSpace(why) != ""
}

// FieldAgreement is the agreement of two annotators on one field.
type FieldAgreement struct {
	N          int      `json:"n"`
	Agreement  float64  `json:"agreement"`
	Kappa      *float64 `json:"kappa"`                // Cohen's; nil when undefined (one category only)
	Unresolved int      `json:"unresolved,omitempty"` // claims left out: a value Unresolved on either side
}

// Pair is the agreement of two annotators (or an annotator and the gold).
type Pair struct {
	A                    string                    `json:"a"`
	B                    string                    `json:"b"`
	Items                int                       `json:"items"` // annotated by both
	Fields               map[string]FieldAgreement `json:"fields"`
	Coverage             *AssignmentCoverage       `json:"coverage,omitempty"`
	AllAssignedAgreement map[string]float64        `json:"all_assigned_agreement,omitempty"`
}

// Agree compares two sets of annotations field by field, over the items both
// have; each annotator's targets are read in its own set.
func Agree(nameA, nameB string, a, b map[string]Annotation) Pair {
	return agree(nameA, nameB, a, b, a, b)
}

// agree compares a and b, reading their targets in setA and setB.
func agree(nameA, nameB string, a, b, setA, setB map[string]Annotation) Pair {
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
			xs, ys = append(xs, FieldIn(a[id], f, setA)), append(ys, FieldIn(b[id], f, setB))
		}
		p.Fields[f] = kappa(xs, ys)
	}
	return p
}

// kappa is raw agreement and Cohen's kappa of two label lists, over the
// pairs neither of whose values is Unresolved.
func kappa(xs, ys []string) FieldAgreement {
	var rx, ry []string
	for i := range xs {
		if xs[i] != Unresolved && ys[i] != Unresolved {
			rx, ry = append(rx, xs[i]), append(ry, ys[i])
		}
	}
	n := len(rx)
	out := FieldAgreement{N: n, Unresolved: len(xs) - n}
	if n == 0 {
		return out
	}
	same := 0
	ca, cb := map[string]int{}, map[string]int{}
	for i := range rx {
		if rx[i] == ry[i] {
			same++
		}
		ca[rx[i]]++
		cb[ry[i]]++
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

// round3 rounds to three decimals, half away from zero. The conversion keeps
// x*1000 from being fused with the addition (an FMA on some processors), so
// that every build, and agreement.py's round3, gives the same figures.
func round3(x float64) float64 { return float64(int(float64(x*1000)+0.5*sign(x))) / 1000 }

func sign(x float64) float64 {
	if x < 0 {
		return -1
	}
	return 1
}

// Disagrees lists the fields on which the annotations of one claim differ;
// sets, when given, are each annotator's annotations (in the order of anns),
// in which its target is read. An Unresolved value differs from any other.
func Disagrees(anns []Annotation, sets ...map[string]Annotation) []string {
	var out []string
	for _, f := range Fields {
		seen := map[string]bool{}
		for i, a := range anns {
			var set map[string]Annotation
			if i < len(sets) {
				set = sets[i]
			}
			seen[FieldIn(a, f, set)] = true
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

var variable = regexp.MustCompile(`\b[A-Z][A-Za-z0-9_]*\b`)

// namesAny tells whether a formula names one of the variables.
func namesAny(f string, vars map[string]bool) bool {
	for _, v := range variable.FindAllString(f, -1) {
		if vars[v] {
			return true
		}
	}
	return false
}

// compact drops the spaces of a formula: "f(a, b)" and "f(a,b)" are the
// same formula.
func compact(formula string) string { return strings.Join(strings.Fields(formula), "") }

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

// Check lists what is wrong with an annotation of a claim of the dataset,
// in the dataset's format (CheckIn without the annotator's other
// annotations, so a target's instance is not looked for).
func (c *Collection) Check(a Annotation, d *Dataset) []string { return c.CheckIn(a, d, nil, 0) }

// CheckIn lists what is wrong with an annotation of a claim of the dataset:
// its shape; its scheme, which must be one of the collection's (or other,
// none); bindings for exactly the scheme's variables; premises,
// assumptions, exceptions and conclusion that are the scheme's formulas
// with the bindings put in; a target that is a claim of the dataset. In
// format 2 (docs/annotation-records.md) also its instances' ids (R12), its
// exact target (R13: the instance, read in set, the annotator's own
// annotations, and the formula), its actors and actions (R14) and contexts
// (R15) by the dataset's case record, and the field contract's nulls and
// markers (R16, in part). It is the core of docs/kialo-batch2/validate.py
// (which also checks the lexicon and the adjudication rules). format is 1
// or 2; 0 means the dataset's, or with no dataset the annotation's own
// (format 2 when its instances have ids).
func (c *Collection) CheckIn(a Annotation, d *Dataset, set map[string]Annotation, format int) []string {
	if format == 0 {
		format = 1
		if d != nil {
			format = d.Format
		} else if hasInstanceIDs(a) {
			format = 2
		}
	}
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
		// R18 (guidelines §5 item 6): fit explicit when every premise-role
		// formula is explicit; assumptions and exceptions may be supplied.
		// A format-1 record keeps the rules it was made with.
		if format == 2 && fit == "explicit" {
			ps, _ := in["premises"].([]any)
			for _, y := range ps {
				pm, _ := y.(map[string]any)
				if pm["role"] == "premise" && pm["status"] != "explicit" {
					add("R18 %v has fit explicit, but its premise %v is %v: fit reconstructed, or the premise stated (guidelines §5 item 6)", in["id"], pm["formula"], pm["status"])
				}
			}
		}
		binds := map[string]string{}
		badBinding := false
		if bm, ok := in["bindings"].(map[string]any); ok {
			for k, v := range bm {
				if format == 2 {
					// Format 2: an atom, or a fragment's {"unstated": why}
					// (R16); null is not a binding.
					if sv, ok := v.(string); ok && sv != "" {
						binds[k] = sv
					} else if !(isMarker(v, "unstated") && fit == "fragment") {
						add("R16 %v binds %s to %s: an atom, or in a fragment {\"unstated\": why}", in["id"], k, jsonText(v))
						badBinding = true
					}
					continue
				}
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
		if badBinding {
			continue // its formulas cannot be compared: said once, above
		}
		sc, ok := c.schemes[wid]
		if !ok {
			add("%s is not a scheme of walton.yml", wid)
			continue
		}
		declared := map[string]bool{}
		for _, v := range sc.Variables {
			declared[v] = true
		}
		if bm, ok := in["bindings"].(map[string]any); ok {
			for k := range bm {
				if !declared[k] {
					add("%s binds undeclared variable %s", wid, k)
				}
			}
		}
		// Format 2 (docs/annotation-records.md §5, as validate.py reads it):
		// a fragment binds every variable of its scheme, those of the
		// formulas it does not state {"unstated": why}, and states only
		// formulas whose variables it binds.
		unstated := map[string]bool{}
		if bm, ok := in["bindings"].(map[string]any); ok && format == 2 {
			for k, v := range bm {
				if isMarker(v, "unstated") {
					unstated[k] = true
				}
			}
		}
		if fit != "fragment" || format == 2 {
			want := append([]string(nil), sc.Variables...)
			var got []string
			for k := range binds {
				got = append(got, k)
			}
			for k := range unstated {
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
				if !namesAny(f, unstated) {
					expected[role+" "+compact(subst(f, binds))]++
				}
			}
		}
		ps, _ := in["premises"].([]any)
		for _, y := range ps {
			pm, _ := y.(map[string]any)
			k := fmt.Sprintf("%v %v", pm["role"], compact(fmt.Sprint(pm["formula"])))
			if expected[k] > 0 {
				expected[k]--
			} else {
				add("%s: %s is not a formula of the scheme with these bindings", wid, k)
			}
		}
		// Every formula of the scheme, but in a fragment that leaves some
		// variables unstated (format 1: any fragment), which states some.
		if fit != "fragment" || format == 2 && len(unstated) == 0 {
			for k, n := range expected {
				if n > 0 {
					add("%s: missing %s", wid, k)
				}
			}
		}
		if len(sc.Conclusions) > 0 {
			cm, _ := in["conclusion"].(map[string]any)
			hasFormula := cm != nil && cm["formula"] != nil && cm["formula"] != ""
			if namesAny(sc.Conclusions[0], unstated) {
				// A fragment whose conclusion names a variable it leaves
				// unstated has a null conclusion formula (§5).
				if hasFormula {
					add("R16 %v is a fragment whose conclusion names a variable it leaves unstated: its formula is null", in["id"])
				}
			} else if got, want := compact(fmt.Sprint(cm["formula"])), compact(subst(sc.Conclusions[0], binds)); (fit != "fragment" || hasFormula || format == 2) && got != want {
				add("%s concludes %s, not %s", wid, got, want)
			}
		}
	}
	if nprim != 1 && (len(insts) > 1 || format == 2) {
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
	// A claim at an exception, an answer, argues that the exception holds, so
	// its primary instance concludes it; a defence of the argument (polarity
	// for) argues that it does not, and concludes its negation (D075; rule R8
	// of validate.py; the replay plays it so).
	if tg, _ := a["target"].(map[string]any); tg != nil {
		if act, _ := tg["actual"].(map[string]any); act != nil && act["part"] == "exception" && !isMarker(act["formula"], "unresolved") {
			exc := ""
			if f := act["formula"]; f != nil {
				exc = compact(fmt.Sprint(f))
			}
			want := exc
			if p, _ := r["polarity"].(string); p == "for" {
				want = "¬" + exc
			}
			if got := primaryConclusion(insts); got != want {
				if got == "" {
					got = "nothing"
				}
				add("an answer at the exception %s concludes %s, not %s (D075)", exc, got, want)
			}
		}
	}
	if format == 2 {
		out = append(out, c.checkIdentity(a, d, set)...)
	}
	return out
}

// primaryConclusion is the conclusion of an annotation's primary instance
// (its only instance, if none is marked), without spaces; "" if it has none.
func primaryConclusion(insts []any) string {
	var own map[string]any
	for _, x := range insts {
		in, _ := x.(map[string]any)
		if p, _ := in["primary"].(bool); p {
			own = in
			break
		}
	}
	if own == nil && len(insts) == 1 {
		own, _ = insts[0].(map[string]any)
	}
	cm, _ := own["conclusion"].(map[string]any)
	if cm == nil || cm["formula"] == nil {
		return ""
	}
	return compact(fmt.Sprint(cm["formula"]))
}

// AssignmentCoverage keeps missing replies visible without counting two missing
// replies as label agreement. Conditional kappa remains a separate diagnostic.
type AssignmentCoverage struct {
	Assigned   int `json:"assigned"`
	AnnotatedA int `json:"annotated_a"`
	AnnotatedB int `json:"annotated_b"`
	Both       int `json:"both"`
	Neither    int `json:"neither"`
}

func AgreeAssigned(nameA, nameB string, a, b map[string]Annotation, assigned []string) Pair {
	aa, bb := map[string]Annotation{}, map[string]Annotation{}
	seen := map[string]bool{}
	c := &AssignmentCoverage{}
	for _, id := range assigned {
		if seen[id] {
			continue
		}
		seen[id] = true
		c.Assigned++
		if a[id] != nil {
			aa[id] = a[id]
			c.AnnotatedA++
		}
		if b[id] != nil {
			bb[id] = b[id]
			c.AnnotatedB++
		}
		if a[id] != nil && b[id] != nil {
			c.Both++
		}
		if a[id] == nil && b[id] == nil {
			c.Neither++
		}
	}
	// Each annotator's targets are read in its whole set, not only in the
	// claims assigned.
	out := agree(nameA, nameB, aa, bb, a, b)
	out.Coverage = c
	out.AllAssignedAgreement = map[string]float64{}
	if c.Assigned > 0 {
		for _, f := range Fields {
			same := 0
			for id := range aa {
				if bb[id] == nil {
					continue
				}
				if x := FieldIn(aa[id], f, a); x != Unresolved && x == FieldIn(bb[id], f, b) {
					same++
				}
			}
			out.AllAssignedAgreement[f] = float64(same) / float64(c.Assigned)
		}
	}
	return out
}
