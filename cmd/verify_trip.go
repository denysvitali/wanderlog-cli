package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	verifyOutputFormat string
)

var verifyTripCmd = &cobra.Command{
	Use:   "verify-trip [trip-id]",
	Short: "Verify and display trip information for debugging",
	Long: `Fetch a trip and display its data in both human-readable and JSON formats.
This command is useful for debugging trip data issues.

The trip ID can be found in the Wanderlog URL:
https://wanderlog.com/view/TRIP_ID/trip-name

Examples:
  wanderlog verify-trip abc123xyz
  wanderlog verify-trip abc123xyz --output json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tripID := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		trip, err := client.GetTripContext(cmd.Context(), tripID)
		if err != nil {
			return fmt.Errorf("fetch trip: %w", err)
		}

		// Output in text format (human readable)
		fmt.Println(ui.HeaderStyle.Render("=== TEXT FORMAT ==="))
		fmt.Println()
		ui.PrintTrip(trip, true)

		// Output in JSON format
		fmt.Println()
		fmt.Println(ui.HeaderStyle.Render("=== JSON FORMAT ==="))
		fmt.Println()
		jsonBytes, err := json.MarshalIndent(trip, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal trip to JSON: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(verifyTripCmd)

	verifyTripCmd.Flags().StringVarP(&verifyOutputFormat, "output", "o", "both",
		"Output format: 'text', 'json', or 'both' (default: both)")
}
