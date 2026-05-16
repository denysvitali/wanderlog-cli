package wanderlog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
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
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/viewOnlyJournal/"+url.PathEscape(journalKey), nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result JournalResponse
	if err := decodeAPIBody("GetViewOnlyJournal", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetJournalStopPolylines computes polylines between journal stops.
func (c *Client) GetJournalStopPolylines(req JournalStopPolylinesRequest) (*JournalPolylinesResponse, error) {
	if err := c.requireAuth("GetJournalStopPolylines"); err != nil {
		return nil, err
	}
	apiResp, err := c.apiJSON(context.Background(), http.MethodPost, "tripPlans/journalStopPolylines", nil, req, true)
	if err != nil {
		return nil, err
	}
	var result JournalPolylinesResponse
	if err := decodeAPIBody("GetJournalStopPolylines", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripExpensesCSV downloads a trip's expense report as CSV bytes.
func (c *Client) GetTripExpensesCSV(tripKey string) ([]byte, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/"+url.PathEscape(tripKey)+"/expensesAsCSV", nil, nil, false)
	if err != nil {
		return nil, err
	}
	if apiResp.StatusCode < 200 || apiResp.StatusCode >= 300 {
		return apiResp.Body, fmt.Errorf("GetTripExpensesCSV: HTTP %d: %s", apiResp.StatusCode, truncateForLog(string(apiResp.Body), 500))
	}
	return apiResp.Body, nil
}

// RegisterTripView bumps the view counter on a shared trip.
func (c *Client) RegisterTripView(tripKey string) error {
	apiResp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/registerView", nil, nil, false)
	if err != nil {
		return err
	}
	return decodeAPIBody("RegisterTripView", apiResp.StatusCode, apiResp.Body, nil)
}

// GetTripUpdateRequired tells clients whether they need to upgrade for this
// trip's schema version.
func (c *Client) GetTripUpdateRequired(tripKey string) (*UpdateRequiredResponse, error) {
	version, err := clientSchemaVersionInt()
	if err != nil {
		return nil, err
	}
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/"+url.PathEscape(tripKey)+"/updateRequired", apiQuery(map[string]string{
		"clientSchemaVersion": fmt.Sprintf("%d", version),
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result UpdateRequiredResponse
	if err := decodeAPIBody("GetTripUpdateRequired", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripDistinction returns the trip's distinction/badge, if any.
func (c *Client) GetTripDistinction(tripKey string) (*DistinctionResponse, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/"+url.PathEscape(tripKey)+"/distinction", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result DistinctionResponse
	if err := decodeAPIBody("GetTripDistinction", apiResp.StatusCode, apiResp.Body, &result); err != nil {
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
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/distinction", nil, body, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("SetTripDistinction", resp.StatusCode, resp.Body, nil)
}

// UpdateTripPlanGeo changes the primary destination geo for a trip.
func (c *Client) UpdateTripPlanGeo(tripKey string, geoID int) error {
	if err := c.requireAuth("UpdateTripPlanGeo"); err != nil {
		return err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodPost, fmt.Sprintf("tripPlans/%s/updateTripPlanGeo/%d", url.PathEscape(tripKey), geoID), nil, nil, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("UpdateTripPlanGeo", resp.StatusCode, resp.Body, nil)
}

// CreateGuideFromTripPlan promotes a trip plan into a published guide.
func (c *Client) CreateGuideFromTripPlan(tripKey string) (*CreateGuideResponse, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/createGuideFromTripPlan", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result CreateGuideResponse
	if err := decodeAPIBody("CreateGuideFromTripPlan", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
