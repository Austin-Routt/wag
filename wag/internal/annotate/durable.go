package annotate

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/wag/internal/llm"
)

// A run records every model reply twice: the call in calls.jsonl, its result
// (an annotation, a review) in the result's file; run.json holds the progress
// and the failures. When a record cannot be written:
//
//   - the run asks nothing new: the annotators and the reviewer finish the
//     call under way, and its own retries stop;
//   - Annotate and Review return the error, and Status shows it (WriteError)
//     with what the files lack (Unsaved);
//   - what was received is kept in memory, counted as unsaved, not as done;
//   - the next Annotate or Review writes every record before it asks
//     anything, and does not start until it can.
//
// A reply that reached calls.jsonl when its result's file could not be
// written is taken from calls.jsonl when the run is opened again, so the
// model is not asked again.

// errNotRecorded stops a claim's next attempt once a record of the run
// could not be written.
var errNotRecorded = errors.New("a record of the run could not be written")

// writeFailed notes a record that could not be written: the first such
// error stands until flushAll writes everything. The calls waiting for
// their turn or for a retry stop at once; an attempt with a provider
// finishes, and is logged (call). Caller holds r.mu.
func (r *Run) writeFailed(what string, err error) {
	if r.writeErr == nil {
		r.writeErr = fmt.Errorf("%s: %w", what, err)
		for pc := range r.calls {
			if !pc.sending {
				pc.cancel()
			}
		}
	}
}

// pendingCall is a model call of the run under way; sending says that one
// of its attempts is with the provider.
type pendingCall struct {
	cancel  context.CancelFunc
	sending bool
}

// call makes one model call for the run; log records each attempt. No
// attempt, the first or a retry, is sent once a record of the run could not
// be written.
func (r *Run) call(ctx context.Context, c *llm.Client, req llm.Request, log func(llm.Attempt)) (*llm.Response, error) {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	pc := &pendingCall{cancel: cancel}
	r.mu.Lock()
	r.calls[pc] = true
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		delete(r.calls, pc)
		r.mu.Unlock()
	}()
	gate := func() error {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.writeErr != nil {
			return errNotRecorded
		}
		pc.sending = true
		return nil
	}
	return c.Complete(llm.WithGate(cctx, gate), req, func(a llm.Attempt) {
		log(a)
		r.mu.Lock()
		pc.sending = false
		failed := r.writeErr != nil
		r.mu.Unlock()
		if failed {
			cancel() // its retries stop
		}
	})
}

// stoppedByWrite is what Annotate and Review return after a record could
// not be written. Caller holds r.mu.
func (r *Run) stoppedByWrite() error {
	if r.writeErr == nil {
		return nil
	}
	return fmt.Errorf("the run could not write %v. It asked the models nothing new after that, and keeps what it received "+
		"until its records are written: the next start writes them first", r.writeErr)
}

func (r *Run) hasWriteErr() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.writeErr != nil
}

// persistAnnotations writes an annotator's file and notes what it holds.
// Caller holds r.mu.
func (r *Run) persistAnnotations(name string) {
	if err := r.saveAnnotations(name); err != nil {
		r.writeFailed("annotations/"+FileName(name)+".json", err)
		return
	}
	held := map[string]bool{}
	for id := range r.anns[name] {
		held[id] = true
	}
	r.onDisk[name] = held
}

// persistReviews writes the reviewer's file and notes which review of each
// claim it holds (a review that failed is replaced when the claim is
// reviewed again). Caller holds r.mu.
func (r *Run) persistReviews() {
	if err := r.saveReviews(); err != nil {
		r.writeFailed("review/model-"+FileName(r.Settings.Reviewer.Name)+".json", err)
		return
	}
	held := map[string]*Review{}
	for id, rv := range r.reviews {
		held[id] = rv
	}
	r.reviewsOnDisk = held
}

// persistRun writes run.json. Caller holds r.mu or owns r.
func (r *Run) persistRun() {
	if err := r.saveRun(); err != nil {
		r.runDirty = true
		r.writeFailed("run.json", err)
		return
	}
	r.runDirty = false
}

// flushCalls appends the call lines not written yet. Caller holds r.mu.
func (r *Run) flushCalls() {
	if len(r.pending) == 0 {
		return
	}
	if err := appendLines(filepath.Join(r.Dir, "calls.jsonl"), r.pending); err != nil {
		r.writeFailed("calls.jsonl", err)
		return
	}
	r.pending = nil
}

