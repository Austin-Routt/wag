package server

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
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

// ProtocolVersion names the text protocol the prompt and the reply follow.
const ProtocolVersion = "wag-protocol/2"

// Prompt limits: a large game (an imported Kialo discussion) is cut to the
// arguments nearest the play. An ordinary game fits whole.
const (
	promptArguments  = 40  // arguments shown with their premises
	promptTargets    = 40  // arguments whose critical questions are listed
	promptHistory    = 40  // moves of the history shown
	promptStatements = 150 // statements listed for moves about them
)

// Prompt is the text a player sees when it is their move: the game, every
// legal move with the help the browser gives (the question, what each
// answer would argue, which schemes can back it, linked in the NASCQ or
// thesis-independent, and their semantic types), and the reply format. The
// statements on the board have references, [s1], [s2], ..., that a move can
// use in place of a statement ID.
func (ss *Session) Prompt() string {
	gs, st := ss.GS, ss.buildState()
	var b strings.Builder
	w := func(format string, args ...interface{}) { fmt.Fprintf(&b, format, args...) }
	me := gs.ActivePlayer()
	if me == nil || gs.Terminated {
		return "The game is over: " + gs.TerminationReason + "\n"
	}
	refs := newRefs(gs)
	ref := func(id string) string { return refs.ref(id) }

	w("# WAG: the Waltonian Argumentation Game. It is your move. (%s)\n\n", ss.protocol())
	if t := strings.TrimSpace(ss.Title); t != "" && !strings.HasPrefix(t, "WAG — ") {
		w("Topic: %s\n\n", t)
	}
	w("You are %s. ", me.ID)
	switch side := gs.Side(me.ID); {
	case side == "pro":
		w("You argue FOR the claim: you win if the claim is IN when the game ends.")
	case side == "con":
		w("You argue AGAINST the claim: you win unless the claim is IN when the game ends.")
	case gs.Config.Cooperative:
		w("You and the other player examine the claim together: neither of you argues a side or wins; the result is the claim's label when the game ends (when both of you pass in one round, or at the round limit).")
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
	if gs.Config.Context != "" {
		w(" Context: %s.", gs.Config.Context)
	}
	w("\n\n")

	w(`## How the game works
- Every argument is an instance of an argumentation scheme (Walton): you fill in the scheme's parts with short phrases, and the scheme's sentences (listed under "The schemes") are built from them. Its premises hold until someone challenges or attacks them.
- Each scheme has critical questions. To ATTACK an argument, answer one of its questions against it; to DEFEND one, answer a question in its favour. Either way you back the answer with an argument of a scheme the question allows; the answer itself is filled in for you.
- You may ASK a question about a premise or the inference without answering it, or CHALLENGE a premise ("why should we accept that?"): the argument stops counting until its author answers, with any argument for the premise.
- You may argue FOR or AGAINST any statement on the board, as for the claim, with a scheme whose conclusion can be that statement.
- Or put forward a NEW argument, PASS or RESIGN.
- After every move the Carneades engine labels each statement IN, OUT or UNDECIDED. An argument is in force when all its premises are IN and no exception or undercutter against it is IN. A statement that is at issue (argued for and against) is decided by the strongest argument in force on each side, not by their sum: under PE (preponderance) the stronger side wins and a tie leaves both OUT; under CCE it must also lead by more than 0.5; under BRD, also the other side must weigh under 0.3.
- Make your parts true and relevant to the discussion, in plain words: the other player can challenge any of them. Cite sources when a scheme asks for them.
`)
	if gs.Config.Objective == game.ObjectiveClaim && gs.Config.Cooperative {
		w("- The inquiry is about the claim: you examine it together, and its label when the game ends is the result; no one wins, and the scores decide nothing. ")
		if gs.Claim == "" {
			w("There is no claim yet: the conclusion of the first argument becomes the claim. ")
		}
		w("The game ends when both players pass in one round")
		if gs.Config.MaxTurns > 0 {
			w(", after round %d,", gs.Config.MaxTurns)
		}
		w(" or when a player resigns.\n")
	} else if gs.Config.Objective == game.ObjectiveClaim {
		w("- The debate is about the claim: pro wins if it is IN when the game ends, con otherwise; the scores do not decide it. ")
		if gs.Claim == "" {
			w("There is no claim yet: the conclusion of pro's first argument becomes the claim. ")
		}
		w("The game ends")
		if gs.Config.MaxTurns > 0 {
			w(" after round %d,", gs.Config.MaxTurns)
		}
		w(" when a side resigns")
		if gs.Config.ConcedeOnPass {
			w(", or when a side passes while the verdict is against it")
		}
		w(".\n")
	}
	if gs.Config.Scoring.Model == game.ScoringLive {
		w("- Scores are live: they count what stands now, so points for an argument go when it falls.\n")
	}
	w("\n")
	if p := gs.Profile; p != nil {
		ss.promptProfile(w, p)
	}
	if gs.Claim != "" {
		c := st.Statements[gs.Claim]
		w("## The claim\n%s %s It is now %s. %s\n\n", ref(gs.Claim), gs.Say(gs.Claim), strings.ToUpper(c.Label), c.Explain)
	}
	ss.promptLastMove(w, me.ID, refs)

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
			w("  C: %s%s [%s]\n", spaced(ref(a.Conclusion)), gs.Say(a.Conclusion), strings.ToUpper(st.Statements[a.Conclusion].Label))
			for i, p := range a.Premises {
				kind := "premise" // in a game with a record set, an assumption is named so (E1.7, B6)
				for _, x := range a.Assumptions {
					if x == p {
						kind = "assumption"
					}
				}
				w("  %s %s%s [%s]\n", kind, spaced(ref(p)), textAt(a.PremiseTexts, i, p), strings.ToUpper(st.Statements[p].Label))
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
			if len(a.EvidenceRefs) > 0 {
				var cited []string
				for _, r := range a.EvidenceRefs {
					state := "admitted"
					if !r.Admitted {
						state = "withdrawn"
					}
					cited = append(cited, r.Binding+" ("+state+")")
				}
				w("  Bindings cited: %s", strings.Join(cited, ", "))
				if a.Ungrounded {
					w("; no longer grounded by admitted evidence (its label is as before)")
				}
				w("\n")
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
			w("You must answer: %s asks about %s: %s", q.Asker, q.Target, q.Question)
			if q.Stmt != "" {
				w(" Answer it with any argument for %s (or the question's defence).", ref(q.Stmt))
			}
			w("\n\n")
		}
	}
	if gs.LastRejection != "" {
		w("Your last move was rejected: %s\n\n", gs.LastRejection)
	}

	ss.promptEvidence(w, st, refs)
	w("## Your legal moves\n")
	ss.promptQuestions(w, st, ids, me.ID, refs)
	ss.promptStatements(w, st, ids, refs)
	w("\n### New argument\nAny scheme, with no target: {\"type\":\"support\",\"scheme_id\":…,\"variable_bindings\":{…}}.")
	if gs.Config.Cooperative {
		w(" In an inquiry, the first argument states the claim.\n")
	} else {
		w(" In a debate, pro's first argument states the claim.\n")
	}
	ss.promptEvidenceActs(w)
	w("\n### Pass or resign\n{\"type\":\"abstain\"} passes")
	if gs.Config.ConcedeOnPass {
		w(" (in a debate, passing while the verdict is against your side concedes)")
	}
	w("; {\"type\":\"resign\"} gives up.\n\n")

	ss.promptSchemes(w)
	if gs.Config.Cooperative {
		// An inquiry has no sides (docs/e1-evidence.md, E1.0).
		w("%s", strings.Replace(replyFormat, "the move helps your side", "the move helps the inquiry", 1))
	} else {
		w("%s", replyFormat)
	}
	return b.String()
}

// answerWords says which literal answer to a question attacks the argument.
func answerWords(q *nascq.CQNode) string {
	if q == nil || q.AnswerEN == nil {
		return ""
	}
	return fmt.Sprintf(" (an attack answers %q, a defence %q)", q.AnswerEN.Against, q.AnswerEN.For)
}

// promptQuestions lists the critical questions the player can answer or
// ask, argument by argument, and the challenges of premises.
func (ss *Session) promptQuestions(w func(string, ...interface{}), st *State, shown []string, me string, refs *refs) {
	gs := ss.GS
	w("\n### Critical questions: answer (attack or defend) or ask\n")
	w("Attack (answer against the argument): {\"type\":\"attack\",\"target_arg_id\":…,\"answers_cq_id\":…,\"target_premise_id\":… (if listed),\"scheme_id\":…,\"variable_bindings\":{…}}; defend (answer in its favour): the same with \"type\":\"support\"; ask: {\"type\":\"question\",\"target_arg_id\":…,\"answers_cq_id\":…,\"target_premise_id\":…}. ")
	var ti []string
	for _, id := range nascq.SortedStringKeys(gs.NASCQ.Schemes) {
		if sc := gs.NASCQ.Schemes[id]; sc.ThesisIndependent {
			ti = append(ti, fmt.Sprintf("%s (%s)", id, sc.SemanticType))
		}
	}
	w("Under each question: the statement an attack or a defence would argue for (filled in for you), and the schemes the NASCQ links to the question, with their semantic type (the kind of claim they support). Every question can also be answered with a thesis-independent scheme: %s.\n", strings.Join(ti, ", "))
	any := false
	var challenges []string
	for _, t := range promptTargetIDs(st, shown, me) {
		opts := moves(gs, t)
		if len(opts) == 0 {
			continue
		}
		a := st.Arguments[t]
		var lines []string
		var chal []string
		for _, o := range opts {
			if o.CQ == nascq.CQPremiseAcceptable {
				chal = append(chal, refs.ref(o.Stmt))
				continue
			}
			q := gs.NASCQ.GetCQ(o.CQ)
			l := fmt.Sprintf("- %s %s (%s), about %s", o.CQ, strings.TrimSuffix(q.Name, " Question"), q.SemanticType, aboutWords(o.Locus))
			if r := refs.ref(o.Stmt); r != "" {
				l += " " + r
				if o.Locus == nascq.LocusPremise {
					l += fmt.Sprintf(" (target_premise_id %q)", refs.id(o.Stmt))
				}
			}
			if o.Question {
				l += "; can be asked alone"
			} else {
				l += "; cannot be asked alone: argue it"
			}
			l += fmt.Sprintf("\n    Q: %s%s\n    attack argues: %s\n    defence argues: %s", gs.QuestionAbout(o.CQ, t, o.Stmt), answerWords(q), o.AttackText, o.DefenceText)
			if b := ss.backers(o.CQ, o.Schemes); b != "" {
				l += "\n    linked: " + b
			}
			for _, typ := range []string{"attack", "support"} {
				for _, sid := range nascq.SortedStringKeys(o.Fixed[typ]) {
					var parts []string
					for _, v := range nascq.SortedStringKeys(o.Fixed[typ][sid]) {
						parts = append(parts, v+" = "+o.Fixed[typ][sid][v])
					}
					l += fmt.Sprintf("\n    %s with %s: the answer fixes %s (leave it out)", map[string]string{"attack": "an attack", "support": "a defence"}[typ], sid, strings.Join(parts, ", "))
				}
			}
			lines = append(lines, l)
		}
		if len(lines) == 0 && len(chal) == 0 {
			continue
		}
		any = true
		w("\n%s (%s's %s): C %s%s\n", t, a.Owner, a.SchemeName, spaced(refs.ref(a.Conclusion)), gs.Say(a.Conclusion))
		for _, l := range lines {
			w("%s\n", l)
		}
		if len(chal) > 0 {
			challenges = append(challenges, fmt.Sprintf("%s: %s", t, strings.Join(chal, " ")))
		}
	}
	if !any {
		w("None yet.\n")
	}
	if len(challenges) > 0 {
		w("\n### Challenges\nAsk why a premise should be accepted (a question: no scheme): {\"type\":\"question\",\"target_arg_id\":…,\"answers_cq_id\":\"%s\",\"target_premise_id\":…}. The premises no critical question is about:\n", nascq.CQPremiseAcceptable)
		for _, c := range challenges {
			w("- %s\n", c)
		}
	}
}

// backers lists the schemes the NASCQ links to a critical question (the
// browser marks them 🔗), with their semantic type; the thesis-independent
// schemes, which back any question, are listed once for all.
func (ss *Session) backers(cqID string, schemes []string) string {
	gs := ss.GS
	linked := map[string]bool{}
	for _, e := range gs.NASCQ.BackupEdges {
		if e.CQ == cqID {
			linked[e.Scheme] = true
		}
	}
	var out []string
	for _, id := range schemes {
		sc := gs.NASCQ.GetScheme(id)
		if sc == nil {
			continue
		}
		if linked[id] {
			out = append(out, fmt.Sprintf("%s (%s)", id, sc.SemanticType))
		}
	}
	return strings.Join(out, ", ")
}

// promptStatements lists the statements on the board with their references,
// and the schemes that can argue for or against each.
func (ss *Session) promptStatements(w func(string, ...interface{}), st *State, shown []string, refs *refs) {
	gs := ss.GS
	// The statements of the arguments shown, and the claim.
	keep := map[string]bool{gs.Claim: true}
	for _, id := range shown {
		a := st.Arguments[id]
		keep[a.Conclusion] = true
		keep[strings.TrimPrefix(a.Conclusion, "¬")] = true
		for _, p := range a.Premises {
			keep[p] = true
		}
	}
	var stmts []string
	for _, id := range refs.order {
		if keep[id] && len(stmts) < promptStatements {
			stmts = append(stmts, id)
		}
	}
	if len(stmts) == 0 {
		return
	}
	var any []string
	for _, id := range nascq.SortedStringKeys(gs.NASCQ.Schemes) {
		if sc := gs.NASCQ.Schemes[id]; sc.AnswerVariable != "" {
			any = append(any, id)
		}
	}
	w("\n### For or against a statement\nArgue for: {\"type\":\"support\",\"target_stmt_id\":\"s3\",\"scheme_id\":…,\"variable_bindings\":{…}}; against: the same with \"type\":\"attack\". ")
	w("Any statement with a reference above ([s3], or [¬s3] for an opposite) can be argued for or against. %s conclude any statement: their answer part becomes the statement (or its opposite), filled in for you. The other schemes only where listed here, with the parts the statement fixes (leave those out):\n", strings.Join(any, ", "))
	acts := gs.NASCQ.EnumerateStatementActions(gs.Bridge(), stmts)
	extra := map[string][]string{}
	for _, t := range acts {
		sc := gs.NASCQ.GetScheme(t.SchemeID)
		if sc == nil || sc.AnswerVariable != "" {
			continue
		}
		var fixed []string
		for _, v := range nascq.SortedStringKeys(t.Fixed) {
			fixed = append(fixed, v+" = "+gs.SayValue(t.Fixed[v]))
		}
		dir := "for"
		if t.Type == "attack" {
			dir = "against"
		}
		extra[t.TargetStmtID+"|"+dir] = append(extra[t.TargetStmtID+"|"+dir], fmt.Sprintf("%s (%s)", t.SchemeID, strings.Join(fixed, ", ")))
	}
	listed := false
	for _, id := range stmts {
		xf, xa := extra[id+"|for"], extra[id+"|against"]
		if len(xf)+len(xa) == 0 {
			continue
		}
		listed = true
		w("- %s", refs.ref(id))
		if id == gs.Claim {
			w(" (the claim)")
		}
		if len(xf) > 0 {
			w(" for it: %s", strings.Join(xf, ", "))
			if len(xa) > 0 {
				w(";")
			}
		}
		if len(xa) > 0 {
			w(" against it: %s", strings.Join(xa, ", "))
		}
		w("\n")
	}
	if !listed {
		w("(none now)\n")
	}
}

// promptProfile says who deliberates and about which actions (D094 (b)).
func (ss *Session) promptProfile(w func(string, ...interface{}), p *game.Profile) {
	w("## Who decides\nThe agent who decides is %s", p.Agent)
	for _, a := range p.Actors {
		if a.ID == p.Agent && a.Text != "" {
			w(" (%s)", a.Text)
		}
	}
	w(". A recommendation (\"A should be done\") is addressed to the agent, so the actions you argue about are the agent's; the game fills in the agent's place (P) itself. The alternatives are compared in this situation: %s.\n", p.Situation)
	var mine []string
	for _, a := range p.Actions {
		if a.Actor != p.Agent {
			continue
		}
		d := a.ID
		if len(a.Aliases) > 0 {
			d += " (also called " + strings.Join(a.Aliases, ", ") + ")"
		}
		mine = append(mine, d)
	}
	if len(mine) > 0 {
		w("The agent's actions so far: %s.", strings.Join(mine, "; "))
	} else {
		w("No action of the agent has been argued about yet.")
	}
	var others []string
	for _, a := range p.Actions {
		if a.Actor != p.Agent {
			others = append(others, fmt.Sprintf("%s (%s's)", a.ID, a.Actor))
		}
	}
	if len(others) > 0 {
		w(" Other actors' actions, which a recommendation to the agent does not argue: %s.", strings.Join(others, "; "))
	}
	w(` To argue about another action of the agent, declare it in the move that first names it, with its actor: "declare": [{"id": "the_action", "actor": %q, "aliases": []}].`+"\n\n", p.Agent)
}

// promptSchemes lists the schemes: parts, sentences, assumptions,
// exceptions, sources, semantic type and the cue words that signal them.
func (ss *Session) promptSchemes(w func(string, ...interface{})) {
	gs := ss.GS
	other := "the other side" // who argues a scheme's exceptions: in an inquiry there are no sides
	if gs.Config.Cooperative {
		other = "either player"
	}
	w("## The schemes\n")
	for _, id := range nascq.SortedStringKeys(gs.NASCQ.Schemes) {
		sc := gs.NASCQ.Schemes[id]
		switch sc.Role {
		case nascq.RoleAnswer:
			w("\n%s — %s (%s; an answer: it weighs as the argument it answers)", id, sc.Name, sc.SemanticType)
		case nascq.RoleModal, nascq.RoleBridge:
			w("\n%s — %s (%s, weight %.2f, as %s)", id, sc.Name, sc.SemanticType, gs.NASCQ.WeightOf(sc), sc.WeighsAs)
		default:
			w("\n%s — %s (%s, weight %.2f)", id, sc.Name, sc.SemanticType, sc.BaseWeight)
		}
		if sc.ThesisIndependent {
			w(", can back an answer to any question")
		}
		if sc.Comparison {
			w(`; the move says how its alternative is better: "comparison": "surely", "sooner" or "cheaper"`)
		}
		w("\n  %s\n  parts:", sc.DescriptionEN)
		vars := game.VariableOrder(nascq.ActionTemplate{Variables: sc.Variables, PremiseTemplatesEN: sc.PremiseTemplatesEN,
			ConclusionTemplateEN: sc.ConclusionTemplateEN, AnswerVariable: sc.AnswerVariable})
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
				w(", filled in for you when the argument answers a question or is about a statement")
			}
		}
		w("\n  reads: %s; therefore %s\n", strings.Join(sc.PremiseTemplatesEN, "; "), sc.ConclusionTemplateEN)
		if t := theoryScheme(gs, sc.CarneadesID); t != nil {
			r := readsOf(gs.Adapter.Graph.Theory.Language, t)
			if len(r.Assumptions) > 0 {
				w("  also assumed: %s\n", strings.Join(r.Assumptions, "; "))
			}
			if len(r.Exceptions) > 0 {
				w("  unless (exceptions %s may argue): %s\n", other, strings.Join(r.Exceptions, "; "))
			}
		}
		switch {
		case sc.Evidence == "required" && gs.Config.RequireEvidence:
			w("  sources: required in this game%s\n", noteOf(sc))
		case sc.Evidence != "":
			w("  sources: recommended%s\n", noteOf(sc))
		}
		if len(sc.CuesEN) > 0 {
			w("  typical wording: %s\n", strings.Join(sc.CuesEN, ", "))
		}
	}
}

