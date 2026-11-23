package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type index struct {
	actors    map[string]Actor
	sources   map[string]Source
	artifacts map[string]Artifact
	reps      map[string]Representation
	claims    map[string]Claim
	targets   map[string]Target
	bindings  map[string]Binding
}

func makeIndex(b *Bundle) index {
	x := index{map[string]Actor{}, map[string]Source{}, map[string]Artifact{}, map[string]Representation{}, map[string]Claim{}, map[string]Target{}, map[string]Binding{}}
	for _, v := range b.Actors {
		x.actors[v.ID] = v
	}
	for _, v := range b.Sources {
		x.sources[v.ID] = v
	}
	for _, v := range b.Artifacts {
		x.artifacts[v.ID] = v
	}
	for _, v := range b.Representations {
		x.reps[v.ID] = v
	}
	for _, v := range b.Claims {
		x.claims[v.ID] = v
	}
	for _, v := range b.Targets {
		x.targets[v.ID] = v
	}
	for _, v := range b.Bindings {
		x.bindings[v.ID] = v
	}
	return x
}

func nonempty(s string) bool { return strings.TrimSpace(s) != "" }
func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}
func digestOK(s string) bool {
	v, err := hex.DecodeString(s)
	return err == nil && len(v) == 32 && s == strings.ToLower(s)
}

