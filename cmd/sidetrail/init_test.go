// Package sidetrail is the test surface for the init subcommand.
package sidetrail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitCreatesDirectory confirms init creates the .sidetrail/ directory.
func TestInitCreatesDirectory(t *testing.T) {
	root := t.TempDir()

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"init", "--root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init returned error: %v\n%s", err, out.String())
	}

	storeDir := filepath.Join(root, storeDirName)
	info, err := os.Stat(storeDir)
	if err != nil {
		t.Fatalf("store dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("store path is not a directory")
	}

	if !strings.Contains(out.String(), "created") {
		t.Errorf("expected 'created' in output, got:\n%s", out.String())
	}
}

// TestInitIdempotent confirms running init twice does not error.
func TestInitIdempotent(t *testing.T) {
	root := t.TempDir()

	for i := 0; i < 2; i++ {
		cmd := newRootCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"init", "--root", root})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("init run %d: %v\n%s", i, err, out.String())
		}
	}
}

// TestInitDefaultRoot confirms init uses CWD when --root is omitted.
func TestInitDefaultRoot(t *testing.T) {
	root := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(root)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init returned error: %v\n%s", err, out.String())
	}

	storeDir := filepath.Join(root, storeDirName)
	if _, err := os.Stat(storeDir); err != nil {
		t.Fatalf("store dir not created in CWD: %v", err)
	}
}
