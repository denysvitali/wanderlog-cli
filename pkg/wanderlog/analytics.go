package wanderlog

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TripAnalytics is a local, deterministic summary derived from a trip plan.
// It intentionally does not make additional network calls, making it useful in
// scripts and for diagnosing itinerary density before requesting route data.
type TripAnalytics struct {
	TripKey       string          `json:"tripKey"`
	Title         string          `json:"title"`
	StartDate     string          `json:"startDate,omitempty"`
	EndDate       string          `json:"endDate,omitempty"`
	Days          int             `json:"days"`
	Sections      int             `json:"sections"`
	DatedSections int             `json:"datedSections"`
	PlaceBlocks   int             `json:"placeBlocks"`
	Notes         int             `json:"notes"`
	Flights       int             `json:"flights"`
	Lodgings      int             `json:"lodgings"`
	Transit       int             `json:"transit"`
	OtherBlocks   int             `json:"otherBlocks"`
	MetadataCount int             `json:"metadataCount"`
	DayLoads      []TripDayLoad   `json:"dayLoads"`
	Expenses      []CurrencyTotal `json:"expenses,omitempty"`
	Warnings      []string        `json:"warnings,omitempty"`
}

type TripDayLoad struct {
	SectionID int    `json:"sectionId"`
	Date      string `json:"date,omitempty"`
	Heading   string `json:"heading,omitempty"`
	Places    int    `json:"places"`
	Blocks    int    `json:"blocks"`
}

type CurrencyTotal struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
}

// AnalyzeTrip derives itinerary and budget health information from trip.
func AnalyzeTrip(trip *TripResponse) (*TripAnalytics, error) {
	if trip == nil {
		return nil, fmt.Errorf("trip is required")
	}

	plan := trip.TripPlan
	result := &TripAnalytics{
		TripKey:       firstNonEmptyString(plan.Key, plan.ViewKey),
		Title:         plan.Title,
		StartDate:     plan.StartDate,
		EndDate:       plan.EndDate,
		Days:          plan.Days,
		Sections:      len(plan.Itinerary.Sections),
		MetadataCount: len(trip.Resources.PlaceMetadata),
		DayLoads:      make([]TripDayLoad, 0, len(plan.Itinerary.Sections)),
	}

	for _, section := range plan.Itinerary.Sections {
		load := TripDayLoad{SectionID: section.ID, Heading: firstNonEmptyString(section.DisplayHeading, section.Heading), Blocks: len(section.Blocks)}
		if section.Date != nil {
			load.Date = *section.Date
			if load.Date != "" {
				result.DatedSections++
			}
		}

		for _, block := range section.Blocks {
			kind := strings.ToLower(strings.TrimSpace(block.Type))
			switch {
			case block.Place != nil || kind == "place":
				result.PlaceBlocks++
				load.Places++
			case block.FlightInfo != nil || kind == "flight":
				result.Flights++
			case block.Hotel != nil || kind == "hotel" || kind == "lodging":
				result.Lodgings++
			case block.Arrive != nil || kind == "train" || kind == "transit":
				result.Transit++
			case kind == "note" || block.NoteIcon != "":
				result.Notes++
			default:
				result.OtherBlocks++
			}
		}
		if load.Places > 6 {
			label := firstNonEmptyString(load.Date, load.Heading, fmt.Sprintf("section %d", load.SectionID))
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s has %d places and may be over-scheduled", label, load.Places))
		}
		result.DayLoads = append(result.DayLoads, load)
	}

	if plan.StartDate != "" && plan.EndDate != "" {
		start, startErr := time.Parse(apiDateFormat, plan.StartDate)
		end, endErr := time.Parse(apiDateFormat, plan.EndDate)
		switch {
		case startErr != nil || endErr != nil:
			result.Warnings = append(result.Warnings, "trip contains an invalid start or end date")
		case end.Before(start):
			result.Warnings = append(result.Warnings, "trip end date is before its start date")
		default:
			expectedDays := int(end.Sub(start).Hours()/24) + 1
			if result.Days == 0 {
				result.Days = expectedDays
			} else if result.Days != expectedDays {
				result.Warnings = append(result.Warnings, fmt.Sprintf("stored duration is %d days; dates span %d days", result.Days, expectedDays))
			}
		}
	}

	if plan.PlaceCount > 0 && result.PlaceBlocks != plan.PlaceCount {
		result.Warnings = append(result.Warnings, fmt.Sprintf("trip reports %d places but itinerary contains %d place blocks", plan.PlaceCount, result.PlaceBlocks))
	}
	if result.MetadataCount < result.PlaceBlocks {
		result.Warnings = append(result.Warnings, "some itinerary places have no corresponding metadata")
	}

	totals := map[string]*CurrencyTotal{}
	for _, expense := range plan.Itinerary.Budget.Expenses {
		currency := strings.ToUpper(strings.TrimSpace(expense.Amount.CurrencyCode))
		if currency == "" {
			currency = "UNKNOWN"
		}
		total := totals[currency]
		if total == nil {
			total = &CurrencyTotal{Currency: currency}
			totals[currency] = total
		}
		total.Amount += expense.Amount.Amount
		total.Count++
	}
	for _, total := range totals {
		result.Expenses = append(result.Expenses, *total)
	}
	sort.Slice(result.Expenses, func(i, j int) bool { return result.Expenses[i].Currency < result.Expenses[j].Currency })

	return result, nil
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
