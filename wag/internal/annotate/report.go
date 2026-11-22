package annotate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AnnotatorReport is one annotator's results.
type AnnotatorReport struct {
	Name                   string         `json:"name"`
	Model                  string         `json:"model"`         // requested model, retained for compatibility
	ServedModels           map[string]int `json:"served_models"` // response counts by provider-reported model, including format retries
	UnknownServedResponses int            `json:"unknown_served_responses"`
	Provider               string         `json:"provider"`
	Annotated              int            `json:"annotated"`
	Failed                 int            `json:"failed"`
	Valid                  int            `json:"valid"`    // annotations the check finds nothing wrong with
	Problems               int            `json:"problems"` // problems found, in all
	Schemes                map[string]int `json:"schemes"`  // primary schemes chosen
	VsGold                 *Pair          `json:"vs_gold,omitempty"`
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
	// InputDrift says why the run's inputs are no longer those it was made
	// with ("" if they are); GoldSHA256 is the gold this report scored against.
	InputDrift string `json:"input_drift,omitempty"`
	GoldSHA256 string `json:"gold_sha256,omitempty"`
	// UnreadableCallLines counts lines of calls.jsonl that could not be read
	// (a crash mid-write); they are left out of served_models.
	UnreadableCallLines int       `json:"unreadable_call_lines,omitempty"`
	Generated           time.Time `json:"generated"`
}

// Report measures the run and writes report.json.
func (r *Run) Report() (*Report, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rep := &Report{Run: r.ID, Dataset: r.Dataset.ID, Items: len(r.Settings.Items), Disagreements: map[string][]string{},
		Verdicts: map[string]int{}, InputDrift: r.drift, GoldSHA256: r.Dataset.Hashes[r.Dataset.GoldFile], Generated: time.Now().UTC()}
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
	served, unknown, bad, err := r.servedModelCounts()
	if err != nil {
		return nil, err
	}
	rep.UnreadableCallLines = bad
	for _, p := range r.Settings.Annotators {
		ar := AnnotatorReport{Name: p.Name, Model: p.Model, Provider: p.Provider, ServedModels: served[p.Name], UnknownServedResponses: unknown[p.Name], Schemes: map[string]int{}}
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
			pr := AgreeAssigned(p.Name, "gold", anns, gold, r.Settings.Items)
			ar.VsGold = &pr
		}
		rep.Annotators = append(rep.Annotators, ar)
	}
	for i, a := range r.Settings.Annotators {
		for _, b := range r.Settings.Annotators[i+1:] {
			rep.Pairs = append(rep.Pairs, AgreeAssigned(a.Name, b.Name, only(r.anns[a.Name]), only(r.anns[b.Name]), r.Settings.Items))
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
			pr := AgreeAssigned("reviewer", "gold", adj, gold, r.Settings.Items)
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

// servedModelCounts reads the append-only calls file, so that a report keeps
// the models that actually answered after the run is reopened, and never takes
// a requested alias for a served model. A line that cannot be read (a crash
// mid-write) is skipped and counted, not fatal.
func (r *Run) servedModelCounts() (served map[string]map[string]int, unknown map[string]int, bad int, err error) {
	served, unknown = map[string]map[string]int{}, map[string]int{}
	f, err := os.Open(filepath.Join(r.Dir, "calls.jsonl"))
	if os.IsNotExist(err) {
		return served, unknown, 0, nil
	}
	if err != nil {
		return nil, nil, 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 256<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var row struct {
			Role string `json:"role"`
			Name string `json:"name"`
			Call struct {
				Response *struct {
					Model string `json:"model"`
				} `json:"response"`
			} `json:"call"`
		}
		if json.Unmarshal(line, &row) != nil {
			bad++
			continue
		}
		if row.Role != "annotator" || row.Call.Response == nil {
			continue
		}
		if row.Call.Response.Model == "" {
			unknown[row.Name]++
			continue
		}
		if served[row.Name] == nil {
			served[row.Name] = map[string]int{}
		}
		served[row.Name][row.Call.Response.Model]++
	}
	if err := sc.Err(); err != nil {
		return nil, nil, 0, fmt.Errorf("calls.jsonl: %w", err)
	}
	return served, unknown, bad, nil
}
