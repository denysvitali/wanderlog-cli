package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

type DirectionsResponse = models.DirectionsResponse

// GetAllDistanceInfoForPlace fetches all-mode (driving, walking, transit, ...)
// distance info from a base place to a list of other places. The bundle sends
// the request body as a JSON-stringified `data` query parameter.
func (c *Client) GetAllDistanceInfoForPlace(payload any) (*DirectionsResponse, error) {
	if payload == nil {
		return nil, fmt.Errorf("GetAllDistanceInfoForPlace: payload is required")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("GetAllDistanceInfoForPlace: marshaling payload: %w", err)
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "directions/allDistanceInfoForPlace/v2", apiQuery(map[string]string{
		"data": string(encoded),
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result DirectionsResponse
	if err := decodeAPIBody("GetAllDistanceInfoForPlace", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDistancesForMode fetches pairwise distances for a fixed travel mode. The
// payload is opaque (driving, walking, etc plus a list of places).
func (c *Client) GetDistancesForMode(payload any) (*DirectionsResponse, error) {
	if payload == nil {
		return nil, fmt.Errorf("GetDistancesForMode: payload is required")
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "directions/distancesForMode", nil, payload, false)
	if err != nil {
		return nil, err
	}
	var result DirectionsResponse
	if err := decodeAPIBody("GetDistancesForMode", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OptimizeRoute returns an optimized ordering for a list of places.
func (c *Client) OptimizeRoute(payload any) (*DirectionsResponse, error) {
	if payload == nil {
		return nil, fmt.Errorf("OptimizeRoute: payload is required")
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "directions/optimizeRoute", nil, payload, false)
	if err != nil {
		return nil, err
	}
	var result DirectionsResponse
	if err := decodeAPIBody("OptimizeRoute", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
