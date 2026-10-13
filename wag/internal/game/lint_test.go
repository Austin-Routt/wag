package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
)

// The shipped data has no lint errors. Its warnings are logged: they are the
// known gaps of the 7-scheme NASCQ (schemes no critical question is asked of).
func TestShippedDataLint(t *testing.T) {
	issues, err := LintData(data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON)
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range issues {
		if i.Error {
			t.Error(i)
		} else {
			t.Log(i)
		}
	}
}

// Each kind of lint error is found.
func TestLintFindsDataErrors(t *testing.T) {
	var schemes, cqs map[string]map[string]interface{}
	json.Unmarshal(data.SchemesJSON, &schemes)
	json.Unmarshal(data.CQsJSON, &cqs)
	schemes["AS:sign"]["carneades_id"] = "no_such_scheme"
	schemes["AS:analogy"]["base_weight"] = 1.5
	schemes["AS:analogy"]["premise_templates_en"] = []string{"{C} is similar", "{Z} holds"}
	delete(schemes["AS:consequences"]["variables"].(map[string]interface{}), "G")
	cqs["CQ:possibility"]["caes_mapping"].(map[string]interface{})["target_premise_pattern"] = "feasible({A})"
	cqs["CQ:likelihood"]["caes_mapping"] = map[string]interface{}{"positive_answer_type": "support", "negative_answer_type": "support"}
	s, _ := json.Marshal(schemes)
	c, _ := json.Marshal(cqs)
	edges := strings.Replace(string(data.EdgesJSON), `"ask_edges": [`, `"ask_edges": [ { "cq": "CQ:evidence", "scheme": "AS:consequences" },
	  { "cq": "CQ:other_goals", "scheme": "AS:values", "locus": "exception", "pattern": "biased({A})" },`, 1)
	yaml := strings.Replace(string(data.CoreYAML), `good/1: "%s is good."`, `good/1: "%s is good for %s."`, 1)

	issues, err := LintData([]byte(yaml), s, c, []byte(edges))
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, i := range issues {
		if i.Error {
			all = append(all, i.String())
		}
	}
	joined := strings.Join(all, "\n")
	for _, want := range []string{
		`AS:sign: carneades_id "no_such_scheme" is not a scheme of the YAML`,
		`AS:analogy: base_weight 1.5 is not in (0, 1]`,
		`AS:analogy: template "{Z} holds" names undeclared variable "Z"`,
		`AS:consequences: Carneades scheme positive_consequences has variable G`,
		`CQ:possibility: asked of AS:practical_reasoning: premise locus feasible/1 is not a premise or assumption of Carneades scheme practical_reasoning1`,
		`CQ:likelihood asked of AS:consequences: no locus`,
		`duplicate ask edge CQ:evidence -> AS:consequences`,
		`good/1 has 2 %s for 1 arguments`,
		`CQ:other_goals: asked of AS:values: exception locus biased/1 is not an exception of Carneades scheme positive_value`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing lint error %q in:\n%s", want, joined)
		}
	}
}
