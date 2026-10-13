package game

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/wag/data"
	"github.com/yourorg/wag/internal/nascq"

	caes "github.com/carneades/carneades-4/src/engine/caes"
)

// modeGame is a shipped-data game in the given mode.
func modeGame(t *testing.T, mode string, players ...string) *GameState {
	t.Helper()
	gs := newShippedGame(t, data.EdgesJSON, players...)
	cfg, err := ModeConfig(mode)
	if err != nil {
		t.Fatal(err)
	}
	gs.Config = cfg
	return gs
}

func question(cq, target string) nascq.ProposedAction {
	return nascq.ProposedAction{Type: "question", AnswersCQID: cq, TargetArgID: target}
}

var abstain = nascq.ProposedAction{Type: "abstain"}

// A question about a premise suspends it until the proponent answers.
func TestQuestionShiftsTheBurdenOfAPremise(t *testing.T) {
	gs := modeGame(t, "classic", "Alice", "Bob")
	a1 := play(t, gs, prSupportAction())
	label := func(id string) caes.Label { return gs.Adapter.Graph.Statements[id].Label }

	q := play(t, gs, question("CQ:possibility", a1))
	if label("possible(carbon_tax)") != caes.Out || label("should_be_performed(carbon_tax)") != caes.Out {
		t.Fatalf("questioned premise %v, conclusion %v; want Out, Out", label("possible(carbon_tax)"), label("should_be_performed(carbon_tax)"))
	}
	if len(gs.Questions) != 1 || !gs.Questions[0].Open() || gs.Questions[0].Owner != "Alice" || gs.Questions[0].ArgID != q {
		t.Fatalf("questions: %+v", gs.Questions)
	}
	if r := gs.MoveHistory[len(gs.MoveHistory)-1]; r.ScoreDelta != 0 {
		t.Fatalf("a question scored %v", r.ScoreDelta)
	}

	ans := play(t, gs, answer("support", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_yes", "D": "politics"}))
	if label("possible(carbon_tax)") != caes.In || label("should_be_performed(carbon_tax)") != caes.In {
		t.Fatalf("answered premise %v, conclusion %v; want In, In", label("possible(carbon_tax)"), label("should_be_performed(carbon_tax)"))
	}
	if gs.Questions[0].AnsweredBy != ans {
		t.Fatalf("question answered by %q, want %s", gs.Questions[0].AnsweredBy, ans)
	}

	// The same question cannot be asked again, and the question argument is
	// not a target.
	r, _ := gs.ExecuteTurn(question("CQ:possibility", a1))
	if r.Accepted || !strings.HasPrefix(r.Error, "Q2") {
		t.Fatalf("repeated question: %+v", r)
	}
	r, _ = gs.ExecuteTurn(question("CQ:evidence", q))
	if r.Accepted || !strings.HasPrefix(r.Error, "S3") {
		t.Fatalf("question about a question: %+v", r)
	}
}

// A question about the inference puts the argument out of force until its
// proponent answers it.
func TestQuestionAboutTheInference(t *testing.T) {
	gs := modeGame(t, "classic", "Alice", "Bob")
	a1 := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:consequences",
		VariableBindings: map[string]string{"A": "carbon_tax", "G": "clean_air"}})
	play(t, gs, question("CQ:evidence", a1))
	if w := gs.Adapter.Graph.Arguments[a1].Weight; w != 0 {
		t.Fatalf("questioned inference: weight %v, want 0", w)
	}
	play(t, gs, answer("support", "CQ:evidence", "AS:expert_opinion", a1, map[string]string{"W": "epa", "D": "environment"}))
	if w := gs.Adapter.Graph.Arguments[a1].Weight; w == 0 || gs.Adapter.Graph.Arguments[a1].Conclusion.Label != caes.In {
		t.Fatalf("answered inference: weight %v, conclusion %v", w, gs.Adapter.Graph.Arguments[a1].Conclusion.Label)
	}
}

