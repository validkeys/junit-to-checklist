// Package main implements a JUnit XML parser that converts test failure
// reports into markdown checklists optimized for AI-assisted debugging.
package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// TestSuites maps the root <testsuites> element in JUnit XML.
type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites"`
	TestSuites []TestSuite `xml:"testsuite"`
}

// TestSuite maps a <testsuite> element containing test cases.
type TestSuite struct {
	Name      string     `xml:"name,attr"`
	TestCases []TestCase `xml:"testcase"`
}

// TestCase maps a <testcase> element with optional failure elements.
type TestCase struct {
	Name     string            `xml:"name,attr"`
	Failures []TestCaseFailure `xml:"failure"`
}

// TestCaseFailure maps the <failure> element within a test case, containing the failure message attribute.
type TestCaseFailure struct {
	Message string `xml:"message,attr"`
}

// Failure represents a single test failure with file, test name, and error message extracted from JUnit XML hierarchy.
type Failure struct {
	File    string
	Test    string
	Message string
}

// parseFile reads a JUnit XML file and extracts all test failures into a flat list.
func parseFile(path string) ([]Failure, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var suites TestSuites
	if err := xml.Unmarshal(data, &suites); err != nil {
		return nil, err
	}

	var failures []Failure
	for _, suite := range suites.TestSuites {
		for _, testCase := range suite.TestCases {
			for _, fail := range testCase.Failures {
				failures = append(failures, Failure{
					File:    suite.Name,
					Test:    testCase.Name,
					Message: fail.Message,
				})
			}
		}
	}

	return failures, nil
}

// parseDir reads all .xml files in a directory (non-recursive) and aggregates failures from each file. Parse errors for individual files are logged to stderr but do not halt processing.
func parseDir(dir string) ([]Failure, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var allFailures []Failure
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".xml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		failures, err := parseFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", path, err)
			continue
		}
		allFailures = append(allFailures, failures...)
	}

	return allFailures, nil
}

// Parse auto-detects whether path is a file or directory using os.Stat and delegates to parseFile or parseDir accordingly.
func Parse(path string) ([]Failure, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return parseDir(path)
	}
	return parseFile(path)
}
