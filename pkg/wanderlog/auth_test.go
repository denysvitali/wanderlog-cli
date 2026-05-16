package wanderlog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestAddAuthHeaders(t *testing.T) {
	client := NewClient()
	logger := newTestLogger(t)
	client.SetLogger(logger)

	client.SetAuth(&AuthCredentials{
		SessionCookie: "s:test-session",
		XSRFToken:     "test-xsrf",
		UserID:        "42",
	})

	t.Run("adds session cookie and xsrf headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com/", nil)
		err := client.addAuthHeaders(req)
		if err != nil {
			t.Fatalf("addAuthHeaders: %v", err)
		}

		if req.Header.Get("X-XSRF-TOKEN") != "test-xsrf" {
			t.Errorf("expected XSRF token header, got: %s", req.Header.Get("X-XSRF-TOKEN"))
		}

		cookies := req.Cookies()
		foundSession := false
		foundXsrf := false
		for _, c := range cookies {
			if c.Name == "connect.sid" && c.Value == "s:test-session" {
				foundSession = true
			}
			if c.Name == "XSRF-TOKEN" && c.Value == "test-xsrf" {
				foundXsrf = true
			}
		}
		if !foundSession {
			t.Error("expected connect.sid cookie")
		}
		if !foundXsrf {
			t.Error("expected XSRF-TOKEN cookie")
		}
	})

	t.Run("returns error when auth is nil", func(t *testing.T) {
		client2 := NewClient()
		client2.SetLogger(newTestLogger(t))
		req, _ := http.NewRequest("GET", "http://example.com/", nil)
		err := client2.addAuthHeaders(req)
		if err == nil {
			t.Fatal("expected error when auth is nil")
		}
		if !strings.Contains(err.Error(), "not authenticated") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestLogin(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/user/login") {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}

			w.Header().Add("Set-Cookie", "connect.sid=s%3Aabc123; Path=/; HttpOnly")
			w.Header().Add("Set-Cookie", "XSRF-TOKEN=test-xsrf; Path=/")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"user":{"id":42,"email":"a@b.com","name":"Alice","username":"alice"}}`))
		}))
		defer server.Close()

		old := BaseURL
		BaseURL = server.URL
		defer func() { BaseURL = old }()

		client := NewClient()
		client.SetLogger(newTestLogger(t))

		creds, err := client.Login("a@b.com", "pass123")
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		if creds.SessionCookie != "s%3Aabc123" {
			t.Errorf("unexpected session cookie: %s", creds.SessionCookie)
		}
		if creds.XSRFToken != "test-xsrf" {
			t.Errorf("unexpected xsrf token: %s", creds.XSRFToken)
		}
		if creds.UserID != "42" {
			t.Errorf("unexpected user ID: %s", creds.UserID)
		}
	})

	t.Run("login failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		old := BaseURL
		BaseURL = server.URL
		defer func() { BaseURL = old }()

		client := NewClient()
		client.SetLogger(newTestLogger(t))

		_, err := client.Login("a@b.com", "wrong")
		if err == nil {
			t.Fatal("expected error for failed login")
		}
	})

	t.Run("missing session cookie", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"user":{"id":42}}`))
		}))
		defer server.Close()

		old := BaseURL
		BaseURL = server.URL
		defer func() { BaseURL = old }()

		client := NewClient()
		client.SetLogger(newTestLogger(t))

		_, err := client.Login("a@b.com", "pass123")
		if err == nil || !strings.Contains(err.Error(), "session cookie not found") {
			t.Fatalf("expected session cookie error, got: %v", err)
		}
	})
}

func TestSetAuth(t *testing.T) {
	client := NewClient()
	creds := &AuthCredentials{SessionCookie: "test", XSRFToken: "xsrf", UserID: "1"}
	client.SetAuth(creds)
	if client.auth != creds {
		t.Error("SetAuth did not store credentials")
	}
}

func newTestLogger(t *testing.T) *logrus.Logger {
	t.Helper()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return logger
}
