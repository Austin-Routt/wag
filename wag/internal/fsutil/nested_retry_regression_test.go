package fsutil

import (
	"errors"
	"path/filepath"
	"testing"
)

// MkdirAll creates all components before it syncs their parent entries.
// Failure at an early parent must not forget the still-unsynced children.
func TestReviewNestedCreationRetainsAllPendingParents(t *testing.T) {
	for _, retryByWrite := range []bool{false, true} {
		name := "mkdir_retry"
		if retryByWrite {
			name = "write_retry"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			a := filepath.Join(root, "a")
			b := filepath.Join(a, "b")
			dir := filepath.Join(b, "c")
			oldSync := SyncDir
			defer func() { SyncDir = oldSync }()
			broken := true
			var synced []string
			SyncDir = func(path string) error {
				synced = append(synced, path)
				if path == root && broken {
					return errors.New("synthetic directory sync failure")
				}
				return oldSync(path)
			}
			if err := MkdirAll(dir, 0755); err == nil {
				t.Fatal("fixture must fail at the earliest new parent")
			}
			broken = false
			synced = nil
			var err error
			if retryByWrite {
				err = WriteFile(filepath.Join(dir, "record.json"), []byte("{}"), 0644)
			} else {
				err = MkdirAll(dir, 0755)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, parent := range []string{root, a, b} {
				if !has(synced, parent) {
					t.Errorf("retry reported success without syncing newly created entry parent %q; calls=%v", parent, synced)
				}
			}
		})
	}
}
