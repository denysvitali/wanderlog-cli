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
	GlobalConfig        = models.GlobalConfig
	SessionStore        = models.SessionStore
	SessionStoreRequest = models.SessionStoreRequest
	SessionPreferences  = models.SessionPreferences
)

// GetGlobalConfig fetches the server's global client configuration.
func (c *Client) GetGlobalConfig() (*GlobalConfig, error) {
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "config/globalConfig", nil, nil, false)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GetGlobalConfig: HTTP %d: %s", resp.StatusCode, truncateForLog(string(resp.Body), 500))
	}
	var cfg GlobalConfig
	if err := json.Unmarshal(resp.Body, &cfg); err != nil {
		return nil, fmt.Errorf("GetGlobalConfig: decoding response: %w", err)
	}
	cfg.Raw = json.RawMessage(resp.Body)
	return &cfg, nil
}

// GetSessionStore returns the authenticated session's key-value store.
func (c *Client) GetSessionStore() (*SessionStore, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "sessionStore", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result SessionStore
	if err := decodeAPIBody("GetSessionStore", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetSessionStoreValue writes a single key into the session store.
func (c *Client) SetSessionStoreValue(key string, value any) error {
	if err := c.requireAuth("SetSessionStoreValue"); err != nil {
		return err
	}
	body := SessionStoreRequest{Key: key, Value: value}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "sessionStore", nil, body, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("SetSessionStoreValue", resp.StatusCode, resp.Body, nil)
}

// GetSessionPreferences returns the locale-scoped session preferences.
func (c *Client) GetSessionPreferences(locale string) (*SessionPreferences, error) {
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "sessionStore/preferences/"+url.PathEscape(locale), nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result SessionPreferences
	if err := decodeAPIBody("GetSessionPreferences", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
