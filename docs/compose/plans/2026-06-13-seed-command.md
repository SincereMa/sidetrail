# sidetrail seed Command Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `seed` command that generates agent prompts from project documents and applies generated records with conflict detection.

**Architecture:** Two-phase agent-driven design: `seed --files` outputs a structured prompt for the host agent to process; `seed --apply` applies agent-generated records with scope+subject conflict detection. No LLM calls bundled in the binary.

**Tech Stack:** Go, Cobra, JSON, filepath.Glob

---

## File Structure

| File | Responsibility |
|------|----------------|
| `cmd/sidetrail/seed.go` | CLI command wiring, flags, output formatting |
| `internal/seed/prompt.go` | Prompt generation from document contents |
| `internal/seed/conflict.go` | Conflict detection (scope+subject matching) |
| `internal/seed/conflict_test.go` | Conflict detection tests |
| `internal/seed/prompt_test.go` | Prompt generation tests |
| `cmd/sidetrail/seed_test.go` | Integration tests for seed command |

---

### Task 1: Conflict Detection Logic

**Covers:** [S3, S4]

**Files:**
- Create: `internal/seed/conflict.go`
- Create: `internal/seed/conflict_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/seed/conflict_test.go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./internal/seed/... -v`
Expected: FAIL with "undefined: DetectConflicts"

- [ ] **Step 3: Write minimal implementation**

```go
// internal/seed/conflict.go
package seed

import (
	"strings"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// Conflict represents a candidate record that conflicts with an existing record.
type Conflict struct {
	Candidate *record.Record `json:"candidate"`
	Existing  *record.Record `json:"existing"`
}

// DetectConflicts compares candidate records against the store and returns
// conflicts (same kind + overlapping scope + similar subject) and
// non-conflicting records.
func DetectConflicts(store *storage.Store, candidates []*record.Record) ([]Conflict, []*record.Record) {
	existing, err := store.ListAll()
	if err != nil {
		return nil, candidates
	}

	var conflicts []Conflict
	var nonConflicting []*record.Record

	for _, c := range candidates {
		conflict := findConflict(c, existing)
		if conflict != nil {
			conflicts = append(conflicts, *conflict)
		} else {
			nonConflicting = append(nonConflicting, c)
		}
	}

	return conflicts, nonConflicting
}

// findConflict checks if candidate conflicts with any existing record.
func findConflict(candidate *record.Record, existing []*record.Record) *Conflict {
	for _, e := range existing {
		if candidate.Kind != e.Kind {
			continue
		}
		if !scopeOverlaps(candidate.Scope, e.Scope) {
			continue
		}
		if subjectsMatch(candidate.Subject, e.Subject) {
			return &Conflict{Candidate: candidate, Existing: e}
		}
	}
	return nil
}

// scopeOverlaps reports whether two scopes overlap (exact match or one is ancestor of other).
func scopeOverlaps(a, b string) bool {
	a = strings.TrimRight(a, "/")
	b = strings.TrimRight(b, "/")
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}

// subjectsMatch reports whether two subjects are similar (case-insensitive prefix match).
func subjectsMatch(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./internal/seed/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/seed/conflict.go internal/seed/conflict_test.go
git commit -m "feat: add conflict detection for seed command"
```

---

### Task 2: Prompt Generation

**Covers:** [S3]

**Files:**
- Create: `internal/seed/prompt.go`
- Create: `internal/seed/prompt_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/seed/prompt_test.go
package seed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratePrompt(t *testing.T) {
	dir := t.TempDir()
	doc1 := filepath.Join(dir, "README.md")
	doc2 := filepath.Join(dir, "ARCHITECTURE.md")
	os.WriteFile(doc1, []byte("# Project\nUse bcrypt for auth."), 0644)
	os.WriteFile(doc2, []byte("# Architecture\nDon't modify billing."), 0644)

	prompt, err := GeneratePrompt([]string{doc1, doc2})
	if err != nil {
		t.Fatal(err)
	}
	if prompt == "" {
		t.Error("expected non-empty prompt")
	}
	if !containsString(prompt, "README.md") {
		t.Error("expected prompt to contain README.md")
	}
	if !containsString(prompt, "ARCHITECTURE.md") {
		t.Error("expected prompt to contain ARCHITECTURE.md")
	}
	if !containsString(prompt, "Use bcrypt for auth.") {
		t.Error("expected prompt to contain document content")
	}
}

func TestGeneratePrompt_EmptyFiles(t *testing.T) {
	_, err := GeneratePrompt([]string{})
	if err == nil {
		t.Error("expected error for empty files")
	}
}

func TestGeneratePrompt_NonexistentFile(t *testing.T) {
	_, err := GeneratePrompt([]string{"/nonexistent/file.md"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./internal/seed/... -v -run TestGeneratePrompt`
