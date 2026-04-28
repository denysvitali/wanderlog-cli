package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Type aliases for backward compatibility
type (
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

// CopyTrip creates a copy of an existing trip
func (c *Client) CopyTrip(sourceKey string) (*CreateTripResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for copying trips")
	}

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithField("sourceKey", sourceKey).Debug("Copying trip")

	resp, err := api.CopyTripPlanWithResponse(context.Background(), sourceKey)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode(),
		"body":   string(resp.Body),
	}).Debug("CopyTrip API response")

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("CopyTrip: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var copyResp models.CopyTripResponse
	if err := json.Unmarshal(resp.Body, &copyResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !copyResp.Success {
		c.logger.WithField("response", string(resp.Body)).Error("Copy trip failed")
		return nil, fmt.Errorf("failed to copy trip - response: %s", string(resp.Body))
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

	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Debug("Restoring trip")

	resp, err := api.RestoreTripPlanWithResponse(context.Background(), tripKey)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOpenAPIBody("RestoreTrip", resp.StatusCode(), resp.Body, nil); err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully restored trip")
	return nil
}

// SendTripInvites sends invites for people to edit a trip plan
func (c *Client) SendTripInvites(tripKey string, req SendInvitesRequest) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for sending invites")
	}

	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":  tripKey,
		"invitees": req.Invitees,
	}).Debug("Sending trip invites")

	invitees := make([]openapi_types.Email, 0, len(req.Invitees))
	for _, invitee := range req.Invitees {
		invitees = append(invitees, openapi_types.Email(invitee))
	}
	var message *string
	if req.Message != "" {
		message = &req.Message
	}
	resp, err := api.SendTripPlanInvitesWithResponse(context.Background(), tripKey, openapi.SendTripPlanInvitesJSONRequestBody{
		Invitees: &invitees,
		Message:  message,
	})
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOpenAPIBody("SendTripInvites", resp.StatusCode(), resp.Body, nil); err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully sent trip invites")
	return nil
}

// ListTripInvites lists all invites that have been sent out for a trip plan
func (c *Client) ListTripInvites(tripKey string) ([]map[string]interface{}, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for listing invites")
	}

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	resp, err := api.ListTripPlanInvitesWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("ListTripInvites: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, fmt.Errorf("ListTripInvites: unexpected response format")
	}

	return *resp.JSON200.Data, nil
}

// SetLike likes or unlikes a trip plan
func (c *Client) SetLike(tripKey string, liked bool) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for liking trips")
	}

	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"liked":   liked,
	}).Debug("Setting like status for trip")

	resp, err := api.SetLikeWithResponse(context.Background(), tripKey, openapi.SetLikeJSONRequestBody{Liked: &liked})
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOpenAPIBody("SetLike", resp.StatusCode(), resp.Body, nil); err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully set like status")
	return nil
}

// GetLikeCount gets whether we've liked a trip plan and the total number of likes
func (c *Client) GetLikeCount(tripKey string) (*LikeCount, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	resp, err := api.GetLikeCountWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("GetLikeCount: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var bulkResp struct {
		Success bool `json:"success"`
		Data    struct {
			Like      bool `json:"like"`
			LikeCount int  `json:"likeCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &bulkResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !bulkResp.Success {
		return nil, fmt.Errorf("failed to get like count")
	}

	return &LikeCount{
		Count:     bulkResp.Data.LikeCount,
		UserLiked: bulkResp.Data.Like,
	}, nil
}

// AddCollaborator adds a new collaborator to a trip plan with edit access
func (c *Client) AddCollaborator(tripKey string, userID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for adding collaborators")
	}

	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"userID":  userID,
	}).Debug("Adding collaborator to trip")

	resp, err := api.AddCollaboratorWithResponse(context.Background(), tripKey, openapi.AddCollaboratorJSONRequestBody{UserId: &userID})
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOpenAPIBody("AddCollaborator", resp.StatusCode(), resp.Body, nil); err != nil {
		return err
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

	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"userID":  userID,
	}).Debug("Removing collaborator from trip")

	resp, err := api.RemoveCollaboratorWithResponse(context.Background(), tripKey, openapi.RemoveCollaboratorJSONRequestBody{UserId: &userID})
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOpenAPIBody("RemoveCollaborator", resp.StatusCode(), resp.Body, nil); err != nil {
		return err
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

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithFields(map[string]interface{}{
		"editKey":     editKey,
		"permissions": permissions,
	}).Debug("Creating/getting share key")

	resp, err := api.GetOrCreateTripPlanKeyWithResponse(context.Background(), editKey, openapi.GetOrCreateTripPlanKeyJSONRequestBody{
		Permissions: &struct {
			CanEdit *bool `json:"canEdit,omitempty"`
			CanView *bool `json:"canView,omitempty"`
		}{
			CanEdit: &permissions.CanEdit,
			CanView: &permissions.CanView,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("GetOrCreateShareKey: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var shareKeyResp ShareKeyResponse
	if err := json.Unmarshal(resp.Body, &shareKeyResp); err != nil {
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
				SectionID:     section.ID,
				FlightNumber:  strconv.Itoa(block.FlightInfo.Number),
				Airline:       block.FlightInfo.Airline.Name,
				AirlineIATA:   block.FlightInfo.Airline.Iata,
				DepartureDate: block.Depart.Date,
				DepartureTime: block.StartTime,
				ArrivalTime:   block.EndTime,
				Origin: FlightAirport{
					IATA: block.Depart.Airport.Iata,
					Name: block.Depart.Airport.Name,
					City: block.Depart.Airport.CityName,
				},
			}
			if section.Date != nil {
				flight.SectionDate = *section.Date
			}
			if block.Arrive != nil {
				flight.ArrivalDate = block.Arrive.Date
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

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithField("tripKey", tripKey).Debug("Exporting trip")

	resp, err := api.ExportTripPlanWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"status":  resp.StatusCode(),
		"body":    string(resp.Body),
	}).Debug("ExportTrip API response")

	var exportResp ExportTripResponse
	if err := decodeOpenAPIBody("ExportTrip", resp.StatusCode(), resp.Body, &exportResp); err != nil {
		return nil, err
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

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"query":     query,
	}).Debug("Autofilling day with suggestions")

	resp, err := api.AutofillDayWithBodyWithResponse(context.Background(), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"status":    resp.StatusCode(),
		"body":      string(resp.Body),
	}).Debug("AutofillDay API response")

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("AutofillDay: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var autofillResp AutofillDayResponse
	if err := json.Unmarshal(resp.Body, &autofillResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !autofillResp.Success {
		return nil, fmt.Errorf("failed to autofill day - response: %s", string(resp.Body))
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

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"itemCount": len(items),
	}).Debug("Adding checklist items")

	resp, err := api.AddChecklistItemsWithBodyWithResponse(context.Background(), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"status":    resp.StatusCode(),
		"body":      string(resp.Body),
	}).Debug("AddChecklistItems API response")

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("AddChecklistItems: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var checklistResp ChecklistSectionResponse
	if err := json.Unmarshal(resp.Body, &checklistResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !checklistResp.Success {
		return nil, fmt.Errorf("failed to add checklist items - response: %s", string(resp.Body))
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

	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"itemID":    itemID,
		"checked":   checked,
	}).Debug("Toggling checklist item")

	resp, err := api.AddChecklistItemsWithBodyWithResponse(context.Background(), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("ToggleChecklistItem: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	var checklistResp ChecklistSectionResponse
	if err := json.Unmarshal(resp.Body, &checklistResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.Info("Successfully toggled checklist item")

	return &checklistResp, nil
}
