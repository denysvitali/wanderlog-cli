package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

// TestMCPIntegration_CompleteFeatureTest creates a trip with all major features
// and exercises all MCP tools to ensure the API is working correctly.
func TestMCPIntegration_CompleteFeatureTest(t *testing.T) {
	skipIntegrationTest(t)

	// 1. Authenticate
	auth, err := loadAuthFromEnvOrKeychain()
	require.NoError(t, err)
	require.NotNil(t, auth)

	ctx := context.Background()

	// 2. Search for geo and create a Japan trip
	geoID := searchGeoIDForLifecycleTest(t, ctx, "Japan")
	tripTitle := "MCP Japan Trip - Complete Feature Test"

	createReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_trip",
			Arguments: map[string]interface{}{
				"title":      tripTitle,
				"geo_id":     geoID,
				"start_date": "2026-05-11",
				"end_date":   "2026-05-17",
				"privacy":    "private",
			},
		},
	}

	createResult, err := handleCreateTrip(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResult)

	var tripKey string
	if createResult.IsError {
		t.Skipf("Trip creation failed: %s", getTextContent(createResult))
	} else {
		tripKey = extractTripKey(getTextContent(createResult))
		require.NotEmpty(t, tripKey, "Failed to extract trip key")
	}

	t.Logf("Created trip: %s (key: %s)", tripTitle, tripKey)

	// Store the trip key for all subtests
	var (
		itinerarySectionID int
		placesAdded        []PlaceData
		copiedTripKeys     []string
	)

	// 3. TEST READ OPERATIONS with created trip
	t.Run("get_trip_shows_initial_state", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip",
				Arguments: map[string]interface{}{
					"trip_id": tripKey,
					"format":  "json",
				},
			},
		}

		result, err := handleGetTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("list_sections_shows_initial_sections", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_sections",
				Arguments: map[string]interface{}{
					"trip_id": tripKey,
					"format":  "json",
				},
			},
		}

		result, err := handleListSections(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	// 4. ADD PLACES TO UNSCHEDULED (Places to visit section)
	t.Run("add_place_to_unscheduled_section", func(t *testing.T) {
		// First get sections to find the "Places to visit" unscheduled section
		sectionID := getUnscheduledPlacesSectionID(t, tripKey)
		if sectionID == 0 {
			t.Skip("No unscheduled Places to visit section found")
		}

		placeData := searchAndGetPlaceData(t, "Senso-ji Temple Tokyo")
		placesAdded = append(placesAdded, placeData)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":   tripKey,
					"name":       placeData.Name,
					"place_id":   placeData.PlaceID,
					"latitude":   placeData.Lat,
					"longitude":  placeData.Lng,
					"section_id": sectionID,
					"text":       "Historic Buddhist temple in Asakusa",
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "add_place should not error: %s", getTextContent(result))
	})

	// 5. ADD PLACES TO ITINERARY (Dated sections for multiple days)
	t.Run("add_places_to_day_1_itinerary", func(t *testing.T) {
		sectionID := getDatedItinerarySectionIDByDate(t, tripKey, "2026-05-11")
		if sectionID == 0 {
			// Try getting any dated section
			sectionID = getDatedItinerarySectionID(t, tripKey)
		}
		if sectionID == 0 {
			t.Skip("No dated itinerary section found")
		}
		itinerarySectionID = sectionID

		places := []string{"Tokyo Tower", "Meiji Jingu Tokyo"}
		for _, query := range places {
			placeData := searchAndGetPlaceData(t, query)
			placesAdded = append(placesAdded, placeData)

			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "add_place",
					Arguments: map[string]interface{}{
						"trip_key":   tripKey,
						"name":       placeData.Name,
						"place_id":   placeData.PlaceID,
						"latitude":   placeData.Lat,
						"longitude":  placeData.Lng,
						"section_id": sectionID,
						"text":       fmt.Sprintf("Visiting %s during lifecycle test", placeData.Name),
					},
				},
			}

			result, err := handleAddPlace(ctx, request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError, "add_place should not error for %s: %s", query, getTextContent(result))
		}
	})

	t.Run("add_places_to_day_2_itinerary", func(t *testing.T) {
		sectionID := getDatedItinerarySectionIDByDate(t, tripKey, "2026-05-12")
		if sectionID == 0 {
			sectionID = getDatedItinerarySectionID(t, tripKey)
		}
		require.NotZero(t, sectionID, "No dated itinerary section found")

		placeData := searchAndGetPlaceData(t, "Fushimi Inari Taisha Kyoto")
		placesAdded = append(placesAdded, placeData)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":   tripKey,
					"name":       placeData.Name,
					"place_id":   placeData.PlaceID,
					"latitude":   placeData.Lat,
					"longitude":  placeData.Lng,
					"section_id": sectionID,
					"text":       "Iconic Kyoto shrine with torii gates",
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "add_place should not error: %s", getTextContent(result))
	})

	// 6. ADD FLIGHT
	t.Run("add_inbound_flight", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_flight",
				Arguments: map[string]interface{}{
					"trip_key":          tripKey,
					"flight_number":     "NH109",
					"departure_date":    "2026-05-11",
					"departure_airport": "JFK",
					"arrival_airport":   "HND",
				},
			},
		}

		result, err := handleAddFlight(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "add_flight should not error: %s", getTextContent(result))
	})

	t.Run("add_outbound_flight", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_flight",
				Arguments: map[string]interface{}{
					"trip_key":          tripKey,
					"flight_number":     "NH110",
					"departure_date":    "2026-05-17",
					"departure_airport": "HND",
					"arrival_airport":   "JFK",
				},
			},
		}

		result, err := handleAddFlight(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "add_flight should not error: %s", getTextContent(result))
	})

	// 7. ADD LODGING USING DEDICATED HANDLER
	t.Run("add_lodging_with_dedicated_handler", func(t *testing.T) {
		// Search for hotels first
		searchReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_hotels",
				Arguments: map[string]interface{}{
					"location":  "Tokyo",
					"check_in":  "2026-05-11",
					"check_out": "2026-05-17",
					"guests":    2,
				},
			},
		}

		hotelResult, err := handleSearchHotels(ctx, searchReq)
		require.NoError(t, err)

		lodgingName := "Hotel Metropolitan Tokyo Marunouchi"
		var placeID string
		var lat, lng float64

		if hotelResult.IsError {
			t.Logf("Hotel search unavailable, using fallback: %s", getTextContent(hotelResult))
		} else {
			lodgings, ok := hotelResult.StructuredContent.(*wanderlog.LodgingSearchResponse)
			if ok && lodgings.Success && len(lodgings.Data) > 0 {
				lodgingName = lodgings.Data[0].Name
			}
		}

		// Get place details for coordinates
		placeData := searchAndGetPlaceData(t, lodgingName+" Tokyo")
		placeID = placeData.PlaceID
		lat = placeData.Lat
		lng = placeData.Lng

		// Use dedicated add_lodging handler
		lodgingReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_lodging",
				Arguments: map[string]interface{}{
					"trip_key":           tripKey,
					"name":               lodgingName,
					"propertyPlaceId":    placeID,
					"latitude":           lat,
					"longitude":          lng,
					"checkInDate":        "2026-05-11",
					"checkOutDate":       "2026-05-17",
					"confirmationNumber": "CONF123456",
					"travelerNames":      []string{"Integration Traveler"},
					"note":               "Tokyo stay for Japan complete feature test",
				},
			},
		}

		lodgingResult, err := handleAddLodging(ctx, lodgingReq)
		require.NoError(t, err)
		require.NotNil(t, lodgingResult)
		assert.False(t, lodgingResult.IsError, "add_lodging should not error: %s", getTextContent(lodgingResult))
		structured, ok := lodgingResult.StructuredContent.(map[string]any)
		require.True(t, ok, "add_lodging should return structured content")
		assert.Equal(t, tripKey, structured["trip_key"])
		assert.NotZero(t, structured["section_id"])
		assert.NotZero(t, structured["block_id"])
	})

	// 8. UPDATE TRIP (title, dates, privacy)
	t.Run("update_trip_title", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					"title":    "Japan Adventure 2026 - Complete Feature Test",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "update_trip should not error: %s", getTextContent(result))
	})

	t.Run("update_trip_dates", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key":   tripKey,
					"start_date": "2026-05-11",
					"end_date":   "2026-05-17",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "update_trip dates should not error: %s", getTextContent(result))
	})

	t.Run("update_trip_privacy", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					"privacy":  "public",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "update_trip privacy should not error: %s", getTextContent(result))
	})

	// 9. ADD BUDGET AND EXPENSES
	t.Run("set_trip_budget", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "set_trip_budget",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					"amount":   450000,
					"currency": "JPY",
				},
			},
		}

		result, err := handleSetTripBudget(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "set_trip_budget should not error: %s", getTextContent(result))
	})

	t.Run("add_trip_expenses", func(t *testing.T) {
		expenses := []map[string]interface{}{
			{
				"description":     "Tokyo hotel deposit",
				"category":        "lodging",
				"amount":          120000,
				"currency":        "JPY",
				"date":            "2026-05-11",
				"associated_date": "2026-05-11",
			},
			{
				"description":     "JR Pass and local transit",
				"category":        "publicTransit",
				"amount":          80000,
				"currency":        "JPY",
				"date":            "2026-05-12",
				"associated_date": "2026-05-12",
			},
			{
				"description":     "Museums and shrines",
				"category":        "sightseeing",
				"amount":          30000,
				"currency":        "JPY",
				"date":            "2026-05-13",
				"associated_date": "2026-05-13",
			},
		}

		for _, expense := range expenses {
			expense["trip_key"] = tripKey
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "add_trip_expense",
					Arguments: expense,
				},
			}

			result, err := handleAddTripExpense(ctx, request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError, "add_trip_expense should not error: %s", getTextContent(result))
		}
	})

	// 10. REORDER PLACES
	t.Run("reorder_places_in_section", func(t *testing.T) {
		require.NotZero(t, itinerarySectionID, "No itinerary section ID was recorded")

		client := wanderlog.NewClient()
		client.SetLogger(logger)
		require.NoError(t, client.EnsureAuthenticated("", ""))

		trip, err := client.GetTrip(tripKey)
		require.NoError(t, err, "Failed to reload trip before reorder")

		originalIDs, originalNames := placeBlockOrderInSection(t, trip, itinerarySectionID)
		require.GreaterOrEqual(t, len(originalIDs), 2, "Need at least two place blocks to test reorder")

		reorderedIDs := make([]int, len(originalIDs))
		reorderedNames := make([]string, len(originalNames))
		for i := range originalIDs {
			reverseIndex := len(originalIDs) - 1 - i
			reorderedIDs[i] = originalIDs[reverseIndex]
			reorderedNames[i] = originalNames[reverseIndex]
		}

		placeIDArgs := make([]string, len(reorderedIDs))
		for i, id := range reorderedIDs {
			placeIDArgs[i] = fmt.Sprintf("%d", id)
		}

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "reorder_places",
				Arguments: map[string]interface{}{
					"trip_key":   tripKey,
					"section_id": itinerarySectionID,
					"place_ids":  strings.Join(placeIDArgs, ","),
				},
			},
		}

		result, err := handleReorderPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "reorder_places should not error: %s", getTextContent(result))

		updatedTrip, err := client.GetTrip(tripKey)
		require.NoError(t, err, "Failed to reload trip after reorder")
		_, updatedNames := placeBlockOrderInSection(t, updatedTrip, itinerarySectionID)
		require.GreaterOrEqual(t, len(updatedNames), len(reorderedNames))
		assert.Equal(t, reorderedNames, updatedNames[:len(reorderedNames)], "Place order should match requested order")

		t.Logf("Reordered section %d places: %s -> %s", itinerarySectionID, strings.Join(originalNames, ", "), strings.Join(updatedNames[:len(reorderedNames)], ", "))
	})

	// 11. LIKE TRIP
	t.Run("like_trip", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "like_trip",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					"liked":    true,
				},
			},
		}

		result, err := handleLikeTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Skip if API unavailable
		if result.IsError && strings.Contains(getTextContent(result), "unavailable") {
			t.Skip("Like API unavailable")
		}
		assert.False(t, result.IsError, "like_trip should not error: %s", getTextContent(result))
	})

	t.Run("get_like_count", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_like_count",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
				},
			},
		}

		result, err := handleGetLikeCount(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Skip if API unavailable
		if result.IsError && strings.Contains(getTextContent(result), "unavailable") {
			t.Skip("Like count API unavailable")
		}
		assert.False(t, result.IsError, "get_like_count should not error: %s", getTextContent(result))
	})

	t.Run("unlike_trip", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "like_trip",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					"liked":    false,
				},
			},
		}

		result, err := handleLikeTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		if result.IsError && strings.Contains(getTextContent(result), "unavailable") {
			t.Skip("Like API unavailable")
		}
		assert.False(t, result.IsError, "unlike_trip should not error: %s", getTextContent(result))
	})

	// 12. SEND TRIP INVITES
	t.Run("send_trip_invites_error_case", func(t *testing.T) {
		// This tests the error case - missing invitees
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "send_trip_invites",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					// Missing invitees
				},
			},
		}

		result, err := handleSendInvites(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError, "send_trip_invites should error without invitees")
	})

	// 13. COPY TRIP
	t.Run("copy_trip", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "copy_trip",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
				},
			},
		}

		result, err := handleCopyTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "copy_trip should not error: %s", getTextContent(result))

		// Extract copied trip key
		copiedTripKey := extractTripKey(getTextContent(result))
		if copiedTripKey != "" {
			t.Logf("Copied trip to: %s", copiedTripKey)
			copiedTripKeys = append(copiedTripKeys, copiedTripKey)
		}
	})

	// 14. GET FLIGHTS
	t.Run("get_flights_shows_added_flights", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_flights",
				Arguments: map[string]interface{}{
					"trip_id": tripKey,
					"format":  "json",
				},
			},
		}

		result, err := handleGetFlights(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "get_flights should not error")
	})

	// 15. GET EXPENSES CSV
	t.Run("get_trip_expenses_csv", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip_expenses_csv",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
				},
			},
		}

		result, err := handleGetTripExpensesCSV(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "get_trip_expenses_csv should not error")
	})

	// 16. FINAL VERIFICATION - Get trip and verify all contents
	t.Run("final_trip_verification", func(t *testing.T) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)
		require.NoError(t, client.EnsureAuthenticated("", ""))

		trip, err := client.GetTrip(tripKey)
		require.NoError(t, err, "Failed to reload trip for verification")

		t.Logf("Final trip state:")
		t.Logf("  Title: %s", trip.TripPlan.Title)
		t.Logf("  Key: %s", tripKey)
		t.Logf("  Start: %s", trip.TripPlan.StartDate)
		t.Logf("  End: %s", trip.TripPlan.EndDate)
		t.Logf("  Sections: %d", len(trip.TripPlan.Itinerary.Sections))
		t.Logf("  Budget: %+v", trip.TripPlan.Itinerary.Budget)

		// Verify sections exist
		assert.NotEmpty(t, trip.TripPlan.Itinerary.Sections, "Trip should have sections")
		assert.Equal(t, float64(450000), trip.TripPlan.Itinerary.Budget.Amount.Amount)
		assert.Equal(t, "JPY", trip.TripPlan.Itinerary.Budget.Amount.CurrencyCode)
		assert.GreaterOrEqual(t, len(trip.TripPlan.Itinerary.Budget.Expenses), 3)

		// Find and verify places in itinerary
		hasPlaces := false
		hasLodging := false
		hasHotelsSection := false
		for _, section := range trip.TripPlan.Itinerary.Sections {
			if section.Type == "hotels" {
				hasHotelsSection = true
			}
			if len(section.Blocks) > 0 {
				for _, block := range section.Blocks {
					if block.Place != nil {
						hasPlaces = true
						t.Logf("  Place in section %d (%s): %s", section.ID, section.Type, block.Place.Name)
					}
					if block.Hotel != nil {
						hasLodging = true
						assert.Equal(t, "place", block.Type, "Lodging must use the app's place block type")
						lodgingName := "<missing place>"
						if block.Place != nil {
							lodgingName = block.Place.Name
						}
						t.Logf("  Lodging in section %d (%s): %s (%s to %s)", section.ID, section.Type, lodgingName, block.Hotel.CheckIn, block.Hotel.CheckOut)
					}
					if block.FlightInfo != nil {
						t.Logf("  Flight in section %d: %s%d", section.ID, block.FlightInfo.Airline.Iata, block.FlightInfo.Number)
					}
				}
			}
		}
		assert.True(t, hasPlaces, "Trip should have places in itinerary sections")
		assert.True(t, hasHotelsSection, "Trip should have a Hotels and lodging section")
		assert.True(t, hasLodging, "Trip should have a lodging block")
	})

	// 17. CLEANUP - Delete the trip
	t.Run("cleanup_delete_trip", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "delete_trip",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
				},
			},
		}

		result, err := handleDeleteTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "delete_trip should not error: %s", getTextContent(result))
		t.Logf("Cleaned up trip: %s", tripKey)

		for _, copiedTripKey := range copiedTripKeys {
			copyDeleteReq := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "delete_trip",
					Arguments: map[string]interface{}{
						"trip_key": copiedTripKey,
					},
				},
			}
			copyDeleteResult, err := handleDeleteTrip(ctx, copyDeleteReq)
			require.NoError(t, err)
			require.NotNil(t, copyDeleteResult)
			assert.False(t, copyDeleteResult.IsError, "delete copied trip should not error: %s", getTextContent(copyDeleteResult))
			t.Logf("Cleaned up copied trip: %s", copiedTripKey)
		}
	})

	t.Logf("=== COMPLETE FEATURE TEST SUMMARY ===")
	t.Logf("Trip Key: %s", tripKey)
	t.Logf("Places Added: %d", len(placesAdded))
	for i, p := range placesAdded {
		t.Logf("  [%d] %s (%s)", i+1, p.Name, p.PlaceID)
	}
	t.Logf("Itinerary Section ID: %d", itinerarySectionID)
	t.Logf("================================")
}

