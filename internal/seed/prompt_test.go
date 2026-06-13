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