// Validate checks record consistency, retained bytes and the finite consent
// protocol. It does not fetch sources, execute evidence content, authenticate
// identities, judge entailment, or alter a live argument graph.
func Validate(b *Bundle, root string) Report {
	r := Report{Format: "wag-evidence-check/0.1", Scope: "record integrity, retained payload identity, selectors and two-party admission for consideration only; no factual or inferential truth certification", Errors: []Issue{}, Warnings: []Issue{}, BindingRevisions: map[string]string{}, Consents: map[string][]string{}, AdmittedBindings: []string{}, DuplicatePayloads: [][]string{}}
	add := func(code, loc, message string) { r.Errors = append(r.Errors, Issue{code, loc, message}) }
	warn := func(code, loc, message string) { r.Warnings = append(r.Warnings, Issue{code, loc, message}) }
	if b == nil {
		add("bundle", "$", "missing bundle")
		return r
	}
	if _, err := json.Marshal(b); err != nil {
		add("model_encoding", "$", err.Error())
		return r
	}
	if err := validText(reflect.ValueOf(b), "$"); err != nil {
		add("model_encoding", "$", err.Error())
		return r
	}
	x := makeIndex(b)
	seen := map[string]bool{}
	id := func(v, loc string) {
		if !nonempty(v) || seen[v] {
			add("id", loc, "missing or duplicate record ID")
		}
		seen[v] = true
	}
	record := func(v Record, loc string) {
		id(v.ID, loc)
		if _, ok := x.actors[v.CreatedBy]; !ok {
			add("actor", loc, "record author is not declared")
		}
		if _, err := time.Parse(time.RFC3339Nano, v.CreatedAt); err != nil {
			add("time", loc, "created_at must be an RFC3339 timestamp")
		}
	}
	if b.Format != Format {
		add("format", "format", "unsupported evidence contract version")
	}
	c := b.Context
	if !nonempty(c.ID) || !nonempty(c.Profile) || c.Policy != Policy {
		add("context", "context", "context/profile must be named and policy must be "+Policy)
	}
	for name, value := range map[string]string{"engine_sha256": c.EngineSHA256, "catalogue_sha256": c.CatalogueSHA256, "nascq_sha256": c.NASCQSHA256} {
		if !digestOK(value) {
			add("pin", "context."+name, "expected lowercase SHA-256 of the named input")
		}
	}
	for _, v := range b.Actors {
		id(v.ID, "actors")
		if !oneOf(v.Kind, "human", "model", "tool") || !nonempty(v.Label) {
			add("actor", v.ID, "actor needs a declared kind and label")
		}
	}
	if len(c.Participants) != 2 || (len(c.Participants) == 2 && c.Participants[0] == c.Participants[1]) {
		add("participants", "context", "policy requires exactly two distinct participants")
	}
	participant := map[string]bool{}
	for _, p := range c.Participants {
		participant[p] = true
		if _, ok := x.actors[p]; !ok {
			add("participants", "context", "participant is not a declared actor")
		}
	}
	for _, v := range b.Sources {
		record(v.Record, "sources")
		if !oneOf(v.Kind, "person", "organization", "document", "device", "model", "unknown") || !nonempty(v.Label) || !nonempty(v.Reference) {
			add("source", v.ID, "source needs kind, label and reference; use unknown explicitly")
		}
	}
	for _, v := range b.Artifacts {
		record(v.Record, "artifacts")
		if _, ok := x.sources[v.SourceID]; !ok {
			add("reference", v.ID, "source_id does not exist")
		}
		if !nonempty(v.Revision) || !nonempty(v.Scope) {
			add("artifact", v.ID, "artifact revision and declared capture scope are required")
		}
		if !oneOf(v.CaptureStatus, "complete", "partial", "failed", "unsupported", "unattempted") {
			add("capture", v.ID, "unknown capture status")
		}
		if v.CaptureStatus == "complete" && (len(v.Omissions) != 0 || v.Failure != "") {
			add("capture", v.ID, "complete capture cannot declare omissions or failure")
		}
		if v.CaptureStatus == "partial" && len(v.Omissions) == 0 {
			add("capture", v.ID, "partial capture must describe known omissions")
		}
		if oneOf(v.CaptureStatus, "failed", "unsupported") && !nonempty(v.Failure) {
			add("capture", v.ID, "failure/unsupported reason is required")
		}
		for _, omission := range v.Omissions {
			if !nonempty(omission) {
				add("capture", v.ID, "empty omission description")
			}
		}
	}
	originals := map[string]int{}
	payloads := map[string][]string{}
	for _, v := range b.Representations {
		record(v.Record, "representations")
		a, ok := x.artifacts[v.ArtifactID]
		if !ok {
			add("reference", v.ID, "artifact_id does not exist")
		}
		if oneOf(a.CaptureStatus, "failed", "unsupported", "unattempted") {
			add("capture", v.ID, "uncaptured artifact cannot own a retained representation")
		}
		if !oneOf(v.Kind, "original", "ocr", "transcript", "translation", "excerpt", "normalized") {
			add("representation", v.ID, "unsupported representation kind")
		}
		if !nonempty(v.Method) {
			add("derivation", v.ID, "capture/transformation method must be recorded")
		}
		if v.Kind == "original" {
			originals[v.ArtifactID]++
			if len(v.ParentIDs) != 0 {
				add("derivation", v.ID, "original bytes have no transformation parents")
			}
		} else if len(v.ParentIDs) == 0 {
			add("derivation", v.ID, "derivative must name retained parent representations")
		}
		parents := map[string]bool{}
		for _, p := range v.ParentIDs {
			parent, ok := x.reps[p]
			if !ok {
				add("reference", v.ID, "transformation parent missing")
			} else if parent.ArtifactID != v.ArtifactID {
				add("derivation", v.ID, "this bounded profile requires parents from the same artifact revision")
			}
			if parents[p] {
				add("derivation", v.ID, "duplicate transformation parent")
			}
			parents[p] = true
		}
		if v.Pages < 0 || v.DurationMS < 0 {
			add("bounds", v.ID, "declared page count/duration cannot be negative")
		}
		if err := checkPayload(root, v.Payload); err != nil {
			add("payload", v.ID, err.Error())
		}
		payloads[v.Payload.SHA256] = append(payloads[v.Payload.SHA256], v.ID)
	}
	for _, a := range b.Artifacts {
		n := originals[a.ID]
		if (a.CaptureStatus == "complete" && n != 1) || n > 1 {
			add("capture", a.ID, "complete capture requires exactly one original representation; multiple originals require separate artifact records")
		}
	}
	colors := map[string]int{}
	var visit func(string) bool
	visit = func(k string) bool {
		if colors[k] == 1 {
			return false
		}
		if colors[k] == 2 {
			return true
		}
		colors[k] = 1
		for _, p := range x.reps[k].ParentIDs {
			if _, ok := x.reps[p]; ok && !visit(p) {
				return false
			}
		}
		colors[k] = 2
		return true
	}
	for _, v := range b.Representations {
		if !visit(v.ID) {
			add("derivation_cycle", v.ID, "transformation ancestry is cyclic")
			break
		}
	}
	for _, v := range b.Claims {
		record(v.Record, "claims")
		rep, ok := x.reps[v.RepresentationID]
		if !ok {
			add("reference", v.ID, "claim representation does not exist")
		}
		if _, ok := x.sources[v.AttributedTo]; !ok {
			add("reference", v.ID, "attribution source is missing; declare unknown explicitly if needed")
		}
		if !oneOf(v.Mode, "quotation", "paraphrase", "observation", "reconstruction") || !nonempty(v.Text) || !nonempty(v.Scope) {
			add("claim", v.ID, "claim needs a mode, text and scope")
		}
		if ok {
			if err := checkSelector(v.Selector, rep, root); err != nil {
				add("selector", v.ID, err.Error())
			}
		}
		if v.Mode == "quotation" && (v.Selector.Kind != "text" || v.Text != v.Selector.Quote) {
			add("quotation", v.ID, "exact quotation requires matching text selector; OCR/translation remains attributed to that derivative")
		}
		if oneOf(v.Selector.Kind, "page", "time", "region") {
			warn("declared_bounds", v.ID, "only declared page/duration/normalized bounds checked; extraction and substantive interpretation require review")
		}
	}
	for _, v := range b.Targets {
		record(v.Record, "targets")
		if !nonempty(v.ArgumentID) || !digestOK(v.ArgumentRevision) || !nonempty(v.Proposition) || !oneOf(v.Kind, "premise", "assumption", "exception", "conclusion", "inference", "source", "annotation") {
			add("target", v.ID, "target needs argument identity/revision, locus and proposition")
		}
	}
	for _, v := range b.Bindings {
		record(v.Record, "bindings")
		if _, ok := x.claims[v.ClaimID]; !ok {
			add("reference", v.ID, "claim_id does not exist")
		}
		if _, ok := x.targets[v.TargetID]; !ok {
			add("reference", v.ID, "target_id does not exist")
		}
		if !oneOf(v.Role, "support", "oppose", "context", "definition", "source_challenge") || !nonempty(v.Interpretation) || !nonempty(v.Warrant) {
			add("binding", v.ID, "binding needs role, offered interpretation and explicit inferential/use warrant")
		}
	}
	for _, v := range b.Assessments {
		record(v.Record, "assessments")
		if _, ok := x.bindings[v.BindingID]; !ok {
			add("reference", v.ID, "assessment binding missing")
		}
		if !nonempty(v.Rationale) || len(v.Dimensions) == 0 {
			add("assessment", v.ID, "assessment must state dimensions and rationale")
		}
		for dim, status := range v.Dimensions {
			if !oneOf(dim, "authenticity", "relevance", "competence", "observation_access", "sincerity", "probative_force", "independence") || !oneOf(status, "supported", "challenged", "unknown", "not_applicable") {
				add("assessment", v.ID, "unknown assessment dimension/status")
			}
		}
		if !oneOf(v.ReviewLevel, "unreviewed", "model_checked", "human_checked") {
			add("assessment", v.ID, "unknown review level")
		}
		if v.ReviewLevel == "human_checked" && x.actors[v.CreatedBy].Kind != "human" {
			add("review_level", v.ID, "human_checked requires a declared human author; identity still requires external attestation")
		}
		for _, p := range v.BasisClaimIDs {
			if _, ok := x.claims[p]; !ok {
				add("reference", v.ID, "assessment basis claim missing")
			}
		}
		f := v.FactCheck
		if !oneOf(f.Status, "not_checked", "supported", "contradicted", "inconclusive") {
			add("fact_check", v.ID, "unknown fact-check status")
		}
		if f.Status == "not_checked" && (f.Method != "" || len(f.BasisClaimIDs) != 0) {
			add("fact_check", v.ID, "not_checked cannot claim a check method or results")
		}
		if f.Status != "not_checked" && (!nonempty(f.Method) || (oneOf(f.Status, "supported", "contradicted") && len(f.BasisClaimIDs) == 0)) {
			add("fact_check", v.ID, "checked status needs method and, for a determination, cited basis")
		}
		for _, p := range f.BasisClaimIDs {
			if _, ok := x.claims[p]; !ok {
				add("reference", v.ID, "fact-check basis claim missing")
			}
		}
	}
	for _, v := range b.Dependencies {
		record(v.Record, "dependencies")
		if !oneOf(v.Relation, "common_origin", "derived", "independence_claim", "unknown") || !nonempty(v.Rationale) || len(v.ArtifactIDs) < 2 {
			add("dependency", v.ID, "dependency needs at least two artifacts, relation and rationale")
		}
		group := map[string]bool{}
		for _, a := range v.ArtifactIDs {
			if _, ok := x.artifacts[a]; !ok {
				add("reference", v.ID, "dependency artifact missing")
			}
			if group[a] {
				add("dependency", v.ID, "dependency repeats one artifact")
			}
			group[a] = true
		}
		if v.Relation == "independence_claim" {
			warn("independence_claim", v.ID, "independence is a recorded claim requiring review, not a certified fact or a probability multiplier")
		}
	}
	for _, ids := range payloads {
		if len(ids) > 1 {
			sort.Strings(ids)
			r.DuplicatePayloads = append(r.DuplicatePayloads, ids)
		}
	}
	sort.Slice(r.DuplicatePayloads, func(i, j int) bool { return r.DuplicatePayloads[i][0] < r.DuplicatePayloads[j][0] })
	if len(r.DuplicatePayloads) > 0 {
		warn("shared_payload", "representations", "identical bytes do not supply independent corroboration")
	}
	for _, v := range b.Events {
		record(v.Record, "events")
	}
	if len(r.Errors) != 0 {
		return r
	}
	for _, v := range b.Bindings {
		digest, err := bindingDigest(b.Context, v, x)
		if err != nil {
			add("fingerprint", v.ID, err.Error())
			return r
		}
		r.BindingRevisions[v.ID] = digest
	}
	state := map[string]map[string]bool{}
	for i, v := range b.Events {
		if v.Sequence != i+1 {
			add("event_order", v.ID, "events must form the complete sequence starting at 1")
			continue
		}
		if !participant[v.CreatedBy] {
			add("authority", v.ID, "only a participant may grant/withdraw their own consent")
			continue
		}
		if !oneOf(v.Type, "consent", "withdraw") || v.Purpose != "consideration" {
			add("event", v.ID, "unsupported event or purpose; no factual acceptance follows from admission")
			continue
		}
		digest, ok := r.BindingRevisions[v.BindingID]
		if !ok || v.BindingRevision != digest {
			add("stale_binding", v.ID, "consent must pin the entire exact binding/context/source/target revision")
			continue
		}
		if state[v.BindingID] == nil {
			state[v.BindingID] = map[string]bool{}
		}
		has := state[v.BindingID][v.CreatedBy]
		if (v.Type == "consent" && has) || (v.Type == "withdraw" && !has) {
			add("consent_transition", v.ID, "duplicate consent or withdrawal without that actor's active consent")
			continue
		}
		state[v.BindingID][v.CreatedBy] = v.Type == "consent"
	}
	// Invalid logs never expose a partly replayed admission as usable output.
	if len(r.Errors) != 0 {
		return r
	}
	for _, v := range b.Bindings {
		ids := []string{}
		for _, a := range c.Participants {
			if state[v.ID][a] {
				ids = append(ids, a)
			}
		}
		sort.Strings(ids)
		r.Consents[v.ID] = ids
		if len(ids) == 2 {
			r.AdmittedBindings = append(r.AdmittedBindings, v.ID)
		}
	}
	sort.Strings(r.AdmittedBindings)
	if len(b.Targets) != 0 {
		warn("external_target", "targets", "internal target references checked; correspondence to a live graph requires its adapter")
	}
	r.Valid = true
	return r
}

