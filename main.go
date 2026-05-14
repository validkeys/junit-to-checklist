package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
)

var (
	// version is set via ldflags during build
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

	// Handle version flag
	if cli.Version {
		fmt.Printf("junit-to-checklist %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	// Path is required if not showing version
	if cli.Path == "" {
		fmt.Fprintf(os.Stderr, "error: <path> argument is required\n")
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the main CLI logic: parses JUnit XML from the input path, generates a markdown checklist, resolves the output file path, writes the result, and prints a success message.
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

	// Validate output path
	if err := validateOutputPath(outputPath); err != nil {
		return err
	}

	// Check if output file exists
	if err := checkOverwrite(outputPath, cli.Force); err != nil {
		return err
	}

	// Write output file
	if err := writeAtomic(outputPath, []byte(checklist), 0644); err != nil {
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

// checkOverwrite returns an error if the output file exists and force is false.
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

// validateOutputPath checks if the output path is safe to write to.
// Rejects paths containing path traversal attempts (..).
func validateOutputPath(path string) error {
	// Check for path traversal before cleaning
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid output path: path traversal not allowed")
	}

	return nil
}

// writeAtomic writes data to path atomically by writing to a temp file
// in the same directory, then renaming on success. This prevents partial
// writes from corrupting the output file.
func writeAtomic(path string, data []byte, perm os.FileMode) error {
	// Create temp file in same directory as target (ensures same filesystem)
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".junit-to-checklist-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write data
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Sync to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close before rename
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil // Prevent defer from closing again

	// Set permissions before rename
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
