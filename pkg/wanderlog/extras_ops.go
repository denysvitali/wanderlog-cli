package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

type (
	LikesBulkRequest              = models.LikesBulkRequest
	LikesBulkResponse             = models.LikesBulkResponse
	CreateTripFromFlightsResponse = models.CreateTripFromFlightsResponse
	MyProfileResponse             = models.MyProfileResponse
	LodgingCheckoutDataResponse   = models.LodgingCheckoutDataResponse
	DealsResponse                 = models.DealsResponse
	RateEmailRequest              = models.RateEmailRequest
)

// GetTripLikesBulk asks the server which of the given trip keys the
// authenticated user has liked.
func (c *Client) GetTripLikesBulk(keys []string) (*LikesBulkResponse, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("GetTripLikesBulk: at least one trip key is required")
	}
	if err := c.requireAuth("GetTripLikesBulk"); err != nil {
		return nil, err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "tripPlans/likes", nil, LikesBulkRequest{Keys: keys}, true)
	if err != nil {
		return nil, err
	}
	var result LikesBulkResponse
	if err := decodeAPIBody("GetTripLikesBulk", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	result.Likes = parseTripLikesBulk(resp.Body, keys)
	if result.Likes == nil {
		c.logger.WithField("tripKeys", keys).Warn("Likes response did not contain a recognized state; leaving like state unknown")
	}
	return &result, nil
}

func parseTripLikesBulk(body []byte, requestedKeys []string) map[string]bool {
	var envelope any
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil
	}
	likes := make(map[string]bool)
	collectTripLikes(envelope, requestedKeys, likes)
	if len(likes) == 0 {
		return nil
	}
	return likes
}

func collectTripLikes(value any, requestedKeys []string, likes map[string]bool) {
	switch typed := value.(type) {
	case bool:
		if len(requestedKeys) == 1 {
			likes[requestedKeys[0]] = typed
		}
	case string:
		for _, key := range requestedKeys {
			if typed == key {
				likes[key] = true
				return
			}
		}
	case []any:
		if len(typed) == len(requestedKeys) {
			allBoolean := true
			for i, item := range typed {
				liked, ok := item.(bool)
				if !ok {
					allBoolean = false
					break
				}
				likes[requestedKeys[i]] = liked
			}
			if allBoolean {
				return
			}
			for _, key := range requestedKeys {
				delete(likes, key)
			}
		}
		allKeys := true
		for _, item := range typed {
			if _, ok := item.(string); !ok {
				allKeys = false
				break
			}
		}
		if allKeys {
			for _, key := range requestedKeys {
				likes[key] = false
			}
			for _, item := range typed {
				key := item.(string)
				for _, requested := range requestedKeys {
					if key == requested {
						likes[requested] = true
						break
					}
				}
			}
			return
		}
		for _, item := range typed {
			collectTripLikes(item, requestedKeys, likes)
		}
	case map[string]any:
		if key, ok := tripLikeObjectKey(typed); ok {
			if liked, ok := tripLikeObjectValue(typed); ok && requestedTripLikeKey(key, requestedKeys) {
				likes[key] = liked
				return
			}
		}
		for _, key := range requestedKeys {
			if candidate, ok := typed[key]; ok {
				if liked, ok := tripLikeBool(candidate); ok {
					likes[key] = liked
				}
			}
		}
		for _, field := range []string{"data", "likes", "tripPlanLikes", "likedTripPlans", "likedKeys"} {
			if nested, ok := typed[field]; ok {
				collectTripLikes(nested, requestedKeys, likes)
			}
		}
	}
}

func requestedTripLikeKey(key string, requestedKeys []string) bool {
	for _, requested := range requestedKeys {
		if key == requested {
			return true
		}
	}
	return false
}

func tripLikeObjectKey(value map[string]any) (string, bool) {
	for _, field := range []string{"key", "tripKey", "tripPlanKey", "id"} {
		if key, ok := value[field].(string); ok && key != "" {
			return key, true
		}
	}
	return "", false
}

func tripLikeObjectValue(value map[string]any) (bool, bool) {
	for _, field := range []string{"liked", "isLiked", "userLiked", "value"} {
		if candidate, ok := value[field]; ok {
			return tripLikeBool(candidate)
		}
	}
	return false, false
}

func tripLikeBool(value any) (bool, bool) {
	if liked, ok := value.(bool); ok {
		return liked, true
	}
	if object, ok := value.(map[string]any); ok {
		return tripLikeObjectValue(object)
	}
	return false, false
}

// CreateTripFromFlights seeds a new trip plan from a flights payload. The
// payload shape is opaque (matches the bundle's flight-import widget output).
func (c *Client) CreateTripFromFlights(payload any) (*CreateTripFromFlightsResponse, error) {
	if payload == nil {
		return nil, fmt.Errorf("CreateTripFromFlights: payload is required")
	}
	if err := c.requireAuth("CreateTripFromFlights"); err != nil {
		return nil, err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "tripPlans/flights", nil, payload, true)
	if err != nil {
		return nil, err
	}
	var result CreateTripFromFlightsResponse
	if err := decodeAPIBody("CreateTripFromFlights", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMyProfileData returns the authenticated user's profile dashboard
// (pinned trips, recent activity). This is distinct from GetUserTrips, which
// hits /api/tripPlans without the dashboard wrapper.
func (c *Client) GetMyProfileData() (*MyProfileResponse, error) {
	if err := c.requireAuth("GetMyProfileData"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/myProfile/", nil, nil, true)
	if err != nil {
		return nil, err
	}
	var result MyProfileResponse
	if err := decodeAPIBody("GetMyProfileData", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLodgingCheckoutData fetches pre-checkout pricing/policy data for a
// lodging offer. Pass the bundle's expected query params (lodgingPropertyId,
// dates, guests, currencyCode, ...).
func (c *Client) GetLodgingCheckoutData(params map[string]string) (*LodgingCheckoutDataResponse, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("GetLodgingCheckoutData: query parameters are required")
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "lodging/checkoutData", apiQuery(params), nil, false)
	if err != nil {
		return nil, err
	}
	var result LodgingCheckoutDataResponse
	if err := decodeAPIBody("GetLodgingCheckoutData", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDealsForUser returns user-targeted travel deals.
func (c *Client) GetDealsForUser() (*DealsResponse, error) {
	if err := c.requireAuth("GetDealsForUser"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "deals", nil, nil, true)
	if err != nil {
		return nil, err
	}
	var result DealsResponse
	if err := decodeAPIBody("GetDealsForUser", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RateEmail scores an auto-parsed forwarded email to improve the parser.
func (c *Client) RateEmail(req RateEmailRequest) error {
	if err := c.requireAuth("RateEmail"); err != nil {
		return err
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("RateEmail: marshaling request: %w", err)
	}
	resp, err := c.apiRequest(context.Background(), http.MethodPost, "emails/rate", nil, encoded, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("RateEmail", resp.StatusCode, resp.Body, nil)
}
