package wanderlog

import (
	"encoding/json"
	"fmt"
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
	raw, err := c.doRaw("GET", "/config/globalConfig", nil, false, "GetGlobalConfig")
	if err != nil {
		return nil, err
	}
	var cfg GlobalConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("GetGlobalConfig: decoding response: %w", err)
	}
	cfg.Raw = json.RawMessage(raw)
	return &cfg, nil
}

// GetSessionStore returns the authenticated session's key-value store.
func (c *Client) GetSessionStore() (*SessionStore, error) {
	var resp SessionStore
	if err := c.doJSON("GET", "/sessionStore", nil, &resp, true, "GetSessionStore"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetSessionStoreValue writes a single key into the session store.
func (c *Client) SetSessionStoreValue(key string, value any) error {
	body := SessionStoreRequest{Key: key, Value: value}
	return c.doJSON("POST", "/sessionStore", body, nil, true, "SetSessionStoreValue")
}

// GetSessionPreferences returns the locale-scoped session preferences.
func (c *Client) GetSessionPreferences(locale string) (*SessionPreferences, error) {
	path := fmt.Sprintf("/sessionStore/preferences/%s", url.PathEscape(locale))
	var resp SessionPreferences
	if err := c.doJSON("GET", path, nil, &resp, false, "GetSessionPreferences"); err != nil {
		return nil, err
	}
	return &resp, nil
}
