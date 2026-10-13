package game

// human.go implements a HumanAgent that plays WAG through an interactive
// CLI wizard. It implements the Agent interface so it slots into the same
// game loop as AI agents.
//
// The wizard guides the player step-by-step:
//   1. Show game state (arguments, scores, turn)
//   2. Choose action type (support / attack / abstain)
//   3. If support: select scheme → fill variables
//   4. If attack: select target → select CQ → select backup scheme → fill variables
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
	// Display current state.
	h.print(RenderState(obs.GameState))

	// Count available attack targets.
	hasAttacks := false
	for _, tmpl := range obs.AvailableActions {
		if tmpl.Type == "attack" {
			hasAttacks = true
			break
		}
	}

	// Action type menu.
	h.print("\n  What would you like to do?\n")
	h.print("    [1] Support — make a new argument\n")
	if hasAttacks {
		h.print("    [2] Attack  — challenge an existing argument\n")
	} else {
		h.print("    [2] Attack  — (no arguments to attack yet)\n")
	}
	h.print("    [3] Abstain — pass your turn\n")

	choice := h.readChoice(1, 3)

	switch choice {
	case 1:
		return h.chooseSupportAction(obs)
	case 2:
		if !hasAttacks {
			h.print("\n  No arguments to attack. Switching to support.\n")
			return h.chooseSupportAction(obs)
		}
		return h.chooseAttackAction(obs)
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
		if tmpl.Type == "support" && !seen[tmpl.SchemeID] {
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
	h.preview(selected, bindings, "therefore: "+nascq.InstantiateTemplate(selected.ConclusionTemplateEN, bindings))

	return nascq.ProposedAction{
		Type:             "support",
		SchemeID:         selected.SchemeID,
		VariableBindings: bindings,
	}
}

// -------------------------------------------------------------------------
// Attack wizard (3-level: target → CQ → backup scheme)
// -------------------------------------------------------------------------

func (h *HumanAgent) chooseAttackAction(obs *Observation) nascq.ProposedAction {
	// Collect attack templates.
	var attacks []nascq.ActionTemplate
	for _, tmpl := range obs.AvailableActions {
		if tmpl.Type == "attack" {
			attacks = append(attacks, tmpl)
		}
	}

	// --- Step 1: Choose target argument ---
	type targetInfo struct {
		argID   string
		display string // "conclusion (label) — scheme by owner"
	}

	seenTarget := make(map[string]bool)
	var targets []targetInfo
	for _, tmpl := range attacks {
		if seenTarget[tmpl.TargetArgID] {
			continue
		}
		seenTarget[tmpl.TargetArgID] = true

		arg := obs.GameState.Adapter.Graph.Arguments[tmpl.TargetArgID]
		if arg == nil {
			continue
		}

		conclusionText := tmpl.TargetArgID
		label := "?"
		if arg.Conclusion != nil {
			conclusionText = arg.Conclusion.Id
			if arg.Conclusion.Text != "" && arg.Conclusion.Text != arg.Conclusion.Id {
				conclusionText = arg.Conclusion.Text
			}
			label = arg.Conclusion.Label.String()
		}

		owner := "?"
		if arg.Metadata != nil {
			if o, ok := arg.Metadata["owner"].(string); ok {
				owner = o
			}
		}

		schemeID := "?"
		if arg.Scheme != nil {
			schemeID = arg.Scheme.Id
		}

		targets = append(targets, targetInfo{
			argID:   tmpl.TargetArgID,
			display: fmt.Sprintf("%s (%s) — %s by %s", conclusionText, label, schemeID, owner),
		})
	}

	if len(targets) == 0 {
		h.print("\n  No valid attack targets.\n")
		return nascq.ProposedAction{Type: "abstain"}
	}

	h.print("\n  Choose target argument to challenge:\n")
	for i, t := range targets {
		h.printf("    [%d] %s: %s\n", i+1, t.argID, t.display)
	}

	targetIdx := h.readChoice(1, len(targets)) - 1
	selectedTarget := targets[targetIdx]

	// --- Step 2: Choose critical question ---
	type cqOption struct {
		id   string
		name string
	}

	seenCQ := make(map[string]bool)
	var cqs []cqOption
	for _, tmpl := range attacks {
		if tmpl.TargetArgID == selectedTarget.argID && tmpl.AnswersCQID != "" && !seenCQ[tmpl.AnswersCQID] {
			seenCQ[tmpl.AnswersCQID] = true
			cqs = append(cqs, cqOption{id: tmpl.AnswersCQID, name: tmpl.CQName})
		}
	}

	if len(cqs) == 0 {
		h.print("\n  No critical questions for this argument.\n")
		return nascq.ProposedAction{Type: "abstain"}
	}

	h.print("\n  Choose critical question:\n")
	for i, cq := range cqs {
		// Show the CQ's question, with the target's own bindings filled in.
		question := cq.name
		if text := obs.GameState.QuestionText(cq.id, selectedTarget.argID); text != "" {
			question = fmt.Sprintf("%s — %s", cq.name, text)
		}
		h.printf("    [%d] %s\n", i+1, question)
	}

	cqIdx := h.readChoice(1, len(cqs)) - 1
	selectedCQ := cqs[cqIdx]

	// --- Step 3: Choose backup scheme ---
	type backupOption struct {
		tmpl nascq.ActionTemplate
	}

	seenScheme := make(map[string]bool)
	var backups []backupOption
	for _, tmpl := range attacks {
		if tmpl.TargetArgID == selectedTarget.argID &&
			tmpl.AnswersCQID == selectedCQ.id &&
			!seenScheme[tmpl.SchemeID] {
			seenScheme[tmpl.SchemeID] = true
			backups = append(backups, backupOption{tmpl: tmpl})
		}
	}

	if len(backups) == 0 {
		h.print("\n  No backup schemes available for this CQ.\n")
		return nascq.ProposedAction{Type: "abstain"}
	}

	h.print("\n  Choose backup scheme:\n")
	for i, opt := range backups {
		weight := 0.0
		scheme := obs.GameState.NASCQ.GetScheme(opt.tmpl.SchemeID)
		if scheme != nil {
			weight = scheme.BaseWeight
		}
		h.printf("    [%d] %s  (weight: %.2f, attack: %s)\n",
			i+1, opt.tmpl.SchemeName, weight, opt.tmpl.CaesAttackType)
	}

	schemeIdx := h.readChoice(1, len(backups)) - 1
	selected := backups[schemeIdx].tmpl

	// --- Step 4: Fill variables ---
	bindings := h.fillVariables(selected)
	h.preview(selected, bindings, "therefore: "+attackEffect(obs.GameState, selected))

	return nascq.ProposedAction{
		Type:             "attack",
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
// and conclusion templates first use them.
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
		// Do NOT sanitize here — the adapter handles sanitization downstream.
		// For support: AddSupport calls SanitizeBinding on each value.
		// For attacks: createAttackPremises calls TextToAtom on English templates.
		bindings[name] = value
	}

	return bindings
}

// preview shows the move before it is submitted: the instantiated premises
// and what follows from them.
func (h *HumanAgent) preview(tmpl nascq.ActionTemplate, bindings map[string]string, effect string) {
	h.print("\n  You argue:\n")
	for _, p := range tmpl.PremiseTemplatesEN {
		h.printf("      %s\n", nascq.InstantiateTemplate(p, bindings))
	}
	h.printf("    %s\n", effect)
}

// attackEffect says what an attack template concludes against its target.
func attackEffect(gs *GameState, tmpl nascq.ActionTemplate) string {
	target := gs.Adapter.Graph.Arguments[tmpl.TargetArgID]
	if target == nil || target.Conclusion == nil {
		return "(target not found)"
	}
	switch tmpl.CaesAttackType {
	case "rebuttal":
		return "it is not the case that: " + statementText(target.Conclusion)
	case "undercutter":
		return fmt.Sprintf("argument %s does not support \"%s\" here", tmpl.TargetArgID, statementText(target.Conclusion))
	case "undermining":
		if p := gs.Adapter.Graph.Statements[tmpl.TargetPremiseID]; p != nil {
			return "it is not the case that: " + statementText(p)
		}
	}
	return tmpl.CaesAttackType + " of " + tmpl.TargetArgID
}

// variableOrder returns the template's variables in the order its premise and
// conclusion templates first use them, then any others in sorted order.
func variableOrder(tmpl nascq.ActionTemplate) []string {
	seen := map[string]bool{}
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
// them (an attack argument has none).
func (gs *GameState) QuestionText(cqID, targetArgID string) string {
	cq := gs.NASCQ.GetCQ(cqID)
	if cq == nil {
		return ""
	}
	return nascq.InstantiateTemplate(cq.QuestionTemplateEN, gs.Adapter.ArgumentBindings(targetArgID))
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
