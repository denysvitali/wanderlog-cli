package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/spf13/cobra"
)

var searchPlacesCmd = &cobra.Command{
	Use:   "search-places [query]",
	Short: "Search for places using Wanderlog's autocomplete API",
	Long:  `Search for places using Wanderlog's autocomplete API with optional location context.`,
	Args:  cobra.ExactArgs(1),
	Run:   runSearchPlaces,
}

func runSearchPlaces(cmd *cobra.Command, args []string) {
	query := args[0]

	// Get location parameters
	latFlag, _ := cmd.Flags().GetString("lat")
	lngFlag, _ := cmd.Flags().GetString("lng")

	var lat, lng float64 = 0.0, 0.0
	var err error

	if latFlag != "" {
		lat, err = strconv.ParseFloat(latFlag, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid latitude: %v\n", err)
			os.Exit(1)
		}
	}

	if lngFlag != "" {
		lng, err = strconv.ParseFloat(lngFlag, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid longitude: %v\n", err)
			os.Exit(1)
		}
	}

	client := wanderlog.NewClient()

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.SearchPlacesWithWanderllog(query, lat, lng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching places: %v\n", err)
		os.Exit(1)
	}

	// Format output based on the requested format
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	default:
		// Human-readable format
		if len(results.Data) == 0 {
			fmt.Printf("No places found for query: %s\n", query)
			return
		}

		fmt.Printf("Found %d places for query: %s\n\n", len(results.Data), query)

		for i, place := range results.Data {
			fmt.Printf("%d. %s\n", i+1, place.Description)
			if place.PlaceID != "" {
				fmt.Printf("   Place ID: %s\n", place.PlaceID)
			}
			if len(place.Types) > 0 {
				fmt.Printf("   Types: %v\n", place.Types)
			}
			if place.Type != "" {
				fmt.Printf("   Type: %s\n", place.Type)
			}
			fmt.Println()
		}
	}
}

func init() {
	searchPlacesCmd.Flags().String("format", "human", "Output format (human, json)")
	searchPlacesCmd.Flags().String("lat", "", "Latitude for location context")
	searchPlacesCmd.Flags().String("lng", "", "Longitude for location context")
	rootCmd.AddCommand(searchPlacesCmd)
}
