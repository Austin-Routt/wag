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

// The candidates registry (docs/process.md §8): schemes, questions,
// exceptions and links that WAG does not have yet but that the literature,
// data or play suggest it may need. Each candidate keeps its evidence, the
// items it comes closest to and why each fails, a sketch checked like a
// proposal's candidates, and examples read in English. Candidates are
// grouped by family and pattern, so that similar ones are compared side by
// side and decided together. The view is generated from the registry.
const (
	CandidatesFile = "docs/candidates.yml" // the registry, edited by hand
	CandidatesView = "docs/candidates.md"  // its readable view, generated
)

// The registry's vocabularies.
var (
	// CandidateKinds: a new scheme; a question or an exception for schemes
	// WAG has; a link (a rule about how schemes meet, or a reading).
	CandidateKinds = []string{"scheme", "question", "exception", "link"}
	// SightingSources are the places a sighting counts from: a Kialo
	// discussion, a game, or an annotated text outside Kialo.
	SightingSources = []string{"kialo", "game", "annotation"}
	// Fits say how well a sighting fits the pattern.
	Fits = []string{"clean", "strain"}
	// Confidences are the finder's, that the candidate is a real gap.
	Confidences = []string{"high", "medium", "low"}
	// Results are the four outcomes of process.md §8.
	Results = []string{"reject", "merge", "domain", "add"}
)

// The stages of a candidate, in order.
const (
	StageCollecting = "collecting" // below the threshold
	StageDrafting   = "drafting"   // over the threshold; the minimum not complete
	StageReady      = "ready"      // the minimum complete: it goes to a person
	StageDecided    = "decided"    // an outcome recorded
)

// CandidateRegister is the candidates registry.
type CandidateRegister struct {
	Note       string           `yaml:"note"`
	Threshold  Threshold        `yaml:"threshold"`
	Groups     []CandidateGroup `yaml:"groups"`
	Candidates []Candidate      `yaml:"candidates"`
}

// Threshold is when a candidate goes to a person: at least Sightings
// verified sightings in at least Places discussions or games. Decision is
// the register's decision on the threshold itself.
type Threshold struct {
	Sightings int    `yaml:"sightings"`
	Places    int    `yaml:"places"`
	Decision  string `yaml:"decision"`
}

// CandidateGroup is a pattern that several candidates share, within one
// family of families.json or across families.
type CandidateGroup struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Families []string `yaml:"families,flow,omitempty"` // none: across families
	About    string   `yaml:"about"`                   // what the members share, and how they differ
}

// Candidate is one scheme, question, exception or link that WAG may need.
type Candidate struct {
	ID         string            `yaml:"id"`
	Name       string            `yaml:"name"`
	Kind       string            `yaml:"kind"`
	Group      string            `yaml:"group"`
	Decision   string            `yaml:"decision,omitempty"` // the register's decision that settles it
	Outcome    *CandidateOutcome `yaml:"outcome,omitempty"`
	Found      []string          `yaml:"found"`      // who found it, and when
	Confidence string            `yaml:"confidence"` // the finder's
	What       string            `yaml:"what"`
	Compares   string            `yaml:"compares"`  // for the group's table: what it weighs, relates or asks
	Concludes  string            `yaml:"concludes"` // for the group's table: its conclusion, or what it attacks
	Closest    []Closest         `yaml:"closest"`
	Related    []Related         `yaml:"related,omitempty"`
	Sources    []string          `yaml:"sources,omitempty"` // the literature, the book's and ND's pages
	Sightings  []Sighting        `yaml:"sightings,omitempty"`
	// Exempt says why the threshold does not apply: a defect or gap shown
	// by analysis (an engine test, a scan of walton.yml), not a pattern
	// that data has to show.
	Exempt   string            `yaml:"exempt,omitempty"`
	Sketch   *Sketch           `yaml:"sketch,omitempty"`
	Examples []ProposalExample `yaml:"examples,omitempty"`
	Engine   []string          `yaml:"engine,omitempty"` // what tests in WAG's engine showed
	Needs    []string          `yaml:"needs,omitempty"`  // what implementing it needs
	Open     []string          `yaml:"open,omitempty"`   // open questions
}

