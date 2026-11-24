package annotate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/fsutil"
)

// A record counts as written once its folder is synced too (fsutil): a
// folder that cannot be synced is a write failure like any other.

// failSync makes fsutil.SyncDir fail for dir until heal is called, and
// counts the folders synced.
func failSync(t *testing.T, dir string) (heal func(), synced *[]string) {
	t.Helper()
	old := fsutil.SyncDir
	fail := true
	var got []string
	fsutil.SyncDir = func(d string) error {
		got = append(got, d)
		if fail && d == dir {
			return errors.New("I/O error")
		}
		return old(d)
	}
	t.Cleanup(func() { fsutil.SyncDir = old })
	return func() { fail = false }, &got
}

func TestAFolderThatCannotBeSyncedIsAWriteFailure(t *testing.T) {
	r, cp, items := soloRun(t, 2)
	heal, _ := failSync(t, filepath.Join(r.Dir, "annotations"))
	err := r.Annotate(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "could not be synced") || cp.total() != 1 {
		t.Fatalf("%v; %d calls", err, cp.total())
	}
	st := r.Status()
	if st.WriteError == "" || st.Progress[0].Done != 0 || st.Progress[0].Unsaved != 1 {
		t.Fatalf("status: %q %+v", st.WriteError, st.Progress[0])
	}
	heal()
	if err := r.Annotate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if cp.total() != 2 || cp.of(items[0]) != 1 {
		t.Fatalf("%d calls, %d for %s", cp.total(), cp.of(items[0]), items[0])
	}
	if got := savedIDs(t, r.annotationsPath("solo")); len(got) != 2 {
		t.Fatalf("the file holds %v", got)
	}
}

// A new call log is synced with its folder: if the folder cannot be, the
// lines are cut off, as for any failed write. A log that has lines is not
// a new entry of its folder.
func TestANewCallLogIsSyncedWithItsFolder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "calls.jsonl")
	heal, synced := failSync(t, dir)
	if err := appendLines(path, [][]byte{[]byte(`{"a":1}`)}); err == nil || !strings.Contains(err.Error(), "I/O error") {
		t.Fatalf("%v", err)
	}
	if b, err := os.ReadFile(path); err != nil || len(b) != 0 {
		t.Fatalf("the lines stay: %q %v", b, err)
	}
	heal()
	if err := appendLines(path, [][]byte{[]byte(`{"a":1}`)}); err != nil {
		t.Fatal(err)
	}
	*synced = nil
	if err := appendLines(path, [][]byte{[]byte(`{"b":2}`)}); err != nil || len(*synced) != 0 {
		t.Fatalf("%v; synced %v", err, *synced)
	}
	if b, _ := os.ReadFile(path); string(b) != "{\"a\":1}\n{\"b\":2}\n" {
		t.Fatalf("%q", b)
	}
}

// A new run's folders are synced into the folder they are made in.
func TestANewRunsFoldersAreSynced(t *testing.T) {
	_, synced := failSync(t, "")
	r, _, _ := soloRun(t, 1)
	root := filepath.Dir(r.Dir)
	want := map[string]bool{root: false, r.Dir: false}
	for _, d := range *synced {
		if _, ok := want[d]; ok {
			want[d] = true
		}
	}
	for d, ok := range want {
		if !ok {
			t.Fatalf("%s was not synced: %v", d, *synced)
		}
	}
}
