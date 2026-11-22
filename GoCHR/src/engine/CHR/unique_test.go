// Copyright © 2026 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Tests of RuleStore.Unique, the set semantics for ground constraints that
// Carneades uses for argument construction.

package chr

import (
	"sort"
	"strings"
	"testing"

	. "github.com/hfried/GoCHR/src/engine/terms"
)

func count(store []string, c string) int {
	n := 0
	for _, s := range store {
		if strings.ReplaceAll(s, " ", "") == c {
			n++
		}
	}
	return n
}

func TestUniqueOffKeepsMultiset(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.AddRule("keep", []string{"p(X)"}, nil, nil, []string{"q(X)"})
	_, store, err := rs.Infer([]string{"p(a)", "p(a)"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if count(store, "p(a)") != 2 {
		t.Errorf("without Unique the store is a multiset: want two p(a), got %v", store)
	}
}

func TestUniqueKeepsOneCopy(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	rs.AddRule("keep", []string{"p(X)"}, nil, nil, []string{"q(X)"})
	_, store, err := rs.Infer([]string{"p(a)", "p(a)", "p(b)"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if count(store, "p(a)") != 1 || count(store, "q(a)") != 1 || count(store, "q(b)") != 1 {
		sort.Strings(store)
		t.Errorf("with Unique: want one p(a), one q(a), one q(b); got %v", store)
	}
}

func TestUniqueKeepsAritiesApart(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	// a rule has to fire: when none does, Infer returns an empty store (PLAN.md 1.3)
	rs.AddRule("start", []string{"go"}, nil, nil, []string{"started"})
	_, store, err := rs.Infer([]string{"go", "p(a)", "p(a,b)", "p(a)", "p(b,a)"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if count(store, "p(a)") != 1 || count(store, "p(a,b)") != 1 || count(store, "p(b,a)") != 1 {
		t.Errorf("with Unique: want one each of p(a), p(a,b), p(b,a); got %v", store)
	}
}

func TestUniqueEndsRederivation(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	// two rules that derive each other's premise: a multiset store runs to the limit
	rs.AddRule("r1", []string{"p(X)"}, nil, nil, []string{"q(X)"})
	rs.AddRule("r2", []string{"q(X)"}, nil, nil, []string{"p(X)"})
	rs.AddRule("r3", []string{"q(X)"}, nil, nil, []string{"s(X)"})
	_, store, err := rs.Infer([]string{"p(a)"}, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if count(store, "s(a)") != 1 || count(store, "p(a)") != 1 || count(store, "q(a)") != 1 {
		t.Errorf("with Unique the cycle ends and the third rule fires once: got %v", store)
	}
}

func TestGroundKeyIsInjective(t *testing.T) {
	pairs := [][2]Term{
		{Atom("a"), String("a")},
		{Int(1), Float(1)},
		{Compound{Functor: "p", Args: []Term{Atom("a")}}, Compound{Functor: "p", Args: []Term{Atom("a"), Atom("b")}}},
		{List{Atom("a")}, Compound{Functor: "a"}},
		{Atom("a,b"), Compound{Functor: "a", Args: []Term{Atom("b")}}},
	}
	for _, p := range pairs {
		if groundKey(p[0]) == groundKey(p[1]) {
			t.Errorf("groundKey(%v) == groundKey(%v) = %q", p[0], p[1], groundKey(p[0]))
		}
	}
}

// A rule's body that re-adds one of the rule's own heads: matching marks the
// heads while the rule fires, and that mark must not make the copy look gone.
func TestUniqueBodyRepeatsItsHead(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	rs.AddRule("again", []string{"r(X)", "c(Y)"}, nil, nil, []string{"w(Y)", "c(Y)"})
	_, store, err := rs.Infer([]string{"r(x)", "c(a)"}, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if count(store, "c(a)") != 1 || count(store, "w(a)") != 1 {
		t.Errorf("with Unique a body that repeats a head adds nothing: want one c(a) and one w(a), got %d and %d", count(store, "c(a)"), count(store, "w(a)"))
	}
}

// A deleted constraint can be added again.
func TestUniqueDeletedCanReturn(t *testing.T) {
	CHRtrace = 0
	rs := MakeRuleStore()
	rs.Unique = true
	rs.AddRule("consume", nil, []string{"p(X)", "t(X)"}, nil, []string{"q(X)"})
	rs.AddRule("back", []string{"q(X)"}, nil, nil, []string{"p(X)"})
	_, store, err := rs.Infer([]string{"p(a)", "t(a)"}, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if count(store, "p(a)") != 1 || count(store, "q(a)") != 1 {
		t.Errorf("p(a), deleted by the first rule, comes back through the second: got %v", store)
	}
}
