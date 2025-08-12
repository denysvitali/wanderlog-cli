package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/spf13/cobra"
)

var placeDetailsCmd = &cobra.Command{
	Use:   "place-details [place-id]",
	Short: "Get detailed information about a place",
	Long:  `Fetch detailed information about a place from Wanderlog's place details API.`,
	Args:  cobra.ExactArgs(1),
	Run:   runPlaceDetails,
}

func runPlaceDetails(cmd *cobra.Command, args []string) {
	placeID := args[0]

	client := wanderlog.NewClient()

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	details, err := client.GetPlaceDetails(placeID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting place details: %v\n", err)
		os.Exit(1)
	}

	// Format output based on the requested format
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
		// Human-readable format
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
}

func init() {
	placeDetailsCmd.Flags().String("format", "human", "Output format (human, json)")
	rootCmd.AddCommand(placeDetailsCmd)
}
