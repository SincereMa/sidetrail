package cortex

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SincereMa/cortex-sidemark/internal/storage"
)

// getOptions carries the flags for the `get` command.
type getOptions struct {
	root  string
	human bool
}

// newGetCmd builds the `cortex get` subcommand. It looks up a
// record by id (exact match, then prefix match) and writes it to
// stdout. The default output is the record's raw JSON, suitable
// for piping to jq or another tool. With --human, a short
// summary line is printed instead.
func newGetCmd() *cobra.Command {
	opts := &getOptions{}
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Fetch a record by id (exact or prefix match)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .cortex/ directory (default: search upward from CWD)")
	cmd.Flags().BoolVar(&opts.human, "human", false, "print a one-line human summary instead of JSON")
	return cmd
}

// runGet resolves the store, looks up the record, and emits it.
// Lookup errors (no match, ambiguity) are returned to the
// caller; the main entry point turns them into a non-zero exit
// code.
func runGet(cmd *cobra.Command, args []string, opts *getOptions) error {
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	r, err := s.Get(args[0])
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if opts.human {
		fmt.Fprintf(out, "%s\t%s\t%s\n", r.ID, r.Kind, r.Subject)
		return nil
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if _, err := out.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if _, err := out.Write([]byte("\n")); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}
	return nil
}
