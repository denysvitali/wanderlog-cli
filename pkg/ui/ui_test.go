package ui

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

func testTripResponse() *wanderlog.TripResponse {
	return &wanderlog.TripResponse{
		Success: true,
		TripPlan: wanderlog.Plan{
			Title:      "Test Trip",
			StartDate:  "2026-05-01",
			EndDate:    "2026-05-07",
			Days:       7,
			PlaceCount: 5,
			ViewCount:  10,
			LikeCount:  3,
			Itinerary: wanderlog.Itinerary{
				Sections: []wanderlog.ItSections{},
			},
		},
	}
}

func testPlaces() []wanderlog.Metadata {
	desc := "A beautiful place to visit"
	return []wanderlog.Metadata{
		{
			Name:                     "Test Place",
			Address:                  "123 Test St",
			Rating:                   4.5,
			Categories:               []string{"restaurant", "cafe"},
			Website:                  "https://example.com",
			InternationalPhoneNumber: "+1-555-1234",
			Description:              &desc,
		},
	}
}

func testSearchResults() []wanderlog.SearchResult {
	return []wanderlog.SearchResult{
		{
			Name:        "Search Result 1",
			Address:     "456 Search Ave",
			Rating:      4.0,
			Categories:  []string{"park"},
			Latitude:    40.7128,
			Longitude:   -74.0060,
			PlaceID:     "ChIJ-test",
			Description: "A nice park",
		},
	}
}

func TestPrintJSON(t *testing.T) {
	out := captureOutput(func() {
		PrintJSON(map[string]string{"key": "value"})
	})
	if !strings.Contains(out, `"key"`) || !strings.Contains(out, `"value"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestPrintTrip(t *testing.T) {
	t.Run("failed trip", func(t *testing.T) {
		trip := &wanderlog.TripResponse{Success: false}
		out := captureOutput(func() { PrintTrip(trip, false) })
		if !strings.Contains(out, "Failed to fetch trip data") {
			t.Errorf("expected failure message, got: %s", out)
		}
	})

	t.Run("successful trip summary", func(t *testing.T) {
		trip := testTripResponse()
		out := captureOutput(func() { PrintTrip(trip, false) })
		if !strings.Contains(out, "Test Trip") {
			t.Errorf("expected trip title, got: %s", out)
		}
		if !strings.Contains(out, "5 places") {
			t.Errorf("expected place count, got: %s", out)
		}
	})

	t.Run("with like count", func(t *testing.T) {
		trip := testTripResponse()
		trip.TripPlan.LikeCount = 5
		out := captureOutput(func() { PrintTrip(trip, false) })
		if !strings.Contains(out, "5 likes") {
			t.Errorf("expected like count, got: %s", out)
		}
	})

	t.Run("with details", func(t *testing.T) {
		trip := testTripResponse()
		out := captureOutput(func() { PrintTrip(trip, true) })
		if !strings.Contains(out, "Detailed Itinerary") {
			t.Errorf("expected detailed itinerary, got: %s", out)
		}
	})
}

func TestPrintPlaces(t *testing.T) {
	t.Run("empty places", func(t *testing.T) {
		out := captureOutput(func() { PrintPlaces([]wanderlog.Metadata{}) })
		if !strings.Contains(out, "No places found") {
			t.Errorf("expected no places message, got: %s", out)
		}
	})

	t.Run("with places", func(t *testing.T) {
		out := captureOutput(func() { PrintPlaces(testPlaces()) })
		if !strings.Contains(out, "Test Place") {
			t.Errorf("expected place name, got: %s", out)
		}
		if !strings.Contains(out, "123 Test St") {
			t.Errorf("expected address, got: %s", out)
		}
		if !strings.Contains(out, "4.5") {
			t.Errorf("expected rating, got: %s", out)
		}
	})

	t.Run("permanently closed", func(t *testing.T) {
		places := testPlaces()
		places[0].PermanentlyClosed = true
		out := captureOutput(func() { PrintPlaces(places) })
		if !strings.Contains(out, "Permanently Closed") {
			t.Errorf("expected permanently closed message, got: %s", out)
		}
	})

	t.Run("long description truncated", func(t *testing.T) {
		longDesc := ""
		for i := 0; i < 100; i++ {
			longDesc += "abcdefghij" // 10 chars, 100 iterations = 1000 chars
		}
		places := []wanderlog.Metadata{
			{Name: "Long Desc Place", Description: &longDesc},
		}
		out := captureOutput(func() { PrintPlaces(places) })
		if !strings.Contains(out, "...") {
			t.Errorf("expected truncated description, got: %s", out)
		}
	})
}

func TestPrintTripMarkdown(t *testing.T) {
	t.Run("failed trip", func(t *testing.T) {
		trip := &wanderlog.TripResponse{Success: false}
		out := captureOutput(func() { PrintTripMarkdown(trip, false) })
		if !strings.Contains(out, "Trip Data Unavailable") {
			t.Errorf("expected unavailable message, got: %s", out)
		}
	})

	t.Run("successful trip", func(t *testing.T) {
		trip := testTripResponse()
		out := captureOutput(func() { PrintTripMarkdown(trip, false) })
		if !strings.Contains(out, "# Test Trip") {
			t.Errorf("expected markdown title, got: %s", out)
		}
		if !strings.Contains(out, "7 days") {
			t.Errorf("expected duration, got: %s", out)
		}
		if !strings.Contains(out, "10") {
			t.Errorf("expected view count, got: %s", out)
		}
	})
}

func TestPrintPlacesMarkdown(t *testing.T) {
	t.Run("empty places", func(t *testing.T) {
		out := captureOutput(func() { PrintPlacesMarkdown([]wanderlog.Metadata{}) })
		if !strings.Contains(out, "No places found") {
			t.Errorf("expected no places message, got: %s", out)
		}
	})

	t.Run("with places", func(t *testing.T) {
		out := captureOutput(func() { PrintPlacesMarkdown(testPlaces()) })
		if !strings.Contains(out, "Test Place") {
			t.Errorf("expected place name, got: %s", out)
		}
		if !strings.Contains(out, "123 Test St") {
			t.Errorf("expected address, got: %s", out)
		}
	})
}

func TestPrintSearchResults(t *testing.T) {
	t.Run("empty results", func(t *testing.T) {
		out := captureOutput(func() { PrintSearchResults([]wanderlog.SearchResult{}) })
		if !strings.Contains(out, "No places found") {
			t.Errorf("expected no results message, got: %s", out)
		}
	})

	t.Run("with results", func(t *testing.T) {
		out := captureOutput(func() { PrintSearchResults(testSearchResults()) })
		if !strings.Contains(out, "Search Result 1") {
			t.Errorf("expected result name, got: %s", out)
		}
		if !strings.Contains(out, "456 Search Ave") {
			t.Errorf("expected address, got: %s", out)
		}
	})
}

func TestPrintSearchResultsMarkdown(t *testing.T) {
	t.Run("empty results", func(t *testing.T) {
		out := captureOutput(func() { PrintSearchResultsMarkdown([]wanderlog.SearchResult{}) })
		if !strings.Contains(out, "No places found") {
			t.Errorf("expected no results message, got: %s", out)
		}
	})

	t.Run("with results", func(t *testing.T) {
		out := captureOutput(func() { PrintSearchResultsMarkdown(testSearchResults()) })
		if !strings.Contains(out, "Search Result 1") {
			t.Errorf("expected result name, got: %s", out)
		}
		if !strings.Contains(out, "40.7128") {
			t.Errorf("expected latitude, got: %s", out)
		}
	})
}
