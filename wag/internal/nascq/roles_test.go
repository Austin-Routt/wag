package nascq

import (
	"encoding/json"
	"strings"
	"testing"
)

// The roles of schemes in the game (D094 (g)) and the question that refuses
// Other (D095 (3)), on a small NASCQ of their own: a consequence scheme, its
// modal variant and the variant's bridge, an answer scheme for the bridge's
// exception, and a source scheme that asserts any statement.
const rolesSchemesJSON = `{
  "AS:cons": {"carneades_id": "positive_consequences", "name": "Positive Consequences", "semantic_type": "action",
    "base_weight": 0.65, "variables": {"A": {"type": "action"}, "G": {"type": "event"}},
    "premise_templates_en": ["Doing {A} brings about {G}", "{G} is good"], "conclusion_template_en": "{A} should be done"},
  "AS:cons_modal": {"carneades_id": "positive_consequences_modal", "name": "Possible Positive Consequences", "semantic_type": "action",
    "base_weight": 0.1, "role": "modal", "weighs_as": "AS:cons",
    "variables": {"A": {"type": "action"}, "G": {"type": "event"}},
    "premise_templates_en": ["Doing {A} may bring about {G}", "{G} is good"], "conclusion_template_en": "{A} may be worth doing for {G}"},
  "AS:bridge_good": {"carneades_id": "bridge_from_possible_good", "name": "From a Possible Good to the Action", "semantic_type": "action",
    "base_weight": 0.2, "role": "bridge", "weighs_as": "AS:cons_modal", "entry_premise": "may_be_worth_performing({A},{G})",
    "variables": {"A": {"type": "action"}, "G": {"type": "event"}},
    "premise_templates_en": ["{A} may be worth doing for {G}", "The chance matters"], "conclusion_template_en": "{A} should be done"},
  "AS:fof": {"carneades_id": "finding_out_first", "name": "Finding Out First", "semantic_type": "causality",
    "base_weight": 0.3, "role": "answer",
    "variables": {"A": {"type": "action"}, "G": {"type": "event"}, "I": {"type": "entity"}},
    "premise_templates_en": ["{I} would settle it", "{I} is done in time", "{I} is worth its costs"],
    "conclusion_template_en": "It is worth finding out first whether {A} brings about {G}"},
  "AS:expert": {"carneades_id": "expert_opinion", "name": "Expert Opinion", "semantic_type": "source",
    "base_weight": 0.6, "thesis_independent": true, "answer_variable": "S",
    "variables": {"W": {"type": "entity"}, "S": {"type": "proposition"}},
    "premise_templates_en": ["{W} is an expert", "{W} asserts {S}"], "conclusion_template_en": "{S}"}
}`

const rolesQuestionsJSON = `{
  "CQ:wf": {"name": "Finding Out First Question", "semantic_type": "causality", "refuse_other": true,
    "question_template_en": "Should finding out whether {A} brings about {G} come first?",
    "caes_mapping": {"positive_answer_type": "support", "negative_answer_type": "undercutter", "target": "argument"}},
  "CQ:slight": {"name": "Too Slight Question", "semantic_type": "value",
    "question_template_en": "Is {G} too slight to count?",
    "caes_mapping": {"positive_answer_type": "support", "negative_answer_type": "undercutter", "target": "argument"}},
  "CQ:likelihood": {"name": "Likelihood Question", "semantic_type": "causality",
    "question_template_en": "Is it likely?",
    "caes_mapping": {"positive_answer_type": "support", "negative_answer_type": "undermining", "target": "premise"}}
}`

const rolesEdgesJSON = `{
  "ask_edges": [
    {"cq": "CQ:wf", "scheme": "AS:bridge_good", "locus": "exception", "pattern": "worth_finding_out_first({A},{G})"},
    {"cq": "CQ:likelihood", "scheme": "AS:cons", "locus": "premise", "pattern": "brings_about({A},{G})"},
    {"cq": "CQ:slight", "scheme": "AS:cons", "locus": "exception", "pattern": "too_slight({A},{G})"},
    {"cq": "CQ:likelihood", "scheme": "AS:cons_modal", "locus": "premise", "pattern": "may_bring_about({A},{G})"}
  ],
  "backup_edges": [
    {"scheme": "AS:fof", "cq": "CQ:wf"}
  ]
}`

