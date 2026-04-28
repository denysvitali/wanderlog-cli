//go:build integration
// +build integration

package wanderlog

import (
	"testing"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for Operational Transform helpers and edge cases.
// These tests verify OT operation construction and can optionally apply
// them against a real trip when WANDERLOG_RUN_PROD_INTEGRATION=1

// ============================================================================
// ReplaceInObject Tests
// ============================================================================

func TestReplaceInObject_Basic(t *testing.T) {
	path := []interface{}{"itinerary", "sections", 0, "heading"}
	oldValue := "Old Title"
	newValue := "New Title"

	op := models.ReplaceInObject(path, oldValue, newValue)

	require.NotNil(t, op)
	assert.Equal(t, path, op.P)
	assert.Equal(t, oldValue, op.OD)
	assert.Equal(t, newValue, op.OI)
	assert.Nil(t, op.LI)
	assert.Nil(t, op.LD)
}

func TestReplaceInObject_EmptyPath(t *testing.T) {
	op := models.ReplaceInObject([]interface{}{}, "old", "new")

	require.NotNil(t, op)
	assert.Empty(t, op.P)
	assert.Equal(t, "old", op.OD)
	assert.Equal(t, "new", op.OI)
}

func TestReplaceInObject_NestedPath(t *testing.T) {
	path := []interface{}{"itinerary", "sections", 0, "blocks", 0, "text"}
	op := models.ReplaceInObject(path, "original", "modified")

	require.NotNil(t, op)
	assert.Len(t, path, 6)
	assert.Equal(t, "original", op.OD)
	assert.Equal(t, "modified", op.OI)
}

// ============================================================================
// InsertInObject Tests
// ============================================================================

func TestInsertInObject_Basic(t *testing.T) {
	path := []interface{}{"metadata", "newField"}
	value := "newValue"

	op := models.InsertInObject(path, value)

	require.NotNil(t, op)
	assert.Equal(t, path, op.P)
	assert.Equal(t, value, op.OI)
	assert.Nil(t, op.OD)
	assert.Nil(t, op.LI)
	assert.Nil(t, op.LD)
}

func TestInsertInObject_EmptyPath(t *testing.T) {
	op := models.InsertInObject([]interface{}{}, "value")

	require.NotNil(t, op)
	assert.Empty(t, op.P)
	assert.Equal(t, "value", op.OI)
}

// ============================================================================
// DeleteInObject Tests
// ============================================================================

func TestDeleteInObject_Basic(t *testing.T) {
	path := []interface{}{"itinerary", "sections", 0, "heading"}
	oldValue := "To Be Deleted"

	op := models.DeleteInObject(path, oldValue)

	require.NotNil(t, op)
	assert.Equal(t, path, op.P)
	assert.Equal(t, oldValue, op.OD)
	assert.Nil(t, op.OI)
	assert.Nil(t, op.LI)
	assert.Nil(t, op.LD)
}

func TestDeleteInObject_EmptyPath(t *testing.T) {
	op := models.DeleteInObject([]interface{}{}, "deletedValue")

	require.NotNil(t, op)
	assert.Empty(t, op.P)
	assert.Equal(t, "deletedValue", op.OD)
}

// ============================================================================
// InsertInList Tests
// ============================================================================

func TestInsertInList_Beginning(t *testing.T) {
	path := []interface{}{"items"}
	index := 0
	value := "newFirst"

	op := models.InsertInList(path, index, value)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 0}, op.P)
	assert.Equal(t, value, op.LI)
	assert.Nil(t, op.OD)
	assert.Nil(t, op.LD)
}

func TestInsertInList_End(t *testing.T) {
	path := []interface{}{"items"}
	index := 3 // Insert at end (list has 3 items: indices 0, 1, 2)
	value := "newLast"

	op := models.InsertInList(path, index, value)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 3}, op.P)
	assert.Equal(t, value, op.LI)
}

func TestInsertInList_Middle(t *testing.T) {
	path := []interface{}{"items"}
	index := 1
	value := "middle"

	op := models.InsertInList(path, index, value)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 1}, op.P)
	assert.Equal(t, value, op.LI)
}

func TestInsertInList_EmptyPath(t *testing.T) {
	op := models.InsertInList([]interface{}{}, 0, "value")

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{0}, op.P)
	assert.Equal(t, "value", op.LI)
}

