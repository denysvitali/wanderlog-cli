package wanderlog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrSessionRejected identifies a credential pair that the Wanderlog server no
// longer accepts. Callers can distinguish this from network/server failures and
// decide whether an explicitly supplied login fallback is appropriate.
var ErrSessionRejected = errors.New("wanderlog session rejected")

// AuthCredentials holds authentication information
type AuthCredentials struct {
	SessionCookie string
	XSRFToken     string
	UserID        string
}

// Validate rejects incomplete credentials. Wanderlog sessions require both the
// session cookie and its matching XSRF token; treating either one alone as an
// authenticated session leads to misleading status and failed mutations.
func (c *AuthCredentials) Validate() error {
	if c == nil {
		return fmt.Errorf("credentials are missing")
	}
	if c.SessionCookie == "" {
		return fmt.Errorf("session cookie is missing")
	}
	if c.XSRFToken == "" {
		return fmt.Errorf("XSRF token is missing")
	}
	return nil
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response from login
type LoginResponse struct {
	Success bool `json:"success"`
	User    struct {
		ID       int    `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"user"`
}

// Login authenticates with the Wanderlog API
func (c *Client) Login(email, password string) (*AuthCredentials, error) {
	return c.LoginContext(context.Background(), email, password)
}

// LoginContext authenticates with Wanderlog using ctx. Canceling ctx cancels
// the in-flight login request.
func (c *Client) LoginContext(ctx context.Context, email, password string) (*AuthCredentials, error) {
	resp, err := c.apiJSON(ctx, http.MethodPost, "user/login", nil, LoginRequest{
		Email:    email,
		Password: password,
	}, false)
	if err != nil {
		return nil, fmt.Errorf("making login request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, resp.Status)
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(resp.Body, &loginResp); err != nil {
		return nil, fmt.Errorf("decoding login response: %w", err)
	}

	if !loginResp.Success {
		return nil, fmt.Errorf("login failed: invalid credentials")
	}
	if loginResp.User.ID == 0 {
		return nil, fmt.Errorf("login failed: user id not found in response")
	}

	// Extract session cookie and XSRF token from response headers
	var sessionCookie, xsrfToken string
	for _, cookie := range (&http.Response{Header: resp.Header}).Cookies() {
		switch cookie.Name {
		case "connect.sid":
			sessionCookie = cookie.Value
		case "XSRF-TOKEN":
			xsrfToken = cookie.Value
		}
	}

	if sessionCookie == "" {
		return nil, fmt.Errorf("session cookie not found in response")
	}
	if xsrfToken == "" {
		return nil, fmt.Errorf("XSRF token not found in response")
	}

	c.logger.WithFields(map[string]interface{}{
		"userID":   loginResp.User.ID,
		"username": loginResp.User.Username,
	}).Info("Successfully authenticated")

	return &AuthCredentials{
		SessionCookie: sessionCookie,
		XSRFToken:     xsrfToken,
		UserID:        fmt.Sprintf("%d", loginResp.User.ID),
	}, nil
}

// ValidateSessionContext verifies the configured session against the current
// user endpoint. A syntactically complete credential pair is not considered
// authenticated until the server returns a concrete user identity.
func (c *Client) ValidateSessionContext(ctx context.Context) error {
	if err := c.auth.Validate(); err != nil {
		return fmt.Errorf("validating session: %w", err)
	}
	resp, err := c.apiRequest(ctx, http.MethodGet, "user", nil, nil, true)
	if err != nil {
		return fmt.Errorf("validating session: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w: HTTP %d", ErrSessionRejected, resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("validating session: HTTP %d: %s", resp.StatusCode, truncateForLog(string(resp.Body), 500))
	}
	var profile struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Body, &profile); err != nil {
		return fmt.Errorf("validating session response: %w", err)
	}
	if profile.ID == 0 {
		return fmt.Errorf("%w: current user identity missing", ErrSessionRejected)
	}
	return nil
}

// SetAuth configures the client with authentication credentials
func (c *Client) SetAuth(creds *AuthCredentials) {
	c.auth = creds
}

// AddAuthHeaders adds authentication headers to a request
func (c *Client) addAuthHeaders(req *http.Request) error {
	if err := c.auth.Validate(); err != nil {
		return fmt.Errorf("not authenticated: %w", err)
	}

	// Add session cookie
	req.AddCookie(&http.Cookie{
		Name:  "connect.sid",
		Value: c.auth.SessionCookie,
	})

	// Add XSRF token header
	req.Header.Set("X-XSRF-TOKEN", c.auth.XSRFToken)
	req.AddCookie(&http.Cookie{
		Name:  "XSRF-TOKEN",
		Value: c.auth.XSRFToken,
	})

	return nil
}
