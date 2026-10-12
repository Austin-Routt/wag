package nascq

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// NASCQ is the in-memory representation of the Network of Argumentation
// Schemes and Critical Questions. After loading, it provides O(1) indexed
// access to ask/backup relationships.
//
// Thread safety: NASCQ is immutable after construction. All query methods
// are safe for concurrent use without locking.
type NASCQ struct {
	Schemes   map[string]*ASNode // NASCQ scheme ID → ASNode
	Questions map[string]*CQNode // NASCQ CQ ID → CQNode

	// askIndex maps scheme_id → []cq_id.
	// "Which CQs can be asked of arguments using this scheme?"
	askIndex map[string][]string

	// backupIndex maps cq_id → []scheme_id (thesis-dependent schemes only).
	// "Which thesis-dependent schemes can back up an answer to this CQ?"
	backupIndex map[string][]string

	// thesisIndep holds all thesis-independent scheme IDs.
	// These implicitly back up ALL CQs — no explicit edges needed.
	thesisIndep []string


	reverseSchemeIndex map[string]string // CarneadesID → NASCQ scheme ID


	// Raw edges stored for validation and export.
	AskEdges    []Edge
	BackupEdges []Edge
}

// ---------------------------------------------------------------------------
// JSON file formats
// ---------------------------------------------------------------------------

// schemesFileMap is the plan's format: map keyed by ID.
// Example: { "AS:practical_reasoning": { "carneades_id": "...", ... } }
type schemesFileMap map[string]*ASNode

// schemesFileArray is the Python prototype's format: array with IDs in each object.
// Example: { "argumentation_schemes": [ { "id": "AS:practical_reasoning", ... } ] }
type schemesFileArray struct {
	Schemes []ASNode `json:"argumentation_schemes"`
}

// cqFileMap is the plan's format.
type cqFileMap map[string]*CQNode

// cqFileArray is the Python prototype's format.
type cqFileArray struct {
	Questions []CQNode `json:"critical_questions"`
}

// edgeFilePlan is the plan's format: separate ask/backup arrays.
type edgeFilePlan struct {
	AskEdges    []Edge `json:"ask_edges"`
	BackupEdges []Edge `json:"backup_edges"`
}

