package server

// access.go — who may do what in a game.
//
// The critical review of 2026-11-23 showed a request from another computer,
// with no credential, deciding a statement as the moderator. Now:
//
//   - A request from this computer (a loopback address, and a Host that names
//     this computer, as the workbench requires) is the owner's: the person who
//     runs the server may do anything in any game, as before.
//   - A request from another computer acts in a game only with a key its owner
//     shared, in the header X-WAG-Key. The moderator's key moderates and
//     decides; a seat's key moves for that seat, on its turn. Either key reads
//     the game, previews moves, writes notes in its holder's name and lets a
//     model or rule bot take its turn.
//   - Another computer starts, loads and imports no game.
//
// The keys are random, made with the game, and never part of its state,
// export or traces. They tell the people in a game apart; they do not make a
// network private: over plain HTTP anyone who can watch the traffic sees them.

import (
	crand "crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// KeyHeader carries a game key in a request from another computer.
const KeyHeader = "X-WAG-Key"

// gameKeys are a game's keys: the moderator's, and one for each seat a
// person plays ("" for a seat the server plays).
type gameKeys struct {
	moderator string
	seats     []string
}

func newKeys(seats []Seat) gameKeys {
	k := gameKeys{moderator: crand.Text(), seats: make([]string, len(seats))}
	for i, seat := range seats {
		if seat.aiKind() == "" {
			k.seats[i] = crand.Text()
		}
	}
	return k
}

// role is who a request acts as in a game.
type role struct {
	owner     bool // from this computer
	moderator bool // with the moderator's key
	seat      int  // with this seat's key; -1 otherwise
}

func (r role) none() bool { return !r.owner && !r.moderator && r.seat < 0 }

// errNoKey answers another computer that has no key for the game, whether or
// not the game exists.
var errNoKey = errors.New("this game answers another computer only with a key its owner shared (" + KeyHeader + ")")

// fromThisComputer reports whether a request comes from this computer and
// names this computer as its host, so that a page of another site whose name
// resolves to this computer is not taken for the owner.
func fromThisComputer(r *http.Request) bool { return loopbackClient(r) && localHost(r) }

func loopbackClient(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	return err == nil && ip != nil && ip.IsLoopback()
}

func localHost(r *http.Request) bool {
	h := r.Host
	if hh, _, err := net.SplitHostPort(h); err == nil {
		h = hh
	}
	h = strings.Trim(h, "[]")
	ip := net.ParseIP(h)
	return h == "localhost" || ip != nil && ip.IsLoopback()
}

// roleIn returns who the request acts as in the game ss (nil if there is no
// such game).
func roleIn(r *http.Request, ss *Session) role {
	if fromThisComputer(r) {
		return role{owner: true, seat: -1}
	}
	k := r.Header.Get(KeyHeader)
	if ss == nil || k == "" {
		return role{seat: -1}
	}
	if sameKey(k, ss.keys.moderator) {
		return role{moderator: true, seat: -1}
	}
	for i, sk := range ss.keys.seats {
		if sameKey(k, sk) {
			return role{seat: i}
		}
	}
	return role{seat: -1}
}

func sameKey(given, key string) bool {
	return key != "" && subtle.ConstantTimeCompare([]byte(given), []byte(key)) == 1
}

// may returns why the role may not take the action in the game (nil if it
// may). An action it does not know is the owner's only.
func (who role) may(ss *Session, action string) error {
	if who.owner {
		return nil
	}
	if who.none() {
		return errNoKey
	}
	switch action {
	case "", "moves", "stmt", "prompt", "export", "traces", "role", "preview", "note", "ai":
		return nil
	case "move", "reply":
		active := ss.GS.ActivePlayerIdx
		if who.seat < 0 {
			return fmt.Errorf("the moderator's key moves for no seat")
		}
		if who.seat != active {
			return fmt.Errorf("it is %s's turn: this key moves for %s", ss.GS.ActivePlayer().ID, ss.GS.Players[who.seat].ID)
		}
		return nil
	case "moderate", "decide":
		if !who.moderator {
			return fmt.Errorf("only the moderator's key moderates or decides")
		}
		return nil
	}
	return fmt.Errorf("%q is for the computer that runs the server", action)
}

// name is the name a role's notes are written in ("" for the owner, who
// names themselves).
func (who role) name(ss *Session) string {
	switch {
	case who.moderator:
		return "moderator"
	case who.seat >= 0:
		return ss.GS.Players[who.seat].ID
	}
	return ""
}

// ownerOnly refuses a request from another computer.
func ownerOnly(w http.ResponseWriter, r *http.Request) bool {
	if fromThisComputer(r) {
		return true
	}
	httpError(w, fmt.Errorf("games are started, loaded and imported on the computer that runs the server"), http.StatusForbidden)
	return false
}

// RoleView is what /api/games/{id}/role says about a request.
type RoleView struct {
	Role string `json:"role"`           // owner, moderator or seat
	Seat string `json:"seat,omitempty"` // the seat, for a seat's key
}

func (who role) view(ss *Session) RoleView {
	switch {
	case who.owner:
		return RoleView{Role: "owner"}
	case who.moderator:
		return RoleView{Role: "moderator"}
	}
	return RoleView{Role: "seat", Seat: ss.GS.Players[who.seat].ID}
}

// KeysView is what /api/games/{id}/keys gives the owner to share.
type KeysView struct {
	Moderator string    `json:"moderator"`
	Seats     []SeatKey `json:"seats"`
}

// SeatKey is the key of a seat a person plays.
type SeatKey struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

func (ss *Session) keysView() KeysView {
	v := KeysView{Moderator: ss.keys.moderator, Seats: []SeatKey{}}
	for i, k := range ss.keys.seats {
		if k != "" {
			v.Seats = append(v.Seats, SeatKey{Name: ss.GS.Players[i].ID, Key: k})
		}
	}
	return v
}
