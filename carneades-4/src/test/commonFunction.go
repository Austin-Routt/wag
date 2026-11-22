package test

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	//	"log"
	"errors"
	"os"
	"testing"
)

// refusedExamples are the examples Infer refuses, with the outcome it reports
// (docs/contracts.md, contract 3): chr-leq uses CHR deletions and guards,
// outside the supported profile (contract 1, PLAN.md D3); paraconsistency
// assumes a statement and its negation, an inconsistent context that is
// reported, not resolved (contract 4, D10).
var refusedExamples = map[string]error{
	"chr-leq.yml":         caes.ErrUnsupported,
	"paraconsistency.yml": caes.ErrInconsistentContext,
}

// skipUnsupported checks that Infer refuses a listed example with its
// outcome, and tells the caller to skip it.
func skipUnsupported(t *testing.T, name string) bool {
	t.Helper()
	want, listed := refusedExamples[name]
	if !listed {
		return false
	}
	f, err := os.Open(yamlDir + name)
	if err != nil {
		t.Errorf("%s: %v", name, err)
		return true
	}
	defer f.Close()
	ag, err := yaml.Import(f)
	if err != nil {
		t.Errorf("%s: %v", name, err)
		return true
	}
	if err := ag.Infer(); !errors.Is(err, want) {
		t.Errorf("%s: Infer returned %v, want %v", name, err, want)
	}
	return true
}

const yamlDir = "../../examples/AGs/YAML/"
const yamlTmp = "/tmp/"
const jsonDir = "../../examples/AGs/AIF/"
const jsonTmp = "/tmp/"
const graphmlTmp = "/tmp/"

func check(t *testing.T, e error) {
	if e != nil {
		t.Errorf(e.Error())
		t.Skip("Skip Test")
	}
}

func checkLabeling(l caes.Labelling, stats map[string]*caes.Statement, expectedLbls map[string]caes.Label) error {
	errStr := ""
	for _, stat := range stats {
		calcLbl := l[stat]
		expLbl, found := expectedLbls[stat.Id]
		if found {
			if expLbl != calcLbl {
				if errStr == "" {
					errStr = fmt.Sprintf(" statement: %s, expected Label: %v, calculated Label: %v \n", stat.Id, expLbl, calcLbl)
				} else {
					errStr = errStr + fmt.Sprintf(" statement: %s, expected Label: %v, calculated Label: %v \n", stat.Id, expLbl, calcLbl)
				}
			}
		}
	}
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}

// exportLoss lists the examples whose expected labels do not survive export
// and import. It is empty since PLAN.md 4.1 and 4.2: the export writes the
// weights, the weighing functions and the expected labels (rebuttal,
// mcda-porsche and lex lost theirs before). checkRoundTrip fails for any
// example that loses its labels, and for a listed one that keeps them, so
// the list cannot go stale.
var exportLoss = map[string]bool{}

func checkRoundTrip(t *testing.T, file string, err error) {
	t.Helper()
	switch {
	case err != nil && !exportLoss[file]:
		t.Errorf("%s: labels after export and import: %v", file, err)
	case err == nil && exportLoss[file]:
		t.Errorf("%s: export and import now keep the expected labels: remove it from exportLoss", file)
	case err != nil:
		t.Logf("%s: known export loss (E-EXPORT): %v", file, err)
	}
}

func printLabeling(l caes.Labelling) {

	for ref_stat, lbl := range l {
		fmt.Printf(" statement: %s    Label: %v\n", ref_stat.Id, lbl)
	}

}
