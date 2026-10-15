package kialo

import (
	"encoding/json"
	"testing"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// A small discussion in the visualizer's format: the thesis T has a pro P
// (strong) and a con C (weak); C has a con D that defeats it; X under D is
// the same claim as P's parent T reused below (a cycle), and an archived pro.
const sample = `{
 "id": "1", "displayTitle": "Should we?", "discussionUrl": "https://example.org/1",
 "thesis_claim_id": "1.1", "top_location_id": "1.10",
 "claims": {
  "1.1": {"id": "1.1", "text": "We should do it."},
  "1.2": {"id": "1.2", "text": "It helps people."},
  "1.3": {"id": "1.3", "text": "It is expensive."},
  "1.4": {"id": "1.4", "text": "It pays for itself within two years."},
  "1.5": {"id": "1.5", "text": "Archived point."}
 },
 "edges": [
  {"location_id": "1.11", "from_claim_id": "1.1", "to_claim_id": "1.2", "relation": "support", "vote_vector": [0,0,0,1,3], "active": true},
  {"location_id": "1.12", "from_claim_id": "1.1", "to_claim_id": "1.3", "relation": "attack", "vote_vector": [1,1,1,0,0], "active": true},
  {"location_id": "1.13", "from_claim_id": "1.3", "to_claim_id": "1.4", "relation": "attack", "vote_vector": [], "active": true},
  {"location_id": "1.14", "from_claim_id": "1.4", "to_claim_id": "1.1", "relation": "support", "vote_vector": [], "active": true},
  {"location_id": "1.15", "from_claim_id": "1.1", "to_claim_id": "1.5", "relation": "support", "vote_vector": [], "active": false}
 ]
}`

func TestImport(t *testing.T) {
	d, err := Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	rels, rep := d.Plan(0)
	if len(rels) != 3 || rep.Cycles != 1 || rep.Inactive != 1 || rep.Pros != 1 || rep.Cons != 2 {
		t.Fatalf("plan: %d relations, report %+v", len(rels), rep)
	}
	sides := map[string]string{}
	for _, r := range rels {
		sides[r.To] = r.Side
	}
	if sides["1.2"] != "pro" || sides["1.3"] != "con" || sides["1.4"] != "pro" {
		t.Fatalf("sides %v", sides)
	}
	ad, err := carneades.NewAdapterFromYAML(data.CoreYAML)
	if err != nil {
		t.Fatal(err)
	}
	n, _ := nascq.LoadFromBytes(data.SchemesJSON, data.CQsJSON, data.EdgesJSON)
	gs, err := Build(d, rels, rep, Config(), ad, n)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Arguments != 3 {
		t.Fatalf("%d arguments", rep.Arguments)
	}
	// The pro (weight 1.0 - well, 3.75/4) beats the con (0.25), which its
	// own con (0.5) defeats anyway: the thesis is In.
	thesis := ad.Graph.Statements[gs.Claim]
	if thesis.Label != caes.In || gs.Say(gs.Claim) != "We should do it." {
		t.Fatalf("thesis %q is %v", gs.Say(gs.Claim), thesis.Label)
	}
	if l := ad.Graph.Statements[ad.FreeStatementID("It is expensive.")].Label; l != caes.Out {
		t.Fatalf("the defeated con is %v", l)
	}
	if errs := gs.CheckInvariants(); len(errs) > 0 {
		t.Fatal(errs)
	}
	// Play continues from there: Con questions the pro's premise.
	var pro string
	for id, a := range ad.Graph.Arguments {
		if a.Metadata["kialo_claim"] == "1.2" {
			pro = id
		}
	}
	gs.ExecuteTurn(nascq.ProposedAction{Type: "abstain"}) // Pro
	r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "question", AnswersCQID: nascq.CQPremiseAcceptable, TargetArgID: pro,
		TargetPremiseID: ad.FreeStatementID("It helps people.")})
	if !r.Accepted {
		t.Fatalf("question on an imported argument: %+v", r)
	}
}

func TestWeight(t *testing.T) {
	for _, c := range []struct {
		votes []int
		want  float64
	}{{nil, DefaultWeight}, {[]int{0, 0, 0, 0, 2}, 1}, {[]int{3, 0, 0, 0, 0}, 0.05}, {[]int{0, 1, 0, 1, 0}, 0.5}} {
		if got := Weight(c.votes); got != c.want {
			t.Errorf("Weight(%v) = %v, want %v", c.votes, got, c.want)
		}
	}
}

// A discussion as an annotation dataset: top-down, with each claim's
// context, in the format of docs/kialo-batch2/items.json.
func TestItems(t *testing.T) {
	d, _ := Parse([]byte(sample))
	b, n, err := d.Items(0, 0)
	if err != nil || n != 4 {
		t.Fatalf("%v %d", err, n)
	}
	var f struct {
		Discussion map[string]string `json:"discussion"`
		Items      []struct {
			ID        string `json:"kialo_claim_id"`
			Depth     int    `json:"depth"`
			Parents   []map[string]string
			Ancestors []map[string]string `json:"ancestors"`
			Siblings  []map[string]string `json:"siblings"`
			Children  []map[string]string `json:"children"`
		} `json:"items"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.Discussion["thesis_claim_id"] != "1.1" || f.Items[0].ID != "1.1" || f.Items[3].ID != "1.4" || f.Items[3].Depth != 2 {
		t.Fatalf("%s", b)
	}
	d4 := f.Items[3]
	if len(d4.Ancestors) != 2 || d4.Ancestors[0]["claim_id"] != "1.3" || d4.Ancestors[0]["relation_of_child"] != "con" || d4.Ancestors[1]["claim_id"] != "1.1" {
		t.Fatalf("ancestors %v", d4.Ancestors)
	}
	if len(f.Items[1].Siblings) != 1 || f.Items[1].Siblings[0]["claim_id"] != "1.3" || len(f.Items[0].Children) != 2 {
		t.Fatalf("siblings %v children %v", f.Items[1].Siblings, f.Items[0].Children)
	}
	if _, n, _ := d.Items(1, 0); n != 3 {
		t.Fatalf("depth 1: %d claims", n)
	}
	if _, n, _ := d.Items(0, 2); n != 2 {
		t.Fatalf("limit 2: %d claims", n)
	}
}
