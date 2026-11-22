package web

import "github.com/carneades/carneades-4/src/engine/caes"

// ExperimentalImport opts the web service in to the legacy CAF, AGXML and
// LKIF converters, as -experimental-import does for check and eval
// (docs/import-profiles.md). Without it an upload in those formats is
// refused before conversion, so the service has the same input profile as
// the command line: native YAML and the restricted AIF projection.
var ExperimentalImport bool

func checkImportProfile(format string) error {
	switch format {
	case "yaml", "aif":
		return nil
	case "caf", "agxml", "lkif":
		if ExperimentalImport {
			return nil
		}
		return caes.NewRunError(caes.ErrUnsupported,
			"%s is an experimental legacy converter; start the server with -experimental-import only after reviewing its conversion policy and loss. The flag does not certify source fidelity", format)
	default:
		return caes.NewRunError(caes.ErrUnsupported, "unknown or unsupported input format: %s", format)
	}
}
