package carneades

import "testing"

// Other's free text that names nothing (no letter or digit) is refused by
// both of the adapter's ways of making an argument of it, as S2 refuses it
// first in the game (the check of step 8, W-4); a blank later premise is
// left out. A Kialo discussion's claims are not checked (AddFreeArgumentOn).
func TestFreeTextThatNamesNothing(t *testing.T) {
	zw := string(rune(0x200b))
	for _, c := range []struct {
		premises []string
		concl    string
	}{
		{[]string{"?"}, "tax meat"},
		{[]string{"lower emissions", "--"}, "tax meat"},
		{[]string{"lower emissions"}, "?"},
		{[]string{zw}, "tax meat"},
	} {
		a := c3(t)
		before := len(a.Graph.Arguments)
		if ids, err := a.AddFreeArgument("pro", c.premises, c.concl, 0.5); err == nil {
			t.Errorf("AddFreeArgument %q -> %q: accepted as %v", c.premises, c.concl, ids)
		}
		if len(a.Graph.Arguments) != before {
			t.Errorf("AddFreeArgument %q -> %q: the refusal left an argument", c.premises, c.concl)
		}
	}
	for _, prem := range [][]string{{"the fall is tiny"}, {"?"}, {"the fall is tiny", "--"}, {zw}} {
		a := c3(t)
		target := c3arg(t, a, "pro", "end_and_means_1", map[string]string{"X": "tax_meat", "Y": "meat_scarce"}, "good(tax_meat)")
		before := len(a.Graph.Arguments)
		_, err := a.AnswerQuestion(Answer{Owner: "con", SchemeID: FreeSchemeID, TargetArgID: target, Locus: LocusArgument, FreePremises: prem, Weight: 0.5})
		if named := prem[0] == "the fall is tiny" && len(prem) == 1; named != (err == nil) || named != (len(a.Graph.Arguments) > before) {
			t.Errorf("an Other answer with the premises %q: %v, %d arguments added", prem, err, len(a.Graph.Arguments)-before)
		}
	}
	a := c3(t)
	if _, err := a.AddFreeArgument("pro", []string{"lower emissions", ""}, "tax meat", 0.5); err != nil {
		t.Fatalf("a blank later premise: %v", err)
	}
}
