package game

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// Say renders a statement in English for players, through the Language
// templates of the scheme file:
//   - the words players typed stand in single quotes, in place of the atoms
//     they became (‘carbon tax’ for carbon_tax), so that the scheme's words
//     and the player's can be told apart and fragments still read;
//   - a statement inside a statement stands in double quotes; one nested
//     deeper stands in square brackets, shortened;
//   - a negation reads through the "¬p/n" template of its predicate where
//     the file has one ("… is not a matter of …"), else as "it is not the
//     case that …".
//
// Statement IDs stay the reference; this is only how they read.
func (gs *GameState) Say(stmtID string) string {
	return sentence(gs.say(stmtID, gs.words(), 0))
}

// SayIn is Say for a statement shown with an argument: the argument's
// conclusion, wherever it occurs inside the statement, reads "C" (the card
// shows the conclusion as C), so premises such as "W asserts that C" and "C
// is a matter of D" do not repeat it.
func (gs *GameState) SayIn(argID, stmtID string) string {
	words := gs.words()
	if a := gs.Adapter.Graph.Arguments[argID]; a != nil && a.Conclusion != nil {
		words = withSelf(words, a.Conclusion.Id)
	}
	return sentence(gs.say(stmtID, words, 0))
}

// SayValue renders the value of a scheme variable (a term) as a part of a
// statement reads: the player's words in single quotes, a statement in
// double quotes.
func (gs *GameState) SayValue(v string) string {
	t, ok := terms.ReadString(v)
	if !ok {
		return "‘" + v + "’"
	}
	return gs.sayArg(t, gs.words(), 0)
}

// selfKey marks, in a words map, the statement that reads "C".
const selfKey = "\x00self"

func withSelf(words map[string]string, stmtID string) map[string]string {
	w := make(map[string]string, len(words)+1)
	for k, v := range words {
		w[k] = v
	}
	w[selfKey] = canonical(stmtID)
	return w
}

// canonical is a statement ID in the form terms print it, negation as a
// leading ¬, so that a statement, the same statement nested in another, and
// "not(p)" for "¬p" compare equal.
func canonical(id string) string {
	neg := false
	for strings.HasPrefix(id, "¬") {
		id, neg = strings.TrimPrefix(id, "¬"), !neg
	}
	t, ok := terms.ReadString(id)
	if !ok {
		if neg {
			return "¬" + id
		}
		return id
	}
	for {
		c, ok := t.(terms.Compound)
		if !ok || c.Functor != "not" || len(c.Args) != 1 {
			break
		}
		t, neg = c.Args[0], !neg
	}
	if neg {
		return "¬" + t.String()
	}
	return t.String()
}

// sentence capitalizes the first letter (after any opening quotes) and ends
// the sentence with a period.
func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	for i, r := range s {
		if r == '‘' || r == '“' || r == '"' || r == '\'' {
			continue
		}
		s = s[:i] + string(unicode.ToUpper(r)) + s[i+utf8.RuneLen(r):]
		break
	}
	if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "?") && !strings.HasSuffix(s, "!") {
		s += "."
	}
	return s
}

// maxNestedWords is how many words of a statement nested two levels deep
// are shown.
const maxNestedWords = 8

func (gs *GameState) say(id string, words map[string]string, depth int) string {
	if self, ok := words[selfKey]; ok && depth > 0 && canonical(id) == self {
		return "C"
	}
	if rest := strings.TrimPrefix(id, "¬"); rest != id {
		if t, ok := terms.ReadString(rest); ok {
			if out, ok := gs.negative(t, words, depth); ok {
				return out
			}
		}
		return "it is not the case that " + lowerFirst(gs.say(rest, words, depth))
	}
	if text := gs.freeText(id); text != "" {
		return strings.TrimSuffix(text, ".")
	}
	t, ok := terms.ReadString(id)
	if !ok || strings.TrimSpace(id) == "" {
		if s := gs.Adapter.Graph.Statements[id]; s != nil && s.Text != "" {
			return s.Text
		}
		return id
	}
	return gs.sayTerm(t, words, depth)
}

// negative renders the negation of t through the "¬p/n" template of its
// predicate, if the Language has one.
func (gs *GameState) negative(t terms.Term, words map[string]string, depth int) (string, bool) {
	c, ok := t.(terms.Compound)
	if !ok {
		return "", false
	}
	tmpl := gs.template("¬"+c.Functor, len(c.Args))
	if tmpl == "" {
		return "", false
	}
	return gs.fill(tmpl, c.Args, words, depth), true
}

