package game

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// View is the game as an agent outside Go sees it, for JSON: the board, the
// scores, and every move the active player may make. It is what JSONAgent
// writes, and what an LLM or other program needs to play: it answers with a
// nascq.ProposedAction (which has JSON tags), filling in the variables of one
// of the actions.
type View struct {
	Turn              int             `json:"turn"`
	ActivePlayer      string          `json:"active_player"`
	Players           []PlayerView    `json:"players"`
	Statements        []StatementView `json:"statements"`
	Arguments         []ArgumentView  `json:"arguments"`
	Issues            []IssueView     `json:"issues,omitempty"`
	Actions           []ActionView    `json:"actions"`
	LastRejection     string          `json:"last_rejection,omitempty"` // why the active player's last move was rejected
	Terminated        bool            `json:"terminated"`
	TerminationReason string          `json:"termination_reason,omitempty"`
}

// PlayerView is a player and their score.
type PlayerView struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

// StatementView is a statement of an argument on the board.
type StatementView struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Label string `json:"label"` // "in", "out" or "undecided"
	Issue string `json:"issue,omitempty"`
}

// ArgumentView is an argument a player put forward.
type ArgumentView struct {
	ID         string   `json:"id"`
	Owner      string   `json:"owner,omitempty"`
	Scheme     string   `json:"scheme"` // NASCQ scheme ID if the argument has one, else the Carneades scheme
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
	Undercuts  []string `json:"undercuts,omitempty"` // for an undercutter: the arguments it undercuts
	Weight     float64  `json:"weight"`              // evaluated weight (0 if undercut or a premise is not in)
}

// IssueView is an issue: competing positions resolved by a proof standard.
type IssueView struct {
	ID        string   `json:"id"`
	Positions []string `json:"positions"`
	Standard  string   `json:"standard"`
}

// ActionView is one legal move, to be completed with variable bindings.
type ActionView struct {
	Type            string                        `json:"type"` // "support" or "attack"
	AttackType      string                        `json:"attack_type,omitempty"`
	SchemeID        string                        `json:"scheme_id"`
	SchemeName      string                        `json:"scheme_name"`
	AnswersCQID     string                        `json:"answers_cq_id,omitempty"`
	Question        string                        `json:"question,omitempty"` // the CQ, with the target's bindings
	TargetArgID     string                        `json:"target_arg_id,omitempty"`
	TargetPremiseID string                        `json:"target_premise_id,omitempty"`
	Variables       map[string]nascq.VariableSpec `json:"variables"`
	VariableOrder   []string                      `json:"variable_order"`
	Premises        []string                      `json:"premises"` // English templates with {Var} placeholders
	Concludes       string                        `json:"concludes"`
}

