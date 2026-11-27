package game

// human.go implements a HumanAgent that plays WAG through an interactive
// CLI wizard. It implements the Agent interface so it slots into the same
// game loop as AI agents.
//
// The wizard guides the player step-by-step:
//   1. Show game state (arguments, scores, round, open questions)
//   2. Choose what to do (argue / for or against a statement / attack /
//      defend / question / pass / resign)
//   3. To argue: select scheme → fill variables
//   4. To argue for or against a statement on the board: select the
//      statement → the side → a scheme that can conclude it → fill variables
//      (no critical question: a rebuttal of a conclusion or an undermining
//      of a premise, docs/game-rules.md)
//   5. To attack or defend through a critical question: select target →
//      select CQ → select backup scheme → fill variables; to question:
//      select target → select CQ
//   6. Return ProposedAction
//
// IO is injectable (io.Reader / io.Writer) for testability. Pass os.Stdin
// and os.Stdout for real CLI use.

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/language"
	"github.com/yourorg/wag/internal/nascq"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// HumanAgent implements Agent via interactive CLI prompts.
type HumanAgent struct {
	name    string
	scanner *bufio.Scanner
	out     io.Writer
	eof     bool // the input has ended
}

// NewHumanAgent creates a human player with the given IO channels.
// For CLI: NewHumanAgent("Alice", os.Stdin, os.Stdout)
// For tests: NewHumanAgent("Alice", strings.NewReader("1\n2\n..."), &buf)
func NewHumanAgent(name string, in io.Reader, out io.Writer) *HumanAgent {
	return &HumanAgent{
		name:    name,
		scanner: bufio.NewScanner(in),
		out:     out,
	}
}

// Compile-time check that HumanAgent implements Agent.
var _ Agent = (*HumanAgent)(nil)

func (h *HumanAgent) Name() string { return h.name }

// ChooseAction displays the game state and walks the player through
// constructing a valid move via numbered menus. Once its input has ended,
// no one is left to play the seat, which resigns: asked again, it would
// read nothing, and a move made of nothing is refused again and again.
func (h *HumanAgent) ChooseAction(obs *Observation) nascq.ProposedAction {
	if h.eof {
		return h.resignAtEnd()
	}
	a := h.choose(obs)
	if h.eof {
		return h.resignAtEnd()
	}
	return a
}

func (h *HumanAgent) resignAtEnd() nascq.ProposedAction {
	h.printf("\n  The input has ended: %s resigns.\n", h.name)
	return nascq.ProposedAction{Type: "resign"}
}

// choose walks the player through a move.
func (h *HumanAgent) choose(obs *Observation) nascq.ProposedAction {
	gs := obs.GameState
	h.print(RenderState(gs))

	// The questions this player must answer come first.
	for _, q := range gs.Questions {
		if q.Open() && q.Owner == obs.ActivePlayer.ID {
			h.printf("\n  ! %s asks you about %s: %s\n    Answer it: Defend → %s → %s.\n",
				q.Asker, q.TargetArgID, gs.QuestionText(q.CQID, q.TargetArgID), q.TargetArgID, gs.NASCQ.GetCQ(q.CQID).Name)
		}
	}

	has := map[string]bool{}
	for _, tmpl := range obs.AvailableActions {
		if tmpl.TargetArgID != "" { // the moves about arguments, through their questions
			has[tmpl.Type] = true
		}
		if tmpl.TargetStmtID != "" { // the moves about statements, with no question
			has["statement"] = true
		}
	}
	none := func(typ string) string {
		if has[typ] {
			return ""
		}
		return " (nothing to do yet)"
	}
	h.print("\n  What would you like to do?\n")
	h.print("    [1] Argue    — put forward a new argument\n")
	h.printf("    [2] For or against — argue for or against a statement on the board, no question needed%s\n", none("statement"))
	h.printf("    [3] Attack   — answer a critical question against an argument%s\n", none("attack"))
	h.printf("    [4] Defend   — answer a critical question in an argument's favour%s\n", none("support"))
	h.printf("    [5] Question — ask a critical question; its proponent must answer%s\n", none("question"))
	h.print("    [6] Pass\n")
	h.print("    [7] Resign\n")

	switch h.readChoice(1, 7) {
	case 1:
		return h.chooseSupportAction(obs)
	case 2:
		return h.chooseStatementAction(obs)
	case 3:
		return h.chooseTargeted(obs, "attack")
	case 4:
		return h.chooseTargeted(obs, "support")
	case 5:
		return h.chooseTargeted(obs, "question")
	case 7:
		return nascq.ProposedAction{Type: "resign"}
	default:
		return nascq.ProposedAction{Type: "abstain"}
	}
}

