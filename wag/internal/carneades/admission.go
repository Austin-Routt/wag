package carneades

// admission.go — what a graph must satisfy before its labels mean anything.
//
// An authored graph (statements, arguments and assumptions written in a YAML
// file, or built from a Kialo discussion) is labelled only if it is admitted:
// its arguments are instances of the schemes they name, and its assumptions
// are consistent. Without this, an argument whose premise is not its scheme's
// premise, or assumptions p and ¬p, were labelled In like any other (the
// critical review of 2026-11-23, its admission probe).
//
// The checks are the engine's own (package validation and CheckContext). WAG's
// profile has these explicit limits and adaptations:
//   - the theory (the language and the schemes as such) is checked by the game's
//     lint and the catalogue's checks, not here: WAG's language declares negated
//     predicates (¬p/1), which the engine's validator reports as not predicate
//     symbols. The patterns of a scheme an argument names must still be terms,
//     or the argument could not be checked against them;
//   - an answer made by the running game (D075) concludes the claim it answers,
//     not its scheme's conclusion (answerForm). Authored input cannot authorize
//     this waiver by setting metadata; its conclusion must fit the scheme;
//   - issue schemes are unsupported here: CheckContext examines their closure
//     on a private graph, but WAG does not infer that closure into the graph it
//     labels. Authored input must supply concrete issues instead.

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
	"github.com/carneades/carneades-4/src/engine/validation"
)

// ConclusionProblem is the engine validator's description of an argument
// whose conclusion is not an instance of its scheme's conclusion.
const ConclusionProblem = "conclusion does not instantiate a conclusion of the scheme with the argument's parameters."

// AnswerKey records, in an argument's metadata, the variable bound to the claim
// answered by AnswerQuestion. It is descriptive metadata, not authorization:
// an authored YAML file cannot use it to waive its scheme's conclusion check.
const AnswerKey = "wag_answer_variable"

// ProfileProblems returns what is wrong with the graph's content under WAG's
// runtime profile: the engine's validation of its statements, issues, arguments,
// assumptions and expected labels, less the answers WAG concludes as the claim
// they answer (answerForm), plus unreadable named patterns and unsupported issue
// schemes. The YAML constructor uses the stricter authored profile, which never
// takes an answer-fold waiver from input metadata. Problems are sorted.
func ProfileProblems(ag *caes.ArgGraph) []validation.Problem {
	return profileProblems(ag, true)
}

func profileProblems(ag *caes.ArgGraph, allowAnswerFold bool) []validation.Problem {
	if ag == nil {
		return nil
	}
	var out []validation.Problem
	for _, p := range validation.Validate(ag) {
		switch p.Category {
		case validation.LANGUAGE, validation.SCHEME, validation.ISCHEME:
			continue // the theory: the lint's and the catalogue's
		case validation.ARGUMENT:
			if p.Description == ConclusionProblem && answerForm(ag.Arguments[p.Id], allowAnswerFold) {
				continue
			}
		}
		out = append(out, p)
	}
	out = append(out, unreadablePatterns(ag)...)
	if ag.Theory != nil {
		for id := range ag.Theory.IssueSchemes {
			out = append(out, validation.Problem{Category: validation.ISCHEME, Id: id,
				Description: "issue schemes are unsupported in WAG's admission profile (authored and runtime graphs); supply concrete issues because WAG does not materialize issue schemes before labelling"})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Id != b.Id {
			return a.Id < b.Id
		}
		if a.Description != b.Description {
			return a.Description < b.Description
		}
		return a.Expression < b.Expression
	})
	return out
}