// replyFormat is the end of every prompt: the reply under protocol 2.
const replyFormat = `
## Your reply
Think it through, then end your reply with one JSON object. An example; use the arguments, questions and references of this game:
{
  "context": ["what you observe on the board that matters for this move", "…"],
  "thoughts": ["your reasoning, step by step", "…"],
  "summary": "one or two sentences: the situation, your reasoning and your move",
  "expect": [{"stmt": "s1", "label": "out"}],
  "alternatives": [{"action": {"type": "question", "target_arg_id": "a1", "answers_cq_id": "CQ:premise_acceptable", "target_premise_id": "s2"}, "why_not": "…"}],
  "confidence": 0.6,
  "action": {"type": "attack", "target_stmt_id": "s2", "scheme_id": "AS:position_to_know", "variable_bindings": {"W": "…", "D": "…"}, "text": "your argument in a sentence or two, as the other player will read it", "evidence": ["a source, if any"]}
}
- "action" is your move; it is the only required field. "type": "support" (a new argument; with a target and a question, a defence; with target_stmt_id, an argument for the statement), "attack", "question", "abstain" (pass) or "resign".
- Leave out of variable_bindings the answer part and the parts a statement fixes: the game fills them in.
- Statements can be given by reference ("s3", "¬s3") or by ID. "text" is your argument in your own words; "evidence" lists your sources. Both are shown to the other player.
- "expect": the labels you predict for some statements ("claim" for the claim) after your move; they are checked. "alternatives": moves you considered and why you did not make them. "confidence": how likely, from 0 to 1, the move helps your side. These three are optional; they are recorded with the move, to study and improve the game.
- A bare move object, as in protocol 1, is still accepted.
`

