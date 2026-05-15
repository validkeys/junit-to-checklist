package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/validkeys/junit-to-checklist/internal/generator"
	"github.com/validkeys/junit-to-checklist/internal/parser"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var cli struct {
	Path    string `arg:"" name:"path" help:"XML file or directory containing JUnit XML reports" type:"path" optional:""`
	Output  string `short:"o" help:"Output file path (default: failing-tests.md in input directory)" type:"path"`
	Force   bool   `short:"f" help:"Overwrite output file if it exists without prompting"`
	Version bool   `short:"v" name:"version" help:"Show version information"`
}

func main() {
	kong.Parse(&cli)

	if cli.Version {
		fmt.Printf("junit-to-checklist %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	if cli.Path == "" {
		fmt.Fprintf(os.Stderr, "error: <path> argument is required\n")
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	failures, err := parser.Parse(cli.Path)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", cli.Path, err)
	}

	checklist := generator.GenerateChecklist(failures)

	outputPath := cli.Output
	if outputPath == "" {
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

	if err := validateOutputPath(outputPath); err != nil {
		return err
	}

	if err := checkOverwrite(outputPath, cli.Force); err != nil {
		return err
	}

	if err := writeAtomic(outputPath, []byte(checklist), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if len(failures) == 0 {
		fmt.Println("✓ No test failures found!")
	} else {
		fmt.Printf("Generated %s with %d failure(s)\n", outputPath, len(failures))
	}

	return nil
}

func checkOverwrite(path string, force bool) error {
	if force {
		return nil
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("output file %s already exists (use --force to overwrite)", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check output file: %w", err)
	}

	return nil
}

func validateOutputPath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid output path: path traversal not allowed")
	}

	return nil
}

func writeAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".junit-to-checklist-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil

	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
