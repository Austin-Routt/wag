package server

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// State is a game as the browser shows it: the board as statements and the
// arguments for and against them (the shape of a Kialo-style reader), every
// argument's critical questions with what happened to each, the players, the
// log, and the moves the player to move can make.
type State struct {
	ID        string        `json:"id"`
	Version   int           `json:"version"`
	Title     string        `json:"title"`
	Source    string        `json:"source,omitempty"`
	Config    ConfigView    `json:"config"`
	Round     int           `json:"round"`
	Active    string        `json:"active"`    // the player to move
	ActiveAI  string        `json:"active_ai"` // "" for a human seat, else the agent kind
	Players   []PlayerState `json:"players"`
	Claim     string        `json:"claim,omitempty"`
	Verdict   string        `json:"verdict,omitempty"`
	Over      bool          `json:"over"`
	Outcome   *game.Outcome `json:"outcome,omitempty"`
	Pending   *PendingState `json:"pending,omitempty"`
	Rejection string        `json:"last_rejection,omitempty"`

	Statements map[string]*StmtState `json:"statements"`
	Arguments  map[string]*ArgState  `json:"arguments"`
	Roots      []string              `json:"roots"` // statements to start reading from: the claim first
	Questions  []QuestionState       `json:"questions"`
	Log        []LogEntry            `json:"log"`
	// CanMove: a human player is to move (the moves about an argument come
	// from /api/games/{id}/moves?target=...).
	CanMove bool `json:"can_move"`
	// Thinking: a Claude seat is waiting for the model.
	Thinking bool `json:"thinking,omitempty"`
	// Notes are the testers' remarks on the game.
	Notes []Note `json:"notes"`
	// SeatKind is the kind of the seat to move (human, llm, random, greedy, claude).
	SeatKind string `json:"seat_kind,omitempty"`
}

// ConfigView is the game's configuration, for display.
type ConfigView struct {
	Mode            string  `json:"mode"`
	Objective       string  `json:"objective"`
	Scoring         string  `json:"scoring"`
	MaxRounds       int     `json:"max_rounds"`
	TargetScore     float64 `json:"target_score,omitempty"`
	TimeLimit       string  `json:"time_limit,omitempty"`
	MoveTimeLimit   string  `json:"move_time_limit,omitempty"`
	AnswerWithin    int     `json:"answer_within,omitempty"`
	Standard        string  `json:"claim_standard,omitempty"`
	ConcedeOnPass   bool    `json:"concede_on_pass,omitempty"`
	AllowOther      bool    `json:"allow_other,omitempty"`
	AnyPremise      bool    `json:"any_premise,omitempty"`
	Context         string  `json:"context,omitempty"`
	Moderated       bool    `json:"moderated,omitempty"`
	RequireEvidence bool    `json:"require_evidence,omitempty"`
}

// PlayerState is a seat.
type PlayerState struct {
	ID       string  `json:"id"`
	Kind     string  `json:"kind"` // human, random, greedy
	Score    float64 `json:"score"`
	Side     string  `json:"side,omitempty"`
	Resigned bool    `json:"resigned,omitempty"`
}

// PendingState is the move awaiting the moderator.
type PendingState struct {
	Player      string               `json:"player"`
	Action      nascq.ProposedAction `json:"action"`
	Description string               `json:"description"` // what it would do
	Weight      float64              `json:"weight"`      // the scheme's base weight
}

// StmtState is a statement with the arguments for and against it.
type StmtState struct {
	ID          string   `json:"id"`
	Text        string   `json:"text"`
	Label       string   `json:"label"` // in, out, undecided
	Issue       string   `json:"issue,omitempty"`
	Standard    string   `json:"standard,omitempty"`
	Opposite    string   `json:"opposite,omitempty"` // the other position of its issue
	Decision    string   `json:"decision,omitempty"` // the moderator's
	Pro         []string `json:"pro"`                // arguments for it that a player made
	Con         []string `json:"con"`                // arguments for its opposite
	Questions   []string `json:"questions"`          // question arguments against it
	PremiseOf   []string `json:"premise_of"`         // arguments it is a premise of
	ExceptionOf []string `json:"exception_of"`       // arguments it is an exception of
	Presumed    []string `json:"presumed_by"`        // players whose presumption holds it
	Undercuts   []string `json:"undercuts,omitempty"`
	Explain     string   `json:"explain"`
}

