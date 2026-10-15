package kialo

import (
	"encoding/json"
	"fmt"
)

// Items writes the discussion as an annotation dataset's items.json (the
// format of docs/kialo-batch2/items.json): the claims reached from the
// thesis, top-down, down to maxDepth levels (0: all), at most limit claims
// (0: all), each with its Kialo context: its parents and their relation,
// its ancestors up to the thesis, its siblings and children, its impact
// votes.
func (d *Discussion) Items(maxDepth, limit int) ([]byte, int, error) {
	rels, _ := d.Plan(maxDepth)
	rel := map[string]string{"support": "pro", "attack": "con"}
	type link struct {
		ClaimID    string `json:"claim_id"`
		Relation   string `json:"relation"`
		LocationID string `json:"location_id,omitempty"`
		Text       string `json:"text,omitempty"`
	}
	type ancestor struct {
		ClaimID         string `json:"claim_id"`
		RelationOfChild string `json:"relation_of_child"`
		Text            string `json:"text"`
	}
	type item struct {
		ID       string     `json:"kialo_claim_id"`
		Depth    int        `json:"depth"`
		Text     string     `json:"text"`
		Parents  []link     `json:"kialo_parents"`
		Ancestor []ancestor `json:"ancestors"`
		Siblings []link     `json:"siblings"`
		Children []link     `json:"children"`
		Votes    []int      `json:"impact_votes,omitempty"`
	}
	parents := map[string][]Relation{}
	children := map[string][]Relation{}
	order := []string{d.ThesisClaimID}
	depth := map[string]int{d.ThesisClaimID: 0}
	for _, r := range rels {
		parents[r.To] = append(parents[r.To], r)
		children[r.From] = append(children[r.From], r)
		if _, ok := depth[r.To]; !ok {
			depth[r.To] = r.Depth
			order = append(order, r.To)
		}
	}
	if limit > 0 && limit < len(order) {
		order = order[:limit]
	}
	var items []item
	for _, id := range order {
		it := item{ID: id, Depth: depth[id], Text: d.Claims[id].Text, Parents: []link{}, Ancestor: []ancestor{}, Siblings: []link{}, Children: []link{}}
		for i, p := range parents[id] {
			it.Parents = append(it.Parents, link{ClaimID: p.From, Relation: rel[p.Relation], LocationID: p.LocationID})
			if i == 0 {
				it.Votes = p.Votes
				for _, s := range children[p.From] {
					if s.To != id {
						it.Siblings = append(it.Siblings, link{ClaimID: s.To, Relation: rel[s.Relation], Text: d.Claims[s.To].Text})
					}
				}
			}
		}
		// The ancestors, by first parents, up to the thesis.
		seen := map[string]bool{id: true}
		for x := id; len(parents[x]) > 0; {
			p := parents[x][0]
			if seen[p.From] {
				break
			}
			seen[p.From] = true
			it.Ancestor = append(it.Ancestor, ancestor{ClaimID: p.From, RelationOfChild: rel[p.Relation], Text: d.Claims[p.From].Text})
			x = p.From
		}
		for _, c := range children[id] {
			it.Children = append(it.Children, link{ClaimID: c.To, Relation: rel[c.Relation], Text: d.Claims[c.To].Text})
		}
		items = append(items, it)
	}
	if len(items) == 0 {
		return nil, 0, fmt.Errorf("no claim")
	}
	b, err := json.MarshalIndent(map[string]any{
		"discussion": map[string]string{"id": d.ID, "title": d.Title, "url": d.URL, "thesis_claim_id": d.ThesisClaimID,
			"thesis": d.Claims[d.ThesisClaimID].Text},
		"note":  "Made by WAG from a capture of the Kialo research visualizer: active Kialo edges only, top-down (parents before children). Impact votes: counts per level 0-4; they prioritise review, they are not evidence.",
		"items": items,
	}, "", " ")
	return b, len(items), err
}
