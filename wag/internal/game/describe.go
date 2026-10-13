package game

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// Say renders a statement in English for players: through the Language
// templates of the scheme file, with nested statements quoted, ¬ as "it is
// not the case that", and the words players typed in place of the atoms they
// became ("carbon tax" for carbon_tax). Statement IDs stay the reference;
// this is only how they read.
func (gs *GameState) Say(stmtID string) string {
	s := gs.say(stmtID, gs.words(), false)
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	s = string(unicode.ToUpper(r)) + s[n:]
	if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "?") {
		s += "."
	}
	return s
}

func (gs *GameState) say(id string, words map[string]string, nested bool) string {
	if rest := strings.TrimPrefix(id, "¬"); rest != id {
		return "it is not the case that " + gs.say(rest, words, true)
	}
	t, ok := terms.ReadString(id)
	if !ok || strings.TrimSpace(id) == "" {
		if s := gs.Adapter.Graph.Statements[id]; s != nil && s.Text != "" {
			return s.Text
		}
		return id
	}
	return gs.sayTerm(t, words, nested)
}

func (gs *GameState) sayTerm(t terms.Term, words map[string]string, nested bool) string {
	c, ok := t.(terms.Compound)
	if !ok {
		s := t.String()
		if w, ok := words[s]; ok {
			return w
		}
		return strings.ReplaceAll(s, "_", " ")
	}
	var out string
	switch {
	case c.Functor == "not" && len(c.Args) == 1:
		out = "it is not the case that " + gs.sayTerm(c.Args[0], words, true)
	case c.Functor == "undercut" && len(c.Args) == 1:
		out = fmt.Sprintf("argument %s does not apply", c.Args[0].String())
	default:
		args := make([]interface{}, len(c.Args))
		for i, a := range c.Args {
			if _, isCompound := a.(terms.Compound); isCompound {
				args[i] = "“" + gs.sayTerm(a, words, true) + "”"
			} else {
				args[i] = gs.sayTerm(a, words, true)
			}
		}
		tmpl := ""
		if th := gs.Adapter.Graph.Theory; th != nil && th.Language != nil {
			tmpl = th.Language[fmt.Sprintf("%s/%d", c.Functor, len(c.Args))]
		}
		if tmpl == "" || strings.Count(tmpl, "%s") != len(args) {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = a.(string)
			}
			out = c.Functor + "(" + strings.Join(parts, ", ") + ")"
		} else {
			out = fmt.Sprintf(tmpl, args...)
			if nested && !strings.HasPrefix(tmpl, "%s") { // "Action %s ..." within a sentence
				out = lowerFirst(out)
			}
		}
	}
	if nested {
		out = strings.TrimSuffix(out, ".")
	}
	return out
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

// Effect says in English what a move made from the template would establish
// about its target (or, for a plain support, what it concludes).
func (gs *GameState) Effect(tmpl nascq.ActionTemplate, bindings map[string]string) string {
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
	not := func(s string) string { return "It is not the case that " + lowerFirst(s) }
	attack := tmpl.Type == "attack"
	switch tmpl.Locus.Kind {
	case nascq.LocusPremise, nascq.LocusException:
		stmt := gs.sayPlain(tmpl.LocusStmtID)
		if attack == (tmpl.Locus.Kind == nascq.LocusPremise) {
			return not(stmt)
		}
		return stmt
	case nascq.LocusConclusion:
		if attack {
			return not(gs.sayPlain(target.Conclusion.Id))
		}
		return gs.sayPlain(target.Conclusion.Id)
	case nascq.LocusArgument:
		concl := gs.sayPlain(target.Conclusion.Id)
		if attack {
			return fmt.Sprintf("%s does not support “%s”", tmpl.TargetArgID, concl)
		}
		return fmt.Sprintf("%s does support “%s”", tmpl.TargetArgID, concl)
	}
	return tmpl.CaesAttackType + " of " + tmpl.TargetArgID
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
		cq = q.Name
	}
	scheme := gs.SchemeName(main)
	concl := ""
	if main != nil && main.Conclusion != nil {
		concl = gs.Say(main.Conclusion.Id)
	}
	switch {
	case a.Type == "question":
		return fmt.Sprintf("%s asks %s about %s: %s", who, ownerOf(gs.Adapter.Graph.Arguments[a.TargetArgID]),
			a.TargetArgID, gs.QuestionText(a.AnswersCQID, a.TargetArgID))
	case a.Type == "attack" && cq != "":
		return fmt.Sprintf("%s attacks %s (%s, %s) with %s %s: %s", who, a.TargetArgID, cq, a.AttackType, scheme, m.ArgIDs[0], concl)
	case a.Type == "attack":
		return fmt.Sprintf("%s attacks %s (%s) with %s %s: %s", who, a.TargetArgID, a.AttackType, scheme, m.ArgIDs[0], concl)
	case a.TargetArgID != "" && cq != "":
		return fmt.Sprintf("%s defends %s (%s) with %s %s: %s", who, a.TargetArgID, cq, scheme, m.ArgIDs[0], concl)
	case a.TargetArgID != "":
		return fmt.Sprintf("%s defends %s with %s %s: %s", who, a.TargetArgID, scheme, m.ArgIDs[0], concl)
	}
	return fmt.Sprintf("%s argues %s with %s: %s", who, m.ArgIDs[0], scheme, concl)
}
