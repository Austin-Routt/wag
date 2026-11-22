package caf

import (
	"strings"
	"testing"
)

const smallCAF = `<caf version="1.3"><statements>
<statement id="p"><descriptions><description lang="en">Premise</description></descriptions></statement>
<statement id="c"><descriptions><description lang="en">Conclusion</description></descriptions></statement>
</statements><arguments><argument id="a"><conclusion statement="c"/><premises><premise statement="p"/></premises></argument></arguments></caf>`

func TestCAF13DefaultsAndTexts(t *testing.T) {
	g, err := Import(strings.NewReader(smallCAF))
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Assumptions) != 0 || len(g.Statements) != 2 || len(g.Issues) != 0 {
		t.Fatal("omitted weight must not manufacture negative assumptions")
	}
	a := g.Arguments["a1"]
	if a.Conclusion.Id != "s2" || a.Premises[0].Stmt.Id != "s1" || a.Conclusion.Text != "Conclusion" || a.Premises[0].Stmt.Text != "Premise" {
		t.Fatal("omitted polarity must default to positive; preserve descriptions")
	}
	if a.Scheme == nil || a.Scheme.Id != "linked" {
		t.Fatal("external scheme names do not bind library schemes")
	}
}

func TestCAF13ExplicitNegativeValuesRemainNegative(t *testing.T) {
	src := strings.Replace(smallCAF, `<statement id="p">`, `<statement id="p" weight="0">`, 1)
	src = strings.Replace(src, `<argument id="a">`, `<argument id="a" pro="false">`, 1)
	src = strings.Replace(src, `<premise statement="p"/>`, `<premise statement="p" positive="false"/>`, 1)
	g, err := Import(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	a := g.Arguments["a1"]
	if len(g.Assumptions) != 1 || g.Assumptions[0] != "¬s1" || a.Premises[0].Stmt.Id != "¬s1" || a.Conclusion.Id != "¬s2" {
		t.Fatal("explicit negative signs/zero weight were overwritten by defaults")
	}
}

func TestCAFRejectsUnreviewedShapes(t *testing.T) {
	for _, bad := range []string{
		strings.Replace(smallCAF, `version="1.3"`, `version="4"`, 1),
		strings.Replace(smallCAF, `<statement id="p">`, `<statement id="p"><text>Unsupported direct text</text>`, 1),
		strings.Replace(smallCAF, `<argument id="a">`, `<argument id="a" direction="con">`, 1),
		strings.Replace(smallCAF, `<argument id="a">`, `<argument id="a" strict="true">`, 1),
		strings.Replace(smallCAF, `statement="p"`, `statement="missing"`, 1),
		strings.Replace(smallCAF, `id="c"`, `id="p"`, 1),
	} {
		if g, err := Import(strings.NewReader(bad)); err == nil || g != nil {
			t.Fatalf("unsupported/malformed CAF was accepted: %s", bad)
		}
	}
}

func TestCAFFowardUndercutterDoesNotPanic(t *testing.T) {
	src := `<caf version="1.3"><statements><statement id="p"/><statement id="c"/><statement id="u" atom="(undercut target)"/></statements><arguments>
<argument id="attack"><conclusion statement="u"/><premises><premise statement="p"/></premises></argument>
<argument id="target"><conclusion statement="c"/><premises><premise statement="p"/></premises></argument>
</arguments></caf>`
	g, err := Import(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Arguments) != 2 {
		t.Fatal("forward-reference placeholder was lost")
	}
	found := false
	for _, a := range g.Arguments {
		if a.Undercutter != nil {
			found = true
		}
	}
	if !found {
		t.Fatal("undercutter link was lost")
	}
}
