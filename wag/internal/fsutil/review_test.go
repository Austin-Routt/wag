package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Follow-ups to the independent review of 18ee7bc (R2, R6).

// failing makes SyncDir fail for the folders in fail until heal, and records
// the folders it syncs.
func failing(t *testing.T, fail ...string) (heal func(), synced *[]string) {
	t.Helper()
	old := SyncDir
	broken := map[string]bool{}
	for _, d := range fail {
		broken[d] = true
	}
	var got []string
	SyncDir = func(dir string) error {
		got = append(got, dir)
		if broken[dir] {
			return errors.New("I/O error")
		}
		return old(dir)
	}
	t.Cleanup(func() { SyncDir = old })
	return func() { broken = map[string]bool{} }, &got
}

func has(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// R2: MkdirAll called again after a folder's sync failed syncs it, rather
// than finding the folder made and reporting success.
func TestMkdirAllAgainSyncsAFolderWhoseSyncFailed(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "a")
	heal, synced := failing(t, root)
	var fe *FolderError
	if err := MkdirAll(dir, 0o755); !errors.As(err, &fe) || fe.Dir != root {
		t.Fatalf("the first MkdirAll: %v", err)
	}
	if err := MkdirAll(dir, 0o755); err == nil {
		t.Fatal("the folder's sync still fails, and MkdirAll reported success")
	}
	heal()
	*synced = nil
	if err := MkdirAll(dir, 0o755); err != nil || !has(*synced, root) {
		t.Fatalf("%v; synced %v", err, *synced)
	}
	*synced = nil
	if err := MkdirAll(dir, 0o755); err != nil || len(*synced) != 0 {
		t.Fatalf("once synced, the folder is not synced again: %v %v", err, *synced)
	}
}

// R2: a file written below a folder whose sync failed syncs that folder too.
func TestAWriteBelowAFolderWhoseSyncFailedSyncsIt(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "games", "g1")
	heal, synced := failing(t, filepath.Join(root, "games"))
	if err := MkdirAll(dir, 0o755); err == nil {
		t.Fatal("fixture: the sync did not fail")
	}
	if err := WriteFile(filepath.Join(dir, "export.json"), []byte("{}"), 0o644); err == nil {
		t.Fatal("written below a folder whose sync fails, and reported done")
	}
	heal()
	*synced = nil
	if err := WriteFile(filepath.Join(dir, "export.json"), []byte("{}"), 0o644); err != nil || !has(*synced, dir) || !has(*synced, filepath.Join(root, "games")) {
		t.Fatalf("%v; synced %v", err, *synced)
	}
	// A write beside it, not below it, does not depend on it.
	other := t.TempDir()
	heal2, _ := failing(t, filepath.Join(root, "games"))
	defer heal2()
	if err := MkdirAll(filepath.Join(root, "games", "g2"), 0o755); err == nil {
		t.Fatal("fixture: the sync did not fail")
	}
	if err := WriteFile(filepath.Join(other, "x.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("a write elsewhere: %v", err)
	}
}

// R6: a link at the temporary path is not followed: the file it names is
// untouched, the link is left, and the write fails.
func TestALinkAtTheTemporaryPathIsLeftAlone(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("links need privileges on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "elsewhere.txt")
	if err := os.WriteFile(target, []byte("theirs"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "export.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path+".tmp"); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(path, []byte("new"), 0o644); err == nil {
		t.Fatal("written through a link")
	}
	if b, _ := os.ReadFile(target); string(b) != "theirs" {
		t.Fatalf("the link's target holds %q", b)
	}
	if fi, err := os.Lstat(path + ".tmp"); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("the link was removed")
	}
	if fi, err := os.Lstat(path); err != nil || !fi.Mode().IsRegular() {
		t.Fatal("the file is no longer a file")
	}
	if b, _ := os.ReadFile(path); string(b) != "old" {
		t.Fatalf("the file holds %q", b)
	}
}

// On Windows a rename fails for a moment while another program has the file
// open: such a failure is tried again; another is not.
func TestATransientRenameFailureIsTriedAgain(t *testing.T) {
	oldRename, oldTransient := renameFile, transient
	defer func() { renameFile, transient = oldRename, oldTransient }()
	busy := errors.New("the file is in use")
	transient = func(err error) bool { return errors.Is(err, busy) }
	dir := t.TempDir()
	path := filepath.Join(dir, "a.json")
	tries := 0
	renameFile = func(from, to string) error {
		tries++
		if tries < 3 {
			return busy
		}
		return os.Rename(from, to)
	}
	if err := WriteFile(path, []byte("new"), 0o644); err != nil || tries != 3 {
		t.Fatalf("%v after %d tries", err, tries)
	}
	tries = 0
	renameFile = func(from, to string) error { tries++; return busy }
	if err := WriteFile(path, []byte("newer"), 0o644); !errors.Is(err, busy) || tries != 7 {
		t.Fatalf("%v after %d tries", err, tries)
	}
	if b, _ := os.ReadFile(path); string(b) != "new" {
		t.Fatalf("the file holds %q", b)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("the temporary file is left")
	}
	tries = 0
	other := errors.New("no such device")
	renameFile = func(from, to string) error { tries++; return other }
	if err := WriteFile(path, []byte("newest"), 0o644); !errors.Is(err, other) || tries != 1 {
		t.Fatalf("%v after %d tries", err, tries)
	}
}

func TestWithin(t *testing.T) {
	sep := string(filepath.Separator)
	a := sep + "a"
	for _, c := range []struct {
		dir, folder string
		want        bool
	}{
		{a + sep + "b", a, true},
		{a + sep + "b" + sep + "c", a, true},
		{a, a, false},
		{sep + "ab", a, false},
		{a, a + sep + "b", false},
		{sep + "x", a, false},
	} {
		if got := within(c.dir, c.folder); got != c.want {
			t.Errorf("within(%q, %q) = %v", c.dir, c.folder, got)
		}
	}
}
