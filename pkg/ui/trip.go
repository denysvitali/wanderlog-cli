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
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C7C7C"))

	DateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F25D94")).
			Bold(true)

	PlaceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C82F6")).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	// Additional color styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#06B6D4")).
			Bold(true)

	LinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Underline(true)

	UrlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Underline(true)

	FlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F97316")).
			Bold(true)

	HotelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EC4899")).
			Bold(true)

	RestaurantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	CategoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6"))

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))

	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	ItalicStyle = lipgloss.NewStyle().
			Italic(true)

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	CountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6366F1")).
			Bold(true)

	UsernameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#14B8A6")).
			Bold(true)

	IdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	TimestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#78716C"))
)

func PrintJSON(data interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
}

func PrintTrip(trip *wanderlog.TripResponse, showDetails bool) {
	if !trip.Success {
		fmt.Println("❌ Failed to fetch trip data")
		return
	}

	plan := trip.TripPlan

	// Trip title and basic info
	fmt.Println(TitleStyle.Render("🌍 " + plan.Title))
	fmt.Println()

	// Trip dates
	if plan.StartDate != "" && plan.EndDate != "" {
		startDate, _ := time.Parse("2006-01-02", plan.StartDate)
		endDate, _ := time.Parse("2006-01-02", plan.EndDate)

		dateInfo := fmt.Sprintf("📅 %s → %s (%d days)",
			startDate.Format("Jan 2, 2006"),
			endDate.Format("Jan 2, 2006"),
			plan.Days)

		fmt.Println(DateStyle.Render(dateInfo))
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

	statsBox := BoxStyle.Render(strings.Join(stats, "  •  "))
	fmt.Println(statsBox)
	fmt.Println()

	if showDetails {
		printTripDetails(trip, &plan)
	} else {
		printTripSummary(&plan)
	}
}

func printTripSummary(plan *wanderlog.Plan) {
	fmt.Println(HeaderStyle.Render("📋 Quick Overview"))
	fmt.Println()

	// Basic trip information
	fmt.Printf("Duration: %d days\n", plan.Days)
	fmt.Printf("Total places: %d\n", plan.PlaceCount)
	fmt.Println()

	fmt.Println(InfoStyle.Render("💡 Use --details flag to see full itinerary"))
}

func printTripDetails(trip *wanderlog.TripResponse, plan *wanderlog.Plan) {
	fmt.Println(HeaderStyle.Render("🗓️  Detailed Itinerary"))
	fmt.Println()

	// Show flights first
	printFlights(plan.Itinerary.Sections)

	// Show sections (cities/destinations)
	printDestinations(plan.Itinerary.Sections, trip.Resources.SectionRecommendations)
}

func printFlights(sections []wanderlog.ItSections) {
	for _, section := range sections {
		if section.Heading == "Flights" && len(section.Blocks) > 0 {
			fmt.Println(HeaderStyle.Render("✈️  Flights"))
			fmt.Println()

			for _, block := range section.Blocks {
				if block.Type == "flight" && block.FlightInfo != nil {
					// Flight header
					flightInfo := fmt.Sprintf("%s %d",
						block.FlightInfo.Airline.Name,
						block.FlightInfo.Number)
					fmt.Println(PlaceStyle.Render("🛫 " + flightInfo))

					// Departure
					departTime, _ := time.Parse("2006-01-02", block.Depart.Date)
					departInfo := fmt.Sprintf("   Depart: %s %s from %s (%s)",
						departTime.Format("Jan 2"),
						block.Depart.Time,
						block.Depart.Airport.Iata,
						block.Depart.Airport.CityName)
					fmt.Println(InfoStyle.Render(departInfo))

					// Arrival
					if block.Arrive != nil {
						arriveTime, _ := time.Parse("2006-01-02", block.Arrive.Date)
						arriveInfo := fmt.Sprintf("   Arrive: %s %s at %s (%s)",
							arriveTime.Format("Jan 2"),
							block.Arrive.Time,
							block.Arrive.Airport.Iata,
							block.Arrive.Airport.CityName)
						fmt.Println(InfoStyle.Render(arriveInfo))
					}
					fmt.Println()
				}
			}
		}
	}
}

func printDestinations(sections []wanderlog.ItSections, sectionRecommendations map[string][]wanderlog.Place) {
	fmt.Println(HeaderStyle.Render("🌍 Destinations"))
	fmt.Println()

	for _, section := range sections {
		// Skip empty sections and special sections
		if section.Heading == "" || section.Heading == "Notes" ||
			section.Heading == "Flights" || section.Heading == "Places to visit" {
			continue
		}

		// Show destination header
		fmt.Println(SubHeaderStyle.Render("📍 " + section.Heading))

		if section.Date != nil && *section.Date != "" {
			sectionDate, _ := time.Parse("2006-01-02", *section.Date)
			fmt.Println(InfoStyle.Render("   " + sectionDate.Format("Monday, Jan 2, 2006")))
		}

		// Show places and notes/blocks for this destination
		hasContent := false
		if len(section.Blocks) > 0 {
			for _, block := range section.Blocks {
				switch block.Type {
				case "place":
					if block.Place != nil && block.Place.Name != "" {
						fmt.Println(InfoStyle.Render("🏢 " + block.Place.Name))
						hasContent = true
					}
				case "note":
					if !block.Text.IsString && len(block.Text.Text.Ops) > 0 && block.Text.Text.Ops[0].Insert != "\n" {
						noteText := strings.TrimSpace(block.Text.Text.Ops[0].Insert)
						if noteText != "" {
							fmt.Println(InfoStyle.Render("   📝 " + noteText))
							hasContent = true
						}
					}
				default:
					// Handle other block types if needed
				}
			}
		}
		if !hasContent {
			fmt.Println(InfoStyle.Render("   No details available"))
		}
		fmt.Println()
	}
}
