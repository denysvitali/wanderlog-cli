package models

import "encoding/json"

// UserProfile mirrors the /api/user response. Fields beyond the common ones are
// passed through via Raw so callers can still inspect everything the API sent.
type UserProfile struct {
	ID                  int             `json:"id"`
	Email               string          `json:"email,omitempty"`
	Name                string          `json:"name,omitempty"`
	Username            string          `json:"username,omitempty"`
	ProfilePictureKey   string          `json:"profilePictureKey,omitempty"`
	Bio                 string          `json:"bio,omitempty"`
	Location            string          `json:"location,omitempty"`
	VisitGeosCount      int             `json:"visitGeosCount,omitempty"`
	CountriesCount      int             `json:"countriesCount,omitempty"`
	IsProUser           bool            `json:"isProUser,omitempty"`
	ShowProfileProBadge bool            `json:"showProfileProBadge,omitempty"`
	UTCOffset           *int            `json:"utcOffset,omitempty"`
	Raw                 json.RawMessage `json:"-"`
}

// UpdateUserRequest is the generic payload for POST /api/user; only non-empty
// fields are included in the JSON sent to the server.
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Bio      string `json:"bio,omitempty"`
	Location string `json:"location,omitempty"`
}

// NotificationsResponse represents the paged notification inbox.
type NotificationsResponse struct {
	Success    bool           `json:"success"`
	Data       []Notification `json:"data"`
	NextOffset *int           `json:"nextOffset,omitempty"`
}

// Notification is intentionally loose: the Wanderlog bundle models this as a
// discriminated union with many variants (likes, comments, invites, etc.).
type Notification struct {
	ID        string          `json:"id"`
	Type      string          `json:"type,omitempty"`
	CreatedAt string          `json:"createdAt,omitempty"`
	Read      bool            `json:"read,omitempty"`
	Title     string          `json:"title,omitempty"`
	Body      string          `json:"body,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

// NotificationSettings is a passthrough — the bundle treats this as a dynamic
// map of setting keys to booleans/strings.
type NotificationSettings struct {
	Success  bool            `json:"success"`
	Settings json.RawMessage `json:"notificationSettings,omitempty"`
}

// UpdateNotificationSettingsRequest is the POST payload.
type UpdateNotificationSettingsRequest struct {
	NotificationSettings json.RawMessage `json:"notificationSettings"`
}

// KeyValueResponse wraps the polymorphic value returned by /api/user/keyValue/{key}.
type KeyValueResponse struct {
	Success bool            `json:"success"`
	Value   json.RawMessage `json:"value,omitempty"`
}

// FollowingResponse pairs each requested userId with whether the current user follows them.
type FollowingResponse struct {
	Success   bool            `json:"success"`
	Following map[string]bool `json:"following,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

// UserAutocompleteResponse is the suggestion list for /api/user/autocomplete/{q}.
type UserAutocompleteResponse struct {
	Success bool               `json:"success"`
	Users   []UserAutocomplete `json:"users"`
}

type UserAutocomplete struct {
	ID                int    `json:"id"`
	Username          string `json:"username,omitempty"`
	Name              string `json:"name,omitempty"`
	ProfilePictureKey string `json:"profilePictureKey,omitempty"`
}

// UserEmailsResponse is the response from /api/user/emails.
type UserEmailsResponse struct {
	Success bool        `json:"success"`
	Emails  []UserEmail `json:"emails"`
}

type UserEmail struct {
	Email     string `json:"email"`
	Verified  bool   `json:"verified,omitempty"`
	IsPrimary bool   `json:"isPrimary,omitempty"`
}

// UsernameTakenResponse is the response from /api/user/isUsernameTaken/{username}.
type UsernameTakenResponse struct {
	Success bool `json:"success"`
	Taken   bool `json:"taken"`
}

// ProfileTripsResponse is the response from /api/tripPlans/profile/{userId|byUsername/*}.
type ProfileTripsResponse struct {
	Success   bool            `json:"success"`
	Profile   json.RawMessage `json:"profile,omitempty"`
	TripPlans json.RawMessage `json:"tripPlans,omitempty"`
	Deleted   json.RawMessage `json:"deleted,omitempty"`
	VisitGeos json.RawMessage `json:"visitGeos,omitempty"`
}
