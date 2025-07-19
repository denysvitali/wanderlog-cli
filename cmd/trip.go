package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	outputFormat string
	showDetails  bool
	fromFile     string
)

var tripCmd = &cobra.Command{
	Use:   "trip [trip-id]",
	Short: "Get trip information",
	Long: `Fetch and display trip information from Wanderlog.

The trip ID can be found in the Wanderlog URL:
https://wanderlog.com/view/TRIP_ID/trip-name

Examples:
  wanderlog trip abc123xyz
  wanderlog trip abc123xyz --format json
  wanderlog trip abc123xyz --details
  wanderlog trip --file trips/trip1.json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if fromFile != "" && len(args) > 0 {
			return fmt.Errorf("cannot specify both trip ID and --file flag")
		}
		if fromFile == "" && len(args) != 1 {
			return fmt.Errorf("requires exactly one trip ID argument when not using --file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var trip *wanderlog.TripResponse
		var err error

		if fromFile != "" {
			// Load from file
			trip, err = wanderlog.LoadTripFromFile(fromFile)
			if err != nil {
				logger.WithError(err).Error("Failed to load trip from file")
				os.Exit(1)
			}
		} else {
			// Load from API
			tripID := args[0]
			client := wanderlog.NewClient()
			client.SetLogger(logger)

			trip, err = client.GetTrip(tripID)
			if err != nil {
				logger.WithError(err).Error("Failed to fetch trip")
				os.Exit(1)
			}
		}

		switch outputFormat {
		case "json":
			ui.PrintJSON(trip)
		default:
			ui.PrintTrip(trip, showDetails)
		}
	},
}

func init() {
	rootCmd.AddCommand(tripCmd)

	tripCmd.Flags().StringVarP(&outputFormat, "format", "f", "pretty", "Output format (pretty, json)")
	tripCmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information")
	tripCmd.Flags().StringVar(&fromFile, "file", "", "Load trip data from local JSON file instead of API")
}
