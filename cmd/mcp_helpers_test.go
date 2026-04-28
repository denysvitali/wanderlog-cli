package cmd

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

// loadAuthFromEnvOrKeychain loads authentication credentials from environment
// variables first, falling back to the system keychain or config file.
// Priority: env vars > keychain > config file
func loadAuthFromEnvOrKeychain() (*wanderlog.AuthCredentials, error) {
	// Try environment variables first
	sessionCookie := os.Getenv("WANDERLOG_AUTH_SESSION_COOKIE")
	xsrfToken := firstNonEmpty(os.Getenv("WANDERLOG_AUTH_XSRF_TOKEN"), os.Getenv("WANDERLOG_AUTH_SESSION_XSRF_TOKEN"))

	if sessionCookie != "" && xsrfToken != "" {
		return &wanderlog.AuthCredentials{
			SessionCookie: sessionCookie,
			XSRFToken:     xsrfToken,
			UserID:        os.Getenv("WANDERLOG_AUTH_USER_ID"),
		}, nil
	}

	email := os.Getenv("WANDERLOG_AUTH_EMAIL")
	password := os.Getenv("WANDERLOG_AUTH_PASSWORD")
	if email != "" && password != "" {
		client := wanderlog.NewClient()
		client.SetLogger(logger)
		return client.Login(email, password)
	}

	// Fall back to keychain
	creds, err := wanderlog.LoadCredentials()
	if err == nil && creds != nil {
		return creds, nil
	}

	// Fall back to config file
	if err := wanderlog.InitConfig(); err != nil {
		return nil, err
	}
	return wanderlog.LoadCredentialsFromConfig()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func testStringPtr(value string) *string {
	return &value
}

func TestFilterGeoGuideCounts(t *testing.T) {
	result := &wanderlog.GeoSearchResult{
		Countries: []wanderlog.GeoIDName{
			{ID: 86647, Name: "Japan"},
			{ID: 86648, Name: "Italy"},
		},
		Cities: []wanderlog.GeoIDName{
			{ID: 1, Name: "Tokyo"},
			{ID: 2, Name: "Kyoto"},
		},
	}

	matches := filterGeoGuideCounts(result, "japan", 10)
	require.Len(t, matches, 1)
	assert.Equal(t, 86647, matches[0].GeoID)
	assert.Equal(t, "Japan", matches[0].Name)
}

func TestTripHasAddedPlaceRequiresRequestedSection(t *testing.T) {
	var trip wanderlog.TripResponse
	require.NoError(t, json.Unmarshal([]byte(`{
		"success": true,
		"tripPlan": {
			"itinerary": {
				"sections": [
					{"id": 10, "blocks": [{"type": "place", "place": {"name": "Tokyo Tower", "place_id": "tokyo-tower"}}]},
					{"id": 11, "blocks": []}
				]
			}
		},
		"resources": {"placeMetadata": [{"name": "Tokyo Tower", "placeId": "tokyo-tower"}]}
	}`), &trip))

	assert.True(t, tripHasAddedPlace(&trip, 10, "Tokyo Tower", "tokyo-tower"))
	assert.False(t, tripHasAddedPlace(&trip, 11, "Tokyo Tower", "tokyo-tower"))
	assert.True(t, tripHasAddedPlace(&trip, 0, "Tokyo Tower", "tokyo-tower"))
}

func TestTripHasAddedFlightRequiresRequestedSectionAndDate(t *testing.T) {
	var trip wanderlog.TripResponse
	require.NoError(t, json.Unmarshal([]byte(`{
		"success": true,
		"tripPlan": {
			"itinerary": {
				"sections": [
					{"id": 10, "blocks": [{"type": "flight", "depart": {"date": "2026-05-11"}, "flightInfo": {"airline": {"iata": "MU"}, "number": 244}}]},
					{"id": 11, "blocks": [{"type": "flight", "depart": {"date": "2026-05-12"}, "flightInfo": {"airline": {"iata": "MU"}, "number": 575}}]}
				]
			}
		}
	}`), &trip))

	assert.True(t, tripHasAddedFlight(&trip, 10, "MU", 244, "2026-05-11"))
	assert.False(t, tripHasAddedFlight(&trip, 11, "MU", 244, "2026-05-11"))
	assert.False(t, tripHasAddedFlight(&trip, 10, "MU", 244, "2026-05-12"))
}

func TestFindFlightsSectionIDPrefersTypedFlightsSection(t *testing.T) {
	var trip wanderlog.TripResponse
	require.NoError(t, json.Unmarshal([]byte(`{
		"success": true,
		"tripPlan": {
			"itinerary": {
				"sections": [
					{"id": 10, "heading": "Flights", "type": "normal", "placeMarkerIcon": "map-marker", "blocks": []},
					{"id": 11, "heading": "Flights", "type": "flights", "placeMarkerIcon": "plane", "blocks": []}
				]
			}
		}
	}`), &trip))

	assert.Equal(t, 11, findFlightsSectionID(&trip))
}

func TestFindFlightsSectionIDFallsBackToManualFlightsShape(t *testing.T) {
	var trip wanderlog.TripResponse
	require.NoError(t, json.Unmarshal([]byte(`{
		"success": true,
		"tripPlan": {
			"itinerary": {
				"sections": [
					{"id": 10, "heading": "", "type": "normal", "placeMarkerIcon": "map-marker", "blocks": []},
					{"id": 12, "heading": "Flights", "mode": "placeList", "placeMarkerIcon": "plane", "blocks": []}
				]
			}
		}
	}`), &trip))

	assert.Equal(t, 12, findFlightsSectionID(&trip))
}

func TestResolveAddPlaceSectionIDRejectsImplicitUnscheduled(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "add_place",
			Arguments: map[string]any{
				"section_id": 0,
			},
		},
	}

	_, _, err := resolveAddPlaceSectionID(nil, "trip-key", request)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "section_id must be a positive itinerary section ID")
}

