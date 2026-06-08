package sidetrail

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// TestHealthEmpty verifies that an empty store prints a message.
func TestHealthEmpty(t *testing.T) {
	dir := t.TempDir()
	initStore(t, dir)
	cmd := newRootCmd()
	cmd.SetArgs([]string{"health", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(buf.String(), "no records found") {
		t.Errorf("expected 'no records found', got %q", buf.String())
	}
}

// TestHealthCounts verifies that the health command counts
// records correctly by kind and status.
func TestHealthCounts(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	now := time.Now().UTC()
	records := []*record.Record{
		{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Kind: record.KindDecision, Scope: "src/a", Subject: "D1", Reason: "r", SourceType: record.SourceHuman, Author: "t", CreatedAt: now, LastVerifiedAt: now, Status: "active"},
		{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAW", Kind: record.KindDecision, Scope: "src/b", Subject: "D2", Reason: "r", SourceType: record.SourceHuman, Author: "t", CreatedAt: now.Add(-time.Hour), LastVerifiedAt: now, Status: "superseded"},
		{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAX", Kind: record.KindConstraint, Scope: "src/a", Subject: "C1", Reason: "r", SourceType: record.SourceHuman, Author: "t", CreatedAt: now.Add(-2 * time.Hour), LastVerifiedAt: now, Status: "active"},
	}
	for _, r := range records {
		if _, err := s.Write(r); err != nil {
			t.Fatal(err)
		}
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"health", "--root", dir})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Total records:    3") {
		t.Errorf("expected total 3, got:\n%s", out)
	}
	if !strings.Contains(out, "Unique scopes:    2") {
		t.Errorf("expected 2 scopes, got:\n%s", out)
	}
}

// TestHealthStale verifies that stale records are flagged.
func TestHealthStale(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	old := time.Now().UTC().AddDate(0, 0, -120)
	now := time.Now().UTC()
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindDecision,
		Scope:          "src/a",
		Subject:        "Old decision",
		Reason:         "r",
		SourceType:     record.SourceHuman,
		Author:         "t",
		CreatedAt:      old,
		LastVerifiedAt: old,
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	_ = now
	cmd := newRootCmd()
	cmd.SetArgs([]string{"health", "--root", dir, "--stale-days", "90"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Stale records") {
		t.Errorf("expected 'Stale records' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Old decision") {
		t.Errorf("expected stale record subject, got:\n%s", out)
	}
}

// TestHealthJSON verifies that --json emits valid JSON.
func TestHealthJSON(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	now := time.Now().UTC()
	r := &record.Record{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Kind:           record.KindDecision,
		Scope:          "src/a",
		Subject:        "D1",
		Reason:         "r",
		SourceType:     record.SourceHuman,
		Author:         "t",
		CreatedAt:      now,
		LastVerifiedAt: now,
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	cmd.SetArgs([]string{"health", "--root", dir, "--json"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var rpt healthReport
	if err := json.Unmarshal(buf.Bytes(), &rpt); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if rpt.Total != 1 {
		t.Errorf("expected total 1, got %d", rpt.Total)
	}
	if rpt.ByKind["decision"] != 1 {
		t.Errorf("expected 1 decision, got %d", rpt.ByKind["decision"])
	}
}
