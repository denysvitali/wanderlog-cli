package models

import "encoding/json"

// LikesBulkRequest is the body for POST /api/tripPlans/likes.
type LikesBulkRequest struct {
	Keys []string `json:"keys"`
}

// LikesBulkResponse mirrors the "is liked" result for a list of trip keys.
type LikesBulkResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// CreateTripFromFlightsResponse wraps the response from
// POST /api/tripPlans/flights, which seeds a new trip from a flight payload.
type CreateTripFromFlightsResponse struct {
	Success  bool             `json:"success"`
	TripPlan *TripPlanSummary `json:"tripPlan,omitempty"`
	Data     json.RawMessage  `json:"data,omitempty"`
}

// MyProfileResponse mirrors GET /api/tripPlans/myProfile/, the dashboard view
// of the authenticated user's trips.
type MyProfileResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// LodgingCheckoutDataResponse mirrors GET /api/lodging/checkoutData.
type LodgingCheckoutDataResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// DealsResponse mirrors GET /api/deals (user-targeted travel deals).
type DealsResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// RateEmailRequest is the body for POST /api/emails/rate, used to score
// auto-parsed forwarded emails.
type RateEmailRequest struct {
	EmailID int             `json:"emailId,omitempty"`
	Rating  string          `json:"rating,omitempty"`
	Extras  json.RawMessage `json:"extras,omitempty"`
}
