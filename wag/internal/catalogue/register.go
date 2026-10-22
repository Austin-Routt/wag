package catalogue

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// The decision register (docs/process.md §5): every open decision of the
// project, with what it is about, what it holds up, who decides it and how
// soon. Many are raised by catalogue entries, as discrepancy decisions that
// begin "Open"; the register lists each such item under a decision, and the
// checker keeps the two in step.
const (
	RegisterFile = "docs/decisions.yml" // the register, edited by hand
	RegisterView = "docs/decisions.md"  // its readable view, generated
)

// The register's vocabularies.
var (
	// DecisionCategories say what a decision is about: a scheme's formulas
	// or questions; how schemes meet (shared predicates and readings); the
	// game's fields, texts and edges; annotation and the gold; a reading of
	// the book; citations; the project itself.
	DecisionCategories = []string{"scheme", "vocabulary", "game", "annotation", "book", "metadata", "project"}
	// DecisionBlocks are the kinds of work a decision holds up.
	DecisionBlocks = []string{"game", "verification", "annotation", "play", "release"}
	Deciders       = []string{"owner", "verifier", "data", "play"}
	// Priority 1 holds up the next steps of the order (the game's fields
	// for the game's schemes, or verifying tier 1); 2, later work; 3 can wait.
	DecisionStatuses = []string{"open", "decided"}
)

// Register is the decision register.
type Register struct {
	Note      string     `yaml:"note,omitempty"`
	Decisions []Decision `yaml:"decisions"`
}

// Decision is one decision, open or taken.
type Decision struct {
	ID       string    `yaml:"id"`
	Title    string    `yaml:"title"`
	Question string    `yaml:"question"`
	Category string    `yaml:"category"`
	Blocks   []string  `yaml:"blocks,flow,omitempty"`
	Decider  string    `yaml:"decider"`
	Priority int       `yaml:"priority"`
	Status   string    `yaml:"status"`
	Items    []ItemRef `yaml:"items,omitempty"` // catalogue discrepancies that raise it
	Docs     []string  `yaml:"docs,omitempty"`  // documents that raise it
	Decided  *Outcome  `yaml:"decided,omitempty"`
}

// ItemRef names a catalogue entry's discrepancy by its key.
type ItemRef struct {
	Entry string `yaml:"entry"`
	Kind  string `yaml:"kind"`
	About string `yaml:"about"`
}

func (r ItemRef) String() string { return r.Entry + ": " + r.Kind + " / " + r.About }

// Outcome is what was decided.
type Outcome struct {
	Date     string `yaml:"date"`
	By       string `yaml:"by"`
	Decision string `yaml:"decision"`
	Recorded string `yaml:"recorded"` // where the decision is recorded
}

var openRE = regexp.MustCompile(`(^|\. )Open\b`)

// IsOpen tells whether a discrepancy's decision leaves something open: a
// sentence of it begins "Open".
func IsOpen(decision string) bool { return openRE.MatchString(decision) }

var decisionID = regexp.MustCompile(`^D[0-9]{3}$`)

// LoadRegister reads the register under the repository's root; nil if
// there is none.
func LoadRegister(root string) (*Register, error) {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(RegisterFile)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ParseRegister(b)
}

// ParseRegister decodes a register; unknown fields are errors.
func ParseRegister(b []byte) (*Register, error) {
	var r Register
	if err := yaml.UnmarshalStrict(b, &r); err != nil {
		return nil, fmt.Errorf("%s: %v", RegisterFile, err)
	}
	return &r, nil
}

// Open counts the open decisions, in all and by priority (index 1 to 3).
func (r *Register) Open() (n int, byPriority [4]int) {
	for _, d := range r.Decisions {
		if d.Status == "open" {
			n++
			if d.Priority >= 1 && d.Priority <= 3 {
				byPriority[d.Priority]++
			}
		}
	}
	return n, byPriority
}

