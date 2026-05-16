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

func TestUpdateMe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/user" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"username":"me","name":"Updated"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	profile, err := client.UpdateMe(UpdateUserRequest{Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdateMe: %v", err)
	}
	if profile.Name != "Updated" {
		t.Errorf("unexpected name: %s", profile.Name)
	}
}

func TestServerLogout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/user/logout" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.ServerLogout(); err != nil {
		t.Fatalf("ServerLogout: %v", err)
	}
}

func TestGetNotifications(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/user/notifications" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"notifications":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetNotifications(0)
	if err != nil {
		t.Fatalf("GetNotifications: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetNotificationsWithOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") != "10" {
			t.Errorf("expected offset=10, got %q", r.URL.Query().Get("offset"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"notifications":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.GetNotifications(10)
	if err != nil {
		t.Fatalf("GetNotifications(10): %v", err)
	}
}

func TestGetNotificationSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/notification/settings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"notify":true,"email":false}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetNotificationSettings()
	if err != nil {
		t.Fatalf("GetNotificationSettings: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestUpdateNotificationSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/user/notification/settings" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"notify":false}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.UpdateNotificationSettings(json.RawMessage(`{"notify":false}`))
	if err != nil {
		t.Fatalf("UpdateNotificationSettings: %v", err)
	}
}

func TestSetKeyValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.HasPrefix(r.URL.Path, "/user/keyValue/") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.SetKeyValue("mykey", "myvalue"); err != nil {
		t.Fatalf("SetKeyValue: %v", err)
	}
}

func TestListFollowing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/user/following/list" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"following":{"123":true}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.ListFollowing([]string{"123", "456"})
	if err != nil {
		t.Fatalf("ListFollowing: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestAutocompleteUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/autocomplete/al" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"users":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.AutocompleteUsers("al")
	if err != nil {
		t.Fatalf("AutocompleteUsers: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestFindUserByEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("email") != "a@b.com" {
			t.Errorf("expected email=a@b.com, got %q", r.URL.Query().Get("email"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"username":"alice"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	profile, err := client.FindUserByEmail("a@b.com")
	if err != nil {
		t.Fatalf("FindUserByEmail: %v", err)
	}
	if profile.ID != 42 {
		t.Errorf("unexpected id: %d", profile.ID)
	}
}

func TestBlockUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/user/block" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.BlockUser("999"); err != nil {
		t.Fatalf("BlockUser: %v", err)
	}
}

func TestGetUserEmails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/emails" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"emails":[{"email":"a@b.com","primary":true}]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetUserEmails()
	if err != nil {
		t.Fatalf("GetUserEmails: %v", err)
	}
	if len(resp.Emails) == 0 || resp.Emails[0].Email != "a@b.com" {
		t.Errorf("unexpected emails: %+v", resp.Emails)
	}
}

func TestGetUserProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/profile/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"trips":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetUserProfile(42)
	if err != nil {
		t.Fatalf("GetUserProfile: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetUserProfileByUsername(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tripPlans/profile/byUsername/alice" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"trips":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetUserProfileByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserProfileByUsername: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
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
