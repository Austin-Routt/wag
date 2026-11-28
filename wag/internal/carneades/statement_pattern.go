package carneades

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/carneades/carneades-4/src/engine/terms"
)

var statementPlaceholder = regexp.MustCompile(`\{([^{}]+)\}`)

// MatchesStatementPattern compares a statement with a NASCQ pattern without
// requiring an argument of that pattern's scheme to be on the board. Only
// {Var} placeholders vary; constants and repeated occurrences retain their
// meaning. The statement is read as the engine keys it, and a variable in it
// is a constant of its own (docs/c5-item-f.md, section K). Parsing and
// matching stay in the Carneades adapter.
func (a *Adapter) MatchesStatementPattern(stmtID, pattern string) bool {
	// A different outer predicate cannot match. Most graph statements are
	// unrelated; avoid parsing the protected pattern for those statements.
	// The shortcut reads the statement as its text shows it, so it is taken
	// only for a name the engine keys as written up to "(" (plainName): the
	// engine reads "¬¬p", "(p)", "+p", "--p" and "p/*c*/(x)" as other forms
	// (the independent check of F1, F2 and F4, E1).
	if i := strings.IndexByte(pattern, '('); i >= 0 && plainName(stmtID) &&
		!strings.ContainsAny(pattern, " \t\r\n\"'") &&
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
	// The statement as the engine keys it (terms.Normalize reads it with the
	// negation identity), each variable a constant of its own: in a game with
	// no profile the claim's variable is in the bridge's exception (W1 of the
	// check; F1 took ground statements only).
	s, ok := terms.ReadFormula(stmtID)
	if !ok {
		return false
	}
	_, ok = terms.Match(p, variablesAsConstants(s), nil)
	return ok
}

// plainName reports whether the engine keys id as its text shows it up to
// its first "(": it starts with a lower-case letter and holds no space,
// quote or comment.
func plainName(id string) bool {
	return id != "" && id[0] >= 'a' && id[0] <= 'z' && !strings.ContainsAny(id, " \t\r\n\"'%") && !strings.Contains(id, "/*")
}

// variablesAsConstants returns t with each variable made a constant named by
// the variable, which a placeholder can stand for and no constant of a
// pattern can equal (its name starts with a character the reader puts in no
// atom). The engine's reader gives each occurrence of X a variable of its
// own, so the occurrences are compared by name.
func variablesAsConstants(t terms.Term) terms.Term {
	switch x := t.(type) {
	case terms.Variable:
		return terms.Atom("\x00" + x.Name)
	case terms.Compound:
		args := make([]terms.Term, len(x.Args))
		for i, arg := range x.Args {
			args[i] = variablesAsConstants(arg)
		}
		x.Args = args
		return x
	case terms.List:
		out := make(terms.List, len(x))
		for i, e := range x {
			out[i] = variablesAsConstants(e)
		}
		return out
	}
	return t
}

// UndercutterStatement reports whether formula, or the statement it negates,
// is one the engine names for an argument's undercutter: undercut/1, as the
// engine reads it. The engine writes undercut(<id>) when it makes argument
// <id>, replacing a statement so named before (and WAG's own paths take one
// over), so a claim so named would be about an argument not yet made (F2 of
// the external review of exchange 3).
func UndercutterStatement(formula string) bool {
	t, ok := terms.ReadString(formula)
	if !ok {
		return false
	}
	for {
		c, ok := t.(terms.Compound)
		if !ok {
			return false
		}
		if c.Functor == "¬" && len(c.Args) == 1 {
			t = c.Args[0]
			continue
		}
		return c.Functor == "undercut" && len(c.Args) == 1
	}
}
