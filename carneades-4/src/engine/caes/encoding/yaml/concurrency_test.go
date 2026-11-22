package yaml

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

const concurrentGraph = `language:
  p/1: "%s is p."
  q/1: "%s is q."
argument_schemes:
  - id: s
    variables: [X]
    premises:
      - p(X)
    conclusions:
      - q(X)
assumptions:
  - p(a)
`

// docs/contracts.md, contract 1: imports and exports from several goroutines
// are serialized and give what one alone gives (the third review's
// TestConcurrentImports; run with -race by tools/vv/engine-tests.sh).
func TestConcurrentImports(t *testing.T) {
	ag, err := Import(strings.NewReader(concurrentGraph))
	if err != nil {
		t.Fatal(err)
	}
	var want bytes.Buffer
	Export(&want, ag)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ag, err := Import(strings.NewReader(concurrentGraph))
				if err != nil {
					t.Error(err)
					return
				}
				var got bytes.Buffer
				Export(&got, ag)
				if got.String() != want.String() {
					t.Errorf("a concurrent import and export gives\n%s\ninstead of\n%s", got.String(), want.String())
					return
				}
			}
		}()
	}
	wg.Wait()
}
