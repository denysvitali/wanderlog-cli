package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
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
	if err := validateProspectiveDates(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	c.logger.WithFields(map[string]interface{}{
		"title":  req.Title,
		"geoIDs": req.GeoIDs,
		"type":   req.Type,
	}).Debug("Creating trip")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans", nil, jsonData, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(resp.Body),
	}).Debug("CreateTrip API response")

	var createResp struct {
		Success  bool                   `json:"success"`
		TripPlan models.TripPlanSummary `json:"tripPlan"`
		Data     models.TripPlanSummary `json:"data"`
	}
	if err := decodeMutationBody("CreateTrip", resp.StatusCode, resp.Body, &createResp); err != nil {
		return nil, err
	}

	tripPlan := createResp.TripPlan
	if tripPlan.Key == "" {
		tripPlan = createResp.Data
	}
	if tripPlan.Key == "" {
		return nil, fmt.Errorf("CreateTrip: successful response is missing trip key")
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

	c.logger.Debug("Creating example trip")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/createExampleTripPlan", nil, nil, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(resp.Body),
	}).Debug("CreateExampleTrip API response")

	// The createExampleTripPlan response uses "data" with viewKey (like CopyTripResponse)
	var exampleResp models.CopyTripResponse
	if err := decodeMutationBody("CreateExampleTrip", resp.StatusCode, resp.Body, &exampleResp); err != nil {
		return nil, err
	}
	key := exampleResp.Data.Key
	if key == "" {
		key = exampleResp.Data.ViewKey
	}
	if key == "" {
		return nil, fmt.Errorf("CreateExampleTrip: successful response is missing trip key")
	}

	// Convert to CreateTripResponse format
	createResp := &CreateTripResponse{
		Success: exampleResp.Success,
		TripPlan: models.TripPlanSummary{
			ID:    exampleResp.Data.ID,
			Key:   key,
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

	c.logger.WithField("tripKey", tripKey).Debug("Deleting trip")

	resp, err := c.apiRequest(context.Background(), http.MethodDelete, "tripPlans/"+url.PathEscape(tripKey), nil, nil, true)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOptionalMutationBody("DeleteTrip", resp.StatusCode, resp.Body); err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully deleted trip")
	return nil
}

// UpdateTrip updates trip metadata (title, dates, privacy) using ShareDB operations
func (c *Client) UpdateTrip(tripKey string, req UpdateTripRequest) error {
	return c.UpdateTripContext(context.Background(), tripKey, req)
}

// UpdateTripContext updates trip metadata and binds all fetch/apply attempts to
// ctx. A ShareDB conflict triggers a fresh fetch and operation rebuild.
func (c *Client) UpdateTripContext(ctx context.Context, tripKey string, req UpdateTripRequest) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for updating trips")
	}

	err := c.retryJSON0MutationContext(ctx, tripKey, "UpdateTrip", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripContext(ctx, tripKey)
		if err != nil {
			return nil, fmt.Errorf("getting current trip: %w", err)
		}
		return buildUpdateTripOperations(trip, req)
	})
	if err != nil {
		return fmt.Errorf("applying operations: %w", err)
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully updated trip")
	return nil
}

