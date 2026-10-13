package game

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/yourorg/wag/internal/carneades"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// Context is a preset for a kind of discussion: the proof standard of the
// claim, whether sources are required, whether a moderator is expected, and
// the weights an audience in that context would give the schemes ("*": every
// scheme). See data/contexts.json; its weights are drafts.
type Context struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Standard        string             `json:"standard"`
	RequireEvidence bool               `json:"require_evidence"`
	Moderated       bool               `json:"moderated,omitempty"`
	Weights         map[string]float64 `json:"weights,omitempty"`
}

// LoadContexts reads context presets: a file {"contexts": [...]} as
// data/contexts.json, or a single preset {"id": ..., ...} as the browser
// saves one.
func LoadContexts(b []byte) ([]Context, error) {
	var f struct {
		Contexts []Context `json:"contexts"`
		Context
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("contexts: %w", err)
	}
	cs := f.Contexts
	if len(cs) == 0 && f.ID != "" {
		cs = []Context{f.Context}
	}
	if len(cs) == 0 {
		return nil, fmt.Errorf("contexts: no preset in the file")
	}
	for _, c := range cs {
		if c.ID == "" {
			return nil, fmt.Errorf("contexts: a preset has no id")
		}
		if err := CheckWeights(c.Weights); err != nil {
			return nil, fmt.Errorf("contexts: %s: %v", c.ID, err)
		}
		switch c.Standard {
		case "", "PE", "CCE", "BRD":
		default:
			return nil, fmt.Errorf("contexts: %s: unknown proof standard %q", c.ID, c.Standard)
		}
	}
	return cs, nil
}

// MergeContexts adds the presets of more to base: a preset with the ID of
// one in base replaces it, the others are added at the end.
func MergeContexts(base, more []Context) []Context {
	out := append([]Context{}, base...)
	for _, c := range more {
		replaced := false
		for i := range out {
			if out[i].ID == c.ID {
				out[i], replaced = c, true
			}
		}
		if !replaced {
			out = append(out, c)
		}
	}
	return out
}

// CheckWeights checks scheme weights: each in (0, 1].
func CheckWeights(w map[string]float64) error {
	for _, id := range carneades.SortedKeys(w) {
		if v := w[id]; v <= 0 || v > 1 {
			return fmt.Errorf("%s weighs %v (want 0 < w <= 1)", id, v)
		}
	}
	return nil
}

// UnknownSchemes returns the keys of weights that name no scheme of the
// game ("*" names every scheme).
func (gs *GameState) UnknownSchemes(w map[string]float64) []string {
	var out []string
	for id := range w {
		if _, ok := gs.NASCQ.Schemes[id]; !ok && id != "*" {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// ClaimUnder returns the label the claim would have if the arguments made
// with schemes were weighed as the weights say (scheme ID, or "*" for every
// scheme; a scheme in neither keeps the weight the data gives it): the
// verdict of another audience on the same dispute. Weights that a moderator
// set, and the weights of imported arguments (Kialo votes), are kept. The
// game is not changed. It returns "" if the game has no claim.
func (gs *GameState) ClaimUnder(w map[string]float64) string {
	if gs.Claim == "" {
		return ""
	}
	moderated := map[string]bool{}
	for _, m := range gs.MoveHistory {
		if m.Moderation != "" && len(m.ArgIDs) > 0 {
			moderated[m.ArgIDs[0]] = true
		}
	}
	labels := gs.Adapter.LabelsWeighted(func(arg *caes.Argument) (float64, bool) {
		if carneades.ArgumentRole(arg) != carneades.RoleMain || moderated[arg.Id] || arg.Metadata["kialo_location"] != nil || arg.Scheme == nil {
			return 0, false
		}
		id := gs.NASCQ.FindNASCQSchemeID(arg.Scheme.Id)
		if id == "" {
			return 0, false
		}
		if v, ok := w[id]; ok {
			return v, true
		}
		if v, ok := w["*"]; ok {
			return v, true
		}
		v, ok := gs.dataWeights[id]
		return v, ok
	})
	switch labels[gs.Claim] {
	case caes.In:
		return "in"
	case caes.Out:
		return "out"
	}
	return "undecided"
}

// applyWeights sets the schemes' base weights from a context's weights.
func (gs *GameState) applyWeights(w map[string]float64) {
	for id, sc := range gs.NASCQ.Schemes {
		if v, ok := w[id]; ok {
			sc.BaseWeight = v
		} else if v, ok := w["*"]; ok {
			sc.BaseWeight = v
		}
	}
}
