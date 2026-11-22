package terms

// NormalizeNegation cancels pairs of explicit negations structurally. It does
// not evaluate arithmetic, coerce values, or reinterpret quoted atoms/strings.
// Children are normalized too, as in the rule store's explicit-negation mode.
func NormalizeNegation(t Term) Term {
	switch x := t.(type) {
	case Compound:
		args := make([]Term, len(x.Args))
		for i, a := range x.Args {
			args[i] = NormalizeNegation(a)
		}
		x.Args = args
		if x.Functor == "¬" && len(args) == 1 {
			if inner, ok := args[0].(Compound); ok && inner.Functor == "¬" && len(inner.Args) == 1 {
				return inner.Args[0]
			}
		}
		return x
	case List:
		out := make(List, len(x))
		for i, a := range x {
			out[i] = NormalizeNegation(a)
		}
		return out
	default:
		return t
	}
}

// ReadFormula uses the ordinary strict reader, then applies the declared
// explicit-negation identity. The syntactic term reader remains unchanged.
func ReadFormula(s string) (Term, bool) {
	t, ok := ReadString(s)
	if !ok {
		return t, false
	}
	return NormalizeNegation(t), true
}

// MatchFormula preserves one-way matching while normalizing the already bound
// parameters before comparison. In particular, X=¬p makes a pattern ¬X mean p.
// Unbound variables retain the ordinary Match behavior.
func MatchFormula(pattern, instance Term, env Bindings) (Bindings, bool) {
	// A variable occurring inside a matched value belongs to that value, not
	// to the pattern. Do not turn it into a newly bindable pattern variable.
	for _, v := range pattern.OccurVars() {
		if value, bound := GetBinding(v, env); bound && !Ground(value, nil) {
			return Match(NormalizeNegation(pattern), NormalizeNegation(instance), env)
		}
	}
	return Match(NormalizeNegation(Substitute(pattern, env)), NormalizeNegation(instance), env)
}
