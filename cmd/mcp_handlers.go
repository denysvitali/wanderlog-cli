package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

// Tool handlers
func handleListTrips(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	trips, err := client.GetUserTrips()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trips: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(trips), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	var result string
	if len(trips.Data) == 0 {
		result = "No trips found."
	} else {
		result = fmt.Sprintf("%d trips:\n", len(trips.Data))
		// Limit to first 20 trips to avoid exceeding char limit
		limit := len(trips.Data)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			trip := trips.Data[i]
			result += fmt.Sprintf("%d. %s (%s) | %d places\n", i+1, trip.Title, trip.Key, trip.PlaceCount)
		}
		if len(trips.Data) > 20 {
			result += fmt.Sprintf("\n... and %d more (use JSON format for full list)\n", len(trips.Data)-20)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func handleGetTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_id", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_id is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(trip), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	plan := trip.TripPlan
	result := fmt.Sprintf("%s (Key: %s)\n", plan.Title, plan.Key)
	if plan.StartDate != "" && plan.EndDate != "" {
		result += fmt.Sprintf("📅 %s to %s (%d days)\n", plan.StartDate, plan.EndDate, plan.Days)
	}
	result += fmt.Sprintf("📍 %d places | 👁 %d views | ❤ %d likes\n", plan.PlaceCount, plan.ViewCount, plan.LikeCount)
	if plan.AuthorBlurb != "" {
		// Truncate description if too long
		desc := plan.AuthorBlurb
		if len(desc) > 200 {
			desc = desc[:197] + "..."
		}
		result += fmt.Sprintf("📝 %s\n", desc)
	}

	return mcp.NewToolResultText(result), nil
}

func handleListPlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_id", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_id is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	// Use GetTrip instead of GetTripPlaces to get the full place metadata
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip: %v", err)), nil
	}

	places := trip.Resources.PlaceMetadata

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(map[string]any{
			"success": true,
			"data": map[string]any{
				"trip_key": tripKey,
				"places":   places,
			},
		}), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	var result string
	if len(places) == 0 {
		result = fmt.Sprintf("No places in trip %s", tripKey)
	} else {
		result = fmt.Sprintf("%s: %d places\n", trip.TripPlan.Title, len(places))

		// Limit to first 15 places to avoid exceeding char limit
		limit := len(places)
		if limit > 15 {
			limit = 15
		}

		for i := 0; i < limit; i++ {
			place := places[i]
			// Compact format: name + rating + address
			name := place.Name
			if place.Rating > 0 {
				name += fmt.Sprintf(" (%.1f⭐)", place.Rating)
			}
			result += fmt.Sprintf("%d. %s\n", i+1, name)

			// Only show address if available (most important info)
			if place.Address != "" {
				// Truncate long addresses
				addr := place.Address
				if len(addr) > 50 {
					addr = addr[:47] + "..."
				}
				result += fmt.Sprintf("   %s\n", addr)
			}
		}
		if len(places) > 15 {
			result += fmt.Sprintf("\n... and %d more (use JSON format for full list)\n", len(places)-15)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func handleListSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_id", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_id is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip: %v", err)), nil
	}

	sections := trip.TripPlan.Itinerary.Sections

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(sections), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	var result string
	if len(sections) == 0 {
		result = fmt.Sprintf("No sections in trip %s", tripKey)
	} else {
		result = fmt.Sprintf("%s: %d sections\n", trip.TripPlan.Title, len(sections))

		for i, section := range sections {
			// Compact format: heading + date + ID + item count
			heading := section.Heading
			if heading == "" {
				heading = "Untitled"
			}
			result += fmt.Sprintf("%d. %s [ID:%d]", i+1, heading, section.ID)

			if section.Date != nil && *section.Date != "" {
				result += fmt.Sprintf(" (%s)", *section.Date)
			}

			if len(section.Blocks) > 0 {
				result += fmt.Sprintf(" - %d items", len(section.Blocks))
			}
			result += "\n"
		}

		result += "\n💡 Use section IDs with add_place tool\n"
	}

	return mcp.NewToolResultText(result), nil
}

func handleGetFlights(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := resolveTripKey(ctx, request, "trip_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	flights, err := client.GetTripFlights(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get flights for trip %s: %v", tripKey, err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(flights), nil
	}

	if len(flights.Data.Flights) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No flights found in trip %s", tripKey)), nil
	}

	result := fmt.Sprintf("✈️ **%d flight(s) in trip %s:**\n\n", len(flights.Data.Flights), tripKey)
	for i, f := range flights.Data.Flights {
		flightStr := fmt.Sprintf("%s%s", f.AirlineIATA, f.FlightNumber)
		result += fmt.Sprintf("%d. **%s**", i+1, flightStr)
		if f.DepartureDate != "" {
			result += fmt.Sprintf(" — %s", f.DepartureDate)
		}
		if f.SectionDate != "" {
			result += fmt.Sprintf(" (section: %s)", f.SectionDate)
		}
		result += "\n"
		if f.Origin.IATA != "" && f.Destination.IATA != "" {
			result += fmt.Sprintf("   %s → %s", f.Origin.IATA, f.Destination.IATA)
		}
		if f.DepartureTime != "" || f.ArrivalTime != "" {
			if f.Origin.IATA != "" || f.Destination.IATA != "" {
				result += " | "
			}
			if f.DepartureTime != "" {
				result += fmt.Sprintf("Depart: %s", f.DepartureTime)
			}
			if f.ArrivalTime != "" {
				if f.DepartureTime != "" {
					result += ", "
				}
				result += fmt.Sprintf("Arrive: %s", f.ArrivalTime)
			}
		}
		result += "\n\n"
	}

	return mcp.NewToolResultText(result), nil
}

func handleGetTripSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	sections, err := client.GetTripSections(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip sections: %v", err)), nil
	}

	// Return structured data directly - more efficient than text formatting
	return mcp.NewToolResultStructuredOnly(sections), nil
}

func optionalIntArgument(request mcp.CallToolRequest, key string) (int, bool) {
	args := request.GetArguments()
	if args == nil {
		return 0, false
	}
	if _, ok := args[key]; !ok {
		return 0, false
	}
	return request.GetInt(key, 0), true
}

func validateDateArgument(name, value string, required bool) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return "", fmt.Errorf("%s is required in YYYY-MM-DD format", name)
		}
		return "", nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", name)
	}
	return value, nil
}

func validateDateRange(startName, startValue, endName, endValue string, required bool) (string, string, error) {
	startValue, err := validateDateArgument(startName, startValue, required)
	if err != nil {
		return "", "", err
	}
	endValue, err = validateDateArgument(endName, endValue, required)
	if err != nil {
		return "", "", err
	}
	if startValue == "" || endValue == "" {
		return startValue, endValue, nil
	}
	start, _ := time.Parse("2006-01-02", startValue)
	end, _ := time.Parse("2006-01-02", endValue)
	if end.Before(start) {
		return "", "", fmt.Errorf("%s must be on or after %s", endName, startName)
	}
	return startValue, endValue, nil
}

