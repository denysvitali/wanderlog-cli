package wanderlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// doJSON performs an authenticated JSON request against the Wanderlog API and
// decodes the response into out. If out is nil the body is discarded. When
// requireAuth is true and no credentials are set, it returns early with an
// auth error. The opNameForErr is used in error messages.
func (c *Client) doJSON(method, path string, body any, out any, requireAuth bool, opNameForErr string) error {
	url := BaseURL + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("%s: marshaling body: %w", opNameForErr, err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("%s: creating request: %w", opNameForErr, err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if requireAuth {
		if c.auth == nil {
			return fmt.Errorf("%s: authentication required", opNameForErr)
		}
		if err := c.addAuthHeaders(req); err != nil {
			return fmt.Errorf("%s: adding auth headers: %w", opNameForErr, err)
		}
	} else if c.auth != nil {
		_ = c.addAuthHeaders(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: request failed: %w", opNameForErr, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: reading response: %w", opNameForErr, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s: HTTP %d: %s", opNameForErr, resp.StatusCode, truncateForLog(string(respBody), 500))
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("%s: decoding response: %w", opNameForErr, err)
	}
	return nil
}

// doRaw is like doJSON but returns the raw response bytes (useful for
// endpoints like expensesAsCSV that don't return JSON).
func (c *Client) doRaw(method, path string, body any, requireAuth bool, opNameForErr string) ([]byte, error) {
	url := BaseURL + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("%s: marshaling body: %w", opNameForErr, err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s: creating request: %w", opNameForErr, err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if requireAuth {
		if c.auth == nil {
			return nil, fmt.Errorf("%s: authentication required", opNameForErr)
		}
		if err := c.addAuthHeaders(req); err != nil {
			return nil, fmt.Errorf("%s: adding auth headers: %w", opNameForErr, err)
		}
	} else if c.auth != nil {
		_ = c.addAuthHeaders(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", opNameForErr, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: reading response: %w", opNameForErr, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, fmt.Errorf("%s: HTTP %d: %s", opNameForErr, resp.StatusCode, truncateForLog(string(respBody), 500))
	}

	return respBody, nil
}

func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
