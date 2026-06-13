package seed

import (
	"testing"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

func TestDetectConflicts_NoExisting(t *testing.T) {
	store := storage.NewStore(t.TempDir())
	candidates := []*record.Record{
		{ID: "test-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Security", SourceType: record.SourceHuman, Author: "test", Status: "active"},
	}
	conflicts, nonConflicting := DetectConflicts(store, candidates)
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
	if len(nonConflicting) != 1 {
		t.Errorf("expected 1 non-conflicting, got %d", len(nonConflicting))
	}
}

func TestDetectConflicts_ExactMatch(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore(dir)
	existing := &record.Record{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Old decision", SourceType: record.SourceHuman, Author: "test", Status: "active"}
	if _, err := store.Write(existing); err != nil {
		t.Fatal(err)
	}
	candidates := []*record.Record{
		{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "New decision", SourceType: record.SourceHuman, Author: "test", Status: "active"},
	}
	conflicts, nonConflicting := DetectConflicts(store, candidates)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}
	if len(nonConflicting) != 0 {
		t.Errorf("expected 0 non-conflicting, got %d", len(nonConflicting))
	}
	if conflicts[0].Existing.ID != "existing-1" {
		t.Errorf("expected existing ID existing-1, got %s", conflicts[0].Existing.ID)
	}
}

func TestDetectConflicts_DifferentKind(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore(dir)
	existing := &record.Record{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"}
	if _, err := store.Write(existing); err != nil {
		t.Fatal(err)
	}
	candidates := []*record.Record{
		{ID: "new-1", Kind: record.KindConstraint, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Constraint", SourceType: record.SourceHuman, Author: "test", Status: "active"},
	}
	conflicts, nonConflicting := DetectConflicts(store, candidates)
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts (different kind), got %d", len(conflicts))
	}
	if len(nonConflicting) != 1 {
		t.Errorf("expected 1 non-conflicting, got %d", len(nonConflicting))
	}
}

func TestDetectConflicts_ScopeOverlap(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore(dir)
	existing := &record.Record{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Password hashing", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"}
	if _, err := store.Write(existing); err != nil {
		t.Fatal(err)
	}
	candidates := []*record.Record{
		{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth/login.go", Subject: "Password hashing", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
	}
	conflicts, _ := DetectConflicts(store, candidates)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict (scope overlap), got %d", len(conflicts))
	}
}

func TestDetectConflicts_SubjectPrefix(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewStore(dir)
	existing := &record.Record{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt for hashing", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"}
	if _, err := store.Write(existing); err != nil {
		t.Fatal(err)
	}
	candidates := []*record.Record{
		{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
	}
	conflicts, _ := DetectConflicts(store, candidates)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict (subject prefix), got %d", len(conflicts))
	}
}
