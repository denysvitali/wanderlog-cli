package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
  wanderlog verify-trip abc123xyz --format json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tripID := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		trip, err := client.GetTrip(tripID)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch trip")
			os.Exit(1)
		}

		// Output in text format (human readable)
		fmt.Println("=== TEXT FORMAT ===")
		fmt.Println()
		ui.PrintTrip(trip, true)

		// Output in JSON format
		fmt.Println()
		fmt.Println("=== JSON FORMAT ===")
		fmt.Println()
		jsonBytes, err := json.MarshalIndent(trip, "", "  ")
		if err != nil {
			logger.WithError(err).Error("Failed to marshal trip to JSON")
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	},
}

func init() {
	rootCmd.AddCommand(verifyTripCmd)

	verifyTripCmd.Flags().StringVarP(&verifyOutputFormat, "format", "f", "both",
		"Output format: 'text', 'json', or 'both' (default: both)")
}