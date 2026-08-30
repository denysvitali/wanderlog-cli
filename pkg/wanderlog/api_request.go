package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiDateFormat = "2006-01-02"

type apiResponse struct {
	StatusCode int
	Status     string
	Header     http.Header
	Body       []byte
}

func (c *Client) apiRequest(ctx context.Context, method, path string, query url.Values, body []byte, authenticated bool) (*apiResponse, error) {
	apiURL, err := buildAPIURL(path, query)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL, reader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if authenticated {
		if err := ensureAuthOrigin(req); err != nil {
			return nil, err
		}
		if err := c.addAuthHeaders(req); err != nil {
			return nil, fmt.Errorf("adding auth headers: %w", err)
		}
	} else if c.auth != nil {
		if err := ensureAuthOrigin(req); err != nil {
			return nil, err
		}
		if err := c.addAuthHeaders(req); err != nil {
			return nil, fmt.Errorf("adding optional auth headers: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := readAPIResponseBody(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &apiResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header.Clone(),
		Body:       respBody,
	}, nil
}

func (c *Client) apiJSON(ctx context.Context, method, path string, query url.Values, body any, authenticated bool) (*apiResponse, error) {
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
	}
	return c.apiRequest(ctx, method, path, query, encoded, authenticated)
}

func buildAPIURL(path string, query url.Values) (string, error) {
	var raw string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		raw = path
	} else {
		trimmed := strings.TrimPrefix(path, "/")
		trimmed = strings.TrimPrefix(trimmed, "api/")
		raw = strings.TrimRight(BaseURL, "/") + "/" + trimmed
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parsing API URL: %w", err)
	}
	values := parsed.Query()
	for key, vals := range query {
		for _, value := range vals {
			values.Add(key, value)
		}
	}
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}

func apiQuery(values map[string]string) url.Values {
	query := url.Values{}
	for key, value := range values {
		if value != "" {
			query.Set(key, value)
		}
	}
	return query
}

func decodeAPIBody(opName string, statusCode int, body []byte, out any) error {
	if statusCode < 200 || statusCode >= 300 {
		bodyText := string(body)
		if msg, ok := knownWanderlogServerError(opName, bodyText); ok {
			return fmt.Errorf("%s: HTTP %d: %s", opName, statusCode, msg)
		}
		return fmt.Errorf("%s: HTTP %d: %s", opName, statusCode, truncateForLog(bodyText, 500))
	}
	if failure := explicitAPIFailure(body); failure != "" {
		return fmt.Errorf("%s: %s", opName, failure)
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%s: decoding response: %w", opName, err)
	}
	return nil
}

// decodeAPIBodyPreserveNumbers is used for JSON0 source documents. Standard
// json.Unmarshal converts every number in an interface value to float64, which
// can round integer IDs and make an operation's old value differ from the
// ShareDB document. json.Number preserves the server's exact numeric token.
func decodeAPIBodyPreserveNumbers(opName string, statusCode int, body []byte, out any) error {
	if statusCode < 200 || statusCode >= 300 {
		bodyText := string(body)
		if msg, ok := knownWanderlogServerError(opName, bodyText); ok {
			return fmt.Errorf("%s: HTTP %d: %s", opName, statusCode, msg)
		}
		return fmt.Errorf("%s: HTTP %d: %s", opName, statusCode, truncateForLog(bodyText, 500))
	}
	if failure := explicitAPIFailure(body); failure != "" {
		return fmt.Errorf("%s: %s", opName, failure)
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := decodeJSONPreserveNumbers(body, out); err != nil {
		return fmt.Errorf("%s: decoding response: %w", opName, err)
	}
	return nil
}

func decodeJSONPreserveNumbers(body []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

// explicitAPIFailure detects Wanderlog's HTTP-200 error envelope while still
// allowing older endpoints that omit the success field entirely.
func explicitAPIFailure(body []byte) string {
	if len(bytes.TrimSpace(body)) == 0 {
		return ""
	}
	var envelope struct {
		Success *bool           `json:"success"`
		Error   json.RawMessage `json:"error"`
		Message string          `json:"message"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.Success == nil || *envelope.Success {
		return ""
	}
	if envelope.Message != "" {
		return envelope.Message
	}
	if len(envelope.Error) > 0 && string(envelope.Error) != "null" {
		var message string
		if json.Unmarshal(envelope.Error, &message) == nil && message != "" {
			return message
		}
		var detail struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(envelope.Error, &detail) == nil && detail.Message != "" {
			return detail.Message
		}
	}
	return "API returned success=false"
}

func parseAPIDate(value, fieldName string) (time.Time, error) {
	parsed, err := time.Parse(apiDateFormat, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing %s date: %w", fieldName, err)
	}
	return parsed, nil
}
