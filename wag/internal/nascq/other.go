package nascq

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// The user-defined scheme and its generic critical questions (EnableOther).
const (
	OtherSchemeID  = "AS:other"
	OtherCarneades = "other" // its scheme ID on Carneades arguments

	CQPremiseAcceptable = "CQ:premise_acceptable"
	CQRelevance         = "CQ:relevance"
	CQCounter           = "CQ:counter"

	// OtherMaxPremises is the number of premise variables (P1 ... P6) of
	// the user-defined scheme.
	OtherMaxPremises = 6
)

// EnableOther adds the user-defined scheme "Other": an argument whose
// premises and conclusion the player writes freely, for arguments that fit
// none of the official schemes (and, recorded, as material for mining new
// schemes). It is thesis-independent, so it can back the answer to any
// critical question, and three generic critical questions can be asked of
// it: whether a premise is acceptable (a premise, named by the asker),
// whether the premises are relevant and sufficient (the inference), and
// whether there is a stronger reason against the conclusion (the
// conclusion). Its weight, 0.5, is below every official scheme's.
//
// Calling it again does nothing.
func (n *NASCQ) EnableOther() {
	if _, ok := n.Schemes[OtherSchemeID]; ok {
		return
	}
	vars := map[string]VariableSpec{"C": {Type: "proposition", Description: "The conclusion, in your own words"}}
	var prem []string
	for i := 1; i <= OtherMaxPremises; i++ {
		v := "P" + strconv.Itoa(i)
		desc := fmt.Sprintf("Premise %d, in your own words (optional)", i)
		if i == 1 {
			desc = "Premise 1, in your own words"
		}
		vars[v] = VariableSpec{Type: "proposition", Description: desc}
		prem = append(prem, "{"+v+"}")
	}
	n.Schemes[OtherSchemeID] = &ASNode{
		ID:                   OtherSchemeID,
		CarneadesID:          OtherCarneades,
		Name:                 "Other (your own scheme)",
		SemanticType:         TypeTruth,
		ThesisIndependent:    true,
		BaseWeight:           0.5,
		Variables:            vars,
		PremiseTemplatesEN:   prem,
		ConclusionTemplateEN: "{C}",
		AnswerVariable:       "C",
		Free:                 true,
		DescriptionEN:        "An argument that fits none of the listed schemes: write its premises and conclusion yourself. It is recorded for finding new schemes.",
	}
	n.Questions[CQPremiseAcceptable] = premiseQuestion()
	for _, q := range []*CQNode{
		{ID: CQRelevance, Name: "Relevance and Sufficiency Question", SemanticType: TypeTruth,
			QuestionTemplateEN: "Are the premises relevant to the conclusion, and together sufficient for it?",
			CaesMapping:        CaesMapping{PositiveAnswerType: "support", NegativeAnswerType: "undercutter", Target: "argument"}},
		{ID: CQCounter, Name: "Counterargument Question", SemanticType: TypeTruth,
			QuestionTemplateEN: "Is there a stronger reason against the conclusion {C}?",
			CaesMapping:        CaesMapping{PositiveAnswerType: "support", NegativeAnswerType: "rebuttal", Target: "conclusion"}},
	} {
		n.Questions[q.ID] = q
	}
	n.AskEdges = append(n.AskEdges,
		Edge{CQ: CQRelevance, Scheme: OtherSchemeID, Locus: LocusArgument},
		Edge{CQ: CQCounter, Scheme: OtherSchemeID, Locus: LocusConclusion},
	)
	if !n.hasAsk(CQPremiseAcceptable, OtherSchemeID) {
		n.AskEdges = append(n.AskEdges, Edge{CQ: CQPremiseAcceptable, Scheme: OtherSchemeID, Locus: LocusPremise})
	}
	n.askIndex = make(map[string][]string)
	n.backupIndex = make(map[string][]string)
	n.buildIndexes()
}

func premiseQuestion() *CQNode {
	return &CQNode{ID: CQPremiseAcceptable, Name: "Premise Acceptability Question", SemanticType: TypeTruth,
		QuestionTemplateEN: "Is this premise acceptable: true, or supported by evidence?",
		CaesMapping:        CaesMapping{PositiveAnswerType: "support", NegativeAnswerType: "undermining", Target: "premise"}}
}

