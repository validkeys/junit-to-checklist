package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
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

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestCLI_Directory(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata")
	outputPath := filepath.Join(testdataDir, "integration", "failing-tests.md")
	os.Remove(outputPath)

	cmd := exec.Command(binaryPath, filepath.Join(testdataDir, "integration"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created at %s", outputPath)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "components/Button.test.tsx") {
		t.Error("Missing frontend test file")
	}
	if !strings.Contains(contentStr, "services/auth.test.ts") {
		t.Error("Missing backend test file")
	}

	if !strings.Contains(contentStr, "**Total failures: 3**") {
		t.Error("Wrong total count")
	}

	os.Remove(outputPath)
}

func TestCLI_SingleFile(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata")
	outputPath := filepath.Join(testdataDir, "failing-tests.md")
	os.Remove(outputPath)

	cmd := exec.Command(binaryPath, filepath.Join(testdataDir, "single-suite.xml"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

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
	testdataDir := filepath.Join("..", "..", "testdata")
	customPath := filepath.Join(os.TempDir(), "custom-checklist.md")
	defer os.Remove(customPath)

	cmd := exec.Command(binaryPath, filepath.Join(testdataDir, "single-suite.xml"), "-o", customPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created at custom path %s", customPath)
	}
}

func TestCLI_NoXMLFiles(t *testing.T) {
	emptyDir := filepath.Join(os.TempDir(), "empty-test-dir")
	os.MkdirAll(emptyDir, 0755)
	defer os.RemoveAll(emptyDir)

	outputPath := filepath.Join(emptyDir, "failing-tests.md")

	cmd := exec.Command(binaryPath, emptyDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "No test failures found") {
		t.Errorf("Expected 'No test failures found' message, got: %s", outputStr)
	}

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
	cmd := exec.Command(binaryPath, "nonexistent/path.xml")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("Expected error for non-existent path, got success")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "error:") {
		t.Errorf("Expected error message, got: %s", outputStr)
	}
}

func TestWriteAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "output.txt")

	data := []byte("test content")
	if err := writeAtomic(path, data, 0644); err != nil {
		t.Fatalf("writeAtomic failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("Content mismatch. Got: %s, Want: %s", content, data)
	}

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
	testdataDir := filepath.Join("..", "..", "testdata")
	outputPath := filepath.Join(t.TempDir(), "output.md")

	if err := os.WriteFile(outputPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	cmd := exec.Command(binaryPath, filepath.Join(testdataDir, "single-suite.xml"), "-o", outputPath)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when output exists without --force")
	}

	if !strings.Contains(string(output), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %s", output)
	}
}

func TestCLI_OverwriteWithForce(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata")
	outputPath := filepath.Join(t.TempDir(), "output.md")

	if err := os.WriteFile(outputPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	cmd := exec.Command(binaryPath, filepath.Join(testdataDir, "single-suite.xml"), "-o", outputPath, "--force")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI failed with --force: %v\nOutput: %s", err, output)
	}

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

	if !strings.Contains(outputStr, "junit-to-checklist") {
		t.Error("Missing program name in version output")
	}

	if !strings.Contains(outputStr, "commit:") {
		t.Error("Missing commit in version output")
	}
	if !strings.Contains(outputStr, "built:") {
		t.Error("Missing build date in version output")
	}
}
