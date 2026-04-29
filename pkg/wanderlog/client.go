package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sirupsen/logrus"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
)

const (
	ClientVersion    = "2"
	DefaultUserAgent = "wanderlog-cli/1.0"
)

var (
	BaseURL = "https://wanderlog.com/api"
)

type Client struct {
	httpClient *http.Client
	logger     *logrus.Logger
	userAgent  string
	auth       *AuthCredentials
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:    logrus.New(),
		userAgent: DefaultUserAgent,
	}
}

func (c *Client) SetLogger(logger *logrus.Logger) {
	c.logger = logger
}

func (c *Client) openAPI() (*openapi.ClientWithResponses, error) {
	return openapi.NewClientWithResponses(
		BaseURL,
		openapi.WithHTTPClient(c.httpClient),
		openapi.WithRequestEditorFn(c.openAPIRequestEditor),
	)
}

func decodeOpenAPIBody(opName string, statusCode int, body []byte, out any) error {
	if statusCode < 200 || statusCode >= 300 {
		bodyText := string(body)
		if msg, ok := knownWanderlogServerError(opName, bodyText); ok {
			return fmt.Errorf("%s: HTTP %d: %s", opName, statusCode, msg)
		}
		return fmt.Errorf("%s: HTTP %d: %s", opName, statusCode, truncateForLog(bodyText, 500))
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%s: decoding response: %w", opName, err)
	}
	return nil
}

func knownWanderlogServerError(opName, body string) (string, bool) {
	if opName == "SearchLodgings" && strings.Contains(body, "Cannot read properties of undefined (reading 'length')") {
		return "Wanderlog lodging search is currently failing server-side before returning hotel results", true
	}
	if strings.Contains(body, "Cannot read properties of undefined (reading 'place_id')") {
		return "Wanderlog add-place is currently failing server-side while processing the place_id payload", true
	}
	return "", false
}

func parseOpenAPIDate(value, fieldName string) (openapi_types.Date, error) {
	parsed, err := time.Parse(openapi_types.DateFormat, value)
	if err != nil {
		return openapi_types.Date{}, fmt.Errorf("parsing %s date: %w", fieldName, err)
	}
	return openapi_types.Date{Time: parsed}, nil
}

