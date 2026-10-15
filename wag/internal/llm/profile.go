package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Profile is a configured model: which provider, which model, how it is
// called. A profiles file (LoadProfiles) holds several; a seat names one.
type Profile struct {
	Name     string `json:"name"`
	Provider string `json:"provider"` // see Providers
	// BaseURL is the endpoint, for the OpenAI-compatible providers (their
	// defaults are in Providers) and to point an SDK provider elsewhere.
	BaseURL string `json:"base_url,omitempty"`
	// APIKeyEnv is the name of the environment variable that holds the key
	// (the provider's usual one if empty). The key itself is never written
	// in a profile.
	APIKeyEnv string `json:"api_key_env,omitempty"`
	// APIKey is a key given another way than the environment (typed in the
	// dashboard). It is never written out: not to a profiles file, a trace,
	// an export or a run's record (json:"-").
	APIKey string `json:"-"`
	// Command runs the "command" provider: the program and its arguments.
	Command []string `json:"command,omitempty"`
	// System replaces the seat's system prompt (to try other wordings).
	System string `json:"system,omitempty"`
	// SchemaFile is a JSON Schema file for Output "schema".
	SchemaFile string `json:"schema_file,omitempty"`
	Params
	// Timeout bounds one call (default 5 minutes).
	Timeout Duration `json:"timeout,omitempty"`
	// MinInterval spaces the calls of this profile: at least this long from
	// one call's start to the next (a free tier's requests per minute).
	MinInterval Duration `json:"min_interval,omitempty"`
	// Retries is how many times a call that failed for a reason that may
	// pass (a rate limit, a server error, a timeout) is tried again
	// (default 3; -1 for none).
	Retries int `json:"retries,omitempty"`
	// Replies are the mock provider's answers, in turn.
	Replies []string `json:"replies,omitempty"`
}

// Duration is a time.Duration written "30s" or "5m" (or a number of
// seconds) in JSON.
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(time.Duration(d).String()) }

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		v, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*d = Duration(v)
		return nil
	}
	var n float64
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("a duration is a string such as \"30s\" or a number of seconds")
	}
	*d = Duration(time.Duration(n * float64(time.Second)))
	return nil
}

// Endpoint is an OpenAI-compatible service's defaults.
type Endpoint struct {
	BaseURL   string
	KeyEnv    string // "" for a local engine, which needs no key
	NeedsKey  bool
	MaxTokens string // the request field for the output limit
	Note      string
}

// Providers are the provider names a profile can use, with what each is.
var Providers = map[string]string{
	"gemini":            "Gemini Developer API, through Google's GenAI SDK (google.golang.org/genai); key GEMINI_API_KEY (or GOOGLE_API_KEY)",
	"anthropic":         "the Claude API, through Anthropic's SDK; key ANTHROPIC_API_KEY or an `ant auth login` profile",
	"openai":            "OpenAI's Chat Completions API; key OPENAI_API_KEY",
	"deepseek":          "DeepSeek's API (OpenAI-compatible); key DEEPSEEK_API_KEY",
	"openai-compatible": "any OpenAI-compatible endpoint: give base_url (and api_key_env if it needs a key)",
	"lmstudio":          "LM Studio's local server",
	"llamacpp":          "llama.cpp's llama-server (its default port, 8080, is also wag-server's: run one of them elsewhere)",
	"sglang":            "SGLang's server",
	"vllm":              "vLLM's server",
	"ollama":            "Ollama's OpenAI-compatible endpoint",
	"command":           "an external program: the request as JSON on its standard input, the answer on its standard output (Guidance, DSPy, a script of your own)",
	"mock":              "scripted answers (replies), for tests and dry runs; no network",
}

