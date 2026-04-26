package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testResponse struct {
	Message string `json:"message"`
}

func TestDoJSON(t *testing.T) {
	t.Run("get request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/test") {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"hello"}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		var out testResponse
		err := client.doJSON("GET", "/test", nil, &out, false, "testOp")
		if err != nil {
			t.Fatalf("doJSON: %v", err)
		}
		if out.Message != "hello" {
			t.Errorf("unexpected response: %+v", out)
		}
	})

	t.Run("post with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["key"] != "value" {
				t.Errorf("unexpected body: %v", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"ok"}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		var out testResponse
		err := client.doJSON("POST", "/test", map[string]string{"key": "value"}, &out, false, "testOp")
		if err != nil {
			t.Fatalf("doJSON: %v", err)
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		client := NewClient()
		client.SetLogger(newTestLogger(t))

		var out testResponse
		err := client.doJSON("GET", "/test", nil, &out, true, "testOp")
		if err == nil || !strings.Contains(err.Error(), "authentication required") {
			t.Fatalf("expected auth error, got: %v", err)
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := newTestClient(t, server)
		var out testResponse
		err := client.doJSON("GET", "/test", nil, &out, false, "testOp")
		if err == nil {
			t.Fatal("expected error for non-200")
		}
	})

	t.Run("nil out skips decoding", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"hello"}`))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		err := client.doJSON("GET", "/test", nil, nil, false, "testOp")
		if err != nil {
			t.Fatalf("doJSON with nil out: %v", err)
		}
	})
}

func TestDoRaw(t *testing.T) {
	t.Run("get raw bytes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write([]byte("col1,col2\nval1,val2"))
		}))
		defer server.Close()

		client := newTestClient(t, server)
		body, err := client.doRaw("GET", "/test", nil, false, "testOp")
		if err != nil {
			t.Fatalf("doRaw: %v", err)
		}
		if !strings.Contains(string(body), "col1") {
			t.Errorf("unexpected body: %s", string(body))
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		client := NewClient()
		client.SetLogger(newTestLogger(t))

		_, err := client.doRaw("GET", "/test", nil, true, "testOp")
		if err == nil || !strings.Contains(err.Error(), "authentication required") {
			t.Fatalf("expected auth error, got: %v", err)
		}
	})
}

func TestTruncateForLog(t *testing.T) {
	t.Run("short string not truncated", func(t *testing.T) {
		result := truncateForLog("hello", 10)
		if result != "hello" {
			t.Errorf("expected 'hello', got '%s'", result)
		}
	})

	t.Run("long string truncated", func(t *testing.T) {
		long := "this is a very long string that should be truncated"
		result := truncateForLog(long, 10)
		if len(result) != 13 { // 10 + "..."
			t.Errorf("expected length 13, got %d: %s", len(result), result)
		}
		if !strings.HasSuffix(result, "...") {
			t.Errorf("expected suffix '...', got '%s'", result)
		}
	})

	t.Run("exact length", func(t *testing.T) {
		result := truncateForLog("exact", 5)
		if result != "exact" {
			t.Errorf("expected 'exact', got '%s'", result)
		}
	})
}
