package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchLodgings(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/lodging/searchLodgings") {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
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
	})

	t.Run("api returns success=false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestGetGooglePriceRates(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/lodging/getGooglePriceRates") {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			// Verify propertyId is in the body
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["propertyId"] != "prop-123" {
				t.Errorf("expected propertyId=prop-123, got %v", body)
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
