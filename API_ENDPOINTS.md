# Wanderlog API Endpoints Documentation

This document provides a comprehensive list of all API endpoints discovered in the Wanderlog web application (from `dist_public_compiled_mainjs.js`) and their implementation status in the CLI.

## API Coverage Summary

### Well Covered (Core Trip Operations)

| Endpoint | Go Implementation | Location |
|----------|-------------------|----------|
| `GET /api/tripPlans/{key}` | `GetTrip()` | `client.go:45` |
| `GET /api/tripPlans/{key}/sections` | `GetTripSections()` | `client.go:106` |
| `POST /api/tripPlans` | `CreateTrip()` | `write_ops.go:50` |
| `DELETE /api/tripPlans/{key}` | `DeleteTrip()` | `write_ops.go:120` |
| `POST /api/tripPlans/{key}/applyOps` | `UpdateTrip()`, `ApplyOperations()` | `write_ops.go:154,394` |
| `POST /api/tripPlans/{key}/sections/{id}/places` | `AddPlace()` | `write_ops.go:246` |
| `DELETE /api/tripPlans/{key}/sections/{id}/places/{id}` | `RemovePlace()` | `write_ops.go:346` |
| `GET /api/tripPlans/{key}/images` | `GetTripImages()` | `visualization.go:135` |
| `POST /api/tripPlans/{key}/like` | `SetLike()` | `write_ops.go:891` |
| `GET /api/tripPlans/{key}/likeCount` | `GetLikeCount()` | `write_ops.go:934` |
| `POST /api/tripPlans/{key}/invite` | `SendTripInvites()` | `write_ops.go:812` |
| `GET /api/tripPlans/{key}/invites` | `ListTripInvites()` | `write_ops.go:855` |
| `POST /api/tripPlans/copy/{key}` | `CopyTrip()` | `write_ops.go:710` |
| `POST /api/tripPlans/{key}/restore` | `RestoreTrip()` | `write_ops.go:778` |
| `GET /api/tripPlans/myProfile/` | `GetUserTrips()` | `visualization.go:97` |

### Partially Covered

| Endpoint | Status | MCP Tool |
|----------|--------|----------|
| `GET /api/flights/allAirlines` | `GetAllAirlines()` | - |
| `GET /api/flights/autocompleteAirport` | `AutocompleteAirport()` | `search_flights` |
| `GET /api/flights/autocompleteAirportWithLocation` | `AutocompleteAirportWithLocation()` | `search_flights` |
| `GET /api/flights/flightStopsLista` | `GetFlightStops()` | - |
| `GET /api/tripPlans/flights` | **Not Implemented** | - |
| `POST /api/lodging/searchLodgings` | `SearchLodgings()` | `search_hotels` |
| `POST /api/lodging/getGooglePriceRates` | `GetGooglePriceRates()` | - |
| `GET /api/placesAPI/getPlaceDetailsAndCardData` | `GetPlaceDetails()` | `get_place_details` |
| `GET /api/placesAPI/autocomplete/v2` | `SearchPlacesWithWanderllog()` | `search_places` |

### Not Covered (Significant Gaps)

**User/Auth (social login missing):**
- `/api/user/loginFacebookAccessToken` - Facebook OAuth login
- `/api/user/loginGoogleAuthCode/v2` - Google OAuth login v2
- `/api/user/loginGoogleIdToken` - Google ID token login
- `/api/user/loginAppleAuthCode` - Apple OAuth login
- `/api/user/register` - User registration
- `/api/user/resetPassword` - Password reset
- `/api/user/isValidPasswordResetToken` - Validate reset token
- `/api/user/profilePicture` - Profile picture upload
- `/api/user/following/*` - Following users
- `/api/user/block` - Block users
- `/api/user/byEmail` - Find user by email

**Social/Feed:**
- `/api/tripPlans/feed` - Trip feed
- `/api/tripPlans/home` - Home feed
- `/api/tripPlans/friendsPlans` - Friends' trips
- `/api/user/leaderboard` - Leaderboard

**Payments (not applicable for CLI):**
- All `/api/payments/*` endpoints (subscription management via Stripe)

**Trip Features:**
- `/api/tripPlans/{key}/export/v2` - Export to Google Maps
- `/api/tripPlans/autofillDay` - Auto-fill day with suggestions
- `/api/tripPlans/checklistSection` - Manage checklist sections
- `/api/tripPlans/browse/guides` - Browse guides
- `/api/tripPlans/landingPage/*` - Landing page content

---

## Implementation Status Legend

- ✅ **Fully Implemented** - Complete implementation with tests
- ✅ **Partial** - Basic implementation exists but may lack full features
- ❌ **Not Implemented** - Endpoint exists in web app but not in CLI

