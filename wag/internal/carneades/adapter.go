// Package carneades provides a concurrency-safe adapter around the
// Carneades-4 CAES argument evaluation engine, tailored for the
// Waltonian Argumentation Game (WAG).
//
// Design principles verified against the carneades-4 source code:
//
//   - All graph mutations are serialized by a sync.Mutex.
//   - Support moves use caes.InstantiateScheme (Carneades built-in)
//     which requires Theory.schemeIndex to be populated via InitSchemeIndex().
//   - Attack moves (rebuttal, undercutter, undermining) are implemented
//     manually because they require specific graph structures (negated
//     statements, issues, undercutter pointers) that InstantiateScheme
//     does not produce directly.
//   - User-supplied English text is sanitized into Prolog quoted atoms
//     before being passed to terms.ReadString.
//   - Statement IDs are normalized via terms.Normalize to match the
//     canonical form used by GroundedLabelling's init method (Labelling.init
//     in caes.go re-keys the statement map by terms.Normalize).
//   - Argument IDs follow the "aN" convention (a1, a2, ...) matching
//     the genArgId logic in InstantiateScheme (caes.go).
//   - Argument weight is set via ConstantWeighingFunction on a per-argument
//     Scheme copy. NEVER set arg.Weight directly — it is computed on demand
//     by Argument.GetWeight (caes.go).
//
// Locking contract: All public methods acquire a.mu. Internal helpers
// (findOrCreateStatement, createArgument, nextArgID, etc.) do NOT
// acquire the mutex; callers must already hold it.
package carneades

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	caes "github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
)

