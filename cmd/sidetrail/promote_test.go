// Package sidetrail is the test surface for the promote subcommand.
package sidetrail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPromoteNoSeeds confirms promote reports nothing to do when
// there is no _seed/ directory.
func TestPromoteNoSeeds(t *testing.T) {
	root := t.TempDir()
	storeDir := filepath.Join(root, storeDirName)
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"promote", "--root", storeDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("promote returned error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "_seed/") || !strings.Contains(output, "sidetrail init") {
		t.Errorf("expected seed directory guidance in output, got: %q", output)
	}
}

// TestPromoteListSeeds confirms promote lists available seeds
// when called with no arguments.
func TestPromoteListSeeds(t *testing.T) {
	root := t.TempDir()
	storeDir := filepath.Join(root, storeDirName)
	seedDir := filepath.Join(storeDir, "_seed")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	seed := `{
  "id": "01TESTSEED00000000000001",
  "kind": "decision",
  "scope": "README.md",
  "subject": "test seed",
  "reason": "test",
  "source_type": "scrape",
  "author": "sidetrail init",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active"
}`
	if err := os.WriteFile(filepath.Join(seedDir, "01TESTSEED00000000000001-test-seed.json"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"promote", "--root", storeDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("promote returned error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "01TESTSEED00000000000001") {
		t.Errorf("expected seed id in output, got: %q", output)
	}
	if !strings.Contains(output, "test seed") {
		t.Errorf("expected subject in output, got: %q", output)
	}
}

// TestPromoteAll confirms --all moves every seed into the
// correct kind directory and removes it from _seed/.
func TestPromoteAll(t *testing.T) {
	root := t.TempDir()
	storeDir := filepath.Join(root, storeDirName)
	seedDir := filepath.Join(storeDir, "_seed")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	seed := `{
  "id": "01TESTSEED00000000000001",
  "kind": "decision",
  "scope": "README.md",
  "subject": "test seed",
  "reason": "test",
  "source_type": "scrape",
  "author": "sidetrail init",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active"
}`
	seedFile := filepath.Join(seedDir, "01TESTSEED00000000000001-test-seed.json")
	if err := os.WriteFile(seedFile, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"promote", "--root", storeDir, "--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("promote returned error: %v\n%s", err, out.String())
	}
	output := out.String()
	if !strings.Contains(output, "promoted 1") {
		t.Errorf("expected 'promoted 1' in output, got: %q", output)
	}
	// Seed file should be gone.
	if _, err := os.Stat(seedFile); !os.IsNotExist(err) {
		t.Error("seed file should have been removed")
	}
	// Record should now be in decisions/.
	decDir := filepath.Join(storeDir, "decisions")
	entries, err := os.ReadDir(decDir)
	if err != nil {
		t.Fatalf("read decisions/: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 record in decisions/, got %d", len(entries))
	}
}

// TestPromoteByID confirms promoting a single seed by ID prefix.
func TestPromoteByID(t *testing.T) {
	root := t.TempDir()
	storeDir := filepath.Join(root, storeDirName)
	seedDir := filepath.Join(storeDir, "_seed")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	seed := `{
  "id": "01TESTSEED00000000000001",
  "kind": "constraint",
  "scope": "billing",
  "subject": "do not change billing",
  "reason": "compliance review pending",
  "source_type": "human",
  "author": "tester",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active"
}`
	seedFile := filepath.Join(seedDir, "01TESTSEED00000000000001-do-not-change-billing.json")
	if err := os.WriteFile(seedFile, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"promote", "--root", storeDir, "01TEST"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("promote returned error: %v\n%s", err, out.String())
	}
	output := out.String()
	if !strings.Contains(output, "promoted 1") {
		t.Errorf("expected 'promoted 1' in output, got: %q", output)
	}
	// Seed should be removed, record should be in constraints/.
	if _, err := os.Stat(seedFile); !os.IsNotExist(err) {
		t.Error("seed file should have been removed")
	}
	conDir := filepath.Join(storeDir, "constraints")
	entries, err := os.ReadDir(conDir)
	if err != nil {
		t.Fatalf("read constraints/: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 record in constraints/, got %d", len(entries))
	}
}

// TestPromoteNotFound confirms an unknown ID returns an error.
func TestPromoteNotFound(t *testing.T) {
	root := t.TempDir()
	storeDir := filepath.Join(root, storeDirName)
	seedDir := filepath.Join(storeDir, "_seed")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Add a seed that won't match the query.
	seed := `{
  "id": "01TESTSEED00000000000001",
  "kind": "decision",
  "scope": "test",
  "subject": "a seed",
  "reason": "test",
  "source_type": "scrape",
  "author": "sidetrail init",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active"
}`
	if err := os.WriteFile(filepath.Join(seedDir, "seed.json"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"promote", "--root", storeDir, "NONEXISTENT"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent seed, got nil")
	}
}
