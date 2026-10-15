package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/annotate"
	"github.com/yourorg/wag/internal/kialo"
	"github.com/yourorg/wag/internal/llm"
)

// Annotation in the dashboard: datasets (built in, or folders in the data
// folder's datasets/), runs (one folder each in annotation/), and their
// work, which goes on in the server, whether or not a browser watches.

// dataset finds a dataset by ID: a built-in one, or folder:NAME for the
// data folder's datasets/NAME.
func (s *Server) dataset(id, dir string) (*annotate.Dataset, error) {
	if name, ok := strings.CutPrefix(id, "folder:"); ok || dir != "" {
		if dir == "" {
			p, err := s.inStore("datasets/" + name)
			if err != nil {
				return nil, err
			}
			dir = p
		}
		d, err := annotate.FromDir(dir)
		if err != nil {
			return nil, err
		}
		d.ID = id
		return d, nil
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
	return nil, fmt.Errorf("no dataset %q", id)
}

// run returns the run with the ID, from memory or from its folder, with the
// clients of its annotators and reviewer attached when the players exist
// with the same service and model.
func (s *Server) run(id string) (*annotate.Run, []string, error) {
	if id == "" || annotate.FileName(id) != id {
		return nil, nil, fmt.Errorf("no run %q", id)
	}
	s.mu.Lock()
	r := s.runs[id]
	s.mu.Unlock()
	if r == nil {
		dir, err := s.inStore("annotation/" + id)
		if err != nil {
			return nil, nil, err
		}
		if r, err = annotate.Open(dir, s.dataset); err != nil {
			return nil, nil, err
		}
		s.mu.Lock()
		if x := s.runs[id]; x != nil {
			r = x
		} else {
			s.runs[id] = r
		}
		s.mu.Unlock()
	}
	var notes []string
	profiles := append([]llm.Profile{}, r.Settings.Annotators...)
	if r.Settings.Reviewer != nil {
		profiles = append(profiles, *r.Settings.Reviewer)
	}
	s.mu.Lock()
	for _, p := range profiles {
		c := s.models[p.Name]
		switch {
		case c == nil:
			notes = append(notes, fmt.Sprintf("%s is not connected: add the player again in Models (%s, %s)", p.Name, p.Provider, p.Model))
		case c.Profile.Provider != p.Provider || c.Profile.Model != p.Model:
			notes = append(notes, fmt.Sprintf("the player %s is now %s %s, not the %s %s this run began with: add a player with the run's model to continue",
				p.Name, c.Profile.Provider, c.Profile.Model, p.Provider, p.Model))
		default:
			r.Attach(p.Name, c)
		}
	}
	s.mu.Unlock()
	return r, notes, nil
}

// runStatus is a run's status with what the dashboard should say about it.
func runStatus(r *annotate.Run, notes []string) map[string]any {
	st := r.Status()
	return map[string]any{"status": st, "notes": notes}
}

// handleAnnotate serves /api/annotate/...
func (s *Server) handleAnnotate(w http.ResponseWriter, r *http.Request) {
	if !local(w, r) {
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/annotate/"), "/"), "/")
	switch {
	case parts[0] == "datasets" && len(parts) == 2 && parts[1] == "kialo":
		s.kialoDataset(w, r)
	case parts[0] == "datasets":
		folder, _ := s.inStore("datasets")
		ds, errs := annotate.Datasets(folder)
		var es []string
		for _, e := range errs {
			es = append(es, e.Error())
		}
		writeJSON(w, map[string]any{"datasets": ds, "folder": folder, "errors": es})
	case parts[0] == "runs" && len(parts) == 1 && r.Method == http.MethodGet:
		dir, err := s.inStore("annotation")
		if err != nil {
			writeJSON(w, map[string]any{"runs": []any{}})
			return
		}
		entries, _ := os.ReadDir(dir)
		out := []any{}
		for i := len(entries) - 1; i >= 0; i-- { // newest first: the IDs begin with the date
			e := entries[i]
			if !e.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(dir, e.Name(), "run.json")); err != nil {
				continue
			}
			run, notes, err := s.run(e.Name())
			if err != nil {
				out = append(out, map[string]any{"id": e.Name(), "error": err.Error()})
				continue
			}
			out = append(out, runStatus(run, notes))
		}
		writeJSON(w, map[string]any{"runs": out})
	case parts[0] == "runs" && len(parts) == 1:
		s.newRun(w, r)
	case parts[0] == "runs" && len(parts) >= 2:
		run, notes, err := s.run(parts[1])
		if err != nil {
			httpError(w, err, http.StatusNotFound)
			return
		}
		action := ""
		if len(parts) > 2 {
			action = parts[2]
		}
		s.runAction(w, r, run, notes, action, parts)
	default:
		httpError(w, fmt.Errorf("unknown %s", r.URL.Path), http.StatusNotFound)
	}
}

// newRun makes a run: POST /api/annotate/runs {dataset, annotators, reviewer, count | items}.
func (s *Server) newRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Dataset    string   `json:"dataset"`
		Annotators []string `json:"annotators"`
		Reviewer   string   `json:"reviewer"`
		Count      int      `json:"count"`
		Items      []string `json:"items"`
		Note       string   `json:"note"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	d, err := s.dataset(req.Dataset, "")
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	clients := map[string]*llm.Client{}
	var ann []llm.Profile
	var rev *llm.Profile
	s.mu.Lock()
	for _, name := range req.Annotators {
		c := s.models[name]
		if c == nil {
			s.mu.Unlock()
			httpError(w, fmt.Errorf("no player %q that can play: connect it in Models", name), http.StatusBadRequest)
			return
		}
		clients[name], ann = c, append(ann, c.Profile)
	}
	if req.Reviewer != "" {
		c := s.models[req.Reviewer]
		if c == nil {
			s.mu.Unlock()
			httpError(w, fmt.Errorf("no player %q that can review: connect it in Models", req.Reviewer), http.StatusBadRequest)
			return
		}
		p := c.Profile
		clients[req.Reviewer], rev = c, &p
	}
	s.mu.Unlock()
	items := req.Items
	if len(items) == 0 {
		items = d.IDs()
		if req.Count > 0 && req.Count < len(items) {
			items = items[:req.Count]
		}
	}
	base, err := s.inStore("annotation")
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	id := time.Now().Format("2006-01-02_150405") + "_" + annotate.FileName(strings.TrimPrefix(d.ID, "folder:"))
	for n := 2; ; n++ {
		if _, err := os.Stat(filepath.Join(base, id)); os.IsNotExist(err) {
			break
		}
		id = fmt.Sprintf("%s_%d", strings.TrimSuffix(id, fmt.Sprintf("_%d", n-1)), n)
	}
	run, err := annotate.New(filepath.Join(base, id), d, annotate.Settings{Annotators: ann, Reviewer: rev, Items: items, Note: req.Note}, clients)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.runs[id] = run
	s.mu.Unlock()
	writeJSON(w, runStatus(run, nil))
}

// runAction serves /api/annotate/runs/{id}[/{action}…].
func (s *Server) runAction(w http.ResponseWriter, r *http.Request, run *annotate.Run, notes []string, action string, parts []string) {
	background := func(f func() error) {
		go func() {
			if err := f(); err != nil {
				fmt.Fprintf(os.Stderr, "annotation run %s: %v\n", run.ID, err)
			}
		}()
		time.Sleep(50 * time.Millisecond) // the status shows it running
		writeJSON(w, runStatus(run, notes))
	}
	post := func() bool {
		if r.Method != http.MethodPost {
			httpError(w, fmt.Errorf("use POST"), http.StatusMethodNotAllowed)
			return false
		}
		return true
	}
	switch action {
	case "":
		writeJSON(w, runStatus(run, notes))
	case "start", "next":
		if !post() {
			return
		}
		var req struct {
			Items []string `json:"items"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if b := run.Busy(); b != "" {
			httpError(w, fmt.Errorf("the run is %s: wait, or stop it", b), http.StatusConflict)
			return
		}
		items := req.Items
		if action == "next" && len(items) == 0 {
			next := run.Status().NextItem
			if next == "" {
				httpError(w, fmt.Errorf("every annotator has annotated every claim of the run"), http.StatusConflict)
				return
			}
			items = []string{next}
		}
		background(func() error { return run.Annotate(s.ctx, items) })
	case "review":
		if !post() {
			return
		}
		var req struct {
			All   bool     `json:"all"`
			Items []string `json:"items"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if run.Settings.Reviewer == nil {
			httpError(w, fmt.Errorf("this run has no reviewer: start a new run with one"), http.StatusBadRequest)
			return
		}
		if b := run.Busy(); b != "" {
			httpError(w, fmt.Errorf("the run is %s: wait, or stop it", b), http.StatusConflict)
			return
		}
		background(func() error { return run.Review(s.ctx, req.Items, !req.All) })
	case "stop":
		if !post() {
			return
		}
		run.Stop()
		writeJSON(w, runStatus(run, notes))
	case "report":
		rep, err := run.Report()
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, rep)
	case "item":
		if len(parts) < 4 {
			httpError(w, fmt.Errorf("which claim?"), http.StatusBadRequest)
			return
		}
		v, err := run.Item(parts[3])
		if err != nil {
			httpError(w, err, http.StatusNotFound)
			return
		}
		writeJSON(w, v)
	case "verdict":
		if !post() {
			return
		}
		var v annotate.Verdict
		if !readJSON(w, r, &v) {
			return
		}
		v.At = time.Time{}
		if err := run.SetVerdict(v); err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		item, _ := run.Item(v.Item)
		writeJSON(w, item)
	case "open":
		if !post() {
			return
		}
		if err := openFolder(run.Dir); err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"opened": run.Dir})
	default:
		httpError(w, fmt.Errorf("unknown action %q", action), http.StatusNotFound)
	}
}

// kialoDataset makes a dataset from a Kialo discussion, as a folder of the
// data folder's datasets/: POST /api/annotate/datasets/kialo with
// {"kialo_id": ID} (a discussion of --kialo-dir) or {"graph": {…}} (a
// graph.json chosen in the browser), and "depth", "limit" (0: all).
func (s *Server) kialoDataset(w http.ResponseWriter, r *http.Request) {
	if !jsonPost(w, r) {
		return
	}
	var req struct {
		KialoID string          `json:"kialo_id"`
		Graph   json.RawMessage `json:"graph"`
		Depth   int             `json:"depth"`
		Limit   int             `json:"limit"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 256<<20)).Decode(&req); err != nil {
		httpError(w, fmt.Errorf("bad request: %v", err), http.StatusBadRequest)
		return
	}
	b := []byte(req.Graph)
	if req.KialoID != "" {
		if s.kialoDir == "" || annotate.FileName(req.KialoID) != req.KialoID {
			httpError(w, fmt.Errorf("no captured discussion %q (start the server with --kialo-dir, or choose a graph.json)", req.KialoID), http.StatusBadRequest)
			return
		}
		var err error
		if b, err = os.ReadFile(filepath.Join(s.kialoDir, req.KialoID, "graph.json")); err != nil {
			httpError(w, err, http.StatusNotFound)
			return
		}
	}
	d, err := kialo.Parse(b)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	items, n, err := d.Items(req.Depth, req.Limit)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	name := "kialo-" + annotate.FileName(d.ID)
	if req.Limit > 0 || req.Depth > 0 {
		name += fmt.Sprintf("-%d-claims", n)
	}
	dir, err := s.inStore("datasets/" + name)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	man, _ := json.MarshalIndent(map[string]string{"name": fmt.Sprintf("Kialo %s — %s (%d claims)", d.ID, d.Title, n),
		"description": "Made from a capture of the Kialo research visualizer (" + d.URL + "). No gold annotation: score the annotators against each other, and review."}, "", " ")
	for f, data := range map[string][]byte{"items.json": items, "dataset.json": man} {
		if err := os.WriteFile(filepath.Join(dir, f), data, 0o644); err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, map[string]any{"dataset": "folder:" + name, "dir": dir, "claims": n})
}
