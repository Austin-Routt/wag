package nascq

import "testing"

func TestLocusOf(t *testing.T) {
	schemes := `{"AS:s": {"carneades_id": "s", "name": "S", "base_weight": 0.5, "variables": {"A": {"type": "action"}}}}`
	cqs := `{
	  "CQ:u": {"name": "U", "caes_mapping": {"positive_answer_type": "support", "negative_answer_type": "undermining", "target_premise_pattern": "possible({A})"}},
	  "CQ:c": {"name": "C", "caes_mapping": {"positive_answer_type": "undercutter", "negative_answer_type": "support"}},
	  "CQ:r": {"name": "R", "caes_mapping": {"positive_answer_type": "rebuttal", "negative_answer_type": "support"}},
	  "CQ:x": {"name": "X", "caes_mapping": {"positive_answer_type": "support", "negative_answer_type": "support"}},
	  "CQ:e": {"name": "E", "caes_mapping": {"positive_answer_type": "rebuttal", "negative_answer_type": "support"}}
	}`
	edges := `{"ask_edges": [
	  {"cq": "CQ:u", "scheme": "AS:s"}, {"cq": "CQ:c", "scheme": "AS:s"}, {"cq": "CQ:r", "scheme": "AS:s"},
	  {"cq": "CQ:x", "scheme": "AS:s"}, {"cq": "CQ:e", "scheme": "AS:s", "locus": "exception", "pattern": "side_effects({A})"}
	]}`
	n, err := LoadFromBytes([]byte(schemes), []byte(cqs), []byte(edges))
	if err != nil {
		t.Fatal(err)
	}
	for cq, want := range map[string]Locus{
		"CQ:u": {Kind: LocusPremise, Pattern: "possible({A})"},
		"CQ:c": {Kind: LocusArgument},
		"CQ:r": {Kind: LocusConclusion},
		"CQ:x": {},
		"CQ:e": {Kind: LocusException, Pattern: "side_effects({A})"}, // the edge's locus wins over caes_mapping
	} {
		if got := n.LocusOf(cq, "AS:s"); got != want {
			t.Errorf("LocusOf(%s) = %+v, want %+v", cq, got, want)
		}
	}
	for kind, at := range map[string]string{LocusPremise: "undermining", LocusException: "undercutter",
		LocusConclusion: "rebuttal", LocusArgument: "undercutter", "": ""} {
		if got := (Locus{Kind: kind}).AttackType(); got != at {
			t.Errorf("Locus{%q}.AttackType() = %q, want %q", kind, got, at)
		}
	}
}
