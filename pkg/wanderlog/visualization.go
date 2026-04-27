package wanderlog

import (
	"context"
	"fmt"
)

// UserTripsResponse represents the response from getting user's trips
type UserTripsResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID             int     `json:"id"`
		Key            string  `json:"key"`
		Title          string  `json:"title"`
		StartDate      string  `json:"startDate"`
		EndDate        string  `json:"endDate"`
		PlaceCount     int     `json:"placeCount"`
		ViewCount      int     `json:"viewCount"`
		LikeCount      int     `json:"likeCount"`
		HeaderImageKey string  `json:"headerImageKey"`
		TopImageKeys   []any   `json:"topImageKeys"`
		EditedAt       string  `json:"editedAt"`
		OpenedAt       string  `json:"openedAt"`
		IsPrimary      bool    `json:"isPrimary"`
		CommentCount   *int    `json:"commentCount"`
		Distinction    *string `json:"distinction"`
		ItemType       string  `json:"itemType"`
		AuthorBlurb    string  `json:"authorBlurb"`
		IsDraft        bool    `json:"isDraft"`
		ImageCount     *int    `json:"imageCount"`
		Type           string  `json:"type"`
		KeyType        string  `json:"keyType"`

		// User who created the trip
		User struct {
			ID                  int    `json:"id"`
			Username            string `json:"username"`
			Name                string `json:"name"`
			ProfilePictureKey   string `json:"profilePictureKey"`
			VisitGeosCount      int    `json:"visitGeosCount"`
			CountriesCount      int    `json:"countriesCount"`
			ShowProfileProBadge bool   `json:"showProfileProBadge"`
			IsProUser           bool   `json:"isProUser"`
		} `json:"user"`

		// Collaborators on the trip
		Collaborators []struct {
			ID                  int     `json:"id"`
			Username            string  `json:"username"`
			Name                string  `json:"name"`
			ProfilePictureKey   *string `json:"profilePictureKey"`
			VisitGeosCount      int     `json:"visitGeosCount"`
			CountriesCount      int     `json:"countriesCount"`
			ShowProfileProBadge bool    `json:"showProfileProBadge"`
			IsProUser           bool    `json:"isProUser"`
		} `json:"collaborators"`
	} `json:"data"`
	RecentlyOpened []interface{} `json:"recentlyOpened"`
}

// TripImagesResponse represents trip images
type TripImagesResponse struct {
	Success bool `json:"success"`
	Images  []struct {
		ID           int    `json:"id"`
		Key          string `json:"key"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnailUrl"`
		Caption      string `json:"caption"`
		PlaceID      string `json:"placeId,omitempty"`
	} `json:"images"`
}

// TripStatsResponse represents trip statistics
type TripStatsResponse struct {
	Success bool `json:"success"`
	Stats   struct {
		TotalDistance  float64 `json:"totalDistance"`
		FlightDistance float64 `json:"flightDistance"`
		GroundDistance float64 `json:"groundDistance"`
		Countries      int     `json:"countries"`
		Cities         int     `json:"cities"`
		TimeZones      int     `json:"timeZones"`
		FlightDuration int     `json:"flightDuration"` // minutes
		EstimatedCost  struct {
			Amount       float64 `json:"amount"`
			CurrencyCode string  `json:"currencyCode"`
		} `json:"estimatedCost"`
	} `json:"stats"`
}

// GetUserTrips retrieves all trips for the authenticated user
func (c *Client) GetUserTrips() (*UserTripsResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.Debug("Fetching user trips")

	resp, err := api.ListUserTripPlansWithResponse(context.Background())
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	var result UserTripsResponse
	if err := decodeOpenAPIBody("GetUserTrips", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}

	c.logger.WithField("tripCount", len(result.Data)).Info("Successfully fetched user trips")

	return &result, nil
}

// GetTripImages retrieves images for a trip
func (c *Client) GetTripImages(tripKey string) (*TripImagesResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithField("tripKey", tripKey).Debug("Fetching trip images")

	resp, err := api.GetTripPlanImagesWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	var result TripImagesResponse
	if err := decodeOpenAPIBody("GetTripImages", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":    tripKey,
		"imageCount": len(result.Images),
	}).Info("Successfully fetched trip images")

	return &result, nil
}

// GetTripPlaces retrieves places for a trip with additional details
func (c *Client) GetTripPlaces(tripKey string) (*TripResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	c.logger.WithField("tripKey", tripKey).Debug("Fetching trip places")

	resp, err := api.GetTripPlanPlacesWithResponse(context.Background(), tripKey)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	var result TripResponse
	if err := decodeOpenAPIBody("GetTripPlaces", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully fetched trip places")

	return &result, nil
}

// LikeTrip likes or unlikes a trip.
//
// Deprecated: Use SetLike in write_ops.go instead.
func (c *Client) LikeTrip(tripKey string, liked bool) error {
	action := "unliked"
	if liked {
		action = "liked"
	}
	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"action":  action,
	}).Debug("Updating trip like status")
	if err := c.SetLike(tripKey, liked); err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"liked":   liked,
	}).Info("Successfully updated trip like status")

	return nil
}

// RegisterView registers a view for analytics.
//
// Deprecated: Use RegisterTripView in journal_ops.go instead.
func (c *Client) RegisterView(tripKey string) error {
	c.logger.WithField("tripKey", tripKey).Debug("Registering trip view")
	if err := c.RegisterTripView(tripKey); err != nil {
		return err
	}

	c.logger.WithField("tripKey", tripKey).Debug("Successfully registered view")

	return nil
}