// ArgState is an argument a player made (or a question).
type ArgState struct {
	ID         string   `json:"id"`
	Owner      string   `json:"owner"`
	Role       string   `json:"role"` // main or question
	Scheme     string   `json:"scheme"`
	SchemeName string   `json:"scheme_name"`
	Free       bool     `json:"free,omitempty"`
	Weight     float64  `json:"weight"`      // evaluated: 0 if out of force
	BaseWeight float64  `json:"base_weight"` // the weight it was made with
	InForce    bool     `json:"in_force"`
	WhyOut     string   `json:"why_out,omitempty"`
	Conclusion string   `json:"conclusion"`
	Premises   []string `json:"premises"`
	Exceptions []string `json:"exceptions,omitempty"`
	// PremiseTexts and ExceptionTexts read the premises and exceptions as
	// the card shows them, with the conclusion as "C" (GameState.SayIn).
	PremiseTexts   []string     `json:"premise_texts,omitempty"`
	ExceptionTexts []string     `json:"exception_texts,omitempty"`
	Undercutter    string       `json:"undercutter,omitempty"`
	UndercutBy     []string     `json:"undercut_by,omitempty"` // arguments for its undercutter
	Text           string       `json:"text,omitempty"`
	Evidence       []string     `json:"evidence,omitempty"`
	Answers        *AnswerOf    `json:"answers,omitempty"` // the critical question this argument answers
	CQs            []CQOnArg    `json:"cqs"`
	Move           int          `json:"move"` // index of the move that made it in the log, -1 if imported
	Moderation     string       `json:"moderation,omitempty"`
	Kialo          *KialoOrigin `json:"kialo,omitempty"`
}

// AnswerOf says which question an answer argument answers.
type AnswerOf struct {
	Target   string `json:"target"`
	CQ       string `json:"cq"`
	CQName   string `json:"cq_name"`
	Positive bool   `json:"positive"` // a defence
	Locus    string `json:"locus"`
}

// KialoOrigin is where an imported argument came from.
type KialoOrigin struct {
	Location string `json:"location"`
	Claim    string `json:"claim"`
}

// CQOnArg is a critical question asked of an argument's scheme, and what
// players did with it.
type CQOnArg struct {
	CQ       string   `json:"cq"`
	Name     string   `json:"name"`
	Question string   `json:"question"`
	Locus    string   `json:"locus"`
	Stmts    []string `json:"stmts,omitempty"` // the premise or exception (several for "any premise")
	AskedBy  string   `json:"asked_by,omitempty"`
	Open     bool     `json:"open,omitempty"`
	Attacks  []string `json:"attacks,omitempty"`
	Defences []string `json:"defences,omitempty"`
}

// QuestionState is an open question.
type QuestionState struct {
	Asker     string `json:"asker"`
	Owner     string `json:"owner"`
	Target    string `json:"target"`
	CQ        string `json:"cq"`
	Question  string `json:"question"`
	Stmt      string `json:"stmt,omitempty"`
	TurnsLeft int    `json:"turns_left,omitempty"`
}

