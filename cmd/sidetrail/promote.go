package sidetrail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// promoteOptions carries the flags for the `promote` command.
type promoteOptions struct {
	root string
	all  bool
}

// newPromoteCmd builds the `sidetrail promote` subcommand. It
// moves seed records from .sidetrail/_seed/ into the active store
// (the kind-specific directory). Seeds are scrape-derived
// candidates waiting for human review; promote is the approval
// gate.
func newPromoteCmd() *cobra.Command {
	opts := &promoteOptions{}
	cmd := &cobra.Command{
		Use:   "promote [<id-or-prefix>...] [--all]",
		Short: "Promote seed records from _seed/ to the active store",
		Long: `promote moves one or more seed records from .sidetrail/_seed/
into the appropriate kind directory (decisions/, constraints/, etc.).

Without arguments and without --all, it lists the available seeds.
With --all, every seed is promoted. With explicit IDs or prefixes,
only matching seeds are promoted.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPromote(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "store root (default: auto-detect)")
	cmd.Flags().BoolVar(&opts.all, "all", false, "promote all seeds")
	return cmd
}

// runPromote resolves the store, lists seeds, and either displays
// them or promotes the specified ones.
func runPromote(cmd *cobra.Command, args []string, opts *promoteOptions) error {
	storeRoot, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(storeRoot)

	seedDir := filepath.Join(storeRoot, "_seed")
	entries, err := os.ReadDir(seedDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(cmd.OutOrStdout(), "no _seed/ directory found; run sidetrail init first\n")
			return nil
		}
		return fmt.Errorf("read _seed/: %w", err)
	}

	var seeds []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			seeds = append(seeds, e)
		}
	}

	// No arguments and no --all: list available seeds or report empty.
	if len(args) == 0 && !opts.all {
		if len(seeds) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "no seeds to promote\n")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "available seeds in %s:\n", seedDir)
		for _, e := range seeds {
			path := filepath.Join(seedDir, e.Name())
			r, err := s.Read(path)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "  %s (unreadable: %v)\n", e.Name(), err)
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-12s  %s\n", r.ID, r.Kind, r.Subject)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\npromote with: sidetrail promote <id> or sidetrail promote --all\n")
		return nil
	}

	// Determine which seeds to promote.
	toPromote := seeds
	if !opts.all {
		toPromote = nil
		for _, arg := range args {
			matched := matchSeeds(seeds, seedDir, s, arg)
			if len(matched) == 0 {
				return fmt.Errorf("no seed matching %q", arg)
			}
			toPromote = append(toPromote, matched...)
		}
	}

	promoted := 0
	for _, e := range toPromote {
		path := filepath.Join(seedDir, e.Name())
		r, err := s.Read(path)
		if err != nil {
			return fmt.Errorf("read seed %q: %w", e.Name(), err)
		}
		if _, err := s.Write(r); err != nil {
			return fmt.Errorf("promote %q: %w", e.Name(), err)
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove seed %q: %w", e.Name(), err)
		}
		promoted++
		fmt.Fprintf(cmd.OutOrStdout(), "promoted %s (%s) -> %s/\n", r.ID, r.Kind, pluralize(string(r.Kind)))
	}

	fmt.Fprintf(cmd.OutOrStdout(), "promoted %d seed(s)\n", promoted)
	return nil
}

// matchSeeds returns seed entries whose ID or ID prefix matches
// arg. It reads each seed file only when needed.
func matchSeeds(seeds []os.DirEntry, seedDir string, s *storage.Store, arg string) []os.DirEntry {
	var out []os.DirEntry
	for _, e := range seeds {
		path := filepath.Join(seedDir, e.Name())
		r, err := s.Read(path)
		if err != nil {
			continue
		}
		if r.ID == arg || strings.HasPrefix(r.ID, arg) {
			out = append(out, e)
		}
	}
	return out
}

// pluralize maps a kind name to its conventional on-disk
// directory name.
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

// Ensure record.Kind is used to avoid unused import.
var _ = record.KindDecision
