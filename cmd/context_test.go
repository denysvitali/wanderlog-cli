package cmd

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestExecuteContextPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	testCommand := &cobra.Command{
		Use:    "context-propagation-test",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Context().Err()
		},
	}
	rootCmd.AddCommand(testCommand)
	t.Cleanup(func() {
		rootCmd.RemoveCommand(testCommand)
		rootCmd.SetArgs(nil)
		rootCmd.SetContext(context.Background())
	})
	rootCmd.SetArgs([]string{testCommand.Name()})

	require.ErrorIs(t, ExecuteContext(ctx), context.Canceled)
}

func TestExecuteContextRejectsNilContext(t *testing.T) {
	//nolint:staticcheck // Explicitly verifies the defensive nil-context contract.
	require.EqualError(t, ExecuteContext(nil), "execute: nil context")
}
