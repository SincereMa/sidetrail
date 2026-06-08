// Package sidetrail implements the sidetrail CLI.
package sidetrail

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// draftOptions carries the flags for the `draft` command.
type draftOptions struct {
	root    string
	subject string
	scope   string
	author  string
}

// newDraftCmd builds the `sidetrail draft` subcommand. It
// creates a draft record file under the .sidetrail/_draft/
// directory. Drafts are complete, schema-valid records that
// are not yet promoted to the main store. They are reviewed
// by humans and promoted with `sidetrail promote`.
func newDraftCmd() *cobra.Command {
	opts := &draftOptions{}
	cmd := &cobra.Command{
		Use:   "draft <kind>",
		Short: "Create a draft record for review",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDraft(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().StringVar(&opts.subject, "subject", "", "short description of the record (required)")
	cmd.Flags().StringVar(&opts.scope, "scope", "", "scope this record applies to (e.g. file path or directory)")
	cmd.Flags().StringVar(&opts.author, "author", "", "author of the record (default: environment user)")
	_ = cmd.MarkFlagRequired("subject")
	return cmd
}

// runDraft creates a draft record in _draft/ directory. The
// draft is a complete, schema-valid record file that can be
// reviewed and promoted to the main store.
func runDraft(cmd *cobra.Command, args []string, opts *draftOptions) error {
	k := record.Kind(args[0])
	if !k.Valid() {
		return fmt.Errorf("unknown kind %q (want decision, constraint, signal, experiment, incident)", args[0])
	}
	id, err := record.NewID()
	if err != nil {
		return err
	}
	author := opts.author
	if author == "" {
		author = defaultAuthor()
	}
	now := time.Now().UTC()
	r := &record.Record{
		ID:             id,
		Kind:           k,
		Scope:          opts.scope,
		Subject:        opts.subject,
		Reason:         "Draft — review and complete before promoting.",
		SourceType:     record.SourceHuman,
		Author:         author,
		CreatedAt:      now,
		LastVerifiedAt: now,
		Status:         "active",
	}
	if k == record.KindDecision {
		r.DecidedAt = &now
	}
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	path, err := s.WriteDraft(r)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s\n", r.ID)
	fmt.Fprintf(out, "%s\n", path)
	return nil
}

// defaultAuthor returns the current OS username, falling back
// to "unknown" if the environment does not provide one.
func defaultAuthor() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Base(home)
	}
	return "unknown"
}
