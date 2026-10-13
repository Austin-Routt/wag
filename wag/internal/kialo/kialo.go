// Package kialo turns a Kialo discussion, as captured by the Kialo research
// visualizer (data/<id>/graph.json), into a WAG game.
//
// Kialo is a tree of claims under a thesis: each claim is a pro (support) or
// a con (attack) of its parent. That is a bipolar, Dung-style graph of free
// text, with no argumentation schemes. Each relation becomes an argument of
// WAG's user-defined scheme "Other" with the child claim as its premise: for
// the parent claim (a pro) or for its opposing position (a con, a rebuttal).
// Classifying them into Walton's schemes is the argument-mining step that
// comes after.
package kialo

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Discussion is the part of a captured Kialo discussion the import needs.
type Discussion struct {
	ID            string           `json:"id"`
	Title         string           `json:"displayTitle"`
	URL           string           `json:"discussionUrl"`
	ThesisClaimID string           `json:"thesis_claim_id"`
	Claims        map[string]Claim `json:"claims"`
	Edges         []Edge           `json:"edges"` // the current relations
}

// Claim is a claim's identity and text.
type Claim struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Author string `json:"author_identity_id"`
}

// Edge says that the claim To supports or attacks the claim From.
type Edge struct {
	LocationID string `json:"location_id"`
	From       string `json:"from_claim_id"` // the parent
	To         string `json:"to_claim_id"`   // the child: the pro or con
	Relation   string `json:"relation"`      // "support" or "attack"
	Votes      []int  `json:"vote_vector"`   // impact votes 0..4: how many of each
	Active     bool   `json:"active"`
}

// Parse reads a graph.json of the Kialo research visualizer.
func Parse(b []byte) (*Discussion, error) {
	var d Discussion
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	if d.ThesisClaimID == "" || d.Claims[d.ThesisClaimID].Text == "" {
		return nil, fmt.Errorf("no thesis claim: is this a graph.json of the Kialo research visualizer?")
	}
	return &d, nil
}

// Relation is one pro or con to import, in the order of a walk from the
// thesis (parents before children).
type Relation struct {
	Edge
	Depth  int
	Side   string  // "pro" or "con": whose case the child claim argues
	Weight float64 // from the impact votes (see Weight)
}

// Plan walks the current relations from the thesis, breadth first, down to
// maxDepth levels (0: all), and returns the relations to import and a report
// of what was left out: relations unreachable from the thesis, with an empty
// claim, or that would close a cycle (a claim reused under its own
// descendant), which WAG's graphs may not have.
func (d *Discussion) Plan(maxDepth int) ([]Relation, *Report) {
	rep := &Report{Discussion: d.ID, Title: d.Title, Claims: len(d.Claims), Depth: maxDepth}
	children := map[string][]Edge{}
	for _, e := range d.Edges {
		if !e.Active {
			rep.Inactive++
			continue
		}
		children[e.From] = append(children[e.From], e)
	}
	for k := range children {
		sort.SliceStable(children[k], func(i, j int) bool { return children[k][i].LocationID < children[k][j].LocationID })
	}
	// reaches[x] holds the claims whose arguments x (transitively) supports:
	// adding child -> parent closes a cycle if child is reachable from parent
	// along those links, i.e. parent already argues for child.
	up := map[string][]string{} // premise claim -> conclusion claims
	reaches := func(from, to string) bool {
		seen := map[string]bool{}
		stack := []string{from}
		for len(stack) > 0 {
			x := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if x == to {
				return true
			}
			if seen[x] {
				continue
			}
			seen[x] = true
			stack = append(stack, up[x]...)
		}
		return false
	}
	side := map[string]string{d.ThesisClaimID: "pro"}
	depth := map[string]int{d.ThesisClaimID: 0}
	queue := []string{d.ThesisClaimID}
	var out []Relation
	visited := map[string]bool{}
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		if visited[parent] {
			continue
		}
		visited[parent] = true
		if maxDepth > 0 && depth[parent] >= maxDepth {
			rep.BeyondDepth += len(children[parent])
			continue
		}
		for _, e := range children[parent] {
			if strings.TrimSpace(d.Claims[e.To].Text) == "" {
				rep.Empty++
				continue
			}
			if reaches(parent, e.To) || e.To == parent {
				rep.Cycles++
				continue
			}
			s := side[parent]
			if e.Relation == "attack" {
				s = map[string]string{"pro": "con", "con": "pro"}[s]
			}
			if prev, ok := side[e.To]; ok && prev != s {
				rep.SideConflicts++
			} else if !ok {
				side[e.To] = s
				depth[e.To] = depth[parent] + 1
			}
			up[e.To] = append(up[e.To], parent)
			out = append(out, Relation{Edge: e, Depth: depth[parent] + 1, Side: s, Weight: Weight(e.Votes)})
			queue = append(queue, e.To)
		}
	}
	reached := 0
	for _, e := range d.Edges {
		if e.Active && visited[e.From] {
			reached++
		}
	}
	rep.Unreachable = len(d.Edges) - rep.Inactive - reached
	if rep.Unreachable < 0 {
		rep.Unreachable = 0
	}
	for _, r := range out {
		if r.Relation == "attack" {
			rep.Cons++
		} else {
			rep.Pros++
		}
		if len(r.Votes) > 0 && sum(r.Votes) > 0 {
			rep.Voted++
		}
	}
	return out, rep
}

// DefaultWeight is the weight of a relation nobody voted on.
const DefaultWeight = 0.5

// Weight turns Kialo impact votes (how many voters gave impact 0, 1, 2, 3,
// 4) into a Carneades weight: the mean impact over 4, at least 0.05 (an
// argument of weight 0 would not count at all); DefaultWeight without votes.
func Weight(votes []int) float64 {
	n, total := 0, 0
	for i, c := range votes {
		n += c
		total += i * c
	}
	if n == 0 {
		return DefaultWeight
	}
	w := float64(total) / float64(n) / 4
	if w < 0.05 {
		w = 0.05
	}
	return w
}

func sum(xs []int) int {
	t := 0
	for _, x := range xs {
		t += x
	}
	return t
}

// Report says what an import did.
type Report struct {
	Discussion    string `json:"discussion"`
	Title         string `json:"title"`
	Claims        int    `json:"claims"`
	Pros          int    `json:"pros"`
	Cons          int    `json:"cons"`
	Voted         int    `json:"voted"`           // relations with impact votes (the others weigh DefaultWeight)
	Inactive      int    `json:"inactive"`        // archived or otherwise inactive relations, not imported
	Unreachable   int    `json:"unreachable"`     // current relations not below the thesis
	Empty         int    `json:"empty"`           // relations to an empty claim
	Cycles        int    `json:"cycles"`          // relations left out because they would close a cycle
	SideConflicts int    `json:"side_conflicts"`  // claims reached as both pro and con (the first side is kept)
	BeyondDepth   int    `json:"beyond_depth"`    // relations below the depth limit
	Arguments     int    `json:"arguments"`       // arguments created
	Depth         int    `json:"depth,omitempty"` // the depth limit (0: none)
}
