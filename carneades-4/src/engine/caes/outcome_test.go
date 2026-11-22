// Copyright © 2026 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed
// with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

package caes

import (
	"errors"
	"fmt"
	"testing"
)

func TestExitStatus(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, 0},
		{NewRunError(ErrInvalidInput, "x"), 1},
		{NewRunError(ErrUnsupported, "x"), 2},
		{NewRunError(ErrInconsistentContext, "x"), 3},
		{NewRunError(ErrLimitReached, "x"), 4},
		{NewRunError(ErrCancelled, "x"), 5},
		{NewRunError(ErrInternal, "x"), 6},
		{errors.New("some other error"), 6},
		{fmt.Errorf("wrapped: %w", NewRunError(ErrLimitReached, "x")), 4},
	}
	for _, c := range cases {
		if got := ExitStatus(c.err); got != c.want {
			t.Errorf("ExitStatus(%v) = %d, want %d", c.err, got, c.want)
		}
	}
	if !errors.Is(NewRunError(ErrLimitReached, "x"), ErrLimitReached) {
		t.Errorf("a RunError does not unwrap to its kind")
	}
}
