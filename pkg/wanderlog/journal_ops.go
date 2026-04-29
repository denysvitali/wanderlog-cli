package wanderlog

import (
	"context"
	"fmt"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
)

type (
	JournalResponse             = models.JournalResponse
	JournalStopPolylinesRequest = models.JournalStopPolylinesRequest
	JournalStop                 = models.JournalStop
	JournalPolyline             = models.JournalPolyline
	JournalPolylinesResponse    = models.JournalPolylinesResponse
	UpdateRequiredResponse      = models.UpdateRequiredResponse
	DistinctionResponse         = models.DistinctionResponse
	CreateGuideResponse         = models.CreateGuideResponse
)

// GetViewOnlyJournal fetches a published journal by its share key.
func (c *Client) GetViewOnlyJournal(journalKey string) (*JournalResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.GetViewOnlyJournalWithResponse(context.Background(), journalKey)
	if err != nil {
		return nil, err
	}
	var result JournalResponse
	if err := decodeOpenAPIBody("GetViewOnlyJournal", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetJournalStopPolylines computes polylines between journal stops.
func (c *Client) GetJournalStopPolylines(req JournalStopPolylinesRequest) (*JournalPolylinesResponse, error) {
	if err := c.requireAuth("GetJournalStopPolylines"); err != nil {
		return nil, err
	}
	editor, err := jsonBodyEditor(req)
	if err != nil {
		return nil, fmt.Errorf("GetJournalStopPolylines: marshaling request: %w", err)
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.FetchJournalStopPolylinesWithResponse(context.Background(), editor)
	if err != nil {
		return nil, err
	}
	var result JournalPolylinesResponse
	if err := decodeOpenAPIBody("GetJournalStopPolylines", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripExpensesCSV downloads a trip's expense report as CSV bytes.
func (c *Client) GetTripExpensesCSV(tripKey string) ([]byte, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.GetTripExpensesCSVWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, err
	}
	if apiResp.StatusCode() < 200 || apiResp.StatusCode() >= 300 {
		return apiResp.Body, fmt.Errorf("GetTripExpensesCSV: HTTP %d: %s", apiResp.StatusCode(), truncateForLog(string(apiResp.Body), 500))
	}
	return apiResp.Body, nil
}

// RegisterTripView bumps the view counter on a shared trip.
func (c *Client) RegisterTripView(tripKey string) error {
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	apiResp, err := api.RegisterTripPlanViewWithResponse(context.Background(), tripKey)
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("RegisterTripView", apiResp.StatusCode(), apiResp.Body, nil)
}

// GetTripUpdateRequired tells clients whether they need to upgrade for this
// trip's schema version.
func (c *Client) GetTripUpdateRequired(tripKey string) (*UpdateRequiredResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	version, err := clientSchemaVersionInt()
	if err != nil {
		return nil, err
	}
	openAPIVersion := version
	apiResp, err := api.CheckIfUpdateRequiredWithResponse(context.Background(), tripKey, &openapi.CheckIfUpdateRequiredParams{
		ClientSchemaVersion: &openAPIVersion,
	})
	if err != nil {
		return nil, err
	}
	var result UpdateRequiredResponse
	if err := decodeOpenAPIBody("GetTripUpdateRequired", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripDistinction returns the trip's distinction/badge, if any.
func (c *Client) GetTripDistinction(tripKey string) (*DistinctionResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.GetTripPlanDistinctionWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, err
	}
	var result DistinctionResponse
	if err := decodeOpenAPIBody("GetTripDistinction", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetTripDistinction updates the trip's distinction.
func (c *Client) SetTripDistinction(tripKey, distinction string) error {
	if err := c.requireAuth("SetTripDistinction"); err != nil {
		return err
	}
	body := struct {
		Distinction string `json:"distinction"`
	}{Distinction: distinction}
	editor, err := jsonBodyEditor(body)
	if err != nil {
		return fmt.Errorf("SetTripDistinction: marshaling request: %w", err)
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	resp, err := api.SetTripPlanDistinctionWithResponse(context.Background(), tripKey, editor)
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("SetTripDistinction", resp.StatusCode(), resp.Body, nil)
}

// UpdateTripPlanGeo changes the primary destination geo for a trip.
func (c *Client) UpdateTripPlanGeo(tripKey string, geoID int) error {
	if err := c.requireAuth("UpdateTripPlanGeo"); err != nil {
		return err
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	resp, err := api.UpdateTripPlanGeoWithResponse(context.Background(), tripKey, geoID)
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("UpdateTripPlanGeo", resp.StatusCode(), resp.Body, nil)
}

// CreateGuideFromTripPlan promotes a trip plan into a published guide.
func (c *Client) CreateGuideFromTripPlan(tripKey string) (*CreateGuideResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.CreateGuideFromTripPlanWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, err
	}
	var result CreateGuideResponse
	if err := decodeOpenAPIBody("CreateGuideFromTripPlan", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
