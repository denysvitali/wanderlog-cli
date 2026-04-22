package models

// TripFlightsResponse represents the response from getting trip flights
type TripFlightsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Flights []TripFlight `json:"flights"`
	} `json:"data"`
}

// TripFlight represents a flight in a trip
type TripFlight struct {
	ID            int           `json:"id,omitzero"`
	FlightNumber  string        `json:"flightNumber"`
	Airline       string        `json:"airline"`
	AirlineIATA   string        `json:"airlineIata,omitempty"`
	Origin        FlightAirport `json:"origin"`
	Destination   FlightAirport `json:"destination"`
	DepartureTime string        `json:"departureTime"`
	ArrivalTime   string        `json:"arrivalTime"`
	DurationMins  int           `json:"durationMins,omitempty"`
	Stops         int           `json:"stops,omitempty"`
	BookingURL    string        `json:"bookingUrl,omitempty"`
	Price         float64       `json:"price,omitempty"`
	Currency      string        `json:"currency,omitempty"`
	Source        string        `json:"source,omitempty"`
	ImageURL      string        `json:"imageUrl,omitempty"`
}

// FlightAirport represents an airport in a flight
type FlightAirport struct {
	IATA     string  `json:"iata"`
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Country  string  `json:"country"`
	Timezone string  `json:"timezone,omitempty"`
	Lat      float64 `json:"lat,omitempty"`
	Lng      float64 `json:"lng,omitempty"`
}

// AutofillDayRequest represents a request to autofill a day with suggestions
type AutofillDayRequest struct {
	TripPlanKey string `json:"tripPlanKey"`
	SectionID   int    `json:"sectionId"`
	Query       string `json:"query,omitempty"`
	Location    *struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location,omitempty"`
}

// AutofillDayResponse represents the response from autofilling a day
type AutofillDayResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Suggestions []AutofillSuggestion `json:"suggestions"`
	} `json:"data"`
}

// AutofillSuggestion represents a single autofill suggestion
type AutofillSuggestion struct {
	PlaceID  string   `json:"placeId,omitempty"`
	Name     string   `json:"name"`
	Address  string   `json:"address,omitempty"`
	Lat      float64  `json:"lat,omitempty"`
	Lng      float64  `json:"lng,omitempty"`
	Rating   float64  `json:"rating,omitempty"`
	Types    []string `json:"types,omitempty"`
	ImageURL string   `json:"imageUrl,omitempty"`
	Score    float64  `json:"score,omitempty"`
}

// ChecklistSectionRequest represents a request to manage checklist sections
type ChecklistSectionRequest struct {
	Action    string          `json:"action"` // "addItems", "removeItems", "toggleItem"
	SectionID int             `json:"sectionId"`
	Items     []ChecklistItem `json:"items,omitempty"`
	ItemID    int             `json:"itemId,omitempty"`
}

// ChecklistItem represents a single checklist item
type ChecklistItem struct {
	ID       int    `json:"id,omitempty"`
	Text     string `json:"text"`
	Checked  bool   `json:"checked,omitempty"`
	Category string `json:"category,omitempty"`
}

// ChecklistSectionResponse represents the response from checklist operations
type ChecklistSectionResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Section struct {
			ID    int             `json:"id"`
			Items []ChecklistItem `json:"items"`
		} `json:"section"`
	} `json:"data"`
}

// ExportTripResponse represents the response from exporting a trip
type ExportTripResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url,omitempty"`
	Data    struct {
		ExportURL string `json:"exportUrl,omitempty"`
	} `json:"data"`
}
