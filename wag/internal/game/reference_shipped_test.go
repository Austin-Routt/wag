package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/wag/data"
)

// The new golden set (step C5.2's step 7; Replay-2 of the external review of
// 26 November): the reference games' table played on the data the game ships
// (package data) with the default profile, which every game made since step
// C5.2 has. The file's first line pins the data's fingerprint and the
// profile, so that a change of either is reported as such before any game
// is compared; the 33 games on the frozen data (TestReferenceDigests) stay as
// they were.
//
// WAG_UPDATE_REFERENCE=1 go test ./internal/game -run TestReferenceDigestsShipped
// rewrites it; read the diff before committing it.

// shippedFingerprint is the shipped data's fingerprint, as the server computes
// it (server.Data.Hash: the four files, LF line endings, each followed by a
// zero byte).
func shippedFingerprint() string {
	h := sha256.New()
	for _, b := range [][]byte{data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON} {
		h.Write(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func TestReferenceDigestsShipped(t *testing.T) {
	const path = "testdata/reference-digests-shipped.txt"
	prof, err := json.Marshal(DefaultProfile())
	if err != nil {
		t.Fatal(err)
	}
	got := []string{fmt.Sprintf("data %s profile %s", shippedFingerprint(), prof)}
	for _, g := range referenceGames() {
		got = append(got, g.playOn(t, data.CoreYAML, data.SchemesJSON, data.CQsJSON, data.EdgesJSON, DefaultProfile()))
	}
	if os.Getenv("WAG_UPDATE_REFERENCE") != "" {
		if err := os.WriteFile(path, []byte(strings.Join(got, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d reference games on the shipped data to %s", len(got)-1, path)
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (WAG_UPDATE_REFERENCE=1 writes it)", err)
	}
	want := strings.Split(strings.TrimRight(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n"), "\n")
	if want[0] != got[0] {
		t.Fatalf("the golden set was made with other data or another profile:\n got  %s\n want %s\n"+
			"(a change of the data's identity rewrites it: WAG_UPDATE_REFERENCE=1, and read the diff)", got[0], want[0])
	}
	if len(want) != len(got) {
		t.Fatalf("%d reference games stored, %d played", len(want)-1, len(got)-1)
	}
	differ := 0
	for i := 1; i < len(got); i++ {
		if got[i] != want[i] {
			differ++
			t.Errorf("game %d differs from the reference\n got  %s\n want %s", i, got[i], want[i])
		}
	}
	if differ == 0 {
		t.Logf("%d reference games on the shipped data reproduced", len(got)-1)
	}
}