// LogEntry is a line of the game's history.
type LogEntry struct {
	Index    int                `json:"index"`
	Round    int                `json:"round"`
	Player   string             `json:"player"`
	Text     string             `json:"text"`
	Accepted bool               `json:"accepted"`
	Deltas   map[string]float64 `json:"deltas,omitempty"`
	Args     []string           `json:"args,omitempty"`
	NL       string             `json:"nl,omitempty"` // the player's own words

	// The move in parts, for the dialogue view.
	Kind     string   `json:"kind"`               // argue, attack, defend, question, pass, resign, decide, rejected, timeout
	Target   string   `json:"target,omitempty"`   // the argument the move answers or questions
	CQ       string   `json:"cq,omitempty"`       // the critical question
	CQName   string   `json:"cq_name,omitempty"`  // its name
	Question string   `json:"question,omitempty"` // its text, about the target
	About    string   `json:"about,omitempty"`    // what of the target it is about, in words
	Answer   string   `json:"answer,omitempty"`   // what the move argues (its main argument's conclusion)
	Scheme   string   `json:"scheme,omitempty"`   // the scheme of the main argument
	Reason   string   `json:"reason,omitempty"`   // why a rejected move was rejected
	Effects  []string `json:"effects,omitempty"`  // what the move changed: the claim, arguments in or out of force
}

// MovesState lists, per argument, the critical questions the player to move
// can raise or answer, and with which schemes; plain arguments can be made
// with any scheme (see the metadata).
type MovesState struct {
	Targets map[string][]MoveOption `json:"targets"`
}

// MoveOption is one critical question about one argument (and one premise,
// for questions about any premise).
type MoveOption struct {
	CQ          string   `json:"cq"`
	Locus       string   `json:"locus"`
	Stmt        string   `json:"stmt,omitempty"`
	Question    bool     `json:"question"` // can be asked alone
	Schemes     []string `json:"schemes"`  // can back an answer either way
	AttackType  string   `json:"attack_type"`
	AttackText  string   `json:"attack_text"`
	DefenceText string   `json:"defence_text"`
}