// CandidateOutcome is what a person decided.
type CandidateOutcome struct {
	Date     string `yaml:"date"`
	By       string `yaml:"by"`
	Result   string `yaml:"result"`         // reject, merge, domain, add
	Into     string `yaml:"into,omitempty"` // merged into, or the domain version of
	Recorded string `yaml:"recorded"`       // where: a proposal, a catalogue entry
}

// Closest is an item WAG has, or another candidate, and why it does not
// cover the candidate.
type Closest struct {
	Item string `yaml:"item"`
	Why  string `yaml:"why"`
}

// Related is another candidate and how the two relate (overlap, merge,
// sibling).
type Related struct {
	Candidate string `yaml:"candidate"`
	How       string `yaml:"how"`
}

// Sighting is one occurrence of the pattern in data. Only a verified
// sighting counts toward the threshold: a second reader has checked the text
// against the note.
type Sighting struct {
	Source   string `yaml:"source"`
	Ref      string `yaml:"ref"`             // a Kialo claim (discussion.claim), a game or text id
	Place    string `yaml:"place,omitempty"` // the discussion or game; a Kialo ref's own
	Fit      string `yaml:"fit"`
	Archived bool   `yaml:"archived,omitempty"` // a Kialo claim the discussion has archived
	Note     string `yaml:"note"`               // what it says and how it fits, in our words
	Found    string `yaml:"found"`
	Verified string `yaml:"verified,omitempty"`
}

// Sketch is the candidate as walton.yml might have it: a new scheme, or a
// scheme WAG has with the question or exception added, under its own id.
// It is checked as a proposal's candidates are, and never applied.
type Sketch struct {
	Note      string              `yaml:"note,omitempty"`
	Schemes   []Scheme            `yaml:"schemes"`
	Language  map[string]string   `yaml:"language,omitempty"`
	Questions map[string][]PQuery `yaml:"questions,omitempty"`
}

var (
	candidateID = regexp.MustCompile(`^C[0-9]{3}$`)
	groupID     = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	kialoRef    = regexp.MustCompile(`^([0-9]+)\.[0-9]+$`)
)

// LoadCandidates reads the registry under the repository's root; nil if
// there is none.
func LoadCandidates(root string) (*CandidateRegister, error) {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(CandidatesFile)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ParseCandidates(b)
}

// ParseCandidates decodes a registry; unknown fields are errors.
func ParseCandidates(b []byte) (*CandidateRegister, error) {
	var r CandidateRegister
	if err := yaml.UnmarshalStrict(b, &r); err != nil {
		return nil, fmt.Errorf("%s: %v", CandidatesFile, err)
	}
	return &r, nil
}

// place is the discussion or game a sighting was seen in.
func (s Sighting) place() string {
	if s.Place != "" {
		return s.Place
	}
	if m := kialoRef.FindStringSubmatch(s.Ref); s.Source == "kialo" && m != nil {
		return m[1]
	}
	return ""
}

// Seen counts a candidate's sightings: all, the verified ones, and the
// places of the verified ones.
func (c *Candidate) Seen() (all, verified, places int) {
	at := map[string]bool{}
	for _, s := range c.Sightings {
		all++
		if s.Verified != "" {
			verified++
			at[s.place()] = true
		}
	}
	return all, verified, len(at)
}

// Met tells whether the candidate's verified sightings meet the threshold,
// or it is exempt.
func (c *Candidate) Met(th Threshold) bool {
	if c.Exempt != "" {
		return true
	}
	_, v, p := c.Seen()
	return v >= th.Sightings && p >= th.Places
}

// sketched reports whether a scheme id is one of the candidate's sketch.
func (c *Candidate) sketched(id string) bool {
	if c.Sketch == nil {
		return false
	}
	return in(schemeIDs(c.Sketch.Schemes), id)
}