// spaced is a reference followed by a space, or "" for none.
func spaced(ref string) string {
	if ref == "" {
		return ""
	}
	return ref + " "
}

// promptLastMove tells the player how its previous move turned out: what it
// did, and how its predictions fared.
func (ss *Session) promptLastMove(w func(string, ...interface{}), player string, refs *refs) {
	for i := len(ss.Traces) - 1; i >= 0; i-- {
		t := ss.Traces[i]
		if t.Player != player || t.Move < 0 {
			continue
		}
		w("## Your last move\n%d. %s\n", t.Move+1, ss.GS.DescribeMove(ss.GS.MoveHistory[t.Move]))
		for _, e := range t.Effects {
			w("   → %s\n", e)
		}
		if len(t.ExpectCheck) > 0 {
			var parts []string
			for _, c := range t.ExpectCheck {
				mark := "right"
				if !c.OK {
					mark = "wrong"
				}
				name := refs.id(c.Stmt)
				if name == "" {
					name = c.Stmt
				}
				parts = append(parts, fmt.Sprintf("%s %s: %s (%s)", name, c.Predicted, c.Actual, mark))
			}
			w("Your predictions: %s.\n", strings.Join(parts, "; "))
		}
		w("\n")
		return
	}
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

// Reply is a player's reply under protocol 2: the move ("action"), and what
// the player says about it, all optional. Context, Thoughts and Summary are
// the player's own account (a rationale, not a faithful trace of how it
// decided); Expect, Alternatives and Confidence can be checked against what
// the move did and what the alternatives would have done (see Trace).
type Reply struct {
	Context      []string             `json:"context,omitempty"`
	Thoughts     []string             `json:"thoughts,omitempty"`
	Summary      string               `json:"summary,omitempty"`
	Expect       []Expectation        `json:"expect,omitempty"`
	Alternatives []Alternative        `json:"alternatives,omitempty"`
	Confidence   *float64             `json:"confidence,omitempty"`
	Action       nascq.ProposedAction `json:"action"`
	// Bare: the reply was a move object alone (protocol 1).
	Bare bool `json:"bare,omitempty"`
}

// Expectation is a label the player predicts a statement will have after
// the move: a reference ("s3"), a statement ID, or "claim".
type Expectation struct {
	Stmt  string `json:"stmt"`
	Label string `json:"label"`
}

// Alternative is a move the player considered and did not make.
type Alternative struct {
	Action nascq.ProposedAction `json:"action"`
	WhyNot string               `json:"why_not,omitempty"`
}

// ParseReply reads the move of a player's reply (see ParseReplyFull).
func ParseReply(reply string) (nascq.ProposedAction, error) {
	r, err := ParseReplyFull(reply)
	if err != nil {
		return nascq.ProposedAction{}, err
	}
	return r.Action, nil
}

// ParseReplyFull reads a player's reply: free text that contains a JSON
// object, either the protocol 2 envelope ({"context": …, "thoughts": …,
// "summary": …, "action": {the move}}) or the move alone (protocol 1). The
// final JSON object is authoritative; malformed final objects/arrays and
// truncated containers are refused instead of falling back to earlier examples.
// A reply may put prose/examples before it or wrap it in a code block. Common variations
// of the field names are accepted ("move" for "action"; in the move,
// "scheme" for "scheme_id", "target" for "target_arg_id", "cq" or
// "question" for "answers_cq_id", "premise" for "target_premise_id",
// "statement" or "stmt" for "target_stmt_id", "slots" or "bindings" for
// "variable_bindings", "sources" for "evidence"; "pass" for "abstain",
// "argue" and "defend" for "support").
func ParseReplyFull(reply string) (*Reply, error) { return parseReply(reply, false) }

// parseReply reads a reply as ParseReplyFull does; with evidence (a game
// with a record set, wag-protocol/2+evidence) it also reads the evidence
// acts' short names and the fields of an act and of a move's citations
// (offer, binding_id, binding, revision, evidence_refs). A game without a
// record set reads neither, as before E1 (docs/e1-evidence.md, E1.7, B3).
func parseReply(reply string, evidence bool) (*Reply, error) {
	objs := jsonCandidates(reply)
	if len(objs) == 0 {
		return nil, fmt.Errorf(`no move found: end the reply with one JSON object such as {"action": {"type": "abstain"}}`)
	}
	// The final JSON candidate is authoritative. A malformed or truncated
	// final answer must not cause an earlier example to become a real move.
	final := objs[len(objs)-1]
	// Well-formed JSON first. The decoder's nesting limit then bounds the
	// duplicate scan below, whose recursion has none of its own (on Go 1.24
	// a reply of millions of "[" overflowed the stack, a fatal error), and a
	// truncated move keeps the decoder's message, which is the feedback the
	// model is given and what a recorded trace replays against.
	if err := json.Unmarshal([]byte(final), new(json.RawMessage)); err != nil {
		return nil, fmt.Errorf("final move JSON: %w", err)
	}
	if err := rejectDuplicateReplyKeys(final); err != nil {
		return nil, fmt.Errorf("final move JSON: %w", err)
	}
	var m map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(final))
	// Binding numbers are phrases: retain their exact JSON spelling instead
	// of rounding large integers/fractions through float64. Confidence is
	// converted separately by number; saved string bindings stay strings.
	decoder.UseNumber()
	if err := decoder.Decode(&m); err != nil {
		return nil, fmt.Errorf("final move JSON: %w", err)
	}
	for _, k := range []string{"action", "move"} {
		v, exists := m[k]
		if !exists {
			continue
		}
		if am, ok := v.(map[string]interface{}); ok {
			a, ok, err := moveOf(am, evidence)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("%q: the final move needs a nonempty type", k)
			}
			r := &Reply{Action: a, Context: strs(m["context"]), Thoughts: strs(m["thoughts"]), Summary: str1(m["summary"])}
			r.Expect = expectations(m["expect"])
			r.Alternatives = alternatives(m["alternatives"], evidence)
			if c, ok := number(m["confidence"]); ok {
				r.Confidence = &c
			}
			return r, nil
		}
		if _, ok := v.(string); !ok {
			return nil, fmt.Errorf("%q: the final move must be an object or a move-type string, not %s", k, jsonKind(v))
		}
	}
	a, ok, err := moveOf(m, evidence)
	if err != nil {
		return nil, err
	}
	if ok {
		return &Reply{Action: a, Bare: true}, nil
	}
	return nil, fmt.Errorf("no move in the final JSON object: give a nonempty type or action object")
}