// Adapter is a concurrency-safe wrapper around a Carneades ArgGraph.
// It provides high-level operations for the Waltonian Argumentation Game:
// creating arguments (support/rebuttal/undercutter/undermining), running
// the grounded labelling evaluation, and checking graph acyclicity.
type Adapter struct {
	mu        sync.Mutex
	Graph     *caes.ArgGraph          // The underlying Carneades argument graph.
	schemeMap map[string]*caes.Scheme // Local cache: scheme ID → *Scheme for fast lookup.
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// NewAdapter creates an Adapter with a fresh, properly initialized argument
// graph. Uses caes.NewArgGraph() which initializes all internal maps including
// the unexported assums map (required by AddAssumption) and a default Theory
// with an empty but initialized schemeIndex (required by InstantiateScheme).
//
// The returned graph has no schemes loaded. To use AddSupport, you must first
// add schemes to Graph.Theory.ArgSchemes and call Graph.Theory.InitSchemeIndex().
// Alternatively, use NewAdapterFromYAML (Block 2) to load a YAML file
// that includes a Theory section.
func NewAdapter() *Adapter {
	ag := caes.NewArgGraph()
	a := &Adapter{Graph: ag}
	a.buildSchemeMap()
	return a
}

// NewAdapterWithGraph wraps an existing ArgGraph in an Adapter.
// The graph should be created via caes.NewArgGraph() or loaded via YAML
// to ensure internal maps (assums, schemeIndex) are properly initialized.
// If graph.Theory has ArgSchemes, call graph.Theory.InitSchemeIndex() before
// wrapping if the index has not been built yet.
func NewAdapterWithGraph(ag *caes.ArgGraph) *Adapter {
	if ag == nil {
		ag = caes.NewArgGraph()
	}
	a := &Adapter{Graph: ag}
	a.buildSchemeMap()
	return a
}

// buildSchemeMap populates the adapter's local scheme cache from the Theory's
// ArgSchemes slice. This is a convenience for makeWeightedScheme and AddSupport
// which need fast scheme lookups by ID.
func (a *Adapter) buildSchemeMap() {
	a.schemeMap = make(map[string]*caes.Scheme)
	if a.Graph.Theory != nil {
		for _, s := range a.Graph.Theory.ArgSchemes {
			if s != nil {
				a.schemeMap[s.Id] = s
			}
		}
		// Ensure undercut/1 is in the Language map so that Language.Apply
		// doesn't print "Functor not defined" warnings for undercutter terms.
		if a.Graph.Theory.Language == nil {
			a.Graph.Theory.Language = make(map[string]string)
		}
		a.Graph.Theory.Language["undercut/1"] = "Argument %s is undercut."
	}
}

// ---------------------------------------------------------------------------
// String helpers (pure functions, no locking needed)
// ---------------------------------------------------------------------------

// quoteForProlog wraps a user-supplied string in Prolog single-quote syntax
// so that terms.ReadString can parse it as a quoted atom.
// Example: `Dr. Marie Curie` → `'Dr. Marie Curie'`
//
// Design decision D7 from the WAG plan: user input must be wrapped in
// Prolog quoted-atom syntax to prevent the term parser from failing on
// spaces, capitals, or special characters.
func quoteForProlog(s string) string {
	escaped := strings.ReplaceAll(s, "'", "\\'")
	return "'" + escaped + "'"
}

// sanitizeBinding is a semantic alias for quoteForProlog, used when
// preparing variable bindings for InstantiateScheme. Replaced by SanitizeBinding in santize.go
//func sanitizeBinding(s string) string {
//	return quoteForProlog(s)
//}

// generateStatementIDFromText creates a deterministic, collision-resistant
// statement ID from arbitrary text. Uses SHA-1 hash prefixed with "stmt_".
// This is the fallback when the text cannot be parsed as a Prolog term.
func generateStatementIDFromText(text string) string {
	h := sha1.Sum([]byte(text))
	return "stmt_" + hex.EncodeToString(h[:])
}

// ---------------------------------------------------------------------------
// Graph helpers (must hold a.mu before calling)
// ---------------------------------------------------------------------------

// findOrCreateStatement finds a statement by its canonical ID or creates it.
//
// Process:
//  1. Try terms.ReadString(idOrText). If it succeeds, use the normalized
//     term string as the ID and the Language-formatted text as the display text.
//  2. If parsing fails, normalize the raw text with terms.Normalize (which
//     itself tries to parse and reprint). If that still doesn't help, generate
//     a SHA-1 based ID.
//
// The normalized ID is used as BOTH the map key and Statement.Id to ensure
// consistency with GroundedLabelling's init method which normalizes map keys
// (Labelling.init in caes.go).
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) findOrCreateStatement(idOrText string) *caes.Statement {
	// Negated statements (¬ prefix) are not parseable Prolog terms.
	// Carneades uses "¬" + id as raw map keys (see caf.go complement()).
	// Skip the term parser to avoid "EOF missing term" stderr warnings.
	if strings.HasPrefix(idOrText, "¬") {
		if stmt, exists := a.Graph.Statements[idOrText]; exists {
			return stmt
		}
		stmt := &caes.Statement{Id: idOrText, Text: idOrText, Metadata: caes.NewMetadata()}
		a.Graph.Statements[idOrText] = stmt
		return stmt
	}
	// Empty strings are not valid statement IDs. Reject early to avoid
	// "EOF missing term" warnings from terms.ReadString("").
	if strings.TrimSpace(idOrText) == "" {
		// Return a placeholder statement rather than crashing.
		const emptyID = "stmt_empty"
		if stmt, exists := a.Graph.Statements[emptyID]; exists {
			return stmt
		}
		stmt := &caes.Statement{Id: emptyID, Text: "(empty)", Metadata: caes.NewMetadata()}
		a.Graph.Statements[emptyID] = stmt
		return stmt
	}

	// Attempt to parse as a Prolog term for canonical representation.
	if t, ok := terms.ReadString(idOrText); ok {
		// Use the canonical term string as the ID.
		id := t.String()
		if stmt, exists := a.Graph.Statements[id]; exists {
			return stmt
		}
		// Format display text via the Theory's Language map if available.
		text := t.String()
		if a.Graph.Theory != nil && a.Graph.Theory.Language != nil {
			if langText := a.Graph.Theory.Language.Apply(t); strings.TrimSpace(langText) != "" {
				text = langText
			}
		}
		stmt := &caes.Statement{Id: id, Text: text, Metadata: caes.NewMetadata()}
		a.Graph.Statements[id] = stmt
		return stmt
	}

	// Not a parseable term — try normalizing the raw text.
	nid := terms.Normalize(idOrText)
	if stmt, exists := a.Graph.Statements[nid]; exists {
		return stmt
	}
	// If normalization returned the same string (no parse), use SHA-1 ID.
	id := nid
	if nid == idOrText {
		id = generateStatementIDFromText(idOrText)
		if stmt, exists := a.Graph.Statements[id]; exists {
			return stmt
		}
	}
	stmt := &caes.Statement{Id: id, Text: idOrText, Metadata: caes.NewMetadata()}
	a.Graph.Statements[id] = stmt
	return stmt
}

// nextArgID returns an unused argument ID of the form "aN" (e.g., "a1", "a2").
// It scans all existing argument IDs and picks the next available numeric suffix.
// This matches the genArgId strategy inside InstantiateScheme (caes.go).
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) nextArgID() string {
	max := 0
	re := regexp.MustCompile(`^a(\d+)$`)
	for id := range a.Graph.Arguments {
		if m := re.FindStringSubmatch(id); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > max {
				max = n
			}
		}
	}
	return fmt.Sprintf("a%d", max+1)
}

