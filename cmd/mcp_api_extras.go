package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

// registerReferenceExtras adds MCP tools for the geo, directions,
// recommendations, places, placesAPI, chat, and miscellaneous endpoints
// discovered in the reference bundle.
func registerReferenceExtras(s *server.MCPServer, readOnly bool) {
	// === Geo (read-only) ===
	s.AddTool(mcp.NewTool("list_geos_with_good_guides",
		mcp.WithDescription("List curated geos that have high-quality Wanderlog guides")),
		handleListGeosWithGoodGuides)
	s.AddTool(mcp.NewTool("list_popular_and_nearby_geos",
		mcp.WithDescription("List popular and nearby geos for the calling user")),
		handleListPopularAndNearbyGeos)
	s.AddTool(mcp.NewTool("find_country_for_ip",
		mcp.WithDescription("Return the country geo associated with the caller's IP address")),
		handleFindCountryForIP)
	s.AddTool(mcp.NewTool("find_nearest_tripadvisor_geo",
		mcp.WithDescription("Find the nearest Tripadvisor-mapped geo for a coordinate"),
		mcp.WithNumber("latitude", mcp.Required(), mcp.Description("Latitude")),
		mcp.WithNumber("longitude", mcp.Required(), mcp.Description("Longitude"))),
		handleFindNearestTripadvisorGeo)
	s.AddTool(mcp.NewTool("find_nearest_geos_to_ip",
		mcp.WithDescription("Find the nearest geos for the caller's IP address")),
		handleFindNearestGeosToIP)
	s.AddTool(mcp.NewTool("find_nearest_kayak_city",
		mcp.WithDescription("Find the nearest Kayak city for a coordinate (used for hotel lookups)"),
		mcp.WithNumber("latitude", mcp.Required(), mcp.Description("Latitude")),
		mcp.WithNumber("longitude", mcp.Required(), mcp.Description("Longitude")),
		mcp.WithString("city_name_to_match", mcp.Required(), mcp.Description("City name to match"))),
		handleFindNearestKayakCity)
	s.AddTool(mcp.NewTool("get_client_geos",
		mcp.WithDescription("Fetch metadata for one or more geos by ID"),
		mcp.WithString("geo_ids", mcp.Required(), mcp.Description("Comma-separated geo IDs"))),
		handleGetClientGeos)
	s.AddTool(mcp.NewTool("list_trip_planner_geos",
		mcp.WithDescription("List geos that have trip-planner content available")),
		handleListTripPlannerGeos)
	s.AddTool(mcp.NewTool("list_countries",
		mcp.WithDescription("List countries"),
		mcp.WithString("language", mcp.Description("UI language code (default en)"), mcp.DefaultString("en"))),
		handleListCountries)
	s.AddTool(mcp.NewTool("list_geo_categories_for_category",
		mcp.WithDescription("List geos that have content for the given keyword category"),
		mcp.WithNumber("keyword_category_id", mcp.Required(), mcp.Description("Keyword category ID")),
		mcp.WithString("language", mcp.Description("UI language code (default en)"), mcp.DefaultString("en"))),
		handleListGeoCategoriesForCategory)
	s.AddTool(mcp.NewTool("list_geo_categories_for_geo",
		mcp.WithDescription("List the keyword categories that have content for the given geo"),
		mcp.WithNumber("geo_id", mcp.Required(), mcp.Description("Geo ID")),
		mcp.WithString("source", mcp.Description("Source channel (e.g. tripPlanner)"), mcp.DefaultString("tripPlanner"))),
		handleListGeoCategoriesForGeo)
	s.AddTool(mcp.NewTool("list_geo_in_month_geos",
		mcp.WithDescription("List 'best places to visit in <month>' content")),
		handleListGeoInMonthGeos)
	s.AddTool(mcp.NewTool("list_keyword_categories",
		mcp.WithDescription("List the top-level keyword taxonomy"),
		mcp.WithString("language", mcp.Description("UI language code (default en)"), mcp.DefaultString("en"))),
		handleListKeywordCategories)
	s.AddTool(mcp.NewTool("search_geo",
		mcp.WithDescription("Free-form search of geos"),
		mcp.WithString("q", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("language", mcp.Description("UI language code"))),
		handleSearchGeo)

	// === Directions (read-only) ===
	s.AddTool(mcp.NewTool("get_all_distance_info_for_place",
		mcp.WithDescription("Get all-mode distance info from one place to a list of others"),
		mcp.WithString("payload", mcp.Required(), mcp.Description("JSON payload (placeId, places, ...)"))),
		handleGetAllDistanceInfoForPlace)
	s.AddTool(mcp.NewTool("get_distances_for_mode",
		mcp.WithDescription("Get pairwise distances for a fixed travel mode"),
		mcp.WithString("payload", mcp.Required(), mcp.Description("JSON payload (mode, places, ...)"))),
		handleGetDistancesForMode)
	s.AddTool(mcp.NewTool("optimize_route",
		mcp.WithDescription("Compute an optimized ordering for a list of places"),
		mcp.WithString("payload", mcp.Required(), mcp.Description("JSON payload (mode, places, ...)"))),
		handleOptimizeRoute)

	// === PlacesAPI (read-only) ===
	s.AddTool(mcp.NewTool("autocomplete_places",
		mcp.WithDescription("Google-style place autocomplete (v1)"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("latitude", mcp.Description("Bias latitude")),
		mcp.WithNumber("longitude", mcp.Description("Bias longitude"))),
		handleAutocompletePlaces)
	s.AddTool(mcp.NewTool("find_place_from_lng_lat",
		mcp.WithDescription("Reverse-geocode a coordinate to a Wanderlog place suggestion"),
		mcp.WithNumber("latitude", mcp.Required(), mcp.Description("Latitude")),
		mcp.WithNumber("longitude", mcp.Required(), mcp.Description("Longitude"))),
		handleFindPlaceFromLngLat)
	s.AddTool(mcp.NewTool("get_map_layer_groups",
		mcp.WithDescription("Get places grouped into map category layers for a viewport / category set"),
		mcp.WithString("payload", mcp.Required(), mcp.Description("JSON payload (bbox, layerGroupIds, ...)"))),
		handleGetMapLayerGroups)
	s.AddTool(mcp.NewTool("get_multiple_place_details",
		mcp.WithDescription("Fetch detail records for multiple Google place IDs in a single call"),
		mcp.WithString("place_ids", mcp.Required(), mcp.Description("Comma-separated place IDs")),
		mcp.WithString("language", mcp.Description("UI language code"), mcp.DefaultString("en"))),
		handleGetMultiplePlaceDetails)
	s.AddTool(mcp.NewTool("get_place_details_v2",
		mcp.WithDescription("Get the detail-only view for a single place (no card data)"),
		mcp.WithString("place_id", mcp.Required(), mcp.Description("Place ID")),
		mcp.WithString("language", mcp.Description("UI language code"), mcp.DefaultString("en"))),
		handleGetPlaceDetailsV2)
	s.AddTool(mcp.NewTool("search_places_google",
		mcp.WithDescription("Google-style places text search"),
		mcp.WithString("payload", mcp.Required(), mcp.Description("JSON payload (input/query, location, ...)"))),
		handleSearchPlacesGoogle)

	// === Places (cards/metadata) (read-only) ===
	s.AddTool(mcp.NewTool("get_places_metadata",
		mcp.WithDescription("Bulk fetch Google place metadata for one or more place IDs"),
		mcp.WithString("place_ids", mcp.Required(), mcp.Description("Comma-separated place IDs")),
		mcp.WithNumber("geo_id", mcp.Description("Optional geo scope")),
		mcp.WithString("list_id", mcp.Description("Optional list ID")),
		mcp.WithString("list_type", mcp.Description("Optional list type")),
		mcp.WithBoolean("get_details", mcp.Description("Include extended details"))),
		handleGetPlacesMetadata)
	s.AddTool(mcp.NewTool("get_place_cards",
		mcp.WithDescription("Bulk fetch lightweight place card data"),
		mcp.WithString("place_ids", mcp.Required(), mcp.Description("Comma-separated place IDs"))),
		handleGetPlaceCards)
	s.AddTool(mcp.NewTool("list_place_page_geos",
		mcp.WithDescription("List geos that have a Wanderlog 'place page' available")),
		handleListPlacePageGeos)

	// === Chat assistant (read-only get/history; getText still read but auth-gated) ===
	s.AddTool(mcp.NewTool("get_trip_plan_assistant_history",
		mcp.WithDescription("Get prior assistant messages for a chat (pass query params)"),
		mcp.WithString("chat_id", mcp.Description("Chat ID")),
		mcp.WithString("page_size", mcp.Description("Page size")),
		mcp.WithString("sent_at_before", mcp.Description("Pagination cursor (millis)"))),
		handleGetTripPlanAssistantHistory)
	s.AddTool(mcp.NewTool("list_trip_plan_assistant_chats",
		mcp.WithDescription("List assistant chat threads for a trip plan"),
		mcp.WithNumber("trip_plan_id", mcp.Required(), mcp.Description("Trip plan ID")),
		mcp.WithString("search", mcp.Description("Search query")),
		mcp.WithNumber("last_item_is_before_millis", mcp.Description("Pagination cursor (millis)")),
		mcp.WithNumber("page_size", mcp.Description("Page size"))),
		handleListTripPlanAssistantChats)
	s.AddTool(mcp.NewTool("get_trip_plan_assistant_initial_chat",
		mcp.WithDescription("Get the seeded initial assistant chat for a trip plan"),
		mcp.WithNumber("trip_plan_id", mcp.Required(), mcp.Description("Trip plan ID"))),
		handleGetTripPlanAssistantInitialChat)

	// === Trip extras (read-only) ===
	s.AddTool(mcp.NewTool("get_trip_likes_bulk",
		mcp.WithDescription("Get the like state for multiple trips at once"),
		mcp.WithString("trip_keys", mcp.Required(), mcp.Description("Comma-separated trip keys"))),
		handleGetTripLikesBulk)
	s.AddTool(mcp.NewTool("get_my_profile_data",
		mcp.WithDescription("Get the authenticated user's profile dashboard")),
		handleGetMyProfileData)

	// === Lodging (read-only) ===
	s.AddTool(mcp.NewTool("get_lodging_checkout_data",
		mcp.WithDescription("Get pre-checkout pricing/policy data for a lodging offer"),
		mcp.WithString("payload", mcp.Required(), mcp.Description("JSON object of query params (lodgingPropertyId, dates, guests, currencyCode, ...)"))),
		handleGetLodgingCheckoutData)

	// === Misc (read-only) ===
	s.AddTool(mcp.NewTool("get_deals_for_user",
		mcp.WithDescription("Get user-targeted travel deals")),
		handleGetDealsForUser)

	// === Write-gated ===
	if readOnly {
		return
	}
	s.AddTool(mcp.NewTool("get_recommended_places",
		mcp.WithDescription("Get recommended places for a trip + geo (uses recommendations/v2)"),
		mcp.WithNumber("trip_plan_id", mcp.Required(), mcp.Description("Trip plan ID")),
		mcp.WithNumber("geo_id", mcp.Required(), mcp.Description("Geo ID")),
		mcp.WithString("input", mcp.Description("Free-text intent (e.g. 'things to do')")),
		mcp.WithString("excluding_place_ids", mcp.Description("Comma-separated place IDs to exclude"))),
		handleGetRecommendedPlaces)
	s.AddTool(mcp.NewTool("mark_recommendation_not_interested",
		mcp.WithDescription("Mark a recommended place as not interested"),
		mcp.WithNumber("trip_plan_id", mcp.Required(), mcp.Description("Trip plan ID")),
		mcp.WithString("maps_place_id", mcp.Required(), mcp.Description("Maps (Google) place ID"))),
		handleMarkRecommendationNotInterested)
	s.AddTool(mcp.NewTool("get_trip_plan_assistant_text",
		mcp.WithDescription("Send a message to the trip-plan assistant and read the streamed response synchronously"),
		mcp.WithString("message", mcp.Required(), mcp.Description("User message")),
		mcp.WithString("chat_id", mcp.Description("Chat ID (omit to start a new chat)")),
		mcp.WithNumber("trip_plan_id", mcp.Description("Trip plan ID")),
		mcp.WithString("trip_plan_key", mcp.Description("Trip plan key")),
		mcp.WithNumber("geo_id", mcp.Description("Geo ID"))),
		handleGetTripPlanAssistantText)
	s.AddTool(mcp.NewTool("get_trip_plan_assistant_highlights",
		mcp.WithDescription("Extract place highlights from a previous assistant message"),
		mcp.WithString("assistant_message", mcp.Required(), mcp.Description("Assistant message text")),
		mcp.WithNumber("trip_plan_id", mcp.Required(), mcp.Description("Trip plan ID")),
		mcp.WithNumber("selected_geo_id", mcp.Description("Selected geo ID")),
		mcp.WithString("chat_id", mcp.Description("Chat ID")),
		mcp.WithString("assistant_chat_item_id", mcp.Description("Assistant chat item ID"))),
		handleGetTripPlanAssistantHighlights)
	s.AddTool(mcp.NewTool("create_trip_from_flights",
		mcp.WithDescription("Create a new trip plan seeded from a flights payload"),
		mcp.WithString("payload", mcp.Required(), mcp.Description("JSON payload (flights array, ...)"))),
		handleCreateTripFromFlights)
	s.AddTool(mcp.NewTool("rate_email",
		mcp.WithDescription("Score an auto-parsed forwarded email"),
		mcp.WithNumber("email_id", mcp.Description("Email ID")),
		mcp.WithString("rating", mcp.Description("Rating (e.g. thumbs_up)")),
		mcp.WithString("extras", mcp.Description("Optional JSON extras"))),
		handleRateEmail)
}

// === Handlers ===

func handleListGeosWithGoodGuides(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runReadOnlyGeo("ListGeosWithGoodGuides", func(c *wanderlog.Client) (any, error) { return c.ListGeosWithGoodGuides() })
}

func handleListPopularAndNearbyGeos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runReadOnlyGeo("ListPopularAndNearbyGeos", func(c *wanderlog.Client) (any, error) { return c.ListPopularAndNearbyGeos() })
}

func handleFindCountryForIP(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runReadOnlyGeo("FindCountryForIP", func(c *wanderlog.Client) (any, error) { return c.FindCountryForIP() })
}

func handleFindNearestTripadvisorGeo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lat := request.GetFloat("latitude", 0)
	lng := request.GetFloat("longitude", 0)
	return runReadOnlyGeo("FindNearestTripadvisorGeo", func(c *wanderlog.Client) (any, error) {
		return c.FindNearestTripadvisorGeo(lat, lng)
	})
}