func resolveTripKey(ctx context.Context, request mcp.CallToolRequest, argName string) (string, error) {
	tripKey := request.GetString(argName, "")
	if tripKey != "" {
		return tripKey, nil
	}
	if defaultTripID, ok := tripIDFromContext(ctx); ok {
		return defaultTripID, nil
	}
	return "", fmt.Errorf("%s is required (either as parameter or default trip key must be set)", argName)
}

func availableSectionDates(sections []wanderlog.ItSections) string {
	dates := make([]string, 0, len(sections))
	for _, section := range sections {
		if section.Date != nil && strings.TrimSpace(*section.Date) != "" {
			dates = append(dates, fmt.Sprintf("%s:%d", strings.TrimSpace(*section.Date), section.ID))
		}
	}
	if len(dates) == 0 {
		return "none"
	}
	return strings.Join(dates, ", ")
}

func loadSectionsForResolution(client *wanderlog.Client, tripKey string) ([]wanderlog.ItSections, error) {
	// First try GetTrip which includes full section data
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		// GetTrip failed, try the dedicated sections endpoint as fallback
		sections, secErr := client.GetTripSections(tripKey)
		if secErr != nil {
			return nil, fmt.Errorf("GetTrip failed: %w, GetTripSections also failed: %w", err, secErr)
		}
		return sections, nil
	}

	// GetTrip succeeded - check if it has sections
	if len(trip.TripPlan.Itinerary.Sections) > 0 {
		return trip.TripPlan.Itinerary.Sections, nil
	}

	// GetTrip succeeded but returned 0 sections - try dedicated endpoint
	// This handles cases where the trip exists but the full trip response
	// has an empty sections array while the sections endpoint returns data
	sections, err := client.GetTripSections(tripKey)
	if err != nil {
		return nil, fmt.Errorf("GetTrip returned no sections and GetTripSections failed: %w", err)
	}
	return sections, nil
}

func resolveSectionFromList(sections []wanderlog.ItSections, sectionID int, hasSectionID bool, sectionDate string, requireDated bool) (int, string, error) {
	if hasSectionID {
		for _, section := range sections {
			if section.ID != sectionID {
				continue
			}
			if requireDated && (section.Date == nil || strings.TrimSpace(*section.Date) == "") {
				return 0, "", fmt.Errorf("section_id %d is not a dated itinerary section; use list_sections and choose a section with a date", sectionID)
			}
			if section.Date != nil && strings.TrimSpace(*section.Date) != "" {
				return section.ID, fmt.Sprintf("%s section ID %d", strings.TrimSpace(*section.Date), section.ID), nil
			}
			return section.ID, fmt.Sprintf("section ID %d", section.ID), nil
		}
		return 0, "", fmt.Errorf("section_id %d was not found in this trip; available dated sections: %s", sectionID, availableSectionDates(sections))
	}

	for _, section := range sections {
		if section.Date != nil && strings.TrimSpace(*section.Date) == sectionDate {
			return section.ID, fmt.Sprintf("%s section ID %d", sectionDate, section.ID), nil
		}
	}

	return 0, "", fmt.Errorf("no itinerary section found for date %s; available dated sections: %s", sectionDate, availableSectionDates(sections))
}

func resolveAddPlaceSectionID(client *wanderlog.Client, tripKey string, request mcp.CallToolRequest) (int, string, error) {
	unscheduled := request.GetBool("unscheduled", false)
	sectionDate, err := validateDateArgument("section_date", request.GetString("section_date", ""), false)
	if err != nil {
		return 0, "", err
	}
	sectionID, hasSectionID := optionalIntArgument(request, "section_id")

	if unscheduled {
		if sectionDate != "" || (hasSectionID && sectionID > 0) {
			return 0, "", fmt.Errorf("use either unscheduled=true or a dated section, not both")
		}
		return 0, "general Places to visit list", nil
	}

	if hasSectionID {
		if sectionDate != "" {
			return 0, "", fmt.Errorf("use either section_id or section_date, not both")
		}
		if sectionID <= 0 {
			return 0, "", fmt.Errorf("section_id must be a positive itinerary section ID. Use list_sections or section_date; pass unscheduled=true to add to the general Places to visit list")
		}
	} else if sectionDate == "" {
		return 0, "", fmt.Errorf("section_id or section_date is required. Use list_sections to find dated sections, or pass unscheduled=true to add to the general Places to visit list")
	}

	sections, err := loadSectionsForResolution(client, tripKey)
	if err != nil {
		return 0, "", fmt.Errorf("failed to resolve section: %w", err)
	}
	return resolveSectionFromList(sections, sectionID, hasSectionID, sectionDate, false)
}

func currentUserID() int {
	auth, err := wanderlog.LoadCredentials()
	if err != nil || auth.UserID == "" {
		return 0
	}
	userID, err := strconv.Atoi(auth.UserID)
	if err != nil {
		return 0
	}
	return userID
}

func placeMatches(name, placeID, candidateName, candidatePlaceID string) bool {
	if placeID != "" && (candidatePlaceID == placeID || strings.EqualFold(candidatePlaceID, placeID)) {
		return true
	}
	return name != "" && strings.EqualFold(strings.TrimSpace(candidateName), strings.TrimSpace(name))
}

func tripHasAddedPlace(trip *wanderlog.TripResponse, sectionID int, name, placeID string) bool {
	if trip == nil {
		return false
	}

	if sectionID > 0 {
		sectionIdx := wanderlog.FindSectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
		if sectionIdx < 0 {
			return false
		}
		for _, block := range trip.TripPlan.Itinerary.Sections[sectionIdx].Blocks {
			if block.Place == nil {
				continue
			}
			if placeMatches(name, placeID, block.Place.Name, block.Place.PlaceID) {
				return true
			}
		}
		return false
	}

	for _, place := range trip.Resources.PlaceMetadata {
		if placeMatches(name, placeID, place.Name, place.PlaceID) {
			return true
		}
	}
	return false
}

func verifyAddedPlacePersisted(client *wanderlog.Client, tripKey string, sectionID int, name, placeID string) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 250 * time.Millisecond)
		}
		trip, err := client.GetTrip(tripKey)
		if err != nil {
			lastErr = err
			continue
		}
		if tripHasAddedPlace(trip, sectionID, name, placeID) {
			return nil
		}
	}
	if lastErr != nil {
		return fmt.Errorf("write request completed, but verification failed while reloading trip: %w", lastErr)
	}
	return fmt.Errorf("write request completed, but the place was not found when the trip was reloaded")
}

