package agxml

import "github.com/carneades/carneades-4/src/engine/caes"

// This admission check prevents a malformed exporter from becoming an empty
// or disconnected graph. Segmentation joints require a separate reviewed
// adapter; the legacy converter accepts one EDU per ADU only.
func (g *Arggraph) validateProfile() error {
	invalid := func(f string, a ...interface{}) error {
		return caes.NewRunError(caes.ErrInvalidInput, "AGXML: "+f, a...)
	}
	unsupported := func(f string, a ...interface{}) error {
		return caes.NewRunError(caes.ErrUnsupported, "experimental AGXML: "+f, a...)
	}
	if len(g.Adus) == 0 {
		return invalid("at least one ADU is required; EDUs alone are not argument statements")
	}
	if len(g.Joints) != 0 {
		return unsupported("segmentation joints are not supported by this conversion profile")
	}
	types := map[string]string{}
	add := func(id, kind string) error {
		if id == "" || types[id] != "" {
			return invalid("empty or duplicate node/edge ID %q", id)
		}
		types[id] = kind
		return nil
	}
	for _, e := range g.Edus {
		if err := add(e.Id, "edu"); err != nil {
			return err
		}
	}
	for _, a := range g.Adus {
		if err := add(a.Id, "adu"); err != nil {
			return err
		}
	}
	for _, e := range g.Edges {
		switch e.Type {
		case "seg", "sup", "exa", "reb", "und", "add":
		default:
			return unsupported("edge %q type %q is unsupported", e.Id, e.Type)
		}
		if err := add(e.Id, e.Type); err != nil {
			return err
		}
	}
	segIn, segOut := map[string]int{}, map[string]int{}
	isArgument := func(kind string) bool { return kind == "sup" || kind == "exa" || kind == "reb" || kind == "und" }
	for _, e := range g.Edges {
		src, trg := types[e.Src], types[e.Trg]
		if src == "" || trg == "" {
			return invalid("edge %q has an empty or undeclared src/trg", e.Id)
		}
		switch e.Type {
		case "seg":
			if src != "edu" || trg != "adu" {
				return unsupported("segmentation must map an EDU directly to an ADU")
			}
			segIn[e.Trg]++
			segOut[e.Src]++
		case "sup", "exa", "reb":
			if src != "adu" || trg != "adu" {
				return invalid("edge %q must join two ADUs", e.Id)
			}
		case "und", "add":
			if src != "adu" || !isArgument(trg) {
				return invalid("edge %q must join an ADU to an argument edge", e.Id)
			}
		}
	}
	for _, a := range g.Adus {
		if segIn[a.Id] != 1 {
			return unsupported("ADU %q requires exactly one EDU segmentation edge", a.Id)
		}
	}
	for _, e := range g.Edus {
		if segOut[e.Id] != 1 {
			return unsupported("EDU %q requires exactly one ADU segmentation edge", e.Id)
		}
	}
	return nil
}
