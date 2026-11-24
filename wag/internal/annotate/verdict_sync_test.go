package annotate

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A verdict whose file is in place but whose folder could not be synced is
// the person's verdict (the review's EV6): the run keeps it, as its file and
// a reopened run do, says that it may not survive a power loss yet, and syncs
// the folder again when asked to (Flush) or at the next verdict.

// savedVerdicts reads review/human.json: claim → choice.
func savedVerdicts(t *testing.T, r *Run) map[string]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(r.Dir, "review", "human.json"))
	if err != nil {
		t.Fatal(err)
	}
	var list []Verdict
	if err := json.Unmarshal(b, &list); err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, v := range list {
		out[v.Item] = v.Choice
	}
	return out
}

func TestAVerdictInPlaceIsKeptUntilItsFolderIsSynced(t *testing.T) {
	r, _, items := soloRun(t, 1)
	if err := r.SetVerdict(Verdict{Item: items[0], Choice: "solo"}); err != nil {
		t.Fatal(err)
	}
	heal, _ := failSync(t, filepath.Join(r.Dir, "review"))
	err := r.SetVerdict(Verdict{Item: items[0], Choice: "none", Note: "neither"})
	var nd *NotDurableError
	if !errors.As(err, &nd) || !strings.Contains(err.Error(), "could not be synced") || !strings.Contains(err.Error(), "power loss") {
		t.Fatalf("%v", err)
	}
	if v := r.verdicts[items[0]]; v.Choice != "none" || v.Note != "neither" {
		t.Fatalf("the run holds %+v, not the verdict its file holds", v)
	}
	if got := savedVerdicts(t, r); got[items[0]] != "none" {
		t.Fatalf("human.json holds %v", got)
	}
	again, err := Open(r.Dir, func(string, string) (*Dataset, error) { return r.Dataset, nil })
	if err != nil {
		t.Fatal(err)
	}
	if v := again.verdicts[items[0]]; v.Choice != "none" {
		t.Fatalf("a reopened run holds %q", v.Choice)
	}
	if st := r.Status(); !strings.Contains(st.VerdictsNotSynced, "could not be synced") || st.Verdicts != 1 {
		t.Fatalf("status: %q, %d verdicts", st.VerdictsNotSynced, st.Verdicts)
	}
	// Asked to write its records, the run syncs the folder again: while it
	// still cannot, the verdicts stay not synced.
	if err := r.Flush(); err == nil || !strings.Contains(err.Error(), "could not be synced") {
		t.Fatalf("flush while the folder cannot be synced: %v", err)
	}
	if r.Status().VerdictsNotSynced == "" {
		t.Fatal("the verdicts are said to be synced")
	}
	heal()
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	if st := r.Status(); st.VerdictsNotSynced != "" || st.WriteError != "" {
		t.Fatalf("after the sync: %q %q", st.VerdictsNotSynced, st.WriteError)
	}
}

// The next verdict syncs the folder again too.
func TestTheNextVerdictSyncsTheVerdictsFolderAgain(t *testing.T) {
	r, _, items := soloRun(t, 2)
	review := filepath.Join(r.Dir, "review")
	heal, synced := failSync(t, review)
	var nd *NotDurableError
	if err := r.SetVerdict(Verdict{Item: items[0], Choice: "solo"}); !errors.As(err, &nd) {
		t.Fatalf("%v", err)
	}
	heal()
	*synced = nil
	if err := r.SetVerdict(Verdict{Item: items[1], Choice: "unsure"}); err != nil {
		t.Fatal(err)
	}
	if st := r.Status(); st.VerdictsNotSynced != "" || st.Verdicts != 2 {
		t.Fatalf("status: %q, %d verdicts", st.VerdictsNotSynced, st.Verdicts)
	}
	if got := savedVerdicts(t, r); got[items[0]] != "solo" || got[items[1]] != "unsure" {
		t.Fatalf("human.json holds %v", got)
	}
	if len(*synced) == 0 || (*synced)[0] != review {
		t.Fatalf("synced %v", *synced)
	}
}

// A verdict that is not in its file is still not kept, whatever the folder.
func TestAVerdictNotInItsFileIsNotKeptAndNotPending(t *testing.T) {
	r, _, items := soloRun(t, 1)
	unblock := block(t, filepath.Join(r.Dir, "review", "human.json.tmp"))
	err := r.SetVerdict(Verdict{Item: items[0], Choice: "solo"})
	var nd *NotDurableError
	if err == nil || errors.As(err, &nd) || !strings.Contains(err.Error(), "not saved") {
		t.Fatalf("%v", err)
	}
	if st := r.Status(); st.Verdicts != 0 || st.VerdictsNotSynced != "" {
		t.Fatalf("status: %q, %d verdicts", st.VerdictsNotSynced, st.Verdicts)
	}
	unblock()
}

// The reviewer's file is in review/ too: its write syncs the folder, and the
// verdicts' entry with it.
func TestTheReviewersWriteSyncsTheVerdictsFolderToo(t *testing.T) {
	r, _ := reviewRun(t, 1, goodReview)
	heal, _ := failSync(t, filepath.Join(r.Dir, "review"))
	var nd *NotDurableError
	if err := r.SetVerdict(Verdict{Item: r.Settings.Items[0], Choice: "solo"}); !errors.As(err, &nd) {
		t.Fatalf("%v", err)
	}
	heal()
	if err := r.Review(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if st := r.Status(); st.VerdictsNotSynced != "" || st.Reviewed != 1 || st.Verdicts != 1 {
		t.Fatalf("after the review: %q, %d reviewed, %d verdicts", st.VerdictsNotSynced, st.Reviewed, st.Verdicts)
	}
}
