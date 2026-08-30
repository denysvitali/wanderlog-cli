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
			expectedOpsCount: 0,
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
			name:         "trip fetch fails",
			tripKey:      "nonexistent",
			tripResponse: `{"error": "not found"}`,
			tripStatus:   http.StatusNotFound,
			expectError:  true,
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
			expectedOpsCount: 0,
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

func TestNukeTripPlacesPreservesMixedBlocksAndExactOldValues(t *testing.T) {
	tripResponse := `{
		"tripPlan":{"id":1,"key":"mixed","itinerary":{"sections":[
			{"id":10,"serverSectionField":"keep","blocks":[
				{"id":1,"type":"place","place":{"name":"Cafe"},"serverOnly":{"keep":true}},
				{"id":2,"type":"note","text":"keep this note"},
				{"id":3,"type":"flight","flight":{"number":"WL123"}}
			]},
			{"id":20,"blocks":[{"id":4,"type":"lodging","lodging":{"name":"Hotel"}}]}
		]}},
		"resources":{"placeMetadata":{"1":{"color":"red","serverOnly":true}}}
	}`
	var gotOps []Operation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(tripResponse))
			return
		}
		var req OperationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode operations: %v", err)
		}
		gotOps = req.Ops
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	oldBaseURL := BaseURL
	BaseURL = server.URL
	defer func() { BaseURL = oldBaseURL }()
	client := NewClient()
	client.auth = &AuthCredentials{SessionCookie: "session", XSRFToken: "token"}
	if err := client.NukeTripPlaces("mixed"); err != nil {
		t.Fatalf("NukeTripPlaces: %v", err)
	}
	if len(gotOps) != 2 {
		t.Fatalf("operations = %d, want one block replacement and one metadata replacement: %#v", len(gotOps), gotOps)
	}
	oldBlocks := gotOps[0].OD.([]any)
	newBlocks := gotOps[0].OI.([]any)
	if len(oldBlocks) != 3 || oldBlocks[0].(map[string]any)["serverOnly"] == nil {
		t.Fatalf("old blocks are not the exact raw snapshot: %#v", oldBlocks)
	}
	if len(newBlocks) != 2 || newBlocks[0].(map[string]any)["type"] != "note" || newBlocks[1].(map[string]any)["type"] != "flight" {
		t.Fatalf("non-place blocks were not preserved: %#v", newBlocks)
	}
	oldMetadata, ok := gotOps[1].OD.(map[string]any)
	if !ok || oldMetadata["1"].(map[string]any)["serverOnly"] != true {
		t.Fatalf("metadata old value is not exact: %#v", gotOps[1].OD)
	}
	if newMetadata, ok := gotOps[1].OI.(map[string]any); !ok || len(newMetadata) != 0 {
		t.Fatalf("metadata was not replaced with an empty object: %#v", gotOps[1].OI)
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
				if r.Method == "GET" && r.URL.Path == "/tripPlans/test-trip" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{
						"success": true,
						"tripPlan": {
							"key": "test-trip",
							"itinerary": {
								"sections": [
									{"id": 99, "heading": "Other", "blocks": []},
									{"id": 1, "heading": "Day 1", "blocks": [{"id": 10, "type": "place", "text": {"ops": [{"insert": "\n"}]}, "serverOnly": {"keep": true}}]}
								]
							}
						},
						"resources": {"placeMetadata": []}
					}`))
					return
				}

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

				// Verify operation (ShareDB JSON0 format)
				if len(opReq.Ops) > 0 {
					op := opReq.Ops[0]
					// For replace operations, we should have both OD (old) and OI (new)
					if op.OD == nil || op.OI == nil {
						t.Error("Expected replace operation with both OD and OI fields")
					}
					// Verify path structure
					expectedPath := []interface{}{"itinerary", "sections", 1, "blocks"}
					if len(op.P) != len(expectedPath) {
						t.Errorf("Expected path length %d, got %d", len(expectedPath), len(op.P))
					}
					oldBlocks, ok := op.OD.([]any)
					if !ok || len(oldBlocks) != 1 || oldBlocks[0].(map[string]any)["serverOnly"] == nil {
						t.Errorf("old blocks did not preserve raw server fields: %#v", op.OD)
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
		if r.Method == "GET" && r.URL.Path == "/tripPlans/test-trip" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"tripPlan": {
					"key": "test-trip",
					"itinerary": {
						"sections": [
							{"id": 99, "heading": "Other", "blocks": []},
							{"id": 2, "heading": "Day 2", "blocks": [], "serverOnly": {"keep": true}}
						]
					}
				},
				"resources": {"placeMetadata": []}
			}`))
			return
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

		// Verify operation (ShareDB JSON0 format)
		if len(opReq.Ops) > 0 {
			op := opReq.Ops[0]
			// For remove operations, we should have LD (list delete) field
			if op.LD == nil {
				t.Error("Expected remove operation with LD field")
			}
			// Verify path structure
			expectedPath := []interface{}{"itinerary", "sections", 1}
			if len(op.P) != len(expectedPath) {
				t.Errorf("Expected path length %d, got %d", len(expectedPath), len(op.P))
			}
			oldSection, ok := op.LD.(map[string]any)
			if !ok || oldSection["serverOnly"] == nil {
				t.Errorf("deleted section did not preserve raw server fields: %#v", op.LD)
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
