package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// syncs records the folders SyncDir is asked to sync, and fails for fail.
func syncs(t *testing.T, fail string) *[]string {
	t.Helper()
	var got []string
	old := SyncDir
	SyncDir = func(dir string) error {
		got = append(got, dir)
		if dir == fail {
			return errors.New("I/O error")
		}
		return old(dir)
	}
	t.Cleanup(func() { SyncDir = old })
	return &got
}

func TestWriteFileReplacesTheFileAndSyncsItsFolder(t *testing.T) {
	dir := t.TempDir()
	got := syncs(t, "")
	path := filepath.Join(dir, "run.json")
	for _, content := range []string{"old", "new"} {
		if err := WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if b, err := os.ReadFile(path); err != nil || string(b) != content {
			t.Fatalf("%q %v", b, err)
		}
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("the temporary file is left")
	}
	if want := []string{dir, dir}; !reflect.DeepEqual(*got, want) {
		t.Fatalf("synced %v, want %v", *got, want)
	}
}

// A temporary file a crash left keeps its mode through O_TRUNC; the file
// written gets the mode asked for (a key file must not become readable).
func TestWriteFileGivesTheModeAskedFor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are Unix")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.json")
	if err := os.WriteFile(path+".tmp", []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(path); err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v %v", fi.Mode(), err)
	}
}

func TestWriteFileLeavesWhatIsNotItsOwn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+".tmp", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(path, []byte("new"), 0o644); err == nil {
		t.Fatal("written through a folder")
	}
	if fi, err := os.Stat(path + ".tmp"); err != nil || !fi.IsDir() {
		t.Fatal("the folder in the way was removed")
	}
	if b, _ := os.ReadFile(path); string(b) != "old" {
		t.Fatalf("the file is %q", b)
	}
}

// When only the folder cannot be synced, the file holds the new bytes and
// the error says the folder: the write is not reported as durable.
func TestWriteFileReportsAFolderThatCannotBeSynced(t *testing.T) {
	dir := t.TempDir()
	syncs(t, dir)
	path := filepath.Join(dir, "a.json")
	err := WriteFile(path, []byte("new"), 0o644)
	var fe *FolderError
	if !errors.As(err, &fe) || fe.Dir != dir {
		t.Fatalf("%v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "new" {
		t.Fatalf("the file is %q", b)
	}
}

func TestMkdirAllSyncsTheFolderOfEachNewFolder(t *testing.T) {
	root := t.TempDir()
	got := syncs(t, "")
	dir := filepath.Join(root, "a", "b", "c")
	if err := MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatal(err)
	}
	if want := []string{root, filepath.Join(root, "a"), filepath.Join(root, "a", "b")}; !reflect.DeepEqual(*got, want) {
		t.Fatalf("synced %v, want %v", *got, want)
	}
	*got = nil
	if err := MkdirAll(dir, 0o755); err != nil || len(*got) != 0 {
		t.Fatalf("an existing folder: %v, synced %v", err, *got)
	}
	syncs(t, filepath.Join(root, "a"))
	var fe *FolderError
	if err := MkdirAll(filepath.Join(root, "a", "x"), 0o755); !errors.As(err, &fe) || fe.Dir != filepath.Join(root, "a") {
		t.Fatalf("%v", err)
	}
	// A file in the way is MkdirAll's error.
	if err := os.WriteFile(filepath.Join(root, "f"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MkdirAll(filepath.Join(root, "f", "g"), 0o755); err == nil {
		t.Fatal("a folder made under a file")
	}
}

func TestSyncDirSyncsAFolder(t *testing.T) {
	if err := SyncDir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := SyncDir(filepath.Join(t.TempDir(), "missing")); err == nil {
			t.Fatal("a missing folder synced")
		}
	}
}

// A file whose bytes cannot be synced is not renamed into place: the old
// file stays, and the temporary one is removed.
func TestAFileThatCannotBeSyncedIsNotRenamedIntoPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.json")
	if err := WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := syncFile
	synced := 0
	syncFile = func(f *os.File) error { synced++; return errors.New("I/O error") }
	defer func() { syncFile = old }()
	if err := WriteFile(path, []byte("new"), 0o644); err == nil || synced != 1 {
		t.Fatalf("%v, %d syncs", err, synced)
	}
	if b, _ := os.ReadFile(path); string(b) != "old" {
		t.Fatalf("the file is %q", b)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("the temporary file is left")
	}
}
