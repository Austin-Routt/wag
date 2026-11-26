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
// carneades_id, or with a base weight outside (0, 1], or a template or
// answer_variable naming an undeclared variable; an ask edge without a locus (named on the edge or
// implied by the CQ's caes_mapping), a premise or exception locus without a
// pattern, or a pattern naming a variable the scheme does not declare;
// a free-text bridge that would replace the backing scheme's conclusion;
// or schemes sharing a carneades_id, which loses their distinct CQ identity.
//
// Warnings: a scheme no CQ is asked of (arguments made with it, including
// attacks, can never be challenged); a CQ asked of no scheme; a CQ whose
// question names a variable that a scheme it is asked of does not declare; a
// thesis-dependent backup edge between different semantic types (the check
// that move validation used to attempt as S4).
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
		if e.Bridge != "" {
			add(true, s, "free-text bridge to %s cannot change a scheme's conclusion; use a typed answer scheme", cq)
		}
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
		switch s.Evidence {
		case "", "required", "recommended":
		default:
			add(true, id, "evidence %q is not \"required\", \"recommended\" or empty", s.Evidence)
		}
		if s.Evidence == "required" && s.EvidenceNoteEN == "" {
			add(false, id, "evidence is required but evidence_note_en does not say what to cite")
		}
		if v := s.AnswerVariable; v != "" {
			if _, ok := s.Variables[v]; !ok {
				add(true, id, "answer_variable %q is not a declared variable", v)
			}
		}
		if len(n.askIndex[id]) == 0 && s.Role != RoleAnswer {
			add(false, id, "no critical question is asked of it: its arguments, attacks included, can never be challenged")
		}
		n.lintRole(id, s, add)
		for _, pair := range s.Distinct {
			if len(pair) != 2 || pair[0] == pair[1] {
				add(true, id, "distinct %v is not two different variables", pair)
				continue
			}
			for _, v := range pair {
				if spec, ok := s.Variables[v]; !ok || spec.Type != VarAction {
					add(true, id, "distinct names %s, which is not an action variable of the scheme", v)
				}
			}
		}
	}
	for _, cid := range SortedStringKeys(byCarneades) {
		if ids := byCarneades[cid]; len(ids) > 1 {
			add(true, cid, "carneades_id is shared by %v; distinct game schemes require distinct engine identities", ids)
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
			where := id + " asked of " + sid
			l := n.LocusOf(id, sid)
			switch l.Kind {
			case LocusPremise, LocusException:
				if l.Pattern == "" {
					add(true, where, "a %s locus needs a pattern naming the %s", l.Kind, l.Kind)
				}
			case LocusConclusion, LocusArgument:
			case "":
				add(true, where, "no locus: the ask edge names none and caes_mapping %+v implies none", q.CaesMapping)
			default:
				add(true, where, "unknown locus %q (premise, exception, conclusion, argument)", l.Kind)
			}
			for _, v := range Placeholders(l.Pattern) {
				if _, ok := s.Variables[v]; !ok {
					add(true, where, "pattern %q names %q, not a variable of the scheme", l.Pattern, v)
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

// lintRole checks a scheme's role (D094 (g)): an answer scheme concludes a
// formula and is offered on its backup edges, so it has one at least, no
// answer variable, and backs no question by being thesis-independent; a
// modal variant and a bridge name the scheme they weigh as, a chain with no
// cycle; only a bridge has an entry premise, and it has one.
func (n *NASCQ) lintRole(id string, s *ASNode, add func(bool, string, string, ...interface{})) {
	switch s.Role {
	case "", RoleAnswer, RoleModal, RoleBridge:
	default:
		add(true, id, "role %q is not %q, %q, %q or empty", s.Role, RoleAnswer, RoleModal, RoleBridge)
		return
	}
	if s.Role == RoleAnswer {
		backed := false
		for _, e := range n.BackupEdges {
			if sid, _ := n.resolveBackupEdge(e); sid == id {
				backed = true
			}
		}
		if !backed {
			add(true, id, "an answer scheme with no backup edge can never be played")
		}
		if s.ThesisIndependent || s.AnswerVariable != "" || s.Free {
			add(true, id, "an answer scheme concludes the formula of the question it answers: it is not thesis-independent, free, or with an answer variable")
		}
	}
	weighs := s.Role == RoleModal || s.Role == RoleBridge
	switch {
	case weighs && s.WeighsAs == "":
		add(true, id, "a %s names the scheme it weighs as (weighs_as)", s.Role)
	case !weighs && s.WeighsAs != "":
		add(true, id, "weighs_as is for a modal variant or a bridge, not a scheme of role %q", s.Role)
	case weighs:
		seen := map[string]bool{id: true}
		for cur := s; cur.WeighsAs != ""; {
			next := n.Schemes[cur.WeighsAs]
			if next == nil {
				add(true, id, "weighs_as %q is not a scheme", cur.WeighsAs)
				break
			}
			if seen[cur.WeighsAs] {
				add(true, id, "weighs_as makes a cycle through %q", cur.WeighsAs)
				break
			}
			seen[cur.WeighsAs] = true
			cur = next
		}
	}
	switch {
	case s.Role == RoleBridge && s.EntryPremise == "":
		add(true, id, "a bridge names its entry premise, which a modal argument supports (entry_premise)")
	case s.Role != RoleBridge && s.EntryPremise != "":
		add(true, id, "entry_premise is for a bridge, not a scheme of role %q", s.Role)
	case s.Role == RoleBridge:
		for _, v := range Placeholders(s.EntryPremise) {
			if _, ok := s.Variables[v]; !ok {
				add(true, id, "entry_premise %q names %q, not a variable of the scheme", s.EntryPremise, v)
			}
		}
	}
}
