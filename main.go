package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

var cli struct {
	Path   string `arg:"" name:"path" help:"XML file or directory containing JUnit XML reports" type:"path"`
	Output string `short:"o" help:"Output file path (default: failing-tests.md in input directory)" type:"path"`
}

func main() {
	kong.Parse(&cli)

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse failures
	failures, err := Parse(cli.Path)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", cli.Path, err)
	}

	// Generate checklist
	checklist := generateChecklist(failures)

	// Resolve output path
	outputPath := cli.Output
	if outputPath == "" {
		// Determine base directory
		info, err := os.Stat(cli.Path)
		if err != nil {
			return fmt.Errorf("failed to stat input path: %w", err)
		}

		if info.IsDir() {
			outputPath = filepath.Join(cli.Path, "failing-tests.md")
		} else {
			outputPath = filepath.Join(filepath.Dir(cli.Path), "failing-tests.md")
		}
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(checklist), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	// Print success message
	if len(failures) == 0 {
		fmt.Println("✓ No test failures found!")
	} else {
		fmt.Printf("Generated %s with %d failure(s)\n", outputPath, len(failures))
	}

	return nil
}