// buildState renders a session.
func (s *Session) buildState() *State {
	gs := s.GS
	ag := gs.Adapter.Graph
	st := &State{
		ID: s.ID, Version: s.Version, Title: s.Title, Source: s.Source,
		Round: gs.CurrentTurn, Active: gs.ActivePlayer().ID, ActiveAI: s.Seats[gs.ActivePlayerIdx].aiKind(),
		Claim: gs.Claim, Over: gs.Terminated, Rejection: gs.LastRejection,
		Statements: map[string]*StmtState{}, Arguments: map[string]*ArgState{},
		Roots: []string{}, Questions: []QuestionState{}, Log: []LogEntry{},
	}
	c := gs.Config
	st.Config = ConfigView{Mode: c.Mode, Objective: c.Objective, Scoring: c.Scoring.Model, MaxRounds: c.MaxTurns,
		TargetScore: c.TargetScore, AnswerWithin: c.AnswerWithin, ConcedeOnPass: c.ConcedeOnPass,
		AllowOther: c.AllowOther, AnyPremise: c.AnyPremise, Context: c.Context, Moderated: c.Moderated, RequireEvidence: c.RequireEvidence}
	if st.Config.Objective == "" {
		st.Config.Objective = game.ObjectivePoints
	}
	if st.Config.Scoring == "" {
		st.Config.Scoring = game.ScoringMove
	}
	if c.Objective == game.ObjectiveClaim {
		st.Config.Standard = game.StandardName(c.ClaimStandard)
	}
	if c.TimeLimit > 0 {
		st.Config.TimeLimit = c.TimeLimit.String()
	}
	if c.MoveTimeLimit > 0 {
		st.Config.MoveTimeLimit = c.MoveTimeLimit.String()
	}
	o := gs.Outcome()
	st.Verdict = o.Verdict
	if gs.Terminated {
		st.Outcome = o
	}
	for i, p := range gs.Players {
		st.Players = append(st.Players, PlayerState{ID: p.ID, Kind: s.Seats[i].Kind, Score: p.Score, Side: gs.Side(p.ID), Resigned: p.Resigned})
	}
	if p := gs.Pending; p != nil {
		st.Pending = &PendingState{Player: p.PlayerID, Action: p.Action, Description: s.describeAction(p.Action)}
		if sc := gs.NASCQ.GetScheme(p.Action.SchemeID); sc != nil {
			st.Pending.Weight = sc.BaseWeight
		}
	}

	// Which move made each argument, and what each move answered.
	madeBy := map[string]int{}
	for i, m := range gs.MoveHistory {
		for _, id := range m.ArgIDs {
			madeBy[id] = i
		}
	}

	stmt := func(x *caes.Statement) *StmtState {
		if ss, ok := st.Statements[x.Id]; ok {
			return ss
		}
		ss := &StmtState{ID: x.Id, Text: gs.Say(x.Id), Label: labelName(x.Label), Pro: []string{}, Con: []string{},
			Questions: []string{}, PremiseOf: []string{}, ExceptionOf: []string{}, Presumed: []string{}}
		if x.Issue != nil {
			ss.Issue = x.Issue.Id
			ss.Standard = game.StandardName(x.Issue.Standard)
			for _, p := range x.Issue.Positions {
				if p != x {
					ss.Opposite = p.Id
				}
			}
		}
		if d := gs.Adapter.Decision(x.Id); d != carneades.DecisionUndecided {
			ss.Decision = d
		}
		st.Statements[x.Id] = ss
		return ss
	}

	ids := carneades.SortedKeys(ag.Arguments)
	carneades.SortArgumentIDs(ids)
	for _, id := range ids {
		arg := ag.Arguments[id]
		if arg.Conclusion == nil {
			continue
		}
		role := carneades.ArgumentRole(arg)
		concl := stmt(arg.Conclusion)
		switch role {
		case carneades.RolePresumption:
			concl.Presumed = append(concl.Presumed, ownerOf(arg))
			continue
		case carneades.RoleException:
			if len(arg.Premises) == 1 {
				stmt(arg.Premises[0].Stmt)
			}
			continue
		case carneades.RoleQuestion:
			st.Arguments[id] = &ArgState{ID: id, Owner: ownerOf(arg), Role: role, Scheme: "question", SchemeName: "Question",
				Weight: arg.Weight, InForce: arg.Weight > 0, Conclusion: arg.Conclusion.Id, Premises: []string{}, CQs: []CQOnArg{}, Move: idx(madeBy, id)}
			if pos := opposite(arg.Conclusion); pos != nil && !arg.Conclusion.IsUndercutter {
				stmt(pos).Questions = append(stmt(pos).Questions, id)
			}
			continue
		}
		if role == "" && arg.Conclusion.IsUndercutter {
			continue
		}
		as := &ArgState{ID: id, Owner: ownerOf(arg), Role: carneades.RoleMain, SchemeName: gs.SchemeName(arg),
			Weight: arg.Weight, InForce: arg.Weight > 0, Conclusion: arg.Conclusion.Id, Premises: []string{}, CQs: []CQOnArg{},
			Move: idx(madeBy, id)}
		if arg.Scheme != nil {
			as.Scheme = gs.NASCQ.FindNASCQSchemeID(arg.Scheme.Id)
			if as.Scheme == "" {
				as.Scheme = arg.Scheme.Id
			}
			as.Free = arg.Scheme.Id == carneades.FreeSchemeID
		}
		as.Text, _ = arg.Metadata["text"].(string)
		as.Evidence, _ = arg.Metadata["evidence"].([]string)
		if loc, ok := arg.Metadata["kialo_location"].(string); ok {
			cl, _ := arg.Metadata["kialo_claim"].(string)
			as.Kialo = &KialoOrigin{Location: loc, Claim: cl}
		}
		for _, p := range arg.Premises {
			if p.Stmt != nil {
				as.Premises = append(as.Premises, p.Stmt.Id)
				as.PremiseTexts = append(as.PremiseTexts, gs.SayIn(id, p.Stmt.Id))
				ps := stmt(p.Stmt)
				ps.PremiseOf = append(ps.PremiseOf, id)
			}
		}
		if u := arg.Undercutter; u != nil {
			as.Undercutter = u.Id
			for _, x := range u.Args {
				switch carneades.ArgumentRole(x) {
				case carneades.RoleException:
					if len(x.Premises) == 1 {
						e := stmt(x.Premises[0].Stmt)
						as.Exceptions = append(as.Exceptions, e.ID)
						as.ExceptionTexts = append(as.ExceptionTexts, gs.SayIn(id, e.ID))
						e.ExceptionOf = append(e.ExceptionOf, id)
					}
				default:
					as.UndercutBy = append(as.UndercutBy, x.Id)
				}
			}
		}
		if i, ok := madeBy[id]; ok {
			m := gs.MoveHistory[i]
			as.BaseWeight = m.Weight
			as.Moderation = m.Moderation
			if a := m.Action; a.TargetArgID != "" && a.Type != "question" && id == m.ArgIDs[0] {
				as.Answers = &AnswerOf{Target: a.TargetArgID, CQ: a.AnswersCQID, Positive: a.Type == "support"}
				if q := gs.NASCQ.GetCQ(a.AnswersCQID); q != nil {
					as.Answers.CQName = q.Name
				}
				if sid := gs.NASCQ.FindNASCQSchemeID(gs.Adapter.Graph.Arguments[a.TargetArgID].Scheme.Id); sid != "" && a.AnswersCQID != "" {
					as.Answers.Locus = gs.NASCQ.LocusOf(a.AnswersCQID, sid).Kind
				}
			}
		} else if arg.Scheme != nil && arg.Scheme.Weight != nil {
			l := caes.Labelling{} // its weight with every premise In
			for _, p := range arg.Premises {
				l[p.Stmt] = caes.In
			}
			as.BaseWeight = arg.Scheme.Weight(arg, l)
		}
		if !as.InForce {
			as.WhyOut = whyOut(gs, arg)
		}
		st.Arguments[id] = as

		// The conclusion's pro, or its opposite's con.
		concl.Pro = append(concl.Pro, id)
		if pos := opposite(arg.Conclusion); pos != nil {
			ps := stmt(pos)
			ps.Con = append(ps.Con, id)
		}
	}
	// Undercutters: which arguments each undercuts.
	for _, id := range ids {
		arg := ag.Arguments[id]
		if arg.Undercutter != nil {
			if us, ok := st.Statements[arg.Undercutter.Id]; ok {
				us.Undercuts = append(us.Undercuts, id)
			}
		}
	}

	// Critical questions on each argument, from the NASCQ and the history.
	for _, id := range ids {
		as := st.Arguments[id]
		if as == nil || as.Role != carneades.RoleMain {
			continue
		}
		for _, q := range gs.NASCQ.GetAskCQs(as.Scheme) {
			l := gs.NASCQ.LocusOf(q.ID, as.Scheme)
			cq := CQOnArg{CQ: q.ID, Name: q.Name, Question: gs.QuestionText(q.ID, id), Locus: l.Kind}
			switch l.Kind {
			case nascq.LocusPremise:
				if l.Pattern == "" { // any premise no other question is about
					covered := gs.NASCQ.CoveredPremises(gs.Bridge(), id, as.Scheme)
					for _, p := range as.Premises {
						if !covered[p] {
							cq.Stmts = append(cq.Stmts, p)
						}
					}
				} else if p := gs.Adapter.PremiseMatching(id, l.Pattern); p != "" {
					cq.Stmts = []string{p}
				}
			case nascq.LocusException:
				if e := gs.Adapter.ExceptionMatching(id, l.Pattern); e != "" {
					cq.Stmts = []string{e}
				}
			}
			for _, m := range gs.MoveHistory {
				if m.Action.TargetArgID != id || m.Action.AnswersCQID != q.ID || len(m.ArgIDs) == 0 {
					continue
				}
				switch m.Action.Type {
				case "question":
					cq.AskedBy = m.PlayerID
				case "attack":
					cq.Attacks = append(cq.Attacks, m.ArgIDs[0])
				case "support":
					cq.Defences = append(cq.Defences, m.ArgIDs[0])
				}
			}
			for _, oq := range gs.Questions {
				if oq.TargetArgID == id && oq.CQID == q.ID && oq.Open() {
					cq.Open = true
				}
			}
			as.CQs = append(as.CQs, cq)
		}
	}

	// Explanations of the labels.
	for _, ss := range st.Statements {
		ss.Explain = explain(gs, st, ss)
	}

	// Roots: the claim, then conclusions that are nobody's premise.
	if gs.Claim != "" {
		if _, ok := st.Statements[gs.Claim]; !ok {
			stmt(ag.Statements[gs.Claim])
		}
		st.Roots = append(st.Roots, gs.Claim)
	}
	var roots []string
	for _, id := range ids {
		as := st.Arguments[id]
		if as == nil || as.Role != carneades.RoleMain || as.Answers != nil {
			continue
		}
		c := st.Statements[as.Conclusion]
		root := c.ID
		if strings.HasPrefix(root, "¬") && c.Opposite != "" {
			root = c.Opposite
		}
		if r := st.Statements[root]; r != nil && len(r.PremiseOf) == 0 && len(r.ExceptionOf) == 0 && root != gs.Claim {
			roots = appendNew(roots, root)
		}
	}
	st.Roots = append(st.Roots, roots...)

	for _, q := range gs.Questions {
		if q.Open() {
			qs := QuestionState{Asker: q.Asker, Owner: q.Owner, Target: q.TargetArgID, CQ: q.CQID,
				Question: gs.QuestionText(q.CQID, q.TargetArgID), Stmt: q.LocusStmtID}
			if gs.Config.AnswerWithin > 0 {
				qs.TurnsLeft = q.TurnsLeft
			}
			st.Questions = append(st.Questions, qs)
		}
	}
	for i, m := range gs.MoveHistory {
		st.Log = append(st.Log, logEntry(gs, i, m, madeBy))
	}
	st.CanMove = !gs.Terminated && gs.Pending == nil && st.ActiveAI == "" && !s.thinking
	st.Thinking = s.thinking
	st.Notes = append([]Note{}, s.Notes...)
	if i := gs.ActivePlayerIdx; i >= 0 && i < len(s.Seats) {
		st.SeatKind = s.Seats[i].Kind
	}
	return st
}