// An exception, or a counterclaim, is the asker's to prove: it cannot be
// raised by a question alone.
func TestQuestionsAboutExceptionsAndConclusionsNeedAnArgument(t *testing.T) {
	gs := modeGame(t, "classic", "Alice", "Bob")
	a1 := play(t, gs, prSupportAction())
	for _, cq := range []string{"CQ:side_effect", "CQ:other_goals"} {
		r, _ := gs.ExecuteTurn(question(cq, a1))
		if r.Accepted || !strings.HasPrefix(r.Error, "Q1") {
			t.Errorf("question %s: %+v", cq, r)
		}
	}
	// The enumeration offers only the questions that can be asked.
	for _, a := range gs.BuildObservation().AvailableActions {
		if a.Type == "question" && a.Locus.Kind != nascq.LocusPremise && a.Locus.Kind != nascq.LocusArgument {
			t.Errorf("enumerated question about the %s: %+v", a.Locus.Kind, a)
		}
	}
}

// In critical-question play, a question left unanswered on the proponent's
// next turn loses the game.
func TestCQModeUnansweredQuestionLoses(t *testing.T) {
	gs := modeGame(t, "cq", "Alice", "Bob")
	a1 := play(t, gs, prSupportAction())
	play(t, gs, question("CQ:possibility", a1))
	r, _ := gs.ExecuteTurn(abstain) // Alice does not answer
	if !r.Terminated || r.Outcome == nil || r.Outcome.Winner != "Bob" || r.Outcome.Loser != "Alice" {
		t.Fatalf("unanswered question: %+v %+v", r, r.Outcome)
	}

	// Answered, the game goes on; questions about the answer must be
	// answered in turn.
	gs = modeGame(t, "cq", "Alice", "Bob")
	a1 = play(t, gs, prSupportAction())
	play(t, gs, question("CQ:possibility", a1))
	ans := play(t, gs, answer("support", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_yes", "D": "politics"}))
	play(t, gs, question("CQ:expertise", ans))
	play(t, gs, answer("support", "CQ:expertise", "AS:position_to_know", ans, map[string]string{"W": "university", "D": "politics"}))
	if gs.Terminated {
		t.Fatalf("game over after every question was answered: %s", gs.TerminationReason)
	}
	// Answering something else does not discharge the obligation.
	play(t, gs, question("CQ:field", ans))
	r, _ = gs.ExecuteTurn(answer("attack", "CQ:side_effect", "AS:sign", a1, map[string]string{"A": "x"}))
	if !r.Terminated || r.Outcome.Loser != "Alice" {
		t.Fatalf("Alice answered another question: %+v", r)
	}
}

func TestResign(t *testing.T) {
	gs := modeGame(t, "infinite", "Alice", "Bob")
	play(t, gs, prSupportAction())
	r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "resign"})
	if !r.Terminated || r.Outcome.Winner != "Alice" || r.Outcome.Loser != "Bob" {
		t.Fatalf("Bob resigned: %+v %+v", r, r.Outcome)
	}

	// With three players, the others play on, and the resigned player's
	// turns are skipped.
	gs = modeGame(t, "infinite", "Alice", "Bob", "Carol")
	gs.ExecuteTurn(abstain)
	r, _ = gs.ExecuteTurn(nascq.ProposedAction{Type: "resign"})
	if r.Terminated || gs.ActivePlayer().ID != "Carol" {
		t.Fatalf("after Bob resigned: terminated %v, active %s", r.Terminated, gs.ActivePlayer().ID)
	}
	gs.ExecuteTurn(abstain)
	if gs.ActivePlayer().ID != "Alice" {
		t.Fatalf("after Carol: active %s", gs.ActivePlayer().ID)
	}
	gs.ExecuteTurn(abstain)
	if gs.ActivePlayer().ID != "Carol" {
		t.Fatalf("Bob's turn was not skipped: active %s", gs.ActivePlayer().ID)
	}
	r, _ = gs.ExecuteTurn(nascq.ProposedAction{Type: "resign"})
	if !r.Terminated || r.Outcome.Winner != "Alice" {
		t.Fatalf("two resigned: %+v %+v", r, r.Outcome)
	}
}

func TestTargetScore(t *testing.T) {
	gs := modeGame(t, "target", "Alice", "Bob")
	gs.Config.TargetScore = 10
	play(t, gs, prSupportAction()) // 7
	play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "ipcc", "D": "climate", "S": "warming_is_real"}})
	r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "nasa", "D": "climate", "S": "seas_rise"}}) // 14
	if !r.Terminated || r.Outcome.Winner != "Alice" || gs.Truncated {
		t.Fatalf("target reached: %+v %+v", r, r.Outcome)
	}
}

