package annotate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AnnotatorSystem is an annotator's system prompt: the task, the
// guidelines, and the resources.
func (d *Dataset) AnnotatorSystem() string {
	var s strings.Builder
	s.WriteString("You annotate claims of a Kialo discussion as instances of Walton's argumentation schemes, exactly as the guidelines below say. " +
		"You work alone: there is no other annotator to consult. For each claim, answer with exactly one JSON object, the item in the guidelines' output format " +
		"(section 4, with the version 3 changes of section 5), with its kialo_claim_id, and nothing else.\n\n# Guidelines\n\n")
	s.WriteString(d.Guidelines)
	s.WriteString("\n\n# Resources\n")
	for _, r := range d.Resources {
		fmt.Fprintf(&s, "\n## %s\n\n```\n%s\n```\n", r.Name, strings.TrimSpace(r.Text))
	}
	return s.String()
}

// AnnotatorPrompt is the message that asks for one claim's annotation: the
// discussion, the primary instances the annotator gave the claims before
// it (so that it can target their formulas), and the claim's entry.
func (d *Dataset) AnnotatorPrompt(it *Item, before map[string]Annotation, order []string) string {
	var u strings.Builder
	fmt.Fprintf(&u, "The discussion:\n```json\n%s\n```\n\n", d.Discussion)
	if prior := priorInstances(order, before); prior != "" {
		fmt.Fprintf(&u, "Your annotations so far (primary instances, for targets):\n%s\n", prior)
	}
	var pretty json.RawMessage = it.Raw
	if b, err := json.MarshalIndent(pretty, "", " "); err == nil {
		pretty = b
	}
	fmt.Fprintf(&u, "Annotate this claim:\n```json\n%s\n```\n", pretty)
	return u.String()
}

// parseAnnotation reads the annotation of the claim id from a reply.
func parseAnnotation(reply, id string) (Annotation, error) {
	got, err := jsonObject(reply)
	if err != nil {
		return nil, err
	}
	if _, ok := got["instances"].([]any); !ok {
		return nil, fmt.Errorf("it has no instances: it is not an annotation in the guidelines' format")
	}
	if v, _ := got["kialo_claim_id"].(string); v == "" {
		got["kialo_claim_id"] = id
	} else if v != id {
		return nil, fmt.Errorf("it is the annotation of %s, not %s", v, id)
	}
	return got, nil
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

// primary returns an annotation's primary instance.
func primary(a Annotation) map[string]any {
	insts, _ := a["instances"].([]any)
	var first map[string]any
	for _, x := range insts {
		in, _ := x.(map[string]any)
		if in == nil {
			continue
		}
		if first == nil {
			first = in
		}
		if p, _ := in["primary"].(bool); p {
			return in
		}
	}
	return first
}

// SchemeOf names an annotation's primary scheme and its fit.
func SchemeOf(a Annotation) string {
	in := primary(a)
	if s, ok := in["scheme"].(map[string]any); ok {
		return fmt.Sprintf("%v (%v)", s["walton_id"], s["fit"])
	}
	return "?"
}

// priorInstances lists the primary instances of the claims annotated
// before: id, scheme, premises, conclusion.
func priorInstances(ids []string, done map[string]Annotation) string {
	var b strings.Builder
	for _, id := range ids {
		it := done[id]
		if it == nil {
			continue
		}
		in := primary(it)
		if in == nil {
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
		fmt.Fprintf(&b, "- %s: %s; %s ⇒ %s\n", id, SchemeOf(it), strings.Join(forms, "; "), concl)
	}
	return b.String()
}

// ReviewerSystem is a reviewer's system prompt: it compares independent
// annotations of a claim and decides, field by field, which is right.
func (d *Dataset) ReviewerSystem() string {
	var s strings.Builder
	s.WriteString(`You review annotations of claims of a Kialo discussion as instances of Walton's argumentation schemes. One or more annotators coded each claim independently, under the guidelines below. For one claim at a time you get their annotations, labelled A, B, C… (you are not told who made them). Check each against the guidelines, the scheme collection and the claim's context, and decide, field by field, which is right; with one annotation, decide field by field whether it is right. Do not prefer an annotation because it is longer, more confident, or agrees with another: two annotators can share a mistake.

Answer with exactly one JSON object, and nothing else:
{"kialo_claim_id": "…",
 "fields": {"scheme": {"choice": "A" | "B" | … | "none", "why": "…"},
            "fit": {…}, "role": {…}, "target": {…}, "polarity": {…}, "premises": {…}, "evidence": {…}},
 "adjudicated": { the claim's annotation as it should be, in the guidelines' output format },
 "confidence": 0 to 1,
 "needs_human": true | false,
 "notes": "what a person should check"}
"choice" names the annotation that is right for that field, or "none" if none is (then "adjudicated" has your own). Set needs_human when the guidelines do not settle it, or you are unsure.

# Guidelines

`)
	s.WriteString(d.Guidelines)
	s.WriteString("\n\n# Resources\n")
	for _, r := range d.Resources {
		fmt.Fprintf(&s, "\n## %s\n\n```\n%s\n```\n", r.Name, strings.TrimSpace(r.Text))
	}
	return s.String()
}

// Letters label the annotations a reviewer sees.
const Letters = "ABCDEFGHIJ"

// ReviewerPrompt asks for the review of one claim's annotations, given in
// the order of names (labelled A, B, …).
func (d *Dataset) ReviewerPrompt(it *Item, names []string, anns map[string]Annotation) string {
	var u strings.Builder
	fmt.Fprintf(&u, "The discussion:\n```json\n%s\n```\n\nThe claim:\n```json\n%s\n```\n", d.Discussion, it.Raw)
	for i, n := range names {
		b, _ := json.MarshalIndent(anns[n], "", " ")
		fmt.Fprintf(&u, "\nAnnotation %c:\n```json\n%s\n```\n", Letters[i], b)
	}
	return u.String()
}
