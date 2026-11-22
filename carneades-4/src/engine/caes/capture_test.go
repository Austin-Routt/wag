// Copyright © 2026 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed
// with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

package caes

import (
	"testing"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// A parameter value is not substituted again: the fresh variable A1 the rule
// engine gives an unbound A is not captured by a scheme variable A1
// (E-CAPTURE).
func TestSubstituteParametersOnce(t *testing.T) {
	q, _ := terms.ReadString("q(A,A1)")
	a1, _ := terms.ReadString("A1")
	a, _ := terms.ReadString("a")
	got := substituteParameters(q, map[string]terms.Term{"A": a1, "A1": a}).String()
	if got != "q(A1,a)" {
		t.Errorf("got %s, want q(A1,a)", got)
	}
}
