// Package sidetrail implements the sidetrail CLI.
package sidetrail

import (
	"github.com/spf13/cobra"

	"github.com/SincereMa/sidetrail/internal/version"
)

// Execute runs the root command and returns any error to the
// caller. main.go handles exit codes; this package stays
// testable.
func Execute() error {
	return newRootCmd().Execute()
}

// newRootCmd builds the root cobra command. It is the single
// place where the CLI's identity, flags, and subcommands are
// wired together.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sidetrail",
		Short:   "SideTrail records long-lived project context",
		Version: version.Version,
		Long: `sidetrail is the CLI for SideTrail, a sidecar that records
long-lived context (decisions, constraints, signals, dependencies) and
makes it available to host agents before they act.

Commands:
  context  — read records relevant to a file
  add      — validate and store a record
  update   — update an existing record
  health   — report project health signals
  init     — create a .sidetrail/ directory
  seed     — seed records from documents or apply agent output`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetVersionTemplate("sidetrail {{.Version}} (commit " + version.Commit + ")\n")
	cmd.AddCommand(
		newAddCmd(),
		newContextCmd(),
		newUpdateCmd(),
		newHealthCmd(),
		newInitCmd(),
		newSeedCmd(),
	)
	return cmd
}
