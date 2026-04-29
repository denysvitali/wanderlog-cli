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
	ops, err := movePlaceOperations(testSections(t), 1, 10, 10, 3)
	if err != nil {
		t.Fatalf("movePlaceOperations: %v", err)
	}
	if got, want := ops[0].P[4], 0; got != want {
		t.Fatalf("delete index = %v, want %v", got, want)
	}
	if got, want := ops[1].P[4], 2; got != want {
		t.Fatalf("insert index = %v, want %v", got, want)
	}
}

func TestReorderPlacesOperationsPreservesUnlistedBlocks(t *testing.T) {
	ops, err := reorderPlacesOperations(testSections(t), 10, []int{3, 1})
	if err != nil {
		t.Fatalf("reorderPlacesOperations: %v", err)
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
		id := reflect.ValueOf(raw).FieldByName("ID").Int()
		ids = append(ids, int(id))
	}
	want := []int{3, 2, 1}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %v, want %v", ids, want)
		}
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
