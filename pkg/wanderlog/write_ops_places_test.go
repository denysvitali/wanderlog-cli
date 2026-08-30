package wanderlog

import (
	"encoding/json"
	"reflect"
	"testing"
)

func testSections(t *testing.T) []ItSections {
	t.Helper()
	var sections []ItSections
	data := []byte(`[
		{"id":10,"type":"normal","blocks":[
			{"id":1,"type":"place","text":{"ops":[{"insert":"one\n"}]},"attachments":[],"upvotedBy":[],"travelMode":null},
			{"id":2,"type":"note","text":{"ops":[{"insert":"note\n"}]},"attachments":[],"upvotedBy":[],"travelMode":null},
			{"id":3,"type":"place","text":{"ops":[{"insert":"three\n"}]},"attachments":[],"upvotedBy":[],"travelMode":null}
		]},
		{"id":20,"type":"normal","blocks":[
			{"id":4,"type":"place","text":{"ops":[{"insert":"four\n"}]},"attachments":[],"upvotedBy":[],"travelMode":null}
		]}
	]`)
	if err := json.Unmarshal(data, &sections); err != nil {
		t.Fatalf("unmarshal sections: %v", err)
	}
	return sections
}

func TestMovePlaceOperationsAdjustsSameSectionForwardMove(t *testing.T) {
	var trip map[string]any
	if err := json.Unmarshal([]byte(`{"tripPlan":{"itinerary":{"sections":[{"id":10,"blocks":[{"id":1,"type":"place","place":{"name":"one"},"serverOnly":true},{"id":2,"type":"note"},{"id":3,"type":"place","place":{"name":"three"}}]}]}}}`), &trip); err != nil {
		t.Fatal(err)
	}
	ops, err := movePlaceRawOperations(trip, 1, 10, 10, 1)
	if err != nil {
		t.Fatalf("movePlaceRawOperations: %v", err)
	}
	if got, want := ops[0].P[4], 0; got != want {
		t.Fatalf("delete index = %v, want %v", got, want)
	}
	if got, want := ops[1].P[4], 1; got != want {
		t.Fatalf("insert index = %v, want %v", got, want)
	}
	deleted := ops[0].LD.(map[string]any)
	if deleted["serverOnly"] != true {
		t.Fatalf("exact raw block was not preserved: %#v", deleted)
	}
}

func TestReorderPlacesOperationsPreservesUnlistedBlocks(t *testing.T) {
	var trip map[string]any
	if err := json.Unmarshal([]byte(`{"tripPlan":{"itinerary":{"sections":[{"id":10,"blocks":[{"id":1,"type":"place","place":{"name":"one"},"unknown":{"keep":true}},{"id":2,"type":"note","unknown":"note"},{"id":3,"type":"place","place":{"name":"three"}}]}]}}}`), &trip); err != nil {
		t.Fatal(err)
	}
	ops, err := reorderPlacesRawOperations(trip, 10, []int{3, 1})
	if err != nil {
		t.Fatalf("reorderPlacesRawOperations: %v", err)
	}
	newBlocks, ok := ops[0].OI.([]any)
	if !ok {
		t.Fatalf("new blocks type = %T, want []any", ops[0].OI)
	}
	if len(newBlocks) != 3 {
		t.Fatalf("new blocks length = %d, want 3", len(newBlocks))
	}
	ids := make([]int, 0, len(newBlocks))
	for _, raw := range newBlocks {
		ids = append(ids, rawInt(raw.(map[string]any)["id"]))
	}
	want := []int{3, 2, 1}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %v, want %v", ids, want)
		}
	}
	if got := newBlocks[1].(map[string]any)["unknown"]; got != "note" {
		t.Fatalf("unlisted note block was not preserved: %#v", newBlocks[1])
	}
	if got := newBlocks[2].(map[string]any)["unknown"]; !reflect.DeepEqual(got, map[string]any{"keep": true}) {
		t.Fatalf("unknown place fields were not preserved: %#v", newBlocks[2])
	}
}

func TestUpdatePlaceVisitTimeOperationsReplacesBlock(t *testing.T) {
	ops, err := updatePlaceVisitTimeOperations(testSections(t), 10, 1, "09:30", "11:00")
	if err != nil {
		t.Fatalf("updatePlaceVisitTimeOperations: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops length = %d, want 1", len(ops))
	}
	if got, want := ops[0].P, []interface{}{"itinerary", "sections", 0, "blocks", 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("path = %v, want %v", got, want)
	}
	newBlock := reflect.ValueOf(ops[0].LI)
	if got, want := newBlock.FieldByName("StartTime").String(), "09:30"; got != want {
		t.Fatalf("start time = %q, want %q", got, want)
	}
	if got, want := newBlock.FieldByName("EndTime").String(), "11:00"; got != want {
		t.Fatalf("end time = %q, want %q", got, want)
	}
}

func TestUpdatePlaceVisitTimeRawOperationsPreservesBlock(t *testing.T) {
	var trip map[string]any
	data := []byte(`{
		"tripPlan": {
			"itinerary": {
				"sections": [{
					"id": 10,
					"blocks": [{
						"id": 1,
						"type": "place",
						"place": {"name": "Cafe"},
						"startTime": "08:00",
						"endTime": "10:00",
						"serverOnlyField": {"keep": true}
					}]
				}]
			}
		}
	}`)
	if err := json.Unmarshal(data, &trip); err != nil {
		t.Fatalf("unmarshal trip: %v", err)
	}

	ops, err := updatePlaceVisitTimeRawOperations(trip, 10, 1, "09:30", "")
	if err != nil {
		t.Fatalf("updatePlaceVisitTimeRawOperations: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops length = %d, want 1", len(ops))
	}
	if got, want := ops[0].P, []interface{}{"itinerary", "sections", 0, "blocks", 0}; !reflect.DeepEqual(got, want) {
		t.Fatalf("path = %v, want %v", got, want)
	}
	newBlock, ok := ops[0].LI.(map[string]any)
	if !ok {
		t.Fatalf("new block type = %T, want map[string]any", ops[0].LI)
	}
	if got, want := newBlock["startTime"], "09:30"; got != want {
		t.Fatalf("startTime = %v, want %q", got, want)
	}
	if got, want := newBlock["endTime"], "10:00"; got != want {
		t.Fatalf("endTime = %v, want preserved %q", got, want)
	}
	if _, ok := newBlock["serverOnlyField"]; !ok {
		t.Fatalf("serverOnlyField was not preserved")
	}
}
