// Package nascq implements the Network of Argumentation Schemes and Critical
// Questions (NASCQ), a directed bipartite graph that governs the structural
// validity of moves in the Waltonian Argumentation Game.
//
// The NASCQ answers: "Is this move structurally valid in argumentation theory?"
// It does NOT evaluate argument content — that's the Carneades engine's job.
//
// Architecture (from the WAG plan, Part I §2):
//
//	NASCQ Layer → validates scheme exists, CQ is applicable, backup edge valid,
//	              semantic type compatible, variables complete
//	Carneades Adapter → translates move into graph operations
//	Carneades Engine → evaluates labels via GroundedLabelling
//	Game Layer → manages turns, scoring, termination
//
// Design decision D1: CQs are NOT Carneades graph nodes. They exist only in
// the NASCQ schema. When a player "answers a CQ," the answer becomes a
// Statement node in the Carneades ArgGraph.
//
// Design decision D2: NASCQ data lives in separate JSON files. Carneades YAML
// is the scheme store for the evaluation engine. NASCQ JSON is the scheme
// network for the game layer. They reference each other by ID (via
// ASNode.CarneadesID) but are maintained independently.
package nascq

// SemanticType classifies the kind of claim a scheme's conclusion makes
// or a CQ's answer expects. Used for thesis-dependent scheme filtering:
// a thesis-dependent scheme can only back up CQs within its semantic domain.
type SemanticType string

const (
	TypeAction      SemanticType = "action"
	TypeCausality   SemanticType = "causality"
	TypeValue       SemanticType = "value"
	TypeProperty    SemanticType = "property"
	TypePossibility SemanticType = "possibility"
	TypeExistence   SemanticType = "existence"
	TypeLikelihood  SemanticType = "likelihood"
	TypeTruth       SemanticType = "truth"
	TypeDefinition  SemanticType = "definition"
	TypeSource      SemanticType = "source"
)

// VariableSpec describes a variable slot in a scheme template.
// The Type and Description fields are used by the UI to guide the player
// in filling variable bindings.
type VariableSpec struct {
	Type        string `json:"type"`        // e.g., "entity", "goal", "action", "proposition"
	Description string `json:"description"` // human-readable hint for the UI
}

// ASNode represents an Argumentation Scheme node in the NASCQ graph.
//
// It carries two kinds of information:
//   - English templates for the UI (PremiseTemplatesEN, ConclusionTemplateEN)
//   - A link to the Carneades engine (CarneadesID → scheme ID in YAML)
//
// This intentional duplication exists because:
//   - Carneades YAML templates use Prolog-syntax predicates (e.g., "credible_source(W,D)")
//   - NASCQ English templates use human-readable text (e.g., "{W} is a credible source")
//   - The UI needs the English version; the engine needs the Prolog version
type ASNode struct {
	ID                   string                  `json:"id"`
	CarneadesID          string                  `json:"carneades_id"`           // links to Scheme.Id in Carneades YAML
	Name                 string                  `json:"name"`                   // human-readable scheme name
	SemanticType         SemanticType            `json:"semantic_type"`          // conclusion semantic category
	ThesisIndependent    bool                    `json:"thesis_independent"`     // if true, backs ALL CQs
	BaseWeight           float64                 `json:"base_weight"`            // default weight for Trusted mode
	Variables            map[string]VariableSpec `json:"variables"`              // variable slots
	PremiseTemplatesEN   []string                `json:"premise_templates_en"`   // English premise patterns
	ConclusionTemplateEN string                  `json:"conclusion_template_en"` // English conclusion pattern

	// AnswerVariable is the variable that states the scheme's conclusion
	// (S of "S is true, since expert W asserts it"). When the scheme backs
	// the answer to a critical question, it is bound to the answer, so the
	// argument's premises are about the answer (asserts(W, ¬possible(A))).
	// A scheme without one concludes something more specific
	// (should_be_performed(A)); its answering argument concludes the answer
	// and leaves its own conclusion as the unstated link (an enthymeme).
	AnswerVariable string `json:"answer_variable,omitempty"`
}

// CaesMapping specifies how a CQ-directed attack translates into a CAES
// graph operation. This is the critical bridge between NASCQ semantics
// (which CQ is being answered) and Carneades mechanics (what graph mutation
// to perform).
//
// Examples:
//   - Side effects CQ: negative answer → undercutter (defeats the argument's inference)
//   - Other means CQ: positive answer → rebuttal (argues for alternative conclusion)
//   - Possibility CQ: negative answer → undermining (attacks the "possible(A)" premise)
type CaesMapping struct {
	PositiveAnswerType   string `json:"positive_answer_type"`             // "support" or attack type
	NegativeAnswerType   string `json:"negative_answer_type"`             // "undercutter", "rebuttal", or "undermining"
	Target               string `json:"target"`                           // "argument", "conclusion", or "premise"
	TargetPremisePattern string `json:"target_premise_pattern,omitempty"` // for undermining: which premise to negate
}

