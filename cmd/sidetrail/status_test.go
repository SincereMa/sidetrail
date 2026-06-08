package sidetrail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// TestStatusArchive verifies that an active record can be
// archived.
func TestStatusArchive(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindDecision,
		Scope:          "src/db",
		Subject:        "Use PostgreSQL",
		Reason:         "Team consensus",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "01ARZ3NDEKTSV4RRFFQ69G5FAV", "archived", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got, err := s.Get("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", got.Status)
	}
}

// TestStatusHide verifies that an active record can be hidden.
func TestStatusHide(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindConstraint,
		Scope:          "src/db",
		Subject:        "No raw SQL",
		Reason:         "Security policy",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "01ARZ3NDEKTSV4RRFFQ69G5FAV", "hidden", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got, err := s.Get("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "hidden" {
		t.Errorf("expected status 'hidden', got %q", got.Status)
	}
}

// TestStatusActivateFromArchived verifies that an archived
// record can be reactivated.
func TestStatusActivateFromArchived(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindDecision,
		Scope:          "src/db",
		Subject:        "Use PostgreSQL",
		Reason:         "Team consensus",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "archived",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "01ARZ3NDEKTSV4RRFFQ69G5FAV", "active", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got, err := s.Get("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "active" {
		t.Errorf("expected status 'active', got %q", got.Status)
	}
}

// TestStatusInvalidTransition verifies that an invalid
// transition is rejected.
func TestStatusInvalidTransition(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindDecision,
		Scope:          "src/db",
		Subject:        "Use PostgreSQL",
		Reason:         "Team consensus",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "superseded",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "01ARZ3NDEKTSV4RRFFQ69G5FAV", "archive", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetErr(buf)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	if !strings.Contains(err.Error(), "cannot transition") {
		t.Errorf("expected 'cannot transition' in error, got %q", err.Error())
	}
}

// TestStatusDryRun verifies that --dry-run does not modify
// the record.
func TestStatusDryRun(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindDecision,
		Scope:          "src/db",
		Subject:        "Use PostgreSQL",
		Reason:         "Team consensus",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "01ARZ3NDEKTSV4RRFFQ69G5FAV", "archived", "--dry-run", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got, err := s.Get("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "active" {
		t.Errorf("dry-run should not modify record, got status %q", got.Status)
	}
	if !strings.Contains(buf.String(), "would change") {
		t.Errorf("expected dry-run output, got %q", buf.String())
	}
}

// TestStatusNotFound verifies that a missing record is
// reported as an error.
func TestStatusNotFound(t *testing.T) {
	dir := t.TempDir()
	initStore(t, dir)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "NOPE", "archive", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetErr(buf)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing record")
	}
	if !strings.Contains(err.Error(), "no record") {
		t.Errorf("expected 'no record' in error, got %q", err.Error())
	}
}

// TestStatusPrefixMatch verifies that status works with a
// unique prefix.
func TestStatusPrefixMatch(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindDecision,
		Scope:          "src/db",
		Subject:        "Use PostgreSQL",
		Reason:         "Team consensus",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "01ARZ3N", "archived", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got, err := s.Get("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", got.Status)
	}
}

// initStore creates a .sidetrail/ directory at root.
func initStore(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".sidetrail"), 0o755); err != nil {
		t.Fatal(err)
	}
}
