package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/llm"
)

// annotate has a model annotate Kialo claims as Walton scheme instances,
// under the batch's guidelines, one claim per call, top-down, and writes
// the annotations in the annotators' format (docs/kialo-batch2/guidelines.md,
// §4), so that validate.py checks them and agreement.py compares them with
// gold.json or another annotator. Every call is traced (calls.jsonl), and
// run.json records the profile and the hashes of everything the model saw.
// A run that stops can be resumed: annotated claims are skipped.
func annotate(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("annotate", flag.ExitOnError)
	specs, file := profilesFlags(fs)
	profile := fs.String("annotator", "", "the profile that annotates (default: the only one given)")
	items := fs.String("items", "../docs/kialo-batch2/items.json", "the claims to annotate (items.json)")
	guidelines := fs.String("guidelines", "../docs/kialo-batch2/guidelines.md", "the annotation guidelines")
	var contexts llm.Specs
	fs.Var(&contexts, "context", "a file the annotator is given besides the guidelines (repeatable; default: walton.yml, the lexicon, the game's schemes, questions and edges)")
	out := fs.String("out", "", "the directory for annotator-NAME.json, calls.jsonl and run.json")
	limit := fs.Int("limit", 0, "annotate at most this many claims (0: all)")
	_ = fs.Parse(args)
	if *out == "" {
		return fmt.Errorf("annotate needs --out")
	}
	ps, err := llm.ProfilesFrom(*file, *specs)
	if err != nil {
		return err
	}
	var p *llm.Profile
	for i := range ps {
		if ps[i].Name == *profile || *profile == "" && len(ps) == 1 {
			p = &ps[i]
		}
	}
	if p == nil {
		return fmt.Errorf("give one profile (--model), or name it with --annotator")
	}
	if n := p.Notice(); n != "" {
		fmt.Fprintln(os.Stderr, "Note:", n)
	}
	c, err := llm.NewClient(*p)
	if err != nil {
		return err
	}
	if len(contexts) == 0 {
		contexts = llm.Specs{"../docs/scheme-registry/walton-clean/walton.yml", "../docs/kialo-batch2/lexicon.json",
			"data/nascq/schemes.json", "data/nascq/critical_questions.json", "data/nascq/edges.json"}
	}
	system, hashes, err := annotationSystem(*guidelines, contexts)
	if err != nil {
		return err
	}
	var in struct {
		Discussion json.RawMessage              `json:"discussion"`
		Items      []map[string]json.RawMessage `json:"items"`
	}
	b, err := os.ReadFile(*items)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &in); err != nil {
		return fmt.Errorf("%s: %v", *items, err)
	}
	hashes[*items] = sha(b)
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	name := strings.NewReplacer("/", "_", ":", "_").Replace(p.Name)
	annPath := filepath.Join(*out, "annotator-"+name+".json")
	done := map[string]map[string]any{}
	if b, err := os.ReadFile(annPath); err == nil { // resume
		var prev struct {
			Items []map[string]any `json:"items"`
		}
		if json.Unmarshal(b, &prev) == nil {
			for _, it := range prev.Items {
				if id, _ := it["kialo_claim_id"].(string); id != "" {
					done[id] = it
				}
			}
		}
	}
	calls, err := os.OpenFile(filepath.Join(*out, "calls.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer calls.Close()
	run := map[string]any{"format": "wag-annotation-run/1", "started": time.Now().UTC(), "profile": p, "inputs_sha256": hashes,
		"system_prompt_sha256": sha([]byte(system)), "failures": map[string]string{}}
	failures := run["failures"].(map[string]string)
	var order []string
	n := 0
	for _, it := range in.Items {
		var id string
		_ = json.Unmarshal(it["kialo_claim_id"], &id)
		order = append(order, id)
		if done[id] != nil || ctx.Err() != nil || *limit > 0 && n >= *limit {
			continue
		}
		n++
		item, err := annotateOne(ctx, c, system, in.Discussion, it, id, order[:len(order)-1], done, func(a llm.Attempt, attempt int) {
			line, _ := json.Marshal(map[string]any{"format": "wag-annotation-call/1", "item": id, "attempt": attempt, "call": a})
			fmt.Fprintln(calls, string(line))
		})
		if err != nil {
			failures[id] = err.Error()
			fmt.Printf("%s: failed: %v\n", id, err)
			if e, ok := err.(*llm.Error); ok && !e.Retryable() && e.Kind != llm.KindBadReply {
				break // a bad key or a refused request will not pass on the next claim
			}
			continue
		}
		done[id] = item
		fmt.Printf("%s: %v\n", id, schemeOf(item))
		if err := writeAnnotations(annPath, name, in.Items, done); err != nil {
			return err
		}
	}
	run["finished"], run["annotated"] = time.Now().UTC(), len(done)
	rb, _ := json.MarshalIndent(run, "", " ")
	if err := os.WriteFile(filepath.Join(*out, "run.json"), rb, 0o644); err != nil {
		return err
	}
	fmt.Printf("%d of %d claims annotated, in %s. Check them: python3 -I ../docs/kialo-batch2/validate.py %s; compare: python3 -I ../docs/kialo-batch2/agreement.py %s ../docs/kialo-batch2/gold.json\n",
		len(done), len(in.Items), annPath, annPath, annPath)
	return ctx.Err()
}

func sha(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// annotationSystem is the annotator's system prompt: the task, the
// guidelines, and the context files.
func annotationSystem(guidelines string, files []string) (string, map[string]string, error) {
	hashes := map[string]string{}
	g, err := os.ReadFile(guidelines)
	if err != nil {
		return "", nil, err
	}
	hashes[guidelines] = sha(g)
	var s strings.Builder
	s.WriteString("You annotate claims of a Kialo discussion as instances of Walton's argumentation schemes, exactly as the guidelines below say. " +
		"You work alone: there is no other annotator to consult. For each claim, answer with exactly one JSON object, the item in the guidelines' output format " +
		"(section 4, with the version 3 changes of section 5), with its kialo_claim_id, and nothing else.\n\n# Guidelines\n\n")
	s.Write(g)
	s.WriteString("\n\n# Resources\n")
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return "", nil, err
		}
		hashes[f] = sha(b)
		fmt.Fprintf(&s, "\n## %s\n\n```\n%s\n```\n", filepath.Base(f), strings.TrimSpace(string(b)))
	}
	return s.String(), hashes, nil
}

