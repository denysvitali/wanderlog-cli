//go:build integration
// +build integration

package wanderlog

import (
	"testing"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/sirupsen/logrus"
)

// TestIntegration_GetTripSections tests the GetTripSections endpoint
func TestIntegration_GetTripSections(t *testing.T) {
	requireProductionIntegrationOptIn(t)

	// Initialize config to load credentials
	if err := InitConfig(); err != nil {
		t.Logf("Warning: Failed to initialize config: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := NewClient()
	client.SetLogger(logger)

	// Use the same authentication as the CLI
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first", err)
	}

	// Create a test trip first
	t.Log("Creating test trip...")
	copyResp, err := client.CopyTrip("vetyiadvqjgikbvx")
	if err != nil {
		t.Fatalf("Failed to copy trip: %v", err)
	}
	tripKey := copyResp.TripPlan.Key
	t.Logf("Created trip: %s", tripKey)

	// Clean up at the end
	defer func() {
		t.Log("Cleaning up test trip...")
		if err := client.DeleteTrip(tripKey); err != nil {
			t.Logf("Warning: Failed to delete test trip: %v", err)
		}
	}()

	// Test GetTripSections
	t.Log("Testing GetTripSections...")
	sections, err := client.GetTripSections(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip sections: %v", err)
	}

	if len(sections) == 0 {
		t.Fatal("Expected at least one section")
	}

	t.Logf("✅ Got %d sections", len(sections))
	for i, section := range sections {
		heading := section.Heading
		if heading == "" {
			heading = section.DisplayHeading
		}
		t.Logf("  Section %d: ID=%d, Heading=%s, Blocks=%d",
			i+1, section.ID, heading, len(section.Blocks))
	}

	// Verify sections match what we'd get from GetTrip
	t.Log("Verifying sections match full trip data...")
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get full trip: %v", err)
	}

	if len(sections) != len(trip.TripPlan.Itinerary.Sections) {
		t.Errorf("Section count mismatch: GetTripSections=%d, GetTrip=%d",
			len(sections), len(trip.TripPlan.Itinerary.Sections))
	}

	for i := range sections {
		if sections[i].ID != trip.TripPlan.Itinerary.Sections[i].ID {
			t.Errorf("Section %d ID mismatch: GetTripSections=%d, GetTrip=%d",
				i, sections[i].ID, trip.TripPlan.Itinerary.Sections[i].ID)
		}
		// The sections endpoint returns DisplayHeading while GetTrip returns Heading
		sectionsHeading := sections[i].DisplayHeading
		tripHeading := trip.TripPlan.Itinerary.Sections[i].Heading
		if sectionsHeading != tripHeading {
			t.Logf("Section %d heading differs (expected from different endpoints): sections=%s, trip=%s",
				i, sectionsHeading, tripHeading)
		}
	}

	t.Log("✅ Sections match full trip data")
}

// TestIntegration_UpdateTrip tests the UpdateTrip endpoint
func TestIntegration_UpdateTrip(t *testing.T) {
	requireProductionIntegrationOptIn(t)

	// Initialize config to load credentials
	if err := InitConfig(); err != nil {
		t.Logf("Warning: Failed to initialize config: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	client := NewClient()
	client.SetLogger(logger)

	// Use the same authentication as the CLI
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first", err)
	}

	// Create a test trip first
	t.Log("Creating test trip...")
	copyResp, err := client.CopyTrip("vetyiadvqjgikbvx")
	if err != nil {
		t.Fatalf("Failed to copy trip: %v", err)
	}
	tripKey := copyResp.TripPlan.Key
	originalTitle := copyResp.TripPlan.Title
	t.Logf("Created trip: %s (title: %s)", tripKey, originalTitle)

	// Clean up at the end
	defer func() {
		t.Log("Cleaning up test trip...")
		if err := client.DeleteTrip(tripKey); err != nil {
			t.Logf("Warning: Failed to delete test trip: %v", err)
		}
	}()

	// Test 1: Update title
	t.Log("Test 1: Updating trip title...")
	newTitle := "Updated Test Trip - Integration Test"
	err = client.UpdateTrip(tripKey, models.UpdateTripRequest{
		Title: newTitle,
	})
	if err != nil {
		t.Fatalf("Failed to update trip title: %v", err)
	}

	// Verify title was updated
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}
	if trip.TripPlan.Title != newTitle {
		t.Errorf("Title not updated: expected '%s', got '%s'", newTitle, trip.TripPlan.Title)
	}
	t.Logf("✅ Title updated successfully to: %s", trip.TripPlan.Title)

	// Test 2: Update dates
	t.Log("Test 2: Updating trip dates...")
	newStartDate := "2025-12-01"
	newEndDate := "2025-12-07"
	err = client.UpdateTrip(tripKey, models.UpdateTripRequest{
		StartDate: newStartDate,
		EndDate:   newEndDate,
	})
	if err != nil {
		t.Fatalf("Failed to update trip dates: %v", err)
	}

	// Verify dates were updated
	trip, err = client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}
	if trip.TripPlan.StartDate != newStartDate {
		t.Errorf("Start date not updated: expected '%s', got '%s'", newStartDate, trip.TripPlan.StartDate)
	}
	if trip.TripPlan.EndDate != newEndDate {
		t.Errorf("End date not updated: expected '%s', got '%s'", newEndDate, trip.TripPlan.EndDate)
	}
	t.Logf("✅ Dates updated successfully: %s to %s", trip.TripPlan.StartDate, trip.TripPlan.EndDate)

	// Test 3: Update privacy
	t.Log("Test 3: Updating trip privacy...")
	newPrivacy := "private"
	err = client.UpdateTrip(tripKey, models.UpdateTripRequest{
		Privacy: newPrivacy,
	})
	if err != nil {
		t.Fatalf("Failed to update trip privacy: %v", err)
	}

	// Verify privacy was updated
	trip, err = client.GetTrip(tripKey)
	if err != nil {
		t.Fatalf("Failed to get trip: %v", err)
	}
	if trip.TripPlan.Privacy != newPrivacy {
		t.Errorf("Privacy not updated: expected '%s', got '%s'", newPrivacy, trip.TripPlan.Privacy)
	}
	t.Logf("✅ Privacy updated successfully to: %s", trip.TripPlan.Privacy)

	t.Log("✅ All update tests passed - title, dates, and privacy can be updated individually")
}
