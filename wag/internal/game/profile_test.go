package game

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/internal/carneades"
	"github.com/yourorg/wag/internal/nascq"
	"gopkg.in/yaml.v2"
)

// The game's profile (D094 (b)): who deliberates, the actions argued about
// with who performs each, and the situation compared. The tests play on the
// shipped data with two schemes of their own: practical reasoning with an
// agent place (pr_agent) and the answer scheme better_means, which backs its
// question about a better means.

const caseJSON = `{
 "format": "wag-case/1", "id": "kialo-test",
 "thesis": {"claim": "1", "action": "tax_meat"},
 "default_context": "present",
 "contexts": [{"id": "present", "text": "Governments deciding now whether to tax meat."}],
 "actors": [{"id": "governments", "text": "the governments that would tax meat"}, {"id": "farmers", "text": "farmers"}],
 "actions": [
  {"id": "tax_meat", "actor": "governments", "aliases": ["meat_tax"]},
  {"id": "end_subsidies", "actor": "governments"},
  {"id": "switch_crops", "actor": "farmers"}
 ]
}`

func caseProfile(t *testing.T) *Profile {
	t.Helper()
	p, err := ProfileFromCase([]byte(caseJSON))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// profileGame is a game with the profile p (nil: none) on the shipped data,
// pr_agent and better_means added.
func profileGame(t *testing.T, p *Profile) *GameState {
	t.Helper()
	core := prevCore
	for k, tmpl := range map[string]string{
		"goal_of_agent/2":            "%s has the goal %s.",
		"realize_more_effectively/3": "%s could realize %s more effectively than by %s.",
		"possible_bring_about/2":     "%s can bring about %s.",
		"more_effective_means/4":     "For %s, %s realizes %s more effectively than %s does.",
	} {
		core = addTemplate(t, core, k, tmpl)
	}
	core = withScheme(t, core, yaml.MapSlice{{Key: "id", Value: "pr_agent"},
		{Key: "variables", Value: []string{"A", "G", "P"}},
		{Key: "premises", Value: []string{"goal_of_agent(P,G)", "brings_about(A,G)"}},
		{Key: "exceptions", Value: []string{"realize_more_effectively(P,G,A)"}},
		{Key: "conclusions", Value: []string{"should_be_performed(A)"}}})
	core = withScheme(t, core, yaml.MapSlice{{Key: "id", Value: "better_means"},
		{Key: "variables", Value: []string{"A", "A2", "G", "P"}},
		{Key: "premises", Value: []string{"goal_of_agent(P,G)", "possible_bring_about(P,A2)", "brings_about(A2,G)", "more_effective_means(P,A2,G,A)"}},
		{Key: "conclusions", Value: []string{"realize_more_effectively(P,G,A)"}}})
	ad, err := carneades.NewAdapterFromYAML(core)
	if err != nil {
		t.Fatal(err)
	}
	var schemes, cqs map[string]json.RawMessage
	if err := json.Unmarshal(prevSchemes, &schemes); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(prevCQs, &cqs); err != nil {
		t.Fatal(err)
	}
	schemes["AS:pr_agent"] = json.RawMessage(`{"carneades_id":"pr_agent","name":"Practical Reasoning of an Agent","semantic_type":"action",
		"base_weight":0.7,"variables":{"A":{"type":"action"},"G":{"type":"goal"},"P":{"type":"agent"}},
		"premise_templates_en":["{P} has the goal {G}","Doing {A} brings about {G}"],"conclusion_template_en":"{A} should be done"}`)
	schemes["AS:better_means"] = json.RawMessage(`{"carneades_id":"better_means","name":"Better Means","semantic_type":"action",
		"base_weight":0.5,"role":"answer","distinct":[["A2","A"]],"comparison":true,
		"variables":{"A":{"type":"action"},"A2":{"type":"action"},"G":{"type":"goal"},"P":{"type":"agent"}},
		"premise_templates_en":["{P} has the goal {G}","{P} can do {A2}","{A2} brings about {G}","{A2} is a better means than {A}"],
		"conclusion_template_en":"{P} could realize {G} better than by {A}"}`)
	cqs["CQ:pa_other_means"] = json.RawMessage(`{"name":"Better Means Question","semantic_type":"action",
		"question_template_en":"Could {P} realize {G} better by another action than {A}?",
		"caes_mapping":{"positive_answer_type":"support","negative_answer_type":"undercutter","target":"argument"}}`)
	var edges map[string][]json.RawMessage
	if err := json.Unmarshal(prevEdges, &edges); err != nil {
		t.Fatal(err)
	}
	edges["ask_edges"] = append(edges["ask_edges"], json.RawMessage(
		`{"cq":"CQ:pa_other_means","scheme":"AS:pr_agent","locus":"exception","pattern":"realize_more_effectively({P},{G},{A})"}`))
	edges["backup_edges"] = append(edges["backup_edges"], json.RawMessage(`{"scheme":"AS:better_means","cq":"CQ:pa_other_means"}`))
	sb, _ := json.Marshal(schemes)
	qb, _ := json.Marshal(cqs)
	eb, _ := json.Marshal(edges)
	n, err := nascq.LoadFromBytes(sb, qb, eb)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultGameConfig()
	cfg.MaxTurns = 0
	cfg.Profile = p
	return NewGame(cfg, ad, n, []string{"Alice", "Bob"})
}

// refused plays a move that must be refused for the reason want.
func refused(t *testing.T, gs *GameState, a nascq.ProposedAction, want string) {
	t.Helper()
	before := gs.Profile
	r, err := gs.ExecuteTurn(a)
	if err != nil {
		t.Fatal(err)
	}
	if r.Accepted || !strings.Contains(gs.LastRejection, want) {
		t.Fatalf("%+v: accepted %v, %q; want a refusal with %q", a, r.Accepted, gs.LastRejection, want)
	}
	if gs.Profile != before {
		t.Fatalf("a refused move changed the profile")
	}
}

func cons(a string) nascq.ProposedAction {
	return nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": a, "G": "lower_emissions"}}
}

// A case record gives the profile: its actors and actions, the thesis
// action's actor as the agent, the default context as the situation.
func TestProfileFromCase(t *testing.T) {
	p := caseProfile(t)
	if p.Agent != "governments" || len(p.Actions) != 3 || p.Source != "case:kialo-test" ||
		p.Situation != "Governments deciding now whether to tax meat." {
		t.Fatalf("%+v", p)
	}
	if a, ok := p.Action("meat_tax"); !ok || a.ID != "tax_meat" {
		t.Errorf("the alias reads as %+v", a)
	}
	// Names are compared as the board's atoms: another spelling is the same
	// action.
	if a, ok := p.Action("Meat Tax"); !ok || a.ID != "tax_meat" {
		t.Errorf("another spelling of the alias reads as %+v", a)
	}
	for name, bad := range map[string]string{
		"format":   strings.Replace(caseJSON, "wag-case/1", "wag-case/2", 1),
		"thesis":   strings.Replace(caseJSON, `"action": "tax_meat"}`, `"action": "tax_fish"}`, 1),
		"context":  strings.Replace(caseJSON, `"default_context": "present"`, `"default_context": "past"`, 1),
		"actor":    strings.Replace(caseJSON, `"actor": "farmers"`, `"actor": "fishers"`, 1),
		"alias":    strings.Replace(caseJSON, `"aliases": ["meat_tax"]`, `"aliases": ["end_subsidies"]`, 1),
		"spelling": strings.Replace(caseJSON, `"aliases": ["meat_tax"]`, `"aliases": ["End Subsidies"]`, 1),
		"twice":    strings.Replace(caseJSON, `{"id": "end_subsidies", "actor": "governments"}`, `{"id": "tax meat", "actor": "governments"}`, 1),
	} {
		if _, err := ProfileFromCase([]byte(bad)); err == nil {
			t.Errorf("a case record with a bad %s was taken", name)
		}
	}
	// The default profile is sound; so is the batch-2 case record's.
	if d := DefaultProfile(); d.Check() != nil || d.Agent != DefaultAgent {
		t.Errorf("the default profile: %+v %v", d, d.Check())
	}
	b, err := os.ReadFile("../../../docs/kialo-batch2/case.json")
	if err != nil {
		t.Fatal(err)
	}
	if p, err := ProfileFromCase(b); err != nil || p.Agent != "governments" || len(p.Actions) != 10 {
		t.Errorf("the batch-2 case record: %v %+v", err, p)
	}
}

// P1: a move binds actions the profile declares, or declares them itself;
// an alias reads as its action; a refused move declares nothing.
func TestProfileDeclaredActions(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	refused(t, gs, cons("ban_meat"), "P1: A = ban_meat is no declared action")
	// A declaration names its actor (D094 (b)): none named, none presumed.
	nameless := cons("ban_meat")
	nameless.Declare = []nascq.ActionDecl{{ID: "ban_meat"}}
	refused(t, gs, nameless, "P1: the declaration of ban_meat names its actor, one of governments, farmers")
	d := cons("ban_meat")
	d.Declare = []nascq.ActionDecl{{ID: "ban_meat", Actor: "governments", Aliases: []string{"meat_ban"}}}
	id := play(t, gs, d)
	if a, ok := gs.Profile.Action("meat_ban"); !ok || a.Actor != "governments" {
		t.Fatalf("the declared action: %+v", a)
	}
	if got := gs.MoveHistory[len(gs.MoveHistory)-1].Action.Declare; len(got) != 1 || got[0].ID != "ban_meat" {
		t.Errorf("the declaration is not recorded with the move: %+v", got)
	}
	if c := gs.Adapter.Graph.Arguments[id].Conclusion.Id; c != "should_be_performed(ban_meat)" {
		t.Errorf("the argument concludes %s", c)
	}
	// An alias is read as its action: the same statement as the id's.
	id = play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "meat_tax", "G": "less_meat_eaten"}})
	if c := gs.Adapter.Graph.Arguments[id].Conclusion.Id; c != "should_be_performed(tax_meat)" {
		t.Errorf("the alias's argument concludes %s", c)
	}
	// A name already declared as another actor's, an unknown actor, an
	// action the move does not argue about: refused, nothing declared.
	again := cons("meat_ban")
	again.Declare = []nascq.ActionDecl{{ID: "meat_ban", Actor: "farmers"}}
	refused(t, gs, again, "P1: meat_ban is already declared, as governments's action (the action ban_meat), not farmers's")
	stranger := cons("fish_tax")
	stranger.Declare = []nascq.ActionDecl{{ID: "fish_tax", Actor: "fishers"}}
	refused(t, gs, stranger, "P1: the declaration of fish_tax names the actor fishers, none of governments, farmers")
	if _, ok := gs.Profile.Action("fish_tax"); ok {
		t.Error("a refused move's declaration joined the profile")
	}
	unargued := cons("tax_meat")
	unargued.Declare = []nascq.ActionDecl{{ID: "tax_fish", Actor: "governments"}}
	refused(t, gs, unargued, "P1: the move declares tax_fish but argues about no such action")
	// A sound declaration on a move refused later (P3) joins nothing either.
	late := cons("tax_fish")
	late.Declare = []nascq.ActionDecl{{ID: "tax_fish", Actor: "farmers"}}
	refused(t, gs, late, "P3: A = tax_fish is farmers's action")
	if _, ok := gs.Profile.Action("tax_fish"); ok {
		t.Error("the declaration of a move refused after it joined the profile")
	}
	// A simulated move leaves the profile as it was.
	before := gs.Profile
	sim := cons("tax_fish")
	sim.Declare = []nascq.ActionDecl{{ID: "tax_fish", Actor: "governments"}}
	if r, err := gs.Simulate(sim); err != nil || !r.Accepted {
		t.Fatalf("the simulated move: %v %+v", err, r)
	}
	if gs.Profile != before {
		t.Error("a simulated move changed the profile")
	}
}

