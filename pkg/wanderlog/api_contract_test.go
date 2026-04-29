package wanderlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type requestContract struct {
	Name           string         `json:"name"`
	Method         string         `json:"method"`
	Path           string         `json:"path"`
	ReferencePath string         `json:"referencePath"`
	Auth           bool           `json:"auth"`
	Query          map[string]any `json:"query,omitempty"`
	Body           any            `json:"body,omitempty"`
}

type capturedAPIRequest struct {
	Method string
	Path   string
	Query  map[string]any
	Body   any
	Header http.Header
}

func TestAPIRequestContracts(t *testing.T) {
	contracts := loadRequestContracts(t)
	referenceEndpoints := loadReferenceEndpoints(t)

	runners := map[string]func(*Client) error{
		"Login": func(c *Client) error {
			_, err := c.Login("user@example.com", "secret")
			return err
		},
		"GetTrip": func(c *Client) error {
			_, err := c.GetTrip("trip-key")
			return err
		},
		"GetTripSections": func(c *Client) error {
			_, err := c.GetTripSections("trip-key")
			return err
		},
		"GetPlaceDetails": func(c *Client) error {
			_, err := c.GetPlaceDetails("place-123")
			return err
		},
		"GetFlightStops": func(c *Client) error {
			_, err := c.GetFlightStops("244", "MU", "2026-05-11")
			return err
		},
		"CreateTrip": func(c *Client) error {
			_, err := c.CreateTrip(CreateTripRequest{
				Title:     "API Contract Trip",
				GeoIDs:    []int{123},
				StartDate: "2026-06-01",
				EndDate:   "2026-06-04",
			})
			return err
		},
		"AddPlace": func(c *Client) error {
			return c.AddPlace("trip-key", 7, AddPlaceRequest{
				Place: AddPlaceInfo{Name: "Tokyo Station", PlaceID: "place-123"},
			})
		},
		"RemovePlace": func(c *Client) error {
			return c.RemovePlace("trip-key", 7, 123)
		},
		"ApplyOperations": func(c *Client) error {
			return c.ApplyOperations("trip-key", []Operation{})
		},
		"SetLike": func(c *Client) error {
			return c.SetLike("trip-key", true)
		},
	}

	for _, contract := range contracts {
		contract := contract
		t.Run(contract.Name, func(t *testing.T) {
			if !hasReferenceEndpoint(referenceEndpoints, contract.ReferencePath) {
				t.Fatalf("reference endpoint %q not found", contract.ReferencePath)
			}

			runner := runners[contract.Name]
			if runner == nil {
				t.Fatalf("missing runner for %s", contract.Name)
			}

			var captured *capturedAPIRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = captureRequest(t, r)
				if contract.Name == "Login" {
					http.SetCookie(w, &http.Cookie{Name: "connect.sid", Value: "session"})
					http.SetCookie(w, &http.Cookie{Name: "XSRF-TOKEN", Value: "xsrf"})
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(responseForContract(contract.Name)))
			}))
			defer server.Close()

			oldBaseURL := BaseURL
			BaseURL = server.URL + "/api"
			t.Cleanup(func() { BaseURL = oldBaseURL })

			client := NewClient()
			if contract.Auth {
				client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf", UserID: "1"})
			}

			if err := runner(client); err != nil {
				t.Fatalf("%s returned error: %v", contract.Name, err)
			}
			if captured == nil {
				t.Fatal("no request captured")
			}

			if captured.Method != contract.Method {
				t.Fatalf("method mismatch: got %s want %s", captured.Method, contract.Method)
			}
			if captured.Path != contract.Path {
				t.Fatalf("path mismatch: got %s want %s", captured.Path, contract.Path)
			}
			if !reflect.DeepEqual(captured.Query, contract.Query) {
				t.Fatalf("query mismatch:\ngot  %#v\nwant %#v", captured.Query, contract.Query)
			}
			if !reflect.DeepEqual(captured.Body, contract.Body) {
				t.Fatalf("body mismatch:\ngot  %#v\nwant %#v", captured.Body, contract.Body)
			}
			if contract.Auth {
				if got := captured.Header.Get("X-XSRF-TOKEN"); got != "xsrf" {
					t.Fatalf("missing X-XSRF-TOKEN header, got %q", got)
				}
				if got := captured.Header.Get("Cookie"); got == "" {
					t.Fatal("missing auth cookies")
				}
			}
		})
	}
}

func loadRequestContracts(t *testing.T) []requestContract {
	t.Helper()
	var contracts []requestContract
	readJSONFixture(t, "artifacts/api-contracts/go_request_contracts.json", &contracts)
	return contracts
}

func loadReferenceEndpoints(t *testing.T) map[string]bool {
	t.Helper()
	var extracted struct {
		EndpointStrings []struct {
			URL string `json:"url"`
		} `json:"endpointStrings"`
		WrappedAxios []struct {
			URL string `json:"url"`
		} `json:"wrappedAxios"`
	}
	readJSONFixture(t, "artifacts/api-contracts/reference_calls.json", &extracted)
	endpoints := map[string]bool{}
	for _, item := range extracted.EndpointStrings {
		if item.URL != "" {
			endpoints[item.URL] = true
		}
	}
	for _, item := range extracted.WrappedAxios {
		if item.URL != "" {
			endpoints[item.URL] = true
		}
	}
	return endpoints
}

func readJSONFixture(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", path))
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
}

func hasReferenceEndpoint(endpoints map[string]bool, expected string) bool {
	if endpoints[expected] {
		return true
	}
	for endpoint := range endpoints {
		if len(expected) > len(endpoint) && expected[:len(endpoint)] == endpoint {
			return true
		}
	}
	return false
}

func captureRequest(t *testing.T, r *http.Request) *capturedAPIRequest {
	t.Helper()
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}
	var parsedBody any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &parsedBody); err != nil {
			t.Fatalf("parsing request body %q: %v", string(rawBody), err)
		}
	}
	query := map[string]any{}
	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			query[key] = values[0]
		} else {
			copied := make([]string, len(values))
			copy(copied, values)
			query[key] = copied
		}
	}
	if len(query) == 0 {
		query = nil
	}
	return &capturedAPIRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  query,
		Body:   parsedBody,
		Header: r.Header.Clone(),
	}
}

func responseForContract(name string) string {
	switch name {
	case "Login":
		return `{"success":true,"user":{"id":1,"email":"user@example.com","name":"User","username":"user"}}`
	case "GetPlaceDetails":
		return `{"success":true,"data":{"details":{"name":"Tokyo Station","place_id":"place-123","geometry":{"location":{"lat":35.6812,"lng":139.7671}}},"cardData":{"placeId":"place-123"}}}`
	case "GetTripSections":
		return `{"success":true,"data":[]}`
	case "CreateTrip":
		return `{"success":true,"tripPlan":{"id":1,"key":"trip-key","title":"API Contract Trip"}}`
	case "SetLike":
		return `{"success":true,"data":true}`
	default:
		return `{"success":true,"tripPlan":{"id":1,"key":"trip-key","title":"API Contract Trip","likeCount":0},"resources":{},"data":[]}`
	}
}
