package wanderlog

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestTripJSONParsing tests that all trip JSON files in trips/ directory
// can be parsed without unknown fields, ensuring our structs are complete
func TestTripJSONParsing(t *testing.T) {
	// Get the trips directory path relative to the project root
	tripsDir := filepath.Join("..", "..", "trips")

	// Read all JSON files in the trips directory
	files, err := filepath.Glob(filepath.Join(tripsDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to read trips directory: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No JSON files found in trips/ directory")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Read the file
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", file, err)
			}

			// Parse with DisallowUnknownFields to catch missing fields
			var trip TripResponse
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.DisallowUnknownFields()

			err = decoder.Decode(&trip)
			if err != nil {
				t.Errorf("Failed to parse %s with strict unmarshaling: %v", filepath.Base(file), err)
				t.Logf("This indicates missing fields in the TripResponse struct")
			} else {
				t.Logf("Successfully parsed %s", filepath.Base(file))
			}
		})
	}
}
