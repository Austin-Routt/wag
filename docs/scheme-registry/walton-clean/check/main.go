// walton-check: load a Carneades-4 argumentation-scheme YAML file the way WAG
// does (caesYaml.Import), and check it against the cleanup gate:
//
//  1. import: no YAML or import error, nothing printed to stderr; additionally a
//     strict YAML decode (yaml.v2 UnmarshalStrict) that rejects unknown keys
//     such as `assumption:` or `langugage:` and duplicate map keys;
//  2. the Carneades validator (src/engine/validation.Validate) reports zero
//     problems, both on the imported graph and after step 3;
//  3. every scheme instantiates (InstantiateScheme with distinct dummy
//     constants): one argument per conclusion, each with every premise and
//     assumption of the scheme, one undercutting argument per exception, every
//     statement ground, nothing printed to stderr;
//  4. every language key is "name/arity" and its template has exactly arity
//     %s (and no other % verb); every predicate/arity used in a formula
//     (inside ¬ too) has a template.
//
// The Carneades importer mutates package-level state, so run one file per
// process: walton-check [-json out.json] file.yml
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/carneades/carneades-4/src/engine/caes"
	caesYaml "github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/terms"
	"github.com/carneades/carneades-4/src/engine/validation"
	yaml "gopkg.in/yaml.v2"
)

// strictFile mirrors every key the Carneades importer reads for a scheme file.
type strictFile struct {
	Meta               map[string]interface{} `yaml:"meta"`
	Language           map[string]string      `yaml:"language"`
	Argument_schemes   []strictScheme         `yaml:"argument_schemes"`
	Issue_schemes      map[string][]string    `yaml:"issue_schemes"`
	Statements         map[string]interface{} `yaml:"statements"`
	Issues             map[string]interface{} `yaml:"issues"`
	Arguments          map[string]interface{} `yaml:"arguments"`
	Assumptions        []string               `yaml:"assumptions"`
	Tests              map[string]interface{} `yaml:"tests"`
	References         map[string]interface{} `yaml:"references"`
	Weighing_functions map[string]interface{} `yaml:"weighing_functions"`
}

type strictScheme struct {
	Id          string                 `yaml:"id"`
	Meta        map[string]interface{} `yaml:"meta"`
	Variables   []string               `yaml:"variables"`
	Weight      interface{}            `yaml:"weight"`
	Premises    []string               `yaml:"premises"`
	Assumptions []string               `yaml:"assumptions"`
	Exceptions  []string               `yaml:"exceptions"`
	Deletions   []string               `yaml:"deletions"`
	Guards      []string               `yaml:"guards"`
	Conclusions []string               `yaml:"conclusions"`
}

type schemeResult struct {
	Index       int      `json:"index"`
	Id          string   `json:"id"`
	Params      []string `json:"params"`
	Arguments   int      `json:"arguments"`
	Conclusions int      `json:"conclusions"`
	Premises    int      `json:"premises_expected"`
	Exceptions  int      `json:"exceptions_expected"`
	Problems    []string `json:"problems"`
	Pass        bool     `json:"pass"`
}

type report struct {
	File             string              `json:"file"`
	YAMLError        string              `json:"yaml_error,omitempty"`
	StrictErrors     []string            `json:"strict_errors"`
	ImportError      string              `json:"import_error,omitempty"`
	ImportStderr     string              `json:"import_stderr"`
	Schemes          int                 `json:"schemes"`
	LanguageKeys     int                 `json:"language_keys"`
	ValidatorBefore  []string            `json:"validator_before_instantiation"`
	ValidatorAfter   []string            `json:"validator_after_instantiation"`
	SchemeResults    []schemeResult      `json:"scheme_results"`
	LanguageProblems []string            `json:"language_problems"`
	UnusedVariables  []string            `json:"declared_but_unused_variables"`
	Gate             map[string]bool     `json:"gate"`
	SchemeFailures   map[string][]string `json:"scheme_failures"`
	GlobalProblems   []string            `json:"global_problems"`
	Pass             bool                `json:"pass"`
}

