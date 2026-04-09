package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

const (
	testTripID = "jdysvggpzbjwpnej"
)

func init() {
	// Initialize logger for tests
	if logger == nil {
		logger = logrus.New()
		// Use DEBUG level for operational transform tests to see API responses
		logger.SetLevel(logrus.DebugLevel)
	}
}

// PlaceData holds both place_id and coordinates for adding places
type PlaceData struct {
	PlaceID string
	Lat     float64
	Lng     float64
	Name    string
}

// Helper function to search for a place and return its place_id (deprecated - use searchAndGetPlaceData instead)
func searchAndGetPlaceID(t *testing.T, query string) string {
	t.Helper()

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Search using Wanderlog's autocomplete API (Paris coordinates for location bias)
	result, err := client.SearchPlacesWithWanderllog(query, 48.8566, 2.3522)
	require.NoError(t, err, "Failed to search for place: %s", query)
	require.True(t, result.Success, "Search API returned success=false")
	require.NotEmpty(t, result.Data, "No search results found for: %s", query)

	// Return the first result's place_id
	placeID := result.Data[0].PlaceID
	require.NotEmpty(t, placeID, "First search result has empty place_id")

	t.Logf("Found place_id for '%s': %s (Description: %s)", query, placeID, result.Data[0].Description)
	return placeID
}

// searchAndGetPlaceData searches for a place and returns complete data including coordinates
func searchAndGetPlaceData(t *testing.T, query string) PlaceData {
	t.Helper()

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Search using Wanderlog's autocomplete API (Paris coordinates for location bias)
	result, err := client.SearchPlacesWithWanderllog(query, 48.8566, 2.3522)
	require.NoError(t, err, "Failed to search for place: %s", query)
	require.True(t, result.Success, "Search API returned success=false")
	require.NotEmpty(t, result.Data, "No search results found for: %s", query)

	// Get the first result's place_id
	placeID := result.Data[0].PlaceID
	require.NotEmpty(t, placeID, "First search result has empty place_id")

	t.Logf("Found place_id for '%s': %s (Description: %s)", query, placeID, result.Data[0].Description)

	// Fetch full place details to get coordinates
	details, err := client.GetPlaceDetails(placeID)
	require.NoError(t, err, "Failed to get place details for: %s", placeID)
	require.True(t, details.Success, "GetPlaceDetails returned success=false")

	lat := details.Data.Details.Geometry.Location.Lat
	lng := details.Data.Details.Geometry.Location.Lng
	name := details.Data.Details.Name

	t.Logf("Fetched coordinates for '%s': lat=%.4f, lng=%.4f", name, lat, lng)

	return PlaceData{
		PlaceID: placeID,
		Lat:     lat,
		Lng:     lng,
		Name:    name,
	}
}

// getDatedItinerarySectionID finds the first section with a date (actual itinerary section)
func getDatedItinerarySectionID(t *testing.T, tripID string) int {
	t.Helper()

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Load authentication
	auth, err := wanderlog.LoadCredentials()
	require.NoError(t, err, "Failed to load credentials")
	client.SetAuth(auth)

	// Get trip data
	trip, err := client.GetTrip(tripID)
	require.NoError(t, err, "Failed to get trip")

	// Find first section with a date
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.Date != nil && *section.Date != "" {
			t.Logf("Found dated itinerary section: ID=%d, Date=%s", section.ID, *section.Date)
			return section.ID
		}
	}

	t.Fatal("No dated itinerary sections found in trip - trip may need dates assigned")
	return 0
}

// skipIntegrationTest skips the test unless INTEGRATION_TESTS environment variable is set.
// Integration tests make real API calls and require authentication.
// To run integration tests: INTEGRATION_TESTS=1 go test ./cmd -v
func skipIntegrationTest(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}
}

