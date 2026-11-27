// Package manifest names a game's inputs by their content (E1,
// docs/e1-evidence.md): the scheme file the game reads, its NASCQ data and the
// Carneades engine compiled into it, each by a full SHA-256, as a
// wag-evidence/0.1 context pins them (engine_sha256, catalogue_sha256,
// nascq_sha256). A pin is an integrity check relative to these bytes: it
// says nothing of what the inputs mean.
package manifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Pins are a game's input fingerprints.
type Pins struct {
	Engine    string `json:"engine_sha256"`
	Catalogue string `json:"catalogue_sha256"` // the scheme file the game reads (core.yml), which the catalogue generates
	NASCQ     string `json:"nascq_sha256"`     // the schemes, questions and edges
}

// Of pins the game's inputs: its scheme file and NASCQ files as it reads
// them, and this build's engine.
func Of(schemeFile, schemes, cqs, edges []byte) Pins {
	return Pins{Engine: EngineSHA256, Catalogue: SchemeFile(schemeFile), NASCQ: NASCQ(schemes, cqs, edges)}
}

// SchemeFile is the scheme file's fingerprint, its line endings read as LF
// (as the game reads them).
func SchemeFile(b []byte) string {
	h := sha256.Sum256(lf(b))
	return hex.EncodeToString(h[:])
}

// NASCQ is the fingerprint of the schemes, questions and edges, in that
// order, each read with LF line endings and followed by a NUL byte.
func NASCQ(schemes, cqs, edges []byte) string {
	h := sha256.New()
	for _, b := range [][]byte{schemes, cqs, edges} {
		h.Write(lf(b))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// EngineDirs are the engine's source folders, relative to the repository's
// root: the vendored Carneades engine and the CHR library it uses.
var EngineDirs = []string{"carneades-4/src/engine", "GoCHR/src/engine"}

// EngineFiles are the engine's module files, relative to the root.
var EngineFiles = []string{"carneades-4/go.mod", "GoCHR/go.mod"}

// EngineSources is the fingerprint of the engine's sources under root (the
// repository's root): every Go file of EngineDirs that is not a test, and
// EngineFiles, by path relative to root (with forward slashes), sorted; for
// each, its path, a NUL, its bytes with line endings read as LF, a NUL.
func EngineSources(root string) (string, error) {
	var paths []string
	for _, d := range EngineDirs {
		err := filepath.WalkDir(filepath.Join(root, d), func(p string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !e.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
				rel, err := filepath.Rel(root, p)
				if err != nil {
					return err
				}
				paths = append(paths, filepath.ToSlash(rel))
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	paths = append(paths, EngineFiles...)
	sort.Strings(paths)
	h := sha256.New()
	for _, p := range paths {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			return "", err
		}
		h.Write([]byte(p))
		h.Write([]byte{0})
		h.Write(lf(b))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func lf(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }
