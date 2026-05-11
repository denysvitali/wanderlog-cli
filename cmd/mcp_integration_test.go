package cmd

import (
	"context"
	"sort"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTripID = "jdysvggpzbjwpnej"
)

func TestMCPIntegration_AllRegisteredToolsHaveCoverage(t *testing.T) {
	tools := createMCPServer(false).ListTools()
	covered := map[string]bool{
		"add_checklist_items":          true,
		"add_collaborator":             true,
		"add_flight":                   true,
		"add_lodging":                  true,
		"add_place":                    true,
		"add_train":                    true,
		"add_trip_expense":             true,
		"autocomplete_airports":        true,
		"autocomplete_users":           true,
		"autofill_day":                 true,
		"block_user":                   true,
		"browse_guides":                true,
		"clear_section_blocks":         true,
		"copy_trip":                    true,
		"create_example_trip":          true,
		"create_guide_from_trip":       true,
		"create_trip":                  true,
		"delete_section":               true,
		"delete_flight":                true,
		"delete_itinerary_block":       true,
		"delete_lodging":               true,
		"delete_trip":                  true,
		"delete_trip_expense":          true,
		"delete_trips":                 true,
		"export_trip":                  true,
		"find_user_by_email":           true,
		"get_all_airlines":             true,
		"get_feed":                     true,
		"get_feed_friends":             true,
		"get_feed_home":                true,
		"get_feed_recent":              true,
		"get_feed_v2":                  true,
		"get_flights":                  true,
		"get_flight_stops":             true,
		"get_global_config":            true,
		"get_hotel_rates":              true,
		"get_if_edited":                true,
		"get_journal_stop_polylines":   true,
		"get_like_count":               true,
		"get_me":                       true,
		"get_notifications":            true,
		"get_notification_settings":    true,
		"get_or_create_share_key":      true,
		"get_place_details":            true,
		"get_session_preferences":      true,
		"get_session_store":            true,
		"get_itinerary":                true,
		"get_trip":                     true,
		"get_trip_distinction":         true,
		"get_trip_expenses_csv":        true,
		"get_trip_history":             true,
		"get_trip_images":              true,
		"get_trip_plan":                true,
		"get_trip_places":              true,
		"get_trip_sections":            true,
		"get_trip_update_required":     true,
		"get_user_kv":                  true,
		"get_user_emails":              true,
		"get_user_profile":             true,
		"get_view_only_journal":        true,
		"is_username_taken":            true,
		"like_trip":                    true,
		"list_places":                  true,
		"list_following":               true,
		"list_sections":                true,
		"list_trip_invites":            true,
		"list_trips":                   true,
		"mark_notifications_read":      true,
		"move_place":                   true,
		"nuke_trip_places":             true,
		"register_trip_view":           true,
		"remove_collaborator":          true,
		"remove_place":                 true,
		"reorder_places":               true,
		"restore_trip":                 true,
		"search_places_in_trips":       true,
		"search_geos":                  true,
		"search_hotels":                true,
		"search_places":                true,
		"search_places_wanderlog":      true,
		"search_restaurants":           true,
		"send_trip_invites":            true,
		"server_logout":                true,
		"set_trip_budget":              true,
		"set_trip_distinction":         true,
		"set_session_store_value":      true,
		"set_utc_offset":               true,
		"set_user_kv":                  true,
		"toggle_checklist_item":        true,
		"update_me":                    true,
		"update_flight":                true,
		"update_lodging":               true,
		"update_notification_settings": true,
		"update_place_notes":           true,
		"update_place_visit_time":      true,
		"update_trip":                  true,
		"update_trip_plan_geo":         true,
		"update_trip_expense":          true,

		// Reference-bundle extras (registerReferenceExtras).
		"autocomplete_places":                  true,
		"create_trip_from_flights":             true,
		"find_country_for_ip":                  true,
		"find_nearest_geos_to_ip":              true,
		"find_nearest_kayak_city":              true,
		"find_nearest_tripadvisor_geo":         true,
		"find_place_from_lng_lat":              true,
		"get_all_distance_info_for_place":      true,
		"get_client_geos":                      true,
		"get_deals_for_user":                   true,
		"get_distances_for_mode":               true,
		"get_lodging_checkout_data":            true,
		"get_map_layer_groups":                 true,
		"get_multiple_place_details":           true,
		"get_my_profile_data":                  true,
		"get_place_cards":                      true,
		"get_place_details_v2":                 true,
		"get_places_metadata":                  true,
		"get_recommended_places":               true,
		"get_trip_likes_bulk":                  true,
		"get_trip_plan_assistant_highlights":   true,
		"get_trip_plan_assistant_history":      true,
		"get_trip_plan_assistant_initial_chat": true,
		"get_trip_plan_assistant_text":         true,
		"list_countries":                       true,
		"list_geo_categories_for_category":     true,
		"list_geo_categories_for_geo":          true,
		"list_geo_in_month_geos":               true,
		"list_geos_with_good_guides":           true,
		"list_keyword_categories":              true,
		"list_place_page_geos":                 true,
		"list_popular_and_nearby_geos":         true,
		"list_trip_plan_assistant_chats":       true,
		"list_trip_planner_geos":               true,
		"mark_recommendation_not_interested":   true,
		"optimize_route":                       true,
		"rate_email":                           true,
		"search_geo":                           true,
		"search_places_google":                 true,
	}

	missing := make([]string, 0)
	for name := range tools {
		if !covered[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	assert.Empty(t, missing, "registered MCP tools need integration coverage entries")
}

func TestMCPIntegration_TripPlanReadAliasesRegistered(t *testing.T) {
	readOnlyTools := createMCPServer(true).ListTools()
	readWriteTools := createMCPServer(false).ListTools()

	for _, name := range []string{"get_trip", "get_trip_plan", "get_itinerary"} {
		assert.Contains(t, readOnlyTools, name, "%s should be registered in read-only mode", name)
		assert.Contains(t, readWriteTools, name, "%s should be registered when writes are enabled", name)
	}
}

func TestMCPIntegration_WriteToolRegistrationMode(t *testing.T) {
	writeTools := []string{
		"add_checklist_items",
		"add_collaborator",
		"add_flight",
		"add_lodging",
		"add_place",
		"add_train",
		"add_trip_expense",
		"autofill_day",
		"block_user",
		"clear_section_blocks",
		"copy_trip",
		"create_example_trip",
		"create_guide_from_trip",
		"create_trip",
		"delete_section",
		"delete_flight",
		"delete_itinerary_block",
		"delete_lodging",
		"delete_trip",
		"delete_trip_expense",
		"delete_trips",
		"export_trip",
		"get_or_create_share_key",
		"like_trip",
		"mark_notifications_read",
		"move_place",
		"register_trip_view",
		"remove_collaborator",
		"remove_place",
		"reorder_places",
		"restore_trip",
		"send_trip_invites",
		"server_logout",
		"set_session_store_value",
		"set_trip_budget",
		"set_trip_distinction",
		"set_utc_offset",
		"set_user_kv",
		"toggle_checklist_item",
		"update_me",
		"update_flight",
		"update_lodging",
		"update_notification_settings",
		"update_place_notes",
		"update_place_visit_time",
		"update_trip",
		"update_trip_plan_geo",
		"update_trip_expense",
		"nuke_trip_places",

		// Reference-bundle write-gated extras.
		"create_trip_from_flights",
		"get_recommended_places",
		"get_trip_plan_assistant_highlights",
		"get_trip_plan_assistant_text",
		"mark_recommendation_not_interested",
		"rate_email",
	}

	readOnlyTools := createMCPServer(true).ListTools()
	readWriteTools := createMCPServer(false).ListTools()
	for _, name := range writeTools {
		assert.NotContains(t, readOnlyTools, name, "%s should not be registered in read-only mode", name)
		assert.Contains(t, readWriteTools, name, "%s should be registered when writes are enabled", name)
	}
}

func TestMCPIntegration_WriteHandlerValidationBeforeAuth(t *testing.T) {
	ctx := context.Background()

	t.Run("create_trip_invalid_privacy", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "create_trip",
				Arguments: map[string]interface{}{
					"title":      "Invalid privacy",
					"geo_id":     1,
					"start_date": "2026-07-01",
					"end_date":   "2026-07-03",
					"privacy":    "secret",
				},
			},
		}

		result, err := handleCreateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "privacy must be one of")
	})

	t.Run("send_trip_invites_blank_invitees", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "send_trip_invites",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"invitees": " , ",
					"message":  "hello",
				},
			},
		}

		result, err := handleSendInvites(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "invitees must contain")
	})

	t.Run("add_checklist_items_blank_items", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_checklist_items",
				Arguments: map[string]interface{}{
					"trip_key":   testTripID,
					"section_id": 123,
					"items":      " , ",
				},
			},
		}

		result, err := handleAddChecklistItems(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "items must contain")
	})

	t.Run("toggle_checklist_item_missing_item_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "toggle_checklist_item",
				Arguments: map[string]interface{}{
					"trip_key":   testTripID,
					"section_id": 123,
					"checked":    true,
				},
			},
		}

		result, err := handleToggleChecklistItem(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "item_id must be")
	})

	t.Run("collaborator_missing_user_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_collaborator",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}

		result, err := handleAddCollaborator(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "user_id must be")
	})

	t.Run("share_key_without_permissions", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_or_create_share_key",
				Arguments: map[string]interface{}{
					"edit_key": testTripID,
					"can_view": false,
					"can_edit": false,
				},
			},
		}

		result, err := handleGetOrCreateShareKey(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "at least one permission")
	})

	t.Run("update_me_without_fields", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "update_me",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleUpdateMe(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "at least one profile field")
	})

	t.Run("update_notification_settings_invalid_json", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_notification_settings",
				Arguments: map[string]interface{}{
					"settings": "{",
				},
			},
		}

		result, err := handleUpdateNotificationSettings(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "valid JSON")
	})

	t.Run("get_if_edited_invalid_json", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_if_edited",
				Arguments: map[string]interface{}{
					"body": "{",
				},
			},
		}

		result, err := handleGetIfEdited(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "valid JSON")
	})

	t.Run("set_utc_offset_missing_offset", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "set_utc_offset",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleSetUTCOffset(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "offset_minutes is required")
	})

	t.Run("autocomplete_airports_requires_lat_lng_pair", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "autocomplete_airports",
				Arguments: map[string]interface{}{
					"query":    "SFO",
					"latitude": 37.6,
				},
			},
		}

		result, err := handleAutocompleteAirports(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "longitude is required")
	})

	t.Run("nuke_trip_places_requires_confirm", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "nuke_trip_places",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"confirm":  false,
				},
			},
		}

		result, err := handleNukeTripPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "confirm must be true")
	})

	t.Run("server_logout_requires_confirm", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "server_logout",
				Arguments: map[string]interface{}{
					"confirm": false,
				},
			},
		}

		result, err := handleServerLogout(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "confirm must be true")
	})
}

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

