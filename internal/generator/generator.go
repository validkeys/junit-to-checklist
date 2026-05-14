package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/validkeys/junit-to-checklist/internal/parser"
)

var ansiPattern = regexp.MustCompile(`\[\d+m`)

func cleanHTMLBlocks(text string) string {
	lines := strings.Split(text, "\n")
	var cleaned []string
	inHTMLBlock := false

	for _, line := range lines {
		if strings.Contains(line, "[36m<html>[39m") || strings.Contains(line, "[36m<html>") {
			inHTMLBlock = true
			continue
		}

		if inHTMLBlock && (strings.Contains(line, "[36m</html>[39m") || strings.Contains(line, "</html>[39m")) {
			inHTMLBlock = false
			continue
		}

		if inHTMLBlock {
			continue
		}

		if strings.Contains(line, "Ignored nodes: comments, script, style") {
			continue
		}

		line = ansiPattern.ReplaceAllString(line, "")

		if strings.TrimSpace(line) == "" {
			continue
		}

		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

func GenerateChecklist(failures []parser.Failure) string {
	var output strings.Builder

	if len(failures) == 0 {
		return "✓ No test failures found!\n"
	}

	output.WriteString("# Test Failures Checklist\n\n")

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

	grouped := make(map[string][]parser.Failure)
	for _, failure := range failures {
		grouped[failure.File] = append(grouped[failure.File], failure)
	}

	var files []string
	for file := range grouped {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		output.WriteString(fmt.Sprintf("## %s\n\n", file))

		for _, failure := range grouped[file] {
			output.WriteString(fmt.Sprintf("- [ ] **%s**\n\n", file))
			output.WriteString(fmt.Sprintf("  - **Test:** %s\n", failure.Test))

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

	return cleanHTMLBlocks(output.String())
}
