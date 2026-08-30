package wanderlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseTripLikesBulkResponseShapes(t *testing.T) {
	tests := []struct {
		name string
		body string
		want map[string]bool
	}{
		{name: "keyed booleans", body: `{"success":true,"data":{"trip-a":true,"trip-b":false}}`, want: map[string]bool{"trip-a": true, "trip-b": false}},
		{name: "keyed objects", body: `{"success":true,"likes":{"trip-a":{"liked":false},"trip-b":{"isLiked":true}}}`, want: map[string]bool{"trip-a": false, "trip-b": true}},
		{name: "positional booleans", body: `{"success":true,"data":[false,true]}`, want: map[string]bool{"trip-a": false, "trip-b": true}},
		{name: "liked key list", body: `{"success":true,"data":["trip-b"]}`, want: map[string]bool{"trip-a": false, "trip-b": true}},
		{name: "empty liked key list", body: `{"success":true,"data":[]}`, want: map[string]bool{"trip-a": false, "trip-b": false}},
		{name: "object list", body: `{"success":true,"data":[{"tripPlanKey":"trip-a","userLiked":true},{"key":"trip-b","liked":false}]}`, want: map[string]bool{"trip-a": true, "trip-b": false}},
		{name: "single boolean", body: `{"success":true,"data":false}`, want: map[string]bool{}},
		{name: "unrecognized", body: `{"success":true,"data":{"total":2}}`, want: map[string]bool{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTripLikesBulk([]byte(tt.body), []string{"trip-a", "trip-b"})
			if got == nil {
				got = map[string]bool{}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseTripLikesBulk() = %#v, want %#v", got, tt.want)
			}
		})
	}

	if got := parseTripLikesBulk([]byte(`{"success":true,"data":false}`), []string{"trip-a"}); !reflect.DeepEqual(got, map[string]bool{"trip-a": false}) {
		t.Fatalf("single-key boolean = %#v", got)
	}
}

func TestGetLikeCountUsesAuthenticatedLikesEndpoint(t *testing.T) {
	var likesCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/tripPlans/trip-a":
			_, _ = io.WriteString(w, `{"success":true,"tripPlan":{"key":"trip-a","likeCount":17,"itinerary":{"sections":[]}},"resources":{"placeMetadata":[]}}`)
		case r.Method == http.MethodPost && r.URL.Path == "/tripPlans/likes":
			likesCalls++
			if r.Header.Get("X-XSRF-TOKEN") != "xsrf" {
				t.Fatalf("missing authentication headers")
			}
			var request LikesBulkRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode likes request: %v", err)
			}
			if !reflect.DeepEqual(request.Keys, []string{"trip-a"}) {
				t.Fatalf("likes keys = %#v", request.Keys)
			}
			_, _ = io.WriteString(w, `{"success":true,"data":{"trip-a":false}}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	result, err := client.GetLikeCount("trip-a")
	if err != nil {
		t.Fatalf("GetLikeCount: %v", err)
	}
	if result.Count != 17 || result.UserLiked || !result.UserLikedKnown {
		t.Fatalf("unexpected like count: %+v", result)
	}
	if likesCalls != 1 {
		t.Fatalf("likes calls = %d, want 1", likesCalls)
	}
}

func TestGetLikeCountLeavesUserStateUnknownWithoutAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/tripPlans/trip-a" {
			t.Fatalf("unexpected unauthenticated request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"success":true,"tripPlan":{"key":"trip-a","likeCount":3,"itinerary":{"sections":[]}},"resources":{"placeMetadata":[]}}`)
	}))
	defer server.Close()

	result, err := NewClient(WithBaseURL(server.URL)).GetLikeCount("trip-a")
	if err != nil {
		t.Fatalf("GetLikeCount: %v", err)
	}
	if result.UserLiked || result.UserLikedKnown {
		t.Fatalf("unauthenticated state must be unknown: %+v", result)
	}
}

func TestGetLikeCountDoesNotClaimFalseForUnrecognizedBulkResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"success":true,"tripPlan":{"key":"trip-a","likeCount":3,"itinerary":{"sections":[]}},"resources":{"placeMetadata":[]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"success":true,"data":{"unexpected":"shape"}}`)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf"})
	result, err := client.GetLikeCount("trip-a")
	if err != nil {
		t.Fatalf("GetLikeCount: %v", err)
	}
	if result.UserLiked || result.UserLikedKnown {
		t.Fatalf("unrecognized state must be unknown: %+v", result)
	}
}
