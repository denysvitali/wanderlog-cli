package wanderlog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseTravelFlightNumber(t *testing.T) {
	airline, number, err := parseTravelFlightNumber(" mu 244 ")
	if err != nil || airline != "MU" || number != 244 {
		t.Fatalf("parseTravelFlightNumber = %q, %d, %v", airline, number, err)
	}
	if _, _, err := parseTravelFlightNumber("244"); err == nil {
		t.Fatal("expected missing airline code to fail")
	}
}

func TestAddFlightReservationContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewClient().AddFlightReservationContext(ctx, AddFlightReservationRequest{TripKey: "trip", FlightNumber: "MU244", DepartureDate: "2026-06-02"})
	if err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestDeleteTrainReservationUsesExactRawBlock(t *testing.T) {
	var request OperationRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"tripPlan":{"itinerary":{"sections":[{"id":10,"blocks":[{"id":77,"type":"train","carrier":"SBB","serverOnly":{"keep":true}}]}]}}}`))
		case http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode operation request: %v", err)
			}
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	oldBaseURL := BaseURL
	BaseURL = server.URL
	defer func() { BaseURL = oldBaseURL }()

	client := NewClient()
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "token"})
	result, err := client.DeleteTrainReservation(DeleteTravelReservationRequest{TripKey: "trip", BlockID: 77})
	if err != nil {
		t.Fatalf("DeleteTrainReservation: %v", err)
	}
	if !result.Success || result.Kind != "train" || result.BlockID != 77 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(request.Ops) != 1 {
		t.Fatalf("ops = %d, want 1", len(request.Ops))
	}
	deleted, ok := request.Ops[0].LD.(map[string]any)
	if !ok {
		t.Fatalf("deleted block type = %T", request.Ops[0].LD)
	}
	if _, ok := deleted["serverOnly"]; !ok {
		t.Fatalf("exact raw block was not preserved: %#v", deleted)
	}
}

func TestAddLodgingReservationRejectsInvalidRangeBeforeNetwork(t *testing.T) {
	_, err := NewClient().AddLodgingReservation(AddLodgingReservationRequest{
		TripKey: "trip", Name: "Hotel", Latitude: 1, Longitude: 2,
		CheckIn: "2026-06-04", CheckOut: "2026-06-03",
	})
	if err == nil || err.Error() != "check-out date must not be before check-in date" {
		t.Fatalf("error = %v", err)
	}
}