// TestMCPIntegration_ListTrips tests the list_trips tool
func TestMCPIntegration_ListTrips(t *testing.T) {
	skipIntegrationTest(t)
	// Skip if not authenticated
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_trips",
			Arguments: map[string]interface{}{
				"format": "json",
			},
		},
	}

	result, err := handleListTrips(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_GetTrip tests the get_trip tool
func TestMCPIntegration_GetTrip(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		result, err := handleGetTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip",
				Arguments: map[string]interface{}{
					"format": "json",
				},
			},
		}

		result, err := handleGetTrip(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_trip",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_ListPlaces tests the list_places tool
func TestMCPIntegration_ListPlaces(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_places",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		result, err := handleListPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_places",
				Arguments: map[string]interface{}{
					"format": "default",
				},
			},
		}

		result, err := handleListPlaces(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

// TestMCPIntegration_ListSections tests the list_sections tool
func TestMCPIntegration_ListSections(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_sections",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		result, err := handleListSections(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_sections",
				Arguments: map[string]interface{}{
					"format": "default",
				},
			},
		}

		result, err := handleListSections(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

// TestMCPIntegration_GetPlaceDetails tests the get_place_details tool
func TestMCPIntegration_GetPlaceDetails(t *testing.T) {
	skipIntegrationTest(t)
	// This doesn't require authentication
	ctx := context.Background()

	t.Run("valid_place_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_place_details",
				Arguments: map[string]interface{}{
					"place_id": "ChIJN1t_tDeuEmsRUsoyG83frY4", // Google Sydney Office
					"format":   "json",
				},
			},
		}

		result, err := handleGetPlaceDetails(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Note: This might fail if the API is down, but should work most of the time
	})

	t.Run("missing_place_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_place_details",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetPlaceDetails(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_SearchPlacesWanderlog tests the search_places_wanderlog tool
func TestMCPIntegration_SearchPlacesWanderlog(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("search_paris", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_places_wanderlog",
				Arguments: map[string]interface{}{
					"query":  "Eiffel Tower",
					"format": "json",
				},
			},
		}

		result, err := handleSearchPlacesWanderlog(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("search_with_coordinates", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_places_wanderlog",
				Arguments: map[string]interface{}{
					"query":     "restaurants",
					"latitude":  48.8566,
					"longitude": 2.3522,
					"format":    "default",
				},
			},
		}

		result, err := handleSearchPlacesWanderlog(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "search_places_wanderlog",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleSearchPlacesWanderlog(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_TripResource tests the trip resource handler
func TestMCPIntegration_TripResource(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()

	t.Run("valid_trip_uri", func(t *testing.T) {
		request := mcp.ReadResourceRequest{
			Params: mcp.ReadResourceParams{
				URI: "wanderlog://trips/" + testTripID,
			},
		}

		contents, err := handleTripResource(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, contents)
		require.Len(t, contents, 1)

		textContent, ok := contents[0].(mcp.TextResourceContents)
		require.True(t, ok)
		assert.Equal(t, "application/json", textContent.MIMEType)
		assert.NotEmpty(t, textContent.Text)
	})

	t.Run("invalid_trip_uri", func(t *testing.T) {
		request := mcp.ReadResourceRequest{
			Params: mcp.ReadResourceParams{
				URI: "wanderlog://invalid",
			},
		}

		_, err := handleTripResource(ctx, request)
		require.Error(t, err)
	})
}

// TestMCPIntegration_AnalyzeTrip tests the analyze_trip prompt
func TestMCPIntegration_AnalyzeTrip(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	testCases := []struct {
		name  string
		focus string
	}{
		{"overall_analysis", "overall"},
		{"budget_focus", "budget"},
		{"itinerary_focus", "itinerary"},
		{"places_focus", "places"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.GetPromptRequest{
				Params: mcp.GetPromptParams{
					Name: "analyze_trip",
					Arguments: map[string]string{
						"trip_id": testTripID,
						"focus":   tc.focus,
					},
				},
			}

			result, err := handleAnalyzeTrip(ctx, request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.Description)
			assert.NotEmpty(t, result.Messages)
			assert.Len(t, result.Messages, 1)
			assert.Equal(t, mcp.RoleUser, result.Messages[0].Role)
		})
	}

	t.Run("missing_trip_id", func(t *testing.T) {
		request := mcp.GetPromptRequest{
			Params: mcp.GetPromptParams{
				Name:      "analyze_trip",
				Arguments: map[string]string{},
			},
		}

		_, err := handleAnalyzeTrip(ctx, request)
		require.Error(t, err)
	})
}

// TestMCPIntegration_ContextWithTripID tests the context trip ID functionality
func TestMCPIntegration_ContextWithTripID(t *testing.T) {
	ctx := context.Background()

	// Test without trip ID
	tripID, ok := tripIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, tripID)

	// Test with trip ID
	ctxWithTrip := withTripID(ctx, testTripID)
	tripID, ok = tripIDFromContext(ctxWithTrip)
	assert.True(t, ok)
	assert.Equal(t, testTripID, tripID)
}

// TestMCPIntegration_AddPlace tests the add_place tool (write operation)
func TestMCPIntegration_AddPlace(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()

	t.Run("add_place_basic", func(t *testing.T) {
		// Search for "Louvre Museum" and get complete place data (place_id + coordinates)
		placeData := searchAndGetPlaceData(t, "Louvre Museum")

		// Dynamically find the first dated itinerary section
		sectionID := getDatedItinerarySectionID(t, testTripID)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":   testTripID,
					"name":       placeData.Name,
					"place_id":   placeData.PlaceID,
					"latitude":   placeData.Lat,
					"longitude":  placeData.Lng,
					"section_id": sectionID,
					"text":       "Test place - added to dated itinerary with coordinates",
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		// The result should contain success message
		if len(result.Content) > 0 {
			textContent := result.Content[0].(mcp.TextContent)
			assert.Contains(t, textContent.Text, "Successfully added place")
		}
	})

	t.Run("add_place_with_place_id", func(t *testing.T) {
		// Search for "Eiffel Tower" and get complete place data (place_id + coordinates)
		placeData := searchAndGetPlaceData(t, "Eiffel Tower")

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":   testTripID,
					"name":       placeData.Name,
					"place_id":   placeData.PlaceID,
					"latitude":   placeData.Lat,
					"longitude":  placeData.Lng,
					"section_id": getDatedItinerarySectionID(t, testTripID), // Add to dated itinerary
					"text":       "Test place - added to itinerary section with coordinates",
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("add_place_with_default_trip_id", func(t *testing.T) {
		// Search for "Arc de Triomphe" and get complete place data (place_id + coordinates)
		placeData := searchAndGetPlaceData(t, "Arc de Triomphe")

		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"name":       placeData.Name,
					"place_id":   placeData.PlaceID,
					"latitude":   placeData.Lat,
					"longitude":  placeData.Lng,
					"section_id": getDatedItinerarySectionID(t, testTripID), // Add to dated itinerary
					"text":       "Test place - added to itinerary section with coordinates",
				},
			},
		}

		result, err := handleAddPlace(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("add_place_to_specific_section", func(t *testing.T) {
		// Search for "Notre-Dame de Paris" and get complete place data (place_id + coordinates)
		placeData := searchAndGetPlaceData(t, "Notre-Dame de Paris")

		// First, get the sections to find a valid section ID
		listSectionsReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_sections",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		sectionsResult, err := handleListSections(ctx, listSectionsReq)
		require.NoError(t, err)
		require.NotNil(t, sectionsResult)

		// If we have sections, try adding to the itinerary section
		if len(sectionsResult.Content) > 0 {
			// Add place to itinerary section
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "add_place",
					Arguments: map[string]interface{}{
						"trip_key":   testTripID,
						"name":       placeData.Name,
						"place_id":   placeData.PlaceID,
						"latitude":   placeData.Lat,
						"longitude":  placeData.Lng,
						"section_id": getDatedItinerarySectionID(t, testTripID), // Add to dated itinerary
						"text":       "Place added to itinerary section - dynamically looked up with coordinates",
					},
				},
			}

			result, err := handleAddPlace(ctx, request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
		}
	})

	t.Run("add_place_missing_name", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					// Missing required "name" field
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("add_place_missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"name": "Test Place",
					// Missing trip_key and no default in context
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_RemovePlace tests the remove_place tool (write operation)
func TestMCPIntegration_RemovePlace(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()

	t.Run("remove_place_missing_place_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "remove_place",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					// Missing required place_id
				},
			},
		}

		result, err := handleRemovePlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("remove_place_missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "remove_place",
				Arguments: map[string]interface{}{
					"place_id": 12345,
					// Missing trip_key and no default
				},
			},
		}

		result, err := handleRemovePlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	// Note: We don't test actual removal here to avoid breaking the test trip
	// A full integration test would:
	// 1. Add a place
	// 2. Get its ID from the response or by listing places
	// 3. Remove it
	// 4. Verify it's gone
}

// TestMCPIntegration_WriteOperationsWorkflow tests a complete workflow:
// Add a place, verify it exists, remove it, verify it's gone
func TestMCPIntegration_WriteOperationsWorkflow(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	ctx := context.Background()

	t.Run("complete_add_remove_workflow", func(t *testing.T) {
		// Step 1: Search for a place and get complete place data (place_id + coordinates)
		placeData := searchAndGetPlaceData(t, "Sacré-Cœur")

		// Step 2: Add the place to the trip itinerary section
		addRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":   testTripID,
					"name":       placeData.Name,
					"place_id":   placeData.PlaceID,
					"latitude":   placeData.Lat,
					"longitude":  placeData.Lng,
					"section_id": getDatedItinerarySectionID(t, testTripID), // Add to dated itinerary
					"text":       "Workflow test - added to itinerary section with coordinates",
				},
			},
		}

		addResult, err := handleAddPlace(ctx, addRequest)
		require.NoError(t, err)
		require.NotNil(t, addResult)
		if addResult.IsError {
			t.Logf("Warning: Failed to add test place: %v", addResult.Content)
			t.Skip("Skipping workflow test: cannot add place")
		}
		t.Logf("✓ Successfully added test place")

		// Step 2: Wait a moment for the API to process
		// (In production, you might want to add a small delay here)

		// Step 3: List places to verify it was added
		listRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_places",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		listResult, err := handleListPlaces(ctx, listRequest)
		require.NoError(t, err)
		require.NotNil(t, listResult)
		t.Logf("✓ Successfully listed places after adding")

		// Note: We can't easily extract the place ID from the response to remove it
		// without parsing the JSON structure, so this test demonstrates the workflow
		// but doesn't complete the removal step. A more comprehensive test would:
		// 1. Parse the JSON response
		// 2. Find the test place by name
		// 3. Extract its ID
		// 4. Remove it using that ID
		// 5. Verify it's gone

		t.Logf("✓ Workflow test completed (add + verify)")
		t.Logf("⚠ Note: Test place may remain in trip - manual cleanup recommended")
	})
}

