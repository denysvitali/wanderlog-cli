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
	tripTitle     string
	tripStartDate string
	tripEndDate   string
	tripPrivacy   string
	tripGeoIDs    []int
	sessionCookie string
	xsrfToken     string
	createExample bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new trip",
	Long: `Create a new trip plan on Wanderlog.

Requires authentication via 'wanderlog login' or environment variables.

Examples:
  wanderlog create --title "Trip to Japan" --geo-id 1
  wanderlog create --title "Europe 2024" --geo-id 7 --start 2024-06-01 --end 2024-06-15
  wanderlog create --title "Private Trip" --geo-id 1 --privacy private
  wanderlog create --example`,
	Run: func(cmd *cobra.Command, args []string) {
		if tripTitle == "" && !createExample {
			logger.Error("Trip title is required")
			os.Exit(1)
		}
		if len(tripGeoIDs) == 0 && !createExample {
			logger.Error("At least one --geo-id is required")
			os.Exit(1)
		}

		// Validate date formats if provided
		if tripStartDate != "" {
			if _, err := time.Parse("2006-01-02", tripStartDate); err != nil {
				logger.WithError(err).Error("Invalid start date format. Use YYYY-MM-DD")
				os.Exit(1)
			}
		}
		if tripEndDate != "" {
			if _, err := time.Parse("2006-01-02", tripEndDate); err != nil {
				logger.WithError(err).Error("Invalid end date format. Use YYYY-MM-DD")
				os.Exit(1)
			}
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Ensure authentication (from flags, env vars, or keychain)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		var resp *wanderlog.CreateTripResponse
		var err error
		if createExample {
			resp, err = client.CreateExampleTrip()
		} else {
			req := wanderlog.CreateTripRequest{
				Title:               tripTitle,
				GeoIDs:              tripGeoIDs,
				InitialMapsPlaceIDs: []int{},
				Type:                "plan",
				StartDate:           tripStartDate,
				EndDate:             tripEndDate,
				Privacy:             tripPrivacy,
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

var deleteCmd = &cobra.Command{
	Use:   "delete [trip-key]",
	Short: "Delete a trip",
	Long: `Delete a trip plan from Wanderlog.

Requires authentication and the trip's edit key.

Examples:
  wanderlog delete abc123xyz
  
WARNING: This action cannot be undone!`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Ensure authentication (from flags, env vars, or keychain)
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

var copyCmd = &cobra.Command{
	Use:   "copy [trip-key]",
	Short: "Copy an existing trip",
	Long: `Create a copy of an existing trip plan.

Examples:
  wanderlog copy abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourceKey := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Ensure authentication (from flags, env vars, or keychain)
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

func init() {
	// root registrations disabled - commands moved under `trips`
	// rootCmd.AddCommand(createCmd)
	// rootCmd.AddCommand(deleteCmd)
	// rootCmd.AddCommand(copyCmd)

	// Create command flags
	createCmd.Flags().StringVarP(&tripTitle, "title", "t", "", "Trip title (required)")
	createCmd.Flags().StringVar(&tripStartDate, "start", "", "Start date (YYYY-MM-DD)")
	createCmd.Flags().StringVar(&tripEndDate, "end", "", "End date (YYYY-MM-DD)")
	createCmd.Flags().StringVar(&tripPrivacy, "privacy", "private", "Trip privacy (public, private, friends)")
	createCmd.Flags().IntSliceVar(&tripGeoIDs, "geo-id", nil, "Wanderlog destination geo ID (repeatable)")
	createCmd.Flags().BoolVar(&createExample, "example", false, "Create Wanderlog's pre-filled example trip")

	// Auth flags for all commands
	for _, cmd := range []*cobra.Command{createCmd, deleteCmd, copyCmd} {
		cmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		cmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
