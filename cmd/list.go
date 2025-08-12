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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your trips",
	Long: `List all trips for the authenticated user.

Requires authentication via 'wanderlog login'.

Examples:
  wanderlog list
  wanderlog list --format json`,
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Ensure authentication (from flags, env vars, or keychain)
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
			printTripsMarkdown(trips)
		default:
			printTripsList(trips)
		}
	},
}

var imagesCmd = &cobra.Command{
	Use:   "images [trip-id]",
	Short: "Show trip images",
	Long: `Display images for a trip.

Examples:
  wanderlog images abc123xyz
  wanderlog images abc123xyz --format json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tripID := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		images, err := client.GetTripImages(tripID)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch trip images")
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			ui.PrintJSON(images)
		case "markdown", "md":
			printImagesMarkdown(images, tripID)
		default:
			printImagesList(images, tripID)
		}
	},
}

func printTripsList(trips *wanderlog.UserTripsResponse) {
	if len(trips.Data) == 0 {
		fmt.Println("📭 No trips found")
		return
	}

	fmt.Printf("📚 Your Trips (%d total)\n\n", len(trips.Data))

	for _, trip := range trips.Data {
		// Trip title with privacy indicator (default to public since privacy field not in response)
		privacy := "🌍" // Default to public
		if trip.IsPrimary {
			privacy = "⭐" // Star for primary trips
		}

		fmt.Printf("%s %s\n", privacy, trip.Title)
		fmt.Printf("   Key: %s\n", trip.Key)

		// Dates
		if trip.StartDate != "" && trip.EndDate != "" {
			startDate, _ := time.Parse("2006-01-02", trip.StartDate)
			endDate, _ := time.Parse("2006-01-02", trip.EndDate)
			days := int(endDate.Sub(startDate).Hours()/24) + 1
			fmt.Printf("   📅 %s → %s (%d days)\n",
				startDate.Format("Jan 2, 2006"),
				endDate.Format("Jan 2, 2006"),
				days)
		}

		// Stats
		stats := []string{
			fmt.Sprintf("📍 %d places", trip.PlaceCount),
			fmt.Sprintf("👀 %d views", trip.ViewCount),
		}
		if trip.LikeCount > 0 {
			stats = append(stats, fmt.Sprintf("❤️ %d likes", trip.LikeCount))
		}

		fmt.Printf("   %s\n", strings.Join(stats, "  •  "))

		// Additional indicators
		if trip.IsPrimary {
			fmt.Printf("   ⭐ Primary Trip\n")
		}
		if trip.IsDraft {
			fmt.Printf("   📝 Draft\n")
		}

		fmt.Println()
	}
}

func printImagesList(images *wanderlog.TripImagesResponse, tripID string) {
	if len(images.Images) == 0 {
		fmt.Printf("📷 No images found for trip %s\n", tripID)
		return
	}

	fmt.Printf("📷 Trip Images (%d total)\n\n", len(images.Images))

	for i, img := range images.Images {
		fmt.Printf("%d. %s\n", i+1, img.Key)
		fmt.Printf("   Size: %dx%d\n", img.Width, img.Height)
		if img.Caption != "" {
			fmt.Printf("   Caption: %s\n", img.Caption)
		}
		if img.PlaceID != "" {
			fmt.Printf("   Place ID: %s\n", img.PlaceID)
		}
		fmt.Printf("   URL: %s\n", img.URL)
		if img.ThumbnailURL != "" {
			fmt.Printf("   Thumbnail: %s\n", img.ThumbnailURL)
		}
		fmt.Println()
	}
}

func printTripsMarkdown(trips *wanderlog.UserTripsResponse) {
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

func printImagesMarkdown(images *wanderlog.TripImagesResponse, tripID string) {
	fmt.Printf("# Trip Images\n\n")
	fmt.Printf("Trip ID: %s\n", tripID)
	fmt.Printf("Total images: %d\n\n", len(images.Images))

	for i, img := range images.Images {
		fmt.Printf("## Image %d\n\n", i+1)
		fmt.Printf("- **Key:** %s\n", img.Key)
		fmt.Printf("- **Size:** %dx%d\n", img.Width, img.Height)
		if img.Caption != "" {
			fmt.Printf("- **Caption:** %s\n", img.Caption)
		}
		if img.PlaceID != "" {
			fmt.Printf("- **Place ID:** %s\n", img.PlaceID)
		}
		fmt.Printf("- **URL:** %s\n", img.URL)
		if img.ThumbnailURL != "" {
			fmt.Printf("- **Thumbnail:** %s\n", img.ThumbnailURL)
		}
		fmt.Println()
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(imagesCmd)

	// Add format flags
	for _, cmd := range []*cobra.Command{listCmd, imagesCmd} {
		cmd.Flags().StringVarP(&outputFormat, "format", "f", "pretty", "Output format (pretty, json, markdown)")
		cmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		cmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}