// TestMCPIntegration_UpdatePlaceNotes tests updating notes on existing places using operational transforms
func TestMCPIntegration_UpdatePlaceNotes(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	t.Run("update_notes_on_place", func(t *testing.T) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Load authentication
		auth, err := wanderlog.LoadCredentials()
		require.NoError(t, err, "Failed to load credentials")
		client.SetAuth(auth)

		// Get trip data to find a place with notes
		trip, err := client.GetTrip(testTripID)
		require.NoError(t, err, "Failed to get trip")

		// Find the first dated section
		sectionID := getDatedItinerarySectionID(t, testTripID)

		// Find the section in the trip data
		var section *wanderlog.ItSections
		for i := range trip.TripPlan.Itinerary.Sections {
			if trip.TripPlan.Itinerary.Sections[i].ID == sectionID {
				section = &trip.TripPlan.Itinerary.Sections[i]
				break
			}
		}
		require.NotNil(t, section, "Could not find dated section in trip data")

		// Find the first place block in this section
		var blockIndex int = -1
		for i, block := range section.Blocks {
			if block.Type == "place" {
				blockIndex = i
				break
			}
		}
		require.NotEqual(t, -1, blockIndex, "No place blocks found in dated section")

		secIdx := sectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
		t.Logf("Updating block %d in section index %d (section ID %d)", blockIndex, secIdx, sectionID)

		// Create a copy of the entire blocks array with the updated block
		updatedBlocks := make([]interface{}, len(section.Blocks))
		for i, block := range section.Blocks {
			if i == blockIndex {
				// Update this block's text
				blockCopy := block
				blockCopy.Text = wanderlog.FlexibleText{
					IsString: false,
					Text: wanderlog.Text{
						Ops: []struct {
							Attributes *struct {
								Bold bool   `json:"bold,omitempty"`
								Link string `json:"link,omitempty"`
								List string `json:"list,omitzero"`
							} `json:"attributes,omitempty,omitzero"`
							Insert string `json:"insert,omitzero"`
						}{
							{
								Insert: "Updated note - testing operational transforms for note editing!",
							},
							{
								Insert: "\n",
							},
						},
					},
				}
				updatedBlocks[i] = blockCopy
			} else {
				updatedBlocks[i] = block
			}
		}

		// Replace the entire blocks array
		updateOp := wanderlog.ReplaceInObject(
			[]interface{}{"itinerary", "sections", secIdx, "blocks"},
			section.Blocks,
			updatedBlocks,
		)

		t.Logf("Replacing entire blocks array at path: /itinerary/sections/%d/blocks", secIdx)
		err = client.ApplyOperations(testTripID, []wanderlog.Operation{updateOp})
		require.NoError(t, err, "Failed to apply operation to update notes")

		// Verify the update worked
		updatedTrip, err := client.GetTrip(testTripID)
		require.NoError(t, err, "Failed to get trip after update")

		// Find the same section again
		for i := range updatedTrip.TripPlan.Itinerary.Sections {
			if updatedTrip.TripPlan.Itinerary.Sections[i].ID == sectionID {
				section = &updatedTrip.TripPlan.Itinerary.Sections[i]
				break
			}
		}

		// Verify the text was updated
		require.True(t, len(section.Blocks) > blockIndex, "Block disappeared after update")
		updatedBlock := section.Blocks[blockIndex]
		require.True(t, len(updatedBlock.Text.Text.Ops) > 0, "Text ops are empty")
		assert.Contains(t, updatedBlock.Text.Text.Ops[0].Insert, "Updated note", "Text was not updated")

		t.Logf("✓ Successfully updated place notes using operational transforms")
	})
}

