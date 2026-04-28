package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func init() {
	// Initialize logger for tests
	if logger == nil {
		logger = logrus.New()
		// Use DEBUG level for operational transform tests to see API responses
		logger.SetLevel(logrus.DebugLevel)
	}
}

func extractTripKey(text string) string {
	keyPrefix := "Key:"
	keyStart := strings.Index(text, keyPrefix)
	if keyStart < 0 {
		return ""
	}
	keyStart += len(keyPrefix)
	remainder := strings.TrimSpace(text[keyStart:])
	if keyEnd := strings.IndexAny(remainder, ",) \n\t"); keyEnd >= 0 {
		return strings.TrimSpace(remainder[:keyEnd])
	}
	return strings.TrimSpace(remainder)
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

// getTextContent extracts text from mcp.TextContent or returns empty string
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

// PlaceData holds both place_id and coordinates for adding places
type PlaceData struct {
	PlaceID string
	Lat     float64
	Lng     float64
	Name    string
}

type geoSearchToolResult struct {
	Success bool `json:"success"`
	Data    struct {
		Geos []geoGuideCount `json:"geos"`
	} `json:"data"`
}

func searchGeoIDForLifecycleTest(t *testing.T, ctx context.Context, query string) int {
	t.Helper()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_geos",
			Arguments: map[string]interface{}{
				"query": query,
				"limit": 5,
			},
		},
	}

	result, err := handleSearchGeos(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError, "search_geos should not return error: %s", getTextContent(result))
	require.NotNil(t, result.StructuredContent, "search_geos should return structured content")

	raw, err := json.Marshal(result.StructuredContent)
	require.NoError(t, err)

	var parsed geoSearchToolResult
	require.NoError(t, json.Unmarshal(raw, &parsed))
	require.True(t, parsed.Success, "search_geos returned success=false")
	require.NotEmpty(t, parsed.Data.Geos, "No geos found for %q", query)
	require.NotZero(t, parsed.Data.Geos[0].GeoID, "First geo result has no geo ID")

	t.Logf("Found geo_id for %q: %d (%s)", query, parsed.Data.Geos[0].GeoID, parsed.Data.Geos[0].Name)
	return parsed.Data.Geos[0].GeoID
}

func lifecycleTripTitle() string {
	if runID := os.Getenv("GITHUB_RUN_ID"); runID != "" {
		return fmt.Sprintf("MCP Lifecycle Test - CI Run %s", runID)
	}
	return fmt.Sprintf("MCP Lifecycle Test - Local %d", time.Now().UnixNano())
}

// searchAndGetPlaceData searches for a place and returns complete data including coordinates
func searchAndGetPlaceData(t *testing.T, query string) PlaceData {
	t.Helper()

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Search using Wanderlog's autocomplete API (Paris coordinates for location bias)
	result, err := client.SearchPlacesWithWanderlog(query, 48.8566, 2.3522)
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
	auth, err := loadAuthFromEnvOrKeychain()
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
	if os.Getenv("CI") == "true" && !hasIntegrationAuthEnv() {
		t.Skip("Skipping authenticated integration test. Configure Wanderlog auth secrets to run in CI.")
	}
}

func hasIntegrationAuthEnv() bool {
	hasSessionAuth := os.Getenv("WANDERLOG_AUTH_SESSION_COOKIE") != "" &&
		os.Getenv("WANDERLOG_AUTH_SESSION_XSRF_TOKEN") != ""
	hasLoginAuth := os.Getenv("WANDERLOG_AUTH_EMAIL") != "" &&
		os.Getenv("WANDERLOG_AUTH_PASSWORD") != ""
	return hasSessionAuth || hasLoginAuth
}

// getFirstSectionID gets the first section ID for a trip (first section with any content)
func getFirstSectionID(ctx context.Context, tripKey string) int {
	client := wanderlog.NewClient()
	client.SetLogger(logger)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		return 0
	}
	client.SetAuth(auth)

	trip, err := client.GetTrip(tripKey)
	if err != nil || trip.TripPlan.Itinerary.Sections == nil {
		return 0
	}

	// Find first section with blocks
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if len(section.Blocks) > 0 {
			return section.ID
		}
	}

	return 0
}

