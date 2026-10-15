package annotate

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/wag/internal/llm"
)

// RunFormat names the format of a run's run.json.
const RunFormat = "wag-annotation-run/2"

// readme is the README.txt of a run's folder: where everything is.
const readme = `This folder is one annotation run of WAG (docs/llm.md, "Annotation").

run.json                  the run: the dataset and its inputs' SHA-256, the annotators'
                          and reviewer's model profiles (no API keys), the claims, progress
annotations/NAME.json     each annotator's annotations, in the guidelines' output format
                          (docs/kialo-batch2/validate.py and agreement.py read these files)
calls.jsonl               every call to a model, one JSON object a line: the whole request,
                          the answer or the error, token usage, latency, request IDs
review/model-NAME.json    a reviewer model's decisions, claim by claim, field by field
review/human.json         your verdicts, claim by claim
report.json               validation of every annotation, and agreement: between the
                          annotators, with the gold annotation, with the reviewer
`

// Settings are what a run was started with.
type Settings struct {
	Dataset    string        `json:"dataset"`
	DatasetDir string        `json:"dataset_dir,omitempty"` // the folder of a dataset that is not built in
	Annotators []llm.Profile `json:"annotators"`
	Reviewer   *llm.Profile  `json:"reviewer,omitempty"`
	Items      []string      `json:"items"`
	Created    time.Time     `json:"created"`
	Note       string        `json:"note,omitempty"`
}

