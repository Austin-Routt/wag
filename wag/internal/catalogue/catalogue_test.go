package catalogue

import (
	"fmt"
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
		{"game text edited", "the game text is", func(e *Entry) { e.CQs[1].Game.Text = "Is it?" }},
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
		{"derived question with a bad id", "needs the id derived/<name>", func(e *Entry) {
			e.CQs = append(e.CQs, derivedCQ("own/1"))
		}},
		{"derived question without its page", "names its source, chapter, page and name there", func(e *Entry) {
			d := derivedCQ("derived/ch3-honesty")
			d.Derived.Page = ""
			e.CQs = append(e.CQs, d)
		}},
		{"derived question the game asks", "a derived question is neither the compendium's nor the game's", func(e *Entry) {
			d := derivedCQ("derived/ch3-honesty")
			d.Game = e.CQs[0].Game
			e.CQs = append(e.CQs, d)
		}},
		{"derived/ question without its source", "says where it comes from", func(e *Entry) {
			d := derivedCQ("derived/ch3-honesty")
			d.Derived = nil
			e.CQs = append(e.CQs, d)
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

// derivedCQ is a derived question (D007, D023) of expert opinion, well
// made but for what a test changes.
func derivedCQ(id string) CQ {
	return CQ{ID: id, Paraphrase: "Is W honest?", Status: "drafted by a model, 2026-10-22",
		Derived: &DerivedCQ{Source: "Walton, Reed and Macagno 2008", Chapter: "3", Page: "92", Name: "Honesty Question"},
		Targets: []Target{{Formula: "untrustworthy(W)", Locus: "exception"}}}
}

// A well-made derived question passes the check and survives regeneration.
func TestDerivedQuestion(t *testing.T) {
	s := sources(t)
	dir := generate(t, s)
	e := readEntry(t, dir, "expert_opinion")
	e.CQs = append(e.CQs, derivedCQ("derived/ch3-honesty"))
	writeEntry(t, dir, e)
	r, err := Check(dir, s, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, msg := range errorsOf(r) {
		if strings.Contains(msg, "derived") {
			t.Errorf("a well-made derived question: %s", msg)
		}
	}
	if _, err := (&Generator{Src: s, Today: "2026-10-22", By: "test"}).Write(dir); err != nil {
		t.Fatal(err)
	}
	if c := findCQ(readEntry(t, dir, "expert_opinion").CQs, "derived/ch3-honesty"); c == nil || c.Derived == nil || c.Derived.Page != "92" {
		t.Errorf("regeneration lost the derived question: %+v", c)
	}
}

// Regenerating keeps what people wrote, refreshes the sources, and adds
// what the sources add.
func TestMergeKeepsWork(t *testing.T) {
	s := sources(t)
	dir := generate(t, s)
	e := readEntry(t, dir, "expert_opinion")
	e.Summary = Text{Text: "Our summary.", Status: "drafted by a model, 2026-10-22"}
	para := e.CQs[0].Paraphrase // the game's text read back: the game's files stay as they are
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
	case m.CQs[0].Paraphrase != para || m.CQs[0].Status != "verified by A. Person, 2026-10-23":
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
	// A game question's text is its paraphrase (D009): a person who rewords
	// it changes the game's files, which the check asks to make again.
	m = readEntry(t, dir, "expert_opinion")
	m.CQs[0].Paraphrase = "How good an expert in D is W?"
	writeEntry(t, dir, m)
	r, _ = Check(dir, s, DefaultOptions())
	stale := strings.Join(errorsOf(r), "\n")
	for _, f := range []string{GameCorePath, GameCQsPath} {
		if !strings.Contains(stale, GameDataDir+"/"+f+": not what the catalogue makes of its game data") {
			t.Errorf("a reworded game question leaves %s as it was:\n%s", f, stale)
		}
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
	e.CQs[1].Game = g                                                           // a person moves it to 2/CQ2 ...
	e.CQs[1].Paraphrase = strings.NewReplacer("{", "", "}", "").Replace(g.Text) // with the question's text (D009)
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
		{"a variable declared twice", func(p *Proposal) { p.Schemes[0].Variables = append(p.Schemes[0].Variables, "P") }, "variable P is declared twice", true},
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
		{"an id the proposal cannot carry", func(p *Proposal) {
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
	p.Schemes = append(p.Schemes, Scheme{ID: "xg_better", Variables: []string{"A", "G", "P"}, Premises: []string{"xg_means(A,G)", "goal_of_agent(P,G)"},
		Conclusions: []string{"realize_more_effectively(P,G,A)"}})
	p.Examples = append(p.Examples, ProposalExample{Scheme: "xg_better", Title: "g1", Text: "t", Use: "answer", Answers: "CQ:other_means",
		Bindings: map[string]string{"A": "the plan", "G": "the aim", "P": "the council"},
		AnswerTo: &AnswerTo{Scheme: "practical_reasoning1", Question: "CQ:other_means", Polarity: "attack",
			Bindings: map[string]string{"G": "the aim", "A": "the plan", "P": "the council"}}})
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
	// A question about the argument itself (sign's CQ:other_event) has no
	// formula to compare.
	p = base()
	x = last(p)
	x.AnswerTo = &AnswerTo{Scheme: "sign", Question: "CQ:other_event", Polarity: "attack"}
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
	// Applied, its templates are walton.yml's (TestAppliedProposalTemplates):
	// the sources here have them.
	p = base()
	last(p).AnswerTo = nil
	p.Applied = "2026-10-22"
	applied := *s
	applied.Language = map[string]string{}
	for k, v := range s.Language {
		applied.Language[k] = v
	}
	for k, v := range p.Language {
		applied.Language[k] = v
	}
	var e []string
	for _, f := range CheckProposal(p, &applied) {
		if f.Error {
			e = append(e, f.Msg)
		}
	}
	if len(e) != 0 {
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

// An applied proposal's view marks each scheme that walton.yml has changed
// since (a later decision), so that its "Now" column is not read as the
// proposed one (the review of proposal D008-D074, version 2, E7).
func TestProposalAppliedChangedSince(t *testing.T) {
	s := sources(t)
	w, ok := findWalton(s, "positive_consequences")
	if !ok {
		t.Fatal("walton.yml has no positive_consequences")
	}
	const mark = "*Changed since this proposal was applied:*"
	p := &Proposal{Decisions: []string{"D999"}, Title: "t", Applied: "2026-10-23", Schemes: []Scheme{w}}
	if v := string(ProposalView(p, s, "x.yml")); strings.Contains(v, mark) {
		t.Errorf("a scheme as walton.yml has it is marked changed:\n%s", v)
	}
	old := w
	old.Exceptions = nil
	p.Schemes = []Scheme{old}
	if v := string(ProposalView(p, s, "x.yml")); !strings.Contains(v, mark) {
		t.Errorf("a scheme walton.yml has changed since is not marked:\n%s", v)
	}
	p.Applied = ""
	if v := string(ProposalView(p, s, "x.yml")); strings.Contains(v, mark) {
		t.Errorf("a proposal not yet applied is marked:\n%s", v)
	}
	// The check of version 3 of proposal D008-D074 (N2): a premise
	// changed since is marked too, and a scheme walton.yml no longer has is
	// marked removed, not shown as a new one.
	p.Applied = "2026-10-23"
	other := w
	other.Premises = append([]string(nil), w.Premises...)
	other.Premises[len(other.Premises)-1] = "bad(G)"
	p.Schemes = []Scheme{other}
	if v := string(ProposalView(p, s, "x.yml")); !strings.Contains(v, mark) {
		t.Errorf("a scheme whose premise walton.yml has changed since is not marked:\n%s", v)
	}
	gone := w
	gone.ID = "gone_since_applied"
	p.Schemes = []Scheme{gone}
	const removed = "*Removed since this proposal was applied:*"
	if v := string(ProposalView(p, s, "x.yml")); !strings.Contains(v, removed) || strings.Contains(v, "A new scheme.") {
		t.Errorf("a scheme walton.yml no longer has is not marked removed:\n%s", v)
	}
	p.Applied = ""
	if v := string(ProposalView(p, s, "x.yml")); strings.Contains(v, removed) || !strings.Contains(v, "A new scheme.") {
		t.Errorf("a new scheme of a proposal not yet applied is marked removed:\n%s", v)
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

// An applied proposal's templates are walton.yml's (cases PC1-PC3,
// docs/d086-first-gold.md §6; the independent review of D040-D071-D097's
// application, W6): a template changed after the application is an error
// of the applied proposal, not of one that proposes it.
func TestAppliedProposalTemplates(t *testing.T) {
	p, err := LoadProposal(filepath.Join(root, "docs", "proposals", "D040-D071-D097.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Applied == "" || p.Language["interference/2"] == "" {
		t.Fatal("D040-D071-D097 is not the applied proposal of interference/2 this test was written for")
	}
	const want = "template interference/2: the proposal is applied, but walton.yml's template is"
	errs := func(p *Proposal, s *Sources) []string {
		var out []string
		for _, f := range CheckProposal(p, s) {
			if f.Error && strings.Contains(f.Msg, "the proposal is applied, but walton.yml's template") {
				out = append(out, f.Msg)
			}
		}
		return out
	}
	// PC3: as committed, none.
	if got := errs(p, sources(t)); len(got) != 0 {
		t.Errorf("committed: %v", got)
	}
	// PC1: walton.yml's template changed after the application.
	s := sources(t)
	changed := *s
	changed.Language = map[string]string{}
	for k, v := range s.Language {
		changed.Language[k] = v
	}
	changed.Language["interference/2"] = "Something keeps %s from being brought about by %s."
	if got := errs(p, &changed); len(got) != 1 || !strings.HasPrefix(got[0], want) {
		t.Errorf("a changed template: %v", got)
	}
	// The template gone from walton.yml is the same error.
	delete(changed.Language, "interference/2")
	if got := errs(p, &changed); len(got) != 1 || !strings.HasPrefix(got[0], want) {
		t.Errorf("a template gone: %v", got)
	}
	// PC2: a proposal not applied proposes the template.
	q := *p
	q.Applied = ""
	changed.Language["interference/2"] = "Something keeps %s from being brought about by %s."
	if got := errs(&q, &changed); len(got) != 0 {
		t.Errorf("not applied: %v", got)
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
			Premises: []string{"brings_about(A,G)", "good(G)"}, Exceptions: []string{"bad_outweighs(A,G)", "too_slight(A,G)", "gain_discounted_by_bearer(A,G)", "tested_x(A)"},
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
			// two formulas from positive_consequences: it lacks bad_outweighs
			s.Schemes[0].Exceptions = []string{"too_slight(A,G)", "gain_discounted_by_bearer(A,G)", "tested_x(A)"}
			c.Sketch = &s
			c.Kind = "scheme"
			c.Relation = &Relation{Kind: "new", Why: "w"}
		}, "only 2 formulas from walton.yml's positive_consequences", false},
		{"a variant far from its base", func(c *Candidate) {
			c.Kind = "scheme"
			c.Relation = &Relation{Kind: "variant", Base: []string{"negative_consequences"}, Why: "w"}
		}, "a variant changes or adds a few formulas of its base, but its sketch is 9 from the nearest", false},
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

// A proposal may link a scheme it proposes to a compendium entry, as
// links.json will have it once the proposal is applied (D008's modal
// variants carry the questions printed under 33): the check reads links.json
// with the link in place, holds the proposal's questions to the maps it
// proposes, and checks the link itself.
// A proposed link of a renamed scheme replaces links.json's link of it, which
// names it by its upstream id (treat for threat): the view shows the entry with
// one link, not two, and an applied proposal finds its link there.
func TestProposalLinksOfARenamedScheme(t *testing.T) {
	s := sources(t)
	e, ok := s.link("35")
	if !ok || len(e.Links) != 1 || e.Links[0].ID != "treat" {
		t.Fatalf("35's links are not the ones this test was written for: %+v", e.Links)
	}
	sch, ok := findWalton(s, "threat")
	if !ok {
		t.Fatal("no threat in walton.yml")
	}
	link := withID(e.Links[0], "threat")
	p := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{sch},
		Links: []ProposedLink{{Entry: "35", WaltonLink: link}}}
	var errs []string
	got := proposedLinks(p, s, func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) })
	g, _ := got.link("35")
	if len(errs) > 0 || len(g.Links) != 1 || g.Links[0].ID != "threat" {
		t.Errorf("before it is applied: %v, the entry's links %+v", errs, g.Links)
	}
	p.Applied = "today"
	for _, f := range CheckProposal(p, s) {
		if strings.Contains(f.Msg, "links.json does not have its link") {
			t.Errorf("applied, links.json's link under the upstream id is not found: %s", f.Msg)
		}
	}
	p.Links[0].Reason += " Changed."
	found := false
	for _, f := range CheckProposal(p, s) {
		found = found || strings.Contains(f.Msg, "does not have its link of threat to 35 as proposed")
	}
	if !found {
		t.Error("applied, a link that differs from links.json's is not reported")
	}
}

func TestProposalLinks(t *testing.T) {
	s := sources(t)
	const sound = `decisions: [D999]
title: t
schemes:
  - id: x_mod
    meta: {title: x, source: s}
    variables: [A, G]
    premises: ["brings_about(A,G)", "good(G)"]
    assumptions: []
    exceptions: ["bad_outweighs(A,G)"]
    conclusions: ["should_be_performed(A)"]
  - id: x_plain
    meta: {title: y, source: s}
    variables: [A, G]
    premises: ["brings_about(A,G)", "good(G)"]
    assumptions: []
    exceptions: []
    conclusions: ["should_be_performed(A)"]
questions:
  x_mod:
    - {id: 33/CQ1, asks: q, target: "brings_about(A,G)"}
    - {id: 33/CQ2, asks: q, target: "brings_about(A,G)"}
    - {id: 33/CQ3, asks: q, target: "bad_outweighs(A,G)"}
links:
  - entry: "33.1"
    id: x_mod
    match: partial
    reason: the reason
    premise_map:
      - compendium_label: Premise
        walton: [{formula: "brings_about(A,G)", locus: premise}, {formula: "good(G)", locus: premise}]
    conclusion_map:
      walton: [{formula: "should_be_performed(A)", locus: conclusion}]
    cq_map:
      - cq: CQ1
        walton: [{formula: "brings_about(A,G)", locus: premise}]
      - cq: CQ2
        walton: [{formula: "brings_about(A,G)", locus: premise}]
      - cq: CQ3
        walton: [{formula: "bad_outweighs(A,G)", locus: exception}]
examples: []
`
	load := func(text string) *Proposal {
		t.Helper()
		f := filepath.Join(t.TempDir(), "p.yml")
		if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		p, err := LoadProposal(f)
		if err != nil {
			t.Fatalf("a proposal with links does not load: %v", err)
		}
		return p
	}
	p := load(sound)
	for _, f := range CheckProposal(p, s) {
		if f.Error || strings.Contains(f.Msg, "links.json maps") || strings.Contains(f.Msg, "is printed for") {
			t.Errorf("a sound link: %s", f)
		}
	}
	if v := string(ProposalView(p, s, "p.yml")); !strings.Contains(v, "## Links to the compendium, as proposed") ||
		!strings.Contains(v, "**x_mod** formalizes 33.1 (partial)") || !strings.Contains(v, "CQ3: `bad_outweighs(A,G)` (exception)") {
		t.Errorf("the view does not show the proposed link:\n%s", v)
	}
	block := sound[strings.Index(sound, "links:"):strings.Index(sound, "examples:")]
	cases := []struct {
		name, old, new, want string
		error                bool
	}{
		{"no link", block, "", "no compendium entry is linked to x_mod", true},
		{"a scheme the proposal does not propose", "    id: x_mod\n", "    id: analogy\n", "analogy is not a proposed scheme", true},
		{"an entry links.json lacks", `entry: "33.1"`, `entry: "99.9"`, "links.json has no compendium entry 99.9", true},
		{"a heading", `entry: "33.1"`, `entry: "33"`, "33 is a heading", true},
		{"a match that is neither", "match: partial", "match: close", `match is exact or partial, not "close"`, true},
		{"no reason", "reason: the reason", "reason: ''", "a link gives its reason", true},
		{"a label the book does not print", "cq: CQ3", "cq: CQ9", "the compendium prints no CQ9 under 33", true},
		{"a formula off the scheme", `{formula: "bad_outweighs(A,G)", locus: exception}`, `{formula: "good_outweighs(A,G)", locus: exception}`,
			"CQ3 maps to good_outweighs(A,G), not a formula of x_mod", true},
		{"a formula at the wrong locus", `{formula: "bad_outweighs(A,G)", locus: exception}`, `{formula: "bad_outweighs(A,G)", locus: premise}`,
			"CQ3 maps to bad_outweighs(A,G) as its premise, but it is x_mod's exception", true},
		{"a premise map off the scheme", `{formula: "good(G)", locus: premise}]`, `{formula: "bad(G)", locus: premise}]`,
			`premise "Premise" maps to bad(G), not a formula of x_mod`, true},
		{"a conclusion map off the scheme", `{formula: "should_be_performed(A)", locus: conclusion}`, `{formula: "¬should_be_performed(A)", locus: conclusion}`,
			"the conclusion maps to ¬should_be_performed(A), not a formula of x_mod", true},
		{"a map to nothing", "      - cq: CQ1\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}]\n",
			"      - cq: CQ1\n        walton: []\n", "CQ1 maps to no formula", true},
		{"a scheme linked twice to one entry", "examples: []", "", "linked twice", true},
		{"a question off the proposed map", `{id: 33/CQ1, asks: q, target: "brings_about(A,G)"}`, `{id: 33/CQ1, asks: q, target: "good(G)"}`,
			"links.json maps 33/CQ1 of x_mod to brings_about(A,G)", false},
		{"an applied proposal whose link links.json lacks", "examples: []", "applied: 2026-11-25\nexamples: []",
			"the proposal is applied, but links.json does not have its link of x_mod to 33.1", true},
		// The review of version 2 (CW1-CW3): a question no link maps, a
		// printed question left out, and maps that are incomplete.
		{"a book question the proposed link does not map", "      - cq: CQ2\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}]\n", "",
			"the proposal's link of x_mod to 33.1 maps no CQ2", true},
		{"a printed question the proposal does not give", "    - {id: 33/CQ2, asks: q, target: \"brings_about(A,G)\"}\n", "",
			"33/CQ2 is printed for 33.1", false},
		{"a link that maps no conclusion", "    conclusion_map:\n      walton: [{formula: \"should_be_performed(A)\", locus: conclusion}]\n", "",
			"the link maps no conclusion", true},
		{"a link that maps fewer premises than the entry prints", "    premise_map:\n      - compendium_label: Premise\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}, {formula: \"good(G)\", locus: premise}]\n", "",
			"33.1 prints 1 premise(s), and the link maps 0", true},
		{"a premise label the entry's links do not use", "compendium_label: Premise", "compendium_label: Major Premise",
			`33.1 prints no premise "Major Premise"`, true},
		// The check of version 3 (N1): a label mapped more often than the
		// entry's links map it, and a question mapped twice.
		{"a premise label mapped twice", "      - compendium_label: Premise\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}, {formula: \"good(G)\", locus: premise}]\n",
			"      - compendium_label: Premise\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}]\n      - compendium_label: Premise\n        walton: [{formula: \"good(G)\", locus: premise}]\n",
			`the link maps 33.1's premise "Premise" 2 times, and the entry's links map it 1`, true},
		{"a question mapped twice", "      - cq: CQ3\n", "      - cq: CQ1\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}]\n      - cq: CQ3\n",
			"the link maps CQ1 twice", true},
		// The review of D022-D093 v1: a question the link maps to no formula,
		// as links.json maps some (an undercutting question), says why, and
		// the proposal gives no question for it, which would have no target.
		{"a question mapped to no formula, and asked", "      - cq: CQ1\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}]\n",
			"      - cq: CQ1\n        note: an undercutter\n        walton: []\n",
			"the proposal's link of x_mod to 33.1 maps CQ1 to no formula, so the question has nothing to target", true},
	}
	for _, c := range cases {
		text := sound
		switch {
		case c.name == "a scheme linked twice to one entry":
			text = strings.Replace(sound, "examples: []", strings.TrimPrefix(block, "links:\n")+"examples: []", 1)
		case strings.Count(sound, c.old) != 1:
			t.Fatalf("%s: %q is not in the proposal exactly once", c.name, c.old)
		default:
			text = strings.Replace(sound, c.old, c.new, 1)
		}
		var got []string
		found := false
		for _, f := range CheckProposal(load(text), s) {
			got = append(got, f.String())
			found = found || strings.Contains(f.Msg, c.want) && f.Error == c.error
		}
		if !found {
			t.Errorf("%s: want a finding with %q (error %v), got %v", c.name, c.want, c.error, got)
		}
	}
	// Mapped to no formula with a note, and not asked: sound, and the view
	// says so.
	undercut := strings.Replace(strings.Replace(sound,
		"      - cq: CQ1\n        walton: [{formula: \"brings_about(A,G)\", locus: premise}]\n",
		"      - cq: CQ1\n        note: an undercutter\n        walton: []\n", 1),
		"    - {id: 33/CQ1, asks: q, target: \"brings_about(A,G)\"}\n", "", 1)
	if undercut == sound || strings.Contains(undercut, "33/CQ1") {
		t.Fatal("the variant with an undercutting CQ1 did not apply")
	}
	up := load(undercut)
	for _, f := range CheckProposal(up, s) {
		if f.Error || strings.Contains(f.Msg, "CQ1") {
			t.Errorf("a question mapped to no formula, with a note, and not asked: %s", f)
		}
	}
	if v := string(ProposalView(up, s, "p.yml")); !strings.Contains(v, "  - CQ1: no formula; an undercutter\n") {
		t.Errorf("the view does not say that CQ1 maps to no formula:\n%s", v)
	}
	// A note that says so itself is not doubled (the check of D022-D093 v2).
	said := load(strings.Replace(undercut, "note: an undercutter", "note: no formula, an undercutter", 1))
	if v := string(ProposalView(said, s, "p.yml")); !strings.Contains(v, "  - CQ1: no formula, an undercutter\n") ||
		strings.Contains(v, "no formula; no formula") {
		t.Errorf("the view doubles a note that says there is no formula:\n%s", v)
	}
	// A link that links.json already has, unchanged, is accepted once the
	// proposal is applied; another scheme's link to the entry is kept.
	if w, ok := findWalton(s, "positive_consequences"); ok {
		e, _ := s.link("33.1")
		var have WaltonLink
		for _, l := range e.Links {
			if l.ID == w.ID {
				have = l
			}
		}
		p := &Proposal{Decisions: []string{"D999"}, Title: "t", Applied: "2026-11-25", Schemes: []Scheme{w},
			Links: []ProposedLink{{Entry: "33.1", WaltonLink: have}}}
		for _, f := range CheckProposal(p, s) {
			if f.Error {
				t.Errorf("an applied proposal whose link links.json has: %s", f)
			}
		}
		p.Links[0].Reason = "another reason"
		found := false
		for _, f := range CheckProposal(p, s) {
			found = found || f.Error && strings.Contains(f.Msg, "links.json does not have its link of positive_consequences to 33.1 as proposed")
		}
		if !found {
			t.Error("an applied proposal whose link differs from links.json's is accepted")
		}
		// The check reads a copy: links.json's link is as it was.
		n := 0
		for _, l := range e.Links {
			if l.ID == w.ID {
				n++
			}
		}
		if after, _ := s.link("33.1"); len(after.Links) != len(e.Links) {
			t.Errorf("after the check, 33.1 has %d links, not %d", len(after.Links), len(e.Links))
		} else {
			for _, l := range after.Links {
				if l.ID == w.ID && linkKey(l) != linkKey(have) {
					t.Error("checking a proposed link changed links.json's link in the sources")
				}
			}
		}
		if n != 1 {
			t.Errorf("links.json links positive_consequences to 33.1 %d times", n)
		}
		// A question that links.json's own link does not map is a warning:
		// links.json must follow when the proposal is applied.
		gap := *s
		gap.Links = append([]LinkEntry(nil), s.Links...)
		for i, en := range gap.Links {
			if en.Number != "33.1" {
				continue
			}
			en.Links = append([]WaltonLink(nil), en.Links...)
			for j, l := range en.Links {
				if l.ID == w.ID {
					var rows []CQMapRow
					for _, r := range l.CQMap {
						if r.CQ != "CQ2" {
							rows = append(rows, r)
						}
					}
					l.CQMap = rows
					en.Links[j] = l
				}
			}
			gap.Links[i] = en
		}
		q := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w},
			Questions: map[string][]PQuery{w.ID: {{ID: "33/CQ2", Asks: "q", Target: "brings_about(A,G)"}}}}
		warned := false
		for _, f := range CheckProposal(q, &gap) {
			if f.Error && strings.Contains(f.Msg, "CQ2") {
				t.Errorf("links.json's own gap is an error: %s", f)
			}
			warned = warned || !f.Error && strings.Contains(f.Msg, "links.json maps no CQ2 of positive_consequences")
		}
		if !warned {
			t.Errorf("a question links.json does not map passes silently: %v", CheckProposal(q, &gap))
		}
		// A question links.json maps to no formula has nothing to target:
		// a warning that says so (the review of D022-D093 v1).
		none := *s
		none.Links = append([]LinkEntry(nil), s.Links...)
		emptied := false
		for i, en := range none.Links {
			if en.Number != "33.1" {
				continue
			}
			en.Links = append([]WaltonLink(nil), en.Links...)
			for j, l := range en.Links {
				if l.ID == w.ID {
					l.CQMap = append([]CQMapRow(nil), l.CQMap...)
					for k, r := range l.CQMap {
						if r.CQ == "CQ2" {
							l.CQMap[k] = CQMapRow{CQ: "CQ2", Note: "no formula"}
							emptied = true
						}
					}
					en.Links[j] = l
				}
			}
			none.Links[i] = en
		}
		if !emptied {
			t.Fatal("links.json maps no CQ2 of positive_consequences to empty")
		}
		warned = false
		for _, f := range CheckProposal(q, &none) {
			if f.Error && strings.Contains(f.Msg, "CQ2") {
				t.Errorf("links.json's question without a formula is an error: %s", f)
			}
			warned = warned || !f.Error && strings.Contains(f.Msg, "links.json maps CQ2 of positive_consequences to no formula, so the question has nothing to target")
		}
		if !warned {
			t.Errorf("a question links.json maps to no formula passes silently: %v", CheckProposal(q, &none))
		}
		// A relinked scheme keeps the book's premise labels, which its own
		// link to the entry gives when no other link does.
		relabel := have
		relabel.PremiseMap = []MapItem{{Label: "Major Premise", Walton: have.PremiseMap[0].Walton}}
		rp := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w},
			Links: []ProposedLink{{Entry: "33.1", WaltonLink: relabel}}}
		found = false
		for _, f := range CheckProposal(rp, s) {
			found = found || f.Error && strings.Contains(f.Msg, `33.1 prints no premise "Major Premise" (its links name Premise)`)
		}
		if !found {
			t.Errorf("a relinked scheme's premise label the book does not print passes: %v", CheckProposal(rp, s))
		}
		// A proposed link replaces the scheme's own: its questions are held
		// to the new map, not the old one as well.
		p = &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w},
			Questions: map[string][]PQuery{w.ID: {{ID: "33/CQ3", Asks: "q", Target: "bad_outweighs(A,G)"}}},
			Links: []ProposedLink{{Entry: "33.1", WaltonLink: WaltonLink{ID: w.ID, Match: "exact", Reason: "r",
				CQMap: []CQMapRow{{CQ: "CQ3", Walton: []FormulaLocus{{Formula: "good(G)", Locus: "premise"}}}}}}}}
		warned = false
		for _, f := range CheckProposal(p, s) {
			warned = warned || !f.Error && strings.Contains(f.Msg, "links.json maps 33/CQ3 of positive_consequences to good(G)")
		}
		if !warned {
			t.Errorf("a proposed link that remaps a question does not replace the scheme's own: %v", CheckProposal(p, s))
		}
	} else {
		t.Error("walton.yml has no positive_consequences")
	}
}

// The check of version 3 of proposal D008-D074 (N1): a link maps every
// premise the entry's links name, so a label mapped twice cannot stand for
// one never mapped.
func TestProposalLinksMapEveryPremise(t *testing.T) {
	s := sources(t)
	e, ok := s.link("33.3")
	if !ok || e.NPremises != 2 {
		t.Fatalf("33.3 is not in links.json with two premises")
	}
	mk := func(second string) *Proposal {
		x := Scheme{ID: "x_mod", Variables: []string{"A", "G"}, Premises: []string{"brings_about(A,G)", "good(G)"},
			Exceptions: []string{"bad_outweighs(A,G)"}, Conclusions: []string{"should_be_performed(A)"}}
		x.Meta.Title, x.Meta.Source = "x", "s"
		l := ProposedLink{Entry: "33.3"}
		l.ID, l.Match, l.Reason = "x_mod", "partial", "the reason"
		l.PremiseMap = []MapItem{
			{Label: "Premise 1", Walton: []FormulaLocus{{Formula: "brings_about(A,G)", Locus: "premise"}}},
			{Label: second, Walton: []FormulaLocus{{Formula: "good(G)", Locus: "premise"}}},
		}
		l.Conclusion = &MapItem{Walton: []FormulaLocus{{Formula: "should_be_performed(A)", Locus: "conclusion"}}}
		return &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{x}, Links: []ProposedLink{l}}
	}
	errs := func(p *Proposal) []string {
		var out []string
		for _, f := range CheckProposal(p, s) {
			if f.Error && strings.Contains(f.Msg, "the link of x_mod to 33.3") {
				out = append(out, f.Msg)
			}
		}
		return out
	}
	if got := errs(mk("Premise 2")); len(got) != 0 {
		t.Errorf("a link that maps both premises: %v", got)
	}
	// A label's count is the most any one of the entry's links maps it, not
	// their total: with a second link that maps each premise once, a third
	// that maps Premise 1 twice still maps it too often.
	p := mk("Premise 2")
	q := mk("Premise 1")
	q.Schemes[0].ID, q.Links[0].ID = "x_two", "x_two"
	p.Schemes, p.Links = append(p.Schemes, q.Schemes[0]), append(p.Links, q.Links[0])
	two := false
	for _, f := range CheckProposal(p, s) {
		two = two || f.Error && strings.Contains(f.Msg, `the link of x_two to 33.3: the link maps 33.3's premise "Premise 1" 2 times, and the entry's links map it 1`)
	}
	if !two {
		t.Error("a link that maps a premise more often than any of the entry's links is accepted when the entry has two links")
	}
	got := strings.Join(errs(mk("Premise 1")), "\n")
	for _, want := range []string{`33.3 prints premise "Premise 2", which the link does not map`,
		`the link maps 33.3's premise "Premise 1" 2 times, and the entry's links map it 1`} {
		if !strings.Contains(got, want) {
			t.Errorf("a link that maps Premise 1 twice and Premise 2 never: want %q, got:\n%s", want, got)
		}
	}
}

// The check of D022-D093 version 2: links.json names a renamed scheme by its
// upstream id (exeptional for exceptional, at 56), and every lookup of a
// scheme's link migrates it: the entry is the scheme's, its printed
// questions are asked of a proposal, a question the link maps is held to
// the formula it maps, and one it maps to no formula has nothing to target.
func TestProposalRenamedSchemeLinks(t *testing.T) {
	s := sources(t)
	if s.migrate("exeptional") != "exceptional" {
		t.Skip("id-migrations.json no longer renames exeptional")
	}
	w, ok := findWalton(s, "exceptional")
	if !ok {
		t.Fatal("walton.yml has no exceptional")
	}
	if es := linkedEntries(s, "exceptional"); !in(es, "56") {
		t.Errorf("linkedEntries(exceptional) = %v, want 56 (links.json names it exeptional)", es)
	}
	found := func(p *Proposal, want string, isErr bool) bool {
		for _, f := range CheckProposal(p, s) {
			if strings.Contains(f.Msg, want) && f.Error == isErr {
				return true
			}
		}
		return false
	}
	p := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w}}
	if !found(p, "56/CQ1 is printed for 56, and the catalogue will carry it for exceptional", false) {
		t.Errorf("a renamed scheme's printed question is not asked of it: %v", CheckProposal(p, s))
	}
	// links.json maps 56/CQ1 to case_exception(A): another target is held to it.
	p.Questions = map[string][]PQuery{"exceptional": {{ID: "56/CQ1", Asks: "q", Target: "rule_can_waived(X)"}}}
	if !found(p, "links.json maps 56/CQ1 of exceptional to case_exception(A)", false) {
		t.Errorf("a renamed scheme's question is not held to its link's formula: %v", CheckProposal(p, s))
	}
	p.Questions = map[string][]PQuery{"exceptional": {{ID: "56/CQ1", Asks: "q", Target: "case_exception(A)"}}}
	if found(p, "56/CQ1", true) || found(p, "links.json maps 56/CQ1", false) {
		t.Errorf("a renamed scheme's question at its link's formula is refused: %v", CheckProposal(p, s))
	}
	// 56/CQ2 maps to no formula: not borrowed, and nothing to target.
	p.Questions = map[string][]PQuery{"exceptional": {{ID: "56/CQ2", Asks: "q", Target: "rule_can_waived(X)"}}}
	if found(p, "borrowed", true) {
		t.Errorf("a renamed scheme's printed question is called borrowed: %v", CheckProposal(p, s))
	}
	if !found(p, "links.json maps CQ2 of exceptional to no formula, so the question has nothing to target", false) {
		t.Errorf("a renamed scheme's question without a formula passes silently: %v", CheckProposal(p, s))
	}
}