func TestResolveAddPlaceSectionIDAllowsExplicitUnscheduled(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "add_place",
			Arguments: map[string]any{
				"unscheduled": true,
			},
		},
	}

	sectionID, label, err := resolveAddPlaceSectionID(nil, "trip-key", request)
	require.NoError(t, err)
	assert.Equal(t, 0, sectionID)
	assert.Equal(t, "general Places to visit list", label)
}

func TestSplitFlightNumber(t *testing.T) {
	airline, number := splitFlightNumber(" mu 244 ")

	assert.Equal(t, "MU", airline)
	assert.Equal(t, 244, number)
}

func TestValidateFlightNumberRejectsMissingAirline(t *testing.T) {
	_, _, err := validateFlightNumber("244")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "airline code")
}

func TestValidateDateRangeRequiresBothDates(t *testing.T) {
	_, _, err := validateDateRange("start_date", "2026-05-11", "end_date", "", true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "end_date is required")
}

func TestValidateDateRangeRejectsReverseRange(t *testing.T) {
	_, _, err := validateDateRange("start_date", "2026-05-20", "end_date", "2026-05-11", true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "end_date must be on or after start_date")
}

func TestResolveSectionFromListByDate(t *testing.T) {
	sections := []wanderlog.ItSections{
		{ID: 1},
		{ID: 2, Date: testStringPtr("2026-05-11")},
	}

	sectionID, label, err := resolveSectionFromList(sections, 0, false, "2026-05-11", true)

	require.NoError(t, err)
	assert.Equal(t, 2, sectionID)
	assert.Equal(t, "2026-05-11 section ID 2", label)
}

func TestResolveSectionFromListRejectsUndatedExplicitSection(t *testing.T) {
	sections := []wanderlog.ItSections{
		{ID: 1},
		{ID: 2, Date: testStringPtr("2026-05-11")},
	}

	_, _, err := resolveSectionFromList(sections, 1, true, "", true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a dated itinerary section")
}

func TestValidateBlockSchemaRejectsPartialAirportStation(t *testing.T) {
	block := map[string]any{
		"type": "flight",
		"depart": map[string]any{
			"type": "airport",
			"date": "2026-05-11",
			"time": "10:00",
		},
		"arrive": map[string]any{
			"date": "2026-05-11",
			"time": "14:00",
		},
	}
	err := validateBlockSchema(block)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type=airport but no airport sub-object")
}

func TestValidateBlockSchemaRejectsAirportWithoutGooglePlace(t *testing.T) {
	block := map[string]any{
		"type": "flight",
		"depart": map[string]any{
			"type": "airport",
			"airport": map[string]any{
				"iata": "SFO",
			},
			"date": "2026-05-11",
			"time": "10:00",
		},
	}
	err := validateBlockSchema(block)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "airport.googlePlace is missing")
}

func TestValidateBlockSchemaAllowsNoTypeAirport(t *testing.T) {
	block := map[string]any{
		"type": "flight",
		"depart": map[string]any{
			"date": "2026-05-11",
			"time": "10:00",
		},
		"arrive": map[string]any{
			"date": "2026-05-11",
			"time": "14:00",
		},
	}
	err := validateBlockSchema(block)
	require.NoError(t, err)
}

func TestValidateBlockSchemaAllowsCompleteAirportStation(t *testing.T) {
	block := map[string]any{
		"type": "flight",
		"depart": map[string]any{
			"type": "airport",
			"airport": map[string]any{
				"googlePlace": map[string]any{
					"placeId": "ChIJE9SXBBhAjYARl8i Luij4BHk",
					"lat":     37.7749,
					"lng":     -122.4194,
				},
				"iata": "SFO",
			},
			"date": "2026-05-11",
			"time": "10:00",
		},
	}
	err := validateBlockSchema(block)
	require.NoError(t, err)
}
