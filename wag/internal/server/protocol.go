package server

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// The text protocol: how any language model (or a person at a chat window)
// plays a seat. The server writes the game as a prompt in plain English, with
// the legal moves and the parts each scheme needs; the player replies in
// free text that ends with one JSON object, the move (ParseReply). The JSON
// is the CLI's and the API's nascq.ProposedAction. docs/protocol.md
// describes it for people.

// Prompt limits: a large game (an imported Kialo discussion) is cut to the
// arguments nearest the play.
const (
	promptArguments = 24 // arguments shown with their premises
	promptTargets   = 10 // arguments whose critical questions are offered
	promptHistory   = 30 // moves of the history shown
)

// Prompt is the text a player sees when it is their move.
func (ss *Session) Prompt() string {
	gs, st := ss.GS, ss.buildState()
	var b strings.Builder
	w := func(format string, args ...interface{}) { fmt.Fprintf(&b, format, args...) }
	me := gs.ActivePlayer()
	if me == nil || gs.Terminated {
		return "The game is over: " + gs.TerminationReason + "\n"
	}

	w("# WAG: the Waltonian Argumentation Game. It is your move.\n\n")
	w("You are %s. ", me.ID)
	switch side := gs.Side(me.ID); {
	case side == "pro":
		w("You argue FOR the claim: you win if the claim is IN when the game ends.")
	case side == "con":
		w("You argue AGAINST the claim: you win unless the claim is IN when the game ends.")
	default:
		w("You play for points.")
	}
	w(" Mode: %s. Round %d", gs.Config.Mode, gs.CurrentTurn)
	if gs.Config.MaxTurns > 0 {
		w(" of %d", gs.Config.MaxTurns)
	}
	w(". Scores:")
	for _, p := range st.Players {
		w(" %s %.1f;", p.ID, p.Score)
	}
	w("\n\n")

	w(`## How the game works
- Every argument is an instance of an argumentation scheme (Walton): you fill in the scheme's parts with short phrases, and the scheme's sentences (below) are built from them. Its premises hold until someone challenges them.
- Each scheme has critical questions. To ATTACK an argument, answer one of its questions against it; to DEFEND one, answer a question in its favour. Either way you back the answer with an argument of a scheme the question allows. The answer itself is filled in for you; you fill in the other parts.
- You may also ASK a question about a premise or the inference without answering it: the argument stops counting until its author answers. Or ARGUE a new argument, PASS or RESIGN.
- After every move the Carneades engine labels each statement IN, OUT or UNDECIDED.
- Make your parts true and relevant to the discussion, in plain words: the other player can challenge any of them. Cite sources when a scheme asks for them.

`)
	if gs.Claim != "" {
		w("## The claim\n%s It is now %s.\n\n", gs.Say(gs.Claim), strings.ToUpper(st.Statements[gs.Claim].Label))
	}

	// The arguments: the most recent, and those for or against the claim.
	ids := promptArgs(st, gs.Claim)
	if len(ids) == 0 {
		w("## The board\nNo arguments yet.\n\n")
	} else {
		w("## The arguments on the board\n")
		if n := countMain(st); n > len(ids) {
			w("(%d of %d arguments: the most recent, and those on the claim.)\n", len(ids), n)
		}
		for _, id := range ids {
			a := st.Arguments[id]
			force := "in force"
			if !a.InForce {
				force = "OUT OF FORCE: " + a.WhyOut
			}
			w("\n%s — %s's %s, weight %.2f, %s\n", id, a.Owner, a.SchemeName, a.BaseWeight, force)
			if a.Answers != nil {
				dir := "against"
				if a.Answers.Positive {
					dir = "for"
				}
				w("  It answers the %s %s %s.\n", a.Answers.CQName, dir, a.Answers.Target)
			}
			w("  C: %s [%s]\n", gs.Say(a.Conclusion), strings.ToUpper(st.Statements[a.Conclusion].Label))
			for i, p := range a.Premises {
				w("  premise: %s [%s]\n", textAt(a.PremiseTexts, i, p), strings.ToUpper(st.Statements[p].Label))
			}
			for i, e := range a.Exceptions {
				if st.Statements[e].Label == "in" {
					w("  exception that holds: %s\n", textAt(a.ExceptionTexts, i, e))
				}
			}
			if a.Text != "" {
				w("  In %s's words: “%s”\n", a.Owner, a.Text)
			}
			if len(a.Evidence) > 0 {
				w("  Sources: %s\n", strings.Join(a.Evidence, "; "))
			}
		}
		w("\n")
	}

	if len(st.Log) > 0 {
		w("## What happened\n")
		from := 0
		if len(st.Log) > promptHistory {
			from = len(st.Log) - promptHistory
			w("(the last %d moves)\n", promptHistory)
		}
		for _, l := range st.Log[from:] {
			w("%d. %s\n", l.Index+1, l.Text)
			for _, e := range l.Effects {
				w("   → %s\n", e)
			}
		}
		w("\n")
	}
	for _, q := range st.Questions {
		if q.Owner == me.ID {
			w("You must answer: %s asks about %s: %s\n\n", q.Asker, q.Target, q.Question)
		}
	}
	if gs.LastRejection != "" {
		w("Your last move was rejected: %s\n\n", gs.LastRejection)
	}

	// The critical questions the player can raise or answer.
	w("## Critical questions you can answer or ask\n")
	targets := promptTargetIDs(st, ids, me.ID)
	any := false
	for _, t := range targets {
		opts := moves(gs, t)
		if len(opts) == 0 {
			continue
		}
		any = true
		a := st.Arguments[t]
		w("\n%s (%s's): %s\n", t, a.Owner, gs.Say(a.Conclusion))
		for _, o := range opts {
			q := gs.NASCQ.GetCQ(o.CQ)
			w("- %s %s", o.CQ, q.Name)
			if o.Stmt != "" {
				w(" about “%s”", strings.TrimSuffix(gs.Say(o.Stmt), "."))
			}
			w(" [about %s]\n", aboutWords(o.Locus))
			w("    question: %s\n", gs.QuestionAbout(o.CQ, t, o.Stmt))
			w("    an attack argues: %s\n", o.AttackText)
			w("    a defence argues: %s\n", o.DefenceText)
			w("    schemes: %s\n", strings.Join(o.Schemes, ", "))
			if o.Locus == nascq.LocusPremise && o.Stmt != "" {
				w("    target_premise_id: %s\n", o.Stmt)
			}
			if !o.Question {
				w("    (an exception or a counterclaim cannot be asked alone: argue it)\n")
			}
		}
	}
	if !any {
		w("None yet.\n")
	}
	w("\n")

	w("## The schemes\n")
	for _, id := range nascq.SortedStringKeys(gs.NASCQ.Schemes) {
		sc := gs.NASCQ.Schemes[id]
		w("\n%s — %s", id, sc.Name)
		if sc.ThesisIndependent {
			w(" (can back an answer to any question)")
		}
		w("\n  %s\n  parts:", sc.DescriptionEN)
		vars := game.VariableOrder(nascq.ActionTemplate{Variables: sc.Variables, PremiseTemplatesEN: sc.PremiseTemplatesEN, ConclusionTemplateEN: sc.ConclusionTemplateEN})
		if sc.AnswerVariable != "" {
			vars = append(vars, sc.AnswerVariable)
		}
		for i, v := range vars {
			spec := sc.Variables[v]
			if i > 0 {
				w(";")
			}
			w(" %s = %s (%s", v, spec.Description, spec.Type)
			if spec.Example != "" {
				w(", e.g. %q", spec.Example)
			}
			w(")")
			if v == sc.AnswerVariable {
				w(", filled in for you when the argument answers a question")
			}
		}
		w("\n  reads: %s; therefore %s\n", strings.Join(sc.PremiseTemplatesEN, "; "), sc.ConclusionTemplateEN)
		if t := theoryScheme(gs, sc.CarneadesID); t != nil {
			r := readsOf(gs.Adapter.Graph.Theory.Language, t)
			if len(r.Assumptions) > 0 {
				w("  also assumed: %s\n", strings.Join(r.Assumptions, "; "))
			}
			if len(r.Exceptions) > 0 {
				w("  unless (exceptions the other side may argue): %s\n", strings.Join(r.Exceptions, "; "))
			}
		}
		switch sc.Evidence {
		case "required":
			w("  sources: required%s\n", noteOf(sc))
		case "recommended":
			w("  sources: recommended%s\n", noteOf(sc))
		}
	}
	w(`
## Your reply
Think it through, then end your reply with one JSON object: your move. For example
{"type":"attack","target_arg_id":"a2","answers_cq_id":"CQ:field","target_premise_id":"…","scheme_id":"AS:position_to_know","variable_bindings":{"W":"…","D":"…"},"text":"your argument in a sentence or two","evidence":["a source, if any"]}
- "type": "support" for a new argument (no target) or a defence (with a target and a question); "attack"; "question" (target, question and, for a premise, target_premise_id; no scheme); "abstain" to pass; "resign".
- Leave the answer part out of variable_bindings: the game fills it in.
- Give target_premise_id when the question is about a premise and a target_premise_id is listed.
`)
	return b.String()
}