// makeWeightedScheme returns a copy of the scheme identified by schemeID,
// with its Weight function replaced by ConstantWeighingFunction(weight).
//
// ConstantWeighingFunction(w) (caes weighing.go) returns:
//   - 0.0 if ANY premise is not In
//   - w   if ALL premises are In
//
// If the scheme is not found in the local cache, a minimal scheme with only
// the constant weight is returned as a fallback. This is safe for attack
// methods where the scheme is primarily informational.
//
// Design decision D6: game-assigned weights use ConstantWeighingFunction
// so that CAES computes weight correctly through its normal GetWeight() path.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) makeWeightedScheme(schemeID string, weight float64) *caes.Scheme {
	base := a.schemeMap[schemeID]
	if base == nil {
		// Fallback: minimal scheme with only the constant weight.
		return &caes.Scheme{
			Id:     schemeID,
			Weight: caes.ConstantWeighingFunction(weight),
		}
	}
	// Shallow copy is sufficient — we only replace the Weight function.
	schemeCopy := *base
	schemeCopy.Weight = caes.ConstantWeighingFunction(weight)
	return &schemeCopy
}

// createArgument builds a new Argument, inserts it into the graph, and links
// it bidirectionally to its conclusion statement (stmt.Args = append(..., arg)).
//
// This is the shared helper for all attack methods (AddRebuttal, AddUndercutter,
// AddUndermining). It handles:
//   - Unique argument ID generation
//   - Scheme copy with constant weight
//   - Premise list construction
//   - Bidirectional linking (both arg.Conclusion = stmt AND stmt.Args includes arg)
//   - Owner metadata attribution
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) createArgument(
	ownerID string,
	schemeID string,
	premises []*caes.Statement,
	conclusion *caes.Statement,
	weight float64,
) (string, error) {
	argID := a.nextArgID()
	scheme := a.makeWeightedScheme(schemeID, weight)

	premiseList := make([]caes.Premise, len(premises))
	for i, p := range premises {
		premiseList[i] = caes.Premise{Stmt: p}
	}

	arg := &caes.Argument{
		Id:       argID,
		Metadata: caes.Metadata{"owner": ownerID, RoleKey: RoleMain, "created_at": time.Now().UTC().Format(time.RFC3339)},
		Scheme:   scheme,
		Premises: premiseList,
		// Conclusion is set below.
	}
	arg.Conclusion = conclusion

	// Insert into graph map.
	a.Graph.Arguments[argID] = arg

	// Bidirectional link: the conclusion must know about this argument.
	// This is essential — without it, GroundedLabelling won't see the argument
	// when evaluating the statement's Supported/Unsupported status.
	conclusion.Args = append(conclusion.Args, arg)

	return argID, nil
}

// detachArgumentFromStatement removes arg from stmt.Args (if present).
// Used by RemoveArgument to clean up the bidirectional link when rolling
// back a move. After this call, the statement no longer lists the argument
// as one of its supporting/attacking arguments.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) detachArgumentFromStatement(arg *caes.Argument, stmt *caes.Statement) {
	if stmt == nil || arg == nil {
		return
	}
	newArgs := make([]*caes.Argument, 0, len(stmt.Args))
	for _, aa := range stmt.Args {
		if aa.Id != arg.Id {
			newArgs = append(newArgs, aa)
		}
	}
	stmt.Args = newArgs
}

// ---------------------------------------------------------------------------
// Public API: Graph building helpers
// ---------------------------------------------------------------------------

// AddStatement adds a statement to the graph with the given ID and text.
// If a statement with that ID already exists, it is returned unchanged.
// The ID is normalized via terms.Normalize for consistency with Carneades internals.
func (a *Adapter) AddStatement(id, text string) *caes.Statement {
	a.mu.Lock()
	defer a.mu.Unlock()
	nid := terms.Normalize(id)
	if s, ok := a.Graph.Statements[nid]; ok {
		return s
	}
	s := &caes.Statement{Id: nid, Text: text, Metadata: caes.NewMetadata()}
	a.Graph.Statements[nid] = s
	return s
}

// AddAssumption marks a statement ID as assumed true. Assumed statements are
// labeled In at the start of GroundedLabelling (Labelling.init in caes.go).
//
// Uses ArgGraph.AddAssumption which writes to BOTH the Assumptions slice
// AND the internal assums map (ArgGraph.AddAssumption in caes.go). The ID is normalized
// to match how GroundedLabelling's init normalizes assumption IDs.
func (a *Adapter) AddAssumption(stmtID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Graph.AddAssumption(terms.Normalize(stmtID))
}