// edgeFileProto is the Python prototype's format: single array with type field.
type edgeFileProto struct {
	Edges []Edge `json:"edges"`
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// LoadFromFiles loads NASCQ from three JSON files on disk.
// It auto-detects whether each file uses the plan's format or the Python
// prototype's format.
func LoadFromFiles(schemesPath, questionsPath, edgesPath string) (*NASCQ, error) {
	schemesData, err := os.ReadFile(schemesPath)
	if err != nil {
		return nil, fmt.Errorf("reading schemes file: %w", err)
	}
	questionsData, err := os.ReadFile(questionsPath)
	if err != nil {
		return nil, fmt.Errorf("reading questions file: %w", err)
	}
	edgesData, err := os.ReadFile(edgesPath)
	if err != nil {
		return nil, fmt.Errorf("reading edges file: %w", err)
	}
	return LoadFromBytes(schemesData, questionsData, edgesData)
}

// LoadFromBytes loads NASCQ from raw JSON bytes. Useful for testing with
// embedded fixtures and for loading from non-filesystem sources.
func LoadFromBytes(schemesJSON, questionsJSON, edgesJSON []byte) (*NASCQ, error) {
	n := &NASCQ{
		Schemes:     make(map[string]*ASNode),
		Questions:   make(map[string]*CQNode),
		askIndex:    make(map[string][]string),
		backupIndex: make(map[string][]string),
	}

	if err := n.loadSchemes(schemesJSON); err != nil {
		return nil, fmt.Errorf("loading schemes: %w", err)
	}
	if err := n.loadQuestions(questionsJSON); err != nil {
		return nil, fmt.Errorf("loading questions: %w", err)
	}
	if err := n.loadEdges(edgesJSON); err != nil {
		return nil, fmt.Errorf("loading edges: %w", err)
	}
	n.buildIndexes()
	return n, nil
}

// ---------------------------------------------------------------------------
// Loaders (auto-detect format)
// ---------------------------------------------------------------------------

func (n *NASCQ) loadSchemes(data []byte) error {
	// Try plan's map format first.
	var mapFormat schemesFileMap
	if err := json.Unmarshal(data, &mapFormat); err == nil && len(mapFormat) > 0 {
		// Check if any entry has a populated Name (distinguishes from empty parse).
		for id, node := range mapFormat {
			if node.Name != "" {
				node.ID = id
				n.Schemes[id] = node
			}
		}
		if len(n.Schemes) > 0 {
			return nil
		}
	}

	// Try Python prototype's array format.
	var arrFormat schemesFileArray
	if err := json.Unmarshal(data, &arrFormat); err == nil && len(arrFormat.Schemes) > 0 {
		for i := range arrFormat.Schemes {
			s := &arrFormat.Schemes[i]
			if s.ID == "" {
				return fmt.Errorf("scheme at index %d has no id", i)
			}
			n.Schemes[s.ID] = s
		}
		return nil
	}

	// Try bare array (no wrapper).
	var bareArr []ASNode
	if err := json.Unmarshal(data, &bareArr); err == nil && len(bareArr) > 0 {
		for i := range bareArr {
			s := &bareArr[i]
			if s.ID == "" {
				return fmt.Errorf("scheme at index %d has no id", i)
			}
			n.Schemes[s.ID] = s
		}
		return nil
	}

	return fmt.Errorf("could not parse schemes JSON (tried map, wrapped array, and bare array formats)")
}

func (n *NASCQ) loadQuestions(data []byte) error {
	// Try plan's map format first.
	var mapFormat cqFileMap
	if err := json.Unmarshal(data, &mapFormat); err == nil && len(mapFormat) > 0 {
		for id, node := range mapFormat {
			if node.Name != "" {
				node.ID = id
				n.Questions[id] = node
			}
		}
		if len(n.Questions) > 0 {
			return nil
		}
	}

	// Try Python prototype's array format.
	var arrFormat cqFileArray
	if err := json.Unmarshal(data, &arrFormat); err == nil && len(arrFormat.Questions) > 0 {
		for i := range arrFormat.Questions {
			q := &arrFormat.Questions[i]
			if q.ID == "" {
				return fmt.Errorf("question at index %d has no id", i)
			}
			n.Questions[q.ID] = q
		}
		return nil
	}

	// Try bare array.
	var bareArr []CQNode
	if err := json.Unmarshal(data, &bareArr); err == nil && len(bareArr) > 0 {
		for i := range bareArr {
			q := &bareArr[i]
			if q.ID == "" {
				return fmt.Errorf("question at index %d has no id", i)
			}
			n.Questions[q.ID] = q
		}
		return nil
	}

	return fmt.Errorf("could not parse questions JSON (tried map, wrapped array, and bare array formats)")
}

func (n *NASCQ) loadEdges(data []byte) error {
	// Try plan's format first: { "ask_edges": [...], "backup_edges": [...] }
	var planFormat edgeFilePlan
	if err := json.Unmarshal(data, &planFormat); err == nil {
		if len(planFormat.AskEdges) > 0 || len(planFormat.BackupEdges) > 0 {
			// Plan format uses CQ/Scheme fields.
			for i := range planFormat.AskEdges {
				e := &planFormat.AskEdges[i]
				e.Type = "ask"
				// Normalize: plan format has CQ and Scheme fields.
				if e.From == "" {
					e.From = e.CQ
				}
				if e.To == "" {
					e.To = e.Scheme
				}
			}
			for i := range planFormat.BackupEdges {
				e := &planFormat.BackupEdges[i]
				e.Type = "backup"
				if e.From == "" {
					e.From = e.Scheme
				}
				if e.To == "" {
					e.To = e.CQ
				}
			}
			n.AskEdges = planFormat.AskEdges
			n.BackupEdges = planFormat.BackupEdges
			return nil
		}
	}

	// Try Python prototype's format: { "edges": [...] } with type field.
	var protoFormat edgeFileProto
	if err := json.Unmarshal(data, &protoFormat); err == nil && len(protoFormat.Edges) > 0 {
		for _, e := range protoFormat.Edges {
			switch e.Type {
			case "ask":
				n.AskEdges = append(n.AskEdges, e)
			case "backup":
				n.BackupEdges = append(n.BackupEdges, e)
			default:
				return fmt.Errorf("unknown edge type %q in edge %s", e.Type, e.ID)
			}
		}
		return nil
	}

	// Try bare array of edges.
	var bareArr []Edge
	if err := json.Unmarshal(data, &bareArr); err == nil && len(bareArr) > 0 {
		for _, e := range bareArr {
			switch e.Type {
			case "ask":
				n.AskEdges = append(n.AskEdges, e)
			case "backup":
				n.BackupEdges = append(n.BackupEdges, e)
			}
		}
		return nil
	}

	return fmt.Errorf("could not parse edges JSON (tried plan, prototype, and bare array formats)")
}

// ---------------------------------------------------------------------------
// Index building
// ---------------------------------------------------------------------------

// buildIndexes populates askIndex, backupIndex, and thesisIndep from the
// loaded edges and scheme metadata. Called once after all data is loaded.
func (n *NASCQ) buildIndexes() {
	// Build ask index: scheme_id → []cq_id
	// Ask edges go CQ → AS, meaning "this CQ can be asked of this scheme."
	for _, e := range n.AskEdges {
		cqID, schemeID := n.resolveAskEdge(e)
		if cqID != "" && schemeID != "" {
			n.askIndex[schemeID] = append(n.askIndex[schemeID], cqID)
		}
	}

	// Build backup index: cq_id → []scheme_id (thesis-dependent only)
	// Backup edges go AS → CQ, meaning "this scheme can back up this CQ."
	for _, e := range n.BackupEdges {
		schemeID, cqID := n.resolveBackupEdge(e)
		if schemeID != "" && cqID != "" {
			n.backupIndex[cqID] = append(n.backupIndex[cqID], schemeID)
		}
	}

	// Collect thesis-independent schemes.
	n.thesisIndep = nil
	for _, id := range sortedSchemeIDs(n.Schemes) { // sorted: map order is random
		if n.Schemes[id].ThesisIndependent {
			n.thesisIndep = append(n.thesisIndep, id)
		}
	}
	n.buildReverseSchemeIndex()
}

// resolveAskEdge extracts (cqID, schemeID) from an ask edge,
// handling both plan format (CQ/Scheme fields) and prototype format (From/To).
func (n *NASCQ) resolveAskEdge(e Edge) (cqID, schemeID string) {
	// Ask edges: CQ → AS. In plan format: CQ=cqID, Scheme=schemeID.
	// In prototype format: From=CQ, To=AS.
	if e.CQ != "" {
		return e.CQ, e.Scheme
	}
	// Prototype format: From should be a CQ, To should be an AS.
	from, to := e.From, e.To
	if strings.HasPrefix(from, "CQ:") && strings.HasPrefix(to, "AS:") {
		return from, to
	}
	// Heuristic: check which is in Questions vs Schemes.
	if _, ok := n.Questions[from]; ok {
		return from, to
	}
	return from, to
}

// resolveBackupEdge extracts (schemeID, cqID) from a backup edge.
func (n *NASCQ) resolveBackupEdge(e Edge) (schemeID, cqID string) {
	// Backup edges: AS → CQ. In plan format: Scheme=schemeID, CQ=cqID.
	if e.Scheme != "" {
		return e.Scheme, e.CQ
	}
	// Prototype format: From=AS, To=CQ.
	from, to := e.From, e.To
	if strings.HasPrefix(from, "AS:") && strings.HasPrefix(to, "CQ:") {
		return from, to
	}
	if _, ok := n.Schemes[from]; ok {
		return from, to
	}
	return from, to
}

// ---------------------------------------------------------------------------
// Query methods (all read-only, safe for concurrent use)
// ---------------------------------------------------------------------------

// GetAskCQs returns the CQ-nodes that can challenge arguments using the
// given scheme. These are the CQs connected via ask edges.
func (n *NASCQ) GetAskCQs(schemeID string) []*CQNode {
	cqIDs := n.askIndex[schemeID]
	result := make([]*CQNode, 0, len(cqIDs))
	for _, id := range cqIDs {
		if q, ok := n.Questions[id]; ok {
			result = append(result, q)
		}
	}
	return result
}

// GetBackupSchemes returns all schemes that can back up an answer to the
// given CQ. This includes:
//   - All thesis-dependent schemes with explicit backup edges to this CQ
//   - ALL thesis-independent schemes (they back every CQ by definition)
func (n *NASCQ) GetBackupSchemes(cqID string) []*ASNode {
	// Explicit backup edges (thesis-dependent).
	depIDs := n.backupIndex[cqID]
	seen := make(map[string]bool, len(depIDs)+len(n.thesisIndep))
	result := make([]*ASNode, 0, len(depIDs)+len(n.thesisIndep))

	for _, sid := range depIDs {
		if s, ok := n.Schemes[sid]; ok {
			result = append(result, s)
			seen[sid] = true
		}
	}
	// All thesis-independent schemes.
	for _, sid := range n.thesisIndep {
		if !seen[sid] {
			if s, ok := n.Schemes[sid]; ok {
				result = append(result, s)
			}
		}
	}
	return result
}

// GetScheme returns an AS-node by its NASCQ ID, or nil if not found.
func (n *NASCQ) GetScheme(id string) *ASNode {
	return n.Schemes[id]
}

// GetCQ returns a CQ-node by its NASCQ ID, or nil if not found.
func (n *NASCQ) GetCQ(id string) *CQNode {
	return n.Questions[id]
}

// ---------------------------------------------------------------------------
// Bipartite structure validation
// ---------------------------------------------------------------------------

// ValidateBipartite checks that the NASCQ graph is a valid bipartite graph:
//   - Ask edges go CQ → AS only (not AS → AS or CQ → CQ)
//   - Backup edges go AS → CQ only
//   - All edge endpoints reference existing nodes
//
// Returns a list of error strings. Empty list means valid.
func (n *NASCQ) ValidateBipartite() []string {
	var errors []string

	for i, e := range n.AskEdges {
		cqID, schemeID := n.resolveAskEdge(e)
		if _, ok := n.Questions[cqID]; !ok {
			errors = append(errors, fmt.Sprintf("ask edge %d: CQ %q not found in questions", i, cqID))
		}
		if _, ok := n.Schemes[schemeID]; !ok {
			errors = append(errors, fmt.Sprintf("ask edge %d: scheme %q not found in schemes", i, schemeID))
		}
	}

	for i, e := range n.BackupEdges {
		schemeID, cqID := n.resolveBackupEdge(e)
		if _, ok := n.Schemes[schemeID]; !ok {
			errors = append(errors, fmt.Sprintf("backup edge %d: scheme %q not found in schemes", i, schemeID))
		}
		if _, ok := n.Questions[cqID]; !ok {
			errors = append(errors, fmt.Sprintf("backup edge %d: CQ %q not found in questions", i, cqID))
		}
	}

	return errors
}

// Statistics returns a summary of the NASCQ graph contents.
type Statistics struct {
	SchemeCount          int `json:"scheme_count"`
	QuestionCount        int `json:"question_count"`
	AskEdgeCount         int `json:"ask_edge_count"`
	BackupEdgeCount      int `json:"backup_edge_count"`
	ThesisIndependentCount int `json:"thesis_independent_count"`
}

func (n *NASCQ) Statistics() Statistics {
	return Statistics{
		SchemeCount:          len(n.Schemes),
		QuestionCount:        len(n.Questions),
		AskEdgeCount:         len(n.AskEdges),
		BackupEdgeCount:      len(n.BackupEdges),
		ThesisIndependentCount: len(n.thesisIndep),
	}
}
