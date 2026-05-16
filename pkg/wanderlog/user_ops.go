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
	UserProfile                       = models.UserProfile
	UpdateUserRequest                 = models.UpdateUserRequest
	NotificationsResponse             = models.NotificationsResponse
	Notification                      = models.Notification
	NotificationSettings              = models.NotificationSettings
	UpdateNotificationSettingsRequest = models.UpdateNotificationSettingsRequest
	KeyValueResponse                  = models.KeyValueResponse
	FollowingResponse                 = models.FollowingResponse
	UserAutocompleteResponse          = models.UserAutocompleteResponse
	UserAutocomplete                  = models.UserAutocomplete
	UserEmailsResponse                = models.UserEmailsResponse
	UserEmail                         = models.UserEmail
	UsernameTakenResponse             = models.UsernameTakenResponse
	ProfileTripsResponse              = models.ProfileTripsResponse
)

func (c *Client) requireAuth(opName string) error {
	if c.auth == nil {
		return fmt.Errorf("%s: authentication required", opName)
	}
	return nil
}

// GetMe fetches the currently authenticated user's profile.
func (c *Client) GetMe() (*UserProfile, error) {
	if err := c.requireAuth("GetMe"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "user", nil, nil, true)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := decodeAPIBody("GetMe", resp.StatusCode, resp.Body, &profile); err != nil {
		return nil, fmt.Errorf("GetMe: decoding response: %w", err)
	}
	profile.Raw = resp.Body
	return &profile, nil
}

// UpdateMe updates the authenticated user's profile.
func (c *Client) UpdateMe(req UpdateUserRequest) (*UserProfile, error) {
	if err := c.requireAuth("UpdateMe"); err != nil {
		return nil, err
	}
	body := map[string]any{}
	if req.Name != "" {
		body["name"] = req.Name
	}
	if req.Username != "" {
		body["username"] = req.Username
	}
	if req.Bio != "" {
		body["bio"] = req.Bio
	}
	if req.Location != "" {
		body["location"] = req.Location
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "user", nil, body, true)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := decodeAPIBody("UpdateMe", resp.StatusCode, resp.Body, &profile); err != nil {
		return nil, fmt.Errorf("UpdateMe: decoding response: %w", err)
	}
	profile.Raw = json.RawMessage(resp.Body)
	return &profile, nil
}

