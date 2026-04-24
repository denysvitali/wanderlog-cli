package wanderlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

// Type aliases for backward compatibility
type (
	CreateTripRequest        = models.CreateTripRequest
	CreateTripResponse       = models.CreateTripResponse
	UpdateTripRequest        = models.UpdateTripRequest
	AddPlaceRequest          = models.AddPlaceRequest
	AddPlaceInfo             = models.AddPlaceInfo
	OperationRequest         = models.OperationRequest
	Operation                = models.Operation
	SendInvitesRequest       = models.SendInvitesRequest
	TripInvite               = models.TripInvite
	LikeCount                = models.LikeCount
	ShareKeyPermissions      = models.ShareKeyPermissions
	ShareKeyResponse         = models.ShareKeyResponse
	TripFlightsResponse      = models.TripFlightsResponse
	TripFlight               = models.TripFlight
	FlightAirport            = models.FlightAirport
	AutofillDayRequest       = models.AutofillDayRequest
	AutofillDayResponse      = models.AutofillDayResponse
	ChecklistSectionRequest  = models.ChecklistSectionRequest
	ChecklistSectionResponse = models.ChecklistSectionResponse
	ChecklistItem            = models.ChecklistItem
	ExportTripResponse       = models.ExportTripResponse
)

// Operation helper functions
var (
	ReplaceInObject = models.ReplaceInObject
	InsertInObject  = models.InsertInObject
	DeleteInObject  = models.DeleteInObject
	InsertInList    = models.InsertInList
	DeleteFromList  = models.DeleteFromList
	ReplaceInList   = models.ReplaceInList
)

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

	if len(req.GeoIDs) == 0 {
		return nil, fmt.Errorf("at least one geo id is required for creating trips")
	}
	if req.Type == "" {
		req.Type = "plan"
	}
	if req.Privacy == "" {
		req.Privacy = "private"
	}
	if req.InitialMapsPlaceIDs == nil {
		req.InitialMapsPlaceIDs = []int{}
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", BaseURL+"/tripPlans", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"title":  req.Title,
		"geoIDs": req.GeoIDs,
		"type":   req.Type,
	}).Debug("Creating trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(respBody),
	}).Debug("CreateTrip API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var createResp struct {
		Success  bool                   `json:"success"`
		TripPlan models.TripPlanSummary `json:"tripPlan"`
		Data     models.TripPlanSummary `json:"data"`
	}
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !createResp.Success {
		c.logger.WithField("response", string(respBody)).Error("Trip creation failed")
		return nil, fmt.Errorf("failed to create trip - response: %s", string(respBody))
	}

	tripPlan := createResp.TripPlan
	if tripPlan.Key == "" {
		tripPlan = createResp.Data
	}

	c.logger.WithFields(map[string]interface{}{
		"tripID": tripPlan.ID,
		"key":    tripPlan.Key,
		"title":  tripPlan.Title,
	}).Info("Successfully created trip")

	return &CreateTripResponse{
		Success:  createResp.Success,
		TripPlan: tripPlan,
	}, nil
}

