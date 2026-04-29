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

	// 2. Search for geo and create a trip
	geoID := searchGeoIDForLifecycleTest(t, ctx, "Paris")
	tripTitle := lifecycleTripTitle()

	createReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_trip",
			Arguments: map[string]interface{}{
				"title":      tripTitle,
				"geo_id":     geoID,
				"start_date": "2026-06-01",
				"end_date":   "2026-06-07",
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

		placeData := searchAndGetPlaceData(t, "Senso-ji Temple")
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
		sectionID := getDatedItinerarySectionIDByDate(t, tripKey, "2026-06-01")
		if sectionID == 0 {
			// Try getting any dated section
			sectionID = getDatedItinerarySectionID(t, tripKey)
		}
		if sectionID == 0 {
			t.Skip("No dated itinerary section found")
		}
		itinerarySectionID = sectionID

		places := []string{"Eiffel Tower", "Louvre Museum"}
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
		sectionID := getDatedItinerarySectionIDByDate(t, tripKey, "2026-06-02")
		if sectionID == 0 {
			t.Skip("No dated itinerary section for 2026-06-02")
		}

		placeData := searchAndGetPlaceData(t, "Notre-Dame de Paris")
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
					"text":       "Medieval Catholic cathedral",
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
					"trip_key":       tripKey,
					"flight_number":  "AF128",
					"departure_date": "2026-06-01",
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
					"trip_key":       tripKey,
					"flight_number":  "AF129",
					"departure_date": "2026-06-07",
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
					"location":  "Paris",
					"check_in":  "2026-06-01",
					"check_out": "2026-06-07",
					"guests":    2,
				},
			},
		}

		hotelResult, err := handleSearchHotels(ctx, searchReq)
		require.NoError(t, err)

		lodgingName := "Hôtel du Louvre"
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
		placeData := searchAndGetPlaceData(t, lodgingName+" Paris")
		placeID = placeData.PlaceID
		lat = placeData.Lat
		lng = placeData.Lng

		// Use dedicated add_lodging handler
		lodgingReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_lodging",
				Arguments: map[string]interface{}{
					"trip_key":          tripKey,
					"name":              lodgingName,
					"place_id":         placeID,
					"latitude":         lat,
					"longitude":        lng,
					"check_in":         "2026-06-01",
					"check_out":        "2026-06-07",
					"confirmation_number": "CONF123456",
					"notes":            "Added via add_lodging handler - Complete Feature Test",
				},
			},
		}

		lodgingResult, err := handleAddLodging(ctx, lodgingReq)
		require.NoError(t, err)
		require.NotNil(t, lodgingResult)
		assert.False(t, lodgingResult.IsError, "add_lodging should not error: %s", getTextContent(lodgingResult))
	})

	// 8. UPDATE TRIP (title, dates, privacy)
	t.Run("update_trip_title", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					"title":    "Paris Adventure 2026 - Complete Feature Test",
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
					"start_date": "2026-06-02",
					"end_date":   "2026-06-08",
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

	// 9. REORDER PLACES (skip - requires internal block IDs, not Google Place IDs)
	t.Run("reorder_places_in_section", func(t *testing.T) {
		// The reorder_places API requires internal block IDs (integers), not Google Place IDs (strings).
		// The placesAdded array contains Google Place IDs which cannot be used with reorder_places.
		// This functionality is tested in other tests. Skip this specific test case.
		t.Skip("Skipping reorder test - requires internal block IDs, not Google Place IDs")
	})

	// 10. LIKE TRIP
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

	// 11. SEND TRIP INVITES
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

	// 12. COPY TRIP
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
			// Don't delete the original, keep both for verification
		}
	})

	// 13. GET FLIGHTS
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

	// 14. GET EXPENSES CSV (empty but should work)
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
		// Expenses will be empty but should not error
		assert.False(t, result.IsError, "get_trip_expenses_csv should not error")
	})

	// 15. FINAL VERIFICATION - Get trip and verify all contents
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

		// Find and verify places in itinerary
		hasPlaces := false
		for _, section := range trip.TripPlan.Itinerary.Sections {
			if len(section.Blocks) > 0 {
				for _, block := range section.Blocks {
					if block.Place != nil {
						hasPlaces = true
						t.Logf("  Place in section %d (%s): %s", section.ID, section.Type, block.Place.Name)
					}
					if block.FlightInfo != nil {
						t.Logf("  Flight in section %d: %s%d", section.ID, block.FlightInfo.Airline.Iata, block.FlightInfo.Number)
					}
				}
			}
		}
		assert.True(t, hasPlaces, "Trip should have places in itinerary sections")
	})

	// 16. CLEANUP - Delete the trip
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
		// Unscheduled sections have ID < 1000000000 and type "placeList"
		if section.ID < 1000000000 && section.Type == "placeList" && section.Heading == "Places to visit" {
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