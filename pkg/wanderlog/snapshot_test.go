//go:build integration
// +build integration

package wanderlog

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sirupsen/logrus"
)

// TestSnapshotTrip creates a trip, performs various operations, and captures
// snapshots of the raw server response after each step to show the differences
func TestSnapshotTrip(t *testing.T) {
	// Initialize config to load credentials from config file
	if err := InitConfig(); err != nil {
		t.Logf("Warning: Failed to initialize config: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := NewClient()
	client.SetLogger(logger)

	// Authenticate
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first or set credentials in config file", err)
	}

	// Step 1: Create a new trip
	t.Log("📸 STEP 1: Creating new trip")
	createReq := CreateTripRequest{
		Title:     fmt.Sprintf("Snapshot Test Trip - %d", time.Now().Unix()),
		StartDate: "2025-12-01",
		EndDate:   "2025-12-03",
		Privacy:   "private",
	}

	createResp, err := client.CreateTrip(createReq)
	if err != nil {
		t.Fatalf("Failed to create trip: %v", err)
	}

	tripKey := createResp.TripPlan.Key
	t.Logf("✅ Created trip: %s", tripKey)
	defer func() {
		// Clean up - delete the trip at the end
		if err := client.DeleteTrip(tripKey); err != nil {
			t.Logf("Warning: Failed to delete test trip: %v", err)
		}
	}()

	snapshot1 := captureSnapshot(t, client, tripKey, "1-initial-creation")
	t.Logf("📸 Snapshot 1: Trip after creation (%d bytes)", len(snapshot1))

	// Step 2: Add first place
	time.Sleep(1 * time.Second)
	t.Log("\n📸 STEP 2: Adding first place")

	searchResp, err := client.SearchPlacesWithWanderllog("Eiffel Tower Paris", 48.8566, 2.3522)
	if err != nil || len(searchResp.Data) == 0 {
		t.Fatal("Failed to find Eiffel Tower")
	}

	placeDetails1, err := client.GetPlaceDetails(searchResp.Data[0].PlaceID)
	if err != nil {
		t.Fatalf("Failed to get place details: %v", err)
	}

	// Get full trip details to access itinerary
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}
	if len(trip.TripPlan.Itinerary.Sections) == 0 {
		t.Fatal("No sections in trip")
	}
	sectionID := trip.TripPlan.Itinerary.Sections[0].ID

	addReq1 := AddPlaceRequest{
		Place: AddPlaceInfo{
			PlaceID: placeDetails1.Data.Details.PlaceID,
			Name:    placeDetails1.Data.Details.Name,
			Geometry: &models.PlaceGeometry{
				Location: models.PlaceLocation{
					Lat: placeDetails1.Data.Details.Geometry.Location.Lat,
					Lng: placeDetails1.Data.Details.Geometry.Location.Lng,
				},
			},
		},
		Text: "Iconic iron tower with amazing views of Paris",
	}

	if err := client.AddPlace(tripKey, sectionID, addReq1); err != nil {
		t.Fatalf("Failed to add first place: %v", err)
	}
	t.Logf("✅ Added: %s", placeDetails1.Data.Details.Name)

	snapshot2 := captureSnapshot(t, client, tripKey, "2-after-first-place")
	t.Logf("📸 Snapshot 2: Trip after adding first place (%d bytes)", len(snapshot2))
	showDiff(t, snapshot1, snapshot2, "Initial → After First Place")

	// Step 3: Add second place
	time.Sleep(1 * time.Second)
	t.Log("\n📸 STEP 3: Adding second place")

	searchResp2, err := client.SearchPlacesWithWanderllog("Louvre Museum Paris", 48.8606, 2.3376)
	if err != nil || len(searchResp2.Data) == 0 {
		t.Fatal("Failed to find Louvre Museum")
	}

	placeDetails2, err := client.GetPlaceDetails(searchResp2.Data[0].PlaceID)
	if err != nil {
		t.Fatalf("Failed to get place details: %v", err)
	}

	addReq2 := AddPlaceRequest{
		Place: AddPlaceInfo{
			PlaceID: placeDetails2.Data.Details.PlaceID,
			Name:    placeDetails2.Data.Details.Name,
			Geometry: &models.PlaceGeometry{
				Location: models.PlaceLocation{
					Lat: placeDetails2.Data.Details.Geometry.Location.Lat,
					Lng: placeDetails2.Data.Details.Geometry.Location.Lng,
				},
			},
		},
		Text: "World's largest art museum",
	}

	if err := client.AddPlace(tripKey, sectionID, addReq2); err != nil {
		t.Fatalf("Failed to add second place: %v", err)
	}
	t.Logf("✅ Added: %s", placeDetails2.Data.Details.Name)

	snapshot3 := captureSnapshot(t, client, tripKey, "3-after-second-place")
	t.Logf("📸 Snapshot 3: Trip after adding second place (%d bytes)", len(snapshot3))
	showDiff(t, snapshot2, snapshot3, "After First Place → After Second Place")

	// Step 4: Remove first place
	time.Sleep(1 * time.Second)
	t.Log("\n📸 STEP 4: Removing first place")

	// Get current trip to find the place block ID
	currentTrip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}

	if len(currentTrip.TripPlan.Itinerary.Sections) == 0 || len(currentTrip.TripPlan.Itinerary.Sections[0].Blocks) == 0 {
		t.Fatal("No blocks to remove")
	}

	firstBlock := currentTrip.TripPlan.Itinerary.Sections[0].Blocks[0]
	if err := client.RemovePlace(tripKey, sectionID, firstBlock.ID); err != nil {
		t.Fatalf("Failed to remove place: %v", err)
	}
	t.Logf("✅ Removed first place")

	snapshot4 := captureSnapshot(t, client, tripKey, "4-after-removal")
	t.Logf("📸 Snapshot 4: Trip after removing first place (%d bytes)", len(snapshot4))
	showDiff(t, snapshot3, snapshot4, "After Second Place → After Removal")

	t.Log("\n🎉 Snapshot test completed successfully!")
	t.Logf("🌐 View trip: https://wanderlog.com/view/%s", tripKey)
}

// captureSnapshot fetches the trip from the server and returns the raw JSON as a pretty-printed string
func captureSnapshot(t *testing.T, client *Client, tripKey, label string) string {
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to capture snapshot '%s': %v", label, err)
	}

	// Pretty print the JSON
	jsonBytes, err := json.MarshalIndent(trip, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal snapshot '%s': %v", label, err)
	}

	// Optionally save to file for inspection
	if os.Getenv("SAVE_SNAPSHOTS") == "1" {
		filename := fmt.Sprintf("/tmp/snapshot_%s.json", label)
		if err := os.WriteFile(filename, jsonBytes, 0644); err != nil {
			t.Logf("Warning: Failed to save snapshot to %s: %v", filename, err)
		} else {
			t.Logf("💾 Saved snapshot to %s", filename)
		}
	}

	return string(jsonBytes)
}

// showDiff displays the unified diff between two snapshots
func showDiff(t *testing.T, before, after, label string) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(before),
		B:        difflib.SplitLines(after),
		FromFile: "Before",
		ToFile:   "After",
		Context:  3,
	}

	diffText, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		t.Logf("Warning: Failed to generate diff for '%s': %v", label, err)
		return
	}

	if diffText == "" {
		t.Logf("📊 Diff (%s): No changes detected", label)
		return
	}

	t.Logf("\n📊 Diff (%s):\n%s", label, diffText)
}