// endpoints are the OpenAI-compatible providers' defaults.
var endpoints = map[string]Endpoint{
	"openai":            {BaseURL: "https://api.openai.com/v1", KeyEnv: "OPENAI_API_KEY", NeedsKey: true, MaxTokens: "max_completion_tokens"},
	"deepseek":          {BaseURL: "https://api.deepseek.com", KeyEnv: "DEEPSEEK_API_KEY", NeedsKey: true, MaxTokens: "max_tokens"},
	"openai-compatible": {MaxTokens: "max_tokens"},
	"lmstudio":          {BaseURL: "http://localhost:1234/v1", MaxTokens: "max_tokens"},
	"llamacpp":          {BaseURL: "http://localhost:8080/v1", MaxTokens: "max_tokens"},
	"sglang":            {BaseURL: "http://localhost:30000/v1", MaxTokens: "max_tokens"},
	"vllm":              {BaseURL: "http://localhost:8000/v1", MaxTokens: "max_tokens"},
	"ollama":            {BaseURL: "http://localhost:11434/v1", MaxTokens: "max_tokens"},
}

// ProviderNames lists the providers, sorted.
func ProviderNames() []string {
	out := make([]string, 0, len(Providers))
	for k := range Providers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Check reports what is wrong with the profile, before any call.
func (p *Profile) Check() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("a profile needs a name")
	}
	if _, ok := Providers[p.Provider]; !ok {
		return fmt.Errorf("profile %s: unknown provider %q (%s)", p.Name, p.Provider, strings.Join(ProviderNames(), ", "))
	}
	switch p.Provider {
	case "command":
		if len(p.Command) == 0 {
			return fmt.Errorf("profile %s: the command provider needs command", p.Name)
		}
	case "mock":
	case "openai-compatible":
		if p.BaseURL == "" {
			return fmt.Errorf("profile %s: openai-compatible needs base_url", p.Name)
		}
		fallthrough
	default:
		if strings.TrimSpace(p.Model) == "" {
			return fmt.Errorf("profile %s: no model", p.Name)
		}
	}
	switch p.Output {
	case "", "text", "json":
	case "schema":
		if len(p.Schema) == 0 && p.SchemaFile == "" {
			return fmt.Errorf("profile %s: output schema needs schema_file", p.Name)
		}
	default:
		return fmt.Errorf("profile %s: output is text, json or schema, not %q", p.Name, p.Output)
	}
	if p.APIKeyEnv != "" && strings.ContainsAny(p.APIKeyEnv, " =") {
		return fmt.Errorf("profile %s: api_key_env is the name of an environment variable, not a key", p.Name)
	}
	return nil
}

// KeyEnv returns the environment variables the profile's key is read from,
// in order.
func (p *Profile) KeyEnv() []string {
	if p.APIKeyEnv != "" {
		return []string{p.APIKeyEnv}
	}
	switch p.Provider {
	case "gemini":
		return []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"}
	case "anthropic":
		return []string{"ANTHROPIC_API_KEY"}
	}
	if e, ok := endpoints[p.Provider]; ok && e.KeyEnv != "" {
		return []string{e.KeyEnv}
	}
	return nil
}

// key returns the profile's key (APIKey, else from the environment), and
// the variable it came from ("" for APIKey).
func (p *Profile) key() (string, string) {
	if k := strings.TrimSpace(p.APIKey); k != "" {
		return k, ""
	}
	for _, v := range p.KeyEnv() {
		if k := strings.TrimSpace(os.Getenv(v)); k != "" {
			return k, v
		}
	}
	return "", ""
}

// LoadProfiles reads a profiles file: {"profiles": [Profile, …]} or a list.
// A field named like a key (api_key) is refused: keys go in environment
// variables.
func LoadProfiles(path string) ([]Profile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseProfiles(b)
}

