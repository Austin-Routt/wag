package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yourorg/wag/internal/llm"
)

// budgetFlags are an invocation's budget (D046, amended by the owner on
// 2026-11-23). No limit is the default, and --unlimited says so explicitly.
// --max-calls caps the client attempts exactly, retries included;
// --max-tokens and --max-cost refuse a call by an estimate, not a billing
// ceiling. Whatever the limits, the invocation counts its calls and the usage
// providers report, prints them, and records them in its output folder
// (finish). The limits cover the clients of one invocation: a resumed
// annotation run starts a new allowance, and adds its own record.
type budgetFlags struct {
	calls, tokens *int
	cost          *float64
	prices        *string
	unlimited     *bool
	pricesSHA256  *string // of the price table as it was read
}

func addBudgetFlags(fs *flag.FlagSet) budgetFlags {
	return budgetFlags{
		calls:        fs.Int("max-calls", 0, "stop this invocation before its N+1th client attempt, counting retries (D046; default: no limit)"),
		tokens:       fs.Int("max-tokens", 0, "refuse a call whose estimated reservation exceeds N accounted tokens; not a provider billing ceiling (D046; default: no limit)"),
		cost:         fs.Float64("max-cost", 0, "refuse a call whose estimated price-table cost exceeds this invocation allowance; not a billing ceiling (D046; default: no limit)"),
		prices:       fs.String("prices", "", `JSON rates per million tokens, with explicit input/output rates, by profile or model ("*" fallback): {"m":{"input":0.3,"output":2.5}}; with or without --max-cost, the spending is priced`),
		unlimited:    fs.Bool("unlimited", false, "no limit, said explicitly: no limit is already the default, and this drops its note; not with --max-calls, --max-tokens or --max-cost (D046)"),
		pricesSHA256: new(string),
	}
}

// decodePrices requires complete, unambiguous rates. Null, omitted, duplicated
// or misspelled rate fields must not become a silent zero or overwrite a rate.
func decodePrices(raw []byte) (map[string]llm.Price, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return nil, fmt.Errorf("prices must be an object keyed by profile/model")
	}
	prices := map[string]llm.Price{}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return nil, err
		}
		name, ok := token.(string)
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("price name must be nonempty text")
		}
		if _, exists := prices[name]; exists {
			return nil, fmt.Errorf("duplicate price name %q", name)
		}
		start, err := d.Token()
		if err != nil || start != json.Delim('{') {
			return nil, fmt.Errorf("price %q needs explicit input and output rates", name)
		}
		rates := map[string]float64{}
		for d.More() {
			token, err := d.Token()
			if err != nil {
				return nil, err
			}
			field, ok := token.(string)
			if !ok || (field != "input" && field != "output") {
				return nil, fmt.Errorf("price %q has an unknown rate field %q", name, field)
			}
			if _, exists := rates[field]; exists {
				return nil, fmt.Errorf("price %q repeats rate %q", name, field)
			}
			token, err = d.Token()
			if err != nil {
				return nil, err
			}
			number, ok := token.(json.Number)
			if !ok {
				return nil, fmt.Errorf("price %q %s must be a finite nonnegative number", name, field)
			}
			value, err := number.Float64()
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				return nil, fmt.Errorf("price %q %s must be a finite nonnegative number", name, field)
			}
			rates[field] = value
		}
		if _, err := d.Token(); err != nil {
			return nil, err
		}
		if len(rates) != 2 {
			return nil, fmt.Errorf("price %q needs explicit input and output rates", name)
		}
		prices[name] = llm.Price{Input: rates["input"], Output: rates["output"]}
	}
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing data after prices")
	}
	return prices, nil
}

// keyed reports whether a profile uses an API key: its provider needs one, or
// the profile names one (a keyed compatible endpoint, a keyed override of a
// local provider). A mock needs none.
func keyed(p llm.Profile) bool {
	if p.Provider == "mock" {
		return false
	}
	return llm.NeedsKey(p.Provider) || strings.TrimSpace(p.APIKeyEnv) != "" || strings.TrimSpace(p.APIKey) != ""
}

// budget builds the invocation's budget. It always returns one: without a
// limit it refuses nothing, and still counts.
func (f budgetFlags) budget() (*llm.Budget, error) {
	b := &llm.Budget{MaxCalls: *f.calls, MaxTokens: *f.tokens, MaxCost: *f.cost}
	if err := b.Check(); err != nil {
		return nil, err
	}
	if *f.unlimited && b.Limited() {
		return nil, fmt.Errorf("--unlimited contradicts --max-calls, --max-tokens and --max-cost: give one or the other (D046)")
	}
	if *f.prices != "" {
		raw, err := os.ReadFile(*f.prices)
		if err != nil {
			return nil, err
		}
		if b.Prices, err = decodePrices(raw); err != nil {
			return nil, fmt.Errorf("--prices %s: %v", *f.prices, err)
		}
		sum := sha256.Sum256(raw)
		*f.pricesSHA256 = hex.EncodeToString(sum[:])
	}
	if err := b.Check(); err != nil {
		return nil, err
	}
	if b.MaxCost > 0 && len(b.Prices) == 0 {
		return nil, fmt.Errorf("--max-cost needs --prices")
	}
	return b, nil
}

