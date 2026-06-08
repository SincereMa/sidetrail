// Package sidetrail is the test surface for the validate subcommand.
package sidetrail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// validRecordJSON is a complete, schema-valid decision record.
const validRecordJSON = `{
  "id": "01HW4F2N8X2JZPM7Q3S5V0K1A1",
  "kind": "decision",
  "scope": "src/foo",
  "subject": "Use ULID for record ids",
  "reason": "Lexicographic time-sortable ids make logs and listings easier.",
  "source_type": "human",
  "author": "tester",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active",
  "decided_at": "2026-01-01T00:00:00Z"
}`

// invalidRecordJSON is missing the kind-specific "decided_at" field.
const invalidRecordJSON = `{
  "id": "01HW4F2N8X2JZPM7Q3S5V0K1A1",
  "kind": "decision",
  "scope": "src/foo",
  "subject": "Use ULID for record ids",
  "reason": "test",
  "source_type": "human",
  "author": "tester",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active"
}`

// TestValidateOK confirms a valid record file passes validation.
func TestValidateOK(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "good.json")
	if err := os.WriteFile(f, []byte(validRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"validate", f})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("validate returned error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok") {
		t.Errorf("expected 'ok' in output, got: %q", out.String())
	}
}

// TestValidateSchemaError confirms an invalid record file is
// reported as failing validation.
func TestValidateSchemaError(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(f, []byte(invalidRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"validate", f})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(out.String(), "fail") {
		t.Errorf("expected 'fail' in output, got: %q", out.String())
	}
}

// TestValidateJSONOutput confirms --json emits a JSON array.
func TestValidateJSONOutput(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(good, []byte(validRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte(invalidRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"validate", "--json", good, bad})
	_ = cmd.Execute()

	var results []validateResult
	if err := json.Unmarshal(out.Bytes(), &results); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nraw: %s", err, out.String())
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].OK {
		t.Error("first result should be OK")
	}
	if results[1].OK {
		t.Error("second result should not be OK")
	}
}

// TestValidateFileNotFound confirms a missing file is reported
// as an error.
func TestValidateFileNotFound(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"validate", "/nonexistent/file.json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// TestValidateMultipleFiles confirms multiple files are all
// validated and reported.
func TestValidateMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(good, []byte(validRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte(invalidRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"validate", good, bad})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error because one file is invalid, got nil")
	}
	output := out.String()
	if !strings.Contains(output, "ok") || !strings.Contains(output, "fail") {
		t.Errorf("expected both ok and fail in output, got: %q", output)
	}
}