// Progress is how far one annotator, or the reviewer, has got.
type Progress struct {
	Name         string `json:"name"`
	Role         string `json:"role"` // annotator, reviewer
	Total        int    `json:"total"`
	Done         int    `json:"done"`
	Failed       int    `json:"failed"`
	Current      string `json:"current,omitempty"`
	Running      bool   `json:"running"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Calls        int    `json:"calls"`
	LastError    string `json:"last_error,omitempty"`
}

// FieldChoice is a reviewer's decision on one field.
type FieldChoice struct {
	Choice    string `json:"choice"`              // A, B, … or none
	Annotator string `json:"annotator,omitempty"` // whom the letter stands for
	Why       string `json:"why,omitempty"`
}

// Review is a reviewer model's review of one claim's annotations.
type Review struct {
	Item        string                 `json:"item"`
	Reviewer    string                 `json:"reviewer"`
	Letters     map[string]string      `json:"letters"` // A → the annotator it labels
	Fields      map[string]FieldChoice `json:"fields,omitempty"`
	Adjudicated Annotation             `json:"adjudicated,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
	NeedsHuman  bool                   `json:"needs_human,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Raw         string                 `json:"raw,omitempty"` // the reply, when it could not be read
}

// Verdict is a person's judgement of one claim's annotations.
type Verdict struct {
	Item   string    `json:"item"`
	Choice string    `json:"choice"` // an annotator's name, "reviewer", "gold", "none" (none is right) or "unsure"
	Note   string    `json:"note,omitempty"`
	By     string    `json:"by,omitempty"`
	At     time.Time `json:"at"`
}

// Run is an annotation run. Its folder (Dir) holds everything (readme).
type Run struct {
	ID       string
	Dir      string
	Dataset  *Dataset
	Settings Settings

	mu       sync.Mutex
	clients  map[string]*llm.Client
	anns     map[string]map[string]Annotation // annotator → claim → annotation
	failures map[string]map[string]string
	reviews  map[string]*Review // claim → the reviewer's review
	verdicts map[string]Verdict
	progress map[string]*Progress
	busy     string // what the run is doing: "", "annotating", "reviewing"
	cancel   context.CancelFunc
	coll     *Collection
}

var unsafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// FileName is a name made safe for a file name.
func FileName(name string) string { return strings.Trim(unsafe.ReplaceAllString(name, "_"), "_") }

// New makes a run in dir (created), for the dataset, with the settings;
// clients are the annotators' and the reviewer's, by profile name.
func New(dir string, d *Dataset, s Settings, clients map[string]*llm.Client) (*Run, error) {
	if len(s.Annotators) == 0 {
		return nil, fmt.Errorf("a run needs at least one annotator")
	}
	names := map[string]bool{}
	for _, p := range s.Annotators {
		if names[p.Name] {
			return nil, fmt.Errorf("annotator %s is chosen twice", p.Name)
		}
		names[p.Name] = true
	}
	if len(s.Items) == 0 {
		s.Items = d.IDs()
	}
	for _, id := range s.Items {
		if d.Item(id) == nil {
			return nil, fmt.Errorf("the dataset has no claim %s", id)
		}
	}
	if s.Created.IsZero() {
		s.Created = time.Now().UTC()
	}
	s.Dataset = d.ID
	if d.Source != "built in" {
		s.DatasetDir = d.Source
	}
	if err := os.MkdirAll(filepath.Join(dir, "annotations"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, "review"), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte(readme), 0o644); err != nil {
		return nil, err
	}
	r := newRun(dir, d, s)
	for name, c := range clients {
		r.clients[name] = c
	}
	if err := r.saveRun(); err != nil {
		return nil, err
	}
	return r, nil
}

func newRun(dir string, d *Dataset, s Settings) *Run {
	r := &Run{ID: filepath.Base(dir), Dir: dir, Dataset: d, Settings: s, clients: map[string]*llm.Client{},
		anns: map[string]map[string]Annotation{}, failures: map[string]map[string]string{}, reviews: map[string]*Review{},
		verdicts: map[string]Verdict{}, progress: map[string]*Progress{}}
	for _, p := range s.Annotators {
		r.anns[p.Name], r.failures[p.Name] = map[string]Annotation{}, map[string]string{}
		r.progress[p.Name] = &Progress{Name: p.Name, Role: "annotator", Total: len(s.Items)}
	}
	if s.Reviewer != nil {
		r.progress["review:"+s.Reviewer.Name] = &Progress{Name: s.Reviewer.Name, Role: "reviewer"}
	}
	if c, err := LoadCollection(d.WaltonYAML()); err == nil {
		r.coll = c
	}
	return r
}

// Open reads the run in dir (its run.json, annotations, reviews and
// verdicts); datasets finds its dataset by ID and folder ("" if built in).
func Open(dir string, datasets func(id, dir string) (*Dataset, error)) (*Run, error) {
	b, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		return nil, err
	}
	var f struct {
		Settings Settings                     `json:"settings"`
		Failures map[string]map[string]string `json:"failures"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("run.json: %v", err)
	}
	d, err := datasets(f.Settings.Dataset, f.Settings.DatasetDir)
	if err != nil {
		return nil, fmt.Errorf("its dataset %s: %v", f.Settings.Dataset, err)
	}
	r := newRun(dir, d, f.Settings)
	for name := range r.anns {
		if b, err := os.ReadFile(r.annotationsPath(name)); err == nil {
			if m, err := parseAnnotations(b); err == nil {
				r.anns[name] = m
			}
		}
		for id, e := range f.Failures[name] {
			if r.anns[name][id] == nil {
				r.failures[name][id] = e
			}
		}
	}
	if s := f.Settings.Reviewer; s != nil {
		if b, err := os.ReadFile(r.reviewPath(s.Name)); err == nil {
			var list []*Review
			if json.Unmarshal(b, &list) == nil {
				for _, rv := range list {
					r.reviews[rv.Item] = rv
				}
			}
		}
	}
	if b, err := os.ReadFile(filepath.Join(dir, "review", "human.json")); err == nil {
		var list []Verdict
		if json.Unmarshal(b, &list) == nil {
			for _, v := range list {
				r.verdicts[v.Item] = v
			}
		}
	}
	r.recount()
	return r, nil
}

// Attach gives the run the client of an annotator or of its reviewer (they
// are not saved: a reopened run needs them again to continue).
func (r *Run) Attach(name string, c *llm.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[name] = c
}

func (r *Run) annotationsPath(name string) string {
	return filepath.Join(r.Dir, "annotations", FileName(name)+".json")
}

func (r *Run) reviewPath(name string) string {
	return filepath.Join(r.Dir, "review", "model-"+FileName(name)+".json")
}

// recount sets the progress counts from the annotations. Caller holds r.mu
// or owns r.
func (r *Run) recount() {
	for name, p := range r.progress {
		if p.Role == "annotator" {
			p.Done, p.Failed = 0, 0
			for _, id := range r.Settings.Items {
				if r.anns[name][id] != nil {
					p.Done++
				} else if r.failures[name][id] != "" {
					p.Failed++
				}
			}
		} else {
			p.Done = len(r.reviews)
		}
	}
}

