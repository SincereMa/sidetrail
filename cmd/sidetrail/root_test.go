// Package sidetrail is the test surface for the root command.
package sidetrail

import (
	"bytes"
	"strings"
	"testing"
)

// TestRootHelp verifies the --help output mentions the project
// name. A passing test here is a smoke check that the long
// description is wired in.
func TestRootHelp(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--help returned error: %v", err)
	}
	if !strings.Contains(out.String(), "SideTrail") {
		t.Fatalf("expected help to mention SideTrail, got: %q", out.String())
	}
}

// TestRootVersion verifies the --version output starts with the
// binary name. The exact version string is set at build time via
// ldflags, so the test only checks the prefix.
func TestRootVersion(t *testing.T) {
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--version returned error: %v", err)
	}
	if !strings.Contains(out.String(), "sidetrail ") {
		t.Fatalf("expected version output to start with 'sidetrail ', got: %q", out.String())
	}
}
