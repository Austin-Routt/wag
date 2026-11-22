// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Import CAES argument graphs from a JSON serialization of the
// Argument Interchange Format (AIF). For more information about AIF, see:
// http://www.argumentinterchange.org/home
// http://www.arg-tech.org/index.php/projects/

// See also:
// Chesnevar,C., McGinnis,J., Modgil,S. Rahwan,I., Reed,C., Simari,G.,
// South,M., Vreeswijk,G. & Willmott,S. (2006)
// "Towards an Argument Interchange Format",
// Knowledge Engineering Review, 21 (4), pp. 293-316

package aif

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"sort"

	"github.com/carneades/carneades-4/src/engine/caes"
)

type AIF struct {
	Edges []Edge `json:"edges"`
	Nodes []Node `json:"nodes"`
}

func newAIF() AIF {
	return AIF{
		Edges: []Edge{},
		Nodes: []Node{},
	}
}

type Edge struct {
	Id       string `json:"edgeID"`
	From     string `json:"fromID"`
	FormEdge string `json:"formEdgeID"`
	To       string `json:"toID"`
}

type Node struct {
	Id        string `json:"nodeID"`
	Text      string `json:"text"`
	Scheme    string `json:"scheme"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
}

// Caes converts an AIF value satisfying ValidateProfile. Callers reading an
// untrusted or unvalidated value must use Import or ValidateProfile first.
// AIF scheme names are provenance, not bindings to executable Walton schemes.
func (ag AIF) Caes() *caes.ArgGraph {
	stmts := make(map[string]*caes.Statement)
	assms := make(map[string]bool) // keys are statement ids
	args := make(map[string]*caes.Argument)
	issues := make(map[string]*caes.Issue)
	nodeType := func(id string) string {
		if stmts[id] != nil {
			return "Statement"
		} else if args[id] != nil {
			return "Argument"
		} else if issues[id] != nil {
			return "Issue"
		} else {
			return ""
		}
	}

	for _, node := range ag.Nodes {
		switch node.Type {
		case "I":
			s := caes.NewStatement()
			s.Id = node.Id
			s.Text = node.Text
			assms[s.Id] = true // may be overridden below
			stmts[s.Id] = s
		case "CA":
			i := caes.NewIssue()
			i.Id = node.Id
			issues[i.Id] = i
		case "RA":
			arg := caes.NewArgument()
			arg.Id = node.Id
			arg.Scheme = caes.BasicSchemes["linked"]
			arg.Metadata["aif_scheme"] = node.Scheme
			arg.Metadata["aif_text"] = node.Text
			arg.Metadata["aif_timestamp"] = node.Timestamp
			args[arg.Id] = arg
		default:
			continue
		}
	}

	for _, edge := range ag.Edges {
		switch nodeType(edge.From) {
		case "Statement":
			s := stmts[edge.From]
			switch nodeType(edge.To) {
			case "Argument":
				arg := args[edge.To]
				p := caes.Premise{Stmt: s}
				arg.Premises = append(arg.Premises, p)
			case "Issue":
				i := issues[edge.To]
				i.Positions = append(i.Positions, s)
				s.Issue = i
			default:
				continue
			}
		case "Issue":
			i := issues[edge.From]
			switch nodeType(edge.To) {
			case "Statement":
				s := stmts[edge.To]
				i.Positions = append(i.Positions, s)
				s.Issue = i
			default:
				continue
			}
		case "Argument":
			arg := args[edge.From]
			switch nodeType(edge.To) {
			case "Statement":
				s := stmts[edge.To]
				arg.Conclusion = s
				s.Args = append(s.Args, arg)
			default:
				continue
			}
		default:
			continue
		}
	}

	cag := caes.NewArgGraph()
	cag.Metadata["import_profile"] = "AIF I/RA/binary-CA: linked arguments and symmetric incompatibility"
	cag.Metadata["import_assumption_policy"] = "I nodes without an incoming RA and without a CA issue are stipulated"

	for i, _ := range stmts {
		s := stmts[i]
		if len(s.Args) > 0 || s.Issue != nil {
			// do not assume statements supported by arguments or at issue
			assms[s.Id] = false
		}
		cag.Statements[s.Id] = s
	}

	for _, issue := range issues {
		cag.Issues[issue.Id] = issue
	}

	for _, arg := range args {
		cag.Arguments[arg.Id] = arg
	}

	for k, assumed := range assms {
		if assumed {
			cag.Assumptions = append(cag.Assumptions, k)
		}
	}
	sort.Strings(cag.Assumptions)
	return cag
}

func Import(inFile io.Reader) (*caes.ArgGraph, error) {
	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return nil, caes.NewRunError(caes.ErrInvalidInput, "AIF: %v", err)
	}
	var aif AIF
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&aif); err != nil {
		return nil, caes.NewRunError(caes.ErrInvalidInput, "AIF: %v", err)
	}
	var trailing interface{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, caes.NewRunError(caes.ErrInvalidInput, "AIF: expected one JSON object")
	}
	if err := aif.ValidateProfile(); err != nil {
		return nil, err
	}
	return aif.Caes(), nil
}
