package wanderlog

import (
	"encoding/json"
	"fmt"
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

// GetMe fetches the currently authenticated user's profile.
func (c *Client) GetMe() (*UserProfile, error) {
	raw, err := c.doRaw("GET", "/user", nil, true, "GetMe")
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return nil, fmt.Errorf("GetMe: decoding response: %w", err)
	}
	profile.Raw = json.RawMessage(raw)
	return &profile, nil
}

// UpdateMe updates the authenticated user's profile.
func (c *Client) UpdateMe(req UpdateUserRequest) (*UserProfile, error) {
	raw, err := c.doRaw("POST", "/user", req, true, "UpdateMe")
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return nil, fmt.Errorf("UpdateMe: decoding response: %w", err)
	}
	profile.Raw = json.RawMessage(raw)
	return &profile, nil
}

// ServerLogout invokes POST /api/user/logout to invalidate the server session.
// Local credential removal is the caller's responsibility (see keychain.go).
func (c *Client) ServerLogout() error {
	return c.doJSON("POST", "/user/logout", nil, nil, true, "ServerLogout")
}

// GetNotifications returns the authenticated user's notification inbox.
// offset is the pagination cursor; pass 0 for the first page.
func (c *Client) GetNotifications(offset int) (*NotificationsResponse, error) {
	path := "/user/notifications"
	if offset > 0 {
		path = fmt.Sprintf("%s?offset=%d", path, offset)
	}
	var resp NotificationsResponse
	if err := c.doJSON("GET", path, nil, &resp, true, "GetNotifications"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MarkNotificationsRead marks the given notification IDs as read.
func (c *Client) MarkNotificationsRead(ids []string) error {
	body := struct {
		NotificationIDs []string `json:"notificationIds"`
	}{NotificationIDs: ids}
	return c.doJSON("POST", "/user/notifications/markRead", body, nil, true, "MarkNotificationsRead")
}

// GetNotificationSettings returns the user's notification settings.
func (c *Client) GetNotificationSettings() (*NotificationSettings, error) {
	var resp NotificationSettings
	if err := c.doJSON("GET", "/user/notification/settings", nil, &resp, true, "GetNotificationSettings"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateNotificationSettings replaces the user's notification settings.
func (c *Client) UpdateNotificationSettings(settings json.RawMessage) (*NotificationSettings, error) {
	body := UpdateNotificationSettingsRequest{NotificationSettings: settings}
	var resp NotificationSettings
	if err := c.doJSON("POST", "/user/notification/settings", body, &resp, true, "UpdateNotificationSettings"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetKeyValue fetches a single value from the user's per-account key-value store.
func (c *Client) GetKeyValue(key string) (json.RawMessage, error) {
	path := fmt.Sprintf("/user/keyValue/%s", url.PathEscape(key))
	var resp KeyValueResponse
	if err := c.doJSON("GET", path, nil, &resp, true, "GetKeyValue"); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// SetKeyValue stores a value in the user's per-account key-value store.
func (c *Client) SetKeyValue(key string, value any) error {
	path := fmt.Sprintf("/user/keyValue/%s", url.PathEscape(key))
	body := struct {
		Value any `json:"value"`
	}{Value: value}
	return c.doJSON("POST", path, body, nil, true, "SetKeyValue")
}

// SetUTCOffset persists the user's timezone offset (in minutes from UTC).
func (c *Client) SetUTCOffset(offsetMinutes int) error {
	body := struct {
		UTCOffset int `json:"utcOffset"`
	}{UTCOffset: offsetMinutes}
	return c.doJSON("POST", "/user/utcOffset", body, nil, true, "SetUTCOffset")
}

// ListFollowing returns follow relationships for the given user IDs.
func (c *Client) ListFollowing(userIDs []string) (*FollowingResponse, error) {
	body := struct {
		UserIDs []string `json:"userIds"`
	}{UserIDs: userIDs}
	raw, err := c.doRaw("POST", "/user/following/list", body, true, "ListFollowing")
	if err != nil {
		return nil, err
	}
	var resp FollowingResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ListFollowing: decoding response: %w", err)
	}
	resp.Raw = json.RawMessage(raw)
	return &resp, nil
}

// AutocompleteUsers searches users by prefix.
func (c *Client) AutocompleteUsers(search string) (*UserAutocompleteResponse, error) {
	path := fmt.Sprintf("/user/autocomplete/%s", url.PathEscape(search))
	var resp UserAutocompleteResponse
	if err := c.doJSON("GET", path, nil, &resp, true, "AutocompleteUsers"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindUserByEmail looks up a user by email address.
func (c *Client) FindUserByEmail(email string) (*UserProfile, error) {
	path := fmt.Sprintf("/user/byEmail?email=%s", url.QueryEscape(email))
	raw, err := c.doRaw("GET", path, nil, true, "FindUserByEmail")
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return nil, fmt.Errorf("FindUserByEmail: decoding response: %w", err)
	}
	profile.Raw = json.RawMessage(raw)
	return &profile, nil
}

// BlockUser blocks the target user ID.
func (c *Client) BlockUser(userID string) error {
	body := struct {
		UserID string `json:"userId"`
	}{UserID: userID}
	return c.doJSON("POST", "/user/block", body, nil, true, "BlockUser")
}

// IsUsernameTaken reports whether a username is already in use.
func (c *Client) IsUsernameTaken(username string) (bool, error) {
	path := fmt.Sprintf("/user/isUsernameTaken/%s", url.PathEscape(username))
	var resp UsernameTakenResponse
	if err := c.doJSON("GET", path, nil, &resp, false, "IsUsernameTaken"); err != nil {
		return false, err
	}
	return resp.Taken, nil
}

// GetUserEmails returns the authenticated user's registered email addresses.
func (c *Client) GetUserEmails() (*UserEmailsResponse, error) {
	var resp UserEmailsResponse
	if err := c.doJSON("GET", "/user/emails", nil, &resp, true, "GetUserEmails"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserProfile fetches another user's profile and public trips by numeric ID.
func (c *Client) GetUserProfile(userID int) (*ProfileTripsResponse, error) {
	path := fmt.Sprintf("/tripPlans/profile/%d", userID)
	var resp ProfileTripsResponse
	if err := c.doJSON("GET", path, nil, &resp, false, "GetUserProfile"); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserProfileByUsername fetches a user's profile by username.
func (c *Client) GetUserProfileByUsername(username string) (*ProfileTripsResponse, error) {
	path := fmt.Sprintf("/tripPlans/profile/byUsername/%s", url.PathEscape(username))
	var resp ProfileTripsResponse
	if err := c.doJSON("GET", path, nil, &resp, false, "GetUserProfileByUsername"); err != nil {
		return nil, err
	}
	return &resp, nil
}
