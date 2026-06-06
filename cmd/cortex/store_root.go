package cortex

import (
	"fmt"
	"os"
	"path/filepath"
)

// findStoreRoot walks upward from start, looking for the first
// directory that contains a `.cortex/` subdirectory. It returns
// the absolute path of that `.cortex/` directory. The walk stops
// at the filesystem root and never follows symlinks into other
// parts of the tree.
//
// The default start is the current working directory. Callers
// may pass an explicit path to override (used by tests).
func findStoreRoot(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs %q: %w", start, err)
	}
	dir := abs
	for {
		candidate := filepath.Join(dir, ".cortex")
		info, err := os.Stat(candidate)
		switch {
		case err == nil && info.IsDir():
			return candidate, nil
		case err != nil && !os.IsNotExist(err):
			return "", fmt.Errorf("stat %q: %w", candidate, err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .cortex/ found from %q upward", abs)
		}
		dir = parent
	}
}

// resolveStoreRoot returns the .cortex/ directory the CLI should
// use. When explicit is non-empty, it must be a directory that
// already contains a .cortex/ subdirectory; when it is empty,
// findStoreRoot searches upward from the current working
// directory.
func resolveStoreRoot(explicit string) (string, error) {
	if explicit == "" {
		return findStoreRoot("")
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
