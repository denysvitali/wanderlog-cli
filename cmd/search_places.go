package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func runSearchPlaces(cmd *cobra.Command, args []string) error {
	query := args[0]

	latFlag, _ := cmd.Flags().GetString("lat")
	lngFlag, _ := cmd.Flags().GetString("lng")

	lat, lng := 0.0, 0.0
	var err error

	if latFlag != "" {
		lat, err = strconv.ParseFloat(latFlag, 64)
		if err != nil {
			return fmt.Errorf("invalid latitude %q: %w", latFlag, err)
		}
	}

	if lngFlag != "" {
		lng, err = strconv.ParseFloat(lngFlag, 64)
		if err != nil {
			return fmt.Errorf("invalid longitude %q: %w", lngFlag, err)
		}
	}

	client := wanderlog.NewClient()

	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.SearchPlacesWithWanderlog(query, lat, lng)
	if err != nil {
		return fmt.Errorf("search places: %w", err)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			return fmt.Errorf("encode JSON: %w", err)
		}
	default:
		if len(results.Data) == 0 {
			fmt.Printf("No places found for query: %s\n", query)
			return nil
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
	return nil
}
