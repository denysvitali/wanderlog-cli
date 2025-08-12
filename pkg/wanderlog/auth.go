package wanderlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AuthCredentials holds authentication information
type AuthCredentials struct {
	SessionCookie string
	XSRFToken     string
	UserID        string
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
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("marshaling login request: %w", err)
	}

	req, err := http.NewRequest("POST", BaseURL+"/user/login", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("creating login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("decoding login response: %w", err)
	}

	if !loginResp.Success {
		return nil, fmt.Errorf("login failed: invalid credentials")
	}

	// Extract session cookie and XSRF token from response headers
	var sessionCookie, xsrfToken string
	for _, cookie := range resp.Cookies() {
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

// SetAuth configures the client with authentication credentials
func (c *Client) SetAuth(creds *AuthCredentials) {
	c.auth = creds
}

// AddAuthHeaders adds authentication headers to a request
func (c *Client) addAuthHeaders(req *http.Request) error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated - call Login() first")
	}

	// Add session cookie
	if c.auth.SessionCookie != "" {
		req.AddCookie(&http.Cookie{
			Name:  "connect.sid",
			Value: c.auth.SessionCookie,
		})
	}

	// Add XSRF token header
	if c.auth.XSRFToken != "" {
		req.Header.Set("X-XSRF-TOKEN", c.auth.XSRFToken)
		req.AddCookie(&http.Cookie{
			Name:  "XSRF-TOKEN",
			Value: c.auth.XSRFToken,
		})
	}

	return nil
}
