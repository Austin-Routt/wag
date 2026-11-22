package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/agxml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/aif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/caf"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/dot"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/graphml"
	cj "github.com/carneades/carneades-4/src/engine/caes/encoding/json"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/lkif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

// evaluateGraph is the common admission and inference boundary of both HTTP
// evaluation routes. Inference is not a substitute for authored-graph
// validation. A failed run must not be labelled or exported as a result.
func evaluateGraph(ctx context.Context, ag *caes.ArgGraph) error {
	problems := validation.Validate(ag)
	if len(problems) > 0 {
		details := make([]string, 0, len(problems))
		for _, p := range problems {
			if p.Expression == "" {
				details = append(details, fmt.Sprintf("%s: %s: %s", p.Category, p.Id, p.Description))
			} else {
				details = append(details, fmt.Sprintf("%s: %s: %s: %s", p.Category, p.Id, p.Description, p.Expression))
			}
		}
		return caes.NewRunError(caes.ErrInvalidInput, "graph validation failed: %s", strings.Join(details, "; "))
	}
	if err := ag.InferContext(ctx); err != nil {
		return err
	}
	ag.ApplyLabelling(ag.GroundedLabelling())
	return nil
}

// The engine's CLI exit codes are not HTTP status codes. Preserve the typed
// outcome in the response body and use a non-success transport status.
func evaluationStatus(err error) int {
	switch {
	case errors.Is(err, caes.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, caes.ErrUnsupported), errors.Is(err, caes.ErrInconsistentContext), errors.Is(err, caes.ErrLimitReached):
		return http.StatusUnprocessableEntity
	case errors.Is(err, caes.ErrCancelled):
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

func inputError(err error) error {
	var runError *caes.RunError
	if errors.As(err, &runError) {
		return err
	}
	return caes.NewRunError(caes.ErrInvalidInput, "%v", err)
}

func newEvalHandler(errorTemplate *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fail := func(err error) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(evaluationStatus(err))
			errorTemplate.Execute(w, err.Error())
		}
		inputFormat := req.FormValue("input-format")
		outputFormat := req.FormValue("output-format")
		// FormValue parses the multipart request, including uploaded parts.
		// The profile gate is before conversion, not before reading the body.
		if req.MultipartForm != nil {
			defer req.MultipartForm.RemoveAll()
		}
		if err := checkImportProfile(inputFormat); err != nil {
			fail(err)
			return
		}
		file, _, err := req.FormFile("datafile")
		if err != nil {
			fail(inputError(err))
			return
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			fail(inputError(err))
			return
		}
		rd := bytes.NewReader(data)
		var ag *caes.ArgGraph
		switch inputFormat {
		case "yaml":
			ag, err = yaml.Import(rd)
		case "aif":
			ag, err = aif.Import(rd)
		case "agxml":
			ag, err = agxml.Import(rd)
		case "lkif":
			ag, err = lkif.Import(rd)
		case "caf":
			ag, err = caf.Import(rd)
		}
		if err != nil {
			fail(inputError(err))
			return
		}
		if err := evaluateGraph(req.Context(), ag); err != nil {
			fail(err)
			return
		}
		switch outputFormat {
		case "yaml":
			yaml.Export(w, ag)
		case "graphml":
			err = graphml.Export(w, ag)
		case "dot":
			err = dot.Export(w, ag)
		case "png", "svg":
			err = exportArgGraph(ag, w, outputFormat)
		default:
			err = caes.NewRunError(caes.ErrUnsupported, "unknown or unsupported output format: %s", outputFormat)
		}
		if err != nil {
			fail(err)
		}
	}
}

// Evaluate an argument graph in YAML (including JSON) format and return the
// resulting argument graph in JSON, or SVG when explicitly requested.
func evalArgGraphHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if req.Method == "OPTIONS" {
		return
	}
	fail := func(err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(evaluationStatus(err))
		json.NewEncoder(w).Encode(err.Error())
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fail(inputError(err))
		return
	}
	ag, err := yaml.Import(bytes.NewReader(body))
	if err != nil {
		fail(inputError(err))
		return
	}
	if err := evaluateGraph(req.Context(), ag); err != nil {
		fail(err)
		return
	}
	if req.Header.Get("Accept") == "image/svg+xml" {
		w.Header().Set("Content-Type", "image/svg+xml")
		if err := exportArgGraph(ag, w, "svg"); err != nil {
			fail(err)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		cj.Export(w, ag)
	}
}