// stderr capture: the Carneades packages write to os.Stderr at call time, so
// pointing os.Stderr at a temp file and reading what was appended after each
// call attributes every message to the call that printed it.
type capture struct {
	f   *os.File
	off int64
}

func newCapture() *capture {
	f, err := os.CreateTemp(".", ".walton-check-stderr-*")
	if err != nil {
		panic(err)
	}
	os.Remove(f.Name())
	return &capture{f: f}
}

func (c *capture) take() string {
	end, _ := c.f.Seek(0, io.SeekEnd)
	if end == c.off {
		return ""
	}
	buf := make([]byte, end-c.off)
	c.f.ReadAt(buf, c.off)
	c.off = end
	return string(buf)
}

var nonIdent = regexp.MustCompile(`[^a-z0-9_]`)

func constFor(i int, v string) string {
	return fmt.Sprintf("k%d_%s", i, nonIdent.ReplaceAllString(strings.ToLower(v), "x"))
}

func problemString(p validation.Problem) string {
	return fmt.Sprintf("[%s] %s: %s: %s", p.Category, p.Id, strings.TrimSpace(p.Description), p.Expression)
}

func sortedProblems(ps []validation.Problem) []string {
	out := []string{}
	for _, p := range ps {
		out = append(out, problemString(p))
	}
	sort.Strings(out)
	return out
}

// usedPredicates returns name/arity for the formula's predicate (inside ¬ as
// well) and for nested compound arguments; isVar reports a bare variable.
func usedPredicates(t terms.Term) (specs []string, nested []string, isBool bool) {
	for {
		c, ok := t.(terms.Compound)
		if ok && c.Functor == "¬" && len(c.Args) == 1 {
			t = c.Args[0]
			continue
		}
		break
	}
	switch t.Type() {
	case terms.VariableType:
		return nil, nil, false
	case terms.BoolType:
		return nil, nil, true
	case terms.AtomType:
		return []string{t.String() + "/0"}, nil, false
	case terms.CompoundType:
		c := t.(terms.Compound)
		specs = []string{c.Functor + "/" + strconv.Itoa(len(c.Args))}
		var walk func(a terms.Term)
		walk = func(a terms.Term) {
			if ac, ok := a.(terms.Compound); ok {
				// ¬ is built in (Language.Apply renders it without a template)
				if !(ac.Functor == "¬" && len(ac.Args) == 1) {
					nested = append(nested, ac.Functor+"/"+strconv.Itoa(len(ac.Args)))
				}
				for _, x := range ac.Args {
					walk(x)
				}
			}
		}
		for _, a := range c.Args {
			walk(a)
		}
		return specs, nested, false
	}
	return nil, nil, false
}

