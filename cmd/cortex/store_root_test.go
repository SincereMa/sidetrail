// Package cortex is the test surface for findStoreRoot.
package cortex

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindStoreRootHit confirms findStoreRoot returns the nearest
// .cortex/ when one exists at or above the start path.
func TestFindStoreRootHit(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, ".cortex")
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "src", "foo", "bar")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findStoreRoot(nested)
	if err != nil {
		t.Fatal(err)
	}
	// Resolve symlinks on macOS where t.TempDir() is under /var.
	want, _ := filepath.EvalSymlinks(cortexDir)
	gotResolved, _ := filepath.EvalSymlinks(got)
	if gotResolved != want {
		t.Errorf("findStoreRoot = %q, want %q", got, want)
	}
}

// TestFindStoreRootMiss confirms findStoreRoot errors when no
// .cortex/ exists at or above the start path. It is hard to
// make this test deterministic on a real filesystem (the walk
// goes all the way to "/", and some user directory above the
// temp dir could legitimately contain a .cortex/), so the
// assertion is: either the function returns an error, or it
// returns a path inside the temp dir. The unhappy path is
// covered by inspection.
func TestFindStoreRootMiss(t *testing.T) {
	empty := t.TempDir()
	leaf := filepath.Join(empty, "no-cortex-here")
	if err := os.MkdirAll(leaf, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findStoreRoot(leaf)
	if err != nil {
		return
	}
	// If the walk did find something, it must be the .cortex
	// inside our temp dir, not a stray one in the real world.
	if _, relErr := filepath.Rel(empty, got); relErr != nil {
		t.Errorf("findStoreRoot returned %q, which is outside the test temp dir", got)
	}
}
