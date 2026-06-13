package sidetrail

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// updateOptions carries the flags for the `update` command.
type updateOptions struct {
	root string
	file string
}

// newUpdateCmd builds the `sidetrail update` subcommand. It
// updates an existing record with partial JSON fields. The
// agent reads the current record, merges the provided fields,
// and writes it back.
func newUpdateCmd() *cobra.Command {
	opts := &updateOptions{}
	cmd := &cobra.Command{
		Use:   "update <id> --file <json-file>",
		Short: "Update an existing record with partial JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.file, "file", "", "JSON file with fields to update (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

// runUpdate reads the existing record, merges the provided JSON
// fields, and writes it back.
func runUpdate(cmd *cobra.Command, args []string, opts *updateOptions) error {
	id := args[0]
	if opts.file == "" {
		return fmt.Errorf("--file is required")
	}

	data, err := os.ReadFile(opts.file)
	if err != nil {
		return fmt.Errorf("read update file: %w", err)
	}

	var updates map[string]interface{}
	if err := json.Unmarshal(data, &updates); err != nil {
		return fmt.Errorf("parse update JSON: %w", err)
	}

	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)

	r, err := s.Get(id)
	if err != nil {
		return err
	}

	existingJSON, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal existing record: %w", err)
	}

	var existing map[string]interface{}
	if err := json.Unmarshal(existingJSON, &existing); err != nil {
		return fmt.Errorf("unmarshal existing record: %w", err)
	}

	for k, v := range updates {
		existing[k] = v
	}

	mergedJSON, err := json.Marshal(existing)
	if err != nil {
		return fmt.Errorf("marshal merged record: %w", err)
	}

	var updated record.Record
	if err := json.Unmarshal(mergedJSON, &updated); err != nil {
		return fmt.Errorf("unmarshal merged record: %w", err)
	}

	if _, err := s.Write(&updated); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", id)
	return nil
}
