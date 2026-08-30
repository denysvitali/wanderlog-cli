package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build metadata is overridden through linker flags in every supported build
// path (Make, Docker, and GoReleaser).
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Print build version information",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "wanderlog %s\ncommit: %s\nbuilt: %s\n", Version, Commit, BuildDate)
			return err
		},
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.AddCommand(newVersionCmd())
}
