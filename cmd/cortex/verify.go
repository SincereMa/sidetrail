package cortex

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/SincereMa/cortex-sidemark/internal/storage"
)

// verifyOptions carries the flags for the `verify` command.
type verifyOptions struct {
	root string
}

// newVerifyCmd builds the `cortex verify` subcommand. It is
// the freshness-renewal primitive: a human (or an opt-in
// scrape) marks a record as still-true, and the sidecar stamps
// `last_verified_at` to the current time and writes the record
// back to disk. The on-disk path is unchanged because the slug
// is derived from the unchanged subject, so the rewrite
// happens in place.
func newVerifyCmd() *cobra.Command {
	opts := &verifyOptions{}
	cmd := &cobra.Command{
		Use:   "verify <id>",
		Short: "Refresh a record's last_verified_at to the current time",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .cortex/ directory (default: search upward from CWD)")
	return cmd
}

// runVerify resolves the store, looks up the record, stamps it,
// and writes it back. The new timestamp is printed on stdout
// at second precision; the on-disk record carries the same
// truncated value so the printed text and the stored JSON
// agree exactly.
func runVerify(cmd *cobra.Command, args []string, opts *verifyOptions) error {
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	r, err := s.Get(args[0])
	if err != nil {
		return err
	}
	now := time.Now().UTC().Truncate(time.Second)
	r.LastVerifiedAt = now
	if _, err := s.Write(r); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", now.Format(time.RFC3339))
	return nil
}
