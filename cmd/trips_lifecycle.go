package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	tripsCreateTitle   string
	tripsCreateStart   string
	tripsCreateEnd     string
	tripsCreatePrivacy string
	tripsCreateGeoIDs  []int
	tripsCreateExample bool
	tripsDeleteYes     bool
)

var tripsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new trip",
	Long: `Create a new trip plan on Wanderlog.

Requires authentication via 'wanderlog login' or environment variables.

Examples:
  wanderlog trips create --title "Trip to Japan" --geo-id 1
  wanderlog trips create --title "Europe 2024" --geo-id 7 --start 2024-06-01 --end 2024-06-15
  wanderlog trips create --title "Private Trip" --geo-id 1 --privacy private
  wanderlog trips create --example`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tripsCreateTitle == "" && !tripsCreateExample {
			return fmt.Errorf("--title is required")
		}
		if len(tripsCreateGeoIDs) == 0 && !tripsCreateExample {
			return fmt.Errorf("at least one --geo-id is required")
		}

		if err := validateDateFlagE(tripsCreateStart, "start"); err != nil {
			return err
		}
		if err := validateDateFlagE(tripsCreateEnd, "end"); err != nil {
			return err
		}
		if tripsCreateStart != "" && tripsCreateEnd != "" {
			start, _ := time.Parse("2006-01-02", tripsCreateStart)
			end, _ := time.Parse("2006-01-02", tripsCreateEnd)
			if end.Before(start) {
				return fmt.Errorf("--end must not be before --start")
			}
		}
		privacy := strings.ToLower(strings.TrimSpace(tripsCreatePrivacy))
		if privacy != "private" && privacy != "public" && privacy != "friends" {
			return fmt.Errorf("invalid privacy %q (valid values: private, public, friends)", tripsCreatePrivacy)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		var resp *wanderlog.CreateTripResponse
		var err error
		if tripsCreateExample {
			resp, err = client.CreateExampleTrip()
		} else {
			req := wanderlog.CreateTripRequest{
				Title:               tripsCreateTitle,
				GeoIDs:              tripsCreateGeoIDs,
				InitialMapsPlaceIDs: []int{},
				Type:                "plan",
				StartDate:           tripsCreateStart,
				EndDate:             tripsCreateEnd,
				Privacy:             privacy,
				IsMapEmbed:          false,
				Language:            "en",
			}
			resp, err = client.CreateTrip(req)
		}
		if err != nil {
			return fmt.Errorf("create trip: %w", err)
		}

		if outputFormat == "json" {
			return ui.PrintJSON(resp)
		}
		fmt.Println(ui.SuccessStyle.Render("🎉 Successfully created trip!"))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Title: %s", resp.TripPlan.Title)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Trip ID: %d", resp.TripPlan.ID)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Trip Key: %s", resp.TripPlan.Key)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Edit Key: %s", resp.TripPlan.EditKey)))
		fmt.Println(ui.UrlStyle.Render(fmt.Sprintf("URL: https://wanderlog.com/view/%s/%s", resp.TripPlan.Key, resp.TripPlan.Title)))
		return nil
	},
}

var tripsDeleteCmd = &cobra.Command{
	Use:   "delete [trip-key]",
	Short: "Delete a trip",
	Long: `Delete a trip plan from Wanderlog.

Requires authentication and the trip's edit key.

Examples:
  wanderlog trips delete abc123xyz

WARNING: This action cannot be undone!`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tripKey := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		confirmed, err := confirmAction(cmd,
			ui.WarningStyle.Render(fmt.Sprintf("WARNING: deleting trip %s cannot be undone.", tripKey)),
			tripsDeleteYes,
		)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println(ui.InfoStyle.Render("Trip deletion canceled."))
			return nil
		}

		if err := client.DeleteTrip(tripKey); err != nil {
			return fmt.Errorf("delete trip: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Deleted trip %s", tripKey), map[string]string{"tripKey": tripKey})
	},
}

var tripsCopyCmd = &cobra.Command{
	Use:   "copy [trip-key]",
	Short: "Copy an existing trip",
	Long: `Create a copy of an existing trip plan.

Examples:
  wanderlog trips copy abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceKey := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		resp, err := client.CopyTrip(sourceKey)
		if err != nil {
			return fmt.Errorf("copy trip: %w", err)
		}

		if outputFormat == "json" {
			return ui.PrintJSON(resp)
		}
		fmt.Println(ui.SuccessStyle.Render("📋 Successfully copied trip!"))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Original: %s", sourceKey)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("New Title: %s", resp.TripPlan.Title)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("New Key: %s", resp.TripPlan.Key)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Edit Key: %s", resp.TripPlan.EditKey)))
		fmt.Println(ui.UrlStyle.Render(fmt.Sprintf("URL: https://wanderlog.com/view/%s/%s", resp.TripPlan.Key, resp.TripPlan.Title)))
		return nil
	},
}

var tripsRestoreCmd = &cobra.Command{
	Use:   "restore [trip-key]",
	Short: "Restore a deleted trip",
	Long: `Restore a soft-deleted trip plan.

Examples:
  wanderlog trips restore abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		if err := client.RestoreTrip(args[0]); err != nil {
			return fmt.Errorf("restore trip: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Restored trip %s", args[0]), map[string]string{"tripKey": args[0]})
	},
}

func init() {
	tripsCmd.AddCommand(tripsCreateCmd, tripsDeleteCmd, tripsCopyCmd, tripsRestoreCmd)

	// create flags
	tripsCreateCmd.Flags().StringVarP(&tripsCreateTitle, "title", "t", "", "Trip title (required)")
	tripsCreateCmd.Flags().StringVar(&tripsCreateStart, "start", "", "Start date (YYYY-MM-DD)")
	tripsCreateCmd.Flags().StringVar(&tripsCreateEnd, "end", "", "End date (YYYY-MM-DD)")
	tripsCreateCmd.Flags().StringVar(&tripsCreatePrivacy, "privacy", "private", "Trip privacy (public, private, friends)")
	tripsCreateCmd.Flags().IntSliceVar(&tripsCreateGeoIDs, "geo-id", nil, "Wanderlog destination geo ID (repeatable)")
	tripsCreateCmd.Flags().BoolVar(&tripsCreateExample, "example", false, "Create Wanderlog's pre-filled example trip")
	tripsDeleteCmd.Flags().BoolVarP(&tripsDeleteYes, "yes", "y", false, "Skip the destructive-operation confirmation")

	// auth flags
	for _, c := range []*cobra.Command{tripsCreateCmd, tripsDeleteCmd, tripsCopyCmd, tripsRestoreCmd} {
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
