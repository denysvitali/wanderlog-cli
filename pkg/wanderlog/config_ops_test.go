package wanderlog

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGlobalConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/globalConfig" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"config":{"featureFlag":true}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	cfg, err := client.GetGlobalConfig()
	if err != nil {
		t.Fatalf("GetGlobalConfig: %v", err)
	}
	if len(cfg.Raw) == 0 {
		t.Error("expected Raw to be populated")
	}
}

func TestSetSessionStoreValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/sessionStore" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.SetSessionStoreValue("foo", "bar"); err != nil {
		t.Fatalf("SetSessionStoreValue: %v", err)
	}
}

func TestGetSessionStore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessionStore" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"store":{}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	resp, err := client.GetSessionStore()
	if err != nil {
		t.Fatalf("GetSessionStore: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestGetSessionPreferences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessionStore/preferences/en" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"preferences":{}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if _, err := client.GetSessionPreferences("en"); err != nil {
		t.Fatalf("GetSessionPreferences: %v", err)
	}
}
