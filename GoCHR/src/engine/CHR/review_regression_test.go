package chr

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// An interrupted search is cancelled even when no tuple can fire.
func TestReviewCancelledUnmatchedPropagation(t *testing.T) {
	rs := propStore(t, rule([]string{"p(X)", "q(X)"}, []string{"r(X)"}))
	done := make(chan struct{})
	close(done)
	rs.Done = done
	if ok, _, err := rs.Infer([]string{"p(a)"}, 10); ok || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled unmatched search: ok=%v err=%v", ok, err)
	}
}

func TestReviewFailedJoinHasWorkBudget(t *testing.T) {
	rs := propStore(t, rule([]string{"p(X)", "q(X)"}, []string{"r(X)"}))
	rs.MaxMatchSteps = 200
	goals := []string{}
	for i := 0; i < 100; i++ {
		goals = append(goals, fmt.Sprintf("p(a%d)", i), fmt.Sprintf("q(b%d)", i))
	}
	if ok, store, err := rs.Infer(goals, 10); ok || !errors.Is(err, ErrLimit) || len(store) != len(goals) {
		t.Fatalf("unsuccessful join should stop without derived facts: ok=%v err=%v len=%d", ok, err, len(store))
	}
	if ok, store, err := rs.Infer([]string{"p(a)", "q(a)"}, 10); !ok || err != nil || count(store, "r(a)") != 1 {
		t.Fatalf("next run resets work budget: ok=%v err=%v store=%v", ok, err, store)
	}
}

func TestReviewApplicationLimitStreamsJoin(t *testing.T) {
	rs := propStore(t, rule([]string{"p(X)", "p(Y)", "p(Z)"}, []string{"result(X,Y,Z)"}))
	rs.MaxMatchSteps = 200
	goals := []string{}
	for i := 0; i < 100; i++ {
		goals = append(goals, fmt.Sprintf("p(a%d)", i))
	}
	_, _, err := rs.Infer(goals, 1)
	if !errors.Is(err, ErrLimit) || !strings.Contains(err.Error(), "1 rule applications") {
		t.Fatalf("one application must stop before enumerating the million-candidate join: %v", err)
	}
}

func TestReviewLogicalPremiseReuseIsExplicit(t *testing.T) {
	rs := propStore(t, rule([]string{"p(X)", "p(Y)"}, []string{"pair(X,Y)"}))
	rs.ReuseConstraints = true
	ok, store, err := rs.Infer([]string{"p(a)"}, 1)
	if !ok || err != nil || count(store, "pair(a,a)") != 1 {
		t.Fatalf("logical conjunction may reuse a fact: ok=%v err=%v store=%v", ok, err, store)
	}
}

func TestReviewStoreQuotasAndDuplicates(t *testing.T) {
	for _, nodes := range []bool{false, true} {
		rs := propStore(t, rule([]string{"p(X)"}, []string{"q(X)"}))
		if nodes {
			rs.MaxStoreNodes = 4
		} else {
			rs.MaxConstraints = 2
		}
		if ok, store, err := rs.Infer([]string{"p(a)", "p(a)"}, 1); !ok || err != nil || len(store) != 2 {
			t.Fatalf("exact quota and duplicate: nodes=%v ok=%v err=%v store=%v", nodes, ok, err, store)
		}
		if nodes {
			rs.MaxStoreNodes = 3
		} else {
			rs.MaxConstraints = 1
		}
		if ok, _, err := rs.Infer([]string{"p(a)"}, 1); ok || !errors.Is(err, ErrLimit) {
			t.Fatalf("over quota: nodes=%v ok=%v err=%v", nodes, ok, err)
		}
	}
}