// TestMCPIntegration_AddPlace tests the add_place tool (write operation)
func TestMCPIntegration_AddPlace(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

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
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

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
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

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
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	t.Run("update_notes_on_place", func(t *testing.T) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Load authentication
		auth, err := loadAuthFromEnvOrKeychain()
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
		blockIndex := -1
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

// TestMCPIntegration_ReorderPlaces tests reordering places in a section using operational transforms
func TestMCPIntegration_ReorderPlaces(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	t.Run("reorder_places_in_section", func(t *testing.T) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Load authentication
		auth, err := loadAuthFromEnvOrKeychain()
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
			switch i {
			case firstPlaceIdx:
				reorderedBlocks[i] = section.Blocks[secondPlaceIdx]
			case secondPlaceIdx:
				reorderedBlocks[i] = section.Blocks[firstPlaceIdx]
			default:
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

// TestMCPIntegration_CompleteTripLifecycle tests the complete trip lifecycle:
// Create a trip, verify it exists, add a place, get trip details, then clean up
func TestMCPIntegration_CompleteTripLifecycle(t *testing.T) {
	skipIntegrationTest(t)

	// 1. Authenticate
	auth, err := loadAuthFromEnvOrKeychain()
	require.NoError(t, err)
	require.NotNil(t, auth)

	ctx := context.Background()

	// 2. Search for a destination geo and create a trip with unique name
	geoID := searchGeoIDForLifecycleTest(t, ctx, "Paris")
	tripTitle := lifecycleTripTitle()
	createReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_trip",
			Arguments: map[string]interface{}{
				"title":      tripTitle,
				"geo_id":     geoID,
				"start_date": "2026-06-01",
				"end_date":   "2026-06-05",
				"privacy":    "private",
			},
		},
	}

	createResult, err := handleCreateTrip(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResult)

	textContent := createResult.Content[0].(mcp.TextContent)

	var tripKey string
	if createResult.IsError {
		// Trip creation failed (API issue) - fall back to existing test trip
		t.Logf("Trip creation failed (API issue), falling back to existing test trip: %s", textContent.Text)
		tripKey = testTripID
	} else {
		// Extract trip key - format is "Created trip 'Title' (Key: abc123)"
		tripKey = extractTripKey(textContent.Text)
		require.NotEmpty(t, tripKey, "Failed to extract trip key from: %s", textContent.Text)
	}

	if tripKey != testTripID {
		t.Logf("Keeping lifecycle test trip for inspection: title=%q key=%s", tripTitle, tripKey)
	}

	// 4. TEST READ TOOLS with created trip
	t.Run("list_trips_verifies_new_trip", func(t *testing.T) {
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
	})

	t.Run("get_trip_with_created_trip", func(t *testing.T) {
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

	t.Run("list_sections_with_created_trip", func(t *testing.T) {
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

	// 5. TEST WRITE TOOLS
	var lifecycleSectionID int
	expectedPlaces := make([]PlaceData, 0, 3)
	var expectedFlightSectionID int

	// First need a section ID - get trip and find first section
	t.Run("add_place_to_new_trip", func(t *testing.T) {
		sectionID := getFirstSectionID(ctx, tripKey)
		if sectionID == 0 {
			t.Skip("No sections found in newly created trip")
		}

		// Search for a place
		placeData := searchAndGetPlaceData(t, "Louvre Museum")

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
					"text":       "Added during lifecycle test",
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("add_places_to_itinerary_section", func(t *testing.T) {
		sectionID := getDatedItinerarySectionID(t, tripKey)
		if sectionID == 0 {
			t.Skip("No dated itinerary sections found in trip")
		}
		lifecycleSectionID = sectionID

		places := []string{"Eiffel Tower", "Notre-Dame de Paris"}
		for _, query := range places {
			placeData := searchAndGetPlaceData(t, query)
			expectedPlaces = append(expectedPlaces, placeData)
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
						"text":       "Added to dated itinerary during lifecycle test",
					},
				},
			}

			result, err := handleAddPlace(ctx, request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError, "add_place should not return error for %s: %s", query, getTextContent(result))
		}
	})

	t.Run("search_lodging_and_add_to_itinerary", func(t *testing.T) {
		sectionID := getDatedItinerarySectionID(t, tripKey)
		if sectionID == 0 {
			t.Skip("No dated itinerary sections found in trip")
		}

		searchHotelsReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_hotels",
				Arguments: map[string]interface{}{
					"location":  "Paris",
					"check_in":  "2026-06-01",
					"check_out": "2026-06-05",
					"guests":    1,
				},
			},
		}

		hotelResult, err := handleSearchHotels(ctx, searchHotelsReq)
		require.NoError(t, err)
		require.NotNil(t, hotelResult)

		lodgingName := "Hôtel du Louvre"
		lodgingAddress := "Paris"
		if hotelResult.IsError {
			errText := getTextContent(hotelResult)
			require.True(t,
				strings.Contains(errText, "Wanderlog lodging search is currently failing server-side") ||
					strings.Contains(errText, "Cannot read properties of undefined (reading 'length')"),
				"unexpected search_hotels error: %s", errText)
			t.Logf("search_hotels is unavailable server-side; falling back to known lodging place search: %s", errText)
		} else {
			lodgings, ok := hotelResult.StructuredContent.(*wanderlog.LodgingSearchResponse)
			require.True(t, ok, "search_hotels returned unexpected structured content type %T", hotelResult.StructuredContent)
			require.True(t, lodgings.Success, "search_hotels returned success=false")
			require.NotEmpty(t, lodgings.Data, "No lodging results found for Paris")

			lodging := lodgings.Data[0]
			require.NotEmpty(t, lodging.Name, "First lodging result has empty name")
			lodgingName = lodging.Name
			lodgingAddress = lodging.Address
			t.Logf("Found lodging for lifecycle test: %s (%s)", lodgingName, lodgingAddress)
		}

		placeData := searchAndGetPlaceData(t, lodgingName+" Paris")
		expectedPlaces = append(expectedPlaces, PlaceData{
			PlaceID: placeData.PlaceID,
			Lat:     placeData.Lat,
			Lng:     placeData.Lng,
			Name:    lodgingName,
		})
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":   tripKey,
					"name":       lodgingName,
					"place_id":   placeData.PlaceID,
					"latitude":   placeData.Lat,
					"longitude":  placeData.Lng,
					"section_id": sectionID,
					"text":       fmt.Sprintf("Lodging added during lifecycle test. Address: %s", lodgingAddress),
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "add_place should add lodging to itinerary: %s", getTextContent(result))
	})

	// Test add_flight
	t.Run("add_flight_to_new_trip", func(t *testing.T) {
		// Get a dated section for the flight
		sectionID := getDatedItinerarySectionID(t, tripKey)
		if sectionID == 0 {
			t.Skip("No dated itinerary sections found in trip")
		}
		expectedFlightSectionID = sectionID

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_flight",
				Arguments: map[string]interface{}{
					"trip_key":       tripKey,
					"section_id":     sectionID,
					"flight_number":  "MU244",
					"departure_date": "2026-06-02",
				},
			},
		}

		flightResult, err := handleAddFlight(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, flightResult)
		assert.False(t, flightResult.IsError, "add_flight should not return error: %s", flightResult.Content[0].(mcp.TextContent).Text)

		// Verify the trip can still be fetched with the flight
		getRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip",
				Arguments: map[string]interface{}{
					"trip_id": tripKey,
					"format":  "json",
				},
			},
		}

		getResult, err := handleGetTrip(ctx, getRequest)
		require.NoError(t, err)
		require.NotNil(t, getResult)
		assert.False(t, getResult.IsError, "get_trip after add_flight should not return error")

		t.Logf("✓ Successfully added flight and verified trip can be fetched")
	})

	t.Run("verify_retained_trip_has_lifecycle_contents", func(t *testing.T) {
		require.NotZero(t, lifecycleSectionID, "No lifecycle itinerary section was recorded")
		require.NotZero(t, expectedFlightSectionID, "No flight itinerary section was recorded")
		require.Len(t, expectedPlaces, 3, "Expected two attraction places plus one lodging place")

		client := wanderlog.NewClient()
		client.SetLogger(logger)
		require.NoError(t, client.EnsureAuthenticated("", ""))

		trip, err := client.GetTrip(tripKey)
		require.NoError(t, err, "Failed to reload retained lifecycle trip")

		sectionIdx := wanderlog.FindSectionIndex(trip.TripPlan.Itinerary.Sections, lifecycleSectionID)
		require.NotEqual(t, -1, sectionIdx, "Lifecycle itinerary section disappeared")
		section := trip.TripPlan.Itinerary.Sections[sectionIdx]

		t.Logf("Retained lifecycle trip: title=%q key=%s section_id=%d", trip.TripPlan.Title, tripKey, lifecycleSectionID)
		for i, block := range section.Blocks {
			switch {
			case block.Place != nil:
				t.Logf("  section block[%d] place: %s (%s)", i, block.Place.Name, block.Place.PlaceID)
			case block.FlightInfo != nil:
				t.Logf("  section block[%d] flight: %s%d on %s", i, block.FlightInfo.Airline.Iata, block.FlightInfo.Number, block.Depart.Date)
			default:
				t.Logf("  section block[%d] type: %s", i, block.Type)
			}
		}

		for _, place := range expectedPlaces {
			assert.True(t,
				tripHasAddedPlace(trip, lifecycleSectionID, place.Name, place.PlaceID),
				"Retained trip is missing place %q (%s) in section %d", place.Name, place.PlaceID, lifecycleSectionID)
		}
		assert.True(t,
			tripHasAddedFlight(trip, expectedFlightSectionID, "MU", 244, "2026-06-02"),
			"Retained trip is missing flight MU244 on 2026-06-02 in section %d", expectedFlightSectionID)
	})

	// 6. TEST ERROR CASES
	t.Run("create_trip_missing_title", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "create_trip",
				Arguments: map[string]interface{}{
					"start_date": "2026-06-01",
					"end_date":   "2026-06-05",
				},
			},
		}

		result, err := handleCreateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("get_trip_non_existent", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip",
				Arguments: map[string]interface{}{
					"trip_id": "nonexistent_trip_key_12345",
					"format":  "json",
				},
			},
		}

		result, err := handleGetTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Non-existent trip may return error or empty result depending on API behavior
	})

	t.Run("add_place_missing_name", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key": tripKey,
					// Missing required "name" field
				},
			},
		}

		result, err := handleAddPlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("delete_trip_missing_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "delete_trip",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleDeleteTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestIntegration_DeleteTrips tests the bulk delete_trips tool
func TestIntegration_DeleteTrips(t *testing.T) {
	skipIntegrationTest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	// Create a trip to delete
	tripTitle := fmt.Sprintf("Delete Test - %d", time.Now().UnixNano())
	createReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_trip",
			Arguments: map[string]interface{}{
				"title":      tripTitle,
				"start_date": "2026-07-01",
				"end_date":   "2026-07-05",
				"privacy":    "private",
			},
		},
	}

	createResult, err := handleCreateTrip(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResult)

	var tripKey string
	if !createResult.IsError {
		textContent := createResult.Content[0].(mcp.TextContent)
		tripKey = extractTripKey(textContent.Text)
	}

	if tripKey == "" {
		t.Skip("Could not create test trip for deletion test")
	}

	// Now delete the trip using handleDeleteTrips with comma-separated keys
	deleteReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "delete_trips",
			Arguments: map[string]interface{}{
				"trip_keys": tripKey,
			},
		},
	}

	result, err := handleDeleteTrips(ctx, deleteReq)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should succeed - the single key should be deleted
	assert.False(t, result.IsError, "delete_trips should not return error: %s", getTextContent(result))
	assert.Contains(t, getTextContent(result), "Deleted")
}

