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

	t.Run("successful google places search", func(t *testing.T) {
		lat, lng := 40.71, -74.00
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("X-Goog-Api-Key") != "test-key" {
				t.Errorf("missing API key header")
			}
			if r.Header.Get("X-Goog-FieldMask") == "" {
				t.Errorf("missing field mask header")
			}

			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["textQuery"] != "coffee" {
				t.Errorf("unexpected query body: %+v", body)
			}
			if _, ok := body["locationBias"]; !ok {
				t.Errorf("expected location bias in body")
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"places":[{"id":"place-1","displayName":{"text":"Coffee Shop"},"formattedAddress":"123 Main St","location":{"latitude":40.71,"longitude":-74},"rating":4.5,"types":["cafe"]}]}`))
		}))
		defer server.Close()

		client := newClientNoAuth(t)
		client.httpClient = newRedirectClient(server)

		resp, err := client.SearchPlaces("coffee", &lat, &lng, "test-key")
		if err != nil {
			t.Fatalf("SearchPlaces: %v", err)
		}
		if !resp.Success || len(resp.Places) != 1 {
			t.Fatalf("unexpected response: %+v", resp)
		}
		if resp.Places[0].Name != "Coffee Shop" || resp.Places[0].PlaceID != "place-1" {
			t.Errorf("unexpected place: %+v", resp.Places[0])
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

	t.Run("adds restaurant type filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if body["includedType"] != "restaurant" {
				t.Errorf("expected restaurant type filter, got %+v", body)
			}
			if !strings.Contains(r.Header.Get("X-Goog-FieldMask"), "places.regularOpeningHours") {
				t.Errorf("expected restaurant field mask, got %q", r.Header.Get("X-Goog-FieldMask"))
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"places":[{"id":"restaurant-1","displayName":{"text":"Pizza Place"},"formattedAddress":"456 Main St","location":{"latitude":41,"longitude":-75},"rating":4.2,"types":["restaurant"]}]}`))
		}))
		defer server.Close()

		client := newClientNoAuth(t)
		client.httpClient = newRedirectClient(server)

		resp, err := client.SearchRestaurants("pizza", nil, nil, "test-key")
		if err != nil {
			t.Fatalf("SearchRestaurants: %v", err)
		}
		if !resp.Success || len(resp.Places) != 1 {
			t.Fatalf("unexpected response: %+v", resp)
		}
		if resp.Places[0].Name != "Pizza Place" || resp.Places[0].PlaceID != "restaurant-1" {
			t.Errorf("unexpected restaurant: %+v", resp.Places[0])
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
						MainText                  string `json:"main_text"`
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
