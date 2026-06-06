// Package cortex is the test surface for the verify subcommand.
package cortex

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyStampsRecord confirms verify updates the on-disk
// record's last_verified_at to a recent timestamp.
func TestVerifyStampsRecord(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, ".cortex")
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	id := seedRecord(t, cortexDir)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"verify", "--root", cortexDir, id})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("verify returned error: %v\n%s", err, out.String())
	}
	stamp := strings.TrimSpace(out.String())
	if stamp == "" {
		t.Fatal("expected a timestamp on stdout, got empty output")
	}
	if !strings.Contains(stamp, "T") {
		t.Errorf("output should be an RFC3339 timestamp, got %q", stamp)
	}

	// Re-fetch and confirm the file on disk has the new stamp.
	gotCmd := newRootCmd()
	var got bytes.Buffer
	gotCmd.SetOut(&got)
	gotCmd.SetErr(&got)
	gotCmd.SetArgs([]string{"get", "--root", cortexDir, id})
	if err := gotCmd.Execute(); err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	if !strings.Contains(got.String(), stamp) {
		t.Errorf("record on disk does not contain the new stamp %q", stamp)
	}
}

// TestVerifyMissing confirms a non-existent id produces an
// error.
func TestVerifyMissing(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, ".cortex")
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"verify", "--root", cortexDir, "01ZZZZZZZZZZZZZZZZZZZZZZZZ"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}
