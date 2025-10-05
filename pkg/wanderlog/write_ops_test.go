package wanderlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCreateTrip(t *testing.T) {
	tests := []struct {
		name           string
		req            CreateTripRequest
		serverResponse string
		serverStatus   int
		expectError    bool
		checkResponse  func(*testing.T, *CreateTripResponse)
	}{
		{
			name: "successful create",
			req: CreateTripRequest{
				Title:     "Test Trip",
				StartDate: "2025-01-01",
				EndDate:   "2025-01-10",
				Privacy:   "private",
			},
			serverResponse: `{
				"success": true,
				"tripPlan": {
					"id": 123,
					"key": "test-trip-key",
					"editKey": "edit-key-123",
					"title": "Test Trip"
				}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
			checkResponse: func(t *testing.T, resp *CreateTripResponse) {
				if !resp.Success {
					t.Error("Expected success to be true")
				}
				if resp.TripPlan.ID != 123 {
					t.Errorf("Expected trip ID 123, got %d", resp.TripPlan.ID)
				}
				if resp.TripPlan.Key != "test-trip-key" {
					t.Errorf("Expected key 'test-trip-key', got '%s'", resp.TripPlan.Key)
				}
			},
		},
		{
			name: "server error",
			req: CreateTripRequest{
				Title: "Test Trip",
			},
			serverResponse: `{"error": "internal server error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "api returns success false",
			req: CreateTripRequest{
				Title: "Test Trip",
			},
			serverResponse: `{"success": false}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/tripPlans") {
					t.Errorf("Expected path to end with /tripPlans, got %s", r.URL.Path)
				}

				// Check headers
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("Expected Content-Type: application/json")
				}

				// Check auth headers
				if r.Header.Get("Cookie") == "" {
					t.Error("Expected Cookie header")
				}
				if r.Header.Get("X-XSRF-TOKEN") == "" {
					t.Error("Expected X-XSRF-TOKEN header")
				}

				// Verify request body
				body, _ := io.ReadAll(r.Body)
				var reqData CreateTripRequest
				if err := json.Unmarshal(body, &reqData); err != nil {
					t.Errorf("Failed to parse request body: %v", err)
				}
				if reqData.Title != tt.req.Title {
					t.Errorf("Expected title '%s', got '%s'", tt.req.Title, reqData.Title)
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Override BaseURL for testing
			oldBaseURL := BaseURL
			BaseURL = server.URL
			defer func() { BaseURL = oldBaseURL }()

			// Create client with auth
			client := NewClient()
			client.auth = &AuthCredentials{
				SessionCookie: "test-session",
				XSRFToken:     "test-token",
			}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			client.SetLogger(logger)

			// Call function
			resp, err := client.CreateTrip(tt.req)

			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check response
			if !tt.expectError && tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

func TestCreateTripRequiresAuth(t *testing.T) {
	client := NewClient()
	client.auth = nil // No auth

	_, err := client.CreateTrip(CreateTripRequest{Title: "Test"})
	if err == nil {
		t.Error("Expected error when auth is nil")
	}
	if !strings.Contains(err.Error(), "authentication required") {
		t.Errorf("Expected 'authentication required' error, got: %v", err)
	}
}

func TestDeleteTrip(t *testing.T) {
	tests := []struct {
		name         string
		tripKey      string
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful delete",
			tripKey:      "test-trip-key",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "not found",
			tripKey:      "nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, tt.tripKey) {
					t.Errorf("Expected path to contain '%s', got %s", tt.tripKey, r.URL.Path)
				}
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			oldBaseURL := BaseURL
			BaseURL = server.URL
			defer func() { BaseURL = oldBaseURL }()

			client := NewClient()
			client.auth = &AuthCredentials{
				SessionCookie: "test-session",
				XSRFToken:     "test-token",
			}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			client.SetLogger(logger)

			err := client.DeleteTrip(tt.tripKey)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestAddPlace(t *testing.T) {
	tests := []struct {
		name           string
		tripKey        string
		sectionID      int
		req            AddPlaceRequest
		serverResponse string
		serverStatus   int
		expectError    bool
	}{
		{
			name:      "successful add with section",
			tripKey:   "test-trip",
			sectionID: 1,
			req: AddPlaceRequest{
				Place: struct {
					PlaceID   string  `json:"place_id"`
					Name      string  `json:"name"`
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					PlaceID:   "ChIJ123",
					Name:      "Test Place",
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				Text: "Test Place",
			},
			serverResponse: `{"success": true}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:      "api error response",
			tripKey:   "test-trip",
			sectionID: 1,
			req: AddPlaceRequest{
				Place: struct {
					PlaceID   string  `json:"place_id"`
					Name      string  `json:"name"`
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					PlaceID: "invalid",
					Name:    "Invalid Place",
				},
			},
			serverResponse: `{"success": false, "error": "Invalid place"}`,
			serverStatus:   http.StatusOK,
			expectError:    true,
		},
		{
			name:      "server error",
			tripKey:   "test-trip",
			sectionID: 1,
			req: AddPlaceRequest{
				Text: "Test",
			},
			serverResponse: `{"error": "internal error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Verify path
				if tt.sectionID > 0 {
					if !strings.Contains(r.URL.Path, "/sections/") {
						t.Errorf("Expected path to contain '/sections/', got %s", r.URL.Path)
					}
				}
				if !strings.Contains(r.URL.Path, tt.tripKey) {
					t.Errorf("Expected path to contain '%s', got %s", tt.tripKey, r.URL.Path)
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			oldBaseURL := BaseURL
			BaseURL = server.URL
			defer func() { BaseURL = oldBaseURL }()

			client := NewClient()
			client.auth = &AuthCredentials{
				SessionCookie: "test-session",
				XSRFToken:     "test-token",
			}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			client.SetLogger(logger)

			err := client.AddPlace(tt.tripKey, tt.sectionID, tt.req)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRemovePlace(t *testing.T) {
	tests := []struct {
		name         string
		tripKey      string
		sectionID    int
		placeID      int
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful remove",
			tripKey:      "test-trip",
			sectionID:    1,
			placeID:      100,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "not found",
			tripKey:      "test-trip",
			sectionID:    1,
			placeID:      999,
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			oldBaseURL := BaseURL
			BaseURL = server.URL
			defer func() { BaseURL = oldBaseURL }()

			client := NewClient()
			client.auth = &AuthCredentials{
				SessionCookie: "test-session",
				XSRFToken:     "test-token",
			}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			client.SetLogger(logger)

			err := client.RemovePlace(tt.tripKey, tt.sectionID, tt.placeID)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestApplyOperations(t *testing.T) {
	tests := []struct {
		name           string
		tripKey        string
		ops            []Operation
		serverResponse string
		serverStatus   int
		expectError    bool
	}{
		{
			name:    "successful operations",
			tripKey: "test-trip",
			ops: []Operation{
				{Type: "replace", Path: "/itinerary/sections/0/blocks", Value: []interface{}{}},
			},
			serverResponse: `{"success": true}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			tripKey:        "test-trip",
			ops:            []Operation{{Type: "invalid"}},
			serverResponse: `{"error": "invalid operation"}`,
			serverStatus:   http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "applyOps") {
					t.Errorf("Expected path to contain 'applyOps', got %s", r.URL.Path)
				}

				// Verify request body
				body, _ := io.ReadAll(r.Body)
				var opReq OperationRequest
				if err := json.Unmarshal(body, &opReq); err != nil {
					t.Errorf("Failed to parse operations: %v", err)
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			oldBaseURL := BaseURL
			BaseURL = server.URL
			defer func() { BaseURL = oldBaseURL }()

			client := NewClient()
			client.auth = &AuthCredentials{
				SessionCookie: "test-session",
				XSRFToken:     "test-token",
			}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			client.SetLogger(logger)

			err := client.ApplyOperations(tt.tripKey, tt.ops)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAddPlaceRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         AddPlaceRequest
		expectError bool
	}{
		{
			name: "valid request",
			req: AddPlaceRequest{
				Place: struct {
					PlaceID   string  `json:"place_id"`
					Name      string  `json:"name"`
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					PlaceID:   "ChIJ123",
					Name:      "Test Place",
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
				Text: "Test Place",
			},
			expectError: false,
		},
		{
			name: "invalid latitude (too high)",
			req: AddPlaceRequest{
				Place: struct {
					PlaceID   string  `json:"place_id"`
					Name      string  `json:"name"`
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					PlaceID:   "ChIJ123",
					Name:      "Test",
					Latitude:  91.0, // Invalid
					Longitude: 0,
				},
			},
			expectError: true,
		},
		{
			name: "invalid longitude (too low)",
			req: AddPlaceRequest{
				Place: struct {
					PlaceID   string  `json:"place_id"`
					Name      string  `json:"name"`
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					PlaceID:   "ChIJ123",
					Name:      "Test",
					Latitude:  0,
					Longitude: -181.0, // Invalid
				},
			},
			expectError: true,
		},
		{
			name: "empty place_id",
			req: AddPlaceRequest{
				Place: struct {
					PlaceID   string  `json:"place_id"`
					Name      string  `json:"name"`
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					PlaceID:   "", // Empty
					Name:      "Test",
					Latitude:  0,
					Longitude: 0,
				},
			},
			expectError: true,
		},
		{
			name: "empty name",
			req: AddPlaceRequest{
				Place: struct {
					PlaceID   string  `json:"place_id"`
					Name      string  `json:"name"`
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					PlaceID:   "ChIJ123",
					Name:      "", // Empty
					Latitude:  0,
					Longitude: 0,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddPlaceRequest(tt.req)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}