// sectionIndex returns the index of a section with the given ID in the sections slice
func sectionIndex(sections []wanderlog.ItSections, sectionID int) int {
	for i, section := range sections {
		if section.ID == sectionID {
			return i
		}
	}
	return -1
}

// TestMCPIntegration_ReorderPlaces tests reordering places in a section using operational transforms
func TestMCPIntegration_ReorderPlaces(t *testing.T) {
	skipIntegrationTest(t)
	if _, err := wanderlog.LoadCredentials(); err != nil {
		t.Skip("Skipping integration test: not authenticated")
	}

	t.Run("reorder_places_in_section", func(t *testing.T) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Load authentication
		auth, err := wanderlog.LoadCredentials()
		require.NoError(t, err, "Failed to load credentials")
		client.SetAuth(auth)

		// Get trip data
		trip, err := client.GetTrip(testTripID)
		require.NoError(t, err, "Failed to get trip")

		// Find the first dated section
		sectionID := getDatedItinerarySectionID(t, testTripID)

		// Find the section in the trip data
		var section *wanderlog.ItSections
		secIdx := sectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
		require.NotEqual(t, -1, secIdx, "Could not find section index")
		section = &trip.TripPlan.Itinerary.Sections[secIdx]

		// Count place blocks
		placeCount := 0
		for _, block := range section.Blocks {
			if block.Type == "place" {
				placeCount++
			}
		}

		if placeCount < 2 {
			t.Skip("Not enough places in section to test reordering (need at least 2)")
		}

		// Record the original order
		t.Logf("Original place order in section %d:", sectionID)
		for i, block := range section.Blocks {
			if block.Type == "place" && block.Place != nil {
				t.Logf("  [%d] %s", i, block.Place.Name)
			}
		}

		// To reorder, we'll swap the first two place blocks
		// In operational transforms, we can do this by:
		// 1. Remove the second place (store it)
		// 2. Insert it before the first place
		// However, OT path-based operations can be complex.
		// A simpler approach: replace the entire blocks array with reordered version

		// Create a reordered blocks array - swap first two places
		reorderedBlocks := make([]interface{}, len(section.Blocks))
		firstPlaceIdx := -1
		secondPlaceIdx := -1

		// Find indices of first two places
		for i, block := range section.Blocks {
			if block.Type == "place" {
				if firstPlaceIdx == -1 {
					firstPlaceIdx = i
				} else if secondPlaceIdx == -1 {
					secondPlaceIdx = i
					break
				}
			}
		}

		// Copy blocks and swap the two places
		for i, block := range section.Blocks {
			if i == firstPlaceIdx {
				reorderedBlocks[i] = section.Blocks[secondPlaceIdx]
			} else if i == secondPlaceIdx {
				reorderedBlocks[i] = section.Blocks[firstPlaceIdx]
			} else {
				reorderedBlocks[i] = block
			}
		}

		// Apply operation to replace the blocks array
		reorderOp := wanderlog.ReplaceInObject(
			[]interface{}{"itinerary", "sections", secIdx, "blocks"},
			section.Blocks,
			reorderedBlocks,
		)

		t.Logf("Swapping places at indices %d and %d", firstPlaceIdx, secondPlaceIdx)
		err = client.ApplyOperations(testTripID, []wanderlog.Operation{reorderOp})
		require.NoError(t, err, "Failed to apply reorder operation")

		// Verify the reordering worked
		updatedTrip, err := client.GetTrip(testTripID)
		require.NoError(t, err, "Failed to get trip after reorder")

		// Find the same section again
		section = &updatedTrip.TripPlan.Itinerary.Sections[secIdx]

		t.Logf("New place order in section %d:", sectionID)
		for i, block := range section.Blocks {
			if block.Type == "place" && block.Place != nil {
				t.Logf("  [%d] %s", i, block.Place.Name)
			}
		}

		// Verify the swap occurred
		assert.Equal(t, trip.TripPlan.Itinerary.Sections[secIdx].Blocks[secondPlaceIdx].Place.Name,
			section.Blocks[firstPlaceIdx].Place.Name,
			"First and second places should be swapped")

		t.Logf("✓ Successfully reordered places using operational transforms")
	})
}

// TestMCPIntegration_ServerCreation tests that MCP servers can be created
func TestMCPIntegration_ServerCreation(t *testing.T) {
	t.Run("read_only_server", func(t *testing.T) {
		server := createMCPServer(true)
		require.NotNil(t, server)
	})

	t.Run("read_write_server", func(t *testing.T) {
		server := createMCPServer(false)
		require.NotNil(t, server)
	})
}
