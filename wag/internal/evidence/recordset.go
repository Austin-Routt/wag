package evidence

import (
	"encoding/json"
	"fmt"
	"strings"
)

// A record set is what a game is given before play (docs/e1-evidence.md,
// choice 2): a wag-evidence/0.2 bundle of sources, artifacts,
// representations, claims and dependencies, captured by the owner's tools.
// The game writes the rest: the context, its own actors, and the targets,
// bindings and events its players' acts make. GameBundle makes the game's
// bundle from a record set.

// GamePrefix begins the ids of the game's records and actors; a record set
// uses none.
const GamePrefix = "game:"

// GameActor is the game, a tool, which writes the targets.
const GameActor = GamePrefix + "wag"

// GameContext is the id of the context a game writes, and GameProfile its
// profile.
const (
	GameContext = GamePrefix + "context"
	GameProfile = "wag-inquiry/1"
)

// PlayerActor is the actor id of the game's player id.
func PlayerActor(id string) string { return GamePrefix + "player:" + id }

// Participant is a player as the game declares it: its id in the game, its
// kind ("human" for a person, "model" for a model seat) and its label (the
// seat's name).
type Participant struct{ ID, Kind, Label string }

// gameKeys are the bundle's keys the game writes: a record set has none of
// them, even empty.
var gameKeys = []string{"context", "targets", "bindings", "assessments", "events"}

// IDLimit is the longest id a record set's record or actor may have.
const IDLimit = 128

// PlainID reports whether id is an identifier a record set may give a record
// or an actor: letters, digits, '.', '_', ':' and '-', beginning with a
// letter or a digit, at most IDLimit characters. A game's prompt writes ids
// as they are, so an id carries no line, space or quotation mark
// (docs/e1-evidence.md, E1.7, B2).
func PlainID(id string) bool {
	return plain(id, IDLimit, ".-_:")
}

// PlainName reports whether name may name a record set in an evidence
// folder: letters, digits, '.', '_' and '-', beginning with a letter or a
// digit, at most 64 characters.
func PlainName(name string) bool {
	return plain(name, 64, ".-_")
}

// plain reports whether s is 1 to limit ASCII letters, digits and the
// punctuation punct, beginning with a letter or a digit.
func plain(s string, limit int, punct string) bool {
	if s == "" || len(s) > limit {
		return false
	}
	for i, r := range s {
		alnum := r < 128 && ('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9')
		if !alnum && (i == 0 || !strings.ContainsRune(punct, r)) {
			return false
		}
	}
	return true
}

// DecodeRecordSet decodes a record set's JSON strictly (Decode: no unknown
// field, duplicate key or null) and refuses what a record set may not hold:
// a key the game writes, even empty; another version than FormatGame; an id,
// its actors' included, that begins with GamePrefix or is not a plain id
// (PlainID); and a record whose author is not one of its actors.
func DecodeRecordSet(raw []byte) (*Bundle, error) {
	b, err := Decode(raw)
	if err != nil {
		return nil, err
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	for _, k := range gameKeys {
		if _, ok := top[k]; ok {
			return nil, fmt.Errorf("a record set has no %q: the game writes its own", k)
		}
	}
	if b.Format != FormatGame {
		return nil, fmt.Errorf("a record set is %s, not %q", FormatGame, b.Format)
	}
	actors := map[string]bool{}
	for _, a := range b.Actors {
		if strings.HasPrefix(a.ID, GamePrefix) {
			return nil, fmt.Errorf("the actor %q begins with %q, which the game's actors use", a.ID, GamePrefix)
		}
		if !PlainID(a.ID) {
			return nil, fmt.Errorf("the actor id %q is not a plain id: letters, digits, '.', '_', ':' and '-', beginning with a letter or a digit, at most %d characters", a.ID, IDLimit)
		}
		actors[a.ID] = true
	}
	for _, r := range b.records() {
		if strings.HasPrefix(r.ID, GamePrefix) {
			return nil, fmt.Errorf("the record %q begins with %q, which the game's records use", r.ID, GamePrefix)
		}
		if !PlainID(r.ID) {
			return nil, fmt.Errorf("the record id %q is not a plain id: letters, digits, '.', '_', ':' and '-', beginning with a letter or a digit, at most %d characters", r.ID, IDLimit)
		}
		if !actors[r.CreatedBy] {
			return nil, fmt.Errorf("the record %s is by %q, who is not one of the record set's actors", r.ID, r.CreatedBy)
		}
	}
	return b, nil
}

// records are a record set's records: its sources, artifacts,
// representations, claims and dependencies.
func (b *Bundle) records() []Record {
	var out []Record
	for _, v := range b.Sources {
		out = append(out, v.Record)
	}
	for _, v := range b.Artifacts {
		out = append(out, v.Record)
	}
	for _, v := range b.Representations {
		out = append(out, v.Record)
	}
	for _, v := range b.Claims {
		out = append(out, v.Record)
	}
	for _, v := range b.Dependencies {
		out = append(out, v.Record)
	}
	return out
}

// GameBundle is the bundle of a game given the record set set (as
// DecodeRecordSet gives it), played by two players with the game's pins:
// the record set's records and actors, the game (GameActor, a tool) and
// the players as actors, and the context the game writes (GameContext,
// GameProfile, the two-party policy, the players as participants, the
// pins), with no target, binding, assessment or event yet. Its records are
// checked (ValidateRecords); the caller checks the payloads (Validate).
func GameBundle(set *Bundle, players []Participant, engine, catalogue, nascq string) (*Bundle, error) {
	if len(players) != 2 {
		return nil, fmt.Errorf("a game with a record set is two players, not %d", len(players))
	}
	b := *set
	b.Actors = append(append([]Actor{}, set.Actors...), Actor{ID: GameActor, Kind: "tool", Label: "WAG"})
	var participants []string
	for _, p := range players {
		if p.Kind != "human" && p.Kind != "model" {
			return nil, fmt.Errorf("the player %s is a %q: a participant is a person or a model (D076)", p.ID, p.Kind)
		}
		b.Actors = append(b.Actors, Actor{ID: PlayerActor(p.ID), Kind: p.Kind, Label: p.Label})
		participants = append(participants, PlayerActor(p.ID))
	}
	b.Context = Context{ID: GameContext, Profile: GameProfile, Policy: Policy, Participants: participants,
		EngineSHA256: engine, CatalogueSHA256: catalogue, NASCQSHA256: nascq}
	b.Targets, b.Bindings, b.Assessments, b.Events = []Target{}, []Binding{}, []Assessment{}, []Event{}
	if r := ValidateRecords(&b); !r.Valid {
		return nil, fmt.Errorf("the record set's records: %s", FirstIssue(r.Errors))
	}
	return &b, nil
}

// FirstIssue says what the first of issues is: its code, where, and the
// message.
func FirstIssue(issues []Issue) string {
	if len(issues) == 0 {
		return "no issue"
	}
	i := issues[0]
	return fmt.Sprintf("%s at %s: %s", i.Code, i.Location, i.Message)
}
