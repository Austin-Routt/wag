// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Carneades Argument Evaluation Structure (CAES)
// This version of CAES supports cyclic argument graphs,
// cumulative arguments and IBIS.

package caes

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// The data types are sorted alphabetically

type Argument struct {
	Id          string
	Metadata    Metadata
	Scheme      *Scheme
	Parameters  []string // the values of the scheme variables, in the same order
	Premises    []Premise
	Conclusion  *Statement
	Undercutter *Statement
	Weight      float64 // for storing the evaluated weight
}

type ArgGraph struct {
	Metadata         Metadata
	Issues           map[string]*Issue // id to *Issue
	Statements       map[string]*Statement
	Arguments        map[string]*Argument
	References       map[string]Metadata // key -> metadata
	Theory           *Theory
	Assumptions      []string         // atomic formulas or statement keys
	assums           map[string]bool  // map representation of the assumptions
	ExpectedLabeling map[string]Label // for testing
	// Questioned lists the statements called into question
	// (docs/contracts.md, contract 5). A questioned default, an assumption
	// that only a scheme's `assumptions:` put there, is not assumed: it then
	// needs support like an ordinary premise. A stipulated statement stays
	// assumed: questioning does not withdraw a stipulation.
	Questioned []string
	// Defaults are the assumptions added by instantiating a scheme, by their
	// normalized key, as opposed to the stipulated ones of the graph itself.
	Defaults map[string]bool
}

type Issue struct {
	Id        string
	Metadata  Metadata
	Positions []*Statement
	Standard  Standard
}

// And IssueScheme is list atomic formulas, which may
// contain schema variables.  Schema variables are denoted
// using Prolog's syntax for variables. Use "..." to indicate
// a variable number of positions, as in this example:
// {"buy(O1)", "...", "buy(On)"}
type IssueScheme []string

type Label int

const (
	Undecided Label = iota
	In
	Out
)

type Labelling map[*Statement]Label

// The keys of a Language map denote the predicate and its arity,
// using Prolog lexical conventions. The values are Go formatting
// strings, for displaying logical formulas in natural language.
// example: {"price/2": "The price of a %v is %v."}
type Language map[string]string

type Metadata map[string]interface{}

type Premise struct {
	Stmt *Statement
	Role string // e.g. major, minor
}

type Scheme struct {
	Id       string
	Metadata Metadata
	// Each parameter is a schema variable, using
	// Prolog syntax for variables, i.e. identifiers starting
	// a capital letter
	Variables   []string // declaration of schema variables
	Weight      WeighingFunction
	Roles       []string // Roles[i] is the role name of Premises[i]
	Premises    []string // list of atomic formulas
	Assumptions []string // list of atomic formulas
	Exceptions  []string // list of atomic formulas
	// Deletions and Guards are extensions for implementing
	// schemes using Constrating Handling Rules (CHR)
	Deletions []string // list of atomic formulas
	Guards    []string // list of atomic formulas
	// Note that multiple conclusions are allowed, as in CHR
	Conclusions []string // list of atomic formulas or schema variables
	// WeightSpec is the weight as written in the input (a weighing function's
	// name, or a constant, criteria or preference specification), kept so
	// that an export can write it back (PLAN.md 4.1)
	WeightSpec interface{}
}

// Proof Standards
type Standard int

const (
	PE  Standard = iota // preponderance of the evidence
	CCE                 // clear and convincing evidence
	BRD                 // beyond reasonable doubt
)

type Statement struct {
	Id            string // a ground atomic formula, using Prolog syntax
	Metadata      Metadata
	Text          string      // natural language
	Issue         *Issue      // nil if not at issue
	Args          []*Argument // concluding with this statement
	Label         Label       // for storing the evaluated label
	IsUndercutter bool        // true if the statement is an undercutter
}

// A Rulebase is a set of constraint handling rules.
// https://en.wikipedia.org/wiki/Constraint_Handling_Rules
// The strings in the methods of the Rulebase interface
// denote terms, represented using Prolog syntax.
type Rulebase interface {
	// Infer returns true if the goals succeed and false if they fail.
	// The list of strings returned represents the constraint store
	Infer(goals []string, max int) (bool, []string, error)
	// AddRule adds a constraint handling rule to the rulebase
	AddRule(name string, keep []string, delete []string, guard []string, body []string) error
}

type Theory struct { // aka Knowledge Base
	Language          Language
	WeighingFunctions map[string]WeighingFunction
	// WeighingSpecs are the named weighing functions as written in the input,
	// kept for export (PLAN.md 4.1)
	WeighingSpecs map[string]interface{}
	ArgSchemes    []*Scheme
	IssueSchemes  map[string]*IssueScheme
	schemeIndex   map[string]*Scheme
}

