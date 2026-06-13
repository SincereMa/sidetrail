# Simplify CLI to Agent-First 4-Command Surface

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce the CLI from 13 subcommands to 4 agent-driven commands: `context`, `add`, `update`, `health`.

**Architecture:** Remove 9 command files (list, ask, get, validate, draft, promote, supersede, verify, status), add 1 new command (update), simplify root.go, remove seed/draft/promote storage logic, update tests and docs.

**Tech Stack:** Go, Cobra, JSON Schema, ULID

---

## File Structure

### Files to Remove
- `cmd/sidetrail/list.go` — subsumed by context
- `cmd/sidetrail/list_test.go`
- `cmd/sidetrail/ask.go` — subsumed by context
- `cmd/sidetrail/ask_test.go`
- `cmd/sidetrail/get.go` — subsumed by add/context
- `cmd/sidetrail/get_test.go`
- `cmd/sidetrail/validate.go` — internal to add/update
- `cmd/sidetrail/validate_test.go`
- `cmd/sidetrail/draft.go` — agent writes directly
- `cmd/sidetrail/draft_test.go`
- `cmd/sidetrail/promote.go` — no seed workflow
- `cmd/sidetrail/promote_test.go`
- `cmd/sidetrail/supersede.go` — replaced by update
- `cmd/sidetrail/supersede_test.go`
- `cmd/sidetrail/verify.go` — replaced by update
- `cmd/sidetrail/verify_test.go`
- `cmd/sidetrail/status.go` — replaced by update
- `cmd/sidetrail/status_test.go`

### Files to Create
- `cmd/sidetrail/update.go` — new command for updating existing records
- `cmd/sidetrail/update_test.go` — tests for update command

### Files to Modify
- `cmd/sidetrail/root.go` — remove 9 command registrations
- `cmd/sidetrail/init.go` — simplify or remove (store auto-created on first add)
- `cmd/sidetrail/init_test.go` — update tests
- `cmd/sidetrail/store_root.go` — keep as-is (used by context, add, update, health)
- `cmd/sidetrail/store_root_test.go` — keep as-is
- `internal/storage/store.go` — remove WriteSeed, WriteDraft, pluralize (move to context), keep Write
- `internal/storage/store_test.go` — update tests
- `README.md` — update CLI surface documentation
- `ROADMAP.md` — update status

### Files to Keep Unchanged
- `cmd/sidetrail/context.go` — unchanged
- `cmd/sidetrail/context_test.go` — unchanged
- `cmd/sidetrail/health.go` — unchanged
- `cmd/sidetrail/health_test.go` — unchanged
- `cmd/sidetrail/add.go` — unchanged (already validates + stores)
- `cmd/sidetrail/add_test.go` — unchanged
- `internal/record/record.go` — unchanged
- `internal/schema/` — unchanged
- `internal/version/` — unchanged
- `main.go` — unchanged

---

## Task 1: Remove list command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/list.go`
- Delete: `cmd/sidetrail/list_test.go`
- Modify: `cmd/sidetrail/root.go:35-50` (remove newListCmd registration)

- [ ] **Step 1: Delete list.go and list_test.go**

```bash
rm cmd/sidetrail/list.go cmd/sidetrail/list_test.go
```

- [ ] **Step 2: Remove newListCmd from root.go**

In `cmd/sidetrail/root.go`, remove `newListCmd()` from the `cmd.AddCommand(...)` call:

```go
cmd.AddCommand(
    newValidateCmd(),
    newAddCmd(),
    newGetCmd(),
    newAskCmd(),
    newContextCmd(),
    newVerifyCmd(),
    newSupersedeCmd(),
    newInitCmd(),
    newPromoteCmd(),
    newDraftCmd(),
    newStatusCmd(),
    newHealthCmd(),
)
```

- [ ] **Step 3: Run tests to verify removal**

```bash
go test ./...
```

