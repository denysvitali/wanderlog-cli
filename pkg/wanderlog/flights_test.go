package wanderlog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAllAirlines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/flights/allAirlines") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[{"iata":"MU","name":"China Eastern"}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.GetAllAirlines()
	if err != nil {
		t.Fatalf("GetAllAirlines: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}

func TestAutocompleteAirport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[{"iata":"JFK","name":"John F Kennedy"}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.AutocompleteAirport("New York")
	if err != nil {
		t.Fatalf("AutocompleteAirport: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}

func TestAutocompleteAirportWithLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "latitude=") {
			t.Errorf("expected lat/lng query params, got: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[{"iata":"LHR","name":"Heathrow"}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.AutocompleteAirportWithLocation("London", 51.5, -0.1)
	if err != nil {
		t.Fatalf("AutocompleteAirportWithLocation: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}

func TestGetFlightStops(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		result, err := client.GetFlightStops("244", "MU", "2026-05-11")
		if err != nil {
			t.Fatalf("GetFlightStops: %v", err)
		}
		if !result.Success {
			t.Error("expected success")
		}
	})

	t.Run("html response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body>Error</body></html>`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		_, err := client.GetFlightStops("244", "MU", "2026-05-11")
		if err == nil || !strings.Contains(err.Error(), "HTML") {
			t.Fatalf("expected HTML error, got: %v", err)
		}
	})
}
