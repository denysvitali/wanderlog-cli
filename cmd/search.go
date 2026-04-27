package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var searchParentCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for places",
	Long:  `Search for places using Wanderlog's autocomplete API.`,
}

var searchGoogleCmd = &cobra.Command{
	Use:    "google [query]",
	Short:  "Deprecated alias for Wanderlog place search",
	Long:   `Deprecated alias. Search is always handled by Wanderlog's autocomplete API.`,
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run:    runSearchPlaces,
}

var searchWanderlogCmd = &cobra.Command{
	Use:     "places [query]",
	Aliases: []string{"wanderlog"},
	Short:   "Search places using Wanderlog's autocomplete API",
	Long: `Search for places using Wanderlog's native autocomplete API with optional location context.

Examples:
  wanderlog search places "Eiffel Tower"
  wanderlog search places "Tokyo Station" --lat 35.6812 --lng 139.7671`,
	Args: cobra.ExactArgs(1),
	Run:  runSearchPlaces,
}

var searchPlaceDetailsCmd = &cobra.Command{
	Use:   "place-details [place-id]",
	Short: "Get detailed information about a place",
	Long: `Fetch detailed information about a place from Wanderlog's place details API.

Examples:
  wanderlog search place-details ChIJLU7jZClu5kcR4PcOOO6p3I0`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		placeID := args[0]

		client := wanderlog.NewClient()

		auth, err := wanderlog.LoadCredentials()
		if err == nil {
			client.SetAuth(auth)
		}

		details, err := client.GetPlaceDetails(placeID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting place details: %v\n", err)
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(details); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Println(ui.TitleStyle.Render(fmt.Sprintf("📍 %s", details.Data.Details.Name)))
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   🆔 %s", details.Data.Details.PlaceID)))
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   🏠 %s", details.Data.Details.FormattedAddress)))

			if details.Data.Details.Rating > 0 {
				fmt.Println(ui.HighlightStyle.Render(fmt.Sprintf("   ⭐ %.1f/5 (%d reviews)",
					details.Data.Details.Rating, details.Data.Details.UserRatingsTotal)))
			}

			if details.Data.Details.Website != "" {
				fmt.Println(ui.UrlStyle.Render(fmt.Sprintf("   🌐 %s", details.Data.Details.Website)))
			}

			if details.Data.Details.InternationalPhoneNumber != "" {
				fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   📞 %s", details.Data.Details.InternationalPhoneNumber)))
			}

			if len(details.Data.Details.Types) > 0 {
				fmt.Println(ui.CategoryStyle.Render(fmt.Sprintf("   🏷️  %v", details.Data.Details.Types)))
			}

			if details.Data.CardData.ReviewsSummary != "" {
				fmt.Println()
				fmt.Println(ui.SubHeaderStyle.Render("📝 Reviews Summary:"))
				fmt.Println(ui.InfoStyle.Render(details.Data.CardData.ReviewsSummary))
			}

			if len(details.Data.CardData.ReasonsToVisit) > 0 {
				fmt.Println()
				fmt.Println(ui.SubHeaderStyle.Render("✨ Reasons to Visit:"))
				for i, reason := range details.Data.CardData.ReasonsToVisit {
					fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   %d. %s", i+1, reason)))
				}
			}

			if len(details.Data.CardData.Tips) > 0 {
				fmt.Println()
				fmt.Println(ui.SubHeaderStyle.Render("💡 Tips:"))
				for i, tip := range details.Data.CardData.Tips {
					fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("   %d. %s", i+1, tip)))
				}
			}

			coords := details.Data.Details.Geometry.Location
			fmt.Println()
			fmt.Println(ui.DimStyle.Render(fmt.Sprintf("📍 Coordinates: %.6f, %.6f", coords.Lat, coords.Lng)))
		}
	},
}

func init() {
	rootCmd.AddCommand(searchParentCmd)
	searchParentCmd.AddCommand(searchGoogleCmd, searchWanderlogCmd, searchPlaceDetailsCmd)

	searchGoogleCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
	searchGoogleCmd.Flags().String("lat", "", "Latitude for location context")
	searchGoogleCmd.Flags().String("lng", "", "Longitude for location context")

	searchWanderlogCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
	searchWanderlogCmd.Flags().String("lat", "", "Latitude for location context")
	searchWanderlogCmd.Flags().String("lng", "", "Longitude for location context")

	searchPlaceDetailsCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
}