// saveRun writes run.json. Caller holds r.mu or owns r.
func (r *Run) saveRun() error {
	ps := r.progressList(false)
	b, err := json.MarshalIndent(map[string]any{"format": RunFormat, "id": r.ID, "settings": r.Settings,
		"dataset": r.Dataset, "inputs_sha256": r.Dataset.Hashes,
		"annotator_system_prompt_sha256": sha([]byte(r.Dataset.AnnotatorSystem())),
		"progress":                       ps, "failures": r.failures, "saved": time.Now().UTC()}, "", " ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(r.Dir, "run.json"), b)
}

// writeFile writes through a temporary file, so a reader never sees half.
func writeFile(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// saveAnnotations writes an annotator's file, in the dataset's order.
// Caller holds r.mu.
func (r *Run) saveAnnotations(name string) error {
	var list []Annotation
	for _, id := range r.Dataset.IDs() {
		if a := r.anns[name][id]; a != nil {
			list = append(list, a)
		}
	}
	b, err := json.MarshalIndent(map[string]any{"annotator": name, "items": list}, "", " ")
	if err != nil {
		return err
	}
	return writeFile(r.annotationsPath(name), b)
}

// logCall appends a call to calls.jsonl.
func (r *Run) logCall(role, name, item string, attempt int, a llm.Attempt) {
	line, _ := json.Marshal(map[string]any{"format": "wag-annotation-call/1", "role": role, "name": name, "item": item, "attempt": attempt, "call": a})
	r.mu.Lock()
	defer r.mu.Unlock()
	if p := r.progress[progressKey(role, name)]; p != nil {
		p.Calls++
		if a.Response != nil {
			if u := a.Response.Usage.InputTokens; u != nil {
				p.InputTokens += *u
			}
			if u := a.Response.Usage.OutputTokens; u != nil {
				p.OutputTokens += *u
			}
		}
	}
	f, err := os.OpenFile(filepath.Join(r.Dir, "calls.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(append(line, '\n'))
}

func progressKey(role, name string) string {
	if role == "reviewer" {
		return "review:" + name
	}
	return name
}

// Busy says what the run is doing ("" if nothing).
func (r *Run) Busy() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.busy
}

// begin marks the run busy with what, and returns the context to do it in;
// an error if it is busy already.
func (r *Run) begin(ctx context.Context, what string) (context.Context, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.busy != "" {
		return nil, fmt.Errorf("the run is %s already", r.busy)
	}
	ctx, r.cancel = context.WithCancel(ctx)
	r.busy = what
	return ctx, nil
}

func (r *Run) end() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
	r.busy, r.cancel = "", nil
	for _, p := range r.progress {
		p.Running, p.Current = false, ""
	}
	r.recount()
	_ = r.saveRun()
}

// Stop stops what the run is doing, after the calls under way.
func (r *Run) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
}

// Annotate has every annotator annotate the given claims (all the run's
// claims if none), in the dataset's order, skipping those it has already
// annotated; the annotators work in parallel. It returns when they are
// done, or stopped. A claim that fails is recorded and retried by the
// next call; an annotator whose key or model is refused stops.
func (r *Run) Annotate(ctx context.Context, items []string) error {
	ctx, err := r.begin(ctx, "annotating")
	if err != nil {
		return err
	}
	defer r.end()
	if len(items) == 0 {
		items = r.Settings.Items
	}
	var wg sync.WaitGroup
	for _, p := range r.Settings.Annotators {
		r.mu.Lock()
		c := r.clients[p.Name]
		r.mu.Unlock()
		if c == nil {
			r.fail(p.Name, "", "no model client: connect its service in Models")
			continue
		}
		wg.Add(1)
		go func(name string, c *llm.Client) {
			defer wg.Done()
			r.annotateAll(ctx, name, c, items)
		}(p.Name, c)
	}
	wg.Wait()
	return ctx.Err()
}

func (r *Run) fail(name, item, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if item != "" {
		r.failures[name][item] = msg
	}
	if p := r.progress[name]; p != nil {
		p.LastError = msg
	}
	r.recount()
	_ = r.saveRun()
}

func (r *Run) annotateAll(ctx context.Context, name string, c *llm.Client, items []string) {
	order := r.Dataset.IDs()
	system := r.Dataset.AnnotatorSystem()
	r.mu.Lock()
	r.progress[name].Running = true
	r.mu.Unlock()
	for _, id := range items {
		if ctx.Err() != nil {
			return
		}
		r.mu.Lock()
		done := r.anns[name][id] != nil
		before := map[string]Annotation{}
		var prior []string
		for _, x := range order {
			if x == id {
				break
			}
			if a := r.anns[name][x]; a != nil {
				before[x], prior = a, append(prior, x)
			}
		}
		r.progress[name].Current = id
		r.mu.Unlock()
		if done {
			continue
		}
		it := r.Dataset.Item(id)
		a, err := r.annotateOne(ctx, name, c, system, it, before, prior)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			r.fail(name, id, err.Error())
			if e, ok := err.(*llm.Error); ok && !e.Retryable() && e.Kind != llm.KindBadReply {
				return // a refused key or model will not pass on the next claim
			}
			continue
		}
		r.mu.Lock()
		r.anns[name][id] = a
		delete(r.failures[name], id)
		r.progress[name].LastError = ""
		r.recount()
		_ = r.saveAnnotations(name)
		_ = r.saveRun()
		r.mu.Unlock()
	}
}

