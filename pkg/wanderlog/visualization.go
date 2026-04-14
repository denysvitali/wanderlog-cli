package wanderlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	url := fmt.Sprintf("%s/tripPlans", BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	if c.auth != nil {
		_ = c.addAuthHeaders(req)
	}

	c.logger.Debug("Fetching user trips")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var tripsResponse UserTripsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tripsResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.WithField("tripCount", len(tripsResponse.Data)).Info("Successfully fetched user trips")

	return &tripsResponse, nil
}

// GetTripImages retrieves images for a trip
func (c *Client) GetTripImages(tripKey string) (*TripImagesResponse, error) {
	url := fmt.Sprintf("%s/tripPlans/%s/images", BaseURL, tripKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	c.logger.WithField("tripKey", tripKey).Debug("Fetching trip images")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var imagesResponse TripImagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&imagesResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":    tripKey,
		"imageCount": len(imagesResponse.Images),
	}).Info("Successfully fetched trip images")

	return &imagesResponse, nil
}

// GetTripPlaces retrieves places for a trip with additional details
func (c *Client) GetTripPlaces(tripKey string) (*TripResponse, error) {
	url := fmt.Sprintf("%s/tripPlans/%s/places", BaseURL, tripKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	c.logger.WithField("tripKey", tripKey).Debug("Fetching trip places")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var placesResponse TripResponse
	if err := json.NewDecoder(resp.Body).Decode(&placesResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.WithField("tripKey", tripKey).Info("Successfully fetched trip places")

	return &placesResponse, nil
}

// LikeTrip likes or unlikes a trip
func (c *Client) LikeTrip(tripKey string, liked bool) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for liking trips")
	}

	likeReq := struct {
		Liked bool `json:"liked"`
	}{
		Liked: liked,
	}

	reqBody, err := json.Marshal(likeReq)
	if err != nil {
		return fmt.Errorf("marshaling like request: %w", err)
	}

	url := fmt.Sprintf("%s/tripPlans/%s/like", BaseURL, tripKey)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if err := c.addAuthHeaders(req); err != nil {
		return fmt.Errorf("adding auth headers: %w", err)
	}

	action := "unliked"
	if liked {
		action = "liked"
	}
	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"action":  action,
	}).Debug("Updating trip like status")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"liked":   liked,
	}).Info("Successfully updated trip like status")

	return nil
}

// RegisterView registers a view for analytics
func (c *Client) RegisterView(tripKey string) error {
	url := fmt.Sprintf("%s/tripPlans/%s/registerView", BaseURL, tripKey)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	// Views can be registered without authentication
	if c.auth != nil {
		_ = c.addAuthHeaders(req)
	}

	c.logger.WithField("tripKey", tripKey).Debug("Registering trip view")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	c.logger.WithField("tripKey", tripKey).Debug("Successfully registered view")

	return nil
}
