// Package catalogue holds WAG's scheme catalogue: one entry per
// argumentation scheme, aligning the book (Walton, Reed & Macagno 2008), the
// Carneades collection (walton.yml) and the game's NASCQ, with a status on
// everything people or models write (docs/plan.md §2).
//
// An entry has two kinds of field:
//
//   - Source fields are copied from the authority for that part and kept in
//     step by the generator: the scheme's id, title and formulas
//     (walton.yml), and the game's critical questions (the NASCQ). The
//     English readings of the formulas are not stored: 173 of walton.yml's
//     393 templates repeat the book's wording, so they are rendered from
//     walton.yml when needed (English).
//     They carry no status and are never edited by hand.
//   - Work fields are drafted, then verified: the book links (from
//     links.json, itself drafted), the families, the summary, the
//     paraphrased critical questions, contrasts, examples, the
//     classification, the dialogue types, the discrepancies. Each carries a
//     status (Status). The generator rewrites a work item only while its
//     status is "generated" or "todo"; once someone drafts, verifies or
//     disputes it, the item is theirs.
package catalogue

import (
	"fmt"
	"regexp"
	"time"
)

// Entry is one scheme of the catalogue, stored as schemes/<id>.yml.
type Entry struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	// Tier 1 schemes are verified first (plan §2); 2 is the rest. Set when
	// the entry is created, then left to people.
	Tier int `yaml:"tier"`
	// Walton is the walton.yml scheme the entry formalizes, nil for a book
	// scheme that walton.yml lacks. Source.
	Walton *WaltonRef `yaml:"walton"`
	// Game is the NASCQ scheme that plays it, if any. Source.
	Game *GameScheme `yaml:"game,omitempty"`
	// Formulas are walton.yml's. Source; nil when Walton is nil.
	Formulas *Formulas `yaml:"formulas"`

	Book          []BookRef      `yaml:"book"`
	Families      []FamilyRef    `yaml:"families"`
	PremiseMap    []PremiseLink  `yaml:"premise_map"`
	CQs           []CQ           `yaml:"critical_questions"`
	Summary       Text           `yaml:"summary"`
	Contrasts     []Contrast     `yaml:"contrasts"`
	Examples      []Example      `yaml:"examples"`
	Discrepancies []Discrepancy  `yaml:"discrepancies"`
	Class         Classification `yaml:"classification"`
	Specialises   Specialises    `yaml:"specialises"`
	DialogueTypes DialogueTypes  `yaml:"dialogue_types"`
	History       []Event        `yaml:"history"`
}

// WaltonRef names the walton.yml scheme and its own citation (meta.source).
type WaltonRef struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Source string `yaml:"source,omitempty"`
}

// GameScheme is the NASCQ scheme that plays the entry: its id, and the
// critical questions it can back an answer to (backup edges).
type GameScheme struct {
	ID    string   `yaml:"id"`
	Backs []string `yaml:"backs,flow,omitempty"`
}

// Formulas are a scheme's variables and formulas by role, in walton.yml's
// order.
type Formulas struct {
	Variables   []string `yaml:"variables,flow"`
	Premises    []string `yaml:"premises,omitempty"`
	Assumptions []string `yaml:"assumptions,omitempty"`
	Exceptions  []string `yaml:"exceptions,omitempty"`
	Conclusions []string `yaml:"conclusions"`
}

// BookRef is a compendium entry (the book's ch. 9) the scheme formalizes.
type BookRef struct {
	Entry string `yaml:"entry"`
	Title string `yaml:"title"`
	Pages string `yaml:"pages"`
	// Parent is the compendium heading the entry sits under, if any.
	Parent string `yaml:"parent,omitempty"`
	// Match is how closely the scheme formalizes the entry: exact, partial.
	Match string `yaml:"match"`
	// CQsPrintedUnder is the entry whose list of critical questions this
	// entry shares (the book prints it once).
	CQsPrintedUnder string `yaml:"cqs_printed_under,omitempty"`
	Note            string `yaml:"note,omitempty"`
	Status          string `yaml:"status"`
}

