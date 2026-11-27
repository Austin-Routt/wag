package game

// The game's profile (decision D094 (b), applying D089, D091 and D092; D063
// says why no agent may pass unchecked): who deliberates, the actions argued
// about with who performs each, and the situation the alternatives are
// compared in. A game made since step C5.2 has one; a game whose record has
// none (made before it) plays as it was made.

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
)

// Profile is a game's profile.
type Profile struct {
	// Agent is the actor who deliberates: the one a recommendation
	// should_be_performed(A) is addressed to (D063), and the P of the schemes
	// with an agent place.
	Agent string `json:"agent"`
	// Actors are the parties the game knows, the agent among them.
	Actors []ProfileActor `json:"actors"`
	// Actions are the actions argued about, each with its one actor and its
	// aliases (D091, D092). A move adds the actions it declares.
	Actions []ProfileAction `json:"actions,omitempty"`
	// Situation is the situation the alternatives are compared in (D089: the
	// current one, fixed for the game). It is shown and recorded, not checked
	// move by move.
	Situation string `json:"situation"`
	// Source says where the profile came from: "case:<id>" for a case
	// record, "" when it was asked or defaulted.
	Source string `json:"source,omitempty"`
}

// ProfileActor is a party of the game.
type ProfileActor struct {
	ID   string `json:"id"`
	Text string `json:"text,omitempty"`
}

// ProfileAction is an action argued about.
type ProfileAction struct {
	ID      string   `json:"id"`
	Actor   string   `json:"actor"`
	Aliases []string `json:"aliases,omitempty"`
}

// DefaultAgent is the agent of a profile no one gave.
const DefaultAgent = "decider"

// DefaultProfile is the profile a game has when none is given: an agent who
// decides, no action yet, the present situation. The board shows it.
func DefaultProfile() *Profile {
	return &Profile{
		Agent:     DefaultAgent,
		Actors:    []ProfileActor{{ID: DefaultAgent, Text: "the one who decides whether to take the actions argued about"}},
		Situation: "the present one",
	}
}