// AddArgumentDirect creates an argument directly without using InstantiateScheme.
// This is useful for programmatic graph construction and testing.
// Returns the ID of the newly created argument.
//
// All statements referenced by premiseIDs and conclusionID must already exist
// in the graph. IDs are normalized before lookup.
func (a *Adapter) AddArgumentDirect(
	ownerID string,
	schemeID string,
	premiseIDs []string,
	conclusionID string,
	weight float64,
) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	premises := make([]*caes.Statement, 0, len(premiseIDs))
	for _, pid := range premiseIDs {
		s := a.Graph.Statements[terms.Normalize(pid)]
		if s == nil {
			return "", fmt.Errorf("premise statement %q not found", pid)
		}
		premises = append(premises, s)
	}

	conclusion := a.Graph.Statements[terms.Normalize(conclusionID)]
	if conclusion == nil {
		return "", fmt.Errorf("conclusion statement %q not found", conclusionID)
	}

	return a.createArgument(ownerID, schemeID, premises, conclusion, weight)
}

// ---------------------------------------------------------------------------
// Public API: Evaluation
// ---------------------------------------------------------------------------

// Evaluate recomputes the grounded labelling and applies it to the graph.
// After this call, every Statement.Label and Argument.Weight is updated.
//
// GroundedLabelling (caes.go) is a fixpoint iteration:
//  1. Init: assumptions → In; everything else → Undecided.
//  2. Loop: unsupported → Out; supported non-issues → In; ready issues → Resolve.
//  3. Stop when no labels change (fixpoint reached).
//
// ApplyLabelling (caes.go) writes labels and weights
// back to the Statement and Argument structs.
func (a *Adapter) Evaluate() (caes.Labelling, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Graph == nil {
		return nil, errors.New("no argument graph loaded")
	}
	l := a.Graph.GroundedLabelling()
	a.Graph.ApplyLabelling(l)
	return l, nil
}

// PremiseMatching returns the ID of the premise of argument argID that a
// premise pattern in NASCQ notation, such as "possible({A})", designates, or ""
// if none does. The pattern's {Var} placeholders are replaced by the
// argument's own bindings of its scheme variables, and the result must equal
// one of its premises. An argument without such bindings (one not made by
// instantiating a scheme) still matches when exactly one of its premises has
// the pattern's predicate and arity.
func (a *Adapter) PremiseMatching(argID, pattern string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	arg := a.Graph.Arguments[argID]
	if arg == nil || strings.TrimSpace(pattern) == "" {
		return ""
	}
	inst := pattern
	for v, val := range argumentBindings(arg) {
		inst = strings.ReplaceAll(inst, "{"+v+"}", val)
	}
	if !strings.Contains(inst, "{") {
		if t, ok := terms.ReadString(inst); ok {
			for _, p := range arg.Premises {
				if p.Stmt != nil && p.Stmt.Id == t.String() {
					return p.Stmt.Id
				}
			}
		}
	}
	functor, arity, ok := patternPredicate(pattern)
	if !ok {
		return ""
	}
	match := ""
	for _, p := range arg.Premises {
		if p.Stmt == nil {
			continue
		}
		t, ok := terms.ReadString(p.Stmt.Id)
		if !ok {
			continue
		}
		if f, ok := terms.Functor(t); ok && f == functor && terms.Arity(t) == arity {
			if match != "" && match != p.Stmt.Id {
				return "" // ambiguous
			}
			match = p.Stmt.Id
		}
	}
	return match
}

// ArgumentBindings returns the values an argument made by instantiating a
// scheme bound to the scheme's variables (variable name -> term), or nil for
// an argument made otherwise (attacks, presumptions, exceptions).
func (a *Adapter) ArgumentBindings(argID string) map[string]string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return argumentBindings(a.Graph.Arguments[argID])
}

func argumentBindings(arg *caes.Argument) map[string]string {
	if arg == nil || arg.Scheme == nil || len(arg.Parameters) == 0 || len(arg.Scheme.Variables) != len(arg.Parameters) {
		return nil
	}
	b := make(map[string]string, len(arg.Parameters))
	for i, v := range arg.Scheme.Variables {
		b[v] = arg.Parameters[i]
	}
	return b
}

