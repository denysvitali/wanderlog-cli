package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
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
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.ListHomeFeedTripPlansWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	var result FeedHomeResponse
	if err := decodeOpenAPIBody("GetFeedHome", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFeed fetches the legacy trip feed.
func (c *Client) GetFeed() (*FeedResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.GetFeedV1WithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	var result FeedResponse
	if err := decodeOpenAPIBody("GetFeed", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFeedV2 fetches the v2 trip feed.
func (c *Client) GetFeedV2() (*FeedResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.GetFeedV2WithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	var result FeedResponse
	if err := decodeOpenAPIBody("GetFeedV2", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFeedMostRecent returns the user's most recently edited trip.
func (c *Client) GetFeedMostRecent() (*FeedRecentResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.GetFeedMostRecentWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	var result FeedRecentResponse
	if err := decodeOpenAPIBody("GetFeedMostRecent", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFriendsPlans fetches trip plans published by the user's friends.
func (c *Client) GetFriendsPlans() (*FriendsPlansResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.ListFriendsTripPlansWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	var result FriendsPlansResponse
	if err := decodeOpenAPIBody("GetFriendsPlans", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripHistory returns the paginated trip edit history.
func (c *Client) GetTripHistory(offset int) (*TripHistoryResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	params := &openapi.ListTripPlansHistoryParams{}
	if offset > 0 {
		params.Offset = &offset
	}
	apiResp, err := api.ListTripPlansHistoryWithResponse(context.Background(), params)
	if err != nil {
		return nil, err
	}
	var result TripHistoryResponse
	if err := decodeOpenAPIBody("GetTripHistory", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
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
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("GetIfEdited: marshaling request: %w", err)
	}
	apiResp, err := api.GetTripPlansIfEditedWithBodyWithResponse(context.Background(), "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var result GetIfEditedResponse
	if err := decodeOpenAPIBody("GetIfEdited", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BrowseGuides returns curated travel guides. When geoID is non-zero the guides
// are scoped to that geography.
func (c *Client) BrowseGuides(geoID int) (*BrowseGuidesResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	var apiRespBody []byte
	var statusCode int
	if geoID > 0 {
		apiResp, err := api.ListGeoPageGoodGuides(context.Background(), geoID)
		if err != nil {
			return nil, err
		}
		defer apiResp.Body.Close()
		apiRespBody, err = io.ReadAll(apiResp.Body)
		if err != nil {
			return nil, fmt.Errorf("BrowseGuides: reading response: %w", err)
		}
		statusCode = apiResp.StatusCode
	} else {
		apiResp, err := api.ListBrowsePageGoodGuides(context.Background())
		if err != nil {
			return nil, err
		}
		defer apiResp.Body.Close()
		apiRespBody, err = io.ReadAll(apiResp.Body)
		if err != nil {
			return nil, fmt.Errorf("BrowseGuides: reading response: %w", err)
		}
		statusCode = apiResp.StatusCode
	}
	var result BrowseGuidesResponse
	if err := decodeOpenAPIBody("BrowseGuides", statusCode, apiRespBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// geoIDName represents a geo entry with ID and name, returned by various geo list endpoints.
type geoIDName struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SearchGeos returns all geographic destinations (countries and cities) from Wanderlog.
// The caller can client-side filter by name since the full list is relatively small.
func (c *Client) SearchGeos() (*BrowseGuidesResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	// Fetch countries
	countriesResp, err := api.ListCountries(context.Background())
	if err != nil {
		return nil, fmt.Errorf("SearchGeos (countries): %w", err)
	}
	defer countriesResp.Body.Close()
	countriesBody, err := io.ReadAll(countriesResp.Body)
	if err != nil {
		return nil, fmt.Errorf("SearchGeos: reading countries response: %w", err)
	}

	// Fetch cities
	citiesResp, err := api.ListGeosWithSearchedCategories(context.Background())
	if err != nil {
		return nil, fmt.Errorf("SearchGeos (cities): %w", err)
	}
	defer citiesResp.Body.Close()
	citiesBody, err := io.ReadAll(citiesResp.Body)
	if err != nil {
		return nil, fmt.Errorf("SearchGeos: reading cities response: %w", err)
	}

	var countries []geoIDName
	var cities []geoIDName

	if err := json.Unmarshal(countriesBody, &countries); err != nil {
		return nil, fmt.Errorf("SearchGeos: parsing countries: %w", err)
	}
	if err := json.Unmarshal(citiesBody, &cities); err != nil {
		return nil, fmt.Errorf("SearchGeos: parsing cities: %w", err)
	}

	// Transform to geoGuideCounts format and combine
	type geoGuideCount struct {
		Name       string `json:"name"`
		GeoID      int    `json:"geoId"`
		GuideCount int    `json:"guideCount"`
	}

	allGeos := make([]geoGuideCount, 0, len(countries)+len(cities))
	for _, c := range countries {
		allGeos = append(allGeos, geoGuideCount{Name: c.Name, GeoID: c.ID})
	}
	for _, c := range cities {
		allGeos = append(allGeos, geoGuideCount{Name: c.Name, GeoID: c.ID})
	}

	combinedData, err := json.Marshal(map[string][]geoGuideCount{"geoGuideCounts": allGeos})
	if err != nil {
		return nil, fmt.Errorf("SearchGeos: marshaling combined geos: %w", err)
	}

	return &BrowseGuidesResponse{
		Success: true,
		Data:    combinedData,
	}, nil
}

func clientSchemaVersionInt() (int, error) {
	v := 0
	if _, err := fmt.Sscanf(ClientVersion, "%d", &v); err != nil {
		return 0, err
	}
	return v, nil
}
