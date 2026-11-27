package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/wag/internal/evidence"
	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/manifest"
)

// A game's record set (docs/e1-evidence.md, choice 2): the file
// NAME/records.json of the server's evidence folder, with its payloads under
// NAME/. A request names it (NewGameRequest.Evidence); the trace's start and
// the export record its name and the SHA-256 of its bytes (RecordSet).

// RecordSet names a game's record set: its name in the evidence folder and
// the SHA-256 of its records.json as the game took it.
type RecordSet struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

// SetEvidence sets the server's evidence folder (wag-llm's --evidence); a
// data folder (SetStore) gives its evidence/ otherwise.
func (s *Server) SetEvidence(dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evidenceDir, s.evidenceSet = dir, true
}

// evidenceFolder is the server's evidence folder, "" if it has none. Caller
// holds s.mu, or no request is being served yet.
func (s *Server) evidenceFolder() string {
	if s.evidenceSet {
		return s.evidenceDir
	}
	if s.store == "" {
		return ""
	}
	return filepath.Join(s.store, "evidence")
}

// readRecordSet reads the record set name from the evidence folder dir: its
// records (evidence.DecodeRecordSet), the SHA-256 of its records.json and
// the folder of its payloads. It refuses a name that is not a plain name,
// and an evidence folder, a NAME folder or a records.json that is a link.
func readRecordSet(dir, name string) (*evidence.Bundle, string, string, error) {
	if dir == "" {
		return nil, "", "", fmt.Errorf("this server has no evidence folder: wag-server's is its data folder's evidence/, and wag-llm takes one with --evidence DIR")
	}
	if name == "" || name == "." || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name {
		return nil, "", "", fmt.Errorf("the record set's name %q is not a plain name: no separator, no ..", name)
	}
	root := filepath.Join(dir, name)
	for _, p := range []string{dir, root} {
		fi, err := os.Lstat(p)
		switch {
		case err != nil:
			return nil, "", "", fmt.Errorf("the record set %s: %v", name, err)
		case fi.Mode()&os.ModeSymlink != 0:
			return nil, "", "", fmt.Errorf("the record set %s: %s is a link, which is not followed", name, p)
		case !fi.IsDir():
			return nil, "", "", fmt.Errorf("the record set %s: %s is not a folder", name, p)
		}
	}
	file := filepath.Join(root, "records.json")
	fi, err := os.Lstat(file)
	if err != nil {
		return nil, "", "", fmt.Errorf("the record set %s: %v", name, err)
	}
	if !fi.Mode().IsRegular() {
		return nil, "", "", fmt.Errorf("the record set %s: its records.json is not a file (a link is not followed)", name)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, "", "", fmt.Errorf("the record set %s: %v", name, err)
	}
	b, err := evidence.DecodeRecordSet(raw)
	if err != nil {
		return nil, "", "", fmt.Errorf("the record set %s: %v", name, err)
	}
	sum := sha256.Sum256(raw)
	return b, hex.EncodeToString(sum[:]), root, nil
}

// evidenceParticipant is a seat as a participant (D076): a person, or a
// model (through its API, or a chatbot a person relays); a rule bot is
// neither, and cannot consent.
func evidenceParticipant(seat Seat, id string) (evidence.Participant, error) {
	switch seat.Kind {
	case "human":
		return evidence.Participant{ID: id, Kind: "human", Label: seat.Name}, nil
	case "model", "claude", "llm":
		return evidence.Participant{ID: id, Kind: "model", Label: seat.Name}, nil
	}
	return evidence.Participant{}, fmt.Errorf("a game with a record set is played by people and models: %s is a %s seat, which cannot consent (D076)", seat.Name, seat.Kind)
}