// annotateOne asks for one claim's annotation, with the discussion's thesis,
// the claim's entry of items.json, and the primary instances annotated so
// far (to target their formulas); a reply that is not one JSON object for
// the claim is sent back once with the reason.
func annotateOne(ctx context.Context, c *llm.Client, system string, discussion json.RawMessage, item map[string]json.RawMessage, id string,
	before []string, done map[string]map[string]any, trace func(llm.Attempt, int)) (map[string]any, error) {
	ib, _ := json.MarshalIndent(item, "", " ")
	var u strings.Builder
	fmt.Fprintf(&u, "The discussion:\n```json\n%s\n```\n\n", discussion)
	if prior := priorInstances(before, done); prior != "" {
		fmt.Fprintf(&u, "Your annotations so far (primary instances, for targets):\n%s\n", prior)
	}
	fmt.Fprintf(&u, "Annotate this claim:\n```json\n%s\n```\n", ib)
	msgs := []llm.Message{{Role: "user", Content: u.String()}}
	for attempt := 1; attempt <= 2; attempt++ {
		resp, err := c.Complete(ctx, c.Request(system, msgs), func(a llm.Attempt) { trace(a, attempt) })
		if err != nil {
			return nil, err
		}
		got, perr := jsonObject(resp.Text)
		if perr == nil {
			if _, ok := got["instances"].([]any); !ok {
				perr = fmt.Errorf("it has no instances: it is not an annotation in the guidelines' format")
			}
		}
		if perr == nil {
			if v, _ := got["kialo_claim_id"].(string); v == "" {
				got["kialo_claim_id"] = id
			} else if v != id {
				perr = fmt.Errorf("it is the annotation of %s, not %s", v, id)
			}
		}
		if perr == nil {
			return got, nil
		}
		if attempt == 2 {
			return nil, &llm.Error{Provider: c.Provider.Name(), Kind: llm.KindBadReply, Message: perr.Error()}
		}
		msgs = append(msgs, llm.Message{Role: "assistant", Content: resp.Text},
			llm.Message{Role: "user", Content: "That reply cannot be used: " + perr.Error() + ". Reply with the one JSON object for " + id + " only."})
	}
	return nil, nil
}

// jsonObject reads the JSON object a reply holds (the whole reply, or from
// its first { to its last }).
func jsonObject(s string) (map[string]any, error) {
	i, j := strings.Index(s, "{"), strings.LastIndex(s, "}")
	if i < 0 || j < i {
		return nil, fmt.Errorf("there is no JSON object in it")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s[i:j+1]), &m); err != nil {
		return nil, fmt.Errorf("its JSON does not parse: %v", err)
	}
	return m, nil
}

// priorInstances lists the primary instances of the claims annotated
// before: id, scheme, premises, conclusion.
func priorInstances(ids []string, done map[string]map[string]any) string {
	var b strings.Builder
	for _, id := range ids {
		it := done[id]
		if it == nil {
			continue
		}
		insts, _ := it["instances"].([]any)
		for _, x := range insts {
			in, _ := x.(map[string]any)
			if p, _ := in["primary"].(bool); !p && len(insts) > 1 {
				continue
			}
			var forms []string
			ps, _ := in["premises"].([]any)
			for _, y := range ps {
				if pm, ok := y.(map[string]any); ok {
					forms = append(forms, fmt.Sprintf("%v %v", pm["role"], pm["formula"]))
				}
			}
			concl := ""
			if cm, ok := in["conclusion"].(map[string]any); ok {
				concl = fmt.Sprint(cm["formula"])
			}
			fmt.Fprintf(&b, "- %s: %v; %s ⇒ %s\n", id, schemeOf(it), strings.Join(forms, "; "), concl)
			break
		}
	}
	return b.String()
}

func schemeOf(item map[string]any) string {
	insts, _ := item["instances"].([]any)
	for _, x := range insts {
		in, _ := x.(map[string]any)
		if p, _ := in["primary"].(bool); p || len(insts) == 1 {
			if s, ok := in["scheme"].(map[string]any); ok {
				return fmt.Sprintf("%v (%v)", s["walton_id"], s["fit"])
			}
		}
	}
	return "?"
}

// writeAnnotations writes the annotations so far, in items.json order.
func writeAnnotations(path, name string, items []map[string]json.RawMessage, done map[string]map[string]any) error {
	pos := map[string]int{}
	for i, it := range items {
		var id string
		_ = json.Unmarshal(it["kialo_claim_id"], &id)
		pos[id] = i
	}
	var list []map[string]any
	for _, it := range done {
		list = append(list, it)
	}
	sort.Slice(list, func(i, j int) bool {
		a, _ := list[i]["kialo_claim_id"].(string)
		b, _ := list[j]["kialo_claim_id"].(string)
		return pos[a] < pos[b]
	})
	b, err := json.MarshalIndent(map[string]any{"annotator": name, "items": list}, "", " ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
