package cortex

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/SincereMa/cortex-sidemark/internal/record"
	"github.com/SincereMa/cortex-sidemark/internal/storage"
)

// initOptions carries the flags for the `init` command.
type initOptions struct {
	root    string
	noWrite bool
}

// initScanPaths enumerates the project paths cortex init looks
// at. The list comes from ADR-0001. Patterns containing a
// slash are taken relative to the project root; the rest are
// globbed at the project root.
var initScanPaths = []string{
	"README*",
	"CONTRIBUTING*",
	"AGENTS*",
	"CLAUDE*",
	"LICENSE*",
	"RUNBOOK*",
	"docs/adr*",
	"docs/decisions",
	"docs/architecture",
	"docs/runbooks",
	".github/PULL_REQUEST_TEMPLATE.md",
	".github/ISSUE_TEMPLATE",
}

// seedBodyLimit caps the body of a seed record at 500 bytes so
// a large README does not bloat the seed.
const seedBodyLimit = 500

// newInitCmd builds the `cortex init` subcommand. It seeds a
// fresh .cortex/ store by walking a fixed list of project
// paths and writing a candidate record for each existing file
// under .cortex/_seed/. Seeds are scrape-derived candidates;
// they stay in _seed/ until a human moves them. The store is
// usable from empty; init is optional.
func newInitCmd() *cobra.Command {
	opts := &initOptions{}
	cmd := &cobra.Command{
		Use:   "init [--root <project>] [--no-write]",
		Short: "Seed a .cortex/ store from existing project docs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "project root where .cortex/ will be created (default: CWD)")
	cmd.Flags().BoolVar(&opts.noWrite, "no-write", false, "do not write seeds; print what would be done")
	return cmd
}

// runInit resolves the project root, scans the canonical
// paths, and writes (or reports) seeds.
func runInit(cmd *cobra.Command, _ []string, opts *initOptions) error {
	projectRoot, err := resolveProjectRoot(opts.root)
	if err != nil {
		return err
	}
	seeds := collectSeeds(projectRoot)
	sort.Slice(seeds, func(i, j int) bool { return seeds[i].path < seeds[j].path })

	if opts.noWrite {
		return reportPlan(cmd, projectRoot, seeds)
	}

	cortexDir := filepath.Join(projectRoot, ".cortex")
	s := storage.NewStore(cortexDir)
	now := time.Now().UTC()
	written := 0
	for _, c := range seeds {
		r, err := buildSeedRecord(c, now)
		if err != nil {
			return fmt.Errorf("build seed for %q: %w", c.path, err)
		}
		if _, err := s.WriteSeed(r); err != nil {
			return fmt.Errorf("write seed for %q: %w", c.path, err)
		}
		written++
	}
	fmt.Fprintf(cmd.OutOrStdout(), "scanned %d path(s); wrote %d seed(s) to %s\n", len(seeds), written, filepath.Join(cortexDir, "_seed"))
	return nil
}

// resolveProjectRoot returns the absolute project root. When
// opts.root is empty, the current working directory is used.
// Unlike the other commands, init's --root is the project root
// (where .cortex/ will be created), not the .cortex/ path
// itself, because the .cortex/ directory does not exist yet.
func resolveProjectRoot(explicit string) (string, error) {
	if explicit == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
		return wd, nil
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

// seedCandidate is one project file we plan to seed from.
type seedCandidate struct {
	path string // path relative to project root
	body string // first seedBodyLimit bytes, or "" if unreadable / binary
}

// collectSeeds walks initScanPaths relative to projectRoot and
// returns one seedCandidate per existing file. A path that
// contains a glob meta-character is expanded with filepath.Glob;
// a path that points at an existing directory is enumerated
// non-recursively. A path that matches nothing is skipped
// silently — a fresh project typically has only a few of the
// canonical paths.
func collectSeeds(projectRoot string) []seedCandidate {
	var out []seedCandidate
	seen := make(map[string]struct{})
	for _, p := range initScanPaths {
		full := filepath.Join(projectRoot, p)
		if hasMeta(p) {
			matches, err := filepath.Glob(full)
			if err != nil {
				continue
			}
			for _, m := range matches {
				info, err := os.Stat(m)
				if err != nil || info.IsDir() {
					continue
				}
				rel := filepath.ToSlash(m[len(projectRoot)+1:])
				if _, dup := seen[rel]; dup {
					continue
				}
				seen[rel] = struct{}{}
				if c, ok := readCandidate(rel, m); ok {
					out = append(out, c)
				}
			}
			continue
		}
		info, err := os.Stat(full)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			if _, dup := seen[p]; dup {
				continue
			}
			seen[p] = struct{}{}
			if c, ok := readCandidate(p, full); ok {
				out = append(out, c)
			}
			continue
		}
		entries, err := os.ReadDir(full)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			rel := filepath.ToSlash(filepath.Join(p, e.Name()))
			if _, dup := seen[rel]; dup {
				continue
			}
			seen[rel] = struct{}{}
			abs := filepath.Join(full, e.Name())
			if c, ok := readCandidate(rel, abs); ok {
				out = append(out, c)
			}
		}
	}
	return out
}

// hasMeta reports whether p contains any glob meta-character.
func hasMeta(p string) bool {
	return strings.ContainsAny(p, "*?[")
}

// readCandidate returns a seedCandidate for a single file. A
// file that cannot be read at all, or whose first chunk looks
// like binary, is reported as (zero, false) and skipped.
func readCandidate(relPath, absPath string) (seedCandidate, bool) {
	f, err := os.Open(absPath)
	if err != nil {
		return seedCandidate{}, false
	}
	defer f.Close()
	buf := make([]byte, seedBodyLimit)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return seedCandidate{}, false
	}
	buf = buf[:n]
	if bytes.IndexByte(buf, 0) >= 0 {
		return seedCandidate{}, false
	}
	return seedCandidate{path: relPath, body: string(buf)}, true
}

// buildSeedRecord turns a seedCandidate into a schema-valid
// record. The Kind is fixed to "decision" (the most general
// kind) and the Status to "active" — the location under
// _seed/ is the signal that the record is a candidate, not
// the Status field.
func buildSeedRecord(c seedCandidate, now time.Time) (*record.Record, error) {
	subject := strings.TrimSuffix(filepath.Base(c.path), filepath.Ext(c.path))
	r := &record.Record{
		Kind:           record.KindDecision,
		Scope:          c.path,
		Subject:        subject,
		Reason:         fmt.Sprintf("Auto-seeded from %q. Review the file and promote to a real record when relevant.", c.path),
		SourceType:     record.SourceScrape,
		Author:         "cortex init",
		CreatedAt:      now,
		LastVerifiedAt: now,
		Status:         "active",
	}
	if c.body != "" {
		r.Body = c.body
	}
	id, err := record.NewID()
	if err != nil {
		return nil, err
	}
	r.ID = id
	return r, nil
}

// reportPlan prints the would-be seed list for a dry run.
func reportPlan(cmd *cobra.Command, projectRoot string, seeds []seedCandidate) error {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "would scan %d path(s) under %s\n", len(initScanPaths), projectRoot)
	fmt.Fprintf(out, "would write %d seed(s) to %s\n", len(seeds), filepath.Join(projectRoot, ".cortex", "_seed"))
	for _, c := range seeds {
		fmt.Fprintf(out, "  - %s\n", c.path)
	}
	return nil
}
