package catalogue

import "testing"

func TestIndexedThreatTemplate(t *testing.T) {
	lang := map[string]string{"assert_bad/2": "I hereby assert that in fact I will see to it that %[2]s occurs if you bring about %[1]s."}
	got := English("assert_bad(X,Y)", lang)
	want := "I hereby assert that in fact I will see to it that Y occurs if you bring about X."
	if got != want {
		t.Fatalf("%q != %q", got, want)
	}
}

func TestLiteralBindingIsNotASecondTemplate(t *testing.T) {
	got := EnglishWith("p(X,Y)", map[string]string{"p/2": "%s then %s"}, map[string]string{"X": "literal %s", "Y": "next"})
	want := "“Literal %s” then “next”."
	if got != want {
		t.Fatalf("%q != %q", got, want)
	}
}
