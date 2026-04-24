package wanderlog

import (
	"fmt"

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
	var resp FeedHomeResponse
	if err := c.doJSON("GET", "/tripPlans/home", nil, &resp, true, "GetFeedHome"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFeed fetches the legacy trip feed.
func (c *Client) GetFeed() (*FeedResponse, error) {
	var resp FeedResponse
	if err := c.doJSON("GET", "/tripPlans/feed", nil, &resp, true, "GetFeed"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFeedV2 fetches the v2 trip feed.
func (c *Client) GetFeedV2() (*FeedResponse, error) {
	var resp FeedResponse
	if err := c.doJSON("GET", "/tripPlans/feed/v2", nil, &resp, true, "GetFeedV2"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFeedMostRecent returns the user's most recently edited trip.
func (c *Client) GetFeedMostRecent() (*FeedRecentResponse, error) {
	var resp FeedRecentResponse
	if err := c.doJSON("GET", "/tripPlans/feed/mostRecentlyEdited", nil, &resp, true, "GetFeedMostRecent"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFriendsPlans fetches trip plans published by the user's friends.
func (c *Client) GetFriendsPlans() (*FriendsPlansResponse, error) {
	var resp FriendsPlansResponse
	if err := c.doJSON("GET", "/tripPlans/friendsPlans", nil, &resp, true, "GetFriendsPlans"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTripHistory returns the paginated trip edit history.
func (c *Client) GetTripHistory(offset int) (*TripHistoryResponse, error) {
	path := "/tripPlans/history"
	if offset > 0 {
		path = fmt.Sprintf("%s?offset=%d", path, offset)
	}
	var resp TripHistoryResponse
	if err := c.doJSON("GET", path, nil, &resp, true, "GetTripHistory"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetIfEdited asks the server which of the given trip plans changed since the
// provided revisions. Useful for cache invalidation.
func (c *Client) GetIfEdited(req GetIfEditedRequest) (*GetIfEditedResponse, error) {
	if req.ClientSchemaVersion == 0 {
		if v, err := clientSchemaVersionInt(); err == nil {
			req.ClientSchemaVersion = v
		}
	}
	var resp GetIfEditedResponse
	if err := c.doJSON("POST", "/tripPlans/getIfEdited", req, &resp, true, "GetIfEdited"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BrowseGuides returns curated travel guides. When geoID is non-zero the guides
// are scoped to that geography.
func (c *Client) BrowseGuides(geoID int) (*BrowseGuidesResponse, error) {
	path := "/tripPlans/browse/guides"
	if geoID > 0 {
		path = fmt.Sprintf("%s/%d", path, geoID)
	}
	var resp BrowseGuidesResponse
	if err := c.doJSON("GET", path, nil, &resp, false, "BrowseGuides"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func clientSchemaVersionInt() (int, error) {
	v := 0
	if _, err := fmt.Sscanf(ClientVersion, "%d", &v); err != nil {
		return 0, err
	}
	return v, nil
}