// -------------------------------------------------------------------------
// Support wizard
// -------------------------------------------------------------------------

func (h *HumanAgent) chooseSupportAction(obs *Observation) nascq.ProposedAction {
	// Collect unique support schemes.
	seen := make(map[string]bool)
	var schemes []nascq.ActionTemplate
	for _, tmpl := range obs.AvailableActions {
		if tmpl.Type == "support" && tmpl.TargetArgID == "" && tmpl.TargetStmtID == "" && !seen[tmpl.SchemeID] {
			seen[tmpl.SchemeID] = true
			schemes = append(schemes, tmpl)
		}
	}

	if len(schemes) == 0 {
		h.print("\n  No support schemes available.\n")
		return nascq.ProposedAction{Type: "abstain"}
	}

	h.print("\n  Choose a scheme:\n")
	for i, s := range schemes {
		varNames := variableOrder(s)
		h.printf("    [%d] %s  (%d vars: %s)\n",
			i+1, s.SchemeName, len(s.Variables), strings.Join(varNames, ", "))
	}

	idx := h.readChoice(1, len(schemes)) - 1
	selected := schemes[idx]

	bindings := h.fillVariables(selected)
	a := nascq.ProposedAction{
		Type:             "support",
		SchemeID:         selected.SchemeID,
		VariableBindings: bindings,
		Declare:          h.declarations(obs.GameState, selected, bindings),
	}
	h.preview(obs.GameState, selected, a)
	return a
}

// -------------------------------------------------------------------------
// Targeted wizard: target → critical question → backup scheme → variables
// -------------------------------------------------------------------------

// chooseTargeted builds a move about an existing argument: an attack, a
// defence (typ "support") or a question (which has no scheme).
func (h *HumanAgent) chooseTargeted(obs *Observation, typ string) nascq.ProposedAction {
	gs := obs.GameState
	var tmpls []nascq.ActionTemplate
	for _, tmpl := range obs.AvailableActions {
		if tmpl.Type == typ && tmpl.TargetArgID != "" {
			tmpls = append(tmpls, tmpl)
		}
	}
	if len(tmpls) == 0 {
		h.print("\n  There is nothing to do that way yet: argue instead.\n")
		return h.chooseSupportAction(obs)
	}

	// --- Step 1: the argument ---
	var targets []string
	seen := map[string]bool{}
	for _, tmpl := range tmpls {
		if !seen[tmpl.TargetArgID] {
			seen[tmpl.TargetArgID] = true
			targets = append(targets, tmpl.TargetArgID)
		}
	}
	h.print("\n  Which argument?\n")
	for i, id := range targets {
		arg := gs.Adapter.Graph.Arguments[id]
		h.printf("    [%d] %s: %s (%s) — %s by %s\n", i+1, id, gs.Say(arg.Conclusion.Id),
			strings.TrimSpace(labelString(arg.Conclusion.Label)), gs.SchemeName(arg), ownerOf(arg))
	}
	target := targets[h.readChoice(1, len(targets))-1]

	// --- Step 2: the critical question ---
	var cqs []nascq.ActionTemplate
	seen = map[string]bool{}
	for _, tmpl := range tmpls {
		if tmpl.TargetArgID == target && !seen[tmpl.AnswersCQID] {
			seen[tmpl.AnswersCQID] = true
			cqs = append(cqs, tmpl)
		}
	}
	h.print("\n  Which critical question?\n")
	for i, c := range cqs {
		asked := ""
		for _, q := range gs.Questions {
			if q.Open() && q.TargetArgID == target && q.CQID == c.AnswersCQID {
				asked = fmt.Sprintf("  [asked by %s]", q.Asker)
			}
		}
		h.printf("    [%d] %s — %s (about the %s)%s\n", i+1, c.CQName, gs.QuestionText(c.AnswersCQID, target), c.Locus.Kind, asked)
	}
	cq := cqs[h.readChoice(1, len(cqs))-1]
	if typ == "question" {
		h.printf("\n  %s\n", gs.Effect(cq, nil))
		return nascq.ProposedAction{Type: "question", AnswersCQID: cq.AnswersCQID, TargetArgID: target}
	}

	// --- Step 3: the scheme backing the answer ---
	var backups []nascq.ActionTemplate
	for _, tmpl := range tmpls {
		if tmpl.TargetArgID == target && tmpl.AnswersCQID == cq.AnswersCQID {
			backups = append(backups, tmpl)
		}
	}
	h.printf("\n  Answer with which scheme? The answer: %s\n", gs.Effect(cq, nil))
	for i, b := range backups {
		weight := gs.MoveWeight(nascq.ProposedAction{SchemeID: b.SchemeID, AnswersCQID: b.AnswersCQID, TargetArgID: target})
		h.printf("    [%d] %s  (weight %.2f)\n", i+1, b.SchemeName, weight)
	}
	selected := backups[h.readChoice(1, len(backups))-1]

	// --- Step 4: the variables ---
	bindings := h.fillVariables(selected)
	a := nascq.ProposedAction{
		Type:             typ,
		AttackType:       selected.AttackType,
		SchemeID:         selected.SchemeID,
		AnswersCQID:      selected.AnswersCQID,
		TargetArgID:      selected.TargetArgID,
		TargetPremiseID:  selected.TargetPremiseID,
		VariableBindings: bindings,
		Declare:          h.declarations(gs, selected, bindings),
	}
	h.preview(gs, selected, a)
	return a
}

