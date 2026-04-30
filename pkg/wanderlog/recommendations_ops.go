package wanderlog

import (
	"context"
	"fmt"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

type (
	RecommendedPlacesRequest               = models.RecommendedPlacesRequest
	RecommendedPlacesResponse              = models.RecommendedPlacesResponse
	MarkRecommendationNotInterestedRequest = models.MarkRecommendationNotInterestedRequest
)

// GetRecommendedPlaces returns place recommendations for a trip + geo, using
// the v2 recommendations endpoint.
func (c *Client) GetRecommendedPlaces(req RecommendedPlacesRequest) (*RecommendedPlacesResponse, error) {
	if req.TripPlanID == 0 {
		return nil, fmt.Errorf("GetRecommendedPlaces: tripPlanId is required")
	}
	if req.GeoID == 0 {
		return nil, fmt.Errorf("GetRecommendedPlaces: geoId is required")
	}
	if err := c.requireAuth("GetRecommendedPlaces"); err != nil {
		return nil, err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "recommendations/v2", nil, req, true)
	if err != nil {
		return nil, err
	}
	var result RecommendedPlacesResponse
	if err := decodeAPIBody("GetRecommendedPlaces", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MarkRecommendationNotInterested tells the recommender that the user is not
// interested in a particular Maps place suggestion for a trip.
func (c *Client) MarkRecommendationNotInterested(req MarkRecommendationNotInterestedRequest) error {
	if req.TripPlanID == 0 {
		return fmt.Errorf("MarkRecommendationNotInterested: tripPlanId is required")
	}
	if req.MapsPlaceID == "" {
		return fmt.Errorf("MarkRecommendationNotInterested: mapsPlaceId is required")
	}
	if err := c.requireAuth("MarkRecommendationNotInterested"); err != nil {
		return err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "recommendations/notInterested", nil, req, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("MarkRecommendationNotInterested", resp.StatusCode, resp.Body, nil)
}
