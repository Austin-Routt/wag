package carneades

// Item F's cases for schemes the game does not play in C5.2
// (docs/c5-item-f.md, F): the engine on walton.yml's schemes, each outcome
// written in that page before the test ran.

import (
	"strings"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// F-continuing: the better-means exception applies to the bridge from a
// continuing effort. The bridge stands on its premises; once
// realize_more_effectively(P,G,A) is In, it is out of force.
func TestItemFContinuingBridgeBetterMeans(t *testing.T) {
	a := c3(t)
	const better = "realize_more_effectively(decider,emissions_fall,tax_meat)"
	c3assume(a, "ought_to_continue(decider,emissions_fall)", "continues_trying(tax_meat,emissions_fall)",
		"deliberating_agent(decider,tax_meat)", "possible_bring_about(decider,tax_meat)")
	br := c3arg(t, a, "pro", "bridge_from_continuing", map[string]string{"P": "decider", "G": "emissions_fall", "A": "tax_meat"},
		"should_be_performed(tax_meat)")
	if got := c3exceptions(a, br); !strings.Contains(strings.Join(got, " "), better) {
		t.Fatalf("the bridge's exceptions are %v, without %s", got, better)
	}
	a.Evaluate()
	if c3undercut(a, br) == caes.In || c3lab(a, "should_be_performed(tax_meat)") != caes.In {
		t.Fatalf("before the better means: undercut %v, should_be_performed %v", c3undercut(a, br), c3lab(a, "should_be_performed(tax_meat)"))
	}
	c3argue(t, a, "con", better)
	a.Evaluate()
	if c3undercut(a, br) != caes.In || c3lab(a, "should_be_performed(tax_meat)") == caes.In {
		t.Fatalf("after the better means: undercut %v, should_be_performed %v", c3undercut(a, br), c3lab(a, "should_be_performed(tax_meat)"))
	}
}

// G-endandmeans: the two end-and-means schemes weigh in mirrored directions.
// end_and_means_bad (a bad end makes its means bad) is defeated by
// means_good_outweighs, end_and_means_1 (a good end makes its means good) by
// means_bad_outweighs, and neither by the other's exception; walton.yml's
// indexed sentences of the two exceptions put each place where it belongs.
func TestItemFEndAndMeansMirrored(t *testing.T) {
	for _, c := range []struct {
		scheme, premise, concl, own, other, text string
	}{
		{"end_and_means_bad", "bad(meat_scarce)", "bad(tax_meat)", "means_good_outweighs(tax_meat,meat_scarce)",
			"means_bad_outweighs(tax_meat,meat_scarce)",
			"As a means to the bad end meat_scarce, tax_meat would have good consequences that matter more than that end."},
		{"end_and_means_1", "good(meat_scarce)", "good(tax_meat)", "means_bad_outweighs(tax_meat,meat_scarce)",
			"means_good_outweighs(tax_meat,meat_scarce)",
			"As a means to the good end meat_scarce, tax_meat would have bad consequences that matter more than that end."},
	} {
		t.Run(c.scheme, func(t *testing.T) {
			a := c3(t)
			c3assume(a, "means_to_end(tax_meat,meat_scarce)", c.premise)
			arg := c3arg(t, a, "pro", c.scheme, map[string]string{"X": "tax_meat", "Y": "meat_scarce"}, c.concl)
			if s := a.Graph.Statements[c.own]; s == nil || s.Text != c.text {
				t.Fatalf("the exception %s reads %+v, want %q", c.own, s, c.text)
			}
			a.Evaluate()
			if c3undercut(a, arg) == caes.In || c3lab(a, c.concl) != caes.In {
				t.Fatalf("unanswered: undercut %v, %s %v", c3undercut(a, arg), c.concl, c3lab(a, c.concl))
			}
			c3argue(t, a, "con", c.other)
			a.Evaluate()
			if c3undercut(a, arg) == caes.In || c3lab(a, c.concl) != caes.In {
				t.Fatalf("the other direction's exception undercut it: %v, %s %v", c3undercut(a, arg), c.concl, c3lab(a, c.concl))
			}
			c3argue(t, a, "con", c.own)
			a.Evaluate()
			if c3undercut(a, arg) != caes.In || c3lab(a, c.concl) == caes.In {
				t.Fatalf("its own exception did not undercut it: %v, %s %v", c3undercut(a, arg), c.concl, c3lab(a, c.concl))
			}
		})
	}
}

// G-unbound: an argument with X or Y unbound is refused; the engine does not
// read the unbound variable as a wildcard.
func TestItemFEndAndMeansUnbound(t *testing.T) {
	for _, b := range []map[string]string{{"X": "tax_meat"}, {"Y": "meat_scarce"}, {"X": "tax_meat", "Y": ""}} {
		a := c3(t)
		before := len(a.Graph.Arguments)
		ids, err := a.AddSchemeArgument("pro", "end_and_means_bad", b, "bad(tax_meat)", w)
		if err == nil {
			t.Errorf("%v: accepted as %v", b, ids)
		}
		if len(a.Graph.Arguments) != before {
			t.Errorf("%v: the refusal left %d arguments", b, len(a.Graph.Arguments)-before)
		}
	}
}
