package models

import "encoding/json"

// PlacesAPIEnvelope wraps thin /api/placesAPI/* responses where the body shape
// varies (some return Google Places shapes, others Wanderlog wrappers).
type PlacesAPIEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// FindPlaceFromLngLatRequest is the body the client JSON-encodes into the
// `request` query parameter for /api/placesAPI/findPlaceFromLngLat.
type FindPlaceFromLngLatRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// PlacesMetadataResponse mirrors /api/places/metadata.
type PlacesMetadataResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// PlacesCardResponse mirrors /api/places/card.
type PlacesCardResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}
