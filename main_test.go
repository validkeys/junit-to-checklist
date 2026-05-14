package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath is set by TestMain and shared across all tests
var binaryPath string

// TestMain builds the binary once before running tests and cleans up after.
func TestMain(m *testing.M) {
	// Build binary in temp location
	tmpDir, err := os.MkdirTemp("", "junit-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "junit-to-checklist")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binary: %v\n%s\n", err, output)
		os.Exit(1)
	}

	// Run tests
	exitCode := m.Run()

	// Cleanup happens via defer
	os.Exit(exitCode)
}

func TestCLI_Directory(t *testing.T) {
	// Clean up any existing output
	outputPath := "testdata/integration/failing-tests.md"
	os.Remove(outputPath)

	// Run CLI against integration directory
	cmd := exec.Command(binaryPath, "testdata/integration")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Check output file created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created at %s", outputPath)
	}

	// Read and verify output content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Should contain both test files
	if !strings.Contains(contentStr, "components/Button.test.tsx") {
		t.Error("Missing frontend test file")
	}
	if !strings.Contains(contentStr, "services/auth.test.ts") {
		t.Error("Missing backend test file")
	}

	// Should have 3 total failures
	if !strings.Contains(contentStr, "**Total failures: 3**") {
		t.Error("Wrong total count")
	}

	// Clean up
	os.Remove(outputPath)
}

func TestCLI_SingleFile(t *testing.T) {
	// Clean up
	outputPath := "testdata/failing-tests.md"
	os.Remove(outputPath)

	// Run against single file
	cmd := exec.Command(binaryPath, "testdata/single-suite.xml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Output should be in same directory as input file
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created at %s", outputPath)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "auth.test.ts") {
		t.Error("Missing expected test file")
	}

	os.Remove(outputPath)
}

func TestCLI_CustomOutput(t *testing.T) {
	customPath := filepath.Join(os.TempDir(), "custom-checklist.md")
	defer os.Remove(customPath)

	// Run with -o flag
	cmd := exec.Command(binaryPath, "testdata/single-suite.xml", "-o", customPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Verify custom path used
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created at custom path %s", customPath)
	}
}

func TestCLI_NoXMLFiles(t *testing.T) {
	// Create empty temp directory
	emptyDir := filepath.Join(os.TempDir(), "empty-test-dir")
	os.MkdirAll(emptyDir, 0755)
	defer os.RemoveAll(emptyDir)

	outputPath := filepath.Join(emptyDir, "failing-tests.md")

	// Run against empty directory
	cmd := exec.Command(binaryPath, emptyDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	// Should report no failures
	outputStr := string(output)
	if !strings.Contains(outputStr, "No test failures found") {
		t.Errorf("Expected 'No test failures found' message, got: %s", outputStr)
	}

	// Output file should still be created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created at %s", outputPath)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if !strings.Contains(string(content), "✓ No test failures found!") {
		t.Error("Output file should contain success message")
	}
}

func TestCLI_NonExistentPath(t *testing.T) {
	// Run against non-existent path
	cmd := exec.Command(binaryPath, "nonexistent/path.xml")
	output, err := cmd.CombinedOutput()

	// Should fail
	if err == nil {
		t.Fatal("Expected error for non-existent path, got success")
	}

	// Should print error to stderr
	outputStr := string(output)
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected error message, got: %s", outputStr)
	}
}

func TestWriteAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.txt")

	// Write data atomically
	data := []byte("test content")
	if err := writeAtomic(path, data, 0644); err != nil {
		t.Fatalf("writeAtomic failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("Content mismatch. Got: %s, Want: %s", content, data)
	}

	// Verify no temp files left behind
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".junit-to-checklist-") {
			t.Errorf("Temp file not cleaned up: %s", entry.Name())
		}
	}
}

func TestValidateOutputPath_Valid(t *testing.T) {
	validPaths := []string{
		"output.md",
		"./output.md",
		"/tmp/output.md",
		"subdir/output.md",
	}

	for _, path := range validPaths {
		if err := validateOutputPath(path); err != nil {
			t.Errorf("Expected path %q to be valid, got error: %v", path, err)
		}
	}
}

func TestValidateOutputPath_Invalid(t *testing.T) {
	invalidPaths := []string{
		"../output.md",
		"subdir/../../output.md",
		"/tmp/../etc/passwd",
	}

	for _, path := range invalidPaths {
		if err := validateOutputPath(path); err == nil {
			t.Errorf("Expected path %q to be invalid, but got no error", path)
		}
	}
}

func TestCLI_OverwriteWithoutForce(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "output.md")

	// Create existing file
	if err := os.WriteFile(outputPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Run without --force should fail
	cmd := exec.Command(binaryPath, "testdata/single-suite.xml", "-o", outputPath)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when output exists without --force")
	}

	if !strings.Contains(string(output), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %s", output)
	}
}

func TestCLI_OverwriteWithForce(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "output.md")

	// Create existing file
	if err := os.WriteFile(outputPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Run with --force should succeed
	cmd := exec.Command(binaryPath, "testdata/single-suite.xml", "-o", outputPath, "--force")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed with --force: %v\nOutput: %s", err, output)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if string(content) == "existing" {
		t.Error("File was not overwritten")
	}
}

func TestCLI_Version(t *testing.T) {
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Should contain version info
	if !strings.Contains(outputStr, "junit-to-checklist") {
		t.Error("Missing program name in version output")
	}

	// Should show commit and date (even if unknown/dev)
	if !strings.Contains(outputStr, "commit:") {
		t.Error("Missing commit in version output")
	}
	if !strings.Contains(outputStr, "built:") {
		t.Error("Missing build date in version output")
	}
}
