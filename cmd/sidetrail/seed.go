package sidetrail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/seed"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// seedOptions carries the flags for the `seed` command.
type seedOptions struct {
	root   string
	files  string
	apply  string
	dryRun bool
	jsonO  bool
}

// newSeedCmd builds the `sidetrail seed` subcommand. It has two modes:
// 1. --files: generate a prompt for the host agent to extract records
// 2. --apply: apply agent-generated records with conflict detection
func newSeedCmd() *cobra.Command {
	opts := &seedOptions{}
	cmd := &cobra.Command{
		Use:   "seed [--files <glob>] [--apply <file>] [--dry-run] [--json]",
		Short: "Seed records from project documents or apply agent-generated records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeed(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.files, "files", "", "glob pattern for project documents to read")
	cmd.Flags().StringVar(&opts.apply, "apply", "", "file containing JSON array of records to apply")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "show what would happen without writing")
	cmd.Flags().BoolVar(&opts.jsonO, "json", false, "emit JSON output instead of text")
	return cmd
}

// runSeed dispatches to the appropriate mode based on flags.
func runSeed(cmd *cobra.Command, _ []string, opts *seedOptions) error {
	if opts.files != "" && opts.apply != "" {
		return fmt.Errorf("--files and --apply are mutually exclusive")
	}
	if opts.files == "" && opts.apply == "" {
		return fmt.Errorf("either --files or --apply is required")
	}

	if opts.files != "" {
		return runSeedFiles(cmd, opts)
	}
	return runSeedApply(cmd, opts)
}

// runSeedFiles reads documents and outputs a prompt for the host agent.
func runSeedFiles(cmd *cobra.Command, opts *seedOptions) error {
	files, err := filepath.Glob(opts.files)
	if err != nil {
		return fmt.Errorf("glob %q: %w", opts.files, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no files matched pattern %q", opts.files)
	}

	prompt, err := seed.GeneratePrompt(context.Background(), files)
	if err != nil {
		return err
	}

	if opts.jsonO {
		output := map[string]string{"prompt": prompt}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	fmt.Fprint(cmd.OutOrStdout(), prompt)
	return nil
}

// runSeedApply reads candidate records and applies them to the store.
func runSeedApply(cmd *cobra.Command, opts *seedOptions) error {
	data, err := os.ReadFile(opts.apply)
	if err != nil {
		return fmt.Errorf("read %q: %w", opts.apply, err)
	}

	var candidates []*record.Record
	if err := json.Unmarshal(data, &candidates); err != nil {
		return fmt.Errorf("parse %q: %w", opts.apply, err)
	}

	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	store := storage.NewStore(root)

	conflicts, nonConflicting, err := seed.DetectConflicts(store, candidates)
	if err != nil {
		return fmt.Errorf("detect conflicts: %w", err)
	}

	if !opts.dryRun {
		for _, r := range nonConflicting {
			if r.ID == "" {
				id, err := record.NewID()
				if err != nil {
					return fmt.Errorf("generate id: %w", err)
				}
				r.ID = id
			}
			if _, err := store.Write(r); err != nil {
				return fmt.Errorf("write %s: %w", r.Subject, err)
			}
		}
	}

	if opts.jsonO {
		return seedApplyJSON(cmd, conflicts, nonConflicting, opts.dryRun)
	}
	return seedApplyTable(cmd, conflicts, nonConflicting, opts.dryRun)
}

// seedApplyJSON outputs the apply result as JSON.
func seedApplyJSON(cmd *cobra.Command, conflicts []seed.Conflict, nonConflicting []*record.Record, dryRun bool) error {
	type applyResult struct {
		DryRun         bool             `json:"dry_run"`
		Conflicts      []seed.Conflict  `json:"conflicts"`
		NonConflicting []*record.Record `json:"non_conflicting"`
	}
	result := applyResult{
		DryRun:         dryRun,
		Conflicts:      conflicts,
		NonConflicting: nonConflicting,
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

// seedApplyTable outputs the apply result as a human-readable table.
func seedApplyTable(cmd *cobra.Command, conflicts []seed.Conflict, nonConflicting []*record.Record, dryRun bool) error {
	out := cmd.OutOrStdout()
	if dryRun {
		fmt.Fprintln(out, "DRY RUN - no changes will be written")
		fmt.Fprintln(out)
	}

	if len(conflicts) > 0 {
		fmt.Fprintf(out, "Conflicts found: %d\n", len(conflicts))
		for _, c := range conflicts {
			fmt.Fprintf(out, "  Existing: %s %s %s (scope: %s)\n", c.Existing.ID, c.Existing.Kind, c.Existing.Subject, c.Existing.Scope)
			fmt.Fprintf(out, "  Candidate: %s %s (scope: %s)\n", c.Candidate.Kind, c.Candidate.Subject, c.Candidate.Scope)
			fmt.Fprintln(out)
		}
	}

	if len(nonConflicting) > 0 {
		fmt.Fprintf(out, "Records to add: %d\n", len(nonConflicting))
		for _, r := range nonConflicting {
			fmt.Fprintf(out, "  %s %s (scope: %s)\n", r.Kind, r.Subject, r.Scope)
		}
	}

	if len(conflicts) == 0 && len(nonConflicting) == 0 {
		fmt.Fprintln(out, "No records to process")
	}

	return nil
}
