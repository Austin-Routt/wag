package language

import "testing"

func TestNamesNothing(t *testing.T) {
	for v, want := range map[string]bool{
		"": true, "   ": true, "?": true, "--": true, "\t\n": true,
		string(rune(0x200b)): true, string(rune(0xfeff)): true, " " + string(rune(0x200b)) + " ": true,
		"a": false, "H": false, "pilot study": false, "2030": false, "税金": false, " x ": false, "?a": false,
	} {
		if got := NamesNothing(v); got != want {
			t.Errorf("NamesNothing(%q) = %v, want %v", v, got, want)
		}
	}
}