func (gs *GameState) template(functor string, arity int) string {
	if th := gs.Adapter.Graph.Theory; th != nil && th.Language != nil {
		if tmpl := th.Language[fmt.Sprintf("%s/%d", functor, arity)]; strings.Count(strings.ReplaceAll(tmpl, "%%", ""), "%s") == arity {
			return tmpl
		}
	}
	return ""
}

// fill instantiates a template with the arguments of a compound: values in
// single quotes, statements in double quotes (see sayArg).
func (gs *GameState) fill(tmpl string, args []terms.Term, words map[string]string, depth int) string {
	vals := make([]interface{}, len(args))
	for i, a := range args {
		vals[i] = gs.sayArg(a, words, depth)
	}
	out := strings.TrimSuffix(fmt.Sprintf(tmpl, vals...), ".")
	if depth > 0 && !strings.HasPrefix(tmpl, "%s") { // "Doing %s ..." within a sentence
		out = lowerFirst(out)
	}
	return out
}

// sayArg renders an argument of a statement: a value as the player's words
// in single quotes, a statement in double quotes.
func (gs *GameState) sayArg(a terms.Term, words map[string]string, depth int) string {
	s := a.String()
	if self, ok := words[selfKey]; ok && canonical(s) == self {
		return "C"
	}
	if _, isCompound := a.(terms.Compound); isCompound || gs.freeText(s) != "" {
		inner := gs.say(s, words, depth+1)
		if depth+1 >= 2 { // [a shortened statement] inside a “quoted” one
			return "[" + shorten(inner, maxNestedWords) + "]"
		}
		return "“" + inner + "”"
	}
	if w, ok := words[s]; ok {
		return "‘" + w + "’"
	}
	return "‘" + strings.ReplaceAll(s, "_", " ") + "’"
}

func (gs *GameState) sayTerm(t terms.Term, words map[string]string, depth int) string {
	c, ok := t.(terms.Compound)
	if !ok {
		s := t.String()
		if text := gs.freeText(s); text != "" {
			return strings.TrimSuffix(text, ".")
		}
		if w, ok := words[s]; ok {
			return w
		}
		return strings.ReplaceAll(s, "_", " ")
	}
	switch {
	case c.Functor == "not" && len(c.Args) == 1:
		if out, ok := gs.negative(c.Args[0], words, depth); ok {
			return out
		}
		return "it is not the case that " + lowerFirst(gs.sayTerm(c.Args[0], words, depth))
	case c.Functor == "undercut" && len(c.Args) == 1:
		return fmt.Sprintf("argument %s does not apply", c.Args[0].String())
	}
	tmpl := gs.template(c.Functor, len(c.Args))
	if tmpl == "" {
		parts := make([]string, len(c.Args))
		for i, a := range c.Args {
			parts[i] = gs.sayArg(a, words, depth)
		}
		return c.Functor + "(" + strings.Join(parts, ", ") + ")"
	}
	return gs.fill(tmpl, c.Args, words, depth)
}

// shorten keeps the first n words of s, closing any quotes it cuts.
func shorten(s string, n int) string {
	f := strings.Fields(s)
	if len(f) <= n {
		return s
	}
	out := strings.Join(f[:n], " ")
	if strings.Count(out, "‘") > strings.Count(out, "’") {
		out += "’"
	}
	if strings.Count(out, "“") > strings.Count(out, "”") {
		out += "”"
	}
	return out + " …"
}

// freeText returns the text of a free-text statement (of the user-defined
// scheme, or imported), whose ID is a hash of its text; "" for others.
func (gs *GameState) freeText(id string) string {
	if !strings.HasPrefix(id, "text_") && !strings.HasPrefix(id, "stmt_") {
		return ""
	}
	if s := gs.Adapter.Graph.Statements[id]; s != nil {
		return s.Text
	}
	return ""
}

// words maps the atoms that players' bindings became to the words they typed
// (the first spelling wins), from the accepted moves.
func (gs *GameState) words() map[string]string {
	w := map[string]string{}
	for _, m := range gs.MoveHistory {
		if len(m.ArgIDs) == 0 {
			continue
		}
		for _, v := range carneades.SortedKeys(m.Action.VariableBindings) {
			typed := strings.TrimSpace(m.Action.VariableBindings[v])
			if atom := carneades.SanitizeBinding(typed); atom != typed && atom != "" {
				if _, ok := w[atom]; !ok {
					w[atom] = typed
				}
			}
		}
	}
	return w
}

