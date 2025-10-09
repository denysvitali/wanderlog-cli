package wanderlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateTripRequest represents a request to create a new trip
type CreateTripRequest struct {
	Title     string `json:"title"`
	StartDate string `json:"startDate,omitempty"` // YYYY-MM-DD format
	EndDate   string `json:"endDate,omitempty"`   // YYYY-MM-DD format
	Privacy   string `json:"privacy,omitempty"`   // "public", "private", "unlisted"
}

// CreateTripResponse represents the response from creating a trip
type CreateTripResponse struct {
	Success  bool `json:"success"`
	TripPlan struct {
		ID      int    `json:"id"`
		Key     string `json:"key"`
		EditKey string `json:"editKey"`
		Title   string `json:"title"`
	} `json:"tripPlan"`
}

// UpdateTripRequest represents a request to update trip metadata
type UpdateTripRequest struct {
	Title     string `json:"title,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Privacy   string `json:"privacy,omitempty"`
}

// AddPlaceRequest represents a request to add a place to a trip
type AddPlaceRequest struct {
	Place AddPlaceInfo `json:"place"`
	Text  string       `json:"text"`
}

// AddPlaceInfo represents the place information when adding a place
type AddPlaceInfo struct {
	PlaceID  string `json:"place_id,omitempty"`  // API uses snake_case
	Name     string `json:"name"`
	Geometry *struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry,omitempty"`
}

// OperationRequest represents a batch operation request
type OperationRequest struct {
	Ops []Operation `json:"ops"`
}

// Operation represents a single ShareDB JSON0 operation
// ShareDB uses a specific format with:
// - p: path as array of keys/indices
// - oi/od: object insert/delete (for replacing object values)
// - li/ld: list insert/delete (for array operations)
type Operation struct {
	P  []interface{} `json:"p"`            // Path as array (e.g., ["itinerary", "sections", 0, "blocks", 1])
	OI interface{}   `json:"oi,omitempty"` // Object insert (new value for replace)
	OD interface{}   `json:"od,omitempty"` // Object delete (old value for replace)
	LI interface{}   `json:"li,omitempty"` // List insert (for array insertions)
	LD interface{}   `json:"ld,omitempty"` // List delete (for array deletions)
}

// ShareDB Operation Helpers

// ReplaceInObject creates a ShareDB operation to replace an object field
// Path should be an array like: []interface{}{"itinerary", "sections", 0, "heading"}
func ReplaceInObject(path []interface{}, oldValue, newValue interface{}) Operation {
	return Operation{
		P:  path,
		OD: oldValue,
		OI: newValue,
	}
}

// InsertInObject creates a ShareDB operation to insert a new object field
func InsertInObject(path []interface{}, value interface{}) Operation {
	return Operation{
		P:  path,
		OI: value,
	}
}

// DeleteInObject creates a ShareDB operation to delete an object field
func DeleteInObject(path []interface{}, oldValue interface{}) Operation {
	return Operation{
		P:  path,
		OD: oldValue,
	}
}

// InsertInList creates a ShareDB operation to insert an item into an array at a specific index
func InsertInList(path []interface{}, index int, value interface{}) Operation {
	pathWithIndex := append(path, index)
	return Operation{
		P:  pathWithIndex,
		LI: value,
	}
}

// DeleteFromList creates a ShareDB operation to delete an item from an array at a specific index
func DeleteFromList(path []interface{}, index int, oldValue interface{}) Operation {
	pathWithIndex := append(path, index)
	return Operation{
		P:  pathWithIndex,
		LD: oldValue,
	}
}

// ReplaceInList creates a ShareDB operation to replace an item in an array
func ReplaceInList(path []interface{}, index int, oldValue, newValue interface{}) Operation {
	pathWithIndex := append(path, index)
	return Operation{
		P:  pathWithIndex,
		LD: oldValue,
		LI: newValue,
	}
}

// FindSectionIndex finds the array index of a section by its ID
func FindSectionIndex(sections []ItSections, sectionID int) int {
	for i, section := range sections {
		if section.ID == sectionID {
			return i
		}
	}
	return -1
}

