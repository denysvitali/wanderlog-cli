# Integration Test Improvement Plan for Wanderlog CLI

## Executive Summary

This document outlines a comprehensive plan to improve integration test coverage for the Wanderlog CLI codebase. The current test suite has solid unit test coverage for many components (~26 test files), but integration test coverage is sparse and concentrated in a few areas. Analysis from 8 specialized agents reveals significant gaps across authentication, trip operations, MCP tools, models, and operational transforms.

**Key Statistics:**
- 26 test files in the project
- ~14 tools fully tested at MCP handler level
- 17+ MCP tools completely untested
- 37+ model types lack JSON round-trip tests
- 9 OT helper functions lack individual unit tests

---

## Priority 1: Critical Functionality (MUST Test)

These features are core to the CLI's value proposition and any failures would directly impact users.

### 1.1 Authentication System

**Agent: agent-auth**

**Current Coverage:**
- `TestAddAuthHeaders` - session cookie and XSRF token headers
- `TestLogin` - success/failure/missing cookie scenarios (mocked)
- `TestSetAuth` - credential storage

**Critical Gaps:**
- Keychain operations (save/load/delete credentials)
- `EnsureAuthenticated()` fallback chain: keychain -> env vars -> flags
- Token refresh and re-login flow
- Session expiration handling
- Invalid credential handling

**Why it's important:**
Authentication is the foundation for all write operations. If auth fails, users cannot manage trips.

**Suggested approach:**
```bash
# Integration test with real credentials (CI-controlled)
INTEGRATION_TESTS=1 WANDERLOG_AUTH_EMAIL=test@example.com WANDERLOG_AUTH_PASSWORD=secret \
  go test -v -run "TestAuth.*Integration" ./pkg/wanderlog
```

**Files to create/modify:**
- `pkg/wanderlog/auth_integration_test.go` - Real auth flow tests
- Mock keychain for CI environments

**Test scenarios to add:**
1. Test `EnsureAuthenticated` with keychain credentials present
2. Test `EnsureAuthenticated` falls back to env vars when keychain is empty
3. Test `EnsureAuthenticated` uses flags when env vars are also missing
4. Test login with invalid credentials returns appropriate error
5. Test session persistence across client instances

---

### 1.2 Trip Operations (Handler Level)

**Agent: agent-trip-ops**

**Current Coverage:**
- `copy_trip` at client level (mocked)
- `delete_trips` MCP handler
- `get_like_count` at client level (mocked)

**Critical Gaps:**
- `get_trip_sections` MCP handler (no handler-level test)
- `copy_trip` MCP handler (no handler-level test)
- Bulk delete with multiple keys (partially covered)
- `get_like_count` MCP handler (no handler-level test)

**Why it's important:**
Trip operations are the primary user workflow. Users must be able to view, copy, and delete trips.

**Suggested approach:**
```go
// TestMCP_GetTripSections
// TestMCP_CopyTrip
// TestMCP_GetLikeCount (at handler level)
// TestMCP_DeleteTrips_MultipleKeys
```

**Files to create/modify:**
- `cmd/mcp_trip_ops_integration_test.go`

---

### 1.3 Place Operations (Complete Workflow)

