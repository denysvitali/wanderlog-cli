package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTripRawPreservesLargeIntegerTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tripPlan":{"itinerary":{"sections":[{"id":1,"blocks":[{"id":2,"type":"place","serverSequence":9007199254740993}]}]}},"resources":{}}`))
	}))
	defer server.Close()
	withTestBaseURL(t, server)

	trip, err := NewClient().GetTripRaw("trip")
	if err != nil {
		t.Fatal(err)
	}
	sections, err := rawItinerarySections(trip)
	if err != nil {
		t.Fatal(err)
	}
	block := sections[0].(map[string]any)["blocks"].([]any)[0].(map[string]any)
	want := "9007199254740993"
	value, ok := block["serverSequence"].(json.Number)
	if !ok || value.String() != want {
		t.Fatalf("raw number = %#v (%T), want json.Number(%s)", block["serverSequence"], block["serverSequence"], want)
	}

	cloned, err := cloneRawMap(block)
	if err != nil {
		t.Fatal(err)
	}
	clonedValue, ok := cloned["serverSequence"].(json.Number)
	if !ok || clonedValue.String() != want {
		t.Fatalf("cloned number = %#v (%T), want json.Number(%s)", cloned["serverSequence"], cloned["serverSequence"], want)
	}
}