// -------------------------------------------------------------------------
// Statement wizard: statement → side → scheme → variables
// -------------------------------------------------------------------------

// chooseStatementAction builds a move about a statement on the board: an
// argument for it, or for its opposite (a rebuttal of a conclusion, an
// undermining of a premise), with a scheme that can conclude it (S9). No
// critical question is named: such moves are the game's as much as the
// answers to questions are (docs/game-rules.md).
func (h *HumanAgent) chooseStatementAction(obs *Observation) nascq.ProposedAction {
	gs := obs.GameState
	var stmts []string
	seen := map[string]bool{}
	for _, tmpl := range obs.AvailableActions {
		if tmpl.TargetStmtID != "" && !seen[tmpl.TargetStmtID] {
			seen[tmpl.TargetStmtID] = true
			stmts = append(stmts, tmpl.TargetStmtID)
		}
	}
	if len(stmts) == 0 {
		h.print("\n  There is no statement on the board yet: argue instead.\n")
		return h.chooseSupportAction(obs)
	}

	// --- Step 1: the statement ---
	h.print("\n  Which statement?\n")
	for i, id := range stmts {
		label := ""
		if st := gs.Adapter.Graph.Statements[id]; st != nil {
			label = " (" + strings.TrimSpace(labelString(st.Label)) + ")"
		}
		h.printf("    [%d] %s%s\n", i+1, gs.Say(id), label)
	}
	stmt := stmts[h.readChoice(1, len(stmts))-1]

	// --- Step 2: the side ---
	var sides []string
	for _, typ := range []string{"support", "attack"} {
		for _, tmpl := range obs.AvailableActions {
			if tmpl.TargetStmtID == stmt && tmpl.Type == typ {
				sides = append(sides, typ)
				break
			}
		}
	}
	if len(sides) == 0 {
		h.print("\n  No scheme can argue about that statement.\n")
		return nascq.ProposedAction{Type: "abstain"}
	}
	typ := sides[0]
	if len(sides) == 2 {
		h.print("\n  For it or against it?\n    [1] For it\n    [2] Against it\n")
		typ = sides[h.readChoice(1, 2)-1]
	} else if typ == "support" {
		h.print("\n  Only an argument for it is possible here.\n")
	} else {
		h.print("\n  Only an argument against it is possible here.\n")
	}

	// --- Step 3: the scheme ---
	var schemes []nascq.ActionTemplate
	for _, tmpl := range obs.AvailableActions {
		if tmpl.TargetStmtID == stmt && tmpl.Type == typ {
			schemes = append(schemes, tmpl)
		}
	}
	h.printf("\n  Argue with which scheme? The argument: %s\n", gs.Effect(schemes[0], nil))
	for i, b := range schemes {
		weight := gs.MoveWeight(nascq.ProposedAction{SchemeID: b.SchemeID})
		h.printf("    [%d] %s  (weight %.2f)\n", i+1, b.SchemeName, weight)
	}
	selected := schemes[h.readChoice(1, len(schemes))-1]

	// --- Step 4: the variables ---
	bindings := h.fillVariables(selected)
	a := selected.Action(bindings)
	a.Declare = h.declarations(gs, selected, bindings)
	h.preview(gs, selected, a)
	return a
}

