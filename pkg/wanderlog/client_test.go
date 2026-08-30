package wanderlog

import (
	"io"
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
		{
			name:           "HTTP 200 API failure",
			tripKey:        "denied",
			serverResponse: `{"success":false,"message":"access denied"}`,
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

func TestDoAPI(t *testing.T) {
	t.Run("get request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"ok"}`))
		}))
		defer server.Close()

		oldBaseURL := BaseURL
		BaseURL = server.URL
		defer func() { BaseURL = oldBaseURL }()

		client := NewClient()
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		client.SetLogger(logger)

		status, body, err := client.DoAPI("GET", "/api/test", nil, nil, false)
		if err != nil {
			t.Fatalf("DoAPI: %v", err)
		}
		if status != http.StatusOK {
			t.Errorf("expected status 200, got %d", status)
		}
		if !strings.Contains(string(body), "ok") {
			t.Errorf("unexpected body: %s", string(body))
		}
	})

	t.Run("post with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "key") {
				t.Errorf("unexpected body: %s", string(body))
			}
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		oldBaseURL := BaseURL
		BaseURL = server.URL
		defer func() { BaseURL = oldBaseURL }()

		client := NewClient()
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		client.SetLogger(logger)

		status, _, err := client.DoAPI("POST", "/api/test", []byte(`{"key":"value"}`), nil, false)
		if err != nil {
			t.Fatalf("DoAPI: %v", err)
		}
		if status != http.StatusCreated {
			t.Errorf("expected status 201, got %d", status)
		}
	})

	t.Run("non-200 status returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`not found`))
		}))
		defer server.Close()

		oldBaseURL := BaseURL
		BaseURL = server.URL
		defer func() { BaseURL = oldBaseURL }()

		client := NewClient()
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		client.SetLogger(logger)

		_, _, err := client.DoAPI("GET", "/api/test", nil, nil, false)
		if err == nil {
			t.Fatal("expected error for non-200")
		}
	})
}

func TestDoAPICredentialBoundaries(t *testing.T) {
	creds := &AuthCredentials{SessionCookie: "secret-session", XSRFToken: "secret-xsrf"}

	t.Run("does not attach optional credentials", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("X-XSRF-TOKEN"); got != "" {
				t.Errorf("unexpected XSRF header: %q", got)
			}
			if got := r.Header.Get("Cookie"); got != "" {
				t.Errorf("unexpected Cookie header: %q", got)
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		defer server.Close()

		client := NewClient()
		client.SetAuth(creds)
		if _, _, err := client.DoAPI(http.MethodGet, server.URL, nil, nil, false); err != nil {
			t.Fatalf("DoAPI: %v", err)
		}
	})

	t.Run("rejects authenticated external URL", func(t *testing.T) {
		external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("external server must not be contacted")
		}))
		defer external.Close()
		base := httptest.NewServer(http.NotFoundHandler())
		defer base.Close()
		oldBaseURL := BaseURL
		BaseURL = base.URL
		defer func() { BaseURL = oldBaseURL }()

		client := NewClient()
		client.SetAuth(creds)
		_, _, err := client.DoAPI(http.MethodGet, external.URL, nil, nil, true)
		if err == nil || !strings.Contains(err.Error(), "refusing to send authentication") {
			t.Fatalf("expected origin rejection, got %v", err)
		}
	})

	t.Run("rejects cross-origin redirect", func(t *testing.T) {
		targetCalled := false
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			targetCalled = true
			if r.Header.Get("X-XSRF-TOKEN") != "" || r.Header.Get("Cookie") != "" {
				t.Error("redirect leaked credentials")
			}
		}))
		defer target.Close()
		source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target.URL, http.StatusFound)
		}))
		defer source.Close()
		oldBaseURL := BaseURL
		BaseURL = source.URL
		defer func() { BaseURL = oldBaseURL }()

		client := NewClient()
		client.SetAuth(creds)
		_, _, err := client.DoAPI(http.MethodGet, "/redirect", nil, nil, true)
		if err == nil || !strings.Contains(err.Error(), "cross-origin redirect") {
			t.Fatalf("expected redirect rejection, got %v", err)
		}
		if targetCalled {
			t.Fatal("redirect target was contacted")
		}
	})
}

func TestReadAPIResponseBodyLimit(t *testing.T) {
	body := strings.NewReader(strings.Repeat("x", MaxAPIResponseBodyBytes+1))
	data, err := readAPIResponseBody(body)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size-limit error, got %v", err)
	}
	if len(data) != MaxAPIResponseBodyBytes {
		t.Fatalf("expected bounded data length %d, got %d", MaxAPIResponseBodyBytes, len(data))
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

func TestSetLoggerNilRestoresSafeDefault(t *testing.T) {
	client := NewClient()
	client.SetLogger(nil)
	if client.logger == nil {
		t.Fatal("SetLogger(nil) left the client with a nil logger")
	}
	client.logger.Debug("nil logger fallback is usable")
}
