//go:build integration
// +build integration

package wanderlog

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// Integration tests for write operations against the real Wanderlog API
// Run with: go test -v -tags=integration ./pkg/wanderlog

const testTripID = "vetyiadvqjgikbvx"

func setupIntegrationClient(t *testing.T) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	client := NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v. Please run 'wanderlog auth login' first", err)
	}

	return client
}

func getTestTripID() string {
	if tripID := os.Getenv("WANDERLOG_TEST_TRIP_ID"); tripID != "" {
		return tripID
	}
	return testTripID
}

func TestIntegration_CreateAndDeleteTrip(t *testing.T) {
	client := setupIntegrationClient(t)

	createReq := CreateTripRequest{
		Title:     "Integration Test Trip",
		StartDate: "2025-11-01",
		EndDate:   "2025-11-07",
		Privacy:   "private",
	}

	createResp, err := client.CreateTrip(createReq)
	if err != nil || !createResp.Success {
		t.Fatalf("Failed to create trip: %v", err)
	}

	tripKey := createResp.TripPlan.Key
	t.Logf("Created trip: %s", tripKey)

	if err := client.DeleteTrip(tripKey); err != nil {
		t.Fatalf("Failed to delete trip: %v", err)
	}
}

func TestIntegration_CopyTrip(t *testing.T) {
	client := setupIntegrationClient(t)
	sourceTripID := getTestTripID()

	copyResp, err := client.CopyTrip(sourceTripID)
	if err != nil {
		t.Fatalf("Failed to copy trip: %v", err)
	}

	t.Logf("Copied to: %s", copyResp.TripPlan.Key)
	_ = client.DeleteTrip(copyResp.TripPlan.Key)
}

func TestIntegration_GetLikeCount(t *testing.T) {
	client := setupIntegrationClient(t)
	tripID := getTestTripID()

	likeCount, err := client.GetLikeCount(tripID)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	t.Logf("Likes: %d, User liked: %v", likeCount.Count, likeCount.UserLiked)
}
