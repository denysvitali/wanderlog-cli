package wanderlog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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

func (c *Client) GetTrip(key string) (*TripResponse, error) {
	url := fmt.Sprintf("%s/tripPlans/%s?clientSchemaVersion=2", BaseURL, key)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	c.logger.WithFields(logrus.Fields{
		"tripKey": key,
		"url":     url,
		"headers": req.Header,
	}).Debug("GetTrip request details")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log response for debugging
	debugBody := string(respBody)
	if len(debugBody) > 500 {
		debugBody = debugBody[:500] + "..."
	}

	c.logger.WithFields(logrus.Fields{
		"tripKey":      key,
		"statusCode":   resp.StatusCode,
		"responseBody": debugBody,
	}).Debug("GetTrip API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s - Response: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var trip TripResponse
	if err := json.Unmarshal(respBody, &trip); err != nil {
		return nil, fmt.Errorf("failed to decode trip response: %w", err)
	}

	// Check if the API returned an error
	if trip.Error != "" {
		return nil, fmt.Errorf("API error: %s", trip.Error)
	}

	return &trip, nil
}

// GetTripSections retrieves only the sections of a trip without the full trip data
func (c *Client) GetTripSections(key string) ([]ItSections, error) {
	url := fmt.Sprintf("%s/tripPlans/%s/sections", BaseURL, key)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	c.logger.WithFields(logrus.Fields{
		"tripKey": key,
		"url":     url,
	}).Debug("GetTripSections request details")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"tripKey":    key,
		"statusCode": resp.StatusCode,
		"bodySize":   len(respBody),
	}).Debug("GetTripSections API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s - Response: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	// The API returns sections wrapped in a success response with "data" key
	var response struct {
		Success bool         `json:"success"`
		Data    []ItSections `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to decode sections response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return response.Data, nil
}

// SearchPlaces searches for places using Google Places API (New)
func (c *Client) SearchPlaces(query string, latitude, longitude *float64, apiKey string) (*PlaceSearchResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"query":     query,
		"latitude":  latitude,
		"longitude": longitude,
	}).Info("Searching places via Google Places API (New)")

	if apiKey == "" {
		return &PlaceSearchResponse{
			Success: false,
			Places:  []SearchResult{},
		}, fmt.Errorf("Google Places API key is required. Please provide an API key using --api-key flag")
	}

	// New Google Places API endpoint
	baseURL := "https://places.googleapis.com/v1/places:searchText"

	// Build request body for new API
	requestBody := map[string]interface{}{
		"textQuery": query,
	}

	// Add location bias if coordinates provided
	if latitude != nil && longitude != nil {
		requestBody["locationBias"] = map[string]interface{}{
			"circle": map[string]interface{}{
				"center": map[string]interface{}{
					"latitude":  *latitude,
					"longitude": *longitude,
				},
				"radius": 50000.0, // 50km radius
			},
		}
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create POST request
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for new API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.formattedAddress,places.location,places.rating,places.types")

	// Make the API request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Google Places API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the response for debugging
	bodyStr := string(body)
	debugBody := bodyStr
	if len(debugBody) > 500 {
		debugBody = debugBody[:500] + "..."
	}
	c.logger.WithFields(logrus.Fields{
		"statusCode": resp.StatusCode,
		"body":       debugBody,
		"url":        baseURL,
	}).Debug("Google Places API response")

	// Check for non-200 status codes
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Google Places API returned status %d: %s", resp.StatusCode, bodyStr)
	}

	// Parse the response - New Places API format
	var googleResp struct {
		Places []struct {
			ID          string `json:"id"`
			DisplayName struct {
				Text string `json:"text"`
			} `json:"displayName"`
			FormattedAddress string `json:"formattedAddress"`
			Location         struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
			Rating float64  `json:"rating"`
			Types  []string `json:"types"`
		} `json:"places"`
	}

	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to decode Google Places API response: %w", err)
	}

	// Convert to our format
	var results []SearchResult
	for _, place := range googleResp.Places {
		result := SearchResult{
			ID:         place.ID,
			Name:       place.DisplayName.Text,
			Address:    place.FormattedAddress,
			PlaceID:    place.ID,
			Latitude:   place.Location.Latitude,
			Longitude:  place.Location.Longitude,
			Rating:     place.Rating,
			Categories: place.Types,
		}
		results = append(results, result)
	}

	c.logger.WithFields(logrus.Fields{
		"query":       query,
		"resultCount": len(results),
	}).Info("Successfully searched places via Google Places API (New)")

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

// WanderlLogAutocompleteResponse represents the response from the Wanderlog autocomplete API
type WanderlLogAutocompleteResponse struct {
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
	url := fmt.Sprintf("https://wanderlog.com/api/placesAPI/getPlaceDetailsAndCardData?placeId=%s&language=en", placeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result PlaceDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("API request was not successful")
	}

	return &result, nil
}

// GetAllAirlines retrieves all available airlines
func (c *Client) GetAllAirlines() (*AirlinesResponse, error) {
	url := fmt.Sprintf("%s/flights/allAirlines", BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result AirlinesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AutocompleteAirport searches for airports by query
func (c *Client) AutocompleteAirport(query string) (*AirportAutocompleteResponse, error) {
	url := fmt.Sprintf("%s/flights/autocompleteAirport?query=%s", BaseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result AirportAutocompleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AutocompleteAirportWithLocation searches for airports by query with location bias
func (c *Client) AutocompleteAirportWithLocation(query string, lat, lng float64) (*AirportAutocompleteResponse, error) {
	url := fmt.Sprintf("%s/flights/autocompleteAirportWithLocation?query=%s&latitude=%f&longitude=%f",
		BaseURL, url.QueryEscape(query), lat, lng)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result AirportAutocompleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFlightStops retrieves flight stops for a given flight number
func (c *Client) GetFlightStops(flightNumber string) (*FlightStopsResponse, error) {
	url := fmt.Sprintf("%s/flights/flightStopsLista?flightNumber=%s", BaseURL, url.QueryEscape(flightNumber))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result FlightStopsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// WanderlLogAutocompleteRequest represents the request for Wanderlog place autocomplete
type WanderlLogAutocompleteRequest struct {
	Input        string `json:"input"`
	SessionToken string `json:"sessiontoken"`
	Location     struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	} `json:"location"`
	Radius   float64 `json:"radius"`
	Language string  `json:"language"`
}

// SearchPlacesWithWanderllog searches for places using Wanderlog's autocomplete API
func (c *Client) SearchPlacesWithWanderllog(query string, lat, lng float64) (*WanderlLogAutocompleteResponse, error) {
	reqData := WanderlLogAutocompleteRequest{
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

	apiURL := fmt.Sprintf("https://wanderlog.com/api/placesAPI/autocomplete/v2?request=%s",
		url.QueryEscape(string(reqJSON)))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result WanderlLogAutocompleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("API request was not successful")
	}

	return &result, nil
}

// SearchLodgings searches for hotels/lodgings for given dates and guest count
func (c *Client) SearchLodgings(query, checkIn, checkOut string, guests int) (*LodgingSearchResponse, error) {
	apiURL := fmt.Sprintf("%s/lodging/searchLodgings", BaseURL)

	requestBody := map[string]interface{}{
		"query":          query,
		"checkIn":        checkIn,
		"checkOut":       checkOut,
		"numberOfGuests": guests,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s - Response: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var result LodgingSearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetGooglePriceRates retrieves Google price rates for a specific lodging property
func (c *Client) GetGooglePriceRates(propertyID string) (*GooglePriceRatesResponse, error) {
	apiURL := fmt.Sprintf("%s/lodging/getGooglePriceRates", BaseURL)

	requestBody := map[string]interface{}{
		"propertyId": propertyID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.auth != nil && c.auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.auth.SessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s - Response: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var result GooglePriceRatesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
