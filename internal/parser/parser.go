package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites"`
	TestSuites []TestSuite `xml:"testsuite"`
}

type TestSuite struct {
	Name      string     `xml:"name,attr"`
	TestCases []TestCase `xml:"testcase"`
}

type TestCase struct {
	Name     string            `xml:"name,attr"`
	Failures []TestCaseFailure `xml:"failure"`
}

type TestCaseFailure struct {
	Message string `xml:"message,attr"`
}

type Failure struct {
	File    string
	Test    string
	Message string
}

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
