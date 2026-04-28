package wanderlog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlexibleText_StringVariant tests unmarshaling a string as FlexibleText
func TestFlexibleText_StringVariant(t *testing.T) {
	jsonData := `"simple text"`
	var ft FlexibleText
	err := json.Unmarshal([]byte(jsonData), &ft)
	require.NoError(t, err)
	assert.True(t, ft.IsString)
	assert.Equal(t, "simple text", ft.String)
	assert.Equal(t, Text{}, ft.Text)
}

// TestFlexibleText_ObjectVariant tests unmarshaling an object as FlexibleText
func TestFlexibleText_ObjectVariant(t *testing.T) {
	jsonData := `{"ops":[{"insert":"hello\n"}]}`
	var ft FlexibleText
	err := json.Unmarshal([]byte(jsonData), &ft)
	require.NoError(t, err)
	assert.False(t, ft.IsString)
	assert.Equal(t, "", ft.String)
	require.Len(t, ft.Text.Ops, 1)
	assert.Equal(t, "hello\n", ft.Text.Ops[0].Insert)
}

// TestFlexibleText_RoundTrip tests marshal after unmarshal
func TestFlexibleText_RoundTrip(t *testing.T) {
	original := FlexibleText{
		IsString: false,
		Text: Text{
			Ops: []struct {
				Attributes *struct {
					Bold bool   `json:"bold,omitempty"`
					Link string `json:"link,omitempty"`
					List string `json:"list,omitzero"`
				} `json:"attributes,omitempty,omitzero"`
				Insert string `json:"insert,omitzero"`
			}{{Insert: "test\n"}},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var parsed FlexibleText
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, original.IsString, parsed.IsString)
	require.Len(t, parsed.Text.Ops, 1)
	assert.Equal(t, "test\n", parsed.Text.Ops[0].Insert)
}

// TestFlexibleText_StringRoundTrip tests marshal and unmarshal for string variant
func TestFlexibleText_StringRoundTrip(t *testing.T) {
	original := FlexibleText{
		IsString: true,
		String:   "simple string value",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var parsed FlexibleText
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.True(t, parsed.IsString)
	assert.Equal(t, "simple string value", parsed.String)
}

// TestAirlinesResponse_JSON tests unmarshaling an AirlinesResponse
func TestAirlinesResponse_JSON(t *testing.T) {
	jsonData := `{
		"success": true,
		"data": [
			{
				"iata": "UA",
				"icao": "UAL",
				"name": "United Airlines",
				"localizedName": "United"
			}
		]
	}`

	var resp AirlinesResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "UA", resp.Data[0].Iata)
	assert.Equal(t, "UAL", resp.Data[0].Icao)
	assert.Equal(t, "United Airlines", resp.Data[0].Name)
}

// TestFlightStopsResponse_JSON tests unmarshaling a FlightStopsResponse
func TestFlightStopsResponse_JSON(t *testing.T) {
	jsonData := `{
		"success": true,
		"data": [
			{
				"depart": {
					"type": "depart",
					"date": "2024-06-15",
					"time": "08:00",
					"airport": {
						"iata": "SFO",
						"name": "San Francisco International",
						"cityName": "San Francisco"
					}
				},
				"arrive": {
					"type": "arrive",
					"date": "2024-06-15",
					"time": "11:30",
					"airport": {
						"iata": "LAX",
						"name": "Los Angeles International",
						"cityName": "Los Angeles"
					}
				}
			}
		]
	}`

	var resp FlightStopsResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	require.Len(t, resp.Data, 1)

	depart := resp.Data[0].Depart
	assert.Equal(t, "depart", depart.Type)
	assert.Equal(t, "2024-06-15", depart.Date)
	assert.Equal(t, "08:00", depart.Time)
	assert.Equal(t, "SFO", depart.Airport.IATA)

	arrive := resp.Data[0].Arrive
	assert.Equal(t, "arrive", arrive.Type)
	assert.Equal(t, "2024-06-15", arrive.Date)
	assert.Equal(t, "11:30", arrive.Time)
	assert.Equal(t, "LAX", arrive.Airport.IATA)
}

// TestLodgingSearchResponse_JSON tests unmarshaling a LodgingSearchResponse
func TestLodgingSearchResponse_JSON(t *testing.T) {
	jsonData := `{
		"success": true,
		"data": [
			{
				"propertyId": "hotel123",
				"name": "Grand Hotel",
				"address": "123 Main St",
				"city": "New York",
				"country": "USA",
				"rating": 4.5,
				"pricePerNight": "$200",
				"currency": "USD",
				"imageUrl": "https://example.com/hotel.jpg"
			}
		]
	}`

	var resp LodgingSearchResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "hotel123", resp.Data[0].PropertyID)
	assert.Equal(t, "Grand Hotel", resp.Data[0].Name)
	assert.Equal(t, "New York", resp.Data[0].City)
	assert.Equal(t, 4.5, resp.Data[0].Rating)
	assert.Equal(t, "$200", resp.Data[0].PricePerNight)
	assert.Equal(t, "USD", resp.Data[0].Currency)
}

// TestPlaceSearchResponse_JSON tests unmarshaling a PlaceSearchResponse
func TestPlaceSearchResponse_JSON(t *testing.T) {
	jsonData := `{
		"success": true,
		"places": [
			{
				"id": "place1",
				"name": "Central Park",
				"address": "New York, NY",
				"place_id": "ChIJnYLx0Ziskan8RJCGHA5AJ3Fs",
				"latitude": 40.7829,
				"longitude": -73.9654,
				"rating": 4.8,
				"categories": ["park", "tourist_attraction"],
				"description": "A large public park"
			}
		]
	}`

	var resp PlaceSearchResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	require.Len(t, resp.Places, 1)
	assert.Equal(t, "place1", resp.Places[0].ID)
	assert.Equal(t, "Central Park", resp.Places[0].Name)
	assert.Equal(t, 40.7829, resp.Places[0].Latitude)
	assert.Equal(t, -73.9654, resp.Places[0].Longitude)
	assert.Equal(t, 4.8, resp.Places[0].Rating)
}

// TestAirlinesResponse_Failure tests unmarshaling a failed AirlinesResponse
func TestAirlinesResponse_Failure(t *testing.T) {
	jsonData := `{"success": false, "data": []}`

	var resp AirlinesResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Len(t, resp.Data, 0)
}

// TestFlightStopsResponse_EmptyData tests unmarshaling FlightStopsResponse with empty data
func TestFlightStopsResponse_EmptyData(t *testing.T) {
	jsonData := `{"success": true, "data": []}`

	var resp FlightStopsResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 0)
}