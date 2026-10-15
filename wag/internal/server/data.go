package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourorg/wag/data"
)

// ReadData reads the game data from a directory laid out as wag/data
// (schemes/core.yml, nascq/*.json, contexts.json); from the data built into
// the program if dir is "".
func ReadData(dir string) (Data, error) {
	d := Data{YAML: data.CoreYAML, Schemes: data.SchemesJSON, CQs: data.CQsJSON, Edges: data.EdgesJSON, Contexts: data.ContextsJSON}
	if dir == "" {
		return d, nil
	}
	for _, x := range []struct {
		dst  *[]byte
		path string
	}{{&d.YAML, "schemes/core.yml"}, {&d.Schemes, "nascq/schemes.json"}, {&d.CQs, "nascq/critical_questions.json"}, {&d.Edges, "nascq/edges.json"}, {&d.Contexts, "contexts.json"}} {
		b, err := os.ReadFile(filepath.Join(dir, x.path))
		if err != nil {
			return d, fmt.Errorf("reading the data: %v", err)
		}
		*x.dst = b
	}
	return d, nil
}
