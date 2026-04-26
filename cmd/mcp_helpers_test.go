package cmd

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func testStringPtr(value string) *string {
	return &value
}

func TestFilterGeoGuideCounts(t *testing.T) {
	data := json.RawMessage(`{
		"geoGuideCounts": [
			{"name": "Japan", "geoId": 86647, "guideCount": 16},
			{"name": "Tokyo", "geoId": 1, "guideCount": 9},
			{"name": "Kyoto", "geoId": 2, "guideCount": 7}
		]
	}`)

	matches, err := filterGeoGuideCounts(data, "japan", 10)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, 86647, matches[0].GeoID)
	assert.Equal(t, "Japan", matches[0].Name)
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
