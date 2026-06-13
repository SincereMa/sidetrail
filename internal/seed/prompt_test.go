package seed

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePrompt(t *testing.T) {
	dir := t.TempDir()
	doc1 := filepath.Join(dir, "README.md")
	doc2 := filepath.Join(dir, "ARCHITECTURE.md")
	os.WriteFile(doc1, []byte("# Project\nUse bcrypt for auth."), 0644)
	os.WriteFile(doc2, []byte("# Architecture\nDon't modify billing."), 0644)

	prompt, err := GeneratePrompt(context.Background(), []string{doc1, doc2})
	if err != nil {
		t.Fatal(err)
	}
	if prompt == "" {
		t.Error("expected non-empty prompt")
	}
	if !strings.Contains(prompt, "README.md") {
		t.Error("expected prompt to contain README.md")
	}
	if !strings.Contains(prompt, "ARCHITECTURE.md") {
		t.Error("expected prompt to contain ARCHITECTURE.md")
	}
	if !strings.Contains(prompt, "Use bcrypt for auth.") {
		t.Error("expected prompt to contain document content")
	}
}

func TestGeneratePrompt_EmptyFiles(t *testing.T) {
	_, err := GeneratePrompt(context.Background(), []string{})
	if err == nil {
		t.Error("expected error for empty files")
	}
}

func TestGeneratePrompt_NonexistentFile(t *testing.T) {
	_, err := GeneratePrompt(context.Background(), []string{"/nonexistent/file.md"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
