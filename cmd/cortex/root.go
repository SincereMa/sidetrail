package cortex

import (
	"github.com/spf13/cobra"

	"github.com/SincereMa/cortex-sidemark/internal/version"
)

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cortex",
		Short:   "Cortex SideMark records long-lived project context",
		Version: version.Version,
		Long: `cortex is the CLI for Cortex SideMark, a sidecar that records
long-lived context (decisions, constraints, signals, dependencies) and
makes it available to host agents before they act.

It is intentionally read-dominant: most calls ask a question; a few
write a record.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetVersionTemplate("cortex {{.Version}} (commit " + version.Commit + ")\n")
	return cmd
}