// rejectDuplicateReplyKeys checks the raw final candidate before map decoding
// can silently overwrite an earlier member. Names are compared exactly after
// JSON unescaping: variable bindings such as A and a remain distinct. Each
// object has its own namespace, including objects nested inside arrays.
func rejectDuplicateReplyKeys(raw string) error {
	d := json.NewDecoder(strings.NewReader(raw))
	d.UseNumber()
	var value func() error
	value = func() error {
		token, err := d.Token()
		if err != nil {
			return err
		}
		delim, compound := token.(json.Delim)
		if !compound {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				token, err := d.Token()
				if err != nil {
					return err
				}
				key, ok := token.(string)
				if !ok {
					return fmt.Errorf("object member name must be a string")
				}
				if seen[key] {
					return fmt.Errorf("duplicate JSON member %q", key)
				}
				seen[key] = true
				if err := value(); err != nil {
					return err
				}
			}
		case '[':
			for d.More() {
				if err := value(); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected delimiter %q", delim)
		}
		_, err = d.Token()
		return err
	}
	if err := value(); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected one JSON value")
	}
	return nil
}

// moveOf reads a move object; ok is false if it has no type. err is set for
// a move whose fields cannot be read (a source that is not text). evidence
// says whether the game has a record set (parseReply).
func moveOf(m map[string]interface{}, evidence bool) (nascq.ProposedAction, bool, error) {
	str := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}
	for _, k := range []string{"type", "move", "action"} {
		if v, exists := m[k]; exists {
			if _, ok := v.(string); !ok {
				return nascq.ProposedAction{}, true, fmt.Errorf("%q must be a move-type string, not %s", k, jsonKind(v))
			}
		}
	}
	t := strings.ToLower(str("type", "move", "action"))
	if t == "" {
		return nascq.ProposedAction{}, false, nil
	}
	// A target, scheme or question given as anything but text is refused by
	// name: dropping it would change the move (a defence without its target
	// is a new argument). null counts as absent.
	textKeys := []string{"scheme_id", "scheme", "target_arg_id", "target", "argument",
		"answers_cq_id", "cq", "question", "critical_question", "target_premise_id", "premise",
		"target_stmt_id", "statement", "stmt", "target_statement", "attack_type", "ground", "text", "argument_text", "words"}
	if evidence {
		textKeys = append(textKeys, "binding_id", "binding", "revision")
	}
	for _, k := range textKeys {
		if v, exists := m[k]; exists && v != nil {
			if _, ok := v.(string); !ok {
				return nascq.ProposedAction{Type: t}, true, fmt.Errorf("%q must be text, not %s", k, jsonKind(v))
			}
		}
	}
	for _, k := range []string{"variable_bindings", "bindings", "slots", "variables", "parts"} {
		v, exists := m[k]
		if !exists || v == nil {
			continue
		}
		vb, ok := v.(map[string]interface{})
		if !ok {
			return nascq.ProposedAction{Type: t}, true, fmt.Errorf("%q must be an object of variable bindings, not %s", k, jsonKind(v))
		}
		// In the order of their names, so that a reply with two faulty parts
		// is refused for the same one each time: a replay compares the
		// reasons, and the next prompt quotes them.
		for _, name := range nascq.SortedStringKeys(vb) {
			x := vb[name]
			switch x.(type) {
			case map[string]interface{}, []interface{}:
				return nascq.ProposedAction{Type: t}, true, fmt.Errorf("%q: the binding of %s must be a phrase, not %s", k, name, jsonKind(x))
			}
		}
	}
	switch t {
	case "pass":
		t = "abstain"
	case "argue", "defend", "defence", "defense", "for":
		t = "support"
	case "against":
		t = "attack"
	case "ask", "challenge":
		t = "question"
	}
	if evidence {
		switch t {
		case "offer":
			t = game.ActOffer
		case "consent":
			t = game.ActConsent
		case "withdraw", "withdrawal":
			t = game.ActWithdraw
		}
	}
	a := nascq.ProposedAction{Type: t,
		SchemeID:        str("scheme_id", "scheme"),
		TargetArgID:     str("target_arg_id", "target", "argument"),
		AnswersCQID:     str("answers_cq_id", "cq", "question", "critical_question"),
		TargetPremiseID: str("target_premise_id", "premise"),
		TargetStmtID:    str("target_stmt_id", "statement", "stmt", "target_statement"),
		AttackType:      str("attack_type"),
		Ground:          str("ground"),
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
	if a.SchemeID != "" && !strings.HasPrefix(a.SchemeID, "AS:") {
		a.SchemeID = "AS:" + a.SchemeID
	}
	if a.AnswersCQID != "" && !strings.HasPrefix(a.AnswersCQID, "CQ:") {
		a.AnswersCQID = "CQ:" + a.AnswersCQID
	}
	// The optional parts are read each on its own: a part that does not read
	// is left out and its error returned, the others kept (a considered move
	// keeps its sources when its declarations do not read).
	var first error
	if ev, err := sourcesOf(m); err != nil {
		first = err
	} else {
		a.Evidence = ev
	}
	if d, err := declarationsOf(m); err != nil {
		if first == nil {
			first = err
		}
	} else {
		a.Declare = d
	}
	if c, err := comparisonOf(m); err != nil {
		if first == nil {
			first = err
		}
	} else {
		a.Comparison = c
	}
	if !evidence {
		return a, true, first
	}
	// An evidence act's parts, and the bindings a move cites
	// (docs/e1-evidence.md), in a game with a record set: read only when
	// given.
	a.BindingID, a.Revision = str("binding_id", "binding"), str("revision")
	if o, err := offerOf(m); err != nil {
		if first == nil {
			first = err
		}
	} else {
		a.Offer = o
	}
	if refs, err := bindingRefsOf(m); err != nil {
		if first == nil {
			first = err
		}
	} else {
		a.EvidenceRefs = refs
	}
	return a, true, first
}

