package wanderlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

// Type aliases for backward compatibility
type (
	CreateTripRequest   = models.CreateTripRequest
	CreateTripResponse  = models.CreateTripResponse
	UpdateTripRequest   = models.UpdateTripRequest
	AddPlaceRequest     = models.AddPlaceRequest
	AddPlaceInfo        = models.AddPlaceInfo
	OperationRequest    = models.OperationRequest
	Operation           = models.Operation
	SendInvitesRequest  = models.SendInvitesRequest
	TripInvite          = models.TripInvite
	LikeCount           = models.LikeCount
	ShareKeyPermissions = models.ShareKeyPermissions
	ShareKeyResponse    = models.ShareKeyResponse
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
		nil, // old value not needed for replace
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
		nil, // old value not needed
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
			nil, // old value not needed for replace
			[]interface{}{},
		))
	}

	// Also clear place metadata
	operations = append(operations, ReplaceInObject(
		[]interface{}{"resources", "placeMetadata"},
		nil, // old value not needed for replace
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
	url := fmt.Sprintf("%s/tripPlans/%s/likeCount", BaseURL, tripKey)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

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

	var likeCount LikeCount
	if err := json.NewDecoder(resp.Body).Decode(&likeCount); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &likeCount, nil
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