// moves lists the critical questions the player to move can raise or answer
// about the argument target.
func moves(gs *game.GameState, target string) []MoveOption {
	ms := &MovesState{Targets: map[string][]MoveOption{}}
	index := map[string]int{}
	for _, t := range gs.NASCQ.EnumerateActions(gs.Bridge(), []string{target}) {
		if t.TargetArgID == "" {
			continue
		}
		key := t.TargetArgID + "|" + t.AnswersCQID + "|" + t.LocusStmtID
		i, ok := index[key]
		if !ok {
			opt := MoveOption{CQ: t.AnswersCQID, Locus: t.Locus.Kind, Stmt: t.LocusStmtID, AttackType: t.Locus.AttackType(), Schemes: []string{}}
			att, def := t, t
			att.Type, def.Type = "attack", "support"
			opt.AttackText = gs.AnswerText(att)
			opt.DefenceText = gs.AnswerText(def)
			ms.Targets[t.TargetArgID] = append(ms.Targets[t.TargetArgID], opt)
			i = len(ms.Targets[t.TargetArgID]) - 1
			index[key] = i
		}
		opt := &ms.Targets[t.TargetArgID][i]
		switch t.Type {
		case "question":
			opt.Question = true
		default:
			opt.Schemes = appendNew(opt.Schemes, t.SchemeID)
		}
	}
	if ms.Targets[target] == nil {
		return []MoveOption{}
	}
	return ms.Targets[target]
}

