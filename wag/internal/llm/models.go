package llm

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ModelInfo is a model a provider offers.
type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
	Details     string `json:"details,omitempty"`
}

// ListModels lists the models the profile's provider offers to its key (the
// profile's model is not needed): Gemini's, Anthropic's, an
// OpenAI-compatible endpoint's (GET /models). Sorted by ID.
func ListModels(ctx context.Context, p Profile) ([]ModelInfo, error) {
	var out []ModelInfo
	switch p.Provider {
	case "gemini":
		list, err := GeminiModels(ctx, p)
		if err != nil {
			return nil, err
		}
		for _, m := range list {
			acts := fmt.Sprint(m["actions"])
			if !strings.Contains(acts, "generateContent") {
				continue // embeddings and the like
			}
			out = append(out, ModelInfo{ID: fmt.Sprint(m["name"]), DisplayName: fmt.Sprint(m["display_name"]),
				Details: fmt.Sprintf("input %v tokens, output %v; thinking %v", m["input_token_limit"], m["output_token_limit"], m["thinking"])})
		}
	case "anthropic":
		a, err := newAnthropic(p)
		if err != nil {
			return nil, err
		}
		it := a.client.Models.ListAutoPaging(ctx, anthropic.ModelListParams{})
		for it.Next() {
			m := it.Current()
			out = append(out, ModelInfo{ID: m.ID, DisplayName: m.DisplayName})
		}
		if err := it.Err(); err != nil {
			return nil, AsError("anthropic", err)
		}
	case "mock":
		out = []ModelInfo{{ID: "mock-1", DisplayName: "Mock (scripted, no network)"}}
	case "command":
		return nil, fmt.Errorf("a command provider has no model list")
	default:
		list, err := OpenAIModels(ctx, p)
		if err != nil {
			return nil, err
		}
		for _, m := range list {
			id := fmt.Sprint(m["id"])
			info := ModelInfo{ID: id}
			if o, ok := m["owned_by"].(string); ok && o != "" {
				info.Details = "owned by " + o
			}
			out = append(out, info)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
