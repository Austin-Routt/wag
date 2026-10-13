package game

import (
	"fmt"
	"strings"

	"github.com/yourorg/wag/internal/carneades"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// CheckInvariants verifies properties that must hold between turns and returns
// one error per violation (none if the state is sound). The tests run it after
// every move of many random games; it is also cheap enough for a debug flag.
//
//   - Graph links: every argument's conclusion and premises are statements of
//     the graph, and every conclusion lists the argument among its Args.
//   - Issues: each statement is a position of at most one issue, and its Issue
//     pointer names that issue (Carneades assumes this when it resolves issues).
//   - Consistency: at most one position of an issue is In, and no statement is
//     In together with its negation "¬" + id.
//   - The game rule: the graph is acyclic.
//   - Labelling: the applied labels equal a fresh grounded labelling, every
//     time it is computed (Carneades walks a Go map, so a graph that breaks an
//     invariant above can label differently from run to run).
//   - Roles: main, presumption and question arguments are owned by a player;
//     exception arguments are not; presumptions and questions have no
//     premises.
//   - Scores: each player's score is the sum of the score deltas recorded for
//     them in the move history and, under ScoringLive, their live score.
//   - Questions: each has its question argument, owned by the asker, and
//     its answer, if any, is in the graph; the claim of a debate is at issue.
func (gs *GameState) CheckInvariants() []error {
	var errs []error
	fail := func(format string, args ...interface{}) { errs = append(errs, fmt.Errorf(format, args...)) }
	ag := gs.Adapter.Graph

	players := map[string]bool{}
	for _, p := range gs.Players {
		players[p.ID] = true
	}
	inGraph := make(map[*caes.Statement]bool, len(ag.Statements))
	for _, s := range ag.Statements {
		inGraph[s] = true
	}
	for _, id := range sortedArgIDs(gs) {
		arg := ag.Arguments[id]
		if arg.Conclusion == nil || !inGraph[arg.Conclusion] {
			fail("argument %s: conclusion missing from the graph", id)
			continue
		}
		linked := false
		for _, x := range arg.Conclusion.Args {
			linked = linked || x == arg
		}
		if !linked {
			fail("argument %s: not listed by its conclusion %s", id, arg.Conclusion.Id)
		}
		for _, p := range arg.Premises {
			if p.Stmt == nil || !inGraph[p.Stmt] {
				fail("argument %s: premise missing from the graph", id)
			}
		}
		owner, _ := arg.Metadata["owner"].(string)
		switch carneades.ArgumentRole(arg) {
		case carneades.RoleMain:
			if !players[owner] {
				fail("argument %s: main argument owned by %q, not a player", id, owner)
			}
		case carneades.RolePresumption, carneades.RoleQuestion:
			if !players[owner] || len(arg.Premises) != 0 {
				fail("argument %s: %s with owner %q and %d premises", id, carneades.ArgumentRole(arg), owner, len(arg.Premises))
			}
		case carneades.RoleException:
			if owner != "" {
				fail("argument %s: exception argument owned by %q", id, owner)
			}
		}
	}

	positionOf := map[*caes.Statement]string{}
	for _, iid := range carneades.SortedKeys(ag.Issues) {
		is := ag.Issues[iid]
		in := 0
		for _, p := range is.Positions {
			if prev, ok := positionOf[p]; ok {
				fail("statement %s is a position of issues %s and %s", p.Id, prev, iid)
			}
			positionOf[p] = iid
			if p.Issue != is {
				fail("statement %s is a position of issue %s but points to another issue", p.Id, iid)
			}
			if p.Label == caes.In {
				in++
			}
		}
		if in > 1 {
			fail("issue %s has %d positions In", iid, in)
		}
	}
	for _, id := range carneades.SortedKeys(ag.Statements) {
		s := ag.Statements[id]
		if s.Issue != nil {
			if _, ok := positionOf[s]; !ok {
				fail("statement %s points to issue %s but is not one of its positions", id, s.Issue.Id)
			}
		}
		if strings.HasPrefix(id, "¬") && s.Label == caes.In {
			if pos := ag.Statements[strings.TrimPrefix(id, "¬")]; pos != nil && pos.Label == caes.In {
				fail("statement %s and its negation are both In", pos.Id)
			}
		}
	}

	if !gs.Adapter.IsAcyclic() {
		fail("the argument graph has a cycle")
	}

	applied := gs.Adapter.StatementLabels()
	for run := 0; run < 3; run++ {
		fresh := gs.Adapter.GroundedLabels()
		for _, id := range carneades.SortedKeys(fresh) {
			if fresh[id] != applied[id] {
				fail("labelling run %d: %s is %v, applied label is %v", run, id, fresh[id], applied[id])
			}
		}
	}

	sum := map[string]float64{}
	for _, m := range gs.MoveHistory {
		for id, d := range m.ScoreDeltas {
			sum[id] += d
		}
		if d := m.ScoreDeltas[m.PlayerID]; d != m.ScoreDelta {
			fail("move by %s in round %d: ScoreDelta %.4f, ScoreDeltas %.4f", m.PlayerID, m.TurnNumber, m.ScoreDelta, d)
		}
	}
	var live map[string]float64
	if gs.Config.live() {
		live = gs.LiveScores()
	}
	for _, p := range gs.Players {
		if d := p.Score - sum[p.ID]; d > 1e-9 || d < -1e-9 {
			fail("player %s: score %.4f but recorded deltas sum to %.4f", p.ID, p.Score, sum[p.ID])
		}
		if live != nil {
			if d := p.Score - live[p.ID]; d > 1e-9 || d < -1e-9 {
				fail("player %s: score %.4f but live score %.4f", p.ID, p.Score, live[p.ID])
			}
		}
	}

	for i, q := range gs.Questions {
		arg := ag.Arguments[q.ArgID]
		if arg == nil || carneades.ArgumentRole(arg) != carneades.RoleQuestion || arg.Metadata["owner"] != q.Asker {
			fail("question %d (%s about %s): no question argument %s by %s", i, q.CQID, q.TargetArgID, q.ArgID, q.Asker)
		}
		if q.AnsweredBy != "" && ag.Arguments[q.AnsweredBy] == nil {
			fail("question %d: answered by %s, which is not in the graph", i, q.AnsweredBy)
		}
	}
	if gs.Claim != "" {
		if c := ag.Statements[gs.Claim]; c == nil || c.Issue == nil {
			fail("the claim %s is not at issue", gs.Claim)
		}
	}
	return errs
}

func sortedArgIDs(gs *GameState) []string {
	ids := carneades.SortedKeys(gs.Adapter.Graph.Arguments)
	carneades.SortArgumentIDs(ids)
	return ids
}