// explain says in a sentence why a statement has its label.
func explain(gs *game.GameState, st *State, ss *StmtState) string {
	switch ss.Decision {
	case carneades.DecisionAccepted:
		return "Accepted by the moderator."
	case carneades.DecisionRejected:
		return "Rejected by the moderator."
	}
	best := func(args []string) (string, float64) {
		id, w := "", 0.0
		for _, a := range args {
			if as := st.Arguments[a]; as != nil && as.Weight > w {
				id, w = a, as.Weight
			}
		}
		return id, w
	}
	if ss.Issue != "" {
		pro, wp := best(ss.Pro)
		opp := st.Statements[ss.Opposite]
		con, wc := "", 0.0
		if opp != nil {
			con, wc = best(opp.Pro)
		}
		if wp == 0 && len(ss.Presumed) > 0 {
			wp = gs.Config.EffectivePresumptionWeight()
			pro = "presumption"
		}
		if wc == 0 && opp != nil && len(opp.Presumed) > 0 {
			wc = gs.Config.EffectivePresumptionWeight()
			con = "presumption"
		}
		if opp != nil {
			for _, q := range opp.Questions {
				if wc < gs.Config.EffectivePresumptionWeight() {
					wc, con = gs.Config.EffectivePresumptionWeight(), "question "+q
				}
			}
		}
		side := func(id string, w float64) string {
			if id == "" {
				return "nothing in force"
			}
			return fmt.Sprintf("%s (%.2f)", id, w)
		}
		verdict := map[string]string{"in": "it wins", "out": "it does not win", "undecided": "undecided"}[ss.Label]
		if ss.Label == "out" && wp > 0 && wp == wc {
			verdict = "a tie: neither side wins"
		}
		return fmt.Sprintf("At issue under %s: for it %s, against it %s; %s.", ss.Standard, side(pro, wp), side(con, wc), verdict)
	}
	switch ss.Label {
	case "in":
		if id, w := best(ss.Pro); id != "" {
			return fmt.Sprintf("Argued by %s (weight %.2f), which is in force.", id, w)
		}
		if len(ss.Presumed) > 0 {
			return "Presumed: it holds until someone attacks or questions it."
		}
		return "Assumed."
	case "out":
		if len(ss.Pro) > 0 {
			return "No argument for it is in force."
		}
		if len(ss.ExceptionOf) > 0 {
			return "An exception: it holds only if someone argues for it."
		}
		return "Nobody has argued for it."
	}
	return "Undecided."
}

