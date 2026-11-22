package dung

import "testing"

func TestExtensionFamilyEqualityIsSymmetric(t *testing.T) {
	empty := NewArgSet()
	a, b := NewArgSet(Arg("a")), NewArgSet(Arg("b"))
	for _, tc := range []struct {
		x, y []ArgSet
		want bool
	}{
		{nil, nil, true}, {nil, []ArgSet{empty}, false},
		{[]ArgSet{a}, []ArgSet{a, b}, false},
		{[]ArgSet{a, b}, []ArgSet{b, a}, true},
		{[]ArgSet{a, a}, []ArgSet{a}, true},
	} {
		if EqualArgSetSlices(tc.x, tc.y) != tc.want || EqualArgSetSlices(tc.y, tc.x) != tc.want {
			t.Fatalf("family comparison disagrees in one or both directions: %v, %v", tc.x, tc.y)
		}
	}
}
