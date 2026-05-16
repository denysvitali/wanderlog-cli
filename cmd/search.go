package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

var searchGeosCmd = &cobra.Command{
	Use:   "geos [query]",
	Short: "Search Wanderlog destination geo IDs",
	Long: `Search for Wanderlog destination geo IDs (countries and cities).
Use the returned geo_id with "wanderlog create trip --geo-id".

This command fetches all available geos and filters client-side by name.
Without a query, it returns all geos (limited by --limit).

Examples:
  wanderlog search geos Japan
  wanderlog search geos Tokyo --limit 5`,
	Args: cobra.ExactArgs(1),
	Run:  runSearchGeos,
}

func runSearchGeos(cmd *cobra.Command, args []string) {
	query := args[0]

	limit, _ := cmd.Flags().GetInt("limit")

	client := wanderlog.NewClient()

	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	result, err := client.SearchGeos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching geos: %v\n", err)
		os.Exit(1)
	}

	if limit <= 0 {
		limit = 10
	}

	var matches []geoIDNameMatch
	queryLower := strings.ToLower(query)

	addMatch := func(name string, id int) {
		if strings.Contains(strings.ToLower(name), queryLower) {
			matches = append(matches, geoIDNameMatch{Name: name, GeoID: id})
		}
	}

	for _, c := range result.Countries {
		addMatch(c.Name, c.ID)
	}
	for _, c := range result.Cities {
		addMatch(c.Name, c.ID)
	}

	if len(matches) > limit {
		matches = matches[:limit]
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(map[string]any{
			"success": true,
			"query":   query,
			"count":   len(matches),
			"geos":    matches,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	default:
		if len(matches) == 0 {
			fmt.Printf("No geos found matching: %s\n", query)
			return
		}

		fmt.Printf("Found %d geo(s) matching %q:\n\n", len(matches), query)
		for _, m := range matches {
			fmt.Printf("  %-30s  (geo_id: %d)\n", m.Name, m.GeoID)
		}
		fmt.Println()
		fmt.Println("Use geo_id with: wanderlog create trip --geo-id <id> ...")
	}
}

type geoIDNameMatch struct {
	Name  string `json:"name"`
	GeoID int    `json:"geoId"`
}

func init() {
	rootCmd.AddCommand(searchParentCmd)
	searchParentCmd.AddCommand(searchGoogleCmd, searchWanderlogCmd, searchPlaceDetailsCmd, searchGeosCmd)

	searchGoogleCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
	searchGoogleCmd.Flags().String("lat", "", "Latitude for location context")
	searchGoogleCmd.Flags().String("lng", "", "Longitude for location context")

	searchWanderlogCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
	searchWanderlogCmd.Flags().String("lat", "", "Latitude for location context")
	searchWanderlogCmd.Flags().String("lng", "", "Longitude for location context")

	searchPlaceDetailsCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")

	searchGeosCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
	searchGeosCmd.Flags().Int("limit", 10, "Maximum number of results to return")
}
