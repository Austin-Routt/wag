//go:build unix

package fsutil

import (
	"errors"
	"syscall"
)

// cannotSyncFolders: the file system does not sync folders (some network
// shares answer EINVAL, EBADF or EOPNOTSUPP); PostgreSQL takes these as done.
func cannotSyncFolders(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.EBADF) || errors.Is(err, syscall.EOPNOTSUPP)
}

// transientRename: a rename on Unix does not fail for a moment.
func transientRename(error) bool { return false }