func theoryScheme(gs *game.GameState, id string) *caes.Scheme {
	if th := gs.Adapter.Graph.Theory; th != nil {
		for _, t := range th.ArgSchemes {
			if t.Id == id {
				return t
			}
		}
	}
	return nil
}

func noteOf(sc *nascq.ASNode) string {
	if sc.EvidenceNoteEN == "" {
		return ""
	}
	return ": " + sc.EvidenceNoteEN
}

func textAt(texts []string, i int, fallback string) string {
	if i < len(texts) {
		return strings.TrimSuffix(texts[i], ".")
	}
	return fallback
}

func countMain(st *State) int {
	n := 0
	for _, a := range st.Arguments {
		if a.Role == "main" {
			n++
		}
	}
	return n
}

// promptArgs picks the arguments to show: those for or against the claim,
// then the most recent, in order.
func promptArgs(st *State, claim string) []string {
	var all []string
	for id, a := range st.Arguments {
		if a.Role == "main" {
			all = append(all, id)
		}
	}
	sort.Slice(all, func(i, j int) bool { return argNum(all[i]) < argNum(all[j]) })
	keep := map[string]bool{}
	if c := st.Statements[claim]; c != nil {
		for _, id := range c.Pro {
			keep[id] = true
		}
		if o := st.Statements[c.Opposite]; o != nil {
			for _, id := range o.Pro {
				keep[id] = true
			}
		}
	}
	for i := len(all) - 1; i >= 0 && len(keep) < promptArguments; i-- {
		keep[all[i]] = true
	}
	var out []string
	for _, id := range all {
		if keep[id] && len(out) < promptArguments {
			out = append(out, id)
		}
	}
	return out
}

