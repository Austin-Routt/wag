package server

import (
	"encoding/json"

	"github.com/yourorg/wag/internal/evidence"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/nascq"
)

// The board's view of a game's evidence (docs/e1-evidence.md, choice 12).
// Every text in it is data, which the board shows as text.

// RefState is a binding an argument cites, and whether it is admitted for
// consideration now: one a move cited was admitted then, so one not
// admitted now had a consent withdrawn since.
type RefState struct {
	Binding  string `json:"binding"`
	Admitted bool   `json:"admitted"`
}

// EvidenceView is a game's evidence: its record set's records, by name and
// hash, its context's id and policy, and the bindings its players made.
type EvidenceView struct {
	Name            string                    `json:"name"`
	SHA256          string                    `json:"sha256"`
	Context         string                    `json:"context"`
	Policy          string                    `json:"policy"`
	Sources         []evidence.Source         `json:"sources"`
	Artifacts       []evidence.Artifact       `json:"artifacts"`
	Representations []evidence.Representation `json:"representations"`
	Claims          []evidence.Claim          `json:"claims"`
	Dependencies    []evidence.Dependency     `json:"dependencies"`
	Bindings        []BindingView             `json:"bindings"`
}

// BindingView is a binding with every part its revision covers that the
// view does not hold already (its target), its full revision, the player
// who offered it, the players whose consent stands, whether it is admitted,
// and whether a consent to it was withdrawn.
type BindingView struct {
	evidence.Binding
	OfferedBy string          `json:"offered_by"`
	Target    evidence.Target `json:"target"`
	Revision  string          `json:"revision"`
	Consents  []string        `json:"consents"`
	Admitted  bool            `json:"admitted"`
	Withdrawn bool            `json:"withdrawn,omitempty"`
}

// decodeAction reads a person's move. A game without a record set ignores
// the fields of an evidence act and of a move's citations, whatever their
// JSON types, as before E1, when they were unknown fields
// (docs/e1-evidence.md, E1.7, B3). Caller holds s.mu.
func (ss *Session) decodeAction(raw json.RawMessage) (nascq.ProposedAction, error) {
	var a nascq.ProposedAction
	if len(raw) == 0 || string(raw) == "null" {
		return a, nil
	}
	if ss.GS.Evidence == nil {
		var m map[string]json.RawMessage
		if json.Unmarshal(raw, &m) == nil {
			for _, k := range []string{"offer", "binding_id", "revision", "evidence_refs"} {
				delete(m, k)
			}
			raw, _ = json.Marshal(m)
		}
	}
	err := json.Unmarshal(raw, &a)
	return a, err
}

// evidenceView is the board's view of the game's evidence, nil without a
// record set.
func evidenceView(gs *game.GameState) *EvidenceView {
	ev := gs.Evidence
	if ev == nil {
		return nil
	}
	b := ev.Bundle
	v := &EvidenceView{Name: ev.Name, SHA256: ev.SHA256, Context: b.Context.ID, Policy: b.Context.Policy,
		Sources: b.Sources, Artifacts: b.Artifacts, Representations: b.Representations, Claims: b.Claims,
		Dependencies: b.Dependencies, Bindings: []BindingView{}}
	withdrawn := map[string]bool{}
	for _, e := range b.Events {
		if e.Type == "withdraw" {
			withdrawn[e.BindingID] = true
		}
	}
	players := map[string]string{}
	for _, p := range gs.Players {
		players[evidence.PlayerActor(p.ID)] = p.ID
	}
	for _, x := range b.Bindings {
		t, _ := ev.Target(x.TargetID)
		rev, _ := ev.Revision(x.ID)
		consents := ev.Consents(x.ID)
		if consents == nil {
			consents = []string{}
		}
		v.Bindings = append(v.Bindings, BindingView{Binding: x, OfferedBy: players[x.CreatedBy], Target: t, Revision: rev,
			Consents: consents, Admitted: ev.Admitted(x.ID), Withdrawn: withdrawn[x.ID]})
	}
	return v
}
