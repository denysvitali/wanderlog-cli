package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

type (
	FeedHomeResponse     = models.FeedHomeResponse
	FeedResponse         = models.FeedResponse
	FeedRecentResponse   = models.FeedRecentResponse
	FriendsPlansResponse = models.FriendsPlansResponse
	TripHistoryResponse  = models.TripHistoryResponse
	GetIfEditedRequest   = models.GetIfEditedRequest
	GetIfEditedResponse  = models.GetIfEditedResponse
	EditCheck            = models.EditCheck
	BrowseGuidesResponse = models.BrowseGuidesResponse
)

// GetFeedHome fetches the authenticated user's home feed.
func (c *Client) GetFeedHome() (*FeedHomeResponse, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/home", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result FeedHomeResponse
	if err := decodeAPIBody("GetFeedHome", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFeed fetches the legacy trip feed.
func (c *Client) GetFeed() (*FeedResponse, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/feed", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result FeedResponse
	if err := decodeAPIBody("GetFeed", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFeedV2 fetches the v2 trip feed.
func (c *Client) GetFeedV2() (*FeedResponse, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/feed/v2", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result FeedResponse
	if err := decodeAPIBody("GetFeedV2", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFeedMostRecent returns the user's most recently edited trip.
func (c *Client) GetFeedMostRecent() (*FeedRecentResponse, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/feed/mostRecentlyEdited", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result FeedRecentResponse
	if err := decodeAPIBody("GetFeedMostRecent", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFriendsPlans fetches trip plans published by the user's friends.
func (c *Client) GetFriendsPlans() (*FriendsPlansResponse, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/friendsPlans", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result FriendsPlansResponse
	if err := decodeAPIBody("GetFriendsPlans", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripHistory returns the paginated trip edit history.
func (c *Client) GetTripHistory(offset int) (*TripHistoryResponse, error) {
	query := url.Values{}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/history", query, nil, false)
	if err != nil {
		return nil, err
	}
	var result TripHistoryResponse
	if err := decodeAPIBody("GetTripHistory", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetIfEdited asks the server which of the given trip plans changed since the
// provided revisions. Useful for cache invalidation.
func (c *Client) GetIfEdited(req GetIfEditedRequest) (*GetIfEditedResponse, error) {
	if req.ClientSchemaVersion == 0 {
		if v, err := clientSchemaVersionInt(); err == nil {
			req.ClientSchemaVersion = v
		}
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("GetIfEdited: marshaling request: %w", err)
	}
	apiResp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/getIfEdited", nil, body, false)
	if err != nil {
		return nil, err
	}
	var result GetIfEditedResponse
	if err := decodeAPIBody("GetIfEdited", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BrowseGuides returns curated travel guides. When geoID is non-zero the guides
// are scoped to that geography.
func (c *Client) BrowseGuides(geoID int) (*BrowseGuidesResponse, error) {
	var apiRespBody []byte
	var statusCode int
	if geoID > 0 {
		apiResp, err := c.apiRequest(context.Background(), http.MethodGet, fmt.Sprintf("tripPlans/browse/guides/%d", geoID), nil, nil, false)
		if err != nil {
			return nil, err
		}
		apiRespBody = apiResp.Body
		statusCode = apiResp.StatusCode
	} else {
		apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/browse/guides", nil, nil, false)
		if err != nil {
			return nil, err
		}
		apiRespBody = apiResp.Body
		statusCode = apiResp.StatusCode
	}
	var result BrowseGuidesResponse
	if err := decodeAPIBody("BrowseGuides", statusCode, apiRespBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GeoIDName represents a geo entry with ID and name, returned by various geo list endpoints.
type GeoIDName struct {
	ID     int       `json:"id"`
	Name   string    `json:"name"`
	Bounds []float64 `json:"bounds"`
}

// GeoSearchResult holds combined country and city geo entries.
type GeoSearchResult struct {
	Countries []GeoIDName
	Cities    []GeoIDName
}

// geoListResponse is the JSON response envelope for geo list endpoints.
type geoListResponse struct {
	Data    []GeoIDName `json:"data"`
	Success bool        `json:"success"`
}

// SearchGeos returns all geographic destinations (countries and cities) from Wanderlog.
// The caller can client-side filter by name since the full list is relatively small.
func (c *Client) SearchGeos() (*GeoSearchResult, error) {
	// Fetch countries
	countriesResp, err := c.apiRequest(context.Background(), http.MethodGet, "geo/countries", nil, nil, false)
	if err != nil {
		return nil, fmt.Errorf("SearchGeos (countries): %w", err)
	}
	countriesBody := countriesResp.Body

	// Fetch cities
	citiesResp, err := c.apiRequest(context.Background(), http.MethodGet, "geo/listGeosWithSearchedCategories", nil, nil, false)
	if err != nil {
		return nil, fmt.Errorf("SearchGeos (cities): %w", err)
	}
	citiesBody := citiesResp.Body

	var parsedCountries geoListResponse
	if err := json.Unmarshal(countriesBody, &parsedCountries); err != nil {
		return nil, fmt.Errorf("SearchGeos: parsing countries: %w", err)
	}
	var parsedCities geoListResponse
	if err := json.Unmarshal(citiesBody, &parsedCities); err != nil {
		return nil, fmt.Errorf("SearchGeos: parsing cities: %w", err)
	}

	return &GeoSearchResult{Countries: parsedCountries.Data, Cities: parsedCities.Data}, nil
}

func clientSchemaVersionInt() (int, error) {
	v := 0
	if _, err := fmt.Sscanf(ClientVersion, "%d", &v); err != nil {
		return 0, err
	}
	return v, nil
}