func handleFindNearestGeosToIP(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runReadOnlyGeo("FindNearestGeosToIP", func(c *wanderlog.Client) (any, error) {
		return c.FindNearestGeosToIP(nil)
	})
}

func handleFindNearestKayakCity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lat := request.GetFloat("latitude", 0)
	lng := request.GetFloat("longitude", 0)
	city, err := request.RequireString("city_name_to_match")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("city_name_to_match is required"), nil //nolint:nilerr
	}
	return runReadOnlyGeo("FindNearestKayakCity", func(c *wanderlog.Client) (any, error) {
		return c.FindNearestKayakCity(lat, lng, city)
	})
}

func handleGetClientGeos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := request.RequireString("geo_ids")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("geo_ids is required"), nil //nolint:nilerr
	}
	ids, perr := parseIntCSVSafe(raw)
	if perr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid geo_ids: %v", perr)), nil
	}
	return runReadOnlyGeo("GetClientGeos", func(c *wanderlog.Client) (any, error) {
		return c.GetClientGeos(ids)
	})
}

func handleListTripPlannerGeos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runReadOnlyGeo("ListTripPlannerGeos", func(c *wanderlog.Client) (any, error) { return c.ListTripPlannerGeos() })
}

func handleListCountries(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	language := request.GetString("language", "en")
	return runReadOnlyGeo("ListCountries", func(c *wanderlog.Client) (any, error) { return c.ListCountries(language) })
}

