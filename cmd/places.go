package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var placesCmd = &cobra.Command{
	Use:   "places [trip-id]",
	Short: "Show places from a trip",
	Long: `Display detailed information about places in a trip including
names, addresses, ratings, and other metadata.

Examples:
  wanderlog places abc123xyz
  wanderlog places --file trips/trip1.json
  wanderlog places abc123xyz --format json
  wanderlog places abc123xyz --format markdown`,
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
			ui.PrintJSON(trip.Resources.PlaceMetadata)
		case "markdown", "md":
			ui.PrintPlacesMarkdown(trip.Resources.PlaceMetadata)
		default:
			ui.PrintPlaces(trip.Resources.PlaceMetadata)
		}
	},
}

func init() {
	// root registration disabled - command moved under `trips places`
	// rootCmd.AddCommand(placesCmd)

	placesCmd.Flags().StringVarP(&outputFormat, "format", "f", "pretty", "Output format (pretty, json, markdown)")
	placesCmd.Flags().StringVar(&fromFile, "file", "", "Load trip data from local JSON file instead of API")
}
