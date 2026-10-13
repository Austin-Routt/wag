package nascq

import (
	"fmt"
	"regexp"
	"sort"
)

// LintIssue is one finding of Lint: an Error makes some move impossible or
// wrong; a Warning is legal data with a consequence the author should know.
type LintIssue struct {
	Error bool
	Where string // node or edge the finding is about
	Msg   string
}

func (i LintIssue) String() string {
	sev := "warning"
	if i.Error {
		sev = "error"
	}
	return fmt.Sprintf("%s: %s: %s", sev, i.Where, i.Msg)
}

// placeholder matches a {Var} placeholder in an English template or pattern.
var placeholder = regexp.MustCompile(`\{([^{}]*)\}`)

// Placeholders returns the variable names used as {Var} in s, in order.
func Placeholders(s string) []string {
	var out []string
	for _, m := range placeholder.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

// Lint checks the NASCQ data for problems no single move reveals. Findings
// are sorted by severity (errors first), then by place.
//
// Errors: an edge to a missing node or a duplicate edge; a scheme without
// carneades_id, or with a base weight outside (0, 1], or a template naming an
// undeclared variable; a CQ whose caes_mapping has no attack type, an
// undermining CQ without target_premise_pattern (it is never enumerated), or
// a pattern naming a variable that a scheme it is asked of does not declare.
//
// Warnings: a scheme no CQ is asked of (arguments made with it, including
// attacks, can never be challenged); a CQ asked of no scheme; a CQ whose
// question names a variable that a scheme it is asked of does not declare; a
// thesis-dependent backup edge between different semantic types (the check
// that move validation used to attempt as S4); and two schemes sharing a
// carneades_id (arguments are mapped back to the smaller NASCQ ID).
func (n *NASCQ) Lint() []LintIssue {
	var out []LintIssue
	add := func(isErr bool, where, format string, args ...interface{}) {
		out = append(out, LintIssue{Error: isErr, Where: where, Msg: fmt.Sprintf(format, args...)})
	}

	for _, msg := range n.ValidateBipartite() {
		add(true, "edges", "%s", msg)
	}
	seen := map[string]bool{}
	for _, e := range n.AskEdges {
		cq, s := n.resolveAskEdge(e)
		if k := "ask " + cq + " " + s; seen[k] {
			add(true, "edges", "duplicate ask edge %s -> %s", cq, s)
		} else {
			seen[k] = true
		}
	}
	for _, e := range n.BackupEdges {
		s, cq := n.resolveBackupEdge(e)
		if k := "backup " + s + " " + cq; seen[k] {
			add(true, "edges", "duplicate backup edge %s -> %s", s, cq)
		} else {
			seen[k] = true
		}
		sch, q := n.Schemes[s], n.Questions[cq]
		if sch == nil || q == nil {
			continue
		}
		if sch.ThesisIndependent {
			add(false, s, "thesis-independent, so its backup edge to %s is redundant", cq)
		} else if sch.SemanticType != "" && q.SemanticType != "" && sch.SemanticType != q.SemanticType {
			add(false, s, "backs %s across semantic types (%s answers %s)", cq, sch.SemanticType, q.SemanticType)
		}
	}

	byCarneades := map[string][]string{}
	for _, id := range sortedSchemeIDs(n.Schemes) {
		s := n.Schemes[id]
		if s.CarneadesID == "" {
			add(true, id, "no carneades_id")
		} else {
			byCarneades[s.CarneadesID] = append(byCarneades[s.CarneadesID], id)
		}
		if !(s.BaseWeight > 0 && s.BaseWeight <= 1) {
			add(true, id, "base_weight %v is not in (0, 1]", s.BaseWeight)
		}
		for _, tmpl := range append(append([]string{}, s.PremiseTemplatesEN...), s.ConclusionTemplateEN) {
			for _, v := range Placeholders(tmpl) {
				if _, ok := s.Variables[v]; !ok {
					add(true, id, "template %q names undeclared variable %q", tmpl, v)
				}
			}
		}
		if len(n.askIndex[id]) == 0 {
			add(false, id, "no critical question is asked of it: its arguments, attacks included, can never be challenged")
		}
	}
	for _, cid := range SortedStringKeys(byCarneades) {
		if ids := byCarneades[cid]; len(ids) > 1 {
			add(false, cid, "carneades_id of %v; its arguments map back to %s", ids, ids[0])
		}
	}

	askedOf := map[string][]string{}
	for sid, cqs := range n.askIndex {
		for _, cq := range cqs {
			askedOf[cq] = append(askedOf[cq], sid)
		}
	}
	for _, id := range SortedStringKeys(n.Questions) {
		q := n.Questions[id]
		at := AttackTypeOf(q)
		if at == "" {
			add(true, id, "caes_mapping %+v has no attack type (rebuttal, undercutter, undermining)", q.CaesMapping)
		}
		pattern := q.CaesMapping.TargetPremisePattern
		if at == "undermining" && pattern == "" {
			add(true, id, "undermining without target_premise_pattern: no premise to attack, so it is never offered")
		}
		schemes := askedOf[id]
		sort.Strings(schemes)
		if len(schemes) == 0 {
			add(false, id, "asked of no scheme")
		}
		for _, sid := range schemes {
			s := n.Schemes[sid]
			if s == nil {
				continue
			}
			for _, v := range Placeholders(pattern) {
				if _, ok := s.Variables[v]; !ok {
					add(true, id, "target_premise_pattern %q names %q, not a variable of %s", pattern, v, sid)
				}
			}
			for _, v := range Placeholders(q.QuestionTemplateEN) {
				if _, ok := s.Variables[v]; !ok {
					add(false, id, "question names {%s}, not a variable of %s", v, sid)
				}
			}
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Error != out[j].Error {
			return out[i].Error
		}
		return out[i].Where < out[j].Where
	})
	return out
}

// SortedStringKeys returns the keys of a string-keyed map in lexical order.
func SortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