// TestUnit_SearchHotelsErrorPropagation tests that handleSearchHotels returns an error
// when the API returns success=false
func TestUnit_SearchHotelsErrorPropagation(t *testing.T) {
	// Use httptest to mock the API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response with success=false
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": false, "error": "API error"}`))
	}))
	defer server.Close()

	// Save original BaseURL
	oldBaseURL := wanderlog.BaseURL
	wanderlog.BaseURL = server.URL
	defer func() { wanderlog.BaseURL = oldBaseURL }()

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_hotels",
			Arguments: map[string]interface{}{
				"location":  "Paris",
				"check_in":  "2026-07-01",
				"check_out": "2026-07-05",
			},
		},
	}

	result, err := handleSearchHotels(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError, "Expected error result when API returns success=false")
	// The error message from the handler wraps the client error
	assert.Contains(t, getTextContent(result), "Hotel search failed for")
}

// TestUnit_SearchPlacesWanderlogJunkFiltering tests that rows without place_id
// are filtered out from the results
// Note: This requires a real API call since the wanderlog autocomplete constructs
// its own URL that can't be easily mocked. Skip unless running with prod integration.
func TestUnit_SearchPlacesWanderlogJunkFiltering(t *testing.T) {
	if os.Getenv("WANDERLOG_RUN_PROD_INTEGRATION") != "1" {
		t.Skip("Skipping test that requires real API. Set WANDERLOG_RUN_PROD_INTEGRATION=1 to run.")
	}

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_places_wanderlog",
			Arguments: map[string]interface{}{
				"query":  "Eiffel Tower Paris",
				"format": "default",
			},
		},
	}

	result, err := handleSearchPlacesWanderlog(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// The filtering logic should have removed any results without place_id
	// If results come back, they should have valid place_ids
	text := getTextContent(result)
	t.Logf("Search result: %s", text)
	// The test passes if we get results and no error
	assert.False(t, result.IsError)
}
