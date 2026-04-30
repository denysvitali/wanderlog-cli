package wanderlog

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type requestContract struct {
	Name             string         `json:"name"`
	Method           string         `json:"method"`
	Path             string         `json:"path"`
	ReferencePath   string         `json:"referencePath"`
	ReferenceMethod string         `json:"referenceMethod,omitempty"`
	Auth             bool           `json:"auth"`
	Query            map[string]any `json:"query,omitempty"`
	Body             any            `json:"body,omitempty"`
}

type capturedAPIRequest struct {
	Method string
	Path   string
	Query  map[string]any
	Body   any
	Header http.Header
}

func TestAPIRequestContracts(t *testing.T) {
	contracts := loadRequestContracts(t)
	referenceEndpoints := loadReferenceEndpoints(t)

	runners := map[string]func(*Client) error{
		"Login": func(c *Client) error {
			_, err := c.Login("user@example.com", "secret")
			return err
		},
		"GetTrip": func(c *Client) error {
			_, err := c.GetTrip("trip-key")
			return err
		},
		"GetTripSections": func(c *Client) error {
			_, err := c.GetTripSections("trip-key")
			return err
		},
		"GetPlaceDetails": func(c *Client) error {
			_, err := c.GetPlaceDetails("place-123")
			return err
		},
		"GetFlightStops": func(c *Client) error {
			_, err := c.GetFlightStops("244", "MU", "2026-05-11")
			return err
		},
		"GetAllAirlines": func(c *Client) error {
			_, err := c.GetAllAirlines()
			return err
		},
		"GetGooglePriceRates": func(c *Client) error {
			_, err := c.GetGooglePriceRates("property-123")
			return err
		},
		"GetGlobalConfig": func(c *Client) error {
			_, err := c.GetGlobalConfig()
			return err
		},
		"GetSessionStore": func(c *Client) error {
			_, err := c.GetSessionStore()
			return err
		},
		"GetFeedHome": func(c *Client) error {
			_, err := c.GetFeedHome()
			return err
		},
		"GetFeed": func(c *Client) error {
			_, err := c.GetFeed()
			return err
		},
		"GetFeedV2": func(c *Client) error {
			_, err := c.GetFeedV2()
			return err
		},
		"GetFeedMostRecent": func(c *Client) error {
			_, err := c.GetFeedMostRecent()
			return err
		},
		"GetFriendsPlans": func(c *Client) error {
			_, err := c.GetFriendsPlans()
			return err
		},
		"BrowseGuides": func(c *Client) error {
			_, err := c.BrowseGuides(0)
			return err
		},
		"GetUserTrips": func(c *Client) error {
			_, err := c.GetUserTrips()
			return err
		},
		"CreateTrip": func(c *Client) error {
			_, err := c.CreateTrip(CreateTripRequest{
				Title:     "API Contract Trip",
				GeoIDs:    []int{123},
				StartDate: "2026-06-01",
				EndDate:   "2026-06-04",
			})
			return err
		},
		"CreateExampleTrip": func(c *Client) error {
			_, err := c.CreateExampleTrip()
			return err
		},
		"DeleteTrip": func(c *Client) error {
			return c.DeleteTrip("trip-key")
		},
		"AddPlace": func(c *Client) error {
			return c.AddPlace("trip-key", 7, AddPlaceRequest{
				Place: AddPlaceInfo{Name: "Tokyo Station", PlaceID: "place-123"},
			})
		},
		"RemovePlace": func(c *Client) error {
			return c.RemovePlace("trip-key", 7, 123)
		},
		"ApplyOperations": func(c *Client) error {
			return c.ApplyOperations("trip-key", []Operation{})
		},
		"SetLike": func(c *Client) error {
			return c.SetLike("trip-key", true)
		},
		"CopyTrip": func(c *Client) error {
			_, err := c.CopyTrip("source-key")
			return err
		},
		"RestoreTrip": func(c *Client) error {
			return c.RestoreTrip("trip-key")
		},
		"SendTripInvites": func(c *Client) error {
			return c.SendTripInvites("trip-key", SendInvitesRequest{
				Invitees: []string{"friend@example.com"},
				Message:  "Join this trip",
			})
		},
		"ListTripInvites": func(c *Client) error {
			_, err := c.ListTripInvites("trip-key")
			return err
		},
		"AddCollaborator": func(c *Client) error {
			return c.AddCollaborator("trip-key", 42)
		},
		"RemoveCollaborator": func(c *Client) error {
			return c.RemoveCollaborator("trip-key", 42)
		},
		"ExportTrip": func(c *Client) error {
			_, err := c.ExportTrip("trip-key")
			return err
		},
		"ToggleChecklistItem": func(c *Client) error {
			_, err := c.ToggleChecklistItem("trip-key", 7, 99, true)
			return err
		},
		"GetMe": func(c *Client) error {
			_, err := c.GetMe()
			return err
		},
		"ServerLogout": func(c *Client) error {
			return c.ServerLogout()
		},
		"GetNotifications": func(c *Client) error {
			_, err := c.GetNotifications(25)
			return err
		},
		"MarkNotificationsRead": func(c *Client) error {
			return c.MarkNotificationsRead([]string{"notification-1"})
		},
		"GetNotificationSettings": func(c *Client) error {
			_, err := c.GetNotificationSettings()
			return err
		},
		"UpdateNotificationSettings": func(c *Client) error {
			_, err := c.UpdateNotificationSettings(json.RawMessage(`{"email":true}`))
			return err
		},
		"FindUserByEmail": func(c *Client) error {
			_, err := c.FindUserByEmail("friend@example.com")
			return err
		},
		"BlockUser": func(c *Client) error {
			return c.BlockUser("42")
		},
		"SetUTCOffset": func(c *Client) error {
			return c.SetUTCOffset(120)
		},
		"GetTripUpdateRequired": func(c *Client) error {
			_, err := c.GetTripUpdateRequired("trip-key")
			return err
		},
		"SetTripDistinction": func(c *Client) error {
			return c.SetTripDistinction("trip-key", "featured")
		},
		"RegisterTripView": func(c *Client) error {
			return c.RegisterTripView("trip-key")
		},
		"GetJournalStopPolylines": func(c *Client) error {
			_, err := c.GetJournalStopPolylines(JournalStopPolylinesRequest{
				Stops: []JournalStop{
					{ID: "a", Lat: 35.6812, Lng: 139.7671, PlaceID: "place-a", StopType: "place"},
					{ID: "b", Lat: 35.6895, Lng: 139.6917, PlaceID: "place-b", StopType: "place"},
				},
				ExistingPolylines: []JournalPolyline{},
			})
			return err
		},
		"GetIfEdited": func(c *Client) error {
			_, err := c.GetIfEdited(GetIfEditedRequest{
				TripPlans: []EditCheck{{Key: "trip-key", LastEditRevision: 4}},
				Platform:  "web",
			})
			return err
		},
		"ListGeosWithGoodGuides": func(c *Client) error {
			_, err := c.ListGeosWithGoodGuides()
			return err
		},
		"ListPopularAndNearbyGeos": func(c *Client) error {
			_, err := c.ListPopularAndNearbyGeos()
			return err
		},
		"FindCountryForIP": func(c *Client) error {
			_, err := c.FindCountryForIP()
			return err
		},
		"FindNearestTripadvisorGeo": func(c *Client) error {
			_, err := c.FindNearestTripadvisorGeo(35.6812, 139.7671)
			return err
		},
		"FindNearestGeosToIP": func(c *Client) error {
			_, err := c.FindNearestGeosToIP(nil)
			return err
		},
		"FindNearestKayakCity": func(c *Client) error {
			_, err := c.FindNearestKayakCity(35.6812, 139.7671, "Tokyo")
			return err
		},
		"GetClientGeos": func(c *Client) error {
			_, err := c.GetClientGeos([]int{12345, 67890})
			return err
		},
		"ListTripPlannerGeos": func(c *Client) error {
			_, err := c.ListTripPlannerGeos()
			return err
		},
		"ListCountries": func(c *Client) error {
			_, err := c.ListCountries("en")
			return err
		},
		"ListGeoCategoriesForCategory": func(c *Client) error {
			_, err := c.ListGeoCategoriesForCategory(42, "en")
			return err
		},
		"ListGeoCategoriesForGeo": func(c *Client) error {
			_, err := c.ListGeoCategoriesForGeo(12345, "tripPlanner")
			return err
		},
		"ListGeoInMonthGeos": func(c *Client) error {
			_, err := c.ListGeoInMonthGeos()
			return err
		},
		"ListKeywordCategories": func(c *Client) error {
			_, err := c.ListKeywordCategories("en")
			return err
		},
		"SearchGeo": func(c *Client) error {
			_, err := c.SearchGeo(map[string]string{"q": "tokyo"})
			return err
		},
		"GetAllDistanceInfoForPlace": func(c *Client) error {
			_, err := c.GetAllDistanceInfoForPlace(map[string]any{"placeId": "place-1"})
			return err
		},
		"GetDistancesForMode": func(c *Client) error {
			_, err := c.GetDistancesForMode(map[string]any{"mode": "driving", "places": []any{}})
			return err
		},
		"OptimizeRoute": func(c *Client) error {
			_, err := c.OptimizeRoute(map[string]any{"mode": "driving", "places": []any{}})
			return err
		},
		"GetRecommendedPlaces": func(c *Client) error {
			_, err := c.GetRecommendedPlaces(RecommendedPlacesRequest{
				TripPlanID: 1,
				GeoID:      12345,
				Input:      "things to do",
			})
			return err
		},
		"MarkRecommendationNotInterested": func(c *Client) error {
			return c.MarkRecommendationNotInterested(MarkRecommendationNotInterestedRequest{
				TripPlanID:  1,
				MapsPlaceID: "place-123",
			})
		},
		"AutocompletePlaces": func(c *Client) error {
			oldFn := autocompleteSessionToken
			autocompleteSessionToken = func() string { return "contract-token" }
			defer func() { autocompleteSessionToken = oldFn }()
			_, err := c.AutocompletePlaces("tokyo", 35.6812, 139.7671)
			return err
		},
		"FindPlaceFromLngLat": func(c *Client) error {
			_, err := c.FindPlaceFromLngLat(35.6812, 139.7671)
			return err
		},
		"GetMapLayerGroups": func(c *Client) error {
			_, err := c.GetMapLayerGroups(map[string]any{"layerGroupIds": []any{}})
			return err
		},
		"GetMultiplePlaceDetails": func(c *Client) error {
			_, err := c.GetMultiplePlaceDetails([]string{"place-1", "place-2"}, "en")
			return err
		},
		"GetPlaceDetailsV2": func(c *Client) error {
			_, err := c.GetPlaceDetailsV2("place-123", "en")
			return err
		},
		"SearchPlacesGoogle": func(c *Client) error {
			_, err := c.SearchPlacesGoogle(map[string]any{"input": "tokyo"})
			return err
		},
		"GetPlacesMetadata": func(c *Client) error {
			_, err := c.GetPlacesMetadata([]string{"place-1", "place-2"}, nil)
			return err
		},
		"GetPlaceCards": func(c *Client) error {
			_, err := c.GetPlaceCards([]string{"place-1", "place-2"})
			return err
		},
		"ListPlacePageGeos": func(c *Client) error {
			_, err := c.ListPlacePageGeos()
			return err
		},
		"GetTripPlanAssistantText": func(c *Client) error {
			_, err := c.GetTripPlanAssistantText(AssistantTextRequest{Message: "what should I do in Tokyo?"})
			return err
		},
		"GetTripPlanAssistantHighlights": func(c *Client) error {
			_, err := c.GetTripPlanAssistantHighlights(AssistantHighlightsRequest{
				AssistantMessage: "Visit Senso-ji",
				TripPlanID:       1,
			})
			return err
		},
		"GetTripPlanAssistantHistory": func(c *Client) error {
			_, err := c.GetTripPlanAssistantHistory(map[string]string{"chatId": "chat-1"})
			return err
		},
		"ListTripPlanAssistantChats": func(c *Client) error {
			_, err := c.ListTripPlanAssistantChats(1, "", 0, 0)
			return err
		},
		"GetTripPlanAssistantInitialChat": func(c *Client) error {
			_, err := c.GetTripPlanAssistantInitialChat(1)
			return err
		},
		"GetTripLikesBulk": func(c *Client) error {
			_, err := c.GetTripLikesBulk([]string{"trip-a", "trip-b"})
			return err
		},
		"CreateTripFromFlights": func(c *Client) error {
			_, err := c.CreateTripFromFlights(map[string]any{"flights": []any{}})
			return err
		},
		"GetMyProfileData": func(c *Client) error {
			_, err := c.GetMyProfileData()
			return err
		},
		"GetLodgingCheckoutData": func(c *Client) error {
			_, err := c.GetLodgingCheckoutData(map[string]string{"lodgingPropertyId": "property-123"})
			return err
		},
		"GetDealsForUser": func(c *Client) error {
			_, err := c.GetDealsForUser()
			return err
		},
		"RateEmail": func(c *Client) error {
			return c.RateEmail(RateEmailRequest{EmailID: 1, Rating: "thumbs_up"})
		},
	}

	for _, contract := range contracts {
		contract := contract
		t.Run(contract.Name, func(t *testing.T) {
			if !hasReferenceEndpoint(referenceEndpoints, contract.ReferencePath, contract.ReferenceMethod) {
				t.Fatalf("reference endpoint %q method %q not found", contract.ReferencePath, contract.ReferenceMethod)
			}

			runner := runners[contract.Name]
			if runner == nil {
				t.Fatalf("missing runner for %s", contract.Name)
			}

			var captured *capturedAPIRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = captureRequest(t, r)
				if contract.Name == "Login" {
					http.SetCookie(w, &http.Cookie{Name: "connect.sid", Value: "session"})
					http.SetCookie(w, &http.Cookie{Name: "XSRF-TOKEN", Value: "xsrf"})
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(responseForContract(contract.Name)))
			}))
			defer server.Close()

			oldBaseURL := BaseURL
			BaseURL = server.URL + "/api"
			t.Cleanup(func() { BaseURL = oldBaseURL })

			client := NewClient()
			if contract.Auth {
				client.SetAuth(&AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf", UserID: "1"})
			}

			if err := runner(client); err != nil {
				t.Fatalf("%s returned error: %v", contract.Name, err)
			}
			if captured == nil {
				t.Fatal("no request captured")
			}

			if captured.Method != contract.Method {
				t.Fatalf("method mismatch: got %s want %s", captured.Method, contract.Method)
			}
			if captured.Path != contract.Path {
				t.Fatalf("path mismatch: got %s want %s", captured.Path, contract.Path)
			}
			if !reflect.DeepEqual(captured.Query, contract.Query) {
				t.Fatalf("query mismatch:\ngot  %#v\nwant %#v", captured.Query, contract.Query)
			}
			if !reflect.DeepEqual(captured.Body, contract.Body) {
				t.Fatalf("body mismatch:\ngot  %#v\nwant %#v", captured.Body, contract.Body)
			}
			if contract.Auth {
				if got := captured.Header.Get("X-XSRF-TOKEN"); got != "xsrf" {
					t.Fatalf("missing X-XSRF-TOKEN header, got %q", got)
				}
				if got := captured.Header.Get("Cookie"); got == "" {
					t.Fatal("missing auth cookies")
				}
			} else {
				if got := captured.Header.Get("X-XSRF-TOKEN"); got != "" {
					t.Fatalf("unexpected X-XSRF-TOKEN header, got %q", got)
				}
			}
		})
	}
}

