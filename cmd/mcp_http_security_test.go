package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeMCPHTTPAddress(t *testing.T) {
	tests := []struct {
		input    string
		want     string
		loopback bool
	}{
		{input: ":8080", want: "127.0.0.1:8080", loopback: true},
		{input: "localhost:8080", want: "localhost:8080", loopback: true},
		{input: "127.0.0.1:8080", want: "127.0.0.1:8080", loopback: true},
		{input: "0.0.0.0:8080", want: "0.0.0.0:8080", loopback: false},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got, loopback, err := normalizeMCPHTTPAddress(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want || loopback != test.loopback {
				t.Fatalf("got (%q, %v), want (%q, %v)", got, loopback, test.want, test.loopback)
			}
		})
	}
}

func TestValidateMCPHTTPConfig(t *testing.T) {
	const token = "0123456789abcdef"
	if got, err := validateMCPHTTPConfig(":8080", token, "", ""); err != nil || got != "127.0.0.1:8080" {
		t.Fatalf("valid loopback config = (%q, %v)", got, err)
	}
	for name, input := range map[string]struct {
		addr, token, cert, key string
	}{
		"missing token": {addr: ":8080"},
		"partial TLS":   {addr: ":8080", token: token, cert: "cert.pem"},
		"remote no TLS": {addr: "0.0.0.0:8080", token: token},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validateMCPHTTPConfig(input.addr, input.token, input.cert, input.key); err == nil {
				t.Fatal("expected invalid configuration")
			}
		})
	}
	if _, err := validateMCPHTTPConfig("0.0.0.0:8080", token, "cert.pem", "key.pem"); err != nil {
		t.Fatalf("valid remote TLS config: %v", err)
	}
}

func TestRunMCPHTTPServerReturnsConfigurationError(t *testing.T) {
	err := runMCPHTTPServer(":8080", false, "", "", "", "")
	if err == nil || !strings.Contains(err.Error(), "bearer token") {
		t.Fatalf("expected normal configuration error, got %v", err)
	}
}

func TestRunMCPServersHonorCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runMCPStdioServerContext(ctx, false, ""); err != nil {
		t.Fatalf("stdio cancellation: %v", err)
	}
	if err := runMCPHTTPServerContext(ctx, ":0", false, "", "0123456789abcdef", "", ""); err != nil {
		t.Fatalf("HTTP cancellation: %v", err)
	}
}

func TestMCPHTTPMiddleware(t *testing.T) {
	const token = "0123456789abcdef"
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := mcpHTTPMiddleware(next, token)

	t.Run("requires bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/mcp", strings.NewReader("{}"))
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", resp.Code)
		}
	})

	t.Run("rejects foreign browser origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/mcp", strings.NewReader("{}"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Origin", "https://evil.example")
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", resp.Code)
		}
	})

	t.Run("accepts authenticated same-origin request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/mcp", strings.NewReader("{}"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Origin", "http://127.0.0.1")
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204: %s", resp.Code, resp.Body.String())
		}
	})

	t.Run("limits request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/mcp", strings.NewReader(strings.Repeat("x", maxMCPHTTPRequestBytes+1)))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413", resp.Code)
		}
	})
}
