package cmd

import (
	"fmt"

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
  wanderlog trips update abc123xyz --title ""
  wanderlog trips update abc123xyz --start 2024-06-01 --end 2024-06-15
  wanderlog trips update abc123xyz --privacy public`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDateFlagE(updateStartDate, "start"); err != nil {
			return err
		}
		if err := validateDateFlagE(updateEndDate, "end"); err != nil {
			return err
		}
		if !cmd.Flags().Changed("title") && !cmd.Flags().Changed("start") && !cmd.Flags().Changed("end") && !cmd.Flags().Changed("privacy") {
			return fmt.Errorf("at least one of --title, --start, --end, or --privacy is required")
		}

		client, err := newClientE(true)
		if err != nil {
			return err
		}

		err = client.UpdateTrip(args[0], wanderlog.UpdateTripRequest{
			Title:      tripsUpdateTitle,
			ClearTitle: cmd.Flags().Changed("title") && tripsUpdateTitle == "",
			StartDate:  updateStartDate,
			EndDate:    updateEndDate,
			Privacy:    updatePrivacy,
		})
		if err != nil {
			return fmt.Errorf("update trip: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Updated trip %s", args[0]), map[string]string{"tripKey": args[0]})
	},
}

var tripsSectionsCmd = &cobra.Command{
	Use:   "sections [trip-key]",
	Short: "List trip sections",
	Long: `List all sections (days) of a trip with their IDs and dates.

Examples:
  wanderlog trips sections abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		sections, err := client.GetTripSectionsContext(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("fetch sections: %w", err)
		}
		return ui.PrintJSON(sections)
	},
}

var tripsFlightsCmd = &cobra.Command{
	Use:   "flights [trip-key]",
	Short: "List flights attached to a trip",
	Long: `List all flights associated with a trip.

Examples:
  wanderlog trips flights abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}

		flights, err := client.GetTripFlights(args[0])
		if err != nil {
			return fmt.Errorf("fetch trip flights: %w", err)
		}
		return ui.PrintJSON(flights)
	},
}

var tripsExportCmd = &cobra.Command{
	Use:   "export [trip-key]",
	Short: "Export a trip to Google Maps",
	Long: `Export a trip to Google Maps format.

Examples:
  wanderlog trips export abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}

		resp, err := client.ExportTrip(args[0])
		if err != nil {
			return fmt.Errorf("export trip: %w", err)
		}
		return ui.PrintJSON(resp)
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
