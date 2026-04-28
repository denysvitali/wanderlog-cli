package cmd

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTripID = "jdysvggpzbjwpnej"
)

// TestMCPIntegration_ListTrips tests the list_trips tool
func TestMCPIntegration_ListTrips(t *testing.T) {
	skipIntegrationTest(t)
	// Fail if not authenticated (credentials are required for integration tests)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_trips",
			Arguments: map[string]interface{}{
				"format": "json",
			},
		},
	}

	result, err := handleListTrips(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_GetTrip tests the get_trip tool
func TestMCPIntegration_GetTrip(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		result, err := handleGetTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip",
				Arguments: map[string]interface{}{
					"format": "json",
				},
			},
		}

		result, err := handleGetTrip(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_trip",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_ListPlaces tests the list_places tool
func TestMCPIntegration_ListPlaces(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_places",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		result, err := handleListPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_places",
				Arguments: map[string]interface{}{
					"format": "default",
				},
			},
		}

		result, err := handleListPlaces(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

// TestMCPIntegration_ListSections tests the list_sections tool
func TestMCPIntegration_ListSections(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_sections",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		result, err := handleListSections(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_sections",
				Arguments: map[string]interface{}{
					"format": "default",
				},
			},
		}

		result, err := handleListSections(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

// TestMCPIntegration_GetPlaceDetails tests the get_place_details tool
func TestMCPIntegration_GetPlaceDetails(t *testing.T) {
	skipIntegrationTest(t)
	// This doesn't require authentication
	ctx := context.Background()

	t.Run("valid_place_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_place_details",
				Arguments: map[string]interface{}{
					"place_id": "ChIJN1t_tDeuEmsRUsoyG83frY4", // Google Sydney Office
					"format":   "json",
				},
			},
		}

		result, err := handleGetPlaceDetails(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Note: This might fail if the API is down, but should work most of the time
	})

	t.Run("missing_place_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_place_details",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetPlaceDetails(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_SearchPlacesWanderlog tests the search_places_wanderlog tool
func TestMCPIntegration_SearchPlacesWanderlog(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("search_paris", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_places_wanderlog",
				Arguments: map[string]interface{}{
					"query":  "Eiffel Tower",
					"format": "json",
				},
			},
		}

		result, err := handleSearchPlacesWanderlog(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("search_with_coordinates", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_places_wanderlog",
				Arguments: map[string]interface{}{
					"query":     "restaurants",
					"latitude":  48.8566,
					"longitude": 2.3522,
					"format":    "default",
				},
			},
		}

		result, err := handleSearchPlacesWanderlog(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "search_places_wanderlog",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleSearchPlacesWanderlog(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_TripResource tests the trip resource handler
func TestMCPIntegration_TripResource(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("valid_trip_uri", func(t *testing.T) {
		request := mcp.ReadResourceRequest{
			Params: mcp.ReadResourceParams{
				URI: "wanderlog://trips/" + testTripID,
			},
		}

		contents, err := handleTripResource(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, contents)
		require.Len(t, contents, 1)

		textContent, ok := contents[0].(mcp.TextResourceContents)
		require.True(t, ok)
		assert.Equal(t, "application/json", textContent.MIMEType)
		assert.NotEmpty(t, textContent.Text)
	})

	t.Run("invalid_trip_uri", func(t *testing.T) {
		request := mcp.ReadResourceRequest{
			Params: mcp.ReadResourceParams{
				URI: "wanderlog://invalid",
			},
		}

		_, err := handleTripResource(ctx, request)
		require.Error(t, err)
	})
}

// TestMCPIntegration_AnalyzeTrip tests the analyze_trip prompt
func TestMCPIntegration_AnalyzeTrip(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	testCases := []struct {
		name  string
		focus string
	}{
		{"overall_analysis", "overall"},
		{"budget_focus", "budget"},
		{"itinerary_focus", "itinerary"},
		{"places_focus", "places"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := mcp.GetPromptRequest{
				Params: mcp.GetPromptParams{
					Name: "analyze_trip",
					Arguments: map[string]string{
						"trip_id": testTripID,
						"focus":   tc.focus,
					},
				},
			}

			result, err := handleAnalyzeTrip(ctx, request)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.Description)
			assert.NotEmpty(t, result.Messages)
			assert.Len(t, result.Messages, 1)
			assert.Equal(t, mcp.RoleUser, result.Messages[0].Role)
		})
	}

	t.Run("missing_trip_id", func(t *testing.T) {
		request := mcp.GetPromptRequest{
			Params: mcp.GetPromptParams{
				Name:      "analyze_trip",
				Arguments: map[string]string{},
			},
		}

		_, err := handleAnalyzeTrip(ctx, request)
		require.Error(t, err)
	})
}

// TestMCPIntegration_ContextWithTripID tests the context trip ID functionality
func TestMCPIntegration_ContextWithTripID(t *testing.T) {
	ctx := context.Background()

	// Test without trip ID
	tripID, ok := tripIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, tripID)

	// Test with trip ID
	ctxWithTrip := withTripID(ctx, testTripID)
	tripID, ok = tripIDFromContext(ctxWithTrip)
	assert.True(t, ok)
	assert.Equal(t, testTripID, tripID)
}

// TestMCPIntegration_GetMe tests the get_me tool
func TestMCPIntegration_GetMe(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_me",
		},
	}

	result, err := handleGetMe(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_GetUserProfile tests the get_user_profile tool
func TestMCPIntegration_GetUserProfile(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("by_username", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_user_profile",
				Arguments: map[string]interface{}{
					"target": "@denysvitali",
				},
			},
		}

		result, err := handleGetUserProfile(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_target", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_user_profile",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetUserProfile(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetNotifications tests the get_notifications tool
func TestMCPIntegration_GetNotifications(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("default_offset", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_notifications",
			},
		}

		result, err := handleGetNotifications(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_offset", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_notifications",
				Arguments: map[string]interface{}{
					"offset": 10,
				},
			},
		}

		result, err := handleGetNotifications(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

// TestMCPIntegration_GetNotificationSettings tests the get_notification_settings tool
func TestMCPIntegration_GetNotificationSettings(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_notification_settings",
		},
	}

	result, err := handleGetNotificationSettings(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_GetUserEmails tests the get_user_emails tool
func TestMCPIntegration_GetUserEmails(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_user_emails",
		},
	}

	result, err := handleGetUserEmails(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_AutocompleteUsers tests the autocomplete_users tool
func TestMCPIntegration_AutocompleteUsers(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("valid_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "autocomplete_users",
				Arguments: map[string]interface{}{
					"query": "denys",
				},
			},
		}

		result, err := handleAutocompleteUsers(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "autocomplete_users",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleAutocompleteUsers(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_IsUsernameTaken tests the is_username_taken tool
func TestMCPIntegration_IsUsernameTaken(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("username_not_taken", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "is_username_taken",
				Arguments: map[string]interface{}{
					"username": "this_username_definitely_does_not_exist_12345",
				},
			},
		}

		result, err := handleIsUsernameTaken(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_username", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "is_username_taken",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleIsUsernameTaken(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_SearchPlaces tests the search_places tool
func TestMCPIntegration_SearchPlaces(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("search_eiffel_tower", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_places",
				Arguments: map[string]interface{}{
					"query":  "Eiffel Tower",
					"format": "json",
				},
			},
		}

		result, err := handleSearchPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("search_with_coordinates", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_places",
				Arguments: map[string]interface{}{
					"query":     "coffee shop",
					"latitude":  48.8566,
					"longitude": 2.3522,
					"format":    "default",
				},
			},
		}

		result, err := handleSearchPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "search_places",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleSearchPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_SearchRestaurants tests the search_restaurants tool
func TestMCPIntegration_SearchRestaurants(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("search_ramen_tokyo", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_restaurants",
				Arguments: map[string]interface{}{
					"query":  "ramen",
					"format": "json",
				},
			},
		}

		result, err := handleSearchRestaurants(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("search_with_coordinates", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_restaurants",
				Arguments: map[string]interface{}{
					"query":     "sushi",
					"latitude":  35.6762,
					"longitude": 139.6503,
					"format":    "default",
				},
			},
		}

		result, err := handleSearchRestaurants(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "search_restaurants",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleSearchRestaurants(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetFlights tests the get_flights tool
func TestMCPIntegration_GetFlights(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_flights",
				Arguments: map[string]interface{}{
					"trip_id": testTripID,
					"format":  "json",
				},
			},
		}

		result, err := handleGetFlights(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_flights",
				Arguments: map[string]interface{}{
					"format": "json",
				},
			},
		}

		result, err := handleGetFlights(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_flights",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetFlights(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetFlightStops tests the get_flight_stops tool
func TestMCPIntegration_GetFlightStops(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("valid_flight", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_flight_stops",
				Arguments: map[string]interface{}{
					"flight_number": "244",
					"airline":       "MU",
					"date":          "2024-06-15",
				},
			},
		}

		result, err := handleGetFlightStops(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_flight_number", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_flight_stops",
				Arguments: map[string]interface{}{
					"airline": "UA",
					"date":    "2024-06-15",
				},
			},
		}

		result, err := handleGetFlightStops(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("missing_airline", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_flight_stops",
				Arguments: map[string]interface{}{
					"flight_number": "100",
					"date":          "2024-06-15",
				},
			},
		}

		result, err := handleGetFlightStops(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("missing_date", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_flight_stops",
				Arguments: map[string]interface{}{
					"flight_number": "244",
					"airline":       "MU",
				},
			},
		}

		result, err := handleGetFlightStops(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetFeedHome tests the get_feed_home tool
func TestMCPIntegration_GetFeedHome(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_feed_home",
		},
	}

	result, err := handleGetFeedHome(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_GetFeedRecent tests the get_feed_recent tool
func TestMCPIntegration_GetFeedRecent(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_feed_recent",
		},
	}

	result, err := handleGetFeedRecent(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_GetFeedFriends tests the get_feed_friends tool
func TestMCPIntegration_GetFeedFriends(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_feed_friends",
		},
	}

	result, err := handleGetFeedFriends(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

// TestMCPIntegration_GetTripHistory tests the get_trip_history tool
func TestMCPIntegration_GetTripHistory(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("default_offset", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip_history",
			},
		}

		result, err := handleGetTripHistory(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_offset", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip_history",
				Arguments: map[string]interface{}{
					"offset": 10,
				},
			},
		}

		result, err := handleGetTripHistory(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

// TestMCPIntegration_BrowseGuides tests the browse_guides tool
func TestMCPIntegration_BrowseGuides(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("without_geo_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "browse_guides",
			},
		}

		result, err := handleBrowseGuides(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_geo_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "browse_guides",
				Arguments: map[string]interface{}{
					"geo_id": 86667, // Japan
				},
			},
		}

		result, err := handleBrowseGuides(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

// TestMCPIntegration_SearchGeos tests the search_geos tool
func TestMCPIntegration_SearchGeos(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("valid_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "search_geos",
				Arguments: map[string]interface{}{
					"query": "Japan",
					"limit": 5,
				},
			},
		}

		result, err := handleSearchGeos(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_query", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "search_geos",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleSearchGeos(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetViewOnlyJournal tests the get_view_only_journal tool
func TestMCPIntegration_GetViewOnlyJournal(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("missing_journal_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_view_only_journal",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetViewOnlyJournal(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetTripExpensesCSV tests the get_trip_expenses_csv tool
func TestMCPIntegration_GetTripExpensesCSV(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_trip_expenses_csv",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetTripExpensesCSV(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetTripDistinction tests the get_trip_distinction tool
func TestMCPIntegration_GetTripDistinction(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	t.Run("missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_trip_distinction",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetTripDistinction(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetGlobalConfig tests the get_global_config tool
// This tool does NOT require authentication
func TestMCPIntegration_GetGlobalConfig(t *testing.T) {
	skipIntegrationTest(t)
	ctx := context.Background()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_global_config",
		},
	}

	result, err := handleGetGlobalConfig(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}