// checkEvidenceGame tells why a game of this configuration and these seats
// takes no record set (docs/e1-evidence.md, choice 2): an inquiry of two
// seats, each a person or a model, not moderated.
func checkEvidenceGame(cfg game.GameConfig, seats []Seat) error {
	if !cfg.Cooperative {
		return fmt.Errorf("a record set is for an inquiry (mode inquiry), not a %s game", cfg.Mode)
	}
	if cfg.Moderated {
		return fmt.Errorf("a moderated game takes no record set: the moderated profile waits for Track D (D076)")
	}
	for _, seat := range seats {
		if _, err := evidenceParticipant(seat, ""); err != nil {
			return err
		}
	}
	return nil
}

// gameEvidence reads the record set name for the game gs, whose seats are
// seats, and checks it in full: its records, with the game's context and
// actors, and its payloads and selectors (evidence.Validate), naming the
// first error. Caller holds s.mu.
func (s *Server) gameEvidence(name string, gs *game.GameState, seats []Seat) (*game.Evidence, error) {
	set, sha, root, err := readRecordSet(s.evidenceFolder(), name)
	if err != nil {
		return nil, err
	}
	var players []evidence.Participant
	for i, seat := range seats {
		p, err := evidenceParticipant(seat, gs.Players[i].ID)
		if err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	ev, err := game.NewEvidence(name, sha, set, players, s.pins.Engine, s.pins.Catalogue, s.pins.NASCQ)
	if err != nil {
		return nil, fmt.Errorf("the record set %s: %v", name, err)
	}
	if r := evidence.Validate(&ev.Bundle, root); !r.Valid {
		return nil, fmt.Errorf("the record set %s: %s", name, evidence.FirstIssue(r.Errors))
	}
	return ev, nil
}

// recordSet names the game's record set, nil for a game without one.
func (ss *Session) recordSet() *RecordSet {
	if ev := ss.GS.Evidence; ev != nil {
		return &RecordSet{Name: ev.Name, SHA256: ev.SHA256}
	}
	return nil
}

// loadEvidence reads again the record set of an exported game, for a load
// or a resume (docs/e1-evidence.md, choice 13): from this server's evidence
// folder, checked in full (its payloads included), its records.json's
// SHA-256 the export's, and the game's pins this server's. A game with a
// record set is not opened read-only: replaying its acts needs the record
// set and the pins it was played with. Caller holds s.mu.
func (s *Server) loadEvidence(e *Export, req *NewGameRequest, gs *game.GameState, inspect bool) (*game.Evidence, error) {
	switch {
	case inspect:
		return nil, fmt.Errorf("a game with a record set is not opened read-only: replaying its acts needs its record set and the pins it was played with; its export stays readable as a file")
	case e.RecordSet == nil || e.RecordSet.Name != req.Evidence:
		return nil, fmt.Errorf("the export's record set and its request's do not agree (%v, %q)", e.RecordSet, req.Evidence)
	case e.Pins == nil || *e.Pins != s.pins:
		return nil, fmt.Errorf("the game was played with other pins than this server's (the game's %s; this server's %s): a game with a record set is replayed with the engine, the scheme file and the NASCQ data it was played with", pinsText(e.Pins), pinsText(&s.pins))
	}
	ev, err := s.gameEvidence(req.Evidence, gs, e.Seats)
	if err != nil {
		return nil, err
	}
	if ev.SHA256 != e.RecordSet.SHA256 {
		return nil, fmt.Errorf("the record set %s has changed since the game: its records.json's SHA-256 is %s, the game's %s", req.Evidence, ev.SHA256, e.RecordSet.SHA256)
	}
	return ev, nil
}

// pinsText says what the pins are, by their first 12 digits.
func pinsText(p *manifest.Pins) string {
	if p == nil {
		return "none recorded"
	}
	cut := func(s string) string {
		if len(s) > 12 {
			return s[:12]
		}
		return s
	}
	return fmt.Sprintf("engine %s, scheme file %s, NASCQ %s", cut(p.Engine), cut(p.Catalogue), cut(p.NASCQ))
}

// Pins are the server's inputs by content (manifest): its engine, scheme
// file and NASCQ data, which a game with a record set is replayed with.
func (s *Server) Pins() manifest.Pins {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pins
}
