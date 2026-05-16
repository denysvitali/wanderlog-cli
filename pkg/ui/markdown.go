package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func PrintTripMarkdown(trip *wanderlog.TripResponse, showDetails bool) {
	if !trip.Success {
		fmt.Println("# Trip Data Unavailable")
		fmt.Println("\nFailed to fetch trip data.")
		return
	}

	plan := trip.TripPlan

	// Trip header
	fmt.Printf("# %s\n\n", plan.Title)

	// Trip metadata
	fmt.Println("## Trip Information")
	fmt.Println()

	if plan.StartDate != "" && plan.EndDate != "" {
		startDate, _ := time.Parse("2006-01-02", plan.StartDate)
		endDate, _ := time.Parse("2006-01-02", plan.EndDate)

		fmt.Printf("- **Dates:** %s to %s\n",
			startDate.Format("January 2, 2006"),
			endDate.Format("January 2, 2006"))
		fmt.Printf("- **Duration:** %d days\n", plan.Days)
	}

	fmt.Printf("- **Places:** %d\n", plan.PlaceCount)
	fmt.Printf("- **Views:** %d\n", plan.ViewCount)

	if plan.LikeCount > 0 {
		fmt.Printf("- **Likes:** %d\n", plan.LikeCount)
	}

	fmt.Println()

	if showDetails {
		printFlightsMarkdown(plan.Itinerary.Sections)
		printDestinationsMarkdown(plan.Itinerary.Sections)
	} else {
		fmt.Printf("*Use --details flag to see complete itinerary with flights and destinations.*\n\n")
	}
}

func printFlightsMarkdown(sections []wanderlog.ItSections) {
	for _, section := range sections {
		if section.Heading == "Flights" && len(section.Blocks) > 0 {
			fmt.Println("## Flights")
			fmt.Println()

			for _, block := range section.Blocks {
				if block.Type == "flight" && block.FlightInfo != nil {
					// Flight details
					fmt.Printf("### %s Flight %d\n\n",
						block.FlightInfo.Airline.Name,
						block.FlightInfo.Number)

					// Departure
					departTime, _ := time.Parse("2006-01-02", block.Depart.Date)
					fmt.Printf("- **Departure:** %s at %s\n",
						departTime.Format("January 2, 2006"),
						block.Depart.Time)
					fmt.Printf("- **From:** %s (%s) - %s\n",
						block.Depart.Airport.Name,
						block.Depart.Airport.Iata,
						block.Depart.Airport.CityName)

					// Arrival
					if block.Arrive != nil {
						arriveTime, _ := time.Parse("2006-01-02", block.Arrive.Date)
						fmt.Printf("- **Arrival:** %s at %s\n",
							arriveTime.Format("January 2, 2006"),
							block.Arrive.Time)
						fmt.Printf("- **To:** %s (%s) - %s\n",
							block.Arrive.Airport.Name,
							block.Arrive.Airport.Iata,
							block.Arrive.Airport.CityName)
					}
					fmt.Println()
				}
			}
		}
	}
}

func printDestinationsMarkdown(sections []wanderlog.ItSections) {
	fmt.Println("## Daily Itinerary")
	fmt.Println()

	for _, section := range sections {
		// Skip empty sections and special sections
		if section.Heading == "" || section.Heading == "Notes" ||
			section.Heading == "Flights" || section.Heading == "Places to visit" {
			continue
		}

		// Destination header
		fmt.Printf("### %s\n\n", section.Heading)

		if section.Date != nil && *section.Date != "" {
			sectionDate, _ := time.Parse("2006-01-02", *section.Date)
			fmt.Printf("**Date:** %s\n\n", sectionDate.Format("Monday, January 2, 2006"))
		}

		// Show blocks for this destination
		if len(section.Blocks) > 0 {
			hasContent := false
			for _, block := range section.Blocks {
				if block.Type == "note" {
					if !block.Text.IsString && len(block.Text.Text.Ops) > 0 && block.Text.Text.Ops[0].Insert != "\n" {
						noteText := strings.TrimSpace(block.Text.Text.Ops[0].Insert)
						if noteText != "" {
							if !hasContent {
								fmt.Println("**Notes:**")
								hasContent = true
							}
							fmt.Printf("- %s\n", noteText)
						}
					}
				}
			}
			if hasContent {
				fmt.Println()
			}
		}
	}
}

func PrintPlacesMarkdown(places []wanderlog.Metadata) {
	if len(places) == 0 {
		fmt.Println("# Places")
		fmt.Println("\nNo places found in this trip.")
		return
	}

	fmt.Println("# Places")
	fmt.Println()

	for _, place := range places {
		// Place header with rating
		header := place.Name
		if place.Rating > 0 {
			header += fmt.Sprintf(" (%.1f★)", place.Rating)
		}
		fmt.Printf("## %s\n\n", header)

		// Basic information
		if place.Address != "" {
			fmt.Printf("- **Address:** %s\n", place.Address)
		}

		if len(place.Categories) > 0 {
			fmt.Printf("- **Categories:** %s\n", strings.Join(place.Categories, ", "))
		}

		if place.Rating > 0 {
			fmt.Printf("- **Rating:** %.1f/5.0\n", place.Rating)
		}

		if place.NumRatings > 0 {
			fmt.Printf("- **Number of Reviews:** %d\n", place.NumRatings)
		}

		if place.Website != "" {
			fmt.Printf("- **Website:** %s\n", place.Website)
		}

		if place.InternationalPhoneNumber != "" {
			fmt.Printf("- **Phone:** %s\n", place.InternationalPhoneNumber)
		}

		// Business status
		if place.BusinessStatus != "" {
			fmt.Printf("- **Business Status:** %s\n", place.BusinessStatus)
		}

		if place.PermanentlyClosed {
			fmt.Printf("- **⚠️ Status:** Permanently Closed\n")
		}

		// Description
		if place.Description != nil && *place.Description != "" {
			fmt.Printf("\n**Description:**\n%s\n", *place.Description)
		}

		// Generated description
		if place.GeneratedDescription != nil && *place.GeneratedDescription != "" {
			fmt.Printf("\n**Summary:**\n%s\n", *place.GeneratedDescription)
		}

		fmt.Println()
	}
}
