package carneades

import (
	"os"
	"reflect"
	"testing"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// The causal exception names its effect (decisions D071 and D097, proposal
// D040-D071-D097, version 2; its effects say these tests before they ran):
// on walton.yml as changed, every argument at the same weight, an argument
// that interference holds for one effect of a cause undercuts the step to
// that effect and leaves a step from the same cause to another effect
// standing; a second argument for the blocked effect, from another cause,
// keeps it acceptable; the same for the modal form. On the frozen format-1
// walton.yml, the planted control: the one-place exception undercuts both.

type causalForm struct{ scheme, cause, effect string }

var causalForms = []causalForm{
	{"cause_to_effect", "has_occurred", "will_occur"},
	{"cause_to_effect_modal", "may_occur", "may_occur"},
}

func causalStep(t *testing.T, a *Adapter, f causalForm, owner, e1, e2 string) string {
	t.Helper()
	c3assume(a, f.cause+"("+e1+")", "causes("+e1+","+e2+")")
	return c3arg(t, a, owner, f.scheme, map[string]string{"E1": e1, "E2": e2}, f.effect+"("+e2+")")
}

func TestCausalExceptionNamesItsEffect(t *testing.T) {
	for _, f := range causalForms {
		t.Run(f.scheme, func(t *testing.T) {
			a := c3(t)
			harvest := causalStep(t, a, f, "p1", "drought", "harvest_failing")
			fires := causalStep(t, a, f, "p2", "drought", "wildfires")
			a.Evaluate()
			for _, s := range []string{f.effect + "(harvest_failing)", f.effect + "(wildfires)"} {
				if c3lab(a, s) != caes.In {
					t.Fatalf("%s is %v before any exception is argued", s, c3lab(a, s))
				}
			}
			if got := c3exceptions(a, harvest); !reflect.DeepEqual(got, []string{"interference(drought,harvest_failing)"}) {
				t.Fatalf("the step to the harvest has the exceptions %v", got)
			}
			c3argue(t, a, "critic", "interference(drought,harvest_failing)")
			a.Evaluate()
			if c3undercut(a, harvest) != caes.In || c3undercut(a, fires) == caes.In {
				t.Fatalf("interference with the harvest undercuts the harvest step %v (want In) and the fires step %v (want not In)",
					c3undercut(a, harvest), c3undercut(a, fires))
			}
			if c3lab(a, f.effect+"(harvest_failing)") == caes.In || c3lab(a, f.effect+"(wildfires)") != caes.In {
				t.Fatalf("after the answer: harvest %v (want not In), fires %v (want In)",
					c3lab(a, f.effect+"(harvest_failing)"), c3lab(a, f.effect+"(wildfires)"))
			}
			// A second argument for the blocked effect, from another cause,
			// is not touched: its exception names its own cause.
			heat := causalStep(t, a, f, "p3", "heatwave", "harvest_failing")
			a.Evaluate()
			if c3undercut(a, heat) == caes.In || c3lab(a, f.effect+"(harvest_failing)") != caes.In {
				t.Fatalf("the argument from the heatwave: undercut %v (want not In); harvest %v (want In)",
					c3undercut(a, heat), c3lab(a, f.effect+"(harvest_failing)"))
			}
		})
	}
}

// The planted control, on the walton.yml the format-1 gold is checked
// against (frozen): its one-place exception, interference(E1), is shared by
// the two steps, so an answer about one effect undercuts both.
func TestCausalOnePlaceExceptionUndercutsBoth(t *testing.T) {
	b, err := os.ReadFile("../../../docs/kialo-batch2/format1/walton.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range causalForms {
		t.Run(f.scheme, func(t *testing.T) {
			a, err := NewAdapterFromYAML(b)
			if err != nil {
				t.Fatal(err)
			}
			harvest := causalStep(t, a, f, "p1", "drought", "harvest_failing")
			fires := causalStep(t, a, f, "p2", "drought", "wildfires")
			if got := c3exceptions(a, harvest); !reflect.DeepEqual(got, []string{"interference(drought)"}) {
				t.Fatalf("the frozen walton.yml's step has the exceptions %v", got)
			}
			c3argue(t, a, "critic", "interference(drought)")
			a.Evaluate()
			if c3undercut(a, harvest) != caes.In || c3undercut(a, fires) != caes.In {
				t.Fatalf("the one-place exception undercuts the harvest step %v and the fires step %v (want both In)",
					c3undercut(a, harvest), c3undercut(a, fires))
			}
		})
	}
}