// whyOut says why an argument is out of force.
func whyOut(gs *game.GameState, arg *caes.Argument) string {
	if u := arg.Undercutter; u != nil && u.Label == caes.In {
		return "undercut: " + gs.Say(u.Id)
	}
	var out []string
	for _, p := range arg.Premises {
		if p.Stmt != nil && p.Stmt.Label != caes.In {
			out = append(out, gs.Say(p.Stmt.Id))
		}
	}
	if len(out) > 0 {
		return "premise not In: " + strings.Join(out, "; ")
	}
	return ""
}

func opposite(s *caes.Statement) *caes.Statement {
	if s.Issue == nil || len(s.Issue.Positions) != 2 {
		return nil
	}
	if s.Issue.Positions[0] == s {
		return s.Issue.Positions[1]
	}
	return s.Issue.Positions[0]
}

func ownerOf(arg *caes.Argument) string {
	o, _ := arg.Metadata["owner"].(string)
	return o
}

func labelName(l caes.Label) string {
	switch l {
	case caes.In:
		return "in"
	case caes.Out:
		return "out"
	}
	return "undecided"
}

func idx(m map[string]int, id string) int {
	if i, ok := m[id]; ok {
		return i
	}
	return -1
}

func appendNew(list []string, x string) []string {
	for _, y := range list {
		if y == x {
			return list
		}
	}
	return append(list, x)
}

// sortedStrings returns the strings sorted.
func sortedStrings(xs []string) []string {
	out := append([]string{}, xs...)
	sort.Strings(out)
	return out
}

