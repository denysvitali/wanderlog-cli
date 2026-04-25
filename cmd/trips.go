package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsCmd = &cobra.Command{
	Use:   "trips",
	Short: "Manage your trips",
	Long:  `List, create, edit, and manage your Wanderlog trips.`,
}

var tripsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your trips",
	Long: `List all trips for the authenticated user.

Requires authentication via 'wanderlog login'.

Examples:
  wanderlog trips list
  wanderlog trips list --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		trips, err := client.GetUserTrips()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch trips")
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			ui.PrintJSON(trips)
		case "markdown", "md":
			tripsListMarkdown(trips)
		default:
			tripsListPretty(trips)
		}
	},
}

var tripsShowCmd = &cobra.Command{
	Use:   "show [trip-id]",
	Short: "Show a trip's details",
	Long: `Fetch and display trip information from Wanderlog.

The trip ID can be found in the Wanderlog URL:
https://wanderlog.com/view/TRIP_ID/trip-name

Examples:
  wanderlog trips show abc123xyz
  wanderlog trips show abc123xyz --format json
  wanderlog trips show abc123xyz --format markdown --details
  wanderlog trips show --file trips/trip1.json`,
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
			trip, err = wanderlog.LoadTripFromFile(fromFile)
			if err != nil {
				logger.WithError(err).Error("Failed to load trip from file")
				os.Exit(1)
			}
		} else {
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
		case "markdown", "md":
			ui.PrintTripMarkdown(trip, showDetails)
		default:
			ui.PrintTrip(trip, showDetails)
		}
	},
}

func init() {
	rootCmd.AddCommand(tripsCmd)
	tripsCmd.AddCommand(tripsListCmd, tripsShowCmd)

	// trips list flags
	tripsListCmd.Flags().StringVarP(&outputFormat, "format", "f", "pretty", "Output format (pretty, json, markdown)")
	tripsListCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsListCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")

	// trips show flags
	tripsShowCmd.Flags().StringVarP(&outputFormat, "format", "f", "pretty", "Output format (pretty, json, markdown)")
	tripsShowCmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information")
	tripsShowCmd.Flags().StringVar(&fromFile, "file", "", "Load trip data from local JSON file instead of API")
}

func tripsListPretty(trips *wanderlog.UserTripsResponse) {
	if len(trips.Data) == 0 {
		fmt.Println("📭 No trips found")
		return
	}

	fmt.Printf("📚 Your Trips (%d total)\n\n", len(trips.Data))

	for _, trip := range trips.Data {
		privacy := "🌍"
		if trip.IsPrimary {
			privacy = "⭐"
		}

		fmt.Printf("%s %s\n", privacy, trip.Title)
		fmt.Printf("   Key: %s\n", trip.Key)

		if trip.StartDate != "" && trip.EndDate != "" {
			startDate, _ := time.Parse("2006-01-02", trip.StartDate)
			endDate, _ := time.Parse("2006-01-02", trip.EndDate)
			days := int(endDate.Sub(startDate).Hours()/24) + 1
			fmt.Printf("   📅 %s → %s (%d days)\n",
				startDate.Format("Jan 2, 2006"),
				endDate.Format("Jan 2, 2006"),
				days)
		}

		stats := []string{
			fmt.Sprintf("📍 %d places", trip.PlaceCount),
			fmt.Sprintf("👀 %d views", trip.ViewCount),
		}
		if trip.LikeCount > 0 {
			stats = append(stats, fmt.Sprintf("❤️ %d likes", trip.LikeCount))
		}

		fmt.Printf("   %s\n", strings.Join(stats, "  •  "))

		if trip.IsPrimary {
			fmt.Printf("   ⭐ Primary Trip\n")
		}
		if trip.IsDraft {
			fmt.Printf("   📝 Draft\n")
		}

		fmt.Println()
	}
}

func tripsListMarkdown(trips *wanderlog.UserTripsResponse) {
	fmt.Printf("# Your Trips\n\n")
	fmt.Printf("Total trips: %d\n\n", len(trips.Data))

	for _, trip := range trips.Data {
		fmt.Printf("## %s\n\n", trip.Title)

		fmt.Printf("- **Trip Key:** %s\n", trip.Key)
		fmt.Printf("- **Type:** %s\n", trip.Type)

		if trip.StartDate != "" && trip.EndDate != "" {
			startDate, _ := time.Parse("2006-01-02", trip.StartDate)
			endDate, _ := time.Parse("2006-01-02", trip.EndDate)
			days := int(endDate.Sub(startDate).Hours()/24) + 1
			fmt.Printf("- **Dates:** %s to %s (%d days)\n",
				startDate.Format("January 2, 2006"),
				endDate.Format("January 2, 2006"),
				days)
		}

		fmt.Printf("- **Places:** %d\n", trip.PlaceCount)
		fmt.Printf("- **Views:** %d\n", trip.ViewCount)
		if trip.LikeCount > 0 {
			fmt.Printf("- **Likes:** %d\n", trip.LikeCount)
		}

		if trip.IsPrimary {
			fmt.Printf("- **Status:** Primary Trip ⭐\n")
		}

		editedAt, _ := time.Parse(time.RFC3339, trip.EditedAt)
		fmt.Printf("- **Last Edited:** %s\n", editedAt.Format("January 2, 2006"))

		fmt.Println()
	}
}
