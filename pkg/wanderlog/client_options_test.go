package wanderlog

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestClientSnapshotsLegacyBaseURL(t *testing.T) {
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"server":"first"}`)
	}))
	defer first.Close()
	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"server":"second"}`)
	}))
	defer second.Close()

	previous := BaseURL
	BaseURL = first.URL
	client := NewClient()
	BaseURL = second.URL
	t.Cleanup(func() { BaseURL = previous })

	_, body, err := client.DoAPI(http.MethodGet, "/snapshot", nil, nil, false)
	if err != nil {
		t.Fatalf("DoAPI: %v", err)
	}
	if got := string(body); !strings.Contains(got, `"first"`) {
		t.Fatalf("client followed a later global BaseURL change: %s", got)
	}
}

func TestWithBaseURLIsolatesClients(t *testing.T) {
	newServer := func(name string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/trips" {
				t.Errorf("unexpected path for %s: %q", name, r.URL.Path)
			}
			if got := r.URL.Query().Get("page"); got != "2" {
				t.Errorf("unexpected query for %s: %q", name, got)
			}
			_, _ = io.WriteString(w, name)
		}))
	}
	first := newServer("first")
	defer first.Close()
	second := newServer("second")
	defer second.Close()

	clients := []*Client{
		NewClient(WithBaseURL(first.URL + "/api/")),
		NewClient(WithBaseURL(second.URL + "/api")),
	}
	for index, client := range clients {
		apiURL, err := client.buildAPIURL("trips", map[string][]string{"page": {"2"}})
		if err != nil {
			t.Fatalf("buildAPIURL: %v", err)
		}
		status, body, err := client.DoAPI(http.MethodGet, apiURL, nil, nil, false)
		if err != nil {
			t.Fatalf("client %d DoAPI: %v", index, err)
		}
		if status != http.StatusOK {
			t.Fatalf("client %d status = %d", index, status)
		}
		want := []string{"first", "second"}[index]
		if got := string(body); got != want {
			t.Fatalf("client %d body = %q, want %q", index, got, want)
		}
	}
}

func TestWithHTTPClient(t *testing.T) {
	requests := 0
	injected := &http.Client{
		Timeout: 7 * time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests++
			if req.URL.String() != "https://example.test/api/ping" {
				t.Errorf("unexpected URL: %s", req.URL)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Request:    req,
			}, nil
		}),
	}
	client := NewClient(WithBaseURL("https://example.test/api"), WithHTTPClient(injected))

	if client.httpClient != injected {
		t.Fatal("WithHTTPClient did not preserve the injected client")
	}
	if _, _, err := client.DoAPI(http.MethodGet, "ping", nil, nil, false); err != nil {
		t.Fatalf("DoAPI: %v", err)
	}
	if requests != 1 {
		t.Fatalf("transport received %d requests, want 1", requests)
	}
}

func TestWithTransportClonesInjectedHTTPClient(t *testing.T) {
	originalTransport := roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("original transport should not be called")
		return nil, nil
	})
	replacementCalls := 0
	replacement := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		replacementCalls++
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Status:     "204 No Content",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})
	injected := &http.Client{Timeout: 9 * time.Second, Transport: originalTransport}
	client := NewClient(
		WithHTTPClient(injected),
		WithTransport(replacement),
		WithBaseURL("https://example.test/api"),
	)

	if client.httpClient == injected {
		t.Fatal("WithTransport mutated rather than cloned the injected client")
	}
	if client.httpClient.Timeout != injected.Timeout {
		t.Fatalf("timeout = %s, want %s", client.httpClient.Timeout, injected.Timeout)
	}
	if _, _, err := client.DoAPI(http.MethodGet, "ping", nil, nil, false); err != nil {
		t.Fatalf("DoAPI: %v", err)
	}
	if replacementCalls != 1 {
		t.Fatalf("replacement transport received %d requests, want 1", replacementCalls)
	}
}

func TestWithBaseURLControlsAuthenticatedOrigin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-XSRF-TOKEN"); got != "token" {
			t.Errorf("X-XSRF-TOKEN = %q", got)
		}
		if got := r.Header.Get("Cookie"); !strings.Contains(got, "session=value") {
			t.Errorf("Cookie = %q", got)
		}
		_, _ = io.WriteString(w, `{}`)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session=value", XSRFToken: "token"})
	if _, _, err := client.DoAPI(http.MethodGet, "/private", nil, nil, true); err != nil {
		t.Fatalf("authenticated DoAPI: %v", err)
	}
}
