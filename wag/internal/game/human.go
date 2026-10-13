package game

// human.go implements a HumanAgent that plays WAG through an interactive
// CLI wizard. It implements the Agent interface so it slots into the same
// game loop as AI agents.
//
// The wizard guides the player step-by-step:
//   1. Show game state (arguments, scores, round, open questions)
//   2. Choose what to do (argue / attack / defend / question / pass / resign)
//   3. To argue: select scheme → fill variables
//   4. To attack or defend: select target → select CQ → select backup scheme
//      → fill variables; to question: select target → select CQ
//   5. Return ProposedAction
//
// IO is injectable (io.Reader / io.Writer) for testability. Pass os.Stdin
// and os.Stdout for real CLI use.

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/nascq"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// HumanAgent implements Agent via interactive CLI prompts.
type HumanAgent struct {
	name    string
	scanner *bufio.Scanner
	out     io.Writer
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
// constructing a valid move via numbered menus.
func (h *HumanAgent) ChooseAction(obs *Observation) nascq.ProposedAction {
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
		if tmpl.TargetArgID != "" {
			has[tmpl.Type] = true
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
	h.printf("    [2] Attack   — answer a critical question against an argument%s\n", none("attack"))
	h.printf("    [3] Defend   — answer a critical question in an argument's favour%s\n", none("support"))
	h.printf("    [4] Question — ask a critical question; its proponent must answer%s\n", none("question"))
	h.print("    [5] Pass\n")
	h.print("    [6] Resign\n")

	switch h.readChoice(1, 6) {
	case 1:
		return h.chooseSupportAction(obs)
	case 2:
		return h.chooseTargeted(obs, "attack")
	case 3:
		return h.chooseTargeted(obs, "support")
	case 4:
		return h.chooseTargeted(obs, "question")
	case 6:
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
		if tmpl.Type == "support" && tmpl.TargetArgID == "" && !seen[tmpl.SchemeID] {
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
	h.preview(obs.GameState, selected, bindings)

	return nascq.ProposedAction{
		Type:             "support",
		SchemeID:         selected.SchemeID,
		VariableBindings: bindings,
	}
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
		weight := 0.0
		if scheme := gs.NASCQ.GetScheme(b.SchemeID); scheme != nil {
			weight = scheme.BaseWeight
		}
		h.printf("    [%d] %s  (weight %.2f)\n", i+1, b.SchemeName, weight)
	}
	selected := backups[h.readChoice(1, len(backups))-1]

	// --- Step 4: the variables ---
	bindings := h.fillVariables(selected)
	h.preview(gs, selected, bindings)
	return nascq.ProposedAction{
		Type:             typ,
		AttackType:       selected.AttackType,
		SchemeID:         selected.SchemeID,
		AnswersCQID:      selected.AnswersCQID,
		TargetArgID:      selected.TargetArgID,
		TargetPremiseID:  selected.TargetPremiseID,
		VariableBindings: bindings,
	}
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
		if strings.TrimSpace(value) == "" {
			// Fallback: use lowercase variable name to avoid empty bindings.
			value = strings.ToLower(name)
			h.printf("    (using default: %q)\n", value)
		}
		// Not sanitized here: the adapter turns each value into an atom,
		// and the board shows the words typed (GameState.Say).
		bindings[name] = value
	}

	return bindings
}

// preview shows the move before it is submitted: the instantiated premises
// and what follows from them.
func (h *HumanAgent) preview(gs *GameState, tmpl nascq.ActionTemplate, bindings map[string]string) {
	shown := bindings
	if tmpl.AnswerVariable != "" {
		shown = map[string]string{tmpl.AnswerVariable: "“" + lowerFirst(gs.answerText(tmpl, gs.Adapter.Graph.Arguments[tmpl.TargetArgID])) + "”"}
		for k, v := range bindings {
			if k != tmpl.AnswerVariable {
				shown[k] = v
			}
		}
	}
	h.print("\n  You argue:\n")
	for _, p := range tmpl.PremiseTemplatesEN {
		h.printf("      %s\n", nascq.InstantiateTemplate(p, shown))
	}
	h.printf("    therefore: %s\n", gs.Effect(tmpl, bindings))
}

// variableOrder returns the template's variables in the order its premise and
// conclusion templates first use them, then any others in sorted order;
// without the answer variable, which the game binds.
func variableOrder(tmpl nascq.ActionTemplate) []string {
	seen := map[string]bool{tmpl.AnswerVariable: true}
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
func (gs *GameState) QuestionText(cqID, targetArgID string) string {
	cq := gs.NASCQ.GetCQ(cqID)
	if cq == nil {
		return ""
	}
	words := gs.words()
	b := gs.Adapter.ArgumentBindings(targetArgID)
	for k, v := range b {
		if w, ok := words[v]; ok {
			b[k] = w
		} else if t, ok := terms.ReadString(v); ok {
			b[k] = gs.sayTerm(t, words, true)
		}
	}
	return nascq.InstantiateTemplate(cq.QuestionTemplateEN, b)
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