// offerFields are the parts of an evidence offer.
var offerFields = []string{"claim_id", "argument_id", "kind", "statement", "role", "interpretation", "warrant"}

// offerOf reads an evidence offer: an object of text parts (offerFields).
// Another value, an unknown part, or a part that is not text is refused by
// name; absent or null is no offer.
func offerOf(m map[string]interface{}) (*nascq.EvidenceOffer, error) {
	v, exists := m["offer"]
	if !exists || v == nil {
		return nil, nil
	}
	o, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("\"offer\" must be an object of %s, not %s", strings.Join(offerFields, ", "), jsonKind(v))
	}
	parts := map[string]string{}
	for _, k := range nascq.SortedStringKeys(o) { // sorted: see moveOf
		x := o[k]
		known := false
		for _, f := range offerFields {
			known = known || f == k
		}
		if !known {
			return nil, fmt.Errorf("\"offer\" has no part %q: its parts are %s", k, strings.Join(offerFields, ", "))
		}
		if x == nil {
			continue
		}
		s, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("\"offer\": %q must be text, not %s", k, jsonKind(x))
		}
		parts[k] = strings.TrimSpace(s)
	}
	return &nascq.EvidenceOffer{ClaimID: parts["claim_id"], ArgumentID: parts["argument_id"], Kind: parts["kind"], Statement: parts["statement"],
		Role: parts["role"], Interpretation: parts["interpretation"], Warrant: parts["warrant"]}, nil
}