type WeighingFunction func(*Argument, Labelling) float64 // [0.0,1.0]

func NewMetadata() Metadata {
	return make(map[string]interface{})
}

func NewIssue() *Issue {
	return &Issue{
		Metadata:  NewMetadata(),
		Positions: []*Statement{},
		Standard:  PE,
	}
}

func NewStatement() *Statement {
	return &Statement{
		Metadata: NewMetadata(),
		Args:     []*Argument{},
	}
}

func DefaultValidityCheck(*Argument) bool {
	return true
}

func NewArgument() *Argument {
	return &Argument{
		Metadata:   NewMetadata(),
		Premises:   []Premise{},
		Parameters: []string{},
	}
}

func NewTheory() *Theory {
	return &Theory{
		Language:          make(map[string]string),
		WeighingFunctions: make(map[string]WeighingFunction),
		ArgSchemes:        []*Scheme{},
		IssueSchemes:      make(map[string]*IssueScheme),
		schemeIndex:       make(map[string]*Scheme),
	}
}

func (t *Theory) InitSchemeIndex() {
	for _, s := range t.ArgSchemes {
		t.schemeIndex[s.Id] = s
	}
}

func NewArgGraph() *ArgGraph {
	return &ArgGraph{
		Metadata:         NewMetadata(),
		Issues:           map[string]*Issue{},
		Statements:       map[string]*Statement{},
		Arguments:        map[string]*Argument{},
		References:       make(map[string]Metadata),
		Assumptions:      []string{},
		assums:           make(map[string]bool),
		Theory:           NewTheory(),
		ExpectedLabeling: map[string]Label{},
		Defaults:         map[string]bool{},
	}
}

// IsQuestioned reports whether a statement has been called into question.
func (ag *ArgGraph) IsQuestioned(id string) bool {
	n := terms.Normalize(id)
	for _, q := range ag.Questioned {
		if terms.Normalize(q) == n {
			return true
		}
	}
	return false
}

// Assumed lists the statements the graph assumes now, normalized: its
// assumptions, except the defaults that have been called into question, and
// the defaults that meet an assumed contrary (docs/contracts.md, contract 5).
// A contrary is the statement's negation or another position of its issue.
// A default yields to a stipulated contrary; two defaults that contradict
// each other are both suspended, and no side is picked.
func (ag *ArgGraph) Assumed() []string {
	cand := []string{}
	seen := map[string]bool{}
	for _, a := range ag.Assumptions {
		n := terms.Normalize(a)
		if seen[n] || (ag.Defaults[n] && ag.IsQuestioned(n)) {
			continue
		}
		seen[n] = true
		cand = append(cand, n)
	}
	out := []string{}
	for _, n := range cand {
		if ag.Defaults[n] && ag.meetsContrary(n, seen) {
			continue
		}
		out = append(out, n)
	}
	return out
}

// meetsContrary reports whether a contrary of the statement n is in the set.
func (ag *ArgGraph) meetsContrary(n string, set map[string]bool) bool {
	if t, ok := terms.ReadFormula(n); ok {
		if c, ok := t.(terms.Compound); ok && c.Functor == "¬" && len(c.Args) == 1 {
			if set[terms.Normalize(c.Args[0].String())] {
				return true
			}
		} else if set[terms.Normalize("¬"+n)] {
			return true
		}
	}
	for k, s := range ag.Statements {
		if s == nil || s.Issue == nil || terms.Normalize(k) != n {
			continue
		}
		for _, p := range s.Issue.Positions {
			if pn := terms.Normalize(p.Id); pn != n && set[pn] {
				return true
			}
		}
	}
	return false
}

func (l Label) String() string {
	switch l {
	case In:
		return "in"
	case Out:
		return "out"
	default:
		return "undecided"
	}
}

func NewLabelling() Labelling {
	return Labelling(make(map[*Statement]Label))
}

func SliceToMap(s []string) map[string]bool {
	result := map[string]bool{}
	for _, k := range s {
		result[k] = true
	}
	return result
}