## Complete API Endpoint Catalog

### Core Trip Operations

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans` | GET | ✅ `GetTrip()` (with key) | `client.go:45` | Multiple tests |
| `/api/tripPlans` | POST | ✅ `CreateTrip()` | `write_ops.go:50` | `TestIntegration_CreateAndDeleteTrip` |
| `/api/tripPlans/:key` | PUT/POST | ✅ `UpdateTrip()` | `write_ops.go:154` | `TestIntegration_UpdateTrip` |
| `/api/tripPlans/:key` | DELETE | ✅ `DeleteTrip()` | `write_ops.go:120` | `TestIntegration_CreateAndDeleteTrip` |
| `/api/tripPlans/:key/restore` | POST | ✅ `RestoreTrip()` | `write_ops.go:778` | `TestIntegration_RestoreTrip` |
| `/api/tripPlans/:key/sections` | GET | ✅ `GetTripSections()` | `client.go:106` | `TestIntegration_GetTripSections` |
| `/api/tripPlans/copy/:key` | POST | ✅ `CopyTrip()` | `write_ops.go:710` | `TestIntegration_CopyTrip` |
| `/api/tripPlans/myProfile/` | GET | ✅ `GetUserTrips()` | `visualization.go:97` | - |

### Place Management

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/sections/:sectionId/place` | POST | ✅ `AddPlace()` | `write_ops.go:246` | `TestIntegration_AddAndRemovePlace` |
| `/api/tripPlans/:key/sections/:sectionId/place/:placeId` | DELETE | ✅ `RemovePlace()` | `write_ops.go:346` | `TestIntegration_AddAndRemovePlace` |
| `/api/tripPlans/:key/applyOps` | POST | ✅ `ApplyOperations()` | `write_ops.go:394` | `TestIntegration_ApplyOperations` |
| `MovePlace` (uses applyOps) | - | ✅ `MovePlace()` | `write_ops.go:588` | `TestMCPIntegration_MovePlace` |
| `ReorderPlaces` (uses applyOps) | - | ✅ `ReorderPlaces()` | `write_ops.go:653` | `TestMCPIntegration_ReorderPlacesTool` |

### Advanced Operations

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `NukeTripPlaces` (custom) | - | ✅ `NukeTripPlaces()` | `write_ops.go:536` | `TestIntegration_NukeTripPlaces` |
| `ClearSectionBlocks` (custom) | - | ✅ `ClearSectionBlocks()` | `write_ops.go:483` | - |
| `DeleteSection` (custom) | - | ✅ `DeleteSection()` | `write_ops.go:509` | - |

### Images & Media

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/images` | GET | ✅ `GetTripImages()` | `visualization.go:135` | `TestIntegration_GetTripImages` |
| `/api/tripPlans/:key/image` | POST | ❌ Not Implemented | - | - |
| `/api/tripPlans/:key/attachment` | POST | ❌ Not Implemented | - | - |

### Flight & Lodging Search

| Endpoint | Method | Implementation | Location | MCP Tool |
|----------|--------|----------------|----------|----------|
| `/api/flights/allAirlines` | GET | ✅ `GetAllAirlines()` | `client.go:474` | - |
| `/api/flights/autocompleteAirport` | GET | ✅ `AutocompleteAirport()` | `client.go:506` | `search_flights` |
| `/api/flights/autocompleteAirportWithLocation` | GET | ✅ `AutocompleteAirportWithLocation()` | `client.go:538` | `search_flights` |
| `/api/flights/flightStopsLista` | GET | ✅ `GetFlightStops()` | `client.go:571` | - |
| `/api/tripPlans/flights` | GET | ❌ Not Implemented | - | - |
| `/api/lodging/searchLodgings` | POST | ✅ `SearchLodgings()` | `client.go:671` | `search_hotels` |
| `/api/lodging/getGooglePriceRates` | POST | ✅ `GetGooglePriceRates()` | `client.go:721` | - |

> **Note:** The MCP tools `search_flights` and `search_hotels` (added in commit `f25b96d`) use the airport autocomplete and lodging search methods respectively. However, the `/api/tripPlans/flights` endpoint for retrieving flights attached to a trip is not yet implemented in the Go client.

### Collaboration

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/invite` | POST | ✅ `SendTripInvites()` | `write_ops.go:812` | - |
| `/api/tripPlans/:key/invites` | GET | ✅ `ListTripInvites()` | `write_ops.go:855` | `TestIntegration_ListTripInvites` |
| `/api/tripPlans/:key/collaborator` | POST | ✅ `AddCollaborator()` | `write_ops.go:969` | - |
| `/api/tripPlans/:key/collaborator` | DELETE | ✅ `RemoveCollaborator()` | `write_ops.go:1016` | - |
| `/api/tripPlans/:editKey/shareKey` | POST | ✅ `GetOrCreateShareKey()` | `write_ops.go:1063` | - |

