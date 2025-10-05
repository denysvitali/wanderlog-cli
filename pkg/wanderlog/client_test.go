package wanderlog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGetTrip(t *testing.T) {
	tests := []struct {
		name           string
		tripKey        string
		serverResponse string
		serverStatus   int
		expectError    bool
		checkTrip      func(*testing.T, *TripResponse)
	}{
		{
			name:    "successful fetch",
			tripKey: "test-trip-key",
			serverResponse: `{
				"tripPlan": {
					"id": 123,
					"key": "test-trip-key",
					"title": "Test Trip",
					"itinerary": {
						"sections": []
					}
				},
				"resources": {
					"placeMetadata": []
				}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
			checkTrip: func(t *testing.T, trip *TripResponse) {
				if trip.TripPlan.ID != 123 {
					t.Errorf("Expected trip ID 123, got %d", trip.TripPlan.ID)
				}
				if trip.TripPlan.Key != "test-trip-key" {
					t.Errorf("Expected key 'test-trip-key', got '%s'", trip.TripPlan.Key)
				}
			},
		},
		{
			name:           "not found",
			tripKey:        "nonexistent",
			serverResponse: `{"error": "not found"}`,
			serverStatus:   http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "invalid json",
			tripKey:        "test",
			serverResponse: `{invalid json}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, tt.tripKey) {
					t.Errorf("Expected path to contain '%s', got %s", tt.tripKey, r.URL.Path)
				}
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Override BaseURL for testing
			oldBaseURL := BaseURL
			BaseURL = server.URL
			defer func() { BaseURL = oldBaseURL }()

			// Create client
			client := NewClient()
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			client.SetLogger(logger)

			// Call function
			trip, err := client.GetTrip(tt.tripKey)

			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check trip
			if !tt.expectError && tt.checkTrip != nil {
				tt.checkTrip(t, trip)
			}
		})
	}
}

func TestMatchesQuery(t *testing.T) {
	client := NewClient()

	description1 := "A beautiful historic landmark"
	description2 := "Modern art museum"

	tests := []struct {
		name        string
		place       Metadata
		query       string
		shouldMatch bool
	}{
		{
			name: "match by name",
			place: Metadata{
				Name:    "Eiffel Tower",
				Address: "Paris, France",
			},
			query:       "eiffel",
			shouldMatch: true,
		},
		{
			name: "match by address",
			place: Metadata{
				Name:    "Some Place",
				Address: "123 Main St, New York",
			},
			query:       "new york",
			shouldMatch: true,
		},
		{
			name: "match by category",
			place: Metadata{
				Name:       "Restaurant",
				Categories: []string{"food", "italian", "dining"},
			},
			query:       "italian",
			shouldMatch: true,
		},
		{
			name: "match by description",
			place: Metadata{
				Name:        "Historic Site",
				Description: &description1,
			},
			query:       "historic",
			shouldMatch: true,
		},
		{
			name: "match by generated description",
			place: Metadata{
				Name:                 "Museum",
				GeneratedDescription: &description2,
			},
			query:       "modern art",
			shouldMatch: true,
		},
		{
			name: "no match",
			place: Metadata{
				Name:    "Tokyo Tower",
				Address: "Tokyo, Japan",
			},
			query:       "paris",
			shouldMatch: false,
		},
		{
			name: "case insensitive match",
			place: Metadata{
				Name: "STATUE OF LIBERTY",
			},
			query:       "statue",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.matchesQuery(tt.place, tt.query)
			if result != tt.shouldMatch {
				t.Errorf("Expected matchesQuery to return %v, got %v", tt.shouldMatch, result)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	if client.logger == nil {
		t.Error("Expected logger to be initialized")
	}

	if client.userAgent != DefaultUserAgent {
		t.Errorf("Expected userAgent to be '%s', got '%s'", DefaultUserAgent, client.userAgent)
	}

	if client.httpClient.Timeout == 0 {
		t.Error("Expected timeout to be set")
	}
}

func TestSetLogger(t *testing.T) {
	client := NewClient()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	client.SetLogger(logger)

	if client.logger != logger {
		t.Error("SetLogger did not set the logger correctly")
	}
}
