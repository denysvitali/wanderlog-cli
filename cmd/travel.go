package cmd

import (
	"fmt"
	"time"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetAllAirlines()
		if err != nil {
			return fmt.Errorf("list airlines: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var travelAirportsCmd = &cobra.Command{
	Use:   "airports [query]",
	Short: "Search airports",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if (travelLat == 0) != (travelLng == 0) {
			return fmt.Errorf("--lat and --lng must be provided together")
		}
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
			return fmt.Errorf("search airports: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var travelFlightStopsCmd = &cobra.Command{
	Use:   "flight-stops [flight-number]",
	Short: "Show stops for a flight number",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flightNum := args[0]
		airline := flightStopsAirline
		date := flightStopsDate

		if airline == "" {
			return fmt.Errorf("--airline is required (for example, UA, BA, or LH)")
		}
		if date == "" {
			return fmt.Errorf("--date is required (YYYY-MM-DD)")
		}
		if err := validateDateFlagE(date, "departure"); err != nil {
			return err
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetFlightStops(flightNum, airline, date)
		if err != nil {
			return fmt.Errorf("get flight stops: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var travelHotelsCmd = &cobra.Command{
	Use:   "hotels [query]",
	Short: "Search hotels/lodging",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDateFlagE(travelHotelCheckIn, "check-in"); err != nil {
			return err
		}
		if err := validateDateFlagE(travelHotelCheckOut, "check-out"); err != nil {
			return err
		}
		if travelHotelGuests < 1 {
			return fmt.Errorf("--guests must be at least 1")
		}
		if travelHotelCheckIn != "" && travelHotelCheckOut != "" {
			checkIn, _ := time.Parse("2006-01-02", travelHotelCheckIn)
			checkOut, _ := time.Parse("2006-01-02", travelHotelCheckOut)
			if !checkOut.After(checkIn) {
				return fmt.Errorf("--check-out must be after --check-in")
			}
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.SearchLodgings(args[0], travelHotelCheckIn, travelHotelCheckOut, travelHotelGuests)
		if err != nil {
			return fmt.Errorf("search hotels: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var travelHotelRatesCmd = &cobra.Command{
	Use:   "hotel-rates [property-id]",
	Short: "Get Google lodging price rates",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetGooglePriceRates(args[0])
		if err != nil {
			return fmt.Errorf("get hotel rates: %w", err)
		}
		return ui.PrintJSON(resp)
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
	travelFlightStopsCmd.Flags().StringVar(&flightStopsAirline, "airline", "", "Airline IATA code (e.g., UA, BA)")
	travelFlightStopsCmd.Flags().StringVar(&flightStopsDate, "date", "", "Departure date (YYYY-MM-DD)")

	for _, c := range []*cobra.Command{travelAirlinesCmd, travelAirportsCmd, travelFlightStopsCmd, travelHotelsCmd, travelHotelRatesCmd} {
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
