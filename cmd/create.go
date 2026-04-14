package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	tripTitle     string
	tripStartDate string
	tripEndDate   string
	tripPrivacy   string
	sessionCookie string
	xsrfToken     string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new trip",
	Long: `Create a new trip plan on Wanderlog.

Requires authentication via 'wanderlog login' or environment variables.

Examples:
  wanderlog create --title "Trip to Japan" 
  wanderlog create --title "Europe 2024" --start 2024-06-01 --end 2024-06-15
  wanderlog create --title "Private Trip" --privacy private`,
	Run: func(cmd *cobra.Command, args []string) {
		if tripTitle == "" {
			logger.Error("Trip title is required")
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

		req := wanderlog.CreateTripRequest{
			Title:     tripTitle,
			StartDate: tripStartDate,
			EndDate:   tripEndDate,
			Privacy:   tripPrivacy,
		}

		resp, err := client.CreateTrip(req)
		if err != nil {
			logger.WithError(err).Error("Failed to create trip")
			os.Exit(1)
		}

		fmt.Printf("🎉 Successfully created trip!\n")
		fmt.Printf("Title: %s\n", resp.TripPlan.Title)
		fmt.Printf("Trip ID: %d\n", resp.TripPlan.ID)
		fmt.Printf("Trip Key: %s\n", resp.TripPlan.Key)
		fmt.Printf("Edit Key: %s\n", resp.TripPlan.EditKey)
		fmt.Printf("URL: https://wanderlog.com/view/%s/%s\n", resp.TripPlan.Key, resp.TripPlan.Title)
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

		fmt.Printf("⚠️  Are you sure you want to delete trip %s? This cannot be undone.\n", tripKey)
		fmt.Print("Type 'yes' to confirm: ")

		var confirmation string
		_, _ = fmt.Scanln(&confirmation)

		if confirmation != "yes" {
			fmt.Println("Trip deletion canceled.")
			return
		}

		err := client.DeleteTrip(tripKey)
		if err != nil {
			logger.WithError(err).Error("Failed to delete trip")
			os.Exit(1)
		}

		fmt.Printf("✅ Successfully deleted trip %s\n", tripKey)
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

		fmt.Printf("📋 Successfully copied trip!\n")
		fmt.Printf("Original: %s\n", sourceKey)
		fmt.Printf("New Title: %s\n", resp.TripPlan.Title)
		fmt.Printf("New Key: %s\n", resp.TripPlan.Key)
		fmt.Printf("Edit Key: %s\n", resp.TripPlan.EditKey)
		fmt.Printf("URL: https://wanderlog.com/view/%s/%s\n", resp.TripPlan.Key, resp.TripPlan.Title)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(copyCmd)

	// Create command flags
	createCmd.Flags().StringVarP(&tripTitle, "title", "t", "", "Trip title (required)")
	createCmd.Flags().StringVar(&tripStartDate, "start", "", "Start date (YYYY-MM-DD)")
	createCmd.Flags().StringVar(&tripEndDate, "end", "", "End date (YYYY-MM-DD)")
	createCmd.Flags().StringVar(&tripPrivacy, "privacy", "public", "Trip privacy (public, private, unlisted)")

	// Auth flags for all commands
	for _, cmd := range []*cobra.Command{createCmd, deleteCmd, copyCmd} {
		cmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		cmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