// bindingRefsOf reads the bindings a move cites: a list of binding ids.
// Another value, or an item that is not text, is refused by name.
func bindingRefsOf(m map[string]interface{}) ([]string, error) {
	v, exists := m["evidence_refs"]
	if !exists || v == nil {
		return nil, nil
	}
	list, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("\"evidence_refs\" must be a list of binding ids, not %s", jsonKind(v))
	}
	var out []string
	for _, x := range list {
		s, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("\"evidence_refs\": a binding id must be text, not %s", jsonKind(x))
		}
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out, nil
}

// sourcesOf reads a move's sources. Sources are text: a citation or a URL.
// null, absent or an empty string is no source (the evidence rule then says
// whether one is needed); any other value is refused, since printing it
// ("<nil>", "false", "map[url:...]") would pass for a citation.
func sourcesOf(m map[string]interface{}) ([]string, error) {
	var out []string
	for _, k := range []string{"evidence", "sources"} {
		switch ev := m[k].(type) {
		case nil:
		case []interface{}:
			for _, x := range ev {
				switch c := x.(type) {
				case nil:
				case string:
					if s := strings.TrimSpace(c); s != "" {
						out = append(out, s)
					}
				default:
					return nil, fmt.Errorf("%q: each source must be text (a citation or a URL), not %s", k, jsonKind(x))
				}
			}
		case string:
			if strings.TrimSpace(ev) != "" {
				out = append(out, strings.TrimSpace(ev))
			}
		default:
			return nil, fmt.Errorf("%q must be text or a list of texts (citations or URLs), not %s", k, jsonKind(ev))
		}
	}
	return out, nil
}