func handleListGeoCategoriesForCategory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetInt("keyword_category_id", 0)
	if id <= 0 {
		return mcp.NewToolResultError("keyword_category_id must be a positive integer"), nil
	}
	language := request.GetString("language", "en")
	return runReadOnlyGeo("ListGeoCategoriesForCategory", func(c *wanderlog.Client) (any, error) {
		return c.ListGeoCategoriesForCategory(id, language)
	})
}

func handleListGeoCategoriesForGeo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	geoID := request.GetInt("geo_id", 0)
	if geoID <= 0 {
		return mcp.NewToolResultError("geo_id must be a positive integer"), nil
	}
	source := request.GetString("source", "tripPlanner")
	return runReadOnlyGeo("ListGeoCategoriesForGeo", func(c *wanderlog.Client) (any, error) {
		return c.ListGeoCategoriesForGeo(geoID, source)
	})
}

func handleListGeoInMonthGeos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runReadOnlyGeo("ListGeoInMonthGeos", func(c *wanderlog.Client) (any, error) { return c.ListGeoInMonthGeos() })
}

func handleListKeywordCategories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	language := request.GetString("language", "en")
	return runReadOnlyGeo("ListKeywordCategories", func(c *wanderlog.Client) (any, error) { return c.ListKeywordCategories(language) })
}

