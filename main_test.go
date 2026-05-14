package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Directory(t *testing.T) {
	// Build binary
	cmd := exec.Command("go", "build", "-o", "junit-to-checklist", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("junit-to-checklist")

	// Clean up any existing output
	outputPath := "testdata/integration/failing-tests.md"
	os.Remove(outputPath)

	// Run CLI against integration directory
	cmd = exec.Command("./junit-to-checklist", "testdata/integration")
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
	cmd := exec.Command("go", "build", "-o", "junit-to-checklist", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("junit-to-checklist")

	// Clean up
	outputPath := "testdata/failing-tests.md"
	os.Remove(outputPath)

	// Run against single file
	cmd = exec.Command("./junit-to-checklist", "testdata/single-suite.xml")
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
	cmd := exec.Command("go", "build", "-o", "junit-to-checklist", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("junit-to-checklist")

	customPath := filepath.Join(os.TempDir(), "custom-checklist.md")
	defer os.Remove(customPath)

	// Run with -o flag
	cmd = exec.Command("./junit-to-checklist", "testdata/single-suite.xml", "-o", customPath)
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
	cmd := exec.Command("go", "build", "-o", "junit-to-checklist", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("junit-to-checklist")

	// Create empty temp directory
	emptyDir := filepath.Join(os.TempDir(), "empty-test-dir")
	os.MkdirAll(emptyDir, 0755)
	defer os.RemoveAll(emptyDir)

	outputPath := filepath.Join(emptyDir, "failing-tests.md")

	// Run against empty directory
	cmd = exec.Command("./junit-to-checklist", emptyDir)
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
	cmd := exec.Command("go", "build", "-o", "junit-to-checklist", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("junit-to-checklist")

	// Run against non-existent path
	cmd = exec.Command("./junit-to-checklist", "nonexistent/path.xml")
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
