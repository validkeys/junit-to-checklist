#!/usr/bin/env node

/**
 * @fileoverview JUnit XML Report Parser
 *
 * Parses JUnit XML test reports and generates a markdown checklist of test failures.
 * This tool is designed to help developers and AI assistants systematically work through
 * test failures by providing a clear, trackable format with file paths and error messages.
 *
 * @author junit-to-checklist contributors
 * @license MIT
 */

const fs = require('fs');
const path = require('path');
const { parseStringPromise } = require('xml2js');

/**
 * Parses JUnit XML reports from a directory and extracts all test failures.
 *
 * Reads all XML files in the specified directory, parses them as JUnit test reports,
 * and extracts information about failed tests including test suite names, test names,
 * and failure messages.
 *
 * @param {string} directory - Path to directory containing JUnit XML report files
 * @returns {Promise<Array<{file: string, test: string, message: string}>>} Array of failure objects containing:
 *   - file: The test suite name (typically the source file path)
 *   - test: The full test name/description
 *   - message: The failure message from the test framework
 * @throws {Error} Logs parsing errors to console.error but continues processing other files
 *
 * @example
 * const failures = await parseJunitReports('./junit-reports');
 * // Returns: [
 * //   {
 * //     file: 'src/example.test.ts',
 * //     test: 'should validate input',
 * //     message: 'expected 5 to equal 10'
 * //   }
 * // ]
 */
async function parseJunitReports(directory) {
  const failures = [];

  // Read all XML files in the directory
  const files = fs.readdirSync(directory).filter(file => file.endsWith('.xml'));

  for (const file of files) {
    const filePath = path.join(directory, file);
    const xmlContent = fs.readFileSync(filePath, 'utf-8');

    try {
      const result = await parseStringPromise(xmlContent);
      const testsuites = result.testsuites?.testsuite || [];

      for (const testsuite of testsuites) {
        const suiteName = testsuite.$.name;
        const testcases = testsuite.testcase || [];

        for (const testcase of testcases) {
          if (testcase.failure) {
            const testName = testcase.$.name;
            const failureMessage = testcase.failure[0].$.message || 'No message';
            
            failures.push({
              file: suiteName,
              test: testName,
              message: failureMessage
            });
          }
        }
      }
    } catch (error) {
      console.error(`Error parsing ${file}:`, error.message);
    }
  }

  return failures;
}

/**
 * Removes verbose HTML DOM dumps and ANSI color codes from failure messages.
 *
 * Test frameworks sometimes include large HTML DOM dumps in failure messages, which
 * can make the output difficult to read and unnecessarily large. This function strips
 * out HTML blocks (identified by <html> tags with ANSI color codes) while preserving
 * other useful error information.
 *
 * @param {string} text - The raw text containing potential HTML blocks and ANSI codes
 * @returns {string} Cleaned text with HTML blocks and ANSI codes removed
 *
 * @example
 * const raw = "Error: test failed\n[36m<html>[39m\n<body>...</body>\n[36m</html>[39m";
 * const clean = cleanHtmlBlocks(raw);
 * // Returns: "Error: test failed"
 */