// CreateTrip creates a new trip plan
func (c *Client) CreateTrip(req CreateTripRequest) (*CreateTripResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for creating trips")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling create trip request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", BaseURL+"/tripPlans", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithField("title", req.Title).Debug("Creating new trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var createResp CreateTripResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !createResp.Success {
		return nil, fmt.Errorf("failed to create trip")
	}

	c.logger.WithFields(map[string]interface{}{
		"tripID": createResp.TripPlan.ID,
		"key":    createResp.TripPlan.Key,
		"title":  createResp.TripPlan.Title,
	}).Info("Successfully created trip")

	return &createResp, nil
}

// DeleteTrip deletes a trip plan
func (c *Client) DeleteTrip(tripKey string) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for deleting trips")
	}

	url := fmt.Sprintf("%s/tripPlans/%s", BaseURL, tripKey)
	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithField("tripKey", tripKey).Debug("Deleting trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully deleted trip")
	return nil
}

// ValidateAddPlaceRequest validates the AddPlaceRequest structure
func ValidateAddPlaceRequest(req AddPlaceRequest) error {
	if req.Place.PlaceID == "" {
		return fmt.Errorf("place_id is required")
	}
	if req.Place.Name == "" {
		return fmt.Errorf("place name is required")
	}
	if req.Place.Latitude < -90 || req.Place.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", req.Place.Latitude)
	}
	if req.Place.Longitude < -180 || req.Place.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", req.Place.Longitude)
	}
	return nil
}

// AddPlace adds a place to a trip section
func (c *Client) AddPlace(tripKey string, sectionID int, req AddPlaceRequest) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for adding places")
	}

	// Validate request
	if err := ValidateAddPlaceRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	var url string
	if sectionID > 0 {
		url = fmt.Sprintf("%s/tripPlans/%s/sections/%d/place", BaseURL, tripKey, sectionID)
	} else {
		url = fmt.Sprintf("%s/tripPlans/%s/sections/place", BaseURL, tripKey)
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling add place request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":     tripKey,
		"sectionID":   sectionID,
		"requestBody": string(reqBody),
		"url":         url,
	}).Debug("AddPlace request details")

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeName": req.Place.Name,
	}).Debug("Adding place to trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":      tripKey,
		"sectionID":    sectionID,
		"placeName":    req.Place.Name,
		"statusCode":   resp.StatusCode,
		"responseBody": string(respBody),
	}).Debug("AddPlace API response")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s - Response: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	// Try to parse the response to check for API-level errors
	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		c.logger.WithField("responseBody", string(respBody)).Warn("Could not parse API response as JSON")
	} else {
		// Check if the response indicates success
		if success, ok := apiResp["success"]; ok {
			if successBool, ok := success.(bool); ok && !successBool {
				// API returned success: false
				errorMsg := "unknown error"
				if msg, ok := apiResp["error"]; ok {
					if msgStr, ok := msg.(string); ok {
						errorMsg = msgStr
					}
				}
				return fmt.Errorf("API request failed: %s", errorMsg)
			}
		}
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"placeName": req.Place.Name,
	}).Info("Successfully added place to trip")

	return nil
}

// RemovePlace removes a place from a trip section
func (c *Client) RemovePlace(tripKey string, sectionID, placeID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for removing places")
	}

	var url string
	if sectionID > 0 {
		url = fmt.Sprintf("%s/tripPlans/%s/sections/%d/place/%d", BaseURL, tripKey, sectionID, placeID)
	} else {
		url = fmt.Sprintf("%s/tripPlans/%s/sections/place/%d", BaseURL, tripKey, placeID)
	}

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeID":   placeID,
	}).Debug("Removing place from trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"placeID": placeID,
	}).Info("Successfully removed place from trip")

	return nil
}

