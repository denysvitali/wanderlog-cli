package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateTripRefetchesAndRebuildsAfterConflict(t *testing.T) {
	getCalls := 0
	applyCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCalls++
			title := "first snapshot"
			if getCalls > 1 {
				title = "concurrent title"
			}
			_, _ = io.WriteString(w, `{"success":true,"tripPlan":{"key":"trip-a","title":"`+title+`","itinerary":{"sections":[]}},"resources":{"placeMetadata":[]}}`)
		case http.MethodPost:
			applyCalls++
			var request OperationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode operations: %v", err)
			}
			if len(request.Ops) != 1 {
				t.Fatalf("operations = %#v", request.Ops)
			}
			wantOld := "first snapshot"
			if applyCalls > 1 {
				wantOld = "concurrent title"
			}
			if request.Ops[0].OD != wantOld || request.Ops[0].OI != "desired title" {
				t.Fatalf("attempt %d operation = %#v, want old %q", applyCalls, request.Ops[0], wantOld)
			}
			if applyCalls == 1 {
				w.WriteHeader(http.StatusConflict)
				_, _ = io.WriteString(w, `{"success":false,"error":"operation conflict"}`)
				return
			}
			_, _ = io.WriteString(w, `{"success":true}`)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	if err := client.UpdateTrip("trip-a", UpdateTripRequest{Title: "desired title"}); err != nil {
		t.Fatalf("UpdateTrip: %v", err)
	}
	if getCalls != 2 || applyCalls != 2 {
		t.Fatalf("calls = GET %d, apply %d; want 2 each", getCalls, applyCalls)
	}
}

func TestRetryJSON0MutationAcceptsNilContext(t *testing.T) {
	called := false
	err := NewClient().retryJSON0MutationContext(nil, "trip-a", "test", func(ctx context.Context) ([]Operation, error) { //nolint:staticcheck // regression verifies defensive nil handling
		called = true
		if ctx == nil {
			t.Fatal("builder received nil context")
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("retryJSON0MutationContext: %v", err)
	}
	if !called {
		t.Fatal("builder was not called")
	}
}

func TestUpdateTripConflictRetryIsBounded(t *testing.T) {
	getCalls := 0
	applyCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getCalls++
			_, _ = io.WriteString(w, `{"success":true,"tripPlan":{"key":"trip-a","title":"old","itinerary":{"sections":[]}},"resources":{"placeMetadata":[]}}`)
			return
		}
		applyCalls++
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"success":false,"error":"operation conflict"}`)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	if err := client.UpdateTrip("trip-a", UpdateTripRequest{Title: "desired"}); err == nil {
		t.Fatal("expected final conflict")
	}
	if getCalls != maxJSON0MutationAttempts || applyCalls != maxJSON0MutationAttempts {
		t.Fatalf("calls = GET %d, apply %d; want %d each", getCalls, applyCalls, maxJSON0MutationAttempts)
	}
}

func TestUpdateTripDoesNotRetryNonConflict(t *testing.T) {
	getCalls := 0
	applyCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getCalls++
			_, _ = io.WriteString(w, `{"success":true,"tripPlan":{"key":"trip-a","title":"old","itinerary":{"sections":[]}},"resources":{"placeMetadata":[]}}`)
			return
		}
		applyCalls++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"success":false,"error":"server error"}`)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	if err := client.UpdateTrip("trip-a", UpdateTripRequest{Title: "desired"}); err == nil {
		t.Fatal("expected server error")
	}
	if getCalls != 1 || applyCalls != 1 {
		t.Fatalf("calls = GET %d, apply %d; want 1 each", getCalls, applyCalls)
	}
}

