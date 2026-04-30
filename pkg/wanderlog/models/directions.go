package models

import "encoding/json"

// DirectionsResponse is the thin success/data envelope used by /api/directions endpoints.
type DirectionsResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}