// patternPredicate returns the predicate and arity of a premise pattern such
// as "would_bring_about({A},{S1},{S2})" (would_bring_about, 3).
func patternPredicate(pattern string) (string, int, bool) {
	pattern = strings.TrimSpace(pattern)
	open := strings.Index(pattern, "(")
	if open <= 0 || !strings.HasSuffix(pattern, ")") {
		return pattern, 0, pattern != "" && !strings.ContainsAny(pattern, "{}(), ")
	}
	args, depth := 1, 0
	for _, r := range pattern[open+1 : len(pattern)-1] {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				args++
			}
		}
	}
	return strings.TrimSpace(pattern[:open]), args, true
}

// GroundedLabels computes a fresh grounded labelling without applying it and
// returns it by statement ID. Comparing it with the applied labels (or with
// another call) checks that the labelling is a function of the graph alone:
// Carneades walks its statement map in Go's random order, so a graph that
// breaks one of its invariants can label differently from call to call.
func (a *Adapter) GroundedLabels() map[string]caes.Label {
	a.mu.Lock()
	defer a.mu.Unlock()
	l := a.Graph.GroundedLabelling()
	out := make(map[string]caes.Label, len(a.Graph.Statements))
	for id, s := range a.Graph.Statements {
		out[id] = l[s]
	}
	return out
}

