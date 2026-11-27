package game

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
)

// Seeded is implemented by agents whose choices are driven by a seeded random
// source. RunGame records each such agent's seed in the GameResult, so that a
// game log identifies everything needed to replay it.
type Seeded interface {
	Seed() int64
}

// InputHashes returns the SHA-256 of the Carneades YAML and of the three NASCQ
// files (schemes, critical questions, edges, hashed in that order with a
// separator), the inputs that together with the seeds determine a game.
func InputHashes(yamlBytes, schemesJSON, cqsJSON, edgesJSON []byte) (yamlSHA, nascqSHA string) {
	y := sha256.Sum256(yamlBytes)
	h := sha256.New()
	for _, b := range [][]byte{schemesJSON, cqsJSON, edgesJSON} {
		h.Write(b)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(y[:]), hex.EncodeToString(h.Sum(nil))
}

// MoveDigest is a SHA-256 over a canonical rendering of the move history:
// every move's player, action (with bindings in sorted order, and its
// declarations, comparison, evidence act and cited bindings when it has
// them), an evidence act's mark (without its time), created argument IDs,
// weight, score delta and the statement labels after it.
// Equal digests mean equality only for this deliberately partial rendering. Wall-clock
// metadata (created_at) is deliberately not part of it.
func (gs *GameState) MoveDigest() string {
	var b strings.Builder
	for _, m := range gs.MoveHistory {
		a := m.Action
		fmt.Fprintf(&b, "%d|%s|%s|%s|%s|%s|%s|%s|%s|%s|%t|", m.TurnNumber, m.PlayerID, a.Type, a.AttackType, a.SchemeID,
			a.AnswersCQID, a.TargetArgID, a.TargetPremiseID, strings.Join(m.ArgIDs, ","), strings.Join(m.Presumptions, ","), m.Abstained)
		for _, k := range carneades.SortedKeys(a.VariableBindings) {
			fmt.Fprintf(&b, "%s=%s,", k, a.VariableBindings[k])
		}
		// A move's declarations and comparison (D094 (b), D089), written
		// only when it has them, so that a game without them keeps the
		// digest it had before they were rendered (step C5.2's check of
		// step 7, W-6).
		for _, d := range a.Declare {
			fmt.Fprintf(&b, "|declare=%s:%s:%s", d.ID, d.Actor, strings.Join(d.Aliases, ";"))
		}
		if a.Comparison != "" {
			fmt.Fprintf(&b, "|comparison=%s", a.Comparison)
		}
		// An evidence act's parts and its mark, and the bindings a move
		// cites (docs/e1-evidence.md), written only when present, as the
		// declarations are; the mark's time is wall-clock metadata.
		if o := a.Offer; o != nil {
			fmt.Fprintf(&b, "|offer=%q:%q:%q:%q:%q:%q:%q", o.ClaimID, o.ArgumentID, o.Kind, o.Statement, o.Role, o.Interpretation, o.Warrant)
		}
		if a.BindingID != "" || a.Revision != "" {
			fmt.Fprintf(&b, "|binding=%q:%q", a.BindingID, a.Revision)
		}
		if len(a.EvidenceRefs) > 0 {
			fmt.Fprintf(&b, "|refs=%q", a.EvidenceRefs)
		}
		if k := m.Evidence; k != nil {
			fmt.Fprintf(&b, "|act=%s:%s:%s:%s", k.Act, k.BindingID, k.BindingRevision, k.ArgumentRevision)
		}
		fmt.Fprintf(&b, "|%.6f|%.6f|", m.Weight, m.ScoreDelta)
		keys := make([]string, 0, len(m.LabelsAfter))
		for k := range m.LabelsAfter {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "%s:%d,", k, m.LabelsAfter[k])
		}
		b.WriteString("\n")
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