// annotateOne asks one annotator for one claim's annotation; a reply that
// is not an annotation of the claim is sent back once, with the reason.
func (r *Run) annotateOne(ctx context.Context, name string, c *llm.Client, system string, it *Item, before map[string]Annotation, prior []string) (Annotation, error) {
	msgs := []llm.Message{{Role: "user", Content: r.Dataset.AnnotatorPrompt(it, before, prior)}}
	for attempt := 1; attempt <= 2; attempt++ {
		resp, err := c.Complete(ctx, c.Request(system, msgs), func(a llm.Attempt) { r.logCall("annotator", name, it.ID, attempt, a) })
		if err != nil {
			return nil, err
		}
		a, perr := parseAnnotation(resp.Text, it.ID)
		if perr == nil {
			return a, nil
		}
		if attempt == 2 {
			return nil, &llm.Error{Provider: c.Provider.Name(), Kind: llm.KindBadReply, Message: perr.Error()}
		}
		msgs = append(msgs, llm.Message{Role: "assistant", Content: resp.Text},
			llm.Message{Role: "user", Content: "That reply cannot be used: " + perr.Error() + ". Reply with the one JSON object for " + it.ID + " only."})
	}
	return nil, nil
}

// Review has the reviewer review the annotations of the given claims (all
// the run's if none) that at least two annotators annotated; with
// onlyDisagreements, only those whose annotations differ on a field. It
// skips claims it has reviewed.
func (r *Run) Review(ctx context.Context, items []string, onlyDisagreements bool) error {
	rp := r.Settings.Reviewer
	if rp == nil {
		return fmt.Errorf("the run has no reviewer")
	}
	r.mu.Lock()
	c := r.clients[rp.Name]
	r.mu.Unlock()
	if c == nil {
		return fmt.Errorf("no model client for the reviewer %s: connect its service in Models", rp.Name)
	}
	ctx, err := r.begin(ctx, "reviewing")
	if err != nil {
		return err
	}
	defer r.end()
	if len(items) == 0 {
		items = r.Settings.Items
	}
	key := "review:" + rp.Name
	system := r.Dataset.ReviewerSystem()
	var todo []string
	r.mu.Lock()
	for _, id := range items {
		names, anns := r.annotationsOf(id)
		if len(names) < 2 || r.reviews[id] != nil && r.reviews[id].Error == "" {
			continue
		}
		if onlyDisagreements && len(Disagrees(anns)) == 0 {
			continue
		}
		todo = append(todo, id)
	}
	p := r.progress[key]
	p.Total, p.Running = len(r.reviews)+len(todo), true
	r.mu.Unlock()
	for _, id := range todo {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		r.mu.Lock()
		names, _ := r.annotationsOf(id)
		anns := map[string]Annotation{}
		for _, n := range names {
			anns[n] = r.anns[n][id]
		}
		p.Current = id
		r.mu.Unlock()
		rv := r.reviewOne(ctx, rp.Name, c, system, r.Dataset.Item(id), names, anns)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		r.mu.Lock()
		r.reviews[id] = rv
		if rv.Error != "" {
			p.LastError = rv.Error
		}
		r.recount()
		_ = r.saveReviews()
		r.mu.Unlock()
	}
	return nil
}

