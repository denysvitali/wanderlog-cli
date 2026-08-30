package wanderlog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GetTripRawContext retrieves a lossless trip document and binds the request to ctx.
func (c *Client) GetTripRawContext(ctx context.Context, key string) (map[string]any, error) {
	resp, err := c.apiRequest(ctx, http.MethodGet, "tripPlans/"+url.PathEscape(key), apiQuery(map[string]string{
		"clientSchemaVersion": ClientVersion,
	}), nil, false)
	if err != nil {
		return nil, err
	}

	var trip map[string]any
	if err := decodeAPIBodyPreserveNumbers("GetTrip", resp.StatusCode, resp.Body, &trip); err != nil {
		return nil, err
	}
	if msg, ok := trip["error"].(string); ok && msg != "" {
		return nil, fmt.Errorf("API error: %s", msg)
	}
	return trip, nil
}

// GetMeContext fetches the authenticated user's profile and binds the request to ctx.
func (c *Client) GetMeContext(ctx context.Context) (*UserProfile, error) {
	if err := c.requireAuth("GetMe"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(ctx, http.MethodGet, "user", nil, nil, true)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := decodeAPIBody("GetMe", resp.StatusCode, resp.Body, &profile); err != nil {
		return nil, fmt.Errorf("GetMe: decoding response: %w", err)
	}
	profile.Raw = resp.Body
	return &profile, nil
}

// GetPlaceDetailsContext fetches place details and binds the request to ctx.
func (c *Client) GetPlaceDetailsContext(ctx context.Context, placeID string) (*PlaceDetailsResponse, error) {
	resp, err := c.apiRequest(ctx, http.MethodGet, "placesAPI/getPlaceDetailsAndCardData", apiQuery(map[string]string{
		"placeId":  placeID,
		"language": "en",
	}), nil, false)
	if err != nil {
		return nil, err
	}

	var result PlaceDetailsResponse
	if err := decodeAPIBody("GetPlaceDetails", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("API request was not successful: %s", truncateForLog(string(resp.Body), 500))
	}
	return &result, nil
}
