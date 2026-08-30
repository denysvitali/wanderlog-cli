package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
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

	c.logger.WithField("sourceKey", sourceKey).Debug("Copying trip")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/copy/"+url.PathEscape(sourceKey), nil, nil, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(resp.Body),
	}).Debug("CopyTrip API response")

	var copyResp models.CopyTripResponse
	if err := decodeMutationBody("CopyTrip", resp.StatusCode, resp.Body, &copyResp); err != nil {
		return nil, err
	}
	key := copyResp.Data.Key
	if key == "" {
		key = copyResp.Data.ViewKey
	}
	if key == "" {
		return nil, fmt.Errorf("CopyTrip: successful response is missing trip key")
	}

	c.logger.WithFields(map[string]interface{}{
		"sourceKey": sourceKey,
		"newKey":    key,
		"title":     copyResp.Data.Title,
	}).Info("Successfully copied trip")

	// Convert to CreateTripResponse format for compatibility
	return &CreateTripResponse{
		Success: copyResp.Success,
		TripPlan: models.TripPlanSummary{
			ID:    copyResp.Data.ID,
			Key:   key,
			Title: copyResp.Data.Title,
		},
	}, nil
}

// RestoreTrip restores a soft-deleted trip plan
func (c *Client) RestoreTrip(tripKey string) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for restoring trips")
	}

	c.logger.WithField("tripKey", tripKey).Debug("Restoring trip")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/restore", nil, nil, true)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOptionalMutationBody("RestoreTrip", resp.StatusCode, resp.Body); err != nil {
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

	c.logger.WithFields(map[string]interface{}{
		"tripKey":  tripKey,
		"invitees": req.Invitees,
	}).Debug("Sending trip invites")

	body := map[string]any{"invitees": req.Invitees}
	if req.Message != "" {
		body["message"] = req.Message
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/invite", nil, body, true)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOptionalMutationBody("SendTripInvites", resp.StatusCode, resp.Body); err != nil {
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

	c.logger.WithField("tripKey", tripKey).Debug("Listing trip invites")

	// DoAPI strips leading / and api/ prefix, so use "tripPlans/%s/invites"
	statusCode, respBody, err := c.DoAPI("GET", fmt.Sprintf("tripPlans/%s/invites", tripKey), nil, nil, true)
	if err != nil {
		return nil, fmt.Errorf("ListTripInvites: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("ListTripInvites: HTTP %d: %s", statusCode, truncateForLog(string(respBody), 500))
	}

	// Parse response - API returns {success, data: [...]}
	var resp struct {
		Success bool                     `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("ListTripInvites: decoding response: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("ListTripInvites: API returned success=false")
	}

	return resp.Data, nil
}

// SetLike likes or unlikes a trip plan
func (c *Client) SetLike(tripKey string, liked bool) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for liking trips")
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"liked":   liked,
	}).Debug("Setting like status for trip")

	body, err := json.Marshal(map[string]bool{"liked": liked})
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	// DoAPI strips leading / and api/ prefix, so use "tripPlans/%s/like"
	statusCode, respBody, err := c.DoAPI("POST", fmt.Sprintf("tripPlans/%s/like", tripKey), body, nil, true)
	if err != nil {
		return fmt.Errorf("SetLike: %w", err)
	}

	if err := decodeOptionalMutationBody("SetLike", statusCode, respBody); err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully set like status")
	return nil
}

// GetLikeCount gets whether we've liked a trip plan and the total number of likes
func (c *Client) GetLikeCount(tripKey string) (*LikeCount, error) {
	c.logger.WithField("tripKey", tripKey).Debug("Getting like count for trip")

	// The /tripPlans/{key}/likeCount endpoint appears to not exist (returns HTML SPA page)
	// Instead, get the like count from the trip data itself
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("GetLikeCount: getting trip: %w", err)
	}

	return &LikeCount{
		Count:     trip.TripPlan.LikeCount,
		UserLiked: false, // UserLiked requires the /likeCount endpoint which doesn't exist
	}, nil
}

// AddCollaborator adds a new collaborator to a trip plan with edit access
func (c *Client) AddCollaborator(tripKey string, userID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for adding collaborators")
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"userID":  userID,
	}).Debug("Adding collaborator to trip")

	resp, err := c.apiJSON(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/collaborator", nil, map[string]any{"userId": userID}, true)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOptionalMutationBody("AddCollaborator", resp.StatusCode, resp.Body); err != nil {
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

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"userID":  userID,
	}).Debug("Removing collaborator from trip")

	resp, err := c.apiJSON(context.Background(), http.MethodDelete, "tripPlans/"+url.PathEscape(tripKey)+"/collaborator", nil, map[string]any{"userId": userID}, true)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if err := decodeOptionalMutationBody("RemoveCollaborator", resp.StatusCode, resp.Body); err != nil {
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

	c.logger.WithFields(map[string]interface{}{
		"editKey":     editKey,
		"permissions": permissions,
	}).Debug("Creating/getting share key")

	resp, err := c.apiJSON(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(editKey)+"/shareKey", nil, map[string]any{
		"permissions": map[string]bool{
			"canEdit": permissions.CanEdit,
			"canView": permissions.CanView,
		},
	}, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetOrCreateShareKey: HTTP %d: %s", resp.StatusCode, truncateForLog(string(resp.Body), 500))
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

	c.logger.WithField("tripKey", tripKey).Debug("Exporting trip")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/export/v2", nil, nil, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"status":  resp.StatusCode,
		"body":    string(resp.Body),
	}).Debug("ExportTrip API response")

	var exportResp ExportTripResponse
	if err := decodeMutationBody("ExportTrip", resp.StatusCode, resp.Body, &exportResp); err != nil {
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

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"query":     query,
	}).Debug("Autofilling day with suggestions")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/autofillDay", nil, reqBody, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"status":    resp.StatusCode,
		"body":      string(resp.Body),
	}).Debug("AutofillDay API response")

	var autofillResp AutofillDayResponse
	if err := decodeMutationBody("AutofillDay", resp.StatusCode, resp.Body, &autofillResp); err != nil {
		return nil, err
	}

	c.logger.WithField("suggestionCount", len(autofillResp.Data.Suggestions)).Info("Successfully autofilled day")

	return &autofillResp, nil
}

// AddChecklistItems adds items to a checklist section in a trip
func (c *Client) AddChecklistItems(tripKey string, sectionID int, items []ChecklistItem) (*ChecklistSectionResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for adding checklist items")
	}
	if sectionID <= 0 {
		return nil, fmt.Errorf("section ID must be greater than zero")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one checklist item is required")
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
		SectionID  int      `json:"sectionId"`
		Items      []string `json:"items"`
	}{
		TripPlanID: trip.TripPlan.ID,
		SectionID:  sectionID,
		Items:      itemTexts,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling checklist request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"itemCount": len(items),
	}).Debug("Adding checklist items")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/checklistSection", nil, reqBody, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"status":    resp.StatusCode,
		"body":      string(resp.Body),
	}).Debug("AddChecklistItems API response")

	var checklistResp ChecklistSectionResponse
	if err := decodeMutationBody("AddChecklistItems", resp.StatusCode, resp.Body, &checklistResp); err != nil {
		return nil, err
	}

	c.logger.WithField("itemCount", len(checklistResp.Data.Section.Items)).Info("Successfully added checklist items")

	return &checklistResp, nil
}

// ToggleChecklistItem toggles a checklist item's checked state
func (c *Client) ToggleChecklistItem(tripKey string, sectionID, itemID int, checked bool) (*ChecklistSectionResponse, error) {
	if c.auth == nil {
		return nil, fmt.Errorf("authentication required for toggling checklist items")
	}
	if sectionID <= 0 {
		return nil, fmt.Errorf("section ID must be greater than zero")
	}
	if itemID <= 0 {
		return nil, fmt.Errorf("item ID must be greater than zero")
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

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"itemID":    itemID,
		"checked":   checked,
	}).Debug("Toggling checklist item")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/checklistSection", nil, reqBody, true)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	var checklistResp ChecklistSectionResponse
	if err := decodeMutationBody("ToggleChecklistItem", resp.StatusCode, resp.Body, &checklistResp); err != nil {
		return nil, err
	}

	c.logger.Info("Successfully toggled checklist item")

	return &checklistResp, nil
}