// rolesGraph is a board with one bridge argument, b1, about A = tax and
// G = g; every scheme's conclusion fits any statement (S9 is not the point).
type rolesGraph struct{}

func (rolesGraph) ArgumentExists(argID string) bool { return argID == "b1" || argID == "c1" }
func (rolesGraph) ArgumentSchemeID(argID string) string {
	return map[string]string{"b1": "AS:bridge_good", "c1": "AS:cons"}[argID]
}
func (rolesGraph) PremiseExists(argID, stmtID string) bool { return true }
func (rolesGraph) PremiseMatching(argID, pattern string) string {
	return strings.NewReplacer("{A}", "tax", "{G}", "g").Replace(pattern)
}
func (rolesGraph) PremiseIDs(argID string) []string { return nil }
func (rolesGraph) ExceptionMatching(argID, pattern string) string {
	if argID != "b1" && argID != "c1" {
		return ""
	}
	return strings.NewReplacer("{A}", "tax", "{G}", "g").Replace(pattern)
}
func (rolesGraph) ArgumentConclusion(argID string) string { return "should_be_performed(tax)" }
func (rolesGraph) Statement(stmtID string) StatementInfo {
	info := StatementInfo{Exists: true, Premise: true, Opposite: "¬" + stmtID}
	if rest, ok := strings.CutPrefix(stmtID, "¬"); ok {
		info.Opposite = rest
	}
	switch stmtID {
	case "worth_finding_out_first(tax,g)":
		info.ExceptionOf = []string{"b1"}
	case "too_slight(tax,g)":
		info.ExceptionOf = []string{"c1"}
	}
	return info
}
func (rolesGraph) Fits(schemeID, stmtID string, given map[string]string) (map[string]string, error) {
	return map[string]string{}, nil
}

func loadRoles(t *testing.T, mutate ...func(*NASCQ)) *NASCQ {
	t.Helper()
	n, err := LoadFromBytes([]byte(rolesSchemesJSON), []byte(rolesQuestionsJSON), []byte(rolesEdgesJSON))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mutate {
		m(n)
	}
	return n
}

var fofBindings = map[string]string{"A": "tax", "G": "g", "I": "a trial"}

// An answer scheme is offered only as an answer, on its backup edges: not
// as a move of its own, nor about a statement; as the answer to the
// bridge's question it is accepted.
func TestAnswerSchemeOnlyAsAnswer(t *testing.T) {
	n := loadRoles(t)
	g := rolesGraph{}
	for _, a := range []ProposedAction{
		{Type: "support", SchemeID: "AS:fof", VariableBindings: fofBindings},
		{Type: "support", SchemeID: "AS:fof", TargetStmtID: "worth_finding_out_first(tax,g)", VariableBindings: fofBindings},
		{Type: "attack", AttackType: "undercutter", SchemeID: "AS:fof", TargetArgID: "b1", VariableBindings: fofBindings},
	} {
		if _, err := n.Validate(a, g); err == nil || !strings.Contains(err.Error(), "S1: Finding Out First is an answer scheme") {
			t.Errorf("%+v: %v", a, err)
		}
	}
	va, err := n.Validate(ProposedAction{Type: "attack", SchemeID: "AS:fof", AnswersCQID: "CQ:wf", TargetArgID: "b1",
		VariableBindings: fofBindings}, g)
	if err != nil {
		t.Fatalf("the answer on its backup edge: %v", err)
	}
	if va.LocusStmtID != "worth_finding_out_first(tax,g)" {
		t.Errorf("the answer is about %q", va.LocusStmtID)
	}
	for _, tmpl := range n.EnumerateActions(g, nil) {
		if tmpl.SchemeID == "AS:fof" {
			t.Errorf("offered as a move of its own: %+v", tmpl)
		}
	}
	for _, tmpl := range n.EnumerateStatementActions(g, []string{"worth_finding_out_first(tax,g)"}) {
		if tmpl.SchemeID == "AS:fof" {
			t.Errorf("offered about a statement: %+v", tmpl)
		}
	}
	offered := false
	for _, tmpl := range n.EnumerateActions(g, []string{"b1"}) {
		if tmpl.SchemeID == "AS:fof" && tmpl.AnswersCQID == "CQ:wf" {
			offered = true
		}
	}
	if !offered {
		t.Error("not offered as the answer to the bridge's question")
	}
}

