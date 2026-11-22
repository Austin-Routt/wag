package web

import (
	"errors"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
)

// The web service admits the same input formats as check and eval: the
// legacy converters only with the explicit opt-in.
func TestWebImportProfile(t *testing.T) {
	defer func(old bool) { ExperimentalImport = old }(ExperimentalImport)
	for _, experimental := range []bool{false, true} {
		ExperimentalImport = experimental
		for _, f := range []string{"yaml", "aif"} {
			if err := checkImportProfile(f); err != nil {
				t.Errorf("experimental=%v: %s refused: %v", experimental, f, err)
			}
		}
		for _, f := range []string{"caf", "agxml", "lkif"} {
			err := checkImportProfile(f)
			if experimental && err != nil {
				t.Errorf("opted in, %s refused: %v", f, err)
			}
			if !experimental && !errors.Is(err, caes.ErrUnsupported) {
				t.Errorf("not opted in, %s admitted (err %v)", f, err)
			}
		}
		if err := checkImportProfile("graphml"); !errors.Is(err, caes.ErrUnsupported) {
			t.Errorf("experimental=%v: unknown format admitted (err %v)", experimental, err)
		}
	}
}
