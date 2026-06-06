package version

import "testing"

func TestVersionDefaults(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must have a non-empty default")
	}
	if Commit == "" {
		t.Fatal("Commit must have a non-empty default")
	}
}