// fakeClock is a clock the test moves forward.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time        { return c.t }
func (c *fakeClock) wait(d time.Duration)  { c.t = c.t.Add(d) }
func newFakeClock() *fakeClock             { return &fakeClock{t: time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)} }
func setClock(gs *GameState, c *fakeClock) { gs.SetClock(c.now) }
func timedGame(t *testing.T, mode string) (*GameState, *fakeClock) {
	gs := modeGame(t, mode, "Alice", "Bob")
	c := newFakeClock()
	setClock(gs, c)
	return gs, c
}

func TestTimedGame(t *testing.T) {
	gs, clock := timedGame(t, "timed")
	play(t, gs, prSupportAction())
	clock.wait(11 * time.Minute)
	r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "support", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "ipcc", "D": "climate", "S": "warming_is_real"}})
	if r.Accepted || !r.Terminated || !gs.Truncated || r.Outcome.Winner != "Alice" {
		t.Fatalf("after the time limit: %+v %+v", r, r.Outcome)
	}
	if s := gs.Adapter.Graph.Statements["warming_is_real"]; s != nil {
		t.Fatal("the move after the time limit was played")
	}
}

func TestSpeedGame(t *testing.T) {
	gs, clock := timedGame(t, "speed")
	play(t, gs, prSupportAction())
	clock.wait(61 * time.Second) // Bob is too slow: his turn passes
	r, _ := gs.ExecuteTurn(question("CQ:possibility", "a1"))
	if !r.TimedOut || !r.Abstained || r.Terminated || gs.ActivePlayer().ID != "Alice" || len(gs.Questions) != 0 {
		t.Fatalf("late move: %+v", r)
	}
	clock.wait(30 * time.Second) // Alice is in time
	play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:expert_opinion",
		VariableBindings: map[string]string{"W": "ipcc", "D": "climate", "S": "warming_is_real"}})

	gs, clock = timedGame(t, "speed")
	gs.Config.MoveTimeoutLoses = true
	clock.wait(2 * time.Minute)
	r, _ = gs.ExecuteTurn(prSupportAction())
	if !r.Terminated || r.Outcome.Loser != "Alice" || r.Outcome.Winner != "Bob" {
		t.Fatalf("late move, timeout loses: %+v %+v", r, r.Outcome)
	}
}

