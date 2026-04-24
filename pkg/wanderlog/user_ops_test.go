package wanderlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// newTestClient returns a Client pointed at the given httptest server with a
// dummy auth set so authenticated endpoints don't short-circuit.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	oldBaseURL := BaseURL
	BaseURL = server.URL
	t.Cleanup(func() { BaseURL = oldBaseURL })

	client := NewClient()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	client.SetLogger(logger)
	client.SetAuth(&AuthCredentials{
		SessionCookie: "test-session",
		XSRFToken:     "test-xsrf",
		UserID:        "1",
	})
	return client
}

func TestGetMe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/user" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if _, err := r.Cookie("connect.sid"); err != nil {
			t.Errorf("expected session cookie to be sent: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"email":"a@b.com","username":"me","name":"Me"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	profile, err := client.GetMe()
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if profile.ID != 42 || profile.Username != "me" {
		t.Errorf("unexpected profile: %+v", profile)
	}
	if len(profile.Raw) == 0 {
		t.Error("expected Raw to be populated")
	}
}

func TestMarkNotificationsRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/user/notifications/markRead" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got struct {
			NotificationIDs []string `json:"notificationIds"`
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if len(got.NotificationIDs) != 2 || got.NotificationIDs[0] != "n1" {
			t.Errorf("unexpected body: %+v", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.MarkNotificationsRead([]string{"n1", "n2"}); err != nil {
		t.Fatalf("MarkNotificationsRead: %v", err)
	}
}

func TestGetKeyValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasPrefix(r.URL.Path, "/user/keyValue/") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"value":{"theme":"dark"}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	val, err := client.GetKeyValue("userPrefs")
	if err != nil {
		t.Fatalf("GetKeyValue: %v", err)
	}
	if !strings.Contains(string(val), "dark") {
		t.Errorf("unexpected value: %s", string(val))
	}
}

func TestSetUTCOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/user/utcOffset" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got struct {
			UTCOffset int `json:"utcOffset"`
		}
		_ = json.Unmarshal(body, &got)
		if got.UTCOffset != 540 {
			t.Errorf("unexpected offset: %d", got.UTCOffset)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.SetUTCOffset(540); err != nil {
		t.Fatalf("SetUTCOffset: %v", err)
	}
}

func TestIsUsernameTaken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || !strings.HasPrefix(r.URL.Path, "/user/isUsernameTaken/") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"taken":true}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	taken, err := client.IsUsernameTaken("foo")
	if err != nil {
		t.Fatalf("IsUsernameTaken: %v", err)
	}
	if !taken {
		t.Error("expected taken=true")
	}
}

func TestGetMeRequiresAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called when unauthenticated")
	}))
	defer server.Close()

	oldBaseURL := BaseURL
	BaseURL = server.URL
	defer func() { BaseURL = oldBaseURL }()

	client := NewClient()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	client.SetLogger(logger)

	if _, err := client.GetMe(); err == nil {
		t.Fatal("expected auth error")
	}
}