// CreateExampleTrip creates a new trip plan with example data (no body required)
func (c *Client) CreateExampleTrip() (*CreateTripResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for creating trips")
	}

	httpReq, err := http.NewRequest("POST", BaseURL+"/tripPlans/createExampleTripPlan", nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.Debug("Creating example trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(respBody),
	}).Debug("CreateExampleTrip API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	// The createExampleTripPlan response uses "data" with viewKey (like CopyTripResponse)
	var exampleResp models.CopyTripResponse
	if err := json.Unmarshal(respBody, &exampleResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !exampleResp.Success {
		c.logger.WithField("response", string(respBody)).Error("Example trip creation failed")
		return nil, fmt.Errorf("failed to create example trip - response: %s", string(respBody))
	}

	// Convert to CreateTripResponse format
	createResp := &CreateTripResponse{
		Success: exampleResp.Success,
		TripPlan: models.TripPlanSummary{
			ID:    exampleResp.Data.ID,
			Key:   exampleResp.Data.Key,
			Title: exampleResp.Data.Title,
		},
	}

	c.logger.WithFields(map[string]interface{}{
		"tripID": createResp.TripPlan.ID,
		"key":    createResp.TripPlan.Key,
		"title":  createResp.TripPlan.Title,
	}).Info("Successfully created example trip")

	return createResp, nil
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

// UpdateTrip updates trip metadata (title, dates, privacy) using ShareDB operations
func (c *Client) UpdateTrip(tripKey string, req UpdateTripRequest) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for updating trips")
	}

	// First, get current trip to get old values for operations
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	// Build operations to replace fields
	ops := []models.Operation{}

	if req.Title != "" && req.Title != trip.TripPlan.Title {
		ops = append(ops, models.ReplaceInObject(
			[]interface{}{"title"},
			trip.TripPlan.Title,
			req.Title,
		))
	}

	if req.StartDate != "" && req.StartDate != trip.TripPlan.StartDate {
		ops = append(ops, models.ReplaceInObject(
			[]interface{}{"startDate"},
			trip.TripPlan.StartDate,
			req.StartDate,
		))
	}

	if req.EndDate != "" && req.EndDate != trip.TripPlan.EndDate {
		ops = append(ops, models.ReplaceInObject(
			[]interface{}{"endDate"},
			trip.TripPlan.EndDate,
			req.EndDate,
		))
	}

	if req.Privacy != "" && req.Privacy != trip.TripPlan.Privacy {
		ops = append(ops, models.ReplaceInObject(
			[]interface{}{"privacy"},
			trip.TripPlan.Privacy,
			req.Privacy,
		))
	}

	if len(ops) == 0 {
		c.logger.Debug("No changes to apply")
		return nil
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"title":     req.Title,
		"startDate": req.StartDate,
		"endDate":   req.EndDate,
		"privacy":   req.Privacy,
		"numOps":    len(ops),
	}).Debug("Updating trip via operations")

	// Apply the operations
	if err := c.ApplyOperations(tripKey, ops); err != nil {
		return fmt.Errorf("applying operations: %w", err)
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully updated trip")

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
	if req.Place.Geometry != nil {
		lat := req.Place.Geometry.Location.Lat
		lng := req.Place.Geometry.Location.Lng
		if lat < -90 || lat > 90 {
			return fmt.Errorf("latitude must be between -90 and 90, got %f", lat)
		}
		if lng < -180 || lng > 180 {
			return fmt.Errorf("longitude must be between -180 and 180, got %f", lng)
		}
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
	clearOp := ReplaceInObject(
		[]interface{}{"itinerary", "sections", sectionID, "blocks"},
		[]interface{}{}, // old value placeholder for ShareDB OD field
		[]interface{}{},
	)

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
	deleteOp := DeleteFromList(
		[]interface{}{"itinerary", "sections"},
		sectionID,
		map[string]interface{}{}, // old value placeholder for ShareDB LD field
	)

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
		operations = append(operations, ReplaceInObject(
			[]interface{}{"itinerary", "sections", i, "blocks"},
			[]interface{}{}, // old value placeholder for ShareDB OD field
			[]interface{}{},
		))
	}

	// Also clear place metadata
	operations = append(operations, ReplaceInObject(
		[]interface{}{"resources", "placeMetadata"},
		[]interface{}{}, // old value placeholder for ShareDB OD field
		[]interface{}{},
	))

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

// MovePlace moves a place from one section to another at a specific position
func (c *Client) MovePlace(tripKey string, placeID, fromSectionID, toSectionID, position int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for moving places")
	}

	// First, get the current trip to find the place data
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	fromIdx := FindSectionIndex(trip.TripPlan.Itinerary.Sections, fromSectionID)
	if fromIdx < 0 {
		return fmt.Errorf("source section %d not found", fromSectionID)
	}

	toIdx := FindSectionIndex(trip.TripPlan.Itinerary.Sections, toSectionID)
	if toIdx < 0 {
		return fmt.Errorf("destination section %d not found", toSectionID)
	}

	// Find the block index of the place in the source section
	blockIdx := -1
	var blockData interface{}
	for i, block := range trip.TripPlan.Itinerary.Sections[fromIdx].Blocks {
		if block.ID == placeID {
			blockIdx = i
			blockData = block
			break
		}
	}
	if blockIdx < 0 {
		return fmt.Errorf("place %d not found in section %d", placeID, fromSectionID)
	}

	// Build operations: delete from source, insert into destination
	ops := []Operation{
		DeleteFromList(
			[]interface{}{"itinerary", "sections", fromIdx, "blocks"},
			blockIdx,
			blockData,
		),
		InsertInList(
			[]interface{}{"itinerary", "sections", toIdx, "blocks"},
			position,
			blockData,
		),
	}

	if err := c.ApplyOperations(tripKey, ops); err != nil {
		return fmt.Errorf("applying move operations: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":       tripKey,
		"placeID":       placeID,
		"fromSectionID": fromSectionID,
		"toSectionID":   toSectionID,
		"position":      position,
	}).Info("Successfully moved place")

	return nil
}

