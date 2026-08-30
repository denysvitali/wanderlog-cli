package wanderlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAddPlaceResolvesMissingGeometryBeforeMutation(t *testing.T) {
	var mutations atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "getPlaceDetailsAndCardData"):
			_, _ = w.Write([]byte(`{"success":true,"data":{"details":{"name":"Cafe","place_id":"place-1","geometry":{"location":{"lat":12.5,"lng":34.5}},"formatted_address":"Main St"}}}`))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/places"):
			mutations.Add(1)
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			places, _ := body["places"].([]any)
			if len(places) != 1 {
				t.Fatalf("places = %#v", body["places"])
			}
			row, _ := places[0].(map[string]any)
			nested, _ := row["place"].(map[string]any)
			geometry, _ := nested["geometry"].(map[string]any)
			if geometry == nil {
				t.Fatalf("missing nested geometry: %#v", row)
			}
			_, _ = w.Write([]byte(`{"success":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := newAuthenticatedTestClient(server.URL)
	err := client.AddPlace("trip", 7, AddPlaceRequest{Place: AddPlaceInfo{PlaceID: "place-1", Name: "Cafe"}})
	if err != nil {
		t.Fatalf("AddPlace: %v", err)
	}
	if mutations.Load() != 1 {
		t.Fatalf("mutations = %d", mutations.Load())
	}
}

func TestAddPlaceDoesNotMutateWhenDetailResolutionFails(t *testing.T) {
	var mutations atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mutations.Add(1)
		}
		http.Error(w, "details unavailable", http.StatusBadGateway)
	}))
	defer server.Close()
	err := newAuthenticatedTestClient(server.URL).AddPlace("trip", 7, AddPlaceRequest{Place: AddPlaceInfo{PlaceID: "place-1", Name: "Cafe"}})
	if err == nil || mutations.Load() != 0 {
		t.Fatalf("error = %v, mutations = %d", err, mutations.Load())
	}
}

func TestSocialTripKeysArePathEscaped(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.RequestURI)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	client := newAuthenticatedTestClient(server.URL)
	if err := client.SetLike("trip/with?reserved", true); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTripInvites("trip/with?reserved"); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if !strings.Contains(path, "trip%2Fwith%3Freserved") {
			t.Fatalf("unescaped path: %q", path)
		}
	}
}

func TestParseTripLikesBulkIgnoresUnrelatedEnvelopeID(t *testing.T) {
	got := parseTripLikesBulk([]byte(`{"id":"request-id","data":{"trip-a":true}}`), []string{"trip-a"})
	if liked, ok := got["trip-a"]; !ok || !liked {
		t.Fatalf("likes = %#v", got)
	}
}
