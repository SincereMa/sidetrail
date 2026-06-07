// Package sidetrail is the test surface for the supersede
// subcommand.
package sidetrail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// supersedeNewRecordJSON is the replacement record used by the
// supersede tests. It is schema-valid, points at the old
// record's scope, and carries a different id.
const supersedeNewRecordJSON = `{
  "id": "01HW4F2N8X2JZPM7Q3S5V0K1A2",
  "kind": "decision",
  "scope": "src/foo",
  "subject": "Use ULID v2 for record ids",
  "reason": "v1 had a monotonicity bug under clock skew; v2 fixes it.",
  "source_type": "human",
  "author": "tester",
  "created_at": "2026-06-01T00:00:00Z",
  "last_verified_at": "2026-06-01T00:00:00Z",
  "status": "active",
  "decided_at": "2026-06-01T00:00:00Z"
}`

// TestSupersedeSwaps confirms supersede updates the old
// record's status and superseded_by, wires the new record's
// supersedes, and writes both to the store.
func TestSupersedeSwaps(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, storeDirName)
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldID := seedRecord(t, cortexDir)

	newFile := filepath.Join(root, "new.json")
	if err := os.WriteFile(newFile, []byte(supersedeNewRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"supersede", "--root", cortexDir, "--new", newFile, oldID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("supersede returned error: %v\n%s", err, out.String())
	}

	// Fetch the old record by id and check its status +
	// superseded_by.
	getOld := newRootCmd()
	var outOld bytes.Buffer
	getOld.SetOut(&outOld)
	getOld.SetErr(&outOld)
	getOld.SetArgs([]string{"get", "--root", cortexDir, oldID})
	if err := getOld.Execute(); err != nil {
		t.Fatalf("get old: %v", err)
	}
	if !strings.Contains(outOld.String(), `"status": "superseded"`) {
		t.Errorf("old record should be marked superseded, got:\n%s", outOld.String())
	}
	if !strings.Contains(outOld.String(), `"superseded_by": "01HW4F2N8X2JZPM7Q3S5V0K1A2"`) {
		t.Errorf("old record should carry superseded_by, got:\n%s", outOld.String())
	}

	// Fetch the new record by id and check its supersedes.
	getNew := newRootCmd()
	var outNew bytes.Buffer
	getNew.SetOut(&outNew)
	getNew.SetErr(&outNew)
	getNew.SetArgs([]string{"get", "--root", cortexDir, "01HW4F2N8X2JZPM7Q3S5V0K1A2"})
	if err := getNew.Execute(); err != nil {
		t.Fatalf("get new: %v", err)
	}
	if !strings.Contains(outNew.String(), `"supersedes": "01HW4F2N8X2JZPM7Q3S5V0K1A1"`) {
		t.Errorf("new record should carry supersedes, got:\n%s", outNew.String())
	}
}

// TestSupersedeSameID confirms a replacement with the same id
// as the old record is rejected (it would be a self-supersede,
// which is meaningless and almost certainly a user error).
func TestSupersedeSameID(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, storeDirName)
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldID := seedRecord(t, cortexDir)

	sameID := strings.ReplaceAll(supersedeNewRecordJSON, "01HW4F2N8X2JZPM7Q3S5V0K1A2", oldID)
	newFile := filepath.Join(root, "new.json")
	if err := os.WriteFile(newFile, []byte(sameID), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"supersede", "--root", cortexDir, "--new", newFile, oldID})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for same-id replacement, got nil")
	}
}

// TestSupersedeDryRun confirms --dry-run does not write.
func TestSupersedeDryRun(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, storeDirName)
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldID := seedRecord(t, cortexDir)

	newFile := filepath.Join(root, "new.json")
	if err := os.WriteFile(newFile, []byte(supersedeNewRecordJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"supersede", "--root", cortexDir, "--new", newFile, "--dry-run", oldID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("dry-run returned error: %v", err)
	}
	// Old record should still be active on disk.
	entries, err := os.ReadDir(filepath.Join(cortexDir, "decisions"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("dry-run should not write, got %d files", len(entries))
	}
}