### Social Features

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/like` | POST | ✅ `SetLike()` | `write_ops.go:891` | `TestIntegration_SetLike` |
| `/api/tripPlans/:key/likeCount` | GET | ✅ `GetLikeCount()` | `write_ops.go:934` | `TestIntegration_GetLikeCount` |

### User Management & Authentication

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/user` | GET/POST | ❌ | Get/update user profile |
| `/api/user/login` | POST | ✅ Partial | User login (implemented in auth.go) |
| `/api/user/logout` | POST | ❌ | User logout |
| `/api/user/register` | POST | ❌ | User registration |
| `/api/user/profilePicture` | POST | ❌ | Update profile picture |
| `/api/user/loginFacebookAccessToken` | POST | ❌ | Facebook OAuth login |
| `/api/user/loginGoogleAuthCode/v2` | POST | ❌ | Google OAuth login v2 |
| `/api/user/loginGoogleIdToken` | POST | ❌ | Google ID token login |
| `/api/user/loginAppleAuthCode` | POST | ❌ | Apple OAuth login |
| `/api/user/createPendingUser` | POST | ❌ | Create pending user |
| `/api/user/loginToken` | GET/POST | ❌ | Token-based login |
| `/api/user/loginToken/login` | POST | ❌ | Login with token |
| `/api/user/activate/:email/:key` | POST | ❌ | Activate user account |
| `/api/user/startResetPassword` | POST | ❌ | Start password reset |
| `/api/user/isValidPasswordResetToken` | POST | ❌ | Validate reset token |
| `/api/user/resetPassword` | POST | ❌ | Reset password |
| `/api/user/changeEmail/:token` | POST | ❌ | Change email address |
| `/api/user/createIncompleteUserSignup` | POST | ❌ | Track incomplete signups |

### User Settings & Preferences

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/user/utcOffset` | POST | ❌ | Set user timezone |
| `/api/user/isUsernameTaken/:username` | GET | ❌ | Check username availability |
| `/api/user/subscribeBlog` | POST | ❌ | Subscribe to blog |
| `/api/user/keyValue/:key` | GET/POST | ❌ | Key-value storage |
| `/api/user/notification/settings` | GET/POST | ❌ | Notification settings |
| `/api/user/notifications` | GET | ❌ | Get notifications |
| `/api/user/notifications/markRead` | POST | ❌ | Mark notifications read |
| `/api/user/emails` | GET | ❌ | Get user emails |
| `/api/user/fcmToken` | POST | ❌ | Firebase Cloud Messaging token |
| `/api/user/saveFlightDealSettings` | POST | ❌ | Flight deal preferences |
| `/api/user/isDeleting` | GET/POST | ❌ | Account deletion status |

### Social Features (Users)

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/user/following/list` | GET | ❌ | List following |
| `/api/user/following/visitedGeo/:geoId` | GET | ❌ | Following who visited location |
| `/api/user/followingMultiple` | POST | ❌ | Follow multiple users |
| `/api/user/mutuallyFollowing` | GET | ❌ | Get mutual followers |
| `/api/user/:userId/follows` | GET | ❌ | Check if user follows |
| `/api/user/:userId/email` | GET | ❌ | Get user email |
| `/api/user/autocomplete/:search` | GET | ❌ | Autocomplete users |
| `/api/user/byEmail` | POST | ❌ | Find user by email |
| `/api/user/leaderboard` | GET | ❌ | Get leaderboard |
| `/api/user/block` | POST | ❌ | Block user |
| `/api/user/combine/:token` | GET/POST | ❌ | Combine accounts |

### Trip Operations - Feed & Discovery

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans` | GET | ✅ | List user trips |
| `/api/tripPlans/feed` | GET | ❌ | Get trip feed |
| `/api/tripPlans/feed/v2` | GET | ❌ | Get trip feed v2 |
| `/api/tripPlans/feed/mostRecentlyEdited` | GET | ❌ | Recently edited trips |
| `/api/tripPlans/home` | GET | ❌ | Home feed |
| `/api/tripPlans/history` | GET | ❌ | Trip history |
| `/api/tripPlans/friendsPlans` | GET | ❌ | Friends' trips |
| `/api/tripPlans/myProfile/` | GET | ✅ | User's profile trips |
| `/api/tripPlans/profile/:userId` | GET | ❌ | User profile trips |
| `/api/tripPlans/profile/byUsername/:username` | GET | ❌ | Profile by username |
| `/api/tripPlans/profile/sampleMapsByUsernames/:usernames` | GET | ❌ | Sample maps by usernames |

### Trip Operations - Guides & Browse

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans/browse/guides` | GET | ❌ | Browse guides |
| `/api/tripPlans/browse/guides/:geoId` | GET | ❌ | Browse guides by location |
| `/api/tripPlans/landingPage/guides` | GET | ❌ | Landing page guides |
| `/api/tripPlans/landingPage/stories` | GET | ❌ | Landing page stories |
| `/api/tripPlans/:key/:geoId/relatedGuides` | GET | ❌ | Related guides |

