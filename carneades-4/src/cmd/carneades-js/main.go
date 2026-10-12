//go:build js

// Fork note: GopherJS command. The compile fix (pointer arguments for the
// GraphML and DOT exporters, the dot package import) comes from the Walton
// Argumentation Game's bundled copy. The build constraint stays, because
// ordinary builds and tests do not have the GopherJS toolchain.

package main

import (
	"bytes"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/dot"      // <-- added
	"github.com/carneades/carneades-4/src/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/gopherjs/gopherjs/js"
)

// GraphMLExportToString exports an argument graph to GraphML format.
// It accepts a pointer to caes.ArgGraph.
func GraphMLExportToString(ag *caes.ArgGraph) (string, error) {
	var b bytes.Buffer
	err := graphml.Export(&b, ag)
	return b.String(), err
}

// DotExportToString exports an argument graph to DOT format.
// It accepts a pointer to caes.ArgGraph.
func DotExportToString(ag *caes.ArgGraph) (string, error) {
	var b bytes.Buffer
	err := dot.Export(&b, ag)   // using the correct dot package
	return b.String(), err
}

func YamlImportFromString(inString string) (*caes.ArgGraph, error) {
	return yaml.Import(bytes.NewBufferString(inString))
}

func YamlExportToString(ag *caes.ArgGraph) string {
	var b bytes.Buffer
	yaml.Export(&b, ag)
	return b.String()
}

func main() {
	js.Global.Set("carneades", map[string]interface{}{
		"eval": evalJs,
	})
}

func evalJs(input string, inputFormat string, outputFormat string) map[string]interface{} {
	var ag *caes.ArgGraph
	var err error
	var str string

	switch inputFormat {
	case "yaml":
		ag, err = YamlImportFromString(input)
	default:
		return map[string]interface{}{
			"err": "invalid input format",
		}
	}

	if err != nil {
		return map[string]interface{}{
			"err": "error trying to import.",
		}
	}

	switch outputFormat {
	case "yaml":
		str = YamlExportToString(ag)   // ag is already a pointer
		err = nil
	case "graphml":
		str, err = GraphMLExportToString(ag)   // pass pointer directly
	case "dot":
		str, err = DotExportToString(ag)       // pass pointer directly
	default:
		return map[string]interface{}{
			"err": "invalid output format",
		}
	}
	if err != nil {
		return map[string]interface{}{
			"err": "error trying to export.",
		}
	} else {
		return map[string]interface{}{
			"result": str,
		}
	}
}
