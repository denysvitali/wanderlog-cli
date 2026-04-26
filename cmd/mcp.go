package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

// tripIDKey is a custom context key for storing the default trip ID.
type tripIDKey struct{}

// withTripID adds a trip ID to the context.
func withTripID(ctx context.Context, tripID string) context.Context {
	return context.WithValue(ctx, tripIDKey{}, tripID)
}

// tripIDFromContext extracts the default trip ID from the context.
func tripIDFromContext(ctx context.Context) (string, bool) {
	tripID, ok := ctx.Value(tripIDKey{}).(string)
	return tripID, ok && tripID != ""
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start Wanderlog CLI as an MCP server",
	Long: `Start the Wanderlog CLI as a Model Context Protocol (MCP) server.
This allows LLMs and other MCP clients to access Wanderlog trip data
and functionality through a standardized protocol.

The server runs in read-only mode by default for security. Use --enable-write
to allow trip modifications.

Read-only tools (always available):
- Listing trips and getting trip details
- Getting trip sections (efficient section queries)
- Viewing places and itineraries
- Searching for places
- Trip analysis and recommendations

Write operations (only with --enable-write):
- Creating, updating, and deleting trips
- Updating trip metadata (title, dates, privacy)
- Adding places to trips
- Removing and moving places
- Reordering places within sections

Examples:
  wanderlog mcp                             # Start read-only MCP server on stdio
  wanderlog mcp --enable-write              # Start read-write MCP server on stdio
  wanderlog mcp --trip-id "abc123"          # Start with default trip ID (trip_id params become optional)
  wanderlog mcp --trip-id "abc123" --enable-write  # Start with default trip ID and write operations
  wanderlog mcp --http :8080                # Start read-only HTTP MCP server on port 8080`,
	Run: func(cmd *cobra.Command, args []string) {
		enableWrite, _ := cmd.Flags().GetBool("enable-write")
		tripID, _ := cmd.Flags().GetString("trip-id")
		if httpAddr, _ := cmd.Flags().GetString("http"); httpAddr != "" {
			runMCPHTTPServer(httpAddr, enableWrite, tripID)
		} else {
			runMCPStdioServer(enableWrite, tripID)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().String("http", "", "HTTP address to serve MCP on (e.g., :8080)")
	mcpCmd.Flags().Bool("enable-write", false, "Enable write operations (add/remove places, etc.)")
	mcpCmd.Flags().String("trip-id", "", "Default trip key to use for all operations (makes trip_id/trip_key parameters optional in tools)")
}

func createMCPServer(readOnly bool) *server.MCPServer {
	// Create MCP server with capabilities
	s := server.NewMCPServer(
		"Wanderlog CLI",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
	)

	// Add list trips tool
	listTripsTools := mcp.NewTool("list_trips",
		mcp.WithDescription("List all trips for the authenticated user"),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(listTripsTools, handleListTrips)

	// Add get trip details tool
	getTripTool := mcp.NewTool("get_trip",
		mcp.WithDescription("Get detailed information about a specific trip"),
		mcp.WithString("trip_id",
			mcp.Description("The trip key to retrieve, not the numeric database ID (optional if default trip key is set)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(getTripTool, handleGetTrip)

	// Add list places tool
	listPlacesTool := mcp.NewTool("list_places",
		mcp.WithDescription("List all places for a specific trip"),
		mcp.WithString("trip_id",
			mcp.Description("The trip key to get places for, not the numeric database ID (optional if default trip key is set)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(listPlacesTool, handleListPlaces)

	// Add list sections tool
	listSectionsTool := mcp.NewTool("list_sections",
		mcp.WithDescription("List all sections/days for a specific trip with their IDs and dates"),
		mcp.WithString("trip_id",
			mcp.Description("The trip key to get sections for, not the numeric database ID (optional if default trip key is set)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(listSectionsTool, handleListSections)

	// Get sections details tool
	getTripSectionsTool := mcp.NewTool("get_trip_sections",
		mcp.WithDescription("Get detailed section information for a trip (more efficient than fetching full trip)"),
		mcp.WithString("trip_key",
			mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
		),
	)
	s.AddTool(getTripSectionsTool, handleGetTripSections)

	// Add write operation tools only if not in read-only mode
	if !readOnly {
		// Add place to trip tool
		addPlaceTool := mcp.NewTool("add_place",
			mcp.WithDescription("Add a place to a trip. Use section_date or a positive section_id from list_sections to place it on the itinerary. Pass unscheduled=true only for the general Places to visit list. If place_id is provided without coordinates, they will be automatically fetched from the Wanderlog API to prevent corrupt place data."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip to add the place to (optional if default trip ID is set)"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the place to add"),
			),
			mcp.WithString("place_id",
				mcp.Description("Google Place ID (recommended - coordinates will be auto-fetched if not provided)"),
			),
			mcp.WithNumber("latitude",
				mcp.Description("Latitude of the place (optional - will be auto-fetched if place_id is provided)"),
			),
			mcp.WithNumber("longitude",
				mcp.Description("Longitude of the place (optional - will be auto-fetched if place_id is provided)"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Positive itinerary section ID from list_sections. Do not pass 0 unless unscheduled=true."),
			),
			mcp.WithString("section_date",
				mcp.Description("Itinerary date (YYYY-MM-DD). The tool resolves it to that day's section ID."),
			),
			mcp.WithBoolean("unscheduled",
				mcp.Description("Set true to add to the general Places to visit list instead of a dated itinerary section."),
				mcp.DefaultBool(false),
			),
			mcp.WithString("text",
				mcp.Description("Additional text/notes for the place (optional)"),
			),
		)
		s.AddTool(addPlaceTool, handleAddPlace)

		addFlightTool := mcp.NewTool("add_flight",
			mcp.WithDescription("Add a flight reservation block to a dated itinerary section. Requires flight_number and departure_date. Airport info is auto-resolved from flight number. Use section_date or a positive section_id from list_sections; section_date defaults to departure_date."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key to add the flight to (optional if default trip key is set)"),
			),
			mcp.WithString("section_date",
				mcp.Description("Itinerary date (YYYY-MM-DD). Defaults to departure_date when omitted."),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Positive itinerary section ID from list_sections."),
			),
			mcp.WithString("flight_number",
				mcp.Required(),
				mcp.Description("Flight number including airline code, e.g. MU244 or UA875"),
			),
			mcp.WithString("departure_date",
				mcp.Required(),
				mcp.Description("Departure date (YYYY-MM-DD)"),
			),
			mcp.WithString("departure_time",
				mcp.Description("Departure time"),
			),
			mcp.WithString("arrival_date",
				mcp.Description("Arrival date (YYYY-MM-DD)"),
			),
			mcp.WithString("arrival_time",
				mcp.Description("Arrival time"),
			),
			mcp.WithString("confirmation_number",
				mcp.Description("Confirmation number"),
			),
			mcp.WithString("notes",
				mcp.Description("Additional notes"),
			),
		)
		s.AddTool(addFlightTool, handleAddFlight)

		// Remove place from trip tool
		removePlaceTool := mcp.NewTool("remove_place",
			mcp.WithDescription("Remove a place from a trip"),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip to remove the place from (optional if default trip ID is set)"),
			),
			mcp.WithNumber("place_id",
				mcp.Required(),
				mcp.Description("The ID of the place to remove"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID to remove the place from (optional)"),
			),
		)
		s.AddTool(removePlaceTool, handleRemovePlace)

		// Move place between sections tool
		movePlaceTool := mcp.NewTool("move_place",
			mcp.WithDescription("Move a place from one section to another, or reorder within the same section. Preserves all place metadata and notes."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("place_id",
				mcp.Required(),
				mcp.Description("The ID of the place to move"),
			),
			mcp.WithNumber("from_section_id",
				mcp.Required(),
				mcp.Description("Source section ID (use list_sections to get section IDs)"),
			),
			mcp.WithNumber("to_section_id",
				mcp.Required(),
				mcp.Description("Destination section ID (can be same as source to reorder)"),
			),
			mcp.WithNumber("position",
				mcp.Required(),
				mcp.Description("Target position index in the destination section (0-based)"),
			),
		)
		s.AddTool(movePlaceTool, handleMovePlace)

		// Reorder places within section tool
		reorderPlacesTool := mcp.NewTool("reorder_places",
			mcp.WithDescription("Reorder places within a section by providing the desired order of place IDs"),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("section_id",
				mcp.Required(),
				mcp.Description("Section ID to reorder places in (use list_sections to get section IDs)"),
			),
			mcp.WithString("place_ids",
				mcp.Required(),
				mcp.Description("Comma-separated list of place IDs in the desired order (e.g., '123,456,789')"),
			),
		)
		s.AddTool(reorderPlacesTool, handleReorderPlaces)
	}

	// Add search places tool
	searchPlacesTool := mcp.NewTool("search_places",
		mcp.WithDescription("Search for places using Google Places API (requires GOOGLE_PLACES_API_KEY environment variable)"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query for places"),
		),
		mcp.WithNumber("latitude",
			mcp.Description("Latitude for location-based search (optional)"),
		),
		mcp.WithNumber("longitude",
			mcp.Description("Longitude for location-based search (optional)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json, markdown)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json", "markdown"),
		),
	)
	s.AddTool(searchPlacesTool, handleSearchPlaces)

	// Add place details tool using Wanderlog's API
	placeDetailsTool := mcp.NewTool("get_place_details",
		mcp.WithDescription("Get detailed information about a place using Wanderlog's place details API"),
		mcp.WithString("place_id",
			mcp.Required(),
			mcp.Description("Google Place ID to get details for"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(placeDetailsTool, handleGetPlaceDetails)

	// Add Wanderlog place search tool
	wanderlogSearchTool := mcp.NewTool("search_places_wanderlog",
		mcp.WithDescription("Search for places using Wanderlog's native autocomplete API"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query for places"),
		),
		mcp.WithNumber("latitude",
			mcp.Description("Latitude for location-based search (optional)"),
		),
		mcp.WithNumber("longitude",
			mcp.Description("Longitude for location-based search (optional)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(wanderlogSearchTool, handleSearchPlacesWanderlog)

	getFlightStopsTool := mcp.NewTool("get_flight_stops",
		mcp.WithDescription("Get airport stops for a known flight number. Note: flight_number is the numeric portion only (e.g., 244), not the full flight code like MU244. Use the separate airline parameter for the airline code."),
		mcp.WithString("flight_number",
			mcp.Required(),
			mcp.Description("Flight number (numeric portion only, e.g., 244 for flight MU244)"),
		),
		mcp.WithString("airline",
			mcp.Required(),
			mcp.Description("Airline IATA code, e.g. UA"),
		),
		mcp.WithString("date",
			mcp.Required(),
			mcp.Description("Departure date (YYYY-MM-DD)"),
		),
	)
	s.AddTool(getFlightStopsTool, handleGetFlightStops)

	// Add search hotels tool
	searchHotelsTool := mcp.NewTool("search_hotels",
		mcp.WithDescription("Search for hotels/lodging"),
		mcp.WithString("location",
			mcp.Required(),
			mcp.Description("Location to search (e.g., 'Paris', 'Times Square NYC')"),
		),
		mcp.WithString("check_in",
			mcp.Required(),
			mcp.Description("Check-in date (YYYY-MM-DD)"),
		),
		mcp.WithString("check_out",
			mcp.Required(),
			mcp.Description("Check-out date (YYYY-MM-DD)"),
		),
		mcp.WithNumber("guests",
			mcp.Description("Number of guests"),
			mcp.DefaultNumber(1),
		),
	)
	s.AddTool(searchHotelsTool, handleSearchHotels)

	// Trip management tools
	createTripTool := mcp.NewTool("create_trip",
		mcp.WithDescription("Create a new trip plan"),
		mcp.WithString("title", mcp.Required(),
			mcp.Description("Trip title")),
		mcp.WithNumber("geo_id", mcp.Required(),
			mcp.Description("Wanderlog destination geo ID. Use search_geos or browse_guides geoGuideCounts first; do not guess this value or use a Google Place ID.")),
		mcp.WithString("start_date",
			mcp.Required(),
			mcp.Description("Start date in YYYY-MM-DD format")),
		mcp.WithString("end_date",
			mcp.Required(),
			mcp.Description("End date in YYYY-MM-DD format")),
		mcp.WithString("privacy",
			mcp.Description("Privacy setting: public, private, or friends"),
			mcp.DefaultString("private"),
			mcp.Enum("public", "private", "friends")),
	)
	s.AddTool(createTripTool, handleCreateTrip)

	deleteTripTool := mcp.NewTool("delete_trip",
		mcp.WithDescription("Delete a trip plan"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key to delete")),
	)
	s.AddTool(deleteTripTool, handleDeleteTrip)

	deleteTripsTool := mcp.NewTool("delete_trips",
		mcp.WithDescription("Delete multiple trip plans in a single request"),
		mcp.WithString("trip_keys", mcp.Required(),
			mcp.Description("Comma-separated list of trip keys to delete")),
	)
	s.AddTool(deleteTripsTool, handleDeleteTrips)

	restoreTripTool := mcp.NewTool("restore_trip",
		mcp.WithDescription("Restore a soft-deleted trip plan"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key to restore")),
	)
	s.AddTool(restoreTripTool, handleRestoreTrip)

	copyTripTool := mcp.NewTool("copy_trip",
		mcp.WithDescription("Create a copy of an existing trip"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key to copy")),
	)
	s.AddTool(copyTripTool, handleCopyTrip)

	updateTripTool := mcp.NewTool("update_trip",
		mcp.WithDescription("Update trip metadata (title, dates, privacy). At least one field must be provided."),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key to update")),
		mcp.WithString("title",
			mcp.Description("New trip title")),
		mcp.WithString("start_date",
			mcp.Description("New start date (YYYY-MM-DD format)")),
		mcp.WithString("end_date",
			mcp.Description("New end date (YYYY-MM-DD format)")),
		mcp.WithString("privacy",
			mcp.Description("Privacy setting: 'public', 'private', or 'unlisted'")),
	)
	s.AddTool(updateTripTool, handleUpdateTrip)

	// Social features
	likeTripTool := mcp.NewTool("like_trip",
		mcp.WithDescription("Like or unlike a trip plan"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key")),
		mcp.WithBoolean("liked", mcp.Required(),
			mcp.Description("true to like, false to unlike")),
	)
	s.AddTool(likeTripTool, handleLikeTrip)

	getLikeCountTool := mcp.NewTool("get_like_count",
		mcp.WithDescription("Get like count and status for a trip"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key")),
	)
	s.AddTool(getLikeCountTool, handleGetLikeCount)

	// Collaboration tools
	sendInvitesTool := mcp.NewTool("send_trip_invites",
		mcp.WithDescription("Send invites to collaborate on a trip"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key")),
		mcp.WithString("invitees", mcp.Required(),
			mcp.Description("Comma-separated list of email addresses")),
		mcp.WithString("message",
			mcp.Description("Optional message to include with the invite")),
	)
	s.AddTool(sendInvitesTool, handleSendInvites)

	listInvitesTool := mcp.NewTool("list_trip_invites",
		mcp.WithDescription("List all invites sent for a trip"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key")),
	)
	s.AddTool(listInvitesTool, handleListInvites)

	// Add trip resource
	tripResource := mcp.NewResource(
		"wanderlog://trips/{trip_id}",
		"Trip Details",
		mcp.WithResourceDescription("Detailed information about a specific Wanderlog trip"),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(tripResource, handleTripResource)

	// Register extended tools (user, feed, journal, config) from mcp_tools.go.
	registerExtendedTools(s, readOnly)

	// Add trip analysis prompt
	analyzeTripsPrompt := mcp.NewPrompt("analyze_trip",
		mcp.WithPromptDescription("Analyze a trip and provide insights or recommendations"),
		mcp.WithArgument("trip_id",
			mcp.ArgumentDescription("The ID of the trip to analyze"),
		),
		mcp.WithArgument("focus",
			mcp.ArgumentDescription("What to focus on (budget, itinerary, places, overall)"),
		),
	)
	s.AddPrompt(analyzeTripsPrompt, handleAnalyzeTrip)

	return s
}

func runMCPStdioServer(enableWrite bool, tripID string) {
	readOnly := !enableWrite
	s := createMCPServer(readOnly)

	mode := "read-only"
	if enableWrite {
		mode = "read-write"
	}

	logFields := map[string]interface{}{"mode": mode}
	if tripID != "" {
		logFields["default_trip_id"] = tripID
	}
	logger.WithFields(logFields).Info("Starting Wanderlog MCP server on stdio")

	var err error
	if tripID != "" {
		// Use context function to inject trip ID
		err = server.ServeStdio(s, server.WithStdioContextFunc(func(ctx context.Context) context.Context {
			return withTripID(ctx, tripID)
		}))
	} else {
		err = server.ServeStdio(s)
	}

	if err != nil {
		logger.WithError(err).Fatal("Failed to start MCP server")
	}
}

func runMCPHTTPServer(addr string, enableWrite bool, tripID string) {
	readOnly := !enableWrite
	s := createMCPServer(readOnly)

	mode := "read-only"
	if enableWrite {
		mode = "read-write"
	}

	logFields := map[string]interface{}{"address": addr, "mode": mode}
	if tripID != "" {
		logFields["default_trip_id"] = tripID
	}
	logger.WithFields(logFields).Info("Starting Wanderlog MCP server on HTTP")

	var httpServer *server.StreamableHTTPServer
	if tripID != "" {
		// Use context function to inject trip ID from HTTP headers or use default
		httpServer = server.NewStreamableHTTPServer(s, server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			return withTripID(ctx, tripID)
		}))
	} else {
		httpServer = server.NewStreamableHTTPServer(s)
	}

	if err := httpServer.Start(addr); err != nil {
		logger.WithError(err).Fatal("Failed to start HTTP MCP server")
	}
}

// Tool handlers
func handleListTrips(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	trips, err := client.GetUserTrips()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trips: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(trips), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	var result string
	if len(trips.Data) == 0 {
		result = "No trips found."
	} else {
		result = fmt.Sprintf("%d trips:\n", len(trips.Data))
		// Limit to first 20 trips to avoid exceeding char limit
		limit := len(trips.Data)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			trip := trips.Data[i]
			result += fmt.Sprintf("%d. %s (%s) | %d places\n", i+1, trip.Title, trip.Key, trip.PlaceCount)
		}
		if len(trips.Data) > 20 {
			result += fmt.Sprintf("\n... and %d more (use JSON format for full list)\n", len(trips.Data)-20)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func handleGetTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_id", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_id is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(trip), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	plan := trip.TripPlan
	result := fmt.Sprintf("%s (Key: %s)\n", plan.Title, plan.Key)
	if plan.StartDate != "" && plan.EndDate != "" {
		result += fmt.Sprintf("📅 %s to %s (%d days)\n", plan.StartDate, plan.EndDate, plan.Days)
	}
	result += fmt.Sprintf("📍 %d places | 👁 %d views | ❤ %d likes\n", plan.PlaceCount, plan.ViewCount, plan.LikeCount)
	if plan.AuthorBlurb != "" {
		// Truncate description if too long
		desc := plan.AuthorBlurb
		if len(desc) > 200 {
			desc = desc[:197] + "..."
		}
		result += fmt.Sprintf("📝 %s\n", desc)
	}

	return mcp.NewToolResultText(result), nil
}

func handleListPlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_id", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_id is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	// Use GetTrip instead of GetTripPlaces to get the full place metadata
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip: %v", err)), nil
	}

	places := trip.Resources.PlaceMetadata

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(places), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	var result string
	if len(places) == 0 {
		result = fmt.Sprintf("No places in trip %s", tripKey)
	} else {
		result = fmt.Sprintf("%s: %d places\n", trip.TripPlan.Title, len(places))

		// Limit to first 15 places to avoid exceeding char limit
		limit := len(places)
		if limit > 15 {
			limit = 15
		}

		for i := 0; i < limit; i++ {
			place := places[i]
			// Compact format: name + rating + address
			name := place.Name
			if place.Rating > 0 {
				name += fmt.Sprintf(" (%.1f⭐)", place.Rating)
			}
			result += fmt.Sprintf("%d. %s\n", i+1, name)

			// Only show address if available (most important info)
			if place.Address != "" {
				// Truncate long addresses
				addr := place.Address
				if len(addr) > 50 {
					addr = addr[:47] + "..."
				}
				result += fmt.Sprintf("   %s\n", addr)
			}
		}
		if len(places) > 15 {
			result += fmt.Sprintf("\n... and %d more (use JSON format for full list)\n", len(places)-15)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func handleListSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_id", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_id is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip: %v", err)), nil
	}

	sections := trip.TripPlan.Itinerary.Sections

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(sections), nil
	}

	// Format as text for default view (concise to stay under 2500 char limit)
	var result string
	if len(sections) == 0 {
		result = fmt.Sprintf("No sections in trip %s", tripKey)
	} else {
		result = fmt.Sprintf("%s: %d sections\n", trip.TripPlan.Title, len(sections))

		for i, section := range sections {
			// Compact format: heading + date + ID + item count
			heading := section.Heading
			if heading == "" {
				heading = "Untitled"
			}
			result += fmt.Sprintf("%d. %s [ID:%d]", i+1, heading, section.ID)

			if section.Date != nil && *section.Date != "" {
				result += fmt.Sprintf(" (%s)", *section.Date)
			}

			if len(section.Blocks) > 0 {
				result += fmt.Sprintf(" - %d items", len(section.Blocks))
			}
			result += "\n"
		}

		result += "\n💡 Use section IDs with add_place tool\n"
	}

	return mcp.NewToolResultText(result), nil
}

func handleGetTripSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	sections, err := client.GetTripSections(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get trip sections: %v", err)), nil
	}

	// Return structured data directly - more efficient than text formatting
	return mcp.NewToolResultStructuredOnly(sections), nil
}

func optionalIntArgument(request mcp.CallToolRequest, key string) (int, bool) {
	args := request.GetArguments()
	if args == nil {
		return 0, false
	}
	if _, ok := args[key]; !ok {
		return 0, false
	}
	return request.GetInt(key, 0), true
}

func validateDateArgument(name, value string, required bool) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return "", fmt.Errorf("%s is required in YYYY-MM-DD format", name)
		}
		return "", nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", name)
	}
	return value, nil
}

func validateDateRange(startName, startValue, endName, endValue string, required bool) (string, string, error) {
	startValue, err := validateDateArgument(startName, startValue, required)
	if err != nil {
		return "", "", err
	}
	endValue, err = validateDateArgument(endName, endValue, required)
	if err != nil {
		return "", "", err
	}
	if startValue == "" || endValue == "" {
		return startValue, endValue, nil
	}
	start, _ := time.Parse("2006-01-02", startValue)
	end, _ := time.Parse("2006-01-02", endValue)
	if end.Before(start) {
		return "", "", fmt.Errorf("%s must be on or after %s", endName, startName)
	}
	return startValue, endValue, nil
}

func resolveTripKey(ctx context.Context, request mcp.CallToolRequest, argName string) (string, error) {
	tripKey := request.GetString(argName, "")
	if tripKey != "" {
		return tripKey, nil
	}
	if defaultTripID, ok := tripIDFromContext(ctx); ok {
		return defaultTripID, nil
	}
	return "", fmt.Errorf("%s is required (either as parameter or default trip key must be set)", argName)
}

func availableSectionDates(sections []wanderlog.ItSections) string {
	dates := make([]string, 0, len(sections))
	for _, section := range sections {
		if section.Date != nil && strings.TrimSpace(*section.Date) != "" {
			dates = append(dates, fmt.Sprintf("%s:%d", strings.TrimSpace(*section.Date), section.ID))
		}
	}
	if len(dates) == 0 {
		return "none"
	}
	return strings.Join(dates, ", ")
}

func loadSectionsForResolution(client *wanderlog.Client, tripKey string) ([]wanderlog.ItSections, error) {
	trip, err := client.GetTrip(tripKey)
	if err == nil && len(trip.TripPlan.Itinerary.Sections) > 0 {
		return trip.TripPlan.Itinerary.Sections, nil
	}

	sections, err := client.GetTripSections(tripKey)
	if err == nil && len(sections) > 0 {
		return sections, nil
	}
	if err != nil {
		return nil, fmt.Errorf("full trip lookup failed and sections endpoint failed: %w", err)
	}
	return nil, fmt.Errorf("full trip lookup returned no itinerary sections")
}

func resolveSectionFromList(sections []wanderlog.ItSections, sectionID int, hasSectionID bool, sectionDate string, requireDated bool) (int, string, error) {
	if hasSectionID {
		for _, section := range sections {
			if section.ID != sectionID {
				continue
			}
			if requireDated && (section.Date == nil || strings.TrimSpace(*section.Date) == "") {
				return 0, "", fmt.Errorf("section_id %d is not a dated itinerary section; use list_sections and choose a section with a date", sectionID)
			}
			if section.Date != nil && strings.TrimSpace(*section.Date) != "" {
				return section.ID, fmt.Sprintf("%s section ID %d", strings.TrimSpace(*section.Date), section.ID), nil
			}
			return section.ID, fmt.Sprintf("section ID %d", section.ID), nil
		}
		return 0, "", fmt.Errorf("section_id %d was not found in this trip; available dated sections: %s", sectionID, availableSectionDates(sections))
	}

	for _, section := range sections {
		if section.Date != nil && strings.TrimSpace(*section.Date) == sectionDate {
			return section.ID, fmt.Sprintf("%s section ID %d", sectionDate, section.ID), nil
		}
	}

	return 0, "", fmt.Errorf("no itinerary section found for date %s; available dated sections: %s", sectionDate, availableSectionDates(sections))
}

func resolveAddPlaceSectionID(client *wanderlog.Client, tripKey string, request mcp.CallToolRequest) (int, string, error) {
	unscheduled := request.GetBool("unscheduled", false)
	sectionDate, err := validateDateArgument("section_date", request.GetString("section_date", ""), false)
	if err != nil {
		return 0, "", err
	}
	sectionID, hasSectionID := optionalIntArgument(request, "section_id")

	if unscheduled {
		if sectionDate != "" || (hasSectionID && sectionID > 0) {
			return 0, "", fmt.Errorf("use either unscheduled=true or a dated section, not both")
		}
		return 0, "general Places to visit list", nil
	}

	if hasSectionID {
		if sectionDate != "" {
			return 0, "", fmt.Errorf("use either section_id or section_date, not both")
		}
		if sectionID <= 0 {
			return 0, "", fmt.Errorf("section_id must be a positive itinerary section ID. Use list_sections or section_date; pass unscheduled=true to add to the general Places to visit list")
		}
	} else if sectionDate == "" {
		return 0, "", fmt.Errorf("section_id or section_date is required. Use list_sections to find dated sections, or pass unscheduled=true to add to the general Places to visit list")
	}

	sections, err := loadSectionsForResolution(client, tripKey)
	if err != nil {
		return 0, "", fmt.Errorf("failed to resolve section: %w", err)
	}
	return resolveSectionFromList(sections, sectionID, hasSectionID, sectionDate, false)
}

func resolveDatedSectionID(client *wanderlog.Client, tripKey string, request mcp.CallToolRequest, defaultDate string) (int, string, error) {
	sectionDate, err := validateDateArgument("section_date", request.GetString("section_date", ""), false)
	if err != nil {
		return 0, "", err
	}
	sectionID, hasSectionID := optionalIntArgument(request, "section_id")

	if sectionDate == "" {
		sectionDate = strings.TrimSpace(defaultDate)
	}

	if hasSectionID {
		if request.GetString("section_date", "") != "" {
			return 0, "", fmt.Errorf("use either section_id or section_date, not both")
		}
		if sectionID <= 0 {
			return 0, "", fmt.Errorf("section_id must be a positive itinerary section ID")
		}
	} else if sectionDate == "" {
		return 0, "", fmt.Errorf("section_id or section_date is required")
	}

	sections, err := loadSectionsForResolution(client, tripKey)
	if err != nil {
		return 0, "", fmt.Errorf("failed to resolve section date: %w", err)
	}
	return resolveSectionFromList(sections, sectionID, hasSectionID, sectionDate, true)
}

func currentUserID() int {
	auth, err := wanderlog.LoadCredentials()
	if err != nil || auth.UserID == "" {
		return 0
	}
	userID, err := strconv.Atoi(auth.UserID)
	if err != nil {
		return 0
	}
	return userID
}

func appendItineraryBlock(client *wanderlog.Client, tripKey string, sectionID int, block map[string]any) error {
	if err := validateBlockSchema(block); err != nil {
		return fmt.Errorf("invalid block schema: %w", err)
	}
	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	sectionIdx := wanderlog.FindSectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
	if sectionIdx < 0 {
		return fmt.Errorf("section %d not found", sectionID)
	}

	maxBlockID := 0
	for _, section := range trip.TripPlan.Itinerary.Sections {
		for _, block := range section.Blocks {
			if block.ID > maxBlockID {
				maxBlockID = block.ID
			}
		}
	}
	if _, ok := block["id"]; !ok {
		block["id"] = maxBlockID + 1
	}

	if _, ok := block["addedBy"]; !ok {
		addedBy := map[string]any{"type": "user"}
		if userID := currentUserID(); userID > 0 {
			addedBy["userId"] = userID
		}
		block["addedBy"] = addedBy
	}
	if _, ok := block["attachments"]; !ok {
		block["attachments"] = []any{}
	}
	if _, ok := block["upvotedBy"]; !ok {
		block["upvotedBy"] = []any{}
	}

	position := len(trip.TripPlan.Itinerary.Sections[sectionIdx].Blocks)
	op := wanderlog.InsertInList(
		[]interface{}{"itinerary", "sections", sectionIdx, "blocks"},
		position,
		block,
	)
	return client.ApplyOperations(tripKey, []wanderlog.Operation{op})
}

func splitFlightNumber(flightNumber string) (string, int) {
	flightNumber = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(flightNumber), " ", ""))
	prefixEnd := 0
	for prefixEnd < len(flightNumber) && (flightNumber[prefixEnd] < '0' || flightNumber[prefixEnd] > '9') {
		prefixEnd++
	}
	number, _ := strconv.Atoi(flightNumber[prefixEnd:])
	return flightNumber[:prefixEnd], number
}

func validateFlightNumber(flightNumber string) (string, int, error) {
	airline, number := splitFlightNumber(flightNumber)
	if airline == "" || number <= 0 {
		return "", 0, fmt.Errorf("flight_number must include an airline code and number, e.g. MU244")
	}
	return airline, number, nil
}

// validateBlockSchema ensures block structures are compatible with the web/mobile app.
// Flight blocks must not set depart.type="airport" or arrive.type="airport" without
// providing a fully populated airport sub-object (with googlePlace), otherwise the
// web app's station parser will crash trying to access undefined fields.
func validateBlockSchema(block map[string]any) error {
	if block["type"] != "flight" {
		return nil
	}

	depart, hasDepart := block["depart"].(map[string]any)
	arrive, hasArrive := block["arrive"].(map[string]any)

	validateStation := func(name string, station map[string]any) error {
		stationType, hasType := station["type"].(string)
		if !hasType || stationType != "airport" {
			return nil // no airport type, no validation needed
		}
		airport, hasAirport := station["airport"].(map[string]any)
		if !hasAirport {
			return fmt.Errorf("flight block %s has type=airport but no airport sub-object; set type=airport only when providing a complete airport object with googlePlace", name)
		}
		if _, hasGooglePlace := airport["googlePlace"]; !hasGooglePlace {
			return fmt.Errorf("flight block %s has type=airport but airport.googlePlace is missing; either provide the complete airport object with googlePlace or omit type=airport entirely", name)
		}
		return nil
	}

	if hasDepart {
		if err := validateStation("depart", depart); err != nil {
			return err
		}
	}
	if hasArrive {
		if err := validateStation("arrive", arrive); err != nil {
			return err
		}
	}
	return nil
}

func handleAddPlace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	name, err := request.RequireString("name")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("name is required"), nil //nolint:nilerr
	}

	placeID := request.GetString("place_id", "")
	latitude := request.GetFloat("latitude", 0)
	longitude := request.GetFloat("longitude", 0)
	text := request.GetString("text", "")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	sectionID, sectionLabel, err := resolveAddPlaceSectionID(client, tripKey, request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// CRITICAL FIX: If place_id is provided but coordinates are missing, fetch them first
	// This prevents creating places without location data, which breaks trips with:
	// "TypeError: Cannot read properties of undefined (reading 'location')"
	if placeID != "" && (latitude == 0 && longitude == 0) {
		logger.WithField("place_id", placeID).Debug("Fetching place details to get coordinates")

		placeDetails, err := client.GetPlaceDetails(placeID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch place details for coordinates: %v. Please provide latitude and longitude parameters.", err)), nil
		}

		// Extract coordinates from place details
		latitude = placeDetails.Data.Details.Geometry.Location.Lat
		longitude = placeDetails.Data.Details.Geometry.Location.Lng

		// Also use the canonical name from the API if user didn't override it
		if name == "" || name == placeID {
			name = placeDetails.Data.Details.Name
		}

		logger.WithFields(map[string]interface{}{
			"place_id": placeID,
			"name":     name,
			"lat":      latitude,
			"lng":      longitude,
		}).Debug("Fetched coordinates from place details")
	}

	// Build the place info with proper geometry structure
	placeInfo := wanderlog.AddPlaceInfo{
		PlaceID: placeID,
		Name:    name,
	}

	// CRITICAL: Always require coordinates when place_id is provided
	// The Wanderlog API technically accepts places without geometry, but they become
	// corrupt and break the trip with "Cannot read properties of undefined (reading 'location')"
	if placeID != "" {
		if latitude == 0 && longitude == 0 {
			return mcp.NewToolResultError("Coordinates (latitude/longitude) are required when adding a place. This should not happen - please report this bug."), nil
		}

		placeInfo.Geometry = &models.PlaceGeometry{
			Location: models.PlaceLocation{
				Lat: latitude,
				Lng: longitude,
			},
		}
	} else if latitude != 0 || longitude != 0 {
		// Only add geometry if coordinates are provided for places without place_id
		placeInfo.Geometry = &models.PlaceGeometry{
			Location: models.PlaceLocation{
				Lat: latitude,
				Lng: longitude,
			},
		}
	}

	req := wanderlog.AddPlaceRequest{
		Place: placeInfo,
		Text:  text,
	}

	err = client.AddPlace(tripKey, sectionID, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add place: %v", err)), nil
	}

	result := fmt.Sprintf("📍 Successfully added place '%s' to trip %s (%s)", name, tripKey, sectionLabel)

	return mcp.NewToolResultText(result), nil
}

func handleAddFlight(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey, err := resolveTripKey(ctx, request, "trip_key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	flightNumber, err := request.RequireString("flight_number")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("flight_number is required"), nil //nolint:nilerr
	}

	departureDate, err := validateDateArgument("departure_date", request.GetString("departure_date", ""), true)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	departureTime := request.GetString("departure_time", "")
	arrivalDate, err := validateDateArgument("arrival_date", request.GetString("arrival_date", ""), false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if arrivalDate == "" {
		arrivalDate = departureDate
	}
	arrivalTime := request.GetString("arrival_time", "")
	confirmationNumber := request.GetString("confirmation_number", "")
	notes := request.GetString("notes", "")

	client := wanderlog.NewClient()
	client.SetLogger(logger)
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	sectionID, sectionLabel, err := resolveDatedSectionID(client, tripKey, request, departureDate)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	airlineIATA, flightNum, err := validateFlightNumber(flightNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	block := map[string]any{
		"type":               "flight",
		"confirmationNumber": confirmationNumber,
		"startTime":          departureTime,
		"endTime":            arrivalTime,
		"depart": map[string]any{
			"date": departureDate,
			"time": departureTime,
		},
		"arrive": map[string]any{
			"date": arrivalDate,
			"time": arrivalTime,
		},
		"flightInfo": map[string]any{
			"airline": map[string]any{
				"iata": airlineIATA,
			},
			"number": flightNum,
		},
		"text":          notes,
		"travelMode":    nil,
		"travelerNames": []any{},
	}

	if err := appendItineraryBlock(client, tripKey, sectionID, block); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add flight: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("✈️ Successfully added flight %s to trip %s (%s)", flightNumber, tripKey, sectionLabel)), nil
}

func handleRemovePlace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	placeID, err := request.RequireInt("place_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_id is required"), nil
	}

	sectionID := request.GetInt("section_id", 0)

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err = client.RemovePlace(tripKey, sectionID, placeID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove place: %v", err)), nil
	}

	result := fmt.Sprintf("🗑️ Successfully removed place %d from trip %s", placeID, tripKey)
	if sectionID > 0 {
		result += fmt.Sprintf(" (Section ID: %d)", sectionID)
	}

	return mcp.NewToolResultText(result), nil
}

// Resource handler
func handleTripResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract trip_id from URI like "wanderlog://trips/abc123"
	uri := request.Params.URI
	tripKey := ""

	// Simple parsing - in production you might want more robust URI parsing
	prefix := "wanderlog://trips/"
	if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
		tripKey = uri[len(prefix):]
	}

	if tripKey == "" {
		return nil, fmt.Errorf("invalid trip URI format")
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	jsonData, err := json.Marshal(trip)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trip data: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleSearchPlaces searches for places using Google Places API
func handleSearchPlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("query is required"), nil
	}

	// Get API key from environment variable
	apiKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if apiKey == "" {
		return mcp.NewToolResultError("GOOGLE_PLACES_API_KEY environment variable is required. Set it with: export GOOGLE_PLACES_API_KEY=your_key_here"), nil
	}

	format := request.GetString("format", "default")

	// Parse optional coordinates
	var lat, lng *float64
	if latStr := request.GetString("latitude", ""); latStr != "" {
		if latNum, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = &latNum
		}
	}
	if lngStr := request.GetString("longitude", ""); lngStr != "" {
		if lngNum, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = &lngNum
		}
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	results, err := client.SearchPlaces(query, lat, lng, apiKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	switch format {
	case "json":
		jsonData, err := json.Marshal(results)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(jsonData)), nil

	case "markdown":
		if len(results.Places) == 0 {
			return mcp.NewToolResultText("# Search Results\n\nNo places found."), nil
		}

		result := "# Search Results\n\n"
		for i, place := range results.Places {
			result += fmt.Sprintf("## %d. %s\n\n", i+1, place.Name)

			if place.Rating > 0 {
				result += fmt.Sprintf("**Rating:** %.1f/5 ⭐\n\n", place.Rating)
			}

			if place.Address != "" {
				result += fmt.Sprintf("**Address:** %s\n\n", place.Address)
			}

			if len(place.Categories) > 0 {
				categories := ""
				for j, cat := range place.Categories {
					categories += cat
					if j < len(place.Categories)-1 {
						categories += ", "
					}
				}
				result += fmt.Sprintf("**Categories:** %s\n\n", categories)
			}

			if place.Description != "" {
				result += fmt.Sprintf("**Description:** %s\n\n", place.Description)
			}

			if place.Website != "" {
				result += fmt.Sprintf("**Website:** %s\n\n", place.Website)
			}

			if place.Latitude != 0 && place.Longitude != 0 {
				result += fmt.Sprintf("**Location:** %.4f, %.4f\n\n", place.Latitude, place.Longitude)
			}

			if place.PlaceID != "" {
				result += fmt.Sprintf("**Place ID:** %s\n\n", place.PlaceID)
			}

			result += "---\n\n"
		}
		return mcp.NewToolResultText(result), nil

	default:
		if len(results.Places) == 0 {
			return mcp.NewToolResultText("🔍 No places found"), nil
		}

		result := "🔍 Search Results\n\n"
		for i, place := range results.Places {
			// Place name with rating
			name := place.Name
			if place.Rating > 0 {
				stars := ""
				for j := 0; j < int(place.Rating) && j < 5; j++ {
					stars += "⭐"
				}
				name += fmt.Sprintf(" %s (%.1f)", stars, place.Rating)
			}

			result += fmt.Sprintf("📍 %s\n", name)

			// Address
			if place.Address != "" {
				result += fmt.Sprintf("   🏠 %s\n", place.Address)
			}

			// Categories
			if len(place.Categories) > 0 {
				categories := ""
				for j, cat := range place.Categories {
					categories += cat
					if j < len(place.Categories)-1 {
						categories += ", "
					}
				}
				result += fmt.Sprintf("   🏷️  %s\n", categories)
			}

			// Description
			if place.Description != "" {
				result += fmt.Sprintf("   📝 %s\n", place.Description)
			}

			// Website
			if place.Website != "" {
				result += fmt.Sprintf("   🌐 %s\n", place.Website)
			}

			// Coordinates
			if place.Latitude != 0 && place.Longitude != 0 {
				result += fmt.Sprintf("   🗺️  %.4f, %.4f\n", place.Latitude, place.Longitude)
			}

			// Place ID
			if place.PlaceID != "" {
				result += fmt.Sprintf("   🆔 %s\n", place.PlaceID)
			}

			if i < len(results.Places)-1 {
				result += "\n"
			}
		}
		return mcp.NewToolResultText(result), nil
	}
}

// Handler for get_place_details tool
func handleGetPlaceDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	placeID, err := request.RequireString("place_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_id is required"), nil
	}

	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	details, err := client.GetPlaceDetails(placeID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting place details: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(details), nil
	}

	result := fmt.Sprintf("🏛️ **%s**\n\n", details.Data.Details.Name)
	result += fmt.Sprintf("📍 **Place ID:** %s\n", details.Data.Details.PlaceID)
	result += fmt.Sprintf("🏠 **Address:** %s\n", details.Data.Details.FormattedAddress)

	if details.Data.Details.Rating > 0 {
		result += fmt.Sprintf("⭐ **Rating:** %.1f/5 (%d reviews)\n",
			details.Data.Details.Rating, details.Data.Details.UserRatingsTotal)
	}

	if details.Data.Details.Website != "" {
		result += fmt.Sprintf("🌐 **Website:** %s\n", details.Data.Details.Website)
	}

	if details.Data.Details.InternationalPhoneNumber != "" {
		result += fmt.Sprintf("📞 **Phone:** %s\n", details.Data.Details.InternationalPhoneNumber)
	}

	if len(details.Data.Details.Types) > 0 {
		result += fmt.Sprintf("🏷️ **Types:** %v\n", details.Data.Details.Types)
	}

	coords := details.Data.Details.Geometry.Location
	result += fmt.Sprintf("🗺️ **Coordinates:** %.6f, %.6f\n", coords.Lat, coords.Lng)

	if details.Data.CardData.ReviewsSummary != "" {
		result += fmt.Sprintf("\n📊 **Reviews Summary:**\n%s\n", details.Data.CardData.ReviewsSummary)
	}

	if len(details.Data.CardData.ReasonsToVisit) > 0 {
		result += "\n✨ **Reasons to Visit:**\n"
		for i, reason := range details.Data.CardData.ReasonsToVisit {
			result += fmt.Sprintf("  %d. %s\n", i+1, reason)
		}
	}

	if len(details.Data.CardData.Tips) > 0 {
		result += "\n💡 **Tips:**\n"
		for i, tip := range details.Data.CardData.Tips {
			result += fmt.Sprintf("  %d. %s\n", i+1, tip)
		}
	}

	return mcp.NewToolResultText(result), nil
}

// Handler for search_places_wanderlog tool
func handleSearchPlacesWanderlog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("query is required"), nil
	}

	lat := request.GetFloat("latitude", 0.0)
	lng := request.GetFloat("longitude", 0.0)
	format := request.GetString("format", "default")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.SearchPlacesWithWanderllog(query, lat, lng)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching places: %v", err)), nil
	}

	if format == "json" {
		return mcp.NewToolResultStructuredOnly(results), nil
	}

	// Filter out junk rows (empty descriptions or placeholder types like just "search")
	validCount := 0
	for _, place := range results.Data {
		if place.Description != "" && place.Description != place.Type {
			validCount++
		}
	}

	if validCount == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No valid places found for query: %s (try being more specific)", query)), nil
	}

	result := fmt.Sprintf("🔍 **Found %d places for query:** %s\n\n", validCount, query)

	validIndex := 0
	for _, place := range results.Data {
		if place.Description == "" || place.Description == place.Type {
			continue
		}
		validIndex++
		result += fmt.Sprintf("**%d. %s**\n", validIndex, place.Description)
		if place.PlaceID != "" {
			result += fmt.Sprintf("   📍 Place ID: %s\n", place.PlaceID)
		}
		if len(place.Types) > 0 {
			result += fmt.Sprintf("   🏷️ Types: %v\n", place.Types)
		}
		if place.Type != "" {
			result += fmt.Sprintf("   🏷️ Type: %s\n", place.Type)
		}
		if validIndex < validCount {
			result += "\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

// Prompt handler
func handleAnalyzeTrip(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	tripID := ""
	focus := "overall"

	if args := request.Params.Arguments; args != nil {
		if id, exists := args["trip_id"]; exists {
			tripID = id
		}
		if f, exists := args["focus"]; exists {
			focus = f
		}
	}

	if tripID == "" {
		return nil, fmt.Errorf("trip_id argument is required")
	}

	var promptText string
	switch focus {
	case "budget":
		promptText = fmt.Sprintf("Please analyze the budget and expenses for trip %s. Look at the costs, suggest ways to save money, and identify any budget concerns.", tripID)
	case "itinerary":
		promptText = fmt.Sprintf("Please analyze the itinerary for trip %s. Look at the schedule, timing, transportation, and suggest optimizations or improvements.", tripID)
	case "places":
		promptText = fmt.Sprintf("Please analyze the places and destinations for trip %s. Evaluate the selection, suggest additional places to visit, and identify any missing must-see locations.", tripID)
	default:
		promptText = fmt.Sprintf("Please provide a comprehensive analysis of trip %s. Include insights on the itinerary, budget, places to visit, and overall trip planning. Suggest improvements and highlight any potential issues.", tripID)
	}

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Analyze trip %s with focus on %s", tripID, focus),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// Handler functions for new MCP tools

func handleCreateTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title := request.GetString("title", "")
	if title == "" {
		return mcp.NewToolResultError("title is required"), nil
	}

	startDate, endDate, err := validateDateRange("start_date", request.GetString("start_date", ""), "end_date", request.GetString("end_date", ""), true)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	privacy := request.GetString("privacy", "private")
	geoID := int(request.GetFloat("geo_id", 0))
	if geoID == 0 {
		return mcp.NewToolResultError("geo_id is required"), nil
	}
	if geoID < 0 {
		return mcp.NewToolResultError("geo_id must be a positive Wanderlog geo ID from search_geos"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	req := wanderlog.CreateTripRequest{
		Title:               title,
		GeoIDs:              []int{geoID},
		InitialMapsPlaceIDs: []int{},
		Type:                "plan",
		StartDate:           startDate,
		EndDate:             endDate,
		Privacy:             privacy,
	}

	resp, err := client.CreateTrip(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create trip: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Created trip '%s' (Key: %s, ID: %d)", resp.TripPlan.Title, resp.TripPlan.Key, resp.TripPlan.ID)
	return mcp.NewToolResultText(result), nil
}

func handleDeleteTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err := client.DeleteTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete trip: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("🗑️ Deleted trip %s", tripKey)), nil
}

func handleDeleteTrips(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKeysStr := request.GetString("trip_keys", "")
	if tripKeysStr == "" {
		return mcp.NewToolResultError("trip_keys is required (comma-separated list)"), nil
	}

	// Parse comma-separated trip keys
	var tripKeys []string
	for _, key := range strings.Split(tripKeysStr, ",") {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			tripKeys = append(tripKeys, trimmed)
		}
	}

	if len(tripKeys) == 0 {
		return mcp.NewToolResultError("trip_keys must contain at least one trip key"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	var deleted []string
	var failed []string

	for _, tripKey := range tripKeys {
		err := client.DeleteTrip(tripKey)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", tripKey, err))
		} else {
			deleted = append(deleted, tripKey)
		}
	}

	if len(failed) > 0 {
		return mcp.NewToolResultError(fmt.Sprintf("Deleted %d trips, but failed to delete: %s", len(deleted), strings.Join(failed, ", "))), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("🗑️ Deleted %d trips: %s", len(deleted), strings.Join(deleted, ", "))), nil
}

func handleRestoreTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err := client.RestoreTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to restore trip: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("♻️ Restored trip %s", tripKey)), nil
}

func handleCopyTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	resp, err := client.CopyTrip(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to copy trip: %v", err)), nil
	}

	result := fmt.Sprintf("📋 Copied trip to '%s' (Key: %s)", resp.TripPlan.Title, resp.TripPlan.Key)
	return mcp.NewToolResultText(result), nil
}

func handleUpdateTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	title := request.GetString("title", "")
	startDate := request.GetString("start_date", "")
	endDate := request.GetString("end_date", "")
	privacy := request.GetString("privacy", "")

	// Validate that at least one field is provided
	if title == "" && startDate == "" && endDate == "" && privacy == "" {
		return mcp.NewToolResultError("At least one field must be provided (title, start_date, end_date, or privacy)"), nil
	}

	// Validate privacy if provided
	if privacy != "" && privacy != "public" && privacy != "private" && privacy != "unlisted" {
		return mcp.NewToolResultError("privacy must be one of: public, private, unlisted"), nil
	}
	if startDate != "" || endDate != "" {
		var err error
		startDate, endDate, err = validateDateRange("start_date", startDate, "end_date", endDate, false)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	updateReq := models.UpdateTripRequest{
		Title:     title,
		StartDate: startDate,
		EndDate:   endDate,
		Privacy:   privacy,
	}

	if err := client.UpdateTrip(tripKey, updateReq); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update trip: %v", err)), nil
	}

	// Build result message showing what was updated
	updates := []string{}
	if title != "" {
		updates = append(updates, fmt.Sprintf("title to '%s'", title))
	}
	if startDate != "" {
		updates = append(updates, fmt.Sprintf("start date to %s", startDate))
	}
	if endDate != "" {
		updates = append(updates, fmt.Sprintf("end date to %s", endDate))
	}
	if privacy != "" {
		updates = append(updates, fmt.Sprintf("privacy to %s", privacy))
	}

	result := fmt.Sprintf("✅ Updated trip %s: %s", tripKey, strings.Join(updates, ", "))
	return mcp.NewToolResultText(result), nil
}

func handleLikeTrip(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	liked := request.GetBool("liked", false)

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err := client.SetLike(tripKey, liked)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to set like: %v", err)), nil
	}

	action := "liked"
	if !liked {
		action = "unliked"
	}
	return mcp.NewToolResultText(fmt.Sprintf("👍 Successfully %s trip %s", action, tripKey)), nil
}

func handleGetLikeCount(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	likeCount, err := client.GetLikeCount(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get like count: %v", err)), nil
	}

	result := fmt.Sprintf("Trip %s has %d likes", tripKey, likeCount.Count)
	if likeCount.UserLiked {
		result += " (you liked this trip)"
	}
	return mcp.NewToolResultText(result), nil
}

func handleSendInvites(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	inviteesStr := request.GetString("invitees", "")
	if inviteesStr == "" {
		return mcp.NewToolResultError("invitees is required"), nil
	}

	message := request.GetString("message", "")

	// Parse comma-separated invitees
	invitees := strings.Split(inviteesStr, ",")
	for i := range invitees {
		invitees[i] = strings.TrimSpace(invitees[i])
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	req := wanderlog.SendInvitesRequest{
		Invitees: invitees,
		Message:  message,
	}

	err := client.SendTripInvites(tripKey, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to send invites: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("📧 Sent invites to %d people for trip %s", len(invitees), tripKey)), nil
}

func handleListInvites(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		return mcp.NewToolResultError("trip_key is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	invites, err := client.ListTripInvites(tripKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list invites: %v", err)), nil
	}

	if len(invites) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No invites found for trip %s", tripKey)), nil
	}

	result := fmt.Sprintf("Invites for trip %s:\n", tripKey)
	for i, invite := range invites {
		result += fmt.Sprintf("%d. %s - Status: %s (Sent: %s)\n", i+1, invite.Email, invite.Status, invite.InvitedAt)
	}

	return mcp.NewToolResultText(result), nil
}

func handleMovePlace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	placeID, err := request.RequireInt("place_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_id is required"), nil
	}

	fromSectionID, err := request.RequireInt("from_section_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("from_section_id is required"), nil
	}

	toSectionID, err := request.RequireInt("to_section_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("to_section_id is required"), nil
	}

	position, err := request.RequireInt("position")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("position is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err = client.MovePlace(tripKey, placeID, fromSectionID, toSectionID, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move place: %v", err)), nil
	}

	result := fmt.Sprintf("🔀 Successfully moved place %d from section %d to section %d at position %d",
		placeID, fromSectionID, toSectionID, position)

	return mcp.NewToolResultText(result), nil
}

func handleReorderPlaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tripKey := request.GetString("trip_key", "")
	if tripKey == "" {
		// Try to get from context
		if defaultTripID, ok := tripIDFromContext(ctx); ok {
			tripKey = defaultTripID
		} else {
			return mcp.NewToolResultError("trip_key is required (either as parameter or default trip ID must be set)"), nil
		}
	}

	sectionID, err := request.RequireInt("section_id")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("section_id is required"), nil
	}

	placeIDsStr, err := request.RequireString("place_ids")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("place_ids is required"), nil
	}

	// Parse comma-separated place IDs
	placeIDStrs := strings.Split(placeIDsStr, ",")
	placeIDs := make([]int, 0, len(placeIDStrs))
	for _, idStr := range placeIDStrs {
		idStr = strings.TrimSpace(idStr)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid place ID '%s': %v", idStr, err)), nil
		}
		placeIDs = append(placeIDs, id)
	}

	if len(placeIDs) == 0 {
		return mcp.NewToolResultError("place_ids must contain at least one place ID"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	err = client.ReorderPlaces(tripKey, sectionID, placeIDs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to reorder places: %v", err)), nil
	}

	result := fmt.Sprintf("📋 Successfully reordered %d places in section %d", len(placeIDs), sectionID)

	return mcp.NewToolResultText(result), nil
}

func handleGetFlightStops(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	flightNumber, err := request.RequireString("flight_number")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("flight_number is required"), nil
	}
	airline, err := request.RequireString("airline")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("airline is required"), nil
	}
	date, err := request.RequireString("date")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("date is required"), nil
	}

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.GetFlightStops(flightNumber, airline, date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting flight stops: %v", err)), nil
	}

	return mcp.NewToolResultStructuredOnly(results), nil
}

// handleSearchHotels searches for hotels/lodging
func handleSearchHotels(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	location, err := request.RequireString("location")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("location is required"), nil
	}

	checkIn, err := request.RequireString("check_in")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("check_in is required"), nil
	}

	checkOut, err := request.RequireString("check_out")
	if err != nil {
		_ = err
		return mcp.NewToolResultError("check_out is required"), nil
	}

	guests := request.GetInt("guests", 1)

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Set up authentication if available
	auth, err := wanderlog.LoadCredentials()
	if err == nil {
		client.SetAuth(auth)
	}

	results, err := client.SearchLodgings(location, checkIn, checkOut, guests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching hotels: %v", err)), nil
	}

	if !results.Success {
		return mcp.NewToolResultError(fmt.Sprintf("Hotel search failed for %s (%s to %s). The lodging API may be unavailable or your session may have expired.", location, checkIn, checkOut)), nil
	}

	return mcp.NewToolResultStructuredOnly(results), nil
}