// A debate: pro's opening argument states the claim, which is at issue;
// the verdict decides; a losing side that passes concedes.
func TestDebate(t *testing.T) {
	gs := modeGame(t, "debate", "Alice", "Bob")
	play(t, gs, prSupportAction())
	if gs.Claim != "should_be_performed(carbon_tax)" || gs.Adapter.Graph.Statements[gs.Claim].Issue == nil {
		t.Fatalf("claim %q not at issue", gs.Claim)
	}
	if gs.Side("Alice") != "pro" || gs.Side("Bob") != "con" {
		t.Fatal("sides")
	}
	// Bob passes while the claim stands: he concedes.
	r, _ := gs.ExecuteTurn(abstain)
	if !r.Terminated || r.Outcome.Verdict != "In" || r.Outcome.WinningSide != "pro" || r.Outcome.Winner != "Alice" {
		t.Fatalf("con passed: %+v %+v", r, r.Outcome)
	}

	// Bob undermines; Alice passes while the claim is Out: she cannot carry
	// her burden, so con wins.
	gs = modeGame(t, "debate", "Alice", "Bob")
	a1 := play(t, gs, prSupportAction())
	play(t, gs, answer("attack", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_no", "D": "politics"}))
	r, _ = gs.ExecuteTurn(abstain)
	if !r.Terminated || r.Outcome.Verdict != "Out" || r.Outcome.Winner != "Bob" {
		t.Fatalf("pro passed: %+v %+v", r, r.Outcome)
	}

	// A debate that runs out of rounds goes by the verdict too.
	gs = modeGame(t, "debate", "Alice", "Bob")
	gs.Config.MaxTurns = 1
	play(t, gs, prSupportAction())
	gs.Config.ConcedeOnPass = false
	r, _ = gs.ExecuteTurn(abstain)
	if !r.Terminated || r.Outcome.Winner != "Alice" {
		t.Fatalf("out of rounds with the claim In: %+v %+v", r, r.Outcome)
	}
}

// Under clear and convincing evidence, a rebuttal that pro outweighs by less
// than 0.5 keeps the claim from being In.
func TestDebateClaimStandard(t *testing.T) {
	gs := modeGame(t, "debate", "Alice", "Bob")
	gs.Config.ClaimStandard = caes.CCE
	a1 := play(t, gs, prSupportAction())
	if gs.Adapter.Graph.Statements[gs.Claim].Label != caes.In {
		t.Fatal("an unopposed 0.70 argument does not meet CCE")
	}
	play(t, gs, answer("attack", "CQ:other_goals", "AS:analogy", a1, map[string]string{"C": "spain"}))
	if l := gs.Adapter.Graph.Statements[gs.Claim].Label; l != caes.Out {
		t.Fatalf("claim %v after a 0.55 rebuttal under CCE, want Out", l)
	}
	if o := gs.Outcome(); o.WinningSide != "con" {
		t.Fatalf("provisional outcome %+v", o)
	}
}

// Solo play: one player argues both sides; self-attacks are moves like any
// other; the result is the verdict, with no winner.
func TestSolo(t *testing.T) {
	gs := modeGame(t, "solo", "Ann")
	a1 := play(t, gs, prSupportAction())
	if gs.ActivePlayer().ID != "Ann" || gs.Claim == "" {
		t.Fatalf("solo: active %s, claim %q", gs.ActivePlayer().ID, gs.Claim)
	}
	play(t, gs, answer("attack", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_no", "D": "politics"}))
	r, _ := gs.ExecuteTurn(abstain)
	if !r.Terminated || r.Outcome.Verdict != "Out" || r.Outcome.Winner != "" || r.Outcome.Draw {
		t.Fatalf("solo end: %+v %+v", r, r.Outcome)
	}
}

// Live scoring takes back the points of a defeated argument, and a second
// argument for the same statement adds nothing unless it is stronger.
func TestLiveScoring(t *testing.T) {
	gs := modeGame(t, "infinite", "Alice", "Bob")
	alice, bob := gs.Players[0], gs.Players[1]
	a1 := play(t, gs, prSupportAction())
	if alice.Score != 7 {
		t.Fatalf("Alice %v, want 7", alice.Score)
	}
	b1 := play(t, gs, answer("attack", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_no", "D": "politics"}))
	if alice.Score != 0 || bob.Score != 7 {
		t.Fatalf("after the undermining: Alice %v, Bob %v; want 0, 7", alice.Score, bob.Score)
	}
	if d := gs.MoveHistory[len(gs.MoveHistory)-1].ScoreDeltas; d["Alice"] != -7 || d["Bob"] != 7 {
		t.Fatalf("deltas %v", d)
	}
	// Alice defends with an equal expert: a tie, so neither side's
	// statement is In; then she undercuts Bob's expert.
	play(t, gs, answer("support", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_yes", "D": "politics"}))
	if alice.Score != 0 || bob.Score != 0 {
		t.Fatalf("after a tie: Alice %v, Bob %v", alice.Score, bob.Score)
	}
	gs.ExecuteTurn(abstain)
	play(t, gs, answer("attack", "CQ:trustworthiness", "AS:position_to_know", b1, map[string]string{"W": "journalist", "D": "media"}))
	// Alice: a1 (7) + her defence of possible (7) + the bias (6.5); the
	// live presets give no undercut bonus.
	if alice.Score != 20.5 || bob.Score != 0 {
		t.Fatalf("after the undercut: Alice %v, Bob %v; want 20.5, 0", alice.Score, bob.Score)
	}
	// With the bonus, the undercut is worth 15 while it holds.
	gs.Config.Scoring.UndercutterBonus = 15
	if s := gs.LiveScores()["Alice"]; s != 35.5 {
		t.Fatalf("with the undercut bonus: Alice %v, want 35.5", s)
	}
	gs.Config.Scoring.UndercutterBonus = 0
	// A second, weaker argument for a statement she already argued for adds
	// nothing.
	before := alice.Score
	gs.ExecuteTurn(abstain)
	play(t, gs, answer("support", "CQ:possibility", "AS:position_to_know", a1, map[string]string{"W": "treasury", "D": "politics"}))
	if alice.Score != before {
		t.Fatalf("a weaker second argument for possible: %v -> %v", before, alice.Score)
	}
}

// Simulate puts back everything the modes added to the game state.
func TestSimulateRestoresModeState(t *testing.T) {
	gs := modeGame(t, "cq", "Alice", "Bob", "Carol")
	gs.Config.Objective = ObjectivePoints
	a1 := play(t, gs, prSupportAction())
	play(t, gs, question("CQ:possibility", a1))
	before := gs.Adapter.Fingerprint()
	q := append([]OpenQuestion{}, gs.Questions...)
	for _, a := range []nascq.ProposedAction{
		{Type: "resign"},
		abstain,
		answer("support", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "dr_yes", "D": "politics"}),
		question("CQ:evidence", a1),
	} {
		if _, err := gs.Simulate(a); err != nil {
			t.Fatal(err)
		}
		if gs.Adapter.Fingerprint() != before || len(gs.Questions) != len(q) || gs.Questions[0] != q[0] || gs.Loser != "" || gs.Terminated {
			t.Fatalf("Simulate(%s) changed the game", a.Type)
		}
		for _, p := range gs.Players {
			if p.Resigned {
				t.Fatalf("Simulate(%s) left %s resigned", a.Type, p.ID)
			}
		}
	}
}

func TestModeConfig(t *testing.T) {
	for _, m := range Modes() {
		c, err := ModeConfig(m.Name)
		if err != nil || c.Mode != m.Name {
			t.Errorf("%s: %v %+v", m.Name, err, c)
		}
	}
	if _, err := ModeConfig("chess"); err == nil {
		t.Error("unknown mode accepted")
	}
}

// Random play keeps every invariant in every mode, with one player in solo
// and two or three otherwise; the clocks of the timed modes run on a fake
// clock that moves 10 to 69 seconds a move.
func TestInvariantsHoldInEveryMode(t *testing.T) {
	games := 6
	if testing.Short() {
		games = 2
	}
	for _, m := range Modes() {
		t.Run(m.Name, func(t *testing.T) {
			for i := 0; i < games; i++ {
				names := []string{"A", "B"}
				if m.Name == "solo" {
					names = names[:1]
				} else if i%2 == 1 {
					names = append(names, "C")
				}
				gs := modeGame(t, m.Name, names...)
				if gs.Config.MaxTurns == 0 {
					gs.Config.MaxTurns = 25
				}
				clock := newFakeClock()
				setClock(gs, clock)
				var agents []Agent
				for j, n := range names {
					agents = append(agents, NewRandomAgent(n, int64(100*i+j), 0.05, 0.4))
				}
				for it := 0; !gs.Terminated && it < 500; it++ {
					clock.wait(time.Duration(10+(it*37)%60) * time.Second)
					act := agents[gs.ActivePlayerIdx].ChooseAction(gs.BuildObservation())
					before := gs.Adapter.Fingerprint()
					r, err := gs.ExecuteTurn(act)
					if err != nil {
						t.Fatal(err)
					}
					if !r.Accepted && !r.Terminated {
						if gs.Adapter.Fingerprint() != before {
							t.Fatalf("rejected %+v changed the graph", act)
						}
						gs.ExecuteTurn(abstain)
					}
					if errs := gs.CheckInvariants(); len(errs) > 0 {
						t.Fatalf("game %d, after %+v: %v", i, act, errs)
					}
				}
				if !gs.Terminated {
					t.Fatalf("game %d did not end", i)
				}
				o := gs.Outcome()
				if (o.Winner == "") != (o.Draw || len(names) == 1 || o.WinningSide != "" && o.Winner == "") {
					t.Fatalf("game %d: outcome %+v", i, o)
				}
				t.Logf("game %d: %s; winner %q, verdict %q, %d moves", i, o.Reason, o.Winner, o.Verdict, len(gs.MoveHistory))
			}
		})
	}
}

// The greedy agent answers the questions it owes in critical-question play
// (not answering loses), and so does the random agent.
func TestAgentsAnswerTheirQuestions(t *testing.T) {
	for _, agent := range []Agent{NewGreedyAgent("Alice", 1, 1), NewRandomAgent("Alice", 1, 0.9, 0.5)} {
		gs := modeGame(t, "cq", "Alice", "Bob")
		a1 := play(t, gs, prSupportAction())
		play(t, gs, question("CQ:possibility", a1))
		act := agent.ChooseAction(gs.BuildObservation())
		if act.Type != "support" || act.TargetArgID != a1 || act.AnswersCQID != "CQ:possibility" {
			t.Fatalf("%T chose %+v", agent, act)
		}
		r, _ := gs.ExecuteTurn(act)
		if !r.Accepted || gs.Terminated || gs.Questions[0].Open() {
			t.Fatalf("%T's answer: %+v", agent, r)
		}
	}
}

// The wizard plays every kind of move: a defence, a question and a
// resignation.
func TestHumanWizardDefendQuestionResign(t *testing.T) {
	gs := modeGame(t, "cq", "Alice", "Bob")
	a1 := play(t, gs, prSupportAction())
	var out strings.Builder
	bob := NewHumanAgent("Bob", strings.NewReader("4\n1\n1\n"), &out) // question → a1 → possibility, the only question about a premise
	act := bob.ChooseAction(gs.BuildObservation())
	if act.Type != "question" || act.TargetArgID != a1 || act.AnswersCQID != "CQ:possibility" {
		t.Fatalf("question: %+v\n%s", act, out.String())
	}
	play(t, gs, act)

	out.Reset()
	alice := NewHumanAgent("Alice", strings.NewReader("3\n1\n4\n3\ndr yes\npolitics\n"), &out) // defend → a1 → possibility → expert
	act = alice.ChooseAction(gs.BuildObservation())
	shown := out.String()
	for _, want := range []string{
		"! Bob asks you about " + a1 + ": Is ‘carbon tax’ actually possible?",
		"[asked by Bob]",
		"The answer: ‘Carbon tax’ is possible (backing a premise of " + a1 + ").",
		"‘Dr yes’ says that C.",
	} {
		if !strings.Contains(shown, want) {
			t.Errorf("wizard output lacks %q:\n%s", want, shown)
		}
	}
	if act.Type != "support" || act.SchemeID != "AS:expert_opinion" || act.AnswersCQID != "CQ:possibility" || act.VariableBindings["S"] != "" {
		t.Fatalf("defence: %+v", act)
	}
	play(t, gs, act)
	if gs.Questions[0].Open() {
		t.Fatal("the defence did not answer the question")
	}

	bob = NewHumanAgent("Bob", strings.NewReader("6\n"), &out)
	if act = bob.ChooseAction(gs.BuildObservation()); act.Type != "resign" {
		t.Fatalf("resign: %+v", act)
	}
}

// Statements read in the players' words, nested statements quoted.
func TestSay(t *testing.T) {
	gs := modeGame(t, "classic", "Alice", "Bob")
	a1 := play(t, gs, nascq.ProposedAction{Type: "support", SchemeID: "AS:practical_reasoning",
		VariableBindings: map[string]string{"A": "carbon tax", "S1": "status quo", "S2": "lower emissions",
			"G": "clean air", "V": "health"}})
	play(t, gs, answer("attack", "CQ:possibility", "AS:expert_opinion", a1, map[string]string{"W": "Dr. No", "D": "politics"}))
	for id, want := range map[string]string{
		"possible(carbon_tax)":                     "‘Carbon tax’ is possible.",
		"¬possible(carbon_tax)":                    "‘Carbon tax’ is not possible.",
		"asserts(dr_no,not(possible(carbon_tax)))": "‘Dr. No’ says that “‘carbon tax’ is not possible”.",
		"undercut(a1)":                             "Argument a1 does not apply.",
	} {
		id := id
		if gs.Adapter.Graph.Statements[id] == nil {
			for sid := range gs.Adapter.Graph.Statements {
				if strings.HasPrefix(sid, "asserts(") {
					id = sid
				}
			}
		}
		if got := gs.Say(id); got != want {
			t.Errorf("Say(%s) = %q, want %q", id, got, want)
		}
	}
}

// A player who resigns while owing an answer leaves the game; the others
// play on.
func TestResignWhileOwingAnAnswer(t *testing.T) {
	gs := modeGame(t, "cq", "Alice", "Bob", "Carol")
	a1 := play(t, gs, prSupportAction())
	play(t, gs, question("CQ:possibility", a1))
	gs.ExecuteTurn(abstain) // Carol
	r, _ := gs.ExecuteTurn(nascq.ProposedAction{Type: "resign"})
	if r.Terminated || gs.Loser != "" || gs.ActivePlayer().ID != "Bob" {
		t.Fatalf("Alice resigned owing an answer: terminated %v (%s), active %s", r.Terminated, r.TerminationReason, gs.ActivePlayer().ID)
	}
}
