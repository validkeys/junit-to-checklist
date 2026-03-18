# JUnit to Checklist

A simple Node.js script that parses JUnit XML test reports and generates a markdown checklist of test failures.

## What It Does

This tool reads JUnit XML reports from your CI/CD pipeline and creates an AI-optimized markdown file (`failing-tests.md`) containing:

- A checkbox for each failed test
- The full file path for each failure
- Test names with their complete hierarchy
- Failure messages showing expected vs actual results
- Grouped by source file for easy navigation

## Installation

```bash
npm install
```

## Usage

```bash
# Parse reports from the default ./junit-reports directory
node parse-failures.js

# Or specify a custom directory
node parse-failures.js path/to/reports

# Using npm script
npm run parse
```

## Output

The script generates `failing-tests.md` with a format like:

```markdown
# Test Failures Checklist

## src/handlers/example.test.ts

- [ ] **src/handlers/example.test.ts** - should validate input correctly
      expected 5 to equal 10

**Total failures: 1**
```

## Why This Exists

- Makes it easy to track which tests need fixing
- Provides clean format for issue trackers or documentation
- AI-optimized with file paths for quick navigation
- Each failure gets its own checkbox for progress tracking