// SchemeName is the name of the NASCQ scheme an argument was made with, or
// what the argument is if it has none.
func (gs *GameState) SchemeName(arg *caes.Argument) string {
	if arg == nil {
		return ""
	}
	switch carneades.ArgumentRole(arg) {
	case carneades.RoleQuestion:
		return "question"
	case carneades.RolePresumption:
		return "presumption"
	case carneades.RoleException:
		return "exception"
	}
	if arg.Scheme == nil {
		return ""
	}
	if s := gs.NASCQ.GetScheme(gs.NASCQ.FindNASCQSchemeID(arg.Scheme.Id)); s != nil {
		return s.Name
	}
	return arg.Scheme.Id
}

// ownerOf returns the owner of an argument, or "" if it has none.
func ownerOf(arg *caes.Argument) string {
	if arg == nil {
		return ""
	}
	o, _ := arg.Metadata["owner"].(string)
	return o
}

// BoardStatements returns the statements that a move about a statement can
// argue for or against, sorted: the claim, and the conclusions and premises
// of the arguments that moves made. Undercutters, the premises of exception
// arguments, and ¬S when S is listed too (an argument against ¬S is one for
// S) are left out.
func (gs *GameState) BoardStatements() []string {
	set := map[string]bool{}
	if gs.Claim != "" {
		set[gs.Claim] = true
	}
	for _, arg := range gs.Adapter.Graph.Arguments {
		if r := carneades.ArgumentRole(arg); r != carneades.RoleMain && r != "" {
			continue
		}
		if c := arg.Conclusion; c != nil && !c.IsUndercutter {
			set[c.Id] = true
		}
		for _, p := range arg.Premises {
			if p.Stmt != nil && !p.Stmt.IsUndercutter {
				set[p.Stmt.Id] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		if rest, neg := strings.CutPrefix(id, "¬"); !neg || !set[rest] {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// Effect says in English what a move made from the template would establish
// about its target (or, for a plain support, what it concludes).
func (gs *GameState) Effect(tmpl nascq.ActionTemplate, bindings map[string]string) string {
	if tmpl.TargetStmtID != "" {
		return gs.answerText(tmpl, nil) + "."
	}
	if tmpl.TargetArgID == "" {
		return nascq.InstantiateTemplate(tmpl.ConclusionTemplateEN, bindings)
	}
	target := gs.Adapter.Graph.Arguments[tmpl.TargetArgID]
	if target == nil || target.Conclusion == nil {
		return "(target not found)"
	}
	answer := gs.answerText(tmpl, target)
	switch tmpl.Type + "/" + tmpl.Locus.Kind {
	case "question/premise":
		return fmt.Sprintf("%s is no longer taken for granted until %s answers.", gs.sayPlain(tmpl.LocusStmtID), ownerOf(target))
	case "question/argument":
		return fmt.Sprintf("%s is out of force until %s answers.", tmpl.TargetArgID, ownerOf(target))
	case "attack/premise":
		return fmt.Sprintf("%s (against a premise of %s).", answer, tmpl.TargetArgID)
	case "support/premise":
		return fmt.Sprintf("%s (backing a premise of %s).", answer, tmpl.TargetArgID)
	case "attack/exception":
		return fmt.Sprintf("%s, so %s does not apply.", answer, tmpl.TargetArgID)
	case "support/exception":
		return fmt.Sprintf("%s (denying an exception to %s).", answer, tmpl.TargetArgID)
	}
	return answer + "."
}

// answerText is the statement a targeted move from the template argues for
// (its conclusion), without a final period.
func (gs *GameState) answerText(tmpl nascq.ActionTemplate, target *caes.Argument) string {
	attack := tmpl.Type == "attack"
	switch tmpl.Locus.Kind {
	case nascq.LocusStatement, nascq.LocusPremise:
		if attack {
			return gs.sayPlain(negation(tmpl.LocusStmtID))
		}
		return gs.sayPlain(tmpl.LocusStmtID)
	case nascq.LocusException:
		if attack {
			return gs.sayPlain(tmpl.LocusStmtID)
		}
		return gs.sayPlain(negation(tmpl.LocusStmtID))
	case nascq.LocusConclusion:
		if attack {
			return gs.sayPlain(negation(target.Conclusion.Id))
		}
		return gs.sayPlain(target.Conclusion.Id)
	case nascq.LocusArgument:
		if attack {
			return fmt.Sprintf("%s's reasons do not support its conclusion", tmpl.TargetArgID)
		}
		return fmt.Sprintf("%s's reasons do support its conclusion", tmpl.TargetArgID)
	}
	return tmpl.CaesAttackType + " of " + tmpl.TargetArgID
}

// negation is the ID of the statement that negates id.
func negation(id string) string {
	if rest := strings.TrimPrefix(id, "¬"); rest != id {
		return rest
	}
	return "¬" + id
}

// sayPlain is Say without the final period.
func (gs *GameState) sayPlain(stmtID string) string {
	return strings.TrimSuffix(gs.Say(stmtID), ".")
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return s
	}
	return string(unicode.ToLower(r)) + s[n:]
}

// DescribeMove is a one-line account of a recorded move, for logs and the CLI.
func (gs *GameState) DescribeMove(m Move) string {
	a := m.Action
	who := m.PlayerID
	switch {
	case a.Type == "decide":
		verb := map[string]string{"accepted": "accepts", "rejected": "rejects", "undecided": "withdraws its decision on"}[m.Decision]
		line := fmt.Sprintf("The moderator %s: %s", verb, gs.Say(m.DecisionStmt))
		if a.Text != "" {
			line += " (" + a.Text + ")"
		}
		return line
	case m.Moderation != "" && len(m.ArgIDs) == 0 && !m.Abstained:
		return fmt.Sprintf("%s's move was %s", who, m.Moderation)
	case m.Resigned:
		return who + " resigns"
	case m.TimedOut && m.Abstained:
		return who + " took too long: the turn passes"
	case m.TimedOut:
		return who + " ran out of time"
	case m.Abstained:
		return who + " passes"
	case len(m.ArgIDs) == 0:
		return fmt.Sprintf("%s's move was rejected", who)
	}
	main := gs.Adapter.Graph.Arguments[m.ArgIDs[0]]
	cq := ""
	if q := gs.NASCQ.GetCQ(a.AnswersCQID); q != nil {
		cq = "the " + q.Name
	}
	tag := fmt.Sprintf("[%s, %s]", gs.SchemeName(main), m.ArgIDs[0])
	concl := ""
	if main != nil && main.Conclusion != nil {
		concl = gs.Say(main.Conclusion.Id)
	}
	target := a.TargetArgID
	if t := gs.Adapter.Graph.Arguments[target]; t != nil && ownerOf(t) != "" && ownerOf(t) != who {
		target = ownerOf(t) + "'s " + target
	}
	switch {
	case a.TargetStmtID != "":
		dir := map[string]string{"support": "for", "attack": "against"}[a.Type]
		return fmt.Sprintf("%s argues %s “%s”: %s %s", who, dir, gs.sayPlain(a.TargetStmtID), concl, tag)
	case a.Type == "question":
		q := gs.QuestionText(a.AnswersCQID, a.TargetArgID)
		if a.TargetPremiseID != "" {
			q = gs.QuestionAbout(a.AnswersCQID, a.TargetArgID, a.TargetPremiseID)
		}
		return fmt.Sprintf("%s asks %s about %s: %s", who, ownerOf(gs.Adapter.Graph.Arguments[a.TargetArgID]), a.TargetArgID, q)
	case a.Type == "attack" && cq != "":
		return fmt.Sprintf("%s answers %s against %s: %s %s", who, cq, target, concl, tag)
	case a.Type == "attack":
		return fmt.Sprintf("%s attacks %s (%s): %s %s", who, target, attackWords(a.AttackType), concl, tag)
	case a.TargetArgID != "" && cq != "":
		return fmt.Sprintf("%s answers %s for %s: %s %s", who, cq, target, concl, tag)
	case a.TargetArgID != "":
		return fmt.Sprintf("%s backs %s: %s %s", who, target, concl, tag)
	}
	return fmt.Sprintf("%s argues: %s %s", who, concl, tag)
}

// attackWords names an attack type for players.
func attackWords(t string) string {
	switch t {
	case "rebuttal":
		return "against its conclusion"
	case "undercutter":
		return "against its inference"
	case "undermining":
		return "against a premise"
	}
	return t
}

// VariableOrder returns the template's variables in the order a player fills
// them in (see variableOrder): without the answer variable.
func VariableOrder(tmpl nascq.ActionTemplate) []string { return variableOrder(tmpl) }

// StandardName is the abbreviation of a proof standard (PE, CCE, BRD).
func StandardName(s caes.Standard) string { return standardName(s) }

// AnswerText is the statement a targeted move from the template argues for,
// in English (see Effect).
func (gs *GameState) AnswerText(tmpl nascq.ActionTemplate) string {
	return gs.answerText(tmpl, gs.Adapter.Graph.Arguments[tmpl.TargetArgID])
}