Expected: PASS (list tests removed, no compilation errors)

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove list command (subsumed by context)"
```

---

## Task 2: Remove ask command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/ask.go`
- Delete: `cmd/sidetrail/ask_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newAskCmd registration)

- [ ] **Step 1: Delete ask.go and ask_test.go**

```bash
rm cmd/sidetrail/ask.go cmd/sidetrail/ask_test.go
```

- [ ] **Step 2: Remove newAskCmd from root.go**

Remove `newAskCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove ask command (subsumed by context)"
```

---

## Task 3: Remove get command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/get.go`
- Delete: `cmd/sidetrail/get_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newGetCmd registration)

- [ ] **Step 1: Delete get.go and get_test.go**

```bash
rm cmd/sidetrail/get.go cmd/sidetrail/get_test.go
```

- [ ] **Step 2: Remove newGetCmd from root.go**

Remove `newGetCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove get command (agent reads JSON directly)"
```

---

## Task 4: Remove validate command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/validate.go`
- Delete: `cmd/sidetrail/validate_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newValidateCmd registration)

- [ ] **Step 1: Delete validate.go and validate_test.go**

```bash
rm cmd/sidetrail/validate.go cmd/sidetrail/validate_test.go
```

- [ ] **Step 2: Remove newValidateCmd from root.go**

Remove `newValidateCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove validate command (internal to add/update)"
```

---

## Task 5: Remove draft command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/draft.go`
- Delete: `cmd/sidetrail/draft_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newDraftCmd registration)

- [ ] **Step 1: Delete draft.go and draft_test.go**

```bash
rm cmd/sidetrail/draft.go cmd/sidetrail/draft_test.go
```

- [ ] **Step 2: Remove newDraftCmd from root.go**

Remove `newDraftCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove draft command (agent writes directly)"
```

---

## Task 6: Remove promote command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/promote.go`
- Delete: `cmd/sidetrail/promote_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newPromoteCmd registration)

- [ ] **Step 1: Delete promote.go and promote_test.go**

```bash
rm cmd/sidetrail/promote.go cmd/sidetrail/promote_test.go
```

- [ ] **Step 2: Remove newPromoteCmd from root.go**

Remove `newPromoteCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove promote command (no seed workflow)"
```

---

## Task 7: Remove supersede command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/supersede.go`
- Delete: `cmd/sidetrail/supersede_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newSupersedeCmd registration)

- [ ] **Step 1: Delete supersede.go and supersede_test.go**

```bash
rm cmd/sidetrail/supersede.go cmd/sidetrail/supersede_test.go
```

- [ ] **Step 2: Remove newSupersedeCmd from root.go**

Remove `newSupersedeCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove supersede command (replaced by update)"
```

---

## Task 8: Remove verify command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/verify.go`
- Delete: `cmd/sidetrail/verify_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newVerifyCmd registration)

- [ ] **Step 1: Delete verify.go and verify_test.go**

```bash
rm cmd/sidetrail/verify.go cmd/sidetrail/verify_test.go
```

- [ ] **Step 2: Remove newVerifyCmd from root.go**

Remove `newVerifyCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove verify command (replaced by update)"
```

---

## Task 9: Remove status command

**Covers:** Simplification — remove subsumed commands

**Files:**
- Delete: `cmd/sidetrail/status.go`
- Delete: `cmd/sidetrail/status_test.go`
- Modify: `cmd/sidetrail/root.go` (remove newStatusCmd registration)

- [ ] **Step 1: Delete status.go and status_test.go**

```bash
rm cmd/sidetrail/status.go cmd/sidetrail/status_test.go
```

- [ ] **Step 2: Remove newStatusCmd from root.go**

Remove `newStatusCmd()` from the `cmd.AddCommand(...)` call.

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove status command (replaced by update)"
```

---

## Task 10: Create update command

**Covers:** New command for updating existing records

**Files:**
- Create: `cmd/sidetrail/update.go`
- Create: `cmd/sidetrail/update_test.go`
- Modify: `cmd/sidetrail/root.go` (add newUpdateCmd registration)

- [ ] **Step 1: Write failing test for update command**

Create `cmd/sidetrail/update_test.go`:

```go
package sidetrail

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

func TestUpdateRecord(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewStore(dir)
	r := &record.Record{
		ID:             "test-id-01",
		Kind:           record.KindDecision,
		Scope:          "src/main.go",
		Subject:        "Use feature flag",
		Reason:         "Because rollout needs to be gradual",
		SourceType:     record.SourceHuman,
		Author:         "test",
		CreatedAt:      time.Now().UTC(),
		LastVerifiedAt: time.Now().UTC(),
		Status:         "active",
	}
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}

	updateJSON := `{"status":"archived"}`
	updateFile := filepath.Join(dir, "update.json")
	if err := os.WriteFile(updateFile, []byte(updateJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"update", "test-id-01", "--file", updateFile, "--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := s.Get("test-id-01")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "archived" {
		t.Errorf("expected status=archived, got %q", updated.Status)
	}
	if updated.Subject != "Use feature flag" {
		t.Errorf("subject should be unchanged, got %q", updated.Subject)
	}
}

func TestUpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	updateJSON := `{"status":"archived"}`
	updateFile := filepath.Join(dir, "update.json")
	if err := os.WriteFile(updateFile, []byte(updateJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"update", "nonexistent-id", "--file", updateFile, "--root", dir})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent record")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./cmd/sidetrail -run TestUpdateRecord -v
```

