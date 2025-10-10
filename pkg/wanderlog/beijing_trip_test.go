//go:build integration
// +build integration

package wanderlog

import (
	"fmt"
	"testing"
	"time"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/sirupsen/logrus"
)

// TestBeijingTripCreation creates a complete week-long trip to Beijing using search
// Run with: go test -v -tags=integration -run TestBeijingTripCreation -timeout 30m ./pkg/wanderlog
func TestBeijingTripCreation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := NewClient()
	client.SetLogger(logger)

	// Use the same authentication as the CLI
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first", err)
	}

	// Calculate dates for next month
	now := time.Now()
	startDate := now.AddDate(0, 1, 0)
	endDate := startDate.AddDate(0, 0, 7)

	t.Log("=== Creating Beijing Week-Long Trip ===")

	// Create the trip
	createReq := CreateTripRequest{
		Title:     fmt.Sprintf("Week in Beijing - %s", now.Format("Jan 2006")),
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Privacy:   "private",
	}

	t.Logf("Creating trip from %s to %s", createReq.StartDate, createReq.EndDate)
	createResp, err := client.CreateTrip(createReq)
	if err != nil {
		t.Fatalf("Failed to create trip: %v", err)
	}

	tripKey := createResp.TripPlan.Key
	t.Logf("✅ Created trip: %s (Key: %s)", createResp.TripPlan.Title, tripKey)

	// Get the trip
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}

	if len(trip.TripPlan.Itinerary.Sections) == 0 {
		t.Fatal("Trip has no sections")
	}

	// Beijing coordinates
	beijingLat := 39.9042
	beijingLng := 116.4074

	// Define daily attractions to search for
	dailyQueries := [][]struct{query, notes string}{
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

	t.Log("\n=== Searching and Adding Places ===")
	totalAdded := 0

	for day, queries := range dailyQueries {
		if day >= len(trip.TripPlan.Itinerary.Sections) {
			continue
		}

		sectionID := trip.TripPlan.Itinerary.Sections[day].ID
		t.Logf("\n--- Day %d ---", day+1)

		for _, q := range queries {
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
			time.Sleep(500 * time.Millisecond)
		}
	}

	t.Logf("\n🎉 Added %d places", totalAdded)
	t.Logf("🌐 https://wanderlog.com/view/%s", tripKey)
}
