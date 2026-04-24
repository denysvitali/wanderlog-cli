//go:build integration
// +build integration

package wanderlog

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// Integration tests for write operations against the real Wanderlog API.
// Run with: WANDERLOG_RUN_PROD_INTEGRATION=1 go test -v -tags=integration ./pkg/wanderlog

const testTripID = "vetyiadvqjgikbvx"

func setupIntegrationClient(t *testing.T) *Client {
	requireProductionIntegrationOptIn(t)

	// Initialize config to load credentials from config file
	if err := InitConfig(); err != nil {
		t.Logf("Warning: Failed to initialize config: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first or set credentials in config file", err)
	}

	return client
}

func getTestTripID() string {
	if tripID := os.Getenv("WANDERLOG_TEST_TRIP_ID"); tripID != "" {
		return tripID
	}
	return testTripID
}

func TestIntegration_CreateAndDeleteTrip(t *testing.T) {
	client := setupIntegrationClient(t)

	createReq := CreateTripRequest{
		Title:               "Integration Test Trip",
		GeoIDs:              []int{1},
		InitialMapsPlaceIDs: []int{},
		Type:                "plan",
		StartDate:           "2025-11-01",
		EndDate:             "2025-11-07",
		Privacy:             "private",
	}

	createResp, err := client.CreateTrip(createReq)
	if err != nil || !createResp.Success {
		t.Fatalf("Failed to create trip: %v", err)
	}

	tripKey := createResp.TripPlan.Key
	t.Logf("Created trip: %s", tripKey)

	if err := client.DeleteTrip(tripKey); err != nil {
		t.Fatalf("Failed to delete trip: %v", err)
	}
}

func TestIntegration_CopyTrip(t *testing.T) {
	client := setupIntegrationClient(t)
	sourceTripID := getTestTripID()

	copyResp, err := client.CopyTrip(sourceTripID)
	if err != nil {
		t.Fatalf("Failed to copy trip: %v", err)
	}

	t.Logf("Copied to: %s", copyResp.TripPlan.Key)
	_ = client.DeleteTrip(copyResp.TripPlan.Key)
}

func TestIntegration_GetLikeCount(t *testing.T) {
	client := setupIntegrationClient(t)
	tripID := getTestTripID()

	likeCount, err := client.GetLikeCount(tripID)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	t.Logf("Likes: %d, User liked: %v", likeCount.Count, likeCount.UserLiked)
}

func TestIntegration_GetTripFlights(t *testing.T) {
	client := setupIntegrationClient(t)
	tripID := getTestTripID()

	flightsResp, err := client.GetTripFlights(tripID)
	if err != nil {
		t.Fatalf("Failed to get trip flights: %v", err)
	}

	t.Logf("Flight count: %d", len(flightsResp.Data.Flights))
	for i, flight := range flightsResp.Data.Flights {
		t.Logf("  Flight %d: %s %s -> %s", i+1, flight.Airline, flight.FlightNumber, flight.Destination.IATA)
	}
}

func TestIntegration_ExportTrip(t *testing.T) {
	client := setupIntegrationClient(t)
	tripID := getTestTripID()

	exportResp, err := client.ExportTrip(tripID)
	if err != nil {
		t.Fatalf("Failed to export trip: %v", err)
	}

	t.Logf("Export URL: %s", exportResp.URL)
	if exportResp.URL == "" {
		t.Logf("Export Data URL: %s", exportResp.Data.ExportURL)
	}
}

func TestIntegration_AutofillDay(t *testing.T) {
	client := setupIntegrationClient(t)
	createResp, err := client.CreateTrip(CreateTripRequest{
		Title:               "Autofill Integration Test Trip",
		GeoIDs:              []int{1},
		InitialMapsPlaceIDs: []int{},
		Type:                "plan",
		StartDate:           "2025-11-01",
		EndDate:             "2025-11-02",
		Privacy:             "private",
	})
	if err != nil {
		t.Fatalf("Failed to create trip: %v", err)
	}
	tripID := createResp.TripPlan.Key
	defer func() { _ = client.DeleteTrip(tripID) }()

	// First get the full trip to find a dated section.
	trip, err := client.GetTrip(tripID)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}

	sectionID := 0
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.Date != nil && *section.Date != "" {
			sectionID = section.ID
			break
		}
	}
	if sectionID == 0 {
		t.Skip("No dated sections found in trip, skipping autofill test")
	}
	t.Logf("Testing autofill for section ID: %d", sectionID)

	autofillResp, err := client.AutofillDay(tripID, sectionID, "restaurant")
	if err != nil {
		t.Fatalf("Failed to autofill day: %v", err)
	}

	t.Logf("Got %d suggestions", len(autofillResp.Data.Suggestions))
	for i, suggestion := range autofillResp.Data.Suggestions {
		t.Logf("  Suggestion %d: %s (%s)", i+1, suggestion.Name, suggestion.Address)
	}
}

func TestIntegration_AddChecklistItems(t *testing.T) {
	client := setupIntegrationClient(t)
	createResp, err := client.CreateTrip(CreateTripRequest{
		Title:               "Checklist Integration Test Trip",
		GeoIDs:              []int{1},
		InitialMapsPlaceIDs: []int{},
		Type:                "plan",
		StartDate:           "2025-11-01",
		EndDate:             "2025-11-02",
		Privacy:             "private",
	})
	if err != nil {
		t.Fatalf("Failed to create trip: %v", err)
	}
	tripID := createResp.TripPlan.Key
	defer func() { _ = client.DeleteTrip(tripID) }()

	// First get the trip sections to find a valid section ID
	sections, err := client.GetTripSections(tripID)
	if err != nil {
		t.Fatalf("Failed to get trip sections: %v", err)
	}

	if len(sections) == 0 {
		t.Skip("No sections found in trip, skipping checklist test")
	}

	sectionID := sections[0].ID
	t.Logf("Testing checklist for section ID: %d", sectionID)

	items := []ChecklistItem{
		{Text: "Passport", Checked: false, Category: "Documents"},
		{Text: "Phone charger", Checked: false, Category: "Electronics"},
	}

	checklistResp, err := client.AddChecklistItems(tripID, sectionID, items)
	if err != nil {
		t.Fatalf("Failed to add checklist items: %v", err)
	}

	t.Logf("Added %d items, section now has %d items", len(items), len(checklistResp.Data.Section.Items))
	for i, item := range checklistResp.Data.Section.Items {
		t.Logf("  Item %d: %s (checked: %v)", i+1, item.Text, item.Checked)
	}
}