Expected: FAIL — `newUpdateCmd` not defined

- [ ] **Step 3: Implement update command**

Create `cmd/sidetrail/update.go`:

```go
package sidetrail

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/storage"
)

// updateOptions carries the flags for the `update` command.
type updateOptions struct {
	root string
	file string
}

// newUpdateCmd builds the `sidetrail update` subcommand. It
// updates an existing record with partial JSON fields. The
// agent reads the current record, merges the provided fields,
// and writes it back.
func newUpdateCmd() *cobra.Command {
	opts := &updateOptions{}
	cmd := &cobra.Command{
		Use:   "update <id> --file <json-file>",
		Short: "Update an existing record with partial JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.file, "file", "", "JSON file with fields to update (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

// runUpdate reads the existing record, merges the provided JSON
// fields, and writes it back.
func runUpdate(cmd *cobra.Command, args []string, opts *updateOptions) error {
	id := args[0]
	if opts.file == "" {
		return fmt.Errorf("--file is required")
	}

	data, err := os.ReadFile(opts.file)
	if err != nil {
		return fmt.Errorf("read update file: %w", err)
	}

	var updates map[string]interface{}
	if err := json.Unmarshal(data, &updates); err != nil {
		return fmt.Errorf("parse update JSON: %w", err)
	}

	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)

	r, err := s.Get(id)
	if err != nil {
		return err
	}

	existingJSON, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal existing record: %w", err)
	}

	var existing map[string]interface{}
	if err := json.Unmarshal(existingJSON, &existing); err != nil {
		return fmt.Errorf("unmarshal existing record: %w", err)
	}

	for k, v := range updates {
		existing[k] = v
	}

	mergedJSON, err := json.Marshal(existing)
	if err != nil {
		return fmt.Errorf("marshal merged record: %w", err)
	}

	if err := storage.ValidateAndWrite(s, r, mergedJSON); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", id)
	return nil
}
```

- [ ] **Step 4: Add ValidateAndWrite helper to storage**

In `internal/storage/store.go`, add:

```go
// ValidateAndWrite rewrites an existing record with merged JSON.
// The record's Kind must remain valid after merge.
func (s *Store) ValidateAndWrite(r *record.Record, mergedJSON []byte) error {
	var updated record.Record
	if err := json.Unmarshal(mergedJSON, &updated); err != nil {
		return fmt.Errorf("unmarshal merged record: %w", err)
	}
	if !updated.Kind.Valid() {
		return fmt.Errorf("invalid kind after merge: %q", updated.Kind)
	}
	updated.ID = r.ID
	_, err := s.Write(&updated)
	return err
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./cmd/sidetrail -run TestUpdateRecord -v
```

Expected: PASS

- [ ] **Step 6: Add update to root.go**

In `cmd/sidetrail/root.go`, add `newUpdateCmd()` to the `cmd.AddCommand(...)` call:

```go
cmd.AddCommand(
    newAddCmd(),
    newContextCmd(),
    newUpdateCmd(),
    newHealthCmd(),
)
```

