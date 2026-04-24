package models

import "encoding/json"

// FeedHomeResponse mirrors /api/tripPlans/home. Most sub-lists are loosely
// typed because the bundle mixes guide cards, hotel deals, and trip summaries.
type FeedHomeResponse struct {
	Success                       bool            `json:"success"`
	FriendsPrivateSharedTripPlans json.RawMessage `json:"friendsPrivateSharedTripPlans,omitempty"`
	FriendsTripPlans              json.RawMessage `json:"friendsTripPlans,omitempty"`
	HeroGuide                     json.RawMessage `json:"heroGuide,omitempty"`
	HotelDeals                    json.RawMessage `json:"hotelDeals,omitempty"`
	OwnGuides                     json.RawMessage `json:"ownGuides,omitempty"`
	OwnTripPlans                  json.RawMessage `json:"ownTripPlans,omitempty"`
	RecommendedGuides             json.RawMessage `json:"recommendedGuides,omitempty"`
	RelatedToUpcomingGuides       json.RawMessage `json:"relatedToUpcomingGuides,omitempty"`
}

// FeedResponse is the shape used by /api/tripPlans/feed and /api/tripPlans/feed/v2.
type FeedResponse struct {
	Success             bool            `json:"success"`
	Data                json.RawMessage `json:"data,omitempty"`
	MostRecentlyEdited  json.RawMessage `json:"mostRecentlyEdited,omitempty"`
	HasSetupFlightDeals *bool           `json:"hasSetupFlightDeals,omitempty"`
}

// FeedRecentResponse is the response from /api/tripPlans/feed/mostRecentlyEdited.
type FeedRecentResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// FriendsPlansResponse is the response from /api/tripPlans/friendsPlans.
type FriendsPlansResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// TripHistoryResponse mirrors /api/tripPlans/history with pagination.
type TripHistoryResponse struct {
	Success    bool            `json:"success"`
	Data       json.RawMessage `json:"data,omitempty"`
	NextOffset *int            `json:"nextOffset,omitempty"`
}

// GetIfEditedRequest matches POST /api/tripPlans/getIfEdited.
type GetIfEditedRequest struct {
	TripPlans           []EditCheck `json:"tripPlans"`
	ClientSchemaVersion int         `json:"clientSchemaVersion"`
	Platform            string      `json:"platform,omitempty"`
}

type EditCheck struct {
	Key              string `json:"key"`
	LastEditedAt     string `json:"lastEditedAt,omitempty"`
	LastEditRevision int    `json:"lastEditRevision,omitempty"`
}

type GetIfEditedResponse struct {
	Success   bool            `json:"success"`
	TripPlans json.RawMessage `json:"tripPlans,omitempty"`
}

// BrowseGuidesResponse is the response from /api/tripPlans/browse/guides[/{geoId}].
type BrowseGuidesResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}