func loadRequestContracts(t *testing.T) []requestContract {
	t.Helper()
	var contracts []requestContract
	readJSONFixture(t, "artifacts/api-contracts/go_request_contracts.json", &contracts)
	return contracts
}

func loadReferenceEndpoints(t *testing.T) map[string]map[string]bool {
	t.Helper()
	var extracted struct {
		EndpointStrings []struct {
			URL    string `json:"url"`
			Source string `json:"source"`
		} `json:"endpointStrings"`
		WrappedAxios []struct {
			Method string `json:"method"`
			URL    string `json:"url"`
		} `json:"wrappedAxios"`
	}
	readJSONFixture(t, "artifacts/api-contracts/reference_calls.json", &extracted)
	endpoints := map[string]map[string]bool{}
	for _, item := range extracted.EndpointStrings {
		if item.URL != "" {
			addReferenceEndpoint(endpoints, item.URL, methodFromSource(item.Source))
		}
	}
	for _, item := range extracted.WrappedAxios {
		if item.URL != "" {
			addReferenceEndpoint(endpoints, item.URL, item.Method)
		}
	}
	return endpoints
}

func readJSONFixture(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", path))
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
}

func addReferenceEndpoint(endpoints map[string]map[string]bool, url, method string) {
	if endpoints[url] == nil {
		endpoints[url] = map[string]bool{}
	}
	endpoints[url][""] = true
	if method != "" {
		endpoints[url][method] = true
	}
}

