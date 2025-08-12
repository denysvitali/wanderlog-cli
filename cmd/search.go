package cmd

import (
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

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for places using Google Places API",
	Long: `Search for places by name, address, or keywords using Google Places API.

This command requires a valid Google Places API key set in the GOOGLE_PLACES_API_KEY environment variable.
You can get one from: https://console.cloud.google.com/apis/library/places-backend.googleapis.com

Examples:
  export GOOGLE_PLACES_API_KEY=your_api_key_here
  wanderlog search "Eiffel Tower"
  wanderlog search "restaurants" --lat 40.7128 --lng -74.0060
  wanderlog search "pizza" --format json`,
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

		// Get API key from environment variable
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
			fmt.Println("   6. Run: wanderlog search \"your query\"")
			os.Exit(1)
		}

		// Parse coordinates if provided
		var lat, lng *float64
		if searchLat != "" && searchLng != "" {
			if parsedLat, err := strconv.ParseFloat(searchLat, 64); err == nil {
				lat = &parsedLat
			} else {
				fmt.Printf("Invalid latitude: %s\n", searchLat)
				os.Exit(1)
			}

			if parsedLng, err := strconv.ParseFloat(searchLng, 64); err == nil {
				lng = &parsedLng
			} else {
				fmt.Printf("Invalid longitude: %s\n", searchLng)
				os.Exit(1)
			}
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

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&outputFormat, "format", "f", "pretty", "Output format (pretty, json, markdown)")
	searchCmd.Flags().StringVar(&searchLat, "lat", "", "Latitude for location-based search")
	searchCmd.Flags().StringVar(&searchLng, "lng", "", "Longitude for location-based search")
}
