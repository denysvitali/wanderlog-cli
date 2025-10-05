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

func TestNukeTripPlaces(t *testing.T) {
	tests := []struct {
		name             string
		tripKey          string
		tripResponse     string
		tripStatus       int
		nukeStatus       int
		expectError      bool
		expectedOpsCount int // Expected number of operations
	}{
		{
			name:    "successful nuke with 2 sections",
			tripKey: "test-trip",
			tripResponse: `{
				"tripPlan": {
					"id": 1,
					"key": "test-trip",
					"itinerary": {
						"sections": [
							{"id": 1, "blocks": []},
							{"id": 2, "blocks": []}
						]
					}
				},
				"resources": {
					"placeMetadata": []
				}
			}`,
			tripStatus:       http.StatusOK,
			nukeStatus:       http.StatusOK,
			expectError:      false,
			expectedOpsCount: 3, // 2 sections + 1 metadata clear
		},
		{
			name:    "trip with no sections",
			tripKey: "empty-trip",
			tripResponse: `{
				"tripPlan": {
					"id": 1,
					"key": "empty-trip",
					"itinerary": {
						"sections": []
					}
				},
				"resources": {
					"placeMetadata": []
				}
			}`,
			tripStatus:  http.StatusOK,
			expectError: false,
			// Should not call applyOps since there are no sections
		},
		{
			name:    "trip fetch fails",
			tripKey: "nonexistent",
			tripResponse: `{"error": "not found"}`,
			tripStatus:  http.StatusNotFound,
			expectError: true,
		},
		{
			name:    "successful nuke with 5 sections",
			tripKey: "big-trip",
			tripResponse: `{
				"tripPlan": {
					"id": 1,
					"key": "big-trip",
					"itinerary": {
						"sections": [
							{"id": 1}, {"id": 2}, {"id": 3}, {"id": 4}, {"id": 5}
						]
					}
				},
				"resources": {
					"placeMetadata": []
				}
			}`,
			tripStatus:       http.StatusOK,
			nukeStatus:       http.StatusOK,
			expectError:      false,
			expectedOpsCount: 6, // 5 sections + 1 metadata
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			var lastOpsCount int

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++

				// First call should be GetTrip
				if callCount == 1 {
					if r.Method != "GET" {
						t.Errorf("First call should be GET, got %s", r.Method)
					}
					w.WriteHeader(tt.tripStatus)
					w.Write([]byte(tt.tripResponse))
					return
				}

				// Second call should be ApplyOperations (if sections exist)
				if callCount == 2 {
					if r.Method != "POST" {
						t.Errorf("Second call should be POST, got %s", r.Method)
					}
					if !strings.Contains(r.URL.Path, "applyOps") {
						t.Errorf("Expected path to contain 'applyOps', got %s", r.URL.Path)
					}

					// Parse and count operations
					body, _ := io.ReadAll(r.Body)
					var opReq OperationRequest
					if err := json.Unmarshal(body, &opReq); err != nil {
						t.Errorf("Failed to parse operations: %v", err)
					}
					lastOpsCount = len(opReq.Ops)

					w.WriteHeader(tt.nukeStatus)
					w.Write([]byte(`{"success": true}`))
					return
				}

				t.Errorf("Unexpected call #%d", callCount)
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

			err := client.NukeTripPlaces(tt.tripKey)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify operations count if expected
			if tt.expectedOpsCount > 0 && lastOpsCount != tt.expectedOpsCount {
				t.Errorf("Expected %d operations, got %d", tt.expectedOpsCount, lastOpsCount)
			}
		})
	}
}

func TestClearSectionBlocks(t *testing.T) {
	tests := []struct {
		name         string
		tripKey      string
		sectionID    int
		serverStatus int
		expectError  bool
	}{
		{
			name:         "successful clear",
			tripKey:      "test-trip",
			sectionID:    1,
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "server error",
			tripKey:      "test-trip",
			sectionID:    1,
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify it's a POST to applyOps
				if r.Method != "POST" {
					t.Errorf("Expected POST, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, "applyOps") {
					t.Errorf("Expected path to contain 'applyOps', got %s", r.URL.Path)
				}

				// Parse operations
				body, _ := io.ReadAll(r.Body)
				var opReq OperationRequest
				if err := json.Unmarshal(body, &opReq); err != nil {
					t.Errorf("Failed to parse operations: %v", err)
				}

				// Should have exactly 1 operation
				if len(opReq.Ops) != 1 {
					t.Errorf("Expected 1 operation, got %d", len(opReq.Ops))
				}

				// Verify operation
				if len(opReq.Ops) > 0 {
					op := opReq.Ops[0]
					if op.Type != "replace" {
						t.Errorf("Expected 'replace' operation, got '%s'", op.Type)
					}
					expectedPath := "/itinerary/sections/1/blocks"
					if op.Path != expectedPath {
						t.Errorf("Expected path '%s', got '%s'", expectedPath, op.Path)
					}
				}

				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(`{"success": true}`))
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

			err := client.ClearSectionBlocks(tt.tripKey, tt.sectionID)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteSection(t *testing.T) {
	tripKey := "test-trip"
	sectionID := 2

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse operations
		body, _ := io.ReadAll(r.Body)
		var opReq OperationRequest
		if err := json.Unmarshal(body, &opReq); err != nil {
			t.Errorf("Failed to parse operations: %v", err)
		}

		// Should have exactly 1 operation
		if len(opReq.Ops) != 1 {
			t.Errorf("Expected 1 operation, got %d", len(opReq.Ops))
		}

		// Verify operation
		if len(opReq.Ops) > 0 {
			op := opReq.Ops[0]
			if op.Type != "remove" {
				t.Errorf("Expected 'remove' operation, got '%s'", op.Type)
			}
			expectedPath := "/itinerary/sections/2"
			if op.Path != expectedPath {
				t.Errorf("Expected path '%s', got '%s'", expectedPath, op.Path)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
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

	err := client.DeleteSection(tripKey, sectionID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
