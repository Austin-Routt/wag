package caes

import (
	"math"
	"testing"
)

// Preserve the shared WAG/Kialo fork's empty-premise behavior. Returning NaN
// makes a premise-less cumulative argument fail to support its conclusion.
func TestCumulativeEmptyPremises(t *testing.T) {
	ag := NewArgGraph()
	conclusion := NewStatement()
	conclusion.Id = "conclusion"
	ag.Statements[conclusion.Id] = conclusion
	arg := NewArgument()
	arg.Id, arg.Conclusion = "a1", conclusion
	arg.Scheme = &Scheme{Id: "empty_cumulative", Weight: CumulativeWeighingFunction}
	ag.Arguments[arg.Id] = arg
	conclusion.Args = []*Argument{arg}
	if got := CumulativeWeighingFunction(arg, NewLabelling()); math.IsNaN(got) || got != 1 {
		t.Fatalf("empty cumulative weight = %v, want 1", got)
	}
	labels := ag.GroundedLabelling()
	ag.ApplyLabelling(labels)
	if conclusion.Label != In || arg.Weight != 1 {
		t.Fatalf("empty cumulative support: conclusion=%s weight=%v, want in/1", conclusion.Label, arg.Weight)
	}
}

func TestCumulativePremiseFraction(t *testing.T) {
	p, q := NewStatement(), NewStatement()
	arg := NewArgument()
	arg.Premises = []Premise{{Stmt: p}, {Stmt: q}}
	for _, tc := range []struct {
		name string
		p, q Label
		want float64
	}{
		{"none", Out, Out, 0}, {"one", In, Out, .5}, {"all", In, In, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CumulativeWeighingFunction(arg, Labelling{p: tc.p, q: tc.q}); got != tc.want {
				t.Fatalf("fraction = %v, want %v", got, tc.want)
			}
		})
	}
}