// promptTargetIDs picks the arguments whose questions are offered: those the
// player owes an answer about, then the most recent shown.
func promptTargetIDs(st *State, shown []string, me string) []string {
	var out []string
	seen := map[string]bool{}
	for _, q := range st.Questions {
		if q.Owner == me && !seen[q.Target] {
			seen[q.Target] = true
			out = append(out, q.Target)
		}
	}
	for i := len(shown) - 1; i >= 0 && len(out) < promptTargets; i-- {
		if !seen[shown[i]] {
			seen[shown[i]] = true
			out = append(out, shown[i])
		}
	}
	return out
}

func argNum(id string) int {
	n := 0
	fmt.Sscanf(strings.TrimPrefix(id, "a"), "%d", &n)
	return n
}

// ParseReply reads a player's reply: free text that contains a JSON object,
// the move. The last object that has a "type" counts, so a reply may think
// aloud first, quote examples, or wrap the JSON in a code block. Common
// variations of the field names are accepted ("scheme" for "scheme_id",
// "target" for "target_arg_id", "cq" or "question" for "answers_cq_id",
// "premise" for "target_premise_id", "slots" or "bindings" for
// "variable_bindings", "sources" for "evidence"; "pass" for "abstain",
// "argue" and "defend" for "support").
func ParseReply(reply string) (nascq.ProposedAction, error) {
	var a nascq.ProposedAction
	objs := jsonObjects(reply)
	for i := len(objs) - 1; i >= 0; i-- {
		var m map[string]interface{}
		if json.Unmarshal([]byte(objs[i]), &m) != nil {
			continue
		}
		str := func(keys ...string) string {
			for _, k := range keys {
				if v, ok := m[k].(string); ok && strings.TrimSpace(v) != "" {
					return strings.TrimSpace(v)
				}
			}
			return ""
		}
		t := strings.ToLower(str("type", "move", "action"))
		if t == "" {
			continue
		}
		switch t {
		case "pass":
			t = "abstain"
		case "argue", "defend", "defence", "defense":
			t = "support"
		case "ask":
			t = "question"
		}
		a = nascq.ProposedAction{Type: t,
			SchemeID:        str("scheme_id", "scheme"),
			TargetArgID:     str("target_arg_id", "target", "argument"),
			AnswersCQID:     str("answers_cq_id", "cq", "question", "critical_question"),
			TargetPremiseID: str("target_premise_id", "premise"),
			AttackType:      str("attack_type"),
			Text:            str("text", "argument_text", "words")}
		for _, k := range []string{"variable_bindings", "bindings", "slots", "variables", "parts"} {
			if vb, ok := m[k].(map[string]interface{}); ok {
				a.VariableBindings = map[string]string{}
				for name, v := range vb {
					if s := strings.TrimSpace(fmt.Sprint(v)); s != "" && v != nil {
						a.VariableBindings[name] = s
					}
				}
				break
			}
		}
		for _, k := range []string{"evidence", "sources"} {
			switch ev := m[k].(type) {
			case []interface{}:
				for _, x := range ev {
					if s := strings.TrimSpace(fmt.Sprint(x)); s != "" {
						a.Evidence = append(a.Evidence, s)
					}
				}
			case string:
				if strings.TrimSpace(ev) != "" {
					a.Evidence = append(a.Evidence, strings.TrimSpace(ev))
				}
			}
		}
		if a.SchemeID != "" && !strings.HasPrefix(a.SchemeID, "AS:") {
			a.SchemeID = "AS:" + a.SchemeID
		}
		if a.AnswersCQID != "" && !strings.HasPrefix(a.AnswersCQID, "CQ:") {
			a.AnswersCQID = "CQ:" + a.AnswersCQID
		}
		return a, nil
	}
	return a, fmt.Errorf(`no move found: end the reply with one JSON object such as {"type":"abstain"}`)
}

// jsonObjects returns the balanced {...} spans of s, outermost only, in
// order. Braces inside JSON strings are skipped.
func jsonObjects(s string) []string {
	var out []string
	depth, start, inStr, esc := 0, -1, false, false
	for i, r := range s {
		switch {
		case inStr:
			switch {
			case esc:
				esc = false
			case r == '\\':
				esc = true
			case r == '"':
				inStr = false
			}
		case r == '"':
			if depth > 0 {
				inStr = true
			}
		case r == '{':
			if depth == 0 {
				start = i
			}
			depth++
		case r == '}':
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					out = append(out, s[start:i+1])
				}
			}
		}
	}
	return out
}