// FamilyRef places the scheme in a family of families.json.
type FamilyRef struct {
	Family   string `yaml:"family"`
	Name     string `yaml:"name"`
	Relation string `yaml:"relation"`
	Note     string `yaml:"note,omitempty"`
	Status   string `yaml:"status"`
}

// PremiseLink maps a premise (or the conclusion) of a compendium entry, by
// its printed label, to the formulas that express it.
type PremiseLink struct {
	Entry    string   `yaml:"entry"`
	Label    string   `yaml:"label"`
	Formulas []Target `yaml:"formulas,flow"`
	Note     string   `yaml:"note,omitempty"`
	Status   string   `yaml:"status"`
}

// Target is a formula of the scheme and its role (locus): premise,
// assumption, exception or conclusion.
type Target struct {
	Formula string `yaml:"formula"`
	Locus   string `yaml:"locus"`
}

// CQ is a critical question of the scheme: the book's (by entry and printed
// label), the game's, or both.
type CQ struct {
	// ID is "<entry>/<label>" for a book question (the entry under which it
	// is printed), "game/<NASCQ id>" for a game question that matches no
	// book question, or "walton/<predicate>" for a question that only
	// walton.yml implies: an assumption or exception that neither the book
	// nor the game asks about, added by whoever drafts the entry.
	ID   string  `yaml:"id"`
	Book *BookCQ `yaml:"book,omitempty"`
	Game *GameCQ `yaml:"game,omitempty"`
	// Targets are the formulas the question puts in doubt. None: the
	// question attacks the inference as a whole (Undercuts), or no formula
	// encodes it yet.
	Targets   []Target `yaml:"targets,flow"`
	Undercuts bool     `yaml:"undercuts,omitempty"`
	Note      string   `yaml:"note,omitempty"`
	// Paraphrase is the question in our own words, with the scheme's
	// variables; never the book's sentence.
	Paraphrase string `yaml:"paraphrase"`
	// Answer says what counts as an answer, and what each answer does.
	Answer string `yaml:"answer"`
	Status string `yaml:"status"`
}

// BookCQ is a critical question as the compendium prints it: its entry and
// label. The book's wording is not reproduced.
type BookCQ struct {
	Entry string `yaml:"entry"`
	Label string `yaml:"label"`
}

// GameCQ is the game's critical question (source: the NASCQ): its id, its
// question template, and the ask edge from this scheme.
type GameCQ struct {
	ID      string `yaml:"id"`
	Text    string `yaml:"text"`
	Locus   string `yaml:"locus"`
	Pattern string `yaml:"pattern,omitempty"`
}

// Text is a piece of our own writing and its status.
type Text struct {
	Text   string `yaml:"text"`
	Status string `yaml:"status"`
}

// Contrast says when to choose this scheme rather than a close relative.
type Contrast struct {
	Scheme string `yaml:"scheme"`
	Text   string `yaml:"text"`
	Status string `yaml:"status"`
}

// Example is a short claim coded with the scheme, or a near miss.
type Example struct {
	Text     string `yaml:"text"`
	Kialo    string `yaml:"kialo,omitempty"`
	Coding   string `yaml:"coding"`
	NearMiss bool   `yaml:"near_miss,omitempty"`
	Status   string `yaml:"status"`
}

// Discrepancy records where the sources disagree, and the decision.
type Discrepancy struct {
	Kind     string `yaml:"kind"`
	About    string `yaml:"about"`
	Detail   string `yaml:"detail"`
	Decision string `yaml:"decision"`
	Status   string `yaml:"status"`
}

// Classification places the scheme in Walton and Macagno's tree (plan §2):
// source-dependent or not; then, for a source-independent scheme, practical
// or epistemic; then, for epistemic reasoning, discovering a rule or entity
// or applying one to a case. Group is one of the tree's groups under that
// heading (Groups). Note says how the scheme was placed when the tree does
// not list it.
type Classification struct {
	Source    string `yaml:"source"`
	Reasoning string `yaml:"reasoning"`
	Epistemic string `yaml:"epistemic"`
	Group     string `yaml:"group"`
	Note      string `yaml:"note,omitempty"`
	Status    string `yaml:"status"`
}

