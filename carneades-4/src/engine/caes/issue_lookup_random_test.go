package caes

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/carneades/carneades-4/src/engine/terms"
)

// Looking positions up by key and scanning all statements give the same
// issues, or both refuse the graph, on 400 random graphs under four issue
// schemes: the negation scheme, a scheme of three patterns, one that swaps
// its arguments and an enumeration. The statements have nested terms, lists,
// numbers, variables and negations. Since statements are visited in key
// order, the scan is a fixed reference: before, it differed between two runs
// on some graphs and could not serve as one. (From pull request #17.)
func TestIssueLookupEqualsScanOnRandomGraphs(t *testing.T) {
	unary := []string{"p", "q", "accepts", "rejects", "abstains", "speed"}
	binary := []string{"wins", "loses"}
	consts := []string{"a", "b", "c", "f(a)", "f(g(b))", "[a,b]", "[a,[b,c]]", "1", "X"}
	build := func(ids []string, lookup bool) (string, error) {
		saved := issueLookup
		issueLookup = lookup
		defer func() { issueLookup = saved }()
		ag := NewArgGraph()
		for name, s := range map[string]IssueScheme{
			"negation": {"S", "¬S"},
			"pair":     {"accepts(X)", "rejects(X)", "abstains(X)"},
			"swap":     {"wins(X,Y)", "loses(Y,X)"},
			"enum":     {"speed(X1)", "...", "speed(Xn)"},
		} {
			s := s
			ag.Theory.IssueSchemes[name] = &s
		}
		for _, id := range ids {
			ag.Statements[id] = &Statement{Id: id, Text: id}
		}
		err := ag.Infer()
		return describeIssues(ag), err
	}
	refused := 0
	for seed := int64(0); seed < 400; seed++ {
		r := rand.New(rand.NewSource(seed))
		set := map[string]bool{}
		for n := 5 + r.Intn(40); len(set) < n; {
			var s string
			if r.Intn(5) == 0 {
				s = fmt.Sprintf("%s(%s,%s)", binary[r.Intn(len(binary))], consts[r.Intn(len(consts))], consts[r.Intn(len(consts))])
			} else {
				s = fmt.Sprintf("%s(%s)", unary[r.Intn(len(unary))], consts[r.Intn(len(consts))])
			}
			if r.Intn(3) == 0 {
				s = "¬" + s
			}
			set[terms.Normalize(s)] = true
		}
		ids := []string{}
		for id := range set {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		byKey, errKey := build(ids, true)
		byScan, errScan := build(ids, false)
		if (errKey == nil) != (errScan == nil) {
			t.Fatalf("seed %d: lookup gives %v, scan gives %v\nstatements: %v", seed, errKey, errScan, ids)
		}
		if errKey != nil {
			refused++
			continue
		}
		if byKey != byScan {
			t.Fatalf("seed %d:\nlookup:\n%s\nscan:\n%s\nstatements: %v", seed, byKey, byScan, ids)
		}
	}
	if refused > 200 {
		t.Errorf("%d of 400 graphs were refused: the test compares too few issue structures", refused)
	}
}
