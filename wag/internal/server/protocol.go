package server

import (
	"encoding/json"
	"fmt"
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

	w("# WAG: the Waltonian Argumentation Game. It is your move. (%s)\n\n", ProtocolVersion)
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
	if gs.Config.Objective == game.ObjectiveClaim {
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
				w("  premise %s%s [%s]\n", spaced(ref(p)), textAt(a.PremiseTexts, i, p), strings.ToUpper(st.Statements[p].Label))
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

	w("## Your legal moves\n")
	ss.promptQuestions(w, st, ids, me.ID, refs)
	ss.promptStatements(w, st, ids, refs)
	w("\n### New argument\nAny scheme, with no target: {\"type\":\"support\",\"scheme_id\":…,\"variable_bindings\":{…}}. In a debate, pro's first argument states the claim.\n")
	w("\n### Pass or resign\n{\"type\":\"abstain\"} passes")
	if gs.Config.ConcedeOnPass {
		w(" (in a debate, passing while the verdict is against your side concedes)")
	}
	w("; {\"type\":\"resign\"} gives up.\n\n")

	ss.promptSchemes(w)
	w("%s", replyFormat)
	return b.String()
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
			l += fmt.Sprintf("\n    Q: %s\n    attack argues: %s\n    defence argues: %s", gs.QuestionAbout(o.CQ, t, o.Stmt), o.AttackText, o.DefenceText)
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

// promptSchemes lists the schemes: parts, sentences, assumptions,
// exceptions, sources, semantic type and the cue words that signal them.
func (ss *Session) promptSchemes(w func(string, ...interface{})) {
	gs := ss.GS
	w("## The schemes\n")
	for _, id := range nascq.SortedStringKeys(gs.NASCQ.Schemes) {
		sc := gs.NASCQ.Schemes[id]
		w("\n%s — %s (%s, weight %.2f)", id, sc.Name, sc.SemanticType, sc.BaseWeight)
		if sc.ThesisIndependent {
			w(", can back an answer to any question")
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
				w("  unless (exceptions the other side may argue): %s\n", strings.Join(r.Exceptions, "; "))
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
// last object that is one or the other counts, so a reply may think aloud
// first, quote examples, or wrap the JSON in a code block. Common variations
// of the field names are accepted ("move" for "action"; in the move,
// "scheme" for "scheme_id", "target" for "target_arg_id", "cq" or
// "question" for "answers_cq_id", "premise" for "target_premise_id",
// "statement" or "stmt" for "target_stmt_id", "slots" or "bindings" for
// "variable_bindings", "sources" for "evidence"; "pass" for "abstain",
// "argue" and "defend" for "support").
func ParseReplyFull(reply string) (*Reply, error) {
	objs := jsonObjects(reply)
	for i := len(objs) - 1; i >= 0; i-- {
		var m map[string]interface{}
		if json.Unmarshal([]byte(objs[i]), &m) != nil {
			continue
		}
		for _, k := range []string{"action", "move"} {
			if am, ok := m[k].(map[string]interface{}); ok {
				if a, ok := moveOf(am); ok {
					r := &Reply{Action: a, Context: strs(m["context"]), Thoughts: strs(m["thoughts"]), Summary: str1(m["summary"])}
					r.Expect = expectations(m["expect"])
					r.Alternatives = alternatives(m["alternatives"])
					if c, ok := number(m["confidence"]); ok {
						r.Confidence = &c
					}
					return r, nil
				}
			}
		}
		if a, ok := moveOf(m); ok {
			return &Reply{Action: a, Bare: true}, nil
		}
	}
	return nil, fmt.Errorf(`no move found: end the reply with one JSON object such as {"action": {"type": "abstain"}}`)
}

// moveOf reads a move object; ok is false if it has no type.
func moveOf(m map[string]interface{}) (nascq.ProposedAction, bool) {
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
		return nascq.ProposedAction{}, false
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
	a := nascq.ProposedAction{Type: t,
		SchemeID:        str("scheme_id", "scheme"),
		TargetArgID:     str("target_arg_id", "target", "argument"),
		AnswersCQID:     str("answers_cq_id", "cq", "question", "critical_question"),
		TargetPremiseID: str("target_premise_id", "premise"),
		TargetStmtID:    str("target_stmt_id", "statement", "stmt", "target_statement"),
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
	return a, true
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

func number(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		var f float64
		if _, err := fmt.Sscanf(strings.TrimSpace(x), "%g", &f); err == nil {
			return f, true
		}
	}
	return 0, false
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

func alternatives(v interface{}) []Alternative {
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
		if a, ok := moveOf(am); ok {
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
	return a
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