// P1 holds for the values a move's target fixes too: a statement about an
// action the profile does not hold (here: a profile replaced under the
// board, as a record made without one would leave it) cannot be argued about.
func TestProfileFixedValues(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	d := cons("ban_meat")
	d.Declare = []nascq.ActionDecl{{ID: "ban_meat", Actor: "governments"}}
	play(t, gs, d)
	gs.Profile = caseProfile(t) // ban_meat no longer declared
	refused(t, gs, nascq.ProposedAction{Type: "attack", SchemeID: "AS:negative_consequences",
		TargetStmtID: "should_be_performed(ban_meat)", VariableBindings: map[string]string{"G": "fewer_farms"}},
		"P1: A = ban_meat is no declared action")
}

// P2: the game binds an agent place to the profile's agent and refuses a
// move that binds it to anyone else; the board shows it as fixed. P3: the
// actions argued are the agent's, aliases included (D063).
func TestProfileAgent(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	pr := nascq.ProposedAction{Type: "support", SchemeID: "AS:pr_agent",
		VariableBindings: map[string]string{"A": "tax_meat", "G": "lower_emissions"}}
	id := play(t, gs, pr)
	found := false
	for _, p := range gs.Adapter.Graph.Arguments[id].Premises {
		found = found || p.Stmt.Id == "goal_of_agent(governments,lower_emissions)"
	}
	if !found {
		t.Errorf("the agent is not bound: %v", gs.Adapter.Graph.Arguments[id].Premises)
	}
	other := pr
	other.VariableBindings = map[string]string{"A": "end_subsidies", "G": "lower_emissions", "P": "farmers"}
	refused(t, gs, other, "P2: P is the game's agent, governments, not farmers")
	refused(t, gs, cons("switch_crops"), "P3: A = switch_crops is farmers's action, not the agent's (governments)")
	for _, av := range gs.BuildView().Actions {
		if av.SchemeID != "AS:pr_agent" {
			continue
		}
		if av.Fixed["P"] != "governments" {
			t.Errorf("the board does not show P as fixed: %+v", av.Fixed)
		}
		if _, ok := av.Variables["P"]; ok {
			t.Error("the board asks the player to bind P")
		}
	}
	if gs.BuildView().Profile == nil {
		t.Error("the board does not show the profile")
	}
}

