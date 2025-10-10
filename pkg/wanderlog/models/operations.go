package models

// OperationRequest represents a batch operation request
type OperationRequest struct {
	Ops []Operation `json:"ops"`
}

// Operation represents a single ShareDB JSON0 operation
// ShareDB uses a specific format with:
// - p: path as array of keys/indices
// - oi/od: object insert/delete (for replacing object values)
// - li/ld: list insert/delete (for array operations)
type Operation struct {
	P  []interface{} `json:"p"`            // Path as array (e.g., ["itinerary", "sections", 0, "blocks", 1])
	OI interface{}   `json:"oi,omitempty"` // Object insert (new value for replace)
	OD interface{}   `json:"od,omitempty"` // Object delete (old value for replace)
	LI interface{}   `json:"li,omitempty"` // List insert (for array insertions)
	LD interface{}   `json:"ld,omitempty"` // List delete (for array deletions)
}

// OperationHelpers provides functions to create common operations

// ReplaceInObject creates a ShareDB operation to replace an object field
// Path should be an array like: []interface{}{"itinerary", "sections", 0, "heading"}
func ReplaceInObject(path []interface{}, oldValue, newValue interface{}) Operation {
	return Operation{
		P:  path,
		OD: oldValue,
		OI: newValue,
	}
}

// InsertInObject creates a ShareDB operation to insert a new object field
func InsertInObject(path []interface{}, value interface{}) Operation {
	return Operation{
		P:  path,
		OI: value,
	}
}

// DeleteInObject creates a ShareDB operation to delete an object field
func DeleteInObject(path []interface{}, oldValue interface{}) Operation {
	return Operation{
		P:  path,
		OD: oldValue,
	}
}

// InsertInList creates a ShareDB operation to insert an item into an array at a specific index
func InsertInList(path []interface{}, index int, value interface{}) Operation {
	pathWithIndex := append(path, index)
	return Operation{
		P:  pathWithIndex,
		LI: value,
	}
}

// DeleteFromList creates a ShareDB operation to delete an item from an array at a specific index
func DeleteFromList(path []interface{}, index int, oldValue interface{}) Operation {
	pathWithIndex := append(path, index)
	return Operation{
		P:  pathWithIndex,
		LD: oldValue,
	}
}

// ReplaceInList creates a ShareDB operation to replace an item in an array
func ReplaceInList(path []interface{}, index int, oldValue, newValue interface{}) Operation {
	pathWithIndex := append(path, index)
	return Operation{
		P:  pathWithIndex,
		LD: oldValue,
		LI: newValue,
	}
}
