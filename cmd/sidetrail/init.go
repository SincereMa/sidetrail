package sidetrail

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initOptions carries the flags for the `init` command.
type initOptions struct {
	root string
}

// newInitCmd builds the `sidetrail init` subcommand. It creates
// a .sidetrail/ directory at the project root. The store is
// usable from empty; init is optional.
func newInitCmd() *cobra.Command {
	opts := &initOptions{}
	cmd := &cobra.Command{
		Use:   "init [--root <project>]",
		Short: "Create a .sidetrail/ directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, opts)
		},
	}
	cmd.Flags().StringVar(&opts.root, "root", "", "project root where .sidetrail/ will be created (default: CWD)")
	return cmd
}

// runInit creates the .sidetrail/ directory.
func runInit(cmd *cobra.Command, _ []string, opts *initOptions) error {
	projectRoot, err := resolveProjectRoot(opts.root)
	if err != nil {
		return err
	}
	storeDir := filepath.Join(projectRoot, storeDirName)
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", storeDir, err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", storeDir)
	return nil
}

// resolveProjectRoot returns the absolute project root. When
// opts.root is empty, the current working directory is used.
func resolveProjectRoot(explicit string) (string, error) {
	if explicit == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
		return wd, nil
	}
	abs, err := filepath.Abs(explicit)
	if err != nil {
		return "", fmt.Errorf("abs %q: %w", explicit, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("--root %q is not a directory", abs)
	}
	return abs, nil
}
