# Wanderlog API Endpoints Implementation

This document lists all trip editing endpoints from the Wanderlog web application and their implementation status in the CLI.

## Implemented Trip Editing Endpoints

### Core Trip Operations

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans` | POST | ✅ `CreateTrip()` | `write_ops.go:141` | `TestIntegration_CreateAndDeleteTrip` |
| `/api/tripPlans/:key` | DELETE | ✅ `DeleteTrip()` | `write_ops.go:194` | `TestIntegration_CreateAndDeleteTrip` |
| `/api/tripPlans/:key/restore` | POST | ✅ `RestoreTrip()` | `write_ops.go:638` | `TestIntegration_RestoreTrip` |
| `/api/tripPlans/copy/:key` | POST | ✅ `CopyTrip()` | `write_ops.go:590` | `TestIntegration_CopyTrip` |

### Place Management

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/sections/:sectionId/place` | POST | ✅ `AddPlace()` | `write_ops.go:249` | `TestIntegration_AddAndRemovePlace` |
| `/api/tripPlans/:key/sections/:sectionId/place/:placeId` | DELETE | ✅ `RemovePlace()` | `write_ops.go:349` | `TestIntegration_AddAndRemovePlace` |
| `/api/tripPlans/:key/applyOps` | POST | ✅ `ApplyOperations()` | `write_ops.go:397` | `TestIntegration_ApplyOperations` |

### Advanced Operations

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `NukeTripPlaces` (custom) | - | ✅ `NukeTripPlaces()` | `write_ops.go:539` | `TestIntegration_NukeTripPlaces` |
| `ClearSectionBlocks` (custom) | - | ✅ `ClearSectionBlocks()` | `write_ops.go:486` | - |
| `DeleteSection` (custom) | - | ✅ `DeleteSection()` | `write_ops.go:512` | - |

### Images & Media

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/images` | GET | ✅ `GetTripImages()` | `visualization.go:135` | `TestIntegration_GetTripImages` |
| `/api/tripPlans/:key/image` | POST | ❌ Not Implemented | - | - |
| `/api/tripPlans/:key/attachment` | POST | ❌ Not Implemented | - | - |

### Collaboration

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/invite` | POST | ✅ `SendTripInvites()` | `write_ops.go:713` | - |
| `/api/tripPlans/:key/invites` | GET | ✅ `ListTripInvites()` | `write_ops.go:763` | `TestIntegration_ListTripInvites` |
| `/api/tripPlans/:key/collaborator` | POST | ✅ `AddCollaborator()` | `write_ops.go:883` | - |
| `/api/tripPlans/:key/collaborator` | DELETE | ✅ `RemoveCollaborator()` | `write_ops.go:930` | - |
| `/api/tripPlans/:editKey/shareKey` | POST | ✅ `GetOrCreateShareKey()` | `write_ops.go:988` | - |

### Social Features

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/like` | POST | ✅ `SetLike()` | `write_ops.go:799` | `TestIntegration_SetLike` |
| `/api/tripPlans/:key/likeCount` | GET | ✅ `GetLikeCount()` | `write_ops.go:848` | `TestIntegration_GetLikeCount` |

### Advanced Features (Not Implemented)

These endpoints are available in the web app but not yet implemented in the CLI:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/tripPlans/:key/export/v2` | POST | Export to Google Maps |
| `/api/tripPlans/:key/createGuideFromTripPlan` | POST | Convert trip to guide |
| `/api/tripPlans/:key/updateTripPlanGeo/:geoId` | POST | Update trip geography |
| `/api/tripPlans/:key/distinction` | GET/POST | Get/set trip distinction |
| `/api/tripPlans/:key/registerView` | POST | Register trip view |
| `/api/tripPlans/autofillDay` | POST | Auto-fill day with suggestions |
| `/api/tripPlans/checklistSection` | POST | Manage checklist sections |

## Data Models

### Request Types

```go
// Create/update trip
type CreateTripRequest struct {
    Title     string `json:"title"`
    StartDate string `json:"startDate,omitempty"` // YYYY-MM-DD
    EndDate   string `json:"endDate,omitempty"`
    Privacy   string `json:"privacy,omitempty"` // "public", "private", "unlisted"
}

// Add place to trip
type AddPlaceRequest struct {
    Place AddPlaceInfo `json:"place"`
    Text  string       `json:"text"`
}

type AddPlaceInfo struct {
    PlaceID  string `json:"place_id,omitempty"`
    Name     string `json:"name"`
    Geometry *struct {
        Location struct {
            Lat float64 `json:"lat"`
            Lng float64 `json:"lng"`
        } `json:"location"`
    } `json:"geometry,omitempty"`
}

// Operational transforms
type Operation struct {
    P  []interface{} `json:"p"`            // Path
    OI interface{}   `json:"oi,omitempty"` // Object insert
    OD interface{}   `json:"od,omitempty"` // Object delete
    LI interface{}   `json:"li,omitempty"` // List insert
    LD interface{}   `json:"ld,omitempty"` // List delete
}

// Collaboration
type SendInvitesRequest struct {
    Invitees []string `json:"invitees"` // Email addresses
    Message  string   `json:"message,omitempty"`
}

type ShareKeyPermissions struct {
    CanEdit bool `json:"canEdit"`
    CanView bool `json:"canView"`
}
```

### Response Types

```go
type CreateTripResponse struct {
    Success  bool `json:"success"`
    TripPlan struct {
        ID      int    `json:"id"`
        Key     string `json:"key"`
        EditKey string `json:"editKey"`
        Title   string `json:"title"`
    } `json:"tripPlan"`
}

type LikeCount struct {
    Count     int  `json:"count"`
    UserLiked bool `json:"userLiked"`
}

type TripInvite struct {
    Email     string `json:"email"`
    InvitedAt string `json:"invitedAt"`
    Status    string `json:"status"` // "pending", "accepted"
}

type ShareKeyResponse struct {
    ShareKey string `json:"shareKey"`
}
```

## Running Integration Tests

All implemented endpoints have integration tests. To run them:

```bash
# Set up authentication (get from browser cookies)
export WANDERLOG_SESSION_COOKIE='s%3A...'
export WANDERLOG_XSRF_TOKEN='...'

# Optional: specify a test trip ID
export WANDERLOG_TEST_TRIP_ID='your-trip-id'

# Run tests
./test_integration.sh
```

Or run specific tests:

```bash
WANDERLOG_INTEGRATION_TEST=1 go test -v -tags=integration ./pkg/wanderlog -run TestIntegration_CreateAndDeleteTrip
```

## ShareDB Operational Transforms

The `ApplyOperations` endpoint uses ShareDB's JSON0 operational transform format. Helper functions are provided:

```go
// Replace object field
ReplaceInObject(path, oldValue, newValue)

// Insert into object
InsertInObject(path, value)

// Delete from object
DeleteInObject(path, oldValue)

// Insert into array
InsertInList(path, index, value)

// Delete from array
DeleteFromList(path, index, oldValue)

// Replace in array
ReplaceInList(path, index, oldValue, newValue)
```

Example:
```go
// Update trip title
ops := []Operation{
    ReplaceInObject(
        []interface{}{"title"},
        "Old Title",
        "New Title",
    ),
}
client.ApplyOperations(tripKey, ops)
```

## Authentication

All write operations require authentication via session cookies and XSRF tokens:

```go
client := wanderlog.NewClient()
auth := &wanderlog.AuthCredentials{
    SessionCookie: "connect.sid=...",
    XSRFToken:     "...",
}
client.SetAuth(auth)
```

See `pkg/wanderlog/auth_helper.go` for automatic credential management via system keychain.