func (n *NASCQ) hasAsk(cqID, schemeID string) bool {
	for _, e := range n.AskEdges {
		if e.CQ == cqID && e.Scheme == schemeID {
			return true
		}
	}
	return false
}

// EnablePremiseQuestion lets any premise of any argument be questioned or
// attacked, not only those a scheme's own critical questions are about. In
// a dialogue a premise can always be challenged ("why should I accept
// that?"); Walton's critical questions point at the typical weak spots,
// which are often exceptions and the conclusion. Practical reasoning, for
// instance, has five critical questions and only one of them (possibility)
// is about one of its premises.
//
// The generic Premise Acceptability Question is asked of every scheme, about
// any premise that none of the scheme's own critical questions is about
// (EnumerateActions); the thesis-independent schemes back its answers.
// Calling it again does nothing.
func (n *NASCQ) EnablePremiseQuestion() {
	if _, ok := n.Questions[CQPremiseAcceptable]; !ok {
		n.Questions[CQPremiseAcceptable] = premiseQuestion()
	}
	added := false
	for _, sid := range sortedSchemeIDs(n.Schemes) {
		if !n.hasAsk(CQPremiseAcceptable, sid) {
			n.AskEdges = append(n.AskEdges, Edge{CQ: CQPremiseAcceptable, Scheme: sid, Locus: LocusPremise,
				Description: "generic: any premise no other critical question is about"})
			added = true
		}
	}
	if added {
		n.askIndex = make(map[string][]string)
		n.backupIndex = make(map[string][]string)
		n.buildIndexes()
	}
}

// CoveredPremises returns the premises of argID that a critical question of
// the scheme with a premise pattern is about.
func (n *NASCQ) CoveredPremises(graph GraphQuerier, argID, schemeID string) map[string]bool {
	out := map[string]bool{}
	for _, cqID := range n.askIndex[schemeID] {
		if l := n.LocusOf(cqID, schemeID); l.Kind == LocusPremise && l.Pattern != "" {
			if id := graph.PremiseMatching(argID, l.Pattern); id != "" {
				out[id] = true
			}
		}
	}
	return out
}

// FreePremises returns the premises of a move made with the user-defined
// scheme, in order (P1, P2, ...), without the empty ones.
func FreePremises(bindings map[string]string) []string {
	var out []string
	for i := 1; i <= OtherMaxPremises; i++ {
		if p := strings.TrimSpace(bindings["P"+strconv.Itoa(i)]); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Suggestion is a scheme Suggest ranks for a natural-language argument.
type Suggestion struct {
	SchemeID string   `json:"scheme_id"`
	Name     string   `json:"name"`
	Score    float64  `json:"score"`
	Matched  []string `json:"matched"` // the cues found in the text
}

// Suggest ranks the schemes by how well a natural-language argument matches
// their cues (CuesEN, then words of the name and templates), best first,
// and returns those that match at all. It is a transparent keyword
// heuristic, a placeholder for a trained or LLM-based scheme classifier:
// the player still chooses the scheme.
func (n *NASCQ) Suggest(text string) []Suggestion {
	t := " " + normalizeText(text) + " "
	var out []Suggestion
	for _, id := range sortedSchemeIDs(n.Schemes) {
		s := n.Schemes[id]
		if s.Free {
			continue
		}
		sg := Suggestion{SchemeID: id, Name: s.Name}
		for _, c := range s.CuesEN {
			c = normalizeText(c)
			if c != "" && strings.Contains(t, " "+c+" ") {
				sg.Score += 1 + 0.5*float64(strings.Count(c, " "))
				sg.Matched = append(sg.Matched, c)
			}
		}
		if sg.Score > 0 {
			out = append(out, sg)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

// normalizeText lowercases and keeps letters, digits and single spaces.
func normalizeText(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '\'', r > 127:
			b.WriteRune(r)
			space = false
		default:
			if !space {
				b.WriteByte(' ')
				space = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}