Expected: FAIL with "undefined: GeneratePrompt"

- [ ] **Step 3: Write minimal implementation**

```go
// internal/seed/prompt.go
package seed

import (
	"fmt"
	"os"
	"strings"
)

// GeneratePrompt reads the specified files and generates a structured prompt
// for the host agent to extract decisions, constraints, and signals.
func GeneratePrompt(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files specified")
	}

	var builder strings.Builder
	builder.WriteString("# SideTrail Seed Prompt\n\n")
	builder.WriteString("Extract decisions, constraints, and signals from the following project documents.\n")
	builder.WriteString("For each item found, generate a JSON record matching the SideTrail schema.\n\n")
	builder.WriteString("## Schema Reference\n\n")
	builder.WriteString("Required fields: id, kind, scope, subject, reason, source_type, author, created_at, last_verified_at, status\n")
	builder.WriteString("Kinds: decision, constraint, signal, experiment, incident\n")
	builder.WriteString("Source types: human, agent-suggested, scrape, derived\n\n")
	builder.WriteString("## Documents\n\n")

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read %q: %w", file, err)
		}
		builder.WriteString(fmt.Sprintf("### %s\n\n", file))
		content := string(data)
		if len(content) > 5000 {
			content = content[:5000] + "\n... (truncated)"
		}
		builder.WriteString(content)
		builder.WriteString("\n\n")
	}

	builder.WriteString("## Output Format\n\n")
	builder.WriteString("Return a JSON array of records. Each record should have:\n")
	builder.WriteString("- kind: decision, constraint, or signal (based on content)\n")
	builder.WriteString("- scope: file or directory path this applies to\n")
	builder.WriteString("- subject: brief title of the item\n")
	builder.WriteString("- reason: why this decision/constraint exists\n")
	builder.WriteString("- status: active\n")
	builder.WriteString("- source_type: derived (since extracted from docs)\n")
	builder.WriteString("- author: agent\n")

	return builder.String(), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./internal/seed/... -v -run TestGeneratePrompt`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/seed/prompt.go internal/seed/prompt_test.go
git commit -m "feat: add prompt generation for seed command"
```

---

### Task 3: Seed Command CLI

**Covers:** [S3, S5]

**Files:**
- Create: `cmd/sidetrail/seed.go`
- Create: `cmd/sidetrail/seed_test.go`
- Modify: `cmd/sidetrail/root.go:39-45`

- [ ] **Step 1: Write the failing test**

```go
// cmd/sidetrail/seed_test.go
package sidetrail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/SincereMa/sidetrail/internal/seed"
)

func TestSeed_Files(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "README.md")
	os.WriteFile(doc, []byte("# Project\nUse bcrypt."), 0644)

	root := filepath.Join(dir, ".sidetrail")
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--files", doc})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --files: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestSeed_Apply(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".sidetrail")
	os.MkdirAll(root, 0755)

	records := []map[string]interface{}{
		{
			"kind":            "decision",
			"scope":           "src/auth",
			"subject":         "Use bcrypt",
			"reason":          "Security",
			"source_type":     "derived",
			"author":          "agent",
			"created_at":      "2026-01-01T00:00:00Z",
			"last_verified_at": "2026-01-01T00:00:00Z",
			"status":          "active",
		},
	}
	data, _ := json.Marshal(records)
	recordsFile := filepath.Join(dir, "records.json")
	os.WriteFile(recordsFile, data, 0644)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--apply", recordsFile, "--root", root})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --apply: %v", err)
	}

	// Verify record was written
	store := seed.NewStore(root)
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 record, got %d", len(all))
	}
}

func TestSeed_ApplyDryRun(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".sidetrail")
	os.MkdirAll(root, 0755)

	records := []map[string]interface{}{
		{
			"kind":            "decision",
			"scope":           "src/auth",
			"subject":         "Use bcrypt",
			"reason":          "Security",
			"source_type":     "derived",
			"author":          "agent",
			"created_at":      "2026-01-01T00:00:00Z",
			"last_verified_at": "2026-01-01T00:00:00Z",
			"status":          "active",
		},
	}
	data, _ := json.Marshal(records)
	recordsFile := filepath.Join(dir, "records.json")
	os.WriteFile(recordsFile, data, 0644)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--apply", recordsFile, "--dry-run", "--root", root})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --apply --dry-run: %v", err)
	}

	// Verify no records were written
	store := seed.NewStore(root)
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 records (dry run), got %d", len(all))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./cmd/sidetrail/... -v -run TestSeed`
Expected: FAIL with "undefined: newSeedCmd"

