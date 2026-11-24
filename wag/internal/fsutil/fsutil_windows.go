//go:build windows

package fsutil

import (
	"errors"
	"syscall"
)

// cannotSyncFolders: Windows does not sync folders (SyncDir returns first).
func cannotSyncFolders(error) bool { return false }

// transientRename: another program (a virus scanner, an indexer) has the
// file open, and the move may work a moment later.
func transientRename(err error) bool {
	var e syscall.Errno
	if !errors.As(err, &e) {
		return false
	}
	const sharingViolation, lockViolation = 32, 33 // ERROR_SHARING_VIOLATION, ERROR_LOCK_VIOLATION
	return e == syscall.ERROR_ACCESS_DENIED || e == sharingViolation || e == lockViolation
}