// Initialize a labelling by making all assumptions In
// other positions of each issue with an assumption Out,
// and unassumed statements without arguments Out.
func (l Labelling) init(ag *ArgGraph) {
	// Normalize the assumptions
	s := []string{}
	for _, k := range ag.Assumptions {
		s = append(s, terms.Normalize(k))
	}
	ag.Assumptions = s

	// Map representation of the assumptions
	// to make membership test efficient; a questioned default is not
	// assumed (contract 5)
	ag.assums = SliceToMap(ag.Assumed())

	// Normalize the keys of the statements table
	m2 := map[string]*Statement{}
	for k, v := range ag.Statements {
		m2[terms.Normalize(k)] = v
	}

	ag.Statements = m2

	// Make all assumed statements In
	for _, s := range ag.Statements {
		if ag.assums[terms.Normalize(s.Id)] {
			l[s] = In
		}
	}
	// For each issue, if some position is In
	// make the undecided positions Out
	// The resulting issue may be inconsistent, with
	// multiple positions being In, if the assumptions are
	// inconsistent.
	for _, i := range ag.Issues {
		// is some position in?
		somePositionIn := false
		for _, p := range i.Positions {
			if l[p] == In {
				somePositionIn = true
				break
			}
		}
		if somePositionIn {
			for _, p := range i.Positions {
				if l[p] == Undecided {
					l[p] = Out
				}
			}
		}
	}
}

// Apply a labelling to an argument graph by setting
// the label property of each statement in the graph to
// its label in the labelling and by setting the weight
// of each argument in the graph to its evaluated weight
// in the labeling.
func (ag ArgGraph) ApplyLabelling(l Labelling) {
	for _, s := range ag.Statements {
		s.Label = l[s]
	}
	for _, arg := range ag.Arguments {
		arg.Weight = arg.GetWeight(l)
	}
}

// Returns In if the argument has been undercut, Out if the argument
// has no undercutter, the undercutter has no arguments,
// or attempts to undercut the argument it have failed, and Undecided otherwise
func (arg *Argument) Undercut(l Labelling) Label {
	if arg.Undercutter == nil {
		return Out // because there is no undercutter
	} else {
		return l[arg.Undercutter]
	}
}

// An argument is applicable if its Undercut property is In, or the undercutter
// is Out and none of the premises are Undecided. Because arguments can be
// cumulative, arguments with Out premises can be applicable. Out premises or
// In undercutters affect the weight of an argument, not its applicability.
// The labels of the premises are irrelevant for applicability if the
// undercutter is In.
func (arg *Argument) Applicable(l Labelling) bool {
	if arg.Undercut(l) == In {
		return true
	}
	if arg.Undercut(l) == Undecided {
		return false
	}
	for _, p := range arg.Premises {
		if l[p.Stmt] == Undecided {
			return false
		}
	}
	return true
}

func (arg *Argument) PropertyValue(p string, l Labelling) (result terms.Term, ok bool) {
	for _, pr := range arg.Premises {
		term1, ok := terms.ReadFormula(pr.Stmt.Id)
		if ok {
			p2, ok := terms.Predicate(term1)
			if ok && p2 == p {
				if l[pr.Stmt] == In {
					return terms.Object(term1)
				} else {
					i := pr.Stmt.Issue
					if i != nil {
						for _, pos := range i.Positions {
							if l[pos] == In {
								term2, ok := terms.ReadFormula(pos.Id)
								if ok {
									return terms.Object(term2)
								} else {
									fmt.Fprintf(os.Stderr, "Could not parse: %s\n", pos.Id)
									return result, false
								}
							}
						}
					}
				}
			}
		}
	}
	return result, false
}

// An issue is ready to be resolved if all the arguments of all its positions are
// undercut or applicable
func (issue *Issue) ReadyToBeResolved(l Labelling) bool {
	for _, position := range issue.Positions {
		for _, arg := range position.Args {
			if !(arg.Undercut(l) == In || arg.Applicable(l)) {
				return false
			}
		}
	}
	return true
}

// Apply a proof standard to check whether w1 is strictly greater than
// w2, where w1 and w2 are argument weights
// Note: PE are indistinguishable in this new model
func (std Standard) greater(w1, w2 float64) bool {
	alpha := 0.5
	beta := 0.3
	switch std {
	case PE:
		return w1 > w2
	case CCE:
		return w1 > w2 && (w1-w2 > alpha)
	case BRD:
		return w1 > w2 && (w1-w2 > alpha) && w2 < beta
	default:
		return false
	}
}