func buildUpdateTripOperations(trip *TripResponse, req UpdateTripRequest) ([]Operation, error) {
	nextStartDate := trip.TripPlan.StartDate
	if req.StartDate != "" {
		nextStartDate = req.StartDate
	}
	nextEndDate := trip.TripPlan.EndDate
	if req.EndDate != "" {
		nextEndDate = req.EndDate
	}
	if req.StartDate != "" || req.EndDate != "" {
		if err := validateProspectiveDates(nextStartDate, nextEndDate); err != nil {
			return nil, err
		}
	}

	// Build operations to replace fields
	ops := []models.Operation{}

	if (req.Title != "" || req.ClearTitle) && req.Title != trip.TripPlan.Title {
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
	if nextStartDate != "" && nextEndDate != "" {
		days, err := tripDays(nextStartDate, nextEndDate)
		if err != nil {
			return nil, err
		}
		if days != trip.TripPlan.Days {
			ops = append(ops, models.ReplaceInObject(
				[]interface{}{"days"},
				trip.TripPlan.Days,
				days,
			))
		}
	}

	if req.Privacy != "" && req.Privacy != trip.TripPlan.Privacy {
		ops = append(ops, models.ReplaceInObject(
			[]interface{}{"privacy"},
			trip.TripPlan.Privacy,
			req.Privacy,
		))
	}

	if len(ops) == 0 {
		return nil, nil
	}
	return ops, nil
}

func tripDays(startDate, endDate string) (int, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return 0, err
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days < 1 {
		return 0, fmt.Errorf("end date must be on or after start date")
	}
	return days, nil
}

func validateProspectiveDates(startDate, endDate string) error {
	if startDate != "" {
		if _, err := time.Parse(apiDateFormat, startDate); err != nil {
			return fmt.Errorf("invalid start date %q: expected YYYY-MM-DD", startDate)
		}
	}
	if endDate != "" {
		if _, err := time.Parse(apiDateFormat, endDate); err != nil {
			return fmt.Errorf("invalid end date %q: expected YYYY-MM-DD", endDate)
		}
	}
	if startDate != "" && endDate != "" {
		if _, err := tripDays(startDate, endDate); err != nil {
			return err
		}
	}
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
	if err := ValidateVisitTime(req.StartTime); err != nil {
		return fmt.Errorf("start_time: %w", err)
	}
	if err := ValidateVisitTime(req.EndTime); err != nil {
		return fmt.Errorf("end_time: %w", err)
	}
	return nil
}

// ValidateVisitTime validates the HH:MM visit time format used by itinerary blocks.
func ValidateVisitTime(value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("15:04", value); err != nil {
		return fmt.Errorf("must use HH:MM 24-hour format")
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
	if req.Place.FormattedAddress != "" {
		place["formatted_address"] = req.Place.FormattedAddress
	}
	if req.Place.URL != "" {
		place["url"] = req.Place.URL
	}
	if req.Place.Website != "" {
		place["website"] = req.Place.Website
	}
	if req.Place.InternationalPhoneNumber != "" {
		place["international_phone_number"] = req.Place.InternationalPhoneNumber
	}
	if len(req.Place.Types) > 0 {
		place["types"] = req.Place.Types
	}
	if req.Place.BusinessStatus != "" {
		place["business_status"] = req.Place.BusinessStatus
	}
	if req.Text != "" {
		place["text"] = quillTextForString(req.Text)
		place["placeNotes"] = req.Text
	}
	if req.StartTime != "" {
		place["startTime"] = req.StartTime
	}
	if req.EndTime != "" {
		place["endTime"] = req.EndTime
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
	for _, key := range []string{
		"formatted_address",
		"url",
		"website",
		"international_phone_number",
		"types",
		"business_status",
	} {
		if value, ok := place[key]; ok {
			place["place"].(map[string]any)[key] = value
		}
	}

	addDuplicates := false
	reqBody, err := json.Marshal(map[string]any{
		"places":        []map[string]any{place},
		"addDuplicates": addDuplicates,
	})
	if err != nil {
		return fmt.Errorf("marshaling add place request: %w", err)
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
		resp, err := c.apiRequest(context.Background(), http.MethodPost, fmt.Sprintf("tripPlans/%s/sections/%d/places", url.PathEscape(tripKey), sectionID), nil, reqBody, true)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode
		respBody = resp.Body
	} else {
		resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/sections/places", nil, reqBody, true)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode
		respBody = resp.Body
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":      tripKey,
		"sectionID":    sectionID,
		"placeName":    req.Place.Name,
		"statusCode":   statusCode,
		"responseBody": string(respBody),
	}).Debug("AddPlace API response")

	if err := decodeMutationBody("AddPlace", statusCode, respBody, nil); err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"placeName": req.Place.Name,
	}).Info("Successfully added place to trip")

	return nil
}

func quillTextForString(text string) map[string]any {
	if text == "" {
		return map[string]any{"ops": []any{map[string]any{"insert": "\n"}}}
	}
	if text[len(text)-1] != '\n' {
		text += "\n"
	}
	return map[string]any{"ops": []any{map[string]any{"insert": text}}}
}
