// Package data embeds WAG's shipped game data: the Carneades scheme YAML and
// the three NASCQ files. The CLI and the tests load these bytes, so there is
// one copy of the data to edit (the CLI and three test files used to carry
// their own copies, and they had drifted apart).
package data

import _ "embed"

// CoreYAML is the Carneades-4 scheme file (schemes/core.yml).
//
//go:embed schemes/core.yml
var CoreYAML []byte

// SchemesJSON holds the NASCQ argumentation-scheme nodes (nascq/schemes.json).
//
//go:embed nascq/schemes.json
var SchemesJSON []byte

// CQsJSON holds the NASCQ critical-question nodes (nascq/critical_questions.json).
//
//go:embed nascq/critical_questions.json
var CQsJSON []byte

// EdgesJSON holds the NASCQ ask and backup edges (nascq/edges.json).
//
//go:embed nascq/edges.json
var EdgesJSON []byte
