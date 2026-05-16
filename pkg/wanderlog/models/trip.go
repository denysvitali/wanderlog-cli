package models

// CreateTripRequest represents a request to create a new trip
type CreateTripRequest struct {
	Title               string `json:"title"`
	GeoIDs              []int  `json:"geoIds"`
	InitialMapsPlaceIDs []int  `json:"initialMapsPlaceIds"`
	InitialEmailID      *int   `json:"initialEmailId"`
	Type                string `json:"type"`                // "plan", "recommendations", or "story"
	StartDate           string `json:"startDate,omitempty"` // YYYY-MM-DD format
	EndDate             string `json:"endDate,omitempty"`   // YYYY-MM-DD format
	Privacy             string `json:"privacy,omitempty"`   // "public", "private", "friends"
	IsMapEmbed          bool   `json:"isMapEmbed"`
	Language            string `json:"language,omitempty"` // Language code (e.g., "en", "it")
}

// CreateTripResponse represents the response from creating a trip
type CreateTripResponse struct {
	Success  bool            `json:"success"`
	TripPlan TripPlanSummary `json:"tripPlan"`
}

// TripPlanSummary represents basic trip plan information
type TripPlanSummary struct {
	ID      int    `json:"id"`
	Key     string `json:"key"`
	EditKey string `json:"editKey"`
	Title   string `json:"title"`
}

// CopyTripResponse represents the response from copying a trip
type CopyTripResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Key     string `json:"key"`
		ViewKey string `json:"viewKey"`
		ID      int    `json:"id"`
		Title   string `json:"title"`
	} `json:"data"`
}

// UpdateTripRequest represents a request to update trip metadata
type UpdateTripRequest struct {
	Title     string `json:"title,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Privacy   string `json:"privacy,omitempty"`
}
