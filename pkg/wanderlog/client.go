package wanderlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	BaseURL          = "https://wanderlog.com/api"
	ClientVersion    = "2"
	DefaultUserAgent = "wanderlog-cli/1.0"
)

type Client struct {
	httpClient *http.Client
	logger     *logrus.Logger
	userAgent  string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:    logrus.New(),
		userAgent: DefaultUserAgent,
	}
}

func (c *Client) SetLogger(logger *logrus.Logger) {
	c.logger = logger
}

func (c *Client) GetTrip(tripID string) (*TripResponse, error) {
	url := fmt.Sprintf("%s/tripPlans/%s?clientSchemaVersion=%s", BaseURL, tripID, ClientVersion)
	
	c.logger.WithFields(logrus.Fields{
		"url":    url,
		"tripID": tripID,
	}).Debug("Fetching trip data")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var tripResponse TripResponse
	if err := json.NewDecoder(resp.Body).Decode(&tripResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"tripID": tripID,
		"title":  tripResponse.TripPlan.Title,
	}).Info("Successfully fetched trip data")

	return &tripResponse, nil
}