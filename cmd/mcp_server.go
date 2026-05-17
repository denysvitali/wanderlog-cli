package cmd

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
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
		mcp.WithDescription("Get the full trip plan for a specific trip key, including itinerary sections and trip resources"),
		mcp.WithString("trip_id",
			mcp.Description("The trip key to retrieve, not the numeric database ID (optional if default trip key is set)"),
		),
		mcp.WithString("trip_key",
			mcp.Description("Alias for trip_id; the shared trip key to retrieve"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (default, json)"),
			mcp.DefaultString("default"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(getTripTool, handleGetTrip)

	getTripPlanTool := mcp.NewTool("get_trip_plan",
		mcp.WithDescription("Get the full Wanderlog trip plan by shared trip key, including itinerary, sections, places, and resources"),
		mcp.WithString("trip_key",
			mcp.Description("The shared trip key to retrieve (optional if default trip key is set)"),
		),
		mcp.WithString("trip_id",
			mcp.Description("Alias for trip_key"),
		),
		mcp.WithString("format",
			mcp.Description("Output format (json by default, default for text summary)"),
			mcp.DefaultString("json"),
			mcp.Enum("default", "json"),
		),
	)
	s.AddTool(getTripPlanTool, handleGetTripPlan)

	getItineraryTool := mcp.NewTool("get_itinerary",
		mcp.WithDescription("Get the structured itinerary for a Wanderlog trip by shared trip key"),
		mcp.WithString("trip_key",
			mcp.Description("The shared trip key to retrieve (optional if default trip key is set)"),
		),
		mcp.WithString("trip_id",
			mcp.Description("Alias for trip_key"),
		),
	)
	s.AddTool(getItineraryTool, handleGetItinerary)

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

	getFlightsTool := mcp.NewTool("get_flights",
		mcp.WithDescription("List all flight blocks currently attached to a trip itinerary. Use after add_flight to verify persisted flights."),
		mcp.WithString("trip_id",
			mcp.Description("The trip key to get flights for, not the numeric database ID (optional if default trip key is set)"),
		),
	)
	s.AddTool(getFlightsTool, handleGetFlights)

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
			mcp.WithString("start_time",
				mcp.Description("Visit start time in HH:MM 24-hour format (optional)"),
			),
			mcp.WithString("end_time",
				mcp.Description("Visit end time in HH:MM 24-hour format (optional)"),
			),
		)
		s.AddTool(addPlaceTool, handleAddPlace)

		addFlightTool := mcp.NewTool("add_flight",
			mcp.WithDescription("Add a flight reservation block to the trip's Flights section, creating that section when needed. Requires flight_number and departure_date. Airport info is auto-resolved from flight number."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key to add the flight to (optional if default trip key is set)"),
			),
			mcp.WithString("section_date",
				mcp.Description("Deprecated; flights are always added to the trip's Flights section."),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Deprecated; flights are always added to the trip's Flights section."),
			),
			mcp.WithString("flight_number",
				mcp.Required(),
				mcp.Description("Flight number including airline code (e.g. MU244). The airline code will be extracted automatically."),
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
			mcp.WithString("departure_airport",
				mcp.Description("Departure airport IATA code override (e.g. 'PVG'). Provide when the flight stops API cannot resolve the airport."),
			),
			mcp.WithString("arrival_airport",
				mcp.Description("Arrival airport IATA code override (e.g. 'NRT'). Provide when the flight stops API cannot resolve the airport."),
			),
			mcp.WithBoolean("unscheduled",
				mcp.Description("Deprecated; flights are always added to the trip's Flights section."),
				mcp.DefaultBool(false),
			),
		)
		s.AddTool(addFlightTool, handleAddFlight)

		updateFlightTool := mcp.NewTool("update_flight",
			mcp.WithDescription("Edit an existing flight reservation block using the app's applyOps block edit path."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key (optional if default trip key is set)"),
			),
			mcp.WithNumber("block_id",
				mcp.Required(),
				mcp.Description("Flight block ID from get_flights/get_trip JSON"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID containing the flight block (optional)"),
			),
			mcp.WithString("flight_number", mcp.Description("New flight number with airline code, e.g. MU244")),
			mcp.WithString("departure_date", mcp.Description("New departure date (YYYY-MM-DD)")),
			mcp.WithString("departure_time", mcp.Description("New departure time")),
			mcp.WithString("arrival_date", mcp.Description("New arrival date (YYYY-MM-DD)")),
			mcp.WithString("arrival_time", mcp.Description("New arrival time")),
			mcp.WithString("confirmation_number", mcp.Description("New confirmation number")),
			mcp.WithString("notes", mcp.Description("Replacement notes")),
			mcp.WithString("departure_airport", mcp.Description("Departure airport IATA override")),
			mcp.WithString("arrival_airport", mcp.Description("Arrival airport IATA override")),
		)
		s.AddTool(updateFlightTool, handleUpdateFlight)

		deleteFlightTool := mcp.NewTool("delete_flight",
			mcp.WithDescription("Delete an existing flight reservation block from the itinerary."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key (optional if default trip key is set)"),
			),
			mcp.WithNumber("block_id",
				mcp.Required(),
				mcp.Description("Flight block ID from get_flights/get_trip JSON"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID containing the flight block (optional)"),
			),
		)
		s.AddTool(deleteFlightTool, handleDeleteFlight)

		// Add lodging to trip tool
		addLodgingTool := mcp.NewTool("add_lodging",
			mcp.WithDescription("Add a lodging/hotel reservation to the trip's Hotels and lodging section, creating that section when needed. Accepts CLI snake_case fields and the app assistant's camelCase lodging fields."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key to add the lodging to (optional if default trip key is set)"),
			),
			mcp.WithString("name",
				mcp.Description("Name of the hotel/lodging. Optional when place_id/propertyPlaceId resolves to place details."),
			),
			mcp.WithString("place_id",
				mcp.Description("Google Place ID for the hotel"),
			),
			mcp.WithString("propertyPlaceId",
				mcp.Description("App-compatible alias for place_id"),
			),
			mcp.WithNumber("latitude",
				mcp.Description("Latitude of the hotel (optional - auto-fetched if place_id provided)"),
			),
			mcp.WithNumber("longitude",
				mcp.Description("Longitude of the hotel (optional - auto-fetched if place_id provided)"),
			),
			mcp.WithString("check_in",
				mcp.Description("Check-in date (YYYY-MM-DD)"),
			),
			mcp.WithString("checkInDate",
				mcp.Description("App-compatible alias for check_in"),
			),
			mcp.WithString("check_out",
				mcp.Description("Check-out date (YYYY-MM-DD)"),
			),
			mcp.WithString("checkOutDate",
				mcp.Description("App-compatible alias for check_out"),
			),
			mcp.WithString("confirmation_number",
				mcp.Description("Confirmation number"),
			),
			mcp.WithString("confirmationNumber",
				mcp.Description("App-compatible alias for confirmation_number"),
			),
			mcp.WithArray("traveler_names",
				mcp.Description("Traveler names"),
			),
			mcp.WithArray("travelerNames",
				mcp.Description("App-compatible alias for traveler_names"),
			),
			mcp.WithString("notes",
				mcp.Description("Additional notes"),
			),
			mcp.WithString("note",
				mcp.Description("App-compatible alias for notes"),
			),
		)
		s.AddTool(addLodgingTool, handleAddLodging)

		updateLodgingTool := mcp.NewTool("update_lodging",
			mcp.WithDescription("Edit an existing lodging/hotel reservation block using the app's applyOps block edit path."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key (optional if default trip key is set)"),
			),
			mcp.WithNumber("block_id",
				mcp.Required(),
				mcp.Description("Lodging block ID from add_lodging/get_trip JSON"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID containing the lodging block (optional)"),
			),
			mcp.WithString("name", mcp.Description("New lodging name")),
			mcp.WithString("place_id", mcp.Description("New Google Place ID")),
			mcp.WithString("propertyPlaceId", mcp.Description("App-compatible alias for place_id")),
			mcp.WithNumber("latitude", mcp.Description("New latitude")),
			mcp.WithNumber("longitude", mcp.Description("New longitude")),
			mcp.WithString("check_in", mcp.Description("New check-in date (YYYY-MM-DD)")),
			mcp.WithString("checkInDate", mcp.Description("App-compatible alias for check_in")),
			mcp.WithString("check_out", mcp.Description("New check-out date (YYYY-MM-DD)")),
			mcp.WithString("checkOutDate", mcp.Description("App-compatible alias for check_out")),
			mcp.WithString("confirmation_number", mcp.Description("New confirmation number")),
			mcp.WithString("confirmationNumber", mcp.Description("App-compatible alias for confirmation_number")),
			mcp.WithArray("traveler_names", mcp.Description("Replacement traveler names")),
			mcp.WithArray("travelerNames", mcp.Description("App-compatible alias for traveler_names")),
			mcp.WithString("notes", mcp.Description("Replacement notes")),
			mcp.WithString("note", mcp.Description("App-compatible alias for notes")),
		)
		s.AddTool(updateLodgingTool, handleUpdateLodging)

		deleteLodgingTool := mcp.NewTool("delete_lodging",
			mcp.WithDescription("Delete an existing lodging/hotel reservation block from the itinerary."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key (optional if default trip key is set)"),
			),
			mcp.WithNumber("block_id",
				mcp.Required(),
				mcp.Description("Lodging block ID from add_lodging/get_trip JSON"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID containing the lodging block (optional)"),
			),
		)
		s.AddTool(deleteLodgingTool, handleDeleteLodging)

		// Add train to trip tool
		addTrainTool := mcp.NewTool("add_train",
			mcp.WithDescription("Add a train reservation block to the trip's Transit section, creating that section when needed. Wanderlog has no dedicated train_number field — pass the full designator (e.g. 'SBB EC 317') in carrier or include it in notes."),
			mcp.WithString("trip_key",
				mcp.Description("The trip key to add the train to (optional if default trip key is set)"),
			),
			mcp.WithString("carrier",
				mcp.Required(),
				mcp.Description("Carrier or train designator shown in the banner, e.g. 'SBB', 'DB', 'Trenitalia', or 'SBB EC 317'"),
			),
			mcp.WithString("departure_place_id",
				mcp.Description("Google Place ID for the departure station or city. Required unless departure_name + latitude/longitude are supplied."),
			),
			mcp.WithString("departure_name",
				mcp.Description("Display name for the departure stop (overrides the Google Places name)"),
			),
			mcp.WithNumber("departure_latitude",
				mcp.Description("Departure latitude (auto-fetched if departure_place_id is provided)"),
			),
			mcp.WithNumber("departure_longitude",
				mcp.Description("Departure longitude (auto-fetched if departure_place_id is provided)"),
			),
			mcp.WithString("departure_date",
				mcp.Required(),
				mcp.Description("Departure date (YYYY-MM-DD)"),
			),
			mcp.WithString("departure_time",
				mcp.Description("Departure time (HH:MM, 24-hour)"),
			),
			mcp.WithString("arrival_place_id",
				mcp.Description("Google Place ID for the arrival station or city. Required unless arrival_name + latitude/longitude are supplied."),
			),
			mcp.WithString("arrival_name",
				mcp.Description("Display name for the arrival stop (overrides the Google Places name)"),
			),
			mcp.WithNumber("arrival_latitude",
				mcp.Description("Arrival latitude (auto-fetched if arrival_place_id is provided)"),
			),
			mcp.WithNumber("arrival_longitude",
				mcp.Description("Arrival longitude (auto-fetched if arrival_place_id is provided)"),
			),
			mcp.WithString("arrival_date",
				mcp.Description("Arrival date (YYYY-MM-DD; defaults to departure_date)"),
			),
			mcp.WithString("arrival_time",
				mcp.Description("Arrival time (HH:MM, 24-hour)"),
			),
			mcp.WithString("confirmation_number",
				mcp.Description("Confirmation/booking reference"),
			),
			mcp.WithString("notes",
				mcp.Description("Additional notes (good place to record train number, coach, seat)"),
			),
		)
		s.AddTool(addTrainTool, handleAddTrain)

		// Remove place from trip tool
		removePlaceTool := mcp.NewTool("remove_place",
			mcp.WithDescription("Remove an itinerary place block from a trip. Use block_id from add_place, get_trip, or list_sections; this is not a Google Place ID."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip to remove the place from (optional if default trip ID is set)"),
			),
			mcp.WithString("trip_id",
				mcp.Description("Alias for trip_key"),
			),
			mcp.WithNumber("block_id",
				mcp.Description("Internal Wanderlog place block ID to remove. This is returned by add_place and is not a Google Place ID."),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID containing the place block (optional; resolved automatically when omitted)"),
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

		updatePlaceNotesTool := mcp.NewTool("update_place_notes",
			mcp.WithDescription("Replace the notes text on a place block without rewriting the whole section"),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("section_id",
				mcp.Required(),
				mcp.Description("Section ID containing the place block"),
			),
			mcp.WithNumber("place_id",
				mcp.Required(),
				mcp.Description("Internal Wanderlog place block ID, not a Google Place ID"),
			),
			mcp.WithString("notes",
				mcp.Required(),
				mcp.Description("Replacement notes text"),
			),
		)
		s.AddTool(updatePlaceNotesTool, handleUpdatePlaceNotes)

		updatePlaceVisitTimeTool := mcp.NewTool("update_place_visit_time",
			mcp.WithDescription("Set the visit start and/or end time on an existing itinerary place block"),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("section_id",
				mcp.Required(),
				mcp.Description("Section ID containing the place block"),
			),
			mcp.WithNumber("place_id",
				mcp.Required(),
				mcp.Description("Internal Wanderlog place block ID, not a Google Place ID"),
			),
			mcp.WithString("start_time",
				mcp.Description("Visit start time in HH:MM 24-hour format"),
			),
			mcp.WithString("end_time",
				mcp.Description("Visit end time in HH:MM 24-hour format"),
			),
		)
		s.AddTool(updatePlaceVisitTimeTool, handleUpdatePlaceVisitTime)

		deleteItineraryBlockTool := mcp.NewTool("delete_itinerary_block",
			mcp.WithDescription("Delete any itinerary block by block ID. Use expected_type to guard the delete for notes, checklists, flights, places, lodging, or other reservation blocks."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("block_id",
				mcp.Required(),
				mcp.Description("Itinerary block ID"),
			),
			mcp.WithNumber("section_id",
				mcp.Description("Section ID containing the block (optional)"),
			),
			mcp.WithString("expected_type",
				mcp.Description("Optional guard, e.g. flight, place, note, checklist"),
			),
		)
		s.AddTool(deleteItineraryBlockTool, handleDeleteItineraryBlock)

		setTripBudgetTool := mcp.NewTool("set_trip_budget",
			mcp.WithDescription("Set or update the total trip budget amount and currency."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("amount",
				mcp.Required(),
				mcp.Description("Budget amount"),
			),
			mcp.WithString("currency",
				mcp.Required(),
				mcp.Description("Currency code, e.g. USD, EUR, JPY"),
			),
		)
		s.AddTool(setTripBudgetTool, handleSetTripBudget)

		addTripExpenseTool := mcp.NewTool("add_trip_expense",
			mcp.WithDescription("Add an expense to the trip budget. Use this for ticket costs, reservations, food, transport, lodging, and other spending."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithString("description",
				mcp.Required(),
				mcp.Description("Expense description"),
			),
			mcp.WithString("category",
				mcp.Description("Expense category: flights, lodging, carRental, publicTransit, food, drinks, sightseeing, activities, shopping, gas, groceries, other"),
				mcp.DefaultString("other"),
			),
			mcp.WithNumber("amount",
				mcp.Required(),
				mcp.Description("Expense amount"),
			),
			mcp.WithString("currency",
				mcp.Required(),
				mcp.Description("Currency code, e.g. USD, EUR, JPY"),
			),
			mcp.WithString("date",
				mcp.Description("Expense date (YYYY-MM-DD). Defaults to today."),
			),
			mcp.WithNumber("block_id",
				mcp.Description("Optional itinerary/reservation block ID to link this expense to"),
			),
			mcp.WithNumber("paid_by_user_id",
				mcp.Description("User ID who paid. Defaults to authenticated user."),
			),
			mcp.WithString("split_with_user_ids",
				mcp.Description("Comma-separated user IDs to split with"),
			),
			mcp.WithString("associated_date",
				mcp.Description("Optional trip date to associate with the expense (YYYY-MM-DD)"),
			),
		)
		s.AddTool(addTripExpenseTool, handleAddTripExpense)

		updateTripExpenseTool := mcp.NewTool("update_trip_expense",
			mcp.WithDescription("Update an existing trip budget expense by expense ID."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("expense_id",
				mcp.Required(),
				mcp.Description("Expense ID"),
			),
			mcp.WithString("description", mcp.Description("New expense description")),
			mcp.WithString("category", mcp.Description("New expense category")),
			mcp.WithNumber("amount", mcp.Description("New expense amount")),
			mcp.WithString("currency", mcp.Description("New currency code")),
			mcp.WithString("date", mcp.Description("New expense date (YYYY-MM-DD)")),
			mcp.WithNumber("block_id", mcp.Description("New linked block ID")),
			mcp.WithBoolean("clear_block_id", mcp.Description("Remove linked block ID")),
			mcp.WithNumber("paid_by_user_id", mcp.Description("New user ID who paid")),
			mcp.WithString("split_with_user_ids", mcp.Description("Comma-separated user IDs to split with")),
			mcp.WithString("associated_date", mcp.Description("New associated trip date (YYYY-MM-DD)")),
			mcp.WithBoolean("clear_associated_date", mcp.Description("Remove associated date")),
		)
		s.AddTool(updateTripExpenseTool, handleUpdateTripExpense)

		deleteTripExpenseTool := mcp.NewTool("delete_trip_expense",
			mcp.WithDescription("Delete a trip budget expense by expense ID."),
			mcp.WithString("trip_key",
				mcp.Description("The key/ID of the trip (optional if default trip ID is set)"),
			),
			mcp.WithNumber("expense_id",
				mcp.Required(),
				mcp.Description("Expense ID"),
			),
		)
		s.AddTool(deleteTripExpenseTool, handleDeleteTripExpense)
	}

	// Add search places tool
	searchPlacesTool := mcp.NewTool("search_places",
		mcp.WithDescription("Search for places using Wanderlog autocomplete"),
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

	// Add search restaurants tool
	searchRestaurantsTool := mcp.NewTool("search_restaurants",
		mcp.WithDescription("Search for restaurants using Wanderlog autocomplete. Great for finding ramen shops, sushi places, izakayas, and other food venues."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Restaurant search query (e.g., 'ramen shop Tokyo', 'sushi Kyoto', 'izakaya Osaka')"),
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
	s.AddTool(searchRestaurantsTool, handleSearchRestaurants)

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
		mcp.WithDescription("Get airport stops for a known flight number. Pass the numeric portion of the flight number (e.g. 244 for flight MU244) plus the airline IATA code separately."),
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

	if !readOnly {
		// Trip management tools
		createTripTool := mcp.NewTool("create_trip",
			mcp.WithDescription("Create a new trip plan"),
			mcp.WithString("title", mcp.Required(),
				mcp.Description("Trip title")),
			mcp.WithNumber("geo_id", mcp.Required(),
				mcp.Description("Wanderlog destination geo ID. Use search_geos first to find the correct ID; do not guess this value or use a Google Place ID.")),
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
			mcp.WithDescription("Delete a trip plan. Accepts trip_key or trip_id; if the MCP server was started with --trip-id, the key can be omitted."),
			mcp.WithString("trip_key",
				mcp.Description("Trip key to delete (optional if trip_id or a default trip key is set)")),
			mcp.WithString("trip_id",
				mcp.Description("Alias for trip_key")),
		)
		s.AddTool(deleteTripTool, handleDeleteTrip)

		deleteTripsTool := mcp.NewTool("delete_trips",
			mcp.WithDescription("Delete multiple trip plans in a single request"),
			mcp.WithArray("trip_keys",
				mcp.Required(),
				mcp.Description("Trip keys to delete. A legacy comma-separated string is also accepted."),
				mcp.WithStringItems(mcp.MinLength(1)),
				mcp.MinItems(1)),
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
	}

	if !readOnly {
		likeTripTool := mcp.NewTool("like_trip",
			mcp.WithDescription("Like or unlike a trip plan"),
			mcp.WithString("trip_key", mcp.Required(),
				mcp.Description("Trip key")),
			mcp.WithBoolean("liked", mcp.Required(),
				mcp.Description("true to like, false to unlike")),
		)
		s.AddTool(likeTripTool, handleLikeTrip)
	}

	getLikeCountTool := mcp.NewTool("get_like_count",
		mcp.WithDescription("Get like count and status for a trip"),
		mcp.WithString("trip_key", mcp.Required(),
			mcp.Description("Trip key")),
	)
	s.AddTool(getLikeCountTool, handleGetLikeCount)

	if !readOnly {
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
	}

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

	// Register API extras (geo, directions, recommendations,
	// places/placesAPI, chat assistant, lodging checkout, misc).
	registerReferenceExtras(s, readOnly)

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
