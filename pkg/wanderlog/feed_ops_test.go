package wanderlog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetFeedHome(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/tripPlans/home" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"ownTripPlans":[],"friendsTripPlans":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetFeedHome()
	if err != nil {
		t.Fatalf("GetFeedHome: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGetTripHistoryWithOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/tripPlans/history" {
			t.Errorf("unexpected path: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("offset") != "20" {
			t.Errorf("expected offset=20, got %q", r.URL.Query().Get("offset"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[],"nextOffset":40}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetTripHistory(20)
	if err != nil {
		t.Fatalf("GetTripHistory: %v", err)
	}
	if resp.NextOffset == nil || *resp.NextOffset != 40 {
		t.Errorf("unexpected nextOffset: %v", resp.NextOffset)
	}
}

func TestBrowseGuidesWithGeoID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/browse/guides/12345" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if _, err := client.BrowseGuides(12345); err != nil {
		t.Fatalf("BrowseGuides: %v", err)
	}
}

func TestBrowseGuidesNoGeoID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/browse/guides") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if _, err := client.BrowseGuides(0); err != nil {
		t.Fatalf("BrowseGuides: %v", err)
	}
}

func TestGetFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/feed" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"tripPlans":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetFeed()
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetFeedV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/feed/v2" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"tripPlans":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetFeedV2()
	if err != nil {
		t.Fatalf("GetFeedV2: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetFeedMostRecent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/feed/mostRecentlyEdited" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"tripPlan":null}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetFeedMostRecent()
	if err != nil {
		t.Fatalf("GetFeedMostRecent: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetFriendsPlans(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/friendsPlans" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"tripPlans":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetFriendsPlans()
	if err != nil {
		t.Fatalf("GetFriendsPlans: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetIfEditedFillsSchemaVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/getIfEdited" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"tripPlans":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.GetIfEdited(GetIfEditedRequest{TripPlans: []EditCheck{{Key: "abc"}}})
	if err != nil {
		t.Fatalf("GetIfEdited: %v", err)
	}
}
