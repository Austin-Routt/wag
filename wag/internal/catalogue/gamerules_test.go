package catalogue

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

// One id is one question (C5.2, step 4; the refinement of D094): every
// scheme that asks it asks it at one formula and one kind of locus, with
// one text.
func TestQuestionFaults(t *testing.T) {
	weighing := []askedAt{
		{Scheme: "AS:practical_reasoning", Locus: "exception", Formula: "bad_outweighs({A},{G})", Text: "Would {A} cost more than {G}?"},
		{Scheme: "AS:consequences", Locus: "exception", Formula: "bad_outweighs({A},{G})", Text: "Would {A} cost more than {G}?"},
	}
	rebuttal := []askedAt{
		{Scheme: "AS:practical_reasoning", Locus: "conclusion", Formula: "should_be_performed({A})", Text: "Bad?"},
		{Scheme: "AS:consequences", Locus: "conclusion", Formula: "should_be_performed({A})", Text: "Bad?"},
	}
	undercut := []askedAt{{Scheme: "AS:sign", Locus: "argument", Text: "Other?"}}
	sound := map[string][]askedAt{"CQ:bad_outweighs": weighing, "CQ:side_effect": rebuttal, "CQ:other_event": undercut}
	if f := questionFaults(sound); len(f) > 0 {
		t.Fatalf("sound questions: %v", f)
	}

	// Before C5.2, CQ:side_effect was an exception of practical reasoning
	// and the conclusion of positive consequences.
	mixed := []askedAt{
		{Scheme: "AS:practical_reasoning", Locus: "exception", Formula: "side_effects({A},{S1})", Text: "Bad?"},
		{Scheme: "AS:consequences", Locus: "conclusion", Formula: "should_be_performed({A})", Text: "Bad?"},
	}
	joined := strings.Join(questionFaults(map[string][]askedAt{"CQ:side_effect": mixed}), "\n")
	for _, want := range []string{
		"CQ:side_effect is asked at conclusion (AS:consequences) and exception (AS:practical_reasoning): one id is one question, asked at one kind of locus",
		"CQ:side_effect is asked at should_be_performed({A}) (AS:consequences) and side_effects({A},{S1}) (AS:practical_reasoning): one id is one question, asked at one formula",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
	// Two formulas at one kind of locus, backed or not (CQ:likelihood
	// before sign's became CQ:correlation).
	two := []askedAt{
		{Scheme: "AS:consequences", Locus: "premise", Formula: "brings_about({A},{G})", Text: "How likely?"},
		{Scheme: "AS:sign", Locus: "premise", Formula: "strong_correlation({A},{B})", Text: "How likely?"},
	}
	f := questionFaults(map[string][]askedAt{"CQ:likelihood": two})
	if len(f) != 1 || !strings.Contains(f[0], "asked at brings_about({A},{G}) (AS:consequences) and strong_correlation({A},{B}) (AS:sign)") {
		t.Errorf("two formulas: %v", f)
	}
	// One formula and locus, two texts (CQ:too_slight before its entries
	// had one paraphrase).
	texts := []askedAt{
		{Scheme: "AS:consequences", Locus: "exception", Formula: "too_slight({A},{G})", Text: "Too slight a gain?"},
		{Scheme: "AS:negative_consequences", Locus: "exception", Formula: "too_slight({A},{G})", Text: "Too slight a harm?"},
	}
	f = questionFaults(map[string][]askedAt{"CQ:too_slight": texts})
	if len(f) != 1 || !strings.Contains(f[0], `CQ:too_slight reads "Too slight a gain?" (AS:consequences) and "Too slight a harm?" (AS:negative_consequences): one id is one question, with one text`) {
		t.Errorf("two texts: %v", f)
	}
}

// A question's caes_mapping follows from its locus.
func TestAnswerTypes(t *testing.T) {
	for locus, want := range map[string]string{
		"premise":    "support undermining premise",
		"exception":  "support undercutter argument",
		"argument":   "support undercutter argument",
		"conclusion": "support rebuttal conclusion",
	} {
		p, n, tg := answerTypes(locus)
		if got := p + " " + n + " " + tg; got != want {
			t.Errorf("%s: %s, want %s", locus, got, want)
		}
	}
}

// A question's record is GameDir's with the made text and effect, in the
// order critical_questions.json writes them; a record that writes either,
// lacks answer words or has an unknown key is refused.
func TestQuestionRecord(t *testing.T) {
	read := func(s string) yaml.MapSlice {
		var m yaml.MapSlice
		if err := yaml.Unmarshal([]byte(s), &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	rec, err := questionRecord("CQ:x", read("name: X\nsemantic_type: truth\nanswer_en: {against: \"no\", for: \"yes\"}\nrefuse_other: true\n"),
		"Is {A} so?", "premise")
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, it := range rec {
		keys = append(keys, it.Key.(string))
	}
	if got := strings.Join(keys, " "); got != "name semantic_type question_template_en answer_en caes_mapping refuse_other" {
		t.Errorf("keys: %s", got)
	}
	if value(rec, "question_template_en") != "Is {A} so?" || value(value(rec, "caes_mapping").(yaml.MapSlice), "negative_answer_type") != "undermining" {
		t.Errorf("record: %v", rec)
	}
	for _, c := range []struct{ rec, want string }{
		{"name: X\nsemantic_type: truth\nquestion_template_en: Is it?\nanswer_en: {against: \"no\", for: \"yes\"}\n", "gives question_template_en, which the generator makes"},
		{"name: X\nsemantic_type: truth\nanswer_en: {against: \"no\", for: \"yes\"}\ncaes_mapping: {}\n", "gives caes_mapping, which the generator makes"},
		{"name: X\nsemantic_type: truth\nanswer_en: {against: \"no\"}\n", "gives no words"},
		{"name: X\nsemantic_type: truth\nanswer_en: {against: \"no\", for: \" \"}\n", "gives no words"},
		{"name: X\nsemantic_type: truth\n", "has no answer_en"},
		{"semantic_type: truth\nanswer_en: {against: \"no\", for: \"yes\"}\n", "has no name"},
		{"name: X\nsemantic_type: truth\nanswer_en: {against: \"no\", for: \"yes\"}\nrefuse: true\n", "has the key refuse, not one of"},
	} {
		if _, err := questionRecord("CQ:x", read(c.rec), "t", "premise"); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%q: %v, want %q", c.rec, err, c.want)
		}
	}
}

// One formula is the same formula with each scheme's variables.
func TestSameFormula(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"bad_outweighs({A},{G})", "bad_outweighs({A},{G})", true},
		{"bad_outweighs({X},{Y})", "bad_outweighs({A},{G})", true},
		{"realize_more_effectively({P},{G},{A})", "realize_more_effectively({Q},{H},{B})", true},
		{"bad_outweighs({A},{A})", "bad_outweighs({A},{G})", false},
		{"bad_outweighs({A},{G})", "bad_outweighs({G},{G})", false},
		{"bad_outweighs({A},{G})", "good_outweighs({A},{G})", false},
		{"bad_outweighs({A},{G})", "bad_outweighs({A},g)", false},
		{"", "bad_outweighs({A},{G})", false},
	} {
		if got := sameFormula(c.a, c.b); got != c.want {
			t.Errorf("%s ~ %s: %v", c.a, c.b, got)
		}
	}
}