func handleSearchGeo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q, err := request.RequireString("q")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("q is required"), nil //nolint:nilerr
	}
	params := map[string]string{"q": q}
	if lang := request.GetString("language", ""); lang != "" {
		params["language"] = lang
	}
	return runReadOnlyGeo("SearchGeo", func(c *wanderlog.Client) (any, error) { return c.SearchGeo(params) })
}

func handleGetAllDistanceInfoForPlace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload, err := parseJSONPayload(request, "payload")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return runReadOnlyGeo("GetAllDistanceInfoForPlace", func(c *wanderlog.Client) (any, error) {
		return c.GetAllDistanceInfoForPlace(payload)
	})
}

func handleGetDistancesForMode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload, err := parseJSONPayload(request, "payload")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return runReadOnlyGeo("GetDistancesForMode", func(c *wanderlog.Client) (any, error) {
		return c.GetDistancesForMode(payload)
	})
}

func handleOptimizeRoute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload, err := parseJSONPayload(request, "payload")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return runReadOnlyGeo("OptimizeRoute", func(c *wanderlog.Client) (any, error) {
		return c.OptimizeRoute(payload)
	})
}

func handleAutocompletePlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	q, err := request.RequireString("query")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("query is required"), nil //nolint:nilerr
	}
	lat := request.GetFloat("latitude", 0)
	lng := request.GetFloat("longitude", 0)
	return runReadOnlyGeo("AutocompletePlaces", func(c *wanderlog.Client) (any, error) {
		return c.AutocompletePlaces(q, lat, lng)
	})
}