func bindingDigest(context Context, b Binding, x index) (string, error) {
	claim := x.claims[b.ClaimID]
	reps := []Representation{}
	artifacts := map[string]Artifact{}
	sources := map[string]Source{claim.AttributedTo: x.sources[claim.AttributedTo]}
	seen := map[string]bool{}
	var walk func(string)
	walk = func(k string) {
		if seen[k] {
			return
		}
		seen[k] = true
		rep := x.reps[k]
		reps = append(reps, rep)
		a := x.artifacts[rep.ArtifactID]
		artifacts[a.ID] = a
		sources[a.SourceID] = x.sources[a.SourceID]
		for _, p := range rep.ParentIDs {
			walk(p)
		}
	}
	walk(claim.RepresentationID)
	sort.Slice(reps, func(i, j int) bool { return reps[i].ID < reps[j].ID })
	// encoding/json sorts string map keys. This named version pins Go JSON
	// serialization; it is not advertised as an independent canonical-JSON standard.
	value := struct {
		Algorithm       string
		Context         Context
		Binding         Binding
		Claim           Claim
		Target          Target
		Representations []Representation
		Artifacts       map[string]Artifact
		Sources         map[string]Source
	}{"wag-binding-go-json/1", context, b, claim, x.targets[b.TargetID], reps, artifacts, sources}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func validText(v reflect.Value, path string) error {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			return validText(v.Elem(), path)
		}
	case reflect.String:
		if !utf8.ValidString(v.String()) || strings.IndexByte(v.String(), 0) >= 0 {
			return fmt.Errorf("%s: invalid UTF-8 or NUL in record text", path)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := validText(v.Field(i), path+"."+v.Type().Field(i).Name); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := validText(v.Index(i), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if err := validText(iter.Key(), path); err != nil {
				return err
			}
			if err := validText(iter.Value(), path); err != nil {
				return err
			}
		}
	}
	return nil
}