// P4: better_means' alternative is another of the agent's actions, not A or
// an alias of A, and the answer records its comparison; its goal premise is
// the answered argument's.
func TestProfileBetterMeans(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	target := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:pr_agent",
		VariableBindings: map[string]string{"A": "tax_meat", "G": "lower_emissions"}})
	ans := func(a2, comparison string) nascq.ProposedAction {
		return nascq.ProposedAction{Type: "attack", SchemeID: "AS:better_means", AnswersCQID: "CQ:pa_other_means",
			TargetArgID: target, Comparison: comparison, VariableBindings: map[string]string{"A2": a2}}
	}
	refused(t, gs, ans("tax_meat", "cheaper"), "P4: A2 and A are the same action (tax_meat)")
	refused(t, gs, ans("meat_tax", "cheaper"), "P4: A2 and A are the same action (tax_meat)")
	refused(t, gs, ans("switch_crops", "cheaper"), "P3: A2 = switch_crops is farmers's action")
	refused(t, gs, ans("end_subsidies", ""), "P4: Better Means records how the alternative is better")
	refused(t, gs, ans("end_subsidies", "better"), "P4: Better Means records how the alternative is better")
	id := play(t, gs, ans("end_subsidies", "cheaper"))
	goal := false
	for _, p := range gs.Adapter.Graph.Arguments[id].Premises {
		goal = goal || p.Stmt.Id == "goal_of_agent(governments,lower_emissions)"
	}
	if !goal {
		t.Error("the answer's goal premise is not the answered argument's")
	}
	if c := gs.MoveHistory[len(gs.MoveHistory)-1].Action.Comparison; c != "cheaper" {
		t.Errorf("the comparison is not recorded: %q", c)
	}
	// Under another audience's weights the answer weighs as the argument it
	// answers (D094 (g)).
	if got := gs.weighedAs(id, "AS:better_means"); got != "AS:pr_agent" {
		t.Errorf("the answer is weighed as %s", got)
	}
}

