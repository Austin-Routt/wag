package carneades

import (
	"fmt"
	"sort"
	"strings"
)

// Fingerprint renders the graph's structure canonically (statements with
// their arguments, issue, undercutter flag and text; arguments with premises,
// conclusion, undercutter and role; issues with positions; assumptions).
// Two graphs with equal fingerprints are structurally identical. Used by the
// rollback tests, and useful for replay checks.
func (a *Adapter) Fingerprint() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var b strings.Builder
	for _, id := range SortedKeys(a.Graph.Statements) {
		s := a.Graph.Statements[id]
		args := make([]string, 0, len(s.Args))
		for _, x := range s.Args {
			args = append(args, x.Id)
		}
		issue := ""
		if s.Issue != nil {
			issue = s.Issue.Id
		}
		fmt.Fprintf(&b, "S %s|%s|%t|%q|%s\n", id, strings.Join(args, ","), s.IsUndercutter, s.Text, issue)
	}
	ids := SortedKeys(a.Graph.Arguments)
	SortArgumentIDs(ids)
	for _, id := range ids {
		x := a.Graph.Arguments[id]
		prem := make([]string, 0, len(x.Premises))
		for _, p := range x.Premises {
			if p.Stmt != nil {
				prem = append(prem, p.Stmt.Id)
			}
		}
		concl, und := "", ""
		if x.Conclusion != nil {
			concl = x.Conclusion.Id
		}
		if x.Undercutter != nil {
			und = x.Undercutter.Id
		}
		fmt.Fprintf(&b, "A %s|%s|%s|%s|%s\n", id, strings.Join(prem, ","), concl, und, ArgumentRole(x))
	}
	for _, id := range SortedKeys(a.Graph.Issues) {
		pos := []string{}
		for _, p := range a.Graph.Issues[id].Positions {
			pos = append(pos, p.Id)
		}
		sort.Strings(pos)
		fmt.Fprintf(&b, "I %s|%s\n", id, strings.Join(pos, ","))
	}
	fmt.Fprintf(&b, "assumptions %s\n", strings.Join(a.Graph.Assumptions, ","))
	return b.String()
}