func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// DoAPI performs a raw request against a Wanderlog API endpoint. It is used by
// the CLI for endpoints discovered in the web/Android bundle before they have a
// typed Go wrapper.
func (c *Client) DoAPI(method, path string, body []byte, headers map[string]string, authenticated bool) (int, []byte, error) {
	apiURL := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		trimmed := strings.TrimPrefix(path, "/")
		trimmed = strings.TrimPrefix(trimmed, "api/")
		apiURL = fmt.Sprintf("%s/%s", BaseURL, trimmed)
	}

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, apiURL, reader)
	if err != nil {
		return 0, nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if authenticated {
		if err := c.addAuthHeaders(req); err != nil {
			return 0, nil, fmt.Errorf("adding auth headers: %w", err)
		}
	} else if c.auth != nil {
		if err := c.addAuthHeaders(req); err != nil {
			return 0, nil, fmt.Errorf("adding optional auth headers: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, respBody, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return resp.StatusCode, respBody, nil
}

func (c *Client) GetTrip(key string) (*TripResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	version := openapi.ClientSchemaVersion(2)

	c.logger.WithFields(logrus.Fields{
		"tripKey": key,
	}).Debug("GetTrip request details")

	resp, err := api.GetTripPlanWithResponse(context.Background(), key, &openapi.GetTripPlanParams{
		ClientSchemaVersion: &version,
	})
	if err != nil {
		return nil, err
	}

	debugBody := string(resp.Body)
	if len(debugBody) > 500 {
		debugBody = debugBody[:500] + "..."
	}

	c.logger.WithFields(logrus.Fields{
		"tripKey":      key,
		"statusCode":   resp.StatusCode(),
		"responseBody": debugBody,
	}).Debug("GetTrip API response")

	var trip TripResponse
	if err := decodeOpenAPIBody("GetTrip", resp.StatusCode(), resp.Body, &trip); err != nil {
		return nil, err
	}

	if trip.Error != "" {
		return nil, fmt.Errorf("API error: %s", trip.Error)
	}

	return &trip, nil
}

// GetTripSections retrieves only the sections of a trip without the full trip data
func (c *Client) GetTripSections(key string) ([]ItSections, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithFields(logrus.Fields{
		"tripKey": key,
	}).Debug("GetTripSections request details")

	resp, err := api.GetTripPlanSectionsWithResponse(context.Background(), key, nil)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"tripKey":    key,
		"statusCode": resp.StatusCode(),
		"bodySize":   len(resp.Body),
	}).Debug("GetTripSections API response")

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("GetTripSections: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var response struct {
		Success bool         `json:"success"`
		Data    []ItSections `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode sections response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return response.Data, nil
}

// SearchPlaces searches for places using Wanderlog's place autocomplete API.
func (c *Client) SearchPlaces(query string, latitude, longitude *float64) (*PlaceSearchResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"query":     query,
		"latitude":  latitude,
		"longitude": longitude,
	}).Info("Searching places via Wanderlog API")

	results, err := c.searchWanderlogPlaces(query, latitude, longitude)
	if err != nil {
		return results, err
	}

	c.logger.WithFields(logrus.Fields{
		"query":       query,
		"resultCount": len(results.Places),
	}).Info("Successfully searched places via Wanderlog API")

	return results, nil
}

// SearchRestaurants searches for restaurants using Wanderlog's place autocomplete API.
func (c *Client) SearchRestaurants(query string, latitude, longitude *float64) (*PlaceSearchResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"query":     query,
		"latitude":  latitude,
		"longitude": longitude,
	}).Info("Searching restaurants via Wanderlog API")

	results, err := c.searchWanderlogPlaces(query, latitude, longitude)
	if err != nil {
		return results, err
	}

	c.logger.WithFields(logrus.Fields{
		"query":       query,
		"resultCount": len(results.Places),
	}).Info("Successfully searched restaurants via Wanderlog API")

	return results, nil
}

func (c *Client) searchWanderlogPlaces(query string, latitude, longitude *float64) (*PlaceSearchResponse, error) {
	lat, lng := 0.0, 0.0
	if latitude != nil {
		lat = *latitude
	}
	if longitude != nil {
		lng = *longitude
	}

	autocompleteResp, err := c.SearchPlacesWithWanderlog(query, lat, lng)
	if err != nil {
		return &PlaceSearchResponse{Success: false, Places: []SearchResult{}}, err
	}

	results := make([]SearchResult, 0, len(autocompleteResp.Data))
	for _, place := range autocompleteResp.Data {
		name := place.StructuredFormatting.MainText
		if name == "" {
			name = place.Description
		}
		address := place.StructuredFormatting.SecondaryText
		if address == "" {
			address = place.SecondaryText
		}
		results = append(results, SearchResult{
			ID:         place.PlaceID,
			Name:       name,
			Address:    address,
			PlaceID:    place.PlaceID,
			Categories: place.Types,
		})
	}

	return &PlaceSearchResponse{
		Success: true,
		Places:  results,
	}, nil
}

// SearchPlacesInTrips searches for places within the user's trips by query
func (c *Client) SearchPlacesInTrips(query string) (*PlaceSearchResponse, error) {
	c.logger.WithField("query", query).Debug("Searching places in user trips")

	// First get user trips
	trips, err := c.GetUserTrips()
	if err != nil {
		return nil, fmt.Errorf("failed to get user trips: %w", err)
	}

	var results []SearchResult

	// Search through each trip's places
	for _, tripData := range trips.Data {
		trip, err := c.GetTrip(tripData.Key)
		if err != nil {
			c.logger.WithField("tripKey", tripData.Key).Debug("Failed to get trip details")
			continue
		}

		// Search through place metadata
		for _, place := range trip.Resources.PlaceMetadata {
			if c.matchesQuery(place, query) {
				result := SearchResult{
					ID:         fmt.Sprintf("%d", place.ID),
					Name:       place.Name,
					Address:    place.Address,
					PlaceID:    place.PlaceID,
					Rating:     place.Rating,
					Categories: place.Categories,
					Website:    place.Website,
				}

				// Set description from generated or regular description
				if place.GeneratedDescription != nil && *place.GeneratedDescription != "" {
					result.Description = *place.GeneratedDescription
				} else if place.Description != nil && *place.Description != "" {
					result.Description = *place.Description
				}

				results = append(results, result)
			}
		}
	}

	c.logger.WithFields(logrus.Fields{
		"query":       query,
		"resultCount": len(results),
	}).Info("Successfully searched places in trips")

	return &PlaceSearchResponse{
		Success: true,
		Places:  results,
	}, nil
}

// matchesQuery checks if a place matches the search query
func (c *Client) matchesQuery(place Metadata, query string) bool {
	query = strings.ToLower(query)

	// Check name
	if strings.Contains(strings.ToLower(place.Name), query) {
		return true
	}

	// Check address
	if strings.Contains(strings.ToLower(place.Address), query) {
		return true
	}

	// Check categories
	for _, category := range place.Categories {
		if strings.Contains(strings.ToLower(category), query) {
			return true
		}
	}

	// Check description
	if place.Description != nil && strings.Contains(strings.ToLower(*place.Description), query) {
		return true
	}

	if place.GeneratedDescription != nil && strings.Contains(strings.ToLower(*place.GeneratedDescription), query) {
		return true
	}

	return false
}

// PlaceDetailsResponse represents the response from the place details API
type PlaceDetailsResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Details struct {
			Name     string `json:"name"`
			PlaceID  string `json:"place_id"`
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
			FormattedAddress         string   `json:"formatted_address"`
			Rating                   float64  `json:"rating"`
			UserRatingsTotal         int      `json:"user_ratings_total"`
			Website                  string   `json:"website"`
			InternationalPhoneNumber string   `json:"international_phone_number"`
			Types                    []string `json:"types"`
			BusinessStatus           string   `json:"business_status"`
			PhotoUrls                []string `json:"photo_urls"`
		} `json:"details"`
		CardData struct {
			ReviewsSummary string   `json:"reviewsSummary"`
			ReasonsToVisit []string `json:"reasonsToVisit"`
			Tips           []string `json:"tips"`
			PlaceID        string   `json:"placeId"`
		} `json:"cardData"`
	} `json:"data"`
}

// WanderlogAutocompleteResponse represents the response from the Wanderlog autocomplete API
type WanderlogAutocompleteResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		PlaceID              string   `json:"place_id"`
		Description          string   `json:"description"`
		Types                []string `json:"types"`
		StructuredFormatting struct {
			MainText                  string `json:"main_text"`
			MainTextMatchedSubstrings []struct {
				Offset int `json:"offset"`
				Length int `json:"length"`
			} `json:"main_text_matched_substrings"`
			SecondaryText string `json:"secondary_text"`
		} `json:"structured_formatting"`
		Type                string        `json:"type,omitempty"`
		Input               string        `json:"input,omitempty"`
		InputTextHighlights []interface{} `json:"inputTextHighlights,omitempty"`
		SeeLocations        bool          `json:"seeLocations,omitempty"`
		SecondaryText       string        `json:"secondaryText,omitempty"`
		CanSeeOnMap         bool          `json:"canSeeOnMap,omitempty"`
	} `json:"data"`
}

// GetPlaceDetails fetches detailed information about a place from Wanderlog's place details API
func (c *Client) GetPlaceDetails(placeID string) (*PlaceDetailsResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	language := "en"
	resp, err := api.GetPlaceDetailsAndCardDataWithResponse(context.Background(), &openapi.GetPlaceDetailsAndCardDataParams{
		PlaceId:  placeID,
		Language: &language,
	})
	if err != nil {
		return nil, err
	}

	var result PlaceDetailsResponse
	if err := decodeOpenAPIBody("GetPlaceDetails", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("API request was not successful: %s", truncateForLog(string(resp.Body), 500))
	}

	return &result, nil
}

// GetAllAirlines retrieves all available airlines
func (c *Client) GetAllAirlines() (*AirlinesResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetAllAirlinesInternalWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	var result AirlinesResponse
	if err := decodeOpenAPIBody("GetAllAirlines", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AutocompleteAirport searches for airports by query (path-based)
func (c *Client) AutocompleteAirport(query string) (*AirportAutocompleteResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.AutocompleteAirportWithResponse(context.Background(), query)
	if err != nil {
		return nil, err
	}
	var result AirportAutocompleteResponse
	if err := decodeOpenAPIBody("AutocompleteAirport", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AutocompleteAirportWithLocation searches for airports by query with location bias (query is path-based)
func (c *Client) AutocompleteAirportWithLocation(query string, lat, lng float64) (*AirportAutocompleteResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.AutocompleteAirportWithLocationWithResponse(context.Background(), query, func(_ context.Context, req *http.Request) error {
		values := req.URL.Query()
		values.Set("latitude", fmt.Sprintf("%f", lat))
		values.Set("longitude", fmt.Sprintf("%f", lng))
		req.URL.RawQuery = values.Encode()
		return nil
	})
	if err != nil {
		return nil, err
	}
	var result AirportAutocompleteResponse
	if err := decodeOpenAPIBody("AutocompleteAirportWithLocation", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFlightStops retrieves flight stops for a given flight.
// The API requires flightNumber (integer string), airline IATA code (airlineIata), and departure date (departDate).
func (c *Client) GetFlightStops(flightNumber, airlineIata, departureDate string) (*FlightStopsResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	date, err := parseOpenAPIDate(departureDate, "departure")
	if err != nil {
		return nil, fmt.Errorf("get flight stops: parsing departure date: %w", err)
	}
	httpResp, err := api.GetFlightStops(context.Background(), &openapi.GetFlightStopsParams{
		FlightNumber: flightNumber,
		AirlineIata:  airlineIata,
		DepartDate:   date,
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("get flight stops: reading response: %w", err)
	}
	if len(body) > 0 && body[0] == '<' {
		return nil, fmt.Errorf("API returned HTML instead of JSON (endpoint may be unavailable)")
	}

	var result FlightStopsResponse
	if err := decodeOpenAPIBody("GetFlightStops", httpResp.StatusCode, body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// WanderlogAutocompleteRequest represents the request for Wanderlog place autocomplete
type WanderlogAutocompleteRequest struct {
	Input        string `json:"input"`
	SessionToken string `json:"sessiontoken"`
	Location     struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	} `json:"location"`
	Radius   float64 `json:"radius"`
	Language string  `json:"language"`
}

// SearchPlacesWithWanderlog searches for places using Wanderlog's autocomplete API
func (c *Client) SearchPlacesWithWanderlog(query string, lat, lng float64) (*WanderlogAutocompleteResponse, error) {
	reqData := WanderlogAutocompleteRequest{
		Input:        query,
		SessionToken: fmt.Sprintf("%d", time.Now().UnixNano()), // Simple session token
		Location: struct {
			Longitude float64 `json:"longitude"`
			Latitude  float64 `json:"latitude"`
		}{
			Longitude: lng,
			Latitude:  lat,
		},
		Radius:   50000, // 50km radius
		Language: "en",
	}

	reqJSON, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.AutocompletePlacesAndSuggestionsWithResponse(context.Background(), &openapi.AutocompletePlacesAndSuggestionsParams{
		Request: string(reqJSON),
	})
	if err != nil {
		return nil, err
	}

	var result WanderlogAutocompleteResponse
	if err := decodeOpenAPIBody("SearchPlacesWithWanderlog", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("API request was not successful")
	}

	return &result, nil
}

// SearchLodgings searches for hotels/lodgings for given dates and guest count.
// The request body matches the React Native app: bounds, dates, and nested guests.
func (c *Client) SearchLodgings(query, checkIn, checkOut string, guests int) (*LodgingSearchResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	startDate, err := parseOpenAPIDate(checkIn, "check-in")
	if err != nil {
		return nil, fmt.Errorf("search lodgings: %w", err)
	}
	endDate, err := parseOpenAPIDate(checkOut, "check-out")
	if err != nil {
		return nil, fmt.Errorf("search lodgings: %w", err)
	}
	bounds, err := c.lookupGeoBounds(query)
	if err != nil {
		return nil, fmt.Errorf("search lodgings: %w", err)
	}
	if guests <= 0 {
		guests = 1
	}

	reqBody, err := json.Marshal(map[string]any{
		"bounds": bounds,
		"dates": map[string]string{
			"startDate": startDate.Format(openapi_types.DateFormat),
			"endDate":   endDate.Format(openapi_types.DateFormat),
		},
		"guests": map[string]any{
			"adultCount":   guests,
			"roomCount":    1,
			"childrenAges": []int{},
		},
		"sources": []string{"google"},
		"filters": map[string]any{
			"hotelClasses":          nil,
			"minGuestRating":        nil,
			"priceRange":            nil,
			"amenities":             nil,
			"propertyTypes":         map[string]any{"accommodationTypes": nil, "lodgingTypes": nil},
			"minBedsInRoom":         nil,
			"propertyName":          "",
			"vacationRentalFilters": nil,
			"hotelOrVacationRental": nil,
		},
		"hotelOrVacationRental": nil,
	})
	if err != nil {
		return nil, fmt.Errorf("search lodgings: marshaling request: %w", err)
	}

	httpResp, err := api.SearchLodgingsWithBody(context.Background(), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("search lodgings: reading response: %w", err)
	}

	var result LodgingSearchResponse
	if err := decodeOpenAPIBody("SearchLodgings", httpResp.StatusCode, body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("lodging search API returned success=false for query %q; response: %s", query, string(body))
	}
	if len(result.Data) == 0 && len(result.Offers) > 0 {
		result.Data = result.Offers
	}

	return &result, nil
}

func (c *Client) lookupGeoBounds(query string) ([]float64, error) {
	geos, err := c.SearchGeos()
	if err != nil {
		return nil, fmt.Errorf("looking up geo bounds for %q: %w", query, err)
	}
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return nil, fmt.Errorf("location is required")
	}
	var fallback []float64
	match := func(geo GeoIDName) bool {
		if len(geo.Bounds) != 4 {
			return false
		}
		name := strings.ToLower(strings.TrimSpace(geo.Name))
		if name == normalized {
			fallback = geo.Bounds
			return true
		}
		if fallback == nil && strings.Contains(name, normalized) {
			fallback = geo.Bounds
		}
		return false
	}
	for _, geo := range geos.Cities {
		if match(geo) {
			return fallback, nil
		}
	}
	for _, geo := range geos.Countries {
		if match(geo) {
			return fallback, nil
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	if bounds, err := c.lookupPlaceBounds(query); err == nil {
		return bounds, nil
	}
	return nil, fmt.Errorf("no geo bounds found for %q; use search_geos and a supported destination name", query)
}

func (c *Client) lookupPlaceBounds(query string) ([]float64, error) {
	results, err := c.SearchPlacesWithWanderlog(query, 0, 0)
	if err != nil {
		return nil, err
	}
	for _, place := range results.Data {
		if place.PlaceID == "" {
			continue
		}
		details, err := c.GetPlaceDetails(place.PlaceID)
		if err != nil || !details.Success {
			continue
		}
		lat := details.Data.Details.Geometry.Location.Lat
		lng := details.Data.Details.Geometry.Location.Lng
		if lat == 0 && lng == 0 {
			continue
		}
		const delta = 0.35
		return []float64{lng - delta, lat - delta, lng + delta, lat + delta}, nil
	}
	return nil, fmt.Errorf("no place geometry found for %q", query)
}

// GetGooglePriceRates retrieves Google price rates for a specific lodging property
func (c *Client) GetGooglePriceRates(propertyID string) (*GooglePriceRatesResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetGooglePriceRatesWithResponse(context.Background(), &openapi.GetGooglePriceRatesParams{
		Id:     propertyID,
		Dates:  "{}",
		Guests: "{}",
	})
	if err != nil {
		return nil, err
	}
	var result GooglePriceRatesResponse
	if err := decodeOpenAPIBody("GetGooglePriceRates", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
