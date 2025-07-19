package ui

import (
	"fmt"
	"strings"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func PrintPlaces(places []wanderlog.Metadata) {
	if len(places) == 0 {
		fmt.Println("🏠 No places found in this trip")
		return
	}

	fmt.Println(titleStyle.Render("🏠 Places"))
	fmt.Println()

	for i, place := range places {
		// Place name with rating
		name := place.Name
		if place.Rating > 0 {
			stars := strings.Repeat("⭐", int(place.Rating))
			if len(stars) > 5 {
				stars = stars[:5]
			}
			name += fmt.Sprintf(" %s (%.1f)", stars, place.Rating)
		}
		
		fmt.Println(placeStyle.Render(fmt.Sprintf("📍 %s", name)))

		// Address
		if place.Address != "" {
			fmt.Println(infoStyle.Render(fmt.Sprintf("   📍 %s", place.Address)))
		}

		// Categories
		if len(place.Categories) > 0 {
			categories := strings.Join(place.Categories, ", ")
			fmt.Println(infoStyle.Render(fmt.Sprintf("   🏷️  %s", categories)))
		}

		// Website
		if place.Website != "" {
			fmt.Println(infoStyle.Render(fmt.Sprintf("   🌐 %s", place.Website)))
		}

		// Phone
		if place.InternationalPhoneNumber != "" {
			fmt.Println(infoStyle.Render(fmt.Sprintf("   📞 %s", place.InternationalPhoneNumber)))
		}

		// Business status
		if place.BusinessStatus != "" && place.BusinessStatus != "OPERATIONAL" {
			fmt.Println(infoStyle.Render(fmt.Sprintf("   ⚠️  Status: %s", place.BusinessStatus)))
		}

		// Permanently closed warning
		if place.PermanentlyClosed {
			fmt.Println(infoStyle.Render("   ❌ Permanently Closed"))
		}

		// Description
		if place.Description != nil && *place.Description != "" {
			desc := *place.Description
			if len(desc) > 100 {
				desc = desc[:97] + "..."
			}
			fmt.Println(infoStyle.Render(fmt.Sprintf("   💬 %s", desc)))
		}

		// Spacer between places
		if i < len(places)-1 {
			fmt.Println()
		}
	}
}