// answerForm reports whether an argument concludes, in WAG's own spelling,
// what its scheme concludes, or is an answer as WAG makes one (D075,
// AnswerQuestion). Its scheme concludes one variable V, and the argument
// concludes the statement V is bound to: the same statement, ¬X written
// not(X) inside a term (StatementTerm). Or its scheme concludes a one-place
// predicate of V that is not a negation (true_in_situation(B)), and the
// argument is marked as an answer for V (AnswerKey), and the caller permits
// runtime answer folding: an answer concludes the claim it answers. The YAML
// constructor does not permit that fold. Only runtime answers are folded so:
// an authored argument from positive
// consequences that concludes the bare action, not should_be_performed of it,
// is not admitted (the review of 2026-11-24, R4). The engine's validator
// expects the scheme's own conclusion and so reports these arguments; the
// game's lint requires one of these shapes of an answer scheme.
func answerForm(arg *caes.Argument, allowAnswerFold bool) bool {
	if arg == nil || arg.Scheme == nil || arg.Conclusion == nil || len(arg.Scheme.Conclusions) != 1 ||
		len(arg.Parameters) != len(arg.Scheme.Variables) {
		return false
	}
	t, ok := terms.ReadString(arg.Scheme.Conclusions[0])
	if !ok {
		return false
	}
	folded := false
	if c, ok := t.(terms.Compound); ok {
		if !allowAnswerFold || len(c.Args) != 1 || IsNegationFunctor(c.Functor) {
			return false
		}
		t, folded = c.Args[0], true
	}
	v, ok := t.(terms.Variable)
	if !ok {
		return false
	}
	if marked, _ := arg.Metadata[AnswerKey].(string); folded && marked != v.Name {
		return false
	}
	for i, name := range arg.Scheme.Variables {
		if name != v.Name {
			continue
		}
		bound, ok1 := terms.ReadString(arg.Parameters[i])
		claim, ok2 := terms.ReadString(StatementTerm(arg.Conclusion.Id))
		return ok1 && ok2 && bound.String() == claim.String()
	}
	return false
}

// IsNegationFunctor reports whether a one-place functor negates its argument:
// ¬, as statements write it, or not, as terms inside a statement do
// (StatementTerm).
func IsNegationFunctor(f string) bool { return f == "¬" || f == "not" }

// unreadablePatterns reports the patterns of the schemes the graph's
// arguments name that are not terms. The engine's validator skips such a
// pattern when it checks an argument ("patterns of schemes are checked
// elsewhere"), so an argument of a scheme with an unreadable premise would
// pass with any premise.
func unreadablePatterns(ag *caes.ArgGraph) []validation.Problem {
	var out []validation.Problem
	seen := map[string]bool{}
	for _, arg := range ag.Arguments {
		if arg == nil || arg.Scheme == nil || caes.IsBasicScheme(arg.Scheme) {
			continue
		}
		s := arg.Scheme
		for _, group := range [][]string{s.Premises, s.Assumptions, s.Exceptions, s.Conclusions} {
			for _, f := range group {
				if t, ok := terms.ReadFormula(f); ok && t != nil {
					continue
				}
				if key := s.Id + "\x00" + f; !seen[key] {
					seen[key] = true
					out = append(out, validation.Problem{Category: validation.SCHEME, Id: s.Id,
						Description: "a pattern of a scheme an argument names is not a term", Expression: f})
				}
			}
		}
	}
	return out
}

// AdmissionError lists what keeps a graph from being admitted.
type AdmissionError struct {
	Problems []validation.Problem // ProfileProblems
	Context  error                // the engine's CheckContext
}

func (e *AdmissionError) Error() string {
	var parts []string
	const shown = 5
	for i, p := range e.Problems {
		if i == shown {
			parts = append(parts, fmt.Sprintf("and %d more", len(e.Problems)-shown))
			break
		}
		s := p.Category.String()
		if p.Id != "" {
			s += " " + p.Id
		}
		s += ": " + strings.TrimSuffix(p.Description, ".")
		if p.Expression != "" {
			s += " (" + p.Expression + ")"
		}
		parts = append(parts, s)
	}
	if e.Context != nil {
		parts = append(parts, "the assumptions are not consistent: "+e.Context.Error())
	}
	return "the graph is not admitted: " + strings.Join(parts, "; ")
}

// Admit checks the graph as it stands: ProfileProblems, and the consistency of
// its assumptions and issues (the engine's CheckContext: no two positions of
// one issue, nor a statement and its negation, are accepted). It returns an
// *AdmissionError, or nil when the graph is admitted. It changes nothing.
func (a *Adapter) Admit() error {
	if a == nil || a.Graph == nil {
		return errors.New("no argument graph loaded")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return admit(a.Graph, true)
}

func admit(ag *caes.ArgGraph, allowAnswerFold bool) error {
	e := &AdmissionError{Problems: profileProblems(ag, allowAnswerFold), Context: ag.CheckContext()}
	if len(e.Problems) == 0 && e.Context == nil {
		return nil
	}
	return e
}
