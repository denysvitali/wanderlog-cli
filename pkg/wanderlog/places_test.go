package wanderlog

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// roundTripperFunc rewrites all outgoing requests to the test server URL.
// This is needed for functions that hardcode the API URL instead of using BaseURL.
type roundTripperFunc struct {
	target *url.URL
}

func (r *roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = r.target.Scheme
	req.URL.Host = r.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func newRedirectClient(server *httptest.Server) *http.Client {
	target, _ := url.Parse(server.URL)
	return &http.Client{
		Transport: &roundTripperFunc{target: target},
	}
}

func TestGetPlaceDetails(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"details":{"name":"Test Place","place_id":"ChIJ-test","geometry":{"location":{"lat":40.71,"lng":-74.00}}}}}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		client.httpClient = newRedirectClient(server)

		result, err := client.GetPlaceDetails("ChIJ-test")
		if err != nil {
			t.Fatalf("GetPlaceDetails: %v", err)
		}
		if !result.Success {
			t.Error("expected success")
		}
		if result.Data.Details.Name != "Test Place" {
			t.Errorf("unexpected name: %s", result.Data.Details.Name)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		client.httpClient = newRedirectClient(server)

		_, err := client.GetPlaceDetails("bad-id")
		if err == nil {
			t.Fatal("expected error for server error")
		}
	})
}

func TestGetTripSections(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/tripPlans/test-key/sections") {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":[{"id":1,"heading":"Day 1","_date":"2026-05-01"}]}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		sections, err := client.GetTripSections("test-key")
		if err != nil {
			t.Fatalf("GetTripSections: %v", err)
		}
		if len(sections) != 1 || sections[0].ID != 1 {
			t.Errorf("unexpected sections: %+v", sections)
		}
	})

	t.Run("api returns success=false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":false}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		_, err := client.GetTripSections("bad-key")
		if err == nil {
			t.Fatal("expected error for success=false")
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		_, err := client.GetTripSections("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-200")
		}
	})
}