// IsAcyclic performs DFS cycle detection on the argument graph.
// WAG enforces acyclic argument graphs as a game rule. Carneades itself
// handles cycles gracefully (labeling cycle participants as Undecided),
// but this check is a WAG constraint enforced at the adapter layer.
//
// The bipartite graph has two node types:
//
//   - Statement → Argument (via stmt.Args: stmt is conclusion of arg)
//
//   - Argument → Statement (via arg.Premises[i].Stmt: arg depends on premise)
//
//   - Argument → Statement (via arg.Undercutter: arg depends on undercutter)
//
//   - Statement → Argument (via its issue: the arguments for the other positions)
//
// A cycle means an argument transitively depends on its own conclusion.
func (a *Adapter) IsAcyclic() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	const (
		white = 0 // unvisited
		gray  = 1 // in current DFS stack — a revisit means cycle
		black = 2 // fully explored
	)
	color := make(map[string]int) // statement ID → color

	var visitStmt func(string) bool
	var visitArg func(string) bool

	visitStmt = func(sid string) bool {
		if color[sid] == gray {
			return false // cycle: this statement is already on the current path
		}
		if color[sid] == black {
			return true // already fully explored, no cycle through here
		}
		color[sid] = gray
		stmt := a.Graph.Statements[sid]
		if stmt != nil {
			for _, arg := range stmt.Args {
				if !visitArg(arg.Id) {
					return false
				}
			}
			// A position of an issue also depends on the arguments for the
			// other positions: the issue is resolved by comparing their weights
			// (review defect #6: a cycle through a rebuttal was not detected and
			// left statements undecided). The other positions' arguments are
			// followed, not the positions themselves, so P and ¬P do not form a
			// cycle by merely sharing an issue.
			if stmt.Issue != nil {
				for _, pos := range stmt.Issue.Positions {
					if pos == nil || pos == stmt {
						continue
					}
					for _, arg := range pos.Args {
						if !visitArg(arg.Id) {
							return false
						}
					}
				}
			}
		}
		color[sid] = black
		return true
	}

	visitArg = func(aid string) bool {
		arg := a.Graph.Arguments[aid]
		if arg == nil {
			return true
		}
		// Follow premise edges.
		for _, p := range arg.Premises {
			if !visitStmt(p.Stmt.Id) {
				return false
			}
		}
		// Follow undercutter edge.
		if arg.Undercutter != nil {
			if !visitStmt(arg.Undercutter.Id) {
				return false
			}
		}
		return true
	}

	for sid := range a.Graph.Statements {
		if color[sid] == white {
			if !visitStmt(sid) {
				return false
			}
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Public API: Game mutations — Support
// ---------------------------------------------------------------------------

// AddSupport creates a supporting argument using Carneades' InstantiateScheme.
//
// InstantiateScheme (caes.go) performs:
//  1. Looks up the scheme by ID in Theory.schemeIndex (then BasicSchemes fallback).
//  2. Binds each scheme variable to the corresponding parameter value.
//  3. Creates Statement nodes for each premise, assumption, and conclusion.
//  4. Creates Argument nodes linking premises to conclusions.
//  5. Creates an undercutter statement and exception arguments automatically.
//
// variableBindings maps scheme variable names to user-supplied values.
// All values are sanitized via sanitizeBinding (Prolog quoted-atom wrapping).
// The parameter order must match scheme.Variables exactly.
//
// After InstantiateScheme runs, newly created arguments are detected by
// comparing the argument map before/after, then attributed with owner metadata
// and assigned the constant weight via makeWeightedScheme.
//
// Prerequisites: Graph.Theory must have the scheme in ArgSchemes, and
// Graph.Theory.InitSchemeIndex() must have been called. If using NewAdapter(),
// you must add schemes and call InitSchemeIndex manually. If using YAML import
// (Block 2), this is handled automatically.
func (a *Adapter) AddSupport(
	ownerID string,
	carneadesSchemeID string,
	variableBindings map[string]string,
	weight float64,
) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Graph == nil {
		return nil, errors.New("no argument graph loaded")
	}

	// Look up scheme to validate variable bindings and get variable order.
	scheme := a.schemeMap[carneadesSchemeID]
	if scheme == nil {
		// Build a sorted list of available schemes for diagnostics.
		avail := make([]string, 0, len(a.schemeMap))
		for k := range a.schemeMap {
			avail = append(avail, k)
		}
		sort.Strings(avail)
		return nil, fmt.Errorf("scheme %q not found in theory; available: %v",
			carneadesSchemeID, avail)
	}

	// Build the parameters slice in the order declared by scheme.Variables.
	// Each value is sanitized for the Prolog term parser.
	missing := []string{}
	params := make([]string, len(scheme.Variables))
	for i, v := range scheme.Variables {
		val, ok := variableBindings[v]
		if !ok {
			missing = append(missing, v)
		} else {
			params[i] = SanitizeBinding(val)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("missing bindings for variable(s): %v", missing)
	}

	// Snapshot current argument IDs to detect new ones after InstantiateScheme.
	before := make(map[string]bool, len(a.Graph.Arguments))
	for id := range a.Graph.Arguments {
		before[id] = true
	}

	// InstantiateScheme mutates the graph in place. It prints errors to stderr
	// rather than returning them (see InstantiateScheme in caes.go). We detect failure
	// by checking if any new arguments were created.
	a.Graph.InstantiateScheme(carneadesSchemeID, params)

	// Identify and attribute newly created arguments. The returned order is
	// deterministic: the scheme's main arguments first, then the exception
	// arguments, each in natural ID order (callers may rely on created[0]
	// being a main argument; map iteration order is random).
	var mains, exceptions []string
	for id := range a.Graph.Arguments {
		if before[id] {
			continue
		}
		if arg := a.Graph.Arguments[id]; arg.Conclusion != nil && arg.Conclusion.IsUndercutter {
			exceptions = append(exceptions, id)
		} else {
			mains = append(mains, id)
		}
	}
	SortArgumentIDs(mains)
	SortArgumentIDs(exceptions)
	created := append(append([]string{}, mains...), exceptions...)
	now := time.Now().UTC().Format(time.RFC3339)
	for _, id := range mains {
		arg := a.Graph.Arguments[id]
		if arg.Metadata == nil {
			arg.Metadata = caes.NewMetadata()
		}
		arg.Metadata["owner"] = ownerID
		arg.Metadata[RoleKey] = RoleMain
		arg.Metadata["created_at"] = now
		// Override the scheme with a constant-weight copy.
		arg.Scheme = a.makeWeightedScheme(carneadesSchemeID, weight)
		arg.Parameters = params
	}
	// Exception arguments (premise: one of the scheme's exceptions; conclusion:
	// the main arguments' undercutter) are structure the scheme brings with it,
	// not a claim of the mover (review defect #2: they used to receive the
	// support scheme and the mover as owner, and so became targets of critical
	// questions). They keep the scheme InstantiateScheme gave them (none, so
	// Carneades' default linked weighing) and have no owner. Their weight does
	// not affect any label: an undercutter is never a position of an issue.
	for _, id := range exceptions {
		arg := a.Graph.Arguments[id]
		if arg.Metadata == nil {
			arg.Metadata = caes.NewMetadata()
		}
		arg.Metadata[RoleKey] = RoleException
		arg.Metadata["instantiated_by"] = ownerID
		arg.Metadata["created_at"] = now
	}
	if len(created) == 0 {
		return nil, fmt.Errorf("InstantiateScheme for scheme %q created no arguments; "+
			"check that bindings parse correctly and scheme has conclusions",
			carneadesSchemeID)
	}
	return created, nil
}

// ---------------------------------------------------------------------------
// Public API: Game mutations — Attacks
// ---------------------------------------------------------------------------

// AddRebuttal creates a rebuttal attack against a statement's conclusion.
//
// Semantics: a rebuttal argues for the position opposing the target
// conclusion (see opposingPosition): its negation ¬target, made a position of
// a new Issue [target, ¬target] with the given proof standard; or, when the
// target is already at issue, the other position of that issue. Rebutting a
// rebuttal's conclusion ¬p therefore argues for p again.
//
// After evaluation, the Issue's Resolve method (caes.go)
// compares the strongest argument weights for each position. With PE,
// the position with strictly greater max-weight wins.
//
// Parameters:
//   - targetConclusionID: ID of an existing statement to rebut.
//   - premiseIDs: IDs of existing statements that serve as premises.
//   - proofStandard: the proof standard for the Issue (caes.PE, caes.CCE, or caes.BRD).
func (a *Adapter) AddRebuttal(
	ownerID string,
	carneadesSchemeID string,
	targetConclusionID string,
	premiseIDs []string,
	weight float64,
	proofStandard caes.Standard,
) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Graph == nil {
		return nil, errors.New("no argument graph loaded")
	}
	if strings.TrimSpace(targetConclusionID) == "" {
		return nil, errors.New("target conclusion ID is empty")
	}

	target := a.Graph.Statements[terms.Normalize(targetConclusionID)]
	if target == nil {
		return nil, fmt.Errorf("target statement %q not found", targetConclusionID)
	}
	return a.argueAgainst(ownerID, carneadesSchemeID, target, premiseIDs, weight, proofStandard)
}

// argueAgainst creates an argument from premiseIDs for the position opposing
// target. It is the common part of AddRebuttal and AddUndermining.
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) argueAgainst(
	ownerID string,
	carneadesSchemeID string,
	target *caes.Statement,
	premiseIDs []string,
	weight float64,
	proofStandard caes.Standard,
) ([]string, error) {
	// Resolve the premises before changing the graph.
	premises := make([]*caes.Statement, 0, len(premiseIDs))
	for _, pid := range premiseIDs {
		ps := a.Graph.Statements[terms.Normalize(pid)]
		if ps == nil {
			return nil, fmt.Errorf("premise statement %q not found", pid)
		}
		premises = append(premises, ps)
	}
	opposing, err := a.opposingPosition(target, proofStandard)
	if err != nil {
		return nil, err
	}
	argID, err := a.createArgument(ownerID, carneadesSchemeID, premises, opposing, weight)
	if err != nil {
		return nil, err
	}
	return []string{argID}, nil
}