// ServerLogout invokes POST /api/user/logout to invalidate the server session.
// Local credential removal is the caller's responsibility (see keychain.go).
func (c *Client) ServerLogout() error {
	if err := c.requireAuth("ServerLogout"); err != nil {
		return err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodPost, "user/logout", nil, nil, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("ServerLogout", resp.StatusCode, resp.Body, nil)
}

// GetNotifications returns the authenticated user's notification inbox.
// offset is the pagination cursor; pass 0 for the first page.
func (c *Client) GetNotifications(offset int) (*NotificationsResponse, error) {
	if err := c.requireAuth("GetNotifications"); err != nil {
		return nil, err
	}

	c.logger.WithField("offset", offset).Debug("Getting notifications")

	// Build URL with optional offset parameter
	path := "/user/notifications"
	if offset > 0 {
		path = fmt.Sprintf("/user/notifications?offset=%d", offset)
	}

	statusCode, respBody, err := c.DoAPI("GET", path, nil, nil, true)
	if err != nil {
		return nil, fmt.Errorf("GetNotifications: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("GetNotifications: HTTP %d: %s", statusCode, truncateForLog(string(respBody), 500))
	}

	var result NotificationsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("GetNotifications: decoding response: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("GetNotifications: API returned success=false")
	}
	return &result, nil
}

// MarkNotificationsRead marks the given notification IDs as read.
func (c *Client) MarkNotificationsRead(ids []string) error {
	if err := c.requireAuth("MarkNotificationsRead"); err != nil {
		return err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "user/notifications/markRead", nil, map[string]any{"notificationIds": ids}, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("MarkNotificationsRead", resp.StatusCode, resp.Body, nil)
}

// GetNotificationSettings returns the user's notification settings.
func (c *Client) GetNotificationSettings() (*NotificationSettings, error) {
	if err := c.requireAuth("GetNotificationSettings"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "user/notification/settings", nil, nil, true)
	if err != nil {
		return nil, err
	}
	var result NotificationSettings
	if err := decodeAPIBody("GetNotificationSettings", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateNotificationSettings replaces the user's notification settings.
func (c *Client) UpdateNotificationSettings(settings json.RawMessage) (*NotificationSettings, error) {
	if err := c.requireAuth("UpdateNotificationSettings"); err != nil {
		return nil, err
	}
	body := UpdateNotificationSettingsRequest{NotificationSettings: settings}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "user/notification/settings", nil, body, true)
	if err != nil {
		return nil, err
	}
	var result NotificationSettings
	if err := decodeAPIBody("UpdateNotificationSettings", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetKeyValue fetches a single value from the user's per-account key-value store.
func (c *Client) GetKeyValue(key string) (json.RawMessage, error) {
	if err := c.requireAuth("GetKeyValue"); err != nil {
		return nil, err
	}
	apiResp, err := c.apiRequest(context.Background(), http.MethodGet, "user/keyValue/"+url.PathEscape(key), nil, nil, true)
	if err != nil {
		return nil, err
	}
	var resp KeyValueResponse
	if err := decodeAPIBody("GetKeyValue", apiResp.StatusCode, apiResp.Body, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// SetKeyValue stores a value in the user's per-account key-value store.
func (c *Client) SetKeyValue(key string, value any) error {
	if err := c.requireAuth("SetKeyValue"); err != nil {
		return err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "user/keyValue/"+url.PathEscape(key), nil, map[string]any{"value": value}, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("SetKeyValue", resp.StatusCode, resp.Body, nil)
}

// SetUTCOffset persists the user's timezone offset (in minutes from UTC).
func (c *Client) SetUTCOffset(offsetMinutes int) error {
	if err := c.requireAuth("SetUTCOffset"); err != nil {
		return err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "user/utcOffset", nil, map[string]any{"utcOffset": offsetMinutes}, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("SetUTCOffset", resp.StatusCode, resp.Body, nil)
}

// ListFollowing returns follow relationships for the given user IDs.
func (c *Client) ListFollowing(userIDs []string) (*FollowingResponse, error) {
	if err := c.requireAuth("ListFollowing"); err != nil {
		return nil, err
	}
	body := struct {
		UserIDs []string `json:"userIds"`
	}{UserIDs: userIDs}
	apiResp, err := c.apiJSON(context.Background(), http.MethodPost, "user/following/list", nil, body, true)
	if err != nil {
		return nil, err
	}
	var result FollowingResponse
	if err := decodeAPIBody("ListFollowing", apiResp.StatusCode, apiResp.Body, &result); err != nil {
		return nil, fmt.Errorf("ListFollowing: decoding response: %w", err)
	}
	result.Raw = json.RawMessage(apiResp.Body)
	return &result, nil
}

// AutocompleteUsers searches users by prefix.
func (c *Client) AutocompleteUsers(search string) (*UserAutocompleteResponse, error) {
	if err := c.requireAuth("AutocompleteUsers"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "user/autocomplete/"+url.PathEscape(search), nil, nil, true)
	if err != nil {
		return nil, err
	}
	var result UserAutocompleteResponse
	if err := decodeAPIBody("AutocompleteUsers", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FindUserByEmail looks up a user by email address.
func (c *Client) FindUserByEmail(email string) (*UserProfile, error) {
	if err := c.requireAuth("FindUserByEmail"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "user/byEmail", apiQuery(map[string]string{"email": email}), nil, true)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := decodeAPIBody("FindUserByEmail", resp.StatusCode, resp.Body, &profile); err != nil {
		return nil, fmt.Errorf("FindUserByEmail: decoding response: %w", err)
	}
	profile.Raw = json.RawMessage(resp.Body)
	return &profile, nil
}

// BlockUser blocks the target user ID.
func (c *Client) BlockUser(userID string) error {
	if err := c.requireAuth("BlockUser"); err != nil {
		return err
	}
	body := struct {
		UserID string `json:"userId"`
	}{UserID: userID}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "user/block", nil, body, true)
	if err != nil {
		return err
	}
	return decodeAPIBody("BlockUser", resp.StatusCode, resp.Body, nil)
}

// IsUsernameTaken reports whether a username is already in use.
func (c *Client) IsUsernameTaken(username string) (bool, error) {
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "user/isUsernameTaken/"+url.PathEscape(username), nil, nil, false)
	if err != nil {
		return false, err
	}
	var result UsernameTakenResponse
	if err := decodeAPIBody("IsUsernameTaken", resp.StatusCode, resp.Body, &result); err != nil {
		return false, err
	}
	return result.Taken, nil
}

// GetUserEmails returns the authenticated user's registered email addresses.
func (c *Client) GetUserEmails() (*UserEmailsResponse, error) {
	if err := c.requireAuth("GetUserEmails"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "user/emails", nil, nil, true)
	if err != nil {
		return nil, err
	}
	var result UserEmailsResponse
	if err := decodeAPIBody("GetUserEmails", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserProfile fetches another user's profile and public trips by numeric ID.
func (c *Client) GetUserProfile(userID int) (*ProfileTripsResponse, error) {
	resp, err := c.apiRequest(context.Background(), http.MethodGet, fmt.Sprintf("tripPlans/profile/%d", userID), nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result ProfileTripsResponse
	if err := decodeAPIBody("GetUserProfile", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserProfileByUsername fetches a user's profile by username.
func (c *Client) GetUserProfileByUsername(username string) (*ProfileTripsResponse, error) {
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "tripPlans/profile/byUsername/"+url.PathEscape(username), nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result ProfileTripsResponse
	if err := decodeAPIBody("GetUserProfileByUsername", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
