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

// registerExtendedTools wires up the user/feed/journal/config MCP tools.
// Read-only tools are always registered; write-gated ones only when
// readOnly is false.
func registerExtendedTools(s *server.MCPServer, readOnly bool) {
	// user tools (read-only)
	s.AddTool(
		mcp.NewTool("get_me",
			mcp.WithDescription("Get the authenticated user's profile"),
		),
		handleGetMe,
	)
	s.AddTool(
		mcp.NewTool("get_user_profile",
			mcp.WithDescription("Get another user's profile and public trips by numeric ID or @username"),
			mcp.WithString("target", mcp.Required(),
				mcp.Description("Numeric user ID or @username")),
		),
		handleGetUserProfile,
	)
	s.AddTool(
		mcp.NewTool("get_notifications",
			mcp.WithDescription("Get the authenticated user's notification inbox"),
			mcp.WithNumber("offset", mcp.Description("Pagination offset (default 0)")),
		),
		handleGetNotifications,
	)
	s.AddTool(
		mcp.NewTool("get_notification_settings",
			mcp.WithDescription("Get the authenticated user's notification settings"),
		),
		handleGetNotificationSettings,
	)
	s.AddTool(
		mcp.NewTool("get_user_emails",
			mcp.WithDescription("List the authenticated user's registered email addresses"),
		),
		handleGetUserEmails,
	)
	s.AddTool(
		mcp.NewTool("autocomplete_users",
			mcp.WithDescription("Autocomplete Wanderlog users by name prefix"),
			mcp.WithString("query", mcp.Required(),
				mcp.Description("Name prefix to search")),
		),
		handleAutocompleteUsers,
	)
	s.AddTool(
		mcp.NewTool("is_username_taken",
			mcp.WithDescription("Check whether a username is already in use"),
			mcp.WithString("username", mcp.Required(),
				mcp.Description("Username to check")),
		),
		handleIsUsernameTaken,
	)

	// feed / discovery (read-only)
	s.AddTool(
		mcp.NewTool("get_feed_home",
			mcp.WithDescription("Fetch the authenticated user's home feed (own trips, friends' trips, recommended guides)"),
		),
		handleGetFeedHome,
	)
	s.AddTool(
		mcp.NewTool("get_feed_recent",
			mcp.WithDescription("Get the authenticated user's most recently edited trip"),
		),
		handleGetFeedRecent,
	)
	s.AddTool(
		mcp.NewTool("get_feed_friends",
			mcp.WithDescription("Get trip plans from friends"),
		),
		handleGetFeedFriends,
	)
	s.AddTool(
		mcp.NewTool("get_trip_history",
			mcp.WithDescription("Get paginated trip edit history"),
			mcp.WithNumber("offset", mcp.Description("Pagination offset (default 0)")),
		),
		handleGetTripHistory,
	)
	s.AddTool(
		mcp.NewTool("browse_guides",
			mcp.WithDescription("Browse curated Wanderlog guides, optionally scoped to a geo ID"),
			mcp.WithNumber("geo_id", mcp.Description("Limit to this geo ID (optional)")),
		),
		handleBrowseGuides,
	)
	s.AddTool(
		mcp.NewTool("search_geos",
			mcp.WithDescription("Search Wanderlog destination geo IDs for create_trip. Use this before creating trips instead of guessing geo_id values."),
			mcp.WithString("query", mcp.Required(),
				mcp.Description("Destination name, e.g. Japan, Tokyo, Kyoto")),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of geos to return"),
				mcp.DefaultNumber(10)),
		),
		handleSearchGeos,
	)

	// journal / advanced trip ops (read-only)
	s.AddTool(
		mcp.NewTool("get_view_only_journal",
			mcp.WithDescription("Fetch a published view-only journal by its journal key"),
			mcp.WithString("journal_key", mcp.Required(),
				mcp.Description("Journal share key")),
		),
		handleGetViewOnlyJournal,
	)
	s.AddTool(
		mcp.NewTool("get_trip_expenses_csv",
			mcp.WithDescription("Download a trip's expenses as CSV (requires authentication)"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
		),
		handleGetTripExpensesCSV,
	)
	s.AddTool(
		mcp.NewTool("get_trip_distinction",
			mcp.WithDescription("Get a trip's distinction/badge (if any)"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
		),
		handleGetTripDistinction,
	)
	s.AddTool(
		mcp.NewTool("get_trip_images",
			mcp.WithDescription("Get images attached to a trip"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
		),
		handleGetTripImages,
	)

	// travel helper APIs (read-only)
	s.AddTool(
		mcp.NewTool("get_all_airlines",
			mcp.WithDescription("List supported airlines with IATA/ICAO codes"),
		),
		handleGetAllAirlines,
	)
	s.AddTool(
		mcp.NewTool("autocomplete_airports",
			mcp.WithDescription("Search airports by name, city, or IATA code"),
			mcp.WithString("query", mcp.Required(),
				mcp.Description("Airport search query")),
			mcp.WithNumber("latitude", mcp.Description("Optional location bias latitude")),
			mcp.WithNumber("longitude", mcp.Description("Optional location bias longitude")),
		),
		handleAutocompleteAirports,
	)
	s.AddTool(
		mcp.NewTool("get_hotel_rates",
			mcp.WithDescription("Get Google hotel price rates for a lodging property"),
			mcp.WithString("property_id", mcp.Required(),
				mcp.Description("Lodging property ID")),
		),
		handleGetHotelRates,
	)

	// config (read-only)
	s.AddTool(
		mcp.NewTool("get_global_config",
			mcp.WithDescription("Fetch the server's global client configuration"),
		),
		handleGetGlobalConfig,
	)

	// Write-gated tools
	if readOnly {
		return
	}
	s.AddTool(
		mcp.NewTool("mark_notifications_read",
			mcp.WithDescription("Mark one or more notifications as read"),
			mcp.WithString("notification_ids", mcp.Required(),
				mcp.Description("Comma-separated notification IDs")),
		),
		handleMarkNotificationsRead,
	)
	s.AddTool(
		mcp.NewTool("set_user_kv",
			mcp.WithDescription("Write a value to the authenticated user's key-value store"),
			mcp.WithString("key", mcp.Required(),
				mcp.Description("Key name")),
			mcp.WithString("value", mcp.Required(),
				mcp.Description("Value (JSON or plain string)")),
		),
		handleSetUserKV,
	)
	s.AddTool(
		mcp.NewTool("register_trip_view",
			mcp.WithDescription("Register a view on a shared trip"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
		),
		handleRegisterTripView,
	)
	s.AddTool(
		mcp.NewTool("create_guide_from_trip",
			mcp.WithDescription("Promote a trip plan into a published guide"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
		),
		handleCreateGuideFromTrip,
	)
	s.AddTool(
		mcp.NewTool("export_trip",
			mcp.WithDescription("Export a trip to Google Maps"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
		),
		handleExportTrip,
	)
	s.AddTool(
		mcp.NewTool("autofill_day",
			mcp.WithDescription("Generate place suggestions for a trip day section"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
			mcp.WithNumber("section_id", mcp.Required(),
				mcp.Description("Section ID")),
			mcp.WithString("query",
				mcp.Description("Optional suggestion query")),
		),
		handleAutofillDay,
	)
	s.AddTool(
		mcp.NewTool("add_checklist_items",
			mcp.WithDescription("Add checklist items to a trip checklist section"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
			mcp.WithNumber("section_id", mcp.Required(),
				mcp.Description("Checklist section ID")),
			mcp.WithString("items", mcp.Required(),
				mcp.Description("Comma-separated checklist item text")),
		),
		handleAddChecklistItems,
	)
	s.AddTool(
		mcp.NewTool("toggle_checklist_item",
			mcp.WithDescription("Set a checklist item's checked state"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
			mcp.WithNumber("section_id", mcp.Required(),
				mcp.Description("Checklist section ID")),
			mcp.WithNumber("item_id", mcp.Required(),
				mcp.Description("Checklist item ID")),
			mcp.WithBoolean("checked", mcp.Required(),
				mcp.Description("Whether the item is checked")),
		),
		handleToggleChecklistItem,
	)
	s.AddTool(
		mcp.NewTool("add_collaborator",
			mcp.WithDescription("Add a Wanderlog user as a trip collaborator"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
			mcp.WithNumber("user_id", mcp.Required(),
				mcp.Description("Wanderlog user ID")),
		),
		handleAddCollaborator,
	)
	s.AddTool(
		mcp.NewTool("remove_collaborator",
			mcp.WithDescription("Remove a Wanderlog user from trip collaborators"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
			mcp.WithNumber("user_id", mcp.Required(),
				mcp.Description("Wanderlog user ID")),
		),
		handleRemoveCollaborator,
	)
	s.AddTool(
		mcp.NewTool("get_or_create_share_key",
			mcp.WithDescription("Get or create a trip share key for the requested permissions"),
			mcp.WithString("edit_key", mcp.Required(),
				mcp.Description("Trip edit key")),
			mcp.WithBoolean("can_view",
				mcp.Description("Share key can view the trip"),
				mcp.DefaultBool(true)),
			mcp.WithBoolean("can_edit",
				mcp.Description("Share key can edit the trip"),
				mcp.DefaultBool(false)),
		),
		handleGetOrCreateShareKey,
	)
}

func ensuredAuthClient() (*wanderlog.Client, error) {
	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return nil, err
	}
	return client, nil
}

func optionalAuthClient() *wanderlog.Client {
	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if auth, err := wanderlog.LoadCredentials(); err == nil {
		client.SetAuth(auth)
	}
	return client
}

func handleGetMe(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	profile, err := client.GetMe()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get profile: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(profile), nil
}

func handleGetUserProfile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, err := request.RequireString("target")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("target is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	if strings.HasPrefix(target, "@") {
		resp, err := client.GetUserProfileByUsername(strings.TrimPrefix(target, "@"))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}
		return mcp.NewToolResultStructuredOnly(resp), nil
	}
	id, err := strconv.Atoi(target)
	if err != nil {
		_ = err
		return mcp.NewToolResultError("target must be a numeric user ID or @username"), nil //nolint:nilerr
	}
	resp, err := client.GetUserProfile(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetNotifications(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	offset := request.GetInt("offset", 0)
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetNotifications(offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetNotificationSettings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetNotificationSettings()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetUserEmails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetUserEmails()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleAutocompleteUsers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("query is required"), nil //nolint:nilerr
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.AutocompleteUsers(query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleIsUsernameTaken(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, err := request.RequireString("username")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("username is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	taken, err := client.IsUsernameTaken(username)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(map[string]interface{}{"username": username, "taken": taken}), nil
}

func handleGetFeedHome(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetFeedHome()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetFeedRecent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetFeedMostRecent()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetFeedFriends(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetFriendsPlans()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetTripHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	offset := request.GetInt("offset", 0)
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetTripHistory(offset)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleBrowseGuides(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	geoID := request.GetInt("geo_id", 0)
	client := optionalAuthClient()
	resp, err := client.BrowseGuides(geoID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

type geoGuideCount struct {
	Name  string `json:"name"`
	GeoID int    `json:"geoId"`
}

func filterGeoGuideCounts(result *wanderlog.GeoSearchResult, query string, limit int) []geoGuideCount {
	query = strings.ToLower(strings.TrimSpace(query))
	if limit <= 0 {
		limit = 10
	}

	matches := make([]geoGuideCount, 0, limit)
	addIfMatch := func(name string, id int) {
		if query == "" || strings.Contains(strings.ToLower(name), query) {
			matches = append(matches, geoGuideCount{Name: name, GeoID: id})
		}
	}

	for _, c := range result.Countries {
		addIfMatch(c.Name, c.ID)
	}
	for _, c := range result.Cities {
		addIfMatch(c.Name, c.ID)
	}

	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

func handleSearchGeos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("query is required"), nil //nolint:nilerr
	}

	client := optionalAuthClient()
	resp, err := client.SearchGeos()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}

	matches := filterGeoGuideCounts(resp, query, request.GetInt("limit", 10))

	return mcp.NewToolResultStructuredOnly(map[string]any{
		"success": true,
		"data": map[string]any{
			"geos": matches,
		},
	}), nil
}

func handleGetViewOnlyJournal(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	journalKey, err := request.RequireString("journal_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("journal_key is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	resp, err := client.GetViewOnlyJournal(journalKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetTripExpensesCSV(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	csv, err := client.GetTripExpensesCSV(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText(string(csv)), nil
}

func handleGetTripDistinction(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	resp, err := client.GetTripDistinction(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetTripImages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	resp, err := client.GetTripImages(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetAllAirlines(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := optionalAuthClient()
	resp, err := client.GetAllAirlines()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleAutocompleteAirports(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("query is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	if lat, ok := request.GetArguments()["latitude"].(float64); ok {
		if lng, ok := request.GetArguments()["longitude"].(float64); ok {
			resp, err := client.AutocompleteAirportWithLocation(query, lat, lng)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
			}
			return mcp.NewToolResultStructuredOnly(resp), nil
		}
	}
	resp, err := client.AutocompleteAirport(query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetHotelRates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	propertyID, err := request.RequireString("property_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("property_id is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	resp, err := client.GetGooglePriceRates(propertyID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleGetGlobalConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := optionalAuthClient()
	cfg, err := client.GetGlobalConfig()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(cfg), nil
}

func handleMarkNotificationsRead(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	idsStr, err := request.RequireString("notification_ids")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("notification_ids is required"), nil //nolint:nilerr
	}
	ids := []string{}
	for _, id := range strings.Split(idsStr, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			ids = append(ids, trimmed)
		}
	}
	if len(ids) == 0 {
		return mcp.NewToolResultError("notification_ids must contain at least one ID"), nil
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	if err := client.MarkNotificationsRead(ids); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Marked %d notification(s) as read", len(ids))), nil
}

func handleSetUserKV(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key, err := request.RequireString("key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("key is required"), nil //nolint:nilerr
	}
	value, err := request.RequireString("value")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("value is required"), nil //nolint:nilerr
	}
	var parsed json.RawMessage
	if jerr := json.Unmarshal([]byte(value), &parsed); jerr != nil {
		parsed = json.RawMessage(fmt.Sprintf("%q", value))
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	if err := client.SetKeyValue(key, parsed); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Wrote key %q", key)), nil
}

func handleRegisterTripView(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	client := optionalAuthClient()
	if err := client.RegisterTripView(tripKey); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Registered view on %s", tripKey)), nil
}

func handleCreateGuideFromTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.CreateGuideFromTripPlan(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleExportTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.ExportTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleAutofillDay(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	sectionID := request.GetInt("section_id", 0)
	if sectionID <= 0 {
		return mcp.NewToolResultError("section_id must be a positive integer"), nil
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.AutofillDay(tripKey, sectionID, request.GetString("query", ""))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func parseCommaList(value string) []string {
	items := []string{}
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func handleAddChecklistItems(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	sectionID := request.GetInt("section_id", 0)
	if sectionID <= 0 {
		return mcp.NewToolResultError("section_id must be a positive integer"), nil
	}
	itemsArg, err := request.RequireString("items")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("items is required"), nil //nolint:nilerr
	}
	texts := parseCommaList(itemsArg)
	if len(texts) == 0 {
		return mcp.NewToolResultError("items must contain at least one item"), nil
	}
	items := make([]wanderlog.ChecklistItem, 0, len(texts))
	for _, text := range texts {
		items = append(items, wanderlog.ChecklistItem{Text: text})
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.AddChecklistItems(tripKey, sectionID, items)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleToggleChecklistItem(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	sectionID := request.GetInt("section_id", 0)
	if sectionID <= 0 {
		return mcp.NewToolResultError("section_id must be a positive integer"), nil
	}
	itemID := request.GetInt("item_id", 0)
	if itemID <= 0 {
		return mcp.NewToolResultError("item_id must be a positive integer"), nil
	}
	checked, err := request.RequireBool("checked")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("checked is required"), nil //nolint:nilerr
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.ToggleChecklistItem(tripKey, sectionID, itemID, checked)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}

func handleAddCollaborator(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	userID := request.GetInt("user_id", 0)
	if userID <= 0 {
		return mcp.NewToolResultError("user_id must be a positive integer"), nil
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	if err := client.AddCollaborator(tripKey, userID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Added collaborator %d", userID)), nil
}

func handleRemoveCollaborator(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := request.RequireString("trip_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("trip_key is required"), nil //nolint:nilerr
	}
	userID := request.GetInt("user_id", 0)
	if userID <= 0 {
		return mcp.NewToolResultError("user_id must be a positive integer"), nil
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	if err := client.RemoveCollaborator(tripKey, userID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Removed collaborator %d", userID)), nil
}

func handleGetOrCreateShareKey(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	editKey, err := request.RequireString("edit_key")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("edit_key is required"), nil //nolint:nilerr
	}
	permissions := wanderlog.ShareKeyPermissions{
		CanView: request.GetBool("can_view", true),
		CanEdit: request.GetBool("can_edit", false),
	}
	if !permissions.CanView && !permissions.CanEdit {
		return mcp.NewToolResultError("at least one permission must be enabled"), nil
	}
	client, err := ensuredAuthClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	resp, err := client.GetOrCreateShareKey(editKey, permissions)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
	}
	return mcp.NewToolResultStructuredOnly(resp), nil
}
