// Package storage reads and writes Cortex SideMark record files
// on disk under a project-local .cortex/ root.
//
// The on-disk layout is one record per file, identified by a
// ULID. Writes are atomic. The store does not interpret record
// contents beyond dispatching to a kind-specific directory.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SincereMa/cortex-sidemark/internal/record"
)

// Store is the on-disk record store rooted at Root. The zero
// value is not usable; construct with NewStore.
type Store struct {
	root string
}

// NewStore returns a Store rooted at root. The root is created
// on first write, not on construction.
func NewStore(root string) *Store {
	return &Store{root: root}
}

// Root returns the directory the store writes under. Useful for
// tests and for surfacing to the user in error messages.
func (s *Store) Root() string { return s.root }

// kindDir returns the on-disk subdirectory for kind k.
func (s *Store) kindDir(k record.Kind) string {
	return filepath.Join(s.root, pluralize(string(k)))
}

// pluralize maps a kind name to its conventional on-disk
// directory name. Centralized here so the layout is one edit.
func pluralize(kind string) string {
	switch kind {
	case "decision":
		return "decisions"
	case "constraint":
		return "constraints"
	case "signal":
		return "signals"
	case "experiment":
		return "experiments"
	case "incident":
		return "incidents"
	}
	return kind + "s"
}

// Write persists r to disk and returns the absolute path of the
// written file. The write is atomic: a temporary file is created
// alongside the destination and renamed into place. The path
// layout is "<root>/<kinddir>/<id>-<slug>.json".
func (s *Store) Write(r *record.Record) (string, error) {
	if !r.Kind.Valid() {
		return "", fmt.Errorf("invalid kind: %q", r.Kind)
	}
	if r.ID == "" {
		return "", fmt.Errorf("record id must not be empty")
	}
	dir := s.kindDir(r.Kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %q: %w", dir, err)
	}
	slug := record.Slug(r.Subject)
	name := fmt.Sprintf("%s-%s.json", r.ID, slug)
	path := filepath.Join(dir, name)

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", fmt.Errorf("write tmp %q: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("rename %q -> %q: %w", tmp, path, err)
	}
	return path, nil
}

// Read loads a record from path. The file is parsed into a
// record.Record; the schema is not re-validated here (the caller
// may have just produced the file).
func (s *Store) Read(path string) (*record.Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	var r record.Record
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", path, err)
	}
	return &r, nil
}

// List returns the absolute paths of every record file of kind k,
// sorted by name (which is also chronological for ULID-based
// names). A missing kind directory returns (nil, nil); it is not
// an error.
func (s *Store) List(k record.Kind) ([]string, error) {
	if !k.Valid() {
		return nil, fmt.Errorf("invalid kind: %q", k)
	}
	dir := s.kindDir(k)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("readdir %q: %w", dir, err)
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}
