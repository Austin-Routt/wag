package lkif

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
)

func (g *ArgumentGraph) validateProfile() error {
	invalid := func(f string, a ...interface{}) error {
		return caes.NewRunError(caes.ErrInvalidInput, "LKIF: "+f, a...)
	}
	unsupported := func(f string, a ...interface{}) error {
		return caes.NewRunError(caes.ErrUnsupported, "experimental LKIF: "+f, a...)
	}
	stmts, ids := map[string]bool{}, map[string]bool{}
	if len(g.Statements.Content) == 0 {
		return invalid("at least one statement is required")
	}
	for _, s := range g.Statements.Content {
		if s.Id == "" || ids[s.Id] {
			return invalid("empty or duplicate ID %q", s.Id)
		}
		ids[s.Id], stmts[s.Id] = true, true
		switch s.Standard {
		case "", "PE", "CCE", "BRD":
		default:
			return unsupported("proof standard %q", s.Standard)
		}
		switch s.Value {
		case "", "true", "false", "unknown":
		default:
			return unsupported("statement value %q", s.Value)
		}
	}
	for _, a := range g.Arguments.Content {
		if a.Id == "" || ids[a.Id] {
			return invalid("empty or duplicate ID %q", a.Id)
		}
		ids[a.Id] = true
		switch a.Direction {
		case "", "pro", "con":
		default:
			return unsupported("argument direction %q", a.Direction)
		}
		if !stmts[a.Conclusion.Statement] {
			return invalid("argument %q has an undeclared conclusion", a.Id)
		}
		for _, p := range a.Premises.Content {
			if !stmts[p.Statement] {
				return invalid("argument %q has an undeclared premise", a.Id)
			}
			switch p.Polarity {
			case "positive", "negative":
			default:
				return unsupported("argument %q needs explicit positive/negative premise polarity", a.Id)
			}
			switch p.Type {
			case "ordinary", "assumption", "exception":
			default:
				return unsupported("argument %q needs an explicit ordinary/assumption/exception premise type", a.Id)
			}
		}
	}
	for _, a := range g.Arguments.Content {
		for index, p := range a.Premises.Content {
			if p.Type == "exception" {
				id := fmt.Sprintf("%s.%d", a.Id, index)
				if ids[id] {
					return invalid("generated exception argument %q collides with a declared ID", id)
				}
				ids[id] = true
			}
		}
	}
	return nil
}
