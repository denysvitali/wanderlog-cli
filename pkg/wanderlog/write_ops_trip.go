package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
)

// Type aliases for backward compatibility
type (
	CreateTripRequest  = models.CreateTripRequest
	CreateTripResponse = models.CreateTripResponse
	UpdateTripRequest  = models.UpdateTripRequest
	AddPlaceRequest    = models.AddPlaceRequest
	AddPlaceInfo       = models.AddPlaceInfo
	OperationRequest   = models.OperationRequest
	Operation          = models.Operation
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
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithFields(map[string]interface{}{
		"title":  req.Title,
		"geoIDs": req.GeoIDs,
		"type":   req.Type,
	}).Debug("Creating trip")

	resp, err := api.CreateTripPlanWithBodyWithResponse(context.Background(), "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode(),
		"body":   string(resp.Body),
	}).Debug("CreateTrip API response")

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("CreateTrip: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var createResp struct {
		Success  bool                   `json:"success"`
		TripPlan models.TripPlanSummary `json:"tripPlan"`
		Data     models.TripPlanSummary `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &createResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !createResp.Success {
		c.logger.WithField("response", string(resp.Body)).Error("Trip creation failed")
		return nil, fmt.Errorf("failed to create trip - response: %s", string(resp.Body))
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

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.Debug("Creating example trip")

	resp, err := api.CreateExampleTripPlanWithResponse(context.Background())
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode(),
		"body":   string(resp.Body),
	}).Debug("CreateExampleTrip API response")

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("CreateExampleTrip: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	// The createExampleTripPlan response uses "data" with viewKey (like CopyTripResponse)
	var exampleResp models.CopyTripResponse
	if err := json.Unmarshal(resp.Body, &exampleResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !exampleResp.Success {
		c.logger.WithField("response", string(resp.Body)).Error("Example trip creation failed")
		return nil, fmt.Errorf("failed to create example trip - response: %s", string(resp.Body))
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

	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Debug("Deleting trip")

	resp, err := api.DeleteTripPlanWithResponse(context.Background(), tripKey)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOpenAPIBody("DeleteTrip", resp.StatusCode(), resp.Body, nil); err != nil {
		return err
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

	place := map[string]any{
		"name": req.Place.Name,
	}
	if req.Place.PlaceID != "" {
		place["place_id"] = req.Place.PlaceID
		place["placeId"] = req.Place.PlaceID
	}
	if req.Place.Geometry != nil {
		place["geometry"] = req.Place.Geometry
	}
	if req.Text != "" {
		place["text"] = req.Text
	}
	// The native add-places endpoint expects autocomplete/detail-shaped rows.
	// Keep the legacy flat fields, but also include the nested shape used by
	// downstream server code that reads row.place.place_id.
	place["place"] = map[string]any{
		"name":     req.Place.Name,
		"place_id": req.Place.PlaceID,
		"placeId":  req.Place.PlaceID,
		"geometry": req.Place.Geometry,
	}

	addDuplicates := false
	reqBody, err := json.Marshal(openapi.AddPlacesRequest{
		Places:        []map[string]any{place},
		AddDuplicates: &addDuplicates,
	})
	if err != nil {
		return fmt.Errorf("marshaling add place request: %w", err)
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":     tripKey,
		"sectionID":   sectionID,
		"requestBody": string(reqBody),
	}).Debug("AddPlace request details")

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeName": req.Place.Name,
	}).Debug("Adding place to trip")

	var statusCode int
	var respBody []byte
	if sectionID > 0 {
		resp, err := api.AddPlacesToTripPlanWithBodyWithResponse(context.Background(), tripKey, sectionID, "application/json", bytes.NewReader(reqBody))
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode()
		respBody = resp.Body
	} else {
		resp, err := api.AddPlacesToTripPlanWithoutSectionWithBodyWithResponse(context.Background(), tripKey, "application/json", bytes.NewReader(reqBody))
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode()
		respBody = resp.Body
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":      tripKey,
		"sectionID":    sectionID,
		"placeName":    req.Place.Name,
		"statusCode":   statusCode,
		"responseBody": string(respBody),
	}).Debug("AddPlace API response")

	if statusCode != http.StatusOK {
		bodyText := string(respBody)
		if msg, ok := knownWanderlogServerError("AddPlace", bodyText); ok {
			return fmt.Errorf("AddPlace: HTTP %d: %s", statusCode, msg)
		}
		return fmt.Errorf("AddPlace: HTTP %d: %s", statusCode, truncateForLog(bodyText, 500))
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
