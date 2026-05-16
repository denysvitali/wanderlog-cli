package models

import "encoding/json"

// JournalResponse mirrors /api/tripPlans/viewOnlyJournal/{journalKey}. The
// journal document itself is large and dynamic so we expose it as RawMessage.
type JournalResponse struct {
	Success bool            `json:"success"`
	Journal json.RawMessage `json:"journal,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// JournalStopPolylinesRequest is the POST payload for /api/tripPlans/journalStopPolylines.
type JournalStopPolylinesRequest struct {
	Stops             []JournalStop     `json:"stops"`
	ExistingPolylines []JournalPolyline `json:"existingPolylines"`
}

type JournalStop struct {
	ID       string  `json:"id,omitempty"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	PlaceID  string  `json:"placeId,omitempty"`
	StopType string  `json:"stopType,omitempty"`
}

type JournalPolyline struct {
	FromStopID string `json:"fromStopId"`
	ToStopID   string `json:"toStopId"`
	Polyline   string `json:"polyline,omitempty"`
}

type JournalPolylinesResponse struct {
	Success   bool              `json:"success"`
	Polylines []JournalPolyline `json:"polylines,omitempty"`
}

// UpdateRequiredResponse is the response from /api/tripPlans/{key}/updateRequired.
type UpdateRequiredResponse struct {
	Success        bool `json:"success"`
	UpdateRequired bool `json:"updateRequired"`
	MinVersion     int  `json:"minVersion,omitempty"`
}

// DistinctionResponse is the response from GET /api/tripPlans/{key}/distinction.
type DistinctionResponse struct {
	Success     bool   `json:"success"`
	Distinction string `json:"distinction,omitempty"`
}

// CreateGuideResponse is the response from POST /api/tripPlans/{key}/createGuideFromTripPlan.
type CreateGuideResponse struct {
	Success bool            `json:"success"`
	Guide   json.RawMessage `json:"guide,omitempty"`
}
