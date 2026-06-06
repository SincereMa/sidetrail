package cortex

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelp(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--help returned error: %v", err)
	}
	if !strings.Contains(out.String(), "Cortex SideMark") {
		t.Fatalf("expected help to mention Cortex SideMark, got: %q", out.String())
	}
}

func TestRootVersion(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--version returned error: %v", err)
	}
	if !strings.Contains(out.String(), "cortex ") {
		t.Fatalf("expected version output to start with 'cortex ', got: %q", out.String())
	}
}
