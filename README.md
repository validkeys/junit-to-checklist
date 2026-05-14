# JUnit to Checklist

A tool that parses JUnit XML test reports and generates a markdown checklist of test failures.

## What It Does

This tool reads JUnit XML reports from your CI/CD pipeline and creates an AI-optimized markdown file (`failing-tests.md`) containing:

- A checkbox for each failed test
- The full file path for each failure
- Test names with their complete hierarchy
- Failure messages showing expected vs actual results
- Grouped by source file for easy navigation

## Installation

**Option 1: Install script (Mac/Linux)**

```bash
git clone https://github.com/validkeys/junit-to-checklist.git
cd junit-to-checklist
./install.sh
```

The script builds the binary and installs it to `/usr/local/bin`, `~/.local/bin`, or `~/bin` (whichever is writable and in your PATH).

**Option 2: Go install**

```bash
go install github.com/validkeys/junit-to-checklist@latest
```

**Option 3: Build from source**

```bash
git clone https://github.com/validkeys/junit-to-checklist.git
cd junit-to-checklist
go build -o junit-to-checklist .
# Move to your preferred location in PATH
```

**Building with version info:**

```bash
# Set version info via ldflags
go build -ldflags "\
  -X main.version=1.0.0 \
  -X main.commit=$(git rev-parse HEAD) \
  -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o junit-to-checklist .
```

## Usage

```bash
# Parse a single XML file
junit-to-checklist path/to/report.xml

# Parse all XML files in a directory
junit-to-checklist path/to/reports

# Specify custom output location
junit-to-checklist path/to/reports -o custom-output.md
```

By default, the output file `failing-tests.md` is created in the same directory as the input.

## Examples

```bash
# Parse CI reports directory
junit-to-checklist ./junit-reports

# Parse single test result
junit-to-checklist ./test-results/backend-tests.xml

# Write to specific location
junit-to-checklist ./junit-reports -o ./docs/test-failures.md
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