func tripHasAddedFlight(trip *wanderlog.TripResponse, sectionID int, airlineIATA string, flightNumber int, departureDate string) bool {
	if trip == nil {
		return false
	}
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if sectionID > 0 && section.ID != sectionID {
			continue
		}
		for _, block := range section.Blocks {
			if block.FlightInfo == nil {
				continue
			}
			if strings.EqualFold(block.FlightInfo.Airline.Iata, airlineIATA) &&
				block.FlightInfo.Number == flightNumber &&
				block.Depart.Date == departureDate {
				return true
			}
		}
	}
	return false
}

func findFlightsSectionID(trip *wanderlog.TripResponse) int {
	if trip == nil {
		return 0
	}
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.Type == "flights" {
			return section.ID
		}
	}
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if strings.EqualFold(section.Heading, "Flights") || section.PlaceMarkerIcon == "plane" {
			return section.ID
		}
	}
	return 0
}

func maxItineraryItemID(trip *wanderlog.TripResponse) int {
	maxID := 0
	if trip == nil {
		return maxID
	}
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.ID > maxID {
			maxID = section.ID
		}
		for _, block := range section.Blocks {
			if block.ID > maxID {
				maxID = block.ID
			}
		}
	}
	return maxID
}

func createFlightsSection(client *wanderlog.Client, tripKey string, trip *wanderlog.TripResponse) (int, error) {
	if trip == nil {
		return 0, fmt.Errorf("trip response is nil")
	}
	sectionID := maxItineraryItemID(trip) + 1
	section := map[string]any{
		"id":               sectionID,
		"heading":          "Flights",
		"displayHeading":   "",
		"date":             nil,
		"type":             "flights",
		"mode":             "placeList",
		"placeMarkerColor": "#3498db",
		"placeMarkerIcon":  "plane",
		"text": map[string]any{
			"ops": []any{map[string]any{"insert": "\n"}},
		},
		"blocks": []any{},
	}
	position := len(trip.TripPlan.Itinerary.Sections)
	op := wanderlog.InsertInList([]interface{}{"itinerary", "sections"}, position, section)
	if err := client.ApplyOperations(tripKey, []wanderlog.Operation{op}); err != nil {
		return 0, err
	}
	return sectionID, nil
}

func ensureFlightsSectionID(client *wanderlog.Client, tripKey string) (int, string, error) {
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return 0, "", fmt.Errorf("getting current trip: %w", err)
	}
	if sectionID := findFlightsSectionID(trip); sectionID > 0 {
		return sectionID, fmt.Sprintf("Flights section ID %d", sectionID), nil
	}
	sectionID, err := createFlightsSection(client, tripKey, trip)
	if err != nil {
		return 0, "", fmt.Errorf("creating Flights section: %w", err)
	}
	return sectionID, fmt.Sprintf("Flights section ID %d", sectionID), nil
}

func verifyAddedFlightPersisted(client *wanderlog.Client, tripKey string, sectionID int, airlineIATA string, flightNumber int, departureDate string) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 250 * time.Millisecond)
		}
		trip, err := client.GetTrip(tripKey)
		if err != nil {
			lastErr = err
			continue
		}
		if tripHasAddedFlight(trip, sectionID, airlineIATA, flightNumber, departureDate) {
			return nil
		}
	}
	if lastErr != nil {
		return fmt.Errorf("write request completed, but verification failed while reloading trip: %w", lastErr)
	}
	return fmt.Errorf("write request completed, but %s%d on %s was not found when the trip was reloaded", airlineIATA, flightNumber, departureDate)
}

func appendItineraryBlock(client *wanderlog.Client, tripKey string, sectionID int, block map[string]any) error {
	if err := validateBlockSchema(block); err != nil {
		return fmt.Errorf("invalid block schema: %w", err)
	}
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	sectionIdx := wanderlog.FindSectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
	if sectionIdx < 0 {
		return fmt.Errorf("section %d not found", sectionID)
	}

	maxBlockID := 0
	for _, section := range trip.TripPlan.Itinerary.Sections {
		for _, block := range section.Blocks {
			if block.ID > maxBlockID {
				maxBlockID = block.ID
			}
		}
	}
	if _, ok := block["id"]; !ok {
		block["id"] = maxBlockID + 1
	}

	if _, ok := block["addedBy"]; !ok {
		addedBy := map[string]any{"type": "user"}
		if userID := currentUserID(); userID > 0 {
			addedBy["userId"] = userID
		}
		block["addedBy"] = addedBy
	}
	if _, ok := block["attachments"]; !ok {
		block["attachments"] = []any{}
	}
	if _, ok := block["upvotedBy"]; !ok {
		block["upvotedBy"] = []any{}
	}

	position := len(trip.TripPlan.Itinerary.Sections[sectionIdx].Blocks)
	op := wanderlog.InsertInList(
		[]interface{}{"itinerary", "sections", sectionIdx, "blocks"},
		position,
		block,
	)
	return client.ApplyOperations(tripKey, []wanderlog.Operation{op})
}

func splitFlightNumber(flightNumber string) (string, int) {
	flightNumber = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(flightNumber), " ", ""))
	prefixEnd := 0
	for prefixEnd < len(flightNumber) && (flightNumber[prefixEnd] < '0' || flightNumber[prefixEnd] > '9') {
		prefixEnd++
	}
	number, _ := strconv.Atoi(flightNumber[prefixEnd:])
	return flightNumber[:prefixEnd], number
}

func validateFlightNumber(flightNumber string) (string, int, error) {
	airline, number := splitFlightNumber(flightNumber)
	if airline == "" || number <= 0 {
		return "", 0, fmt.Errorf("flight_number must include an airline code and number, e.g. MU244")
	}
	return airline, number, nil
}

// getGooglePlaceForAirport searches for an airport by name and IATA code,
// then fetches detailed place information to construct a googlePlace object.
func getGooglePlaceForAirport(client *wanderlog.Client, airportName, iataCode, cityName string) map[string]any {
	// Search for the airport with name + IATA code
	query := fmt.Sprintf("%s Airport %s", airportName, iataCode)
	if cityName != "" {
		query = fmt.Sprintf("%s %s airport", cityName, airportName)
	}

	results, err := client.SearchPlaces(query, nil, nil)
	if err != nil || !results.Success || len(results.Places) == 0 {
		logger.WithFields(logrus.Fields{
			"query": query,
			"error": err,
		}).Warn("Failed to search for airport place details")
		return nil
	}

	// Find the best matching result (prefer exact IATA match in description)
	var bestMatch *wanderlog.SearchResult
	for i, place := range results.Places {
		// Look for IATA code in the address or description
		if strings.Contains(strings.ToUpper(place.Address), iataCode) ||
			strings.Contains(strings.ToUpper(place.Name), iataCode) {
			bestMatch = &results.Places[i]
			break
		}
		// Use first result if no exact match
		if i == 0 && bestMatch == nil {
			bestMatch = &results.Places[i]
		}
	}

	if bestMatch == nil {
		return nil
	}

	// Fetch detailed place information
	details, err := client.GetPlaceDetails(bestMatch.PlaceID)
	if err != nil || !details.Success {
		logger.WithFields(logrus.Fields{
			"placeID": bestMatch.PlaceID,
			"error":   err,
		}).Warn("Failed to get place details for airport")
		return nil
	}

	// Construct a googlePlace-like object with the data we have
	googlePlace := map[string]any{
		"place_id":           details.Data.Details.PlaceID,
		"name":               details.Data.Details.Name,
		"formatted_address":  details.Data.Details.FormattedAddress,
		"rating":             details.Data.Details.Rating,
		"user_ratings_total": details.Data.Details.UserRatingsTotal,
		"types":              details.Data.Details.Types,
		"business_status":    details.Data.Details.BusinessStatus,
		"url":                fmt.Sprintf("https://maps.google.com/?cid=%s", details.Data.Details.PlaceID),
		"geometry": map[string]any{
			"location": map[string]float64{
				"lat": details.Data.Details.Geometry.Location.Lat,
				"lng": details.Data.Details.Geometry.Location.Lng,
			},
		},
	}

	return googlePlace
}

