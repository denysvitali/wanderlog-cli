package wanderlog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
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

	t.Run("rejects incomplete credentials", func(t *testing.T) {
		client2 := NewClient()
		client2.SetAuth(&AuthCredentials{SessionCookie: "session-only"})
		req, _ := http.NewRequest("GET", "http://example.com/", nil)
		err := client2.addAuthHeaders(req)
		if err == nil || !strings.Contains(err.Error(), "XSRF token is missing") {
			t.Fatalf("expected incomplete credential error, got %v", err)
		}
	})
}

func TestEnsureAuthenticatedRejectsPartialExplicitCredentials(t *testing.T) {
	client := NewClient()
	err := client.EnsureAuthenticated("session-only", "")
	if err == nil || !strings.Contains(err.Error(), "XSRF token is missing") {
		t.Fatalf("expected incomplete credential error, got %v", err)
	}
}

func TestEnsureAuthenticatedValidatesSession(t *testing.T) {
	clearAuthEnvironment(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/user" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			return
		}
		if r.Header.Get("X-XSRF-TOKEN") != "valid-xsrf" {
			t.Error("validation request did not include the XSRF token")
		}
		if cookie, err := r.Cookie("connect.sid"); err != nil || cookie.Value != "valid-session" {
			t.Errorf("validation request session cookie = %v, %v", cookie, err)
		}
		_, _ = w.Write([]byte(`{"id":42,"email":"user@example.com"}`))
	}))
	defer server.Close()
	withAuthTestBaseURL(t, server.URL)

	client := NewClient()
	if err := client.EnsureAuthenticated("valid-session", "valid-xsrf"); err != nil {
		t.Fatalf("EnsureAuthenticated: %v", err)
	}
}

func TestEnsureAuthenticatedRejectedSessionRefreshesFromEnvironment(t *testing.T) {
	clearAuthEnvironment(t)
	t.Setenv("WANDERLOG_AUTH_SESSION_COOKIE", "expired-session")
	t.Setenv("WANDERLOG_AUTH_XSRF_TOKEN", "expired-xsrf")
	t.Setenv("WANDERLOG_AUTH_EMAIL", "user@example.com")
	t.Setenv("WANDERLOG_AUTH_PASSWORD", "environment-password")

	validationCalls := 0
	loginCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/user":
			validationCalls++
			w.WriteHeader(http.StatusUnauthorized)
		case r.Method == http.MethodPost && r.URL.Path == "/user/login":
			loginCalls++
			var request LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode login: %v", err)
				return
			}
			if request.Email != "user@example.com" || request.Password != "environment-password" {
				t.Errorf("unexpected login request: %+v", request)
			}
			w.Header().Add("Set-Cookie", "connect.sid=refreshed-session; Path=/; HttpOnly")
			w.Header().Add("Set-Cookie", "XSRF-TOKEN=refreshed-xsrf; Path=/")
			_, _ = w.Write([]byte(`{"success":true,"user":{"id":42,"email":"user@example.com"}}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	withAuthTestBaseURL(t, server.URL)

	client := NewClient()
	if err := client.EnsureAuthenticatedContext(context.Background(), "", ""); err != nil {
		t.Fatalf("EnsureAuthenticatedContext: %v", err)
	}
	if validationCalls != 1 || loginCalls != 1 {
		t.Fatalf("validation calls = %d, login calls = %d", validationCalls, loginCalls)
	}
	if client.auth == nil || client.auth.SessionCookie != "refreshed-session" || client.auth.XSRFToken != "refreshed-xsrf" {
		t.Fatalf("refreshed credentials not installed: %+v", client.auth)
	}
}

func TestEnsureAuthenticatedRefreshesRejectedKeychainSession(t *testing.T) {
	clearAuthEnvironment(t)
	t.Setenv("WANDERLOG_DISABLE_KEYCHAIN", "")
	t.Setenv("WANDERLOG_AUTH_EMAIL", "user@example.com")
	t.Setenv("WANDERLOG_AUTH_PASSWORD", "environment-password")
	keyring.MockInit()
	if err := SaveCredentials(&AuthCredentials{SessionCookie: "expired-session", XSRFToken: "expired-xsrf"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = DeleteCredentials() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/user":
			w.WriteHeader(http.StatusUnauthorized)
		case r.Method == http.MethodPost && r.URL.Path == "/user/login":
			w.Header().Add("Set-Cookie", "connect.sid=refreshed-keychain-session; Path=/; HttpOnly")
			w.Header().Add("Set-Cookie", "XSRF-TOKEN=refreshed-keychain-xsrf; Path=/")
			_, _ = w.Write([]byte(`{"success":true,"user":{"id":42}}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	withAuthTestBaseURL(t, server.URL)

	client := NewClient()
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("EnsureAuthenticated: %v", err)
	}
	stored, err := LoadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	if stored == nil || stored.SessionCookie != "refreshed-keychain-session" || stored.XSRFToken != "refreshed-keychain-xsrf" {
		t.Fatalf("keychain was not refreshed: %+v", stored)
	}
}

