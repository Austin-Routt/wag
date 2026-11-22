package terms

import "testing"

func TestNormalizeExplicitNegation(t *testing.T) {
	for _, tc := range []struct{ input, expected string }{
		{"¬¬p", "p"}, {"¬¬¬p", "¬p"}, {"¬¬¬¬p", "p"},
		{"box(¬¬p,[¬¬¬q,1,1.25,\"¬¬p\",'two words'])", "box(p,[¬q,1,1.25,\"¬¬p\",'two words'])"},
		{"p(-2,1.0000001)", "p(-2,1.0000001)"},
	} {
		got, ok := ReadFormula(tc.input)
		want, wok := ReadString(tc.expected)
		if !ok || !wok || !Equal(got, want) {
			t.Errorf("%s: got %v, expected %v", tc.input, got, want)
		}
		if !Equal(got, NormalizeNegation(got)) {
			t.Errorf("normalization not idempotent for %s", tc.input)
		}
	}
}

func TestFormulaMatchNormalizesBoundNegationWithoutCapture(t *testing.T) {
	x, y := NewVariable("X"), NewVariable("Y")
	pattern, _ := ReadString("¬X")
	negP, _ := ReadString("¬p")
	p, _ := ReadString("p")
	if _, ok := MatchFormula(pattern, p, AddBinding(x, negP, nil)); !ok {
		t.Fatal("X=¬p did not make ¬X match p")
	}
	pattern, _ = ReadString("box(X)")
	actual, _ := ReadString("box(a)")
	if _, ok := MatchFormula(pattern, actual, AddBinding(x, y, nil)); ok {
		t.Fatal("captured a variable inside an already bound value")
	}
}
