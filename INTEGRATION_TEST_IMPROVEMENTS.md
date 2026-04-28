# Integration Test Improvement Plan for Wanderlog CLI

## Executive Summary

This document outlines a comprehensive plan to improve integration test coverage for the Wanderlog CLI codebase. The current test suite has solid unit test coverage for many components (~26 test files), but integration test coverage is sparse and concentrated in a few areas. This plan categorizes improvements by priority based on criticality and impact.

---

## Priority 1: Critical Functionality (MUST Test)

These features are core to the CLI's value proposition and any failures would directly impact users.

### 1.1 Authentication System

**What needs to be tested:**
- Full login flow with real credentials (currently only mocked)
- Session token persistence and retrieval from keychain
- XSRF token handling
- `EnsureAuthenticated()` fallback chain: keychain -> env vars -> flags
- Logout and session cleanup
- Auth refresh when session expires

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
- Modify `auth_helper.go` to support test injection points

### 1.2 Trip CRUD Operations

**What needs to be tested:**
- **Create Trip**: Full creation with geo_id, dates, privacy settings
- **Get Trip**: Fetching existing trip, handling non-existent trips
- **Update Trip**: Title, dates, privacy mutations (partially covered in `sections_test.go`)
- **Delete Trip**: Deletion and cleanup verification
- **Copy Trip**: Creating trip copies

**Why it's important:**
Trip CRUD is the primary user workflow. Users must be able to create, read, update, and delete trips reliably.

**Current coverage:**
- `write_ops_test.go` has unit tests for CreateTrip, DeleteTrip (mocked)
- `write_ops_integration_test.go` has `TestIntegration_CreateAndDeleteTrip`
- `sections_test.go` has `TestIntegration_UpdateTrip`

**Gaps:**
- No integration test for `GetTrip` with real data
- No integration test for `CopyTrip` with cleanup verification
- No test for bulk `DeleteTrips` (only in `mcp_integration_write_test.go`)

**Suggested approach:**
```go
// TestIntegration_CompleteTripLifecycle
// 1. Create trip with all parameters
// 2. Fetch trip and verify all fields
// 3. Update title, verify
// 4. Update dates, verify
// 5. Update privacy, verify
// 6. Copy trip, verify copy exists
// 7. Delete original, verify deletion
// 8. Delete copy, verify cleanup
```

### 1.3 Place Operations (Add/Remove/Move/Reorder)

**What needs to be tested:**
- **Add Place**: With coordinates, without coordinates, to dated section
- **Remove Place**: With valid/invalid place IDs
- **Move Place**: Between sections with position tracking
- **Reorder Places**: Within a section, verify order persists
- **List Places**: Verify added places appear correctly

**Why it's important:**
Places are the core content of any trip. Users must be able to add, remove, and organize places.

**Current coverage:**
- `write_ops_test.go` has unit tests for AddPlace, RemovePlace (mocked)
- `cmd/mcp_integration_write_test.go` has `TestMCPIntegration_AddPlace`, `TestMCPIntegration_RemovePlace`
- `TestMCPIntegration_UpdatePlaceNotes` and `TestMCPIntegration_ReorderPlaces` test operational transforms directly

**Gaps:**
- No integration test that adds a place, extracts its ID, removes it, and verifies removal
- No integration test for `MovePlace` MCP tool
- No test for `ReorderPlaces` via MCP tool (only via direct ApplyOperations)
- No test for adding place without coordinates (auto-fetch)

**Suggested approach:**
```go
// TestIntegration_AddPlaceExtractAndRemove
// 1. Search for place to get place_id
// 2. Add place to section
// 3. List places, parse response to extract new place ID
// 4. Remove the place
// 5. List places again, verify removal

// TestIntegration_MovePlace
// 1. Add place to section A
// 2. Move place to section B at specific position
// 3. Verify place appears in section B at correct position
// 4. Verify place no longer in section A (or reordered if same section)
```

### 1.4 Section Operations

**What needs to be tested:**
- **List Sections**: Verify section structure and dates
- **Get Section Details**: Block content, metadata
- **Create Section**: If applicable
- **Update Section**: Heading, notes
- **Delete Section**: Cleanup verification

**Why it's important:**
Sections represent days in a trip itinerary. Proper section handling is critical.

**Current coverage:**
- `sections_test.go` has `TestIntegration_GetTripSections` and `TestIntegration_UpdateTrip`

**Gaps:**
- No integration test for section deletion
- No test for section with blocks (places, flights, notes)
- No test for section ordering or reorganization

---

## Priority 2: Important Functionality (SHOULD Test)

These features enhance user experience but aren't strictly required for basic usage.

### 2.1 Search Operations

**What needs to be tested:**
- **search_places**: Query with/without coordinates, verify result structure
- **search_restaurants**: Restaurant-specific search
- **search_geos**: Destination search for trip creation
- **get_place_details**: Detailed place info retrieval
- **search_hotels**: Lodging search with check-in/out dates

**Why it's important:**
Search is the primary discovery mechanism for places to add to trips.

