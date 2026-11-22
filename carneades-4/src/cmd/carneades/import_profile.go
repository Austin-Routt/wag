package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/carneades/carneades-4/src/engine/caes"
)

func checkImportProfile(format string, experimental bool) error {
	switch format {
	case "yaml", "aif":
		return nil
	case "caf", "agxml", "lkif":
		if experimental {
			return nil
		}
		return caes.NewRunError(caes.ErrUnsupported,
			"%s is an experimental legacy converter; use -experimental-import only after reviewing its conversion policy and loss. The flag does not certify source fidelity", format)
	default:
		return caes.NewRunError(caes.ErrUnsupported, "unknown input format %q", format)
	}
}

// Importers historically return ordinary errors for malformed syntax. Preserve
// that invalid-input status, while keeping a profile refusal distinct.
func failImport(err error) {
	fmt.Fprintln(os.Stderr, err)
	if errors.Is(err, caes.ErrUnsupported) {
		os.Exit(caes.ExitStatus(caes.ErrUnsupported))
	}
	os.Exit(caes.ExitStatus(caes.ErrInvalidInput))
}
