package game

import (
	"encoding/json"
	"fmt"
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

// LoadContexts reads the context presets.
func LoadContexts(b []byte) ([]Context, error) {
	var f struct {
		Contexts []Context `json:"contexts"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("contexts: %w", err)
	}
	for _, c := range f.Contexts {
		for id, w := range c.Weights {
			if w <= 0 || w > 1 {
				return nil, fmt.Errorf("contexts: %s weighs %s %v (want 0 < w <= 1)", c.ID, id, w)
			}
		}
	}
	return f.Contexts, nil
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
