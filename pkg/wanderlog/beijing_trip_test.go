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

// TestBeijingTripCreation creates a complete week-long trip to Beijing using search
// Run with: go test -v -tags=integration -run TestBeijingTripCreation -timeout 30m ./pkg/wanderlog
func TestBeijingTripCreation(t *testing.T) {
	// Initialize config to load credentials from config file
	if err := InitConfig(); err != nil {
		t.Logf("Warning: Failed to initialize config: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	client := NewClient()
	client.SetLogger(logger)

	// Use the same authentication as the CLI
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first or set credentials in config file", err)
	}

	t.Log("=== Creating Beijing Week-Long Trip ===")

	// WORKAROUND: CreateTrip has a server bug, so we copy an existing trip instead
	// TODO: Fix CreateTrip endpoint once server issue is resolved
	t.Log("Copying trip from template (CreateTrip endpoint has server issues)")
	copyResp, err := client.CopyTrip("vetyiadvqjgikbvx") // Copy the test trip
	if err != nil {
		t.Fatalf("Failed to copy trip: %v", err)
	}

	tripKey := copyResp.TripPlan.Key
	t.Logf("✅ Copied trip: %s (Key: %s)", copyResp.TripPlan.Title, tripKey)

	// Set trip dates for a week-long Beijing trip
	now := time.Now()
	startDate := now.AddDate(0, 1, 0)     // Start next month
	endDate := startDate.AddDate(0, 0, 7) // 7-day trip

	t.Logf("Setting trip dates: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	updateReq := UpdateTripRequest{
		Title:     fmt.Sprintf("Week in Beijing - %s", now.Format("Jan 2006")),
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Privacy:   "private",
	}

	if err := client.UpdateTrip(tripKey, updateReq); err != nil {
		t.Fatalf("Failed to update trip dates: %v", err)
	}
	t.Log("✅ Updated trip with dates")

	// Wait for server to process the update
	time.Sleep(2 * time.Second)

	// Get the trip to see the new daily sections
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}

	if len(trip.TripPlan.Itinerary.Sections) == 0 {
		t.Fatal("Trip has no sections")
	}

	t.Logf("Trip now has %d sections (should have 7 daily sections + 2 default sections)", len(trip.TripPlan.Itinerary.Sections))

	// Beijing coordinates
	beijingLat := 39.9042
	beijingLng := 116.4074

	// Define daily attractions to search for
	dailyQueries := [][]struct{ query, notes string }{
		{ // Day 1
			{"Forbidden City Beijing", "Start early at 8:30 AM. Explore Imperial Palace (3-4 hours)."},
			{"Jingshan Park Beijing", "Panoramic views of Forbidden City. Best for sunset!"},
			{"Wangfujing Street Beijing", "Evening food street. Try traditional snacks."},
		},
		{ // Day 2
			{"Temple of Heaven Beijing", "Morning Tai Chi. Explore Echo Wall and Hall of Prayer."},
			{"798 Art District Beijing", "Contemporary art in former factories. Great cafes."},
		},
		{ // Day 3
			{"Mutianyu Great Wall Beijing", "FULL DAY: Leave at 7 AM. Cable car up, toboggan down!"},
		},
		{ // Day 4
			{"Summer Palace Beijing", "Imperial gardens. Rent boat on Kunming Lake."},
			{"Houhai Lake Beijing", "Evening lakeside bars and paddleboats."},
		},
		{ // Day 5
			{"Lama Temple Beijing", "Tibetan Buddhist temple. 18m sandalwood Buddha."},
			{"Nanluoguxiang Hutong Beijing", "Historic alley. Shops and traditional architecture."},
		},
		{ // Day 6
			{"Tiananmen Square Beijing", "World's largest square. Bring ID for security."},
			{"National Museum of China Beijing", "Free admission. 2-3 hours. Closed Mondays."},
		},
		{ // Day 7
			{"Beijing Zoo Panda House", "See pandas in morning when most active."},
			{"Silk Street Market Beijing", "Final souvenirs. Haggle hard!"},
		},
	}

	// Capture initial snapshot
	t.Log("\n=== Taking Initial Snapshot ===")
	prevSnapshot := captureBeijingSnapshot(t, client, tripKey, "0-initial")

	// Find daily sections (those with dates)
	dailySections := []ItSections{}
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.Date != nil && *section.Date != "" {
			dailySections = append(dailySections, section)
		}
	}

	if len(dailySections) < 7 {
		t.Logf("Warning: Expected 7 daily sections, got %d. Available sections:", len(dailySections))
		for i, section := range trip.TripPlan.Itinerary.Sections {
			dateStr := "no date"
			if section.Date != nil {
				dateStr = *section.Date
			}
			t.Logf("  Section %d: %s (date: %s, type: %s)", i, section.Heading, dateStr, section.Type)
		}
	}

	t.Log("\n=== Searching and Adding Places ===")
	totalAdded := 0

	for day, queries := range dailyQueries {
		if day >= len(dailySections) {
			t.Logf("⚠️  Skipping day %d - no daily section available", day+1)
			continue
		}

		section := dailySections[day]
		sectionID := section.ID
		dateStr := ""
		if section.Date != nil {
			dateStr = *section.Date
		}
		t.Logf("\n--- Day %d (%s) - %s ---", day+1, dateStr, section.Heading)

		dayStartSnapshot := prevSnapshot

		for placeIdx, q := range queries {
			t.Logf("🔍 Searching: %s", q.query)

			searchResp, err := client.SearchPlacesWithWanderllog(q.query, beijingLat, beijingLng)
			if err != nil || searchResp == nil || len(searchResp.Data) == 0 {
				t.Logf("⚠️  No results for '%s'", q.query)
				continue
			}

			// Get first result
			autocompleteResult := searchResp.Data[0]
			t.Logf("   Found: %s", autocompleteResult.Description)

			// Get detailed place info to get coordinates
			placeDetails, err := client.GetPlaceDetails(autocompleteResult.PlaceID)
			if err != nil {
				t.Logf("⚠️  Failed to get details for '%s': %v", autocompleteResult.Description, err)
				continue
			}

			addReq := AddPlaceRequest{
				Place: AddPlaceInfo{
					PlaceID: placeDetails.Data.Details.PlaceID,
					Name:    placeDetails.Data.Details.Name,
					Geometry: &models.PlaceGeometry{
						Location: models.PlaceLocation{
							Lat: placeDetails.Data.Details.Geometry.Location.Lat,
							Lng: placeDetails.Data.Details.Geometry.Location.Lng,
						},
					},
				},
				Text: q.notes,
			}

			if err := client.AddPlace(tripKey, sectionID, addReq); err != nil {
				t.Logf("❌ Failed to add: %v", err)
				continue
			}

			t.Logf("✅ Added: %s", placeDetails.Data.Details.Name)
			totalAdded++

			// Take snapshot after each place addition
			time.Sleep(500 * time.Millisecond)
			newSnapshot := captureBeijingSnapshot(t, client, tripKey, fmt.Sprintf("day%d-place%d", day+1, placeIdx+1))
			showBeijingDiff(t, prevSnapshot, newSnapshot, fmt.Sprintf("Day %d - After adding '%s'", day+1, placeDetails.Data.Details.Name))
			prevSnapshot = newSnapshot
		}

		// Show cumulative diff for the day
		if len(queries) > 0 {
			showBeijingDiff(t, dayStartSnapshot, prevSnapshot, fmt.Sprintf("Day %d - Complete", day+1))
		}
	}

	t.Logf("\n🎉 Added %d places", totalAdded)
	t.Logf("🌐 https://wanderlog.com/view/%s", tripKey)
}

