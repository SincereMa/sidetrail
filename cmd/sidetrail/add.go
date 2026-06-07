package sidetrail

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// addOptions carries the flags for the `add` command.
type addOptions struct {
	root string
}

// newAddCmd builds the `sidetrail add` subcommand. It validates a
// record file against the schema and writes it to the project
// store. The new record's id and the absolute path of the
// written file are printed on stdout; the id is also printed on
// its own line so it is easy to capture in a shell.
func newAddCmd() *cobra.Command {
	opts := &addOptions{}
	cmd := &cobra.Command{
		Use:   "add <file>",
		Short: "Validate a record file and add it to the store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	return cmd
}

// runAdd loads the file, validates it, and writes the record to
// the store. The flag's --root value, when set, overrides the
// upward search.
func runAdd(cmd *cobra.Command, args []string, opts *addOptions) error {
	r, err := record.LoadFile(args[0])
	if err != nil {
		return err
	}
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	if existsInStore(s, r.ID) {
		return fmt.Errorf("record %q already exists in store", r.ID)
	}
	path, err := s.Write(r)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s\n", r.ID)
	fmt.Fprintf(out, "%s\n", path)
	return nil
}

// existsInStore reports whether the store already holds a record
// with the given id. It is used to make `add` an idempotency-
// guarded operation: a second add of the same id is an error
// rather than a silent overwrite.
func existsInStore(s *storage.Store, id string) bool {
	if id == "" {
		return false
	}
	_, err := s.Get(id)
	return err == nil
}
