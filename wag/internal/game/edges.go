package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/nascq"
)

// EdgeCase is one way the NASCQ lets a critical question be answered: the
// question asked of an argument of a scheme, at its locus, answered for or
// against the argument with a backing scheme, as the game plays it on a
// sample argument. EdgeCases lists them all, for review: what each answer
// means in the question's words, what it concludes, and on which premises.
type EdgeCase struct {
	Scheme   string `json:"scheme"` // the scheme the question is asked of
	CQ       string `json:"cq"`
	Locus    string `json:"locus"`
	Question string `json:"question"` // about the sample argument
	Backer   string `json:"backer"`   // the scheme that backs the answer
	Polarity string `json:"polarity"` // "against" (an attack) or "for" (a defence)
	Literal  string `json:"literal,omitempty"`
	Effect   string `json:"effect"` // undermining, undercutter, rebuttal; support for a defence
	// Concludes is the statement the answering argument concludes, and
	// Reading the argument as it reads.
	Concludes string   `json:"concludes"`
	Reading   []string `json:"reading,omitempty"`
	Bridge    string   `json:"bridge,omitempty"`
	// Expected is what the locus and the polarity say the answer concludes;
	// Error is why the game rejected the answer, or why it concludes
	// something else ("" if it does as expected).
	Expected string `json:"expected"`
	Error    string `json:"error,omitempty"`
	// Offered is false for a backing scheme the NASCQ links to the question
	// that the game does not offer here (S9: its conclusion cannot be the
	// answer).
	Offered bool `json:"offered"`
}

// sampleBindings are the example values of a scheme's variables (their
// example_en), less the ones a move need not give.
func sampleBindings(s *nascq.ASNode, skip func(string) bool) map[string]string {
	b := map[string]string{}
	for v, spec := range s.Variables {
		if skip != nil && skip(v) {
			continue
		}
		ex := spec.Example
		if ex == "" {
			ex = strings.ToLower(v) + " of the example"
		}
		b[v] = ex
	}
	return b
}

// EdgeCases plays, for every ask edge of the NASCQ, an argument of its
// scheme with example values, then every answer the NASCQ allows to its
// question there (each backing scheme, for and against), each on a fresh
// game from newGame, and records what it does.
func EdgeCases(newGame func() (*GameState, error)) ([]EdgeCase, error) {
	gs, err := newGame()
	if err != nil {
		return nil, err
	}
	n := gs.NASCQ
	var out []EdgeCase
	for _, sid := range nascq.SortedStringKeys(n.Schemes) {
		ts := n.Schemes[sid]
		if ts.Free {
			continue
		}
		for _, q := range n.GetAskCQs(sid) {
			g, err := newGame()
			if err != nil {
				return nil, err
			}
			g.Config.MaxTurns = 0
			r, err := g.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: sid, VariableBindings: sampleBindings(ts, nil)})
			if err != nil || !r.Accepted {
				return nil, fmt.Errorf("the sample argument of %s: %v %v", sid, err, r.Error)
			}
			arg := r.ArgIDs[0]
			if _, err := g.ExecuteTurn(nascq.ProposedAction{Type: "abstain"}); err != nil {
				return nil, err
			}
			offered := map[string]bool{}
			for _, t := range g.NASCQ.EnumerateActions(g.Bridge(), []string{arg}) {
				if t.AnswersCQID != q.ID || t.Type == "question" || t.TargetArgID != arg {
					continue
				}
				offered[t.SchemeID+"/"+t.Type] = true
				out = append(out, g.edgeCase(sid, q, arg, t))
			}
			for _, bs := range g.NASCQ.GetBackupSchemes(q.ID) {
				for _, typ := range []string{"attack", "support"} {
					if !offered[bs.ID+"/"+typ] {
						c := EdgeCase{Scheme: sid, CQ: q.ID, Locus: g.NASCQ.LocusOf(q.ID, sid).Kind, Backer: bs.ID, Polarity: polarity(typ),
							Error: "not offered: the scheme's conclusion cannot be the answer here (S9)"}
						out = append(out, c)
					}
				}
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Scheme != b.Scheme {
			return a.Scheme < b.Scheme
		}
		if a.CQ != b.CQ {
			return a.CQ < b.CQ
		}
		if a.Backer != b.Backer {
			return a.Backer < b.Backer
		}
		return a.Polarity < b.Polarity
	})
	return out, nil
}

func polarity(typ string) string {
	if typ == "attack" {
		return "against"
	}
	return "for"
}

