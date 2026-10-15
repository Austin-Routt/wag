package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/wag/internal/llm"
)

// The dashboard's workbench: connecting model services (keys typed in the
// browser), adding players, and annotation runs (workbench_annotate.go).
// Everything it keeps is in one data folder (Server.store):
//
//	players.json        the players added in the dashboard (model profiles, no keys)
//	keys.json           the keys you asked to remember (unencrypted; only if you ask)
//	games/              every game with a model seat: its trace file and export
//	annotation/         one folder per annotation run (see its README.txt)
//	datasets/           your own annotation datasets (a folder each, with items.json)

// keyStore holds the keys typed in the dashboard: in memory, and in
// keys.json for those to remember.
type keyStore struct {
	mu         sync.Mutex
	keys       map[string]string
	remembered map[string]bool
	file       string
}

func (k *keyStore) get(service string) string {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.keys[service]
}

func (k *keyStore) load() {
	if k.file == "" {
		return
	}
	b, err := os.ReadFile(k.file)
	if err != nil {
		return
	}
	var m map[string]string
	if json.Unmarshal(b, &m) == nil {
		for s, v := range m {
			k.keys[s], k.remembered[s] = v, true
		}
	}
}

// set keeps a key ("" forgets it), and writes keys.json.
func (k *keyStore) set(service, key string, remember bool) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if key == "" {
		delete(k.keys, service)
		delete(k.remembered, service)
	} else {
		k.keys[service], k.remembered[service] = key, remember
	}
	if k.file == "" {
		if remember {
			return fmt.Errorf("this server has no data folder to remember keys in")
		}
		return nil
	}
	m := map[string]string{}
	for s, v := range k.keys {
		if k.remembered[s] {
			m[s] = v
		}
	}
	if len(m) == 0 {
		if err := os.Remove(k.file); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	b, _ := json.MarshalIndent(m, "", " ")
	return os.WriteFile(k.file, b, 0o600)
}

// SetStore makes dir (created) the server's data folder: games with model
// seats are saved there, annotation runs are made there, and the players
// added in the dashboard and the keys to remember are kept there. It loads
// those players and keys.
func (s *Server) SetStore(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	for _, sub := range []string{"", "games", "annotation", "datasets"} {
		if err := os.MkdirAll(filepath.Join(abs, sub), 0o755); err != nil {
			return err
		}
	}
	s.mu.Lock()
	s.store = abs
	s.keys.file = filepath.Join(abs, "keys.json")
	s.keys.load()
	var ps []llm.Profile
	if b, err := os.ReadFile(filepath.Join(abs, "players.json")); err == nil {
		if ps, err = llm.ParseProfiles(b); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("players.json: %v", err)
		}
	}
	s.players = ps
	s.mu.Unlock()
	for _, p := range ps {
		s.connect(p) // a player whose key is missing waits for it
	}
	return nil
}

// Store is the server's data folder ("" if none).
func (s *Server) Store() string { return s.store }

// connect makes the client of a dashboard player, with the key typed for
// its service if any, and offers it as a seat; the error says why not.
func (s *Server) connect(p llm.Profile) error {
	p.APIKey = s.keys.get(p.Provider)
	c, err := llm.NewClient(p)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		delete(s.models, p.Name)
		s.playerErr[p.Name] = err.Error()
		s.meta = nil
		return err
	}
	delete(s.playerErr, p.Name)
	s.models[p.Name] = c
	s.meta = nil
	return nil
}

// savePlayers writes players.json. Caller holds s.mu.
func (s *Server) savePlayers() error {
	if s.store == "" {
		return nil
	}
	b, err := json.MarshalIndent(map[string]any{"profiles": s.players}, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.store, "players.json"), b, 0o644)
}

// local guards the workbench API: the request must come from this computer
// (keys are typed here, and calls spend the account's quota), name this
// computer as its host (no DNS rebinding), and, if it changes anything, be
// JSON from this page's origin (no request forged by another site).
func local(w http.ResponseWriter, r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if ip := net.ParseIP(host); err != nil || ip == nil || !ip.IsLoopback() {
		httpError(w, fmt.Errorf("the model workbench answers only on this computer"), http.StatusForbidden)
		return false
	}
	h := r.Host
	if hh, _, err := net.SplitHostPort(h); err == nil {
		h = hh
	}
	h = strings.Trim(h, "[]")
	if ip := net.ParseIP(h); h != "localhost" && (ip == nil || !ip.IsLoopback()) {
		httpError(w, fmt.Errorf("the model workbench answers only at localhost"), http.StatusForbidden)
		return false
	}
	if r.Method != http.MethodGet {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			httpError(w, fmt.Errorf("send JSON"), http.StatusUnsupportedMediaType)
			return false
		}
		if o := r.Header.Get("Origin"); o != "" {
			if u, err := url.Parse(o); err != nil || u.Host != r.Host {
				httpError(w, fmt.Errorf("a request from another site"), http.StatusForbidden)
				return false
			}
		}
	}
	return true
}

