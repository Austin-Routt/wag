package tgf

import (
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/dung"
)

func TestStrictAttackOnlyProfile(t *testing.T) {
	for _, input := range []string{
		"a\na\n#\n", "a\nb\n", "a\n# extra\n", "a\n#\na\n",
		"a\n#\na missing\n", "a\nb\n#\na b support:1\n", "a\nb\n#\na b attack\n",
	} {
		if _, err := Import(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted unsupported or malformed TGF: %q", input)
		}
	}
	af, err := Import(strings.NewReader("a Optional node text\nb\n#\na b\na b"))
	if err != nil || len(af.Args()) != 2 || len(af.Atks()[dung.Arg("b")]) != 1 {
		t.Fatalf("valid AF or duplicate-edge set semantics failed: %v, %v", af, err)
	}
	if !af.GroundedExtension().Equals(dung.NewArgSet(dung.Arg("a"))) {
		t.Fatal("attack orientation changed")
	}
	if empty, err := Import(strings.NewReader("#\n")); err != nil || len(empty.Args()) != 0 {
		t.Fatalf("explicit empty AF rejected: %v", err)
	}
}
