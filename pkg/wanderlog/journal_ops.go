package wanderlog

import (
	"fmt"
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
	path := fmt.Sprintf("/tripPlans/viewOnlyJournal/%s", url.PathEscape(journalKey))
	var resp JournalResponse
	if err := c.doJSON("GET", path, nil, &resp, false, "GetViewOnlyJournal"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetJournalStopPolylines computes polylines between journal stops.
func (c *Client) GetJournalStopPolylines(req JournalStopPolylinesRequest) (*JournalPolylinesResponse, error) {
	var resp JournalPolylinesResponse
	if err := c.doJSON("POST", "/tripPlans/journalStopPolylines", req, &resp, true, "GetJournalStopPolylines"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTripExpensesCSV downloads a trip's expense report as CSV bytes.
func (c *Client) GetTripExpensesCSV(tripKey string) ([]byte, error) {
	path := fmt.Sprintf("/tripPlans/%s/expensesAsCSV", url.PathEscape(tripKey))
	return c.doRaw("GET", path, nil, true, "GetTripExpensesCSV")
}

// RegisterTripView bumps the view counter on a shared trip.
func (c *Client) RegisterTripView(tripKey string) error {
	path := fmt.Sprintf("/tripPlans/%s/registerView", url.PathEscape(tripKey))
	return c.doJSON("POST", path, nil, nil, false, "RegisterTripView")
}

// GetTripUpdateRequired tells clients whether they need to upgrade for this
// trip's schema version.
func (c *Client) GetTripUpdateRequired(tripKey string) (*UpdateRequiredResponse, error) {
	path := fmt.Sprintf("/tripPlans/%s/updateRequired?clientSchemaVersion=%s", url.PathEscape(tripKey), ClientVersion)
	var resp UpdateRequiredResponse
	if err := c.doJSON("GET", path, nil, &resp, false, "GetTripUpdateRequired"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTripDistinction returns the trip's distinction/badge, if any.
func (c *Client) GetTripDistinction(tripKey string) (*DistinctionResponse, error) {
	path := fmt.Sprintf("/tripPlans/%s/distinction", url.PathEscape(tripKey))
	var resp DistinctionResponse
	if err := c.doJSON("GET", path, nil, &resp, false, "GetTripDistinction"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetTripDistinction updates the trip's distinction.
func (c *Client) SetTripDistinction(tripKey, distinction string) error {
	path := fmt.Sprintf("/tripPlans/%s/distinction", url.PathEscape(tripKey))
	body := struct {
		Distinction string `json:"distinction"`
	}{Distinction: distinction}
	return c.doJSON("POST", path, body, nil, true, "SetTripDistinction")
}

// CreateGuideFromTripPlan promotes a trip plan into a published guide.
func (c *Client) CreateGuideFromTripPlan(tripKey string) (*CreateGuideResponse, error) {
	path := fmt.Sprintf("/tripPlans/%s/createGuideFromTripPlan", url.PathEscape(tripKey))
	var resp CreateGuideResponse
	if err := c.doJSON("POST", path, nil, &resp, true, "CreateGuideFromTripPlan"); err != nil {
		return nil, err
	}
	return &resp, nil
}