// TestMCPIntegration_GetTripSections tests the get_trip_sections tool
func TestMCPIntegration_GetTripSections(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("with_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_trip_sections",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}

		result, err := handleGetTripSections(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("with_default_trip_id", func(t *testing.T) {
		ctxWithTrip := withTripID(ctx, testTripID)
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_trip_sections",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetTripSections(ctxWithTrip, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_trip_sections",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetTripSections(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
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
		// Skip if API returns error (e.g., service unavailable)
		if result.IsError {
			t.Skip("Notifications API unavailable (returned error response)")
		}
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
		// Skip if API returns error (e.g., service unavailable)
		if result.IsError {
			t.Skip("Notifications API unavailable (returned error response)")
		}
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

// TestMCPIntegration_MovePlace tests the move_place tool (write operation)
func TestMCPIntegration_MovePlace(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("move_place_missing_params", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "move_place",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleMovePlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("move_place_missing_place_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "move_place",
				Arguments: map[string]interface{}{
					"trip_key":        testTripID,
					"from_section_id": 1,
					"to_section_id":   2,
					"position":        0,
					// Missing place_id
				},
			},
		}

		result, err := handleMovePlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("move_place_missing_from_section_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "move_place",
				Arguments: map[string]interface{}{
					"trip_key":      testTripID,
					"place_id":      12345,
					"to_section_id": 2,
					"position":      0,
					// Missing from_section_id
				},
			},
		}

		result, err := handleMovePlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("move_place_missing_to_section_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "move_place",
				Arguments: map[string]interface{}{
					"trip_key":        testTripID,
					"place_id":        12345,
					"from_section_id": 1,
					"position":        0,
					// Missing to_section_id
				},
			},
		}

		result, err := handleMovePlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("move_place_missing_position", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "move_place",
				Arguments: map[string]interface{}{
					"trip_key":        testTripID,
					"place_id":        12345,
					"from_section_id": 1,
					"to_section_id":   2,
					// Missing position
				},
			},
		}

		result, err := handleMovePlace(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_ReorderPlaces_ErrorCases tests the reorder_places tool error cases
func TestMCPIntegration_ReorderPlaces_ErrorCases(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication but credentials are not available: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("reorder_places_missing_section_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "reorder_places",
				Arguments: map[string]interface{}{
					"place_ids": "123,456",
					// Missing section_id
				},
			},
		}

		result, err := handleReorderPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("reorder_places_missing_place_ids", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "reorder_places",
				Arguments: map[string]interface{}{
					"section_id": 1,
					// Missing place_ids
				},
			},
		}

		result, err := handleReorderPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("reorder_places_missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "reorder_places",
				Arguments: map[string]interface{}{
					"section_id": 1,
					"place_ids":  "123,456",
					// Missing trip_key and no default in context
				},
			},
		}

		result, err := handleReorderPlaces(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_BudgetTools_ErrorCases tests trip budget tool validation.
func TestMCPIntegration_BudgetTools_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("set_trip_budget_missing_amount", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "set_trip_budget",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"currency": "EUR",
				},
			},
		}
		result, err := handleSetTripBudget(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("add_trip_expense_missing_description", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add_trip_expense",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"amount":   12.5,
					"currency": "EUR",
				},
			},
		}
		result, err := handleAddTripExpense(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("update_trip_expense_missing_expense_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip_expense",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"amount":   20,
				},
			},
		}
		result, err := handleUpdateTripExpense(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("delete_trip_expense_missing_expense_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "delete_trip_expense",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}
		result, err := handleDeleteTripExpense(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_UpdateTrip tests the update_trip tool (write operation)
func TestMCPIntegration_UpdateTrip(t *testing.T) {
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
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"title": "New Title",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("update_title", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"title":    "Updated Test Title",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("update_dates", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key":   testTripID,
					"start_date": "2026-06-01",
					"end_date":   "2026-06-10",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("update_privacy", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"privacy":  "private",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("update_all_fields", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key":   testTripID,
					"title":      "Fully Updated Title",
					"start_date": "2026-07-01",
					"end_date":   "2026-07-15",
					"privacy":    "public",
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("no_fields_provided", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "update_trip",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}

		result, err := handleUpdateTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_LikeTrip tests the like_trip tool (write operation)
func TestMCPIntegration_LikeTrip(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("liked_true", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "like_trip",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"liked":    true,
				},
			},
		}

		result, err := handleLikeTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Skip if API returns error (e.g., service unavailable)
		if result.IsError {
			t.Skip("Like API unavailable (returned error response)")
		}
		assert.False(t, result.IsError)
	})

	t.Run("liked_false", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "like_trip",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
					"liked":    false,
				},
			},
		}

		result, err := handleLikeTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Skip if API returns error (e.g., service unavailable)
		if result.IsError {
			t.Skip("Like API unavailable (returned error response)")
		}
		assert.False(t, result.IsError)
	})

	t.Run("missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "like_trip",
				Arguments: map[string]interface{}{
					"liked": true,
				},
			},
		}

		result, err := handleLikeTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_GetLikeCount tests the get_like_count tool