// ApplyOperations applies a batch of operations to a trip (for complex edits)
func (c *Client) ApplyOperations(tripKey string, ops []Operation) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for applying operations")
	}

	opReq := OperationRequest{Ops: ops}
	reqBody, err := json.Marshal(opReq)
	if err != nil {
		return fmt.Errorf("marshaling operations request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/%s/applyOps", BaseURL, tripKey)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":    tripKey,
		"operations": len(ops),
	}).Debug("Applying operations to trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":      tripKey,
		"operations":   len(ops),
		"statusCode":   resp.StatusCode,
		"responseBody": string(respBody),
	}).Debug("ApplyOperations API response")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s - Response: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	// Try to parse the response to check for API-level errors
	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		c.logger.WithField("responseBody", string(respBody)).Warn("Could not parse API response as JSON")
	} else {
		// Check if the response indicates success
		if success, ok := apiResp["success"]; ok {
			if successBool, ok := success.(bool); ok && !successBool {
				// API returned success: false
				errorMsg := "unknown error"
				if msg, ok := apiResp["error"]; ok {
					if msgStr, ok := msg.(string); ok {
						errorMsg = msgStr
					}
				}
				// Also check for messages array
				if messages, ok := apiResp["messages"]; ok {
					if msgArray, ok := messages.([]interface{}); ok && len(msgArray) > 0 {
						if firstMsg, ok := msgArray[0].(string); ok {
							errorMsg = firstMsg
						}
					}
				}
				return fmt.Errorf("API request failed: %s", errorMsg)
			}
		}
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":    tripKey,
		"operations": len(ops),
	}).Info("Successfully applied operations to trip")

	return nil
}

// ClearSectionBlocks removes all blocks from a specific section using operational transforms
func (c *Client) ClearSectionBlocks(tripKey string, sectionID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for clearing section blocks")
	}

	// Create an operation to replace the blocks array with an empty array
	clearOp := Operation{
		Type:  "replace",
		Path:  fmt.Sprintf("/itinerary/sections/%d/blocks", sectionID),
		Value: []interface{}{},
	}

	err := c.ApplyOperations(tripKey, []Operation{clearOp})
	if err != nil {
		return fmt.Errorf("failed to clear section blocks: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
	}).Info("Successfully cleared all blocks from section")

	return nil
}

// DeleteSection removes an entire section from a trip using operational transforms
func (c *Client) DeleteSection(tripKey string, sectionID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for deleting sections")
	}

	// Create an operation to remove the section
	deleteOp := Operation{
		Type: "remove",
		Path: fmt.Sprintf("/itinerary/sections/%d", sectionID),
	}

	err := c.ApplyOperations(tripKey, []Operation{deleteOp})
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
	}).Info("Successfully deleted section")

	return nil
}

// NukeTripPlaces removes all place blocks from all sections in a trip using operational transforms
// This function first fetches the trip to determine which sections exist, then clears them
func (c *Client) NukeTripPlaces(tripKey string) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for nuking trip places")
	}

	// First fetch the trip to see what sections actually exist
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("failed to fetch trip: %w", err)
	}

	if len(trip.TripPlan.Itinerary.Sections) == 0 {
		c.logger.WithField("tripKey", tripKey).Info("No sections found in trip, nothing to clear")
		return nil
	}

	// Build operations only for sections that exist
	operations := []Operation{}
	for i := range trip.TripPlan.Itinerary.Sections {
		operations = append(operations, Operation{
			Type:  "replace",
			Path:  fmt.Sprintf("/itinerary/sections/%d/blocks", i),
			Value: []interface{}{},
		})
	}

	// Also clear place metadata
	operations = append(operations, Operation{
		Type:  "replace",
		Path:  "/resources/placeMetadata",
		Value: map[string]interface{}{},
	})

	c.logger.WithFields(map[string]interface{}{
		"tripKey":         tripKey,
		"sectionsCleared": len(trip.TripPlan.Itinerary.Sections),
	}).Debug("Clearing sections from trip")

	err = c.ApplyOperations(tripKey, operations)
	if err != nil {
		return fmt.Errorf("failed to nuke trip places: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":  tripKey,
		"sections": len(trip.TripPlan.Itinerary.Sections),
	}).Info("Successfully nuked all place data from trip")

	return nil
}

// CopyTrip creates a copy of an existing trip
func (c *Client) CopyTrip(sourceKey string) (*CreateTripResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for copying trips")
	}

	url := fmt.Sprintf("%s/tripPlans/copy/%s", BaseURL, sourceKey)
	httpReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithField("sourceKey", sourceKey).Debug("Copying trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var copyResp CreateTripResponse
	if err := json.NewDecoder(resp.Body).Decode(&copyResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !copyResp.Success {
		return nil, fmt.Errorf("failed to copy trip")
	}

	c.logger.WithFields(map[string]interface{}{
		"sourceKey": sourceKey,
		"newKey":    copyResp.TripPlan.Key,
		"title":     copyResp.TripPlan.Title,
	}).Info("Successfully copied trip")

	return &copyResp, nil
}
