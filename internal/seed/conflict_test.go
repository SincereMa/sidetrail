package seed

import (
	"testing"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

func TestDetectConflicts(t *testing.T) {
	tests := []struct {
		name             string
		existing         []*record.Record
		candidates       []*record.Record
		wantConflicts    int
		wantNonConflict  int
		wantExistingID   string
	}{
		{
			name:            "no existing records",
			existing:        nil,
			candidates:      []*record.Record{
				{ID: "test-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Security", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   0,
			wantNonConflict: 1,
		},
		{
			name: "exact match",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Old decision", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "New decision", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   1,
			wantNonConflict: 0,
			wantExistingID:  "existing-1",
		},
		{
			name: "different kind no conflict",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindConstraint, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Constraint", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   0,
			wantNonConflict: 1,
		},
		{
			name: "scope overlap child",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Password hashing", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth/login.go", Subject: "Password hashing", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   1,
			wantNonConflict: 0,
		},
		{
			name: "subject prefix match",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt for hashing", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   1,
			wantNonConflict: 0,
		},
		{
			name: "empty candidate subject no conflict",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   0,
			wantNonConflict: 1,
		},
		{
			name: "empty existing subject no conflict",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "Use bcrypt", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   0,
			wantNonConflict: 1,
		},
		{
			name: "both subjects empty conflict",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "src/auth", Subject: "", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   1,
			wantNonConflict: 0,
			wantExistingID:  "existing-1",
		},
		{
			name: "empty scopes with different subjects",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "", Subject: "Use bcrypt", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "", Subject: "Use bcrypt", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   1,
			wantNonConflict: 0,
			wantExistingID:  "existing-1",
		},
		{
			name: "empty scopes different subjects no conflict",
			existing: []*record.Record{
				{ID: "existing-1", Kind: record.KindDecision, Scope: "", Subject: "Use bcrypt", Reason: "Old", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			candidates: []*record.Record{
				{ID: "new-1", Kind: record.KindDecision, Scope: "", Subject: "Use argon2", Reason: "New", SourceType: record.SourceHuman, Author: "test", Status: "active"},
			},
			wantConflicts:   0,
			wantNonConflict: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewStore(t.TempDir())
			for _, r := range tt.existing {
				if _, err := store.Write(r); err != nil {
					t.Fatal(err)
				}
			}

			conflicts, nonConflicting, err := DetectConflicts(store, tt.candidates)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(conflicts) != tt.wantConflicts {
				t.Errorf("conflicts: got %d, want %d", len(conflicts), tt.wantConflicts)
			}
			if len(nonConflicting) != tt.wantNonConflict {
				t.Errorf("non-conflicting: got %d, want %d", len(nonConflicting), tt.wantNonConflict)
			}
			if tt.wantExistingID != "" && len(conflicts) > 0 && conflicts[0].Existing.ID != tt.wantExistingID {
				t.Errorf("existing ID: got %s, want %s", conflicts[0].Existing.ID, tt.wantExistingID)
			}
		})
	}
}


