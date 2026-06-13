package sidetrail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// e2eSetup creates a temp project root with .sidetrail/ directory.
// It returns the .sidetrail/ directory path (the store root).
func e2eSetup(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	storeDir := filepath.Join(root, storeDirName)
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return storeDir
}

// e2eWriteRecord writes a record JSON file and returns the path.
func e2eWriteRecord(t *testing.T, dir, filename, json string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// e2eRunCmd executes a sidetrail command and returns output.
func e2eRunCmd(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	fullArgs := append([]string{}, args...)
	fullArgs = append(fullArgs, "--root", root)
	cmd.SetArgs(fullArgs)
	err := cmd.Execute()
	return out.String(), err
}

// TestE2E_InitAddContextHealth tests the basic happy path:
// init → add → context → health.
func TestE2E_InitAddContextHealth(t *testing.T) {
	root := e2eSetup(t)

	// 1. Init
	out, err := e2eRunCmd(t, root, "init")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if !strings.Contains(out, "created") {
		t.Errorf("init output: %s", out)
	}

	// 2. Add a decision record
	decisionJSON := `{
		"id": "e2e-decision-001",
		"kind": "decision",
		"scope": "src/auth/login.go",
		"subject": "Use bcrypt for password hashing",
		"reason": "OWASP recommended, good compatibility",
		"source_type": "human",
		"author": "zhangsan",
		"created_at": "2026-06-13T10:00:00Z",
		"last_verified_at": "2026-06-13T10:00:00Z",
		"status": "active",
		"decided_at": "2026-06-13T10:00:00Z"
	}`
	decisionFile := e2eWriteRecord(t, root, "decision.json", decisionJSON)
	out, err = e2eRunCmd(t, root, "add", decisionFile)
	if err != nil {
		t.Fatalf("add decision: %v\n%s", err, out)
	}
	if !strings.Contains(out, "e2e-decision-001") {
		t.Errorf("add output should contain id, got: %s", out)
	}

	// 3. Add a constraint record
	constraintJSON := `{
		"id": "e2e-constraint-001",
		"kind": "constraint",
		"scope": "src/billing",
		"subject": "Do not modify billing code",
		"reason": "Compliance review pending, frozen until Q3",
		"source_type": "human",
		"author": "lisi",
		"created_at": "2026-06-13T10:00:00Z",
		"last_verified_at": "2026-06-13T10:00:00Z",
		"status": "active"
	}`
	constraintFile := e2eWriteRecord(t, root, "constraint.json", constraintJSON)
	out, err = e2eRunCmd(t, root, "add", constraintFile)
	if err != nil {
		t.Fatalf("add constraint: %v\n%s", err, out)
	}

	// 4. Context for src/auth/login.go should return the decision
	out, err = e2eRunCmd(t, root, "context", "--file", "src/auth/login.go", "--json")
	if err != nil {
		t.Fatalf("context: %v\n%s", err, out)
	}
	var recs []*record.Record
	if err := json.Unmarshal([]byte(out), &recs); err != nil {
		t.Fatalf("context output not JSON: %v\n%s", err, out)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record for src/auth/login.go, got %d", len(recs))
	}
	if recs[0].ID != "e2e-decision-001" {
		t.Errorf("expected decision record, got %s", recs[0].ID)
	}

	// 5. Context for src/billing/ should return the constraint
	out, err = e2eRunCmd(t, root, "context", "--file", "src/billing/invoice.go", "--json")
	if err != nil {
		t.Fatalf("context: %v\n%s", err, out)
	}
	if err := json.Unmarshal([]byte(out), &recs); err != nil {
		t.Fatalf("context output not JSON: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record for src/billing/invoice.go, got %d", len(recs))
	}
	if recs[0].ID != "e2e-constraint-001" {
		t.Errorf("expected constraint record, got %s", recs[0].ID)
	}

	// 6. Health should show 2 records
	out, err = e2eRunCmd(t, root, "health")
	if err != nil {
		t.Fatalf("health: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Total records:    2") {
		t.Errorf("health output: %s", out)
	}
	if !strings.Contains(out, "decision") {
		t.Errorf("health should mention decision kind: %s", out)
	}
	if !strings.Contains(out, "constraint") {
		t.Errorf("health should mention constraint kind: %s", out)
	}
}

// TestE2E_SupersedeWorkflow tests the supersede pattern:
// add old record → add new record → update old record status.
func TestE2E_SupersedeWorkflow(t *testing.T) {
	root := e2eSetup(t)

	// 1. Add old decision
	oldJSON := `{
		"id": "e2e-old-001",
		"kind": "decision",
		"scope": "src/auth/login.go",
		"subject": "Use bcrypt for password hashing",
		"reason": "OWASP recommended",
		"source_type": "human",
		"author": "zhangsan",
		"created_at": "2026-06-01T00:00:00Z",
		"last_verified_at": "2026-06-01T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-01T00:00:00Z"
	}`
	oldFile := e2eWriteRecord(t, root, "old.json", oldJSON)
	out, err := e2eRunCmd(t, root, "add", oldFile)
	if err != nil {
		t.Fatalf("add old: %v\n%s", err, out)
	}

	// 2. Add new decision with supersedes field
	newJSON := `{
		"id": "e2e-new-001",
		"kind": "decision",
		"scope": "src/auth/login.go",
		"subject": "Switch to argon2 for password hashing",
		"reason": "bcrypt not resistant enough to GPU attacks",
		"source_type": "human",
		"author": "zhangsan",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"supersedes": "e2e-old-001",
		"decided_at": "2026-06-13T00:00:00Z"
	}`
	newFile := e2eWriteRecord(t, root, "new.json", newJSON)
	out, err = e2eRunCmd(t, root, "add", newFile)
	if err != nil {
		t.Fatalf("add new: %v\n%s", err, out)
	}

	// 3. Update old record status to superseded
	updateJSON := `{"status":"superseded","superseded_by":"e2e-new-001"}`
	updateFile := e2eWriteRecord(t, root, "update.json", updateJSON)
	out, err = e2eRunCmd(t, root, "update", "e2e-old-001", "--file", updateFile)
	if err != nil {
		t.Fatalf("update old: %v\n%s", err, out)
	}

	// 4. Verify old record is superseded
	out, err = e2eRunCmd(t, root, "context", "--file", "src/auth/login.go", "--json")
	if err != nil {
		t.Fatalf("context: %v\n%s", err, out)
	}
	var recs []*record.Record
	if err := json.Unmarshal([]byte(out), &recs); err != nil {
		t.Fatalf("context not JSON: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d", len(recs))
	}

	var oldRec, newRec *record.Record
	for _, r := range recs {
		if r.ID == "e2e-old-001" {
			oldRec = r
		} else if r.ID == "e2e-new-001" {
			newRec = r
		}
	}
	if oldRec == nil {
		t.Fatal("old record not found in context")
	}
	if newRec == nil {
		t.Fatal("new record not found in context")
	}
	if oldRec.Status != "superseded" {
		t.Errorf("old record status: want superseded, got %s", oldRec.Status)
	}
	if oldRec.SupersededBy != "e2e-new-001" {
		t.Errorf("old record superseded_by: want e2e-new-001, got %s", oldRec.SupersededBy)
	}
	if newRec.Supersedes != "e2e-old-001" {
		t.Errorf("new record supersedes: want e2e-old-001, got %s", newRec.Supersedes)
	}

	// 5. Health should show 2 records, 1 active chain
	out, err = e2eRunCmd(t, root, "health")
	if err != nil {
		t.Fatalf("health: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Total records:    2") {
		t.Errorf("health: %s", out)
	}
	if !strings.Contains(out, "Active chains:    1") {
		t.Errorf("health should show 1 active chain: %s", out)
	}
}

// TestE2E_MultipleKinds tests adding records of different kinds.
func TestE2E_MultipleKinds(t *testing.T) {
	root := e2eSetup(t)

	records := []struct {
		id       string
		kind     string
		scope    string
		subj     string
		extra    string
		statuses string
	}{
		{"r-dec-001", "decision", "src/main.go", "Use Go modules", `"decided_at":"2026-06-13T00:00:00Z"`, "active"},
		{"r-con-001", "constraint", "src/legacy/", "Do not touch legacy code", "", "active"},
		{"r-sig-001", "signal", "src/api/", "API latency trending up", "", "active"},
		{"r-exp-001", "experiment", "src/cache/", "Try Redis for caching", `"started_at":"2026-06-13T00:00:00Z"`, "in_progress"},
		{"r-inc-001", "incident", "src/db/", "DB connection pool exhausted", `"occurred_at":"2026-06-13T00:00:00Z"`, "investigating"},
	}

	for _, rec := range records {
		extraField := ""
		if rec.extra != "" {
			extraField = "," + rec.extra
		}
		json := `{
			"id": "` + rec.id + `",
			"kind": "` + rec.kind + `",
			"scope": "` + rec.scope + `",
			"subject": "` + rec.subj + `",
			"reason": "test",
			"source_type": "human",
			"author": "tester",
			"created_at": "2026-06-13T00:00:00Z",
			"last_verified_at": "2026-06-13T00:00:00Z",
			"status": "` + rec.statuses + `"` + extraField + `
		}`
		f := e2eWriteRecord(t, root, rec.id+".json", json)
		out, err := e2eRunCmd(t, root, "add", f)
		if err != nil {
			t.Fatalf("add %s: %v\n%s", rec.id, err, out)
		}
	}

	// Health should show all 5 kinds
	out, err := e2eRunCmd(t, root, "health")
	if err != nil {
		t.Fatalf("health: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Total records:    5") {
		t.Errorf("health total: %s", out)
	}
	for _, kind := range []string{"decision", "constraint", "signal", "experiment", "incident"} {
		if !strings.Contains(out, kind) {
			t.Errorf("health should mention %s: %s", kind, out)
		}
	}
}

// TestE2E_ContextAncestors tests that context walks ancestor directories.
func TestE2E_ContextAncestors(t *testing.T) {
	root := e2eSetup(t)

	// Add record at directory scope
	dirJSON := `{
		"id": "e2e-dir-001",
		"kind": "constraint",
		"scope": "src/auth",
		"subject": "Auth module constraints",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active"
	}`
	f := e2eWriteRecord(t, root, "dir.json", dirJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Add record at file scope
	fileJSON := `{
		"id": "e2e-file-001",
		"kind": "decision",
		"scope": "src/auth/login.go",
		"subject": "Login implementation",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-13T00:00:00Z"
	}`
	f = e2eWriteRecord(t, root, "file.json", fileJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Context for src/auth/login.go should return both
	out, err := e2eRunCmd(t, root, "context", "--file", "src/auth/login.go", "--json")
	if err != nil {
		t.Fatalf("context: %v\n%s", err, out)
	}
	var recs []*record.Record
	if err := json.Unmarshal([]byte(out), &recs); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records (file + ancestor), got %d", len(recs))
	}

	// Context with radius=1 should return file + immediate parent
	// We have records at "src/auth" and "src/auth/login.go"
	// With radius=1, ancestorScopes returns ["src/auth/login.go", "src/auth"]
	// Both records match, so we get 2 records
	out, err = e2eRunCmd(t, root, "context", "--file", "src/auth/login.go", "--radius", "1", "--json")
	if err != nil {
		t.Fatalf("context: %v\n%s", err, out)
	}
	if err := json.Unmarshal([]byte(out), &recs); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(recs) != 2 {
		t.Errorf("expected 2 records with radius=1 (file + parent 'src/auth'), got %d", len(recs))
	}
}

// TestE2E_UpdatePreservesFields tests that update only changes specified fields.
func TestE2E_UpdatePreservesFields(t *testing.T) {
	root := e2eSetup(t)

	// Add a record with many fields
	addJSON := `{
		"id": "e2e-preserve-001",
		"kind": "decision",
		"scope": "src/api/handler.go",
		"subject": "Use structured logging",
		"reason": "Better observability",
		"source_type": "human",
		"author": "wangwu",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-13T00:00:00Z",
		"tags": ["logging", "observability"]
	}`
	f := e2eWriteRecord(t, root, "add.json", addJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Update only status
	updateJSON := `{"status":"archived"}`
	f = e2eWriteRecord(t, root, "update.json", updateJSON)
	if _, err := e2eRunCmd(t, root, "update", "e2e-preserve-001", "--file", f); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Read back and verify all fields preserved
	s := storage.NewStore(root)
	r, err := s.Get("e2e-preserve-001")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if r.Status != "archived" {
		t.Errorf("status: want archived, got %s", r.Status)
	}
	if r.Subject != "Use structured logging" {
		t.Errorf("subject changed: %s", r.Subject)
	}
	if r.Reason != "Better observability" {
		t.Errorf("reason changed: %s", r.Reason)
	}
	if r.Author != "wangwu" {
		t.Errorf("author changed: %s", r.Author)
	}
	if len(r.Tags) != 2 {
		t.Errorf("tags changed: %v", r.Tags)
	}
}

// TestE2E_StoreAutoDiscovery tests that commands find .sidetrail/ by walking up.
func TestE2E_StoreAutoDiscovery(t *testing.T) {
	// Create nested project structure
	projectRoot := t.TempDir()
	storeDir := filepath.Join(projectRoot, storeDirName)
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Add a record
	addJSON := `{
		"id": "e2e-discover-001",
		"kind": "decision",
		"scope": "src/main.go",
		"subject": "Auto-discovered store",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-13T00:00:00Z"
	}`
	addFile := filepath.Join(projectRoot, "record.json")
	if err := os.WriteFile(addFile, []byte(addJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run from a nested subdirectory without --root
	nestedDir := filepath.Join(projectRoot, "src", "deep", "nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(nestedDir)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"add", addFile})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add from nested dir: %v\n%s", err, out.String())
	}

	// Health should find the record
	cmd2 := newRootCmd()
	out.Reset()
	cmd2.SetOut(&out)
	cmd2.SetErr(&out)
	cmd2.SetArgs([]string{"health"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("health from nested dir: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Total records:    1") {
		t.Errorf("health from nested dir: %s", out.String())
	}
}

// TestE2E_ErrorCases tests various error conditions.
func TestE2E_ErrorCases(t *testing.T) {
	root := e2eSetup(t)

	// Add with invalid JSON
	badJSON := `not valid json`
	badFile := e2eWriteRecord(t, root, "bad.json", badJSON)
	_, err := e2eRunCmd(t, root, "add", badFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Update nonexistent record
	updateJSON := `{"status":"archived"}`
	f := e2eWriteRecord(t, root, "update.json", updateJSON)
	_, err = e2eRunCmd(t, root, "update", "nonexistent-id", "--file", f)
	if err == nil {
		t.Error("expected error for nonexistent record")
	}

	// Context with missing --file
	_, err = e2eRunCmd(t, root, "context")
	if err == nil {
		t.Error("expected error for missing --file")
	}

	// Add duplicate ID
	addJSON := `{
		"id": "e2e-dup-001",
		"kind": "decision",
		"scope": "src/test.go",
		"subject": "Test",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-13T00:00:00Z"
	}`
	f = e2eWriteRecord(t, root, "dup.json", addJSON)
	_, err = e2eRunCmd(t, root, "add", f)
	if err != nil {
		t.Fatalf("first add: %v", err)
	}
	_, err = e2eRunCmd(t, root, "add", f)
	if err == nil {
		t.Error("expected error for duplicate ID")
	}
}

// TestE2E_HealthJSON tests health command with --json flag.
func TestE2E_HealthJSON(t *testing.T) {
	root := e2eSetup(t)

	// Add a record
	addJSON := `{
		"id": "e2e-health-001",
		"kind": "decision",
		"scope": "src/main.go",
		"subject": "Health check test",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-13T00:00:00Z"
	}`
	f := e2eWriteRecord(t, root, "record.json", addJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Health with --json
	out, err := e2eRunCmd(t, root, "health", "--json")
	if err != nil {
		t.Fatalf("health: %v\n%s", err, out)
	}
	var rpt healthReport
	if err := json.Unmarshal([]byte(out), &rpt); err != nil {
		t.Fatalf("health JSON invalid: %v\n%s", err, out)
	}
	if rpt.Total != 1 {
		t.Errorf("total: want 1, got %d", rpt.Total)
	}
	if rpt.ByKind["decision"] != 1 {
		t.Errorf("by_kind decision: want 1, got %d", rpt.ByKind["decision"])
	}
	if rpt.ScopeCount != 1 {
		t.Errorf("scope_count: want 1, got %d", rpt.ScopeCount)
	}
}

// TestE2E_StaleDetection tests that stale records are detected.
func TestE2E_StaleDetection(t *testing.T) {
	root := e2eSetup(t)

	// Add a stale record (verified 120 days ago)
	staleJSON := `{
		"id": "e2e-stale-001",
		"kind": "decision",
		"scope": "src/old.go",
		"subject": "Old decision",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-01-01T00:00:00Z",
		"last_verified_at": "2026-01-01T00:00:00Z",
		"status": "active",
		"decided_at": "2026-01-01T00:00:00Z"
	}`
	f := e2eWriteRecord(t, root, "stale.json", staleJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Health should flag stale record
	out, err := e2eRunCmd(t, root, "health", "--stale-days", "90")
	if err != nil {
		t.Fatalf("health: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Stale records") {
		t.Errorf("expected stale records warning: %s", out)
	}
	if !strings.Contains(out, "Old decision") {
		t.Errorf("expected stale record subject: %s", out)
	}
}

// TestE2E_TagsPreservation tests that tags are preserved through update.
func TestE2E_TagsPreservation(t *testing.T) {
	root := e2eSetup(t)

	addJSON := `{
		"id": "e2e-tags-001",
		"kind": "constraint",
		"scope": "src/config/",
		"subject": "Config constraints",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"tags": ["security", "config"]
	}`
	f := e2eWriteRecord(t, root, "tags.json", addJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Update status only
	updateJSON := `{"status":"archived"}`
	f = e2eWriteRecord(t, root, "update.json", updateJSON)
	if _, err := e2eRunCmd(t, root, "update", "e2e-tags-001", "--file", f); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Verify tags preserved
	s := storage.NewStore(root)
	r, err := s.Get("e2e-tags-001")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(r.Tags) != 2 {
		t.Errorf("tags: want [security config], got %v", r.Tags)
	}
}

// TestE2E_TimeOrdering tests that context returns records newest first.
func TestE2E_TimeOrdering(t *testing.T) {
	root := e2eSetup(t)

	// Add two records with different timestamps
	oldJSON := `{
		"id": "e2e-time-old",
		"kind": "decision",
		"scope": "src/order.go",
		"subject": "Old decision",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-01T00:00:00Z",
		"last_verified_at": "2026-06-01T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-01T00:00:00Z"
	}`
	f := e2eWriteRecord(t, root, "old.json", oldJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add old: %v", err)
	}

	newJSON := `{
		"id": "e2e-time-new",
		"kind": "decision",
		"scope": "src/order.go",
		"subject": "New decision",
		"reason": "test",
		"source_type": "human",
		"author": "tester",
		"created_at": "2026-06-13T00:00:00Z",
		"last_verified_at": "2026-06-13T00:00:00Z",
		"status": "active",
		"decided_at": "2026-06-13T00:00:00Z"
	}`
	f = e2eWriteRecord(t, root, "new.json", newJSON)
	if _, err := e2eRunCmd(t, root, "add", f); err != nil {
		t.Fatalf("add new: %v", err)
	}

	// Context should return newest first
	out, err := e2eRunCmd(t, root, "context", "--file", "src/order.go", "--json")
	if err != nil {
		t.Fatalf("context: %v\n%s", err, out)
	}
	var recs []*record.Record
	if err := json.Unmarshal([]byte(out), &recs); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d", len(recs))
	}
	if recs[0].ID != "e2e-time-new" {
		t.Errorf("first record should be newest, got %s", recs[0].ID)
	}
	if recs[1].ID != "e2e-time-old" {
		t.Errorf("second record should be oldest, got %s", recs[1].ID)
	}
}
