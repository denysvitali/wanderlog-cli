package models

// CreateTripRequest represents a request to create a new trip
type CreateTripRequest struct {
	Title     string `json:"title"`
	StartDate string `json:"startDate,omitempty"` // YYYY-MM-DD format
	EndDate   string `json:"endDate,omitempty"`   // YYYY-MM-DD format
	Privacy   string `json:"privacy,omitempty"`   // "public", "private", "unlisted"
}

// CreateTripResponse represents the response from creating a trip
type CreateTripResponse struct {
	Success  bool           `json:"success"`
	TripPlan TripPlanSummary `json:"tripPlan"`
}

// TripPlanSummary represents basic trip plan information
type TripPlanSummary struct {
	ID      int    `json:"id"`
	Key     string `json:"key"`
	EditKey string `json:"editKey"`
	Title   string `json:"title"`
}

// UpdateTripRequest represents a request to update trip metadata
type UpdateTripRequest struct {
	Title     string `json:"title,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Privacy   string `json:"privacy,omitempty"`
}