// validateBlockSchema ensures block structures are compatible with the web/mobile app.
// Flight blocks must not set depart.type="airport" or arrive.type="airport" without
// providing a fully populated airport sub-object (with googlePlace), otherwise the
// web app's station parser will crash trying to access undefined fields.
func validateBlockSchema(block map[string]any) error {
	if block["type"] != "flight" {
		return nil
	}

	depart, hasDepart := block["depart"].(map[string]any)
	arrive, hasArrive := block["arrive"].(map[string]any)

	validateStation := func(name string, station map[string]any) error {
		stationType, hasType := station["type"].(string)
		if !hasType || stationType != "airport" {
			return nil // no airport type, no validation needed
		}
		airport, hasAirport := station["airport"].(map[string]any)
		if !hasAirport {
			return fmt.Errorf("flight block %s has type=airport but no airport sub-object; set type=airport only when providing a complete airport object with googlePlace", name)
		}
		if _, hasGooglePlace := airport["googlePlace"]; !hasGooglePlace {
			return fmt.Errorf("flight block %s has type=airport but airport.googlePlace is missing; either provide the complete airport object with googlePlace or omit type=airport entirely", name)
		}
		return nil
	}

	if hasDepart {
		if err := validateStation("depart", depart); err != nil {
			return err
		}
	}
	if hasArrive {
		if err := validateStation("arrive", arrive); err != nil {
			return err
		}
	}
	return nil
}

func handleAddPlace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	name, err := request.RequireString("name")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("name is required"), nil //nolint:nilerr
	}

	placeID := request.GetString("place_id", "")
	latitude := request.GetFloat("latitude", 0)
	longitude := request.GetFloat("longitude", 0)
	text := request.GetString("text", "")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	sectionID, sectionLabel, err := resolveAddPlaceSectionID(client, tripKey, request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// CRITICAL FIX: If place_id is provided but coordinates are missing, fetch them first
	// This prevents creating places without location data, which breaks trips with:
	// "TypeError: Cannot read properties of undefined (reading 'location')"
	if placeID != "" && (latitude == 0 && longitude == 0) {
		logger.WithField("place_id", placeID).Debug("Fetching place details to get coordinates")

		placeDetails, err := client.GetPlaceDetails(placeID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch place details for coordinates: %v. Please provide latitude and longitude parameters.", err)), nil
		}

		// Extract coordinates from place details
		latitude = placeDetails.Data.Details.Geometry.Location.Lat
		longitude = placeDetails.Data.Details.Geometry.Location.Lng

		// Also use the canonical name from the API if user didn't override it
		if name == "" || name == placeID {
			name = placeDetails.Data.Details.Name
		}

		logger.WithFields(map[string]interface{}{
			"place_id": placeID,
			"name":     name,
			"lat":      latitude,
			"lng":      longitude,
		}).Debug("Fetched coordinates from place details")
	}

	// Build the place info with proper geometry structure
	placeInfo := wanderlog.AddPlaceInfo{
		PlaceID: placeID,
		Name:    name,
	}

	// CRITICAL: Always require coordinates when place_id is provided
	// The Wanderlog API technically accepts places without geometry, but they become
	// corrupt and break the trip with "Cannot read properties of undefined (reading 'location')"
	if placeID != "" {
		if latitude == 0 && longitude == 0 {
			return mcp.NewToolResultError("Coordinates (latitude/longitude) are required when adding a place. This should not happen - please report this bug."), nil
		}

		placeInfo.Geometry = &models.PlaceGeometry{
			Location: models.PlaceLocation{
				Lat: latitude,
				Lng: longitude,
			},
		}
	} else if latitude != 0 || longitude != 0 {
		// Only add geometry if coordinates are provided for places without place_id
		placeInfo.Geometry = &models.PlaceGeometry{
			Location: models.PlaceLocation{
				Lat: latitude,
				Lng: longitude,
			},
		}
	}

	req := wanderlog.AddPlaceRequest{
		Place: placeInfo,
		Text:  text,
	}

	err = client.AddPlace(tripKey, sectionID, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add place '%s' to trip %s (section %s): %v", name, tripKey, sectionLabel, err)), nil
	}
	if err := verifyAddedPlacePersisted(client, tripKey, sectionID, name, placeID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add place '%s' to trip %s (section %s): %v", name, tripKey, sectionLabel, err)), nil
	}

	result := fmt.Sprintf("📍 Successfully added place '%s' to trip %s (%s)", name, tripKey, sectionLabel)

	return mcp.NewToolResultText(result), nil
}

