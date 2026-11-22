package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/agxml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/aif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/caf"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/dot"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/lkif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

const helpEval = `
usage: carneades eval [-f input-format] [-experimental-import] [-t output-format] [-o output-file] [input-file]

Evaluates an argument graph and prints the result in the selected output format.
The argument graph is first checked for syntactic and semantic errors and
evaluted only if no errors where found.  Any problems found are printed
to stderr.

If no input-file is specified, input is read from stdin. 

The -f flag ("from") specifies the format of the input file: Currently 
yaml and a restricted aif profile are supported. (default: yaml)
The legacy caf, agxml and lkif converters require -experimental-import.
That opt-in does not certify conversion fidelity or skip input validation.
The AIF profile accepts I/RA/binary-CA structures only; CA becomes symmetric
incompatibility and RA becomes a linked argument. Scheme names are provenance,
not executable scheme-library bindings. Dialogical nodes are refused.
I nodes without an incoming RA and without a CA issue are stipulated; this
conversion policy does not establish their truth or general AIFdb conformance.

The "yaml" format is the native format of Carneades 4.x.  YAML is a schemeless
data interchange format. See http://yaml.org/ for general information about
YAML.  For examples of Carneades argument graphs in this format, see

    https://github.com/carneades/carneades-4/examples/AGs/YAML

The "aif" format is the JSON serialization of the Argument Interchange 
Format (AIF).  For further information about AIF, see:

    http://www.argumentinterchange.org/developers
    http://www.arg-tech.org/index.php/projects/

The "agxml" format is the XML schema used by the
annotated corpus of argumentative microtexts Github project.
For further information about the agxml XML schema, see:

    https://github.com/peldszus/arg-microtext

The "caf" format is the Carneades Argument Format (CAF), the native format of 
Carneades 3. For further information about the CAF and Carneades 3, see::

	https://github.com/carneades/carneades-3/blob/master/schemas/CAF.rnc
	https://github.com/carneades/carneades-3
	
The "lkif" format is the Legal Knowledge Interchange Format
(LKIF) XML schema. Only the first argument graph in the LKIF file
is translated. For more information about LKIF, see:
Gordon, T. F. The Legal Knowledge Interchange Format (LKIF).
Deliverable 4.1, European Commission, 2008.
http://www.tfgordon.de/publications/files/GordonLKIF2008.pdf

LKIF is the native format of Carneades 2, a desktop argument mapping
tool with a graphical user interface. Also known as the Carneades Editor.
https://github.com/carneades/carneades-2

The -t flag ("to") specifies the output format of the evaluated argument
graph. Currently graphml, dot, and yaml are supported. (default: graphml)	

GraphML is an XML schema for directed graphs.  Graphml is supported by 
several graph editors and visualizations tools.  For further information see:

    http://graphml.graphdrawing.org/

The graphml produced by Carneades is intended for use with the yEd
graph editor. For more information about yEd, see:

    https://www.yworks.com/en/products/yfiles/yed/
	
The "dot" format is the native format of the open source GraphViz network
visualization project. For further information see:

	http://graphviz.org/

The -o flag specifies the output file name. If the -o flag is not used, 
output goes to stdout.
`

func evalCmd() {
	eval := flag.NewFlagSet("eval", flag.ContinueOnError)
	fromFlag := eval.String("f", "yaml", "the format of the source file")
	experimentalFlag := eval.Bool("experimental-import", false, "opt in to a legacy converter; does not certify conversion fidelity")
	toFlag := eval.String("t", "graphml", "the format of the output file")
	outFileFlag := eval.String("o", "", "the filename of the output file")

	var inFile *os.File
	var outFile *os.File
	var err error

	if err := eval.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	// An interrupt or a termination signal cancels the evaluation: the run
	// stops, nothing is printed, and the exit status is 5 (docs/contracts.md,
	// contract 3).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cancelled := func() {
		fmt.Fprintf(os.Stderr, "carneades eval: %v\n", caes.NewRunError(caes.ErrCancelled, "the evaluation was stopped by a signal"))
		os.Exit(caes.ExitStatus(caes.ErrCancelled))
	}

	if err := checkImportProfile(*fromFlag, *experimentalFlag); err != nil {
		failImport(err)
		return
	}
	if !contains(outputFormats, *toFlag) {
		log.Fatal(fmt.Errorf("unsupported output format: %s\n", *toFlag))
		return
	}
	switch eval.NArg() {
	case 0:
		inFile = os.Stdin
	case 1:
		inFile, err = os.Open(eval.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal(fmt.Errorf("incorrect number of arguments after the command flags; should be 0, to read from stdin, or 1, naming the input file\n"))
		return
	}
	if *outFileFlag == "" {
		outFile = os.Stdout
	} else {
		outFile, err = os.Create(*outFileFlag)
		if err != nil {
			log.Fatal(fmt.Errorf("%s\n", err))
			return
		}
	}

	var ag *caes.ArgGraph

	switch *fromFlag {
	case "yaml":
		ag, err = yaml.Import(inFile)
		inFile.Close()
		if err != nil {
			failImport(err)
			return
		}
	case "agxml":
		ag, err = agxml.Import(inFile)
		inFile.Close()
		if err != nil {
			failImport(err)
			return
		}
	case "aif":
		ag, err = aif.Import(inFile)
		inFile.Close()
		if err != nil {
			failImport(err)
			return
		}
	case "lkif":
		ag, err = lkif.Import(inFile)
		inFile.Close()
		if err != nil {
			failImport(err)
			return
		}
	case "caf":
		ag, err = caf.Import(inFile)
		inFile.Close()
		if err != nil {
			failImport(err)
			return
		}
	default:
		log.Fatal(fmt.Errorf("unknown or unsupported input format: %s\n", *fromFlag))
		return
	}

	// Validate the argument graph
	problems := validation.Validate(ag)

	// Print out any problems found to standard out
	for _, p := range problems {
		if p.Expression == "" {
			fmt.Fprintf(os.Stderr, "%s: %s: %s\n", p.Category, p.Id, p.Description)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s: %s: %s\n", p.Category, p.Id, p.Description, p.Expression)
		}
	}

	if len(problems) > 0 {
		// The graph is not evaluated: invalid input (docs/contracts.md, contract 3).
		os.Exit(caes.ExitStatus(caes.ErrInvalidInput))
	}
	if err := ag.CheckContext(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(caes.ExitStatus(err))
	}
	{
		// Apply the theory of the argument graph, if any, to
		// derive further arguments. A run that does not complete
		// prints no evaluation, and the exit status says why. An
		// interrupt or a termination signal cancels the run (exit 5,
		// docs/contracts.md, contract 3).
		if err := ag.InferContext(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "carneades eval: %v\n", err)
			os.Exit(caes.ExitStatus(err))
		}

		// evaluate the argument graph, using grounded semantics
		// and update the labels of the statements in the argument graph
		l := ag.GroundedLabelling()
		// fmt.Printf("labelling=%v\n", l)
		ag.ApplyLabelling(l)
		if ctx.Err() != nil {
			cancelled()
		}

		switch *toFlag {
		case "yaml":
			yaml.Export(outFile, ag)
			outFile.Close()
		case "graphml":
			err = graphml.Export(outFile, ag)
			outFile.Close()
			if err != nil {
				log.Fatal(err)
				return
			}
		case "dot":
			err = dot.Export(outFile, ag)
			outFile.Close()
			if err != nil {
				log.Fatal(err)
				return
			}
		default:
			log.Fatal(fmt.Errorf("unknown or unsupported output format: %s\n", *toFlag))
			return
		}
	}
}
