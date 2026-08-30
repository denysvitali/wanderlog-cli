package wanderlog

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestAPIErrorHTTPMetadataAndBounds(t *testing.T) {
	body := []byte(`{"error":{"message":"temporarily unavailable"},"padding":"` + strings.Repeat("界", MaxAPIErrorBodyBytes) + `"}`)
	err := decodeAPIBody("ListTrips", http.StatusServiceUnavailable, body, nil)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Operation != "ListTrips" {
		t.Fatalf("Operation = %q, want ListTrips", apiErr.Operation)
	}
	if apiErr.HTTPStatus != http.StatusServiceUnavailable {
		t.Fatalf("HTTPStatus = %d, want %d", apiErr.HTTPStatus, http.StatusServiceUnavailable)
	}
	if !apiErr.Retryable {
		t.Fatal("Retryable = false, want true for HTTP 503")
	}
	if apiErr.Message != "temporarily unavailable" {
		t.Fatalf("Message = %q", apiErr.Message)
	}
	if len(apiErr.Body) > MaxAPIErrorBodyBytes || !utf8.ValidString(apiErr.Body) {
		t.Fatalf("bounded body has %d bytes and valid UTF-8 = %v", len(apiErr.Body), utf8.ValidString(apiErr.Body))
	}
	if len(err.Error()) > len("ListTrips: HTTP 503: ")+MaxAPIErrorMessageBytes {
		t.Fatalf("Error() is unexpectedly large: %d bytes", len(err.Error()))
	}
}

func TestAPIErrorRetryability(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{status: http.StatusBadRequest, want: false},
		{status: http.StatusRequestTimeout, want: true},
		{status: http.StatusTooEarly, want: true},
		{status: http.StatusTooManyRequests, want: true},
		{status: http.StatusInternalServerError, want: true},
	}
	for _, test := range tests {
		err := apiHTTPError("Operation", test.status, []byte(http.StatusText(test.status)))
		if err.Retryable != test.want {
			t.Errorf("HTTP %d Retryable = %v, want %v", test.status, err.Retryable, test.want)
		}
	}
}

func TestAPIErrorForSuccessEnvelopeAndDecodeFailure(t *testing.T) {
	t.Run("success false", func(t *testing.T) {
		body := []byte(`{"success":false,"message":"denied"}`)
		err := decodeAPIBody("GetTrip", http.StatusOK, body, nil)
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("error type = %T, want *APIError", err)
		}
		if apiErr.HTTPStatus != http.StatusOK || apiErr.Retryable || apiErr.Message != "denied" || apiErr.Body != string(body) {
			t.Fatalf("APIError = %#v", apiErr)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		var result map[string]any
		err := decodeAPIBody("GetTrip", http.StatusOK, []byte(`{`), &result)
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("error type = %T, want *APIError", err)
		}
		var syntaxErr *json.SyntaxError
		if !errors.As(err, &syntaxErr) {
			t.Fatalf("underlying error = %v, want json.SyntaxError", err)
		}
	})
}

func TestAPIRequestTransportErrorIsStructured(t *testing.T) {
	client := NewClient()
	transportErr := &net.DNSError{Err: "temporary failure", Name: "wanderlog.test", IsTemporary: true}
	client.httpClient.Transport = apiErrorRoundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportErr
	})

	_, err := client.apiRequest(t.Context(), http.MethodGet, "tripPlans", nil, nil, false)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Operation != "GET tripPlans" || apiErr.HTTPStatus != 0 || !apiErr.Retryable {
		t.Fatalf("APIError = %#v", apiErr)
	}
	if !errors.Is(err, transportErr) {
		t.Fatal("transport cause is not preserved")
	}
}

func TestAPIRequestRejectsNilContextWithStructuredError(t *testing.T) {
	client := NewClient()
	//nolint:staticcheck // Explicitly verifies the defensive nil-context contract.
	_, err := client.apiRequest(nil, http.MethodPost, "tripPlans", nil, nil, false)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Operation != "POST tripPlans" || apiErr.HTTPStatus != 0 || apiErr.Retryable {
		t.Fatalf("APIError = %#v", apiErr)
	}
	if !strings.Contains(apiErr.Message, "nil Context") {
		t.Fatalf("Message = %q, want nil-context detail", apiErr.Message)
	}
}

func TestDoAPIHTTPErrorIsStructuredAndRedactsQueryFromOperation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"slow down"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	status, body, err := client.DoAPI(http.MethodGet, "limited?token=secret", nil, nil, false)
	if status != http.StatusTooManyRequests || string(body) != `{"message":"slow down"}` {
		t.Fatalf("status, body = %d, %q", status, body)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Operation != "GET limited" || apiErr.HTTPStatus != http.StatusTooManyRequests || !apiErr.Retryable || apiErr.Message != "slow down" {
		t.Fatalf("APIError = %#v", apiErr)
	}
	if strings.Contains(apiErr.Error(), "secret") {
		t.Fatalf("error exposed query value: %v", apiErr)
	}
}

type apiErrorRoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn apiErrorRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
