package language

import "testing"

func TestTemplateRolesAndLiteralValues(t *testing.T) {
	for _, tc := range []struct {
		name, template string
		args           []string
		want           string
	}{
		{"ordered", "%s brings about %s", []string{"agent", "act"}, "agent brings about act"},
		{"indexed", "%[2]s follows %[1]s", []string{"cause", "effect"}, "effect follows cause"},
		{"mixed", "%[2]s / %s / %[1]s", []string{"one", "two", "three"}, "two / three / one"},
		{"repeated", "%[1]s repeats %[1]s", []string{"word"}, "word repeats word"},
		{"literal", "100%%: %s and %s", []string{"literal %s", "literal %[1]s"}, "100%: literal %s and literal %[1]s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Render(tc.template, tc.args)
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestTemplateRejectsInvalidOrMissingRoles(t *testing.T) {
	for _, template := range []string{"%[0]s %s", "%[3]s %s", "%[x]s %s", "%[1]s", "%s %v", "%s %s %s", "%s %", "%s %[1]s"} {
		if err := Validate(template, 2); err == nil {
			t.Errorf("accepted invalid template %q", template)
		}
	}
}

func TestTemplateHugeDeclaredArityDoesNotAllocateByArity(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if err := Validate("%[1]s", maxInt); err == nil {
		t.Fatal("huge incomplete argument coverage was accepted")
	}
}