### Trip Operations - Advanced

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans/createExampleTripPlan` | POST | ❌ | Create example trip |
| `/api/tripPlans/flights` | GET | ❌ | Get flights |
| `/api/tripPlans/:key/export/v2` | POST | ❌ | Export to Google Maps |
| `/api/tripPlans/:key/createGuideFromTripPlan` | POST | ❌ | Convert trip to guide |
| `/api/tripPlans/:key/updateTripPlanGeo/:geoId` | POST | ❌ | Update trip geography |
| `/api/tripPlans/:key/distinction` | GET/POST | ❌ | Get/set trip distinction |
| `/api/tripPlans/:key/registerView` | POST | ❌ | Register trip view |
| `/api/tripPlans/:key/updateRequired` | GET | ❌ | Check if update required |
| `/api/tripPlans/getIfEdited` | POST | ❌ | Get if edited |
| `/api/tripPlans/:key/sections` | GET | ✅ | Get trip sections |
| `/api/tripPlans/:key/sections/:sectionId/place` | POST | ✅ | Add place to section |
| `/api/tripPlans/:key/sections/:sectionId/place/:placeId` | DELETE | ✅ | Remove place from section |
| `/api/tripPlans/autofillDay` | POST | ❌ | Auto-fill day with suggestions |
| `/api/tripPlans/checklistSection` | POST | ❌ | Manage checklist sections |

### Trip Operations - Journal & View Only

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans/viewOnlyJournal/:journalKey` | GET | ❌ | View-only journal |
| `/api/tripPlans/viewOnlyJournal/mobile/:journalKey` | GET | ❌ | Mobile view-only journal |
| `/api/tripPlans/:key/expensesAsCSV` | GET | ❌ | Export expenses as CSV |

### Payments & Subscriptions

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/payments/subscriptionsInfo` | GET | ❌ | Subscription info |
| `/api/payments/discountInfo/v2` | GET | ❌ | Discount information |
| `/api/payments/ensureStripeSubscriptionUpdated` | POST | ❌ | Update Stripe subscription |
| `/api/payments/newStripeSubscriptionInfo` | GET | ❌ | New subscription info |
| `/api/payments/maybeStartMobileSubscription` | POST | ❌ | Start mobile subscription |
| `/api/payments/updateSubscriptionCanceled` | POST | ❌ | Cancel subscription |
| `/api/payments/latestSubscriptionPrice` | GET | ❌ | Get latest price |
| `/api/payments/changeStripeSubscription` | POST | ❌ | Change subscription |
| `/api/payments/startStripeTrial` | POST | ❌ | Start trial |
| `/api/payments/logSubscriptionError` | POST | ❌ | Log subscription error |
| `/api/payments/proDiscountLandingPage/:slug` | GET | ❌ | Pro discount landing |

### Admin & System

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans/admin/recent/plans` | GET | ❌ | Recent plans (admin) |
| `/api/tripPlans/admin/recent/recommendations` | GET | ❌ | Recent recommendations (admin) |
| `/api/mailboxes/:id` | GET | ❌ | Get mailbox |
| `/api/sessionStore` | GET/POST | ❌ | Session storage |
| `/api/sessionStore/preferences/:key` | GET/POST | ❌ | Session preferences |
| `/api/config/globalConfig` | GET | ❌ | Global configuration |

### Analytics

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/analytics/firebaseAppInstanceId` | POST | ❌ | Firebase instance ID |
| `/api/analytics/googleAnalyticsClientId` | POST | ❌ | GA client ID |
| `/api/analytics/trackTestEvent` | POST | ❌ | Track test event |

## Notes

1. The CLI focuses on core trip management functionality
2. OAuth providers and social features are generally not implemented
3. Payment/subscription endpoints are not needed for CLI usage
4. Some endpoints may require specific cookies or tokens from the web app
5. The `/api/tripPlans/:key/applyOps` endpoint is the primary method for complex trip edits using ShareDB operational transforms

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
