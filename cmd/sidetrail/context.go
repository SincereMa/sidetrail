package sidetrail

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// contextOptions carries the flags for the `context` command.
type contextOptions struct {
	root   string
	file   string
	radius int
	limit  int
	jsonO  bool
}

// newContextCmd builds the `sidetrail context` subcommand. It is
// the file-scoped aggregate the host agent calls before
// editing: "what do I know about this file and the directories
// above it?" The host agent points at the file; the sidecar
// walks ancestors and returns matching records, newest first.
func newContextCmd() *cobra.Command {
	opts := &contextOptions{limit: 50}
	cmd := &cobra.Command{
		Use:   "context --file <path> [--radius N] [--limit N] [--json]",
		Short: "Aggregate records relevant to a file path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContext(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.file, "file", "", "file path whose context to gather (required)")
	cmd.Flags().IntVar(&opts.radius, "radius", 0, "ancestor levels to include (0 = walk to root, N>0 = walk N levels)")
	cmd.Flags().IntVar(&opts.limit, "limit", 50, "maximum number of records to return")
	cmd.Flags().BoolVar(&opts.jsonO, "json", false, "emit JSON array of records instead of a table")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

// runContext resolves the store, gathers context for the file,
// and emits results. The file is not required to exist on disk;
// the function is a retrieval over stored scopes.
func runContext(cmd *cobra.Command, _ []string, opts *contextOptions) error {
	if opts.file == "" {
		return fmt.Errorf("--file is required")
	}
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	recs, err := s.ContextFor(opts.file, opts.radius, opts.limit)
	if err != nil {
		return err
	}
	if opts.jsonO {
		return writeRecordsJSON(cmd, recs)
	}
	writeRecordsTable(cmd, recs)
	return nil
}

// writeRecordsJSON writes recs as a JSON array to the command's stdout.
func writeRecordsJSON(cmd *cobra.Command, recs []*record.Record) error {
	data, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = cmd.OutOrStdout().Write(append(data, '\n'))
	return err
}

// writeRecordsTable writes recs as a tab-separated table with a header.
func writeRecordsTable(cmd *cobra.Command, recs []*record.Record) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "ID\tKIND\tSUBJECT\tCREATED_AT")
	for _, r := range recs {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", r.ID, r.Kind, r.Subject, r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}
}