// A role scheme's record agrees with the formulas the game plays (rule
// (d)).
func TestRoleFaults(t *testing.T) {
	scheme := func(id, nascq string, roles map[string][]string) *gameIn {
		var rec yaml.MapSlice
		if err := yaml.Unmarshal([]byte(nascq), &rec); err != nil {
			t.Fatal(err)
		}
		g := &gameIn{id: id, entry: &Entry{ID: id, Game: &GameScheme{ID: id, NASCQ: rec}}, roles: roles, vars: map[string]bool{}}
		for _, v := range roles["variables"] {
			g.vars[v] = true
		}
		return g
	}
	cons := scheme("AS:consequences", "carneades_id: c\n", map[string][]string{"variables": {"A", "G"}})
	modal := scheme("AS:modal", "role: modal\nweighs_as: AS:consequences\n", map[string][]string{"variables": {"A", "G"}})
	bridge := scheme("AS:bridge", "role: bridge\nweighs_as: AS:modal\nentry_premise: may_be_worth_performing({A},{G})\n",
		map[string][]string{"variables": {"A", "G"}, "premises": {"may_be_worth_performing(A,G)", "matters_to_decision(A,G)"}})
	answer := scheme("AS:bad_outweighs_good", "role: answer\n",
		map[string][]string{"variables": {"A", "H", "G"}, "conclusions": {"bad_outweighs(A,G)"}})
	schemes := []*gameIn{cons, modal, bridge, answer}
	byID := map[string]*gameIn{}
	for _, g := range schemes {
		byID[g.id] = g
	}
	asked := map[string][]askedAt{
		"CQ:bad_outweighs": {{Scheme: "AS:consequences", Locus: "exception", Formula: "bad_outweighs({A},{G})"}},
		"CQ:side_effect":   {{Scheme: "AS:consequences", Locus: "conclusion", Formula: "should_be_performed({A})"}},
		"CQ:likelihood":    {{Scheme: "AS:consequences", Locus: "premise", Formula: "brings_about({A},{G})"}},
	}
	if f := roleFaults(schemes, byID, asked, map[string][]string{"AS:bad_outweighs_good": {"CQ:bad_outweighs"}}); len(f) > 0 {
		t.Fatalf("sound roles: %v", f)
	}
	for _, c := range []struct {
		name  string
		backs map[string][]string
		edit  func()
		want  string
	}{
		{"an answer backing a question about the conclusion", map[string][]string{"AS:bad_outweighs_good": {"CQ:side_effect"}}, nil,
			"AS:bad_outweighs_good backs CQ:side_effect, which AS:consequences asks at its conclusion: an answer scheme answers an exception or an assumption"},
		{"an answer concluding another formula", map[string][]string{"AS:bad_outweighs_good": {"CQ:likelihood"}}, nil,
			"AS:bad_outweighs_good concludes bad_outweighs({A},{G}), not brings_about({A},{G}), which AS:consequences asks CQ:likelihood at"},
		{"a modal weighing as no game scheme", nil, func() { modal.entry.Game.NASCQ[1].Value = "AS:none" },
			`AS:modal weighs as "AS:none", which is no game scheme`},
		{"a bridge's entry premise that is not its premise", nil, func() { bridge.entry.Game.NASCQ[2].Value = "may_be_worth_performing({G},{A})" },
			`AS:bridge's entry premise "may_be_worth_performing({G},{A})" is none of its premises`},
		{"an unknown role", nil, func() { answer.entry.Game.NASCQ[0].Value = "witness" },
			"AS:bad_outweighs_good: its role witness is not one of answer, modal, bridge"},
	} {
		saved := []yaml.MapSlice{copyMap(modal.entry.Game.NASCQ), copyMap(bridge.entry.Game.NASCQ), copyMap(answer.entry.Game.NASCQ)}
		if c.edit != nil {
			c.edit()
		}
		f := strings.Join(roleFaults(schemes, byID, asked, c.backs), "\n")
		if !strings.Contains(f, c.want) {
			t.Errorf("%s: %q, want %q", c.name, f, c.want)
		}
		modal.entry.Game.NASCQ, bridge.entry.Game.NASCQ, answer.entry.Game.NASCQ = saved[0], saved[1], saved[2]
	}
}

func copyMap(m yaml.MapSlice) yaml.MapSlice { return append(yaml.MapSlice{}, m...) }

// A paraphrase is plain: the generator braces its variables.
func TestParaphraseFault(t *testing.T) {
	for p, want := range map[string]string{
		"Can P actually do A?":   "",
		"":                       "has no paraphrase",
		"Can {P} actually do A?": "has braces",
	} {
		if got := paraphraseFault(p); (want == "") != (got == "") || !strings.Contains(got, want) {
			t.Errorf("%q: %q, want %q", p, got, want)
		}
	}
}
