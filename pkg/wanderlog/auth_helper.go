package wanderlog

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// EnsureAuthenticated ensures the client has authentication credentials,
// either from explicit parameters, environment variables, config file, or stored keychain
// If session tokens are available along with email/password, it will use the session tokens
// and fall back to re-login with email/password if the session is invalid
func (c *Client) EnsureAuthenticated(sessionCookie, xsrfToken string) error {
	// Priority order:
	// 1. Explicit credentials (flags)
	// 2. Session tokens from environment variables/config file
	// 3. Stored keychain
	// 4. Email/password from config file (auto-login and save new session)

	// If explicit credentials provided, use them
	if sessionCookie != "" || xsrfToken != "" {
		creds := &AuthCredentials{
			SessionCookie: sessionCookie,
			XSRFToken:     xsrfToken,
		}
		c.SetAuth(creds)
		return nil
	}

	envSession := os.Getenv("WANDERLOG_AUTH_SESSION_COOKIE")
	envXSRF := firstNonEmpty(os.Getenv("WANDERLOG_AUTH_XSRF_TOKEN"), os.Getenv("WANDERLOG_AUTH_SESSION_XSRF_TOKEN"))
	if envSession != "" && envXSRF != "" {
		c.SetAuth(&AuthCredentials{
			SessionCookie: envSession,
			XSRFToken:     envXSRF,
			UserID:        os.Getenv("WANDERLOG_AUTH_USER_ID"),
		})
		c.logger.Debug("Using session credentials from environment variables")
		return nil
	}

	envEmail := os.Getenv("WANDERLOG_AUTH_EMAIL")
	envPassword := os.Getenv("WANDERLOG_AUTH_PASSWORD")
	if envEmail != "" && envPassword != "" {
		c.logger.Debug("Logging in with email/password from environment variables")
		creds, err := c.Login(envEmail, envPassword)
		if err != nil {
			return fmt.Errorf("login with environment credentials failed: %w", err)
		}
		c.SetAuth(creds)
		return nil
	}

	// Try to load session tokens from viper (env vars or config file)
	viperSession := viper.GetString("auth.session.cookie")
	viperXSRF := viper.GetString("auth.session.xsrf_token")
	viperUserID := viper.GetString("auth.session.user_id")

	// Also check if email/password are available for fallback
	viperEmail := viper.GetString("auth.email")
	viperPassword := viper.GetString("auth.password")

	if viperSession != "" && viperXSRF != "" {
		creds := &AuthCredentials{
			SessionCookie: viperSession,
			XSRFToken:     viperXSRF,
			UserID:        viperUserID,
		}
		c.logger.Debug("Using session credentials from config file or environment variables")
		c.SetAuth(creds)
		return nil
	}

	// If we have session but no XSRF token, and we have email/password, re-login
	if viperSession != "" && viperXSRF == "" && viperEmail != "" && viperPassword != "" {
		c.logger.Debug("Session cookie found but no XSRF token, re-authenticating with email/password")
		creds, err := c.Login(viperEmail, viperPassword)
		if err != nil {
			return fmt.Errorf("re-login failed: %w", err)
		}
		c.SetAuth(creds)

		// Save the new session tokens to config file
		if err := SaveCredentialsToConfig(creds, viperEmail, viperPassword); err != nil {
			c.logger.WithError(err).Warn("Failed to update config file with new session tokens")
		} else {
			c.logger.Debug("Updated config file with new XSRF token")
		}

		return nil
	}

	// Try to load from keychain
	if HasStoredCredentials() {
		creds, err := LoadCredentials()
		if err != nil {
			return fmt.Errorf("loading stored credentials: %w", err)
		}
		c.logger.Debug("Using credentials from keychain")
		c.SetAuth(creds)
		return nil
	}

	// If no session tokens available but email/password are, login and save new session
	if viperEmail != "" && viperPassword != "" {
		c.logger.Debug("No valid session found, logging in with email/password from config file")
		creds, err := c.Login(viperEmail, viperPassword)
		if err != nil {
			return fmt.Errorf("login with config credentials failed: %w", err)
		}
		c.SetAuth(creds)

		// Save the new session tokens to config file
		if err := SaveCredentialsToConfig(creds, viperEmail, viperPassword); err != nil {
			c.logger.WithError(err).Warn("Failed to update config file with new session tokens")
		} else {
			c.logger.Debug("Updated config file with new session tokens")
		}

		return nil
	}

	return fmt.Errorf("authentication required - run 'wanderlog login', set credentials in config file, or provide --session and --xsrf flags")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
