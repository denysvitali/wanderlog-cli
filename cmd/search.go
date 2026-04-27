package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	searchLat string
	searchLng string
)

var searchParentCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for places",
	Long:  `Search for places using Google Places API or Wanderlog's native autocomplete.`,
}

var searchGoogleCmd = &cobra.Command{
	Use:   "google [query]",
	Short: "Search places using Google Places API",
	Long: `Search for places by name, address, or keywords using Google Places API.

This command requires a valid Google Places API key set in the GOOGLE_PLACES_API_KEY environment variable.
You can get one from: https://console.cloud.google.com/apis/library/places-backend.googleapis.com

Examples:
  export GOOGLE_PLACES_API_KEY=your_api_key_here
  wanderlog search google "Eiffel Tower"
  wanderlog search google "restaurants" --lat 40.7128 --lng -74.0060
  wanderlog search google "pizza" --format json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one search query argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		apiKey := os.Getenv("GOOGLE_PLACES_API_KEY")
		if apiKey == "" {
			fmt.Println("❌ Google Places API key is required")
			fmt.Println("")
			fmt.Println("💡 To get a Google Places API key:")
			fmt.Println("   1. Go to https://console.cloud.google.com/")
			fmt.Println("   2. Create or select a project")
			fmt.Println("   3. Enable the Places API")
			fmt.Println("   4. Create credentials (API key)")
			fmt.Println("   5. Set environment variable:")
			fmt.Println("      export GOOGLE_PLACES_API_KEY=your_key_here")
			fmt.Println("   6. Run: wanderlog search google \"your query\"")
			os.Exit(1)
		}

		var lat, lng *float64
		if searchLat != "" && searchLng != "" {
			parsedLat, err := strconv.ParseFloat(searchLat, 64)
			if err != nil {
				fmt.Printf("Invalid latitude: %s\n", searchLat)
				os.Exit(1)
			}
			lat = &parsedLat

			parsedLng, err := strconv.ParseFloat(searchLng, 64)
			if err != nil {
				fmt.Printf("Invalid longitude: %s\n", searchLng)
				os.Exit(1)
			}
			lng = &parsedLng
		}

		results, err := client.SearchPlaces(query, lat, lng, apiKey)
		if err != nil {
			logger.WithError(err).Error("Place search failed")
			fmt.Printf("❌ %s\n", err.Error())
			os.Exit(1)
		}

		switch outputFormat {
		case "json":
			ui.PrintJSON(results)
		case "markdown", "md":
			ui.PrintSearchResultsMarkdown(results.Places)
		default:
			ui.PrintSearchResults(results.Places)
		}
	},
}

var searchWanderlogCmd = &cobra.Command{
	Use:   "wanderlog [query]",
	Short: "Search places using Wanderlog's autocomplete API",
	Long: `Search for places using Wanderlog's native autocomplete API with optional location context.

Examples:
  wanderlog search wanderlog "Eiffel Tower"
  wanderlog search wanderlog "Tokyo Station" --lat 35.6812 --lng 139.7671`,
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

		format, _ := cmd.Flags().GetString("format")
		switch format {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(details); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Place: %s\n", details.Data.Details.Name)
			fmt.Printf("Place ID: %s\n", details.Data.Details.PlaceID)
			fmt.Printf("Address: %s\n", details.Data.Details.FormattedAddress)

			if details.Data.Details.Rating > 0 {
				fmt.Printf("Rating: %.1f/5 (%d reviews)\n",
					details.Data.Details.Rating, details.Data.Details.UserRatingsTotal)
			}

			if details.Data.Details.Website != "" {
				fmt.Printf("Website: %s\n", details.Data.Details.Website)
			}

			if details.Data.Details.InternationalPhoneNumber != "" {
				fmt.Printf("Phone: %s\n", details.Data.Details.InternationalPhoneNumber)
			}

			if len(details.Data.Details.Types) > 0 {
				fmt.Printf("Types: %v\n", details.Data.Details.Types)
			}

			if details.Data.CardData.ReviewsSummary != "" {
				fmt.Printf("\nReviews Summary:\n%s\n", details.Data.CardData.ReviewsSummary)
			}

			if len(details.Data.CardData.ReasonsToVisit) > 0 {
				fmt.Printf("\nReasons to Visit:\n")
				for i, reason := range details.Data.CardData.ReasonsToVisit {
					fmt.Printf("  %d. %s\n", i+1, reason)
				}
			}

			if len(details.Data.CardData.Tips) > 0 {
				fmt.Printf("\nTips:\n")
				for i, tip := range details.Data.CardData.Tips {
					fmt.Printf("  %d. %s\n", i+1, tip)
				}
			}

			coords := details.Data.Details.Geometry.Location
			fmt.Printf("\nCoordinates: %.6f, %.6f\n", coords.Lat, coords.Lng)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchParentCmd)
	searchParentCmd.AddCommand(searchGoogleCmd, searchWanderlogCmd, searchPlaceDetailsCmd)

	searchGoogleCmd.Flags().StringVarP(&outputFormat, "format", "f", "pretty", "Output format (pretty, json, markdown)")
	searchGoogleCmd.Flags().StringVar(&searchLat, "lat", "", "Latitude for location-based search")
	searchGoogleCmd.Flags().StringVar(&searchLng, "lng", "", "Longitude for location-based search")

	searchWanderlogCmd.Flags().String("format", "human", "Output format (human, json)")
	searchWanderlogCmd.Flags().String("lat", "", "Latitude for location context")
	searchWanderlogCmd.Flags().String("lng", "", "Longitude for location context")

	searchPlaceDetailsCmd.Flags().String("format", "human", "Output format (human, json)")
}
