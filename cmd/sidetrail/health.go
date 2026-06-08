package sidetrail

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// healthOptions carries the flags for the `health` command.
type healthOptions struct {
	root      string
	jsonO     bool
	staleDays int
}

// newHealthCmd builds the `sidetrail health` subcommand. It
// scans the store and reports project health signals:
//
//   - Record counts by kind and status
//   - Stale records (last_verified_at older than --stale-days)
//   - Active supersede chains
//   - Scope coverage
//
// The output is a human-readable summary suitable for agents
// to pull before acting. With --json, a structured object is
// emitted.
func newHealthCmd() *cobra.Command {
	opts := &healthOptions{staleDays: 90}
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Report project health signals",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHealth(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "explicit path to a .sidetrail/ directory (default: search upward from CWD)")
	cmd.Flags().BoolVar(&opts.jsonO, "json", false, "emit structured JSON instead of a table")
	cmd.Flags().IntVar(&opts.staleDays, "stale-days", 90, "records unverified longer than this are flagged stale")
	return cmd
}

// healthReport is the structured output of the health command.
type healthReport struct {
	Total       int               `json:"total"`
	ByKind      map[string]int    `json:"by_kind"`
	ByStatus    map[string]int    `json:"by_status"`
	Stale       []*record.Record  `json:"stale"`
	ActiveChain int               `json:"active_chains"`
	ScopeCount  int               `json:"scope_count"`
	Oldest      string            `json:"oldest,omitempty"`
	Newest      string            `json:"newest,omitempty"`
}

// runHealth scans the store and emits the health report.
func runHealth(cmd *cobra.Command, _ []string, opts *healthOptions) error {
	root, err := resolveStoreRoot(opts.root)
	if err != nil {
		return err
	}
	s := storage.NewStore(root)
	all, err := s.ListAll()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no records found")
		return nil
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -opts.staleDays)
	rpt := &healthReport{
		Total:    len(all),
		ByKind:   make(map[string]int),
		ByStatus: make(map[string]int),
	}
	scopes := make(map[string]struct{})
	for _, r := range all {
		rpt.ByKind[string(r.Kind)]++
		rpt.ByStatus[r.Status]++
		if r.Scope != "" {
			scopes[r.Scope] = struct{}{}
		}
		if r.Status == "active" && r.Supersedes != "" {
			rpt.ActiveChain++
		}
		if r.LastVerifiedAt.Before(cutoff) && (r.Status == "active" || r.Status == "in_progress" || r.Status == "investigating") {
			rpt.Stale = append(rpt.Stale, r)
		}
	}
	rpt.ScopeCount = len(scopes)
	if len(all) > 0 {
		rpt.Oldest = all[len(all)-1].CreatedAt.Format("2006-01-02")
		rpt.Newest = all[0].CreatedAt.Format("2006-01-02")
	}
	if opts.jsonO {
		return writeHealthJSON(cmd, rpt)
	}
	writeHealthTable(cmd, rpt, opts.staleDays)
	return nil
}

// writeHealthJSON emits the report as JSON.
func writeHealthJSON(cmd *cobra.Command, rpt *healthReport) error {
	data, err := json.MarshalIndent(rpt, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return err
}

// writeHealthTable emits the report as a human-readable table.
func writeHealthTable(cmd *cobra.Command, rpt *healthReport, staleDays int) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Total records:    %d\n", rpt.Total)
	fmt.Fprintf(out, "Active chains:    %d\n", rpt.ActiveChain)
	fmt.Fprintf(out, "Unique scopes:    %d\n", rpt.ScopeCount)
	fmt.Fprintf(out, "Date range:       %s to %s\n", rpt.Oldest, rpt.Newest)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "By kind:")
	for _, k := range []string{"decision", "constraint", "signal", "experiment", "incident"} {
		if n := rpt.ByKind[k]; n > 0 {
			fmt.Fprintf(out, "  %-14s %d\n", k, n)
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "By status:")
	for _, s := range []string{"active", "superseded", "archived", "hidden", "in_progress", "succeeded", "failed", "inconclusive", "abandoned", "investigating", "mitigated", "resolved"} {
		if n := rpt.ByStatus[s]; n > 0 {
			fmt.Fprintf(out, "  %-14s %d\n", s, n)
		}
	}
	if len(rpt.Stale) > 0 {
		fmt.Fprintf(out, "\nStale records (unverified >%d days):\n", staleDays)
		for _, r := range rpt.Stale {
			days := int(time.Since(r.LastVerifiedAt).Hours() / 24)
			fmt.Fprintf(out, "  %s  %s  %s  (%d days)\n", r.ID, r.Kind, r.Subject, days)
		}
	}
}
