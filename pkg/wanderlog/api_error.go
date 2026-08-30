package wanderlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	// MaxAPIErrorOperationBytes bounds endpoint names and raw request paths.
	MaxAPIErrorOperationBytes = 256
	// MaxAPIErrorMessageBytes limits the human-readable detail included in an
	// APIError and its Error string.
	MaxAPIErrorMessageBytes = 500
	// MaxAPIErrorBodyBytes limits the response excerpt retained for callers.
	// Successful response bodies continue to use MaxAPIResponseBodyBytes.
	MaxAPIErrorBodyBytes = 4 << 10
)

// APIError is a machine-inspectable failure returned by a Wanderlog API call.
// HTTPStatus is zero when no response was received. Body is always a bounded
// response excerpt and may be empty for transport or local decoding errors.
type APIError struct {
	Operation  string
	HTTPStatus int
	Retryable  bool
	Message    string
	Body       string
	Cause      error
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	prefix := e.Operation
	if prefix == "" {
		prefix = "Wanderlog API"
	}
	if e.HTTPStatus != 0 {
		return fmt.Sprintf("%s: HTTP %d: %s", prefix, e.HTTPStatus, e.Message)
	}
	return fmt.Sprintf("%s: %s", prefix, e.Message)
}

// Unwrap exposes the underlying transport or decoding failure, when present.
func (e *APIError) Unwrap() error { return e.Cause }

func newAPIError(operation string, status int, message string, body []byte, cause error) *APIError {
	message = boundedAPIErrorText(message, MaxAPIErrorMessageBytes)
	if message == "" {
		if status != 0 {
			message = http.StatusText(status)
		}
		if message == "" {
			message = "API request failed"
		}
	}
	return &APIError{
		Operation:  boundedAPIErrorText(operation, MaxAPIErrorOperationBytes),
		HTTPStatus: status,
		Retryable:  retryableAPIError(status, cause),
		Message:    message,
		Body:       boundedAPIErrorText(string(body), MaxAPIErrorBodyBytes),
		Cause:      cause,
	}
}

func apiRequestOperation(method, path string) string {
	endpoint := path
	if parsed, err := url.Parse(path); err == nil {
		endpoint = parsed.EscapedPath()
	}
	if endpoint == "" {
		endpoint = "/"
	}
	return method + " " + endpoint
}

func retryableAPIError(status int, cause error) bool {
	if status == http.StatusRequestTimeout || status == http.StatusTooEarly || status == http.StatusTooManyRequests {
		return true
	}
	if status >= 500 && status <= 599 {
		return true
	}
	if status != 0 || cause == nil {
		return false
	}
	var networkError net.Error
	return errors.As(cause, &networkError)
}

func boundedAPIErrorText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return strings.Repeat(".", limit)
	}
	value = value[:limit-3]
	for len(value) > 0 && !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value + "..."
}

func apiHTTPError(operation string, status int, body []byte) *APIError {
	bodyText := string(body)
	message := bodyText
	if known, ok := knownWanderlogServerError(operation, bodyText); ok {
		message = known
	} else if extracted := apiErrorMessage(body); extracted != "" {
		message = extracted
	}
	return newAPIError(operation, status, message, body, nil)
}

// apiErrorMessage extracts common message shapes without requiring a
// particular success-envelope variant.
func apiErrorMessage(body []byte) string {
	var envelope struct {
		Error    any    `json:"error"`
		Message  string `json:"message"`
		Messages []any  `json:"messages"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return ""
	}
	if envelope.Message != "" {
		return envelope.Message
	}
	if len(envelope.Messages) > 0 {
		if message := apiMessageValue(envelope.Messages[0]); message != "" {
			return message
		}
	}
	return apiMessageValue(envelope.Error)
}

func apiMessageValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case map[string]any:
		message, _ := typed["message"].(string)
		return message
	default:
		return ""
	}
}
