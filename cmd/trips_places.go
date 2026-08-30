package cmd

import (
	"fmt"

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
			return ui.PrintJSON(trip.Resources.PlaceMetadata)
		case "markdown", "md":
			ui.PrintPlacesMarkdown(trip.Resources.PlaceMetadata)
		default:
			ui.PrintPlaces(trip.Resources.PlaceMetadata)
		}
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		tripID := args[0]

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		images, err := client.GetTripImages(tripID)
		if err != nil {
			return fmt.Errorf("fetch trip images: %w", err)
		}

		switch outputFormat {
		case "json":
			return ui.PrintJSON(images)
		case "markdown", "md":
			tripsImagesMarkdown(images, tripID)
		default:
			tripsImagesPretty(images, tripID)
		}
		return nil
	},
}

var tripsExpensesCmd = &cobra.Command{
	Use:   "expenses [trip-key]",
	Short: "Download a trip's expenses as CSV",
	Long: `Download a trip's expenses as CSV.

Examples:
  wanderlog trips expenses abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		csv, err := client.GetTripExpensesCSV(args[0])
		if err != nil {
			return fmt.Errorf("fetch expenses CSV: %w", err)
		}
		if _, err := cmd.OutOrStdout().Write(csv); err != nil {
			return fmt.Errorf("write expenses CSV: %w", err)
		}
		if len(csv) > 0 && csv[len(csv)-1] != '\n' {
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}
		return nil
	},
}

func init() {
	tripsCmd.AddCommand(tripsPlacesCmd, tripsImagesCmd, tripsExpensesCmd)

	tripsPlacesCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json, markdown)")
	tripsPlacesCmd.Flags().StringVar(&fromFile, "file", "", "Load trip data from local JSON file instead of API")

	tripsImagesCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json, markdown)")
	tripsImagesCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsImagesCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")

	tripsExpensesCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsExpensesCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
}

func tripsImagesPretty(images *wanderlog.TripImagesResponse, tripID string) {
	if len(images.Images) == 0 {
		fmt.Println(ui.WarningStyle.Render(fmt.Sprintf("📷 No images found for trip %s", ui.SafeText(tripID))))
		return
	}

	fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("📷 Trip Images (%d total)", len(images.Images))))
	fmt.Println()

	for i, img := range images.Images {
		fmt.Printf("%d. %s\n", i+1, ui.PlaceStyle.Render(ui.SafeText(img.Key)))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Size: %dx%d", img.Width, img.Height)))
		if img.Caption != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Caption: %s", ui.SafeText(img.Caption))))
		}
		if img.PlaceID != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Place ID: %s", ui.SafeText(img.PlaceID))))
		}
		fmt.Println(ui.UrlStyle.Render(fmt.Sprintf("   URL: %s", ui.SafeText(img.URL))))
		if img.ThumbnailURL != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   Thumbnail: %s", ui.SafeText(img.ThumbnailURL))))
		}
		fmt.Println()
	}
}

func tripsImagesMarkdown(images *wanderlog.TripImagesResponse, tripID string) {
	fmt.Printf("# Trip Images\n\n")
	fmt.Printf("Trip ID: %s\n", ui.MarkdownInline(tripID))
	fmt.Printf("Total images: %d\n\n", len(images.Images))

	for i, img := range images.Images {
		fmt.Printf("## Image %d\n\n", i+1)
		fmt.Printf("- **Key:** %s\n", ui.MarkdownInline(img.Key))
		fmt.Printf("- **Size:** %dx%d\n", img.Width, img.Height)
		if img.Caption != "" {
			fmt.Printf("- **Caption:** %s\n", ui.MarkdownInline(img.Caption))
		}
		if img.PlaceID != "" {
			fmt.Printf("- **Place ID:** %s\n", ui.MarkdownInline(img.PlaceID))
		}
		fmt.Printf("- **URL:** %s\n", ui.MarkdownInline(img.URL))
		if img.ThumbnailURL != "" {
			fmt.Printf("- **Thumbnail:** %s\n", ui.MarkdownInline(img.ThumbnailURL))
		}
		fmt.Println()
	}
}