// Apply the proof standard of an issue to each of its positions and update
// the labelling accordingly. After resolving the issue, at most
// one of its positions will be In and all the others will be Out.
// (No position will remain Undecided.) The issue is assumed to be ready to be
// resolved before this method is called.
func (issue *Issue) Resolve(l Labelling) {
	var maxArgWeight = make(map[*Statement]float64)
	for _, p := range issue.Positions {
		maxArgWeight[p] = 0.0
		for _, arg := range p.Args {
			w := arg.GetWeight(l)
			if w > maxArgWeight[p] {
				maxArgWeight[p] = w
			}
		}
	}
	var winner *Statement
PositionLoop:
	for _, p1 := range issue.Positions {
		if maxArgWeight[p1] == 0.0 {
			continue // the winner must be supported by at least one good argument
		}
		winner = p1 // assumption
		// look for another position which is at least as strong as p1
		for _, p2 := range issue.Positions {
			if p2 != p1 &&
				!issue.Standard.greater(maxArgWeight[p1], maxArgWeight[p2]) {
				winner = nil // found an alternative which is at least as good
				continue PositionLoop
			}
		}
		if winner != nil {
			break // winning position found
		}
	}
	// update the labels
	for _, p := range issue.Positions {
		if p == winner {
			l[p] = In
		} else {
			l[p] = Out
		}
	}
}

// A argument has 0.0 weight if it is undercut or inapplicable.
// Otherwise, if a scheme has been applied, it is the weight assigned by
// the evaluator of the scheme.  Otherwise it is the weight assigned
// by the default evaluator, LinkedArgument.
func (arg *Argument) GetWeight(l Labelling) float64 {
	if arg.Undercut(l) == In || !arg.Applicable(l) {
		return 0.0
	} else if arg.Scheme != nil {
		return arg.Scheme.Weight(arg, l)
	} else {
		// apply the default weighing function
		return LinkedWeighingFunction(arg, l)
	}
}

// A statement is supported if it is the conclusion of at least one
// applicable argument with weight greater than 0.0.
func (stmt *Statement) Supported(l Labelling) bool {
	for _, arg := range stmt.Args {
		if arg.Applicable(l) && arg.GetWeight(l) > 0 {
			return true
		}
	}
	return false
}

// A statement is unsupported if it has no arguments or
// all of its arguments are applicable and weigh 0.0 or less
func (stmt *Statement) Unsupported(l Labelling) bool {
	for _, arg := range stmt.Args {
		if !arg.Applicable(l) || arg.GetWeight(l) > 0 {
			return false
		}
	}
	return true
}

// Returns the grounded labelling of an argument graph.
// The argument graph is not modified.
func (ag *ArgGraph) GroundedLabelling() Labelling {
	l := NewLabelling()
	l.init(ag)
	var changed bool
	for {
		changed = false // assumption
		// Try to label Undecided statements
		for _, stmt := range ag.Statements {
			if l[stmt] == Undecided {
				if stmt.Unsupported(l) {
					// make unsupported statements Out
					l[stmt] = Out
					changed = true
				} else if stmt.Issue == nil && stmt.Supported(l) {
					// make supported nonissues In
					l[stmt] = In
					changed = true
				} else if stmt.Issue != nil && stmt.Issue.ReadyToBeResolved(l) {
					// Apply proof standards to label the positions of issues
					// ready to be resolved. Only a label that actually changed
					// counts as progress: otherwise a statement the issue does
					// not list would keep the loop running for ever.
					before := positionLabels(stmt.Issue, l)
					stmt.Issue.Resolve(l)
					if positionLabels(stmt.Issue, l) != before {
						changed = true
					}
				}
			}
		}
		// return if a fixpoint has been found
		if !changed {
			return l
		}
	}
}

// positionLabels summarizes the labels of the positions of an issue, to see
// whether resolving the issue changed any of them.
func positionLabels(issue *Issue, l Labelling) string {
	s := make([]byte, 0, len(issue.Positions))
	for _, p := range issue.Positions {
		s = append(s, byte('0'+int(l[p])))
	}
	return string(s)
}

// An argument graph is inconsistent if it assumes more than one position of
// some issue, or both a statement and its negation (docs/contracts.md,
// contract 4). The answer is computed from the assumptions as they are now,
// not from the index that labelling builds.
func (ag *ArgGraph) inconsistentPositions() bool {
	assumed := map[string]bool{}
	for _, a := range ag.Assumed() {
		assumed[a] = true
	}
	for _, issue := range ag.Issues {
		found := false
		for _, p := range issue.Positions {
			if assumed[terms.Normalize(p.Id)] {
				if found {
					return true
				}
				found = true
			}
		}
	}
	for a := range assumed {
		if t, ok := terms.ReadFormula(a); ok {
			if c, ok := t.(terms.Compound); ok && c.Functor == "¬" && len(c.Args) == 1 && assumed[c.Args[0].String()] {
				return true
			}
		}
	}
	return false
}

