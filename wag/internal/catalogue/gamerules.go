package catalogue

// The rules of the game's questions that step C5.2 adds to the generator
// (docs/c5-game.md, C5.2, step 4): one locus for each question id, and one
// formula for an id that backup edges back.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// askedAt is where a scheme asks a question: the ask edge's locus and its
// pattern (with the scheme's variables in braces).
type askedAt struct {
	Scheme, Locus, Pattern string
}

// questionLociFaults reports the questions that break the rules of a
// question's locus: (a) a question id that backup edges back is asked at the
// same formula and kind of locus by every scheme that asks it (decision
// D095 (1)), so that an answer scheme is never offered where it cannot
// conclude the target; (b) every question id is asked at one kind of locus,
// so that its record's caes_mapping, derived from that locus, is the effect
// of every edge (D009). asked maps a question id to where each scheme asks
// it; backed names the ids that backup edges back.
func questionLociFaults(asked map[string][]askedAt, backed map[string]bool) []string {
	var faults []string
	ids := make([]string, 0, len(asked))
	for id := range asked {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		edges := asked[id]
		loci, formulas := map[string][]string{}, map[string][]string{}
		for _, e := range edges {
			loci[e.Locus] = append(loci[e.Locus], e.Scheme)
			formulas[e.Pattern] = append(formulas[e.Pattern], e.Scheme)
		}
		if len(loci) > 1 {
			faults = append(faults, fmt.Sprintf("%s is asked at %s: a question is asked at one kind of locus, so that its effect is the same on every edge (split it, one id for each)",
				id, describe(loci)))
		}
		if backed[id] && len(formulas) > 1 {
			faults = append(faults, fmt.Sprintf("%s is backed by answer schemes and asked at %s: a backed question is asked at one formula wherever it is asked (D095 (1))",
				id, describe(formulas)))
		}
	}
	return faults
}

// describe lists the values of a key (a locus or a formula) with the
// schemes that have each.
func describe(by map[string][]string) string {
	keys := make([]string, 0, len(by))
	for k := range by {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		name := k
		if name == "" {
			name = "no formula"
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", name, strings.Join(by[k], ", ")))
	}
	return strings.Join(parts, " and ")
}

// answerTypes is the caes_mapping a question's locus gives it: what a
// positive answer does, what a negative one does, and what it is about.
func answerTypes(locus string) (positive, negative, target string) {
	switch locus {
	case "premise":
		return "support", "undermining", "premise"
	case "conclusion":
		return "support", "rebuttal", "conclusion"
	default: // an exception, or the inference
		return "support", "undercutter", "argument"
	}
}

// placeholder is a {Var} placeholder of a question's text.
var placeholder = regexp.MustCompile(`\{([^{}]*)\}`)

// textFault reports why a game question's text is not its entry's
// paraphrase with the scheme's variables in braces (decision D009: the
// game's texts are the catalogue's paraphrases), or "". The braces say
// which capitals are variables, so a sentence that begins with the article
// "A" stays as written.
func textFault(text, paraphrase string, vars map[string]bool) string {
	for _, m := range placeholder.FindAllStringSubmatch(text, -1) {
		if !vars[m[1]] {
			return fmt.Sprintf("{%s} is not a variable of the scheme", m[1])
		}
	}
	if plain := placeholder.ReplaceAllString(text, "$1"); plain != paraphrase {
		return fmt.Sprintf("without its braces it is %q, not the paraphrase %q", plain, paraphrase)
	}
	return ""
}
