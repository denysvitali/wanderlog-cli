package wanderlog

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// EnsureAuthenticated configures a complete credential pair from explicit
// flags, environment variables, the system keychain, or the legacy config
// session. Passwords are only accepted from the environment and are never
// loaded from the plaintext config file.
func (c *Client) EnsureAuthenticated(sessionCookie, xsrfToken string) error {
	return c.EnsureAuthenticatedContext(context.Background(), sessionCookie, xsrfToken)
}

// EnsureAuthenticatedContext resolves and verifies credentials using ctx. If
// the server rejects a complete session, it may re-login only with an explicit
// WANDERLOG_AUTH_EMAIL/WANDERLOG_AUTH_PASSWORD pair; config-file passwords are
// never read.
func (c *Client) EnsureAuthenticatedContext(ctx context.Context, sessionCookie, xsrfToken string) error {
	// A failed resolution must never leave previously configured credentials on
	// the client for a caller to accidentally reuse.
	c.SetAuth(nil)

	// Priority order:
	// 1. Explicit credentials (flags)
	// 2. Session tokens from environment variables
	// 3. Stored keychain
	// 4. Legacy session tokens from the config file
	// 5. Email/password from environment variables
	envEmail := os.Getenv("WANDERLOG_AUTH_EMAIL")
	envPassword := os.Getenv("WANDERLOG_AUTH_PASSWORD")
	refreshRejectedSession := func(source string, persistRefresh bool, validationErr error) error {
		c.SetAuth(nil)
		if !errors.Is(validationErr, ErrSessionRejected) {
			return fmt.Errorf("validating %s credentials: %w", source, validationErr)
		}
		if envEmail == "" && envPassword == "" {
			return fmt.Errorf("validating %s credentials: %w", source, validationErr)
		}
		if envEmail == "" || envPassword == "" {
			return fmt.Errorf("%s credentials were rejected and both WANDERLOG_AUTH_EMAIL and WANDERLOG_AUTH_PASSWORD are required for refresh", source)
		}
		c.logger.WithField("source", source).Debug("Session rejected; logging in with environment credentials")
		creds, err := c.LoginContext(ctx, envEmail, envPassword)
		if err != nil {
			return fmt.Errorf("refreshing rejected %s credentials: %w", source, err)
		}
		c.SetAuth(creds)
		if persistRefresh {
			if err := SaveCredentials(creds); err != nil {
				// Authentication succeeded, so a temporarily unavailable keychain
				// must not break the current headless/environment-backed process.
				c.logger.WithError(err).Warn("Session refreshed, but the keychain could not be updated")
			} else {
				c.logger.Debug("Updated refreshed session credentials in the keychain")
			}
		}
		return nil
	}
	useSession := func(source string, creds *AuthCredentials, persistRefresh bool) error {
		if err := creds.Validate(); err != nil {
			return fmt.Errorf("invalid %s credentials: %w", source, err)
		}
		c.SetAuth(creds)
		if err := c.ValidateSessionContext(ctx); err != nil {
			return refreshRejectedSession(source, persistRefresh, err)
		}
		c.logger.WithField("source", source).Debug("Verified session credentials")
		return nil
	}

	// If explicit credentials provided, use them
	if sessionCookie != "" || xsrfToken != "" {
		return useSession("explicit session", &AuthCredentials{
			SessionCookie: sessionCookie,
			XSRFToken:     xsrfToken,
		}, false)
	}

	envSession := os.Getenv("WANDERLOG_AUTH_SESSION_COOKIE")
	envXSRF := firstNonEmpty(os.Getenv("WANDERLOG_AUTH_XSRF_TOKEN"), os.Getenv("WANDERLOG_AUTH_SESSION_XSRF_TOKEN"))
	if envSession != "" || envXSRF != "" {
		return useSession("environment session", &AuthCredentials{
			SessionCookie: envSession,
			XSRFToken:     envXSRF,
			UserID:        os.Getenv("WANDERLOG_AUTH_USER_ID"),
		}, false)
	}

	// Prefer the keychain over the legacy config-file session. Headless and
	// hermetic environments may explicitly disable keychain access.
	var keychainErr error
	if os.Getenv("WANDERLOG_DISABLE_KEYCHAIN") != "1" {
		var creds *AuthCredentials
		creds, keychainErr = LoadCredentials()
		if creds != nil {
			return useSession("keychain session", creds, true)
		}
	}

	// Try to load legacy session tokens from viper (env vars or config file).
	viperSession := viper.GetString("auth.session.cookie")
	viperXSRF := viper.GetString("auth.session.xsrf_token")
	viperUserID := viper.GetString("auth.session.user_id")

	if viperSession != "" || viperXSRF != "" {
		return useSession("config session", &AuthCredentials{
			SessionCookie: viperSession,
			XSRFToken:     viperXSRF,
			UserID:        viperUserID,
		}, false)
	}
	if envEmail != "" || envPassword != "" {
		if envEmail == "" || envPassword == "" {
			return fmt.Errorf("both WANDERLOG_AUTH_EMAIL and WANDERLOG_AUTH_PASSWORD are required")
		}
		c.logger.Debug("Logging in with email/password from environment variables")
		creds, err := c.LoginContext(ctx, envEmail, envPassword)
		if err != nil {
			return fmt.Errorf("login with environment credentials failed: %w", err)
		}
		c.SetAuth(creds)
		return nil
	}
	if keychainErr != nil {
		return fmt.Errorf("loading stored credentials: %w", keychainErr)
	}

	return fmt.Errorf("authentication required - run 'wanderlog login', set complete session credentials, or provide --session and --xsrf flags")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