// Heading is the tree's heading the classification falls under: dependent
// (source-dependent), practical, discovery or application; "" if not
// placed yet.
func (c Classification) Heading() string {
	switch {
	case c.Source == "dependent":
		return "dependent"
	case c.Reasoning == "practical":
		return "practical"
	default:
		return c.Epistemic
	}
}

// Specialises names the scheme this one is a domain version of, if any.
type Specialises struct {
	Scheme string `yaml:"scheme"`
	Note   string `yaml:"note"`
	Status string `yaml:"status"`
}

// DialogueTypes are the dialogue types in which the scheme is a proper move,
// and those in which it is suspect (docs/dialogue-types.md §7).
type DialogueTypes struct {
	Proper  []string `yaml:"proper,flow"`
	Suspect []string `yaml:"suspect,flow"`
	Note    string   `yaml:"note"`
	Status  string   `yaml:"status"`
}

// Event is a line of the entry's history. Field-level history is the git
// log of the file.
type Event struct {
	Date string `yaml:"date"`
	By   string `yaml:"by"`
	What string `yaml:"what"`
}

// Statuses.
const (
	Generated = "generated" // made by the generator from a source that is itself only drafted
	Todo      = "todo"      // empty, to be written
)

var statusRE = regexp.MustCompile(`^(?:generated|todo|(drafted|verified) by ([^,:]+), (\d{4}-\d{2}-\d{2})|(disputed) by ([^,:]+), (\d{4}-\d{2}-\d{2}): (.+))$`)

// Status is a parsed status: "generated", "todo", "drafted by WHO, DATE",
// "verified by WHO, DATE" or "disputed by WHO, DATE: REASON".
type Status struct {
	State, By, Date, Reason string
}

// ParseStatus parses a status string.
func ParseStatus(s string) (Status, error) {
	m := statusRE.FindStringSubmatch(s)
	if m == nil {
		return Status{}, fmt.Errorf("status %q: want generated, todo, \"drafted by WHO, YYYY-MM-DD\", \"verified by WHO, YYYY-MM-DD\" or \"disputed by WHO, YYYY-MM-DD: REASON\"", s)
	}
	var st Status
	switch {
	case m[1] != "":
		st = Status{State: m[1], By: m[2], Date: m[3]}
	case m[4] != "":
		st = Status{State: m[4], By: m[5], Date: m[6], Reason: m[7]}
	default:
		return Status{State: s}, nil
	}
	if _, err := time.Parse("2006-01-02", st.Date); err != nil {
		return Status{}, fmt.Errorf("status %q: bad date %q", s, st.Date)
	}
	return st, nil
}

// owned reports whether a work item belongs to the generator: it may be
// rewritten while its status is generated or todo.
func owned(status string) bool { return status == Generated || status == Todo || status == "" }

// Vocabularies.
var (
	Loci              = []string{"premise", "assumption", "exception", "conclusion"}
	DialogueTypeNames = []string{"persuasion", "inquiry", "information_seeking", "negotiation", "deliberation", "eristic"}
	SourceValues      = []string{"", "dependent", "independent"}
	ReasoningValues   = []string{"", "practical", "epistemic"}
	EpistemicValues   = []string{"", "discovery", "application"}
	// Groups are the thematic groups of Walton and Macagno's tree (2015, p.
	// 22; Aberdein 2020, Table 4), by heading.
	Groups = map[string][]string{
		"discovery":   {"establishing rules", "finding entities"},
		"application": {"based on cases", "defeasible rule-based", "chained"},
		"practical":   {"instrumental", "values", "value-based"},
		"dependent":   {"position to know", "ad hominem", "popular acceptance"},
	}
	DiscrepancyKinds = []string{
		"partial-link",     // the scheme formalizes the book entry only in part (links.json)
		"no-formalization", // a book scheme that walton.yml lacks
		"not-in-book",      // a walton.yml scheme the compendium does not list
		"game-differs",     // the game's copy of the scheme (core.yml) differs from walton.yml
		"link-formula",     // links.json names a formula that is not the scheme's
		"book-walton",      // the book and walton.yml say different things (drafted by people)
		"other",
	}
)

func in(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
