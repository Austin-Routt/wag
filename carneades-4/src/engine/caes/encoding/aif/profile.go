// Copyright © 2026 The Carneades Authors
// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can obtain
// one at http://mozilla.org/MPL/2.0/.

package aif

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/carneades/carneades-4/src/engine/caes"
)

// encoding/json otherwise uses the last value for a repeated key, even with
// DisallowUnknownFields. Struct-field matching is also case-insensitive, so
// reject equivalent keys under Unicode simple folding before typed decoding.
func rejectDuplicateKeys(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	var value func() error
	value = func() error {
		token, err := d.Token()
		if err != nil {
			return err
		}
		delim, compound := token.(json.Delim)
		if !compound {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]string{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				folded := strings.Map(func(r rune) rune {
					minimum := r
					for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
						if next < minimum {
							minimum = next
						}
					}
					return minimum
				}, name)
				if previous, duplicate := seen[folded]; duplicate {
					return fmt.Errorf("duplicate or case-equivalent JSON keys %q and %q", previous, name)
				}
				seen[folded] = name
				if err := value(); err != nil {
					return err
				}
			}
		case '[':
			for d.More() {
				if err := value(); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected delimiter %q", delim)
		}
		_, err = d.Token()
		return err
	}
	if err := value(); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected one JSON object")
	}
	return nil
}

// ValidateProfile checks the deliberately narrow structural import profile.
// It is not an AIF/AIFdb conformance checker. RA nodes become linked arguments;
// each binary CA node is interpreted as a symmetric incompatibility issue.
// This policy does not implement general directed AIF conflict, dialogical
// nodes, preferences, or source-scheme instantiation.
func (ag AIF) ValidateProfile() error {
	invalid := func(format string, args ...interface{}) error {
		return caes.NewRunError(caes.ErrInvalidInput, "AIF: "+format, args...)
	}
	unsupported := func(format string, args ...interface{}) error {
		return caes.NewRunError(caes.ErrUnsupported, "AIF profile: "+format, args...)
	}
	if ag.Nodes == nil || ag.Edges == nil {
		return invalid("nodes and edges arrays are required (use [] for an empty array)")
	}
	nodes := make(map[string]Node, len(ag.Nodes))
	for _, n := range ag.Nodes {
		if n.Id == "" {
			return invalid("node ID is empty")
		}
		if _, duplicate := nodes[n.Id]; duplicate {
			return invalid("duplicate node ID %q", n.Id)
		}
		switch n.Type {
		case "I", "RA", "CA":
		default:
			return unsupported("node %q has unsupported type %q; only I, RA and binary CA are accepted", n.Id, n.Type)
		}
		nodes[n.Id] = n
	}
	in, out := map[string][]string{}, map[string][]string{}
	ids := map[string]bool{}
	pairs := map[[2]string]bool{}
	for _, e := range ag.Edges {
		if e.Id == "" || ids[e.Id] {
			return invalid("empty or duplicate edge ID %q", e.Id)
		}
		ids[e.Id] = true
		from, fromOK := nodes[e.From]
		to, toOK := nodes[e.To]
		if !fromOK || !toOK {
			return invalid("edge %q refers to an undeclared node", e.Id)
		}
		pair := [2]string{e.From, e.To}
		if pairs[pair] {
			return invalid("duplicate relation %q -> %q", e.From, e.To)
		}
		pairs[pair] = true
		if !(from.Type == "I" && (to.Type == "RA" || to.Type == "CA") ||
			(from.Type == "RA" || from.Type == "CA") && to.Type == "I") {
			return unsupported("edge %q has unsupported incidence %s -> %s", e.Id, from.Type, to.Type)
		}
		in[e.To] = append(in[e.To], e.From)
		out[e.From] = append(out[e.From], e.To)
	}
	positionOf := map[string]string{}
	for _, n := range ag.Nodes {
		switch n.Type {
		case "RA":
			if len(in[n.Id]) == 0 || len(out[n.Id]) != 1 {
				return invalid("RA node %q needs at least one I premise and exactly one I conclusion", n.Id)
			}
		case "CA":
			if len(in[n.Id]) != 1 || len(out[n.Id]) != 1 || in[n.Id][0] == out[n.Id][0] {
				return unsupported("CA node %q must connect two distinct I positions, one incoming and one outgoing", n.Id)
			}
			for _, p := range []string{in[n.Id][0], out[n.Id][0]} {
				if previous := positionOf[p]; previous != "" {
					return unsupported("I node %q participates in multiple conflicts (%q, %q)", p, previous, n.Id)
				}
				positionOf[p] = n.Id
			}
		}
	}
	return nil
}