func handleAddFlight(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := resolveTripKey(ctx, request, "trip_key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	flightNumber, err := request.RequireString("flight_number")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("flight_number is required"), nil //nolint:nilerr
	}

	departureDate, err := validateDateArgument("departure_date", request.GetString("departure_date", ""), true)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	departureTime := request.GetString("departure_time", "")
	arrivalDate, err := validateDateArgument("arrival_date", request.GetString("arrival_date", ""), false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if arrivalDate == "" {
		arrivalDate = departureDate
	}
	arrivalTime := request.GetString("arrival_time", "")
	confirmationNumber := request.GetString("confirmation_number", "")
	notes := request.GetString("notes", "")
	userDepartAirport := strings.ToUpper(strings.TrimSpace(request.GetString("departure_airport", "")))
	userArriveAirport := strings.ToUpper(strings.TrimSpace(request.GetString("arrival_airport", "")))

	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	sectionID, sectionLabel, err := ensureFlightsSectionID(client, tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve Flights section for trip %s: %v", tripKey, err)), nil
	}

	airlineIATA, flightNum, err := validateFlightNumber(flightNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Fetch flight stops to get airport information for proper UI display
	var departAirport, arriveAirport map[string]any
	flightStops, err := client.GetFlightStops(strconv.Itoa(flightNum), airlineIATA, departureDate)
	if err == nil && len(flightStops.Data) > 0 {
		firstLeg := flightStops.Data[0]
		if firstLeg.Depart.Airport.IATA != "" {
			departAirport = map[string]any{
				"cityName": firstLeg.Depart.Airport.CityName,
				"iata":     firstLeg.Depart.Airport.IATA,
				"name":     firstLeg.Depart.Airport.Name,
			}
		}
		if firstLeg.Arrive.Airport.IATA != "" {
			arriveAirport = map[string]any{
				"cityName": firstLeg.Arrive.Airport.CityName,
				"iata":     firstLeg.Arrive.Airport.IATA,
				"name":     firstLeg.Arrive.Airport.Name,
			}
		}
		// Override arrival date/time from API if not explicitly provided
		if arrivalDate == departureDate && firstLeg.Arrive.Date != "" {
			arrivalDate = firstLeg.Arrive.Date
		}
		if arrivalTime == "" && firstLeg.Arrive.Time != "" {
			arrivalTime = firstLeg.Arrive.Time
		}
	}

	// Apply user-provided airport IATA overrides. These take precedence
	// over whatever the flight stops API resolved, and also handle the
	// case where the API returned no data.
	if userDepartAirport != "" {
		if departAirport == nil {
			departAirport = map[string]any{"iata": userDepartAirport}
		} else {
			departAirport["iata"] = userDepartAirport
		}
	}
	if userArriveAirport != "" {
		if arriveAirport == nil {
			arriveAirport = map[string]any{"iata": userArriveAirport}
		} else {
			arriveAirport["iata"] = userArriveAirport
		}
	}

	block := map[string]any{
		"type":               "flight",
		"confirmationNumber": confirmationNumber,
		"startTime":          departureTime,
		"endTime":            arrivalTime,
		"depart": map[string]any{
			"type": "depart",
			"date": departureDate,
			"time": departureTime,
		},
		"arrive": map[string]any{
			"type": "arrive",
			"date": arrivalDate,
			"time": arrivalTime,
		},
		"flightInfo": map[string]any{
			"airline": map[string]any{
				"iata": airlineIATA,
			},
			"number": flightNum,
		},
		"text": map[string]any{
			"ops": []any{
				map[string]any{"insert": notes + "\n"},
			},
		},
		"travelerNames": []any{},
	}

	// Add airport info if we fetched it from the API
	if departAirport != nil {
		block["depart"].(map[string]any)["airport"] = departAirport
		// Try to fetch googlePlace data for the departure airport
		if googlePlace := getGooglePlaceForAirport(client, departAirport["name"].(string), departAirport["iata"].(string), departAirport["cityName"].(string)); googlePlace != nil {
			departAirport["googlePlace"] = googlePlace
		}
	}
	if arriveAirport != nil {
		block["arrive"].(map[string]any)["airport"] = arriveAirport
		// Try to fetch googlePlace data for the arrival airport
		if googlePlace := getGooglePlaceForAirport(client, arriveAirport["name"].(string), arriveAirport["iata"].(string), arriveAirport["cityName"].(string)); googlePlace != nil {
			arriveAirport["googlePlace"] = googlePlace
		}
	}

	if err := appendItineraryBlock(client, tripKey, sectionID, block); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add flight %s to trip %s (%s): %v", flightNumber, tripKey, sectionLabel, err)), nil
	}
	if err := verifyAddedFlightPersisted(client, tripKey, sectionID, airlineIATA, flightNum, departureDate); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add flight %s to trip %s (%s): %v", flightNumber, tripKey, sectionLabel, err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("✈️ Successfully added flight %s to trip %s (%s)", flightNumber, tripKey, sectionLabel)), nil
}

func handleRemovePlace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	placeID, err := request.RequireInt("place_id")
	if err != nil {
		return mcp.NewToolResultError("place_id is required"), nil //nolint:nilerr
	}

	sectionID := request.GetInt("section_id", 0)

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err = client.RemovePlace(tripKey, sectionID, placeID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove place %d from trip %s (section %d): %v", placeID, tripKey, sectionID, err)), nil
	}

	result := fmt.Sprintf("🗑️ Successfully removed place %d from trip %s", placeID, tripKey)
	if sectionID > 0 {
		result += fmt.Sprintf(" (Section ID: %d)", sectionID)
	}

	return mcp.NewToolResultText(result), nil
}

// Resource handler
func handleTripResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract trip_id from URI like "wanderlog://trips/abc123"
	uri := request.Params.URI
	tripKey := ""

	// Simple parsing - in production you might want more robust URI parsing
	prefix := "wanderlog://trips/"
	if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
		tripKey = uri[len(prefix):]
	}

	if tripKey == "" {
		return nil, fmt.Errorf("invalid trip URI format")
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	jsonData, err := json.Marshal(trip)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trip data: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleSearchPlaces searches for places using Wanderlog's autocomplete API
func handleSearchPlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil //nolint:nilerr
	}

	format := request.GetString("format", "default")

	// Parse optional coordinates
	var lat, lng *float64
	if latStr := request.GetString("latitude", ""); latStr != "" {
		if latNum, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = &latNum
		}
	}
	if lngStr := request.GetString("longitude", ""); lngStr != "" {
		if lngNum, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = &lngNum
		}
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	results, err := client.SearchPlaces(query, lat, lng)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	// Filter out junk results (empty place_id or empty name) from autocomplete
	validIdx := 0
	for _, p := range results.Places {
		if p.PlaceID != "" && p.Name != "" {
			results.Places[validIdx] = p
			validIdx++
		}
	}
	results.Places = results.Places[:validIdx]

	switch format {
	case "json":
		jsonData, err := json.Marshal(results)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(jsonData)), nil

	case "markdown":
		if len(results.Places) == 0 {
			return mcp.NewToolResultText("# Search Results\n\nNo places found."), nil
		}

		result := "# Search Results\n\n"
		for i, place := range results.Places {
			result += fmt.Sprintf("## %d. %s\n\n", i+1, place.Name)

			if place.Rating > 0 {
				result += fmt.Sprintf("**Rating:** %.1f/5 ⭐\n\n", place.Rating)
			}

			if place.Address != "" {
				result += fmt.Sprintf("**Address:** %s\n\n", place.Address)
			}

			if len(place.Categories) > 0 {
				categories := ""
				for j, cat := range place.Categories {
					categories += cat
					if j < len(place.Categories)-1 {
						categories += ", "
					}
				}
				result += fmt.Sprintf("**Categories:** %s\n\n", categories)
			}

			if place.Description != "" {
				result += fmt.Sprintf("**Description:** %s\n\n", place.Description)
			}

			if place.Website != "" {
				result += fmt.Sprintf("**Website:** %s\n\n", place.Website)
			}

			if place.Latitude != 0 && place.Longitude != 0 {
				result += fmt.Sprintf("**Location:** %.4f, %.4f\n\n", place.Latitude, place.Longitude)
			}

			if place.PlaceID != "" {
				result += fmt.Sprintf("**Place ID:** %s\n\n", place.PlaceID)
			}

			result += "---\n\n"
		}
		return mcp.NewToolResultText(result), nil

	default:
		if len(results.Places) == 0 {
			return mcp.NewToolResultText("🔍 No places found"), nil
		}

		result := "🔍 Search Results\n\n"
		for i, place := range results.Places {
			// Place name with rating
			name := place.Name
			if place.Rating > 0 {
				stars := ""
				for j := 0; j < int(place.Rating) && j < 5; j++ {
					stars += "⭐"
				}
				name += fmt.Sprintf(" %s (%.1f)", stars, place.Rating)
			}

			result += fmt.Sprintf("📍 %s\n", name)

			// Address
			if place.Address != "" {
				result += fmt.Sprintf("   🏠 %s\n", place.Address)
			}

			// Categories
			if len(place.Categories) > 0 {
				categories := ""
				for j, cat := range place.Categories {
					categories += cat
					if j < len(place.Categories)-1 {
						categories += ", "
					}
				}
				result += fmt.Sprintf("   🏷️  %s\n", categories)
			}

			// Description
			if place.Description != "" {
				result += fmt.Sprintf("   📝 %s\n", place.Description)
			}

			// Website
			if place.Website != "" {
				result += fmt.Sprintf("   🌐 %s\n", place.Website)
			}

			// Coordinates
			if place.Latitude != 0 && place.Longitude != 0 {
				result += fmt.Sprintf("   🗺️  %.4f, %.4f\n", place.Latitude, place.Longitude)
			}

			// Place ID
			if place.PlaceID != "" {
				result += fmt.Sprintf("   🆔 %s\n", place.PlaceID)
			}

			if i < len(results.Places)-1 {
				result += "\n"
			}
		}
		return mcp.NewToolResultText(result), nil
	}
}