func methodFromSource(source string) string {
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if containsMethodLiteral(source, method) {
			return method
		}
	}
	return ""
}

func containsMethodLiteral(source, method string) bool {
	return strings.Contains(source, `method: "`+method+`"`) ||
		strings.Contains(source, `method:"`+method+`"`) ||
		strings.Contains(source, `method: '`+method+`'`) ||
		strings.Contains(source, `method:'`+method+`'`)
}

func hasReferenceEndpoint(endpoints map[string]map[string]bool, expected, method string) bool {
	if methods := endpoints[expected]; len(methods) > 0 {
		return referenceMethodMatches(methods, method)
	}
	for endpoint, methods := range endpoints {
		if len(expected) > len(endpoint) && expected[:len(endpoint)] == endpoint {
			return referenceMethodMatches(methods, method)
		}
	}
	return false
}

func referenceMethodMatches(methods map[string]bool, method string) bool {
	if method == "" || methods[method] {
		return true
	}
	for existing := range methods {
		if existing != "" {
			return false
		}
	}
	return methods[""]
}

func captureRequest(t *testing.T, r *http.Request) *capturedAPIRequest {
	t.Helper()
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}
	var parsedBody any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &parsedBody); err != nil {
			t.Fatalf("parsing request body %q: %v", string(rawBody), err)
		}
	}
	query := map[string]any{}
	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			query[key] = values[0]
		} else {
			copied := make([]string, len(values))
			copy(copied, values)
			query[key] = copied
		}
	}
	if len(query) == 0 {
		query = nil
	}
	return &capturedAPIRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  query,
		Body:   parsedBody,
		Header: r.Header.Clone(),
	}
}

