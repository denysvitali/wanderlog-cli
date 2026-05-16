package ui

import (
	"fmt"
	"strings"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

// PrintSearchResults prints place search results in a pretty format
func PrintSearchResults(results []wanderlog.SearchResult) {
	if len(results) == 0 {
		fmt.Println("🔍 No places found matching your search query")
		fmt.Println("💡 Try a different search term or check your spelling")
		return
	}

	fmt.Println(TitleStyle.Render(fmt.Sprintf("🔍 Found %d place(s)", len(results))))
	fmt.Println()

	for i, place := range results {
		// Place name with rating
		name := place.Name
		if place.Rating > 0 {
			stars := strings.Repeat("⭐", int(place.Rating))
			if len(stars) > 5 {
				stars = stars[:5]
			}
			name += fmt.Sprintf(" %s (%.1f)", stars, place.Rating)
		}

		fmt.Println(PlaceStyle.Render(fmt.Sprintf("📍 %s", name)))

		// Address
		if place.Address != "" {
			fmt.Println(InfoStyle.Render(fmt.Sprintf("   🏠 %s", place.Address)))
		}

		// Categories
		if len(place.Categories) > 0 {
			categories := strings.Join(place.Categories, ", ")
			fmt.Println(InfoStyle.Render(fmt.Sprintf("   🏷️  %s", categories)))
		}

		// Description
		if place.Description != "" {
			fmt.Println(InfoStyle.Render(fmt.Sprintf("   📝 %s", place.Description)))
		}

		// Website
		if place.Website != "" {
			fmt.Println(InfoStyle.Render(fmt.Sprintf("   🌐 %s", place.Website)))
		}

		// Coordinates
		if place.Latitude != 0 && place.Longitude != 0 {
			fmt.Println(InfoStyle.Render(fmt.Sprintf("   🗺️  %.4f, %.4f", place.Latitude, place.Longitude)))
		}

		// Place ID
		if place.PlaceID != "" {
			fmt.Println(InfoStyle.Render(fmt.Sprintf("   🆔 %s", place.PlaceID)))
		}

		// Spacer between places
		if i < len(results)-1 {
			fmt.Println()
		}
	}
}

// PrintSearchResultsMarkdown prints place search results in Markdown format
func PrintSearchResultsMarkdown(results []wanderlog.SearchResult) {
	if len(results) == 0 {
		fmt.Println("# Search Results\n\nNo places found matching your search query.\nTry a different search term or check your spelling.")
		return
	}

	fmt.Printf("# Search Results\n\nFound %d place(s):\n\n", len(results))

	for i, place := range results {
		fmt.Printf("## %d. %s\n\n", i+1, place.Name)

		if place.Rating > 0 {
			fmt.Printf("**Rating:** %.1f/5 ⭐\n\n", place.Rating)
		}

		if place.Address != "" {
			fmt.Printf("**Address:** %s\n\n", place.Address)
		}

		if len(place.Categories) > 0 {
			fmt.Printf("**Categories:** %s\n\n", strings.Join(place.Categories, ", "))
		}

		if place.Description != "" {
			fmt.Printf("**Description:** %s\n\n", place.Description)
		}

		if place.Website != "" {
			fmt.Printf("**Website:** %s\n\n", place.Website)
		}

		if place.Latitude != 0 && place.Longitude != 0 {
			fmt.Printf("**Location:** %.4f, %.4f\n\n", place.Latitude, place.Longitude)
		}

		if place.PlaceID != "" {
			fmt.Printf("**Place ID:** %s\n\n", place.PlaceID)
		}

		fmt.Println("---")
		fmt.Println()
	}
}