// handleSearchRestaurants searches for restaurants using Wanderlog's autocomplete API
func handleSearchRestaurants(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil //nolint:nilerr
	}

	format := request.GetString("format", "default")

	var lat, lng *float64
	if latStr := request.GetString("latitude", ""); latStr != "" {
		if latNum, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = &latNum
		}
	}
	if lngStr := request.GetString("longitude", ""); lngStr != "" {
		if lngNum, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = &lngNum
		}
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	results, err := client.SearchRestaurants(query, lat, lng)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Restaurant search failed: %v", err)), nil
	}

	switch format {
	case "json":
		jsonData, err := json.Marshal(results)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(jsonData)), nil

	case "markdown":
		if len(results.Places) == 0 {
			return mcp.NewToolResultText("# Restaurant Search Results\n\nNo restaurants found."), nil
		}

		result := "# Restaurant Search Results\n\n"
		for i, place := range results.Places {
			result += fmt.Sprintf("## %d. %s\n\n", i+1, place.Name)

			if place.Rating > 0 {
				result += fmt.Sprintf("**Rating:** %.1f/5 ⭐\n\n", place.Rating)
			}

			if place.Address != "" {
				result += fmt.Sprintf("**Address:** %s\n\n", place.Address)
			}

			if len(place.Categories) > 0 {
				categories := ""
				for j, cat := range place.Categories {
					categories += cat
					if j < len(place.Categories)-1 {
						categories += ", "
					}
				}
				result += fmt.Sprintf("**Categories:** %s\n\n", categories)
			}

			if place.Description != "" {
				result += fmt.Sprintf("**Description:** %s\n\n", place.Description)
			}

			if place.Website != "" {
				result += fmt.Sprintf("**Website:** %s\n\n", place.Website)
			}

			if place.Latitude != 0 && place.Longitude != 0 {
				result += fmt.Sprintf("**Location:** %.4f, %.4f\n\n", place.Latitude, place.Longitude)
			}

			if place.PlaceID != "" {
				result += fmt.Sprintf("**Place ID:** %s\n\n", place.PlaceID)
			}

			result += "---\n\n"
		}
		return mcp.NewToolResultText(result), nil

	default:
		if len(results.Places) == 0 {
			return mcp.NewToolResultText("🍽️ No restaurants found"), nil
		}

		result := "🍽️ Restaurant Search Results\n\n"
		for i, place := range results.Places {
			name := place.Name
			if place.Rating > 0 {
				stars := ""
				for j := 0; j < int(place.Rating) && j < 5; j++ {
					stars += "⭐"
				}
				name += fmt.Sprintf(" %s (%.1f)", stars, place.Rating)
			}

			result += fmt.Sprintf("📍 %s\n", name)

			if place.Address != "" {
				result += fmt.Sprintf("   🏠 %s\n", place.Address)
			}

			if len(place.Categories) > 0 {
				categories := ""
				for j, cat := range place.Categories {
					categories += cat
					if j < len(place.Categories)-1 {
						categories += ", "
					}
				}
				result += fmt.Sprintf("   🏷️  %s\n", categories)
			}

			if place.Description != "" {
				result += fmt.Sprintf("   📝 %s\n", place.Description)
			}

			if place.Website != "" {
				result += fmt.Sprintf("   🌐 %s\n", place.Website)
			}

			if place.Latitude != 0 && place.Longitude != 0 {
				result += fmt.Sprintf("   🗺️  %.4f, %.4f\n", place.Latitude, place.Longitude)
			}

			if place.PlaceID != "" {
				result += fmt.Sprintf("   🆔 %s\n", place.PlaceID)
			}

			if i < len(results.Places)-1 {
				result += "\n"
			}
		}
		return mcp.NewToolResultText(result), nil
	}
}

// Handler for get_place_details tool
func handleGetPlaceDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	placeID, err := request.RequireString("place_id")
	if err != nil {
		return mcp.NewToolResultError("place_id is required"), nil //nolint:nilerr
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	details, err := client.GetPlaceDetails(placeID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting place details: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(details), nil
	}

	result := fmt.Sprintf("🏛️ **%s**\n\n", details.Data.Details.Name)
	result += fmt.Sprintf("📍 **Place ID:** %s\n", details.Data.Details.PlaceID)
	result += fmt.Sprintf("🏠 **Address:** %s\n", details.Data.Details.FormattedAddress)

	if details.Data.Details.Rating > 0 {
		result += fmt.Sprintf("⭐ **Rating:** %.1f/5 (%d reviews)\n",
			details.Data.Details.Rating, details.Data.Details.UserRatingsTotal)
	}

	if details.Data.Details.Website != "" {
		result += fmt.Sprintf("🌐 **Website:** %s\n", details.Data.Details.Website)
	}

	if details.Data.Details.InternationalPhoneNumber != "" {
		result += fmt.Sprintf("📞 **Phone:** %s\n", details.Data.Details.InternationalPhoneNumber)
	}

	if len(details.Data.Details.Types) > 0 {
		result += fmt.Sprintf("🏷️ **Types:** %v\n", details.Data.Details.Types)
	}

	coords := details.Data.Details.Geometry.Location
	result += fmt.Sprintf("🗺️ **Coordinates:** %.6f, %.6f\n", coords.Lat, coords.Lng)

	if details.Data.CardData.ReviewsSummary != "" {
		result += fmt.Sprintf("\n📊 **Reviews Summary:**\n%s\n", details.Data.CardData.ReviewsSummary)
	}

	if len(details.Data.CardData.ReasonsToVisit) > 0 {
		result += "\n✨ **Reasons to Visit:**\n"
		for i, reason := range details.Data.CardData.ReasonsToVisit {
			result += fmt.Sprintf("  %d. %s\n", i+1, reason)
		}
	}

	if len(details.Data.CardData.Tips) > 0 {
		result += "\n💡 **Tips:**\n"
		for i, tip := range details.Data.CardData.Tips {
			result += fmt.Sprintf("  %d. %s\n", i+1, tip)
		}
	}

	return mcp.NewToolResultText(result), nil
}

