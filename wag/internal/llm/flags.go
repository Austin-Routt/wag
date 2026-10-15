package llm

import (
	"fmt"
	"strings"
)

// Specs is a repeatable command-line flag of profiles (ParseSpec).
type Specs []string

func (s *Specs) String() string     { return strings.Join(*s, " ") }
func (s *Specs) Set(v string) error { *s = append(*s, v); return nil }

// ProfilesFrom reads the profiles of a profiles file (if file is not "")
// and of command-line specs; a later one replaces an earlier one of the
// same name.
func ProfilesFrom(file string, specs []string) ([]Profile, error) {
	var out []Profile
	if file != "" {
		ps, err := LoadProfiles(file)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", file, err)
		}
		out = ps
	}
	for _, s := range specs {
		p, err := ParseSpec(s)
		if err != nil {
			return nil, fmt.Errorf("--model %s: %v", s, err)
		}
		replaced := false
		for i := range out {
			if out[i].Name == p.Name {
				out[i], replaced = p, true
			}
		}
		if !replaced {
			out = append(out, p)
		}
	}
	return out, nil
}

// Notice is what a person should know before a profile sends anything:
// "" if nothing.
func (p *Profile) Notice() string {
	switch p.Provider {
	case "gemini":
		return "On the Gemini API's unpaid (free) tier, Google may use the prompts and answers to improve its products, " +
			"and people may review them (Gemini API Additional Terms): send only public, non-confidential material. " +
			"Free-tier rate limits are low: set min_interval (e.g. 6s for 10 requests a minute)."
	case "command":
		return "The command provider runs " + strings.Join(p.Command, " ") + " for every call."
	}
	return ""
}