func safePath(root, rel string) (string, error) {
	if !filepath.IsLocal(rel) || strings.Contains(rel, "\\") {
		return "", fmt.Errorf("payload path is not portable and local")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	path := abs
	for _, part := range strings.Split(filepath.ToSlash(filepath.Clean(rel)), "/") {
		path = filepath.Join(path, part)
		info, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("payload path contains a symbolic link")
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("payload is not a regular file")
	}
	return path, nil
}

func checkPayload(root string, p Payload) error {
	if !digestOK(p.SHA256) || p.Bytes < 0 {
		return fmt.Errorf("payload requires SHA-256 and a nonnegative byte length")
	}
	if _, _, err := mime.ParseMediaType(p.MediaType); err != nil || !strings.Contains(p.MediaType, "/") {
		return fmt.Errorf("payload media type is invalid")
	}
	f, err := openPayload(root, p.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	prefix := make([]byte, 1025)
	n, err := f.Read(prefix)
	if err != nil && err != io.EOF {
		return err
	}
	if n <= 1024 && isLFSPointer(prefix[:n]) {
		return fmt.Errorf("unresolved Git LFS pointer is not a retained payload")
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return err
	}
	if size != p.Bytes || hex.EncodeToString(h.Sum(nil)) != p.SHA256 {
		return fmt.Errorf("retained payload hash/length mismatch")
	}
	return nil
}

func openPayload(root, rel string) (*os.File, error) {
	if _, err := safePath(root, rel); err != nil {
		return nil, err
	}
	// Root.Open confines lookup even if a path component changes after Lstat.
	// Inputs must still be immutable for a consumer to reuse a prior receipt.
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	f, err := dir.Open(rel)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, fmt.Errorf("opened payload is not a regular file")
	}
	return f, nil
}

func isLFSPointer(raw []byte) bool {
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "version https://git-lfs.github.com/spec/v1" {
		return false
	}
	oid, size := false, false
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "oid sha256:"):
			if oid || !digestOK(strings.TrimPrefix(line, "oid sha256:")) {
				return false
			}
			oid = true
		case strings.HasPrefix(line, "size "):
			n, err := strconv.ParseInt(strings.TrimPrefix(line, "size "), 10, 64)
			if size || err != nil || n < 0 {
				return false
			}
			size = true
		case strings.HasPrefix(line, "ext-"):
			// Extension lines are retained only to recognize an unresolved pointer.
		default:
			return false
		}
	}
	return oid && size
}

