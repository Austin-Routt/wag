package game

import "github.com/carneades/carneades-4/src/engine/caes"

// SchemeID preserves the exact game scheme played. Historical
// graphs without that metadata can recover a game ID only when the engine
// identity has one unambiguous mapping in the loaded profile.
func (gs *GameState) SchemeID(arg *caes.Argument) string {
	if arg == nil {
		return ""
	}
	if id, ok := arg.Metadata["nascq_scheme_id"].(string); ok && id != "" {
		return id
	}
	if arg.Scheme == nil {
		return ""
	}
	id := ""
	for sid, scheme := range gs.NASCQ.Schemes {
		if scheme.CarneadesID == arg.Scheme.Id {
			if id != "" {
				return arg.Scheme.Id // ambiguous legacy identity: do not guess
			}
			id = sid
		}
	}
	if id != "" {
		return id
	}
	return arg.Scheme.Id
}

// CurrentSchemeID identifies an argument only if its recorded game identity
// and engine scheme agree with the loaded profile. Historical identities can
// be displayed without silently acquiring a different profile's CQs/weights.
// This is not a cross-version migration: the episode still needs a pinned
// profile to establish that unchanged IDs have unchanged meanings.
func (gs *GameState) CurrentSchemeID(arg *caes.Argument) string {
	if arg == nil || arg.Scheme == nil {
		return ""
	}
	id := gs.SchemeID(arg)
	s := gs.NASCQ.GetScheme(id)
	if s == nil || s.CarneadesID != arg.Scheme.Id {
		return ""
	}
	return id
}