// Helper function to get the unscheduled "Places to visit" section ID
func getUnscheduledPlacesSectionID(t *testing.T, tripKey string) int {
	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Skipf("Auth required: %v", err)
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil || trip.TripPlan.Itinerary.Sections == nil {
		return 0
	}

	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.Heading == "Places to visit" && section.Date == nil {
			return section.ID
		}
	}
	return 0
}

// Helper function to get a dated itinerary section by specific date
func getDatedItinerarySectionIDByDate(t *testing.T, tripKey string, date string) int {
	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Skipf("Auth required: %v", err)
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil || trip.TripPlan.Itinerary.Sections == nil {
		return 0
	}

	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.Date != nil && *section.Date == date && section.Type == "day" {
			return section.ID
		}
	}
	return 0
}

func placeBlockOrderInSection(t *testing.T, trip *wanderlog.TripResponse, sectionID int) ([]int, []string) {
	t.Helper()

	sectionIdx := wanderlog.FindSectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
	require.NotEqual(t, -1, sectionIdx, "section %d not found", sectionID)

	ids := []int{}
	names := []string{}
	for _, block := range trip.TripPlan.Itinerary.Sections[sectionIdx].Blocks {
		if block.Place == nil {
			continue
		}
		ids = append(ids, block.ID)
		names = append(names, block.Place.Name)
	}
	return ids, names
}
