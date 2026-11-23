package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ioFixtureJSON(t *testing.T, b *Bundle) []byte {
	t.Helper()
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func ioReplacePayload(t *testing.T, b *Bundle, root string, raw []byte) {
	t.Helper()
	p := &b.Representations[0].Payload
	if err := os.WriteFile(filepath.Join(root, p.Path), raw, 0600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	p.SHA256, p.Bytes = hex.EncodeToString(sum[:]), int64(len(raw))
	b.Claims[0].Mode, b.Claims[0].Text = "paraphrase", "A source excerpt for structural testing."
	b.Claims[0].Selector = Selector{Kind: "whole"}
}

func TestIndependentIOExactIntegerDecode(t *testing.T) {
	b, _ := fixture(t)
	b.Representations[0].Payload.Bytes = 9007199254740993 // 2^53+1, beyond exact float64 integers.
	got, err := Decode(ioFixtureJSON(t, b))
	if err != nil || got.Representations[0].Payload.Bytes != b.Representations[0].Payload.Bytes {
		t.Fatalf("int64 byte length changed during JSON decoding: got=%v err=%v", got, err)
	}
	for _, number := range []string{"1.5", "9223372036854775808", "1e309"} {
		raw := strings.Replace(string(ioFixtureJSON(t, b)), `"bytes":9007199254740993`, `"bytes":`+number, 1)
		if _, err := Decode([]byte(raw)); err == nil {
			t.Fatalf("invalid integer byte length accepted: %s", number)
		}
	}
}

func TestIndependentIOUnicodeScalarIntegrity(t *testing.T) {
	b, _ := fixture(t)
	b.Sources[0].Label = "UNICODE_SENTINEL"
	raw := string(ioFixtureJSON(t, b))
	for _, bad := range []string{`"\ud800"`, `"\udc00"`, `"\ud800\ud800"`} {
		candidate := strings.Replace(raw, `"UNICODE_SENTINEL"`, bad, 1)
		if got, err := Decode([]byte(candidate)); err == nil {
			t.Errorf("unpaired surrogate silently replaced in recorded source: %q -> %q", bad, got.Sources[0].Label)
		}
	}
	for _, good := range []string{`"\ud83d\ude42"`, `"\ufffd"`} {
		candidate := strings.Replace(raw, `"UNICODE_SENTINEL"`, good, 1)
		if _, err := Decode([]byte(candidate)); err != nil {
			t.Errorf("valid Unicode scalar JSON refused: %s: %v", good, err)
		}
	}
}

func TestIndependentIONativeNonfiniteRegion(t *testing.T) {
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		b, root := fixture(t)
		b.Representations[0].Pages = 1
		b.Claims[0].Mode = "paraphrase"
		b.Claims[0].Selector = Selector{Kind: "region", Page: 1, X: bad, Y: 0, Width: 0.5, Height: 0.5}
		r := Validate(b, root)
		if r.Valid || len(r.BindingRevisions) != 0 || len(r.AdmittedBindings) != 0 {
			t.Errorf("nonfinite native coordinate admitted or fingerprinted: %v: %+v", bad, r)
		}
	}
}

func TestIndependentIOPayloadConfinement(t *testing.T) {
	for _, kind := range []string{"parent traversal", "absolute path", "file symlink", "directory symlink", "directory payload"} {
		t.Run(kind, func(t *testing.T) {
			b, root := fixture(t)
			consent(t, b, root, "p1", "reading-use", "consent")
			consent(t, b, root, "p2", "reading-use", "consent")
			outside := t.TempDir()
			raw, err := os.ReadFile(filepath.Join(root, "report.txt"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(outside, "report.txt"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "parent traversal":
				b.Representations[0].Payload.Path = "../report.txt"
			case "absolute path":
				b.Representations[0].Payload.Path = filepath.Join(outside, "report.txt")
			case "file symlink":
				if err := os.Symlink(filepath.Join(outside, "report.txt"), filepath.Join(root, "link.txt")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
				b.Representations[0].Payload.Path = "link.txt"
			case "directory symlink":
				if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
				b.Representations[0].Payload.Path = "linked/report.txt"
			case "directory payload":
				if err := os.Mkdir(filepath.Join(root, "folder"), 0700); err != nil {
					t.Fatal(err)
				}
				b.Representations[0].Payload.Path = "folder"
			}
			r := Validate(b, root)
			if r.Valid || len(r.AdmittedBindings) != 0 || len(r.Consents) != 0 {
				t.Fatalf("invalid payload exposed usable admission: %+v", r)
			}
		})
	}
}

func TestIndependentIOContentIntegrityFailClosed(t *testing.T) {
	b, root := fixture(t)
	consent(t, b, root, "p1", "reading-use", "consent")
	consent(t, b, root, "p2", "reading-use", "consent")
	// Change the actual file without changing its declared revision or hash.
	if err := os.WriteFile(filepath.Join(root, "report.txt"), []byte("The instrument reported 80 °C.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r := Validate(b, root)
	if r.Valid || len(r.AdmittedBindings) != 0 || len(r.Consents) != 0 {
		t.Fatalf("changed retained bytes preserved usable consent: %+v", r)
	}
}

func TestIndependentIOLFSHeaderIsNotAlwaysPointer(t *testing.T) {
	b, root := fixture(t)
	ioReplacePayload(t, b, root, []byte("version https://git-lfs.github.com/spec/v1\nThis is tutorial text explaining the header, not a pointer object.\n"))
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("ordinary retained text was mistaken for an unresolved pointer: %+v", r.Errors)
	}
	ioReplacePayload(t, b, root, []byte("version https://git-lfs.github.com/spec/v1\noid sha256:"+strings.Repeat("a", 64)+"\nsize 123\n"))
	if r := Validate(b, root); r.Valid {
		t.Fatal("well-formed unresolved LFS pointer accepted as content")
	}
}

func TestIndependentIOTextSelectorUsesCodePoints(t *testing.T) {
	b, root := fixture(t)
	ioReplacePayload(t, b, root, []byte("é🙂x\n"))
	b.Claims[0].Mode, b.Claims[0].Text = "quotation", "🙂"
	b.Claims[0].Selector = Selector{Kind: "text", Start: 1, End: 2, Quote: "🙂"}
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("code-point selector refused: %+v", r.Errors)
	}
	b.Claims[0].Selector.Start, b.Claims[0].Selector.End = 2, 6 // UTF-8 byte offsets, invalid code-point range.
	if r := Validate(b, root); r.Valid {
		t.Fatal("byte offsets accepted as Unicode code-point offsets")
	}
}

func TestIndependentIORequiredZeroValuedFields(t *testing.T) {
	b, root := fixture(t)
	ioReplacePayload(t, b, root, []byte{})
	if r := Validate(b, root); !r.Valid {
		t.Fatalf("explicitly empty retained file must remain valid: %+v", r.Errors)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(ioFixtureJSON(t, b), &raw); err != nil {
		t.Fatal(err)
	}
	reps := raw["representations"].([]interface{})
	delete(reps[0].(map[string]interface{})["payload"].(map[string]interface{}), "bytes")
	missingBytes, _ := json.Marshal(raw)
	if _, err := Decode(missingBytes); err == nil {
		t.Error("omitted byte count silently became zero")
	}

	for _, tc := range []struct {
		selector Selector
		field    string
	}{
		{Selector{Kind: "text", Start: 0, End: 1, Quote: "T"}, "start"},
		{Selector{Kind: "text", Start: 0, End: 1, Quote: "T"}, "end"},
		{Selector{Kind: "text", Start: 0, End: 1, Quote: "T"}, "quote"},
		{Selector{Kind: "page", Page: 1}, "page"},
		{Selector{Kind: "time", StartMS: 0, EndMS: 1000}, "start_ms"},
		{Selector{Kind: "time", StartMS: 0, EndMS: 1000}, "end_ms"},
		{Selector{Kind: "region", Page: 1, X: 0, Y: 0, Width: 0.5, Height: 0.5}, "x"},
		{Selector{Kind: "region", Page: 1, X: 0, Y: 0, Width: 0.5, Height: 0.5}, "y"},
		{Selector{Kind: "region", Page: 1, X: 0, Y: 0, Width: 0.5, Height: 0.5}, "width"},
		{Selector{Kind: "region", Page: 1, X: 0, Y: 0, Width: 0.5, Height: 0.5}, "height"},
	} {
		t.Run(tc.selector.Kind+" missing "+tc.field, func(t *testing.T) {
			b, _ := fixture(t)
			b.Claims[0].Mode = "paraphrase"
			b.Claims[0].Selector = tc.selector
			if err := json.Unmarshal(ioFixtureJSON(t, b), &raw); err != nil {
				t.Fatal(err)
			}
			claims := raw["claims"].([]interface{})
			delete(claims[0].(map[string]interface{})["selector"].(map[string]interface{}), tc.field)
			candidate, _ := json.Marshal(raw)
			if _, err := Decode(candidate); err == nil {
				t.Errorf("missing explicit selector %s accepted", tc.field)
			}
		})
	}
}

func TestIndependentIOOptionalEmptyListsMayBeOmitted(t *testing.T) {
	b, root := fixture(t)
	var raw map[string]interface{}
	if err := json.Unmarshal(ioFixtureJSON(t, b), &raw); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"assessments", "dependencies", "events"} {
		delete(raw, field)
	}
	candidate, _ := json.Marshal(raw)
	got, err := Decode(candidate)
	if err != nil {
		t.Fatalf("optional empty lists made mandatory: %v", err)
	}
	if r := Validate(got, root); !r.Valid {
		t.Fatalf("omitted optional lists invalidate bundle: %+v", r.Errors)
	}
}

func TestIndependentIONativeInvalidUTF8(t *testing.T) {
	b, root := fixture(t)
	b.Sources[0].Label = string([]byte{0xff})
	if r := Validate(b, root); r.Valid || len(r.BindingRevisions) != 0 {
		t.Fatalf("native invalid UTF-8 was normalized and fingerprinted: %+v", r)
	}
}
