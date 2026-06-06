// Package record is the test surface for LoadFile.
package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// validDecisionJSON is the minimal valid decision record used as
// the baseline for LoadFile tests. It exercises one of the
// kind-specific required fields (decided_at) and the absolute
// minimum of the universal ones.
const validDecisionJSON = `{
  "id": "01HW4F2N8X2JZPM7Q3S5V0K1A1",
  "kind": "decision",
  "scope": "src/foo",
  "subject": "Use ULID for record ids",
  "reason": "Lexicographic time-sortable ids make logs and listings easier.",
  "source_type": "human",
  "author": "tester",
  "created_at": "2026-01-01T00:00:00Z",
  "last_verified_at": "2026-01-01T00:00:00Z",
  "status": "active",
  "decided_at": "2026-01-01T00:00:00Z"
}`

// TestLoadFileOK confirms that a file containing a schema-valid
// record is parsed into a *Record with its fields preserved.
func TestLoadFileOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ok.json")
	if err := writeFile(path, validDecisionJSON); err != nil {
		t.Fatal(err)
	}
	r, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID != "01HW4F2N8X2JZPM7Q3S5V0K1A1" {
		t.Errorf("id mismatch: %q", r.ID)
	}
	if r.Kind != KindDecision {
		t.Errorf("kind mismatch: %q", r.Kind)
	}
	if r.Subject != "Use ULID for record ids" {
		t.Errorf("subject mismatch: %q", r.Subject)
	}
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !r.CreatedAt.Equal(want) {
		t.Errorf("created_at mismatch: %v", r.CreatedAt)
	}
}

// TestLoadFileSchemaError confirms that a file whose content does
// not validate is rejected with a descriptive error wrapping the
// schema failure.
func TestLoadFileSchemaError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	// Missing the kind-specific "decided_at".
	if err := writeFile(path, strings.TrimSuffix(validDecisionJSON, ",\n  \"decided_at\": \"2026-01-01T00:00:00Z\"\n}")+"\n}"); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected schema error, got nil")
	}
	if !strings.Contains(err.Error(), "validate") {
		t.Errorf("error should mention validation, got: %v", err)
	}
}

// TestLoadFileMissing confirms that a path that does not exist
// surfaces the underlying read error.
func TestLoadFileMissing(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "nope.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// writeFile is a one-liner test helper.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
