package wanderlog

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "wanderlog-cli"
	userKey     = "auth"
)

// SaveCredentials securely stores authentication credentials in the system keychain
func SaveCredentials(creds *AuthCredentials) error {
	if creds == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	// Marshal credentials to JSON for storage
	credsJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}

	// Store in system keychain
	err = keyring.Set(serviceName, userKey, string(credsJSON))
	if err != nil {
		return fmt.Errorf("storing credentials in keychain: %w", err)
	}

	return nil
}

// LoadCredentials retrieves authentication credentials from the system keychain
func LoadCredentials() (*AuthCredentials, error) {
	// Get from system keychain
	credsJSON, err := keyring.Get(serviceName, userKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil // No credentials stored
		}
		return nil, fmt.Errorf("retrieving credentials from keychain: %w", err)
	}

	// Unmarshal JSON
	var creds AuthCredentials
	if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
		return nil, fmt.Errorf("unmarshaling credentials: %w", err)
	}

	return &creds, nil
}

// DeleteCredentials removes stored credentials from the system keychain
func DeleteCredentials() error {
	err := keyring.Delete(serviceName, userKey)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("deleting credentials from keychain: %w", err)
	}
	return nil
}

// HasStoredCredentials checks if credentials are stored in the keychain
func HasStoredCredentials() bool {
	_, err := keyring.Get(serviceName, userKey)
	return err == nil
}