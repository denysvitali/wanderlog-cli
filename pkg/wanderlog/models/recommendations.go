package models

import "encoding/json"

// RecommendedPlacesRequest is the body for POST /api/recommendations/v2.
type RecommendedPlacesRequest struct {
	TripPlanID        int      `json:"tripPlanId"`
	GeoID             int      `json:"geoId"`
	Input             string   `json:"input,omitempty"`
	ExcludingPlaceIDs []string `json:"excludingPlaceIds,omitempty"`
}

// RecommendedPlacesResponse wraps the v2 recommendation response.
type RecommendedPlacesResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// MarkRecommendationNotInterestedRequest is the body for POST
// /api/recommendations/notInterested.
type MarkRecommendationNotInterestedRequest struct {
	TripPlanID   int    `json:"tripPlanId"`
	MapsPlaceID  string `json:"mapsPlaceId"`
}
