package cmd

import (
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		if tripsCreateTitle == "" && !tripsCreateExample {
			logger.Error("Trip title is required")
			os.Exit(1)
		}
		if len(tripsCreateGeoIDs) == 0 && !tripsCreateExample {
			logger.Error("At least one --geo-id is required")
			os.Exit(1)
		}

		if tripsCreateStart != "" {
			if _, err := time.Parse("2006-01-02", tripsCreateStart); err != nil {
				logger.WithError(err).Error("Invalid start date format. Use YYYY-MM-DD")
				os.Exit(1)
			}
		}
		if tripsCreateEnd != "" {
			if _, err := time.Parse("2006-01-02", tripsCreateEnd); err != nil {
				logger.WithError(err).Error("Invalid end date format. Use YYYY-MM-DD")
				os.Exit(1)
			}
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
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
				Privacy:             tripsCreatePrivacy,
				IsMapEmbed:          false,
				Language:            "en",
			}
			resp, err = client.CreateTrip(req)
		}
		if err != nil {
			logger.WithError(err).Error("Failed to create trip")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render("🎉 Successfully created trip!"))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Title: %s", resp.TripPlan.Title)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Trip ID: %d", resp.TripPlan.ID)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Trip Key: %s", resp.TripPlan.Key)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Edit Key: %s", resp.TripPlan.EditKey)))
		fmt.Println(ui.UrlStyle.Render(fmt.Sprintf("URL: https://wanderlog.com/view/%s/%s", resp.TripPlan.Key, resp.TripPlan.Title)))
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
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		fmt.Println(ui.WarningStyle.Render(fmt.Sprintf("⚠️  Are you sure you want to delete trip %s? This cannot be undone.", tripKey)))
		fmt.Print("Type 'yes' to confirm: ")

		var confirmation string
		_, _ = fmt.Scanln(&confirmation)

		if confirmation != "yes" {
			fmt.Println(ui.InfoStyle.Render("Trip deletion canceled."))
			return
		}

		err := client.DeleteTrip(tripKey)
		if err != nil {
			logger.WithError(err).Error("Failed to delete trip")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✅ Successfully deleted trip %s", tripKey)))
	},
}

var tripsCopyCmd = &cobra.Command{
	Use:   "copy [trip-key]",
	Short: "Copy an existing trip",
	Long: `Create a copy of an existing trip plan.

Examples:
  wanderlog trips copy abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourceKey := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		resp, err := client.CopyTrip(sourceKey)
		if err != nil {
			logger.WithError(err).Error("Failed to copy trip")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render("📋 Successfully copied trip!"))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Original: %s", sourceKey)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("New Title: %s", resp.TripPlan.Title)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("New Key: %s", resp.TripPlan.Key)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Edit Key: %s", resp.TripPlan.EditKey)))
		fmt.Println(ui.UrlStyle.Render(fmt.Sprintf("URL: https://wanderlog.com/view/%s/%s", resp.TripPlan.Key, resp.TripPlan.Title)))
	},
}

var tripsRestoreCmd = &cobra.Command{
	Use:   "restore [trip-key]",
	Short: "Restore a deleted trip",
	Long: `Restore a soft-deleted trip plan.

Examples:
  wanderlog trips restore abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		if err := client.RestoreTrip(args[0]); err != nil {
			logger.WithError(err).Error("Failed to restore trip")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Restored trip %s", args[0]), map[string]string{"tripKey": args[0]})
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

	// auth flags
	for _, c := range []*cobra.Command{tripsCreateCmd, tripsDeleteCmd, tripsCopyCmd, tripsRestoreCmd} {
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
