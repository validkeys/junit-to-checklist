package generator

import (
	"strings"
	"testing"

	"github.com/validkeys/junit-to-checklist/internal/parser"
)

func TestCleanHTMLBlocks_HTMLBlock(t *testing.T) {
	input := `Error: test failed
[36m<html>[39m
  <body>
    <div>content</div>
  </body>
[36m</html>[39m
Actual error message`

	expected := `Error: test failed
Actual error message`

	result := cleanHTMLBlocks(input)
	if result != expected {
		t.Errorf("HTML block not removed correctly.\nGot:\n%s\nExpected:\n%s", result, expected)
	}
}

func TestCleanHTMLBlocks_ANSICodes(t *testing.T) {
	input := "[90mDebug info[39m\n[1mBold text[22m\nNormal text"
	expected := "Debug info\nBold text\nNormal text"

	result := cleanHTMLBlocks(input)
	if result != expected {
		t.Errorf("ANSI codes not stripped.\nGot:\n%s\nExpected:\n%s", result, expected)
	}
}

func TestCleanHTMLBlocks_IgnoredNodes(t *testing.T) {
	input := `Error message
Ignored nodes: comments, script, style
More error details`

	expected := `Error message
More error details`

	result := cleanHTMLBlocks(input)
	if result != expected {
		t.Errorf("Ignored nodes line not removed.\nGot:\n%s\nExpected:\n%s", result, expected)
	}
}

func TestCleanHTMLBlocks_CleanInput(t *testing.T) {
	input := "Clean error message\nNo special codes\nJust plain text"
	expected := input

	result := cleanHTMLBlocks(input)
	if result != expected {
		t.Errorf("Clean input modified.\nGot:\n%s\nExpected:\n%s", result, expected)
	}
}

func TestCleanHTMLBlocks_Empty(t *testing.T) {
	input := ""
	expected := ""

	result := cleanHTMLBlocks(input)
	if result != expected {
		t.Errorf("Empty input not handled correctly. Got: '%s'", result)
	}
}

func TestCleanHTMLBlocks_MultipleBlocks(t *testing.T) {
	input := `First error
[36m<html>[39m
  <div>block1</div>
[36m</html>[39m
Between blocks
[36m<html>
  <div>block2</div>
</html>[39m
Final message`

	expected := `First error
Between blocks
Final message`

	result := cleanHTMLBlocks(input)
	if result != expected {
		t.Errorf("Multiple HTML blocks not removed.\nGot:\n%s\nExpected:\n%s", result, expected)
	}
}

func TestCleanHTMLBlocks_EmptyAfterClean(t *testing.T) {
	input := "[90m[39m\n[1m[22m\n   \n"
	expected := ""

	result := cleanHTMLBlocks(input)
	if result != expected {
		t.Errorf("Expected empty output after cleaning whitespace/ANSI. Got: '%s'", result)
	}
}

func TestGenerateChecklist_Empty(t *testing.T) {
	failures := []parser.Failure{}
	expected := "✓ No test failures found!\n"

	result := GenerateChecklist(failures)
	if result != expected {
		t.Errorf("Empty failures not handled correctly.\nGot:\n%s\nExpected:\n%s", result, expected)
	}
}

func TestGenerateChecklist_SingleFailure(t *testing.T) {
	failures := []parser.Failure{
		{
			File:    "auth.test.ts",
			Test:    "login should fail with invalid password",
			Message: "Expected status 401 but got 200",
		},
	}

	result := GenerateChecklist(failures)

	if !strings.Contains(result, "# Test Failures Checklist") {
		t.Error("Missing header")
	}
	if !strings.Contains(result, "## Instructions for AI") {
		t.Error("Missing AI instructions")
	}
	if !strings.Contains(result, "## auth.test.ts") {
		t.Error("Missing file heading")
	}
	if !strings.Contains(result, "- [ ] **auth.test.ts**") {
		t.Error("Missing checkbox")
	}
	if !strings.Contains(result, "**Test:** login should fail with invalid password") {
		t.Error("Missing test name")
	}
	if !strings.Contains(result, "**Error:**") {
		t.Error("Missing error section")
	}
	if !strings.Contains(result, "- Expected status 401 but got 200") {
		t.Error("Missing error message")
	}
	if !strings.Contains(result, "**Total failures: 1**") {
		t.Error("Missing total count")
	}
}

func TestGenerateChecklist_MultipleFailuresSameFile(t *testing.T) {
	failures := []parser.Failure{
		{
			File:    "api.test.ts",
			Test:    "GET /users should return list",
			Message: "Timeout after 5000ms",
		},
		{
			File:    "api.test.ts",
			Test:    "POST /users should create user",
			Message: "Status 500: Internal Server Error",
		},
	}

	result := GenerateChecklist(failures)

	fileHeadingCount := strings.Count(result, "## api.test.ts")
	if fileHeadingCount != 1 {
		t.Errorf("Expected 1 file heading for api.test.ts, got %d", fileHeadingCount)
	}

	checkboxCount := strings.Count(result, "- [ ] **api.test.ts**")
	if checkboxCount != 2 {
		t.Errorf("Expected 2 checkboxes, got %d", checkboxCount)
	}

	if !strings.Contains(result, "**Total failures: 2**") {
		t.Error("Wrong total count")
	}
}

func TestGenerateChecklist_MultipleFiles(t *testing.T) {
	failures := []parser.Failure{
		{
			File:    "auth.test.ts",
			Test:    "login test",
			Message: "auth error",
		},
		{
			File:    "db.test.ts",
			Test:    "connection test",
			Message: "db error",
		},
	}

	result := GenerateChecklist(failures)

	if !strings.Contains(result, "## auth.test.ts") {
		t.Error("Missing auth.test.ts heading")
	}
	if !strings.Contains(result, "## db.test.ts") {
		t.Error("Missing db.test.ts heading")
	}
}

func TestGenerateChecklist_MultilineError(t *testing.T) {
	failures := []parser.Failure{
		{
			File:    "test.ts",
			Test:    "sample test",
			Message: "Error line 1\nError line 2\nError line 3",
		},
	}

	result := GenerateChecklist(failures)

	if !strings.Contains(result, "- Error line 1") {
		t.Error("Missing error line 1")
	}
	if !strings.Contains(result, "- Error line 2") {
		t.Error("Missing error line 2")
	}
	if !strings.Contains(result, "- Error line 3") {
		t.Error("Missing error line 3")
	}
}

func TestGenerateChecklist_HTMLCleaning(t *testing.T) {
	failures := []parser.Failure{
		{
			File:    "test.ts",
			Test:    "html test",
			Message: "Error before\n[36m<html>[39m\n  <body></body>\n[36m</html>[39m\nError after",
		},
	}

	result := GenerateChecklist(failures)

	if strings.Contains(result, "<html>") || strings.Contains(result, "<body>") {
		t.Error("HTML block not cleaned from output")
	}
	if !strings.Contains(result, "Error before") {
		t.Error("Missing error before HTML block")
	}
	if !strings.Contains(result, "Error after") {
		t.Error("Missing error after HTML block")
	}
}