// Service is a model service as the dashboard shows it.
type Service struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	About     string   `json:"about"`
	NeedsKey  bool     `json:"needs_key"`
	KeyEnv    []string `json:"key_env,omitempty"`
	KeySource string   `json:"key_source,omitempty"` // dashboard, remembered, environment variable NAME
	KeyHint   string   `json:"key_hint,omitempty"`   // the key's last 4 characters
	BaseURL   string   `json:"base_url,omitempty"`   // the default endpoint
	Local     bool     `json:"local,omitempty"`
	GetKey    string   `json:"get_key,omitempty"` // where to get a key
	Note      string   `json:"note,omitempty"`
}

// services are the services the dashboard offers, in its order.
var services = []struct{ id, name, getKey string }{
	{"gemini", "Google Gemini", "https://aistudio.google.com/apikey"},
	{"anthropic", "Anthropic Claude", "https://console.anthropic.com/settings/keys"},
	{"openai", "OpenAI", "https://platform.openai.com/api-keys"},
	{"deepseek", "DeepSeek", "https://platform.deepseek.com/api_keys"},
	{"lmstudio", "LM Studio (on this computer)", ""},
	{"ollama", "Ollama (on this computer)", ""},
	{"llamacpp", "llama.cpp server (on this computer)", ""},
	{"sglang", "SGLang (on this computer)", ""},
	{"vllm", "vLLM (on this computer)", ""},
	{"openai-compatible", "Another OpenAI-compatible server", ""},
	{"mock", "Test service (scripted answers, no network)", ""},
}

func (s *Server) serviceList() []Service {
	var out []Service
	for _, sv := range services {
		p := llm.Profile{Provider: sv.id}
		x := Service{ID: sv.id, Name: sv.name, About: llm.Providers[sv.id], NeedsKey: llm.NeedsKey(sv.id), KeyEnv: p.KeyEnv(),
			BaseURL: llm.DefaultBaseURL(sv.id), Local: strings.Contains(sv.name, "this computer") || sv.id == "openai-compatible", GetKey: sv.getKey,
			Note: p.Notice()}
		s.keys.mu.Lock()
		k, rem := s.keys.keys[sv.id], s.keys.remembered[sv.id]
		s.keys.mu.Unlock()
		switch {
		case k != "" && rem:
			x.KeySource = "remembered"
		case k != "":
			x.KeySource = "dashboard"
		default:
			for _, v := range x.KeyEnv {
				if e := strings.TrimSpace(os.Getenv(v)); e != "" {
					x.KeySource, k = "environment variable "+v, e
					break
				}
			}
		}
		if len(k) >= 8 {
			x.KeyHint = "…" + k[len(k)-4:]
		}
		out = append(out, x)
	}
	return out
}

// PlayerView is a player as the dashboard lists it.
type PlayerView struct {
	llm.Profile
	From    string `json:"from"` // dashboard, command line
	Ready   bool   `json:"ready"`
	Problem string `json:"problem,omitempty"`
}

func (s *Server) playerList() []PlayerView {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []PlayerView{}
	seen := map[string]bool{}
	for _, p := range s.players {
		v := PlayerView{Profile: p, From: "dashboard", Ready: s.models[p.Name] != nil, Problem: s.playerErr[p.Name]}
		out = append(out, v)
		seen[p.Name] = true
	}
	for _, name := range sortedKeys(s.models) {
		if !seen[name] {
			out = append(out, PlayerView{Profile: s.models[name].Profile, From: "command line", Ready: true})
		}
	}
	return out
}

// playerRequest is a player as the dashboard adds it.
type playerRequest struct {
	Name           string   `json:"name"`
	Service        string   `json:"service"`
	Model          string   `json:"model"`
	BaseURL        string   `json:"base_url"`
	Temperature    *float64 `json:"temperature"`
	MaxOutput      int      `json:"max_output_tokens"`
	Effort         string   `json:"effort"`
	ThinkingBudget *int     `json:"thinking_budget"`
	Output         string   `json:"output"`
	MinInterval    float64  `json:"min_interval_seconds"`
	Timeout        float64  `json:"timeout_seconds"`
	Retries        int      `json:"retries"`
}