// ParseProfiles is LoadProfiles on the file's content.
func ParseProfiles(b []byte) ([]Profile, error) {
	var raw json.RawMessage = bytes.TrimSpace(b)
	var wrap struct {
		Profiles []json.RawMessage `json:"profiles"`
	}
	var list []json.RawMessage
	if len(raw) > 0 && raw[0] == '[' {
		if err := json.Unmarshal(raw, &list); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(raw, &wrap); err != nil {
			return nil, err
		}
		list = wrap.Profiles
	}
	var out []Profile
	seen := map[string]bool{}
	for i, r := range list {
		var m map[string]json.RawMessage
		if err := json.Unmarshal(r, &m); err != nil {
			return nil, fmt.Errorf("profile %d: %v", i+1, err)
		}
		for k := range m {
			if lk := strings.ToLower(k); lk == "api_key" || lk == "apikey" || lk == "key" || lk == "token" {
				return nil, fmt.Errorf("profile %d: no keys in profiles: put the key in an environment variable and name it in api_key_env", i+1)
			}
		}
		dec := json.NewDecoder(bytes.NewReader(r))
		dec.DisallowUnknownFields()
		var p Profile
		if err := dec.Decode(&p); err != nil {
			return nil, fmt.Errorf("profile %d: %v", i+1, err)
		}
		if err := p.Check(); err != nil {
			return nil, err
		}
		if seen[p.Name] {
			return nil, fmt.Errorf("two profiles are named %s", p.Name)
		}
		seen[p.Name] = true
		out = append(out, p)
	}
	return out, nil
}

// ParseSpec reads a profile from the command line: [NAME=]PROVIDER:MODEL,
// then options as key=value separated by commas (temperature, top_p, top_k,
// max_output_tokens, seed, effort, thinking_budget, output, base_url,
// api_key_env, min_interval, timeout, retries). Without NAME, the profile
// is named PROVIDER:MODEL.
//
//	gemini:gemini-2.5-flash,temperature=0.7,min_interval=6s
//	local=lmstudio:qwen3-8b,output=json
func ParseSpec(spec string) (Profile, error) {
	var p Profile
	head, opts, _ := strings.Cut(spec, ",")
	if name, rest, ok := strings.Cut(head, "="); ok && !strings.Contains(name, ":") {
		p.Name, head = name, rest
	}
	prov, model, _ := strings.Cut(head, ":")
	p.Provider, p.Model = strings.TrimSpace(prov), strings.TrimSpace(model)
	if p.Name == "" {
		p.Name = head
	}
	for _, kv := range strings.Split(opts, ",") {
		if strings.TrimSpace(kv) == "" {
			continue
		}
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return p, fmt.Errorf("%q: options are key=value", kv)
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		var err error
		switch k {
		case "temperature", "top_p":
			var f float64
			if f, err = strconv.ParseFloat(v, 64); err == nil {
				if k == "temperature" {
					p.Temperature = &f
				} else {
					p.TopP = &f
				}
			}
		case "top_k", "max_output_tokens", "thinking_budget", "retries":
			var n int
			if n, err = strconv.Atoi(v); err == nil {
				switch k {
				case "top_k":
					p.TopK = &n
				case "max_output_tokens":
					p.MaxOutputTokens = n
				case "thinking_budget":
					p.ThinkingBudget = &n
				default:
					p.Retries = n
				}
			}
		case "seed":
			var n int64
			if n, err = strconv.ParseInt(v, 10, 64); err == nil {
				p.Seed = &n
			}
		case "effort":
			p.Effort = v
		case "output":
			p.Output = v
		case "schema_file":
			p.SchemaFile = v
		case "base_url":
			p.BaseURL = v
		case "api_key_env":
			p.APIKeyEnv = v
		case "min_interval", "timeout":
			var d time.Duration
			if d, err = time.ParseDuration(v); err == nil {
				if k == "timeout" {
					p.Timeout = Duration(d)
				} else {
					p.MinInterval = Duration(d)
				}
			}
		case "api_key", "key", "token":
			return p, fmt.Errorf("no keys on the command line: put the key in an environment variable and name it with api_key_env")
		default:
			return p, fmt.Errorf("unknown option %q", k)
		}
		if err != nil {
			return p, fmt.Errorf("%s: %v", k, err)
		}
	}
	return p, p.Check()
}

// DefaultBaseURL is the endpoint a provider uses when a profile gives none
// ("" for the SDK providers, which know their own).
func DefaultBaseURL(provider string) string { return endpoints[provider].BaseURL }

// NeedsKey reports whether the provider needs an API key (local servers do
// not; the "anthropic" provider can also use an `ant auth login` profile).
func NeedsKey(provider string) bool {
	switch provider {
	case "gemini", "anthropic":
		return true
	}
	return endpoints[provider].NeedsKey
}
