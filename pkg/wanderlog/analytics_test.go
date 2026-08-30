package wanderlog

import (
	"encoding/json"
	"testing"
)

func TestAnalyzeTripPreservesMixedBlockCountsAndWarnings(t *testing.T) {
	var trip TripResponse
	if err := json.Unmarshal([]byte(`{
		"success":true,
		"tripPlan":{
			"key":"demo","title":"Demo","startDate":"2026-06-01","endDate":"2026-06-02","days":1,"placeCount":3,
			"itinerary":{"budget":{"expenses":[{"id":1,"amount":{"amount":12.5,"currencyCode":"usd"}},{"id":2,"amount":{"amount":7.5,"currencyCode":"USD"}}]},"sections":[
				{"id":10,"date":"2026-06-01","blocks":[{"id":1,"type":"place","place":{"name":"A"}},{"id":2,"type":"note"},{"id":3,"type":"flight","flightInfo":{"number":12}}]},
				{"id":20,"date":"2026-06-02","blocks":[{"id":4,"type":"hotel","hotel":{"checkIn":"2026-06-01"}},{"id":5,"type":"train","arrive":{"type":"train"}}]}
			]}
		},
		"resources":{"placeMetadata":[]}
	}`), &trip); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	got, err := AnalyzeTrip(&trip)
	if err != nil {
		t.Fatalf("AnalyzeTrip: %v", err)
	}
	if got.PlaceBlocks != 1 || got.Notes != 1 || got.Flights != 1 || got.Lodgings != 1 || got.Transit != 1 {
		t.Fatalf("unexpected block counts: %+v", got)
	}
	if got.Days != 1 || len(got.Warnings) < 3 {
		t.Fatalf("expected duration/place/metadata warnings, got %+v", got.Warnings)
	}
	if len(got.Expenses) != 1 || got.Expenses[0].Currency != "USD" || got.Expenses[0].Amount != 20 || got.Expenses[0].Count != 2 {
		t.Fatalf("unexpected expenses: %+v", got.Expenses)
	}
}

func TestAnalyzeTripRejectsNil(t *testing.T) {
	if _, err := AnalyzeTrip(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestAnalyzeTripClassifiesLodgingWithPlaceMetadataAsLodging(t *testing.T) {
	var trip TripResponse
	if err := json.Unmarshal([]byte(`{
		"tripPlan":{"itinerary":{"sections":[{"id":1,"blocks":[{
			"id":2,"type":"place","place":{"name":"Hotel Example"},
			"hotel":{"checkInDate":"2026-06-01","checkOutDate":"2026-06-02"}
		}]}]}},"resources":{"placeMetadata":[]}
	}`), &trip); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	got, err := AnalyzeTrip(&trip)
	if err != nil {
		t.Fatalf("AnalyzeTrip: %v", err)
	}
	if got.Lodgings != 1 || got.PlaceBlocks != 0 {
		t.Fatalf("expected one lodging and no place block, got %+v", got)
	}
}