func TestEnsureAuthenticatedEnvironmentLoginWorksWithoutKeychain(t *testing.T) {
	clearAuthEnvironment(t)
	t.Setenv("WANDERLOG_AUTH_EMAIL", "user@example.com")
	t.Setenv("WANDERLOG_AUTH_PASSWORD", "environment-password")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/user/login" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			return
		}
		w.Header().Add("Set-Cookie", "connect.sid=environment-session; Path=/; HttpOnly")
		w.Header().Add("Set-Cookie", "XSRF-TOKEN=environment-xsrf; Path=/")
		_, _ = w.Write([]byte(`{"success":true,"user":{"id":42}}`))
	}))
	defer server.Close()
	withAuthTestBaseURL(t, server.URL)

	client := NewClient()
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("EnsureAuthenticated: %v", err)
	}
	if client.auth == nil || client.auth.SessionCookie != "environment-session" {
		t.Fatalf("environment credentials not installed: %+v", client.auth)
	}
}

func TestEnsureAuthenticatedDoesNotFallbackForServerFailure(t *testing.T) {
	clearAuthEnvironment(t)
	t.Setenv("WANDERLOG_AUTH_EMAIL", "user@example.com")
	t.Setenv("WANDERLOG_AUTH_PASSWORD", "environment-password")
	loginCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			loginCalls++
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	withAuthTestBaseURL(t, server.URL)

	client := NewClient()
	err := client.EnsureAuthenticated("session", "xsrf")
	if err == nil || errors.Is(err, ErrSessionRejected) {
		t.Fatalf("expected non-rejection validation failure, got %v", err)
	}
	if loginCalls != 0 {
		t.Fatalf("login fallback attempted %d times", loginCalls)
	}
	if client.auth != nil {
		t.Fatalf("rejected credentials remain installed: %+v", client.auth)
	}
}

func TestEnsureAuthenticatedContextHonorsCancellation(t *testing.T) {
	clearAuthEnvironment(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := NewClient()
	err := client.EnsureAuthenticatedContext(ctx, "session", "xsrf")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if client.auth != nil {
		t.Fatalf("credentials remain installed after canceled validation: %+v", client.auth)
	}
}

func TestEnsureAuthenticatedNeverUsesConfigPassword(t *testing.T) {
	clearAuthEnvironment(t)
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("auth.session.cookie", "expired-session")
	viper.Set("auth.session.xsrf_token", "expired-xsrf")
	viper.Set("auth.email", "config@example.com")
	viper.Set("auth.password", "forbidden-config-password")
	loginCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			loginCalls++
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	withAuthTestBaseURL(t, server.URL)

	client := NewClient()
	err := client.EnsureAuthenticated("", "")
	if !errors.Is(err, ErrSessionRejected) {
		t.Fatalf("expected rejected config session, got %v", err)
	}
	if loginCalls != 0 {
		t.Fatalf("config password triggered %d login attempts", loginCalls)
	}
}

func clearAuthEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"WANDERLOG_AUTH_SESSION_COOKIE",
		"WANDERLOG_AUTH_XSRF_TOKEN",
		"WANDERLOG_AUTH_SESSION_XSRF_TOKEN",
		"WANDERLOG_AUTH_USER_ID",
		"WANDERLOG_AUTH_EMAIL",
		"WANDERLOG_AUTH_PASSWORD",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("WANDERLOG_DISABLE_KEYCHAIN", "1")
}

func withAuthTestBaseURL(t *testing.T, value string) {
	t.Helper()
	old := BaseURL
	BaseURL = value
	t.Cleanup(func() { BaseURL = old })
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

	t.Run("missing xsrf token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Set-Cookie", "connect.sid=session; Path=/; HttpOnly")
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
		if err == nil || !strings.Contains(err.Error(), "XSRF token not found") {
			t.Fatalf("expected XSRF token error, got: %v", err)
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