**Current coverage:**
- `search_test.go` has unit tests for SearchPlaces, SearchRestaurants, SearchPlacesWithWanderlog, matchesQuery
- `cmd/mcp_integration_test.go` has `TestMCPIntegration_SearchPlacesWanderlog`
- `mcp_integration_write_test.go` has hotel search within lifecycle test

**Gaps:**
- No integration test for `search_places` with real Google Places API
- No integration test for `search_geos` with real Wanderlog API
- No test for `search_restaurants`
- No test for `get_place_details` with place_id validation

**Suggested approach:**
```go
// TestIntegration_SearchPlacesReal
// 1. Search for "Eiffel Tower Paris"
// 2. Verify results have place_id, name, address
// 3. Verify results can be used in add_place

// TestIntegration_SearchGeosForTripCreation
// 1. Search for "Tokyo"
// 2. Extract geo_id
// 3. Create trip with that geo_id
// 4. Verify trip has correct destination
```

### 2.2 Flight Operations

**What needs to be tested:**
- **GetTripFlights**: Retrieve flights for a trip
- **Add Flight**: Add flight to trip with flight number, dates
- **GetFlightStops**: Query stops for a flight
- **AutocompleteAirport**: Airport search

**Why it's important:**
Flights are a common travel component and users need to track them.

**Current coverage:**
- `flights_test.go` has unit tests for GetAllAirlines, AutocompleteAirport, GetFlightStops
- `write_ops_integration_test.go` has `TestIntegration_GetTripFlights`

**Gaps:**
- No integration test for `add_flight` MCP tool
- No integration test for adding flight, then retrieving trip to verify
- No test for flight with multiple stops

**Suggested approach:**
```go
// TestIntegration_AddFlightAndVerify
// 1. Create new trip
// 2. Add flight MU244 on specific date
// 3. Get trip sections, find flights section
// 4. Verify flight appears in correct section

// TestIntegration_FlightStopsReal
// 1. Query flight stops for MU244
// 2. Verify stops include departure/arrival airports
```

### 2.3 Social Features

**What needs to be tested:**
- **Like/Unlike Trip**: Toggle like status
- **GetLikeCount**: Retrieve like count and user status
- **GetTripDistinction**: Badge/info retrieval

**Why it's important:**
Social features enable trip sharing and discovery.

**Current coverage:**
- `write_ops_integration_test.go` has `TestIntegration_LikeTrip`, `TestIntegration_GetLikeCount`
- `write_ops_social.go` has related operations

**Gaps:**
- No test for like/unlike cycle (like, verify count, unlike, verify count)
- No test for distinction/badge display

### 2.4 Lodging/Hotel Operations

**What needs to be tested:**
- **SearchHotels**: Hotel search with location and dates
- **GetGooglePriceRates**: Price rate retrieval
- **Add lodging to trip as place**

**Why it's important:**
Users frequently need lodging information during trip planning.

**Current coverage:**
- `lodging_test.go` has unit tests for SearchLodgings, GetGooglePriceRates
- `mcp_integration_write_test.go` handles lodging search outage gracefully

**Gaps:**
- No integration test for full hotel search -> add as place workflow
- No test for `GetGooglePriceRates` with real property ID

### 2.5 Feed Operations

**What needs to be tested:**
- **GetFeedHome**: User's home feed
- **GetFeedMostRecent**: Most recently edited trip
- **GetFriendsPlans**: Friends' public trips
- **BrowseGuides**: Guide discovery

**Why it's important:**
Feed operations support trip discovery and collaboration.

**Current coverage:**
- `feed_ops_test.go` has unit tests for all feed operations (mocked)

**Gaps:**
- No integration tests for feed operations with real data
- No test for feed pagination

### 2.6 User Operations

**What needs to be tested:**
- **GetMe**: Current user profile
- **GetUserProfile**: Other user profiles
- **GetNotifications**: Notification inbox
- **MarkNotificationsRead**: Mark notifications as read

**Why it's important:**
User operations support profile management and collaboration.

**Current coverage:**
- `user_ops_test.go` has unit tests (mocked)

**Gaps:**
- No integration tests for user operations
- No test for notification workflow

---

## Priority 3: Nice-to-Have (Edge Cases, Error Scenarios)

These are valuable for robustness but lower priority than P1/P2 items.

### 3.1 Error Handling

**What needs to be tested:**
- Invalid trip key returns appropriate error
- Missing required fields return validation errors
- Network failures are handled gracefully
- Rate limiting is respected (with delays)
- API server errors (5xx) are handled

**Suggested approach:**
```go
// TestErrorHandling
func TestIntegration_InvalidTripKey(t *testing.T) {
    // Attempt operations with obviously invalid keys
    // Verify appropriate error messages
}

func TestIntegration_MissingRequiredFields(t *testing.T) {
    // Create trip without title
    // Create trip without dates
    // Verify validation errors
}
```

### 3.2 Edge Cases

**What needs to be tested:**
- Empty trip (no places, no sections)
- Trip with very long title
- Unicode in place names and notes
- Concurrent operations on same trip
- Large number of places in single section

