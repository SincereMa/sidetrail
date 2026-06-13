package sidetrail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/SincereMa/sidetrail/internal/storage"
)

func TestSeed_Files(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "README.md")
	os.WriteFile(doc, []byte("# Project\nUse bcrypt."), 0644)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--files", doc})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --files: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestSeed_Apply(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".sidetrail")
	os.MkdirAll(root, 0755)

	records := []map[string]interface{}{
		{
			"kind":             "decision",
			"scope":            "src/auth",
			"subject":          "Use bcrypt",
			"reason":           "Security",
			"source_type":      "derived",
			"author":           "agent",
			"created_at":       "2026-01-01T00:00:00Z",
			"last_verified_at": "2026-01-01T00:00:00Z",
			"status":           "active",
		},
	}
	data, _ := json.Marshal(records)
	recordsFile := filepath.Join(dir, "records.json")
	os.WriteFile(recordsFile, data, 0644)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--apply", recordsFile, "--root", root})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --apply: %v", err)
	}

	// Verify record was written
	store := storage.NewStore(root)
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 record, got %d", len(all))
	}
}

func TestSeed_ApplyDryRun(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".sidetrail")
	os.MkdirAll(root, 0755)

	records := []map[string]interface{}{
		{
			"kind":             "decision",
			"scope":            "src/auth",
			"subject":          "Use bcrypt",
			"reason":           "Security",
			"source_type":      "derived",
			"author":           "agent",
			"created_at":       "2026-01-01T00:00:00Z",
			"last_verified_at": "2026-01-01T00:00:00Z",
			"status":           "active",
		},
	}
	data, _ := json.Marshal(records)
	recordsFile := filepath.Join(dir, "records.json")
	os.WriteFile(recordsFile, data, 0644)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--apply", recordsFile, "--dry-run", "--root", root})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --apply --dry-run: %v", err)
	}

	// Verify no records were written
	store := storage.NewStore(root)
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 records (dry run), got %d", len(all))
	}
}