func (q playerRequest) profile() llm.Profile {
	p := llm.Profile{Name: strings.TrimSpace(q.Name), Provider: q.Service, BaseURL: strings.TrimSpace(q.BaseURL), Retries: q.Retries,
		MinInterval: llm.Duration(time.Duration(q.MinInterval * float64(time.Second))), Timeout: llm.Duration(time.Duration(q.Timeout * float64(time.Second))),
		Params: llm.Params{Model: strings.TrimSpace(q.Model), Temperature: q.Temperature, MaxOutputTokens: q.MaxOutput, Effort: q.Effort,
			ThinkingBudget: q.ThinkingBudget, Output: q.Output}}
	if p.Name == "" {
		p.Name = p.Model
		if i := strings.LastIndex(p.Name, "/"); i >= 0 {
			p.Name = p.Name[i+1:]
		}
	}
	return p
}

// handleLLM serves /api/llm/...: services, keys, models, players.
func (s *Server) handleLLM(w http.ResponseWriter, r *http.Request) {
	if !local(w, r) {
		return
	}
	what := strings.TrimPrefix(r.URL.Path, "/api/llm/")
	switch what {
	case "services":
		writeJSON(w, map[string]any{"services": s.serviceList(), "store": s.store})
	case "key", "forget":
		var req struct {
			Service  string `json:"service"`
			Key      string `json:"key"`
			Remember bool   `json:"remember"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if _, ok := llm.Providers[req.Service]; !ok {
			httpError(w, fmt.Errorf("unknown service %q", req.Service), http.StatusBadRequest)
			return
		}
		key := strings.TrimSpace(req.Key)
		if what == "forget" {
			key = ""
		} else if key == "" || strings.ContainsAny(key, " \t\r\n") {
			httpError(w, fmt.Errorf("paste the key itself: no spaces"), http.StatusBadRequest)
			return
		}
		if err := s.keys.set(req.Service, key, req.Remember); err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		s.reconnect(req.Service)
		writeJSON(w, map[string]any{"services": s.serviceList(), "players": s.playerList()})
	case "models":
		var req struct {
			Service string `json:"service"`
			BaseURL string `json:"base_url"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		p := llm.Profile{Name: req.Service, Provider: req.Service, BaseURL: strings.TrimSpace(req.BaseURL), APIKey: s.keys.get(req.Service)}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		list, err := llm.ListModels(ctx, p)
		if err != nil {
			httpError(w, err, http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]any{"models": list})
	case "test":
		var req playerRequest
		if !readJSON(w, r, &req) {
			return
		}
		p := req.profile()
		if p.Name == "" {
			p.Name = "test"
		}
		p.APIKey = s.keys.get(p.Provider)
		p.Retries = -1
		c, err := llm.NewClient(p)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
		defer cancel()
		var att llm.Attempt
		resp, err := c.Complete(ctx, c.Request("Answer in one word.", []llm.Message{{Role: "user", Content: "Reply with the single word: ready"}}),
			func(a llm.Attempt) { att = a })
		out := map[string]any{"latency_ms": att.Latency.Milliseconds()}
		if err != nil {
			out["error"] = err.Error()
		} else {
			out["text"], out["model"], out["usage"], out["finish_reason"] = strings.TrimSpace(resp.Text), resp.Model, resp.Usage, resp.FinishReason
		}
		writeJSON(w, out)
	case "players":
		if r.Method == http.MethodGet {
			writeJSON(w, map[string]any{"players": s.playerList()})
			return
		}
		var req playerRequest
		if !readJSON(w, r, &req) {
			return
		}
		p := req.profile()
		if err := p.Check(); err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		replaced := false
		for i := range s.players {
			if s.players[i].Name == p.Name {
				s.players[i], replaced = p, true
			}
		}
		if !replaced {
			if s.models[p.Name] != nil {
				s.mu.Unlock()
				httpError(w, fmt.Errorf("a player named %s comes from the command line: choose another name", p.Name), http.StatusConflict)
				return
			}
			s.players = append(s.players, p)
		}
		err := s.savePlayers()
		s.mu.Unlock()
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		cerr := s.connect(p)
		res := map[string]any{"players": s.playerList()}
		if cerr != nil {
			res["warning"] = "added, but it cannot play yet: " + cerr.Error()
		}
		writeJSON(w, res)
	case "players/remove":
		var req struct {
			Name string `json:"name"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		s.mu.Lock()
		found := false
		for i := range s.players {
			if s.players[i].Name == req.Name {
				s.players = append(s.players[:i], s.players[i+1:]...)
				found = true
				break
			}
		}
		if found {
			delete(s.models, req.Name)
			delete(s.playerErr, req.Name)
			s.meta = nil
		}
		err := s.savePlayers()
		s.mu.Unlock()
		if !found {
			httpError(w, fmt.Errorf("no dashboard player %q", req.Name), http.StatusNotFound)
			return
		}
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"players": s.playerList()})
	default:
		httpError(w, fmt.Errorf("unknown %s", r.URL.Path), http.StatusNotFound)
	}
}

// reconnect remakes the clients of the dashboard players of a service (its
// key changed).
func (s *Server) reconnect(service string) {
	s.mu.Lock()
	var ps []llm.Profile
	for _, p := range s.players {
		if p.Provider == service {
			ps = append(ps, p)
		}
	}
	s.mu.Unlock()
	for _, p := range ps {
		s.connect(p)
	}
}

// inStore resolves a path relative to the data folder, refusing one that
// leaves it.
func (s *Server) inStore(rel string) (string, error) {
	if s.store == "" {
		return "", fmt.Errorf("this server has no data folder (start it with --store)")
	}
	p := filepath.Join(s.store, filepath.FromSlash(rel))
	if p != s.store && !strings.HasPrefix(p, s.store+string(filepath.Separator)) {
		return "", fmt.Errorf("%q is not in the data folder", rel)
	}
	return p, nil
}

// openFolder shows a folder in the system's file manager.
func openFolder(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// StoreEntry is a saved game or a run, for the dashboard's lists.
type StoreEntry struct {
	Path     string    `json:"path"` // relative to the data folder
	Title    string    `json:"title"`
	Modified time.Time `json:"modified"`
	Files    []string  `json:"files"`
}

func listDir(store, sub string) []StoreEntry {
	entries, _ := os.ReadDir(filepath.Join(store, sub))
	out := []StoreEntry{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, _ := e.Info()
		x := StoreEntry{Path: sub + "/" + e.Name(), Title: e.Name()}
		if info != nil {
			x.Modified = info.ModTime()
		}
		files, _ := os.ReadDir(filepath.Join(store, sub, e.Name()))
		for _, f := range files {
			if !f.IsDir() {
				x.Files = append(x.Files, f.Name())
			}
		}
		out = append(out, x)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Modified.After(out[j].Modified) })
	return out
}

// handleStore serves /api/store: the data folder and what is in it;
// /api/store/open shows a folder of it; /api/store/file?path= sends a file.
func (s *Server) handleStore(w http.ResponseWriter, r *http.Request) {
	if !local(w, r) {
		return
	}
	switch strings.TrimPrefix(r.URL.Path, "/api/store") {
	case "", "/":
		if s.store == "" {
			writeJSON(w, map[string]any{"dir": ""})
			return
		}
		writeJSON(w, map[string]any{"dir": s.store, "games": listDir(s.store, "games"), "annotation": listDir(s.store, "annotation"),
			"datasets": listDir(s.store, "datasets"), "sep": string(filepath.Separator)})
	case "/open":
		var req struct {
			Path string `json:"path"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		p, err := s.inStore(req.Path)
		if err == nil {
			err = openFolder(p)
		}
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{"opened": p})
	case "/file":
		p, err := s.inStore(r.URL.Query().Get("path"))
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			httpError(w, fmt.Errorf("no file %s", r.URL.Query().Get("path")), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(p))
		http.ServeFile(w, r, p)
	default:
		httpError(w, fmt.Errorf("unknown %s", r.URL.Path), http.StatusNotFound)
	}
}

// autosave writes a game with a model seat to the data folder: its trace
// file and its export, after every change. Caller holds s.mu.
func (s *Server) autosave(ss *Session) {
	if s.store == "" {
		return
	}
	hasModel := false
	for _, seat := range ss.Seats {
		if seat.aiKind() == "model" {
			hasModel = true
		}
	}
	if !hasModel {
		return
	}
	dir := filepath.Join(s.store, "games", ss.Created.Format("2006-01-02_150405")+"_game"+ss.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "saving the game:", err)
		return
	}
	ss.SavedTo = dir
	var tb bytes.Buffer
	s.encodeTraces(&tb, ss)
	eb, _ := json.MarshalIndent(ss.export(s.contexts), "", " ")
	for name, b := range map[string][]byte{"trace.jsonl": tb.Bytes(), "export.json": eb} {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "saving the game:", err)
		}
	}
}
