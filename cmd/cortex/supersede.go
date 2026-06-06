package cortex

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SincereMa/cortex-sidemark/internal/record"
	"github.com/SincereMa/cortex-sidemark/internal/storage"
)

// supersedeOptions carries the flags for the `supersede`
// command.
type supersedeOptions struct {
	root   string
	new    string
	dryRun bool
}

// newSupersedeCmd builds the `cortex supersede` subcommand.
// It marks an existing record as superseded and, in the same
// transaction, writes a new replacement record. The
// relationship is wired in both directions: the old record's
// status becomes "superseded" and its superseded_by is set to
// the new id; the new record's supersedes is set to the old id
// (unless the input file already provided one).
//
// `--new` is the path to a complete, schema-valid record file.
// The file's id may be pre-set or may be left to the sidecar
// to generate. If pre-set, the new id must not already exist
// in the store.
func newSupersedeCmd() *cobra.Command {
	opts := &supersedeOptions{}
	cmd := &cobra.Command{
		Use:   "supersede <old-id> --new <file>",
		Short: "Mark a record as superseded and add a replacement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSupersede(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .cortex/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.new, "new", "", "path to a record file that replaces the old record (required)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "do not write; print what would change")
	_ = cmd.MarkFlagRequired("new")
	return cmd
}

// runSupersede orchestrates the swap.
func runSupersede(cmd *cobra.Command, args []string, opts *supersedeOptions) error {
	if opts.new == "" {
		return fmt.Errorf("--new is required")
	}
	newRec, err := record.LoadFile(opts.new)
	if err != nil {
		return err
	}
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	oldRec, err := s.Get(args[0])
	if err != nil {
		return err
	}
	if newRec.ID == oldRec.ID {
		return fmt.Errorf("replacement must not have the same id as the record it supersedes")
	}
	if newRec.Supersedes == "" {
		newRec.Supersedes = oldRec.ID
	}
	oldRec.Status = "superseded"
	oldRec.SupersededBy = newRec.ID

	if opts.dryRun {
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "would update old: %s -> status=superseded, superseded_by=%s\n", oldRec.ID, newRec.ID)
		fmt.Fprintf(out, "would write new: %s -> supersedes=%s\n", newRec.ID, oldRec.ID)
		return nil
	}
	if _, err := s.Write(oldRec); err != nil {
		return fmt.Errorf("write old: %w", err)
	}
	if _, err := s.Write(newRec); err != nil {
		return fmt.Errorf("write new: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s\n", oldRec.ID, newRec.ID)
	return nil
}
