package carneades

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/carneades/carneades-4/src/engine/terms"
)

var statementPlaceholder = regexp.MustCompile(`\{([^{}]+)\}`)

// MatchesStatementPattern compares a ground statement with a NASCQ pattern
// without requiring an argument of that pattern's scheme to be on the board.
// Only {Var} placeholders vary; constants and repeated occurrences retain
// their meaning. Parsing and matching stay in the Carneades adapter.
func (a *Adapter) MatchesStatementPattern(stmtID, pattern string) bool {
	// A different outer predicate cannot match. Most graph statements are
	// unrelated; avoid parsing the protected pattern for those statements.
	if i := strings.IndexByte(pattern, '('); i >= 0 &&
		!strings.ContainsAny(stmtID+pattern, " \t\r\n\"'") &&
		!strings.ContainsAny(pattern[:i], "{}") && !strings.HasPrefix(stmtID, pattern[:i+1]) {
		return false
	}
	// Check before making variables so an unmarked variable named, for
	// example, PatternVar0 cannot collide with a generated placeholder.
	groundPattern, ok := terms.ReadString(statementPlaceholder.ReplaceAllString(pattern, "placeholder"))
	if !ok || !terms.Ground(groundPattern, nil) {
		return false
	}
	names := map[string]string{}
	allowed := map[string]bool{}
	pattern = statementPlaceholder.ReplaceAllStringFunc(pattern, func(p string) string {
		if v, ok := names[p]; ok {
			return v
		}
		v := fmt.Sprintf("PatternVar%d", len(names))
		names[p], allowed[v] = v, true
		return v
	})
	p, ok := terms.ReadString(pattern)
	if !ok {
		return false
	}
	for _, v := range p.OccurVars() {
		if !allowed[v.Name] {
			return false
		}
	}
	s, ok := terms.ReadString(stmtID)
	if !ok || !terms.Ground(s, nil) {
		return false
	}
	_, ok = terms.Match(p, s, nil)
	return ok
}
