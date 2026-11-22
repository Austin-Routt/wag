package caf

import (
	"encoding/xml"
	"math"

	"github.com/carneades/carneades-4/src/engine/caes"
)

// CAF 1.3 explicitly defaults statement weight to 0.5 and argument/premise
// polarity to positive. Go zero values invert all three meanings. An absent
// evaluation value is unknown, not a computed false value.
func (s *Statement) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Statement
	v := plain{Weight: 0.5, Value: 0.5}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	*s = Statement(v)
	return nil
}

func (a *Argument) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Argument
	v := plain{Pro: true, Weight: 0.5, Value: 0.5}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	*a = Argument(v)
	return nil
}

func (p *Premise) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type plain Premise
	v := plain{Positive: true}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	*p = Premise(v)
	return nil
}

func (g *CAF) validateProfile() error {
	invalid := func(f string, a ...interface{}) error { return caes.NewRunError(caes.ErrInvalidInput, "CAF: "+f, a...) }
	unsupported := func(f string, a ...interface{}) error {
		return caes.NewRunError(caes.ErrUnsupported, "experimental CAF: "+f, a...)
	}
	if g.Version != "" && g.Version != "1.3" {
		return unsupported("version %q is not CAF 1.3", g.Version)
	}
	if len(g.Unknown) != 0 || len(g.UnknownAttrs) != 0 {
		return unsupported("unrecognized root elements or attributes")
	}
	if len(g.Statements.Content) == 0 {
		return invalid("at least one statement is required")
	}
	ids, stmts, args := map[string]bool{}, map[string]bool{}, map[string]bool{}
	unit := func(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) && x >= 0 && x <= 1 }
	for _, s := range g.Statements.Content {
		if s.Id == "" || ids[s.Id] {
			return invalid("empty or duplicate ID %q", s.Id)
		}
		ids[s.Id], stmts[s.Id] = true, true
		if len(s.Unknown) != 0 || len(s.UnknownAttrs) != 0 {
			return unsupported("statement %q has unrecognized fields; text belongs in descriptions/description", s.Id)
		}
		if !unit(s.Weight) || !unit(s.Value) {
			return invalid("statement %q has a weight/value outside [0,1]", s.Id)
		}
		switch s.Standard {
		case "", "PE", "CCE", "BRD":
		default:
			return unsupported("proof standard %q", s.Standard)
		}
	}
	for _, a := range g.Arguments.Content {
		if a.Id == "" || ids[a.Id] {
			return invalid("empty or duplicate ID %q", a.Id)
		}
		ids[a.Id], args[a.Id] = true, true
		if a.Strict {
			return unsupported("strict argument %q has no reviewed strict-rule mapping", a.Id)
		}
		if len(a.Unknown) != 0 || len(a.UnknownAttrs) != 0 {
			return unsupported("argument %q has unrecognized fields; polarity is the pro attribute", a.Id)
		}
		if !unit(a.Weight) || !unit(a.Value) {
			return invalid("argument %q has a weight/value outside [0,1]", a.Id)
		}
		if !stmts[a.Conclusion.Statement] {
			return invalid("argument %q has an undeclared conclusion", a.Id)
		}
		for _, p := range a.Premises.Content {
			if len(p.Unknown) != 0 || len(p.UnknownAttrs) != 0 {
				return unsupported("argument %q has an unrecognized premise field", a.Id)
			}
			if !stmts[p.Statement] {
				return invalid("argument %q has an undeclared premise", a.Id)
			}
		}
	}
	for _, s := range g.Statements.Content {
		match := undercutRegExp.FindStringSubmatch(s.Atom)
		if len(match) == 3 && !args[match[2]] {
			return invalid("undercutter statement %q refers to an undeclared argument", s.Id)
		}
	}
	return nil
}