// A game without a profile (made before step C5.2) plays without its rules;
// a game with a claim declares the claim's action as the agent's.
func TestProfileOffAndClaim(t *testing.T) {
	gs := profileGame(t, nil)
	play(t, gs, cons("anything_at_all"))
	cfg := DefaultGameConfig()
	cfg.Profile, cfg.Objective, cfg.Claim = DefaultProfile(), ObjectiveClaim, "should_be_performed(carbon_tax)"
	g2 := NewGame(cfg, profileGame(t, nil).Adapter, profileGame(t, nil).NASCQ, []string{"Alice", "Bob"})
	if a, ok := g2.Profile.Action("carbon_tax"); !ok || a.Actor != DefaultAgent {
		t.Errorf("the claim's action: %+v", a)
	}
	if cfg.Profile.Actions != nil {
		t.Error("the configured profile was changed")
	}
	if err := caseProfile(t).CheckClaim("should_be_performed(switch_crops)"); err == nil ||
		!strings.Contains(err.Error(), "farmers's action, not the agent's") {
		t.Errorf("a claim about another actor's action: %v", err)
	}
	if err := caseProfile(t).CheckClaim("¬should_be_performed(tax_meat)"); err != nil {
		t.Error(err)
	}
}

// A computer player's move is made fit for the profile: its agent place
// bound to the agent, the actions it names declared as the agent's.
func TestForProfile(t *testing.T) {
	gs := profileGame(t, caseProfile(t))
	a := gs.forProfile(nascq.ProposedAction{Type: "support", SchemeID: "AS:pr_agent",
		VariableBindings: map[string]string{"A": "the proposal", "G": "lower_emissions", "P": "something else"}})
	if a.VariableBindings["P"] != "governments" || len(a.Declare) != 1 || a.Declare[0].ID != "the proposal" || a.Declare[0].Actor != "governments" {
		t.Fatalf("%+v", a)
	}
	play(t, gs, a)
	if gs.vocabulary()[nascq.VarAction][0] != "tax_meat" {
		t.Errorf("the vocabulary's actions: %v", gs.vocabulary()[nascq.VarAction])
	}
}