// substituteParameters replaces each variable of t named in params by its
// value, once: the values are not substituted again.
func substituteParameters(t terms.Term, params map[string]terms.Term) terms.Term {
	switch x := t.(type) {
	case terms.Variable:
		if v, ok := params[x.Name]; ok {
			return v
		}
		return x
	case terms.Compound:
		args := make([]terms.Term, len(x.Args))
		for i, a := range x.Args {
			args[i] = substituteParameters(a, params)
		}
		return terms.Compound{Functor: x.Functor, Id: x.Id, Prio: x.Prio, Args: args}
	case terms.List:
		l := make(terms.List, len(x))
		for i, a := range x {
			l[i] = substituteParameters(a, params)
		}
		return l
	}
	return t
}

// negationText words the negation of a statement whose text is inner (or,
// without a text, whose term is raw).
func negationText(inner, raw string) string {
	inner = strings.TrimSpace(inner)
	if inner == "" || inner == raw {
		return "¬" + raw
	}
	// lower-case a sentence-initial article or pronoun, never a name
	if sp := strings.IndexByte(inner, ' '); sp > 0 {
		switch inner[:sp] {
		case "The", "A", "An", "There", "It", "This", "That", "These", "Those", "Some", "No", "All", "Every", "Each":
			inner = strings.ToLower(inner[:1]) + inner[1:]
		}
	}
	return "It is not the case that " + inner
}

// Apply a language to a term to construct a string,
// usually to represent the term in natural language.
func (l Language) Apply(term1 terms.Term) string {
	functor, ok := terms.Functor(term1)
	if !ok {
		return ""
	}
	arity := terms.Arity(term1)

	if arity == 0 {
		return l[functor+"/0"]
	}

	if arity > 0 {

		args := []interface{}{}
		for _, arg := range term1.(terms.Compound).Args {
			args = append(args, arg.String())
		}

		// Use Sprintf to instantiate the template of the functor in the
		// language spec
		spec := functor + "/" + strconv.Itoa(arity)
		template, ok := l[spec]
		if !ok {
			if functor == "¬" && arity == 1 {
				// the language applies inside a negation (E-NEGTEXT)
				return negationText(l.Apply(term1.(terms.Compound).Args[0]), args[0].(string))
			}
			fmt.Fprintf(os.Stderr, "Functor not defined in the language: %v\n", spec)
			return term1.String()
		}

		return fmt.Sprintf(template, args...)
	}

	return ""
}