// ReorderPlaces reorders places within a section by replacing the blocks list
func (c *Client) ReorderPlaces(tripKey string, sectionID int, placeIDs []int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for reordering places")
	}

	// First, get the current trip to find the section data
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	sectionIdx := FindSectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
	if sectionIdx < 0 {
		return fmt.Errorf("section %d not found", sectionID)
	}

	section := trip.TripPlan.Itinerary.Sections[sectionIdx]

	// Build a map of block ID -> block data
	blockMap := make(map[int]interface{})
	for _, block := range section.Blocks {
		blockMap[block.ID] = block
	}

	// Build the new blocks list in the requested order
	newBlocks := make([]interface{}, 0, len(placeIDs))
	for _, id := range placeIDs {
		block, ok := blockMap[id]
		if !ok {
			return fmt.Errorf("place %d not found in section %d", id, sectionID)
		}
		newBlocks = append(newBlocks, block)
	}

	// Replace the entire blocks array
	ops := []Operation{
		ReplaceInObject(
			[]interface{}{"itinerary", "sections", sectionIdx, "blocks"},
			section.Blocks,
			newBlocks,
		),
	}

	if err := c.ApplyOperations(tripKey, ops); err != nil {
		return fmt.Errorf("applying reorder operations: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeIDs":  placeIDs,
	}).Info("Successfully reordered places")

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

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(respBody),
	}).Debug("CopyTrip API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var copyResp models.CopyTripResponse
	if err := json.Unmarshal(respBody, &copyResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !copyResp.Success {
		c.logger.WithField("response", string(respBody)).Error("Copy trip failed")
		return nil, fmt.Errorf("failed to copy trip - response: %s", string(respBody))
	}

	c.logger.WithFields(map[string]interface{}{
		"sourceKey": sourceKey,
		"newKey":    copyResp.Data.Key,
		"title":     copyResp.Data.Title,
	}).Info("Successfully copied trip")

	// Convert to CreateTripResponse format for compatibility
	return &CreateTripResponse{
		Success: copyResp.Success,
		TripPlan: models.TripPlanSummary{
			ID:    copyResp.Data.ID,
			Key:   copyResp.Data.Key,
			Title: copyResp.Data.Title,
		},
	}, nil
}

// RestoreTrip restores a soft-deleted trip plan
func (c *Client) RestoreTrip(tripKey string) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for restoring trips")
	}

	url := fmt.Sprintf("%s/tripPlans/%s/restore", BaseURL, tripKey)
	httpReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithField("tripKey", tripKey).Debug("Restoring trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully restored trip")
	return nil
}

// SendTripInvites sends invites for people to edit a trip plan
func (c *Client) SendTripInvites(tripKey string, req SendInvitesRequest) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for sending invites")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling send invites request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/%s/invite", BaseURL, tripKey)
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
		"tripKey":  tripKey,
		"invitees": req.Invitees,
	}).Debug("Sending trip invites")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully sent trip invites")
	return nil
}

// ListTripInvites lists all invites that have been sent out for a trip plan
func (c *Client) ListTripInvites(tripKey string) ([]TripInvite, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for listing invites")
	}

	url := fmt.Sprintf("%s/tripPlans/%s/invites", BaseURL, tripKey)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var invites []TripInvite
	if err := json.NewDecoder(resp.Body).Decode(&invites); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return invites, nil
}

// SetLike likes or unlikes a trip plan
func (c *Client) SetLike(tripKey string, liked bool) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for liking trips")
	}

	reqBody, err := json.Marshal(models.SetLikeRequest{Liked: liked})
	if err != nil {
		return fmt.Errorf("marshaling set like request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/%s/like", BaseURL, tripKey)
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
		"tripKey": tripKey,
		"liked":   liked,
	}).Debug("Setting like status for trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully set like status")
	return nil
}

