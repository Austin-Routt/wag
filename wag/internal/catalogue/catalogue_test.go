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
			Schemes:   []Scheme{{ID: "x_new", Variables: []string{"P", "C"}, Premises: []string{"in_case(P,C)", "similar_case(C)"}, Conclusions: []string{"P"}}},
			Questions: map[string][]PQuery{"x_new": {{ID: "7/CQ2", Asks: "q", Target: "in_case(P,C)"}}},
			Examples: []ProposalExample{
				{Scheme: "x_new", Title: "a", Text: "t", Use: "argument", Bindings: map[string]string{"P": "p", "C": "c"},
					Attacks: []Attack{{Question: "7/CQ2", Text: "x"}, {Question: Challenge, Target: "similar_case(C)", Text: "y"}}},
				{Scheme: "x_new", Title: "b", Text: "t", Use: "argument", Bindings: map[string]string{"P": "p", "C": "c"}},
				{Scheme: "x_new", Title: "c", Text: "t", Use: "answer", Answers: "CQ:x", Bindings: map[string]string{"P": "p", "C": "c"}},
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
		{"a question off the scheme", func(p *Proposal) { p.Questions["x_new"][0].Target = "in_case(C,P)" }, "not a formula of the scheme", true},
		{"an unknown question", func(p *Proposal) { p.Examples[0].Attacks[0].Question = "7/CQ9" }, "7/CQ9 is not a question", true},
		{"an answer without its question", func(p *Proposal) { p.Examples[2].Answers = "" }, "says which question", true},
		{"one argument only", func(p *Proposal) { p.Examples = append(p.Examples[:1], p.Examples[2:]...) }, "the minimum is two", false},
		{"no answer", func(p *Proposal) { p.Examples = append(p.Examples[:2], p.Examples[3]) }, "no example shows it", false},
		{"no near miss", func(p *Proposal) { p.Examples = p.Examples[:3] }, "no near miss", false},
		{"a question never pressed", func(p *Proposal) { p.Examples[0].Attacks = nil }, "pressed in no example", false},
		{"a walton/ question on a premise", func(p *Proposal) {
			p.Questions["x_new"][0].ID = "walton/in_case"
			p.Examples[0].Attacks[0].Question = "walton/in_case"
		}, "a walton/ question targets an assumption or exception", true},
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
		{"a challenge of a premise a question targets", func(p *Proposal) { p.Examples[0].Attacks[1].Target = "in_case(P,C)" }, "targeted by question 7/CQ2", true},
		{"an answer from a scheme without a bare conclusion", func(p *Proposal) { p.Schemes[0].Conclusions = []string{"in_case(C,P)"} }, "cannot use it as an answer", false},
	}
	// A challenge of an assumption is allowed: the game presumes it as a
	// premise.
	p := base()
	p.Schemes[0].Assumptions = []string{"in_case(C,P)"}
	p.Examples[0].Attacks[1].Target = "in_case(C,P)"
	for _, f := range CheckProposal(p, s) {
		if f.Error {
			t.Errorf("a challenge of an assumption: %s", f)
		}
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

func bindAll(s *Sources, id string) map[string]string {
	m := map[string]string{}
	if w, ok := findWalton(s, id); ok {
		for _, v := range w.Variables {
			m[v] = strings.ToLower(v)
		}
	}
	return m
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
