package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/catalogue"
	"github.com/yourorg/wag/internal/server"
)

// catalogueDirs resolves the repository's root and the catalogue's
// directory: given, or found from the working directory.
func catalogueDirs(repo, dir string) (string, string, error) {
	if repo == "" {
		r, err := catalogue.FindRoot(".")
		if err != nil {
			return "", "", err
		}
		repo = r
	}
	if dir == "" {
		dir = filepath.Join(repo, filepath.FromSlash(catalogue.Dir))
	}
	return repo, dir, nil
}

// runCatalogueGen generates the catalogue's skeleton from the sources,
// keeping what people and models have written (wag/data/catalogue/README.md).
func runCatalogueGen(repo, dir, by string) int {
	repo, dir, err := catalogueDirs(repo, dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// The game's files first (step C5.1's switch, docs/c5-game.md): the
	// catalogue's game data is their source, so the entries are then made
	// from what it says. Only the repository's own catalogue writes them.
	f, err := catalogue.ReadFiles(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	own := filepath.Join(repo, filepath.FromSlash(catalogue.Dir))
	if _, err := os.Stat(filepath.Join(dir, catalogue.GameDir)); err != nil {
		fmt.Printf("note: %s has no %s/; its entries' game parts are made from the repository's game files\n", dir, catalogue.GameDir)
	} else {
		var g *catalogue.GameGen
		if f, g, err = catalogue.WithGame(f, dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if sameDir(dir, own) {
			written, err := catalogue.WriteGameFiles(filepath.Join(repo, filepath.FromSlash(catalogue.GameDataDir)), g)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
			fmt.Printf("%s: the game's files from the catalogue, %d of %d written (%s)\n", catalogue.GameDataDir, len(written), len(catalogue.GamePaths), fingerprints(g))
		} else {
			fmt.Printf("note: %s is not the repository's catalogue: its entries are made from the game's files it makes (%s), which are not written\n", dir, fingerprints(g))
		}
	}
	src, err := catalogue.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	g := &catalogue.Generator{Src: src, Today: time.Now().Format("2006-01-02"), By: by}
	r, err := g.Write(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("%s: %d entries created, %d updated, %d unchanged; %s and %s/ written\n", dir, r.Created, r.Updated, r.Unchanged, catalogue.IndexFile, catalogue.GameDir)
	for _, o := range r.Orphans {
		fmt.Printf("note: %s is not made from the sources any more; left in place (the check reports it)\n", o)
	}
	reg, err := catalogue.LoadRegister(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if reg != nil {
		if err := os.WriteFile(filepath.Join(repo, filepath.FromSlash(catalogue.RegisterView)), catalogue.RegisterMarkdown(reg), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("%s written\n", catalogue.RegisterView)
	}
	cands, err := catalogue.LoadCandidates(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if cands != nil {
		if err := os.WriteFile(filepath.Join(repo, filepath.FromSlash(catalogue.CandidatesView)), catalogue.CandidatesMarkdown(cands, reg, src), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("%s written\n", catalogue.CandidatesView)
	}
	return 0
}

// fingerprints names the generated files' data fingerprint (the data_sha256
// a game records) and the fingerprint of the catalogue's values they were
// made from (core.yml's header).
func fingerprints(g *catalogue.GameGen) string {
	d := server.Data{YAML: g.Files[catalogue.GameCorePath], Schemes: g.Files[catalogue.GameSchemesPath],
		CQs: g.Files[catalogue.GameCQsPath], Edges: g.Files[catalogue.GameEdgesPath]}
	return "data " + d.Hash() + ", catalogue values " + g.Inputs
}

// sameDir tells whether two paths name the same directory.
func sameDir(a, b string) bool {
	x, err1 := os.Stat(a)
	y, err2 := os.Stat(b)
	return err1 == nil && err2 == nil && os.SameFile(x, y)
}

// runGameGen generates the game's files from the catalogue (docs/c5-game.md,
// C5.1): the NASCQ files and core.yml, into out (default: the repository's
// wag/data).
func runGameGen(repo, dir, out string) int {
	repo, dir, err := catalogueDirs(repo, dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if out == "" {
		if !sameDir(dir, filepath.Join(repo, filepath.FromSlash(catalogue.Dir))) {
			fmt.Fprintf(os.Stderr, "%s is not the repository's catalogue: give --game-out, so that the repository's game files are left as they are\n", dir)
			return 1
		}
		out = filepath.Join(repo, filepath.FromSlash(catalogue.GameDataDir))
	}
	g, err := catalogue.GameFromCatalogue(repo, dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	written, err := catalogue.WriteGameFiles(out, g)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("%s: the game's files from the catalogue, %d of %d written (%s)\n", out, len(written), len(catalogue.GamePaths), fingerprints(g))
	return 0
}

// runCatalogueCheck checks the catalogue against the sources and, given
// the book's text, checks that no text of ours reproduces it. Status 1 on
// an error.
func runCatalogueCheck(repo, dir, books string) int {
	repo, dir, err := catalogueDirs(repo, dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	src, err := catalogue.Load(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	opt := catalogue.DefaultOptions()
	var paths []string
	for _, p := range strings.Split(books, ",") {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		// The transcription of the book's pp. 1-138 in the repository, if it
		// is still there; the compendium (ch. 9) is not.
		if p := filepath.Join(repo, "docs", "book-catalogue", "catalogue.json"); fileExists(p) {
			paths = []string{p}
		}
	}
	if len(paths) > 0 {
		c, err := catalogue.LoadCorpus(paths)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		opt.Book = c
		fmt.Printf("copy check against %d book sentences from %s\n", c.Sentences(), strings.Join(paths, ", "))
	}
	// The decision register (docs/process.md §5), when there is one.
	if opt.Register, err = catalogue.LoadRegister(repo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if opt.Register != nil {
		opt.RegisterView = filepath.Join(repo, filepath.FromSlash(catalogue.RegisterView))
	}
	// The candidates registry (docs/process.md §8), when there is one.
	if opt.Candidates, err = catalogue.LoadCandidates(repo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if opt.Candidates != nil {
		opt.CandidatesView = filepath.Join(repo, filepath.FromSlash(catalogue.CandidatesView))
	}
	r, err := catalogue.Check(dir, src, opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, f := range r.Findings {
		fmt.Println(f)
	}
	fmt.Println(r.Summary())
	if r.Copied {
		fmt.Printf("copy check: %d texts of ours compared with the book\n", r.Texts)
	}
	if n := r.Errors(); n > 0 {
		fmt.Printf("%d error(s)\n", n)
		return 1
	}
	fmt.Println("no errors")
	return 0
}

// runCatalogueShow prints a walton.yml scheme's formulas by role with their
// English readings, for whoever drafts or verifies its entry.
func runCatalogueShow(repo, id string) int {
	repo, _, err := catalogueDirs(repo, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	src, err := catalogue.Load(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, w := range src.Walton {
		if w.ID != id {
			continue
		}
		fmt.Printf("%s: %s\nvariables: %s\n", w.ID, w.Meta.Title, strings.Join(w.Variables, ", "))
		for _, r := range []struct {
			role string
			fs   []string
		}{{"premise", w.Premises}, {"assumption", w.Assumptions}, {"exception", w.Exceptions}, {"conclusion", w.Conclusions}} {
			for _, f := range r.fs {
				fmt.Printf("%-10s  %s\n            %s\n", r.role, f, catalogue.English(f, src.Language))
			}
		}
		return 0
	}
	fmt.Fprintf(os.Stderr, "no walton.yml scheme %q\n", id)
	return 1
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// runProposal checks a proposal for a decision that changes walton.yml,
// writes its view beside it, and, given a path, walton.yml with the
// proposal applied, for the cleaning's gate. Status 1 on an error.
func runProposal(repo, path, waltonOut string) int {
	repo, _, err := catalogueDirs(repo, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	src, err := catalogue.Load(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	p, err := catalogue.LoadProposal(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	errs := 0
	for _, f := range catalogue.CheckProposal(p, src) {
		fmt.Println(f)
		if f.Error {
			errs++
		}
	}
	view := catalogue.ProposalViewPath(path)
	if err := os.WriteFile(view, catalogue.ProposalView(p, src, filepath.Base(path)), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	candidates := 0
	if p.Candidates != nil {
		candidates = len(p.Candidates.Schemes)
	}
	fmt.Printf("%s written: %d proposed scheme(s), %d candidate(s), %d example(s)\n", view, len(p.Schemes), candidates, len(p.Examples))
	if waltonOut != "" {
		raw, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(catalogue.SourcePaths.Walton)))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		out, err := catalogue.ProposalWalton(raw, p)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if err := os.WriteFile(waltonOut, out, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("%s written: walton.yml with the proposal applied (run walton-check on it)\n", waltonOut)
	}
	if errs > 0 {
		fmt.Printf("%d error(s)\n", errs)
		return 1
	}
	fmt.Println("no errors")
	return 0
}