**Current Coverage:**
- Unit tests for AddPlace, RemovePlace with mocks
- MCP integration tests for AddPlace (adds to test trip, doesn't clean up)
- MCP integration test for RemovePlace (error cases only)

**Critical Gaps:**
- No integration test that adds a place, extracts its ID, removes it, and verifies removal
- No test for MovePlace MCP tool (both client and handler)
- No test for ReorderPlaces MCP tool (only via direct ApplyOperations)
- No test for adding place without coordinates (auto-fetch via place_id)

**Why it's important:**
Places are the core content of any trip. Users must be able to add, remove, move, and organize places.

**Suggested approach:**
```go
// TestIntegration_AddPlaceExtractAndRemove
// 1. Search for place to get place_id
// 2. Add place to section
// 3. List places, parse response to extract new place ID
// 4. Remove the place
// 5. List places again, verify removal

// TestMCP_MovePlace
// 1. Get section IDs for a trip
// 2. Add place to section A
// 3. Call move_place to move to section B at position 0
// 4. Verify place appears in section B at correct position
// 5. Verify place no longer in section A (or reordered if same section)

// TestMCP_ReorderPlaces
// 1. Get section with 3+ places
// 2. Call reorder_places with reversed order
// 3. Verify order persisted
```

---

### 1.4 Operational Transforms (Complete Coverage)

**Agent: agent-operational-transforms**

**Current Coverage:**
- `ApplyOperations` unit tests (mocked)
- `ReplaceInObject` via UpdatePlaceNotes and ReorderPlaces

**Critical Gaps:**
- `InsertInObject` - no direct unit test
- `DeleteInObject` - no direct unit test
- `InsertInList` - no direct unit test
- `DeleteFromList` - no direct unit test
- `ReplaceInList` - no direct unit test
- `MovePlace` OT - no direct unit test
- `ClearSectionBlocks` - no direct unit test
- `DeleteSection` - no direct unit test
- `NukeTripPlaces` - no direct unit test
- Edge cases: empty paths, concurrent modifications, path validation

**Why it's important:**
Operational transforms are the foundation of all write operations. Without comprehensive OT tests, regressions can occur silently.

**Suggested approach:**
```go
// TestOT_InsertInObject
// TestOT_DeleteInObject
// TestOT_InsertInList
// TestOT_DeleteFromList
// TestOT_ReplaceInList
// TestOT_MovePlace
// TestOT_ClearSectionBlocks
// TestOT_DeleteSection
// TestOT_NukeTripPlaces
// TestOT_EdgeCase_EmptyPath
// TestOT_EdgeCase_ConcurrentModifications
```

**Files to create:**
- `pkg/wanderlog/ot_integration_test.go` (or expand `write_ops_test.go`)

---

## Priority 2: Important Functionality (SHOULD Test)

### 2.1 MCP Extended Tools (17+ Untested)

**Agent: agent-mcp-server**

**Current Coverage:**
- Core read tools: list_trips, get_trip, list_places, list_sections, get_place_details, search_places_wanderlog
- Write tools: add_place, remove_place, create_trip, delete_trips, add_flight
- Lifecycle test: complete trip creation with places, lodging, flights

**Critical Gaps (17+ tools untested):**
| Tool | Status | Priority |
|------|--------|----------|
| get_me | NOT TESTED | P2 |
| get_user_profile | NOT TESTED | P2 |
| get_notifications | NOT TESTED | P2 |
| get_notification_settings | NOT TESTED | P3 |
| mark_notifications_read | NOT TESTED | P3 |
| get_user_emails | NOT TESTED | P3 |
| autocomplete_users | NOT TESTED | P3 |
| is_username_taken | NOT TESTED | P3 |
| get_feed_home | NOT TESTED | P2 |
| get_feed_recent | NOT TESTED | P2 |
| get_feed_friends | NOT TESTED | P2 |
| get_trip_history | NOT TESTED | P2 |
| browse_guides | NOT TESTED | P2 |
| search_geos | NOT TESTED | P2 |
| get_view_only_journal | NOT TESTED | P3 |
| get_trip_expenses_csv | NOT TESTED | P3 |
| get_trip_distinction | NOT TESTED | P3 |
| get_global_config | NOT TESTED | P3 |
| HTTP transport mode | NOT TESTED | P2 |
| CLI flag combinations | NOT TESTED | P2 |

**Why it's important:**
The MCP server exposes 30+ tools. Users and AI assistants rely on these tools for trip management.

**Suggested approach:**
```go
// cmd/mcp_extended_integration_test.go
// Separate file for extended tool tests

// Priority P2 tests:
TestMCP_GetMe
TestMCP_GetUserProfile
TestMCP_GetNotifications
TestMCP_GetFeedHome
TestMCP_GetFeedRecent
TestMCP_GetFeedFriends
TestMCP_GetTripHistory
TestMCP_BrowseGuides
TestMCP_SearchGeos
TestMCP_HTTPTransport
TestMCP_CLIFlagCombinations
```

---

### 2.2 Travel Search Tools (4 Untested)

**Agent: agent-travel-search**

**Current Coverage:**
- `search_hotels` MCP handler
- `search_places_wanderlog` MCP handler

**Critical Gaps:**
- `search_places` MCP handler (Google Places API)
- `search_restaurants` MCP handler
- `get_trip_flights` MCP handler (not the client method)
- `get_flight_stops` MCP handler

**Why it's important:**
Travel search is the primary discovery mechanism for places, lodging, and flights.

**Suggested approach:**
```go
// TestMCP_SearchPlaces
// TestMCP_SearchRestaurants
// TestMCP_GetTripFlights
// TestMCP_GetFlightStops
```

---

### 2.3 CLI Command Behavioral Tests

**Agent: agent-cli-commands**

**Current Coverage:**
- Command registration tests only

**Critical Gaps:**
- All behavioral tests for:
  - `trip` command (view trip details)
  - `places` command (list places)
  - `create` command (create new trip)
  - `edit` subcommands (update-trip, add-place, remove-place, etc.)
  - `travel` subcommands (airlines, airports, flight-stops, hotels)
  - `user` subcommands (profile, notifications, settings)
  - `feed` subcommands (home, recent, friends, history)
  - `journal` command
  - `expenses` command
- Argument parsing tests
- Flag handling tests
- Output formatting tests (pretty vs JSON vs markdown)

**Why it's important:**
Users interact with the CLI via commands. Regression in command behavior breaks user workflows.

**Suggested approach:**
```go
// cmd/cliBehavioral_test.go
TestCLI_TripCommand
TestCLI_CreateCommand
TestCLI_EditCommands
TestCLI_TravelCommands
TestCLI_UserCommands
TestCLI_FeedCommands
TestCLI_FlagCombinations
TestCLI_OutputFormats
```

---

### 2.4 User/Feed/Journal Tools (17 Untested)

**Agent: agent-user-feed-journal**

**Current Coverage:**
- 14 tools partially tested at unit level with mocks

**Critical Gaps (17 tools need MCP handler tests):**
1. `get_me` - Current user profile
2. `get_user_profile` - Other user profiles
3. `get_notifications` - Notification inbox
4. `get_notification_settings` - User notification preferences
5. `get_user_emails` - Registered email addresses
6. `autocomplete_users` - User search by name
7. `is_username_taken` - Username availability check
8. `get_feed_home` - Home feed (own + friends' trips)
9. `get_feed_recent` - Most recently edited trip
10. `get_feed_friends` - Friends' public trips
11. `get_trip_history` - Trip edit history with pagination
12. `browse_guides` - Curated guide discovery
13. `search_geos` - Destination search for trip creation
14. `get_view_only_journal` - Shared journal view
15. `get_trip_expenses_csv` - Expense export
16. `get_trip_distinction` - Trip badge/info
17. `get_global_config` - Server configuration

**Why it's important:**
These tools support profile management, social features, and trip discovery.

**Suggested approach:**
```go
// cmd/mcp_user_feed_journal_test.go
TestMCP_GetMe
TestMCP_GetUserProfile
TestMCP_GetNotifications
TestMCP_GetFeedHome
TestMCP_GetFeedRecent
TestMCP_GetFeedFriends
TestMCP_GetTripHistory
TestMCP_BrowseGuides
TestMCP_SearchGeos
// ... etc.
```

---

### 2.5 Section Operations

**Current Coverage:**
- `TestIntegration_GetTripSections`
- `TestIntegration_UpdateTrip`

**Critical Gaps:**
- Section deletion
- Section with blocks (places, flights, notes)
- Section reordering/organization
- Block-level operations (add block, remove block, update block)

---

## Priority 3: Nice-to-Have (Edge Cases, Model Tests)

### 3.1 Model JSON Round-Trip Tests

**Agent: agent-models**

**Current Coverage:**
- TripResponse JSON parsing only

**Critical Gaps (37+ models untested):**
- FlexibleText marshaling/unmarshaling
- All response wrappers (success/error structures)
- Travel models (Flight, Airline, Airport, Lodging)
- User models (Profile, Notifications, Settings)
- Feed models (TripPlan, Guide)
- Block models (PlaceBlock, FlightBlock, NoteBlock, etc.)

**Why it's important:**
Models are the foundation of all API communication. Malformed JSON can cause silent failures.

**Suggested approach:**
```go
// pkg/wanderlog/models_test.go
// Add JSON round-trip tests for each model type

TestFlexibleText_MarshalUnmarshal
TestTripPlan_MarshalUnmarshal
TestPlaceBlock_MarshalUnmarshal
TestFlightBlock_MarshalUnmarshal
TestLodgingSearchResponse_MarshalUnmarshal
// ... etc.

// Use table-driven tests with JSON fixtures
```

**Files to create:**
- `pkg/wanderlog/testdata/*.json` - JSON fixtures for all model types

---

### 3.2 OT Edge Cases

**Current Coverage:**
- Basic ApplyOperations with mocked server

**Critical Gaps:**
- Nested path operations (e.g., `["itinerary", "sections", 0, "blocks", 0, "text"]`)
- Array index boundaries (negative, out-of-bounds)
- Partial updates vs full replacements
- Conflict resolution scenarios
- Invalid path validation

---

### 3.3 Error Handling Integration Tests

**What needs to be tested:**
- Invalid trip key returns appropriate error
- Missing required fields return validation errors
- Network failures are handled gracefully
- Rate limiting is respected (with delays)
- API server errors (5xx) are handled
- Malformed JSON responses

---

### 3.4 MCP Server Edge Cases

**What needs to be tested:**
- HTTP transport mode (not just stdio)
- Concurrent tool calls
- Resource URI variations (invalid URIs)
- Prompt variations with different focus areas
- Empty results handling
- Large result pagination

---

### 3.5 CLI Command Edge Cases

**What needs to be tested:**
- Empty trip (no places, no sections)
- Trip with very long title
- Unicode in place names and notes
- Large number of places in single section
- Concurrent operations on same trip

---

## Test Infrastructure Recommendations

### A. Test Fixtures

Create reusable test fixtures for:
- Authenticated client setup
- Common trip templates
- Place data for popular destinations (Paris, Tokyo, NYC)
- JSON fixtures for all model types

```go
// pkg/wanderlog/testfixtures/fixtures.go
var (
    AuthenticatedClient *Client
    ParisTripTemplate   CreateTripRequest
    TestPlaces          map[string]PlaceData
)

// pkg/wanderlog/testdata/*.json
```

### B. Test Helpers

```go
// SkipIfNoAuth skips tests that need real credentials
func skipIfNoAuth(t *testing.T) {
    if os.Getenv("INTEGRATION_TESTS") != "1" {
        t.Skip("Set INTEGRATION_TESTS=1 to run")
    }
    if os.Getenv("CI") == "true" && !hasAuthEnv() {
        t.Skip("CI without auth secrets")
    }
}

// authenticatedClient returns a client with real auth
func authenticatedClient(t *testing.T) *Client {
    client := NewClient()
    if err := client.EnsureAuthenticated("", ""); err != nil {
        t.Skipf("Auth required: %v", err)
    }
    return client
}
```

### C. CI Integration

```bash
# .github/workflows/integration-tests.yml
env:
  WANDERLOG_AUTH_SESSION_COOKIE: ${{ secrets.WANDERLOG_SESSION_COOKIE }}
  WANDERLOG_AUTH_SESSION_XSRF_TOKEN: ${{ secrets.WANDERLOG_XSRF_TOKEN }}
```

---

## Consolidated Testing Matrix

### Priority 1 (Critical)

| Feature | Unit Tests | Integration Tests | MCP Tests | Agent |
|---------|------------|-------------------|-----------|-------|
| Auth (Login) | Yes | **GAP** | N/A | auth |
| Auth (EnsureAuthenticated) | Partial | **GAP** | N/A | auth |
| Auth (Keychain) | **GAP** | **GAP** | N/A | auth |
| CopyTrip (client) | Yes | Yes | Yes | trip-ops |
| CopyTrip (handler) | **GAP** | **GAP** | **GAP** | trip-ops |
| GetTripSections (handler) | **GAP** | **GAP** | **GAP** | trip-ops |
| AddPlace workflow | Yes | Partial | Yes | - |
| RemovePlace (complete) | Yes | **GAP** | Partial | - |
| MovePlace | **GAP** | **GAP** | **GAP** | - |
| ReorderPlaces (MCP) | Yes | **GAP** | **GAP** | - |
| OT: InsertInObject | **GAP** | **GAP** | N/A | OT |
| OT: DeleteInObject | **GAP** | **GAP** | N/A | OT |
| OT: InsertInList | **GAP** | **GAP** | N/A | OT |
| OT: DeleteFromList | **GAP** | **GAP** | N/A | OT |
| OT: ReplaceInList | **GAP** | **GAP** | N/A | OT |
| OT: MovePlace | **GAP** | **GAP** | N/A | OT |
| OT: ClearSectionBlocks | **GAP** | **GAP** | N/A | OT |
| OT: DeleteSection | **GAP** | **GAP** | N/A | OT |
| OT: NukeTripPlaces | **GAP** | **GAP** | N/A | OT |

### Priority 2 (Important)

| Feature | MCP Handler Tests | Agent |
|---------|------------------|-------|
| get_me | **GAP** | user-feed-journal |
| get_user_profile | **GAP** | user-feed-journal |
| get_notifications | **GAP** | user-feed-journal |
| get_feed_home | **GAP** | user-feed-journal |
| get_feed_recent | **GAP** | user-feed-journal |
| get_feed_friends | **GAP** | user-feed-journal |
| get_trip_history | **GAP** | user-feed-journal |
| browse_guides | **GAP** | user-feed-journal |
| search_geos | **GAP** | user-feed-journal |
| search_places | **GAP** | travel-search |
| search_restaurants | **GAP** | travel-search |
| get_trip_flights (handler) | **GAP** | travel-search |
| get_flight_stops (handler) | **GAP** | travel-search |
| HTTP transport | **GAP** | mcp-server |
| CLI flag combos | **GAP** | cli-commands |
| All CLI behavioral | **GAP** | cli-commands |

### Priority 3 (Nice-to-Have)

| Feature | Gap | Agent |
|---------|-----|-------|
| FlexibleText JSON | **GAP** | models |
| 37+ models JSON | **GAP** | models |
| Error handling edge cases | **GAP** | - |
| MCP HTTP transport | **GAP** | mcp-server |
| Concurrent tool calls | **GAP** | mcp-server |
| Unicode/large payloads | **GAP** | cli-commands |

---

## Implementation Roadmap

### Phase 1: Critical Gaps (Week 1-2)
1. Create `auth_integration_test.go` for real auth flow + keychain
2. Add `EnsureAuthenticated` table-driven tests
3. Add OT helper function tests (InsertInObject, DeleteInObject, InsertInList, DeleteFromList, etc.)
4. Add `TestMCP_GetTripSections` handler test
5. Add `TestMCP_CopyTrip` handler test
6. Add `TestMCP_MovePlace` integration test
7. Add complete AddPlace -> RemovePlace workflow test

### Phase 2: Important Gaps (Week 3-4)
1. Add all 17+ MCP extended tool tests
2. Add 4 travel search MCP handler tests
3. Add CLI behavioral tests for core commands
4. Add feed operations integration tests
5. Add user operations integration tests
6. Add HTTP transport test

### Phase 3: Nice-to-Have (Week 5+)
1. Add JSON round-trip tests for 37+ models
2. Add comprehensive error handling tests
3. Add edge case tests (Unicode, large payloads)
4. Implement test fixtures and helpers
5. Document testing patterns

---

## Appendix: Running Tests

```bash
# Run all unit tests (fast, no network required)
go test -v ./...

# Run integration tests with production API
INTEGRATION_TESTS=1 go test -v -tags=integration ./pkg/wanderlog

# Run MCP integration tests
INTEGRATION_TESTS=1 go test -v -tags=integration ./cmd

# Run specific integration test
INTEGRATION_TESTS=1 go test -v -tags=integration -run TestIntegration_CreateAndDeleteTrip ./pkg/wanderlog

# Run with snapshots
SAVE_SNAPSHOTS=1 go test -v -tags=integration ./pkg/wanderlog

# Run with real credentials in CI
WANDERLOG_AUTH_SESSION_COOKIE=xxx WANDERLOG_AUTH_SESSION_XSRF_TOKEN=yyy \
  go test -v -tags=integration ./...
```
