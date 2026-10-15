// Package data embeds WAG's shipped game data: the Carneades scheme YAML and
// the three NASCQ files. The CLI and the tests load these bytes, so there is
// one copy of the data to edit (the CLI and three test files used to carry
// their own copies, and they had drifted apart).
package data

import "embed"

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

// ContextsJSON holds the context presets (contexts.json): proof standard,
// evidence and scheme weights for kinds of discussion.
//
//go:embed contexts.json
var ContextsJSON []byte

// Annotation holds the annotation datasets the dashboard and wag-llm offer
// (annotation/<id>/dataset.json, its items, guidelines, gold) and the
// scheme collection annotators are given (annotation/walton.yml). They are
// copies of docs/kialo-batch2 and docs/scheme-registry/walton-clean/walton.yml;
// a test keeps them equal.
//
//go:embed annotation
var Annotation embed.FS