// GetLikeCount gets whether we've liked a trip plan and the total number of likes
func (c *Client) GetLikeCount(tripKey string) (*LikeCount, error) {
	reqBody, err := json.Marshal(map[string][]string{"keys": {tripKey}})
	if err != nil {
		return nil, fmt.Errorf("marshaling like count request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/likes", BaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	// Auth is optional for this endpoint
	if c.auth != nil {
		if err := c.addAuthHeaders(httpReq); err != nil {
			return nil, fmt.Errorf("adding auth headers: %w", err)
		}
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var bulkResp struct {
		Success bool `json:"success"`
		Data    []struct {
			Like      bool `json:"like"`
			LikeCount int  `json:"likeCount"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bulkResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !bulkResp.Success {
		return nil, fmt.Errorf("failed to get like count")
	}
	if len(bulkResp.Data) == 0 {
		return &LikeCount{}, nil
	}

	return &LikeCount{
		Count:     bulkResp.Data[0].LikeCount,
		UserLiked: bulkResp.Data[0].Like,
	}, nil
}

// AddCollaborator adds a new collaborator to a trip plan with edit access
func (c *Client) AddCollaborator(tripKey string, userID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for adding collaborators")
	}

	reqBody, err := json.Marshal(models.CollaboratorRequest{UserID: userID})
	if err != nil {
		return fmt.Errorf("marshaling add collaborator request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/%s/collaborator", BaseURL, tripKey)
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
		"tripKey": tripKey,
		"userID":  userID,
	}).Debug("Adding collaborator to trip")

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
		"userID":  userID,
	}).Info("Successfully added collaborator")

	return nil
}

// RemoveCollaborator removes a tripmate from a trip plan
func (c *Client) RemoveCollaborator(tripKey string, userID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for removing collaborators")
	}

	reqBody, err := json.Marshal(models.CollaboratorRequest{UserID: userID})
	if err != nil {
		return fmt.Errorf("marshaling remove collaborator request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/%s/collaborator", BaseURL, tripKey)
	httpReq, err := http.NewRequest("DELETE", url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"userID":  userID,
	}).Debug("Removing collaborator from trip")

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
		"userID":  userID,
	}).Info("Successfully removed collaborator")

	return nil
}

// GetOrCreateShareKey creates or gets a share key with matching permissions
func (c *Client) GetOrCreateShareKey(editKey string, permissions ShareKeyPermissions) (*ShareKeyResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for creating share keys")
	}

	type shareKeyRequest struct {
		Permissions ShareKeyPermissions `json:"permissions"`
	}
	reqBody, err := json.Marshal(shareKeyRequest{Permissions: permissions})
	if err != nil {
		return nil, fmt.Errorf("marshaling share key request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/%s/shareKey", BaseURL, editKey)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"editKey":     editKey,
		"permissions": permissions,
	}).Debug("Creating/getting share key")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var shareKeyResp ShareKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&shareKeyResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.WithField("shareKey", shareKeyResp.ShareKey).Info("Successfully created/got share key")

	return &shareKeyResp, nil
}

// GetTripFlights retrieves all flights associated with a trip plan
func (c *Client) GetTripFlights(tripKey string) (*TripFlightsResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for getting trip flights")
	}

	c.logger.WithField("tripKey", tripKey).Debug("Getting trip flights")
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("getting trip: %w", err)
	}

	var flightsResp TripFlightsResponse
	flightsResp.Success = true
	for _, section := range trip.TripPlan.Itinerary.Sections {
		for _, block := range section.Blocks {
			if block.FlightInfo == nil {
				continue
			}

			flight := TripFlight{
				ID:            block.ID,
				FlightNumber:  strconv.Itoa(block.FlightInfo.Number),
				Airline:       block.FlightInfo.Airline.Name,
				AirlineIATA:   block.FlightInfo.Airline.Iata,
				DepartureTime: block.StartTime,
				ArrivalTime:   block.EndTime,
				Origin: FlightAirport{
					IATA: block.Depart.Airport.Iata,
					Name: block.Depart.Airport.Name,
					City: block.Depart.Airport.CityName,
				},
			}
			if block.Arrive != nil {
				flight.Destination = FlightAirport{
					IATA: block.Arrive.Airport.Iata,
					Name: block.Arrive.Airport.Name,
					City: block.Arrive.Airport.CityName,
				}
			}
			flightsResp.Data.Flights = append(flightsResp.Data.Flights, flight)
		}
	}

	c.logger.WithField("flightCount", len(flightsResp.Data.Flights)).Info("Successfully retrieved trip flights")

	return &flightsResp, nil
}

// ExportTrip exports a trip plan to Google Maps
func (c *Client) ExportTrip(tripKey string) (*ExportTripResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for exporting trips")
	}

	url := fmt.Sprintf("%s/tripPlans/%s/export/v2", BaseURL, tripKey)
	httpReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithField("tripKey", tripKey).Debug("Exporting trip")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"status":  resp.StatusCode,
		"body":    string(respBody),
	}).Debug("ExportTrip API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var exportResp ExportTripResponse
	if err := json.Unmarshal(respBody, &exportResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.Info("Successfully exported trip")

	return &exportResp, nil
}

// AutofillDay fills a day section with place suggestions
func (c *Client) AutofillDay(tripKey string, sectionID int, query string) (*AutofillDayResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for autofilling days")
	}

	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("getting trip: %w", err)
	}
	geoID := 0
	if len(trip.Resources.Geos) > 0 {
		geoID = trip.Resources.Geos[0].ID
	}
	if geoID == 0 {
		return nil, fmt.Errorf("trip has no geo id")
	}
	sectionDate := ""
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.ID == sectionID && section.Date != nil {
			sectionDate = *section.Date
			break
		}
	}
	if sectionDate == "" {
		return nil, fmt.Errorf("section %d has no date", sectionID)
	}

	reqBody, err := json.Marshal(AutofillDayRequest{
		TripPlanKey: tripKey,
		TripPlanID:  trip.TripPlan.ID,
		SectionID:   sectionID,
		SectionDate: sectionDate,
		GeoID:       geoID,
		Query:       query,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling autofill request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/autofillDay", BaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"query":     query,
	}).Debug("Autofilling day with suggestions")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"status":    resp.StatusCode,
		"body":      string(respBody),
	}).Debug("AutofillDay API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var autofillResp AutofillDayResponse
	if err := json.Unmarshal(respBody, &autofillResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !autofillResp.Success {
		return nil, fmt.Errorf("failed to autofill day - response: %s", string(respBody))
	}

	c.logger.WithField("suggestionCount", len(autofillResp.Data.Suggestions)).Info("Successfully autofilled day")

	return &autofillResp, nil
}

// AddChecklistItems adds items to a checklist section in a trip
func (c *Client) AddChecklistItems(tripKey string, sectionID int, items []ChecklistItem) (*ChecklistSectionResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for adding checklist items")
	}

	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("getting trip: %w", err)
	}

	itemTexts := make([]string, 0, len(items))
	for _, item := range items {
		itemTexts = append(itemTexts, item.Text)
	}
	reqBody, err := json.Marshal(struct {
		TripPlanID int      `json:"tripPlanId"`
		Items      []string `json:"items"`
	}{
		TripPlanID: trip.TripPlan.ID,
		Items:      itemTexts,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling checklist request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/checklistSection", BaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"itemCount": len(items),
	}).Debug("Adding checklist items")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"status":    resp.StatusCode,
		"body":      string(respBody),
	}).Debug("AddChecklistItems API response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var checklistResp ChecklistSectionResponse
	if err := json.Unmarshal(respBody, &checklistResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !checklistResp.Success {
		return nil, fmt.Errorf("failed to add checklist items - response: %s", string(respBody))
	}

	c.logger.WithField("itemCount", len(checklistResp.Data.Section.Items)).Info("Successfully added checklist items")

	return &checklistResp, nil
}

// ToggleChecklistItem toggles a checklist item's checked state
func (c *Client) ToggleChecklistItem(tripKey string, sectionID, itemID int, checked bool) (*ChecklistSectionResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for toggling checklist items")
	}

	reqBody, err := json.Marshal(ChecklistSectionRequest{
		Action:    "toggleItem",
		SectionID: sectionID,
		ItemID:    itemID,
		Checked:   checked,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling checklist request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/checklistSection", BaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("adding auth headers: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"itemID":    itemID,
		"checked":   checked,
	}).Debug("Toggling checklist item")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var checklistResp ChecklistSectionResponse
	if err := json.Unmarshal(respBody, &checklistResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.Info("Successfully toggled checklist item")

	return &checklistResp, nil
}
