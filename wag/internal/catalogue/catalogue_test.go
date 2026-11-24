package catalogue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

const root = "../../.."

func sources(t *testing.T) *Sources {
	t.Helper()
	s, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// generate writes a fresh catalogue into a temporary directory.
func generate(t *testing.T, s *Sources) string {
	t.Helper()
	dir := t.TempDir()
	g := &Generator{Src: s, Today: "2026-10-22", By: "test"}
	if _, err := g.Write(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func errorsOf(r *Report) []string {
	var out []string
	for _, f := range r.Findings {
		if f.Error {
			out = append(out, f.String())
		}
	}
	return out
}

func readEntry(t *testing.T, dir, id string) *Entry {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, SchemesDir, id+".yml"))
	if err != nil {
		t.Fatal(err)
	}
	e, err := Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func writeEntry(t *testing.T, dir string, e *Entry) {
	t.Helper()
	b, err := Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, SchemesDir, e.ID+".yml"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// A fresh skeleton passes the check, and covers every scheme, compendium
// scheme, book question and game question.
func TestGeneratedCatalogueChecks(t *testing.T) {
	s := sources(t)
	dir := generate(t, s)
	r, err := Check(dir, s, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if errs := errorsOf(r); len(errs) > 0 {
		t.Fatalf("%d errors, e.g.:\n%s", len(errs), strings.Join(errs[:min(len(errs), 10)], "\n"))
	}
	books := 0
	for _, l := range s.Links {
		if !l.HeadingOnly && len(l.Links) == 0 {
			books++
		}
	}
	if want := len(s.Walton) + books; r.Entries != want {
		t.Errorf("%d entries, want %d (%d walton.yml schemes and %d book schemes walton.yml lacks)", r.Entries, want, len(s.Walton), books)
	}
	// Every game question is in the entry of its scheme.
	for _, a := range s.AskEdges {
		e := readEntry(t, dir, s.migrate(s.Game[a.Scheme].CarneadesID))
		found := false
		for _, q := range e.CQs {
			if q.Game != nil && q.Game.ID == a.CQ {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: game question %s missing", e.ID, a.CQ)
		}
	}
}

// The committed catalogue is current: generating it again changes nothing,
// and it passes the check.
func TestCommittedCatalogueIsCurrent(t *testing.T) {
	s := sources(t)
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	if _, err := os.Stat(dir); err != nil {
		t.Skip("no catalogue yet")
	}
	r, err := Check(dir, s, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if errs := errorsOf(r); len(errs) > 0 {
		t.Fatalf("the catalogue fails the check (run wag-cli --catalogue-check):\n%s", strings.Join(errs[:min(len(errs), 20)], "\n"))
	}
	entries, _, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	g := &Generator{Src: s, Today: "2026-10-22", By: "test"}
	for _, gen := range g.Build() {
		old := entries[gen.ID+".yml"]
		a, _ := Marshal(Merge(old, gen))
		b, _ := Marshal(old)
		if string(a) != string(b) {
			t.Errorf("%s.yml is not what wag-cli --catalogue-gen makes of the sources: run it", gen.ID)
		}
	}
}

// The checker catches each kind of fault.
func TestCheckFindsFaults(t *testing.T) {
	s := sources(t)
	cases := []struct {
		name, want string
		edit       func(e *Entry)
	}{
		{"formula edited", "formulas: not walton.yml's", func(e *Entry) { e.Formulas.Premises[0] = "expert(W)" }},
		{"formulas reordered", "formulas: not walton.yml's", func(e *Entry) {
			e.Formulas.Premises[0], e.Formulas.Premises[1] = e.Formulas.Premises[1], e.Formulas.Premises[0]
		}},
		{"question targets a foreign formula", "is not a formula of walton.yml's expert_opinion", func(e *Entry) {
			e.CQs[0].Status = "drafted by a model, 2026-10-22"
			e.CQs[0].Paraphrase = "Is W good?"
			e.CQs[0].Targets = []Target{{Formula: "good(W)", Locus: "premise"}}
		}},
		{"question in the wrong role", "is a premise of walton.yml's expert_opinion, not an exception", func(e *Entry) {
			e.CQs[0].Status = "drafted by a model, 2026-10-22"
			e.CQs[0].Paraphrase = "Is W an expert at all?"
			e.CQs[0].Targets[0].Locus = "exception"
		}},
		{"game question dropped", "game question CQ:expertise is missing", func(e *Entry) { e.CQs[0].Game = nil }},
		{"book question dropped", "question 2/CQ1 (compendium) is missing", func(e *Entry) { e.CQs = e.CQs[1:] }},
		{"game text edited", "the game part is not the NASCQ's", func(e *Entry) { e.CQs[1].Game.Text = "Is it?" }},
		{"bad status", "status \"done\"", func(e *Entry) { e.Summary.Status = "done" }},
		{"bad date", "bad date", func(e *Entry) { e.Summary = Text{Text: "x", Status: "verified by A. Person, 2026-13-40"} }},
		{"verified without text", "summary: verified by A. Person, 2026-10-22, but no text", func(e *Entry) { e.Summary.Status = "verified by A. Person, 2026-10-22" }},
		{"unknown dialogue type", "dialogue type \"debate\"", func(e *Entry) { e.DialogueTypes.Proper = []string{"debate"} }},
		{"unknown contrast", "contrast with \"no_such_scheme\": no such entry", func(e *Entry) {
			e.Contrasts = append(e.Contrasts, Contrast{Scheme: "no_such_scheme", Status: Todo})
		}},
		{"book entry dropped", "book entry 2 (links.json) is missing", func(e *Entry) { e.Book = nil }},
		{"family dropped", "family F01 (families.json) is missing", func(e *Entry) { e.Families = nil }},
		{"bad classification", "classification", func(e *Entry) { e.Class.Reasoning = "deductive" }},
		{"group outside its heading", "group \"instrumental\" under heading \"dependent\"", func(e *Entry) {
			e.Class = Classification{Source: "dependent", Group: "instrumental", Status: "drafted by a model, 2026-10-22"}
		}},
		{"classification not placed", "but not placed down to a group", func(e *Entry) {
			e.Class = Classification{Source: "independent", Reasoning: "epistemic", Status: "drafted by a model, 2026-10-22"}
		}},
		{"walton question with a bad id", "needs the id walton/<predicate>", func(e *Entry) {
			e.CQs = append(e.CQs, CQ{ID: "own/1", Paraphrase: "Is W honest?", Status: "drafted by a model, 2026-10-22",
				Targets: []Target{{Formula: "untrustworthy(W)", Locus: "exception"}}})
		}},
		{"walton question on a premise", "targets one assumption or exception", func(e *Entry) {
			e.CQs = append(e.CQs, CQ{ID: "walton/expert", Paraphrase: "Is W an expert?", Status: "drafted by a model, 2026-10-22",
				Targets: []Target{{Formula: "expert(W,D)", Locus: "premise"}}})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := generate(t, s)
			e := readEntry(t, dir, "expert_opinion")
			tc.edit(e)
			writeEntry(t, dir, e)
			r, err := Check(dir, s, DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			errs := strings.Join(errorsOf(r), "\n")
			if !strings.Contains(errs, tc.want) {
				t.Errorf("want an error containing %q; got:\n%s", tc.want, errs)
			}
		})
	}
	// A discrepancy the sources show must stay.
	dir := generate(t, s)
	e := readEntry(t, dir, "sign")
	e.Discrepancies = nil
	writeEntry(t, dir, e)
	r, _ := Check(dir, s, DefaultOptions())
	if errs := strings.Join(errorsOf(r), "\n"); !strings.Contains(errs, "discrepancy game-differs") {
		t.Errorf("a dropped discrepancy is not reported:\n%s", errs)
	}
	// A missing entry, and a stale index.
	dir = generate(t, s)
	os.Remove(filepath.Join(dir, SchemesDir, "threat.yml"))
	r, _ = Check(dir, s, DefaultOptions())
	errs := strings.Join(errorsOf(r), "\n")
	if !strings.Contains(errs, "no entry for threat") || !strings.Contains(errs, "index.md: out of date") {
		t.Errorf("a missing entry or a stale index is not reported:\n%s", errs)
	}
}

// Regenerating keeps what people wrote, refreshes the sources, and adds
// what the sources add.
func TestMergeKeepsWork(t *testing.T) {
	s := sources(t)
	dir := generate(t, s)
	e := readEntry(t, dir, "expert_opinion")
	e.Summary = Text{Text: "Our summary.", Status: "drafted by a model, 2026-10-22"}
	e.CQs[0].Paraphrase = "How good an expert in D is W?"
	e.CQs[0].Status = "verified by A. Person, 2026-10-23"
	e.CQs[0].Game.Text = "stale text"                          // source: refreshed
	e.CQs[1].Note = "a generated item edited without a status" // still generated: overwritten
	e.Formulas.Premises[0] = "x(W)"                            // source: refreshed
	e.Examples = []Example{{Text: "Dr X says so.", Coding: "expert_opinion(W=dr_x)", Status: "drafted by a model, 2026-10-22"}}
	e.Tier = 2
	writeEntry(t, dir, e)
	os.Remove(filepath.Join(dir, SchemesDir, "threat.yml"))
	g := &Generator{Src: s, Today: "2026-11-01", By: "test"}
	res, err := g.Write(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Created != 1 || res.Updated != 1 {
		t.Errorf("result %+v, want 1 created (threat), 1 updated (expert_opinion)", res)
	}
	m := readEntry(t, dir, "expert_opinion")
	switch {
	case m.Summary.Text != "Our summary.":
		t.Error("the drafted summary was lost")
	case m.CQs[0].Paraphrase != "How good an expert in D is W?" || m.CQs[0].Status != "verified by A. Person, 2026-10-23":
		t.Error("the verified question was lost")
	case m.CQs[0].Game == nil || m.CQs[0].Game.Text == "stale text":
		t.Error("the game part of a kept question was not refreshed")
	case m.CQs[1].Note == "a generated item edited without a status":
		t.Error("a generated item was not regenerated")
	case m.Formulas.Premises[0] == "x(W)":
		t.Error("the formulas were not refreshed")
	case len(m.Examples) != 1:
		t.Error("the example was lost")
	case m.Tier != 2:
		t.Error("the tier set by a person was lost")
	case m.History[0].Date != "2026-10-22":
		t.Error("the history was rewritten")
	}
	if readEntry(t, dir, "threat").History[0].Date != "2026-11-01" {
		t.Error("a new entry should be dated the day it is made")
	}
	r, _ := Check(dir, s, DefaultOptions())
	if errs := errorsOf(r); len(errs) > 0 {
		t.Errorf("after a merge: %s", strings.Join(errs, "\n"))
	}
	// Generating twice changes nothing.
	res, _ = g.Write(dir)
	if res.Created+res.Updated != 0 {
		t.Errorf("a second run changed files: %+v", res)
	}
}

// A kept question that took a game question from another keeps it, and the
// game question appears once.
func TestMergeMovesGameQuestion(t *testing.T) {
	s := sources(t)
	dir := generate(t, s)
	e := readEntry(t, dir, "expert_opinion")
	g := e.CQs[0].Game // CQ:expertise, matched to 2/CQ1
	e.CQs[0].Game = nil
	e.CQs[1].Game = nil
	e.CQs[1].Game = g // a person moves it to 2/CQ2 ...
	e.CQs[1].Paraphrase = "Is S in D?"
	e.CQs[1].Status = "verified by A. Person, 2026-10-23"
	writeEntry(t, dir, e)
	gen := &Generator{Src: s, Today: "2026-10-22", By: "test"}
	if _, err := gen.Write(dir); err != nil {
		t.Fatal(err)
	}
	m := readEntry(t, dir, "expert_opinion")
	n := 0
	for _, q := range m.CQs {
		if q.Game != nil && q.Game.ID == "CQ:expertise" {
			n++
			if q.ID != "2/CQ2" {
				t.Errorf("CQ:expertise is on %s", q.ID)
			}
		}
	}
	if n != 1 {
		t.Errorf("CQ:expertise appears %d times", n)
	}
	// CQ:field, which the person dropped from 2/CQ2, stands on its own or
	// returns to a generated item.
	found := false
	for _, q := range m.CQs {
		if q.Game != nil && q.Game.ID == "CQ:field" {
			found = true
		}
	}
	if !found {
		t.Error("CQ:field was lost")
	}
	r, _ := Check(dir, s, DefaultOptions())
	if errs := errorsOf(r); len(errs) > 0 {
		t.Errorf("after moving a game question: %s", strings.Join(errs, "\n"))
	}
}

// The copy check catches the book's sentence, even with its variables
// renamed, and lets a paraphrase through.
func TestCopyCheck(t *testing.T) {
	s := sources(t)
	book := NewCorpus([]string{
		"Is the source named in the argument an expert in the field that the proposition A is in?",
		"How credible is E as an expert source?",
	})
	opt := DefaultOptions()
	opt.Book = book
	cases := []struct {
		text  string
		error bool
		warn  bool
	}{
		{"Is the source named in the argument an expert in the field that the claim S is in?", true, false},
		{"How credible is W as an expert source?", true, false},
		{"How far can W's word be trusted on matters of D?", false, false},
	}
	for _, tc := range cases {
		dir := generate(t, s)
		e := readEntry(t, dir, "expert_opinion")
		e.CQs[0].Paraphrase = tc.text
		e.CQs[0].Status = "drafted by a model, 2026-10-22"
		writeEntry(t, dir, e)
		r, err := Check(dir, s, opt)
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Join(errorsOf(r), "\n")
		if tc.error != strings.Contains(got, "reproduces") {
			t.Errorf("%q: error %v, want %v (%s)", tc.text, !tc.error, tc.error, got)
		}
		if r.Texts != 1 {
			t.Errorf("%q: %d texts compared, want 1", tc.text, r.Texts)
		}
	}
}

func TestStatus(t *testing.T) {
	for _, s := range []string{"generated", "todo", "drafted by a model, 2026-10-22", "verified by A. Routt, 2026-11-02", "disputed by A. Routt, 2026-11-02: the book has a modal conclusion"} {
		if _, err := ParseStatus(s); err != nil {
			t.Errorf("%q: %v", s, err)
		}
	}
	for _, s := range []string{"", "done", "verified", "verified by A. Routt", "verified by A. Routt, 2 Nov 2026", "disputed by X, 2026-11-02"} {
		if _, err := ParseStatus(s); err == nil {
			t.Errorf("%q: accepted", s)
		}
	}
}

func TestEnglish(t *testing.T) {
	lang := map[string]string{"expert/2": "%s is an expert in %s.", "asserts/2": "%s asserts that %s is true.", "based_on_evidence/1": "The assertion %s is based on evidence.", "more_reliable/0": "Another sign is more reliable."}
	for f, want := range map[string]string{
		"expert(W,D)":                     "W is an expert in D.",
		"¬expert(W,D)":                    "It is not the case that W is an expert in D.",
		"based_on_evidence(asserts(W,S))": "The assertion “W asserts that S is true” is based on evidence.",
		"S":                               "S.",
		"¬S":                              "It is not the case that S.",
		"more_reliable":                   "Another sign is more reliable.",
		"unknown(X, a)":                   "Unknown(X, a).",
	} {
		if got := English(f, lang); got != want {
			t.Errorf("%s: %q, want %q", f, got, want)
		}
	}
}

// The committed decision register lists every open catalogue decision, its
// references resolve, and docs/decisions.md is its current view.
func TestCommittedRegister(t *testing.T) {
	reg, err := LoadRegister(root)
	if err != nil {
		t.Fatal(err)
	}
	if reg == nil {
		t.Skip("no register")
	}
	entries, _, err := ReadDir(filepath.Join(root, filepath.FromSlash(Dir)))
	if err != nil {
		t.Fatal(err)
	}
	var all []*Entry
	for _, e := range entries {
		all = append(all, e)
	}
	for _, f := range CheckRegister(reg, all) {
		if f.Error {
			t.Error(f)
		}
	}
	view, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(RegisterView)))
	if err != nil || strings.ReplaceAll(string(view), "\r\n", "\n") != string(RegisterMarkdown(reg)) {
		t.Errorf("%s is not the register's current view: run wag-cli --catalogue-gen", RegisterView)
	}
}

// The register's check catches each kind of fault.
func TestRegisterFindsFaults(t *testing.T) {
	entry := &Entry{ID: "x", Discrepancies: []Discrepancy{
		{Kind: "other", About: "a", Decision: "No change. Open: whether to add an exception."},
		{Kind: "other", About: "b", Decision: "Open: whether to rename it."},
		{Kind: "other", About: "c", Decision: "Done: renamed."},
	}}
	ok := Decision{ID: "D001", Title: "t", Question: "q", Category: "scheme", Decider: "owner", Priority: 1, Status: "open",
		Items: []ItemRef{{"x", "other", "a"}, {"x", "other", "b"}}}
	cases := []struct {
		name  string
		edit  func(d *Decision)
		want  string
		error bool
	}{
		{"an open item not listed", func(d *Decision) { d.Items = d.Items[:1] }, "other / b leaves a decision open", true},
		{"a missing entry", func(d *Decision) { d.Items = append(d.Items, ItemRef{"y", "other", "a"}) }, `no catalogue entry "y"`, true},
		{"a missing discrepancy", func(d *Decision) { d.Items = append(d.Items, ItemRef{"x", "other", "z"}) }, "has no discrepancy other / z", true},
		{"a bad id", func(d *Decision) { d.ID = "7" }, "write D and three digits", true},
		{"a bad category", func(d *Decision) { d.Category = "misc" }, `category "misc"`, true},
		{"a bad block", func(d *Decision) { d.Blocks = []string{"everything"} }, `blocks "everything"`, true},
		{"a bad priority", func(d *Decision) { d.Priority = 4 }, "priority 4", true},
		{"decided without an outcome", func(d *Decision) { d.Status = "decided" }, "decided needs", true},
		{"nowhere raised", func(d *Decision) { d.Items = nil }, "say where it is raised", true},
		{"an item no longer open", func(d *Decision) { d.Items = append(d.Items, ItemRef{"x", "other", "c"}) }, "no longer says Open", false},
	}
	if fs := CheckRegister(&Register{Decisions: []Decision{ok}}, []*Entry{entry}); len(fs) != 0 {
		t.Fatalf("a sound register has findings: %v", fs)
	}
	for _, c := range cases {
		d := ok
		d.Items = append([]ItemRef(nil), ok.Items...)
		c.edit(&d)
		var got []string
		found := false
		for _, f := range CheckRegister(&Register{Decisions: []Decision{d}}, []*Entry{entry}) {
			got = append(got, f.String())
			if strings.Contains(f.Msg, c.want) && f.Error == c.error {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: want a finding with %q (error %v), got %v", c.name, c.want, c.error, got)
		}
	}
	if !IsOpen("Open: x") || !IsOpen("No change. Open, for the owner: y") || IsOpen("Done: the open question is closed.") || IsOpen("Reopen it") {
		t.Error("IsOpen does not tell an open decision")
	}
}

// A formula reads with its variables bound to phrases, and a reading that
// opens with a quotation is capitalized inside it.
func TestEnglishWith(t *testing.T) {
	lang := map[string]string{"in_case/2": "%s is true in case %s.", "similar_to/2": "%s is similar to %s."}
	bind := map[string]string{"Ps": "the tax cut smoking", "Cs": "the tobacco tax", "Ct": "a meat tax", "Pt": "a meat tax cuts meat eating"}
	for _, c := range []struct{ formula, want string }{
		{"in_case(Ps,Cs)", "“The tax cut smoking” is true in case “the tobacco tax”."},
		{"similar_to(Cs,Ct)", "“The tobacco tax” is similar to “a meat tax”."},
		{"Pt", "A meat tax cuts meat eating."},
		{"in_case(Ps,X)", "“The tax cut smoking” is true in case X."},
	} {
		if got := EnglishWith(c.formula, lang, bind); got != c.want {
			t.Errorf("%s: got %q, want %q", c.formula, got, c.want)
		}
	}
}

// Applying a proposal to walton.yml changes only its schemes' blocks and its
// templates; every other byte stays (a YAML round trip would turn a
// variable Y into a boolean).
func TestProposalWalton(t *testing.T) {
	raw := []byte("meta:\n  title: t\nlanguage:\n  a/1: \"%s is a.\"\n  b/1: \"%s is b.\"\n\nargument_schemes:\n  - id: s1\n    variables: [X,Y]\n    premises:\n      - a(X)\n    conclusions:\n      - b(Y)\n\n  - id: s2\n    variables: [Y]\n    premises:\n      - a(Y)\n    conclusions:\n      - b(Y)\n")
	p := &Proposal{
		Schemes:  []Scheme{{ID: "s1", Variables: []string{"Z"}, Premises: []string{"a(Z)"}, Conclusions: []string{"c(Z)"}}, {ID: "s3", Variables: []string{"Y"}, Premises: []string{"a(Y)"}, Conclusions: []string{"Y"}}},
		Language: map[string]string{"c/1": "%s is c, in a long template that a YAML marshaller would wrap onto a second line somewhere."},
		Candidates: &Candidates{Schemes: []Scheme{{ID: "s4", Variables: []string{"Y"}, Premises: []string{"d(Y)"}, Conclusions: []string{"Y"}}},
			Language: map[string]string{"d/1": "%s is d."}},
	}
	out, err := ProposalWalton(raw, p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{"  b/1: \"%s is b.\"\n  c/1: \"%s is c, in a long", "  - id: s2\n    variables: [Y]\n", "  - id: s3\n", "    - c(Z)"} {
		if !strings.Contains(s, want) {
			t.Errorf("the result lacks %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "a(X)") || strings.Contains(s, "true") {
		t.Errorf("the old s1 block, or a boolean, is left:\n%s", s)
	}
	if strings.Contains(s, "s4") || strings.Contains(s, "d/1") {
		t.Errorf("a candidate was applied:\n%s", s)
	}
	var w struct {
		Schemes []Scheme `yaml:"argument_schemes"`
	}
	if err := yaml.UnmarshalStrict(out, &struct {
		Meta     map[string]string `yaml:"meta"`
		Language map[string]string `yaml:"language"`
		Schemes  *[]Scheme         `yaml:"argument_schemes"`
	}{Schemes: &w.Schemes}); err != nil || len(w.Schemes) != 3 || w.Schemes[1].Variables[0] != "Y" {
		t.Errorf("the result does not read back as three schemes (%v): %+v", err, w.Schemes)
	}
}

// The proposal check catches each kind of fault, and asks for the minimum
// set of examples.
func TestProposalFindsFaults(t *testing.T) {
	s := sources(t)
	base := func() *Proposal {
		return &Proposal{Decisions: []string{"D999"}, Title: "t",
			Schemes: []Scheme{{ID: "x_new", Variables: []string{"P", "C"}, Premises: []string{"in_case(P,C)", "similar_case(C)"},
				Assumptions: []string{"in_case(C,P)"}, Conclusions: []string{"P"}}},
			Questions: map[string][]PQuery{"x_new": {{ID: "walton/in_case", Asks: "q", Target: "in_case(C,P)"}}},
			Examples: []ProposalExample{
				{Scheme: "x_new", Title: "a", Text: "t", Use: "argument", Bindings: map[string]string{"P": "p", "C": "c"},
					Attacks: []Attack{{Question: "walton/in_case", Text: "x"}, {Question: Challenge, Target: "similar_case(C)", Text: "y"}}},
				{Scheme: "x_new", Title: "b", Text: "t", Use: "argument", Bindings: map[string]string{"P": "p", "C": "c"}},
				{Scheme: "x_new", Title: "c", Text: "t", Use: "answer", Answers: "CQ:x", Bindings: map[string]string{"P": "p", "C": "c"},
					AnswerTo: &AnswerTo{Scheme: "x_new", Question: "walton/in_case", Polarity: "attack", Bindings: map[string]string{"C": "c", "P": "p"}}},
				{Scheme: "analogy", Title: "d", Text: "t", Use: "near-miss", Of: "x_new", Bindings: bindAll(s, "analogy")},
				{Scheme: "x_cand", Title: "e", Text: "t", Use: "argument", Bindings: map[string]string{"P": "p"}},
			},
			Candidates: &Candidates{Schemes: []Scheme{{ID: "x_cand", Variables: []string{"P"}, Premises: []string{"x_odd(P)"}, Conclusions: []string{"P"}}},
				Language: map[string]string{"x_odd/1": "%s is odd."}},
		}
	}
	if fs := CheckProposal(base(), s); len(fs) != 0 {
		t.Fatalf("a sound proposal has findings: %v", fs)
	}
	cases := []struct {
		name  string
		edit  func(p *Proposal)
		want  string
		error bool
	}{
		{"an unbound variable", func(p *Proposal) { delete(p.Examples[0].Bindings, "C") }, "variable C is not bound", true},
		{"a stray binding", func(p *Proposal) { p.Examples[0].Bindings["Q"] = "q" }, "Q is not a variable", true},
		{"no template", func(p *Proposal) { p.Schemes[0].Premises = []string{"nowhere(P,C)"}; p.Questions = nil }, "no template for nowhere/2", true},
		{"an undeclared variable", func(p *Proposal) { p.Schemes[0].Variables = []string{"P"} }, "variable C is used but not declared", true},
		{"a variable in no premise", func(p *Proposal) {
			p.Schemes[0].Variables = append(p.Schemes[0].Variables, "D")
			p.Schemes[0].Conclusions = []string{"in_case(P,D)"}
		}, "variable D occurs in no premise", true},
		{"a question off the scheme", func(p *Proposal) { p.Questions["x_new"][0].Target = "in_case(P,P)" }, "not a formula of the scheme", true},
		{"an unknown question", func(p *Proposal) { p.Examples[0].Attacks[0].Question = "7/CQ9" }, "7/CQ9 is not a question", true},
		{"an answer without its question", func(p *Proposal) { p.Examples[2].Answers = "" }, "says which question", true},
		{"one argument only", func(p *Proposal) { p.Examples = append(p.Examples[:1], p.Examples[2:]...) }, "the minimum is two", false},
		{"no answer", func(p *Proposal) { p.Examples = append(p.Examples[:2], p.Examples[3]) }, "no example shows it", false},
		{"no near miss", func(p *Proposal) { p.Examples = p.Examples[:3] }, "no near miss", false},
		{"a question never pressed", func(p *Proposal) { p.Examples[0].Attacks = nil }, "pressed in no example", false},
		{"a walton/ question on a premise", func(p *Proposal) { p.Questions["x_new"][0].Target = "in_case(P,C)" },
			"a walton/ question targets an assumption or exception", true},
		{"a challenge of a conclusion", func(p *Proposal) { p.Examples[0].Attacks[1].Target = "P" }, "a challenge targets a premise", true},
		{"a question with a target", func(p *Proposal) { p.Examples[0].Attacks[0].Target = "in_case(P,C)" }, "only a challenge", true},
		{"a candidate that is in walton.yml", func(p *Proposal) {
			p.Candidates.Schemes[0].ID = "analogy"
			p.Examples[4].Scheme = "analogy"
			p.Examples[4].Bindings = bindAll(s, "analogy")
		}, "a candidate is a new scheme", true},
		{"a candidate's template in walton.yml", func(p *Proposal) { p.Candidates.Language["in_case/2"] = "x" }, "already in walton.yml", true},
		{"a candidate without an example", func(p *Proposal) { p.Examples = p.Examples[:4] }, "candidate x_cand: no example", false},
		{"questions for an unknown scheme", func(p *Proposal) { p.Questions["x_typo"] = p.Questions["x_new"] }, "x_typo, which is not a proposed scheme", true},
		{"candidate questions for a proposed scheme", func(p *Proposal) {
			p.Candidates.Questions = map[string][]PQuery{"x_new": p.Questions["x_new"]}
		}, "x_new, which is not a candidate", true},
		{"a question given twice", func(p *Proposal) { p.Questions["x_new"] = append(p.Questions["x_new"], p.Questions["x_new"][0]) }, "given twice", true},
		{"a label the book does not print", func(p *Proposal) {
			p.Questions["x_new"][0].ID = "7/CQ9"
			p.Examples[0].Attacks[0].Question = "7/CQ9"
		}, "prints no CQ9 under 7", true},
		{"a challenge of a formula a question targets", func(p *Proposal) { p.Examples[0].Attacks[1].Target = "in_case(C,P)" }, "targeted by question walton/in_case", true},
		// The catalogue must be able to carry a proposed scheme's questions.
		{"a book label on a scheme no entry links", func(p *Proposal) {
			p.Questions["x_new"][0].ID = "7/CQ2"
			p.Examples[0].Attacks[0].Question = "7/CQ2"
		}, "no compendium entry is linked to x_new", true},
		{"an id the catalogue cannot carry", func(p *Proposal) {
			p.Questions["x_new"][0].ID = "C999/Q1"
			p.Examples[0].Attacks[0].Question = "C999/Q1"
		}, "neither a book question", true},
		// Ids, and the schemes examples name, are unique and resolve (the
		// review of 24 November 2026).
		{"a scheme proposed twice", func(p *Proposal) { p.Schemes = append(p.Schemes, p.Schemes[0]) }, "x_new is proposed twice", true},
		{"a candidate sketched twice", func(p *Proposal) { p.Candidates.Schemes = append(p.Candidates.Schemes, p.Candidates.Schemes[0]) },
			"candidate x_cand is sketched twice", true},
		{"a near miss of nothing", func(p *Proposal) { p.Examples[3].Of = "x_typo" }, "near_miss_of names no scheme \"x_typo\"", true},
		{"a near miss of itself", func(p *Proposal) { p.Examples[3].Of = "analogy" }, "a near miss of analogy is coded with analogy itself", true},
		{"a near_miss_of on an argument", func(p *Proposal) { p.Examples[1].Of = "analogy" }, "only a near miss names near_miss_of", true},
		{"an answer that names nothing it answers", func(p *Proposal) { p.Examples[2].AnswerTo = nil }, "an answer names what it answers (answer_to", true},
		{"answer_to on an argument", func(p *Proposal) { p.Examples[1].AnswerTo = p.Examples[2].AnswerTo }, "only an answer names answer_to", true},
	}
	// A book label must belong to an entry linked to the scheme.
	if w, ok := findWalton(s, "analogy"); ok {
		p := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w},
			Questions: map[string][]PQuery{"analogy": {{ID: "33/CQ1", Asks: "q", Target: w.Premises[0]}}}}
		found := false
		for _, f := range CheckProposal(p, s) {
			found = found || f.Error && strings.Contains(f.Msg, "33 is not a compendium entry of analogy")
		}
		if !found {
			t.Error("a label of an entry not linked to the scheme is accepted")
		}
	}
	// A challenge of an assumption is allowed: the game presumes it as a
	// premise.
	p := base()
	p.Schemes[0].Assumptions = append(p.Schemes[0].Assumptions, "similar_case(P)")
	p.Examples[0].Attacks[1].Target = "similar_case(P)"
	for _, f := range CheckProposal(p, s) {
		if f.Error {
			t.Errorf("a challenge of an assumption: %s", f)
		}
	}
	// A candidate's sketch, which is never applied, may borrow another
	// entry's question; a proposed scheme may not (the cases above).
	p = base()
	p.Candidates.Questions = map[string][]PQuery{"x_cand": {{ID: "33/CQ1", Asks: "q", Target: "x_odd(P)"}}}
	p.Examples[4].Attacks = []Attack{{Question: "33/CQ1", Text: "z"}}
	for _, f := range CheckProposal(p, s) {
		if f.Error {
			t.Errorf("a sketch's borrowed question: %s", f)
		}
	}
	// A book label is any label the compendium prints for the scheme's own
	// entry, not only CQn: 54.1 prints a rebuttal factor.
	if w, ok := findWalton(s, "rule_physical_world_1"); ok {
		p := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w},
			Questions: map[string][]PQuery{w.ID: {{ID: "54.1/Rebuttal Factor 1", Asks: "q", Target: w.Premises[0]}}}}
		for _, f := range CheckProposal(p, s) {
			if f.Error {
				t.Errorf("a printed rebuttal factor of the scheme's own entry: %s", f)
			}
		}
	} else {
		t.Error("walton.yml has no rule_physical_world_1")
	}
	for _, c := range cases {
		p := base()
		c.edit(p)
		var got []string
		found := false
		for _, f := range CheckProposal(p, s) {
			got = append(got, f.String())
			if strings.Contains(f.Msg, c.want) && f.Error == c.error {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: want a finding with %q (error %v), got %v", c.name, c.want, c.error, got)
		}
	}
}

// An answer example names what it answers (answer_to): the check finds the
// formula the question targets, where it is in the argument answered, and
// so what the answer must conclude, and compares it with the example's
// conclusion, both instantiated (the review of 24 November 2026: a free-text
// answers field let an unrelated answer pass).
func TestProposalAnswerReferences(t *testing.T) {
	s := sources(t)
	base := func() *Proposal {
		goal := Scheme{ID: "xg_goal", Variables: []string{"A", "G"}, Premises: []string{"xg_aim(G)", "xg_means(A,G)"},
			Exceptions: []string{"xg_costly(A,G)"}, Conclusions: []string{"xg_do(A)"}}
		costly := Scheme{ID: "xg_costly_answer", Variables: []string{"A", "G", "H"}, Premises: []string{"xg_harm(A,H)", "xg_worse(H,G)"},
			Conclusions: []string{"xg_costly(A,G)"}}
		answer := func(x ProposalExample) ProposalExample {
			x.Scheme, x.Text, x.Use, x.Answers = "xg_costly_answer", "t", "answer", "walton/xg_costly of the goal argument"
			if x.Bindings == nil {
				x.Bindings = map[string]string{"A": "the plan", "G": "the aim", "H": "the harm"}
			}
			x.AnswerTo = &AnswerTo{Scheme: "xg_goal", Question: "walton/xg_costly", Polarity: "attack",
				Bindings: map[string]string{"A": "the plan", "G": "the aim"}}
			return x
		}
		bind := map[string]string{"A": "a", "G": "g"}
		bind3 := map[string]string{"A": "a", "G": "g", "H": "h"}
		return &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{goal, costly},
			Language: map[string]string{"xg_aim/1": "%s is an aim.", "xg_means/2": "%s serves %s.", "xg_costly/2": "%s costs more than %s is worth.",
				"xg_do/1": "%s should be done.", "xg_harm/2": "%s does the harm %s.", "xg_worse/2": "%s matters more than %s."},
			Questions: map[string][]PQuery{"xg_goal": {{ID: "walton/xg_costly", Asks: "q", Target: "xg_costly(A,G)"}}},
			Examples: []ProposalExample{
				{Scheme: "xg_goal", Title: "a1", Text: "t", Use: "argument", Bindings: bind, Attacks: []Attack{{Question: "walton/xg_costly", Text: "x"}}},
				{Scheme: "xg_goal", Title: "a2", Text: "t", Use: "argument", Bindings: bind},
				{Scheme: "analogy", Title: "n1", Text: "t", Use: "near-miss", Of: "xg_goal", Bindings: bindAll(s, "analogy")},
				{Scheme: "xg_costly_answer", Title: "b1", Text: "t", Use: "argument", Bindings: bind3},
				{Scheme: "xg_costly_answer", Title: "b2", Text: "t", Use: "argument", Bindings: bind3},
				{Scheme: "analogy", Title: "n2", Text: "t", Use: "near-miss", Of: "xg_costly_answer", Bindings: bindAll(s, "analogy")},
				answer(ProposalExample{Title: "the answer"}),
			},
		}
	}
	errs := func(p *Proposal) []string {
		var out []string
		for _, f := range CheckProposal(p, s) {
			if f.Error {
				out = append(out, f.Msg)
			}
		}
		return out
	}
	if e := errs(base()); len(e) != 0 {
		t.Fatalf("a sound proposal has errors: %v", e)
	}
	last := func(p *Proposal) *ProposalExample { return &p.Examples[len(p.Examples)-1] }
	cases := []struct {
		name string
		edit func(p *Proposal)
		want string
	}{
		{"another argument's binding", func(p *Proposal) { last(p).AnswerTo.Bindings["G"] = "another aim" }, `it concludes xg_costly("the plan","the aim")`},
		{"a defence that concludes the exception", func(p *Proposal) { last(p).AnswerTo.Polarity = "defence" }, `must conclude ¬xg_costly("the plan","the aim")`},
		{"a polarity that is neither", func(p *Proposal) { last(p).AnswerTo.Polarity = "for" }, "polarity is attack or defence"},
		{"an unknown scheme", func(p *Proposal) { last(p).AnswerTo.Scheme = "xg_typo" }, `answer_to: no scheme "xg_typo"`},
		{"a question the scheme does not have", func(p *Proposal) { last(p).AnswerTo.Question = "walton/xg_other" }, "walton/xg_other is not a question of xg_goal"},
		{"a target the question does not have", func(p *Proposal) { last(p).AnswerTo.Target = "xg_means(A,G)" }, "xg_means(A,G) is not what walton/xg_costly targets"},
		{"a variable of the target left unbound", func(p *Proposal) { delete(last(p).AnswerTo.Bindings, "G") }, "answer_to binds no G"},
		{"a binding of no variable of the scheme", func(p *Proposal) { last(p).AnswerTo.Bindings["H"] = "h" }, "H is not a variable of xg_goal"},
	}
	for _, c := range cases {
		p := base()
		c.edit(p)
		e := errs(p)
		found := false
		for _, m := range e {
			found = found || strings.Contains(m, c.want)
		}
		if !found {
			t.Errorf("%s: want an error with %q, got %v", c.name, c.want, e)
		}
	}
	// A challenge names its target, a premise of the argument; an attack on
	// it concludes its negation.
	p := base()
	x := last(p)
	x.Scheme, x.Form = "xg_goal", "negated"
	x.Bindings = map[string]string{"A": "the plan", "G": "the aim"}
	x.AnswerTo = &AnswerTo{Scheme: "xg_goal", Question: Challenge, Target: "xg_do(A)", Polarity: "attack", Bindings: map[string]string{"A": "the plan"}}
	if e := errs(p); !strings.Contains(strings.Join(e, "\n"), "a challenge targets a premise or assumption") {
		t.Errorf("a challenge of a conclusion: %v", e)
	}
	// A book label of a walton.yml scheme resolves through links.json: 33's
	// CQ3, printed for positive consequences (33.1), targets its exception
	// bad_outweighs(A,G); an answer that attacks there concludes it.
	if w, ok := findWalton(s, "positive_consequences"); ok {
		p := base()
		x := last(p)
		x.AnswerTo = &AnswerTo{Scheme: w.ID, Question: "33/CQ3", Polarity: "attack", Bindings: map[string]string{"A": "the plan", "G": "the aim"}}
		e := errs(p)
		want := `it concludes xg_costly("the plan","the aim")`
		if !strings.Contains(strings.Join(e, "\n"), `must conclude bad_outweighs("the plan","the aim")`) || !strings.Contains(strings.Join(e, "\n"), want) {
			t.Errorf("an answer to 33/CQ3 that concludes another formula: %v", e)
		}
		x.AnswerTo.Question = "33/CQ9"
		if e := errs(p); !strings.Contains(strings.Join(e, "\n"), "33/CQ9 is not a question of positive_consequences") {
			t.Errorf("an unprinted label: %v", e)
		}
	} else {
		t.Error("walton.yml has no positive_consequences")
	}
	// A proposed scheme's book question targets what links.json maps the
	// label to, or the proposal is warned that links.json must follow.
	if w, ok := findWalton(s, "positive_consequences"); ok {
		p := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w},
			Questions: map[string][]PQuery{w.ID: {{ID: "33/CQ3", Asks: "q", Target: w.Premises[0]}}}}
		warned := false
		for _, f := range CheckProposal(p, s) {
			warned = warned || !f.Error && strings.Contains(f.Msg, "links.json maps 33/CQ3 of positive_consequences to bad_outweighs(A,G)")
		}
		if !warned {
			t.Errorf("a book question retargeted without a warning: %v", CheckProposal(p, s))
		}
		p.Questions[w.ID][0].Target = "bad_outweighs(A,G)"
		for _, f := range CheckProposal(p, s) {
			if strings.Contains(f.Msg, "links.json maps") {
				t.Errorf("a book question on its mapped target: %s", f)
			}
		}
	}
	// Phrases are compared without regard to case or spacing.
	p = base()
	last(p).AnswerTo.Bindings["A"] = "The  Plan"
	if e := errs(p); len(e) != 0 {
		t.Errorf("a binding that differs only in case and spaces: %v", e)
	}
	// A game question of a walton.yml scheme resolves through its ask edge:
	// CQ:other_means targets practical_reasoning1's realize_more_effectively.
	p = base()
	p.Schemes = append(p.Schemes, Scheme{ID: "xg_better", Variables: []string{"A", "G"}, Premises: []string{"xg_means(A,G)"},
		Conclusions: []string{"realize_more_effectively(G,A)"}})
	p.Examples = append(p.Examples, ProposalExample{Scheme: "xg_better", Title: "g1", Text: "t", Use: "answer", Answers: "CQ:other_means",
		Bindings: map[string]string{"A": "the plan", "G": "the aim"},
		AnswerTo: &AnswerTo{Scheme: "practical_reasoning1", Question: "CQ:other_means", Polarity: "attack", Bindings: map[string]string{"G": "the aim", "A": "the plan"}}})
	if e := errs(p); len(e) != 0 {
		t.Errorf("an answer to a game question: %v", e)
	}
	last(p).AnswerTo.Question = "CQ:no_such_question"
	if e := errs(p); !strings.Contains(strings.Join(e, "\n"), "CQ:no_such_question is not a question of practical_reasoning1") {
		t.Errorf("an unknown game question: %v", e)
	}
	// A walton/ question of a walton.yml scheme outside the proposal names
	// its exception: positive_consequences' walton/bad_outweighs.
	p = base()
	last(p).AnswerTo = &AnswerTo{Scheme: "positive_consequences", Question: "walton/bad_outweighs", Polarity: "attack",
		Bindings: map[string]string{"A": "the plan", "G": "the aim"}}
	if e := errs(p); !strings.Contains(strings.Join(e, "\n"), `must conclude bad_outweighs("the plan","the aim")`) {
		t.Errorf("a walton/ question of a walton.yml scheme: %v", e)
	}
	// A question about the argument itself (evidence) has no formula to
	// compare.
	p = base()
	x = last(p)
	x.AnswerTo = &AnswerTo{Scheme: "positive_consequences", Question: "CQ:evidence", Polarity: "attack"}
	if e := errs(p); len(e) != 0 {
		t.Errorf("an answer to a question about the argument: %v", e)
	}
	// A challenge of a premise, answered by an attack, concludes the
	// premise's negation.
	p = base()
	p.Schemes = append(p.Schemes, Scheme{ID: "xg_aimless", Variables: []string{"G", "H"}, Premises: []string{"xg_worse(H,G)"},
		Conclusions: []string{"xg_aim(G)"}})
	p.Examples = append(p.Examples, ProposalExample{Scheme: "xg_aimless", Title: "c1", Text: "t", Use: "answer", Answers: "a challenge",
		Form: "negated", Bindings: map[string]string{"G": "the aim", "H": "h"},
		AnswerTo: &AnswerTo{Scheme: "xg_goal", Question: Challenge, Target: "xg_aim(G)", Polarity: "attack", Bindings: map[string]string{"G": "the aim"}}})
	if e := errs(p); len(e) != 0 {
		t.Errorf("an attack on a challenged premise: %v", e)
	}
	last(p).Form = ""
	if e := errs(p); !strings.Contains(strings.Join(e, "\n"), `must conclude ¬xg_aim("the aim")`) {
		t.Errorf("an attack on a premise that concludes the premise: %v", e)
	}
	// An applied proposal, written before answer_to, is warned, not refused.
	p = base()
	last(p).AnswerTo = nil
	p.Applied = "2026-10-22"
	if e := errs(p); len(e) != 0 {
		t.Errorf("an applied proposal's answer without answer_to: %v", e)
	}
}

func bindAll(s *Sources, id string) map[string]string {
	m := map[string]string{}
	if w, ok := findWalton(s, id); ok {
		for _, v := range w.Variables {
			m[v] = strings.ToLower(v)
		}
	}
	return m
}

// An applied proposal's view says so, since its "Now" columns then show the
// changed walton.yml.
func TestProposalApplied(t *testing.T) {
	s := sources(t)
	p := &Proposal{Decisions: []string{"D999"}, Title: "t", Applied: "2026-10-22, through edits.py"}
	if v := string(ProposalView(p, s, "x.yml")); !strings.Contains(v, "**Applied:** 2026-10-22, through edits.py.") {
		t.Errorf("the view does not say the proposal is applied:\n%s", v)
	}
}

// Every committed proposal checks without errors, and its view is current.
func TestCommittedProposals(t *testing.T) {
	s := sources(t)
	files, _ := filepath.Glob(filepath.Join(root, "docs", "proposals", "*.yml"))
	for _, f := range files {
		p, err := LoadProposal(f)
		if err != nil {
			t.Error(err)
			continue
		}
		for _, fd := range CheckProposal(p, s) {
			if fd.Error {
				t.Errorf("%s: %s", filepath.Base(f), fd)
			}
		}
		view, err := os.ReadFile(ProposalViewPath(f))
		if err != nil || strings.ReplaceAll(string(view), "\r\n", "\n") != string(ProposalView(p, s, filepath.Base(f))) {
			t.Errorf("%s is not the proposal's current view: run wag-cli --proposal %s", ProposalViewPath(filepath.Base(f)), filepath.Base(f))
		}
	}
}

// The candidates registry passes its check, and its view is current.
func TestCommittedCandidates(t *testing.T) {
	cr, err := LoadCandidates(root)
	if err != nil {
		t.Fatal(err)
	}
	if cr == nil {
		t.Skip("no candidates registry")
	}
	reg, err := LoadRegister(root)
	if err != nil {
		t.Fatal(err)
	}
	src := sources(t)
	for _, f := range CheckCandidates(cr, reg, src) {
		if f.Error {
			t.Error(f)
		}
	}
	view, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(CandidatesView)))
	if err != nil || strings.ReplaceAll(string(view), "\r\n", "\n") != string(CandidatesMarkdown(cr, reg, src)) {
		t.Errorf("%s is not the registry's current view: run wag-cli --catalogue-gen", CandidatesView)
	}
}

// The candidates check catches each kind of fault, and the stages follow
// the evidence and the minimum.
func TestCandidatesFindFaults(t *testing.T) {
	src := sources(t)
	reg := &Register{Decisions: []Decision{
		{ID: "D001", Status: "open"}, {ID: "D002", Status: "decided"}, {ID: "D047", Status: "open"},
	}}
	kialo := func(ref string) Sighting {
		return Sighting{Source: "kialo", Ref: ref, Fit: "clean", Note: "n", Found: "f", Verified: "v"}
	}
	sketch := &Sketch{
		Schemes: []Scheme{{ID: "sketch_x", Variables: []string{"A", "G"},
			Premises: []string{"brings_about(A,G)", "good(G)"}, Exceptions: []string{"bad_outweighs(A,G)", "tested_x(A)"},
			Conclusions: []string{"should_be_performed(A)"}}},
		Language:  map[string]string{"tested_x/1": "%s is tested."},
		Questions: map[string][]PQuery{"sketch_x": {{ID: "C001/Q1", Asks: "Tested?", Target: "tested_x(A)"}}},
	}
	use := func(title string) ProposalExample {
		return ProposalExample{Scheme: "sketch_x", Title: title, Text: "t", Invented: true, Use: "argument",
			Bindings: map[string]string{"A": "a", "G": "g"}}
	}
	near := ProposalExample{Scheme: "positive_consequences", Title: "near", Text: "t", Invented: true, Use: "near-miss",
		Of: "sketch_x", Bindings: map[string]string{"A": "a", "G": "g"}}
	ok := Candidate{ID: "C001", Name: "x", Kind: "exception", Group: "g", Decision: "D001", Found: []string{"f"},
		Relation:   &Relation{Kind: "addition", Base: []string{"positive_consequences"}, Why: "w"},
		Confidence: "medium", What: "w", Compares: "c", Concludes: "c",
		Closest:   []Closest{{Item: "positive_consequences", Why: "w"}},
		Sightings: []Sighting{kialo("1.1"), kialo("1.2"), kialo("2.1")},
		Sketch:    sketch, Examples: []ProposalExample{use("u1"), use("u2"), use("u3"), near}}
	base := func(cs ...Candidate) *CandidateRegister {
		return &CandidateRegister{Threshold: Threshold{Sightings: 3, Places: 2, Decision: "D047"},
			Groups: []CandidateGroup{{ID: "g", Name: "G", Families: []string{"F15"}, About: "a"}}, Candidates: cs}
	}
	if fs := CheckCandidates(base(ok), reg, src); len(fs) != 0 {
		t.Fatalf("a sound registry has findings: %v", fs)
	}
	if s := ok.Stage(base().Threshold); s != StageReady {
		t.Errorf("a complete candidate is %s, want ready", s)
	}
	cases := []struct {
		name  string
		edit  func(c *Candidate)
		want  string
		error bool
	}{
		{"a bad id", func(c *Candidate) { c.ID = "X1" }, "write C and three digits", true},
		{"a bad kind", func(c *Candidate) { c.Kind = "rule" }, `kind "rule"`, true},
		{"no group", func(c *Candidate) { c.Group = "h" }, `no group "h"`, true},
		{"an unknown decision", func(c *Candidate) { c.Decision = "D999" }, "no decision D999", true},
		{"a taken decision without an outcome", func(c *Candidate) { c.Decision = "D002" }, "no outcome is recorded", false},
		{"over the threshold without a decision", func(c *Candidate) { c.Decision = "" }, "meets the threshold", false},
		{"a bad outcome", func(c *Candidate) {
			c.Decision = "D002"
			c.Outcome = &CandidateOutcome{Date: "d", By: "b", Result: "keep", Recorded: "r"}
		}, `outcome "keep"`, true},
		{"a merge into nothing", func(c *Candidate) {
			c.Decision = "D002"
			c.Outcome = &CandidateOutcome{Date: "d", By: "b", Result: "merge", Recorded: "r"}
		}, "names the item", true},
		{"a bad Kialo ref", func(c *Candidate) { c.Sightings = append(c.Sightings, kialo("abc")) }, "discussion.claim", true},
		{"a sighting twice", func(c *Candidate) { c.Sightings = append(c.Sightings, kialo("1.1")) }, "listed twice", true},
		{"a bad fit", func(c *Candidate) { c.Sightings[0].Fit = "good" }, `fit "good"`, true},
		{"a sketch with no template", func(c *Candidate) {
			s := *sketch
			s.Language = nil
			c.Sketch = &s
		}, "no template for tested_x/1", true},
		{"an example not of the sketch", func(c *Candidate) {
			x := near
			x.Use = "argument"
			c.Examples = append(c.Examples, x)
		}, "coded with the candidate's sketch", true},
		{"a closest item without why", func(c *Candidate) { c.Closest = []Closest{{Item: "x"}} }, "why it does not cover", true},
		{"no relation", func(c *Candidate) { c.Relation = nil }, "say how it stands to what WAG has", true},
		{"a bad relation", func(c *Candidate) { c.Relation = &Relation{Kind: "cousin", Base: []string{"analogy"}, Why: "w"} }, `relation "cousin"`, true},
		{"a relation that does not fit the kind", func(c *Candidate) {
			c.Relation = &Relation{Kind: "variant", Base: []string{"positive_consequences"}, Why: "w"}
		}, "of kind exception cannot be a variant", true},
		{"a base that does not exist", func(c *Candidate) {
			c.Relation = &Relation{Kind: "addition", Base: []string{"nowhere"}, Why: "w"}
		}, "base nowhere is neither", true},
		{"a new pattern with a base", func(c *Candidate) {
			c.Kind = "scheme"
			c.Relation = &Relation{Kind: "new", Base: []string{"analogy"}, Why: "w"}
		}, "a new pattern has no base", true},
		{"a scheme one formula from walton.yml", func(c *Candidate) {
			c.Kind = "scheme"
			c.Relation = &Relation{Kind: "variant", Base: []string{"positive_consequences"}, Why: "w"}
		}, "within one formula of walton.yml's positive_consequences (adds exception tested_x/1)", true},
		{"an addition far from its base", func(c *Candidate) {
			c.Relation = &Relation{Kind: "addition", Base: []string{"negative_consequences"}, Why: "w"}
		}, "an addition is one formula from its base", false},
		{"a new pattern near a walton.yml scheme", func(c *Candidate) {
			s := *sketch
			s.Schemes = []Scheme{s.Schemes[0]}
			s.Schemes[0].Exceptions = []string{"tested_x(A)"}
			c.Sketch = &s
			c.Kind = "scheme"
			c.Relation = &Relation{Kind: "new", Why: "w"}
		}, "only 2 formulas from walton.yml's positive_consequences", false},
		{"a variant far from its base", func(c *Candidate) {
			c.Kind = "scheme"
			c.Relation = &Relation{Kind: "variant", Base: []string{"negative_consequences"}, Why: "w"}
		}, "a variant changes or adds a few formulas of its base, but its sketch is 7 from the nearest", false},
		{"a composition that shares nothing with a base", func(c *Candidate) {
			c.Kind = "scheme"
			c.Relation = &Relation{Kind: "composition", Base: []string{"positive_consequences", "analogy"}, Why: "w"}
		}, "shares no formula with analogy", false},
	}
	for _, k := range cases {
		c := ok
		c.Sightings = append([]Sighting(nil), ok.Sightings...)
		c.Examples = append([]ProposalExample(nil), ok.Examples...)
		k.edit(&c)
		var got []string
		found := false
		for _, f := range CheckCandidates(base(c), reg, src) {
			got = append(got, f.String())
			if strings.Contains(f.Msg, k.want) && f.Error == k.error {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: want a finding with %q (error %v), got %v", k.name, k.want, k.error, got)
		}
	}
	// Two candidates: a shared sighting is noted; a sketch id or a template
	// given twice is an error.
	other := ok
	other.ID = "C002"
	got := CheckCandidates(base(ok, other), reg, src)
	var msgs []string
	for _, f := range got {
		msgs = append(msgs, f.String())
	}
	all := strings.Join(msgs, "\n")
	for _, want := range []string{"1.1 is listed under C001 and C002", "sketch sketch_x is also C001's"} {
		if !strings.Contains(all, want) {
			t.Errorf("want %q among %v", want, msgs)
		}
	}
	// A listing that names the others (also) settles the note; also naming
	// a candidate that does not list the claim is an error.
	acked := ok
	acked.ID = "C002"
	acked.Sketch, acked.Examples = nil, nil
	acked.Kind = "link"
	acked.Relation = &Relation{Kind: "rule", Base: []string{"positive_consequences"}, Why: "w"}
	acked.Sightings = []Sighting{kialo("1.1"), kialo("3.1"), kialo("4.1")}
	acked.Sightings[0].Also = []string{"C001"}
	acked.Sightings[1].Also = []string{"C001"}
	msgs = nil
	for _, f := range CheckCandidates(base(ok, acked), reg, src) {
		msgs = append(msgs, f.String())
	}
	all = strings.Join(msgs, "\n")
	if strings.Contains(all, "1.1 is listed under") {
		t.Errorf("an acknowledged listing is still noted: %v", msgs)
	}
	if !strings.Contains(all, "sighting 3.1: also C001, which does not list it") {
		t.Errorf("want an error for also naming a candidate that does not list the claim, got %v", msgs)
	}
	// The stages: an unverified sighting does not count; an exemption does.
	c := ok
	c.Sightings = []Sighting{kialo("1.1"), kialo("2.1"), {Source: "kialo", Ref: "2.2", Fit: "clean", Note: "n", Found: "f"}}
	if s := c.Stage(base().Threshold); s != StageCollecting {
		t.Errorf("two verified sightings: %s, want collecting", s)
	}
	c.Exempt = "a gap shown by analysis"
	if s := c.Stage(base().Threshold); s != StageReady {
		t.Errorf("an exempt candidate: %s, want ready", s)
	}
	c = ok
	c.Examples = c.Examples[:2]
	if s := c.Stage(base().Threshold); s != StageDrafting {
		t.Errorf("no near miss and two examples: %s, want drafting", s)
	}
}