// -------------------------------------------------------------------------
// Variable filling
// -------------------------------------------------------------------------

// fillVariables asks for the template's variables in the order the premise
// and conclusion templates first use them (not the answer variable, which
// the game binds to the answer).
func (h *HumanAgent) fillVariables(tmpl nascq.ActionTemplate) map[string]string {
	vars := tmpl.Variables
	bindings := make(map[string]string, len(vars))
	names := variableOrder(tmpl)

	h.print("\n  Fill in the variables:\n")
	for _, name := range names {
		spec := vars[name]
		h.printf("    %s (%s — %s): ", name, spec.Type, spec.Description)
		value := h.readLine()
		// A value that names nothing (blank, punctuation alone) binds
		// nothing (S2): ask again, rather than fill in the variable's letter,
		// which the game would take as a name. Other's premises after the
		// first may be left blank. At the end of the input the move is not
		// made: the seat resigns (ChooseAction).
		optional := tmpl.SchemeID == nascq.OtherSchemeID && name != "P1" && strings.HasPrefix(name, "P")
		for language.NamesNothing(value) && !(optional && value == "") && !h.eof {
			h.printf("    %s names nothing: type a word, or a number in digits: ", name)
			value = h.readLine()
		}
		// Not sanitized here: the adapter turns each value into an atom,
		// and the board shows the words typed (GameState.Say).
		bindings[name] = value
	}

	return bindings
}

// declarations are the declarations of the new actions a move binds (D094
// (b)): for each action place whose value the game's profile does not hold,
// the person says whose action it is, the agent's unless they choose
// another actor, and the move declares it.
func (h *HumanAgent) declarations(gs *GameState, tmpl nascq.ActionTemplate, bindings map[string]string) []nascq.ActionDecl {
	if gs.Profile == nil {
		return nil
	}
	var out []nascq.ActionDecl
	seen := map[string]bool{}
	actors := gs.Profile.actorIDs()
	for _, v := range variableOrder(tmpl) {
		x := strings.TrimSpace(bindings[v])
		if tmpl.Variables[v].Type != nascq.VarAction || x == "" || seen[carneades.TextToAtom(x)] {
			continue
		}
		seen[carneades.TextToAtom(x)] = true
		if _, ok := gs.Profile.Action(x); ok {
			continue
		}
		actor := gs.Profile.Agent
		if len(actors) > 1 {
			h.printf("\n  %q is a new action: whose action is it?\n", x)
			order := append([]string{actor}, without(actors, actor)...)
			for i, a := range order {
				note := ""
				if a == actor {
					note = " (the agent, who decides)"
				}
				h.printf("    [%d] %s%s\n", i+1, a, note)
			}
			actor = order[h.readChoice(1, len(order))-1]
		} else {
			h.printf("    (%q is a new action: the move declares it as %s's, the agent's)\n", x, actor)
		}
		out = append(out, nascq.ActionDecl{ID: x, Actor: actor})
	}
	return out
}

// without is a list without one of its strings.
func without(list []string, x string) []string {
	var out []string
	for _, s := range list {
		if s != x {
			out = append(out, s)
		}
	}
	return out
}