// A question that refuses Other (D095 (3)) takes no answer made with it, a
// well-made one included, and does not offer it; a scheme that asserts the
// answer (an answer variable) still answers, and Other answers elsewhere.
func TestQuestionRefusesOther(t *testing.T) {
	n := loadRoles(t)
	n.EnableOther()
	g := rolesGraph{}
	other := ProposedAction{Type: "attack", SchemeID: OtherSchemeID, AnswersCQID: "CQ:wf", TargetArgID: "b1",
		VariableBindings: map[string]string{"P1": "a trial would show it in time and cheaply", "P2": "it would settle it"}}
	_, err := n.Validate(other, g)
	if err == nil || !strings.Contains(err.Error(), `S5: CQ "CQ:wf" refuses Other`) || !strings.Contains(err.Error(), "Finding Out First") {
		t.Fatalf("Other at the question: %v", err)
	}
	for _, tmpl := range n.EnumerateActions(g, []string{"b1"}) {
		if tmpl.SchemeID == OtherSchemeID && tmpl.AnswersCQID == "CQ:wf" {
			t.Errorf("Other offered at the question: %+v", tmpl)
		}
	}
	if _, err := n.Validate(ProposedAction{Type: "attack", SchemeID: "AS:expert", AnswersCQID: "CQ:wf", TargetArgID: "b1",
		VariableBindings: map[string]string{"W": "dr_smith"}}, g); err != nil {
		t.Errorf("a source that asserts the exception: %v", err)
	}
	if _, err := n.Validate(ProposedAction{Type: "attack", SchemeID: OtherSchemeID, AnswersCQID: "CQ:likelihood", TargetArgID: "c1",
		VariableBindings: map[string]string{"P1": "rarely"}}, g); err != nil {
		t.Errorf("Other at a question that does not refuse it: %v", err)
	}
	// The exception argued by a move about it, or against its negation, is
	// refused with Other too; arguing against the exception is not.
	for _, a := range []ProposedAction{
		{Type: "support", SchemeID: OtherSchemeID, TargetStmtID: "worth_finding_out_first(tax,g)", VariableBindings: map[string]string{"P1": "it could be found out"}},
		{Type: "attack", SchemeID: OtherSchemeID, TargetStmtID: "¬worth_finding_out_first(tax,g)", VariableBindings: map[string]string{"P1": "it could be found out"}},
	} {
		if _, err := n.Validate(a, g); err == nil || !strings.Contains(err.Error(), `S5: worth_finding_out_first(tax,g) is the exception CQ "CQ:wf" asks about`) {
			t.Errorf("%s with Other: %v", a.Type, err)
		}
	}
	if _, err := n.Validate(ProposedAction{Type: "attack", SchemeID: OtherSchemeID, TargetStmtID: "worth_finding_out_first(tax,g)",
		VariableBindings: map[string]string{"P1": "there is no time to find out"}}, g); err != nil {
		t.Errorf("Other against the exception: %v", err)
	}
	if _, err := n.Validate(ProposedAction{Type: "support", SchemeID: "AS:expert", TargetStmtID: "worth_finding_out_first(tax,g)",
		VariableBindings: map[string]string{"W": "dr_smith"}}, g); err != nil {
		t.Errorf("a source for the exception, about the statement: %v", err)
	}
	// An exception whose question does not refuse Other is argued with it.
	if _, err := n.Validate(ProposedAction{Type: "support", SchemeID: OtherSchemeID, TargetStmtID: "too_slight(tax,g)",
		VariableBindings: map[string]string{"P1": "few would notice"}}, g); err != nil {
		t.Errorf("Other at an exception whose question takes it: %v", err)
	}
}

// Backs (D095 (3), D099 (4)): the schemes on a question's backup edges and
// the thesis-independent ones back it, Other too unless the question refuses
// it; a scheme with an answer variable that is neither does not.
func TestBacks(t *testing.T) {
	n := loadRoles(t)
	n.EnableOther()
	n.Schemes["AS:signlike"] = &ASNode{ID: "AS:signlike", Name: "Sign-like", AnswerVariable: "B",
		Variables: map[string]VariableSpec{"A": {Type: "event"}, "B": {Type: "event"}}}
	for _, c := range []struct {
		cq, scheme string
		want       bool
	}{
		{"CQ:wf", "AS:fof", true},          // its backup edge
		{"CQ:wf", "AS:expert", true},       // a source-based scheme backs every question
		{"CQ:wf", OtherSchemeID, false},    // the question refuses Other
		{"CQ:wf", "AS:signlike", false},    // concludes any statement, backs nothing
		{"CQ:wf", "AS:cons", false},        // no edge
		{"CQ:slight", OtherSchemeID, true}, // a question that takes Other
		{"CQ:slight", "AS:fof", false},
	} {
		if got := n.Backs(c.cq, n.Schemes[c.scheme]); got != c.want {
			t.Errorf("Backs(%s, %s) = %v, want %v", c.cq, c.scheme, got, c.want)
		}
	}
	if n.Backs("CQ:wf", nil) {
		t.Error("no scheme backs a question")
	}
}