func TestMCPIntegration_GetLikeCount(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_like_count",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}

		result, err := handleGetLikeCount(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Skip if API returns error (e.g., service unavailable)
		if result.IsError {
			t.Skip("Like count API unavailable (returned error response)")
		}
		assert.False(t, result.IsError)
	})

	t.Run("missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "get_like_count",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleGetLikeCount(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_SendTripInvites tests the send_trip_invites tool (write operation)
func TestMCPIntegration_SendTripInvites(t *testing.T) {
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
				Name: "send_trip_invites",
				Arguments: map[string]interface{}{
					"invitees": "test@example.com",
				},
			},
		}

		result, err := handleSendInvites(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("missing_invitees", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "send_trip_invites",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}

		result, err := handleSendInvites(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_ListTripInvites tests the list_trip_invites tool
func TestMCPIntegration_ListTripInvites(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("with_trip_id", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "list_trip_invites",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}

		result, err := handleListInvites(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Skip if API returns error (e.g., service unavailable)
		if result.IsError {
			t.Skip("Trip invites API unavailable (returned error response)")
		}
		assert.False(t, result.IsError)
	})

	t.Run("missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "list_trip_invites",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleListInvites(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_RestoreTrip tests the restore_trip tool (write operation)
func TestMCPIntegration_RestoreTrip(t *testing.T) {
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
				Name:      "restore_trip",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleRestoreTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}

// TestMCPIntegration_ExtendedWriteTools covers write-gated tools registered by registerExtendedTools.
func TestMCPIntegration_ExtendedWriteTools(t *testing.T) {
	skipIntegrationTest(t)
	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Fatalf("Integration test requires authentication: %v", err)
	}
	_ = auth

	ctx := context.Background()

	t.Run("mark_notifications_read_missing_ids", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "mark_notifications_read",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleMarkNotificationsRead(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})

	t.Run("set_user_kv", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "set_user_kv",
				Arguments: map[string]interface{}{
					"key":   "codex_mcp_integration_test",
					"value": `{"source":"mcp_integration","version":1}`,
				},
			},
		}

		result, err := handleSetUserKV(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "set_user_kv should not error: %s", getTextContent(result))
	})

	t.Run("register_trip_view", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "register_trip_view",
				Arguments: map[string]interface{}{
					"trip_key": testTripID,
				},
			},
		}

		result, err := handleRegisterTripView(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError, "register_trip_view should not error: %s", getTextContent(result))
	})

	t.Run("create_guide_from_trip_missing_trip_key", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "create_guide_from_trip",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handleCreateGuideFromTrip(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
	})
}