// CQNode represents a Critical Question node in the NASCQ graph.
//
// CQs serve as bridges between schemes. They encode:
//   - What vulnerability a scheme has (via ask edges)
//   - How that vulnerability can be exploited (via CaesMapping)
//   - Which schemes can provide answers (via backup edges)
type CQNode struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name"`                 // human-readable CQ name
	SemanticType       SemanticType `json:"semantic_type"`        // answer semantic category
	QuestionTemplateEN string       `json:"question_template_en"` // English question pattern
	CaesMapping        CaesMapping  `json:"caes_mapping"`         // how to translate to CAES operations
}

// Edge represents a connection between an AS-node and a CQ-node.
// The Type field determines the direction and meaning:
//   - "ask": CQ → AS (this CQ challenges arguments using this scheme)
//   - "backup": AS → CQ (this scheme can support an answer to this CQ)
type Edge struct {
	ID          string `json:"id,omitempty"`     // optional edge identifier
	From        string `json:"from,omitempty"`   // source node ID (Python prototype format)
	To          string `json:"to,omitempty"`     // target node ID (Python prototype format)
	Type        string `json:"type,omitempty"`   // "ask" or "backup" (Python prototype format)
	CQ          string `json:"cq,omitempty"`     // CQ node ID (plan format)
	Scheme      string `json:"scheme,omitempty"` // AS node ID (plan format)
	Description string `json:"description,omitempty"`

	// Locus and Pattern (ask edges only) say which part of the scheme the
	// question is about; see Locus. Without them the CQ's caes_mapping
	// decides (NASCQ.LocusOf).
	Locus   string `json:"locus,omitempty"`
	Pattern string `json:"pattern,omitempty"`
}

// Locus kinds: the part of an argument a critical question is about.
//
// Gordon, Prakken and Walton (2007) model the critical questions of a scheme
// as its premises, assumptions and exceptions, and the Carneades scheme files
// declare exactly these; Pyon and Ri (2022, §6) build on that model: the
// positive answer to a critical question is a premise of the scheme. A
// critical question asked of a scheme is therefore about
//
//   - a premise or assumption P (LocusPremise): the positive answer is P, the
//     negative answer ¬P undermines the argument;
//   - an exception E (LocusException): the negative answer is E, which
//     undercuts the argument through the scheme's exception argument, the
//     positive answer is ¬E;
//   - the conclusion C (LocusConclusion): the negative answer rebuts C, the
//     positive answer supports C;
//   - the inference as a whole, when no formula of the scheme states the
//     question (LocusArgument): the negative answer undercuts the argument,
//     the positive answer denies the undercutter.
const (
	LocusPremise    = "premise"
	LocusException  = "exception"
	LocusConclusion = "conclusion"
	LocusArgument   = "argument"
)

// Locus is where a critical question applies to the arguments of one scheme.
type Locus struct {
	Kind    string `json:"kind"`              // LocusPremise, LocusException, LocusConclusion or LocusArgument
	Pattern string `json:"pattern,omitempty"` // the premise or exception, with {Var} placeholders
}

// AttackType is the CAES attack that a negative answer about this locus makes.
func (l Locus) AttackType() string {
	switch l.Kind {
	case LocusPremise:
		return "undermining"
	case LocusConclusion:
		return "rebuttal"
	case LocusException, LocusArgument:
		return "undercutter"
	}
	return ""
}

// ProposedAction represents a player's intended move before validation.
// The NASCQ layer validates this against its structural rules (S1–S8)
// and produces a ValidatedAction if all checks pass.
type ProposedAction struct {
	Type             string            `json:"type"`              // "support" or "attack"
	AttackType       string            `json:"attack_type"`       // "rebuttal", "undercutter", "undermining" (attacks only)
	TargetArgID      string            `json:"target_arg_id"`     // ID of argument being attacked
	TargetPremiseID  string            `json:"target_premise_id"` // for undermining: which premise
	SchemeID         string            `json:"scheme_id"`         // NASCQ scheme ID (e.g., "AS:expert_opinion")
	AnswersCQID      string            `json:"answers_cq_id"`     // CQ being answered (empty for direct support)
	VariableBindings map[string]string `json:"variable_bindings"` // {VarName: "user text"}
}

// ValidatedAction is the output of NASCQ validation. It contains
// the resolved Carneades scheme ID, instantiated templates, the CAES
// attack type (from CaesMapping), and the scheme's base weight.
// This is passed to the Carneades adapter for execution.
type ValidatedAction struct {
	ProposedAction
	CarneadesSchemeID string   `json:"carneades_scheme_id"` // resolved from ASNode.CarneadesID
	InstPremises      []string `json:"inst_premises"`       // substituted English premises
	InstConclusion    string   `json:"inst_conclusion"`     // substituted English conclusion
	CaesAttackType    string   `json:"caes_attack_type"`    // attack type implied by the locus ("" for supports and defences)
	Weight            float64  `json:"weight"`              // scheme BaseWeight

	// For a move with a target: the part of the target it is about, and the
	// premise or exception statement for those loci.
	Locus       Locus  `json:"locus"`
	LocusStmtID string `json:"locus_stmt_id,omitempty"`

	// AnswerVariable is the backing scheme's AnswerVariable on a move with a
	// target: the adapter binds it to the answer (S2 does not require it).
	AnswerVariable string `json:"answer_variable,omitempty"`
}

// IsDefence reports whether the move answers in the target's favour.
func (v *ValidatedAction) IsDefence() bool { return v.Type == "support" && v.TargetArgID != "" }