// The profile's check refuses an actor, an action or an alias that names
// nothing (no letter or digit), naming it, and marks the refusal so that a
// server can name the build that replays a game recorded with one (the
// check of step 8, W-3). Other refusals are not so marked.
func TestProfileNamesThatNameNothing(t *testing.T) {
	good := func() *Profile {
		p := DefaultProfile()
		p.Actors = append(p.Actors, ProfileActor{ID: "farmers"})
		p.Actions = []ProfileAction{{ID: "tax_meat", Actor: DefaultAgent, Aliases: []string{"meat tax"}}, {ID: "strike", Actor: "farmers"}}
		return p
	}
	if err := good().Check(); err != nil {
		t.Fatalf("a good profile: %v", err)
	}
	for _, c := range []struct {
		name string
		edit func(*Profile)
		says string
	}{
		{"an actor ?", func(p *Profile) { p.Actors[1].ID = "?"; p.Actions[1].Actor = "?" }, `an actor whose id, "?", names nothing`},
		{"an action ①", func(p *Profile) { p.Actions[1].ID = "\u2460" }, "names nothing"},
		{"an alias —", func(p *Profile) { p.Actions[0].Aliases = []string{"\u2014"} }, "the action \"tax_meat\" has a name, \"\u2014\", that names nothing"},
		{"a zero-width alias", func(p *Profile) { p.Actions[0].Aliases = []string{string(rune(0x200b))} }, "names nothing"},
	} {
		p := good()
		c.edit(p)
		err := p.Check()
		if err == nil || !strings.Contains(err.Error(), c.says) || !errors.Is(err, ErrNamesNothing) {
			t.Errorf("%s: %v (marked: %v)", c.name, err, errors.Is(err, ErrNamesNothing))
		}
	}
	p := good()
	p.Actions[1].ID = "tax_meat"
	if err := p.Check(); err == nil || errors.Is(err, ErrNamesNothing) {
		t.Errorf("an action twice: %v, marked %v", err, errors.Is(err, ErrNamesNothing))
	}
}
