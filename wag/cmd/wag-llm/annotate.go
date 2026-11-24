package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/llm"
)

// annotate has one or more models annotate a dataset's claims (the
// built-in Kialo batch, or a folder with items.json), each claim in its own
// call, the models in parallel; then, with --reviewer, a model reviews
// where they disagree. Everything goes in the run's folder (--out; its
// README.txt says what each file is). Run it again on the same folder to
// continue: annotated claims are skipped. The dashboard (wag-server,
// Annotate) makes the same folders.
func annotateCmd(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("annotate", flag.ExitOnError)
	specs, file := profilesFlags(fs)
	annotators := fs.String("annotators", "", "the profiles that annotate, comma-separated (default: every profile given, but the reviewer)")
	reviewer := fs.String("reviewer", "", "a profile that reviews the annotations where the annotators disagree")
	dataset := fs.String("dataset", "kialo-13292-batch2", "a built-in dataset's ID, or a folder with items.json (and optionally guidelines.md, gold.json, lexicon.json, case.json)")
	out := fs.String("out", "", "the run's folder (created; an existing run there continues)")
	limit := fs.Int("limit", 0, "annotate only the first N claims (0: all)")
	independent := fs.Bool("independent-first-pass", false, "withhold this annotator's earlier answers; preserve source context")
	all := fs.Bool("review-all", false, "with --reviewer: review every claim two annotators annotated, not only those where they disagree")
	limits := addBudgetFlags(fs)
	_ = fs.Parse(args)
	if *out == "" {
		return fmt.Errorf("annotate needs --out")
	}
	ps, err := llm.ProfilesFrom(*file, *specs)
	if err != nil {
		return err
	}
	byName := map[string]llm.Profile{}
	for _, p := range ps {
		byName[p.Name] = p
	}
	var ann []llm.Profile
	if *annotators != "" {
		for _, n := range strings.Split(*annotators, ",") {
			p, ok := byName[strings.TrimSpace(n)]
			if !ok {
				return fmt.Errorf("no profile %q", n)
			}
			ann = append(ann, p)
		}
	} else {
		for _, p := range ps {
			if p.Name != *reviewer {
				ann = append(ann, p)
			}
		}
	}
	var rev *llm.Profile
	if *reviewer != "" {
		p, ok := byName[*reviewer]
		if !ok {
			return fmt.Errorf("no profile %q for the reviewer", *reviewer)
		}
		rev = &p
	}
	budget, err := limits.budget()
	if err != nil {
		return err
	}
	started := time.Now()
	clients := map[string]*llm.Client{}
	for _, p := range append(append([]llm.Profile{}, ann...), ps...) {
		if _, done := clients[p.Name]; done {
			continue
		}
		if n := p.Notice(); n != "" {
			fmt.Fprintln(os.Stderr, "Note:", n)
		}
		c, err := llm.NewClient(p)
		if err != nil {
			return fmt.Errorf("profile %s: %v", p.Name, err)
		}
		c.Budget = budget
		clients[p.Name] = c
	}
	find := func(id string) (*annotate.Dataset, error) {
		if st, err := os.Stat(id); err == nil && st.IsDir() {
			return annotate.FromDir(id)
		}
		ds, err := annotate.Builtin()
		if err != nil {
			return nil, err
		}
		for _, d := range ds {
			if d.ID == id {
				return d, nil
			}
		}
		return nil, fmt.Errorf("no dataset %q: give a built-in ID or a folder", id)
	}
	var run *annotate.Run
	if _, err := os.Stat(filepath.Join(*out, "run.json")); err == nil {
		if run, err = annotate.Open(*out, func(id, dir string) (*annotate.Dataset, error) {
			if dir != "" {
				return annotate.FromDir(dir)
			}
			return find(id)
		}); err != nil {
			return err
		}
		modeChanged := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "independent-first-pass" && *independent != run.Settings.IndependentFirstPass {
				modeChanged = true
			}
		})
		if modeChanged {
			return fmt.Errorf("annotation mode differs from the saved run; create a new run")
		}
		if drift := run.InputDrift(); drift != "" {
			return fmt.Errorf("%s", drift)
		}
		for _, p := range ps {
			if err := run.CheckProfile(p); err != nil {
				return err
			}
		}
		fmt.Printf("Continuing the run in %s.\n", *out)
	} else {
		d, err := find(*dataset)
		if err != nil {
			return err
		}
		if len(ann) == 0 {
			return fmt.Errorf("annotate needs at least one annotator profile (--model)")
		}
		items := d.IDs()
		if *limit > 0 && *limit < len(items) {
			items = items[:*limit]
		}
		if run, err = annotate.New(*out, d, annotate.Settings{Annotators: ann, Reviewer: rev, Items: items, IndependentFirstPass: *independent}, nil); err != nil {
			return err
		}
	}
	// The run calls its annotators and reviewer, as profiled now: a continued
	// run, those saved in it, whatever this invocation's --annotators.
	var called []llm.Profile
	for _, s := range run.Settings.Annotators {
		if p, ok := byName[s.Name]; ok {
			called = append(called, p)
		}
	}
	if r := run.Settings.Reviewer; r != nil {
		if p, ok := byName[r.Name]; ok {
			called = append(called, p)
		}
	}
	limits.note(os.Stderr, budget, called)
	defer limits.finish(os.Stderr, run.Dir, "annotate", budget, started)
	for name, c := range clients {
		if err := run.Attach(name, c); err != nil {
			return err
		}
	}
	fmt.Printf("Annotating %d claims of %s with %d annotators; the files are in %s.\n", len(run.Settings.Items), run.Dataset.Name,
		len(run.Settings.Annotators), run.Dir)
	if err := run.Annotate(ctx, nil); err != nil {
		return unsavedNote(run, err)
	}
	// An annotator that stopped before the end (a quota used up, a refused
	// key, replies that keep failing the check) says why: it left claims it
	// did not try.
	for _, p := range run.Status().Progress {
		if p.Role == "annotator" && p.Done+p.Unsaved+p.Failed < p.Total && p.LastError != "" {
			fmt.Fprintf(os.Stderr, "%s: %s\n", p.Name, p.LastError)
		}
	}
	if run.Settings.Reviewer != nil {
		fmt.Printf("Reviewing with %s.\n", run.Settings.Reviewer.Name)
		if err := run.Review(ctx, nil, !*all); err != nil {
			return unsavedNote(run, err)
		}
	}
	rep, err := run.Report()
	if err != nil {
		return err
	}
	for _, a := range rep.Annotators {
		line := fmt.Sprintf("%s: %d annotated, %d failed, %d with no problem found", a.Name, a.Annotated, a.Failed, a.Valid)
		if a.VsGold != nil {
			line += fmt.Sprintf("; scheme agreement with the gold %.2f", a.VsGold.Fields["scheme"].Agreement)
		}
		fmt.Println(line)
	}
	for _, p := range rep.Pairs {
		fmt.Printf("%s and %s: scheme agreement %.2f over %d claims\n", p.A, p.B, p.Fields["scheme"].Agreement, p.Items)
	}
	b, _ := json.Marshal(rep.Verdicts)
	fmt.Printf("Report: %s (verdicts %s)\n", filepath.Join(run.Dir, "report.json"), b)
	return ctx.Err()
}

// unsavedNote tries once more to write what the run could not write, and
// says what this command's end loses if it still cannot.
func unsavedNote(run *annotate.Run, err error) error {
	if run.Status().WriteError == "" {
		return err
	}
	if run.Flush() == nil {
		fmt.Fprintln(os.Stderr, "Its records are written now: continue the run to carry on.")
		return err
	}
	if u := run.Status().Unsaved; len(u) > 0 {
		fmt.Fprintf(os.Stderr, "Not saved: %s. What did not reach calls.jsonl is lost when this command ends; "+
			"a reply that did is taken from it when the run is continued.\n", strings.Join(u, "; "))
	}
	return err
}
