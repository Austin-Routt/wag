package caes

import "github.com/carneades/carneades-4/src/engine/terms"

// normalizeGraph is used only on a private inference/context copy. It makes
// explicit double negation share the same statement identity as the rule
// store. Distinct objects with one normalized key are rejected rather than
// dropping one object's text, arguments or issue membership.
func (ag *ArgGraph) normalizeGraph() error {
	statements := make(map[string]*Statement, len(ag.Statements))
	for key, stmt := range ag.Statements {
		t, ok := terms.ReadFormula(key)
		if !ok || stmt == nil {
			return NewRunError(ErrInvalidInput, "invalid statement key or object %q", key)
		}
		n := t.String()
		if stmt.Id != "" && terms.Normalize(stmt.Id) != n {
			return NewRunError(ErrInvalidInput, "statement key %q disagrees with its id %q", key, stmt.Id)
		}
		if old, exists := statements[n]; exists && old != stmt {
			return NewRunError(ErrInvalidInput, "multiple statement objects normalize to %q", n)
		}
		stmt.Id = n
		statements[n] = stmt
	}
	ag.Statements = statements
	for i, s := range ag.Assumptions {
		ag.Assumptions[i] = terms.Normalize(s)
	}
	for i, s := range ag.Questioned {
		ag.Questioned[i] = terms.Normalize(s)
	}
	defaults := map[string]bool{}
	for key, active := range ag.Defaults {
		defaults[terms.Normalize(key)] = defaults[terms.Normalize(key)] || active
	}
	ag.Defaults = defaults
	expected := map[string]Label{}
	for key, label := range ag.ExpectedLabeling {
		n := terms.Normalize(key)
		if old, exists := expected[n]; exists && old != label {
			return NewRunError(ErrInvalidInput, "conflicting expected labels normalize to %q", n)
		}
		expected[n] = label
	}
	ag.ExpectedLabeling = expected
	return nil
}
