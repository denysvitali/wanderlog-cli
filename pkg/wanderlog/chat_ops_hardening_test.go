package wanderlog

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newAuthenticatedTestClient(baseURL string) *Client {
	client := NewClient(WithBaseURL(baseURL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	return client
}

func TestAssistantStreamAccumulatesContentAndMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		_, _ = w.Write([]byte("{\"type\":\"chatMetadata\",\"data\":{\"id\":1}}\n" +
			"{\"type\":\"content\",\"data\":\"hello \"}\n" +
			"{\"type\":\"content\",\"data\":\"world\"}\n"))
	}))
	defer server.Close()

	got, err := newAuthenticatedTestClient(server.URL).GetTripPlanAssistantTextContext(
		context.Background(), AssistantTextRequest{Message: "plan my day"},
	)
	if err != nil {
		t.Fatalf("GetTripPlanAssistantTextContext: %v", err)
	}
	if got.Content != "hello world" || string(got.ChatMetadata) != `{"id":1}` || len(got.Events) != 3 {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestAssistantStreamReportsApplicationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{\"success\":false,\"message\":\"quota exhausted\"}\n"))
	}))
	defer server.Close()

	_, err := newAuthenticatedTestClient(server.URL).GetTripPlanAssistantText(AssistantTextRequest{Message: "hello"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Message != "quota exhausted" {
		t.Fatalf("error = %#v", err)
	}
}

func TestAssistantStreamRejectsMalformedAndOversizedContentEvents(t *testing.T) {
	t.Run("non-string content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("{\"type\":\"content\",\"data\":{}}\n"))
		}))
		defer server.Close()
		_, err := newAuthenticatedTestClient(server.URL).GetTripPlanAssistantText(AssistantTextRequest{Message: "hello"})
		if err == nil || !strings.Contains(err.Error(), "must be a JSON string") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("oversized event", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"type":"content","data":"` + strings.Repeat("x", MaxAssistantStreamEventBytes) + `"}`))
		}))
		defer server.Close()
		_, err := newAuthenticatedTestClient(server.URL).GetTripPlanAssistantText(AssistantTextRequest{Message: "hello"})
		if err == nil || !strings.Contains(err.Error(), "event exceeds") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestAssistantStreamHonorsCanceledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := newAuthenticatedTestClient(server.URL).GetTripPlanAssistantTextContext(ctx, AssistantTextRequest{Message: "hello"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
