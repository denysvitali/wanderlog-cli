package models

import "encoding/json"

// GlobalConfig is a passthrough of /api/config/globalConfig. The bundle accesses
// many nested keys; rather than enumerate them here we keep the raw JSON.
type GlobalConfig struct {
	Success bool            `json:"success"`
	Config  json.RawMessage `json:"config,omitempty"`
	Raw     json.RawMessage `json:"-"`
}

// SessionStore represents the client-wide session store.
type SessionStore struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// SessionStoreRequest sets a single key in the session store.
type SessionStoreRequest struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// SessionPreferences represents /api/sessionStore/preferences/{locale}.
type SessionPreferences struct {
	Success     bool            `json:"success"`
	Preferences json.RawMessage `json:"preferences,omitempty"`
}