// needsFlush reports whether the files lack a record the run holds, or a
// write failed. Caller holds r.mu.
func (r *Run) needsFlush() bool {
	return r.writeErr != nil || r.runDirty || len(r.pending) > 0 || len(r.unsaved()) > 0
}

// flushAll writes every record the files lack; nil when they hold them all,
// which ends a write failure. Caller holds r.mu.
func (r *Run) flushAll() error {
	r.writeErr = nil
	r.flushCalls()
	for _, p := range r.Settings.Annotators {
		if len(r.unsavedAnnotations(p.Name)) > 0 {
			r.persistAnnotations(p.Name)
		}
	}
	if r.Settings.Reviewer != nil && len(r.unsavedReviews()) > 0 {
		r.persistReviews()
	}
	r.recount()
	r.persistRun()
	return r.writeErr
}

// Flush writes the records the run could not write before, if it can now.
func (r *Run) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.needsFlush() {
		return nil
	}
	if r.busy != "" {
		return fmt.Errorf("the run is %s", r.busy)
	}
	return r.flushAll()
}

// unsavedAnnotations lists, in the dataset's order, the claims whose
// annotation by name the run holds and its file does not. Caller holds r.mu.
func (r *Run) unsavedAnnotations(name string) []string {
	var out []string
	for _, id := range r.Dataset.IDs() {
		if r.anns[name][id] != nil && !r.onDisk[name][id] {
			out = append(out, id)
		}
	}
	return out
}

// unsavedReviews lists the claims whose review the run holds and the
// reviewer's file does not. Caller holds r.mu.
func (r *Run) unsavedReviews() []string {
	var out []string
	for _, id := range r.Dataset.IDs() {
		if rv := r.reviews[id]; rv != nil && r.reviewsOnDisk[id] != rv {
			out = append(out, id)
		}
	}
	return out
}

// unsaved says what the run holds that its files do not. Caller holds r.mu.
func (r *Run) unsaved() []string {
	var out []string
	mark := func(key string, ids []string) string {
		var xs []string
		for _, id := range ids {
			if r.fromLog[key+id] {
				id += " (taken from calls.jsonl)"
			}
			xs = append(xs, id)
		}
		return strings.Join(xs, ", ")
	}
	for _, p := range r.Settings.Annotators {
		if ids := r.unsavedAnnotations(p.Name); len(ids) > 0 {
			out = append(out, fmt.Sprintf("annotations of %s: %s", p.Name, mark("annotator\x00"+p.Name+"\x00", ids)))
		}
	}
	if rp := r.Settings.Reviewer; rp != nil {
		if ids := r.unsavedReviews(); len(ids) > 0 {
			out = append(out, fmt.Sprintf("reviews of %s: %s", rp.Name, mark("reviewer\x00", ids)))
		}
	}
	if n := len(r.pending); n == 1 {
		out = append(out, "1 line of calls.jsonl")
	} else if n > 1 {
		out = append(out, fmt.Sprintf("%d lines of calls.jsonl", n))
	}
	if r.runDirty {
		out = append(out, "run.json (progress and failures)")
	}
	return out
}