func checkSelector(s Selector, r Representation, root string) error {
	if !oneOf(s.Kind, "whole", "text", "page", "time", "region") {
		return fmt.Errorf("unknown selector kind")
	}
	if s.Kind == "text" {
		if s.Start < 0 || s.End <= s.Start || !nonempty(s.Quote) {
			return fmt.Errorf("text selector needs nonempty quote and ordered code-point offsets")
		}
		media, _, _ := mime.ParseMediaType(r.Payload.MediaType)
		if !strings.HasPrefix(media, "text/") {
			return fmt.Errorf("text selector requires a retained text representation")
		}
		f, err := openPayload(root, r.Payload.Path)
		if err != nil {
			return err
		}
		raw, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return err
		}
		// Bind the selector to the exact same bytes whose identity is checked,
		// rather than trusting a prior hash followed by a separate unchecked read.
		sum := sha256.Sum256(raw)
		if int64(len(raw)) != r.Payload.Bytes || hex.EncodeToString(sum[:]) != r.Payload.SHA256 {
			return fmt.Errorf("selector payload hash/length mismatch")
		}
		if !utf8.Valid(raw) {
			return fmt.Errorf("text selector needs valid UTF-8; retain normalization as a derivative")
		}
		chars := []rune(string(raw))
		if s.End > int64(len(chars)) || string(chars[s.Start:s.End]) != s.Quote {
			return fmt.Errorf("text selector does not match retained code points")
		}
	}
	if oneOf(s.Kind, "page", "region") && (s.Page < 1 || r.Pages < 1 || s.Page > r.Pages) {
		return fmt.Errorf("page selector outside declared page bounds")
	}
	if s.Kind == "time" && (s.StartMS < 0 || s.EndMS <= s.StartMS || r.DurationMS <= 0 || s.EndMS > r.DurationMS) {
		return fmt.Errorf("time selector outside declared duration")
	}
	if s.Kind == "region" && (s.X < 0 || s.Y < 0 || s.Width <= 0 || s.Height <= 0 || s.X+s.Width > 1 || s.Y+s.Height > 1) {
		return fmt.Errorf("region must be positive and within normalized page bounds")
	}
	// Unused selector coordinates must be empty; silently carrying stale fields
	// would make the quoted region ambiguous across readers.
	if s.Kind != "text" && (s.Start != 0 || s.End != 0 || s.Quote != "") {
		return fmt.Errorf("text fields on a non-text selector")
	}
	if !oneOf(s.Kind, "page", "region") && s.Page != 0 {
		return fmt.Errorf("page field on another selector")
	}
	if s.Kind != "time" && (s.StartMS != 0 || s.EndMS != 0) {
		return fmt.Errorf("time fields on another selector")
	}
	if s.Kind != "region" && (s.X != 0 || s.Y != 0 || s.Width != 0 || s.Height != 0) {
		return fmt.Errorf("region fields on another selector")
	}
	return nil
}