func handleFindPlaceFromLngLat(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lat := request.GetFloat("latitude", 0)
	lng := request.GetFloat("longitude", 0)
	return runReadOnlyGeo("FindPlaceFromLngLat", func(c *wanderlog.Client) (any, error) {
		return c.FindPlaceFromLngLat(lat, lng)
	})
}

func handleGetMapLayerGroups(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload, err := parseJSONPayload(request, "payload")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return runReadOnlyGeo("GetMapLayerGroups", func(c *wanderlog.Client) (any, error) {
		return c.GetMapLayerGroups(payload)
	})
}

func handleGetMultiplePlaceDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := request.RequireString("place_ids")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_ids is required"), nil //nolint:nilerr
	}
	ids := splitCSV(raw)
	if len(ids) == 0 {
		return mcp.NewToolResultError("at least one place id is required"), nil
	}
	language := request.GetString("language", "en")
	return runReadOnlyGeo("GetMultiplePlaceDetails", func(c *wanderlog.Client) (any, error) {
		return c.GetMultiplePlaceDetails(ids, language)
	})
}

func handleGetPlaceDetailsV2(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	placeID, err := request.RequireString("place_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_id is required"), nil //nolint:nilerr
	}
	language := request.GetString("language", "en")
	return runReadOnlyGeo("GetPlaceDetailsV2", func(c *wanderlog.Client) (any, error) {
		return c.GetPlaceDetailsV2(placeID, language)
	})
}

