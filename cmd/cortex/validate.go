package cortex

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SincereMa/cortex-sidemark/internal/schema"
)

// validateOptions carries the flags for the `validate` command.
type validateOptions struct {
	jsonOutput bool
}

// newValidateCmd builds the `cortex validate` subcommand. It
// reports each file's outcome on its own line (or in a JSON array
// with --json) and exits non-zero if any file fails.
func newValidateCmd() *cobra.Command {
	opts := &validateOptions{}
	cmd := &cobra.Command{
		Use:   "validate <file> [<file>...]",
		Short: "Validate one or more record files against the record schema",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd, args, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "emit JSON error output")
	return cmd
}

// validateResult is the per-file outcome reported by `validate`.
// It is also the shape of each entry in the --json output array.
type validateResult struct {
	File  string `json:"file"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// runValidate reads each file, validates it against the record
// schema, and emits a result. It always reports every file's
// outcome before returning, so a single invocation can be used as
// a CI gate on a directory of files.
func runValidate(cmd *cobra.Command, args []string, opts *validateOptions) error {
	results := make([]validateResult, 0, len(args))
	exitCode := 0
	for _, path := range args {
		data, err := os.ReadFile(path)
		if err != nil {
			results = append(results, validateResult{File: path, Error: err.Error()})
			exitCode = 1
			continue
		}
		if err := schema.ValidateRecord(data); err != nil {
			results = append(results, validateResult{File: path, Error: err.Error()})
			exitCode = 1
			continue
		}
		results = append(results, validateResult{File: path, OK: true})
	}

	if opts.jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
	} else {
		for _, r := range results {
			if r.OK {
				fmt.Fprintf(cmd.OutOrStdout(), "ok    %s\n", r.File)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "fail  %s: %s\n", r.File, r.Error)
			}
		}
	}
	if exitCode != 0 {
		return fmt.Errorf("validation failed")
	}
	return nil
}