// A modal variant weighs as its categorical counterpart, a bridge as its
// modal variant, so as the counterpart too, with no multiplier; an answer
// scheme as the argument it answers (D094 (g)). The counterpart's weight is
// the one a context sets.
func TestWeightsByRole(t *testing.T) {
	n := loadRoles(t)
	g := rolesGraph{}
	weight := func(a ProposedAction) float64 {
		t.Helper()
		va, err := n.Validate(a, g)
		if err != nil {
			t.Fatal(err)
		}
		return va.Weight
	}
	modal := ProposedAction{Type: "support", SchemeID: "AS:cons_modal", VariableBindings: map[string]string{"A": "tax", "G": "g"}}
	bridge := ProposedAction{Type: "support", SchemeID: "AS:bridge_good", VariableBindings: map[string]string{"A": "tax", "G": "g"}}
	answer := ProposedAction{Type: "attack", SchemeID: "AS:fof", AnswersCQID: "CQ:wf", TargetArgID: "b1", VariableBindings: fofBindings}
	for name, a := range map[string]ProposedAction{"modal": modal, "bridge": bridge, "answer": answer} {
		if w := weight(a); w != 0.65 {
			t.Errorf("%s weighs %v, want the consequence scheme's 0.65", name, w)
		}
	}
	n.Schemes["AS:cons"].BaseWeight = 0.4 // a context's weight
	for name, a := range map[string]ProposedAction{"modal": modal, "bridge": bridge, "answer": answer} {
		if w := weight(a); w != 0.4 {
			t.Errorf("after the context: %s weighs %v, want 0.4", name, w)
		}
	}
	if w := weight(ProposedAction{Type: "support", SchemeID: "AS:cons", VariableBindings: map[string]string{"A": "tax", "G": "g"}}); w != 0.4 {
		t.Errorf("the consequence scheme weighs %v", w)
	}
}

// Lint refuses a role the game cannot play.
func TestLintRoles(t *testing.T) {
	if issues := errorsOf(loadRoles(t).Lint()); len(issues) > 0 {
		t.Fatalf("the sound fixture: %v", issues)
	}
	// An answer scheme needs no question of its own (D075's reasoned none):
	// its premises are challenged as any premise is.
	for _, i := range loadRoles(t).Lint() {
		if i.Where == "AS:fof" {
			t.Errorf("the answer scheme: %s", i)
		}
	}
	for _, c := range []struct {
		name string
		m    func(*NASCQ)
		want string
	}{
		{"unknown role", func(n *NASCQ) { n.Schemes["AS:cons"].Role = "helper" }, `role "helper" is not`},
		{"answer without backup edge", func(n *NASCQ) { n.BackupEdges = nil; n.buildIndexes() }, "an answer scheme with no backup edge"},
		{"answer with an answer variable", func(n *NASCQ) { n.Schemes["AS:fof"].AnswerVariable = "A" }, "an answer scheme concludes the formula"},
		{"modal without weighs_as", func(n *NASCQ) { n.Schemes["AS:cons_modal"].WeighsAs = "" }, "a modal names the scheme it weighs as"},
		{"weighs_as on an argument", func(n *NASCQ) { n.Schemes["AS:cons"].WeighsAs = "AS:cons_modal" }, "weighs_as is for a modal variant or a bridge"},
		{"weighs_as nowhere", func(n *NASCQ) { n.Schemes["AS:cons_modal"].WeighsAs = "AS:none" }, `weighs_as "AS:none" is not a scheme`},
		{"weighs_as cycle", func(n *NASCQ) {
			n.Schemes["AS:cons_modal"].WeighsAs = "AS:bridge_good"
		}, "weighs_as makes a cycle"},
		{"bridge without entry premise", func(n *NASCQ) { n.Schemes["AS:bridge_good"].EntryPremise = "" }, "a bridge names its entry premise"},
		{"entry premise on a modal", func(n *NASCQ) { n.Schemes["AS:cons_modal"].EntryPremise = "x({A})" }, "entry_premise is for a bridge"},
		{"entry premise names a stranger", func(n *NASCQ) { n.Schemes["AS:bridge_good"].EntryPremise = "may_be_worth_performing({A},{H})" }, `names "H"`},
		{"distinct names one variable twice", func(n *NASCQ) { n.Schemes["AS:cons"].Distinct = [][]string{{"A", "A"}} }, "is not two different variables"},
		{"distinct names no action", func(n *NASCQ) { n.Schemes["AS:cons"].Distinct = [][]string{{"A", "G"}} }, "distinct names G, which is not an action variable"},
	} {
		issues := errorsOf(loadRoles(t, c.m).Lint())
		if !strings.Contains(strings.Join(issues, "\n"), c.want) {
			t.Errorf("%s: %v, want %q", c.name, issues, c.want)
		}
	}
}

