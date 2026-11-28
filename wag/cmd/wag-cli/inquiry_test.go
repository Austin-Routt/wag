package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// C2 at the terminal (docs/e1-evidence.md): an inquiry of one player is
// refused, and so is one that loses on a late move (C6). The terminal plays
// an inquiry to its end, and a series of them, naming no winner (C1).
func TestC2TheTerminalRefusesAnInquiryOfOnePlayer(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "wag-cli")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	out, err := exec.Command(bin, "--mode", "inquiry", "--p1", "random", "--p2", "none", "--turns", "2").CombinedOutput()
	if err == nil || !strings.Contains(string(out), "an inquiry is two players") {
		t.Fatalf("an inquiry of one player: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "--mode", "inquiry", "--p1", "random", "--p2", "greedy", "--move-time", "60s", "--timeout-loses").CombinedOutput()
	if err == nil || !strings.Contains(string(out), "late move") {
		t.Fatalf("C6: an inquiry that loses on a late move: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "--mode", "inquiry", "--p1", "random", "--p2", "greedy", "--turns", "3").CombinedOutput()
	if s := string(out); err != nil || !strings.Contains(s, "GAME OVER") || !strings.Contains(s, "Verdict:") ||
		!strings.Contains(s, "no one wins") || strings.Contains(s, "Winner") || strings.Contains(s, "TIE") {
		t.Fatalf("C1: an inquiry at the terminal: %v\n%s", err, out)
	}
	out, err = exec.Command(bin, "--mode", "inquiry", "--p1", "random", "--p2", "greedy", "--turns", "3", "--games", "2").CombinedOutput()
	if s := string(out); err != nil || !strings.Contains(s, "verdict") || strings.Contains(s, "wins") || strings.Contains(s, "ties") {
		t.Fatalf("C1: a series of inquiries at the terminal: %v\n%s", err, out)
	}
}

// U4 (docs/c5-item-f.md): the terminal refuses a claim in the engine's
// undercutter namespace before the game starts.
func TestU4TheTerminalRefusesAClaimInTheUndercuttersNamespace(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "wag-cli")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	for _, claim := range []string{"undercut(a11)", "¬undercut(a11)"} {
		out, err := exec.Command(bin, "--claim", claim, "--p1", "random", "--p2", "greedy", "--turns", "2").CombinedOutput()
		if err == nil || !strings.Contains(string(out), "names the engine's undercutter of an argument") {
			t.Fatalf("%s: %v\n%s", claim, err, out)
		}
	}
	out, err := exec.Command(bin, "--claim", "worth_finding_out_first(tax_meat,demand_falls)", "--p1", "random", "--p2", "greedy", "--turns", "2").CombinedOutput()
	if err != nil || strings.Contains(string(out), "undercutter") {
		t.Fatalf("another claim: %v\n%s", err, out)
	}
}
