package annotate

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
)

// The identity contract's checks (docs/annotation-records.md, D092) on a
// format-2 annotation, the part the Go check shares with validate.py: R12
// (ids), R13 (exact targets), R14 (actors and actions), R15 (contexts) and
// R16's nulls and markers. R4, R9, R10, R11 and R17 are validate.py's.

// instanceID is an instance's id, <claim>#<n>: a positive number, no
// leading zero.
var instanceID = regexp.MustCompile(`^(.+)#([1-9][0-9]*)$`)

// hasInstanceIDs tells whether an annotation's instances carry ids, as
// format 2's do.
func hasInstanceIDs(a Annotation) bool {
	insts, _ := a["instances"].([]any)
	for _, x := range insts {
		if in, _ := x.(map[string]any); in != nil {
			if _, ok := in["id"]; ok {
				return true
			}
		}
	}
	return false
}

// jsonText is a value as JSON, for messages.
func jsonText(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

// thesisID is the discussion's thesis, from items.json's discussion; "" if
// it names none.
func (d *Dataset) thesisID() string {
	var x struct {
		Thesis string `json:"thesis_claim_id"`
	}
	_ = json.Unmarshal(d.Discussion, &x)
	return x.Thesis
}

// checkIdentity is CheckIn's format-2 part.
func (c *Collection) checkIdentity(a Annotation, d *Dataset, set map[string]Annotation) []string {
	var out []string
	add := func(f string, x ...any) { out = append(out, fmt.Sprintf(f, x...)) }
	claim, _ := a["kialo_claim_id"].(string)
	var cs *Case
	if d != nil {
		cs = d.Case
	}
	// R12: every instance has an id of its own claim, used once, not retired.
	retired := map[string]bool{}
	if rv, ok := a["retired"]; ok {
		list, ok := rv.([]any)
		if !ok {
			add("R12 its retired ids are not a list of instance ids")
		}
		for _, x := range list {
			s, ok := x.(string)
			if !ok || !instanceID.MatchString(s) {
				add("R12 its retired ids are not a list of instance ids")
				break
			}
			retired[s] = true
		}
	}
	insts, _ := a["instances"].([]any)
	seen := map[string]bool{}
	var own map[string]any
	for _, x := range insts {
		in, _ := x.(map[string]any)
		if in == nil {
			continue
		}
		id, _ := in["id"].(string)
		m := instanceID.FindStringSubmatch(id)
		switch {
		case m == nil || claim != "" && m[1] != claim:
			add("R12 an instance has the id %s, not %s#<n>", jsonText(in["id"]), claim)
		case seen[id]:
			add("R12 two instances have the id %s", id)
		case retired[id]:
			add("R12 the instance %s has a retired id", id)
		}
		seen[id] = true
		p, ok := in["primary"].(bool)
		if !ok {
			add("R16 the instance %s says nothing of primary (true or false)", id)
		}
		if p && own == nil {
			own = in
		}
		if cx, ok := in["context"]; ok && cs != nil {
			if s, _ := cx.(string); !cs.contexts[s] {
				add("R15 the instance %s is in the context %s, which is not one of the case's", id, jsonText(cx))
			}
		}
		if cs != nil {
			out = append(out, c.checkActors(in, id, cs)...)
		}
	}
	if own == nil && len(insts) == 1 {
		own, _ = insts[0].(map[string]any)
	}
	// R16: the target and the polarity, null only for the thesis; no stored
	// re-parenting.
	tg, _ := a["target"].(map[string]any)
	role, _ := a["role"].(map[string]any)
	if _, ok := tg["re_parented"]; ok {
		add("R16 it records re_parented, which is derived from the target, not recorded")
	}
	actual, isMap := tg["actual"].(map[string]any)
	if d != nil && d.thesisID() != "" {
		if claim == d.thesisID() {
			if tg["actual"] != nil || tg["kialo_parent"] != nil || tg["kialo_relation"] != nil {
				add("R16 the thesis has no target: its kialo_parent, kialo_relation and actual are null")
			}
			if role["polarity"] != nil {
				add("R16 the thesis has no polarity (null)")
			}
		} else {
			if !isMap {
				add("R16 its actual target is %s: only the thesis has none", jsonText(tg["actual"]))
			}
			if p := role["polarity"]; p != "for" && p != "against" {
				add("R16 its polarity is %s: for or against, null only for the thesis", jsonText(p))
			}
		}
	}
	// R13: the target names an instance of its claim and a formula of it.
	var tp map[string]any
	if isMap && len(actual) > 0 {
		tp = c.checkTarget(&out, "its target", actual["claim"], actual["instance"], actual["part"], actual["formula"], set)
	}
	if others, ok := tg["other_kialo_parents"].([]any); ok {
		for _, x := range others {
			o, _ := x.(map[string]any)
			if part, _ := o["part"].(string); part == "premise" || part == "exception" || part == "conclusion" {
				c.checkTarget(&out, fmt.Sprintf("its target in %v", o["claim_id"]), o["claim_id"], o["instance"], o["part"], o["formula"], set)
			}
		}
	}
	// R15: the claim's own argument is in its target instance's context.
	if cs != nil && own != nil && tp != nil {
		if x, y := contextOf(own, cs), contextOf(tp, cs); x != y {
			add("R15 its primary instance is in the context %s, its target %v in %s", jsonText(x), actual["instance"], jsonText(y))
		}
	}
	return out
}

// contextOf is an instance's context: its own, or the case's default.
func contextOf(in map[string]any, cs *Case) string {
	if x, ok := in["context"].(string); ok {
		return x
	}
	return cs.DefaultContext
}

// checkTarget checks one exact target (R13, and R16's shape of it): the
// instance it names, its part and its formula. It returns the instance,
// looked up in the annotator's own annotations (set), or nil when it cannot
// be read: unresolved, not well named, or not in set.
func (c *Collection) checkTarget(out *[]string, where string, claimV, ref, part, formula any, set map[string]Annotation) map[string]any {
	add := func(f string, x ...any) { *out = append(*out, fmt.Sprintf(f, x...)) }
	claim, _ := claimV.(string)
	if isMarker(ref, "unresolved") {
		return nil
	}
	id, ok := ref.(string)
	if !ok || id == "" {
		add("R16 %s names no instance (an instance id, or {\"unresolved\": why})", where)
		return nil
	}
	if m := instanceID.FindStringSubmatch(id); m == nil || m[1] != claim {
		add("R13 %s names the instance %s, which is not an instance of %s", where, id, claim)
		return nil
	}
	pt, _ := part.(string)
	switch {
	case isMarker(part, "unresolved"):
	case pt == "inference":
		if formula != nil {
			add("R16 %s is the inference, which has no formula, but names %s", where, jsonText(formula))
		}
	case pt == "premise" || pt == "exception" || pt == "conclusion":
		if fs, ok := formula.(string); (!ok || fs == "") && !isMarker(formula, "unresolved") {
			add("R16 %s names no formula of its %s", where, pt)
		}
	default:
		add("R13 %s has the part %s: premise, exception, conclusion or inference", where, jsonText(part))
	}
	t := set[claim]
	if t == nil {
		return nil // the annotator has not annotated its target: nothing to look the instance up in
	}
	in := targetInstance(set, claim, id)
	if in == nil {
		add("R13 %s names the instance %s, which is not an instance of %s", where, id, claim)
		return nil
	}
	if fs, ok := formula.(string); ok && fs != "" && (pt == "premise" || pt == "exception" || pt == "conclusion") {
		found := false
		for _, f := range formulasOf(in, pt) {
			found = found || f == compact(fs)
		}
		if !found {
			add("R13 %s names %s, which is not a %s of %s", where, fs, pt, id)
		}
	}
	return in
}

// formulasOf is an instance's formulas in a part of a target: premise (its
// premises and assumptions), exception, or conclusion.
func formulasOf(in map[string]any, part string) []string {
	if part == "conclusion" {
		cm, _ := in["conclusion"].(map[string]any)
		if f, ok := cm["formula"].(string); ok && f != "" {
			return []string{compact(f)}
		}
		return nil
	}
	roles := map[string]bool{}
	switch part {
	case "premise":
		roles["premise"], roles["assumption"] = true, true
	case "exception":
		roles["exception"] = true
	}
	var out []string
	ps, _ := in["premises"].([]any)
	for _, y := range ps {
		pm, _ := y.(map[string]any)
		if r, _ := pm["role"].(string); roles[r] {
			if f, ok := pm["formula"].(string); ok {
				out = append(out, compact(f))
			}
		}
	}
	return out
}

// checkActors is R14 for one instance: an agent's place holds an actor of
// the case, an action's place an action of it (or an alias of one), and an
// instance with an agent binds only that agent's actions: who performs an
// action comes from the case's mapping, not from the action's name (D063,
// D091).
func (c *Collection) checkActors(in map[string]any, id string, cs *Case) []string {
	s, _ := in["scheme"].(map[string]any)
	wid, _ := s["walton_id"].(string)
	sc, ok := c.schemes[wid]
	if !ok {
		return nil
	}
	bm, _ := in["bindings"].(map[string]any)
	t := typedVars(sc)
	agents := map[string]bool{}
	for v, k := range t {
		if b, ok := bm[v].(string); ok && b != "" && k == "agent" {
			agents[b] = true
		}
	}
	var out []string
	if len(agents) > 1 {
		out = append(out, fmt.Sprintf("R14 %s names more than one agent: %v", id, sortedSet(agents)))
	}
	vars := make([]string, 0, len(t))
	for v := range t {
		vars = append(vars, v)
	}
	sort.Strings(vars)
	for _, v := range vars {
		b, ok := bm[v].(string)
		if !ok || b == "" {
			continue
		}
		switch t[v] {
		case "agent":
			if !cs.actors[b] {
				out = append(out, fmt.Sprintf("R14 %s binds the agent %s to %s, which is not an actor of the case", id, v, b))
			}
		case "action":
			act, ok := cs.canon[b]
			if !ok {
				out = append(out, fmt.Sprintf("R14 %s binds the action %s to %s, which is not an action of the case", id, v, b))
				continue
			}
			if len(agents) == 1 {
				for ag := range agents {
					if cs.actorOf[act] != ag {
						out = append(out, fmt.Sprintf("R14 %s binds %s to %s, which is %s's action, not %s's", id, v, b, cs.actorOf[act], ag))
					}
				}
			}
		}
	}
	return out
}

func sortedSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
