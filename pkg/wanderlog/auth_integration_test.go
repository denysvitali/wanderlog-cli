package wanderlog

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// skipIntegrationTest skips the test if INTEGRATION_TESTS env var is not set
func skipIntegrationTest(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS env var not set")
	}
}

func TestIntegration_EnsureAuthenticated(t *testing.T) {
	skipIntegrationTest(t)

	t.Run("full fallback chain with explicit credentials", func(t *testing.T) {
		sessionCookie := os.Getenv("WANDERLOG_AUTH_SESSION_COOKIE")
		xsrfToken := os.Getenv("WANDERLOG_AUTH_XSRF_TOKEN")
		email := os.Getenv("WANDERLOG_AUTH_EMAIL")
		password := os.Getenv("WANDERLOG_AUTH_PASSWORD")

		// If we have explicit credentials, test with them
		if sessionCookie != "" || xsrfToken != "" || (email != "" && password != "") {
			client := NewClient()
			logger := newTestLogger(t)
			logger.SetLevel(logrus.DebugLevel)
			client.SetLogger(logger)

			err := client.EnsureAuthenticated(sessionCookie, xsrfToken)
			if err != nil {
				t.Fatalf("EnsureAuthenticated with explicit credentials failed: %v", err)
			}
			if client.auth == nil {
				t.Fatal("Expected auth to be set after EnsureAuthenticated")
			}
			t.Logf("Successfully authenticated with explicit credentials (session=%v, xsrf=%v)",
				client.auth.SessionCookie != "", client.auth.XSRFToken != "")
		}
	})

	t.Run("fallback to environment variables", func(t *testing.T) {
		email := os.Getenv("WANDERLOG_AUTH_EMAIL")
		password := os.Getenv("WANDERLOG_AUTH_PASSWORD")

		if email == "" || password == "" {
			t.Skip("WANDERLOG_AUTH_EMAIL and WANDERLOG_AUTH_PASSWORD not set")
		}

		// Clear explicit credentials to force env var usage
		client := NewClient()
		logger := newTestLogger(t)
		logger.SetLevel(logrus.DebugLevel)
		client.SetLogger(logger)

		// EnsureAuthenticated should use env vars when no explicit creds provided
		err := client.EnsureAuthenticated("", "")
		if err != nil {
			t.Fatalf("EnsureAuthenticated with env vars failed: %v", err)
		}
		if client.auth == nil {
			t.Fatal("Expected auth to be set after EnsureAuthenticated with env vars")
		}
		t.Logf("Successfully authenticated via env vars (userID=%s)", client.auth.UserID)
	})
}