// declarationsOf reads the actions a move declares (D094 (b)): a list of
// objects, each with its id, its actor and its aliases if any.
func declarationsOf(m map[string]interface{}) ([]nascq.ActionDecl, error) {
	var out []nascq.ActionDecl
	switch d := m["declare"].(type) {
	case nil:
	case []interface{}:
		for _, x := range d {
			switch v := x.(type) {
			case map[string]interface{}:
				decl := nascq.ActionDecl{}
				for _, k := range nascq.SortedStringKeys(v) { // sorted: see moveOf
					f := v[k]
					switch k {
					case "id", "actor":
						t, ok := f.(string)
						if !ok {
							return nil, fmt.Errorf(`"declare": an action's %q must be text, not %s`, k, jsonKind(f))
						}
						if k == "id" {
							decl.ID = strings.TrimSpace(t)
						} else {
							decl.Actor = strings.TrimSpace(t)
						}
					case "aliases":
						list, ok := f.([]interface{})
						if f != nil && !ok {
							return nil, fmt.Errorf(`"declare": an action's aliases must be a list of texts, not %s`, jsonKind(f))
						}
						for _, y := range list {
							t, ok := y.(string)
							if !ok {
								return nil, fmt.Errorf(`"declare": each alias must be text, not %s`, jsonKind(y))
							}
							if t = strings.TrimSpace(t); t != "" {
								decl.Aliases = append(decl.Aliases, t)
							}
						}
					default:
						return nil, fmt.Errorf(`"declare": an action has an id, an actor and aliases, not %q`, k)
					}
				}
				out = append(out, decl)
			default:
				return nil, fmt.Errorf(`"declare": each action must be an object with its id and actor, not %s`, jsonKind(x))
			}
		}
	default:
		return nil, fmt.Errorf(`"declare" must be a list of actions, not %s`, jsonKind(d))
	}
	return out, nil
}

// comparisonOf reads how a move's alternative is better (D089).
func comparisonOf(m map[string]interface{}) (string, error) {
	switch c := m["comparison"].(type) {
	case nil:
		return "", nil
	case string:
		return strings.ToLower(strings.TrimSpace(c)), nil
	default:
		return "", fmt.Errorf(`"comparison" must be text ("surely", "sooner" or "cheaper"), not %s`, jsonKind(c))
	}
}

