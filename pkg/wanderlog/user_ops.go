package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	openapitypes "github.com/oapi-codegen/runtime/types"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
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

func jsonBodyEditor(body any) (openapi.RequestEditorFn, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return func(_ context.Context, req *http.Request) error {
		req.Body = io.NopCloser(bytes.NewReader(encoded))
		req.ContentLength = int64(len(encoded))
		req.Header.Set("Content-Type", "application/json")
		return nil
	}, nil
}

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
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetLoggedInUserWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := decodeOpenAPIBody("GetMe", resp.StatusCode(), resp.Body, &profile); err != nil {
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
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	body := openapi.UpdateProfileJSONRequestBody{}
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
	resp, err := api.UpdateProfileWithResponse(context.Background(), body)
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := decodeOpenAPIBody("UpdateMe", resp.StatusCode(), resp.Body, &profile); err != nil {
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
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	resp, err := api.LogoutWithResponse(context.Background())
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("ServerLogout", resp.StatusCode(), resp.Body, nil)
}

// GetNotifications returns the authenticated user's notification inbox.
// offset is the pagination cursor; pass 0 for the first page.
func (c *Client) GetNotifications(offset int) (*NotificationsResponse, error) {
	if err := c.requireAuth("GetNotifications"); err != nil {
		return nil, err
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	editors := []openapi.RequestEditorFn{}
	if offset > 0 {
		editors = append(editors, func(_ context.Context, req *http.Request) error {
			values := req.URL.Query()
			values.Set("offset", fmt.Sprintf("%d", offset))
			req.URL.RawQuery = values.Encode()
			return nil
		})
	}
	apiResp, err := api.GetNotificationsWithResponse(context.Background(), editors...)
	if err != nil {
		return nil, err
	}
	var result NotificationsResponse
	if err := decodeOpenAPIBody("GetNotifications", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MarkNotificationsRead marks the given notification IDs as read.
func (c *Client) MarkNotificationsRead(ids []string) error {
	if err := c.requireAuth("MarkNotificationsRead"); err != nil {
		return err
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	resp, err := api.MarkNotificationsReadWithResponse(context.Background(), openapi.MarkNotificationsReadJSONRequestBody{NotificationIds: &ids})
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("MarkNotificationsRead", resp.StatusCode(), resp.Body, nil)
}

// GetNotificationSettings returns the user's notification settings.
func (c *Client) GetNotificationSettings() (*NotificationSettings, error) {
	if err := c.requireAuth("GetNotificationSettings"); err != nil {
		return nil, err
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetNotificationSettingsWithResponse(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	var result NotificationSettings
	if err := decodeOpenAPIBody("GetNotificationSettings", resp.StatusCode(), resp.Body, &result); err != nil {
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
	editor, err := jsonBodyEditor(body)
	if err != nil {
		return nil, fmt.Errorf("UpdateNotificationSettings: marshaling request: %w", err)
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.UpdateNotificationSettingsWithResponse(context.Background(), editor)
	if err != nil {
		return nil, err
	}
	var result NotificationSettings
	if err := decodeOpenAPIBody("UpdateNotificationSettings", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetKeyValue fetches a single value from the user's per-account key-value store.
func (c *Client) GetKeyValue(key string) (json.RawMessage, error) {
	if err := c.requireAuth("GetKeyValue"); err != nil {
		return nil, err
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.GetKeyValueWithResponse(context.Background(), key)
	if err != nil {
		return nil, err
	}
	var resp KeyValueResponse
	if err := decodeOpenAPIBody("GetKeyValue", apiResp.StatusCode(), apiResp.Body, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// SetKeyValue stores a value in the user's per-account key-value store.
func (c *Client) SetKeyValue(key string, value any) error {
	if err := c.requireAuth("SetKeyValue"); err != nil {
		return err
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	resp, err := api.SetKeyValueWithResponse(context.Background(), key, openapi.SetKeyValueJSONRequestBody{Value: value})
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("SetKeyValue", resp.StatusCode(), resp.Body, nil)
}

// SetUTCOffset persists the user's timezone offset (in minutes from UTC).
func (c *Client) SetUTCOffset(offsetMinutes int) error {
	if err := c.requireAuth("SetUTCOffset"); err != nil {
		return err
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	resp, err := api.UpdateUTCOffsetWithResponse(context.Background(), openapi.UpdateUTCOffsetJSONRequestBody{UtcOffset: &offsetMinutes})
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("SetUTCOffset", resp.StatusCode(), resp.Body, nil)
}

// ListFollowing returns follow relationships for the given user IDs.
func (c *Client) ListFollowing(userIDs []string) (*FollowingResponse, error) {
	if err := c.requireAuth("ListFollowing"); err != nil {
		return nil, err
	}
	body := struct {
		UserIDs []string `json:"userIds"`
	}{UserIDs: userIDs}
	editor, err := jsonBodyEditor(body)
	if err != nil {
		return nil, fmt.Errorf("ListFollowing: marshaling request: %w", err)
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	apiResp, err := api.ListFollowingWithResponse(context.Background(), editor)
	if err != nil {
		return nil, err
	}
	var result FollowingResponse
	if err := decodeOpenAPIBody("ListFollowing", apiResp.StatusCode(), apiResp.Body, &result); err != nil {
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
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.AutocompleteUserWithResponse(context.Background(), search)
	if err != nil {
		return nil, err
	}
	var result UserAutocompleteResponse
	if err := decodeOpenAPIBody("AutocompleteUsers", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FindUserByEmail looks up a user by email address.
func (c *Client) FindUserByEmail(email string) (*UserProfile, error) {
	if err := c.requireAuth("FindUserByEmail"); err != nil {
		return nil, err
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetUserInfoByEmailWithResponse(context.Background(), &openapi.GetUserInfoByEmailParams{Email: openapitypes.Email(email)})
	if err != nil {
		return nil, err
	}
	var profile UserProfile
	if err := decodeOpenAPIBody("FindUserByEmail", resp.StatusCode(), resp.Body, &profile); err != nil {
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
	editor, err := jsonBodyEditor(body)
	if err != nil {
		return fmt.Errorf("BlockUser: marshaling request: %w", err)
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}
	resp, err := api.BlockUserWithResponse(context.Background(), editor)
	if err != nil {
		return err
	}
	return decodeOpenAPIBody("BlockUser", resp.StatusCode(), resp.Body, nil)
}

// IsUsernameTaken reports whether a username is already in use.
func (c *Client) IsUsernameTaken(username string) (bool, error) {
	api, err := c.openAPI()
	if err != nil {
		return false, err
	}
	resp, err := api.IsUsernameTakenWithResponse(context.Background(), username)
	if err != nil {
		return false, err
	}
	var result UsernameTakenResponse
	if err := decodeOpenAPIBody("IsUsernameTaken", resp.StatusCode(), resp.Body, &result); err != nil {
		return false, err
	}
	return result.Taken, nil
}

// GetUserEmails returns the authenticated user's registered email addresses.
func (c *Client) GetUserEmails() (*UserEmailsResponse, error) {
	if err := c.requireAuth("GetUserEmails"); err != nil {
		return nil, err
	}
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetScannedEmailsWithResponse(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	var result UserEmailsResponse
	if err := decodeOpenAPIBody("GetUserEmails", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserProfile fetches another user's profile and public trips by numeric ID.
func (c *Client) GetUserProfile(userID int) (*ProfileTripsResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetProfileDataByUserIdWithResponse(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	var result ProfileTripsResponse
	if err := decodeOpenAPIBody("GetUserProfile", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserProfileByUsername fetches a user's profile by username.
func (c *Client) GetUserProfileByUsername(username string) (*ProfileTripsResponse, error) {
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}
	resp, err := api.GetProfileByUsernameWithResponse(context.Background(), username)
	if err != nil {
		return nil, err
	}
	var result ProfileTripsResponse
	if err := decodeOpenAPIBody("GetUserProfileByUsername", resp.StatusCode(), resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
