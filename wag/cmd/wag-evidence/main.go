// wag-evidence validates an offline evidence snapshot; it never fetches content
// or changes a game's graph. See docs/evidence-contract.md for the exact scope.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourorg/wag/internal/evidence"
)

func main() {
	root := flag.String("root", "", "payload root (default: bundle's directory)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: wag-evidence [-root DIR] bundle.json")
		os.Exit(2)
	}
	path := flag.Arg(0)
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	b, err := evidence.Decode(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "evidence JSON:", err)
		os.Exit(1)
	}
	if *root == "" {
		*root = filepath.Dir(path)
	}
	r := evidence.Validate(b, *root)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !r.Valid {
		os.Exit(1)
	}
}
