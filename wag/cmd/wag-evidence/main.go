// wag-evidence validates an offline evidence snapshot; it never fetches content
// or changes a game's graph. See docs/evidence-contract.md for the exact scope.
// With --record-set, the file is a game's record set (docs/e1-evidence.md),
// checked alone: with the game's actors and a placeholder context of two
// players, as a game would take it.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/wag/internal/evidence"
)

func main() {
	root := flag.String("root", "", "payload root (default: bundle's directory)")
	recordSet := flag.Bool("record-set", false, "the file is a game's record set (NAME/records.json): checked alone, with a placeholder context of two players")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: wag-evidence [-root DIR] [--record-set] bundle.json")
		os.Exit(2)
	}
	path := flag.Arg(0)
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	var b *evidence.Bundle
	if *recordSet {
		set, err := evidence.DecodeRecordSet(raw)
		if err != nil {
			fmt.Fprintln(os.Stderr, "record set:", err)
			os.Exit(1)
		}
		// A placeholder context: two players and pins of zeros, which a game
		// replaces with its own.
		zero := strings.Repeat("0", 64)
		b, err = evidence.GameBundle(set, []evidence.Participant{{ID: "A", Kind: "human", Label: "placeholder player A"}, {ID: "B", Kind: "human", Label: "placeholder player B"}}, zero, zero, zero)
		if err != nil {
			fmt.Fprintln(os.Stderr, "record set:", err)
			os.Exit(1)
		}
	} else if b, err = evidence.Decode(raw); err != nil {
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
