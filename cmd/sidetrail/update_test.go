package sidetrail

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

func TestUpdateRecord(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "test-id-01",
		Kind:           record.KindDecision,
		Scope:          "src/main.go",
		Subject:        "Use feature flag",
		Reason:         "Because rollout needs to be gradual",
		SourceType:     record.SourceHuman,
		Author:         "test",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}

	updateJSON := `{"status":"archived"}`
	updateFile := filepath.Join(dir, "update.json")
	if err := os.WriteFile(updateFile, []byte(updateJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"update", "test-id-01", "--file", updateFile, "--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := s.Get("test-id-01")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "archived" {
		t.Errorf("expected status=archived, got %q", updated.Status)
	}
	if updated.Subject != "Use feature flag" {
		t.Errorf("subject should be unchanged, got %q", updated.Subject)
	}
}

func TestUpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	updateJSON := `{"status":"archived"}`
	updateFile := filepath.Join(dir, "update.json")
	if err := os.WriteFile(updateFile, []byte(updateJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"update", "nonexistent-id", "--file", updateFile, "--root", dir})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent record")
	}
}

func TestUpdateMultipleFields(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "test-id-02",
		Kind:           record.KindConstraint,
		Scope:          "src/auth.go",
		Subject:        "No new deps",
		Reason:         "Security review pending",
		SourceType:     record.SourceHuman,
		Author:         "test",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}

	updateJSON := `{"status":"superseded","reason":"Security review completed"}`
	updateFile := filepath.Join(dir, "update.json")
	if err := os.WriteFile(updateFile, []byte(updateJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"update", "test-id-02", "--file", updateFile, "--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := s.Get("test-id-02")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "superseded" {
		t.Errorf("expected status=superseded, got %q", updated.Status)
	}
	if updated.Reason != "Security review completed" {
		t.Errorf("expected reason updated, got %q", updated.Reason)
	}
}
