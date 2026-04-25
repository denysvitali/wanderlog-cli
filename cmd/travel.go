package cmd

import (
	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var newTravelCmd = &cobra.Command{
	Use:   "travel",
	Short: "Search flights and lodging helpers",
}

var travelAirlinesCmd = &cobra.Command{
	Use:   "airlines",
	Short: "List all airlines",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetAllAirlines()
		if err != nil {
			logger.WithError(err).Error("Failed to list airlines")
			return
		}
		ui.PrintJSON(resp)
	},
}

var travelAirportsCmd = &cobra.Command{
	Use:   "airports [query]",
	Short: "Search airports",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		var resp interface{}
		var err error
		if travelLat != 0 || travelLng != 0 {
			resp, err = client.AutocompleteAirportWithLocation(args[0], travelLat, travelLng)
		} else {
			resp, err = client.AutocompleteAirport(args[0])
		}
		if err != nil {
			logger.WithError(err).Error("Failed to search airports")
			return
		}
		ui.PrintJSON(resp)
	},
}

var travelFlightStopsCmd = &cobra.Command{
	Use:   "flight-stops [flight-number]",
	Short: "Show stops for a flight number",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetFlightStops(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to get flight stops")
			return
		}
		ui.PrintJSON(resp)
	},
}

var travelHotelsCmd = &cobra.Command{
	Use:   "hotels [query]",
	Short: "Search hotels/lodging",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateDateFlag(travelHotelCheckIn, "check-in")
		validateDateFlag(travelHotelCheckOut, "check-out")

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.SearchLodgings(args[0], travelHotelCheckIn, travelHotelCheckOut, travelHotelGuests)
		if err != nil {
			logger.WithError(err).Error("Failed to search hotels")
			return
		}
		ui.PrintJSON(resp)
	},
}

var travelHotelRatesCmd = &cobra.Command{
	Use:   "hotel-rates [property-id]",
	Short: "Get Google lodging price rates",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetGooglePriceRates(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to get hotel rates")
			return
		}
		ui.PrintJSON(resp)
	},
}

var (
	travelLat           float64
	travelLng           float64
	travelHotelCheckIn  string
	travelHotelCheckOut string
	travelHotelGuests   int
)

func init() {
	rootCmd.AddCommand(newTravelCmd)
	newTravelCmd.AddCommand(travelAirlinesCmd, travelAirportsCmd, travelFlightStopsCmd, travelHotelsCmd, travelHotelRatesCmd)

	travelAirportsCmd.Flags().Float64Var(&travelLat, "lat", 0, "Latitude for location bias")
	travelAirportsCmd.Flags().Float64Var(&travelLng, "lng", 0, "Longitude for location bias")
	travelHotelsCmd.Flags().StringVar(&travelHotelCheckIn, "check-in", "", "Check-in date (YYYY-MM-DD)")
	travelHotelsCmd.Flags().StringVar(&travelHotelCheckOut, "check-out", "", "Check-out date (YYYY-MM-DD)")
	travelHotelsCmd.Flags().IntVar(&travelHotelGuests, "guests", 1, "Number of guests")

	for _, c := range []*cobra.Command{travelAirlinesCmd, travelAirportsCmd, travelFlightStopsCmd, travelHotelsCmd, travelHotelRatesCmd} {
		c.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