func (ag *ArgGraph) InstantiateScheme(id string, parameters []string) {
	genArgId := func() string {
		prefix := "a"
		// Assume exisiting arguments have been given ids using the
		// system. Thus initializing i using the number of existing
		// arguments is likely to produce an unused id.
		i := len(ag.Arguments) + 1
		// increment i until no argument id with this index is already in use
		for _, ok := ag.Arguments[prefix+strconv.Itoa(i)]; ok; _, ok = ag.Arguments[prefix+strconv.Itoa(i)] {
			i++
		}
		return prefix + strconv.Itoa(i)
	}

	if ag.Theory != nil {
		scheme, ok := ag.Theory.schemeIndex[id]
		if !ok {
			scheme, ok = BasicSchemes[id]
		}
		if ok {
			// bind each schema variable to its value
			if len(scheme.Variables) != len(parameters) {
				fmt.Fprintf(os.Stderr, "Scheme formal (%v) and actual parameters (%v) do not match\n", scheme.Variables, parameters)
				return
			}

			// The parameters replace the scheme's variables once, all at the
			// same time: a value that holds a variable (a fresh one the rule
			// engine named A1, say) is not substituted again, so a scheme
			// variable that happens to be called A1 cannot capture it
			// (E-CAPTURE).
			bindings := map[string]terms.Term{}
			for i, varName := range scheme.Variables {
				// v := terms.NewVariable(varName)
				t, ok := terms.ReadFormula(parameters[i])
				if ok {
					bindings[varName] = t
				} else {
					fmt.Fprintf(os.Stderr, "Could not parse parameter: %v\n", parameters[i])
				}
			}

			// construct the premises and conclusions,
			// adding new statements to the graph
			premises := []Premise{}
			conclusions := []*Statement{}

			addPremises := func(l []string, assumptions bool) {
				for i, p := range l {
					term1, ok := terms.ReadFormula(p)
					if ok {
						term2 := terms.NormalizeNegation(substituteParameters(term1, bindings))
						// Leave argument(S,P) premises implicit; Enthymeme!
						pred, ok := terms.Predicate(term2)
						if ok && pred == "argument" {
							continue
						}
						stmt, ok := ag.Statements[term2.String()]
						if !ok {
							s := Statement{Id: term2.String(),
								Text: ag.Theory.Language.Apply(term2)}
							ag.Statements[term2.String()] = &s
							stmt = &s

						}
						if assumptions {
							ag.addDefault(term2.String())
						}
						role := ""
						if i < len(scheme.Roles) {
							role = scheme.Roles[i]
						}
						premises = append(premises, Premise{Role: role, Stmt: stmt})
					} else {
						fmt.Fprintf(os.Stderr, "Could not parse term: %v\n", p)
					}
				}
			}

			addPremises(scheme.Premises, false)
			addPremises(scheme.Deletions, false)
			// add the assumptions as additional premises
			addPremises(scheme.Assumptions, true)

			// instantiate the conclusions of the scheme
			for _, c := range scheme.Conclusions {
				term1, ok := terms.ReadFormula(c)
				if ok {
					term2 := terms.NormalizeNegation(substituteParameters(term1, bindings))
					stmt, ok := ag.Statements[term2.String()]
					if !ok {
						s := Statement{Id: term2.String(),
							Text: ag.Theory.Language.Apply(term2)}
						ag.Statements[term2.String()] = &s
						stmt = &s
					}
					conclusions = append(conclusions, stmt)
				} else {
					fmt.Fprintf(os.Stderr, "Could not parse term: %v\n", c)
				}
			}

			// construct an argument for each conclusion and add
			// the argument to the graph.  All these arguments
			// share the same undercutter.

			if len(conclusions) > 0 {
				var uc Statement    // the undercutter
				argId := genArgId() // pseudo-argument id, representing the set of all arguments constructed by the scheme
				// To Do: this is a hack.  Clean this up by allowing arguments
				// to have multiple conclusions. This is a fairly big change, requiring
				// modifications to the import and export translators and the YAML and JSON representation of arguments.

				// Construct the undercutter statement and
				// add it to the statements of the graph
				ucid := "undercut(" + argId + ")"
				uc = Statement{Id: ucid,
					Text:          argId + " is undercut.",
					IsUndercutter: true}
				ag.Statements[terms.Normalize(ucid)] = &uc
				for _, c := range conclusions {
					// Construct an argument for each conclusion and add it to the graph
					argId := genArgId()
					arg := Argument{Id: argId,
						Scheme:      scheme,
						Parameters:  parameters,
						Premises:    premises,
						Undercutter: &uc,
						Conclusion:  c}
					ag.Arguments[argId] = &arg
					c.Args = append(c.Args, &arg)
				}

				// instantiate the exceptions of the scheme
				exceptions := []*Statement{}
				for _, e := range scheme.Exceptions {
					term1, ok := terms.ReadFormula(e)
					if ok {
						term2 := terms.NormalizeNegation(substituteParameters(term1, bindings))
						stmt, ok := ag.Statements[term2.String()]
						if !ok {
							s := Statement{Id: term2.String(),
								Text: ag.Theory.Language.Apply(term2)}
							ag.Statements[term2.String()] = &s
							stmt = &s
						}
						exceptions = append(exceptions, stmt)
					} else {
						fmt.Fprintf(os.Stderr, "Could not parse term: %v\n", e)
					}
				}

				// construct an undercutting argument for each exception
				// and add it to the argument graph
				for _, e := range exceptions {
					argId := genArgId()

					// Construct an undercutter statement (for the undercutter of
					// undercutter!) and add it to the statements of the graph

					ucid := "undercut(" + argId + ")"
					uc2 := Statement{Id: ucid,
						Text: argId + " is undercut."}
					ag.Statements[terms.Normalize(ucid)] = &uc2

					// Construct the argument and add it to the graph
					arg := Argument{Id: argId,
						Premises:    []Premise{Premise{Stmt: e}},
						Undercutter: &uc2,
						Conclusion:  &uc}
					ag.Arguments[argId] = &arg
					uc.Args = append(uc.Args, &arg)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "No scheme with this id: %v\n", id)
		}
	}
}

// AddAssumption explicitly stipulates a statement, including one previously
// introduced as a scheme default. Questioning does not withdraw a stipulation.
// Clear aliases too: an API client may not yet have normalized its graph.
func (ag *ArgGraph) AddAssumption(stmt_id string) {
	n := terms.Normalize(stmt_id)
	for id := range ag.Defaults {
		if terms.Normalize(id) == n {
			delete(ag.Defaults, id)
		}
	}
	ag.appendAssumption(n)
}

// appendAssumption records an already normalized assumption without changing
// its origin. Scheme defaults must not pass through the stipulating API.
func (ag *ArgGraph) appendAssumption(id string) {
	if ag.assums == nil {
		ag.assums = make(map[string]bool)
	}
	ag.assums[id] = true
	ag.Assumptions = append(ag.Assumptions, id)
}

// addDefault assumes a statement because a scheme lists it under its
// assumptions, unless the graph already assumes it, and records it as a
// default (contract 5).
func (ag *ArgGraph) addDefault(stmt_id string) {
	n := terms.Normalize(stmt_id)
	for _, a := range ag.Assumptions {
		if terms.Normalize(a) == n {
			return
		}
	}
	if ag.Defaults == nil {
		ag.Defaults = map[string]bool{}
	}
	ag.Defaults[n] = true
	ag.appendAssumption(n)
}

// makeIssue: match the patterns of an issue scheme against the
// statements of the argument graph.  If more than one statement
// matches, make them positions of an issue, creating the issue
// if one does not already exist and adding it to the argument graph.
// Every statement may be a position of at most one issue.  No statement
// is made a position of some issue if this constraint would be violated.
// If some pattern is not synatically correct and thus cannot be parsed,
// an error is returned and the argument graph is left unchanged.
// If all goes well, the argument graph is updated and nil is returned
// issueLookup lets issue construction look positions up by key when the
// remaining patterns are ground; tests turn it off to compare with the scan.
var issueLookup = true

// readIssuePatterns rejects the complete pattern list before any graph write.
func readIssuePatterns(patterns []string) ([]terms.Term, error) {
	if len(patterns) < 2 {
		return nil, fmt.Errorf("fewer than two patterns")
	}
	parsed := make([]terms.Term, len(patterns))
	for i, p := range patterns {
		if i == 1 && len(patterns) == 3 && p == "..." {
			continue
		}
		t, ok := terms.ReadFormula(p)
		if !ok {
			return nil, fmt.Errorf("cannot parse pattern %q", p)
		}
		parsed[i] = t
	}
	return parsed, nil
}

func (ag *ArgGraph) makeIssue(issueScheme string, patterns []string) error {
	return ag.makeIssueWithWork(issueScheme, patterns, nil)
}

func (ag *ArgGraph) makeIssueWithWork(issueScheme string, patterns []string, work *issueWork) (err error) {
	parsed, err := readIssuePatterns(patterns)
	if err != nil {
		return err
	}
	pattern := parsed[0]
	// Use a stable traversal for anchors, indexed aliases and fallback scans.
	// Candidate bindings below are local; sorting must not conceal a choice of
	// only the first value of a variable introduced by a later pattern.
	wffs := make([]string, 0, len(ag.Statements))
	for wff := range ag.Statements {
		wffs = append(wffs, wff)
	}
	sort.Strings(wffs)
	// API-created maps may use valid formula keys with noncanonical spacing.
	// Index their parsed forms without renaming keys or replacing graph objects.
	// Keep all aliases, then use Match below to preserve the scan's typed equality.
	type issueEntry struct {
		key  string
		term terms.Term
		stmt *Statement
	}
	byTerm := map[string][]issueEntry{}
	if issueLookup && !(len(patterns) == 3 && patterns[1] == "...") {
		for _, key := range wffs {
			stmt := ag.Statements[key]
			if err := work.step(); err != nil {
				return err
			}
			if term, ok := terms.ReadFormula(key); ok {
				byTerm[term.String()] = append(byTerm[term.String()], issueEntry{key, term, stmt})
			}
		}
	}
	for _, wff1 := range wffs {
		stmt := ag.Statements[wff1]
		if err := work.step(); err != nil {
			return err
		}
		term1, ok := terms.ReadFormula(wff1)
		if !ok {
			fmt.Fprintf(os.Stderr, "Statement key not a term: %v\n", wff1)
			continue
		}
		var bindings terms.Bindings
		bindings, ok = terms.MatchFormula(pattern, term1, bindings)
		if !ok {
			continue // terms do not match
		} else {
			candidates := []*Statement{stmt}

			// Check if the issue scheme defines an enumeration.
			isEnumeration := len(patterns) == 3 && patterns[1] == "..."

			// Create a copy of bindings with all variables with names ending
			// in integer indexes unbound
			var bindings2 terms.Bindings
			for env := bindings; env != nil; env = env.Next {
				v := env.Var
				t := env.T
				suffix := v.Name[len(v.Name)-1:]
				_, err := strconv.Atoi(suffix)
				if err != nil {
					// the variable does not end with an integer suffix
					// so keep its binding
					bindings2 = terms.AddBinding(v, t, bindings2)
				}
			}

			// When every remaining pattern is ground under the bindings of
			// the first (as "¬S" is once S is bound), its positions are
			// looked up by key instead of found by a scan of all statements.
			// The scan made issue construction quadratic in the size of the
			// graph, so that a graph of some 800 statements exhausted the
			// issue-visit budget (docs/contracts.md, contract 3).
			// Only a ground statement gives ground bindings: statement
			// keys with variables keep the scan, which matches them.
			ground := !isEnumeration && issueLookup && terms.Ground(term1, nil)
			keys := []string{}
			if ground {
				for _, pattern2 := range parsed[1:] {
					inst := terms.Substitute(pattern2, bindings)
					if !terms.Ground(inst, nil) {
						ground = false
						break
					}
					keys = append(keys, terms.Normalize(inst.String()))
				}
			}
			if ground {
				seen := map[string]bool{wff1: true}
				for j, k := range keys {
					if err := work.step(); err != nil {
						return err
					}
					for _, entry := range byTerm[k] {
						if err := work.step(); err != nil {
							return err
						}
						if seen[entry.key] {
							continue
						}
						if _, match := terms.MatchFormula(parsed[j+1], entry.term, bindings); match {
							seen[entry.key] = true
							candidates = append(candidates, entry.stmt)
						}
					}
				}
			}

			// Otherwise, for each matching statement, iterate over the
			// statements again to try to find other positions of the issue.
			// Whether or not a statement is a position depends on the
			// remaining patterns of the issue scheme and, in particular,
			// whether or not the issue scheme is an enumeration.

			for _, wff2 := range wffs {
				if ground {
					break
				}
				stmt2 := ag.Statements[wff2]
				if err := work.step(); err != nil {
					return err
				}
				if wff2 == wff1 {
					// skip the matching statement found previously
					continue
				}
				term2, ok := terms.ReadFormula(wff2)
				if !ok {
					fmt.Fprintf(os.Stderr, "Statement key not a term: %v\n", wff2)
					continue
				}

				match := false
				if !isEnumeration {
					// Each remaining pattern supplies alternative positions
					// under the first pattern's anchor bindings. A candidate's
					// fresh bindings must not constrain the next candidate or
					// another alternative pattern. Repeated variables within
					// one pattern are still checked by that single match.
					for _, pattern2 := range parsed[1:] {
						if err := work.step(); err != nil {
							return err
						}
						_, match = terms.MatchFormula(pattern2, term2, bindings)
						if match {
							break
						}
					}
				} else {
					// Use a fresh copy of bindings2 for enumeration issue patterns
					var b2copy terms.Bindings
					for env := bindings2; env != nil; env = env.Next {
						k := env.Var
						v := env.T
						b2copy = terms.AddBinding(k, v, b2copy)
					}
					b2copy, match = terms.MatchFormula(pattern, term2, b2copy)
				}
				if !match {
					continue // terms do not match
				} else {
					candidates = append(candidates, stmt2)
				}
			}

			// Check whether any of the candidates found are already at issue
			// and, if so, that they are all positions of the same issue.
			// No statement may be a position of more than one issue.
			// Add candidates which do not violate the single issue constraint
			// to the list of positions.
			var issue *Issue
			var positions = []*Statement{}
			for _, c := range candidates {
				if c.Issue != nil {
					if issue == nil {
						issue = c.Issue
					} else if c.Issue != issue {
						// found a conflict, due to statements being positions of different issues
						return fmt.Errorf("statement %s connects distinct existing issues; explicit issue resolution is required", c.Id)
					}
				}
				positions = append(positions, c)
			}

			// Do not make an issue when there are less than two positions
			if len(positions) < 2 {
				continue
			}

			// Add the statements which are not at issue to the existing
			// issue, if any, or create a new issue and add all the statements
			// found as positions of the new issue.
			if issue == nil {
				// create a new issue
				issue = NewIssue()
				issue.Positions = positions
				for _, pos := range positions {
					pos.Issue = issue
				}
				// generate an id for the new issue
				i := len(ag.Issues) + 1
				prefix := "i"
				id := prefix + strconv.Itoa(i)
				_, existing := ag.Issues[id]
				for existing {
					i++
					id = prefix + strconv.Itoa(i)
					_, existing = ag.Issues[id]
				}
				// add the new issue to the argument graph
				issue.Id = id
				ag.Issues[id] = issue
			} else {
				// add new positions to the existing issue: link both ways,
				// so that labelling resolves them with the other positions
				for _, pos := range positions {
					if pos.Issue == nil {
						pos.Issue = issue
						issue.Positions = append(issue.Positions, pos)
					}
				}
			}
		}
	}
	return err
}
