// Package evidence checks retained evidence records and replays a bounded
// admission policy. It never assigns truth, argument weights or engine labels.
package evidence

const Format = "wag-evidence/0.1"
const Policy = "two-party-consideration/1"

// Bundle is an offline snapshot. Native WAG moves still use citation strings;
// this contract does not silently migrate or reinterpret those historical moves.
type Bundle struct {
	Format          string           `json:"format"`
	Context         Context          `json:"context"`
	Actors          []Actor          `json:"actors"`
	Sources         []Source         `json:"sources"`
	Artifacts       []Artifact       `json:"artifacts"`
	Representations []Representation `json:"representations"`
	Claims          []Claim          `json:"claims"`
	Targets         []Target         `json:"targets"`
	Bindings        []Binding        `json:"bindings"`
	Assessments     []Assessment     `json:"assessments"`
	Dependencies    []Dependency     `json:"dependencies"`
	Events          []Event          `json:"events"`
}

type Context struct {
	ID              string   `json:"id"`
	Profile         string   `json:"profile"`
	Policy          string   `json:"policy"`
	Participants    []string `json:"participants"`
	EngineSHA256    string   `json:"engine_sha256"`
	CatalogueSHA256 string   `json:"catalogue_sha256"`
	NASCQSHA256     string   `json:"nascq_sha256"`
}

type Actor struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"` // human, model, tool; identity is declared, not authenticated here
	Label string `json:"label"`
}

type Record struct {
	ID        string `json:"id"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

// Source identity is descriptive. Expert competence is assessed in context,
// rather than encoded as a globally trusted source kind.
type Source struct {
	Record
	Kind      string `json:"kind"` // person, organization, document, device, model, unknown
	Label     string `json:"label"`
	Reference string `json:"reference"`
}

type Artifact struct {
	Record
	SourceID      string   `json:"source_id"`
	Revision      string   `json:"revision"`
	CaptureStatus string   `json:"capture_status"` // complete, partial, failed, unsupported, unattempted
	Scope         string   `json:"scope"`
	Omissions     []string `json:"omissions"`
	Failure       string   `json:"failure"`
}

type Payload struct {
	Path      string `json:"path"` // portable relative path within the explicitly supplied root
	SHA256    string `json:"sha256"`
	Bytes     int64  `json:"bytes"`
	MediaType string `json:"media_type"`
}

type Representation struct {
	Record
	ArtifactID string   `json:"artifact_id"`
	Kind       string   `json:"kind"` // original, ocr, transcript, translation, excerpt, normalized
	ParentIDs  []string `json:"parent_ids"`
	Method     string   `json:"method"`
	Payload    Payload  `json:"payload"`
	Pages      int      `json:"pages"`       // declared metadata; PDF pagination is not independently parsed
	DurationMS int64    `json:"duration_ms"` // declared metadata; media duration is not independently parsed
}

// Text selectors use Unicode code-point offsets into an exact retained UTF-8
// representation. Other selectors receive structural/bounds checks only.
type Selector struct {
	Kind    string  `json:"kind"` // whole, text, page, time, region
	Start   int64   `json:"start"`
	End     int64   `json:"end"`
	Quote   string  `json:"quote"`
	Page    int     `json:"page"`
	StartMS int64   `json:"start_ms"`
	EndMS   int64   `json:"end_ms"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

// A Claim records what a source is reported to say, or an explicitly marked
// observation/reconstruction. Its inclusion does not assert its truth.
type Claim struct {
	Record
	RepresentationID string   `json:"representation_id"`
	AttributedTo     string   `json:"attributed_to"` // source ID, possibly an explicitly unknown source
	Mode             string   `json:"mode"`          // quotation, paraphrase, observation, reconstruction
	Text             string   `json:"text"`
	Scope            string   `json:"scope"`
	Selector         Selector `json:"selector"`
}

// Targets describe a pinned argument snapshot. This checker verifies internal
// references, not correspondence with a running game's graph; the adapter must
// perform that separate check before using these records in a game.
type Target struct {
	Record
	ArgumentID       string `json:"argument_id"`
	ArgumentRevision string `json:"argument_revision"`
	Kind             string `json:"kind"` // premise, assumption, exception, conclusion, inference, source, annotation
	Proposition      string `json:"proposition"`
}

type Binding struct {
	Record
	ClaimID        string `json:"claim_id"`
	TargetID       string `json:"target_id"`
	Role           string `json:"role"` // support, oppose, context, definition, source_challenge
	Interpretation string `json:"interpretation"`
	Warrant        string `json:"warrant"`
}

type Assessment struct {
	Record
	BindingID     string            `json:"binding_id"`
	Dimensions    map[string]string `json:"dimensions"`
	Rationale     string            `json:"rationale"`
	BasisClaimIDs []string          `json:"basis_claim_ids"`
	ReviewLevel   string            `json:"review_level"` // unreviewed, model_checked, human_checked
	FactCheck     FactCheck         `json:"fact_check"`
}

type FactCheck struct {
	Status        string   `json:"status"` // not_checked, supported, contradicted, inconclusive
	Method        string   `json:"method"`
	BasisClaimIDs []string `json:"basis_claim_ids"`
}

// Dependencies record claims about source relations, not calculated probability
// or a declaration that two agreeing assessors are statistically independent.
type Dependency struct {
	Record
	ArtifactIDs []string `json:"artifact_ids"`
	Relation    string   `json:"relation"` // common_origin, derived, independence_claim, unknown
	Rationale   string   `json:"rationale"`
}

type Event struct {
	Record
	Sequence        int    `json:"sequence"`
	Type            string `json:"type"` // consent, withdraw
	BindingID       string `json:"binding_id"`
	BindingRevision string `json:"binding_revision"`
	Purpose         string `json:"purpose"` // consideration; never acceptance-as-true
}

type Issue struct {
	Code     string `json:"code"`
	Location string `json:"location"`
	Message  string `json:"message"`
}

type Report struct {
	Format            string              `json:"format"`
	Valid             bool                `json:"valid"`
	Scope             string              `json:"scope"`
	Errors            []Issue             `json:"errors"`
	Warnings          []Issue             `json:"warnings"`
	BindingRevisions  map[string]string   `json:"binding_revisions"`
	Consents          map[string][]string `json:"consents"`
	AdmittedBindings  []string            `json:"admitted_for_consideration"`
	DuplicatePayloads [][]string          `json:"duplicate_payload_groups"`
}