// annotationsOf returns the names of the annotators who annotated the claim,
// in a rotating order (so that no annotator is always A), and their
// annotations in that order. Caller holds r.mu.
func (r *Run) annotationsOf(id string) ([]string, []Annotation) {
	var names []string
	for _, p := range r.Settings.Annotators {
		if r.anns[p.Name][id] != nil {
			names = append(names, p.Name)
		}
	}
	if len(names) > len(Letters) {
		names = names[:len(Letters)]
	}
	if n := len(names); n > 1 {
		h := fnv.New32a()
		h.Write([]byte(id))
		k := int(h.Sum32() % uint32(n))
		names = append(append([]string{}, names[k:]...), names[:k]...)
	}
	var anns []Annotation
	for _, n := range names {
		anns = append(anns, r.anns[n][id])
	}
	return names, anns
}

func (r *Run) reviewOne(ctx context.Context, name string, c *llm.Client, system string, it *Item, names []string, anns map[string]Annotation) *Review {
	rv := &Review{Item: it.ID, Reviewer: name, Letters: map[string]string{}}
	for i, n := range names {
		rv.Letters[string(Letters[i])] = n
	}
	resp, err := c.Complete(ctx, c.Request(system, []llm.Message{{Role: "user", Content: r.Dataset.ReviewerPrompt(it, names, anns)}}),
		func(a llm.Attempt) { r.logCall("reviewer", name, it.ID, 1, a) })
	if err != nil {
		rv.Error = err.Error()
		return rv
	}
	obj, err := jsonObject(resp.Text)
	if err != nil {
		rv.Error, rv.Raw = "the review cannot be read: "+err.Error(), resp.Text
		return rv
	}
	var parsed struct {
		Fields      map[string]FieldChoice `json:"fields"`
		Adjudicated Annotation             `json:"adjudicated"`
		Confidence  float64                `json:"confidence"`
		NeedsHuman  bool                   `json:"needs_human"`
		Notes       string                 `json:"notes"`
	}
	b, _ := json.Marshal(obj)
	if err := json.Unmarshal(b, &parsed); err != nil {
		rv.Error, rv.Raw = "the review cannot be read: "+err.Error(), resp.Text
		return rv
	}
	rv.Fields, rv.Adjudicated, rv.Confidence, rv.NeedsHuman, rv.Notes = parsed.Fields, parsed.Adjudicated, parsed.Confidence, parsed.NeedsHuman, parsed.Notes
	for f, ch := range rv.Fields {
		ch.Annotator = rv.Letters[strings.ToUpper(strings.TrimSpace(ch.Choice))]
		rv.Fields[f] = ch
	}
	if rv.Adjudicated != nil {
		rv.Adjudicated["kialo_claim_id"] = it.ID
	}
	return rv
}

// saveReviews writes the reviewer's file. Caller holds r.mu.
func (r *Run) saveReviews() error {
	var list []*Review
	for _, id := range r.Dataset.IDs() {
		if rv := r.reviews[id]; rv != nil {
			list = append(list, rv)
		}
	}
	b, err := json.MarshalIndent(list, "", " ")
	if err != nil {
		return err
	}
	return writeFile(r.reviewPath(r.Settings.Reviewer.Name), b)
}

