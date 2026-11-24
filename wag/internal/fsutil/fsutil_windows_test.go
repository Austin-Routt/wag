//go:build windows

package fsutil

import (
	"fmt"
	"syscall"
	"testing"
)

// Another program holding the file makes a rename fail for a moment; a
// missing file does not pass.
func TestTransientRenameOnWindows(t *testing.T) {
	for errno, want := range map[syscall.Errno]bool{
		syscall.ERROR_ACCESS_DENIED:  true,
		32:                           true, // ERROR_SHARING_VIOLATION
		33:                           true, // ERROR_LOCK_VIOLATION
		syscall.ERROR_FILE_NOT_FOUND: false,
	} {
		if got := transientRename(fmt.Errorf("rename: %w", errno)); got != want {
			t.Errorf("%v: %v", errno, got)
		}
	}
}
