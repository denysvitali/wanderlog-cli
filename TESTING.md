# Testing Guide

## Integration Tests

This project includes comprehensive integration tests that interact with the real Wanderlog API.

### Prerequisites

1. **Authentication Required**: Run `./wanderlog auth login` before running integration tests
2. **Network Access**: Tests make real API calls to wanderlog.com
3. **Rate Limiting**: Some tests add delays to respect API rate limits

### Running Tests

```bash
# Run all integration tests
go test -v -tags=integration -timeout 30m ./pkg/wanderlog

# Run specific test
go test -v -tags=integration -run TestSnapshotTrip -timeout 10m ./pkg/wanderlog
go test -v -tags=integration -run TestBeijingTripCreation -timeout 30m ./pkg/wanderlog

# Run with snapshot saving enabled
SAVE_SNAPSHOTS=1 go test -v -tags=integration -run TestSnapshotTrip ./pkg/wanderlog
```

### Available Integration Tests

#### `TestSnapshotTrip` - Trip Modification Snapshots
Creates a Paris trip and captures server snapshots after each operation:
1. Initial trip creation
2. Add first place (Eiffel Tower)
3. Add second place (Louvre Museum)
4. Update trip title
5. Remove first place

Shows unified diffs between each step to demonstrate how the server state changes.

**Usage:**
```bash
# Run with snapshot saving
SAVE_SNAPSHOTS=1 go test -v -tags=integration -run TestSnapshotTrip ./pkg/wanderlog

# Snapshots saved to /tmp/snapshot_*.json
ls /tmp/snapshot_*.json
```

#### `TestBeijingTripCreation` - Real-World Trip Planning
Creates a complete week-long Beijing itinerary using the search API:
- Searches for 15+ attractions dynamically
- Adds places across 7 days with detailed notes
- Captures snapshots after each place addition
- Shows diffs to track trip evolution

**Usage:**
```bash
# Run with verbose output
go test -v -tags=integration -run TestBeijingTripCreation -timeout 30m ./pkg/wanderlog

# Save all snapshots for inspection
SAVE_SNAPSHOTS=1 go test -v -tags=integration -run TestBeijingTripCreation -timeout 30m ./pkg/wanderlog

# Snapshots saved to /tmp/beijing_snapshot_*.json
```

#### `TestIntegration_*` - Individual Operations
Tests for specific operations:
- `TestIntegration_CreateAndDeleteTrip`
- `TestIntegration_CopyTrip`
- `TestIntegration_LikeTrip`
- `TestIntegration_GetLikeCount`

### Snapshot Features

The snapshot tests capture the raw, unprocessed JSON response from the server after each modification. This is useful for:

1. **Debugging**: See exactly what the server returns
2. **Verification**: Confirm operations work as expected
3. **Documentation**: Understand the API response structure
4. **Regression Testing**: Detect unexpected API changes

**Unified Diffs**: Tests display unified diffs showing:
- Lines added (prefixed with `+`)
- Lines removed (prefixed with `-`)
- Context lines (unchanged)

### Unit Tests

Run unit tests without the integration tag:

```bash
# Run all unit tests
go test -v ./pkg/wanderlog

# Run specific unit test
go test -v -run TestCreateTrip ./pkg/wanderlog
```

### API Request Contract Tests

`TestAPIRequestContracts` captures Go client requests with a local test server
and compares them against `artifacts/api-contracts/go_request_contracts.json`.
It also checks that every contract maps back to an endpoint extracted from
`artifacts/decompiled/wanderlog_decompiled.js`.

```bash
make generate
go test -v -run TestAPIRequestContracts ./pkg/wanderlog
```

### Test Organization

```
pkg/wanderlog/
├── *_test.go              # Unit tests (no integration tag)
├── *_integration_test.go  # Integration tests (require auth)
├── snapshot_test.go       # Snapshot-based testing
└── beijing_trip_test.go   # Real-world trip creation
```

### Best Practices

1. **Clean Up**: Tests clean up created trips automatically
2. **Rate Limits**: Tests include delays between API calls
3. **Error Handling**: Tests continue on non-fatal errors (e.g., search failures)
4. **Logging**: Use `-v` flag for detailed test output
5. **Timeouts**: Long-running tests have explicit timeouts

### Troubleshooting

**Authentication Errors:**
```
Failed to authenticate: authentication required
```
**Solution:** Run `./wanderlog auth login` first

**Timeout Errors:**
```
panic: test timed out after 2m0s
```
**Solution:** Increase timeout with `-timeout 30m`

**Search Failures:**
```
⚠️  No results for 'Place Name'
```
**Solution:** This is expected - tests skip places that can't be found

### Environment Variables

- `SAVE_SNAPSHOTS=1`: Save JSON snapshots to /tmp for inspection
- `WANDERLOG_SESSION_COOKIE`: Override session cookie (not recommended)
- `WANDERLOG_XSRF_TOKEN`: Override XSRF token (not recommended)

Use the CLI's built-in authentication instead of environment variables.