func main() {
	jsonOut := flag.String("json", "", "write the report as JSON to this file")
	dumpOut := flag.String("dump", "", "write the imported schemes and language as JSON to this file")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("usage: walton-check [-json out.json] file.yml")
		os.Exit(2)
	}
	path := flag.Arg(0)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read:", err)
		os.Exit(2)
	}

	realStderr := os.Stderr
	cap := newCapture()
	os.Stderr = cap.f
	defer func() { os.Stderr = realStderr }()

	r := report{File: path, StrictErrors: []string{}, ValidatorBefore: []string{}, ValidatorAfter: []string{},
		LanguageProblems: []string{}, UnusedVariables: []string{}, Gate: map[string]bool{}}

	// ---- 1. import ------------------------------------------------------
	var generic interface{}
	if err := yaml.Unmarshal(data, &generic); err != nil {
		r.YAMLError = err.Error()
	}
	var sf strictFile
	if err := yaml.UnmarshalStrict(data, &sf); err != nil {
		for _, l := range strings.Split(err.Error(), "\n") {
			l = strings.TrimSpace(l)
			if l != "" && l != "yaml: unmarshal errors:" {
				r.StrictErrors = append(r.StrictErrors, l)
			}
		}
	}
	ag, err := caesYaml.Import(bytes.NewReader(data))
	r.ImportStderr = cap.take()
	if err != nil {
		r.ImportError = err.Error()
	}
	r.Gate["1_import"] = r.YAMLError == "" && r.ImportError == "" && r.ImportStderr == "" && len(r.StrictErrors) == 0
	if err != nil || ag == nil || ag.Theory == nil {
		finish(&r, *jsonOut, realStderr)
		return
	}
	theory := ag.Theory
	if *dumpOut != "" {
		type dumpScheme struct {
			Id          string      `json:"id"`
			Title       interface{} `json:"title"`
			Variables   []string    `json:"variables"`
			Premises    []string    `json:"premises"`
			Assumptions []string    `json:"assumptions"`
			Exceptions  []string    `json:"exceptions"`
			Conclusions []string    `json:"conclusions"`
		}
		d := struct {
			Schemes  []dumpScheme      `json:"schemes"`
			Language map[string]string `json:"language"`
		}{Language: theory.Language}
		for _, s := range theory.ArgSchemes {
			d.Schemes = append(d.Schemes, dumpScheme{s.Id, s.Metadata["title"], s.Variables, s.Premises, s.Assumptions, s.Exceptions, s.Conclusions})
		}
		b, _ := json.MarshalIndent(d, "", " ")
		os.WriteFile(*dumpOut, b, 0644)
	}
	r.Schemes = len(theory.ArgSchemes)
	r.LanguageKeys = len(theory.Language)

	// ---- 2a. validator on the imported graph -----------------------------
	vBefore := validation.Validate(ag)
	r.ValidatorBefore = sortedProblems(vBefore)
	if s := cap.take(); s != "" {
		r.ValidatorBefore = append(r.ValidatorBefore, "stderr during validation: "+s)
	}

	// ---- 3. instantiate every scheme ------------------------------------
	allInst := true
	for i, s := range theory.ArgSchemes {
		sr := schemeResult{Index: i + 1, Id: s.Id, Problems: []string{},
			Conclusions: len(s.Conclusions), Premises: len(s.Premises) + len(s.Assumptions), Exceptions: len(s.Exceptions)}
		params := []string{}
		var b terms.Bindings
		for _, v := range s.Variables {
			c := constFor(i+1, v)
			params = append(params, c)
			b = terms.AddBinding(terms.NewVariable(v), terms.Atom(c), b)
		}
		sr.Params = params
		expect := func(f string) (string, bool) {
			t, ok := terms.ReadString(f)
			if !ok {
				return "", false
			}
			return terms.Substitute(t, b).String(), true
		}
		groundCheck := func(kind, id string) {
			t, ok := terms.ReadString(id)
			if !ok {
				sr.Problems = append(sr.Problems, fmt.Sprintf("%s statement %q is not a term", kind, id))
				return
			}
			if !terms.Ground(t, nil) {
				sr.Problems = append(sr.Problems, fmt.Sprintf("%s statement %q is not ground", kind, id))
			}
		}
		textCheck := func(kind string, st *caes.Statement) {
			if strings.Contains(st.Text, "%!") {
				sr.Problems = append(sr.Problems, fmt.Sprintf("%s statement %q renders with a format error: %q", kind, st.Id, st.Text))
			}
		}
		cap.take()
		before := map[string]bool{}
		for id := range ag.Arguments {
			before[id] = true
		}
		ag.InstantiateScheme(s.Id, params)
		if e := cap.take(); e != "" {
			sr.Problems = append(sr.Problems, "stderr: "+strings.TrimSpace(strings.ReplaceAll(e, "\n", " | ")))
		}
		main := []*caes.Argument{}
		for id, a := range ag.Arguments {
			if !before[id] && a.Scheme == s {
				main = append(main, a)
			}
		}
		sort.Slice(main, func(x, y int) bool { return main[x].Id < main[y].Id })
		sr.Arguments = len(main)
		if len(main) == 0 {
			sr.Problems = append(sr.Problems, "no argument built")
		}
		if len(main) != len(s.Conclusions) {
			sr.Problems = append(sr.Problems, fmt.Sprintf("%d arguments built for %d conclusions", len(main), len(s.Conclusions)))
		}
		formulas := append(append([]string{}, s.Premises...), s.Assumptions...)
		for _, f := range formulas {
			if _, ok := expect(f); !ok {
				sr.Problems = append(sr.Problems, fmt.Sprintf("premise/assumption %q does not parse", f))
			}
		}
		concls := map[string]bool{}
		for _, c := range s.Conclusions {
			e, ok := expect(c)
			if !ok {
				sr.Problems = append(sr.Problems, fmt.Sprintf("conclusion %q does not parse", c))
				continue
			}
			concls[e] = true
		}
		for _, a := range main {
			if a.Conclusion == nil {
				sr.Problems = append(sr.Problems, a.Id+": no conclusion")
			} else {
				groundCheck("conclusion", a.Conclusion.Id)
				if !concls[a.Conclusion.Id] {
					sr.Problems = append(sr.Problems, fmt.Sprintf("%s: conclusion %q is not an instance of a scheme conclusion", a.Id, a.Conclusion.Id))
				}
				if t, ok := terms.ReadString(a.Conclusion.Id); ok && t.Type() != terms.AtomType {
					textCheck("conclusion", a.Conclusion)
				}
			}
			if len(a.Premises) != len(formulas) {
				sr.Problems = append(sr.Problems, fmt.Sprintf("%s: %d premises, scheme has %d premises+assumptions", a.Id, len(a.Premises), len(formulas)))
			} else {
				for j, p := range a.Premises {
					e, _ := expect(formulas[j])
					if p.Stmt == nil || p.Stmt.Id != e {
						got := "<nil>"
						if p.Stmt != nil {
							got = p.Stmt.Id
						}
						sr.Problems = append(sr.Problems, fmt.Sprintf("%s: premise %d is %q, expected %q", a.Id, j+1, got, e))
						continue
					}
					groundCheck("premise", p.Stmt.Id)
					textCheck("premise", p.Stmt)
				}
			}
		}
		assumed := map[string]bool{}
		for _, a := range ag.Assumptions {
			assumed[a] = true
		}
		for _, f := range s.Assumptions {
			if e, ok := expect(f); ok && !assumed[e] {
				sr.Problems = append(sr.Problems, fmt.Sprintf("assumption %q was not added to the graph's assumptions", e))
			}
		}
		if len(main) > 0 {
			uc := main[0].Undercutter
			if uc == nil {
				sr.Problems = append(sr.Problems, "no undercutter statement")
			} else {
				if len(uc.Args) != len(s.Exceptions) {
					sr.Problems = append(sr.Problems, fmt.Sprintf("%d exception arguments for %d exceptions", len(uc.Args), len(s.Exceptions)))
				}
				for j, ea := range uc.Args {
					if len(ea.Premises) != 1 || ea.Premises[0].Stmt == nil {
						sr.Problems = append(sr.Problems, "exception argument without its premise")
						continue
					}
					id := ea.Premises[0].Stmt.Id
					if j < len(s.Exceptions) {
						if e, ok := expect(s.Exceptions[j]); !ok || e != id {
							sr.Problems = append(sr.Problems, fmt.Sprintf("exception %d is %q, expected instance of %q", j+1, id, s.Exceptions[j]))
						}
					}
					groundCheck("exception", id)
					textCheck("exception", ea.Premises[0].Stmt)
				}
			}
		}
		for _, f := range s.Exceptions {
			if _, ok := expect(f); !ok {
				sr.Problems = append(sr.Problems, fmt.Sprintf("exception %q does not parse", f))
			}
		}
		sr.Pass = len(sr.Problems) == 0
		if !sr.Pass {
			allInst = false
		}
		r.SchemeResults = append(r.SchemeResults, sr)
	}
	r.Gate["3_instantiate"] = allInst

	// ---- 2b. validator after instantiation ------------------------------
	vAfter := validation.Validate(ag)
	r.ValidatorAfter = sortedProblems(vAfter)
	if s := cap.take(); s != "" {
		r.ValidatorAfter = append(r.ValidatorAfter, "stderr during validation: "+s)
	}
	r.Gate["2_validator"] = len(r.ValidatorBefore) == 0 && len(r.ValidatorAfter) == 0

	// ---- 4. language ------------------------------------------------------
	lp := []string{}
	keys := []string{}
	for k := range theory.Language {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tmpl := theory.Language[k]
		i := strings.LastIndex(k, "/")
		if i < 0 {
			lp = append(lp, fmt.Sprintf("key %q has no /arity", k))
			continue
		}
		name := k[:i]
		arity, err := strconv.Atoi(k[i+1:])
		if err != nil || arity < 0 {
			lp = append(lp, fmt.Sprintf("key %q: arity is not an integer", k))
			continue
		}
		if t, ok := terms.ReadString(name); !ok || t.Type() != terms.AtomType {
			lp = append(lp, fmt.Sprintf("key %q: %q is not a predicate symbol", k, name))
		}
		plain := strings.ReplaceAll(tmpl, "%%", "")
		ns := strings.Count(plain, "%s")
		np := strings.Count(plain, "%")
		if ns != arity || np != arity {
			lp = append(lp, fmt.Sprintf("key %q: template has %d %%s and %d %% for arity %d: %q", k, ns, np, arity, tmpl))
		}
	}
	for _, s := range theory.ArgSchemes {
		declared := map[string]bool{}
		for _, v := range s.Variables {
			declared[v] = false
		}
		check := func(kind string, fs []string) {
			for _, f := range fs {
				t, ok := terms.ReadString(f)
				if !ok {
					lp = append(lp, fmt.Sprintf("%s: %s %q does not parse", s.Id, kind, f))
					continue
				}
				for _, v := range t.OccurVars() {
					if _, d := declared[v.Name]; d {
						declared[v.Name] = true
					}
				}
				specs, nested, isBool := usedPredicates(t)
				if isBool {
					lp = append(lp, fmt.Sprintf("%s: %s %q is read as a boolean constant", s.Id, kind, f))
				}
				for _, sp := range specs {
					if _, ok := theory.Language[sp]; !ok {
						lp = append(lp, fmt.Sprintf("%s: %s %q: no language template %s", s.Id, kind, f, sp))
					}
				}
				for _, sp := range nested {
					if _, ok := theory.Language[sp]; !ok {
						lp = append(lp, fmt.Sprintf("%s: %s %q: nested term %s has no language template", s.Id, kind, f, sp))
					}
				}
			}
		}
		check("premise", s.Premises)
		check("assumption", s.Assumptions)
		check("exception", s.Exceptions)
		check("conclusion", s.Conclusions)
		for _, v := range s.Variables {
			if !declared[v] {
				r.UnusedVariables = append(r.UnusedVariables, s.Id+": "+v)
			}
		}
	}
	r.LanguageProblems = lp
	r.Gate["4_language"] = len(lp) == 0

	// ---- per-scheme view of checks 2-4 ------------------------------------
	ids := map[string]bool{}
	for _, s := range theory.ArgSchemes {
		ids[s.Id] = true
	}
	r.SchemeFailures = map[string][]string{}
	r.GlobalProblems = []string{}
	seen := map[string]bool{}
	for _, p := range append(append([]validation.Problem{}, vBefore...), vAfter...) {
		sid := ""
		if p.Category == validation.SCHEME && ids[p.Id] {
			sid = p.Id
		} else if p.Category == validation.ARGUMENT {
			if a, ok := ag.Arguments[p.Id]; ok && a.Scheme != nil {
				sid = a.Scheme.Id
			}
		}
		msg := "validator: " + problemString(p)
		if seen[msg] {
			continue
		}
		seen[msg] = true
		if sid == "" {
			r.GlobalProblems = append(r.GlobalProblems, msg)
		} else {
			r.SchemeFailures[sid] = append(r.SchemeFailures[sid], msg)
		}
	}
	for _, sr := range r.SchemeResults {
		for _, x := range sr.Problems {
			r.SchemeFailures[sr.Id] = append(r.SchemeFailures[sr.Id], "instantiate: "+x)
		}
	}
	for _, x := range lp {
		if i := strings.Index(x, ": "); i > 0 && ids[x[:i]] {
			r.SchemeFailures[x[:i]] = append(r.SchemeFailures[x[:i]], "language: "+x[i+2:])
		} else {
			r.GlobalProblems = append(r.GlobalProblems, "language: "+x)
		}
	}
	finish(&r, *jsonOut, realStderr)
}

