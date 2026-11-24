// Package fsutil writes files so that a crash or a power loss leaves a file
// whole, old or new, and so that what a write reported as done survives a
// power loss, as far as the operating system allows.
//
// A file's bytes are synced before it is renamed into place, and the folder
// is synced after, so that the rename, and a new file's name, are on the disk
// too: a synced file whose folder entry is not may still vanish at a power
// loss. A folder whose sync failed is remembered, and synced again before a
// later write or MkdirAll in it, or below it, reports success. Windows cannot
// sync a folder (its file system journals the entries itself), and some file
// systems refuse to (some network shares); there a folder's sync does nothing.
package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// SyncDir makes the entries of the folder dir durable: the files created,
// renamed or removed in it. Where folders cannot be synced (Windows; a file
// system that answers EINVAL, EBADF or EOPNOTSUPP, as PostgreSQL also allows)
// it does nothing. It is a variable so that tests can make it fail.
var SyncDir = func(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = syncFile(f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil && cannotSyncFolders(err) {
		return nil
	}
	return err
}

// syncFile syncs a file's bytes; a variable so that tests can make it fail.
var syncFile = func(f *os.File) error { return f.Sync() }

// renameFile and transient are the rename and what makes it worth trying
// again (rename); variables so that tests can make them fail.
var (
	renameFile = os.Rename
	transient  = transientRename
)

// pending are the folders whose sync failed: their entries may not survive a
// power loss until a sync of them succeeds.
var (
	pendingMu sync.Mutex
	pending   = map[string]bool{}
)

// SyncFolder syncs the folder dir, and the folders at and above it whose sync
// failed before. A failure is remembered, and tried again by the next write,
// MkdirAll or SyncFolder there.
func SyncFolder(dir string) error { return settle(dir, true) }

// settle syncs the folders at and above dir whose sync failed before, and dir
// itself if self. (dir exists, so the folders above it do too.)
func settle(dir string, self bool) error {
	abs := absolute(dir)
	pendingMu.Lock()
	defer pendingMu.Unlock()
	var todo []string
	if self {
		todo = append(todo, abs)
	}
	var above []string
	for d := range pending {
		if d != abs && within(abs, d) || d == abs && !self {
			above = append(above, d)
		}
	}
	sort.Strings(above)
	for _, d := range append(todo, above...) {
		if err := SyncDir(d); err != nil {
			pending[d] = true
			return &FolderError{Dir: d, Err: err}
		}
		delete(pending, d)
	}
	return nil
}

func absolute(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		return abs
	}
	return filepath.Clean(dir)
}

// within reports whether dir is in folder, at any depth.
func within(dir, folder string) bool {
	rel, err := filepath.Rel(folder, dir)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// WriteFile writes b to path through a temporary file, path + ".tmp": the
// temporary file is made new (never through a link), written and synced,
// renamed over path, and the folder is synced (SyncFolder). A reader never
// sees half a file. The file's mode is perm less the umask. A file left at
// the temporary path by a crash is replaced; anything else there (a folder,
// a link) is not this write's: it is left, and the write fails (tests use a
// folder there to make writes fail). The temporary file is removed if the
// write fails before the rename. If only a folder's sync fails, path holds
// the new bytes, but they may not survive a power loss, and the error says
// so (a *FolderError).
func WriteFile(path string, b []byte, perm fs.FileMode) error {
	tmp := path + ".tmp"
	if fi, err := os.Lstat(tmp); err == nil {
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("%s is in the way: it is not a file this write made", tmp)
		}
		if err := os.Remove(tmp); err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	// O_EXCL: a new file, never one a link names.
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err == nil {
		err = syncFile(f)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return SyncFolder(filepath.Dir(path))
}

// rename moves the temporary file into place. On Windows a virus scanner or
// an indexer that has the file open makes the move fail for a moment, so a
// transient failure is tried again, for about half a second in all.
func rename(from, to string) error {
	wait := 10 * time.Millisecond
	for i := 0; ; i++ {
		err := renameFile(from, to)
		if err == nil || i == 6 || !transient(err) {
			return err
		}
		time.Sleep(wait)
		wait *= 2
	}
}

// FolderError is a folder that could not be synced after a file in it was
// written or created: the file is there, but may not survive a power loss.
type FolderError struct {
	Dir string
	Err error
}

func (e *FolderError) Error() string {
	return "the folder " + e.Dir + " could not be synced: " + e.Err.Error()
}

func (e *FolderError) Unwrap() error { return e.Err }

// MkdirAll makes the folder dir and any missing parents, syncs the folder
// each new one was made in, and syncs again the folders at and above dir
// whose sync failed before, so that the folders survive a power loss.
func MkdirAll(dir string, perm fs.FileMode) error {
	var made []string
	for d := filepath.Clean(dir); ; d = filepath.Dir(d) {
		if _, err := os.Stat(d); err == nil {
			break
		} else if !errors.Is(err, fs.ErrNotExist) {
			break // MkdirAll reports it
		}
		made = append(made, d)
		if filepath.Dir(d) == d {
			break
		}
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return err
	}
	for i := len(made) - 1; i >= 0; i-- {
		if err := SyncFolder(filepath.Dir(made[i])); err != nil {
			return err
		}
	}
	return settle(dir, false)
}
