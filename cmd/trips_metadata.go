package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsUpdateCmd = &cobra.Command{
	Use:   "update [trip-key]",
	Short: "Update trip title, dates, or privacy",
	Long: `Update trip title, dates, or privacy settings.

Examples:
  wanderlog trips update abc123xyz --title "New Title"
  wanderlog trips update abc123xyz --start 2024-06-01 --end 2024-06-15
  wanderlog trips update abc123xyz --privacy public`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateDateFlag(updateStartDate, "start")
		validateDateFlag(updateEndDate, "end")
		if tripsUpdateTitle == "" && updateStartDate == "" && updateEndDate == "" && updatePrivacy == "" {
			logger.Error("At least one of --title, --start, --end, or --privacy is required")
			os.Exit(1)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err := client.UpdateTrip(args[0], wanderlog.UpdateTripRequest{
			Title:     tripsUpdateTitle,
			StartDate: updateStartDate,
			EndDate:   updateEndDate,
			Privacy:   updatePrivacy,
		})
		if err != nil {
			logger.WithError(err).Error("Failed to update trip")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Updated trip %s", args[0]), map[string]string{"tripKey": args[0]})
	},
}

var tripsSectionsCmd = &cobra.Command{
	Use:   "sections [trip-key]",
	Short: "List trip sections",
	Long: `List all sections (days) of a trip with their IDs and dates.

Examples:
  wanderlog trips sections abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		sections, err := client.GetTripSections(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch sections")
			os.Exit(1)
		}
		ui.PrintJSON(sections)
	},
}

var tripsFlightsCmd = &cobra.Command{
	Use:   "flights [trip-key]",
	Short: "List flights attached to a trip",
	Long: `List all flights associated with a trip.

Examples:
  wanderlog trips flights abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		flights, err := client.GetTripFlights(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch trip flights")
			os.Exit(1)
		}
		ui.PrintJSON(flights)
	},
}

var tripsExportCmd = &cobra.Command{
	Use:   "export [trip-key]",
	Short: "Export a trip to Google Maps",
	Long: `Export a trip to Google Maps format.

Examples:
  wanderlog trips export abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		resp, err := client.ExportTrip(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to export trip")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var tripsUpdateTitle string

func init() {
	tripsCmd.AddCommand(tripsUpdateCmd, tripsSectionsCmd, tripsFlightsCmd, tripsExportCmd)

	tripsUpdateCmd.Flags().StringVarP(&tripsUpdateTitle, "title", "t", "", "Trip title")
	tripsUpdateCmd.Flags().StringVar(&updateStartDate, "start", "", "Start date (YYYY-MM-DD)")
	tripsUpdateCmd.Flags().StringVar(&updateEndDate, "end", "", "End date (YYYY-MM-DD)")
	tripsUpdateCmd.Flags().StringVar(&updatePrivacy, "privacy", "", "Trip privacy (public, private, unlisted)")

	for _, c := range []*cobra.Command{tripsUpdateCmd, tripsSectionsCmd, tripsFlightsCmd, tripsExportCmd} {
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