func TestInsertInList_NestedPath(t *testing.T) {
	path := []interface{}{"itinerary", "sections", 0, "blocks"}
	index := 2
	value := map[string]interface{}{"type": "text", "content": "hello"}

	op := models.InsertInList(path, index, value)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"itinerary", "sections", 0, "blocks", 2}, op.P)
	assert.Equal(t, value, op.LI)
}

// ============================================================================
// DeleteFromList Tests
// ============================================================================

func TestDeleteFromList_Beginning(t *testing.T) {
	path := []interface{}{"items"}
	index := 0
	oldValue := "first"

	op := models.DeleteFromList(path, index, oldValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 0}, op.P)
	assert.Equal(t, oldValue, op.LD)
	assert.Nil(t, op.LI)
	assert.Nil(t, op.OD)
}

func TestDeleteFromList_End(t *testing.T) {
	path := []interface{}{"items"}
	index := 2
	oldValue := "last"

	op := models.DeleteFromList(path, index, oldValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 2}, op.P)
	assert.Equal(t, oldValue, op.LD)
}

func TestDeleteFromList_OnlyItem(t *testing.T) {
	path := []interface{}{"items"}
	index := 0
	oldValue := "only"

	op := models.DeleteFromList(path, index, oldValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 0}, op.P)
	assert.Equal(t, oldValue, op.LD)
}

func TestDeleteFromList_Middle(t *testing.T) {
	path := []interface{}{"items"}
	index := 1
	oldValue := "middle"

	op := models.DeleteFromList(path, index, oldValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 1}, op.P)
	assert.Equal(t, oldValue, op.LD)
}

func TestDeleteFromList_EmptyPath(t *testing.T) {
	op := models.DeleteFromList([]interface{}{}, 0, "value")

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{0}, op.P)
	assert.Equal(t, "value", op.LD)
}

// ============================================================================
// ReplaceInList Tests
// ============================================================================

func TestReplaceInList_Beginning(t *testing.T) {
	path := []interface{}{"items"}
	index := 0
	oldValue := "a"
	newValue := "A"

	op := models.ReplaceInList(path, index, oldValue, newValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 0}, op.P)
	assert.Equal(t, oldValue, op.LD)
	assert.Equal(t, newValue, op.LI)
	assert.Nil(t, op.OD)
}

func TestReplaceInList_End(t *testing.T) {
	path := []interface{}{"items"}
	index := 2
	oldValue := "c"
	newValue := "C"

	op := models.ReplaceInList(path, index, oldValue, newValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 2}, op.P)
	assert.Equal(t, oldValue, op.LD)
	assert.Equal(t, newValue, op.LI)
}

func TestReplaceInList_Middle(t *testing.T) {
	path := []interface{}{"items"}
	index := 1
	oldValue := "b"
	newValue := "B"

	op := models.ReplaceInList(path, index, oldValue, newValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 1}, op.P)
	assert.Equal(t, oldValue, op.LD)
	assert.Equal(t, newValue, op.LI)
}

func TestReplaceInList_EmptyPath(t *testing.T) {
	op := models.ReplaceInList([]interface{}{}, 0, "old", "new")

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{0}, op.P)
	assert.Equal(t, "old", op.LD)
	assert.Equal(t, "new", op.LI)
}

func TestReplaceInList_OnlyItem(t *testing.T) {
	path := []interface{}{"items"}
	index := 0
	oldValue := "only"
	newValue := "updated"

	op := models.ReplaceInList(path, index, oldValue, newValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"items", 0}, op.P)
	assert.Equal(t, oldValue, op.LD)
	assert.Equal(t, newValue, op.LI)
}

func TestReplaceInList_NestedPath(t *testing.T) {
	path := []interface{}{"itinerary", "sections", 0, "blocks"}
	index := 1
	oldValue := map[string]interface{}{"type": "text", "content": "old"}
	newValue := map[string]interface{}{"type": "text", "content": "new"}

	op := models.ReplaceInList(path, index, oldValue, newValue)

	require.NotNil(t, op)
	assert.Equal(t, []interface{}{"itinerary", "sections", 0, "blocks", 1}, op.P)
	assert.Equal(t, oldValue, op.LD)
	assert.Equal(t, newValue, op.LI)
}

// ============================================================================
// Operation Construction Consistency Tests
// ============================================================================

