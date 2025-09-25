package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	subHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C7C7C"))

	dateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F25D94")).
			Bold(true)

	placeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C82F6")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)
)

func PrintJSON(data interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

func PrintTrip(trip *wanderlog.TripResponse, showDetails bool) {
	if !trip.Success {
		fmt.Println("❌ Failed to fetch trip data")
		return
	}

	plan := trip.TripPlan

	// Trip title and basic info
	fmt.Println(titleStyle.Render("🌍 " + plan.Title))
	fmt.Println()

	// Trip dates
	if plan.StartDate != "" && plan.EndDate != "" {
		startDate, _ := time.Parse("2006-01-02", plan.StartDate)
		endDate, _ := time.Parse("2006-01-02", plan.EndDate)

		dateInfo := fmt.Sprintf("📅 %s → %s (%d days)",
			startDate.Format("Jan 2, 2006"),
			endDate.Format("Jan 2, 2006"),
			plan.Days)

		fmt.Println(dateStyle.Render(dateInfo))
		fmt.Println()
	}

	// Trip stats
	stats := []string{
		fmt.Sprintf("📍 %d places", plan.PlaceCount),
		fmt.Sprintf("👀 %d views", plan.ViewCount),
	}
	if plan.LikeCount > 0 {
		stats = append(stats, fmt.Sprintf("❤️  %d likes", plan.LikeCount))
	}

	statsBox := boxStyle.Render(strings.Join(stats, "  •  "))
	fmt.Println(statsBox)
	fmt.Println()

	if showDetails {
		printTripDetails(trip, &plan)
	} else {
		printTripSummary(&plan)
	}
}

func printTripSummary(plan *wanderlog.Plan) {
	fmt.Println(headerStyle.Render("📋 Quick Overview"))
	fmt.Println()

	// Basic trip information
	fmt.Printf("Duration: %d days\n", plan.Days)
	fmt.Printf("Total places: %d\n", plan.PlaceCount)
	fmt.Println()

	fmt.Println(infoStyle.Render("💡 Use --details flag to see full itinerary"))
}

func printTripDetails(trip *wanderlog.TripResponse, plan *wanderlog.Plan) {
	fmt.Println(headerStyle.Render("🗓️  Detailed Itinerary"))
	fmt.Println()

	// Show flights first
	printFlights(plan.Itinerary.Sections)

	// Show sections (cities/destinations)
	printDestinations(plan.Itinerary.Sections, trip.Resources.SectionRecommendations)
}

func printFlights(sections []wanderlog.ItSections) {
	for _, section := range sections {
		if section.Heading == "Flights" && len(section.Blocks) > 0 {
			fmt.Println(headerStyle.Render("✈️  Flights"))
			fmt.Println()
			
			for _, block := range section.Blocks {
				if block.Type == "flight" && block.FlightInfo != nil {
					// Flight header
					flightInfo := fmt.Sprintf("%s %d", 
						block.FlightInfo.Airline.Name, 
						block.FlightInfo.Number)
					fmt.Println(placeStyle.Render("🛫 " + flightInfo))
					
					// Departure
					departTime, _ := time.Parse("2006-01-02", block.Depart.Date)
					departInfo := fmt.Sprintf("   Depart: %s %s from %s (%s)",
						departTime.Format("Jan 2"),
						block.Depart.Time,
						block.Depart.Airport.Iata,
						block.Depart.Airport.CityName)
					fmt.Println(infoStyle.Render(departInfo))
					
					// Arrival
					if block.Arrive != nil {
						arriveTime, _ := time.Parse("2006-01-02", block.Arrive.Date)
						arriveInfo := fmt.Sprintf("   Arrive: %s %s at %s (%s)",
							arriveTime.Format("Jan 2"),
							block.Arrive.Time,
							block.Arrive.Airport.Iata,
							block.Arrive.Airport.CityName)
						fmt.Println(infoStyle.Render(arriveInfo))
					}
					fmt.Println()
				}
			}
		}
	}
}

func printDestinations(sections []wanderlog.ItSections, sectionRecommendations map[string][]wanderlog.Place) {
	fmt.Println(headerStyle.Render("🌍 Destinations"))
	fmt.Println()

	for _, section := range sections {
		// Skip empty sections and special sections
		if section.Heading == "" || section.Heading == "Notes" ||
		   section.Heading == "Flights" || section.Heading == "Places to visit" {
			continue
		}

		// Show destination header
		fmt.Println(subHeaderStyle.Render("📍 " + section.Heading))

		if section.Date != nil && *section.Date != "" {
			sectionDate, _ := time.Parse("2006-01-02", *section.Date)
			fmt.Println(infoStyle.Render("   " + sectionDate.Format("Monday, Jan 2, 2006")))
		}

		// Show places and notes/blocks for this destination
		hasContent := false
		if len(section.Blocks) > 0 {
			for _, block := range section.Blocks {
				switch block.Type {
				case "place":
					if block.Place != nil && block.Place.Name != "" {
						fmt.Println(infoStyle.Render("🏢 " + block.Place.Name))
						hasContent = true
					}
				case "note":
					if len(block.Text.Ops) > 0 && block.Text.Ops[0].Insert != "\n" {
						noteText := strings.TrimSpace(block.Text.Ops[0].Insert)
						if noteText != "" {
							fmt.Println(infoStyle.Render("   📝 " + noteText))
							hasContent = true
						}
					}
				default:
					// Handle other block types if needed
				}
			}
		}
		if !hasContent {
			fmt.Println(infoStyle.Render("   No details available"))
		}
		fmt.Println()
	}
}
