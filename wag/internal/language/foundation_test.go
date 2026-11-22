package language

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Compatibility with the pinned source library's indexed dialect. This checks
// role ordering against Go fmt, not the source interpretation of each scheme.
func TestPinnedFoundationIndexedTemplates(t *testing.T) {
	raw, err := os.ReadFile("testdata/foundation_indexed_templates.json")
	if err != nil {
		t.Fatal(err)
	}
	var templates map[string]string
	if err := json.Unmarshal(raw, &templates); err != nil {
		t.Fatal(err)
	}
	if len(templates) != 24 {
		t.Fatal("pinned fixture changed without provenance update")
	}
	for key, template := range templates {
		t.Run(key, func(t *testing.T) {
			arity, err := strconv.Atoi(key[strings.LastIndex(key, "/")+1:])
			if err != nil {
				t.Fatal(err)
			}
			args := make([]string, arity)
			values := make([]interface{}, arity)
			for i := range args {
				args[i] = fmt.Sprintf("ROLE_%d", i+1)
				values[i] = args[i]
			}
			got, err := Render(template, args)
			want := fmt.Sprintf(template, values...)
			if err != nil || got != want {
				t.Fatalf("got %q, %v; want %q", got, err, want)
			}
		})
	}
}
