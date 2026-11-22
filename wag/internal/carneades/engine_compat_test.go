package carneades

import (
	"reflect"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

const upgradeDefaultYAML = `language:
  p/1: "Premise %s"
  d/1: "Default %s"
  c/1: "Conclusion %s"
argument_schemes:
  - id: defaults_case
    variables: [X]
    premises: [p(X)]
    assumptions: [d(X)]
    conclusions: [c(X)]
`

func upgradeAdapter(t *testing.T) *Adapter {
	t.Helper()
	a, err := NewAdapterFromYAML([]byte(upgradeDefaultYAML))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// Rejected moves and simulation must restore the candidate engine's new
// assumption provenance, not only the legacy Assumptions slice.
func TestEngineUpgradeCheckpointNewDefault(t *testing.T) {
	a := upgradeAdapter(t)
	cp := a.Checkpoint()
	if _, err := a.AddSupport("owner", "defaults_case", map[string]string{"X": "a"}, 1); err != nil {
		t.Fatal(err)
	}
	if !a.Graph.Defaults["d(a)"] {
		t.Fatal("fixture did not instantiate a default")
	}
	a.Restore(cp)
	if len(a.Graph.Defaults) != 0 {
		t.Fatalf("rollback leaked default provenance: %v", a.Graph.Defaults)
	}
}

// A tentative stipulation promotes a default; restoring it must restore
// its status, so a later question still suspends the default.
func TestEngineUpgradeCheckpointPromotion(t *testing.T) {
	a := upgradeAdapter(t)
	if _, err := a.AddSupport("owner", "defaults_case", map[string]string{"X": "a"}, 1); err != nil {
		t.Fatal(err)
	}
	cp := a.Checkpoint()
	a.AddAssumption("d(a)")
	a.Restore(cp)
	a.Graph.Questioned = []string{"d(a)"}
	labels, err := a.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	if got := labels[a.Graph.Statements["d(a)"]]; got != caes.Out {
		t.Fatalf("restored default must be suspended when questioned, got %s; defaults=%v", got, a.Graph.Defaults)
	}
}

func TestEngineUpgradeCheckpointQuestions(t *testing.T) {
	a := upgradeAdapter(t)
	a.Graph.Questioned = []string{"p(a)"}
	cp := a.Checkpoint()
	a.Graph.Questioned[0] = "d(a)"
	a.Graph.Questioned = append(a.Graph.Questioned, "c(a)")
	a.Restore(cp)
	if !reflect.DeepEqual(a.Graph.Questioned, []string{"p(a)"}) {
		t.Fatalf("question state not restored: %v", a.Graph.Questioned)
	}
}

// WAG retracts scheme defaults and uses weighted presumption arguments.
// The removed assumptions must not leave engine provenance behind.
func TestEngineUpgradeRetractionClearsDefaultOrigin(t *testing.T) {
	a := upgradeAdapter(t)
	n := a.AssumptionCount()
	if _, err := a.AddSupport("owner", "defaults_case", map[string]string{"X": "a"}, 1); err != nil {
		t.Fatal(err)
	}
	if got := a.RetractAssumptionsSince(n); !reflect.DeepEqual(got, []string{"d(a)"}) {
		t.Fatalf("wrong retraction: %v", got)
	}
	if len(a.Graph.Defaults) != 0 {
		t.Fatalf("retraction leaked default provenance: %v", a.Graph.Defaults)
	}
}

func TestEngineUpgradeModeratorWithdrawalClearsOrigin(t *testing.T) {
	a := upgradeAdapter(t)
	if _, err := a.AddSupport("owner", "defaults_case", map[string]string{"X": "a"}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Decide("d(a)", DecisionUndecided); err != nil {
		t.Fatal(err)
	}
	if len(a.Graph.Defaults) != 0 {
		t.Fatalf("withdrawal leaked default provenance: %v", a.Graph.Defaults)
	}
}

func TestEngineUpgradeFingerprintIncludesProvenance(t *testing.T) {
	a := upgradeAdapter(t)
	a.AddStatement("p(a)", "P")
	a.AddAssumption("p(a)")
	before := a.Fingerprint()
	a.Graph.Defaults["p(a)"] = true
	if a.Fingerprint() == before {
		t.Error("fingerprint hides a stipulation-to-default change")
	}
	before = a.Fingerprint()
	a.Graph.Questioned = []string{"p(a)"}
	if a.Fingerprint() == before {
		t.Error("fingerprint hides a questioning change")
	}
}

func TestEngineUpgradeRetractionPreservesRetainedOrigins(t *testing.T) {
	a := upgradeAdapter(t)
	a.AddStatement("d(a)", "Default")
	a.Graph.Assumptions = []string{"d(a)", "d( a )"}
	a.Graph.Defaults["d(a)"] = true
	a.Graph.Questioned = []string{"d(a)"}
	a.RetractAssumptionsSince(1)
	if !a.Graph.Defaults["d(a)"] || !reflect.DeepEqual(a.Graph.Questioned, []string{"d(a)"}) {
		t.Fatalf("retained origin/question history changed: defaults=%v questioned=%v", a.Graph.Defaults, a.Graph.Questioned)
	}
	if _, err := a.Decide("d(a)", DecisionUndecided); err != nil {
		t.Fatal(err)
	}
	if len(a.Graph.Defaults) != 0 || !reflect.DeepEqual(a.Graph.Questioned, []string{"d(a)"}) {
		t.Fatalf("withdrawal must remove origin while retaining question history: defaults=%v questioned=%v", a.Graph.Defaults, a.Graph.Questioned)
	}
}

func TestEngineUpgradeCheckpointProvenanceCopyIsolation(t *testing.T) {
	a := upgradeAdapter(t)
	a.Graph.Defaults["d(a)"] = true
	cp := a.Checkpoint()
	a.Graph.Defaults["d(a)"] = false
	a.Restore(cp)
	if !a.Graph.Defaults["d(a)"] {
		t.Fatal("checkpoint shared the live default map")
	}
	a.Graph.Defaults["d(a)"] = false
	a.Restore(cp)
	if !a.Graph.Defaults["d(a)"] {
		t.Fatal("restore shared the saved default map")
	}
}