func handleSearchPlacesGoogle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload, err := parseJSONPayload(request, "payload")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return runReadOnlyGeo("SearchPlacesGoogle", func(c *wanderlog.Client) (any, error) {
		return c.SearchPlacesGoogle(payload)
	})
}

func handleGetPlacesMetadata(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := request.RequireString("place_ids")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_ids is required"), nil //nolint:nilerr
	}
	ids := splitCSV(raw)
	opts := map[string]string{}
	if v := request.GetInt("geo_id", 0); v > 0 {
		opts["geoId"] = strconv.Itoa(v)
	}
	if v := request.GetString("list_id", ""); v != "" {
		opts["listId"] = v
	}
	if v := request.GetString("list_type", ""); v != "" {
		opts["listType"] = v
	}
	if request.GetBool("get_details", false) {
		opts["getDetails"] = "true"
	}
	return runReadOnlyGeo("GetPlacesMetadata", func(c *wanderlog.Client) (any, error) {
		return c.GetPlacesMetadata(ids, opts)
	})
}

func handleGetPlaceCards(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := request.RequireString("place_ids")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_ids is required"), nil //nolint:nilerr
	}
	ids := splitCSV(raw)
	return runReadOnlyGeo("GetPlaceCards", func(c *wanderlog.Client) (any, error) {
		return c.GetPlaceCards(ids)
	})
}

func handleListPlacePageGeos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return runReadOnlyGeo("ListPlacePageGeos", func(c *wanderlog.Client) (any, error) { return c.ListPlacePageGeos() })
}

func handleGetTripPlanAssistantHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	params := map[string]string{}
	if v := request.GetString("chat_id", ""); v != "" {
		params["chatId"] = v
	}
	if v := request.GetString("page_size", ""); v != "" {
		params["pageSize"] = v
	}
	if v := request.GetString("sent_at_before", ""); v != "" {
		params["sentAtBefore"] = v
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetTripPlanAssistantHistory(params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleListTripPlanAssistantChats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripPlanID := request.GetInt("trip_plan_id", 0)
	if tripPlanID <= 0 {
		return mcp.NewToolResultError("trip_plan_id must be a positive integer"), nil
	}
	search := request.GetString("search", "")
	lastItemIsBefore := int64(request.GetInt("last_item_is_before_millis", 0))
	pageSize := request.GetInt("page_size", 0)
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.ListTripPlanAssistantChats(tripPlanID, search, lastItemIsBefore, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetTripPlanAssistantInitialChat(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripPlanID := request.GetInt("trip_plan_id", 0)
	if tripPlanID <= 0 {
		return mcp.NewToolResultError("trip_plan_id must be a positive integer"), nil
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetTripPlanAssistantInitialChat(tripPlanID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetTripLikesBulk(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := request.RequireString("trip_keys")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_keys is required"), nil //nolint:nilerr
	}
	keys := splitCSV(raw)
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetTripLikesBulk(keys)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetMyProfileData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetMyProfileData()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetLodgingCheckoutData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw, err := request.RequireString("payload")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("payload is required"), nil //nolint:nilerr
	}
	var params map[string]string
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid payload JSON: %v", err)), nil
	}
	return runReadOnlyGeo("GetLodgingCheckoutData", func(c *wanderlog.Client) (any, error) {
		return c.GetLodgingCheckoutData(params)
	})
}

func handleGetDealsForUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetDealsForUser()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetRecommendedPlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripPlanID := request.GetInt("trip_plan_id", 0)
	if tripPlanID <= 0 {
		return mcp.NewToolResultError("trip_plan_id must be a positive integer"), nil
	}
	geoID := request.GetInt("geo_id", 0)
	if geoID <= 0 {
		return mcp.NewToolResultError("geo_id must be a positive integer"), nil
	}
	req := wanderlog.RecommendedPlacesRequest{
		TripPlanID: tripPlanID,
		GeoID:      geoID,
		Input:      request.GetString("input", ""),
	}
	if v := request.GetString("excluding_place_ids", ""); v != "" {
		req.ExcludingPlaceIDs = splitCSV(v)
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetRecommendedPlaces(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleMarkRecommendationNotInterested(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripPlanID := request.GetInt("trip_plan_id", 0)
	if tripPlanID <= 0 {
		return mcp.NewToolResultError("trip_plan_id must be a positive integer"), nil
	}
	mapsPlaceID, err := request.RequireString("maps_place_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("maps_place_id is required"), nil //nolint:nilerr
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	if err := client.MarkRecommendationNotInterested(wanderlog.MarkRecommendationNotInterestedRequest{
		TripPlanID:  tripPlanID,
		MapsPlaceID: mapsPlaceID,
	}); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText("Marked place as not interested"), nil
}

func handleGetTripPlanAssistantText(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, err := request.RequireString("message")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("message is required"), nil //nolint:nilerr
	}
	req := wanderlog.AssistantTextRequest{
		Message:     message,
		ChatID:      request.GetString("chat_id", ""),
		TripPlanID:  request.GetInt("trip_plan_id", 0),
		TripPlanKey: request.GetString("trip_plan_key", ""),
		GeoID:       request.GetInt("geo_id", 0),
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetTripPlanAssistantText(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetTripPlanAssistantHighlights(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	assistantMessage, err := request.RequireString("assistant_message")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("assistant_message is required"), nil //nolint:nilerr
	}
	tripPlanID := request.GetInt("trip_plan_id", 0)
	if tripPlanID <= 0 {
		return mcp.NewToolResultError("trip_plan_id must be a positive integer"), nil
	}
	req := wanderlog.AssistantHighlightsRequest{
		AssistantMessage:    assistantMessage,
		TripPlanID:          tripPlanID,
		SelectedGeoID:       request.GetInt("selected_geo_id", 0),
		ChatID:              request.GetString("chat_id", ""),
		AssistantChatItemID: request.GetString("assistant_chat_item_id", ""),
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetTripPlanAssistantHighlights(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleCreateTripFromFlights(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload, err := parseJSONPayload(request, "payload")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.CreateTripFromFlights(payload)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleRateEmail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	req := wanderlog.RateEmailRequest{
		EmailID: request.GetInt("email_id", 0),
		Rating:  request.GetString("rating", ""),
	}
	if extras := request.GetString("extras", ""); extras != "" {
		req.Extras = json.RawMessage(extras)
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	if err := client.RateEmail(req); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText("Email rated"), nil
}

// runReadOnlyGeo executes a read-only call against an optional-auth client and
// returns the structured result. The opName argument is used in the error
// message if the call fails.
func runReadOnlyGeo(opName string, fn func(*wanderlog.Client) (any, error)) (*mcp.CallToolResult, error) {
	client := optionalAuthClient()
	resp, err := fn(client)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("%s failed: %v", opName, err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func parseJSONPayload(request mcp.CallToolRequest, name string) (any, error) {
	raw, err := request.RequireString(name)
	if err != nil {
		_ = err
		return nil, fmt.Errorf("%s is required", name)
	}
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("invalid %s JSON: %w", name, err)
	}
	return payload, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseIntCSVSafe(raw string) ([]int, error) {
	parts := splitCSV(raw)
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("not an integer: %q", p)
		}
		out = append(out, n)
	}
	return out, nil
}
