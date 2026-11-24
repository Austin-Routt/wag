//go:build !unix && !windows

package fsutil

func cannotSyncFolders(error) bool { return false }

func transientRename(error) bool { return false }