### 3.3 MCP Server Specific

**What needs to be tested:**
- HTTP transport mode (not just stdio)
- Concurrent tool calls
- Resource URI variations
- Prompt variations with different focus areas

### 3.4 Operational Transform Edge Cases

**What needs to be tested:**
- Nested path operations
- Array index operations
- Partial updates vs full replacements
- Conflict resolution scenarios

---

## Test Infrastructure Recommendations

### A. Test Fixtures

Create reusable test fixtures for:
- Authenticated client setup
- Common trip templates
- Place data for popular destinations (Paris, Tokyo, NYC)

```go
// pkg/wanderlog/testfixtures/fixtures.go
var (
    AuthenticatedClient *Client // Setup via test helper
    ParisTripTemplate   CreateTripRequest
    TestPlaces          map[string]PlaceData
)
```

### B. Test Helpers

Create helper functions to reduce boilerplate:

```go
// RequiresRealAuth skips tests that need real credentials
func requiresRealAuth(t *testing.T) {
    if os.Getenv("INTEGRATION_TESTS") != "1" {
        t.Skip("Set INTEGRATION_TESTS=1 to run")
    }
}

// authenticatedClient returns a client with real auth loaded
func authenticatedClient(t *testing.T) *Client {
    client := NewClient()
    if err := client.EnsureAuthenticated("", ""); err != nil {
        t.Skipf("Auth required: %v", err)
    }
    return client
}
```

### C. CI Integration

Ensure tests can run in CI with proper secrets management:

```bash
# .github/workflows/integration-tests.yml
env:
  WANDERLOG_AUTH_SESSION_COOKIE: ${{ secrets.WANDERLOG_SESSION_COOKIE }}
  WANDERLOG_AUTH_SESSION_XSRF_TOKEN: ${{ secrets.WANDERLOG_XSRF_TOKEN }}
```

### D. Test Data Management

- Use snapshot tests to capture API responses for debugging
- Implement `SAVE_SNAPSHOTS=1` for local development
- Keep test trips isolated and clean them up in `defer`

---

## Testing Matrix

| Feature Area | Unit Tests | Integration Tests | MCP Tests | Priority |
|-------------|------------|-------------------|-----------|----------|
| Auth (Login) | Yes | **GAP** | N/A | P1 |
| Auth (Session) | Partial | **GAP** | N/A | P1 |
| CreateTrip | Yes | Yes | Yes | P1 |
| GetTrip | Yes | **GAP** | Yes | P1 |
| UpdateTrip | Yes | Yes | **GAP** | P1 |
| DeleteTrip | Yes | Yes | Yes | P1 |
| CopyTrip | Yes | Yes | Yes | P1 |
| ListPlaces | Partial | Partial | Yes | P1 |
| AddPlace | Yes | Partial | Yes | P1 |
| RemovePlace | Yes | **GAP** | Partial | P1 |
| MovePlace | N/A | **GAP** | **GAP** | P1 |
| ReorderPlaces | Yes | **GAP** | **GAP** | P1 |
| ListSections | Yes | Yes | Yes | P1 |
| SearchPlaces | Yes | **GAP** | Yes | P2 |
| SearchRestaurants | Yes | **GAP** | **GAP** | P2 |
| SearchGeos | N/A | **GAP** | **GAP** | P2 |
| GetPlaceDetails | Yes | **GAP** | Yes | P2 |
| SearchHotels | Yes | **GAP** | Yes | P2 |
| GetTripFlights | Yes | Yes | **GAP** | P2 |
| AddFlight | N/A | **GAP** | Yes | P2 |
| LikeTrip | Yes | Yes | **GAP** | P2 |
| GetLikeCount | Yes | Yes | **GAP** | P2 |
| Feed (Home/Recent/Friends) | Yes | **GAP** | **GAP** | P2 |
| User Profile | Yes | **GAP** | **GAP** | P2 |
| Notifications | Yes | **GAP** | **GAP** | P2 |
| Journal | Yes | **GAP** | **GAP** | P3 |
| Operational Transforms | Yes | Yes | Yes | P1 |

---

## Implementation Roadmap

### Phase 1: Critical Gaps (Week 1-2)
1. Create `auth_integration_test.go` for real auth flow
2. Expand `TestIntegration_CompleteTripLifecycle` to cover all CRUD operations
3. Add integration test for AddPlace -> extract ID -> RemovePlace workflow
4. Add integration test for MovePlace
5. Add integration test for ReorderPlaces via MCP tool

### Phase 2: Important Gaps (Week 3-4)
1. Create `search_integration_test.go` for real search operations
2. Add flight integration tests (add flight, verify, get stops)
3. Add social features integration tests (like/unlike cycle)
4. Add feed operations integration tests
5. Add user operations integration tests

### Phase 3: Polish (Week 5+)
1. Add comprehensive error handling tests
2. Add edge case tests
3. Implement test fixtures and helpers
4. Document testing patterns
5. Add performance/load tests if needed

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