// caseRecord is the part of a case record (format wag-case/1,
// docs/annotation-records.md) a profile is made from.
type caseRecord struct {
	Format string `json:"format"`
	ID     string `json:"id"`
	Thesis struct {
		Action string `json:"action"`
	} `json:"thesis"`
	DefaultContext string `json:"default_context"`
	Contexts       []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"contexts"`
	Actors  []ProfileActor  `json:"actors"`
	Actions []ProfileAction `json:"actions"`
}

// ProfileFromCase makes a profile from a case record: its actors and
// actions; its agent, the thesis action's actor; its situation, the default
// context's text.
func ProfileFromCase(b []byte) (*Profile, error) {
	var c caseRecord
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("the case record: %v", err)
	}
	if c.Format != "wag-case/1" {
		return nil, fmt.Errorf("the case record's format is %q, not wag-case/1", c.Format)
	}
	p := &Profile{Actors: c.Actors, Actions: c.Actions, Source: "case:" + c.ID}
	for _, a := range c.Actions {
		if a.ID == c.Thesis.Action {
			p.Agent = a.Actor
		}
	}
	if p.Agent == "" {
		return nil, fmt.Errorf("the case record's thesis action %q is none of its actions", c.Thesis.Action)
	}
	for _, x := range c.Contexts {
		if x.ID == c.DefaultContext {
			p.Situation = x.Text
		}
	}
	if p.Situation == "" {
		return nil, fmt.Errorf("the case record's default context %q is none of its contexts", c.DefaultContext)
	}
	return p, p.Check()
}

// Check reports what makes a profile unusable: an agent that is no actor,
// an action without a known actor, an id or alias given twice.
func (p *Profile) Check() error {
	var errs []string
	actors := map[string]bool{}
	for _, a := range p.Actors {
		if strings.TrimSpace(a.ID) == "" {
			errs = append(errs, "an actor without an id")
		} else if actors[a.ID] {
			errs = append(errs, fmt.Sprintf("the actor %s twice", a.ID))
		}
		actors[a.ID] = true
	}
	if !actors[p.Agent] {
		errs = append(errs, fmt.Sprintf("the agent %q is none of the actors", p.Agent))
	}
	names := map[string]string{} // by the atom the board names it with
	ids := map[string]bool{}
	for _, a := range p.Actions {
		if ids[carneades.TextToAtom(a.ID)] {
			errs = append(errs, fmt.Sprintf("the action %s twice", a.ID))
		}
		ids[carneades.TextToAtom(a.ID)] = true
		if !actors[a.Actor] {
			errs = append(errs, fmt.Sprintf("the action %s: its actor %q is none of the actors", a.ID, a.Actor))
		}
		for _, n := range append([]string{a.ID}, a.Aliases...) {
			if strings.TrimSpace(n) == "" {
				errs = append(errs, fmt.Sprintf("the action %s has an empty name", a.ID))
				continue
			}
			if o, ok := names[carneades.TextToAtom(n)]; ok && o != a.ID {
				errs = append(errs, fmt.Sprintf("%s names both %s and %s", n, o, a.ID))
			}
			names[carneades.TextToAtom(n)] = a.ID
		}
	}
	if len(errs) > 0 {
		sort.Strings(errs)
		return fmt.Errorf("the profile: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Action returns the action a name (an id or an alias) names. Names are
// compared as the board writes them, as atoms (carneades.TextToAtom): "tax
// meat", "Tax-Meat" and tax_meat name one statement, so one action.
func (p *Profile) Action(name string) (*ProfileAction, bool) {
	atom := carneades.TextToAtom(name)
	for i := range p.Actions {
		a := &p.Actions[i]
		if carneades.TextToAtom(a.ID) == atom {
			return a, true
		}
		for _, al := range a.Aliases {
			if carneades.TextToAtom(al) == atom {
				return a, true
			}
		}
	}
	return nil, false
}

// Comparisons are the ways an alternative can be better (D089): it reaches
// the goal more surely, sooner, or at lower cost.
var Comparisons = []string{"surely", "sooner", "cheaper"}

// declared returns a copy of the profile with a move's declarations added,
// or why they cannot be. A declaration names its actor (D094 (b): "a
// journaled move declares an action with its actor"): an actor nobody named
// would let any agent pass (D063). A move declares only the actions it
// argues about (argued: the atoms of the values its action places take),
// in the move that first argues about each. A declaration of an action the
// profile already holds, with the same actor, declares nothing new but the
// aliases it adds; with another actor it is refused.
func (p *Profile) declared(decls []nascq.ActionDecl, argued map[string]bool) (*Profile, error) {
	q := p.clone()
	for _, d := range decls {
		d.ID, d.Actor = strings.TrimSpace(d.ID), strings.TrimSpace(d.Actor)
		if d.ID == "" {
			return nil, fmt.Errorf("P1: a declared action has no id")
		}
		if d.Actor == "" {
			return nil, fmt.Errorf("P1: the declaration of %s names its actor, one of %s (D094 (b))", d.ID, strings.Join(q.actorIDs(), ", "))
		}
		actor, ok := q.actorID(d.Actor)
		if !ok {
			return nil, fmt.Errorf("P1: the declaration of %s names the actor %s, none of %s", d.ID, d.Actor, strings.Join(q.actorIDs(), ", "))
		}
		names := append([]string{d.ID}, d.Aliases...)
		bound := false
		for _, n := range names {
			bound = bound || argued[carneades.TextToAtom(n)]
		}
		if !bound {
			return nil, fmt.Errorf("P1: the move declares %s but argues about no such action: a move declares the actions it argues about (D094 (b))", d.ID)
		}
		var known *ProfileAction
		for _, n := range names {
			a, ok := q.Action(n)
			switch {
			case !ok:
			case known != nil && a.ID != known.ID:
				return nil, fmt.Errorf("P1: %s is already declared (the action %s)", n, a.ID)
			default:
				known = a
			}
		}
		if known == nil {
			q.Actions = append(q.Actions, ProfileAction{ID: d.ID, Actor: actor, Aliases: d.Aliases})
			continue
		}
		if known.Actor != actor {
			return nil, fmt.Errorf("P1: %s is already declared, as %s's action (the action %s), not %s's", d.ID, known.Actor, known.ID, actor)
		}
		for _, n := range names {
			if _, ok := q.Action(n); !ok {
				known.Aliases = append(known.Aliases, n)
			}
		}
	}
	if err := q.Check(); err != nil {
		return nil, fmt.Errorf("P1: %v", err)
	}
	return q, nil
}

// actorID returns the id of the actor a name names, compared as atoms, as
// actions' names are.
func (p *Profile) actorID(name string) (string, bool) {
	atom := carneades.TextToAtom(name)
	for _, a := range p.Actors {
		if carneades.TextToAtom(a.ID) == atom {
			return a.ID, true
		}
	}
	return "", false
}

// actorIDs are the profile's actors' ids, in order.
func (p *Profile) actorIDs() []string {
	var out []string
	for _, a := range p.Actors {
		out = append(out, a.ID)
	}
	return out
}

func (p *Profile) clone() *Profile {
	q := *p
	q.Actors = append([]ProfileActor(nil), p.Actors...)
	q.Actions = make([]ProfileAction, len(p.Actions))
	for i, a := range p.Actions {
		a.Aliases = append([]string(nil), a.Aliases...)
		q.Actions[i] = a
	}
	return &q
}

// beforeValidation applies the profile to a move before NASCQ validates it:
// its declarations (P1), the agent bound to every agent place (P2), and each
// action place's alias read as its action. It returns the move to validate
// and the profile the move would leave.
func (gs *GameState) beforeValidation(a nascq.ProposedAction) (nascq.ProposedAction, *Profile, error) {
	if gs.Profile == nil {
		if len(a.Declare) > 0 {
			return a, nil, fmt.Errorf("P1: this game has no profile (it was made before step C5.2), so it declares no action")
		}
		return a, nil, nil
	}
	// A question argues about nothing: it declares nothing, and the scheme
	// and values it may carry are not its own (S0 sets them aside).
	s := gs.NASCQ.GetScheme(a.SchemeID)
	if a.Type == "question" || s == nil {
		if len(a.Declare) > 0 {
			return a, nil, fmt.Errorf("P1: the move declares %s but argues about no action: a move declares the actions it argues about (D094 (b))", a.Declare[0].ID)
		}
		return a, gs.Profile, nil
	}
	argued := map[string]bool{}
	for _, v := range nascq.SortedStringKeys(s.Variables) {
		if x, ok := a.VariableBindings[v]; ok && s.Variables[v].Type == nascq.VarAction {
			argued[carneades.TextToAtom(x)] = true
		}
	}
	p, err := gs.Profile.declared(a.Declare, argued)
	if err != nil {
		return a, nil, err
	}
	b := make(map[string]string, len(a.VariableBindings)+1)
	for k, v := range a.VariableBindings {
		b[k] = v
	}
	for _, v := range nascq.SortedStringKeys(s.Variables) {
		switch s.Variables[v].Type {
		case nascq.VarAgent:
			if x, ok := b[v]; ok && carneades.TextToAtom(x) != carneades.TextToAtom(p.Agent) {
				return a, nil, fmt.Errorf("P2: %s is the game's agent, %s, not %s (the profile)", v, p.Agent, x)
			}
			b[v] = p.Agent
		case nascq.VarAction:
			if x, ok := b[v]; ok {
				act, found := p.Action(x)
				if !found {
					return a, nil, fmt.Errorf("P1: %s = %s is no declared action: declare it with the move (declare: its id, actor and aliases)", v, x)
				}
				b[v] = act.ID
			}
		}
	}
	a.VariableBindings = b
	return a, p, nil
}

// Validate validates a move as a turn would, without playing it: the
// profile's rules (P1 to P4) around NASCQ's (S0 to S9). The preview uses
// it, so that it accepts the moves the game accepts.
func (gs *GameState) Validate(a nascq.ProposedAction) (*nascq.ValidatedAction, error) {
	a, p, err := gs.beforeValidation(a)
	if err != nil {
		return nil, err
	}
	va, err := gs.NASCQ.Validate(a, gs.bridge)
	if err != nil {
		return nil, err
	}
	if err := gs.afterValidation(va, p); err != nil {
		return nil, err
	}
	return va, nil
}

// afterValidation applies the profile to the values a validated move binds,
// its own and the ones its target fixes: each action declared (P1) and the
// agent's (P3), the actions a scheme keeps apart distinct, and a recorded
// comparison where the scheme asks for one (P4).
func (gs *GameState) afterValidation(va *nascq.ValidatedAction, p *Profile) error {
	if p == nil {
		return nil
	}
	s := gs.NASCQ.GetScheme(va.SchemeID)
	if s == nil {
		return nil
	}
	value := func(v string) string {
		if x, ok := va.Fixed[v]; ok {
			return x
		}
		return va.VariableBindings[v]
	}
	ids := map[string]string{}
	for _, v := range nascq.SortedStringKeys(s.Variables) {
		if s.Variables[v].Type != nascq.VarAction {
			continue
		}
		x := value(v)
		act, ok := p.Action(x)
		if !ok {
			return fmt.Errorf("P1: %s = %s is no declared action", v, x)
		}
		if act.Actor != p.Agent {
			return fmt.Errorf("P3: %s = %s is %s's action, not the agent's (%s): a recommendation names no agent but the one addressed (D063)", v, act.ID, act.Actor, p.Agent)
		}
		ids[v] = act.ID
	}
	for _, pair := range s.Distinct {
		if len(pair) == 2 && ids[pair[0]] != "" && ids[pair[0]] == ids[pair[1]] {
			return fmt.Errorf("P4: %s and %s are the same action (%s), aliases included: the alternative must be another action (D089)", pair[0], pair[1], ids[pair[0]])
		}
	}
	if s.Comparison && !containsString(Comparisons, va.Comparison) {
		return fmt.Errorf("P4: %s records how the alternative is better (comparison: %s; D089)", s.Name, strings.Join(Comparisons, ", "))
	}
	return nil
}

// practicalClaim matches a recommendation, should_be_performed(A) or its
// negation, and gives its negation sign and its action.
var practicalClaim = regexp.MustCompile(`^(¬?)\s*should_be_performed\s*\(\s*([^(),]*[^(),\s])\s*\)$`)

// claimAction returns the action a claim recommends or advises against, as
// written, or "" if it is not a recommendation.
func claimAction(claim string) string {
	if m := practicalClaim.FindStringSubmatch(strings.TrimSpace(claim)); m != nil {
		return m[2]
	}
	return ""
}

// canonicalClaim is a recommendation written with its action's id: the
// claim names the action the moves about it bind (P1 reads an alias as its
// action), so that an argument for the action concludes the claim. A name
// the profile does not hold is written as the atom the board makes of it.
func (p *Profile) canonicalClaim(claim string) string {
	m := practicalClaim.FindStringSubmatch(strings.TrimSpace(claim))
	if m == nil {
		return claim
	}
	id := carneades.TextToAtom(m[2])
	if act, ok := p.Action(m[2]); ok {
		id = act.ID
	}
	return m[1] + "should_be_performed(" + id + ")"
}

// CheckClaim reports why a game's claim cannot be argued under the profile:
// a recommendation of an action the profile gives to another actor (D063).
func (p *Profile) CheckClaim(claim string) error {
	a := claimAction(claim)
	if a == "" {
		return nil
	}
	if act, ok := p.Action(a); ok && act.Actor != p.Agent {
		return fmt.Errorf("the claim %s recommends %s, which is %s's action, not the agent's (%s): a recommendation names no agent but the one addressed (D063)",
			claim, act.ID, act.Actor, p.Agent)
	}
	return nil
}

// withClaim is the profile with the action of the game's claim declared, as
// the agent's, if it is a recommendation of an action the profile does not
// hold: the claim is the first thing argued about.
func (p *Profile) withClaim(claim string) *Profile {
	q := p.clone()
	if a := claimAction(claim); a != "" {
		if _, ok := q.Action(a); !ok {
			q.Actions = append(q.Actions, ProfileAction{ID: carneades.TextToAtom(a), Actor: q.Agent})
		}
	}
	return q
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// agentFixed is a template's fixed values with its agent places bound to the
// profile's agent (P2): the game binds them, so a player does not.
func (gs *GameState) agentFixed(t nascq.ActionTemplate) map[string]string {
	if gs.Profile == nil {
		return t.Fixed
	}
	var out map[string]string
	for _, v := range nascq.SortedStringKeys(t.Variables) {
		if t.Variables[v].Type != nascq.VarAgent {
			continue
		}
		if out == nil {
			out = make(map[string]string, len(t.Fixed)+1)
			for k, x := range t.Fixed {
				out[k] = x
			}
		}
		out[v] = gs.Profile.Agent
	}
	if out == nil {
		return t.Fixed
	}
	return out
}

// forProfile is a computer player's move made fit for the profile: its
// agent places bound to the agent, and the actions it names that the
// profile does not hold declared, as the agent's (a computer player argues
// about the agent's actions only).
func (gs *GameState) forProfile(a nascq.ProposedAction) nascq.ProposedAction {
	if gs.Profile == nil {
		return a
	}
	s := gs.NASCQ.GetScheme(a.SchemeID)
	if s == nil {
		return a
	}
	b := make(map[string]string, len(a.VariableBindings))
	for k, v := range a.VariableBindings {
		b[k] = v
	}
	seen := map[string]bool{}
	for _, v := range nascq.SortedStringKeys(s.Variables) {
		switch s.Variables[v].Type {
		case nascq.VarAgent:
			b[v] = gs.Profile.Agent
		case nascq.VarAction:
			x, ok := b[v]
			if !ok || seen[x] {
				continue
			}
			if _, declared := gs.Profile.Action(x); !declared {
				a.Declare = append(a.Declare, nascq.ActionDecl{ID: x, Actor: gs.Profile.Agent})
				seen[x] = true
			}
		}
	}
	a.VariableBindings = b
	return a
}
