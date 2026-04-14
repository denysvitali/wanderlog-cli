package models

// AddPlaceRequest represents a request to add a place to a trip
type AddPlaceRequest struct {
	Place AddPlaceInfo `json:"place"`
	Text  string       `json:"text"`
}

// AddPlaceInfo represents the place information when adding a place
type AddPlaceInfo struct {
	PlaceID  string         `json:"place_id,omitempty"` // API uses snake_case
	Name     string         `json:"name"`
	Geometry *PlaceGeometry `json:"geometry,omitempty"`
}

// PlaceGeometry represents the geographic location of a place
type PlaceGeometry struct {
	Location PlaceLocation `json:"location"`
}

// PlaceLocation represents latitude and longitude coordinates
type PlaceLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
