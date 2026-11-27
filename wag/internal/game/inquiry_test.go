package game

import (
	"strings"
	"testing"
)

// C2 and C6 (docs/e1-evidence.md), at the game, which the server and the
// terminal both ask: an inquiry is two players, and takes no setting that
// makes a winner or a loser.
func TestC2C6TheInquirysCheck(t *testing.T) {
	cfg, err := ModeConfig("inquiry")
	if err != nil || !cfg.Cooperative || cfg.Objective != ObjectiveClaim || cfg.AbstentionThreshold != 1 {
		t.Fatalf("the inquiry's configuration: %+v, %v", cfg, err)
	}
	for n, ok := range map[int]bool{1: false, 2: true, 3: false} {
		if err := CheckInquiry(cfg, n); (err == nil) != ok {
			t.Errorf("C2: %d players: %v", n, err)
		}
	}
	for name, c := range map[string]struct {
		edit func(*GameConfig)
		says string
	}{
		"a target score":        {func(c *GameConfig) { c.TargetScore = 50 }, "target score"},
		"an answer deadline":    {func(c *GameConfig) { c.AnswerWithin = 1 }, "answer deadline"},
		"conceding on a pass":   {func(c *GameConfig) { c.ConcedeOnPass = true }, "concedes"},
		"losing on a late move": {func(c *GameConfig) { c.MoveTimeoutLoses = true }, "late move"},
	} {
		c2 := cfg
		c.edit(&c2)
		if err := CheckInquiry(c2, 2); err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("C6, %s: %v", name, err)
		}
	}
	debate, _ := ModeConfig("debate")
	if err := CheckInquiry(debate, 3); err != nil {
		t.Errorf("a debate is checked as an inquiry: %v", err)
	}
}
