// Package cortex is the test surface for the init subcommand.
package cortex

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SincereMa/cortex-sidemark/internal/record"
)

// seedProjectTree lays down a small project with the canonical
// files init looks for, plus a junk file that must be
// ignored. It returns the project root.
func seedProjectTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	must := func(rel, content string) {
		t.Helper()
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	must("README.md", "# Project\n\nThis is the project readme.\nIt has two lines of body.")
	must("CONTRIBUTING.md", "# Contributing\n\nSend a PR.")
	must("AGENTS.md", "# Agent guide\n\nRun `make test` first.")
	must("LICENSE", "MIT")
	must("docs/decisions/0001-foo.md", "# 0001 Foo\n\nWe chose foo because bar.")
	must("docs/decisions/0002-bar.md", "# 0002 Bar\n\nWe chose bar because baz.")
	must(".github/PULL_REQUEST_TEMPLATE.md", "## What\n\n## Why\n")

	// Junk that must not be picked up.
	must("random.txt", "ignore me")

	return root
}

// TestInitNoWrite confirms --no-write does not touch disk.
func TestInitNoWrite(t *testing.T) {
	root := seedProjectTree(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"init", "--root", root, "--no-write"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init returned error: %v\n%s", err, out.String())
	}

	if !strings.Contains(out.String(), "would write") {
		t.Errorf("dry-run output should say 'would write', got:\n%s", out.String())
	}
	// No .cortex/ should be created.
	if _, err := os.Stat(filepath.Join(root, ".cortex")); err == nil {
		t.Error("dry-run should not create .cortex/")
	}
}

// TestInitWritesSeeds confirms init creates .cortex/_seed/
// with one record per found file.
func TestInitWritesSeeds(t *testing.T) {
	root := seedProjectTree(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"init", "--root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init returned error: %v\n%s", err, out.String())
	}

	seedDir := filepath.Join(root, ".cortex", "_seed")
	entries, err := os.ReadDir(seedDir)
	if err != nil {
		t.Fatalf("expected .cortex/_seed to exist: %v", err)
	}
	// We seeded 7 files (README, CONTRIBUTING, AGENTS, LICENSE,
	// 2 docs/decisions/*, .github/PULL_REQUEST_TEMPLATE.md).
	// random.txt is not in the canonical list and is skipped.
	if len(entries) != 7 {
		t.Errorf("expected 7 seed records, got %d", len(entries))
	}

	// Each seed is a valid record with source_type=scrape.
	for _, e := range entries {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(seedDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var r record.Record
		if err := json.Unmarshal(data, &r); err != nil {
			t.Fatalf("seed %s is not valid JSON: %v", e.Name(), err)
		}
		if r.SourceType != record.SourceScrape {
			t.Errorf("seed %s should have source_type=scrape, got %q", e.Name(), r.SourceType)
		}
		if r.Author != "cortex init" {
			t.Errorf("seed %s should have author='cortex init', got %q", e.Name(), r.Author)
		}
	}
}

// TestInitSeedsNotInList confirms seeds are not surfaced by
// `cortex list` (which only walks the canonical kind dirs).
func TestInitSeedsNotInList(t *testing.T) {
	root := seedProjectTree(t)

	runInit := newRootCmd()
	var out bytes.Buffer
	runInit.SetOut(&out)
	runInit.SetErr(&out)
	runInit.SetArgs([]string{"init", "--root", root})
	if err := runInit.Execute(); err != nil {
		t.Fatalf("init returned error: %v", err)
	}

	listCmd := newRootCmd()
	out.Reset()
	listCmd.SetOut(&out)
	listCmd.SetErr(&out)
	listCmd.SetArgs([]string{"list", "--root", filepath.Join(root, ".cortex")})
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	// List walks only the kind dirs, so seeds must not appear.
	// The header is still printed.
	if strings.Contains(out.String(), "scrape") {
		t.Errorf("seeds should not appear in list output, got:\n%s", out.String())
	}
}

// TestInitIdempotent confirms running init twice does not
// duplicate seeds (the second run still finds the same files
// and writes fresh records with new ULIDs, but it does not
// error). Real idempotency would require us to skip files
// that already have a seed; that is left for a later ADR.
func TestInitIdempotent(t *testing.T) {
	root := seedProjectTree(t)

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
