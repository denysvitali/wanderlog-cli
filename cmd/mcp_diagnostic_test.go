package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// TestDiagnostic_AddPlaceStepByStep performs step-by-step diagnostics to identify what causes trip corruption
func TestDiagnostic_AddPlaceStepByStep(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	// Helper to check trip health
	checkTripHealth := func(stepName string) {
		t.Logf("\n=== Checking trip health after: %s ===", stepName)

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
		require.NoError(t, err, "Failed to get trip at step: %s", stepName)
		require.NotNil(t, result, "Null result at step: %s", stepName)

		if result.IsError {
			t.Fatalf("❌ TRIP IS BROKEN at step '%s'! Error: %v", stepName, result.Content)
		}

		// Try to parse the JSON to verify it's valid
		if len(result.Content) > 0 {
			textContent := result.Content[0].(mcp.TextContent)
			var tripData map[string]interface{}
			err := json.Unmarshal([]byte(textContent.Text), &tripData)
			require.NoError(t, err, "Failed to parse trip JSON at step: %s", stepName)
			t.Logf("✓ Trip is healthy - JSON is valid")
		}
	}

	t.Run("step_by_step_diagnosis", func(t *testing.T) {
		// Step 0: Check initial trip health
		checkTripHealth("INITIAL STATE (before any operations)")

		// Step 1: Search for a place and fetch full data (place_id + coordinates)
		t.Logf("\n=== STEP 1: Searching for place and fetching coordinates ===")
		placeData := searchAndGetPlaceData(t, "Louvre Museum")
		t.Logf("✓ Found place_id: %s with coords: %.4f, %.4f", placeData.PlaceID, placeData.Lat, placeData.Lng)
		checkTripHealth("After searching (no modifications yet)")

		// Step 2: Add place with MINIMAL data (just name and place_id, NO coordinates)
		t.Logf("\n=== STEP 2: Adding place WITHOUT coordinates (name + place_id only) ===")
		minimalRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"name":     "Test Without Coords",
					"place_id": placeData.PlaceID,
					// NO latitude/longitude - this is what was causing the corruption!
				},
			},
		}

		result, err := handleAddPlace(ctx, minimalRequest)
		require.NoError(t, err)
		require.NotNil(t, result)

		if result.IsError {
			t.Logf("❌ Add place WITHOUT coords FAILED: %v", result.Content)
		} else {
			t.Logf("✓ Add place WITHOUT coords SUCCESS")
		}

		checkTripHealth("After adding place WITHOUT coordinates (THIS MAY CAUSE CORRUPTION)")

		// Step 3: Add place WITH coordinates (the correct way)
		t.Logf("\n=== STEP 3: Adding place WITH coordinates (CORRECT METHOD) ===")
		placeData2 := searchAndGetPlaceData(t, "Eiffel Tower")

		withCoordsRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":  testTripID,
					"name":      placeData2.Name,
					"place_id":  placeData2.PlaceID,
					"latitude":  placeData2.Lat,
					"longitude": placeData2.Lng,
				},
			},
		}

		result2, err := handleAddPlace(ctx, withCoordsRequest)
		require.NoError(t, err)
		require.NotNil(t, result2)

		if result2.IsError {
			t.Logf("❌ Add place WITH coords FAILED: %v", result2.Content)
		} else {
			t.Logf("✓ Add place WITH coords SUCCESS")
		}

		checkTripHealth("After adding place WITH coordinates")

		// Step 4: Add place with ALL fields (coordinates + text)
		t.Logf("\n=== STEP 4: Adding place with ALL fields (coords + text) ===")
		placeData3 := searchAndGetPlaceData(t, "Arc de Triomphe")

		fullRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_place",
				Arguments: map[string]interface{}{
					"trip_key":  testTripID,
					"name":      placeData3.Name,
					"place_id":  placeData3.PlaceID,
					"text":      "Complete test with all fields",
					"latitude":  placeData3.Lat,
					"longitude": placeData3.Lng,
				},
			},
		}

		result3, err := handleAddPlace(ctx, fullRequest)
		require.NoError(t, err)
		require.NotNil(t, result3)

		if result3.IsError {
			t.Logf("❌ Add place with all fields FAILED: %v", result3.Content)
		} else {
			t.Logf("✓ Add place with all fields SUCCESS")
		}

		checkTripHealth("After adding place with all fields")

		// Final summary
		t.Logf("\n=== DIAGNOSTIC COMPLETE ===")
		t.Logf("✓ All steps completed")
		t.Logf("⚠ If trip is broken, it broke at Step 2 (adding without coordinates)")
		t.Logf("⚠ WARNING: Trip now has 3 test places added - may need manual cleanup")
	})
}
