package parser

import (
	"testing"
)

func TestParseFile_SingleSuite(t *testing.T) {
	failures, err := parseFile("../../testdata/single-suite.xml")
	if err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}

	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}

	f := failures[0]
	if f.File != "auth.test.ts" {
		t.Errorf("expected File='auth.test.ts', got '%s'", f.File)
	}
	if f.Test != "login should fail with invalid password" {
		t.Errorf("expected Test='login should fail with invalid password', got '%s'", f.Test)
	}
	if f.Message != "Expected status 401 but got 200" {
		t.Errorf("expected Message='Expected status 401 but got 200', got '%s'", f.Message)
	}
}

func TestParseFile_NoFailures(t *testing.T) {
	failures, err := parseFile("../../testdata/no-failures.xml")
	if err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}

	if len(failures) != 0 {
		t.Errorf("expected 0 failures, got %d", len(failures))
	}
}

func TestParseFile_Malformed(t *testing.T) {
	_, err := parseFile("../../testdata/malformed.xml")
	if err == nil {
		t.Fatal("expected error for malformed XML, got nil")
	}
}

func TestParseFile_MultiSuite(t *testing.T) {
	failures, err := parseFile("../../testdata/multi-suite.xml")
	if err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}

	if len(failures) != 2 {
		t.Fatalf("expected 2 failures, got %d", len(failures))
	}

	f0 := failures[0]
	if f0.File != "api.test.ts" {
		t.Errorf("expected failures[0].File='api.test.ts', got '%s'", f0.File)
	}
	if f0.Test != "GET /users should return list" {
		t.Errorf("expected failures[0].Test='GET /users should return list', got '%s'", f0.Test)
	}
	if f0.Message != "Timeout after 5000ms" {
		t.Errorf("expected failures[0].Message='Timeout after 5000ms', got '%s'", f0.Message)
	}

	f1 := failures[1]
	if f1.File != "db.test.ts" {
		t.Errorf("expected failures[1].File='db.test.ts', got '%s'", f1.File)
	}
	if f1.Test != "connection pool should initialize" {
		t.Errorf("expected failures[1].Test='connection pool should initialize', got '%s'", f1.Test)
	}
	if f1.Message != "ECONNREFUSED 127.0.0.1:5432" {
		t.Errorf("expected failures[1].Message='ECONNREFUSED 127.0.0.1:5432', got '%s'", f1.Message)
	}
}

func TestParseDir(t *testing.T) {
	failures, err := parseDir("../../testdata")
	if err != nil {
		t.Fatalf("parseDir failed: %v", err)
	}

	if len(failures) != 3 {
		t.Errorf("expected 3 failures from testdata, got %d", len(failures))
	}
}

func TestParseDir_Empty(t *testing.T) {
	emptyDir := t.TempDir()

	failures, err := parseDir(emptyDir)
	if err != nil {
		t.Fatalf("parseDir failed: %v", err)
	}

	if len(failures) != 0 {
		t.Errorf("expected 0 failures from empty dir, got %d", len(failures))
	}
}

func TestParse_File(t *testing.T) {
	failures, err := Parse("../../testdata/single-suite.xml")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(failures) != 1 {
		t.Errorf("expected 1 failure, got %d", len(failures))
	}
}

func TestParse_Directory(t *testing.T) {
	failures, err := Parse("../../testdata")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(failures) != 3 {
		t.Errorf("expected 3 failures from testdata, got %d", len(failures))
	}
}

func TestParse_NonExistent(t *testing.T) {
	_, err := Parse("testdata/nonexistent.xml")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}
