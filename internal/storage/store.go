// Package storage reads and writes SideTrail record files
// on disk under a project-local .sidetrail/ root.
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

	"github.com/SincereMa/sidetrail/internal/record"
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
	return s.writeToDir(r, dir)
}

// WriteSeed persists r under the .sidetrail/_seed/ subdirectory.
// Seeds are scrape-derived candidates waiting for human review;
// they live in their own subdirectory so the canonical kind
// listings (decisions, constraints, ...) do not surface them
// without an explicit --include-seed flag (a future PR).
//
// The Kind field on a seed record is still required to be valid;
// WriteSeed is "where to write" not "what to validate".
func (s *Store) WriteSeed(r *record.Record) (string, error) {
	if !r.Kind.Valid() {
		return "", fmt.Errorf("invalid kind: %q", r.Kind)
	}
	if r.ID == "" {
		return "", fmt.Errorf("record id must not be empty")
	}
	dir := filepath.Join(s.root, "_seed")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %q: %w", dir, err)
	}
	return s.writeToDir(r, dir)
}

// writeToDir is the common path of Write and WriteSeed. The
// caller has already validated the record and ensured dir
// exists. The atomic-write plumbing lives here so both entry
// points behave identically.
func (s *Store) writeToDir(r *record.Record, dir string) (string, error) {
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

// ListAll reads every record under the store, across all kinds.
// Results are returned sorted by CreatedAt descending (newest
// first). A store with no records returns (nil, nil). Read errors
// on any individual file are returned to the caller; the store
// does not silently skip a corrupt file.
func (s *Store) ListAll() ([]*record.Record, error) {
	var out []*record.Record
	for _, k := range allKinds {
		paths, err := s.List(k)
		if err != nil {
			return nil, err
		}
		for _, p := range paths {
			r, err := s.Read(p)
			if err != nil {
				return nil, err
			}
			out = append(out, r)
		}
	}
	sortByCreatedAtDesc(out)
	return out, nil
}

// ListKind reads every record of kind k under the store. Results
// are returned sorted by CreatedAt descending. A missing kind
// directory returns (nil, nil).
func (s *Store) ListKind(k record.Kind) ([]*record.Record, error) {
	paths, err := s.List(k)
	if err != nil {
		return nil, err
	}
	out := make([]*record.Record, 0, len(paths))
	for _, p := range paths {
		r, err := s.Read(p)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	sortByCreatedAtDesc(out)
	return out, nil
}

// Get finds a record by id. The lookup is exact-match first; if
// no record has id exactly equal to id, the store falls back to
// prefix match on the on-disk filename. A prefix that matches
// more than one file is reported as an ambiguity error so the
// caller can disambiguate.
func (s *Store) Get(id string) (*record.Record, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}
	exactPath, err := s.findExact(id)
	if err != nil {
		return nil, err
	}
	if exactPath != "" {
		return s.Read(exactPath)
	}
	matches, err := s.findPrefix(id)
	if err != nil {
		return nil, err
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no record with id %q", id)
	case 1:
		return s.Read(matches[0])
	default:
		return nil, fmt.Errorf("id %q is ambiguous; matches %d records", id, len(matches))
	}
}

// findExact returns the on-disk path of the record file whose
// "<id>-" prefix matches id exactly, or "" if none exists. It
// searches every kind directory once and never errors for a
// missing directory.
func (s *Store) findExact(id string) (string, error) {
	prefix := id + "-"
	for _, k := range allKinds {
		paths, err := s.List(k)
		if err != nil {
			return "", err
		}
		for _, p := range paths {
			if strings.HasPrefix(filepath.Base(p), prefix) {
				return p, nil
			}
		}
	}
	return "", nil
}

// findPrefix returns the on-disk paths of every record file whose
// "<id>-" prefix starts with id. It is used as the fallback path
// for Get when no exact match exists.
func (s *Store) findPrefix(id string) ([]string, error) {
	var matches []string
	for _, k := range allKinds {
		paths, err := s.List(k)
		if err != nil {
			return nil, err
		}
		for _, p := range paths {
			if strings.HasPrefix(filepath.Base(p), id) {
				matches = append(matches, p)
			}
		}
	}
	sort.Strings(matches)
	return matches, nil
}

// allKinds is the canonical iteration order for cross-kind
// listings. Defining it once here keeps the on-disk layout and
// the in-memory listing order in lock-step.
var allKinds = []record.Kind{
	record.KindDecision,
	record.KindConstraint,
	record.KindSignal,
	record.KindExperiment,
	record.KindIncident,
}

// sortByCreatedAtDesc sorts recs in place by CreatedAt descending.
func sortByCreatedAtDesc(recs []*record.Record) {
	sort.SliceStable(recs, func(i, j int) bool {
		return recs[i].CreatedAt.After(recs[j].CreatedAt)
	})
}

// WriteDraft persists r under the .sidetrail/_draft/ subdirectory.
// Drafts are complete, schema-valid records waiting for human
// review before being promoted to the main store via `sidetrail promote`.
func (s *Store) WriteDraft(r *record.Record) (string, error) {
	if !r.Kind.Valid() {
		return "", fmt.Errorf("invalid kind: %q", r.Kind)
	}
	if r.ID == "" {
		return "", fmt.Errorf("record id must not be empty")
	}
	dir := filepath.Join(s.root, "_draft")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %q: %w", dir, err)
	}
	return s.writeToDir(r, dir)
}

