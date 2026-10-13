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
//	--claude     offer Claude seats, played through the Claude API with the
//	             text protocol (needs ANTHROPIC_API_KEY; moves cost API credits);
//	             --claude-model (default claude-opus-5-5), --claude-effort (default medium)
//
// The engine runs in this process; the browser shows the game and sends the
// moves. Several tabs can show the same game, for example the players and a
// moderator.
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/yourorg/wag/data"
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
	flag.Parse()

	d := server.Data{YAML: data.CoreYAML, Schemes: data.SchemesJSON, CQs: data.CQsJSON, Edges: data.EdgesJSON, Contexts: data.ContextsJSON}
	if *dataDir != "" {
		for _, x := range []struct {
			dst  *[]byte
			path string
		}{{&d.YAML, "schemes/core.yml"}, {&d.Schemes, "nascq/schemes.json"}, {&d.CQs, "nascq/critical_questions.json"}, {&d.Edges, "nascq/edges.json"}, {&d.Contexts, "contexts.json"}} {
			b, err := os.ReadFile(filepath.Join(*dataDir, x.path))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading data:", err)
				os.Exit(1)
			}
			*x.dst = b
		}
	}
	srv, err := server.New(d, *kialoDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if *claude {
		srv.SetClaude(server.NewClaude(*claudeModel, *claudeEffort))
		fmt.Printf("Claude seats play with %s (effort %s).\n", *claudeModel, *claudeEffort)
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
	if err := http.Serve(ln, srv.Handler()); err != nil {
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