func responseForContract(name string) string {
	switch name {
	case "Login":
		return `{"success":true,"user":{"id":1,"email":"user@example.com","name":"User","username":"user"}}`
	case "GetMe", "FindUserByEmail":
		return `{"id":1,"email":"user@example.com","name":"User","username":"user"}`
	case "GetPlaceDetails":
		return `{"success":true,"data":{"details":{"name":"Tokyo Station","place_id":"place-123","geometry":{"location":{"lat":35.6812,"lng":139.7671}}},"cardData":{"placeId":"place-123"}}}`
	case "GetTripSections":
		return `{"success":true,"data":[]}`
	case "CreateTrip":
		return `{"success":true,"tripPlan":{"id":1,"key":"trip-key","title":"API Contract Trip"}}`
	case "CreateExampleTrip":
		return `{"success":true,"data":{"id":1,"key":"trip-key","title":"API Contract Trip"}}`
	case "CopyTrip":
		return `{"success":true,"data":{"id":2,"key":"copied-key","title":"Copied Trip"}}`
	case "ListTripInvites", "GetNotifications", "GetAllAirlines", "GetJournalStopPolylines":
		return `{"success":true,"data":[]}`
	case "GetNotificationSettings", "UpdateNotificationSettings":
		return `{"success":true,"notificationSettings":{"email":true}}`
	case "GetIfEdited":
		return `{"success":true,"tripPlans":[]}`
	case "GetTripUpdateRequired":
		return `{"success":true,"updateRequired":false}`
	case "GetTripDistinction":
		return `{"success":true,"distinction":"featured"}`
	case "GetGooglePriceRates":
		return `{"success":true,"data":{"propertyId":"property-123","rates":[]}}`
	case "ExportTrip":
		return `{"success":true,"data":{"exportUrl":"https://example.com/export"}}`
	case "ToggleChecklistItem":
		return `{"success":true,"data":{"section":{"id":7,"items":[]}}}`
	case "GetGlobalConfig":
		return `{"success":true,"config":{}}`
	case "SetLike", "GetFeedHome", "GetFeed", "GetFeedV2", "GetFeedMostRecent", "GetFriendsPlans", "BrowseGuides", "GetSessionStore":
		return `{"success":true,"data":true}`
	default:
		return `{"success":true,"tripPlan":{"id":1,"key":"trip-key","title":"API Contract Trip","likeCount":0},"resources":{},"data":[]}`
	}
}
