package game

// Regression test for defect #2 of the 2026-10-06 review: the exception
// arguments InstantiateScheme creates were given the support scheme and the
// mover as owner, so 76 of the 95 enumerated attack moves targeted them
// (e.g. "other goals?" asked of "side effects => undercut(a1)").

import (
	"testing"

	"github.com/yourorg/wag/internal/carneades"
)

func TestExceptionArgumentsAreNotMovesAndNotTargets(t *testing.T) {
	gs := makeCoreGame(t)
	r, _ := gs.ExecuteTurn(prSupportAction())
	if !r.Accepted {
		t.Fatalf("support rejected: %s", r.Error)
	}
	main := r.ArgIDs[0]
	exceptions := 0
	for _, id := range r.ArgIDs[1:] {
		a := gs.Adapter.Graph.Arguments[id]
		exceptions++
		if carneades.ArgumentRole(a) != carneades.RoleException {
			t.Errorf("%s: role %q, want exception", id, carneades.ArgumentRole(a))
		}
		if o, ok := a.Metadata["owner"]; ok {
			t.Errorf("%s: exception argument has owner %v", id, o)
		}
		if a.Scheme != nil {
			t.Errorf("%s: exception argument carries scheme %q", id, a.Scheme.Id)
		}
	}
	if exceptions != 4 { // practical_reasoning1 declares four exceptions
		t.Fatalf("%d exception arguments, want 4", exceptions)
	}
	if carneades.ArgumentRole(gs.Adapter.Graph.Arguments[main]) != carneades.RoleMain || gs.Adapter.ArgumentOwner(main) != "Alice" {
		t.Fatalf("main argument %s: role %q owner %q", main, carneades.ArgumentRole(gs.Adapter.Graph.Arguments[main]), gs.Adapter.ArgumentOwner(main))
	}
	attacks := 0
	for _, a := range gs.BuildObservation().AvailableActions {
		if a.Type != "attack" {
			continue
		}
		attacks++
		if a.TargetArgID != main {
			t.Errorf("attack template %s/%s targets %s, not the main argument %s", a.AnswersCQID, a.SchemeID, a.TargetArgID, main)
		}
	}
	if attacks == 0 {
		t.Fatal("no attack template on the main argument")
	}
	t.Logf("%d attack templates, all on the main argument %s (the snapshot of 2026-10-06 enumerated 95, 19 on each of the five arguments: 76 on exception arguments)", attacks, main)
}