// Ask returns records whose scope matches the pattern, optionally
// filtered by kind and tag, sorted newest first and capped at
// limit. A non-positive limit means "no cap".
//
// The scope match rule is record.MatchScope: exact match or
// strict descendant. An empty scope pattern returns nil; the
// caller is expected to require a scope.
func (s *Store) Ask(scope, kind, tag string, limit int) ([]*record.Record, error) {
	all, err := s.ListAll()
	if err != nil {
		return nil, err
	}
	out := make([]*record.Record, 0, len(all))
	for _, r := range all {
		if scope != "" && !record.MatchScope(r.Scope, scope) {
			continue
		}
		if kind != "" && string(r.Kind) != kind {
			continue
		}
		if tag != "" && !recordHasTag(r, tag) {
			continue
		}
		out = append(out, r)
	}
	sortByCreatedAtDesc(out)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ContextFor returns records relevant to a file: the records
// whose scope equals the file path, plus the records whose scope
// is an ancestor directory of the file path, up to `radius`
// levels. Results are sorted newest first and capped at limit.
//
// Radius semantics: radius=0 keeps only the file path itself;
// radius=1 also includes the immediate parent directory;
// radius=2 adds the grandparent; and so on. A non-positive
// radius means "walk all the way to the filesystem root".
//
// The file path is used as-is. It is not resolved against the
// store root and is not required to exist on disk; the function
// is a retrieval over stored scopes, not a filesystem walk.
func (s *Store) ContextFor(file string, radius, limit int) ([]*record.Record, error) {
	if file == "" {
		return nil, fmt.Errorf("file must not be empty")
	}
	patterns := ancestorScopes(file, radius)
	all, err := s.ListAll()
	if err != nil {
		return nil, err
	}
	out := make([]*record.Record, 0, len(all))
	seen := make(map[string]struct{}, len(all))
	for _, r := range all {
		if !scopeMatchesAny(r.Scope, patterns) {
			continue
		}
		if _, dup := seen[r.ID]; dup {
			continue
		}
		seen[r.ID] = struct{}{}
		out = append(out, r)
	}
	sortByCreatedAtDesc(out)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// recordHasTag reports whether r has tag in its Tags slice. An
// empty tag list never matches.
func recordHasTag(r *record.Record, tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// ancestorScopes returns the candidate scopes for ContextFor:
// the file path itself, then each parent directory up to radius
// levels deep. A non-positive radius means "no limit".
func ancestorScopes(file string, radius int) []string {
	clean := filepath.Clean(file)
	dir := filepath.Dir(clean)
	out := []string{clean}
	if dir == clean || dir == "." || dir == string(filepath.Separator) {
		return out
	}
	for i := 0; ; i++ {
		out = append(out, dir)
		if radius > 0 && i+1 >= radius {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return out
}

// scopeMatchesAny reports whether scope matches any pattern in
// patterns. An exact equality on any pattern is sufficient;
// descendant match is not applied here because the patterns come
// from a fixed ancestor walk, so equality already captures
// "the file's own scope" and "an ancestor directory's scope".
func scopeMatchesAny(scope string, patterns []string) bool {
	for _, p := range patterns {
		if scope == p {
			return true
		}
	}
	return false
}
