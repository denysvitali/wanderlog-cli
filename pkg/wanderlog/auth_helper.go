package wanderlog

import "fmt"

// EnsureAuthenticated ensures the client has authentication credentials,
// either from explicit parameters, environment variables, or stored keychain
func (c *Client) EnsureAuthenticated(sessionCookie, xsrfToken string) error {
	// If explicit credentials provided, use them
	if sessionCookie != "" || xsrfToken != "" {
		creds := &AuthCredentials{
			SessionCookie: sessionCookie,
			XSRFToken:     xsrfToken,
		}
		c.SetAuth(creds)
		return nil
	}

	// Try to load from keychain
	if HasStoredCredentials() {
		creds, err := LoadCredentials()
		if err != nil {
			return fmt.Errorf("loading stored credentials: %w", err)
		}
		c.SetAuth(creds)
		return nil
	}

	return fmt.Errorf("authentication required - run 'wanderlog login' or provide --session and --xsrf flags")
}