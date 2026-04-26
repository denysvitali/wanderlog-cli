package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSearchPlaces(t *testing.T) {
	t.Run("no api key", func(t *testing.T) {
		client := newClientNoAuth(t)
		resp, err := client.SearchPlaces("test", nil, nil, "")
		if err == nil {
			t.Fatal("expected error for empty API key")
		}
		if resp != nil && resp.Success {
			t.Error("expected success=false when no API key")
		}
	})

	t.Run("missing api key returns error", func(t *testing.T) {
		client := newClientNoAuth(t)
		_, err := client.SearchPlaces("test", nil, nil, "")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestSearchRestaurants(t *testing.T) {
	t.Run("no api key returns error", func(t *testing.T) {
		client := newClientNoAuth(t)
		_, err := client.SearchRestaurants("pizza", nil, nil, "")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestSearchPlacesInTrips(t *testing.T) {
	t.Run("get user trips fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		_, err := client.SearchPlacesInTrips("test")
		if err == nil {
			t.Fatal("expected error when get user trips fails")
		}
	})
}

func TestSearchPlacesWithWanderllog(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			resp := WanderlLogAutocompleteResponse{
				Success: true,
				Data: []struct {
					PlaceID              string   `json:"place_id"`
					Description          string   `json:"description"`
					Types                []string `json:"types"`
					StructuredFormatting struct {
						MainText                  string   `json:"main_text"`
						MainTextMatchedSubstrings []struct {
							Offset int `json:"offset"`
							Length int `json:"length"`
						} `json:"main_text_matched_substrings"`
						SecondaryText string `json:"secondary_text"`
					} `json:"structured_formatting"`
					Type                string        `json:"type,omitempty"`
					Input               string        `json:"input,omitempty"`
					InputTextHighlights []interface{} `json:"inputTextHighlights,omitempty"`
					SeeLocations        bool          `json:"seeLocations,omitempty"`
					SecondaryText       string        `json:"secondaryText,omitempty"`
					CanSeeOnMap         bool          `json:"canSeeOnMap,omitempty"`
				}{
					{PlaceID: "ChIJ-test", Description: "Test Place, City"},
				},
			}
			b, _ := json.Marshal(resp)
			_, _ = w.Write(b)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		client.httpClient = newRedirectClient(server)

		result, err := client.SearchPlacesWithWanderllog("test", 40.71, -74.00)
		if err != nil {
			t.Fatalf("SearchPlacesWithWanderllog: %v", err)
		}
		if !result.Success {
			t.Error("expected success")
		}
		if len(result.Data) == 0 || result.Data[0].PlaceID != "ChIJ-test" {
			t.Errorf("unexpected results: %+v", result.Data)
		}
	})
}

func TestMatchesQueryInSearch(t *testing.T) {
	client := NewClient()

	desc := "A beautiful historic landmark"

	tests := []struct {
		name        string
		place       Metadata
		query       string
		shouldMatch bool
	}{
		{
			name:        "match by name",
			place:       Metadata{Name: "Eiffel Tower"},
			query:       "eiffel",
			shouldMatch: true,
		},
		{
			name:        "match by address",
			place:       Metadata{Name: "Place", Address: "123 Main St, New York"},
			query:       "new york",
			shouldMatch: true,
		},
		{
			name:        "match by category",
			place:       Metadata{Name: "Restaurant", Categories: []string{"italian"}},
			query:       "italian",
			shouldMatch: true,
		},
		{
			name:        "match by description",
			place:       Metadata{Name: "Site", Description: &desc},
			query:       "historic",
			shouldMatch: true,
		},
		{
			name:        "no match",
			place:       Metadata{Name: "Tokyo Tower"},
			query:       "paris",
			shouldMatch: false,
		},
		{
			name:        "case insensitive",
			place:       Metadata{Name: "STATUE OF LIBERTY"},
			query:       "statue",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.matchesQuery(tt.place, tt.query)
			if result != tt.shouldMatch {
				t.Errorf("matchesQuery(%q) = %v, want %v", tt.query, result, tt.shouldMatch)
			}
		})
	}
}

// newClientNoAuth creates a client without auth for testing error paths
func newClientNoAuth(t *testing.T) *Client {
	t.Helper()
	client := NewClient()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	client.SetLogger(logger)
	return client
}
