package wanderlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "wanderlog-cli"
	userKey     = "auth"
)

// CredentialsFilePath returns the location of the plaintext fallback
// credential store, used when no system keychain backend is available.
// WANDERLOG_CREDENTIALS_FILE overrides the default so headless setups and
// tests can redirect it.
func CredentialsFilePath() string {
	if path := os.Getenv("WANDERLOG_CREDENTIALS_FILE"); path != "" {
		return path
	}
	return filepath.Join(xdg.ConfigHome, "wanderlog", "credentials.json")
}

// keychainDisabled reports whether the operator explicitly forced plaintext
// storage with WANDERLOG_DISABLE_KEYCHAIN=1.
func keychainDisabled() bool {
	return os.Getenv("WANDERLOG_DISABLE_KEYCHAIN") == "1"
}

// keychainUnavailable reports whether err means no OS keychain backend exists
// at all — headless Linux without D-Bus/Secret Service, or an unsupported
// platform — rather than a working keychain failing. Only the former may
// silently fall back to plaintext storage.
func keychainUnavailable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "dbus") ||
		strings.Contains(message, "secret service") ||
		strings.Contains(message, "org.freedesktop.secrets")
}

// SaveCredentials stores authentication credentials in the system keychain,
// falling back to a chmod-0600 plaintext file when no keychain backend exists
// (e.g. headless Linux without D-Bus) or when WANDERLOG_DISABLE_KEYCHAIN=1.
// Only session tokens are ever stored, never the account password.
func SaveCredentials(creds *AuthCredentials) error {
	if err := creds.Validate(); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	// Marshal credentials to JSON for storage
	credsJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}

	if !keychainDisabled() {
		if err := keyring.Set(serviceName, userKey, string(credsJSON)); err != nil {
			if !keychainUnavailable(err) {
				return fmt.Errorf("storing credentials in keychain: %w", err)
			}
			logrus.WithError(err).Warnf(
				"System keychain unavailable; storing session credentials in plaintext at %s",
				CredentialsFilePath(),
			)
			return saveCredentialsFile(credsJSON)
		}
		// The keychain is authoritative again; drop any credentials a previous
		// plaintext fallback left behind so stale sessions cannot resurface.
		if err := os.Remove(CredentialsFilePath()); err != nil && !os.IsNotExist(err) {
			logrus.WithError(err).Warn("Failed to remove superseded plaintext credentials")
		}
		return nil
	}

	return saveCredentialsFile(credsJSON)
}

// LoadCredentials retrieves stored authentication credentials from the system
// keychain, falling back to the plaintext file when the keychain is disabled,
// unavailable, or empty. It returns nil credentials when nothing is stored.
func LoadCredentials() (*AuthCredentials, error) {
	if !keychainDisabled() {
		credsJSON, err := keyring.Get(serviceName, userKey)
		switch {
		case err == nil:
			return parseStoredCredentials(credsJSON)
		case errors.Is(err, keyring.ErrNotFound):
			// Nothing in the keychain; check the plaintext fallback.
		case keychainUnavailable(err):
			logrus.WithError(err).Debug("System keychain unavailable; checking plaintext credential fallback")
			// No keychain backend; check the plaintext fallback.
		default:
			return nil, fmt.Errorf("retrieving credentials from keychain: %w", err)
		}
	}
	return loadCredentialsFile()
}

// DeleteCredentials removes stored credentials from every local backend: the
// system keychain and the plaintext fallback file.
func DeleteCredentials() error {
	var errs []error
	if !keychainDisabled() {
		err := keyring.Delete(serviceName, userKey)
		switch {
		case err == nil, errors.Is(err, keyring.ErrNotFound), keychainUnavailable(err):
			// Nothing stored, or no backend to delete from.
		default:
			errs = append(errs, fmt.Errorf("deleting credentials from keychain: %w", err))
		}
	}
	if err := os.Remove(CredentialsFilePath()); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("deleting plaintext credentials: %w", err))
	}
	return errors.Join(errs...)
}

// HasStoredCredentials checks if credentials are stored in any local backend
func HasStoredCredentials() bool {
	creds, err := LoadCredentials()
	return err == nil && creds != nil
}

// saveCredentialsFile writes credsJSON to the plaintext fallback path with
// owner-only permissions.
func saveCredentialsFile(credsJSON []byte) error {
	path := CredentialsFilePath()
	if err := writeFileAtomically(path, credsJSON); err != nil {
		return fmt.Errorf("storing credentials in plaintext fallback %s: %w", path, err)
	}
	return nil
}

// loadCredentialsFile reads credentials from the plaintext fallback path. It
// returns nil credentials when no file exists.
func loadCredentialsFile() (*AuthCredentials, error) {
	data, err := os.ReadFile(CredentialsFilePath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading plaintext credentials: %w", err)
	}
	return parseStoredCredentials(string(data))
}

func parseStoredCredentials(credsJSON string) (*AuthCredentials, error) {
	// Unmarshal JSON
	var creds AuthCredentials
	if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
		return nil, fmt.Errorf("unmarshaling credentials: %w", err)
	}
	if err := creds.Validate(); err != nil {
		return nil, fmt.Errorf("stored credentials are invalid: %w", err)
	}
	return &creds, nil
}