// recoverFromLog takes from calls.jsonl the replies whose results the files
// lack: for each claim of the run with no annotation by an annotator, or no
// review, the last reply that reads as one. They are held as unsaved, and
// the next start writes them instead of asking again. A line that cannot be
// read (a crash mid-write) is skipped, as the report skips it; a review is
// taken only from a line that names its letters. A log that cannot be read
// through makes the run damaged, and nothing is taken from it: continuing
// could ask again for what it holds. Caller owns r.
func (r *Run) recoverFromLog() {
	if !r.lacksResults() {
		return
	}
	f, err := os.Open(filepath.Join(r.Dir, "calls.jsonl"))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		r.loadErrors = append(r.loadErrors, "calls.jsonl: "+err.Error())
		return
	}
	defer f.Close()
	inRun := map[string]bool{}
	for _, id := range r.Settings.Items {
		inRun[id] = true
	}
	anns := map[string]map[string]Annotation{}
	reviews := map[string]*Review{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 256<<20)
	for sc.Scan() {
		var row struct { // a call line, without the request (the long part)
			Role    string            `json:"role"`
			Name    string            `json:"name"`
			Item    string            `json:"item"`
			Letters map[string]string `json:"letters"`
			Reply   string            `json:"reply_base64"`
			Call    struct {
				Error    json.RawMessage `json:"error"`
				Response *llm.Response   `json:"response"`
			} `json:"call"`
		}
		if json.Unmarshal(sc.Bytes(), &row) != nil || !inRun[row.Item] || row.Call.Response == nil ||
			len(row.Call.Error) > 0 && string(row.Call.Error) != "null" {
			continue
		}
		if row.Reply != "" { // the reply as it came, not as JSON could hold it
			b, err := base64.StdEncoding.DecodeString(row.Reply)
			if err != nil {
				continue
			}
			row.Call.Response.Text = string(b)
		}
		switch {
		case row.Role == "annotator" && r.anns[row.Name] != nil:
			if a, err := parseAnnotation(row.Call.Response.Text, row.Item); err == nil {
				if anns[row.Name] == nil {
					anns[row.Name] = map[string]Annotation{}
				}
				anns[row.Name][row.Item] = a
			}
		case row.Role == "reviewer" && r.Settings.Reviewer != nil && row.Name == r.Settings.Reviewer.Name && len(row.Letters) > 0:
			rv := &Review{Item: row.Item, Reviewer: row.Name, Letters: row.Letters}
			if readReview(rv, row.Call.Response.Text) == nil {
				reviews[row.Item] = rv
			}
		}
	}
	if err := sc.Err(); err != nil {
		r.loadErrors = append(r.loadErrors, "calls.jsonl: "+err.Error())
		return
	}
	for name, m := range anns {
		for id, a := range m {
			if r.anns[name][id] == nil {
				r.anns[name][id] = a
				delete(r.failures[name], id)
				r.fromLog["annotator\x00"+name+"\x00"+id] = true
			}
		}
	}
	for id, rv := range reviews {
		if old := r.reviews[id]; old == nil || old.Error != "" {
			r.reviews[id] = rv
			r.fromLog["reviewer\x00"+id] = true
		}
	}
}

// lacksResults reports whether a claim of the run has no annotation by an
// annotator, or no review that could be read (when the run has a reviewer):
// only then can calls.jsonl hold a result the files lack. Caller owns r.
func (r *Run) lacksResults() bool {
	for _, id := range r.Settings.Items {
		for _, p := range r.Settings.Annotators {
			if r.anns[p.Name][id] == nil {
				return true
			}
		}
		if r.Settings.Reviewer != nil && (r.reviews[id] == nil || r.reviews[id].Error != "") {
			return true
		}
	}
	return false
}

// appendFile is the call log as appendLines uses it (a test fakes a disk
// that fails part way).
type appendFile interface {
	Stat() (os.FileInfo, error)
	ReadAt([]byte, int64) (int, error)
	Write([]byte) (int, error)
	Truncate(int64) error
	Sync() error
	Close() error
}

var openAppend = func(path string) (appendFile, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
}

// appendLines appends lines to a log, all of them or none: a write that
// fails part way is cut off, and the lines are written again whole. Half a
// line already at the end (a crash, or a failed write that could not be cut
// off) is ended first, so that it stays a line of its own, which readers
// skip, and joins no other. When it returns nil the lines' bytes are synced
// (not the folder's entry for a new log); an error from closing the file
// after that is not a loss.
func appendLines(path string, lines [][]byte) error {
	f, err := openAppend(path)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	var buf []byte
	if size := fi.Size(); size > 0 {
		last := make([]byte, 1)
		if _, err := f.ReadAt(last, size-1); err != nil {
			f.Close()
			return err
		}
		if last[0] != '\n' {
			buf = append(buf, '\n')
		}
	}
	for _, l := range lines {
		buf = append(append(buf, l...), '\n')
	}
	_, err = f.Write(buf)
	if err == nil {
		err = f.Sync()
	}
	if err != nil {
		if terr := f.Truncate(fi.Size()); terr != nil {
			err = fmt.Errorf("%v; the lines written in part could not be cut off (%v)", err, terr)
		}
		f.Close()
		return err
	}
	f.Close()
	return nil
}

// writeFile writes through a temporary file, so a reader never sees half a
// file; when it returns nil the file's bytes are synced (not the folder's
// entry for the rename). A temporary file it made is removed if it fails;
// anything else at that path is left.
func writeFile(path string, b []byte) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
	}
	return err
}
