package sidetrail

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// askOptions carries the flags for the `ask` command.
type askOptions struct {
	root  string
	scope string
	kind  string
	tag   string
	limit int
	jsonO bool
}

// newAskCmd builds the `sidetrail ask` subcommand. It is a
// structured query, not natural-language Q&A. The host agent
// (or its human) supplies a scope pattern; the sidecar returns
// the records that match, newest first, with optional kind and
// tag filters and a hard cap on the result size.
func newAskCmd() *cobra.Command {
	opts := &askOptions{limit: 50}
	cmd := &cobra.Command{
		Use:   "ask --scope <pattern> [--kind K] [--tag T] [--limit N] [--json]",
		Short: "Query records whose scope matches a pattern",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAsk(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.scope, "scope", "", "scope pattern to match (required)")
	cmd.Flags().StringVar(&opts.kind, "kind", "", "filter to one kind (decision, constraint, signal, experiment, incident)")
	cmd.Flags().StringVar(&opts.tag, "tag", "", "filter to records carrying this tag")
	cmd.Flags().IntVar(&opts.limit, "limit", 50, "maximum number of records to return")
	cmd.Flags().BoolVar(&opts.jsonO, "json", false, "emit JSON array of records instead of a table")
	_ = cmd.MarkFlagRequired("scope")
	return cmd
}

// runAsk resolves the store, runs the query, and emits results.
func runAsk(cmd *cobra.Command, _ []string, opts *askOptions) error {
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	recs, err := s.Ask(opts.scope, opts.kind, opts.tag, opts.limit)
	if err != nil {
		return err
	}
	if opts.jsonO {
		return writeRecordsJSON(cmd, recs)
	}
	writeRecordsTable(cmd, recs)
	return nil
}

// writeRecordsJSON writes recs as a JSON array to the command's
// stdout. It is shared by ask, list, and context so the agent
// contract is uniform.
func writeRecordsJSON(cmd *cobra.Command, recs []*record.Record) error {
	data, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = cmd.OutOrStdout().Write(append(data, '\n'))
	return err
}
