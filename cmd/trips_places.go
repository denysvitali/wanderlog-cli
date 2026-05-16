package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsPlacesCmd = &cobra.Command{
	Use:   "places [trip-id]",
	Short: "Show places from a trip",
	Long: `Display detailed information about places in a trip including
names, addresses, ratings, and other metadata.

Examples:
  wanderlog trips places abc123xyz
  wanderlog trips places --file trips/trip1.json
  wanderlog trips places abc123xyz --output json`,
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
			ui.PrintJSON(trip.Resources.PlaceMetadata)
		case "markdown", "md":
			ui.PrintPlacesMarkdown(trip.Resources.PlaceMetadata)
		default:
			ui.PrintPlaces(trip.Resources.PlaceMetadata)
		}
	},
}

var tripsImagesCmd = &cobra.Command{
	Use:   "images [trip-id]",
	Short: "Show trip images",
	Long: `Display images for a trip.

Examples:
  wanderlog trips images abc123xyz
  wanderlog trips images abc123xyz --output json`,
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
			tripsImagesMarkdown(images, tripID)
		default:
			tripsImagesPretty(images, tripID)
		}
	},
}

var tripsExpensesCmd = &cobra.Command{
	Use:   "expenses [trip-key]",
	Short: "Download a trip's expenses as CSV",
	Long: `Download a trip's expenses as CSV.

Examples:
  wanderlog trips expenses abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		csv, err := client.GetTripExpensesCSV(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch expenses CSV")
			os.Exit(1)
		}
		_, _ = os.Stdout.Write(csv)
		if len(csv) > 0 && csv[len(csv)-1] != '\n' {
			fmt.Println()
		}
	},
}

func init() {
	tripsCmd.AddCommand(tripsPlacesCmd, tripsImagesCmd, tripsExpensesCmd)

	tripsPlacesCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json, markdown)")
	tripsPlacesCmd.Flags().StringVar(&fromFile, "file", "", "Load trip data from local JSON file instead of API")

	tripsImagesCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json, markdown)")
	tripsImagesCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsImagesCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")

	tripsExpensesCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format")
	tripsExpensesCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsExpensesCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
}

func tripsImagesPretty(images *wanderlog.TripImagesResponse, tripID string) {
	if len(images.Images) == 0 {
		fmt.Println(ui.WarningStyle.Render(fmt.Sprintf("📷 No images found for trip %s", tripID)))
		return
	}

	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("📷 Trip Images (%d total)", len(images.Images))))
	fmt.Println()

	for i, img := range images.Images {
		fmt.Printf("%d. %s\n", i+1, ui.PlaceStyle.Render(img.Key))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Size: %dx%d", img.Width, img.Height)))
		if img.Caption != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Caption: %s", img.Caption)))
		}
		if img.PlaceID != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Place ID: %s", img.PlaceID)))
		}
		fmt.Println(ui.UrlStyle.Render(fmt.Sprintf("   URL: %s", img.URL)))
		if img.ThumbnailURL != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Thumbnail: %s", img.ThumbnailURL)))
		}
		fmt.Println()
	}
}

func tripsImagesMarkdown(images *wanderlog.TripImagesResponse, tripID string) {
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
