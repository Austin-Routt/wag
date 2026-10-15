// WAG server — play the Waltonian Argumentation Game in a browser.
//
// Usage:
//
//	go run ./cmd/wag-server            # serves http://127.0.0.1:8080 and opens the browser
//
// Flags:
//
//	--addr       address to listen on (default 127.0.0.1:8080; only this machine)
//	--data       directory with schemes/core.yml and nascq/*.json (default: the data built in)
//	--kialo-dir  directory of captured Kialo discussions (<dir>/<id>/graph.json, as in the
//	             Kialo research visualizer's data folder), offered for import
//	--open       open the browser (default true)
//	--model      offer seats played by a language model through an API:
//	             [NAME=]PROVIDER:MODEL[,option=value…], repeatable, e.g.
//	             --model gemini:gemini-2.5-flash,min_interval=6s (key in GEMINI_API_KEY);
//	             providers: gemini, anthropic, openai, deepseek, lmstudio, llamacpp,
//	             sglang, vllm, ollama, openai-compatible, command, mock (wag-llm providers)
//	--models     a JSON file of such profiles (docs/llm.md)
//	--store      the data folder (default: WAG in your home folder, e.g.
//	             C:\Users\you\WAG): the players and keys added in the dashboard's
//	             Models tab, games with a model seat, annotation runs, datasets
//	--claude     offer Claude seats (the profile "claude", through the Claude API;
//	             needs ANTHROPIC_API_KEY; moves cost API credits);
//	             --claude-model (default claude-opus-5-5), --claude-effort (default medium)
//	--contexts   a JSON file of context presets to offer besides the built-in ones
//
// The engine runs in this process; the browser shows the game and sends the
// moves. Several tabs can show the same game, for example the players and a
// moderator.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"

	"github.com/yourorg/wag/internal/game"
	"github.com/yourorg/wag/internal/llm"
	"github.com/yourorg/wag/internal/server"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "address to listen on")
	dataDir := flag.String("data", "", "directory with schemes/core.yml and nascq/*.json (default: built-in data)")
	kialoDir := flag.String("kialo-dir", "", "directory of captured Kialo discussions (<dir>/<id>/graph.json)")
	open := flag.Bool("open", true, "open the browser")
	claude := flag.Bool("claude", false, "offer Claude seats, played through the Claude API (needs ANTHROPIC_API_KEY or an `ant auth login` profile; the moves cost API credits)")
	claudeModel := flag.String("claude-model", "claude-opus-5-5", "the model Claude seats use")
	claudeEffort := flag.String("claude-effort", "medium", "the effort Claude seats use: low, medium, high, xhigh or max")
	var modelSpecs llm.Specs
	flag.Var(&modelSpecs, "model", "offer seats played by a language model: [NAME=]PROVIDER:MODEL[,option=value…] (repeatable; keys in environment variables)")
	modelsFile := flag.String("models", "", "a JSON file of model profiles to offer as seats")
	home, _ := os.UserHomeDir()
	storeDir := flag.String("store", filepath.Join(home, "WAG"), "the data folder: dashboard players and remembered keys, games with model seats, annotation runs, datasets")
	contextsFile := flag.String("contexts", "", "a JSON file of context presets (as data/contexts.json, or one preset as the browser saves it) to offer besides the built-in ones; a preset with a built-in's id replaces it")
	flag.Parse()

	d, err := server.ReadData(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	srv, err := server.New(d, *kialoDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if *contextsFile != "" {
		b, err := os.ReadFile(*contextsFile)
		if err == nil {
			var cs []game.Context
			if cs, err = game.LoadContexts(b); err == nil {
				srv.AddContexts(cs)
				fmt.Printf("Context presets added from %s: %d.\n", *contextsFile, len(cs))
			}
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading the context presets:", err)
			os.Exit(1)
		}
	}
	if *storeDir != "" {
		if err := srv.SetStore(*storeDir); err != nil {
			fmt.Fprintln(os.Stderr, "Error with the data folder:", err)
			os.Exit(1)
		}
		fmt.Println("Your data folder (games with model seats, annotation runs, players):", srv.Store())
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	srv.SetContext(ctx)
	if *claude {
		if err := srv.SetClaude(*claudeModel, *claudeEffort); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Claude seats play with %s (effort %s).\n", *claudeModel, *claudeEffort)
	}
	profiles, err := llm.ProfilesFrom(*modelsFile, modelSpecs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	for _, p := range profiles {
		c, err := llm.NewClient(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in model profile %s: %v\n", p.Name, err)
			os.Exit(1)
		}
		srv.AddModel(c)
		fmt.Printf("Model seats: %s (%s, %s).\n", p.Name, p.Provider, p.Model)
		if n := p.Notice(); n != "" {
			fmt.Println("  Note:", n)
		}
	}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Println("WAG is running at", url)
	fmt.Println("Press Ctrl+C to stop.")
	if *open {
		openBrowser(url)
	}
	hs := &http.Server{Handler: srv.Handler()}
	go func() {
		<-ctx.Done() // Ctrl+C: stop the model calls, then the server
		_ = hs.Shutdown(context.Background())
	}()
	if err := hs.Serve(ln); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("Open", url, "in your browser.")
	}
}
