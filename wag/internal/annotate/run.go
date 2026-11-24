package annotate

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

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
budget.jsonl              each wag-llm annotate invocation's budget, one JSON object a
                          line: its limits, or none, and the calls and usage it counted
                          (D046; the dashboard records none yet)

If a file cannot be written, the run asks the models nothing new and says so; what it
received is kept until the next start writes it, before asking anything. A reply that
reached calls.jsonl is taken from it when the run is opened again, not asked again.
`

// Settings are what a run was started with.
type Settings struct {
	IndependentFirstPass bool          `json:"independent_first_pass"` // withhold earlier annotations, retaining source context
	Dataset              string        `json:"dataset"`
	DatasetDir           string        `json:"dataset_dir,omitempty"` // the folder of a dataset that is not built in
	Annotators           []llm.Profile `json:"annotators"`
	Reviewer             *llm.Profile  `json:"reviewer,omitempty"`
	Items                []string      `json:"items"`
	Created              time.Time     `json:"created"`
	Note                 string        `json:"note,omitempty"`
}

// Progress is how far one annotator, or the reviewer, has got.
type Progress struct {
	Name         string `json:"name"`
	Role         string `json:"role"` // annotator, reviewer
	Total        int    `json:"total"`
	Done         int    `json:"done"`    // results in the files
	Unsaved      int    `json:"unsaved"` // results held in memory that the files lack (a write failed)
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
	// These identify the material available when this later verdict was made.
	InputsSHA256    map[string]string `json:"inputs_sha256,omitempty"`
	InputDrift      string            `json:"input_drift,omitempty"`
	AssessmentBuild string            `json:"assessment_build,omitempty"`
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
	// inputHashes and promptHash are the run's identity: the hashes of the
	// inputs and of the annotator's system prompt when the run was created.
	// They are written to run.json unchanged for the run's life.
	inputHashes map[string]string
	promptHash  string
	// drift says why the run's inputs are no longer the ones it was made
	// with ("" if they are). Such a run can be read, reported on and given
	// verdicts, but not continued (begin refuses).
	drift string
	// builds are the WAG builds that have written this run, oldest first.
	builds        []string
	scoringInputs []string // pinned scoring-only filenames, not model inputs
	loadErrors    []string // damaged saved result files: inspectable, never silently overwritten
	curPrompt     string   // digest of the loaded dataset's annotator system prompt (immutable after load)
	// The records the files lack (durable.go): the first write that failed
	// since they were last all written, the call lines not written, what
	// each annotator's file and the reviewer's file hold, whether run.json
	// is older than the run, and the results taken from calls.jsonl.
	writeErr      error
	pending       [][]byte
	onDisk        map[string]map[string]bool
	reviewsOnDisk map[string]*Review
	runDirty      bool
	fromLog       map[string]bool
	calls         map[*pendingCall]bool // the model calls under way (call)
}

var unsafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// FileName is a name made safe for a file name.
func FileName(name string) string { return strings.Trim(unsafe.ReplaceAllString(name, "_"), "_") }

// checkResultNames prevents distinct annotators from sharing a saved file or
// a progress key after filename sanitizing or reviewer-role qualification.
func checkResultNames(s Settings) error {
	files := map[string]string{}
	for _, p := range s.Annotators {
		filename := strings.ToLower(FileName(p.Name)) // one file on case-insensitive file systems
		if filename == "" {
			return fmt.Errorf("annotator %q has no usable result filename", p.Name)
		}
		if other, ok := files[filename]; ok {
			return fmt.Errorf("annotators %q and %q share result filename %q; choose distinct names", other, p.Name, filename)
		}
		files[filename] = p.Name
		if s.Reviewer != nil && p.Name == progressKey("reviewer", s.Reviewer.Name) {
			return fmt.Errorf("annotator %q collides with the reviewer's progress key; choose another name", p.Name)
		}
	}
	return nil
}

// New makes a run in dir (created), for the dataset, with the settings;
// clients are the annotators' and the reviewer's, by profile name.
func New(dir string, d *Dataset, s Settings, clients map[string]*llm.Client) (*Run, error) {
	if len(s.Annotators) == 0 {
		return nil, fmt.Errorf("a run needs at least one annotator")
	}
	// Freeze schema bytes and mutable parameter maps before writing the run.
	s.Annotators = append([]llm.Profile(nil), s.Annotators...)
	for i, p := range s.Annotators {
		q, err := freezeProfile(p)
		if err != nil {
			return nil, err
		}
		s.Annotators[i] = q
	}
	if s.Reviewer != nil {
		q, err := freezeProfile(*s.Reviewer)
		if err != nil {
			return nil, err
		}
		s.Reviewer = &q
	}
	if _, err := os.Stat(filepath.Join(dir, "run.json")); err == nil {
		return nil, fmt.Errorf("a run already exists in %s: open it instead", dir)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := checkResultNames(s); err != nil {
		return nil, err
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
	seenItems := map[string]bool{}
	for _, id := range s.Items {
		if seenItems[id] {
			return nil, fmt.Errorf("claim %s assigned twice", id)
		}
		seenItems[id] = true
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
		if err := r.Attach(name, c); err != nil {
			return nil, err
		}
	}
	if err := r.saveRun(); err != nil {
		return nil, err
	}
	return r, nil
}

func newRun(dir string, d *Dataset, s Settings) *Run {
	r := &Run{ID: filepath.Base(dir), Dir: dir, Dataset: d, Settings: s, clients: map[string]*llm.Client{},
		anns: map[string]map[string]Annotation{}, failures: map[string]map[string]string{}, reviews: map[string]*Review{},
		verdicts: map[string]Verdict{}, progress: map[string]*Progress{}, onDisk: map[string]map[string]bool{},
		reviewsOnDisk: map[string]*Review{}, fromLog: map[string]bool{}, calls: map[*pendingCall]bool{}}
	r.inputHashes = copyHashes(d.Hashes)
	r.promptHash = sha([]byte(d.AnnotatorSystem()))
	r.curPrompt = r.promptHash
	r.scoringInputs = []string{}
	if d.GoldFile != "" && !d.annotationFiles[d.GoldFile] {
		r.scoringInputs = append(r.scoringInputs, d.GoldFile)
	}
	for _, p := range s.Annotators {
		r.anns[p.Name], r.failures[p.Name], r.onDisk[p.Name] = map[string]Annotation{}, map[string]string{}, map[string]bool{}
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
		Settings      Settings                     `json:"settings"`
		Inputs        map[string]string            `json:"inputs_sha256"`
		Prompt        string                       `json:"annotator_system_prompt_sha256"`
		Builds        []string                     `json:"builds"`
		ScoringInputs []string                     `json:"scoring_only_inputs"`
		Failures      map[string]map[string]string `json:"failures"`
	}
	if !json.Valid(b) {
		return nil, fmt.Errorf("run.json: invalid JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&f); err != nil {
		return nil, fmt.Errorf("run.json: %v", err)
	}
	d, err := datasets(f.Settings.Dataset, f.Settings.DatasetDir)
	if err != nil {
		return nil, fmt.Errorf("its dataset %s: %v", f.Settings.Dataset, err)
	}
	r := newRun(dir, d, f.Settings)
	if err := checkResultNames(f.Settings); err != nil {
		r.loadErrors = append(r.loadErrors, err.Error())
	}
	r.builds = f.Builds
	if f.ScoringInputs != nil {
		r.scoringInputs = append([]string{}, f.ScoringInputs...)
	}
	if len(f.Inputs) == 0 || f.Prompt == "" {
		// A run made before its inputs were pinned: readable, not continued.
		r.inputHashes, r.promptHash = nil, ""
	} else {
		r.inputHashes, r.promptHash = copyHashes(f.Inputs), f.Prompt
	}
	if err := r.checkInputs(); err != nil {
		r.drift = err.Error()
	}
	for name := range r.anns {
		if b, err := os.ReadFile(r.annotationsPath(name)); err == nil {
			if m, err := parseAnnotations(b); err == nil {
				r.anns[name] = m
				for id := range m {
					r.onDisk[name][id] = true
				}
			} else {
				r.loadErrors = append(r.loadErrors, "annotations/"+FileName(name)+".json: "+err.Error())
			}
		} else if !os.IsNotExist(err) {
			r.loadErrors = append(r.loadErrors, err.Error())
		}
		for id, e := range f.Failures[name] {
			if r.anns[name][id] == nil {
				r.failures[name][id] = e
			}
		}
	}
	// Reviews and verdicts are read as strictly as annotations: a review or
	// verdict that names no claim, or a second one for the same claim, makes
	// the run damaged instead of being dropped or silently replacing the
	// first. The person's verdicts are the human judgments a gold is made of.
	if s := f.Settings.Reviewer; s != nil {
		name := "review/model-" + FileName(s.Name) + ".json"
		if b, err := os.ReadFile(r.reviewPath(s.Name)); err == nil {
			var list []*Review
			err := checkedJSON(b, &list)
			if err == nil {
				err = canonicalEach(b, jsonFields(Review{}))
			}
			if err == nil {
				for i, rv := range list {
					switch {
					case rv == nil || strings.TrimSpace(rv.Item) == "":
						r.loadErrors = append(r.loadErrors, fmt.Sprintf("%s: review %d names no claim (invalid review entry)", name, i+1))
					case r.reviews[rv.Item] != nil:
						r.loadErrors = append(r.loadErrors, fmt.Sprintf("%s: two reviews of %s", name, rv.Item))
					default:
						r.reviews[rv.Item], r.reviewsOnDisk[rv.Item] = rv, rv
					}
				}
			} else {
				r.loadErrors = append(r.loadErrors, name+": "+err.Error())
			}
		} else if !os.IsNotExist(err) {
			r.loadErrors = append(r.loadErrors, err.Error())
		}
	}
	if b, err := os.ReadFile(filepath.Join(dir, "review", "human.json")); err == nil {
		var list []*Verdict
		err := checkedJSON(b, &list)
		if err == nil {
			err = canonicalEach(b, jsonFields(Verdict{}))
		}
		if err == nil {
			for i, v := range list {
				if v == nil || strings.TrimSpace(v.Item) == "" {
					r.loadErrors = append(r.loadErrors, fmt.Sprintf("review/human.json: verdict %d names no claim", i+1))
				} else if _, dup := r.verdicts[v.Item]; dup {
					r.loadErrors = append(r.loadErrors, fmt.Sprintf("review/human.json: two verdicts for %s", v.Item))
				} else {
					r.verdicts[v.Item] = *v
				}
			}
		} else {
			r.loadErrors = append(r.loadErrors, "review/human.json: "+err.Error())
		}
	} else if !os.IsNotExist(err) {
		r.loadErrors = append(r.loadErrors, err.Error())
	}
	r.refreshDrift()
	if r.drift == "" {
		// Not for a run that cannot be continued: what it could not write
		// could never be written.
		r.recoverFromLog() // a call log it cannot read damages the run (loadErrors)
	}
	r.recount()
	return r, nil
}

// Attach gives the run the client of an annotator or of its reviewer (they
// are not saved: a reopened run needs them again to continue).
func (r *Run) Attach(name string, c *llm.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c != nil {
		if name != c.Profile.Name {
			delete(r.clients, name)
			return fmt.Errorf("client %s cannot be attached as %s", c.Profile.Name, name)
		}
		if err := r.CheckProfile(c.Profile); err != nil {
			delete(r.clients, name)
			return err
		}
	}
	r.clients[name] = c
	return nil
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
			p.Done, p.Unsaved, p.Failed = 0, 0, 0
			for _, id := range r.Settings.Items {
				switch {
				case r.anns[name][id] != nil && r.onDisk[name][id]:
					p.Done++
				case r.anns[name][id] != nil:
					p.Unsaved++
				case r.failures[name][id] != "":
					p.Failed++
				}
			}
		} else {
			p.Done, p.Unsaved = 0, 0
			for id, rv := range r.reviews {
				if r.reviewsOnDisk[id] == rv {
					p.Done++
				} else {
					p.Unsaved++
				}
			}
		}
	}
}

// saveRun writes run.json. Caller holds r.mu or owns r.
// The identity it writes is always the pinned one, so saving a run whose
// inputs have drifted (a verdict recorded later) misrepresents nothing.
func (r *Run) saveRun() error {
	if b := buildID(); len(r.builds) == 0 || r.builds[len(r.builds)-1] != b {
		r.builds = append(r.builds, b)
	}
	ps := r.progressList(false)
	b, err := json.MarshalIndent(map[string]any{"format": RunFormat, "id": r.ID, "settings": r.Settings,
		"dataset": r.Dataset, "inputs_sha256": r.inputHashes,
		"annotator_system_prompt_sha256": r.promptHash, "builds": r.builds, "scoring_only_inputs": r.scoringInputs,
		"progress": ps, "failures": r.failures, "saved": time.Now().UTC()}, "", " ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(r.Dir, "run.json"), b)
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

// logCall appends a call to calls.jsonl, with a review's letters, which say
// whose annotation each letter labels. A line that cannot be written is
// kept for the next start, and the run asks nothing new meanwhile.
func (r *Run) logCall(role, name, item string, attempt int, a llm.Attempt, letters map[string]string) {
	row := map[string]any{"format": "wag-annotation-call/1", "role": role, "name": name, "item": item, "attempt": attempt, "call": a}
	if letters != nil {
		row["letters"] = letters
	}
	if a.Response != nil && !utf8.ValidString(a.Response.Text) {
		// JSON replaces the bytes that are not UTF-8; the reply as it came
		// is kept beside, so that what is read back is what the run read.
		row["reply_base64"] = base64.StdEncoding.EncodeToString([]byte(a.Response.Text))
	}
	line, err := json.Marshal(row)
	if err != nil {
		// Not with what a call holds; if it were, the line says so.
		line, _ = json.Marshal(map[string]any{"format": "wag-annotation-call/1", "role": role, "name": name, "item": item, "attempt": attempt,
			"unrecorded": err.Error()})
	}
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
	r.pending = append(r.pending, line)
	r.flushCalls()
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
func copyHashes(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Damaged reports whether saved result files could not be read when the run
// was opened. Such a run can be inspected but not written; a caller that
// keeps runs open should open it again once the files are repaired.
func (r *Run) Damaged() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.loadErrors) > 0
}

// checkInputs compares the run's pinned identity with its inputs as they are
// now, leaving out the scoring-only inputs (a gold file no annotator is given).
func (r *Run) checkInputs() error {
	if len(r.loadErrors) > 0 {
		return fmt.Errorf("saved run data cannot be read (%s); inspect or recover those files before continuing", strings.Join(r.loadErrors, "; "))
	}
	for _, p := range r.profiles() {
		if p.Output == "schema" && len(p.Schema) == 0 {
			return fmt.Errorf("profile %s has no pinned schema contents; it can be read but not continued: create a new run", p.Name)
		}
	}
	if r.inputHashes == nil || r.promptHash == "" {
		return fmt.Errorf("the run was made before its inputs were pinned; it can be read but not continued: create a new run")
	}
	var changed []string
	scoring := map[string]bool{}
	for _, k := range r.scoringInputs {
		scoring[k] = true
	}
	if r.Dataset.GoldFile != "" {
		scoring[r.Dataset.GoldFile] = true
	}
	for _, k := range sortedUnion(r.inputHashes, r.Dataset.Hashes) {
		if scoring[k] && !r.Dataset.annotationFiles[k] {
			continue
		}
		if !r.Dataset.sameInput(k, r.inputHashes[k]) {
			changed = append(changed, k)
		}
	}
	if r.promptHash != r.curPrompt && !r.Dataset.samePrompt(r.promptHash) {
		changed = append(changed, "the annotator's system prompt")
	}
	if len(changed) > 0 {
		return fmt.Errorf("the run's inputs have changed since it was made (%s); it can be read but not continued: create a new run", strings.Join(changed, ", "))
	}
	return nil
}

func sortedUnion(a, b map[string]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range []map[string]string{a, b} {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	sort.Strings(out)
	return out
}

// InputDrift says why the run's inputs are no longer those it was made with,
// or "" if they are.
func (r *Run) InputDrift() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshDrift()
	return r.drift
}

// refreshDrift requires r.mu or exclusive ownership of a newly opened run.
func (r *Run) refreshDrift() {
	r.drift = ""
	if err := r.checkInputs(); err != nil {
		r.drift = err.Error()
	}
}

func (r *Run) profiles() []llm.Profile {
	saved := append([]llm.Profile{}, r.Settings.Annotators...)
	if r.Settings.Reviewer != nil {
		saved = append(saved, *r.Settings.Reviewer)
	}
	return saved
}

// freezeProfile resolves the actual output schema and copies mutable values.
// Saved profiles intentionally exclude the API key, which is operational.
func freezeProfile(p llm.Profile) (llm.Profile, error) {
	if p.Output == "schema" && len(p.Schema) == 0 {
		b, err := os.ReadFile(p.SchemaFile)
		if err != nil {
			return p, fmt.Errorf("profile %s schema: %w", p.Name, err)
		}
		p.Schema = b
	}
	b, err := json.Marshal(p)
	if err != nil {
		return p, fmt.Errorf("profile %s: %w", p.Name, err)
	}
	var q llm.Profile
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&q); err != nil {
		return p, err
	}
	return q, nil
}

// CheckProfile refuses a profile that would continue the run's annotator or
// reviewer of the same name with another provider, model, endpoint, prompt,
// schema or parameters. Timeouts, call spacing, retries and the key's source
// may change.
func (r *Run) CheckProfile(p llm.Profile) error {
	q, err := freezeProfile(p)
	if err != nil {
		return err
	}
	for _, s := range r.profiles() {
		if s.Name == p.Name && s.Output == "schema" && len(s.Schema) == 0 {
			return fmt.Errorf("profile %s has no saved schema contents: create a new run", p.Name)
		}
		if s.Name == p.Name && callerOf(s) != callerOf(q) {
			return fmt.Errorf("profile %s is not the one this run was made with (provider, model, endpoint, prompt, schema or parameters differ): create a new run", p.Name)
		}
	}
	return nil
}

func callerOf(p llm.Profile) string {
	b, _ := json.Marshal(struct {
		Provider, BaseURL, System, SchemaFile string
		Command                               []string
		Params                                llm.Params
	}{p.Provider, p.BaseURL, p.System, p.SchemaFile, p.Command, p.Params})
	return string(b)
}

// buildID names the WAG build that is running: its VCS revision, whether the
// tree was modified, and the Go version ("unknown" parts when not recorded).
func buildID() string {
	rev, mod, gov := "unknown", "", runtime.Version()
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					mod = "+modified"
				}
			}
		}
	}
	return rev + mod + " " + gov
}

func (r *Run) begin(ctx context.Context, what string) (context.Context, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshDrift()
	if r.drift != "" {
		return nil, fmt.Errorf("%s", r.drift)
	}
	for _, p := range r.profiles() {
		if c := r.clients[p.Name]; c != nil {
			if c.Profile.Name != p.Name {
				return nil, fmt.Errorf("attached client identity differs for %s", p.Name)
			}
			if err := r.CheckProfile(c.Profile); err != nil {
				return nil, err
			}
		}
	}
	if r.busy != "" {
		return nil, fmt.Errorf("the run is %s already", r.busy)
	}
	if r.needsFlush() {
		if err := r.flushAll(); err != nil {
			return nil, fmt.Errorf("the run cannot write its records, so it asks the models nothing new: %v. "+
				"What it received is kept until they are written: free space or fix the folder, then start again", err)
		}
	}
	ctx, r.cancel = context.WithCancel(ctx)
	r.busy = what
	return ctx, nil
}

// end marks the run idle and writes run.json; it returns the error of a
// record that could not be written while the run was busy, or now.
func (r *Run) end() error {
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
	r.persistRun()
	return r.stoppedByWrite()
}

// finish ends what begin began: the report, then end. A record that could
// not be written is the error returned, before any other.
func (r *Run) finish(err *error) {
	_, _ = r.Report() // the folder is complete even if no one asks for the report
	if werr := r.end(); werr != nil {
		*err = werr
	}
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
func (r *Run) Annotate(ctx context.Context, items []string) (err error) {
	ctx, err = r.begin(ctx, "annotating")
	if err != nil {
		return err
	}
	defer r.finish(&err)
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

// failure says why a call failed, first in plain words if a quota is used
// up (the provider's own message is long, and says it last).
func failure(e *llm.Error, err error) string {
	if e != nil && e.Exhausted() {
		return fmt.Sprintf("quota used up: %s asks to wait %s before the next call, so this stopped; carry on after that. %v",
			e.Provider, e.RetryAfter.Round(time.Minute), err)
	}
	return err.Error()
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
	r.persistRun()
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
		if r.writeErr != nil {
			r.mu.Unlock()
			return
		}
		done := r.anns[name][id] != nil
		before := map[string]Annotation{}
		var prior []string
		for _, x := range order {
			if r.Settings.IndependentFirstPass {
				break
			}
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
			if ctx.Err() != nil || r.hasWriteErr() {
				return // stopped: the claim is not failed, it is still to do
			}
			e, _ := err.(*llm.Error)
			r.fail(name, id, failure(e, err))
			if e != nil && !e.Retryable() && e.Kind != llm.KindBadReply {
				return // a refused key or model, or a quota used up, will not pass on the next claim
			}
			continue
		}
		r.mu.Lock()
		r.anns[name][id] = a
		delete(r.failures[name], id)
		r.progress[name].LastError = ""
		r.persistAnnotations(name) // a failure stops the loop at its next claim
		r.recount()
		r.persistRun()
		r.mu.Unlock()
	}
}

// annotateOne asks one annotator for one claim's annotation; a reply that
// is not an annotation of the claim is sent back once, with the reason.
func (r *Run) annotateOne(ctx context.Context, name string, c *llm.Client, system string, it *Item, before map[string]Annotation, prior []string) (Annotation, error) {
	msgs := []llm.Message{{Role: "user", Content: r.Dataset.AnnotatorPrompt(it, before, prior)}}
	for attempt := 1; attempt <= 2; attempt++ {
		resp, err := r.call(ctx, c, c.Request(system, msgs), func(a llm.Attempt) { r.logCall("annotator", name, it.ID, attempt, a, nil) })
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
func (r *Run) Review(ctx context.Context, items []string, onlyDisagreements bool) (err error) {
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
	ctx, err = r.begin(ctx, "reviewing")
	if err != nil {
		return err
	}
	defer r.finish(&err)
	if len(items) == 0 {
		items = r.Settings.Items
	}
	key := "review:" + rp.Name
	system := r.Dataset.ReviewerSystem()
	var todo []string
	r.mu.Lock()
	for _, id := range items {
		names, anns := r.annotationsOf(id)
		if len(names) < r.toReview() || r.reviews[id] != nil && r.reviews[id].Error == "" {
			continue
		}
		// With one annotator there is nothing to disagree with: the
		// reviewer checks every annotation.
		if onlyDisagreements && len(r.Settings.Annotators) > 1 && len(Disagrees(anns)) == 0 {
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
		if werr := r.stoppedByWrite(); werr != nil {
			r.mu.Unlock()
			return werr
		}
		names, _ := r.annotationsOf(id)
		anns := map[string]Annotation{}
		for _, n := range names {
			anns[n] = r.anns[n][id]
		}
		p.Current = id
		r.mu.Unlock()
		rv, err := r.reviewOne(ctx, rp.Name, c, system, r.Dataset.Item(id), names, anns)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil && r.hasWriteErr() {
			continue // stopped: the claim is still to review (the loop returns the error)
		}
		e, _ := err.(*llm.Error)
		if err != nil {
			rv.Error = failure(e, err)
		}
		r.mu.Lock()
		r.reviews[id] = rv
		if rv.Error != "" {
			p.LastError = rv.Error
		}
		r.persistReviews()
		r.recount()
		r.mu.Unlock()
		if e != nil && !e.Retryable() {
			return err // a refused key or model, or a quota used up, will not pass on the next claim
		}
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

// reviewOne asks the reviewer about one claim. A reply it cannot read is
// recorded in the review's Error; the error returned is the call's.
func (r *Run) reviewOne(ctx context.Context, name string, c *llm.Client, system string, it *Item, names []string, anns map[string]Annotation) (*Review, error) {
	rv := &Review{Item: it.ID, Reviewer: name, Letters: map[string]string{}}
	for i, n := range names {
		rv.Letters[string(Letters[i])] = n
	}
	resp, err := r.call(ctx, c, c.Request(system, []llm.Message{{Role: "user", Content: r.Dataset.ReviewerPrompt(it, names, anns)}}),
		func(a llm.Attempt) { r.logCall("reviewer", name, it.ID, 1, a, rv.Letters) })
	if err != nil {
		rv.Error = err.Error()
		return rv, err
	}
	if err := readReview(rv, resp.Text); err != nil {
		rv.Error, rv.Raw = err.Error(), resp.Text
	}
	return rv, nil
}

// readReview fills a review from the reviewer's reply.
func readReview(rv *Review, text string) error {
	obj, err := jsonObject(text)
	if err != nil {
		return fmt.Errorf("the review cannot be read: %v", err)
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
		return fmt.Errorf("the review cannot be read: %v", err)
	}
	rv.Fields, rv.Adjudicated, rv.Confidence, rv.NeedsHuman, rv.Notes = parsed.Fields, parsed.Adjudicated, parsed.Confidence, parsed.NeedsHuman, parsed.Notes
	for f, ch := range rv.Fields {
		ch.Annotator = rv.Letters[strings.ToUpper(strings.TrimSpace(ch.Choice))]
		rv.Fields[f] = ch
	}
	if rv.Adjudicated != nil {
		rv.Adjudicated["kialo_claim_id"] = rv.Item
	}
	return nil
}

// toReview is how many annotations a claim needs to be reviewed: two to
// compare, or the one of a run with one annotator, to check.
func (r *Run) toReview() int {
	if len(r.Settings.Annotators) == 1 {
		return 1
	}
	return 2
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
	if len(r.loadErrors) > 0 {
		return fmt.Errorf("saved run data is unreadable; recover it before writing verdicts")
	}
	r.refreshDrift()
	v.InputsSHA256, v.InputDrift, v.AssessmentBuild = copyHashes(r.Dataset.Hashes), r.drift, buildID()
	prev, had := r.verdicts[v.Item]
	if v.Choice == "" && v.Note == "" {
		delete(r.verdicts, v.Item)
	} else {
		r.verdicts[v.Item] = v
	}
	// A verdict that is not written is not kept: the run holds what its
	// file holds, and the person is told to give it again.
	undo := func(err error) error {
		if had {
			r.verdicts[v.Item] = prev
		} else {
			delete(r.verdicts, v.Item)
		}
		return fmt.Errorf("the verdict is not saved: %v", err)
	}
	var list []Verdict
	for _, id := range r.Dataset.IDs() {
		if x, ok := r.verdicts[id]; ok {
			list = append(list, x)
		}
	}
	b, err := json.MarshalIndent(list, "", " ")
	if err != nil {
		return undo(err)
	}
	if err := writeFile(filepath.Join(r.Dir, "review", "human.json"), b); err != nil {
		return undo(err)
	}
	return nil
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
	Reviewable int         `json:"reviewable"` // claims with enough annotations to review (Run.Review)
	// InputDrift says why the run can be read but not continued ("" if it can be).
	InputDrift string `json:"input_drift,omitempty"`
	// WriteError is a record the run could not write: it asks the models
	// nothing new until its next start has written what the files lack
	// (Unsaved), which can also be what a reopened run took from calls.jsonl.
	WriteError string   `json:"write_error,omitempty"`
	Unsaved    []string `json:"unsaved,omitempty"`
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
	r.refreshDrift()
	s := Status{ID: r.ID, Dir: r.Dir, Dataset: r.Dataset, Settings: r.Settings, Busy: r.busy, Reviewed: len(r.reviews), Verdicts: len(r.verdicts), InputDrift: r.drift,
		Unsaved: r.unsaved()}
	if r.writeErr != nil {
		s.WriteError = r.writeErr.Error()
	}
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
		if len(names) >= r.toReview() {
			s.Reviewable++
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
