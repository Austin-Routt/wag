package evidence

import (
	"strings"
	"testing"
)

// X5 (docs/e1-evidence.md, E1.7, A6; the outcome written before the test
// ran): the records-only check checks a selector's structure, as the full
// check does, without reading its bytes: a selector of an unknown kind, and
// a text selector whose end precedes its start, are refused, naming the
// claim's selector.
func TestX5TheRecordsOnlyCheckChecksTheSelectors(t *testing.T) {
	for name, edit := range map[string]func(*Selector){
		"a selector of kind line": func(s *Selector) { s.Kind = "line" },
		"a text selector's end before its start": func(s *Selector) {
			s.Start, s.End = 5, 2
		},
	} {
		t.Run(name, func(t *testing.T) {
			b, root := fixture(t)
			edit(&b.Claims[0].Selector)
			for check, r := range map[string]Report{"records only": ValidateRecords(b), "full": Validate(b, root)} {
				named := false
				for _, e := range r.Errors {
					named = named || (e.Code == "selector" && e.Location == b.Claims[0].ID)
				}
				if r.Valid || !named {
					t.Fatalf("%s: valid %v, errors %+v", check, r.Valid, r.Errors)
				}
				if check == "records only" && strings.Contains(firstText(r), "payload") {
					t.Fatalf("records only read a payload: %+v", r.Errors)
				}
			}
		})
	}
}

// firstText is the report's first error, as text.
func firstText(r Report) string {
	if len(r.Errors) == 0 {
		return ""
	}
	return r.Errors[0].Code + " " + r.Errors[0].Message
}