// captureBeijingSnapshot fetches the trip and returns the raw JSON
func captureBeijingSnapshot(t *testing.T, client *Client, tripKey, label string) string {
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to capture snapshot '%s': %v", label, err)
	}

	jsonBytes, err := json.MarshalIndent(trip, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal snapshot '%s': %v", label, err)
	}

	// Save to file if requested
	if os.Getenv("SAVE_SNAPSHOTS") == "1" {
		filename := fmt.Sprintf("/tmp/beijing_snapshot_%s.json", label)
		if err := os.WriteFile(filename, jsonBytes, 0644); err != nil {
			t.Logf("Warning: Failed to save snapshot to %s: %v", filename, err)
		} else {
			t.Logf("💾 Saved snapshot to %s", filename)
		}
	}

	return string(jsonBytes)
}

// showBeijingDiff displays the unified diff between two snapshots
func showBeijingDiff(t *testing.T, before, after, label string) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(before),
		B:        difflib.SplitLines(after),
		FromFile: "Before",
		ToFile:   "After",
		Context:  2,
	}

	diffText, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		t.Logf("Warning: Failed to generate diff for '%s': %v", label, err)
		return
	}

	if diffText == "" {
		return
	}

	// Only show a summary to avoid overwhelming output
	lines := difflib.SplitLines(diffText)
	if len(lines) > 50 {
		t.Logf("\n📊 Diff (%s): %d lines changed (showing first 30 lines)", label, len(lines))
		for i := 0; i < 30 && i < len(lines); i++ {
			t.Log(lines[i])
		}
		t.Logf("... (%d more lines)", len(lines)-30)
	} else {
		t.Logf("\n📊 Diff (%s):\n%s", label, diffText)
	}
}
