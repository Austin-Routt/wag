package catalogue

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

// The fixes of the independent review of C5.1 (wag-workspace
// verify/c5/c51/review), each with the case that showed the fault.

// findings returns a report's errors on a file, or containing a text.
func findings(r *Report, file, text string) []string {
	var out []string
	for _, f := range r.Findings {
		if f.Error && (file == "" || f.File == file) && strings.Contains(f.Msg, text) {
			out = append(out, f.String())
		}
	}
	return out
}

// F2: the game part's status is the entry's once someone sets it; backs are
// copied from game/, and a question's game text is made from its paraphrase
// (step C5.2), and the check says so.
func TestGamePartStatusAndCopiedFields(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := generate(t, s)
	e := readEntry(t, dir, "sign")
	e.Game.Status = "verified by A Person, 2026-11-26"
	writeEntry(t, dir, e)
	src := regenerate(t, f, dir)
	if got := readEntry(t, dir, "sign").Game.Status; got != "verified by A Person, 2026-11-26" {
		t.Errorf("the game part's status after --catalogue-gen: %q", got)
	}
	r, err := Check(dir, src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if errs := errorsOf(r); len(errs) > 0 {
		t.Errorf("a verified game part: %v", errs)
	}
	e = readEntry(t, dir, "sign")
	e.Game.Status = "bogus"
	writeEntry(t, dir, e)
	if r, err = Check(dir, src, DefaultOptions()); err != nil {
		t.Fatal(err)
	}
	if len(findings(r, "sign.yml", "game: status")) == 0 {
		t.Errorf("a malformed game status: no error; errors %v", errorsOf(r))
	}

	dir = generate(t, s)
	e = readEntry(t, dir, "sign")
	e.Game.Backs = append(e.Game.Backs, "CQ:extra")
	for i := range e.CQs {
		if e.CQs[i].Game != nil {
			e.CQs[i].Game.Text += " Really?"
			break
		}
	}
	writeEntry(t, dir, e)
	if r, err = Check(dir, s, DefaultOptions()); err != nil {
		t.Fatal(err)
	}
	if len(findings(r, "sign.yml", "game/order.yml")) == 0 || len(findings(r, "sign.yml", "made from the question's paraphrase")) == 0 {
		t.Errorf("edited backs and game text: the check does not say where they are changed: %v", errorsOf(r))
	}
}

// F3: a comment under a marker in core.yml.in stays there: one run of
// --catalogue-gen is a fixpoint. Lines of a block under a marker are
// refused.
func TestTemplateCommentUnderAMarker(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := generate(t, s)
	path := filepath.Join(dir, GameDir, GameCoreTemplate)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const note = "    # a note under the marker\n"
	tpl := strings.Replace(string(b), coreMarker+"sign\n", coreMarker+"sign\n"+note, 1)
	if tpl == string(b) {
		t.Fatal("no marker for sign")
	}
	if err := os.WriteFile(path, []byte(tpl), 0o644); err != nil {
		t.Fatal(err)
	}
	src := regenerate(t, f, dir)
	if got, _ := os.ReadFile(path); !strings.Contains(string(got), coreMarker+"sign\n"+note) {
		t.Error("core.yml.in lost the comment")
	}
	if !strings.Contains(string(src.files.Core), note) {
		t.Error("core.yml lacks the comment")
	}
	r, err := Check(dir, src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if errs := errorsOf(r); len(errs) > 0 {
		t.Errorf("after one run: %v", errs)
	}
	es, d := freshGame(t)
	d.CoreTemplate = strings.Replace(d.CoreTemplate, coreMarker+"sign\n", coreMarker+"sign\n    premises: [a(X)]\n", 1)
	if _, err := GenerateGame(es, d, s.GameSources()); err == nil || !strings.Contains(err.Error(), "is followed by lines of a block") {
		t.Errorf("a block under a marker: %v", err)
	}
}

// F4: an override that is walton.yml's formulas, written in another order
// of keys, is a fixpoint after one run: the files are as without it.
func TestOverrideInAnotherKeyOrder(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := generate(t, s)
	e := readEntry(t, dir, "expert_opinion")
	w := s.GameSources().Formulas["expert_opinion"]
	list := func(xs []string) []interface{} {
		out := make([]interface{}, len(xs))
		for i, x := range xs {
			out[i] = x
		}
		return out
	}
	e.Game.Core = append(e.Game.Core,
		yaml.MapItem{Key: "premises", Value: list(w.Premises)},
		yaml.MapItem{Key: "variables", Value: list(w.Variables)},
		yaml.MapItem{Key: "exceptions", Value: list(w.Exceptions)},
		yaml.MapItem{Key: "assumptions", Value: list(w.Assumptions)},
		yaml.MapItem{Key: "conclusions", Value: list(w.Conclusions)})
	writeEntry(t, dir, e)
	src := regenerate(t, f, dir)
	if !bytes.Equal(src.files.Core, f.Core) {
		t.Error("core.yml differs from the committed one")
	}
	if c := readEntry(t, dir, "expert_opinion").Game.Core; value(c, "premises") != nil {
		t.Errorf("the override is kept: %v", c)
	}
	r, err := Check(dir, src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if errs := errorsOf(r); len(errs) > 0 {
		t.Errorf("after one run: %v", errs)
	}
}

// F10: the check reads CRLF line endings as LF, in the game's files and in
// game/.
func TestCheckReadsCRLF(t *testing.T) {
	s := sources(t)
	f, err := ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	crlf := func(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\n"), []byte("\r\n")) }
	f.Core, f.Schemes, f.CQs, f.Edges = crlf(f.Core), crlf(f.Schemes), crlf(f.CQs), crlf(f.Edges)
	src, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	dir := generate(t, s)
	for _, name := range []string{GameQuestionsFile, GameOrderFile, GameCoreTemplate} {
		p := filepath.Join(dir, GameDir, name)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, crlf(b), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r, err := Check(dir, src, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if errs := gameErrors(r); len(errs) > 0 {
		t.Errorf("CRLF files: %v", errs)
	}
}

// F11: one template error is one error, not one per scheme.
func TestOneTemplateErrorIsOne(t *testing.T) {
	s := sources(t)
	es, d := freshGame(t)
	d.CoreTemplate += coreMarker + "nothing\n"
	_, err := GenerateGame(es, d, s.GameSources())
	if err == nil {
		t.Fatal("no error")
	}
	if n := strings.Count(err.Error(), "has no marker"); n > 0 {
		t.Errorf("%d \"has no marker\" errors besides the template's:\n%v", n, err)
	}
}

// F12: a formula names only its scheme's variables; a backed question is
// asked.
func TestGameGenVariablesAndBackedQuestions(t *testing.T) {
	s := sources(t)
	es, d := freshGame(t)
	e := es["sign.yml"]
	for i, it := range e.Game.Core {
		if it.Key == "variables" {
			vs := it.Value.([]interface{})
			e.Game.Core[i].Value = vs[:len(vs)-1]
		}
	}
	if _, err := GenerateGame(es, d, s.GameSources()); err == nil || !strings.Contains(err.Error(), "which is not one of its variables") {
		t.Errorf("a variable the list lacks: %v", err)
	}
	es, d = freshGame(t)
	backed := d.Order.Backup[0].CQ
	var keep []GameEdge
	for _, a := range d.Order.Ask {
		if a.CQ != backed {
			keep = append(keep, a)
		}
	}
	d.Order.Ask = keep
	for _, e := range es {
		for i := range e.CQs {
			if q := e.CQs[i].Game; q != nil && q.ID == backed {
				e.CQs[i].Game = nil
			}
		}
	}
	if _, err := GenerateGame(es, d, s.GameSources()); err == nil || !strings.Contains(err.Error(), "backs a question no scheme asks") {
		t.Errorf("a backed question no scheme asks (%s): %v", backed, err)
	}
}

// Case XA3 (docs/c5-game.md, C5.3's switch): a backup edge's across mark
// names a decision the register has, and has decided.
func TestAcrossMarkNamesADecidedDecision(t *testing.T) {
	s := sources(t)
	reg, err := LoadRegister(root)
	if err != nil || reg == nil {
		t.Fatalf("the register: %v", err)
	}
	open := ""
	for _, d := range reg.Decisions {
		if d.Status != "decided" {
			open = d.ID
			break
		}
	}
	if open == "" {
		t.Skip("no open decision in the register")
	}
	const first = "backup_edges:\n- cq: CQ:side_effect\n  scheme: AS:sign\n"
	for _, c := range []struct{ id, want string }{
		{"D999", "which the register does not have"},
		{open, "not decided"},
		{"D009", ""},
	} {
		dir := generate(t, s)
		p := filepath.Join(dir, GameDir, GameOrderFile)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), first) {
			t.Fatalf("order.yml's backup edges do not start with %q", first)
		}
		if err := os.WriteFile(p, []byte(strings.Replace(string(b), first, first+"  across: "+c.id+"\n", 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		opt := DefaultOptions()
		opt.Register = reg
		r, err := Check(dir, s, opt)
		if err != nil {
			t.Fatal(err)
		}
		got := findings(r, GameDir+"/"+GameOrderFile, "across semantic types by "+c.id)
		if c.want == "" && len(got) != 0 || c.want != "" && (len(got) != 1 || !strings.Contains(got[0], c.want)) {
			t.Errorf("across %s: %v, want %q", c.id, got, c.want)
		}
	}
}