function cleanHtmlBlocks(text) {
  // Remove verbose HTML DOM dumps with ANSI color codes
  const lines = text.split('\n');
  const cleaned = [];
  let inHtmlBlock = false;
  let skippedBlocks = 0;

  for (let i = 0; i < lines.length; i++) {
    let line = lines[i];

    // Check if this line starts an HTML block (contains ANSI codes and <html>)
    if (line.includes('[36m<html>[39m') || line.includes('[36m<html>')) {
      inHtmlBlock = true;
      skippedBlocks++;
      continue;
    }

    // Check if this line ends an HTML block
    if (inHtmlBlock && (line.includes('[36m</html>[39m') || line.includes('</html>[39m'))) {
      inHtmlBlock = false;
      continue;
    }

    // Skip lines that are part of an HTML block
    if (inHtmlBlock) {
      continue;
    }

    // Also skip lines that contain "Ignored nodes: comments, script, style" which precede HTML blocks
    if (line.includes('Ignored nodes: comments, script, style')) {
      continue;
    }

    // Clean up trailing ANSI codes and orphaned lines
    // Remove common ANSI color codes: [90m, [1m, [22m, [39m, etc.
    line = line.replace(/\[\d+m/g, '');

    // Skip lines that are now empty or only whitespace after ANSI removal
    if (line.trim() === '') {
      continue;
    }

    // Keep all other lines
    cleaned.push(line);
  }

  if (skippedBlocks > 0) {
    console.log(`  Cleaned ${skippedBlocks} HTML block(s) from failure messages`);
  }

  return cleaned.join('\n');
}

/**
 * Generates a markdown checklist from an array of test failures.
 *
 * Creates a structured markdown document with:
 * - Instructions for AI assistants and developers
 * - Grouped test failures by source file
 * - Checkboxes for tracking progress
 * - Full error messages and test names
 * - Total failure count
 *
 * The generated format is optimized for use with AI coding assistants, encouraging
 * systematic, one-at-a-time fixing of test failures.
 *
 * @param {Array<{file: string, test: string, message: string}>} failures - Array of test failure objects
 * @returns {string} Markdown-formatted checklist with HTML blocks cleaned
 *
 * @example
 * const failures = [
 *   { file: 'test.ts', test: 'should work', message: 'failed' }
 * ];
 * const markdown = generateChecklist(failures);
 * // Returns formatted markdown with checkboxes and instructions
 */
function generateChecklist(failures) {
  let output = '';

  if (failures.length === 0) {
    output = '✓ No test failures found!\n';
    return output;
  }

  output += '# Test Failures Checklist\n\n';

  // Add AI instructions
  output += '## Instructions for AI\n\n';
  output += '**Work through these tests ONE AT A TIME.** After fixing each test:\n';
  output += '1. Update the checkbox from `- [ ]` to `- [x]`\n';
  output += '2. Move to the next test\n';
  output += '3. Do not skip ahead or work on multiple tests simultaneously\n\n';

  output += '### How to Run Tests\n\n';
  output += 'To run a specific test file:\n';
  output += '```bash\n';
  output += 'cd <package-root-directory>\n';
  output += 'npx vitest path/to/file\n';
  output += '```\n\n';
  output += 'Example:\n';
  output += '```bash\n';
  output += 'npx vitest src/handlers/example.test.ts\n';
  output += '```\n\n';

  output += '### Be Systematic About Debugging\n\n';
  output += '**First:** State the current problem -- whether it\'s an error or otherwise.\n\n';
  output += '**Then:**\n\n';
  output += '1. Isolate a failing test with `.only`\n';
  output += '2. Add comprehensive diagnostic console logging to the entire call stack\n';
  output += '3. Run the test\n';
  output += '4. Trace the diagnostic logging\n';
  output += '5. Determine the root cause\n';
  output += '6. Present the root cause to the user along with a proposed solution. If there are several options, present the options for solving.\n\n';
  output += '---\n\n';

  // Group failures by file
  const groupedFailures = failures.reduce((acc, failure) => {
    if (!acc[failure.file]) {
      acc[failure.file] = [];
    }
    acc[failure.file].push(failure);
    return acc;
  }, {});

  // Output checklist
  for (const [file, fileFailures] of Object.entries(groupedFailures)) {
    output += `## ${file}\n\n`;

    fileFailures.forEach(failure => {
      // Clean the message and format it properly
      const messageLines = failure.message.split('\n').filter(line => line.trim());

      output += `- [ ] **${file}**\n\n`;
      output += `  - **Test:** ${failure.test}\n`;

      // Add error message lines as bullet points
      if (messageLines.length > 0) {
        output += `  - **Error:**\n`;
        for (let i = 0; i < messageLines.length; i++) {
          output += `    - ${messageLines[i]}\n`;
        }
      }

      // Add blank line after each test
      output += '\n';
    });
  }

  output += `\n**Total failures: ${failures.length}**\n`;

  // Clean HTML blocks before returning
  return cleanHtmlBlocks(output);
}

/**
 * Main entry point for the CLI tool.
 *
 * Orchestrates the parsing of JUnit reports and generation of the markdown checklist.
 * Accepts an optional directory path as a command-line argument, defaulting to
 * './junit-reports'. Writes the output to 'failing-tests.md' in the current directory.
 *
 * @async
 * @returns {Promise<void>} Resolves when the checklist has been generated and written to disk
 * @throws {Error} Exits with code 1 if the reports directory doesn't exist
 *
 * @example
 * // Run with default directory
 * node parse-failures.js
 *
 * @example
 * // Run with custom directory
 * node parse-failures.js ./custom-reports
 */
async function main() {
  const reportsDir = process.argv[2] || './junit-reports';
  const outputFile = 'failing-tests.md';

  if (!fs.existsSync(reportsDir)) {
    console.error(`Error: Directory ${reportsDir} does not exist`);
    process.exit(1);
  }

  console.log(`Parsing JUnit reports from: ${reportsDir}`);
  
  const failures = await parseJunitReports(reportsDir);
  const checklistOutput = generateChecklist(failures);
  
  // Write to file
  fs.writeFileSync(outputFile, checklistOutput, 'utf-8');
  console.log(`✓ Generated ${outputFile} with ${failures.length} failure(s)`);
}

main().catch(console.error);
