package catalogue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
