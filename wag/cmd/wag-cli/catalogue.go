package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/catalogue"
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
	src, err := catalogue.Load(repo)
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
	fmt.Printf("%s: %d entries created, %d updated, %d unchanged; %s written\n", dir, r.Created, r.Updated, r.Unchanged, catalogue.IndexFile)
	for _, o := range r.Orphans {
		fmt.Printf("note: %s is not made from the sources any more; left in place (the check reports it)\n", o)
	}
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