// logEntry is a move of the history, in parts and in words.
func logEntry(gs *game.GameState, i int, m game.Move, madeBy map[string]int) LogEntry {
	a := m.Action
	e := LogEntry{Index: i, Round: m.TurnNumber, Player: m.PlayerID, Text: gs.DescribeMove(m),
		Accepted: len(m.ArgIDs) > 0 || m.Abstained || m.Resigned || a.Type == "decide",
		Deltas:   m.ScoreDeltas, Args: m.ArgIDs, NL: a.Text, Target: a.TargetArgID, CQ: a.AnswersCQID}
	if q := gs.NASCQ.GetCQ(a.AnswersCQID); q != nil {
		e.CQName = q.Name
		if a.TargetArgID != "" {
			e.Question = gs.QuestionAbout(a.AnswersCQID, a.TargetArgID, a.TargetPremiseID)
		}
	}
	switch {
	case a.Type == "decide":
		e.Kind = "decide"
	case m.Resigned:
		e.Kind = "resign"
	case m.TimedOut:
		e.Kind = "timeout"
	case m.Abstained:
		e.Kind = "pass"
	case !e.Accepted:
		e.Kind, e.Reason = "rejected", m.Rejection
		if e.Reason == "" && m.Moderation != "" {
			e.Reason = m.Moderation
		}
		e.Text = fmt.Sprintf("%s's move was rejected", m.PlayerID)
		if e.Reason != "" {
			e.Text += ": " + e.Reason
		}
	case a.Type == "question":
		e.Kind = "question"
	case a.Type == "attack":
		e.Kind = "attack"
	case a.TargetArgID != "":
		e.Kind = "defend"
	default:
		e.Kind = "argue"
	}
	if len(m.ArgIDs) > 0 && e.Kind != "question" {
		main := gs.Adapter.Graph.Arguments[m.ArgIDs[0]]
		if main != nil && main.Conclusion != nil {
			e.Answer = gs.Say(main.Conclusion.Id)
			e.Scheme = gs.SchemeName(main)
		}
	}
	if a.TargetArgID != "" && a.AnswersCQID != "" {
		if t := gs.Adapter.Graph.Arguments[a.TargetArgID]; t != nil && t.Scheme != nil {
			if sid := gs.NASCQ.FindNASCQSchemeID(t.Scheme.Id); sid != "" {
				e.About = aboutWords(gs.NASCQ.LocusOf(a.AnswersCQID, sid).Kind)
			}
		}
	}
	if e.Accepted {
		e.Effects = effects(gs, m, madeBy)
	}
	return e
}

func aboutWords(locus string) string {
	return map[string]string{nascq.LocusPremise: "a premise", nascq.LocusException: "an exception",
		nascq.LocusConclusion: "the conclusion", nascq.LocusArgument: "the inference"}[locus]
}

// effects says what a move changed: the claim's label, and which arguments
// that were already on the board went out of force or came back into it.
// An argument is in force when its premises are In and its undercutter is
// not (exceptions undercut).
func effects(gs *game.GameState, m game.Move, madeBy map[string]int) []string {
	var out []string
	before, after := m.LabelsBefore, m.LabelsAfter
	if before == nil || after == nil {
		return nil
	}
	if c := gs.Claim; c != "" {
		if b, a := before[c], after[c]; b != a {
			out = append(out, fmt.Sprintf("The claim is now %s (it was %s).", strings.ToUpper(labelName(a)), labelName(b)))
		}
	}
	mine := map[string]bool{}
	for _, id := range m.ArgIDs {
		mine[id] = true
	}
	inForce := func(arg *caes.Argument, labels map[string]caes.Label) bool {
		for _, p := range arg.Premises {
			if p.Stmt != nil && labels[p.Stmt.Id] != caes.In {
				return false
			}
		}
		return arg.Undercutter == nil || labels[arg.Undercutter.Id] != caes.In
	}
	var ids []string
	for id := range madeBy {
		ids = append(ids, id)
	}
	carneades.SortArgumentIDs(ids)
	for _, id := range ids {
		arg := gs.Adapter.Graph.Arguments[id]
		if arg == nil || mine[id] || carneades.ArgumentRole(arg) != carneades.RoleMain {
			continue
		}
		if _, existed := before[arg.Conclusion.Id]; !existed {
			continue
		}
		b, a := inForce(arg, before), inForce(arg, after)
		switch {
		case b && !a:
			out = append(out, fmt.Sprintf("%s (%s's) no longer holds.", id, ownerOf(arg)))
		case !b && a:
			out = append(out, fmt.Sprintf("%s (%s's) holds again.", id, ownerOf(arg)))
		}
	}
	return out
}