- [ ] **Step 3: Write minimal implementation**

```go
// cmd/sidetrail/seed.go
package sidetrail

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/seed"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// seedOptions carries the flags for the `seed` command.
type seedOptions struct {
	root    string
	files   string
	apply   string
	dryRun  bool
	jsonO   bool
}

// newSeedCmd builds the `sidetrail seed` subcommand. It has two modes:
// 1. --files: generate a prompt for the host agent to extract records
// 2. --apply: apply agent-generated records with conflict detection
func newSeedCmd() *cobra.Command {
	opts := &seedOptions{}
	cmd := &cobra.Command{
		Use:   "seed [--files <glob>] [--apply <file>] [--dry-run] [--json]",
		Short: "Seed records from project documents or apply agent-generated records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeed(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.files, "files", "", "glob pattern for project documents to read")
	cmd.Flags().StringVar(&opts.apply, "apply", "file containing JSON array of records to apply")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "show what would happen without writing")
	cmd.Flags().BoolVar(&opts.jsonO, "json", false, "emit JSON output instead of text")
	return cmd
}

// runSeed dispatches to the appropriate mode based on flags.
func runSeed(cmd *cobra.Command, _ []string, opts *seedOptions) error {
	if opts.files != "" && opts.apply != "" {
		return fmt.Errorf("--files and --apply are mutually exclusive")
	}
	if opts.files == "" && opts.apply == "" {
		return fmt.Errorf("either --files or --apply is required")
	}

	if opts.files != "" {
		return runSeedFiles(cmd, opts)
	}
	return runSeedApply(cmd, opts)
}

// runSeedFiles reads documents and outputs a prompt for the host agent.
func runSeedFiles(cmd *cobra.Command, opts *seedOptions) error {
	files, err := filepath.Glob(opts.files)
	if err != nil {
		return fmt.Errorf("glob %q: %w", opts.files, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no files matched pattern %q", opts.files)
	}

	prompt, err := seed.GeneratePrompt(files)
	if err != nil {
		return err
	}

	if opts.jsonO {
		output := map[string]string{"prompt": prompt}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	fmt.Fprint(cmd.OutOrStdout(), prompt)
	return nil
}

// runSeedApply reads candidate records and applies them to the store.
func runSeedApply(cmd *cobra.Command, opts *seedOptions) error {
	data, err := os.ReadFile(opts.apply)
	if err != nil {
		return fmt.Errorf("read %q: %w", opts.apply, err)
	}

	var candidates []*record.Record
	if err := json.Unmarshal(data, &candidates); err != nil {
		return fmt.Errorf("parse %q: %w", opts.apply, err)
	}

	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	store := storage.NewStore(root)

	conflicts, nonConflicting := seed.DetectConflicts(store, candidates)

	if opts.jsonO {
		return seedApplyJSON(cmd, conflicts, nonConflicting, opts.dryRun)
	}
	return seedApplyTable(cmd, conflicts, nonConflicting, opts.dryRun)
}

// seedApplyJSON outputs the apply result as JSON.
func seedApplyJSON(cmd *cobra.Command, conflicts []seed.Conflict, nonConflicting []*record.Record, dryRun bool) error {
	type applyResult struct {
		DryRun         bool             `json:"dry_run"`
		Conflicts      []seed.Conflict  `json:"conflicts"`
		NonConflicting []*record.Record `json:"non_conflicting"`
	}
	result := applyResult{
		DryRun:         dryRun,
		Conflicts:      conflicts,
		NonConflicting: nonConflicting,
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

// seedApplyTable outputs the apply result as a human-readable table.
func seedApplyTable(cmd *cobra.Command, conflicts []seed.Conflict, nonConflicting []*record.Record, dryRun bool) error {
	out := cmd.OutOrStdout()
	if dryRun {
		fmt.Fprintln(out, "DRY RUN - no changes will be written")
		fmt.Fprintln(out)
	}

	if len(conflicts) > 0 {
		fmt.Fprintf(out, "Conflicts found: %d\n", len(conflicts))
		for _, c := range conflicts {
			fmt.Fprintf(out, "  Existing: %s %s %s (scope: %s)\n", c.Existing.ID, c.Existing.Kind, c.Existing.Subject, c.Existing.Scope)
			fmt.Fprintf(out, "  Candidate: %s %s (scope: %s)\n", c.Candidate.Kind, c.Candidate.Subject, c.Candidate.Scope)
			fmt.Fprintln(out)
		}
	}

	if len(nonConflicting) > 0 {
		fmt.Fprintf(out, "Records to add: %d\n", len(nonConflicting))
		for _, r := range nonConflicting {
			fmt.Fprintf(out, "  %s %s (scope: %s)\n", r.Kind, r.Subject, r.Scope)
		}
	}

	if len(conflicts) == 0 && len(nonConflicting) == 0 {
		fmt.Fprintln(out, "No records to process")
	}

	return nil
}
```

