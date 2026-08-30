package wanderlog

import (
	"strings"
	"testing"
)

func TestDecodeAPIBodyExplicitFailure(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "message", body: `{"success":false,"message":"denied"}`, want: "denied"},
		{name: "string error", body: `{"success":false,"error":"expired"}`, want: "expired"},
		{name: "object error", body: `{"success":false,"error":{"message":"conflict"}}`, want: "conflict"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := decodeAPIBody("TestOperation", 200, []byte(test.body), &map[string]any{})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want message containing %q", err, test.want)
			}
		})
	}
}

func TestDecodeAPIBodyAllowsMissingSuccess(t *testing.T) {
	var result map[string]any
	if err := decodeAPIBody("TestOperation", 200, []byte(`{"data":"ok"}`), &result); err != nil {
		t.Fatal(err)
	}
	if result["data"] != "ok" {
		t.Fatalf("decoded result = %#v", result)
	}
}
