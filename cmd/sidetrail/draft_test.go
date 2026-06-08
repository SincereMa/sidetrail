package sidetrail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDraftCreatesFile verifies that `sidetrail draft decision`
// creates a draft file under _draft/.
func TestDraftCreatesFile(t *testing.T) {
	dir := t.TempDir()
	initStore(t, dir)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"draft", "decision", "--subject", "Use PostgreSQL", "--scope", "src/db", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %d: %q", len(lines), out)
	}
	draftDir := filepath.Join(dir, "_draft")
	entries, err := os.ReadDir(draftDir)
	if err != nil {
		t.Fatalf("readdir _draft: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 draft file, got %d", len(entries))
	}
	if !strings.HasSuffix(entries[0].Name(), ".json") {
		t.Errorf("draft file should end with .json, got %q", entries[0].Name())
	}
}

// TestDraftInvalidKind verifies that an unknown kind is rejected.
func TestDraftInvalidKind(t *testing.T) {
	dir := t.TempDir()
	initStore(t, dir)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"draft", "bogus", "--subject", "test", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetErr(buf)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid kind")
	}
	if !strings.Contains(err.Error(), "unknown kind") {
		t.Errorf("expected 'unknown kind' in error, got %q", err.Error())
	}
}

// TestDraftDefaultAuthor verifies that the author defaults to
// USER env var.
func TestDraftDefaultAuthor(t *testing.T) {
	old := os.Getenv("USER")
	os.Setenv("USER", "testuser")
	defer os.Setenv("USER", old)
	dir := t.TempDir()
	initStore(t, dir)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"draft", "decision", "--subject", "test", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	draftDir := filepath.Join(dir, "_draft")
	entries, _ := os.ReadDir(draftDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 draft, got %d", len(entries))
	}
	data, _ := os.ReadFile(filepath.Join(draftDir, entries[0].Name()))
	if !strings.Contains(string(data), "testuser") {
		t.Errorf("expected author 'testuser' in draft, got:\n%s", data)
	}
}
