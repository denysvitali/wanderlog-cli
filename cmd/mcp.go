package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
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
- Viewing places and itineraries
- Searching for places
- Trip analysis and recommendations

Write operations (only with --enable-write):
- Adding places to trips
- Removing places from trips

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
	mcpCmd.Flags().String("trip-id", "", "Default trip ID to use for all operations (makes trip_id parameter optional in tools)")
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
			mcp.Description("The ID of the trip to retrieve (optional if default trip ID is set)"),
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
			mcp.Description("The ID of the trip to get places for (optional if default trip ID is set)"),
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
			mcp.Description("The ID of the trip to get sections for (optional if default trip ID is set)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(listSectionsTool, handleListSections)

	// Add write operation tools only if not in read-only mode
	if !readOnly {
		// Add place to trip tool
		addPlaceTool := mcp.NewTool("add_place",
			mcp.WithDescription("Add a place to a trip"),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip to add the place to (optional if default trip ID is set)"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the place to add"),
			),
			mcp.WithString("place_id",
				mcp.Description("Google Place ID (optional)"),
			),
			mcp.WithNumber("latitude",
				mcp.Description("Latitude of the place (optional)"),
			),
			mcp.WithNumber("longitude",
				mcp.Description("Longitude of the place (optional)"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID to add the place to (optional)"),
			),
			mcp.WithString("text",
				mcp.Description("Additional text/notes for the place (optional)"),
			),
		)
		s.AddTool(addPlaceTool, handleAddPlace)

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

	// Add trip resource
	tripResource := mcp.NewResource(
		"wanderlog://trips/{trip_id}",
		"Trip Details",
		mcp.WithResourceDescription("Detailed information about a specific Wanderlog trip"),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(tripResource, handleTripResource)

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

	// Format as text for default view
	var result string
	if len(trips.Data) == 0 {
		result = "No trips found."
	} else {
		result = fmt.Sprintf("Found %d trips:\n\n", len(trips.Data))
		for i, trip := range trips.Data {
			result += fmt.Sprintf("%d. %s (Key: %s)\n", i+1, trip.Title, trip.Key)
			if trip.StartDate != "" && trip.EndDate != "" {
				result += fmt.Sprintf("   Dates: %s to %s\n", trip.StartDate, trip.EndDate)
			}
			result += fmt.Sprintf("   Places: %d, Views: %d, Likes: %d\n", trip.PlaceCount, trip.ViewCount, trip.LikeCount)
			result += "\n"
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

	// Format as text for default view
	plan := trip.TripPlan
	result := fmt.Sprintf("Trip: %s\n", plan.Title)
	result += fmt.Sprintf("Key: %s\n", plan.Key)
	if plan.StartDate != "" {
		result += fmt.Sprintf("Start Date: %s\n", plan.StartDate)
	}
	if plan.EndDate != "" {
		result += fmt.Sprintf("End Date: %s\n", plan.EndDate)
	}
	result += fmt.Sprintf("Days: %d\n", plan.Days)
	result += fmt.Sprintf("Places: %d\n", plan.PlaceCount)
	result += fmt.Sprintf("Views: %d, Likes: %d\n", plan.ViewCount, plan.LikeCount)
	if plan.AuthorBlurb != "" {
		result += fmt.Sprintf("Description: %s\n", plan.AuthorBlurb)
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

	// Format as text for default view (matching CLI output)
	var result string
	if len(places) == 0 {
		result = fmt.Sprintf("No places found for trip %s.", tripKey)
	} else {
		result = fmt.Sprintf("Trip: %s\n", trip.TripPlan.Title)
		result += fmt.Sprintf("Found %d places:\n\n", len(places))

		for i, place := range places {
			// Place name with rating
			name := place.Name
			if place.Rating > 0 {
				name += fmt.Sprintf(" ⭐ (%.1f)", place.Rating)
			}
			result += fmt.Sprintf("%d. %s\n", i+1, name)

			// Address
			if place.Address != "" {
				result += fmt.Sprintf("   📍 %s\n", place.Address)
			}

			// Categories
			if len(place.Categories) > 0 {
				categories := ""
				for j, cat := range place.Categories {
					if j > 0 {
						categories += ", "
					}
					categories += cat
				}
				result += fmt.Sprintf("   🏷️  %s\n", categories)
			}

			// Website
			if place.Website != "" {
				result += fmt.Sprintf("   🌐 %s\n", place.Website)
			}

			// Phone
			if place.InternationalPhoneNumber != "" {
				result += fmt.Sprintf("   📞 %s\n", place.InternationalPhoneNumber)
			}

			// Description
			if place.Description != nil && *place.Description != "" {
				desc := *place.Description
				if len(desc) > 100 {
					desc = desc[:97] + "..."
				}
				result += fmt.Sprintf("   💬 %s\n", desc)
			}

			result += "\n"
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

	// Format as text for default view
	var result string
	if len(sections) == 0 {
		result = fmt.Sprintf("No sections found for trip %s.", tripKey)
	} else {
		result = fmt.Sprintf("Trip: %s\n", trip.TripPlan.Title)
		result += fmt.Sprintf("Found %d sections/days:\n\n", len(sections))

		for i, section := range sections {
			// Section header with ID and heading
			result += fmt.Sprintf("%d. %s (ID: %d)\n", i+1, section.Heading, section.ID)

			// Date
			if section.Date != nil && *section.Date != "" {
				result += fmt.Sprintf("   📅 %s\n", *section.Date)
			}

			// Section type and mode
			if section.Type != "" {
				result += fmt.Sprintf("   📋 Type: %s", section.Type)
				if section.Mode != "" {
					result += fmt.Sprintf(" (%s)", section.Mode)
				}
				result += "\n"
			}

			// Number of blocks/items
			if len(section.Blocks) > 0 {
				result += fmt.Sprintf("   📍 %d items\n", len(section.Blocks))
			}

			// Notes/text if available
			if len(section.Text.Ops) > 0 {
				for _, op := range section.Text.Ops {
					if op.Insert != "\n" && strings.TrimSpace(op.Insert) != "" {
						result += fmt.Sprintf("   📝 %s\n", strings.TrimSpace(op.Insert))
						break // Only show first meaningful text
					}
				}
			}

			result += "\n"
		}

		result += fmt.Sprintf("\n💡 Section IDs can be used with add_place tool to add places to specific days\n")
		result += fmt.Sprintf("   (Note: write operations require --enable-write flag)\n")
	}

	return mcp.NewToolResultText(result), nil
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
		return mcp.NewToolResultError("name is required"), nil
	}

	placeID := request.GetString("place_id", "")
	latitude := request.GetFloat("latitude", 0)
	longitude := request.GetFloat("longitude", 0)
	sectionID := request.GetInt("section_id", 0)
	text := request.GetString("text", "")

	client := wanderlog.NewClient()
	client.SetLogger(logger)

	// Ensure authentication
	if err := client.EnsureAuthenticated("", ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}

	// Build the place info with proper geometry structure
	placeInfo := wanderlog.AddPlaceInfo{
		PlaceID: placeID,
		Name:    name,
	}

	// Only add geometry if coordinates are provided
	if latitude != 0 || longitude != 0 {
		placeInfo.Geometry = &struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		}{
			Location: struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}{
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

	result := fmt.Sprintf("📍 Successfully added place '%s' to trip %s", name, tripKey)
	if sectionID > 0 {
		result += fmt.Sprintf(" (Section ID: %d)", sectionID)
	}

	return mcp.NewToolResultText(result), nil
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
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	trip, err := client.GetTrip(tripKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %v", err)
	}

	jsonData, err := json.Marshal(trip)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trip data: %v", err)
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

	if len(results.Data) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No places found for query: %s", query)), nil
	}

	result := fmt.Sprintf("🔍 **Found %d places for query:** %s\n\n", len(results.Data), query)

	for i, place := range results.Data {
		result += fmt.Sprintf("**%d. %s**\n", i+1, place.Description)
		if place.PlaceID != "" {
			result += fmt.Sprintf("   📍 Place ID: %s\n", place.PlaceID)
		}
		if len(place.Types) > 0 {
			result += fmt.Sprintf("   🏷️ Types: %v\n", place.Types)
		}
		if place.Type != "" {
			result += fmt.Sprintf("   🏷️ Type: %s\n", place.Type)
		}
		if i < len(results.Data)-1 {
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
