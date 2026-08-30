package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func authenticatedTestClient() *Client {
	client := NewClient()
	client.auth = &AuthCredentials{SessionCookie: "session", XSRFToken: "token"}
	return client
}

func withTestBaseURL(t *testing.T, server *httptest.Server) {
	t.Helper()
	oldBaseURL := BaseURL
	BaseURL = server.URL
	t.Cleanup(func() { BaseURL = oldBaseURL })
}

func TestCreateExampleTripUsesViewKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":7,"viewKey":"view-only-key","title":"Example"}}`))
	}))
	defer server.Close()
	withTestBaseURL(t, server)

	response, err := authenticatedTestClient().CreateExampleTrip()
	if err != nil {
		t.Fatal(err)
	}
	if response.TripPlan.Key != "view-only-key" {
		t.Fatalf("trip key = %q, want viewKey fallback", response.TripPlan.Key)
	}
}

func TestCopyTripUsesViewKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":8,"viewKey":"copied-view-key","title":"Copy"}}`))
	}))
	defer server.Close()
	withTestBaseURL(t, server)

	response, err := authenticatedTestClient().CopyTrip("source")
	if err != nil {
		t.Fatal(err)
	}
	if response.TripPlan.Key != "copied-view-key" {
		t.Fatalf("trip key = %q, want viewKey fallback", response.TripPlan.Key)
	}
}

func TestUpdateTripRejectsInvalidProspectiveDatesBeforeMutation(t *testing.T) {
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			mutations++
		}
		_, _ = w.Write([]byte(`{"success":true,"tripPlan":{"key":"trip","startDate":"2026-09-01","endDate":"2026-09-10","days":10,"itinerary":{"sections":[]}},"resources":{}}`))
	}))
	defer server.Close()
	withTestBaseURL(t, server)

	err := authenticatedTestClient().UpdateTrip("trip", UpdateTripRequest{StartDate: "2026-09-20"})
	if err == nil || !strings.Contains(err.Error(), "end date must be on or after start date") {
		t.Fatalf("error = %v, want reversed-date validation", err)
	}
	if mutations != 0 {
		t.Fatalf("sent %d mutation requests after validation failure", mutations)
	}
}

func TestCreateTripRejectsInvalidDatesBeforeRequest(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"success":true,"tripPlan":{"key":"unexpected"}}`))
	}))
	defer server.Close()
	withTestBaseURL(t, server)

	_, err := authenticatedTestClient().CreateTrip(CreateTripRequest{
		Title: "Invalid", GeoIDs: []int{1}, StartDate: "2026-09-10", EndDate: "2026-09-01",
	})
	if err == nil || !strings.Contains(err.Error(), "end date must be on or after start date") {
		t.Fatalf("error = %v, want reversed-date validation", err)
	}
	if requests != 0 {
		t.Fatalf("sent %d requests after validation failure", requests)
	}
}

func TestAddChecklistItemsSendsSectionID(t *testing.T) {
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"success":true,"tripPlan":{"id":42,"key":"trip","itinerary":{"sections":[]}},"resources":{}}`))
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"section":{"id":9,"items":[]}}}`))
	}))
	defer server.Close()
	withTestBaseURL(t, server)

	if _, err := authenticatedTestClient().AddChecklistItems("trip", 9, []ChecklistItem{{Text: "Passport"}}); err != nil {
		t.Fatal(err)
	}
	if got := rawInt(requestBody["sectionId"]); got != 9 {
		t.Fatalf("sectionId = %d, want 9; body = %#v", got, requestBody)
	}
	if got := rawInt(requestBody["tripPlanId"]); got != 42 {
		t.Fatalf("tripPlanId = %d, want 42; body = %#v", got, requestBody)
	}
}

func TestToggleChecklistItemRejectsSuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error":"item not found"}`))
	}))
	defer server.Close()
	withTestBaseURL(t, server)

	_, err := authenticatedTestClient().ToggleChecklistItem("trip", 9, 11, true)
	if err == nil || !strings.Contains(err.Error(), "item not found") {
		t.Fatalf("error = %v, want API failure", err)
	}
}