// BuildView renders the game for the active player.
func (gs *GameState) BuildView() *View {
	v := &View{
		Turn:              gs.CurrentTurn,
		ActivePlayer:      gs.ActivePlayer().ID,
		LastRejection:     gs.LastRejection,
		Terminated:        gs.Terminated,
		TerminationReason: gs.TerminationReason,
	}
	for _, p := range gs.Players {
		v.Players = append(v.Players, PlayerView{ID: p.ID, Score: p.Score})
	}

	ag := gs.Adapter.Graph
	undercuts := map[*caes.Statement][]string{}
	for _, id := range sortedArgIDs(gs) {
		if u := ag.Arguments[id].Undercutter; u != nil {
			undercuts[u] = append(undercuts[u], id)
		}
	}
	onBoard := map[*caes.Statement]bool{}
	for _, id := range sortedArgIDs(gs) {
		arg := ag.Arguments[id]
		role := carneades.ArgumentRole(arg)
		if arg.Conclusion == nil || role == carneades.RoleException || role == carneades.RolePresumption ||
			(role == "" && arg.Conclusion.IsUndercutter) {
			continue
		}
		av := ArgumentView{ID: id, Conclusion: arg.Conclusion.Id, Weight: arg.Weight}
		av.Owner, _ = arg.Metadata["owner"].(string)
		if arg.Scheme != nil {
			av.Scheme = arg.Scheme.Id
			if nid := gs.NASCQ.FindNASCQSchemeID(arg.Scheme.Id); nid != "" {
				av.Scheme = nid
			}
		}
		for _, p := range arg.Premises {
			av.Premises = append(av.Premises, p.Stmt.Id)
			onBoard[p.Stmt] = true
		}
		if arg.Conclusion.IsUndercutter {
			av.Undercuts = undercuts[arg.Conclusion]
		}
		onBoard[arg.Conclusion] = true
		v.Arguments = append(v.Arguments, av)
	}
	for _, id := range carneades.SortedKeys(ag.Issues) {
		is := ag.Issues[id]
		iv := IssueView{ID: id, Standard: standardName(is.Standard)}
		for _, p := range is.Positions {
			iv.Positions = append(iv.Positions, p.Id)
			onBoard[p] = true
		}
		v.Issues = append(v.Issues, iv)
	}
	for _, id := range carneades.SortedKeys(ag.Statements) {
		s := ag.Statements[id]
		if !onBoard[s] {
			continue
		}
		sv := StatementView{ID: id, Text: statementText(s), Label: strings.ToLower(s.Label.String())}
		if s.Issue != nil {
			sv.Issue = s.Issue.Id
		}
		v.Statements = append(v.Statements, sv)
	}

	for _, a := range gs.BuildObservation().AvailableActions {
		av := ActionView{Type: a.Type, AttackType: a.CaesAttackType, SchemeID: a.SchemeID, SchemeName: a.SchemeName,
			AnswersCQID: a.AnswersCQID, TargetArgID: a.TargetArgID, TargetPremiseID: a.TargetPremiseID,
			Variables: a.Variables, VariableOrder: variableOrder(a), Premises: a.PremiseTemplatesEN,
			Concludes: a.ConclusionTemplateEN}
		if a.Type == "attack" {
			av.Question = gs.QuestionText(a.AnswersCQID, a.TargetArgID)
			av.Concludes = attackEffect(gs, a)
		}
		v.Actions = append(v.Actions, av)
	}
	return v
}

func standardName(s caes.Standard) string {
	switch s {
	case caes.PE:
		return "PE"
	case caes.CCE:
		return "CCE"
	case caes.BRD:
		return "BRD"
	}
	return fmt.Sprintf("standard(%d)", s)
}

// JSONAgent plays over JSON lines, for programs outside Go (an LLM harness,
// a Python RL loop): each time it is to move it writes the View as one line
// to out, and reads one line from in: a nascq.ProposedAction such as
//
//	{"type":"support","scheme_id":"AS:sign","variable_bindings":{"A":"smoke","B":"fire"}}
//	{"type":"attack","scheme_id":"AS:sign","answers_cq_id":"CQ:possibility","target_arg_id":"a1",
//	 "variable_bindings":{"A":"no_majority","B":"cannot_pass"}}
//	{"type":"abstain"}
//
// A rejected move is followed by a new View whose last_rejection says why. A
// line that is not an action is reported the same way; end of input abstains.
type JSONAgent struct {
	name     string
	in       *bufio.Scanner
	out      io.Writer
	parseErr string
}

// NewJSONAgent creates a JSONAgent reading actions from in and writing views to out.
func NewJSONAgent(name string, in io.Reader, out io.Writer) *JSONAgent {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	return &JSONAgent{name: name, in: sc, out: out}
}

func (j *JSONAgent) Name() string { return j.name }

func (j *JSONAgent) ChooseAction(obs *Observation) nascq.ProposedAction {
	v := obs.GameState.BuildView()
	if j.parseErr != "" {
		v.LastRejection = j.parseErr
		j.parseErr = ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nascq.ProposedAction{Type: "abstain"}
	}
	fmt.Fprintf(j.out, "%s\n", b)
	if !j.in.Scan() {
		return nascq.ProposedAction{Type: "abstain"}
	}
	var a nascq.ProposedAction
	if err := json.Unmarshal(j.in.Bytes(), &a); err != nil {
		j.parseErr = "your line is not an action: " + err.Error()
		return nascq.ProposedAction{Type: "not-json"}
	}
	return a
}
