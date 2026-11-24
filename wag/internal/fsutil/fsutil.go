// Package fsutil writes files so that a crash or a power loss leaves a file
// whole, old or new, and so that what a write reported as done survives a
// power loss, as far as the operating system allows.
//
// A file's bytes are synced before it is renamed into place, and the folder
// is synced after, so that the rename, and a new file's name, are on the disk
// too: a synced file whose folder entry is not may still vanish at a power
// loss. Windows cannot sync a folder (its file system journals the entries
// itself), so there SyncDir does nothing.
package fsutil

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// SyncDir makes the entries of the folder dir durable: the files created,
// renamed or removed in it. It is a variable so that tests can make it fail.
var SyncDir = func(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = f.Sync()
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// syncFile syncs a file's bytes; a variable so that tests can make it fail.
var syncFile = func(f *os.File) error { return f.Sync() }

// WriteFile writes b to path through a temporary file, path + ".tmp": the
// temporary file is written and synced, renamed over path, and the folder is
// synced. A reader never sees half a file. The temporary file is removed if
// the write fails before the rename; anything else at that path is left (a
// folder there makes every write fail, which tests use). If only the
// folder's sync fails, path holds the new bytes, but they may not survive a
// power loss, and the error says so.
func WriteFile(path string, b []byte, perm fs.FileMode) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	// A temporary file left by a crash keeps its mode through O_TRUNC.
	err = f.Chmod(perm)
	if err == nil || runtime.GOOS == "windows" {
		_, err = f.Write(b)
	}
	if err == nil {
		err = syncFile(f)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	if err := SyncDir(filepath.Dir(path)); err != nil {
		return &FolderError{Dir: filepath.Dir(path), Err: err}
	}
	return nil
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

// MkdirAll makes the folder dir and any missing parents, and syncs the
// folder each new one was made in, so that the new folders survive a power
// loss.
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
		parent := filepath.Dir(made[i])
		if err := SyncDir(parent); err != nil {
			return &FolderError{Dir: parent, Err: err}
		}
	}
	return nil
}
