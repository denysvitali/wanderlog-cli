package wanderlog

import (
	"strings"
	"testing"
)

func TestDecodeMutationBodyRejectsFalseAndMalformedEnvelopes(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "success false", body: `{"success":false,"error":"rejected"}`, want: "rejected"},
		{name: "malformed JSON", body: `{`, want: "decoding mutation response"},
		{name: "missing success", body: `{"data":{}}`, want: "missing boolean success"},
		{name: "non-boolean success", body: `{"success":"yes"}`, want: "decoding mutation response"},
		{name: "empty", body: ``, want: "empty mutation response"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decodeMutationBody("Mutation", 200, []byte(tt.body), nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestDecodeMutationBodyDecodesSuccessfulPayload(t *testing.T) {
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := decodeMutationBody("Mutation", 200, []byte(`{"success":true,"data":{"id":7}}`), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.ID != 7 {
		t.Fatalf("data.id = %d, want 7", response.Data.ID)
	}
}
