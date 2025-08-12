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
	Place struct {
		PlaceID   string  `json:"place_id"`
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"place"`
	Text string `json:"text"`
}

// OperationRequest represents a batch operation request
type OperationRequest struct {
	Ops []Operation `json:"ops"`
}

// Operation represents a single operation in the operational transform system
type Operation struct {
	Type     string      `json:"type"`
	Path     string      `json:"path,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	OldValue interface{} `json:"oldValue,omitempty"`
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

// AddPlace adds a place to a trip section
func (c *Client) AddPlace(tripKey string, sectionID int, req AddPlaceRequest) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for adding places")
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
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
func (c *Client) NukeTripPlaces(tripKey string) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for nuking trip places")
	}

	// Try multiple approaches to clear problematic place data
	operations := []Operation{
		// Try to clear common section blocks that might contain places
		{Type: "replace", Path: "/itinerary/sections/0/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/1/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/2/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/3/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/4/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/5/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/6/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/7/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/8/blocks", Value: []interface{}{}},
		{Type: "replace", Path: "/itinerary/sections/9/blocks", Value: []interface{}{}},
		// Clear any place metadata that might be corrupted
		{Type: "replace", Path: "/resources/placeMetadata", Value: map[string]interface{}{}},
	}

	err := c.ApplyOperations(tripKey, operations)
	if err != nil {
		return fmt.Errorf("failed to nuke trip places: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
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