func errorsOf(issues []LintIssue) []string {
	var out []string
	for _, i := range issues {
		if i.Error {
			out = append(out, i.String())
		}
	}
	return out
}

// A declaration is an object with its id, actor and aliases (D094 (b)).
func TestActionDeclJSON(t *testing.T) {
	var a ProposedAction
	if err := json.Unmarshal([]byte(`{"type":"support","declare":[{"id":"ban_meat","actor":"farmers","aliases":["meat ban"]}],"comparison":"cheaper"}`), &a); err != nil {
		t.Fatal(err)
	}
	if len(a.Declare) != 1 || a.Declare[0].ID != "ban_meat" || a.Declare[0].Actor != "farmers" || a.Declare[0].Aliases[0] != "meat ban" || a.Comparison != "cheaper" {
		t.Fatalf("%+v", a)
	}
	if err := json.Unmarshal([]byte(`{"declare":["ban_meat"]}`), &a); err == nil {
		t.Error("a bare id was taken as a declaration")
	}
}

// Cases XA1 and XA2 (docs/c5-game.md, C5.3's switch): a backup edge whose
// scheme answers a question of another semantic type is warned of, unless
// it is marked across by a decision (the catalogue checks the decision).
func TestLintAcross(t *testing.T) {
	n := loadRoles(t)
	across := func() int {
		c := 0
		for _, i := range n.Lint() {
			if strings.Contains(i.Msg, "across semantic types") {
				c++
			}
		}
		return c
	}
	if c := across(); c != 0 {
		t.Fatalf("the fixture: %d warnings", c)
	}
	n.Questions["CQ:wf"].SemanticType = "action" // Finding Out First is causality
	if c := across(); c != 1 {
		t.Errorf("XA1: %d warnings, want 1", c)
	}
	n.BackupEdges[0].Across = "D009"
	if c := across(); c != 0 {
		t.Errorf("XA2: %d warnings, want 0", c)
	}
}

// Case XA5 (docs/c5-game.md, C5.3's switch; added after its independent
// review, N3, before this test): a mark on a backup edge that does not cross
// semantic types (the same type, or either has none) suppresses no warning,
// so it is stale, and the lint says so.
func TestLintAcrossStale(t *testing.T) {
	n := loadRoles(t)
	stale := func() int {
		c := 0
		for _, i := range n.Lint() {
			if strings.Contains(i.Msg, "does not cross semantic types") {
				if i.Error {
					t.Errorf("a stale mark is a warning, not an error: %+v", i)
				}
				c++
			}
		}
		return c
	}
	if c := stale(); c != 0 {
		t.Fatalf("the fixture: %d warnings", c)
	}
	n.BackupEdges[0].Across = "D009" // the edge is of one type
	if c := stale(); c != 1 {
		t.Errorf("XA5, the same type: %d warnings, want 1", c)
	}
	n.Questions["CQ:wf"].SemanticType = "" // a type missing
	if c := stale(); c != 1 {
		t.Errorf("XA5, a type missing: %d warnings, want 1", c)
	}
	n.Questions["CQ:wf"].SemanticType = "action" // now it crosses, as XA2
	if c := stale(); c != 0 {
		t.Errorf("XA5, an edge that crosses: %d warnings, want 0", c)
	}
}
