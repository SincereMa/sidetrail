package sidetrail

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/storage"
)

// statusOptions carries the flags for the `status` command.
type statusOptions struct {
	root   string
	dryRun bool
}

// newStatusCmd builds the `sidetrail status` subcommand. It
// transitions a record's status field. The allowed transitions
// are:
//
//   active → superseded  (via `supersede`, not here)
//   active → archived    (via `archive`)
//   active → hidden      (via `hide`)
//   archived → active    (via `activate`)
//   hidden → active      (via `activate`)
//
// The command does not chain transitions; each call makes one
// atomic status change.
func newStatusCmd() *cobra.Command {
	opts := &statusOptions{}
	cmd := &cobra.Command{
		Use:   "status <id> <activate|archive|hide>",
		Short: "Transition a record's status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "do not write; print what would change")
	return cmd
}

// validTransitions maps the current status to the set of
// statuses it may transition to.
var validTransitions = map[string][]string{
	"active":      {"archived", "hidden"},
	"archived":    {"active"},
	"hidden":      {"active"},
	"superseded":  {},
	"in_progress": {"succeeded", "failed", "inconclusive", "abandoned"},
	"succeeded":   {},
	"failed":      {},
	"inconclusive": {},
	"abandoned":   {},
	"investigating": {"mitigated", "resolved"},
	"mitigated":   {"resolved"},
	"resolved":    {},
}

// runStatus reads the record, validates the transition, applies
// it, and writes the record back. A --dry-run prints the
// proposed change without touching the file.
func runStatus(cmd *cobra.Command, args []string, opts *statusOptions) error {
	id, action := args[0], args[1]
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	r, err := s.Get(id)
	if err != nil {
		return err
	}
	targets, ok := validTransitions[r.Status]
	if !ok {
		return fmt.Errorf("record %q has status %q which has no defined transitions", r.ID, r.Status)
	}
	found := false
	for _, t := range targets {
		if t == action {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot transition from %q to %q (allowed: %v)", r.Status, action, targets)
	}
	if opts.dryRun {
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "would change %s status: %s -> %s\n", r.ID, r.Status, action)
		return nil
	}
	r.Status = action
	if _, err := s.Write(r); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s %s\n", r.ID, r.Status)
	return nil
}
