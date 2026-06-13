// Package sidetrail is the test surface for the add subcommand.
package sidetrail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// validAddRecord is the JSON for a complete, schema-valid
// decision record. It is written to a temp file by the tests
// and fed to `sidetrail add`.
const validAddRecord = `{
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

// setupAddStore creates a temp directory tree with a .sidetrail/
// inside it, plus a record file at the leaf. It returns the
// .sidetrail/ path and the record file path.
func setupAddStore(t *testing.T) (cortexDir, recordFile string) {
	t.Helper()
	root := t.TempDir()
	cortexDir = filepath.Join(root, storeDirName)
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	recordFile = filepath.Join(root, "record.json")
	if err := os.WriteFile(recordFile, []byte(validAddRecord), 0o644); err != nil {
		t.Fatal(err)
	}
	return cortexDir, recordFile
}

// TestAddOK confirms a valid record file is written under the
// .sidetrail/ root and the id is reported on stdout.
func TestAddOK(t *testing.T) {
	cortexDir, recordFile := setupAddStore(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", "--root", cortexDir, recordFile})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add returned error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "01HW4F2N8X2JZPM7Q3S5V0K1A1") {
		t.Errorf("expected id in output, got: %q", out.String())
	}
	// The record file should now exist under decisions/.
	entries, err := os.ReadDir(filepath.Join(cortexDir, "decisions"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 record file, got %d", len(entries))
	}
}

// TestAddSchemaError confirms an invalid record file is rejected
// with the schema error surfaced on the command's error path.
func TestAddSchemaError(t *testing.T) {
	cortexDir, recordFile := setupAddStore(t)
	// Replace the file with one missing the required "kind" field
	// so the schema rejects it.
	if err := os.WriteFile(recordFile, []byte(`{
  "id": "01HW4F2N8X2JZPM7Q3S5V0K1A1",
  "scope": "src/foo",
  "subject": "Use ULID for record ids",
  "reason": "test",
  "source_type": "human",
  "author": "tester",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active"
}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", "--root", cortexDir, recordFile})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected schema error, got nil")
	}
}

// TestAddIdempotency confirms adding the same id twice errors
// the second time, leaving the file on disk unchanged.
func TestAddIdempotency(t *testing.T) {
	cortexDir, recordFile := setupAddStore(t)

	first := newRootCmd()
	var out bytes.Buffer
	first.SetOut(&out)
	first.SetErr(&out)
	first.SetArgs([]string{"add", "--root", cortexDir, recordFile})
	if err := first.Execute(); err != nil {
		t.Fatalf("first add: %v\n%s", err, out.String())
	}

	second := newRootCmd()
	out.Reset()
	second.SetOut(&out)
	second.SetErr(&out)
	second.SetArgs([]string{"add", "--root", cortexDir, recordFile})
	err := second.Execute()
	if err == nil {
		t.Fatal("expected duplicate-id error, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention duplicate, got: %v", err)
	}
}
