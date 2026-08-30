package cmd

import (
	"fmt"
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
  wanderlog trips list --output json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		trips, err := client.GetUserTripsContext(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetch trips: %w", err)
		}

		switch outputFormat {
		case "json":
			return ui.PrintJSON(trips)
		case "markdown", "md":
			tripsListMarkdown(trips)
		default:
			tripsListPretty(trips)
		}
		return nil
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
  wanderlog trips show abc123xyz --output json
  wanderlog trips show abc123xyz --output markdown --details
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var trip *wanderlog.TripResponse
		var err error

		if fromFile != "" {
			trip, err = wanderlog.LoadTripFromFile(fromFile)
			if err != nil {
				return fmt.Errorf("load trip from file: %w", err)
			}
		} else {
			tripID := args[0]
			client := wanderlog.NewClient()
			client.SetLogger(logger)

			trip, err = client.GetTripContext(cmd.Context(), tripID)
			if err != nil {
				return fmt.Errorf("fetch trip: %w", err)
			}
		}

		switch outputFormat {
		case "json":
			return ui.PrintJSON(trip)
		case "markdown", "md":
			ui.PrintTripMarkdown(trip, showDetails)
		default:
			ui.PrintTrip(trip, showDetails)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tripsCmd)
	tripsCmd.AddCommand(tripsListCmd, tripsShowCmd)

	// trips list flags
	tripsListCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json, markdown)")
	tripsListCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsListCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")

	// trips show flags
	tripsShowCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json, markdown)")
	tripsShowCmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information")
	tripsShowCmd.Flags().StringVar(&fromFile, "file", "", "Load trip data from local JSON file instead of API")
}

func tripsListPretty(trips *wanderlog.UserTripsResponse) {
	if len(trips.Data) == 0 {
		fmt.Println(ui.WarningStyle.Render("📭 No trips found"))
		return
	}

	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("📚 Your Trips (%d total)", len(trips.Data))))
	fmt.Println()

	for _, trip := range trips.Data {
		privacy := "🌍"
		if trip.IsPrimary {
			privacy = "⭐"
		}

		fmt.Printf("%s %s\n", privacy, ui.PlaceStyle.Render(ui.SafeText(trip.Title)))
		fmt.Println(ui.IdStyle.Render(fmt.Sprintf("   Key: %s", ui.SafeText(trip.Key))))

		if trip.StartDate != "" && trip.EndDate != "" {
			startDate, _ := time.Parse("2006-01-02", trip.StartDate)
			endDate, _ := time.Parse("2006-01-02", trip.EndDate)
			days := int(endDate.Sub(startDate).Hours()/24) + 1
			fmt.Println(ui.DateStyle.Render(fmt.Sprintf("   📅 %s → %s (%d days)",
				startDate.Format("Jan 2, 2006"),
				endDate.Format("Jan 2, 2006"),
				days)))
		}

		stats := []string{
			ui.InfoStyle.Render(fmt.Sprintf("📍 %d places", trip.PlaceCount)),
			ui.InfoStyle.Render(fmt.Sprintf("👀 %d views", trip.ViewCount)),
		}
		if trip.LikeCount > 0 {
			stats = append(stats, ui.SuccessStyle.Render(fmt.Sprintf("❤️ %d likes", trip.LikeCount)))
		}

		fmt.Printf("   %s\n", strings.Join(stats, ui.SeparatorStyle.Render("  •  ")))

		if trip.IsPrimary {
			fmt.Println(ui.HighlightStyle.Render("   ⭐ Primary Trip"))
		}
		if trip.IsDraft {
			fmt.Println(ui.WarningStyle.Render("   📝 Draft"))
		}

		fmt.Println()
	}
}

func tripsListMarkdown(trips *wanderlog.UserTripsResponse) {
	fmt.Printf("# Your Trips\n\n")
	fmt.Printf("Total trips: %d\n\n", len(trips.Data))

	for _, trip := range trips.Data {
		fmt.Printf("## %s\n\n", ui.MarkdownInline(trip.Title))

		fmt.Printf("- **Trip Key:** %s\n", ui.MarkdownInline(trip.Key))
		fmt.Printf("- **Type:** %s\n", ui.MarkdownInline(trip.Type))

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