func TestClearSectionBlocksRefetchesAndRebuildsAfterConflict(t *testing.T) {
	getCalls := 0
	applyCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCalls++
			blockID := 11
			if getCalls > 1 {
				blockID = 22
			}
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"success":true,"tripPlan":{"itinerary":{"sections":[{"id":7,"blocks":[{"id":%d,"type":"place"}]}]}}}`,
				blockID,
			))
		case http.MethodPost:
			applyCalls++
			var request OperationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode operations: %v", err)
			}
			if len(request.Ops) != 1 {
				t.Fatalf("operations = %#v", request.Ops)
			}
			oldBlocks, ok := request.Ops[0].OD.([]any)
			if !ok || len(oldBlocks) != 1 {
				t.Fatalf("attempt %d old blocks = %#v", applyCalls, request.Ops[0].OD)
			}
			oldBlock, ok := oldBlocks[0].(map[string]any)
			if !ok {
				t.Fatalf("attempt %d old block = %#v", applyCalls, oldBlocks[0])
			}
			wantID := 11
			if applyCalls > 1 {
				wantID = 22
			}
			if gotID := rawInt(oldBlock["id"]); gotID != wantID {
				t.Fatalf("attempt %d old block ID = %d, want %d", applyCalls, gotID, wantID)
			}
			if applyCalls == 1 {
				w.WriteHeader(http.StatusConflict)
				_, _ = io.WriteString(w, `{"success":false,"error":"operation conflict"}`)
				return
			}
			_, _ = io.WriteString(w, `{"success":true}`)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	if err := client.ClearSectionBlocks("trip-a", 7); err != nil {
		t.Fatalf("ClearSectionBlocks: %v", err)
	}
	if getCalls != 2 || applyCalls != 2 {
		t.Fatalf("calls = GET %d, apply %d; want 2 each", getCalls, applyCalls)
	}
}

func TestSetTripBudgetRefetchesAndRebuildsAfterConflict(t *testing.T) {
	getCalls := 0
	applyCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCalls++
			oldAmount := 100
			if getCalls > 1 {
				oldAmount = 200
			}
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"success":true,"tripPlan":{"itinerary":{"sections":[],"budget":{"amount":{"amount":%d,"currencyCode":"USD"},"expenses":[]}}}}`,
				oldAmount,
			))
		case http.MethodPost:
			applyCalls++
			var request OperationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode operations: %v", err)
			}
			if len(request.Ops) != 1 {
				t.Fatalf("operations = %#v", request.Ops)
			}
			oldAmount, ok := request.Ops[0].OD.(map[string]any)
			if !ok {
				t.Fatalf("attempt %d old amount = %#v", applyCalls, request.Ops[0].OD)
			}
			wantOld := 100
			if applyCalls > 1 {
				wantOld = 200
			}
			if got := rawInt(oldAmount["amount"]); got != wantOld {
				t.Fatalf("attempt %d old amount = %d, want %d", applyCalls, got, wantOld)
			}
			newAmount, ok := request.Ops[0].OI.(map[string]any)
			if !ok || rawInt(newAmount["amount"]) != 300 || newAmount["currencyCode"] != "EUR" {
				t.Fatalf("attempt %d new amount = %#v", applyCalls, request.Ops[0].OI)
			}
			if applyCalls == 1 {
				w.WriteHeader(http.StatusConflict)
				_, _ = io.WriteString(w, `{"success":false,"error":"operation conflict"}`)
				return
			}
			_, _ = io.WriteString(w, `{"success":true}`)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	if err := client.SetTripBudget("trip-a", 300, "eur"); err != nil {
		t.Fatalf("SetTripBudget: %v", err)
	}
	if getCalls != 2 || applyCalls != 2 {
		t.Fatalf("calls = GET %d, apply %d; want 2 each", getCalls, applyCalls)
	}
}

func TestAppendTravelBlockReturnsIDFromSuccessfulRebuild(t *testing.T) {
	getCalls := 0
	applyCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCalls++
			blocks := `[{"id":5,"type":"text"}]`
			if getCalls > 1 {
				blocks = `[{"id":5,"type":"text"},{"id":20,"type":"text"}]`
			}
			_, _ = io.WriteString(w, `{"success":true,"tripPlan":{"itinerary":{"sections":[{"id":7,"type":"flights","blocks":`+blocks+`}]}}}`)
		case http.MethodPost:
			applyCalls++
			var request OperationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode operations: %v", err)
			}
			if len(request.Ops) != 1 {
				t.Fatalf("operations = %#v", request.Ops)
			}
			inserted, ok := request.Ops[0].LI.(map[string]any)
			if !ok {
				t.Fatalf("attempt %d inserted block = %#v", applyCalls, request.Ops[0].LI)
			}
			wantID, wantPosition := 8, 1
			if applyCalls > 1 {
				wantID, wantPosition = 21, 2
			}
			if got := rawInt(inserted["id"]); got != wantID {
				t.Fatalf("attempt %d inserted ID = %d, want %d", applyCalls, got, wantID)
			}
			if got := rawInt(request.Ops[0].P[len(request.Ops[0].P)-1]); got != wantPosition {
				t.Fatalf("attempt %d position = %d, want %d", applyCalls, got, wantPosition)
			}
			if applyCalls == 1 {
				w.WriteHeader(http.StatusConflict)
				_, _ = io.WriteString(w, `{"success":false,"error":"operation conflict"}`)
				return
			}
			_, _ = io.WriteString(w, `{"success":true}`)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	block := map[string]any{"type": "flight"}
	blockID, err := client.appendTravelBlock(context.Background(), "trip-a", 7, block)
	if err != nil {
		t.Fatalf("appendTravelBlock: %v", err)
	}
	if blockID != 21 {
		t.Fatalf("returned block ID = %d, want 21", blockID)
	}
	if _, mutated := block["id"]; mutated {
		t.Fatalf("appendTravelBlock mutated caller block: %#v", block)
	}
	if getCalls != 2 || applyCalls != 2 {
		t.Fatalf("calls = GET %d, apply %d; want 2 each", getCalls, applyCalls)
	}
}
