package cortex

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SincereMa/cortex-sidemark/internal/record"
	"github.com/SincereMa/cortex-sidemark/internal/storage"
)

// listOptions carries the flags for the `list` command.
type listOptions struct {
	root  string
	kind  string
	limit int
	jsonO bool
}

// newListCmd builds the `cortex list` subcommand. It enumerates
// records under the .cortex/ store, optionally filtered by
// --kind, capped at --limit (default 50), and printed as JSON
// (--json) or as a human-readable table (default).
func newListCmd() *cobra.Command {
	opts := &listOptions{limit: 50}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List records in the store, newest first",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .cortex/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.kind, "kind", "", "filter to one kind (decision, constraint, signal, experiment, incident)")
	cmd.Flags().IntVar(&opts.limit, "limit", 50, "maximum number of records to return")
	cmd.Flags().BoolVar(&opts.jsonO, "json", false, "emit JSON array of records instead of a table")
	return cmd
}

// runList reads the records, applies --kind and --limit, and
// emits them. An unknown --kind is reported as a validation
// error, not silently treated as "all".
func runList(cmd *cobra.Command, _ []string, opts *listOptions) error {
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	records, err := fetchRecords(s, opts.kind)
	if err != nil {
		return err
	}
	if opts.limit > 0 && len(records) > opts.limit {
		records = records[:opts.limit]
	}
	if opts.jsonO {
		return emitJSON(cmd, records)
	}
	emitTable(cmd, records)
	return nil
}

// fetchRecords returns the records for the requested kind, or
// all records when kind is empty. A non-empty unknown kind is
// reported as a validation error.
func fetchRecords(s *storage.Store, kind string) ([]*record.Record, error) {
	if kind == "" {
		return s.ListAll()
	}
	k := record.Kind(kind)
	if !k.Valid() {
		return nil, fmt.Errorf("unknown kind %q (want decision, constraint, signal, experiment, incident)", kind)
	}
	return s.ListKind(k)
}

// emitJSON writes the records as a JSON array.
func emitJSON(cmd *cobra.Command, recs []*record.Record) error {
	data, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = cmd.OutOrStdout().Write(append(data, '\n'))
	return err
}

// emitTable writes the records as a tab-separated table with a
// header. Width is not adjusted; tab characters are the
// alignment mechanism so agents can parse the columns
// positionally.
func emitTable(cmd *cobra.Command, recs []*record.Record) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "ID\tKIND\tSUBJECT\tCREATED_AT")
	for _, r := range recs {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", r.ID, r.Kind, r.Subject, r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}
}