// Handler for search_places_wanderlog tool
func handleSearchPlacesWanderlog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil //nolint:nilerr
	}

	lat := request.GetFloat("latitude", 0.0)
	lng := request.GetFloat("longitude", 0.0)
	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.SearchPlacesWithWanderlog(query, lat, lng)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching places: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(results), nil
	}

	// Filter out junk rows: need a place_id AND a real description (non-empty, not just a type label)
	validCount := 0
	for _, place := range results.Data {
		if place.PlaceID != "" && place.Description != "" && place.Description != place.Type {
			validCount++
		}
	}

	if validCount == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No valid places found for query: %s (try being more specific)", query)), nil
	}

	result := fmt.Sprintf("🔍 **Found %d places for query:** %s\n\n", validCount, query)

	validIndex := 0
	for _, place := range results.Data {
		if place.PlaceID == "" || place.Description == "" || place.Description == place.Type {
			continue
		}
		validIndex++
		result += fmt.Sprintf("**%d. %s**\n", validIndex, place.Description)
		result += fmt.Sprintf("   📍 Place ID: %s\n", place.PlaceID)
		if len(place.Types) > 0 {
			result += fmt.Sprintf("   🏷️ Types: %v\n", place.Types)
		}
		if place.Type != "" && place.Type != place.Description {
			result += fmt.Sprintf("   🏷️ Type: %s\n", place.Type)
		}
		if validIndex < validCount {
			result += "\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

// Prompt handler
func handleAnalyzeTrip(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	tripID := ""
	focus := "overall"

	if args := request.Params.Arguments; args != nil {
		if id, exists := args["trip_id"]; exists {
			tripID = id
		}
		if f, exists := args["focus"]; exists {
			focus = f
		}
	}

	if tripID == "" {
		return nil, fmt.Errorf("trip_id argument is required")
	}

	var promptText string
	switch focus {
	case "budget":
		promptText = fmt.Sprintf("Please analyze the budget and expenses for trip %s. Look at the costs, suggest ways to save money, and identify any budget concerns.", tripID)
	case "itinerary":
		promptText = fmt.Sprintf("Please analyze the itinerary for trip %s. Look at the schedule, timing, transportation, and suggest optimizations or improvements.", tripID)
	case "places":
		promptText = fmt.Sprintf("Please analyze the places and destinations for trip %s. Evaluate the selection, suggest additional places to visit, and identify any missing must-see locations.", tripID)
	default:
		promptText = fmt.Sprintf("Please provide a comprehensive analysis of trip %s. Include insights on the itinerary, budget, places to visit, and overall trip planning. Suggest improvements and highlight any potential issues.", tripID)
	}

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Analyze trip %s with focus on %s", tripID, focus),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// Handler functions for new MCP tools

func handleCreateTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title := request.GetString("title", "")
	if title == "" {
		return mcp.NewToolResultError("title is required"), nil
	}

	startDate, endDate, err := validateDateRange("start_date", request.GetString("start_date", ""), "end_date", request.GetString("end_date", ""), true)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	privacy := request.GetString("privacy", "private")
	geoID := int(request.GetFloat("geo_id", 0))
	if geoID == 0 {
		return mcp.NewToolResultError("geo_id is required"), nil
	}
	if geoID < 0 {
		return mcp.NewToolResultError("geo_id must be a positive Wanderlog geo ID from search_geos"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	req := wanderlog.CreateTripRequest{
		Title:               title,
		GeoIDs:              []int{geoID},
		InitialMapsPlaceIDs: []int{},
		Type:                "plan",
		StartDate:           startDate,
		EndDate:             endDate,
		Privacy:             privacy,
	}

	resp, err := client.CreateTrip(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create trip '%s' (geo_id=%d, dates %s to %s): %v", title, geoID, startDate, endDate, err)), nil
	}

	result := fmt.Sprintf("✅ Created trip '%s' (Key: %s, ID: %d)", resp.TripPlan.Title, resp.TripPlan.Key, resp.TripPlan.ID)
	return mcp.NewToolResultText(result), nil
}

func handleDeleteTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err := client.DeleteTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete trip %s: %v", tripKey, err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("🗑️ Deleted trip %s", tripKey)), nil
}

func handleDeleteTrips(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKeysStr := request.GetString("trip_keys", "")
	if tripKeysStr == "" {
		return mcp.NewToolResultError("trip_keys is required (comma-separated list)"), nil
	}

	// Parse comma-separated trip keys
	var tripKeys []string
	for _, key := range strings.Split(tripKeysStr, ",") {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			tripKeys = append(tripKeys, trimmed)
		}
	}

	if len(tripKeys) == 0 {
		return mcp.NewToolResultError("trip_keys must contain at least one trip key"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	var deleted []string
	var failed []string

	for _, tripKey := range tripKeys {
		err := client.DeleteTrip(tripKey)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", tripKey, err))
		} else {
			deleted = append(deleted, tripKey)
		}
	}

	if len(failed) > 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Deleted %d/%d trips. Failed to delete: %s", len(deleted), len(tripKeys), strings.Join(failed, "; "))), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("🗑️ Deleted %d trips: %s", len(deleted), strings.Join(deleted, ", "))), nil
}

func handleRestoreTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err := client.RestoreTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to restore trip: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("♻️ Restored trip %s", tripKey)), nil
}

func handleCopyTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	resp, err := client.CopyTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to copy trip: %v", err)), nil
	}

	result := fmt.Sprintf("📋 Copied trip to '%s' (Key: %s)", resp.TripPlan.Title, resp.TripPlan.Key)
	return mcp.NewToolResultText(result), nil
}

func handleUpdateTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	title := request.GetString("title", "")
	startDate := request.GetString("start_date", "")
	endDate := request.GetString("end_date", "")
	privacy := request.GetString("privacy", "")

	// Validate that at least one field is provided
	if title == "" && startDate == "" && endDate == "" && privacy == "" {
		return mcp.NewToolResultError("At least one field must be provided (title, start_date, end_date, or privacy)"), nil
	}

	// Validate privacy if provided
	if privacy != "" && privacy != "public" && privacy != "private" && privacy != "unlisted" {
		return mcp.NewToolResultError("privacy must be one of: public, private, unlisted"), nil
	}
	if startDate != "" || endDate != "" {
		var err error
		startDate, endDate, err = validateDateRange("start_date", startDate, "end_date", endDate, false)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	updateReq := models.UpdateTripRequest{
		Title:     title,
		StartDate: startDate,
		EndDate:   endDate,
		Privacy:   privacy,
	}

	if err := client.UpdateTrip(tripKey, updateReq); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update trip %s: %v", tripKey, err)), nil
	}

	// Build result message showing what was updated
	updates := []string{}
	if title != "" {
		updates = append(updates, fmt.Sprintf("title to '%s'", title))
	}
	if startDate != "" {
		updates = append(updates, fmt.Sprintf("start date to %s", startDate))
	}
	if endDate != "" {
		updates = append(updates, fmt.Sprintf("end date to %s", endDate))
	}
	if privacy != "" {
		updates = append(updates, fmt.Sprintf("privacy to %s", privacy))
	}

	result := fmt.Sprintf("✅ Updated trip %s: %s", tripKey, strings.Join(updates, ", "))
	return mcp.NewToolResultText(result), nil
}

func handleLikeTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	liked := request.GetBool("liked", false)

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err := client.SetLike(tripKey, liked)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set like: %v", err)), nil
	}

	action := "liked"
	if !liked {
		action = "unliked"
	}
	return mcp.NewToolResultText(fmt.Sprintf("👍 Successfully %s trip %s", action, tripKey)), nil
}

func handleGetLikeCount(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	likeCount, err := client.GetLikeCount(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get like count: %v", err)), nil
	}

	result := fmt.Sprintf("Trip %s has %d likes", tripKey, likeCount.Count)
	if likeCount.UserLiked {
		result += " (you liked this trip)"
	}
	return mcp.NewToolResultText(result), nil
}

func handleSendInvites(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	inviteesStr := request.GetString("invitees", "")
	if inviteesStr == "" {
		return mcp.NewToolResultError("invitees is required"), nil
	}

	message := request.GetString("message", "")

	// Parse comma-separated invitees
	invitees := strings.Split(inviteesStr, ",")
	for i := range invitees {
		invitees[i] = strings.TrimSpace(invitees[i])
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	req := wanderlog.SendInvitesRequest{
		Invitees: invitees,
		Message:  message,
	}

	err := client.SendTripInvites(tripKey, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to send invites: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("📧 Sent invites to %d people for trip %s", len(invitees), tripKey)), nil
}

func handleListInvites(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	invites, err := client.ListTripInvites(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list invites: %v", err)), nil
	}

	if len(invites) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No invites found for trip %s", tripKey)), nil
	}

	result := fmt.Sprintf("Invites for trip %s:\n", tripKey)
	for i, invite := range invites {
		email, _ := invite["email"].(string)
		status, _ := invite["status"].(string)
		invitedAt, _ := invite["invitedAt"].(string)
		result += fmt.Sprintf("%d. %s - Status: %s (Sent: %s)\n", i+1, email, status, invitedAt)
	}

	return mcp.NewToolResultText(result), nil
}

func handleMovePlace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	placeID, err := request.RequireInt("place_id")
	if err != nil {
		return mcp.NewToolResultError("place_id is required"), nil //nolint:nilerr
	}

	fromSectionID, err := request.RequireInt("from_section_id")
	if err != nil {
		return mcp.NewToolResultError("from_section_id is required"), nil //nolint:nilerr
	}

	toSectionID, err := request.RequireInt("to_section_id")
	if err != nil {
		return mcp.NewToolResultError("to_section_id is required"), nil //nolint:nilerr
	}

	position, err := request.RequireInt("position")
	if err != nil {
		return mcp.NewToolResultError("position is required"), nil //nolint:nilerr
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err = client.MovePlace(tripKey, placeID, fromSectionID, toSectionID, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move place %d from section %d to section %d in trip %s: %v", placeID, fromSectionID, toSectionID, tripKey, err)), nil
	}

	result := fmt.Sprintf("🔀 Successfully moved place %d from section %d to section %d at position %d",
		placeID, fromSectionID, toSectionID, position)

	return mcp.NewToolResultText(result), nil
}

func handleReorderPlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	sectionID, err := request.RequireInt("section_id")
	if err != nil {
		return mcp.NewToolResultError("section_id is required"), nil //nolint:nilerr
	}

	placeIDsStr, err := request.RequireString("place_ids")
	if err != nil {
		return mcp.NewToolResultError("place_ids is required"), nil //nolint:nilerr
	}

	// Parse comma-separated place IDs
	placeIDStrs := strings.Split(placeIDsStr, ",")
	placeIDs := make([]int, 0, len(placeIDStrs))
	for _, idStr := range placeIDStrs {
		idStr = strings.TrimSpace(idStr)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid place ID '%s': %v", idStr, err)), nil
		}
		placeIDs = append(placeIDs, id)
	}

	if len(placeIDs) == 0 {
		return mcp.NewToolResultError("place_ids must contain at least one place ID"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err = client.ReorderPlaces(tripKey, sectionID, placeIDs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reorder %d places in section %d for trip %s: %v", len(placeIDs), sectionID, tripKey, err)), nil
	}

	result := fmt.Sprintf("📋 Successfully reordered %d places in section %d", len(placeIDs), sectionID)

	return mcp.NewToolResultText(result), nil
}

func handleGetFlightStops(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flightNumber, err := request.RequireString("flight_number")
	if err != nil {
		return mcp.NewToolResultError("flight_number is required"), nil //nolint:nilerr
	}
	airline, err := request.RequireString("airline")
	if err != nil {
		return mcp.NewToolResultError("airline is required"), nil //nolint:nilerr
	}
	date, err := request.RequireString("date")
	if err != nil {
		return mcp.NewToolResultError("date is required"), nil //nolint:nilerr
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.GetFlightStops(flightNumber, airline, date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting flight stops: %v", err)), nil
	}

	return mcp.NewToolResultStructuredOnly(results), nil
}

// handleSearchHotels searches for hotels/lodging
func handleSearchHotels(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	location, err := request.RequireString("location")
	if err != nil {
		return mcp.NewToolResultError("location is required"), nil //nolint:nilerr
	}

	checkIn, err := request.RequireString("check_in")
	if err != nil {
		return mcp.NewToolResultError("check_in is required"), nil //nolint:nilerr
	}

	checkOut, err := request.RequireString("check_out")
	if err != nil {
		return mcp.NewToolResultError("check_out is required"), nil //nolint:nilerr
	}

	guests := request.GetInt("guests", 1)

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.SearchLodgings(location, checkIn, checkOut, guests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Hotel search failed for %s (%s to %s, %d guests): %v", location, checkIn, checkOut, guests, err)), nil
	}

	return mcp.NewToolResultStructuredOnly(results), nil
}
