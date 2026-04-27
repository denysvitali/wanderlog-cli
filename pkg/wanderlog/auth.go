package wanderlog

import (
	"context"
	"fmt"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
	openapi_types "github.com/oapi-codegen/runtime/types"
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
	api, err := c.openAPI()
	if err != nil {
		return nil, err
	}

	resp, err := api.LoginWithResponse(context.Background(), openapi.LoginJSONRequestBody{
		Email:    openapi_types.Email(email),
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("making login request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode(), resp.Status())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("decoding login response: empty JSON response")
	}

	loginResp := resp.JSON200
	if loginResp.Success == nil || !*loginResp.Success {
		return nil, fmt.Errorf("login failed: invalid credentials")
	}
	if loginResp.User == nil || loginResp.User.Id == nil {
		return nil, fmt.Errorf("login failed: user id not found in response")
	}

	// Extract session cookie and XSRF token from response headers
	var sessionCookie, xsrfToken string
	for _, cookie := range resp.HTTPResponse.Cookies() {
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
		"userID":   *loginResp.User.Id,
		"username": derefString(loginResp.User.Username),
	}).Info("Successfully authenticated")

	return &AuthCredentials{
		SessionCookie: sessionCookie,
		XSRFToken:     xsrfToken,
		UserID:        fmt.Sprintf("%d", *loginResp.User.Id),
	}, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
