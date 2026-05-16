package wanderlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchLodgings(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/geo/countries") {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
				return
			}
			if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/geo/listGeosWithSearchedCategories") {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"success":true,"data":[{"id":1,"name":"Tokyo","bounds":[138.29911,34.57763,141.24051,36.44085]}]}`))
				return
			}
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/lodging/searchLodgings") {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decoding body: %v", err)
			}
			if _, ok := payload["bounds"]; !ok {
				t.Fatalf("expected bounds in request body: %s", string(body))
			}
			guests, ok := payload["guests"].(map[string]any)
			if !ok {
				t.Fatalf("expected nested guests in request body: %s", string(body))
			}
			if _, ok := guests["childrenAges"]; !ok {
				t.Fatalf("expected guests.childrenAges in request body: %s", string(body))
			}
			if payload["sources"] != nil {
				t.Fatalf("expected app-compatible sources=null in request body: %s", string(body))
			}
			if payload["sortBy"] != "ratings" {
				t.Fatalf("expected app-compatible sortBy=ratings in request body: %s", string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"isComplete":true,"offers":[{"offerId":"offer-1","source":"google","priceRate":{"amount":367,"currencyCode":"CHF"},"lodging":{"id":{"lodgingId":"prop-1","type":"google"},"name":"Tokyo Hotel","rating":{"source":"Google","value":4.8},"images":[{"url":"https://example.com/hotel.jpg"}]}}]}}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		result, err := client.SearchLodgings("Tokyo", "2026-06-01", "2026-06-07", 1)
		if err != nil {
			t.Fatalf("SearchLodgings: %v", err)
		}
		if !result.Success {
			t.Error("expected success")
		}
		if len(result.Data) != 1 || result.Data[0].PropertyID != "prop-1" || result.Data[0].Name != "Tokyo Hotel" {
			t.Fatalf("expected offers to be normalized into data, got %+v", result.Data)
		}
		if result.Data[0].Rating != 4.8 || result.Data[0].PricePerNight != "367" || result.Data[0].Currency != "CHF" {
			t.Fatalf("expected nested lodging fields to parse, got %+v", result.Data[0])
		}
	})

	t.Run("api returns success=false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if writeTestGeoEndpoints(w, r) {
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":false,"error":"no results"}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		_, err := client.SearchLodgings("Nowhere", "2026-06-01", "2026-06-07", 1)
		if err == nil {
			t.Fatal("expected error for success=false")
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if writeTestGeoEndpoints(w, r) {
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		_, err := client.SearchLodgings("Tokyo", "2026-06-01", "2026-06-07", 1)
		if err == nil {
			t.Fatal("expected error for non-200")
		}
	})
}

func writeTestGeoEndpoints(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/geo/countries") {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
		return true
	}
	if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/geo/listGeosWithSearchedCategories") {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[{"id":1,"name":"Tokyo","bounds":[138.29911,34.57763,141.24051,36.44085]},{"id":2,"name":"Nowhere","bounds":[1,2,3,4]}]}`))
		return true
	}
	return false
}

func TestGetGooglePriceRates(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/lodging/getGooglePriceRates") {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			if r.URL.Query().Get("id") != "prop-123" {
				t.Errorf("expected id=prop-123, got %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"propertyId":"prop-123","rates":[]}}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		result, err := client.GetGooglePriceRates("prop-123")
		if err != nil {
			t.Fatalf("GetGooglePriceRates: %v", err)
		}
		if !result.Success {
			t.Error("expected success")
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		_, err := client.GetGooglePriceRates("bad-prop")
		if err == nil {
			t.Fatal("expected error for bad request")
		}
	})
}
