// Package version is the test surface for the version package.
package version

import "testing"

// TestVersionDefaults ensures the unbuilt binary still has
// non-empty Version and Commit values. A regression here would
// mean the package was changed in a way that broke the
// ldflags-injection contract.
func TestVersionDefaults(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must have a non-empty default")
	}
	if Commit == "" {
		t.Fatal("Commit must have a non-empty default")
	}
}
