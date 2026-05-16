//go:build integration
// +build integration

package wanderlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
)

// debugAPITest runs raw API calls to debug failing endpoints
// Run with: WANDERLOG_RUN_PROD_INTEGRATION=1 go test -v -tags=integration ./pkg/wanderlog -run TestDebugAPI
func TestDebugAPI(t *testing.T) {
	requireProductionIntegrationOptIn(t)

	// Initialize config to load credentials from config file
	if err := InitConfig(); err != nil {
		t.Logf("Warning: Failed to initialize config: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	client := NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first or set credentials in config file", err)
	}

	testTripID := getTestTripID(t)
	t.Logf("Using trip ID: %s", testTripID)

	// Test GetNotifications
	t.Run("GetNotifications", func(t *testing.T) {
		testGetNotificationsRaw(client, logger)
	})

	// Test SetLike (like_trip)
	t.Run("SetLike", func(t *testing.T) {
		testSetLikeRaw(client, testTripID, logger)
	})

	// Test GetLikeCount
	t.Run("GetLikeCount", func(t *testing.T) {
		testGetLikeCountRaw(client, testTripID, logger)
	})

	// Test ListTripInvites
	t.Run("ListTripInvites", func(t *testing.T) {
		testListTripInvitesRaw(client, testTripID, logger)
	})
}

func testGetNotificationsRaw(client *Client, logger *logrus.Logger) {
	logger.Info("=== Testing GetNotifications (raw) ===")

	// Build the URL
	url := BaseURL + "/user/notifications"
	logger.Infof("URL: %s", url)

	// Create request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logger.Errorf("Failed to create request: %v", err)
		return
	}
	printRequest(req, logger)

	// Add auth headers
	if err := client.addAuthHeaders(req); err != nil {
		logger.Errorf("Failed to add auth headers: %v", err)
		return
	}
	printRequest(req, logger)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		logger.Errorf("Request failed: %v", err)
		return
	}
	printResponse(resp, logger)

	// Test with offset
	logger.Info("--- GetNotifications with offset=10 ---")
	req2, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logger.Errorf("Failed to create request: %v", err)
		return
	}
	printRequest(req2, logger)

	if err := client.addAuthHeaders(req2); err != nil {
		logger.Errorf("Failed to add auth headers: %v", err)
		return
	}
	printRequest(req2, logger)

	// Manually add offset to query
	q := req2.URL.Query()
	q.Set("offset", "10")
	req2.URL.RawQuery = q.Encode()
	logger.Infof("URL with query: %s", req2.URL.String())
	printRequest(req2, logger)

	resp2, err := client.httpClient.Do(req2)
	if err != nil {
		logger.Errorf("Request failed: %v", err)
		return
	}
	printResponse(resp2, logger)
}

func testSetLikeRaw(client *Client, tripKey string, logger *logrus.Logger) {
	logger.Info("=== Testing SetLike (raw) ===")

	// Build the URL
	url := fmt.Sprintf("%s/tripPlans/%s/like", BaseURL, tripKey)
	logger.Infof("URL: %s", url)

	// Test liked=true
	logger.Info("--- SetLike with liked=true ---")
	liked := true
	body, _ := json.Marshal(map[string]interface{}{"liked": liked})
	logger.Infof("Request body: %s", string(body))

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Errorf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	printRequest(req, logger)

	if err := client.addAuthHeaders(req); err != nil {
		logger.Errorf("Failed to add auth headers: %v", err)
		return
	}
	printRequest(req, logger)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		logger.Errorf("Request failed: %v", err)
		return
	}
	printResponse(resp, logger)

	// Test liked=false
	logger.Info("--- SetLike with liked=false ---")
	body2, _ := json.Marshal(map[string]interface{}{"liked": false})
	logger.Infof("Request body: %s", string(body2))

	req2, err := http.NewRequest("POST", url, bytes.NewReader(body2))
	if err != nil {
		logger.Errorf("Failed to create request: %v", err)
		return
	}
	req2.Header.Set("Content-Type", "application/json")
	printRequest(req2, logger)

	if err := client.addAuthHeaders(req2); err != nil {
		logger.Errorf("Failed to add auth headers: %v", err)
		return
	}
	printRequest(req2, logger)

	resp2, err := client.httpClient.Do(req2)
	if err != nil {
		logger.Errorf("Request failed: %v", err)
		return
	}
	printResponse(resp2, logger)
}

func testGetLikeCountRaw(client *Client, tripKey string, logger *logrus.Logger) {
	logger.Info("=== Testing GetLikeCount (raw) ===")

	// GetLikesBulk expects POST with JSON body: {"keys": [tripKey]}
	url := BaseURL + "/tripPlans/likes/bulk"
	logger.Infof("URL: %s", url)

	body, _ := json.Marshal(map[string][]string{"keys": {tripKey}})
	logger.Infof("Request body: %s", string(body))

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Errorf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	printRequest(req, logger)

	if err := client.addAuthHeaders(req); err != nil {
		logger.Errorf("Failed to add auth headers: %v", err)
		return
	}
	printRequest(req, logger)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		logger.Errorf("Request failed: %v", err)
		return
	}
	printResponse(resp, logger)
}

func testListTripInvitesRaw(client *Client, tripKey string, logger *logrus.Logger) {
	logger.Info("=== Testing ListTripInvites (raw) ===")

	url := fmt.Sprintf("%s/tripPlans/%s/invites", BaseURL, tripKey)
	logger.Infof("URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Errorf("Failed to create request: %v", err)
		return
	}
	printRequest(req, logger)

	if err := client.addAuthHeaders(req); err != nil {
		logger.Errorf("Failed to add auth headers: %v", err)
		return
	}
	printRequest(req, logger)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		logger.Errorf("Request failed: %v", err)
		return
	}
	printResponse(resp, logger)
}

// Helper to print the raw HTTP request
func printRequest(req *http.Request, logger *logrus.Logger) {
	fmt.Printf("\n=== REQUEST ===\n")
	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("URL: %s\n", req.URL.String())
	fmt.Printf("Headers:\n")
	for k, v := range req.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(body))
		if len(body) > 0 {
			fmt.Printf("Body: %s\n", string(body))
		}
	}
}

// Helper to print the raw HTTP response
func printResponse(resp *http.Response, logger *logrus.Logger) {
	fmt.Printf("\n=== RESPONSE ===\n")
	fmt.Printf("Status: %s (%d)\n", resp.Status, resp.StatusCode)
	fmt.Printf("Headers:\n")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Body (error reading): %v\n", err)
	} else {
		resp.Body = io.NopCloser(bytes.NewReader(body))
		if len(body) > 0 {
			// Try to pretty print JSON
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
				fmt.Printf("Body:\n%s\n", prettyJSON.String())
			} else {
				fmt.Printf("Body: %s\n", string(body))
			}
		}
	}
}