// opposingPosition returns the statement that an attack on target argues for,
// making sure that the two are the positions of one issue:
//
//   - a target already at issue: the other position of its issue (so the
//     opposing position of ¬p, the conclusion of a rebuttal, is p);
//   - otherwise its complement, ¬target (or p for a target ¬p), with a new
//     issue [p, ¬p] under proofStandard.
//
// Carneades resolves an issue from each position's Issue pointer and assumes
// that every statement is a position of at most one issue. Creating ¬¬p and a
// second issue [¬p, ¬¬p] for a rebuttal of a rebuttal broke that assumption:
// ¬p was a position of two issues, and its label depended on Go's map order
// (in 21 of 30 runs ¬p and ¬¬p were both In, in the other 9 p stayed Out
// although the counter-rebuttal outweighed the rebuttal).
//
// IMPORTANT: Caller must hold a.mu.
func (a *Adapter) opposingPosition(target *caes.Statement, proofStandard caes.Standard) (*caes.Statement, error) {
	if is := target.Issue; is != nil {
		if len(is.Positions) != 2 {
			return nil, fmt.Errorf("statement %q is one of %d positions of issue %q; argue for one of the other positions instead",
				target.Id, len(is.Positions), is.Id)
		}
		if is.Positions[0] == target {
			return is.Positions[1], nil
		}
		return is.Positions[0], nil
	}
	pos, neg := target, (*caes.Statement)(nil)
	if strings.HasPrefix(target.Id, "¬") {
		neg = target
		pos = a.findOrCreateStatement(strings.TrimPrefix(target.Id, "¬"))
	} else {
		negID := "¬" + target.Id
		_, existed := a.Graph.Statements[negID]
		neg = a.findOrCreateStatement(negID)
		if !existed {
			neg.Text = "It is not the case that: " + target.Text
		}
	}
	if pos.Issue != nil || neg.Issue != nil {
		return nil, fmt.Errorf("the complement of %q is already a position of another issue", target.Id)
	}
	issueID := "issue:" + pos.Id
	if _, exists := a.Graph.Issues[issueID]; exists {
		return nil, fmt.Errorf("issue %q already exists without %q as a position", issueID, target.Id)
	}
	issue := &caes.Issue{
		Id:        issueID,
		Metadata:  caes.NewMetadata(),
		Positions: []*caes.Statement{pos, neg},
		Standard:  proofStandard,
	}
	a.Graph.Issues[issueID] = issue
	pos.Issue = issue
	neg.Issue = issue
	if target == pos {
		return neg, nil
	}
	return pos, nil
}

