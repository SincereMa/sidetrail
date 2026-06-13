package sidetrail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

func TestSeed_Files(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "README.md")
	if err := os.WriteFile(doc, []byte("# Project\nUse bcrypt."), 0644); err != nil {
		t.Fatal(err)
	}

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
	if err := os.WriteFile(recordsFile, data, 0644); err != nil {
		t.Fatal(err)
	}

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
	if err := os.WriteFile(recordsFile, data, 0644); err != nil {
		t.Fatal(err)
	}

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

func TestSeed_MutualExclusivity(t *testing.T) {
	dir := t.TempDir()

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
	if err := os.WriteFile(recordsFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("both flags", func(t *testing.T) {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"seed", "--files", "*.md", "--apply", recordsFile})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error when both --files and --apply are provided")
		}
	})

	t.Run("no flags", func(t *testing.T) {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"seed"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error when neither --files nor --apply is provided")
		}
	})
}

func TestSeed_FilesJSON(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "README.md")
	if err := os.WriteFile(doc, []byte("# Project\nUse bcrypt."), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--files", doc, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --files --json: %v", err)
	}

	var output map[string]string
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if _, ok := output["prompt"]; !ok {
		t.Error("JSON output missing 'prompt' key")
	}
}

func TestSeed_ApplyJSON(t *testing.T) {
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
	if err := os.WriteFile(recordsFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--apply", recordsFile, "--root", root, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --apply --json: %v", err)
	}

	var output struct {
		DryRun         bool            `json:"dry_run"`
		Conflicts      []interface{}   `json:"conflicts"`
		NonConflicting []interface{}   `json:"non_conflicting"`
	}
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if output.DryRun {
		t.Error("expected dry_run=false")
	}
	if len(output.NonConflicting) != 1 {
		t.Errorf("expected 1 non-conflicting record, got %d", len(output.NonConflicting))
	}
}

func TestSeed_ApplyConflict(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".sidetrail")
	os.MkdirAll(root, 0755)

	// Write an existing record to the store
	store := storage.NewStore(root)
	existing := &record.Record{
		ID:                 "existing-id",
		Kind:               record.KindDecision,
		Scope:              "src/auth",
		Subject:            "Use bcrypt",
		Reason:             "Security",
		SourceType:         record.SourceDerived,
		Author:             "agent",
		CreatedAt:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		LastVerifiedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:             "active",
	}
	if _, err := store.Write(existing); err != nil {
		t.Fatal(err)
	}

	// Candidate that conflicts (same kind + overlapping scope + similar subject)
	candidates := []map[string]interface{}{
		{
			"kind":             "decision",
			"scope":            "src/auth/utils",
			"subject":          "Use bcrypt for hashing",
			"reason":           "Updated security requirement",
			"source_type":      "derived",
			"author":           "agent",
			"created_at":       "2026-02-01T00:00:00Z",
			"last_verified_at": "2026-02-01T00:00:00Z",
			"status":           "active",
		},
	}
	data, _ := json.Marshal(candidates)
	recordsFile := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(recordsFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--apply", recordsFile, "--root", root})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --apply: %v", err)
	}

	// Verify conflict was reported
	output := buf.String()
	if len(output) == 0 {
		t.Fatal("expected output about conflicts")
	}

	// Verify no new records were written (only the original exists)
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 record (conflict should prevent write), got %d", len(all))
	}
}
