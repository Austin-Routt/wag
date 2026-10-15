package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/kialo"
	"github.com/yourorg/wag/internal/nascq"
)

// runKialoCheck imports every discussion of dir (<dir>/<id>/graph.json), and
// prints for each: its claims; the claims whose text another claim also has;
// whether it imports; and what the import of before 2026-10-15 made of it,
// which merged claims of the same text into one statement and refused the
// game when that closed a loop. It returns the exit status: 1 if a
// discussion does not import.
func runKialoCheck(f dataFiles, dir string) int {
	paths, _ := filepath.Glob(filepath.Join(dir, "*", "graph.json"))
	if len(paths) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no <id>/graph.json in %s\n", dir)
		return 1
	}
	n, err := nascq.LoadFromBytes(f.schemes, f.cqs, f.edges)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading the NASCQ: %v\n", err)
		return 1
	}
	build := func(d *kialo.Discussion, keys string) (*kialo.Report, int, error) {
		rels, rep := d.Plan(0)
		rep.Keys = keys
		ad, err := carneades.NewAdapterFromYAML(f.yaml)
		if err != nil {
			return rep, 0, err
		}
		_, err = kialo.Build(d, rels, rep, kialo.Config(), ad, n)
		return rep, len(ad.Graph.Statements), err
	}
	fmt.Fprintf(say, "%-12s %7s %9s %9s  %-8s %s\n", "discussion", "claims", "relations", "same text", "import", "the import before 2026-10-15 (claims merged by text)")
	var total, shared, bad, refused int
	for _, p := range paths {
		name := filepath.Base(filepath.Dir(p))
		b, err := os.ReadFile(p)
		var d *kialo.Discussion
		if err == nil {
			d, err = kialo.Parse(b)
		}
		if err != nil {
			fmt.Fprintf(say, "%-12s %v\n", name, err)
			bad++
			continue
		}
		total++
		dup := sameText(d)
		if dup > 0 {
			shared++
		}
		rep, stmts, err := build(d, kialo.KeysByClaim)
		now := "ok"
		if err != nil {
			now = "FAILS: " + err.Error()
			bad++
		}
		old := "the same game"
		if _, oldStmts, err := build(d, kialo.KeysByText); err != nil {
			old = "refused: " + err.Error()
			refused++
		} else if oldStmts != stmts {
			old = fmt.Sprintf("%d statement(s) fewer: claims merged", stmts-oldStmts)
		}
		fmt.Fprintf(say, "%-12s %7d %9d %9d  %-8s %s\n", name, len(d.Claims), rep.Pros+rep.Cons, dup, now, old)
	}
	fmt.Fprintf(say, "\n%d discussions: %d with claims that share a text; %d refused by the import before 2026-10-15; %d that do not import now.\n",
		total, shared, refused, bad)
	if bad > 0 {
		return 1
	}
	return 0
}

// sameText counts the claims of a discussion whose text (compared as free
// text is: case, spaces and a final period aside) another claim also has.
func sameText(d *kialo.Discussion) int {
	by := map[string]int{}
	for _, c := range d.Claims {
		if t := carneades.NormalText(c.Text); strings.TrimSpace(t) != "" {
			by[t]++
		}
	}
	n := 0
	for _, k := range by {
		if k > 1 {
			n += k
		}
	}
	return n
}