// preview shows the move before it is submitted: its argument as it would
// read on the board, and what it establishes.
func (h *HumanAgent) preview(gs *GameState, tmpl nascq.ActionTemplate, a nascq.ProposedAction) {
	bindings := a.VariableBindings
	rd, err := gs.Read(a)
	if err != nil {
		h.printf("\n  This move would be rejected: %v\n", err)
		return
	}
	h.printf("\n  You argue that C: %s\n  because:\n", rd.Conclusion)
	for _, p := range rd.Premises {
		h.printf("      %s\n", p)
	}
	if tmpl.TargetArgID != "" || tmpl.TargetStmtID != "" {
		h.printf("    therefore: %s\n", gs.Effect(tmpl, bindings))
	}
}

// variableOrder returns the template's variables in the order its premise and
// conclusion templates first use them, then any others in sorted order;
// without the variables the game binds (the answer variable, and those the
// statement of a move about a statement fixes).
func variableOrder(tmpl nascq.ActionTemplate) []string {
	seen := map[string]bool{tmpl.AnswerVariable: true}
	for v := range tmpl.Fixed {
		seen[v] = true
	}
	var out []string
	for _, t := range append(append([]string{}, tmpl.PremiseTemplatesEN...), tmpl.ConclusionTemplateEN) {
		for _, v := range nascq.Placeholders(t) {
			if _, ok := tmpl.Variables[v]; ok && !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	for _, v := range sortedKeys(tmpl.Variables) {
		if !seen[v] {
			out = append(out, v)
		}
	}
	return out
}

// QuestionText is the critical question's text with the target argument's
// own bindings in place of its {Var} placeholders, as far as the target has
// them (a question argument has none), in the words players typed.
// QuestionAbout is QuestionText for a question about one statement of the
// target (a premise or an exception): the generic premise question, which
// has no words of its own for the premise, names it.
func (gs *GameState) QuestionAbout(cqID, targetArgID, stmtID string) string {
	if cqID == nascq.CQPremiseAcceptable && stmtID != "" {
		return "Why should we accept that “" + strings.TrimSuffix(gs.SayIn(targetArgID, stmtID), ".") + "”?"
	}
	return gs.QuestionText(cqID, targetArgID)
}

func (gs *GameState) QuestionText(cqID, targetArgID string) string {
	cq := gs.NASCQ.GetCQ(cqID)
	if cq == nil {
		return ""
	}
	words := gs.words()
	b := gs.Adapter.ArgumentBindings(targetArgID)
	// An Other move about an existing statement can omit C. Recover its
	// actual conclusion for this question without changing recorded bindings.
	if cqID == nascq.CQCounter {
		if _, hasC := b["C"]; !hasC {
			if arg := gs.Adapter.Graph.Arguments[targetArgID]; arg != nil &&
				arg.Scheme != nil && arg.Scheme.Id == nascq.OtherCarneades &&
				arg.Conclusion != nil {
				if b == nil {
					b = make(map[string]string)
				}
				b["C"] = arg.Conclusion.Id
			}
		}
	}
	for k, v := range b {
		if t, ok := terms.ReadString(v); ok {
			b[k] = gs.sayArg(t, words, 0)
		} else if w, ok := words[v]; ok {
			b[k] = "‘" + w + "’"
		}
	}
	return sentence(nascq.InstantiateTemplate(cq.QuestionTemplateEN, b))
}

// -------------------------------------------------------------------------
// IO helpers
// -------------------------------------------------------------------------

func (h *HumanAgent) print(s string) {
	fmt.Fprint(h.out, s)
}

func (h *HumanAgent) printf(format string, args ...interface{}) {
	fmt.Fprintf(h.out, format, args...)
}

func (h *HumanAgent) readLine() string {
	if h.scanner.Scan() {
		return strings.TrimSpace(h.scanner.Text())
	}
	h.eof = true
	return ""
}

func (h *HumanAgent) readChoice(min, max int) int {
	for {
		h.printf("  > ")
		line := h.readLine()
		if line == "" {
			// EOF or empty — return min as safe default.
			return min
		}
		var n int
		if _, err := fmt.Sscanf(line, "%d", &n); err == nil && n >= min && n <= max {
			return n
		}
		h.printf("  Please enter a number between %d and %d.\n", min, max)
	}
}

// sortedKeys returns the keys of a VariableSpec map in sorted order,
// ensuring consistent display across runs.
func sortedKeys(m map[string]nascq.VariableSpec) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