// Missing lists what the candidate still lacks of the minimum a person
// needs to decide on it (process.md §8): the threshold; the items it comes
// closest to and why each fails; for a scheme, question or exception, a
// sketch, three examples of its use and a near miss.
func (c *Candidate) Missing(th Threshold) []string {
	var out []string
	if !c.Met(th) {
		_, v, p := c.Seen()
		out = append(out, fmt.Sprintf("the threshold (%d verified in %d places; %d in %d needed)", v, p, th.Sightings, th.Places))
	}
	if len(c.Closest) == 0 {
		out = append(out, "the closest items")
	}
	if c.Kind == "link" {
		return out
	}
	if c.Sketch == nil || len(c.Sketch.Schemes) == 0 {
		out = append(out, "a sketch")
	}
	uses, near := 0, 0
	for _, x := range c.Examples {
		switch {
		case x.Use == "near-miss" && c.sketched(x.Of):
			near++
		case x.Use != "near-miss" && c.sketched(x.Scheme):
			uses++
		}
	}
	if uses < 3 {
		out = append(out, fmt.Sprintf("examples (%d of 3)", uses))
	}
	if near == 0 {
		out = append(out, "a near miss")
	}
	return out
}

// Stage is where the candidate stands.
func (c *Candidate) Stage(th Threshold) string {
	switch {
	case c.Outcome != nil:
		return StageDecided
	case !c.Met(th):
		return StageCollecting
	case len(c.Missing(th)) > 0:
		return StageDrafting
	}
	return StageReady
}

// asProposal is the candidate's sketch and examples as a proposal's
// candidates, for the proposal check and the rendering.
func (c *Candidate) asProposal() *Proposal {
	p := &Proposal{Decisions: []string{c.ID}, Title: c.Name, Examples: c.Examples}
	if s := c.Sketch; s != nil {
		p.Candidates = &Candidates{Note: s.Note, Schemes: s.Schemes, Language: s.Language, Questions: s.Questions}
	}
	return p
}