// note says, once, that profiles with an API key run with no limit, by
// default: such calls may be billed. It is not given under a limit, nor with
// --unlimited, which says no limit explicitly.
func (f budgetFlags) note(w io.Writer, b *llm.Budget, used []llm.Profile) {
	if b.Limited() || *f.unlimited {
		return
	}
	seen := map[string]bool{}
	var names []string
	for _, p := range used {
		if keyed(p) && !seen[p.Name] {
			seen[p.Name] = true
			names = append(names, p.Name)
		}
	}
	if len(names) == 0 {
		return
	}
	sort.Strings(names)
	who := "profile " + names[0] + " uses an API key and runs"
	if len(names) > 1 {
		who = "profiles " + strings.Join(names, ", ") + " use an API key and run"
	}
	fmt.Fprintf(w, "Note: %s with no limit, D046's default. --max-calls, --max-tokens or --max-cost set one; --unlimited says so explicitly and drops this note.\n", who)
}

// budgetRecord is one invocation's line in budget.jsonl, in its output
// folder: its limits, or none and whether by default, and what it used.
type budgetRecord struct {
	Command   string `json:"command"`
	Started   string `json:"started"`
	Ended     string `json:"ended"`
	Unlimited bool   `json:"unlimited"`
	// Default: no limit because none was given (D046's default), not because
	// of --unlimited.
	Default      bool      `json:"default"`
	MaxCalls     int       `json:"max_calls,omitempty"`
	MaxTokens    int       `json:"max_tokens,omitempty"`
	MaxCost      float64   `json:"max_cost,omitempty"`
	Prices       string    `json:"prices,omitempty"`        // the price table's file name, without its folder
	PricesSHA256 string    `json:"prices_sha256,omitempty"` // and its content's hash
	Spent        llm.Spent `json:"spent"`
	Refused      bool      `json:"refused"`
	Halted       string    `json:"halted,omitempty"`
}

// finish prints what the invocation used and appends its record to
// dir/budget.jsonl. An invocation that made no folder called no model, and
// is not recorded.
func (f budgetFlags) finish(w io.Writer, dir, command string, b *llm.Budget, started time.Time) {
	reportBudget(w, b, *f.unlimited)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return
	}
	rec := budgetRecord{Command: command, Started: started.UTC().Format(time.RFC3339), Ended: time.Now().UTC().Format(time.RFC3339),
		Unlimited: !b.Limited(), Default: !b.Limited() && !*f.unlimited, MaxCalls: b.MaxCalls, MaxTokens: b.MaxTokens, MaxCost: b.MaxCost,
		Spent: b.Spent(), Refused: b.Refused(), Halted: b.Halted()}
	if *f.prices != "" {
		rec.Prices, rec.PricesSHA256 = filepath.Base(*f.prices), *f.pricesSHA256
	}
	line, err := json.Marshal(rec)
	if err == nil {
		var fh *os.File
		if fh, err = os.OpenFile(filepath.Join(dir, "budget.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			_, err = fh.Write(append(line, '\n'))
			if cerr := fh.Close(); err == nil {
				err = cerr
			}
		}
	}
	if err != nil {
		fmt.Fprintln(w, "The budget's record was not written:", err)
	}
}

// reportBudget prints what the invocation used, and against which limits.
func reportBudget(w io.Writer, b *llm.Budget, explicit bool) {
	if b == nil {
		return
	}
	s := b.Spent()
	estimated := b.MaxTokens > 0 || b.MaxCost > 0
	if !b.Limited() {
		how := "the default"
		if explicit {
			how = "--unlimited"
		}
		fmt.Fprintf(w, "Invocation budget: no limit (%s); %s, %d reported tokens", how, plural(s.Calls, "client attempt"), s.Tokens)
	} else {
		fmt.Fprintf(w, "Invocation budget: %s", plural(s.Calls, "client attempt"))
		if b.MaxCalls > 0 {
			fmt.Fprintf(w, " of %d", b.MaxCalls)
		}
		if estimated {
			fmt.Fprintf(w, ", %d accounted tokens", s.Tokens)
		} else {
			fmt.Fprintf(w, ", %d reported tokens", s.Tokens)
		}
		if b.MaxTokens > 0 {
			fmt.Fprintf(w, " of %d", b.MaxTokens)
		}
	}
	switch {
	case b.MaxCost > 0:
		fmt.Fprintf(w, ", estimated price-table cost %.6f of %.6f", s.Cost, b.MaxCost)
	case len(b.Prices) > 0:
		fmt.Fprintf(w, ", price-table cost %.6f", s.Cost)
	}
	if s.Unreported > 0 {
		if estimated {
			fmt.Fprintf(w, "; %s reported no usage, counted at the estimate", plural(s.Unreported, "attempt"))
		} else {
			fmt.Fprintf(w, "; %s reported no usage, not counted in the tokens", plural(s.Unreported, "attempt"))
		}
	}
	if estimated {
		fmt.Fprint(w, "; token and cost limits are estimates, not billing ceilings")
	}
	if h := b.Halted(); h != "" {
		fmt.Fprintf(w, "; stopped: %s", h)
	}
	fmt.Fprintln(w, ".")
}

// plural counts a thing: "1 client attempt", "2 client attempts".
func plural(n int, noun string) string {
	if n != 1 {
		noun += "s"
	}
	return fmt.Sprintf("%d %s", n, noun)
}