- [ ] **Step 4: Register seed command in root.go**

Modify `cmd/sidetrail/root.go:39-45`:

```go
	cmd.AddCommand(
		newAddCmd(),
		newContextCmd(),
		newUpdateCmd(),
		newHealthCmd(),
		newInitCmd(),
		newSeedCmd(),
	)
```

- [ ] **Step 5: Add missing import**

Modify `cmd/sidetrail/seed.go:1-12` to add `path/filepath` import:

```go
package sidetrail

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/seed"
	"github.com/SincereMa/sidetrail/internal/storage"
)
```

- [ ] **Step 6: Run test to verify it passes**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./cmd/sidetrail/... -v -run TestSeed`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add cmd/sidetrail/seed.go cmd/sidetrail/seed_test.go cmd/sidetrail/root.go
git commit -m "feat: add seed command for bootstrapping records"
```

---

### Task 4: Integration Tests

**Covers:** [S3, S4, S5]

**Files:**
- Modify: `cmd/sidetrail/seed_test.go`

- [ ] **Step 1: Write the failing test**

Add to `cmd/sidetrail/seed_test.go`:

```go
func TestSeed_ApplyWithConflict(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".sidetrail")
	os.MkdirAll(root, 0755)

	// Create existing record
	existing := map[string]interface{}{
		"id":              "existing-1",
		"kind":            "decision",
		"scope":           "src/auth",
		"subject":         "Use bcrypt",
		"reason":          "Old decision",
		"source_type":     "human",
		"author":          "test",
		"created_at":      "2026-01-01T00:00:00Z",
		"last_verified_at": "2026-01-01T00:00:00Z",
		"status":          "active",
	}
	existingData, _ := json.Marshal(existing)
	existingFile := filepath.Join(root, "decisions", "existing-1-use-bcrypt.json")
	os.MkdirAll(filepath.Dir(existingFile), 0755)
	os.WriteFile(existingFile, existingData, 0644)

	// Candidate that conflicts
	candidates := []map[string]interface{}{
		{
			"kind":            "decision",
			"scope":           "src/auth",
			"subject":         "Use bcrypt for hashing",
			"reason":          "New decision",
			"source_type":     "derived",
			"author":          "agent",
			"created_at":      "2026-01-02T00:00:00Z",
			"last_verified_at": "2026-01-02T00:00:00Z",
			"status":          "active",
		},
	}
	data, _ := json.Marshal(candidates)
	recordsFile := filepath.Join(dir, "records.json")
	os.WriteFile(recordsFile, data, 0644)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"seed", "--apply", recordsFile, "--root", root})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed --apply: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Conflicts found: 1") {
		t.Errorf("expected conflict message, got: %s", output)
	}
	if !strings.Contains(output, "existing-1") {
		t.Errorf("expected existing record ID in output, got: %s", output)
	}

	// Verify no new record was written (conflict prevents auto-add)
	store := storage.NewStore(root)
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 record (conflict prevented add), got %d", len(all))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./cmd/sidetrail/... -v -run TestSeed_ApplyWithConflict`
Expected: FAIL (missing storage import or conflict behavior)

- [ ] **Step 3: Fix implementation if needed**

The test should pass with current implementation. If not, fix the implementation.

- [ ] **Step 4: Run all tests**

Run: `cd /Users/sincere/Projects/sidetrail && go test ./...`
Expected: ALL PASS

- [ ] **Step 5: Run linter**

Run: `cd /Users/sincere/Projects/sidetrail && go vet ./...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/sidetrail/seed_test.go
git commit -m "test: add integration tests for seed command"
```

---

### Task 5: Documentation Update

**Covers:** [S5]

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README.md CLI table**

Add to README.md after the `init` row in the CLI surface table:

```markdown
| `sidetrail seed --files <glob>` | Generate agent prompt from project documents |
| `sidetrail seed --apply <file>` | Apply agent-generated records with conflict detection |
```

- [ ] **Step 2: Add seed section to Quick start**

Add after the init section in README.md:

```markdown
**2. Seed records from existing docs (optional):**

```bash
# Generate prompt for agent to extract records
sidetrail seed --files "./docs/**/*.md"
# Agent processes documents and generates records.json

# Apply generated records (dry run first)
sidetrail seed --apply records.json --dry-run
sidetrail seed --apply records.json
```
```

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add seed command documentation"
```
