package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "github.com/json-to-terraform/parser/internal/handler" // register handlers
	"github.com/json-to-terraform/parser/internal/diagram"
	"github.com/json-to-terraform/parser/internal/parser"
)

func main() {
	input := flag.String("input", "", "Path to diagram JSON file (or - for stdin)")
	output := flag.String("o", "output", "Output directory for Terraform files")
	noTfvars := flag.Bool("no-tfvars", false, "Do not generate terraform.tfvars")
	parallel := flag.Int("parallel", 0, "Max parallel nodes per tier (0 = auto)")
	jsonOut := flag.Bool("json", false, "Output errors as JSON")
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: parser -input <file|-> [-o output] [-no-tfvars] [-parallel N] [-json]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var data []byte
	var err error
	if *input == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(*input)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	var d diagram.Diagram
	if err := json.Unmarshal(data, &d); err != nil {
		fmt.Fprintf(os.Stderr, "parse JSON: %v\n", err)
		os.Exit(1)
	}

	opts := parser.DefaultOptions()
	opts.EmitTfvars = !*noTfvars
	opts.MaxParallel = *parallel
	p := parser.New(opts)
	result, err := p.Parse(&d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		if *jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(result)
		} else {
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "ERROR [%s] %s\n", e.NodeID, e.Message)
				if e.Suggestion != "" {
					fmt.Fprintf(os.Stderr, "  suggestion: %s\n", e.Suggestion)
				}
			}
			for _, w := range result.Warnings {
				fmt.Fprintf(os.Stderr, "WARN [%s] %s\n", w.NodeID, w.Message)
			}
		}
		os.Exit(1)
	}

	if err := os.MkdirAll(*output, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	for name, content := range result.TerraformFiles {
		path := filepath.Join(*output, name)
		if err := os.WriteFile(path, content, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Println("wrote", path)
	}
}