// CheckCandidates checks the registry's form, and against the decision
// register and the sources:
//   - ids, groups, kinds, fits, confidences and outcomes are well formed,
//     groups name families that families.json has, and every candidate's
//     group exists;
//   - a candidate's decision is in the register, open until an outcome is
//     recorded and decided after; a candidate over the threshold has one
//     (a warning otherwise);
//   - a sighting says where it was seen, how it fits and what it shows, and
//     is not listed twice; a claim listed under two candidates is noted;
//   - the sketch and examples pass the proposal check (CheckProposal), each
//     example uses the sketch or is a near miss of it, and no two sketches
//     share a scheme id or give one template two texts.
func CheckCandidates(cr *CandidateRegister, reg *Register, src *Sources) []Finding {
	var out []Finding
	bad := func(f string, a ...any) {
		out = append(out, Finding{File: CandidatesFile, Error: true, Msg: fmt.Sprintf(f, a...)})
	}
	warn := func(f string, a ...any) {
		out = append(out, Finding{File: CandidatesFile, Msg: fmt.Sprintf(f, a...)})
	}
	decisions := map[string]*Decision{}
	if reg != nil {
		for i := range reg.Decisions {
			decisions[reg.Decisions[i].ID] = &reg.Decisions[i]
		}
	}
	families := map[string]bool{}
	for _, f := range src.Families {
		families[f.ID] = true
	}
	th := cr.Threshold
	if th.Sightings < 1 || th.Places < 1 {
		bad("threshold: give the sightings and places needed")
	}
	if reg != nil && th.Decision != "" && decisions[th.Decision] == nil {
		bad("threshold: no decision %s in %s", th.Decision, RegisterFile)
	}
	groups := map[string]bool{}
	members := map[string]int{}
	for _, g := range cr.Groups {
		if !groupID.MatchString(g.ID) {
			bad("group id %q: write lower case words joined by hyphens", g.ID)
		}
		if groups[g.ID] {
			bad("group %s: the id is used twice", g.ID)
		}
		groups[g.ID] = true
		if strings.TrimSpace(g.Name) == "" || strings.TrimSpace(g.About) == "" {
			bad("group %s: a name and what its members share are needed", g.ID)
		}
		for _, f := range g.Families {
			if !families[f] {
				bad("group %s: family %s is not in families.json", g.ID, f)
			}
		}
	}
	ids := map[string]bool{}
	for _, c := range cr.Candidates {
		if ids[c.ID] {
			bad("%s: the id is used twice", c.ID)
		}
		ids[c.ID] = true
	}
	refs := map[string][]string{}       // a sighting's ref: the candidates it is listed under
	sketchOf := map[string]string{}     // a sketch's scheme id: its candidate
	templates := map[string][2]string{} // a sketch's template: its text and candidate
	for _, c := range cr.Candidates {
		where := c.ID
		if !candidateID.MatchString(c.ID) {
			bad("candidate id %q: write C and three digits", c.ID)
		}
		if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.What) == "" {
			bad("%s: a name and what it is are needed", where)
		}
		if strings.TrimSpace(c.Compares) == "" || strings.TrimSpace(c.Concludes) == "" {
			bad("%s: say what it compares and what it concludes, for its group's table", where)
		}
		if !in(CandidateKinds, c.Kind) {
			bad("%s: kind %q is not one of %s", where, c.Kind, strings.Join(CandidateKinds, ", "))
		}
		if !groups[c.Group] {
			bad("%s: no group %q", where, c.Group)
		}
		members[c.Group]++
		if !in(Confidences, c.Confidence) {
			bad("%s: confidence %q is not one of %s", where, c.Confidence, strings.Join(Confidences, ", "))
		}
		if len(c.Found) == 0 {
			bad("%s: say who found it (found)", where)
		}
		for _, x := range c.Closest {
			if strings.TrimSpace(x.Item) == "" || strings.TrimSpace(x.Why) == "" {
				bad("%s: a closest item needs the item and why it does not cover the candidate", where)
			}
		}
		for _, r := range c.Related {
			if !ids[r.Candidate] || r.Candidate == c.ID {
				bad("%s: related %q is not another candidate", where, r.Candidate)
			}
			if strings.TrimSpace(r.How) == "" {
				bad("%s: say how %s is related", where, r.Candidate)
			}
		}
		// The decision.
		d := decisions[c.Decision]
		switch {
		case c.Decision == "":
			if c.Outcome != nil {
				bad("%s: an outcome needs the decision that took it", where)
			} else if reg != nil && c.Met(th) {
				warn("%s meets the threshold: give it a decision in %s", where, RegisterFile)
			}
		case reg != nil && d == nil:
			bad("%s: no decision %s in %s", where, c.Decision, RegisterFile)
		case reg != nil && c.Outcome == nil && d.Status != "open":
			warn("%s: decision %s is taken, but no outcome is recorded here", where, c.Decision)
		case reg != nil && c.Outcome != nil && d.Status == "open":
			warn("%s: an outcome is recorded, but decision %s is still open", where, c.Decision)
		}
		if o := c.Outcome; o != nil {
			if !in(Results, o.Result) {
				bad("%s: outcome %q is not one of %s", where, o.Result, strings.Join(Results, ", "))
			}
			if o.Date == "" || o.By == "" || o.Recorded == "" {
				bad("%s: an outcome needs its date, by and where it is recorded", where)
			}
			if (o.Result == "merge" || o.Result == "domain") && o.Into == "" {
				bad("%s: a %s outcome names the item (into)", where, o.Result)
			}
		}
		// The sightings.
		seen := map[string]bool{}
		for _, s := range c.Sightings {
			if !in(SightingSources, s.Source) {
				bad("%s: sighting %s: source %q is not one of %s", where, s.Ref, s.Source, strings.Join(SightingSources, ", "))
			}
			if s.Source == "kialo" && !kialoRef.MatchString(s.Ref) {
				bad("%s: sighting %q: a Kialo ref is discussion.claim", where, s.Ref)
			}
			if s.place() == "" {
				bad("%s: sighting %s: say where it was seen (place)", where, s.Ref)
			}
			if !in(Fits, s.Fit) {
				bad("%s: sighting %s: fit %q is not one of %s", where, s.Ref, s.Fit, strings.Join(Fits, ", "))
			}
			if strings.TrimSpace(s.Note) == "" || s.Found == "" {
				bad("%s: sighting %s: a note and who found it are needed", where, s.Ref)
			}
			if seen[s.Ref] {
				bad("%s: sighting %s is listed twice", where, s.Ref)
			}
			seen[s.Ref] = true
			refs[s.Ref] = append(refs[s.Ref], c.ID)
		}
		// The sketch and the examples.
		for _, f := range CheckProposal(c.asProposal(), src) {
			f.File = CandidatesFile
			f.Msg = where + ": " + f.Msg
			out = append(out, f)
		}
		if s := c.Sketch; s != nil {
			for _, sc := range s.Schemes {
				if o, ok := sketchOf[sc.ID]; ok {
					bad("%s: sketch %s is also %s's", where, sc.ID, o)
				}
				sketchOf[sc.ID] = c.ID
			}
			keys := make([]string, 0, len(s.Language))
			for k := range s.Language {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				text := strings.TrimSpace(s.Language[k])
				if o, ok := templates[k]; ok && o[0] != text {
					bad("%s: template %s differs from %s's", where, k, o[1])
				}
				templates[k] = [2]string{text, c.ID}
			}
		}
		for i, x := range c.Examples {
			if x.Use == "near-miss" {
				if !c.sketched(x.Of) {
					bad("%s: example %d (%s): a near miss of the candidate names one of its sketch's schemes (near_miss_of)", where, i+1, x.Title)
				}
			} else if !c.sketched(x.Scheme) {
				bad("%s: example %d (%s): an example of use is coded with the candidate's sketch", where, i+1, x.Title)
			}
		}
	}
	for _, g := range cr.Groups {
		if members[g.ID] == 0 {
			warn("group %s has no candidates", g.ID)
		}
	}
	keys := make([]string, 0, len(refs))
	for k, v := range refs {
		if len(v) > 1 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, Finding{File: CandidatesFile, Msg: fmt.Sprintf("note: %s is listed under %s: one pattern, or a claim showing two?", k, strings.Join(refs[k], " and "))})
	}
	return out
}