func finish(r *report, jsonOut string, realStderr *os.File) {
	os.Stderr = realStderr
	r.Pass = true
	for _, k := range []string{"1_import", "2_validator", "3_instantiate", "4_language"} {
		if !r.Gate[k] {
			r.Pass = false
		}
	}
	p := func(format string, a ...interface{}) { fmt.Printf(format, a...) }
	p("walton-check %s\n\n", r.File)
	p("== 1. import ==\n")
	if r.YAMLError != "" {
		p("YAML error: %s\n", r.YAMLError)
	}
	if r.ImportError != "" {
		p("import error: %s\n", r.ImportError)
	}
	p("strict YAML decode (unknown or duplicate keys): %d error(s)\n", len(r.StrictErrors))
	for _, e := range r.StrictErrors {
		p("  %s\n", e)
	}
	p("stderr during import: %q\n", r.ImportStderr)
	p("schemes: %d, language keys: %d\n\n", r.Schemes, r.LanguageKeys)
	p("== 2. Carneades validator ==\n")
	p("before instantiation: %d problem(s)\n", len(r.ValidatorBefore))
	for _, x := range r.ValidatorBefore {
		p("  %s\n", x)
	}
	p("after instantiating every scheme: %d problem(s)\n", len(r.ValidatorAfter))
	for _, x := range r.ValidatorAfter {
		p("  %s\n", x)
	}
	p("\n== 3. InstantiateScheme ==\n")
	nfail := 0
	for _, sr := range r.SchemeResults {
		st := "PASS"
		if !sr.Pass {
			st = "FAIL"
			nfail++
		}
		p("%3d %-42s %s  args=%d/%d premises+assumptions=%d exceptions=%d\n", sr.Index, sr.Id, st, sr.Arguments, sr.Conclusions, sr.Premises, sr.Exceptions)
		for _, x := range sr.Problems {
			p("      - %s\n", x)
		}
	}
	p("schemes failing instantiation: %d of %d\n", nfail, len(r.SchemeResults))
	p("\n== 4. language ==\n")
	p("%d problem(s)\n", len(r.LanguageProblems))
	for _, x := range r.LanguageProblems {
		p("  %s\n", x)
	}
	p("\ndeclared variables used in no formula (informational): %d\n", len(r.UnusedVariables))
	for _, x := range r.UnusedVariables {
		p("  %s\n", x)
	}
	p("\n== per scheme (checks 2-4) ==\n")
	failing := []string{}
	for id := range r.SchemeFailures {
		failing = append(failing, id)
	}
	sort.Strings(failing)
	p("schemes passing every check: %d of %d\n", r.Schemes-len(failing), r.Schemes)
	for _, id := range failing {
		p("  FAIL %s (%d problem(s))\n", id, len(r.SchemeFailures[id]))
	}
	p("problems not tied to one scheme (language keys etc.): %d\n", len(r.GlobalProblems))
	p("\n== GATE ==\n")
	for _, k := range []string{"1_import", "2_validator", "3_instantiate", "4_language"} {
		v := "PASS"
		if !r.Gate[k] {
			v = "FAIL"
		}
		p("%-14s %s\n", k, v)
	}
	if r.Pass {
		p("GATE: PASS\n")
	} else {
		p("GATE: FAIL\n")
	}
	if jsonOut != "" {
		b, _ := json.MarshalIndent(r, "", "  ")
		os.WriteFile(jsonOut, b, 0644)
	}
	if !r.Pass {
		os.Exit(1)
	}
}