// AddUndercutter creates an undercutter attack against an argument.
//
// Semantics: an undercutter claims that the inference step of the target
// argument is inapplicable. In Carneades, this is modeled by:
//  1. Finding or creating an undercutter statement for the target argument.
//  2. Setting targetArg.Undercutter to point at that statement.
//  3. Creating a new argument concluding the undercutter statement.
//
// When the undercutter statement is labeled In, GetWeight (caes.go)
// returns 0.0 for the target argument, effectively defeating it.
//
// BUG FIX (critical): InstantiateScheme auto-generates an undercutter statement
// for every scheme instantiation (see InstantiateScheme in caes.go). The undercutter's ID
// uses a pseudo-argument-ID, so it may be "undercut(a3)" even though the actual
// argument is "a5". If we blindly create "undercut(a5)" and overwrite
// targetArg.Undercutter, we orphan the original undercutter and all its
// exception arguments. Instead, we REUSE the existing undercutter if one exists.
// This way, a new argument concluding the existing undercutter adds another
// reason for the undercutter to be In, which is the correct Carneades semantics.
//
// Parameters:
//   - targetArgID: ID of an existing argument to undercut.
//   - premiseIDs: IDs of existing statements that serve as premises.
func (a *Adapter) AddUndercutter(
	ownerID string,
	carneadesSchemeID string,
	targetArgID string,
	premiseIDs []string,
	weight float64,
) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Graph == nil {
		return nil, errors.New("no argument graph loaded")
	}

	targetArg := a.Graph.Arguments[targetArgID]
	if targetArg == nil {
		return nil, fmt.Errorf("target argument %q not found", targetArgID)
	}

	// BUG FIX: Reuse existing undercutter statement if one already exists.
	// This preserves the exception arguments created by InstantiateScheme.
	var ucStmt *caes.Statement
	if targetArg.Undercutter != nil {
		// Reuse — a new argument concluding this statement adds another
		// reason for the undercutter to be In.
		ucStmt = targetArg.Undercutter
	} else {
		// No existing undercutter — create one.
		ucID := fmt.Sprintf("undercut(%s)", targetArgID)
		ucStmt = a.findOrCreateStatement(ucID)
		ucStmt.Text = fmt.Sprintf("Argument %s is not applicable", targetArgID)
		ucStmt.IsUndercutter = true
		targetArg.Undercutter = ucStmt
	}

	// Resolve premise statements.
	premises := make([]*caes.Statement, 0, len(premiseIDs))
	for _, pid := range premiseIDs {
		ps := a.Graph.Statements[terms.Normalize(pid)]
		if ps == nil {
			return nil, fmt.Errorf("premise statement %q not found", pid)
		}
		premises = append(premises, ps)
	}

	argID, err := a.createArgument(ownerID, carneadesSchemeID, premises, ucStmt, weight)
	if err != nil {
		return nil, err
	}
	return []string{argID}, nil
}

// AddUndermining creates an undermining attack against a premise statement.
//
// Semantics: undermining argues that a premise of an existing argument is false.
// This is structurally identical to a rebuttal, but targets a premise rather
// than a conclusion: the new argument concludes the position opposing the
// premise (¬premise, at issue with the premise; see opposingPosition).
//
// If the negated premise wins the Issue resolution, the original premise
// becomes Out. Any argument that uses that premise will then have its weight
// reduced (or zeroed, for ConstantWeighingFunction/LinkedWeighingFunction
// which require all premises In).
//
// Parameters:
//   - targetPremiseID: ID of an existing statement (used as a premise) to undermine.
//   - premiseIDs: IDs of existing statements that serve as premises for the attack.
//   - proofStandard: the proof standard for the Issue.
func (a *Adapter) AddUndermining(
	ownerID string,
	carneadesSchemeID string,
	targetPremiseID string,
	premiseIDs []string,
	weight float64,
	proofStandard caes.Standard,
) ([]string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Graph == nil {
		return nil, errors.New("no argument graph loaded")
	}

	if strings.TrimSpace(targetPremiseID) == "" {
		return nil, errors.New("target premise ID is empty")
	}

	target := a.Graph.Statements[terms.Normalize(targetPremiseID)]
	if target == nil {
		return nil, fmt.Errorf("target premise %q not found", targetPremiseID)
	}
	return a.argueAgainst(ownerID, carneadesSchemeID, target, premiseIDs, weight, proofStandard)
}