// candidateTexts are the registry's texts of ours, for the copy check.
func candidateTexts(cr *CandidateRegister) []struct{ where, text string } {
	var out []struct{ where, text string }
	add := func(where, text string) { out = append(out, struct{ where, text string }{where, text}) }
	for _, g := range cr.Groups {
		add("group "+g.ID, g.About)
	}
	for _, c := range cr.Candidates {
		add(c.ID+" what", c.What)
		for _, x := range c.Closest {
			add(c.ID+" closest "+x.Item, x.Why)
		}
		for _, s := range c.Sightings {
			add(c.ID+" sighting "+s.Ref, s.Note)
		}
		for i, x := range c.Examples {
			add(fmt.Sprintf("%s example %d", c.ID, i+1), x.Text)
			add(fmt.Sprintf("%s example %d note", c.ID, i+1), x.Note)
			for _, a := range x.Attacks {
				add(fmt.Sprintf("%s example %d attack", c.ID, i+1), a.Text)
			}
		}
		for _, l := range [][]string{c.Engine, c.Needs, c.Open} {
			for _, t := range l {
				add(c.ID, t)
			}
		}
	}
	return out
}

// CandidatesMarkdown renders the registry as a readable view
// (CandidatesView): an index, each group with a table that compares its
// members, then each candidate in full, its sketch and examples rendered.
func CandidatesMarkdown(cr *CandidateRegister, reg *Register, src *Sources) []byte {
	var b bytes.Buffer
	th := cr.Threshold
	famName := map[string]string{}
	for _, f := range src.Families {
		famName[f.ID] = f.Name
	}
	status := map[string]string{}
	if reg != nil {
		for _, d := range reg.Decisions {
			status[d.ID] = d.Status
		}
	}
	stages := map[string]int{}
	for i := range cr.Candidates {
		stages[cr.Candidates[i].Stage(th)]++
	}
	b.WriteString("# Candidates\n\n")
	b.WriteString("Generated by `wag-cli --catalogue-gen` from [`candidates.yml`](candidates.yml), the candidates registry ([`process.md`](process.md) §8). Edit that file, not this one.\n\n")
	fmt.Fprintf(&b, "**%d candidates in %d groups:** %d ready, %d drafting, %d collecting, %d decided.\n\n",
		len(cr.Candidates), len(cr.Groups), stages[StageReady], stages[StageDrafting], stages[StageCollecting], stages[StageDecided])
	if cr.Note != "" {
		b.WriteString(strings.TrimSpace(cr.Note) + "\n\n")
	}
	fmt.Fprintf(&b, "**The threshold.** A candidate goes to a person once it has at least %d verified sightings in at least %d discussions or games (%s decides whether that is right). A sighting is verified when a second reader has checked the text against its note; only verified sightings count.\n\n", th.Sightings, th.Places, orDash(th.Decision))
	b.WriteString("**The stages.**\n")
	b.WriteString("- **collecting:** below the threshold.\n")
	b.WriteString("- **drafting:** over the threshold, but the minimum is not complete.\n")
	b.WriteString("- **ready:** the minimum is complete: the threshold, the closest items and why each fails, and, for a scheme, question or exception, a sketch that passes the proposal check, three examples of its use and a near miss. It goes to a person through its decision.\n")
	b.WriteString("- **decided:** reject, merge, domain version or add. An addition becomes a proposal (§5).\n\n")
	b.WriteString("## Index\n\n| Group | Id | Candidate | Kind | Seen | Stage | Decision |\n|---|---|---|---|---|---|---|\n")
	for _, g := range cr.Groups {
		for i := range cr.Candidates {
			c := &cr.Candidates[i]
			if c.Group != g.ID {
				continue
			}
			fmt.Fprintf(&b, "| %s | [%s](#%s) | %s | %s | %s | %s | %s |\n", cell(g.Name), c.ID, anchor(c), cell(c.Name), c.Kind,
				seenCell(c, th), c.Stage(th), decisionCell(c.Decision, status))
		}
	}
	b.WriteString("\n")
	for _, g := range cr.Groups {
		fmt.Fprintf(&b, "## %s\n\n", g.Name)
		var fams []string
		for _, f := range g.Families {
			fams = append(fams, fmt.Sprintf("%s (%s)", f, famName[f]))
		}
		if len(fams) == 0 {
			fams = []string{"across families"}
		}
		fmt.Fprintf(&b, "*Families:* %s.\n\n%s\n\n", strings.Join(fams, "; "), strings.Join(strings.Fields(g.About), " "))
		b.WriteString("| Id | Candidate | Kind | Compares | Concludes | Closest | Seen | Stage |\n|---|---|---|---|---|---|---|---|\n")
		for i := range cr.Candidates {
			c := &cr.Candidates[i]
			if c.Group != g.ID {
				continue
			}
			var closest []string
			for _, x := range c.Closest {
				closest = append(closest, x.Item)
			}
			fmt.Fprintf(&b, "| [%s](#%s) | %s | %s | %s | %s | %s | %s | %s |\n", c.ID, anchor(c), cell(c.Name), c.Kind, cell(c.Compares),
				cell(c.Concludes), cell(strings.Join(closest, "; ")), seenCell(c, th), c.Stage(th))
		}
		b.WriteString("\n")
		for i := range cr.Candidates {
			c := &cr.Candidates[i]
			if c.Group == g.ID {
				writeCandidate(&b, c, th, status, src)
			}
		}
	}
	return b.Bytes()
}

