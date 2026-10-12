package carneades

import (
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// Regression test for review defect #5: rollback must remove everything a
// move created, not only its arguments.
func TestRestoreUndoesEveryKindOfMutation(t *testing.T) {
	a := makeAdapterWithScheme()
	a.AddStatement("base", "Base")
	a.AddStatement("target", "Target")
	if _, err := a.AddArgumentDirect("Alice", "s", []string{"base"}, "target", 0.6); err != nil {
		t.Fatal(err)
	}
	a.AddAssumption("base")
	a.Evaluate()
	before := a.Fingerprint()
	cp := a.Checkpoint()

	ids, err := a.AddSupport("Bob", "test_support", map[string]string{"P": "x", "C": "y"}, 0.7) // new statements and an undercutter
	if err != nil {
		t.Fatal(err)
	}
	a.AddAssumption("premise(x)")
	if _, err := a.AddRebuttal("Bob", "s", "target", []string{"base"}, 0.7, caes.PE); err != nil { // ¬target and an issue
		t.Fatal(err)
	}
	if _, err := a.AddUndercutter("Bob", "s", "a1", []string{"base"}, 0.7); err != nil { // sets a1's undercutter
		t.Fatal(err)
	}
	if _, err := a.AddPresumption("Bob", "base", 0.01); err != nil {
		t.Fatal(err)
	}
	if a.Fingerprint() == before {
		t.Fatal("the mutations changed nothing")
	}
	a.Restore(cp)
	if after := a.Fingerprint(); after != before {
		t.Fatalf("Restore left differences.\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if a.Graph.Arguments["a1"].Undercutter != nil {
		t.Fatal("a1's undercutter pointer was not restored")
	}
	a.Evaluate() // the restored graph evaluates as before
	if a.Fingerprint() != before {
		t.Fatal("evaluation after Restore changed the structure")
	}
	_ = ids
}

// RemoveArguments (the old rollback) leaves residue; documents why Restore exists.
func TestRemoveArgumentsAloneLeavesResidue(t *testing.T) {
	a := makeAdapterWithScheme()
	before := a.Fingerprint()
	ids, _ := a.AddSupport("X", "test_support", map[string]string{"P": "x", "C": "y"}, 0.5)
	a.AddAssumption("premise(x)")
	_ = a.RemoveArguments(ids)
	if a.Fingerprint() == before {
		t.Fatal("expected residue from RemoveArguments")
	}
}
