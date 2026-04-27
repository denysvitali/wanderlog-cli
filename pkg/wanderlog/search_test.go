package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSearchPlaces(t *testing.T) {
	t.Run("successful wanderlog places search", func(t *testing.T) {
		lat, lng := 40.71, -74.00
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}

			var reqBody map[string]any
			if err := json.Unmarshal([]byte(r.URL.Query().Get("request")), &reqBody); err != nil {
				t.Fatalf("decode request query: %v", err)
			}
			if reqBody["input"] != "coffee" {
				t.Errorf("unexpected request query: %+v", reqBody)
			}
			location, ok := reqBody["location"].(map[string]any)
			if !ok || location["latitude"] != lat || location["longitude"] != lng {
				t.Errorf("unexpected location: %+v", reqBody["location"])
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":[{"place_id":"place-1","description":"Coffee Shop, 123 Main St","types":["cafe"],"structured_formatting":{"main_text":"Coffee Shop","secondary_text":"123 Main St"}}]}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)

		resp, err := client.SearchPlaces("coffee", &lat, &lng)
		if err != nil {
			t.Fatalf("SearchPlaces: %v", err)
		}
		if !resp.Success || len(resp.Places) != 1 {
			t.Fatalf("unexpected response: %+v", resp)
		}
		if resp.Places[0].Name != "Coffee Shop" || resp.Places[0].PlaceID != "place-1" {
			t.Errorf("unexpected place: %+v", resp.Places[0])
		}
		if resp.Places[0].Address != "123 Main St" {
			t.Errorf("unexpected address: %+v", resp.Places[0])
		}
	})
}

func TestSearchRestaurants(t *testing.T) {
	t.Run("uses wanderlog places search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/placesAPI/autocomplete/v2" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":[{"place_id":"restaurant-1","description":"Pizza Place, 456 Main St","types":["restaurant"],"structured_formatting":{"main_text":"Pizza Place","secondary_text":"456 Main St"}}]}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)

		resp, err := client.SearchRestaurants("pizza", nil, nil)
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

func TestSearchPlacesWithWanderlog(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			resp := WanderlogAutocompleteResponse{
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

		result, err := client.SearchPlacesWithWanderlog("test", 40.71, -74.00)
		if err != nil {
			t.Fatalf("SearchPlacesWithWanderlog: %v", err)
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
