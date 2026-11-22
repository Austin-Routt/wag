// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Import Dung AFs represented using the Trivial Graph Format
// See <https://en.wikipedia.org/wiki/Trivial_Graph_Format>
package tgf

import (
	"bufio"
	"fmt"
	"github.com/carneades/carneades-4/src/engine/dung"
	"io"
	"strings"
)

func Import(inFile io.Reader) (af dung.AF, err error) {
	reader := bufio.NewReader(inFile)
	args := make([]dung.Arg, 0, 50)
	atks := make(map[dung.Arg][]dung.Arg, 50)
	nodeList := true // false if reading the list of edges has begun
	declared := map[dung.Arg]bool{}
	edges := map[[2]dung.Arg]bool{}
	lineNumber := 0
	eof := false
	for !eof {
		line, readErr := reader.ReadString('\n')
		err = readErr
		lineNumber++
		if err == io.EOF {
			err = nil // io.EOF isn't really an error
			eof = true
		} else if err != nil {
			return af, err // finish immediately for real errors
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if nodeList {
			if fields[0] == "#" {
				if len(fields) != 1 {
					return af, fmt.Errorf("TGF line %d: the separator must be # alone", lineNumber)
				}
				nodeList = false // start of edges list
				continue
			}
			arg := dung.Arg(fields[0])
			if declared[arg] {
				return af, fmt.Errorf("TGF line %d: duplicate argument %q", lineNumber, arg)
			}
			declared[arg] = true
			args = append(args, arg) // optional node text is not part of Dung semantics
		} else {
			// In this Dung profile every edge is an attack. Do not silently
			// turn a labelled support edge (e.g. a Kialo export) into an attack.
			if len(fields) != 2 {
				return af, fmt.Errorf("TGF line %d: expected two argument IDs; edge labels and malformed edges are unsupported", lineNumber)
			}
			from, to := dung.Arg(fields[0]), dung.Arg(fields[1])
			if !declared[from] || !declared[to] {
				return af, fmt.Errorf("TGF line %d: attack refers to an undeclared argument", lineNumber)
			}
			pair := [2]dung.Arg{from, to}
			if !edges[pair] {
				edges[pair] = true
				atks[to] = append(atks[to], from)
			}
		}
	}
	if nodeList {
		return af, fmt.Errorf("TGF: missing # separator")
	}
	return dung.NewAF(args, atks), err
}