// strs reads a string or a list of strings.
func strs(v interface{}) []string {
	switch x := v.(type) {
	case string:
		if s := strings.TrimSpace(x); s != "" {
			return []string{s}
		}
	case []interface{}:
		var out []string
		for _, y := range x {
			if s := strings.TrimSpace(fmt.Sprint(y)); s != "" && y != nil {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func str1(v interface{}) string {
	return strings.Join(strs(v), " ")
}

// number reads a finite number: NaN or an infinity ("NaN", "inf", 1e400) is
// no number, since a record holding one cannot be saved as JSON.
func number(v interface{}) (float64, bool) {
	var f float64
	switch x := v.(type) {
	case json.Number:
		var err error
		if f, err = x.Float64(); err != nil {
			return 0, false
		}
	case float64:
		f = x
	case string:
		if _, err := fmt.Sscanf(strings.TrimSpace(x), "%g", &f); err != nil {
			return 0, false
		}
	default:
		return 0, false
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}

func expectations(v interface{}) []Expectation {
	list, _ := v.([]interface{})
	var out []Expectation
	for _, x := range list {
		if m, ok := x.(map[string]interface{}); ok {
			e := Expectation{Stmt: str1(m["stmt"]), Label: strings.ToLower(str1(m["label"]))}
			if e.Stmt == "" {
				e.Stmt = str1(m["statement"])
			}
			if e.Stmt != "" && e.Label != "" {
				out = append(out, e)
			}
		}
	}
	return out
}

func alternatives(v interface{}, evidence bool) []Alternative {
	list, _ := v.([]interface{})
	var out []Alternative
	for _, x := range list {
		m, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		alt := Alternative{WhyNot: str1(m["why_not"])}
		am, ok := m["action"].(map[string]interface{})
		if !ok {
			am = m
		}
		// A considered move with a part that does not read is still recorded,
		// without that part (moveOf keeps the others).
		if a, ok, _ := moveOf(am, evidence); ok && a.Type != "" {
			alt.Action = a
			out = append(out, alt)
		}
	}
	return out
}

// refs numbers the statements on the board, s1, s2, ..., in the order the
// game introduced them (the claim first), so that a reference keeps its
// number for the whole game.
type refs struct {
	order []string
	num   map[string]int
}

func newRefs(gs *game.GameState) *refs {
	on := map[string]bool{}
	for _, id := range gs.BoardStatements() {
		on[id] = true
	}
	r := &refs{num: map[string]int{}}
	add := func(id string) {
		if on[id] && r.num[id] == 0 {
			r.order = append(r.order, id)
			r.num[id] = len(r.order)
		}
	}
	add(gs.Claim)
	visit := func(argID string) {
		// Only arguments made by moves: an exception statement is numbered
		// when someone argues about it, not when its scheme declared it, so
		// a number, once given, never changes.
		if arg := gs.Adapter.Graph.Arguments[argID]; arg != nil && carneades.ArgumentRole(arg) != carneades.RoleException &&
			carneades.ArgumentRole(arg) != carneades.RolePresumption && carneades.ArgumentRole(arg) != carneades.RoleQuestion {
			if arg.Conclusion != nil {
				add(arg.Conclusion.Id)
				add(strings.TrimPrefix(arg.Conclusion.Id, "¬"))
			}
			for _, p := range arg.Premises {
				if p.Stmt != nil {
					add(p.Stmt.Id)
				}
			}
		}
	}
	for _, m := range gs.MoveHistory {
		for _, id := range m.ArgIDs {
			visit(id)
		}
	}
	var rest []string
	for id := range gs.Adapter.Graph.Arguments {
		rest = append(rest, id)
	}
	carneades.SortArgumentIDs(rest)
	for _, id := range rest { // arguments not made by moves (an imported discussion)
		visit(id)
	}
	return r
}

// ref is the reference of a statement as the prompt shows it, "[s3]", or ""
// if it has none.
func (r *refs) ref(id string) string {
	if x := r.id(id); x != "" {
		return "[" + x + "]"
	}
	return ""
}

// id is the reference of a statement as a move gives it, "s3", or "" if it
// has none; the negation ¬S of a numbered statement S is "¬s3".
func (r *refs) id(id string) string {
	if n := r.num[id]; n > 0 {
		return fmt.Sprintf("s%d", n)
	}
	if pos, neg := strings.CutPrefix(id, "¬"); neg && r.num[pos] > 0 {
		return fmt.Sprintf("¬s%d", r.num[pos])
	}
	return ""
}

// resolve returns the statement a reference ("s3", "[s3]", "¬s3") names, or
// id itself if it is not a reference.
func (r *refs) resolve(id string) string {
	x := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(id), "["), "]")
	if pos, neg := strings.CutPrefix(x, "¬"); neg {
		if s := r.resolve(pos); s != pos {
			return "¬" + s
		}
		return id
	}
	var n int
	if _, err := fmt.Sscanf(x, "s%d", &n); err == nil && fmt.Sprintf("s%d", n) == x && n >= 1 && n <= len(r.order) {
		return r.order[n-1]
	}
	return id
}

// resolveAction replaces the references in a move by statement IDs.
func (r *refs) resolveAction(a nascq.ProposedAction) nascq.ProposedAction {
	a.TargetStmtID = r.resolve(a.TargetStmtID)
	a.TargetPremiseID = r.resolve(a.TargetPremiseID)
	if a.Offer != nil {
		o := *a.Offer
		o.Statement = r.resolve(o.Statement)
		a.Offer = &o
	}
	return a
}

// jsonCandidates returns outermost object/array spans, including a trailing
// unfinished span. Quoted braces/brackets inside a container are not boundaries.
// Including unfinished candidates prevents fallback to an earlier example.
func jsonCandidates(s string) []string {
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
		case r == '{' || r == '[':
			if depth == 0 {
				// A bracket in prose ("[s3]", "[1]", "[a link](…)") is not
				// a move container; a JSON array that could hold or be a
				// move starts with {, [, " or is empty.
				if r == '[' && !arrayStart(s[i+1:]) {
					continue
				}
				start = i
			}
			depth++
		case r == '}' || r == ']':
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					out = append(out, s[start:i+1])
					start = -1
				}
			}
		}
	}
	if depth > 0 && start >= 0 {
		out = append(out, s[start:])
	}
	return out
}

// arrayStart reports whether the text after a '[' begins like a JSON array
// of containers or strings, or an empty array.
func arrayStart(rest string) bool {
	if strings.HasPrefix(rest, "]") {
		return true // []
	}
	t := strings.TrimLeft(rest, " \t\r\n")
	return t != "" && strings.ContainsRune("{[\"", rune(t[0]))
}

// jsonKind names the JSON type of a decoded value, for error messages.
func jsonKind(v interface{}) string {
	switch v.(type) {
	case bool:
		return "true or false"
	case float64, json.Number:
		return "a number"
	case map[string]interface{}:
		return "an object"
	case []interface{}:
		return "a list"
	case nil:
		return "null"
	}
	return fmt.Sprintf("%T", v)
}
