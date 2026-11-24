//go:build unix

package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// A file system that cannot sync folders answers EINVAL (some network
// shares): the folder's sync counts as done, as the file's own sync does.
func TestAFolderSyncTheFileSystemRefusesCountsAsDone(t *testing.T) {
	old := syncFile
	syncFile = func(f *os.File) error {
		if fi, err := f.Stat(); err == nil && fi.IsDir() {
			return syscall.EINVAL
		}
		return f.Sync()
	}
	defer func() { syncFile = old }()
	dir := t.TempDir()
	if err := SyncDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(filepath.Join(dir, "a.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Another error is not.
	syncFile = func(f *os.File) error {
		if fi, err := f.Stat(); err == nil && fi.IsDir() {
			return errors.New("I/O error")
		}
		return f.Sync()
	}
	if err := SyncDir(dir); err == nil {
		t.Fatal("an I/O error counted as done")
	}
}

// R7: the umask is honoured: a file is made with the mode asked for less the
// umask, as os.WriteFile makes it.
func TestTheUmaskIsHonoured(t *testing.T) {
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)
	dir := t.TempDir()
	if err := WriteFile(filepath.Join(dir, "a.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	a, _ := os.Stat(filepath.Join(dir, "a.json"))
	b, _ := os.Stat(filepath.Join(dir, "b.json"))
	if a.Mode().Perm() != 0o600 || a.Mode().Perm() != b.Mode().Perm() {
		t.Fatalf("WriteFile: %v; os.WriteFile: %v", a.Mode(), b.Mode())
	}
}