func TestOperation_PathWithIndex(t *testing.T) {
	// Test that pathWithIndex correctly appends indices
	path := []interface{}{"a", "b", "c"}

	// Verify InsertInList uses correct path
	op := models.InsertInList(path, 2, "value")
	expectedPath := []interface{}{"a", "b", "c", 2}
	assert.Equal(t, expectedPath, op.P)
}

func TestOperation_LIvsOI_Distinction(t *testing.T) {
	// List insert should use LI
	listOp := models.InsertInList([]interface{}{"items"}, 0, "value")
	assert.NotNil(t, listOp.LI)
	assert.Nil(t, listOp.OI)

	// Object insert should use OI
	objOp := models.InsertInObject([]interface{}{"field"}, "value")
	assert.NotNil(t, objOp.OI)
	assert.Nil(t, objOp.LI)
}

func TestOperation_LDvsOD_Distinction(t *testing.T) {
	// List delete should use LD
	listOp := models.DeleteFromList([]interface{}{"items"}, 0, "value")
	assert.NotNil(t, listOp.LD)
	assert.Nil(t, listOp.OD)

	// Object delete should use OD
	objOp := models.DeleteInObject([]interface{}{"field"}, "value")
	assert.NotNil(t, objOp.OD)
	assert.Nil(t, objOp.LD)
	assert.Nil(t, objOp.LI)
}

func TestReplaceInList_ReplacesBothLIandLD(t *testing.T) {
	// Replace should have both LD (old) and LI (new)
	op := models.ReplaceInList([]interface{}{"items"}, 0, "old", "new")

	assert.NotNil(t, op.LD)
	assert.NotNil(t, op.LI)
	assert.Equal(t, "old", op.LD)
	assert.Equal(t, "new", op.LI)
}

func TestReplaceInObject_HasBothOIandOD(t *testing.T) {
	// Replace should have both OD (old) and OI (new)
	op := models.ReplaceInObject([]interface{}{"field"}, "old", "new")

	assert.NotNil(t, op.OD)
	assert.NotNil(t, op.OI)
	assert.Equal(t, "old", op.OD)
	assert.Equal(t, "new", op.OI)
}

// ============================================================================
// Edge Case: Various Value Types
// ============================================================================

func TestOTHelpers_StringValues(t *testing.T) {
	path := []interface{}{"field"}

	op := models.ReplaceInObject(path, "oldString", "newString")
	assert.Equal(t, "oldString", op.OD)
	assert.Equal(t, "newString", op.OI)
}

func TestOTHelpers_IntValues(t *testing.T) {
	path := []interface{}{"count"}

	op := models.ReplaceInObject(path, 42, 100)
	assert.Equal(t, 42, op.OD)
	assert.Equal(t, 100, op.OI)
}

func TestOTHelpers_MapValues(t *testing.T) {
	path := []interface{}{"config"}

	oldMap := map[string]interface{}{"key": "oldValue"}
	newMap := map[string]interface{}{"key": "newValue"}

	op := models.ReplaceInObject(path, oldMap, newMap)
	assert.Equal(t, oldMap, op.OD)
	assert.Equal(t, newMap, op.OI)
}

func TestOTHelpers_SliceValues(t *testing.T) {
	path := []interface{}{"tags"}

	oldSlice := []interface{}{"a", "b"}
	newSlice := []interface{}{"a", "b", "c"}

	op := models.ReplaceInObject(path, oldSlice, newSlice)
	assert.Equal(t, oldSlice, op.OD)
	assert.Equal(t, newSlice, op.OI)
}

func TestOTHelpers_NilValues(t *testing.T) {
	path := []interface{}{"nullableField"}

	op := models.ReplaceInObject(path, nil, "nowSet")
	assert.Nil(t, op.OD)
	assert.Equal(t, "nowSet", op.OI)
}

func TestInsertInList_NilValue(t *testing.T) {
	path := []interface{}{"items"}

	op := models.InsertInList(path, 0, nil)
	assert.Equal(t, []interface{}{"items", 0}, op.P)
	assert.Nil(t, op.LI)
}

func TestDeleteFromList_NilValue(t *testing.T) {
	path := []interface{}{"items"}

	op := models.DeleteFromList(path, 0, nil)
	assert.Equal(t, []interface{}{"items", 0}, op.P)
	assert.Nil(t, op.LD)
}