// SetVerdict records a person's verdict on a claim.
func (r *Run) SetVerdict(v Verdict) error {
	if r.Dataset.Item(v.Item) == nil {
		return fmt.Errorf("no claim %s", v.Item)
	}
	ok := map[string]bool{"reviewer": true, "gold": true, "none": true, "unsure": true, "": true}
	for _, p := range r.Settings.Annotators {
		ok[p.Name] = true
	}
	if !ok[v.Choice] {
		return fmt.Errorf("unknown choice %q", v.Choice)
	}
	if v.At.IsZero() {
		v.At = time.Now().UTC()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if v.Choice == "" && v.Note == "" {
		delete(r.verdicts, v.Item)
	} else {
		r.verdicts[v.Item] = v
	}
	var list []Verdict
	for _, id := range r.Dataset.IDs() {
		if x, ok := r.verdicts[id]; ok {
			list = append(list, x)
		}
	}
	b, err := json.MarshalIndent(list, "", " ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(r.Dir, "review", "human.json"), b)
}

// Status is a run at a glance.
type Status struct {
	ID         string      `json:"id"`
	Dir        string      `json:"dir"`
	Dataset    *Dataset    `json:"dataset"`
	Settings   Settings    `json:"settings"`
	Busy       string      `json:"busy,omitempty"`
	Progress   []*Progress `json:"progress"`
	Reviewed   int         `json:"reviewed"`
	Verdicts   int         `json:"verdicts"`
	Missing    []string    `json:"missing_clients,omitempty"` // annotators or reviewer without a client
	NextItem   string      `json:"next_item,omitempty"`       // the first claim not every annotator has annotated
	Files      []string    `json:"files"`
	Disagreed  int         `json:"disagreed"` // claims whose annotations differ on a field
	Annotated2 int         `json:"annotated_by_two"`
}

// progressList lists the annotators' progress in their order, then the
// reviewer's (copies, if copy). Caller holds r.mu.
func (r *Run) progressList(copy bool) []*Progress {
	keys := []string{}
	for _, p := range r.Settings.Annotators {
		keys = append(keys, p.Name)
	}
	if rp := r.Settings.Reviewer; rp != nil {
		keys = append(keys, "review:"+rp.Name)
	}
	var out []*Progress
	for _, k := range keys {
		if p := r.progress[k]; p != nil {
			if copy {
				x := *p
				p = &x
			}
			out = append(out, p)
		}
	}
	return out
}

// Status returns the run's status.
func (r *Run) Status() Status {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := Status{ID: r.ID, Dir: r.Dir, Dataset: r.Dataset, Settings: r.Settings, Busy: r.busy, Reviewed: len(r.reviews), Verdicts: len(r.verdicts)}
	s.Progress = r.progressList(true)
	for _, p := range r.Settings.Annotators {
		if r.clients[p.Name] == nil {
			s.Missing = append(s.Missing, p.Name)
		}
	}
	if rp := r.Settings.Reviewer; rp != nil && r.clients[rp.Name] == nil {
		s.Missing = append(s.Missing, rp.Name)
	}
	for _, id := range r.Settings.Items {
		all := true
		for _, p := range r.Settings.Annotators {
			if r.anns[p.Name][id] == nil {
				all = false
			}
		}
		if !all && s.NextItem == "" {
			s.NextItem = id
		}
		names, anns := r.annotationsOf(id)
		if len(names) >= 2 {
			s.Annotated2++
			if len(Disagrees(anns)) > 0 {
				s.Disagreed++
			}
		}
	}
	s.Files = r.files()
	return s
}

// files lists the run's files, relative to its folder. Caller holds r.mu.
func (r *Run) files() []string {
	var out []string
	_ = filepath.Walk(r.Dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && !strings.HasSuffix(p, ".tmp") {
			rel, _ := filepath.Rel(r.Dir, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

// ItemView is one claim with everything about it: its annotations, their
// problems, the review, the verdict, the gold.
type ItemView struct {
	ID          string                `json:"id"`
	Item        json.RawMessage       `json:"item"`
	Annotations map[string]Annotation `json:"annotations"`
	Problems    map[string][]string   `json:"problems"`
	Failures    map[string]string     `json:"failures,omitempty"`
	Disagree    []string              `json:"disagree"`
	Gold        Annotation            `json:"gold,omitempty"`
	Review      *Review               `json:"review,omitempty"`
	Verdict     *Verdict              `json:"verdict,omitempty"`
}

// Item returns the view of a claim.
func (r *Run) Item(id string) (*ItemView, error) {
	it := r.Dataset.Item(id)
	if it == nil {
		return nil, fmt.Errorf("no claim %s", id)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	v := &ItemView{ID: id, Item: it.Raw, Annotations: map[string]Annotation{}, Problems: map[string][]string{}, Failures: map[string]string{},
		Gold: r.Dataset.Gold[id], Review: r.reviews[id]}
	var anns []Annotation
	for _, p := range r.Settings.Annotators {
		if a := r.anns[p.Name][id]; a != nil {
			v.Annotations[p.Name] = a
			anns = append(anns, a)
			if r.coll != nil {
				v.Problems[p.Name] = r.coll.Check(a, r.Dataset)
			}
		} else if e := r.failures[p.Name][id]; e != "" {
			v.Failures[p.Name] = e
		}
	}
	v.Disagree = Disagrees(anns)
	if x, ok := r.verdicts[id]; ok {
		v.Verdict = &x
	}
	return v, nil
}