- [ ] **Step 7: Run all tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add update command for partial record updates"
```

---

## Task 11: Remove seed/draft storage logic

**Covers:** Simplification — remove unused storage methods

**Files:**
- Modify: `internal/storage/store.go` (remove WriteSeed, WriteDraft, pluralize)

- [ ] **Step 1: Remove WriteSeed method**

In `internal/storage/store.go`, remove the `WriteSeed` method (lines 77-97).

- [ ] **Step 2: Remove WriteDraft method**

In `internal/storage/store.go`, remove the `WriteDraft` method (lines 300-315).

- [ ] **Step 3: Remove pluralize function from storage**

In `internal/storage/store.go`, remove the `pluralize` function (lines 43-57). Keep the `kindDir` method but inline the pluralization:

```go
func (s *Store) kindDir(k record.Kind) string {
	return filepath.Join(s.root, string(k)+"s")
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "refactor: remove seed/draft storage methods"
```

---

## Task 12: Simplify init command

**Covers:** Simplification — remove seed workflow from init

**Files:**
- Modify: `cmd/sidetrail/init.go` (simplify to just create directory)
- Modify: `cmd/sidetrail/init_test.go` (update tests)

- [ ] **Step 1: Simplify init.go**

Replace the entire `cmd/sidetrail/init.go` with:

```go
package sidetrail

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initOptions carries the flags for the `init` command.
type initOptions struct {
	root string
}

// newInitCmd builds the `sidetrail init` subcommand. It creates
// a .sidetrail/ directory at the project root. The store is
// usable from empty; init is optional.
func newInitCmd() *cobra.Command {
	opts := &initOptions{}
	cmd := &cobra.Command{
		Use:   "init [--root <project>]",
		Short: "Create a .sidetrail/ directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "project root where .sidetrail/ will be created (default: CWD)")
	return cmd
}

// runInit creates the .sidetrail/ directory.
func runInit(cmd *cobra.Command, _ []string, opts *initOptions) error {
	projectRoot, err := resolveProjectRoot(opts.root)
	if err != nil {
		return err
	}
	storeDir := filepath.Join(projectRoot, storeDirName)
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", storeDir, err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", storeDir)
	return nil
}

// resolveProjectRoot returns the absolute project root. When
// opts.root is empty, the current working directory is used.
func resolveProjectRoot(explicit string) (string, error) {
	if explicit == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
		return wd, nil
	}
	abs, err := filepath.Abs(explicit)
	if err != nil {
		return "", fmt.Errorf("abs %q: %w", explicit, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("--root %q is not a directory", abs)
	}
	return abs, nil
}
```

- [ ] **Step 2: Update init_test.go**

Replace `cmd/sidetrail/init_test.go` with:

```go
package sidetrail

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	cmd := newRootCmd()
	cmd.SetArgs([]string{"init", "--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	storeDir := filepath.Join(dir, storeDirName)
	info, err := os.Stat(storeDir)
	if err != nil {
		t.Fatalf("store dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("store path is not a directory")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: simplify init to just create directory"
```

---

## Task 13: Update root.go final state

**Covers:** Simplification — final root.go with 4 commands

**Files:**
- Modify: `cmd/sidetrail/root.go`

- [ ] **Step 1: Verify root.go has exactly 4 commands**

The final `cmd.AddCommand(...)` call should be:

```go
cmd.AddCommand(
    newAddCmd(),
    newContextCmd(),
    newUpdateCmd(),
    newHealthCmd(),
)
```

- [ ] **Step 2: Update Long description**

Update the `Long` field in root.go to reflect the simplified surface:

```go
Long: `sidetrail is the CLI for SideTrail, a sidecar that records
long-lived context (decisions, constraints, signals, dependencies) and
makes it available to host agents before they act.

Commands:
  context  — read records relevant to a file
  add      — validate and store a record
  update   — update an existing record
  health   — report project health signals`,
```

- [ ] **Step 3: Run tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor: finalize root.go with 4-command surface"
```

---

## Task 14: Update README.md

**Covers:** Documentation — update CLI surface docs

**Files:**
- Modify: `README.md` (CLI surface section)

- [ ] **Step 1: Update CLI surface table**

Replace the 13-command table in README.md with:

```markdown
## CLI surface

The `sidetrail` binary is agent-driven with four commands:

| Command | Purpose |
| --- | --- |
| `sidetrail context --file <path>` | Aggregate records relevant to a file path |
| `sidetrail add <file>` | Validate a record file and add it to the store |
| `sidetrail update <id> --file <json>` | Update an existing record with partial JSON |
| `sidetrail health` | Report project health signals |

The `.sidetrail/` store is discovered by walking upward from the
current working directory unless `--root` points elsewhere. The
store is auto-created on first `add`.
```

- [ ] **Step 2: Update How it works section**

Simplify the "How it works" section to focus on agent-driven usage.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "docs: update README for simplified CLI surface"
```

---

## Task 15: Update ROADMAP.md

**Covers:** Documentation — update status

**Files:**
- Modify: `ROADMAP.md`

- [ ] **Step 1: Update Current status section**

Update the "Completed" list to reflect the 4-command surface.

- [ ] **Step 2: Update Outstanding tasks**

Remove completed items, add any remaining work.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "docs: update ROADMAP for simplified CLI"
```

---

## Task 16: Final verification

**Covers:** Verification — ensure everything works

**Files:**
- None (verification only)

- [ ] **Step 1: Run all tests**

```bash
go test ./...
```

Expected: PASS

- [ ] **Step 2: Run go vet**

```bash
go vet ./...
```

Expected: PASS

- [ ] **Step 3: Build binary**

```bash
go build -o bin/sidetrail .
```

Expected: binary built successfully

- [ ] **Step 4: Test CLI manually**

```bash
bin/sidetrail --help
bin/sidetrail context --help
bin/sidetrail add --help
bin/sidetrail update --help
bin/sidetrail health --help
```

Expected: all commands show correct help text

- [ ] **Step 5: Final commit (if any fixes needed)**

```bash
git add -A
git commit -m "chore: final verification and cleanup"
```
