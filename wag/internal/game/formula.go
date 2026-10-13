package game

import (
	"fmt"
	"strings"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// FormulaText is how a scheme's formula reads through the Language
// templates, with its variables as {X} placeholders ("Doing {A} in the
// situation {S1} would lead to {S2}"); a formula that is a bare variable
// reads "{S}". The scheme card and the text protocol show schemes this way,
// in the same words the board uses once the variables are filled in.
func FormulaText(lang caes.Language, formula string) string {
	t, ok := terms.ReadString(formula)
	if !ok {
		return formula
	}
	return formulaText(lang, t)
}

func formulaText(lang caes.Language, t terms.Term) string {
	c, ok := t.(terms.Compound)
	if !ok {
		return "{" + t.String() + "}"
	}
	tmpl := lang[fmt.Sprintf("%s/%d", c.Functor, len(c.Args))]
	if tmpl == "" || strings.Count(strings.ReplaceAll(tmpl, "%%", ""), "%s") != len(c.Args) {
		return t.String()
	}
	args := make([]interface{}, len(c.Args))
	for i, a := range c.Args {
		args[i] = formulaText(lang, a)
	}
	return strings.TrimSuffix(fmt.Sprintf(tmpl, args...), ".")
}
