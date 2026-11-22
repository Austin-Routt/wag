// Copyright © 2026 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed
// with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

package caes

import (
	"errors"
	"fmt"
)

// How a run of the engine ends: one of seven outcomes (docs/contracts.md,
// contract 3 of the Walton scheme library). A run that completes returns no
// error. Every other outcome is an error that wraps one of the six values
// below, so that callers can test it with errors.Is.
var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnsupported         = errors.New("unsupported")
	ErrInconsistentContext = errors.New("inconsistent context")
	ErrLimitReached        = errors.New("limit reached")
	ErrCancelled           = errors.New("cancelled")
	ErrInternal            = errors.New("internal failure")
)

// RunError is an error with one of the outcomes above as its kind.
type RunError struct {
	Kind   error  // one of ErrInvalidInput ... ErrInternal
	Detail string // what happened, for a person
}

func (e *RunError) Error() string {
	if e.Detail == "" {
		return e.Kind.Error()
	}
	return e.Kind.Error() + ": " + e.Detail
}

func (e *RunError) Unwrap() error { return e.Kind }

// NewRunError returns an error of the given kind.
func NewRunError(kind error, format string, args ...interface{}) error {
	return &RunError{Kind: kind, Detail: fmt.Sprintf(format, args...)}
}

// ExitStatus maps an outcome to the exit status of the carneades command:
// 0 complete, 1 invalid input, 2 unsupported, 3 inconsistent context,
// 4 limit reached, 5 cancelled, 6 internal failure. An error of no known
// kind counts as an internal failure.
func ExitStatus(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, ErrInvalidInput):
		return 1
	case errors.Is(err, ErrUnsupported):
		return 2
	case errors.Is(err, ErrInconsistentContext):
		return 3
	case errors.Is(err, ErrLimitReached):
		return 4
	case errors.Is(err, ErrCancelled):
		return 5
	default:
		return 6
	}
}
