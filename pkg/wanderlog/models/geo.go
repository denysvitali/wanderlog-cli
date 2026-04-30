package models

import "encoding/json"

// GeoEnvelope is a thin success/data wrapper used by most /api/geo endpoints.
type GeoEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}