// CheckRegister checks the register's own form, and keeps it in step with
// the catalogue's entries:
//   - every discrepancy decision that is open is listed under an open
//     decision of the register (an error otherwise);
//   - every item the register names exists (an error otherwise);
//   - an open decision's items are still open, and a taken decision's are
//     not (warnings: one side was updated and not the other).
func CheckRegister(reg *Register, entries []*Entry) []Finding {
	var out []Finding
	bad := func(f string, a ...any) {
		out = append(out, Finding{File: RegisterFile, Error: true, Msg: fmt.Sprintf(f, a...)})
	}
	warn := func(f string, a ...any) {
		out = append(out, Finding{File: RegisterFile, Msg: fmt.Sprintf(f, a...)})
	}
	items := map[ItemRef]*Discrepancy{}
	byID := map[string]*Entry{}
	for _, e := range entries {
		byID[e.ID] = e
		for i := range e.Discrepancies {
			d := &e.Discrepancies[i]
			items[ItemRef{e.ID, d.Kind, d.About}] = d
		}
	}
	listed := map[ItemRef]bool{}
	seen := map[string]bool{}
	for _, d := range reg.Decisions {
		where := d.ID
		if !decisionID.MatchString(d.ID) {
			bad("decision id %q: write D and three digits", d.ID)
		}
		if seen[d.ID] {
			bad("%s: the id is used twice", where)
		}
		seen[d.ID] = true
		if strings.TrimSpace(d.Title) == "" || strings.TrimSpace(d.Question) == "" {
			bad("%s: a title and a question are needed", where)
		}
		if !in(DecisionCategories, d.Category) {
			bad("%s: category %q is not one of %s", where, d.Category, strings.Join(DecisionCategories, ", "))
		}
		for _, b := range d.Blocks {
			if !in(DecisionBlocks, b) {
				bad("%s: blocks %q, not one of %s", where, b, strings.Join(DecisionBlocks, ", "))
			}
		}
		if !in(Deciders, d.Decider) {
			bad("%s: decider %q is not one of %s", where, d.Decider, strings.Join(Deciders, ", "))
		}
		if d.Priority < 1 || d.Priority > 3 {
			bad("%s: priority %d is not 1, 2 or 3", where, d.Priority)
		}
		switch d.Status {
		case "open":
			if d.Decided != nil {
				bad("%s: open, but it has an outcome", where)
			}
		case "decided":
			if o := d.Decided; o == nil || o.Date == "" || o.By == "" || o.Decision == "" || o.Recorded == "" {
				bad("%s: decided needs the outcome's date, by, decision and where it is recorded", where)
			}
		default:
			bad("%s: status %q is not open or decided", where, d.Status)
		}
		if len(d.Items) == 0 && len(d.Docs) == 0 {
			bad("%s: say where it is raised (items or docs)", where)
		}
		for _, it := range d.Items {
			disc, ok := items[it]
			if !ok {
				if byID[it.Entry] == nil {
					bad("%s: no catalogue entry %q", where, it.Entry)
				} else {
					bad("%s: %s has no discrepancy %s / %s", where, it.Entry, it.Kind, it.About)
				}
				continue
			}
			open := IsOpen(disc.Decision)
			switch {
			case d.Status == "open" && open:
				listed[it] = true
			case d.Status == "open":
				warn("%s is open, but %s no longer says Open: mark the decision decided, or drop the item", where, it)
			case open:
				warn("%s is decided, but %s still says Open: update the entry", where, it)
			}
		}
	}
	var keys []ItemRef
	for k, d := range items {
		if IsOpen(d.Decision) && !listed[k] {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, k := range keys {
		out = append(out, Finding{File: k.Entry + ".yml", Error: true,
			Msg: fmt.Sprintf("discrepancy %s / %s leaves a decision open that no open decision of %s lists", k.Kind, k.About, RegisterFile)})
	}
	return out
}

var priorityName = [4]string{"", "Priority 1: holds up the next steps", "Priority 2: holds up later work", "Priority 3: can wait"}

// RegisterMarkdown renders the register as a readable view (RegisterView).
func RegisterMarkdown(reg *Register) []byte {
	var b bytes.Buffer
	n, byP := reg.Open()
	deciders := map[string]int{}
	for _, d := range reg.Decisions {
		if d.Status == "open" {
			deciders[d.Decider]++
		}
	}
	b.WriteString("# Open decisions\n\n")
	b.WriteString("Generated by `wag-cli --catalogue-gen` from [`decisions.yml`](decisions.yml), the decision register ([`process.md`](process.md) §5). Edit that file, not this one.\n\n")
	fmt.Fprintf(&b, "**%d open:** %d of priority 1, %d of priority 2, %d of priority 3. **Who decides:**", n, byP[1], byP[2], byP[3])
	sep := " "
	for _, k := range Deciders {
		if deciders[k] > 0 {
			fmt.Fprintf(&b, "%s%s %d", sep, k, deciders[k])
			sep = ", "
		}
	}
	b.WriteString(".\n\n")
	if reg.Note != "" {
		b.WriteString(strings.TrimSpace(reg.Note) + "\n\n")
	}
	for p := 1; p <= 3; p++ {
		var rows []Decision
		for _, d := range reg.Decisions {
			if d.Status == "open" && d.Priority == p {
				rows = append(rows, d)
			}
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s\n\n", priorityName[p])
		b.WriteString("| Id | Decision | About | Holds up | Decides | Raised in |\n|---|---|---|---|---|---|\n")
		for _, d := range rows {
			fmt.Fprintf(&b, "| %s | **%s.** %s | %s | %s | %s | %s |\n", d.ID, cell(d.Title), cell(d.Question), d.Category,
				orDash(strings.Join(d.Blocks, ", ")), d.Decider, cell(raisedIn(d)))
		}
		b.WriteString("\n")
	}
	var done []Decision
	for _, d := range reg.Decisions {
		if d.Status == "decided" {
			done = append(done, d)
		}
	}
	if len(done) > 0 {
		b.WriteString("## Decided\n\n| Id | Decision | Date | By | Outcome | Recorded in |\n|---|---|---|---|---|---|\n")
		for _, d := range done {
			o := d.Decided
			fmt.Fprintf(&b, "| %s | **%s** | %s | %s | %s | %s |\n", d.ID, cell(d.Title), o.Date, o.By, cell(o.Decision), cell(o.Recorded))
		}
		b.WriteString("\n")
	}
	return b.Bytes()
}

func raisedIn(d Decision) string {
	var parts []string
	for _, it := range d.Items {
		parts = append(parts, "`"+it.Entry+"` ("+it.About+")")
	}
	parts = append(parts, d.Docs...)
	return strings.Join(parts, "; ")
}

func cell(s string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(s), " "), "|", `\|`)
}

func orDash(s string) string {
	if s == "" {
		return "–"
	}
	return s
}
