package web

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
)

const webRule = `language:
  p/1: "%s is a premise"
  q/1: "%s is a consequence"
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    conclusions: [q(X)]
statements:
  p(a): Premise a
assumptions: [p(a)]
`

const wrongAuthoredArgument = `language:
  p/1: "%s is a premise"
  q/1: "%s is another premise"
  r/1: "%s is a result"
argument_schemes:
  - id: s
    variables: [X]
    premises: [p(X)]
    conclusions: [r(X)]
statements:
  q(a): The wrong premise
  r(a): The result
arguments:
  a1:
    scheme: s
    parameters: [a]
    premises: [q(a)]
    conclusion: r(a)
assumptions: [q(a)]
`

func formRequest(t *testing.T, url, format, graph string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	if err := m.WriteField("input-format", format); err != nil {
		t.Fatal(err)
	}
	if err := m.WriteField("output-format", "yaml"); err != nil {
		t.Fatal(err)
	}
	w, err := m.CreateFormFile("datafile", "graph.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(graph)); err != nil {
		t.Fatal(err)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", m.FormDataContentType())
	return r
}

func testFormHandler() http.HandlerFunc {
	return newEvalHandler(template.Must(template.New("error").Parse("<p>{{.}}</p>")))
}

func rawRequest(t *testing.T, url, graph string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(graph))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/yaml")
	return r
}

// These run both real handlers over HTTP. The old form route swallowed Infer
// errors; the old raw route skipped authored-argument validation entirely.
func TestHTTPEvaluationRefusesFailedGraphs(t *testing.T) {
	defer func(old bool) { ExperimentalImport = old }(ExperimentalImport)
	ExperimentalImport = false
	cases := []struct {
		name, graph, outcome string
		status               int
	}{
		{"contradictory stipulations", "statements:\n  p: P\n  '¬p': Not P\nassumptions: [p, '¬p']\n", "inconsistent context", http.StatusUnprocessableEntity},
		{"unsupported guard", strings.Replace(webRule, "    conclusions:", "    guards: [X == a]\n    conclusions:", 1), "unsupported", http.StatusUnprocessableEntity},
		{"wrong predicate", wrongAuthoredArgument, "invalid input", http.StatusBadRequest},
		{"wrong instance", strings.ReplaceAll(wrongAuthoredArgument, "q(a)", "p(b)"), "invalid input", http.StatusBadRequest},
		{"unknown YAML field", "statements: {p: P}\nassumption: [p]\n", "invalid input", http.StatusBadRequest},
	}
	for _, raw := range []bool{false, true} {
		name, handler := "form", testFormHandler()
		if raw {
			name, handler = "raw", evalArgGraphHandler
		}
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(handler)
			defer srv.Close()
			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) {
					r := formRequest(t, srv.URL, "yaml", c.graph)
					if raw {
						r = rawRequest(t, srv.URL, c.graph)
					}
					resp, err := srv.Client().Do(r)
					if err != nil {
						t.Fatal(err)
					}
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Fatal(err)
					}
					if resp.StatusCode != c.status || !strings.Contains(string(body), c.outcome) {
						t.Fatalf("status=%d, body=%s; want %d and %s", resp.StatusCode, body, c.status, c.outcome)
					}
					if raw {
						var message string
						if err := json.Unmarshal(body, &message); err != nil || resp.Header.Get("Content-Type") != "application/json" {
							t.Fatalf("raw error is not a JSON string: content-type=%s body=%s error=%v", resp.Header.Get("Content-Type"), body, err)
						}
					}
					if strings.Contains(string(body), "statements:") || strings.Contains(string(body), `"statements":`) {
						t.Fatalf("failed graph was exported as a result: %s", body)
					}
				})
			}
			// Rejection must not replace ordinary inference with blanket refusal.
			r := formRequest(t, srv.URL, "yaml", webRule)
			if raw {
				r = rawRequest(t, srv.URL, webRule)
			}
			resp, err := srv.Client().Do(r)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			graph, err := yaml.Import(resp.Body)
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Fatalf("valid inference refused: status=%d, error=%v", resp.StatusCode, err)
			}
			q := graph.Statements["q(a)"]
			if q == nil || q.Label != caes.In {
				t.Fatal("valid rule did not derive and label q(a)")
			}
		})
	}
}

func TestHTTPImportProfile(t *testing.T) {
	defer func(old bool) { ExperimentalImport = old }(ExperimentalImport)
	graphs := map[string]string{
		"yaml":  "statements: {p: Premise}\nassumptions: [p]\n",
		"aif":   `{"nodes":[{"nodeID":"p","type":"I","text":"Premise"}],"edges":[]}`,
		"caf":   `<caf version="1.3"><statements><statement id="p" weight="1"><descriptions><description lang="en">Premise</description></descriptions></statement></statements><arguments/></caf>`,
		"agxml": `<arggraph id="g"><edu id="e">Premise</edu><adu id="p" type="pro"/><edge id="s" src="e" trg="p" type="seg"/></arggraph>`,
		"lkif":  `<lkif><argument-graphs><argument-graph><statements><statement id="p" value="true"><s>Premise</s></statement></statements><arguments/></argument-graph></argument-graphs></lkif>`,
	}
	for _, experimental := range []bool{false, true} {
		ExperimentalImport = experimental
		srv := httptest.NewServer(testFormHandler())
		for format, graph := range graphs {
			resp, err := srv.Client().Do(formRequest(t, srv.URL, format, graph))
			if err != nil {
				t.Fatal(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			legacy := format == "caf" || format == "agxml" || format == "lkif"
			if legacy && !experimental {
				if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(string(body), "unsupported") {
					t.Errorf("%s without opt-in: status=%d body=%s", format, resp.StatusCode, body)
				}
			} else if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "label: in") {
				t.Errorf("%s experimental=%v: positive import failed: status=%d body=%s", format, experimental, resp.StatusCode, body)
			}
		}
		// Opt-in must reach, not bypass, each converter's admission checks.
		for _, format := range []string{"caf", "agxml", "lkif"} {
			resp, err := srv.Client().Do(formRequest(t, srv.URL, format, "<wrong/>"))
			if err != nil {
				t.Fatal(err)
			}
			body, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			want := http.StatusUnprocessableEntity
			if experimental {
				want = http.StatusBadRequest
			}
			if resp.StatusCode != want || strings.Contains(string(body), "statements:") {
				t.Errorf("malformed %s experimental=%v: status=%d body=%s", format, experimental, resp.StatusCode, body)
			}
		}
		resp, err := srv.Client().Do(formRequest(t, srv.URL, "graphml", "<graphml/>"))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("unknown format accepted: status=%d", resp.StatusCode)
		}
		srv.Close()
	}
}

func TestHTTPEvaluationUsesRequestCancellation(t *testing.T) {
	for _, raw := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		r, h := formRequest(t, "/carneades/eval", "yaml", webRule), testFormHandler()
		if raw {
			r, h = rawRequest(t, "/carneades/eval-arg-graph", webRule), evalArgGraphHandler
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r.WithContext(ctx))
		if w.Code != http.StatusRequestTimeout || !strings.Contains(w.Body.String(), "cancelled") {
			t.Fatalf("raw=%v: cancelled request returned %d: %s", raw, w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "statements:") || strings.Contains(w.Body.String(), `"statements":`) {
			t.Fatalf("raw=%v: cancelled request exported a graph", raw)
		}
	}
}