// edgeCase plays one answer template on a simulation of g.
func (g *GameState) edgeCase(sid string, q *nascq.CQNode, arg string, t nascq.ActionTemplate) EdgeCase {
	bs := g.NASCQ.Schemes[t.SchemeID]
	c := EdgeCase{Scheme: sid, CQ: q.ID, Locus: t.Locus.Kind, Backer: t.SchemeID, Polarity: polarity(t.Type), Offered: true,
		Question: g.QuestionAbout(q.ID, arg, t.LocusStmtID), Effect: t.CaesAttackType}
	if c.Effect == "" {
		c.Effect = "support"
	}
	if q.AnswerEN != nil {
		c.Literal = q.AnswerEN.For
		if t.Type == "attack" {
			c.Literal = q.AnswerEN.Against
		}
	}
	// What the locus and polarity say the answer concludes.
	target := g.Adapter.Graph.Arguments[arg]
	switch t.Locus.Kind {
	case nascq.LocusPremise, nascq.LocusException:
		c.Expected = t.LocusStmtID
		if (t.Locus.Kind == nascq.LocusPremise) == (t.Type == "attack") {
			c.Expected = g.Bridge().Statement(t.LocusStmtID).Opposite
		}
	case nascq.LocusConclusion:
		c.Expected = target.Conclusion.Id
		if t.Type == "attack" {
			c.Expected = g.Bridge().Statement(target.Conclusion.Id).Opposite
		}
	}
	// The answer's values differ from the sample argument's (an expert
	// questioned is not the expert who answers).
	vals := sampleBindings(bs, t.Bound)
	for k, v := range vals {
		vals[k] = v + " (2)"
	}
	a := nascq.ProposedAction{Type: t.Type, TargetArgID: arg, AnswersCQID: q.ID, SchemeID: t.SchemeID, TargetPremiseID: t.TargetPremiseID,
		AttackType: t.AttackType, VariableBindings: vals}
	r, err := g.SimulateThen(a, func(after *GameState) {
		n := len(after.MoveHistory)
		if n == 0 || len(after.MoveHistory[n-1].ArgIDs) == 0 {
			return
		}
		m := after.MoveHistory[n-1]
		x := after.Adapter.Graph.Arguments[m.ArgIDs[0]]
		if x == nil || x.Conclusion == nil {
			return
		}
		c.Concludes, c.Bridge = x.Conclusion.Id, m.Bridge
		for _, p := range x.Premises {
			if p.Stmt != nil {
				c.Reading = append(c.Reading, after.SayIn(x.Id, p.Stmt.Id))
			}
		}
		c.Reading = append(c.Reading, "therefore "+after.Say(x.Conclusion.Id))
		// An answer about the inference concludes the target's undercutter
		// (made on demand), or its opposite.
		if t.Locus.Kind == nascq.LocusArgument {
			if u := after.Adapter.Graph.Arguments[arg].Undercutter; u != nil {
				c.Expected = u.Id
				if t.Type == "support" {
					c.Expected = after.Bridge().Statement(u.Id).Opposite
				}
			}
		}
	})
	switch {
	case err != nil:
		c.Error = err.Error()
	case !r.Accepted:
		c.Error = r.Error
	case c.Concludes != c.Expected:
		c.Error = "concludes " + c.Concludes + ", not " + c.Expected
	}
	return c
}

// EdgeTable writes the edge cases as Markdown, grouped by scheme and
// question.
func EdgeTable(cases []EdgeCase) string {
	var b strings.Builder
	last := ""
	for _, c := range cases {
		if k := c.Scheme + " " + c.CQ; k != last {
			last = k
			fmt.Fprintf(&b, "\n### %s — %s (%s)\n\n", c.Scheme, c.CQ, c.Locus)
			if c.Question != "" {
				fmt.Fprintf(&b, "%s\n\n", c.Question)
			}
			b.WriteString("| Answer | Backed by | Effect | Concludes | The answering argument | Check |\n|---|---|---|---|---|---|\n")
		}
		ans := c.Polarity
		if c.Literal != "" {
			ans += " (“" + c.Literal + "”)"
		}
		check := "✓"
		if c.Error != "" {
			check = c.Error
		}
		if c.Bridge != "" {
			check += "; bridge: " + c.Bridge
		}
		fmt.Fprintf(&b, "| %s | %s | %s | `%s` | %s | %s |\n", ans, c.Backer, c.Effect, c.Concludes,
			strings.ReplaceAll(strings.Join(c.Reading, "; "), "|", "/"), strings.ReplaceAll(check, "|", "/"))
	}
	return b.String()
}
