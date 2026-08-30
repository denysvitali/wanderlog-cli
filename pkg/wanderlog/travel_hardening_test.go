package wanderlog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestValidateTravelDateTimeRange(t *testing.T) {
	tests := []struct {
		name                   string
		departureDate, depTime string
		arrivalDate, arrTime   string
		want                   string
	}{
		{name: "malformed date", departureDate: "06/01/2026", arrivalDate: "2026-06-01", want: "YYYY-MM-DD"},
		{name: "malformed time", departureDate: "2026-06-01", depTime: "9am", arrivalDate: "2026-06-01", want: "HH:MM"},
		{name: "reversed dates", departureDate: "2026-06-02", arrivalDate: "2026-06-01", want: "arrival date"},
		{name: "reversed same-day times", departureDate: "2026-06-01", depTime: "11:00", arrivalDate: "2026-06-01", arrTime: "10:59", want: "arrival time"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTravelDateTimeRange(test.departureDate, test.depTime, test.arrivalDate, test.arrTime)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestAddTravelBlockCreatesSectionAndFirstBlockAtomically(t *testing.T) {
	var requests atomic.Int32
	var operation OperationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"tripPlan":{"itinerary":{"sections":[{"id":4,"type":"normal","blocks":[{"id":9,"type":"place"}]}]}},"resources":{"placeMetadata":[]}}`))
		case http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(&operation); err != nil {
				t.Fatalf("decode operation: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	sectionID, blockID, err := newAuthenticatedTestClient(server.URL).addTravelBlock(
		context.Background(), "trip", "flights", map[string]any{"type": "flight"},
	)
	if err != nil {
		t.Fatalf("addTravelBlock: %v", err)
	}
	if sectionID != 10 || blockID != 11 || requests.Load() != 2 || len(operation.Ops) != 1 {
		t.Fatalf("section=%d block=%d requests=%d operation=%+v", sectionID, blockID, requests.Load(), operation)
	}
	if got := operation.Ops[0].P; len(got) != 3 || got[0] != "itinerary" || got[1] != "sections" {
		t.Fatalf("operation path = %#v", got)
	}
	section, ok := operation.Ops[0].LI.(map[string]any)
	if !ok {
		t.Fatalf("inserted section = %T", operation.Ops[0].LI)
	}
	blocks, _ := section["blocks"].([]any)
	if len(blocks) != 1 {
		t.Fatalf("section blocks = %#v", section["blocks"])
	}
}

func TestAddTravelBlockDoesNotReuseIconOnlyNormalSection(t *testing.T) {
	var operation OperationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"tripPlan":{"itinerary":{"sections":[{"id":4,"type":"normal","heading":"Day 1","placeMarkerIcon":"plane","blocks":[]}]}},"resources":{"placeMetadata":[]}}`))
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&operation); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	_, _, err := newAuthenticatedTestClient(server.URL).addTravelBlock(context.Background(), "trip", "flights", map[string]any{"type": "flight"})
	if err != nil {
		t.Fatal(err)
	}
	path := operation.Ops[0].P
	if len(path) != 3 {
		t.Fatalf("expected top-level section insertion, got path %#v", path)
	}
}

func TestUpdateLodgingRejectsProspectiveInvertedRangeWithoutMutation(t *testing.T) {
	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts.Add(1)
		}
		_, _ = w.Write([]byte(`{"tripPlan":{"itinerary":{"sections":[{"id":3,"blocks":[{"id":7,"type":"place","place":{"name":"Hotel"},"hotel":{"checkIn":"2026-06-01","checkOut":"2026-06-04"}}]}]}}}`))
	}))
	defer server.Close()
	newCheckIn := "2026-06-05"
	_, err := newAuthenticatedTestClient(server.URL).UpdateLodgingReservation(UpdateLodgingReservationRequest{
		TripKey: "trip", BlockID: 7, CheckIn: &newCheckIn,
	})
	if err == nil || !strings.Contains(err.Error(), "check-out date") || posts.Load() != 0 {
		t.Fatalf("error = %v, posts = %d", err, posts.Load())
	}
}