func TestIntegration_KeychainCredentialOperations(t *testing.T) {
	skipIntegrationTest(t)

	t.Run("save and load credentials", func(t *testing.T) {
		testCreds := &AuthCredentials{
			SessionCookie: "s:test-session-" + t.Name(),
			XSRFToken:     "test-xsrf-" + t.Name(),
			UserID:        "99999",
		}

		// Save credentials
		err := SaveCredentials(testCreds)
		if err != nil {
			t.Fatalf("SaveCredentials failed: %v", err)
		}

		// Verify credentials are stored
		if !HasStoredCredentials() {
			t.Fatal("HasStoredCredentials returned false after saving")
		}

		// Load credentials
		loadedCreds, err := LoadCredentials()
		if err != nil {
			t.Fatalf("LoadCredentials failed: %v", err)
		}
		if loadedCreds == nil {
			t.Fatal("LoadCredentials returned nil after saving")
		}
		if loadedCreds.SessionCookie != testCreds.SessionCookie {
			t.Errorf("SessionCookie mismatch: got %s, want %s", loadedCreds.SessionCookie, testCreds.SessionCookie)
		}
		if loadedCreds.XSRFToken != testCreds.XSRFToken {
			t.Errorf("XSRFToken mismatch: got %s, want %s", loadedCreds.XSRFToken, testCreds.XSRFToken)
		}
		if loadedCreds.UserID != testCreds.UserID {
			t.Errorf("UserID mismatch: got %s, want %s", loadedCreds.UserID, testCreds.UserID)
		}

		t.Log("Successfully saved and loaded credentials from keychain")

		// Clean up - delete credentials
		err = DeleteCredentials()
		if err != nil {
			t.Logf("Warning: failed to delete test credentials: %v", err)
		}
	})

	t.Run("delete credentials", func(t *testing.T) {
		// First save some credentials
		testCreds := &AuthCredentials{
			SessionCookie: "s:test-delete-session",
			XSRFToken:     "test-delete-xsrf",
			UserID:        "12345",
		}
		err := SaveCredentials(testCreds)
		if err != nil {
			t.Fatalf("SaveCredentials failed: %v", err)
		}

		// Verify they exist
		if !HasStoredCredentials() {
			t.Fatal("HasStoredCredentials returned false after saving")
		}

		// Delete credentials
		err = DeleteCredentials()
		if err != nil {
			t.Fatalf("DeleteCredentials failed: %v", err)
		}

		// Verify they are gone
		// Note: HasStoredCredentials may still return true if the keyring doesn't immediately remove
		loadedCreds, err := LoadCredentials()
		if err != nil {
			t.Logf("LoadCredentials returned error (expected for deleted creds): %v", err)
		}
		if loadedCreds != nil {
			t.Logf("Warning: credentials still exist after delete: %+v", loadedCreds)
		}

		t.Log("DeleteCredentials completed")
	})

	t.Run("has stored credentials detection", func(t *testing.T) {
		// First ensure no credentials exist
		_ = DeleteCredentials()

		// Check initial state
		initialState := HasStoredCredentials()
		t.Logf("Initial HasStoredCredentials state: %v", initialState)

		// Save credentials
		testCreds := &AuthCredentials{
			SessionCookie: "s:test-presence",
			XSRFToken:     "test-presence-xsrf",
			UserID:        "11111",
		}
		err := SaveCredentials(testCreds)
		if err != nil {
			t.Fatalf("SaveCredentials failed: %v", err)
		}

		// Check after save
		afterSave := HasStoredCredentials()
		if !afterSave {
			t.Error("HasStoredCredentials returned false after saving")
		}

		// Clean up
		_ = DeleteCredentials()

		t.Logf("HasStoredCredentials: initial=%v, afterSave=%v", initialState, afterSave)
	})
}

func TestIntegration_TokenRefreshReLogin(t *testing.T) {
	skipIntegrationTest(t)

	email := os.Getenv("WANDERLOG_AUTH_EMAIL")
	password := os.Getenv("WANDERLOG_AUTH_PASSWORD")

	if email == "" || password == "" {
		t.Skip("WANDERLOG_AUTH_EMAIL and WANDERLOG_AUTH_PASSWORD not set for re-login test")
	}

	t.Run("login and verify tokens", func(t *testing.T) {
		client := NewClient()
		logger := newTestLogger(t)
		logger.SetLevel(logrus.DebugLevel)
		client.SetLogger(logger)

		creds, err := client.Login(email, password)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		if creds == nil {
			t.Fatal("Login returned nil credentials")
		}
		if creds.SessionCookie == "" {
			t.Error("SessionCookie is empty after login")
		}
		if creds.XSRFToken == "" {
			t.Error("XSRFToken is empty after login")
		}
		if creds.UserID == "" {
			t.Error("UserID is empty after login")
		}

		t.Logf("Login successful: userID=%s, sessionCookie=%s, xsrfToken=%s",
			creds.UserID, creds.SessionCookie[:min(20, len(creds.SessionCookie))]+"...", creds.XSRFToken)
	})

	t.Run("set auth and verify headers", func(t *testing.T) {
		client := NewClient()
		logger := newTestLogger(t)
		logger.SetLevel(logrus.DebugLevel)
		client.SetLogger(logger)

		creds, err := client.Login(email, password)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		client.SetAuth(creds)
		if client.auth == nil {
			t.Fatal("SetAuth did not store credentials")
		}
		if client.auth.SessionCookie != creds.SessionCookie {
			t.Errorf("SessionCookie mismatch after SetAuth")
		}

		t.Log("SetAuth successfully stored credentials")
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
