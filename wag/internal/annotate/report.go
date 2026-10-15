package annotate

import (
	"encoding/json"
	"path/filepath"
	"time"
)

// AnnotatorReport is one annotator's results.
type AnnotatorReport struct {
	Name      string         `json:"name"`
	Model     string         `json:"model"`
	Provider  string         `json:"provider"`
	Annotated int            `json:"annotated"`
	Failed    int            `json:"failed"`
	Valid     int            `json:"valid"`    // annotations the check finds nothing wrong with
	Problems  int            `json:"problems"` // problems found, in all
	Schemes   map[string]int `json:"schemes"`  // primary schemes chosen
	VsGold    *Pair          `json:"vs_gold,omitempty"`
}

// ReviewerReport is the reviewer's results.
type ReviewerReport struct {
	Name       string                    `json:"name"`
	Reviewed   int                       `json:"reviewed"`
	Errors     int                       `json:"errors"`
	NeedsHuman int                       `json:"needs_human"`
	Chosen     map[string]map[string]int `json:"chosen"` // field → annotator (or none) → times chosen
	VsGold     *Pair                     `json:"adjudicated_vs_gold,omitempty"`
}

// Report is what a run's annotations add up to.
type Report struct {
	Run           string              `json:"run"`
	Dataset       string              `json:"dataset"`
	Items         int                 `json:"items"`
	Annotators    []AnnotatorReport   `json:"annotators"`
	Pairs         []Pair              `json:"pairs"`
	Reviewer      *ReviewerReport     `json:"reviewer,omitempty"`
	Disagreements map[string][]string `json:"disagreements"` // claim → the fields its annotations differ on
	Verdicts      map[string]int      `json:"verdicts"`      // choice → claims
	Generated     time.Time           `json:"generated"`
}

// Report measures the run and writes report.json.
func (r *Run) Report() (*Report, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rep := &Report{Run: r.ID, Dataset: r.Dataset.ID, Items: len(r.Settings.Items), Disagreements: map[string][]string{},
		Verdicts: map[string]int{}, Generated: time.Now().UTC()}
	in := map[string]bool{}
	for _, id := range r.Settings.Items {
		in[id] = true
	}
	only := func(m map[string]Annotation) map[string]Annotation {
		out := map[string]Annotation{}
		for id, a := range m {
			if in[id] {
				out[id] = a
			}
		}
		return out
	}
	gold := only(r.Dataset.Gold)
	for _, p := range r.Settings.Annotators {
		ar := AnnotatorReport{Name: p.Name, Model: p.Model, Provider: p.Provider, Schemes: map[string]int{}}
		anns := only(r.anns[p.Name])
		ar.Annotated, ar.Failed = len(anns), len(r.failures[p.Name])
		for _, a := range anns {
			ar.Schemes[FieldOf(a, "scheme")]++
			if r.coll != nil {
				if n := len(r.coll.Check(a, r.Dataset)); n == 0 {
					ar.Valid++
				} else {
					ar.Problems += n
				}
			}
		}
		if len(gold) > 0 && len(anns) > 0 {
			pr := Agree(p.Name, "gold", anns, gold)
			ar.VsGold = &pr
		}
		rep.Annotators = append(rep.Annotators, ar)
	}
	for i, a := range r.Settings.Annotators {
		for _, b := range r.Settings.Annotators[i+1:] {
			rep.Pairs = append(rep.Pairs, Agree(a.Name, b.Name, only(r.anns[a.Name]), only(r.anns[b.Name])))
		}
	}
	for _, id := range r.Settings.Items {
		if names, anns := r.annotationsOf(id); len(names) >= 2 {
			if d := Disagrees(anns); len(d) > 0 {
				rep.Disagreements[id] = d
			}
		}
	}
	if rp := r.Settings.Reviewer; rp != nil {
		rr := &ReviewerReport{Name: rp.Name, Chosen: map[string]map[string]int{}}
		adj := map[string]Annotation{}
		for id, rv := range r.reviews {
			if !in[id] {
				continue
			}
			rr.Reviewed++
			if rv.Error != "" {
				rr.Errors++
				continue
			}
			if rv.NeedsHuman {
				rr.NeedsHuman++
			}
			for f, ch := range rv.Fields {
				who := ch.Annotator
				if who == "" {
					who = "none"
				}
				if rr.Chosen[f] == nil {
					rr.Chosen[f] = map[string]int{}
				}
				rr.Chosen[f][who]++
			}
			if rv.Adjudicated != nil {
				adj[id] = rv.Adjudicated
			}
		}
		if len(gold) > 0 && len(adj) > 0 {
			pr := Agree("reviewer", "gold", adj, gold)
			rr.VsGold = &pr
		}
		rep.Reviewer = rr
	}
	for id, v := range r.verdicts {
		if in[id] {
			rep.Verdicts[v.Choice]++
		}
	}
	b, err := json.MarshalIndent(rep, "", " ")
	if err != nil {
		return nil, err
	}
	return rep, writeFile(filepath.Join(r.Dir, "report.json"), b)
}
