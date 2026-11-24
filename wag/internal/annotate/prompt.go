package annotate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AnnotatorSystem is an annotator's system prompt: the task, the
// guidelines, and the resources.
func (d *Dataset) AnnotatorSystem() string {
	return d.annotatorSystem(func(s string) string { return s })
}

// annotatorSystem is AnnotatorSystem with the files' text converted (their
// line endings: a run made from a Windows checkout pinned that prompt).
func (d *Dataset) annotatorSystem(conv func(string) string) string {
	var s strings.Builder
	s.WriteString("You annotate claims of a Kialo discussion as instances of Walton's argumentation schemes, exactly as the guidelines below say. " +
		"You work alone: there is no other annotator to consult. For each claim, answer with exactly one JSON object, the item in the guidelines' output format " +
		"(section 4, with the version 3 changes of section 5), with its kialo_claim_id, and nothing else.\n\n# Guidelines\n\n")
	s.WriteString(conv(d.Guidelines))
	s.WriteString("\n\n# Resources\n")
	for _, r := range d.Resources {
		fmt.Fprintf(&s, "\n## %s\n\n```\n%s\n```\n", r.Name, strings.TrimSpace(conv(r.Text)))
	}
	return s.String()
}

// noCaseRecord ends each claim's prompt in a format-2 run on a dataset
// without a case record: the guidelines bind agents and actions by one
// (docs/annotation-records.md §3). It is not in the system prompt, which a
// run pins: a dataset's default format follows its gold, which scoring may
// remove.
const noCaseRecord = "\nThis dataset has no case record (case.json): bind agents and actions to the atoms the claims name, and give no instance a context.\n"

// samePrompt tells whether a pinned hash is the annotator's system prompt,
// its files' line endings read either way.
func (d *Dataset) samePrompt(pinned string) bool {
	lf := func(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
	crlf := func(s string) string { return strings.ReplaceAll(lf(s), "\n", "\r\n") }
	return pinned == sha([]byte(d.annotatorSystem(lf))) || pinned == sha([]byte(d.annotatorSystem(crlf)))
}

// AnnotatorPrompt is the message that asks for one claim's annotation: the
// discussion, the instances the annotator gave the claims before it, each
// by its id (so that it can target any of them and their formulas exactly;
// in format 1, their primary instances), and the claim's entry.
func (d *Dataset) AnnotatorPrompt(it *Item, before map[string]Annotation, order []string) string {
	var u strings.Builder
	fmt.Fprintf(&u, "The discussion:\n```json\n%s\n```\n\n", d.Discussion)
	if prior := priorInstances(order, before); prior != "" {
		if strings.Contains(prior, "#") {
			fmt.Fprintf(&u, "Your annotations so far (every instance, by its id, for targets):\n%s\n", prior)
		} else {
			fmt.Fprintf(&u, "Your annotations so far (primary instances, for targets):\n%s\n", prior)
		}
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
	if err := checkedJSON([]byte(s[i:j+1]), &m); err != nil {
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

// priorInstances lists the instances of the claims annotated before. An
// annotation whose instances have ids (format 2) is listed instance by
// instance: id, primary or not, scheme, premises, conclusion; one without
// (format 1) by its primary instance, under its claim's id.
func priorInstances(ids []string, done map[string]Annotation) string {
	var b strings.Builder
	line := func(in map[string]any) string {
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
		return strings.Join(forms, "; ") + " ⇒ " + concl
	}
	for _, id := range ids {
		it := done[id]
		if it == nil {
			continue
		}
		if !hasInstanceIDs(it) {
			if in := primary(it); in != nil {
				fmt.Fprintf(&b, "- %s: %s; %s\n", id, SchemeOf(it), line(in))
			}
			continue
		}
		insts, _ := it["instances"].([]any)
		for _, x := range insts {
			in, _ := x.(map[string]any)
			if in == nil {
				continue
			}
			s, _ := in["scheme"].(map[string]any)
			tag := ""
			if p, _ := in["primary"].(bool); p {
				tag = ", primary"
			}
			fmt.Fprintf(&b, "- %v (%v (%v)%s): %s\n", in["id"], s["walton_id"], s["fit"], tag, line(in))
		}
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
            "fit": {…}, "role": {…}, "target": {…}, "polarity": {…}, "premises": {…}, "evidence": {…}, "instances": {…}},
 "adjudicated": { the claim's annotation as it should be, in the guidelines' output format },
 "confidence": 0 to 1,
 "needs_human": true | false,
 "notes": "what a person should check"}
"choice" names the annotation that is right for that field, or "none" if none is (then "adjudicated" has your own). "role" includes the question answered, "target" the target's instance and formula, "premises" the bindings, and "instances" how the claim is divided into inferences. In "adjudicated", a target's instance is named as in the annotation you follow for the target. Set needs_human when the guidelines do not settle it, or you are unsure.

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
