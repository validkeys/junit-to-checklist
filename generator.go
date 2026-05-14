package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ansiPattern matches ANSI color codes in the form [<digits>m
var ansiPattern = regexp.MustCompile(`\[\d+m`)

// cleanHTMLBlocks strips HTML blocks (between <html> tags), ANSI color codes, and 'Ignored nodes' lines from text. Lines that are empty after cleaning are removed. Returns cleaned text with remaining lines joined.
func cleanHTMLBlocks(text string) string {
	lines := strings.Split(text, "\n")
	var cleaned []string
	inHTMLBlock := false

	for _, line := range lines {
		// Detect HTML block start
		if strings.Contains(line, "[36m<html>[39m") || strings.Contains(line, "[36m<html>") {
			inHTMLBlock = true
			continue
		}

		// Detect HTML block end
		if inHTMLBlock && (strings.Contains(line, "[36m</html>[39m") || strings.Contains(line, "</html>[39m")) {
			inHTMLBlock = false
			continue
		}

		// Skip lines inside HTML block
		if inHTMLBlock {
			continue
		}

		// Skip "Ignored nodes" lines
		if strings.Contains(line, "Ignored nodes: comments, script, style") {
			continue
		}

		// Strip ANSI codes
		line = ansiPattern.ReplaceAllString(line, "")

		// Skip empty/whitespace lines after cleaning
		if strings.TrimSpace(line) == "" {
			continue
		}

		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// generateChecklist generates a markdown checklist from test failures. Failures are grouped by file with deterministic (sorted) ordering. Empty failure list returns a success message. Output is passed through cleanHTMLBlocks before returning.
func generateChecklist(failures []Failure) string {
	var output strings.Builder

	if len(failures) == 0 {
		return "✓ No test failures found!\n"
	}

	output.WriteString("# Test Failures Checklist\n\n")

	// AI instructions section
	output.WriteString("## Instructions for AI\n\n")
	output.WriteString("**Work through these tests ONE AT A TIME.** After fixing each test:\n")
	output.WriteString("1. Update the checkbox from `- [ ]` to `- [x]`\n")
	output.WriteString("2. Move to the next test\n")
	output.WriteString("3. Do not skip ahead or work on multiple tests simultaneously\n\n")

	output.WriteString("### How to Run Tests\n\n")
	output.WriteString("To run a specific test file:\n")
	output.WriteString("```bash\n")
	output.WriteString("cd <package-root-directory>\n")
	output.WriteString("npx vitest path/to/file\n")
	output.WriteString("```\n\n")
	output.WriteString("Example:\n")
	output.WriteString("```bash\n")
	output.WriteString("npx vitest src/handlers/example.test.ts\n")
	output.WriteString("```\n\n")

	output.WriteString("### Be Systematic About Debugging\n\n")
	output.WriteString("**First:** State the current problem -- whether it's an error or otherwise.\n\n")
	output.WriteString("**Then:**\n\n")
	output.WriteString("1. Isolate a failing test with `.only`\n")
	output.WriteString("2. Add comprehensive diagnostic console logging to the entire call stack\n")
	output.WriteString("3. Run the test\n")
	output.WriteString("4. Trace the diagnostic logging\n")
	output.WriteString("5. Determine the root cause\n")
	output.WriteString("6. Present the root cause to the user along with a proposed solution. If there are several options, present the options for solving.\n\n")
	output.WriteString("---\n\n")

	// Group failures by file
	grouped := make(map[string][]Failure)
	for _, failure := range failures {
		grouped[failure.File] = append(grouped[failure.File], failure)
	}

	// Sort file names for deterministic output
	var files []string
	for file := range grouped {
		files = append(files, file)
	}
	sort.Strings(files)

	// Generate checklist per file
	for _, file := range files {
		output.WriteString(fmt.Sprintf("## %s\n\n", file))

		for _, failure := range grouped[file] {
			// Checkbox uses file name (matches JS implementation line 229)
			output.WriteString(fmt.Sprintf("- [ ] **%s**\n\n", file))
			output.WriteString(fmt.Sprintf("  - **Test:** %s\n", failure.Test))

			// Split message into lines and filter empty
			messageLines := strings.Split(failure.Message, "\n")
			var filteredLines []string
			for _, line := range messageLines {
				if strings.TrimSpace(line) != "" {
					filteredLines = append(filteredLines, line)
				}
			}

			if len(filteredLines) > 0 {
				output.WriteString("  - **Error:**\n")
				for _, line := range filteredLines {
					output.WriteString(fmt.Sprintf("    - %s\n", line))
				}
			}

			output.WriteString("\n")
		}
	}

	output.WriteString(fmt.Sprintf("\n**Total failures: %d**\n", len(failures)))

	// Clean HTML blocks before returning
	return cleanHTMLBlocks(output.String())
}
