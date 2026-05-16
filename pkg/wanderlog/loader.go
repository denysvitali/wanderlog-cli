package wanderlog

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadTripFromFile loads trip data from a local JSON file
func LoadTripFromFile(filePath string) (*TripResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	var tripResponse TripResponse
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tripResponse); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}

	return &tripResponse, nil
}
