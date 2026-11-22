package chr

import (
	"fmt"
	"sync"
	"testing"

	. "github.com/hfried/GoCHR/src/engine/terms"
)

// The spelling maps are initialized once, also when several goroutines spell
// at the same time (the third review's TestConcurrentSpellInitialization; run
// with -race by tools/vv/engine-tests.sh).
func TestConcurrentSpellInitialization(t *testing.T) {
	start := make(chan struct{})
	results := make([]string, 8)
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			v, ok := Spell2text(List{String("\"a\"")})
			results[i] = fmt.Sprint(v, ok)
		}(i)
	}
	close(start)
	wg.Wait()
	v, ok := Spell2text(List{String("\"a\"")})
	want := fmt.Sprint(v, ok)
	for i, r := range results {
		if r != want {
			t.Errorf("goroutine %d spelled %q, alone it is %q", i, r, want)
		}
	}
}