// The check of version 3 of proposal D008-D074 (W3): a printed question
// that the link maps to no formula (the catalogue's undercutting question)
// is not asked of the proposal, which could give it no target.
func TestProposalPrintedQuestionsWithoutFormula(t *testing.T) {
	s := sources(t)
	w, ok := findWalton(s, "memory")
	if !ok {
		t.Fatal("walton.yml has no memory")
	}
	e, ok := s.link("60")
	if !ok {
		t.Fatal("links.json has no 60")
	}
	empty := map[string]bool{}
	for _, l := range e.Links {
		if l.ID != "memory" {
			continue
		}
		for _, r := range l.CQMap {
			if len(r.Walton) == 0 {
				empty[r.CQ] = true
			}
		}
	}
	if len(empty) == 0 {
		t.Skip("links.json now maps every question of 60 to a formula")
	}
	p := &Proposal{Decisions: []string{"D999"}, Title: "t", Schemes: []Scheme{w}}
	for _, f := range CheckProposal(p, s) {
		for label := range empty {
			if strings.Contains(f.Msg, "60/"+label+" is printed for 60") {
				t.Errorf("a printed question that links.json maps to no formula is asked of the proposal: %s", f)
			}
		}
	}
}

// Cases PC4 and PC5 (docs/d086-first-gold.md §6; the independent review of
// D098 and R21, W2): an applied proposal's template that walton.yml no
// longer has is a later decision's when a later applied proposal set
// walton.yml's (no error; the view marks it), and an error otherwise.
func TestAppliedProposalTemplatesLater(t *testing.T) {
	p, err := LoadProposal(filepath.Join(root, "docs", "proposals", "D040-D071-D097.yml"))
	if err != nil {
		t.Fatal(err)
	}
	const decided = "Something interferes with %s bringing about %s."
	errs := func(s *Sources) []string {
		var out []string
		for _, f := range CheckProposal(p, s) {
			if f.Error && strings.Contains(f.Msg, "the proposal is applied, but walton.yml's template") {
				out = append(out, f.Msg)
			}
		}
		return out
	}
	with := func(later ...AppliedTemplate) *Sources {
		s := sources(t)
		changed := *s
		changed.Language = map[string]string{}
		for k, v := range s.Language {
			changed.Language[k] = v
		}
		changed.Language["interference/2"] = decided
		changed.Applied = map[string][]AppliedTemplate{}
		for k, v := range s.Applied {
			changed.Applied[k] = append([]AppliedTemplate(nil), v...)
		}
		changed.Applied["interference/2"] = append(changed.Applied["interference/2"], later...)
		return &changed
	}
	// PC4: a later applied proposal carries walton.yml's template.
	s := with(AppliedTemplate{Title: "The decided wording of interference", Date: "2026-12-01", Template: decided})
	if got := errs(s); len(got) != 0 {
		t.Errorf("PC4: %v", got)
	}
	view := string(ProposalView(p, s, "D040-D071-D097.yml"))
	if !strings.Contains(view, "*Changed since this proposal was applied:* `walton.yml`'s template for `interference/2` is now “"+decided+"”, from the later proposal “The decided wording of interference” (applied 2026-12-01).") {
		t.Errorf("PC4: the view does not mark the template")
	}
	// PC5: none does; one applied before this proposal does not count, nor
	// one whose template is another.
	for name, later := range map[string][]AppliedTemplate{
		"none":    nil,
		"earlier": {{Title: "An earlier proposal", Date: "2026-11-01", Template: decided}},
		"another": {{Title: "A later proposal", Date: "2026-12-01", Template: "Something else stops %s bringing about %s."}},
	} {
		if got := errs(with(later...)); len(got) != 1 {
			t.Errorf("PC5 %s: %v", name, got)
		}
	}
	// The committed proposals, read by Load: interference/2 is this
	// proposal's own, which is no later decision.
	if at := sources(t).Applied["interference/2"]; len(at) != 1 || at[0].Title != p.Title {
		t.Errorf("the applied templates of interference/2: %+v", at)
	}
}