func writeCandidate(b *bytes.Buffer, c *Candidate, th Threshold, status map[string]string, src *Sources) {
	fmt.Fprintf(b, "### %s. %s\n\n", c.ID, c.Name)
	stage := c.Stage(th)
	if m := c.Missing(th); stage == StageDrafting || stage == StageCollecting {
		stage += "; missing: " + strings.Join(m, ", ")
	}
	fmt.Fprintf(b, "**Kind:** %s. **Decision:** %s. **Stage:** %s. **Found:** %s. **Finder's confidence:** %s.\n\n",
		c.Kind, decisionCell(c.Decision, status), stage, strings.Join(c.Found, "; "), c.Confidence)
	if o := c.Outcome; o != nil {
		into := ""
		if o.Into != "" {
			into = " (" + o.Into + ")"
		}
		fmt.Fprintf(b, "**Outcome:** %s%s, %s, by %s; recorded in %s.\n\n", o.Result, into, o.Date, o.By, o.Recorded)
	}
	fmt.Fprintf(b, "%s\n\n", strings.Join(strings.Fields(c.What), " "))
	if len(c.Closest) > 0 {
		b.WriteString("**Closest, and why each falls short:**\n")
		for _, x := range c.Closest {
			fmt.Fprintf(b, "- %s: %s\n", x.Item, strings.Join(strings.Fields(x.Why), " "))
		}
		b.WriteString("\n")
	}
	if len(c.Related) > 0 {
		b.WriteString("**Related candidates:**\n")
		for _, r := range c.Related {
			fmt.Fprintf(b, "- %s: %s\n", r.Candidate, strings.Join(strings.Fields(r.How), " "))
		}
		b.WriteString("\n")
	}
	if len(c.Sources) > 0 {
		b.WriteString("**Sources:**\n")
		for _, s := range c.Sources {
			fmt.Fprintf(b, "- %s\n", strings.Join(strings.Fields(s), " "))
		}
		b.WriteString("\n")
	}
	if c.Exempt != "" {
		fmt.Fprintf(b, "**Exempt from the threshold:** %s\n\n", strings.Join(strings.Fields(c.Exempt), " "))
	}
	if len(c.Sightings) > 0 {
		all, v, p := c.Seen()
		met := "not met"
		if c.Met(th) {
			met = "met"
		}
		fmt.Fprintf(b, "**Sightings:** %d, %d verified, in %d places; the threshold is %s.\n\n", all, v, p, met)
		b.WriteString("| Where | Ref | Fit | What it shows | Found | Verified |\n|---|---|---|---|---|---|\n")
		for _, s := range c.Sightings {
			ref := s.Ref
			if s.Archived {
				ref += " (archived)"
			}
			fmt.Fprintf(b, "| %s %s | %s | %s | %s | %s | %s |\n", s.Source, s.place(), ref, s.Fit, cell(s.Note), cell(s.Found), orDash(cell(s.Verified)))
		}
		b.WriteString("\n")
	}
	if s := c.Sketch; s != nil {
		b.WriteString("**The sketch** (checked as a proposal's candidates are; not in `walton.yml`):\n\n")
		if s.Note != "" {
			b.WriteString(strings.Join(strings.Fields(s.Note), " ") + "\n\n")
		}
		var hb bytes.Buffer
		for _, sc := range s.Schemes {
			writeScheme(&hb, src, sc, s.Questions[sc.ID], "A sketch: not in `walton.yml`.")
		}
		writeTemplates(&hb, "**Its templates:**", s.Language)
		// One level down: the candidate is ###.
		b.WriteString(strings.ReplaceAll("\n"+hb.String(), "\n### ", "\n#### ")[1:])
	}
	if len(c.Examples) > 0 {
		p := c.asProposal()
		schemes, lang := proposalSchemes(p, src)
		b.WriteString("**Examples, rendered:**\n\n")
		for i, x := range c.Examples {
			writeExample(b, "####", i+1, x, schemes[x.Scheme], lang, p.questions(x.Scheme), p.candidate(x.Scheme))
		}
	}
	for _, l := range []struct {
		title string
		items []string
	}{{"What WAG's engine showed", c.Engine}, {"What implementing it needs", c.Needs}, {"Open questions", c.Open}} {
		if len(l.items) == 0 {
			continue
		}
		fmt.Fprintf(b, "**%s:**\n", l.title)
		for _, t := range l.items {
			fmt.Fprintf(b, "- %s\n", strings.Join(strings.Fields(t), " "))
		}
		b.WriteString("\n")
	}
}

// anchor is the candidate's heading anchor in the view.
func anchor(c *Candidate) string {
	var b strings.Builder
	for _, r := range strings.ToLower(c.ID + " " + c.Name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}

func seenCell(c *Candidate, th Threshold) string {
	all, v, p := c.Seen()
	if c.Exempt != "" && all == 0 {
		return "exempt"
	}
	if all == 0 {
		return "–"
	}
	s := fmt.Sprintf("%d verified in %d", v, p)
	if v < all {
		s += fmt.Sprintf(" (%d sightings)", all)
	}
	if c.Met(th) {
		s += "; met"
	}
	return s
}

func decisionCell(id string, status map[string]string) string {
	if id == "" {
		return "–"
	}
	if s := status[id]; s != "" {
		return id + " (" + s + ")"
	}
	return id
}
