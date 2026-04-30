# Wanderlog API Endpoints Documentation

This document provides a comprehensive list of all API endpoints discovered in the Wanderlog web/Android bundle (`wanderlog_main/res/raw/dist_public_compiled_mainjs.asset`) and their implementation status in the CLI.

The APK-derived bundle currently exposes 85 quoted `/api/...` endpoint constants via:

```sh
rg -o "['\"](/api/[^'\"]+)['\"]" wanderlog_main/res/raw/dist_public_compiled_mainjs.asset
```

Strongly typed wrappers cover the core trip, search, flight, lodging, collaboration, sharing, checklist, and export flows. For endpoints that are discovered but not modeled yet, the CLI now includes `wanderlog api`, which can call any `/api/...` path with optional authentication and JSON bodies.

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

### Trip Planning Features (Recently Added)

| Endpoint | Go Implementation | Location | Integration Test |
|----------|-------------------|----------|------------------|
| `GET /api/tripPlans/{key}/flights` | `GetTripFlights()` | `write_ops.go:1125` | `TestIntegration_GetTripFlights` |
| `POST /api/tripPlans/{key}/export/v2` | `ExportTrip()` | `write_ops_social.go` | `TestIntegration_ExportTrip` |
| `POST /api/tripPlans/autofillDay` | `AutofillDay()` | `write_ops_social.go` | `TestIntegration_AutofillDay` |
| `POST /api/tripPlans/checklistSection` | `AddChecklistItems()`, `ToggleChecklistItem()` | `write_ops_social.go` | `TestIntegration_AddChecklistItems` |

### Covered Helper APIs

| Endpoint | Go Implementation | CLI Command | MCP Tool |
|----------|-------------------|-------------|----------|
| `GET /api/flights/allAirlines` | `GetAllAirlines()` | `wanderlog travel airlines` | `get_all_airlines` |
| `GET /api/flights/autocompleteAirport` | `AutocompleteAirport()` | `wanderlog travel airports` | `autocomplete_airports` |
| `GET /api/flights/autocompleteAirportWithLocation` | `AutocompleteAirportWithLocation()` | `wanderlog travel airports --lat --lng` | `autocomplete_airports` |
| `GET /api/flights/flightStopsLista` | `GetFlightStops()` | `wanderlog travel flight-stops` | `get_flight_stops` |
| `POST /api/lodging/searchLodgings` | `SearchLodgings()` | `wanderlog travel hotels` | `search_hotels` |
| `POST /api/lodging/getGooglePriceRates` | `GetGooglePriceRates()` | `wanderlog travel hotel-rates` | `get_hotel_rates` |
| `GET /api/placesAPI/getPlaceDetailsAndCardData` | `GetPlaceDetails()` | `wanderlog place-details` | `get_place_details` |
| `GET /api/placesAPI/autocomplete/v2` | `SearchPlacesWithWanderlog()` | `wanderlog search-places` | `search_places`, `search_places_wanderlog` |

### Raw API Coverage

Any APK-discovered endpoint can be exercised through:

```sh
wanderlog api /api/config/globalConfig
wanderlog api /api/user/notification/settings --auth
wanderlog api /api/sessionStore --method POST --auth --body '{"key":"value"}'
```

Use the typed commands where available; use `wanderlog api` for admin, payments, analytics, social feed, notification, profile, session-store, and other less stable endpoints.

### Not Strongly Typed Yet

**User (Recently Added):**

| Endpoint | Go Implementation | CLI Command | MCP Tool |
|----------|-------------------|-------------|----------|
| `GET /api/user` | `GetMe()` | `wanderlog user profile` | `get_me` |
| `POST /api/user` | `UpdateMe()` | - | - |
| `POST /api/user/logout` | `ServerLogout()` | `wanderlog user server-logout` | - |
| `GET /api/user/notifications` | `GetNotifications()` | `wanderlog user notifications` | `get_notifications` |
| `POST /api/user/notifications/markRead` | `MarkNotificationsRead()` | `wanderlog user mark-read` | `mark_notifications_read` (write) |
| `GET/POST /api/user/notification/settings` | `GetNotificationSettings()`, `UpdateNotificationSettings()` | `wanderlog user settings`, `settings-set` | `get_notification_settings` |
| `GET/POST /api/user/keyValue/:key` | `GetKeyValue()`, `SetKeyValue()` | `wanderlog user kv-get/kv-set` | `set_user_kv` (write) |
| `POST /api/user/utcOffset` | `SetUTCOffset()` | `wanderlog user utc-offset` | - |
| `POST /api/user/following/list` | `ListFollowing()` | `wanderlog user following` | - |
| `GET /api/user/autocomplete/:search` | `AutocompleteUsers()` | `wanderlog user search` | `autocomplete_users` |
| `GET /api/user/byEmail` | `FindUserByEmail()` | `wanderlog user by-email` | - |
| `POST /api/user/block` | `BlockUser()` | `wanderlog user block` | - |
| `GET /api/user/isUsernameTaken/:username` | `IsUsernameTaken()` | `wanderlog user username-taken` | `is_username_taken` |
| `GET /api/user/emails` | `GetUserEmails()` | `wanderlog user emails` | `get_user_emails` |
| `GET /api/tripPlans/profile/:userId` | `GetUserProfile()` | `wanderlog user profile <id>` | `get_user_profile` |
| `GET /api/tripPlans/profile/byUsername/:username` | `GetUserProfileByUsername()` | `wanderlog user profile @<name>` | `get_user_profile` |

**Feed & Discovery (Recently Added):**

| Endpoint | Go Implementation | CLI Command | MCP Tool |
|----------|-------------------|-------------|----------|
| `GET /api/tripPlans/home` | `GetFeedHome()` | `wanderlog feed home` | `get_feed_home` |
| `GET /api/tripPlans/feed` | `GetFeed()` | `wanderlog feed legacy` | - |
| `GET /api/tripPlans/feed/v2` | `GetFeedV2()` | `wanderlog feed v2` | - |
| `GET /api/tripPlans/feed/mostRecentlyEdited` | `GetFeedMostRecent()` | `wanderlog feed recent` | `get_feed_recent` |
| `GET /api/tripPlans/friendsPlans` | `GetFriendsPlans()` | `wanderlog feed friends` | `get_feed_friends` |
| `GET /api/tripPlans/history` | `GetTripHistory()` | `wanderlog feed history` | `get_trip_history` |
| `POST /api/tripPlans/getIfEdited` | `GetIfEdited()` | `wanderlog get-if-edited` | - |
| `GET /api/tripPlans/browse/guides[/:geoId]` | `BrowseGuides()` | `wanderlog feed guides` | `browse_guides` |

**Journal & Advanced Trip Ops (Recently Added):**

| Endpoint | Go Implementation | CLI Command | MCP Tool |
|----------|-------------------|-------------|----------|
| `GET /api/tripPlans/viewOnlyJournal/:journalKey` | `GetViewOnlyJournal()` | `wanderlog journal <key>` | `get_view_only_journal` |
| `POST /api/tripPlans/journalStopPolylines` | `GetJournalStopPolylines()` | - | - |
| `GET /api/tripPlans/:key/expensesAsCSV` | `GetTripExpensesCSV()` | `wanderlog expenses` | `get_trip_expenses_csv` |
| `POST /api/tripPlans/:key/registerView` | `RegisterTripView()` | `wanderlog register-view` | `register_trip_view` (write) |
| `GET /api/tripPlans/:key/updateRequired` | `GetTripUpdateRequired()` | `wanderlog update-required` | - |
| `GET/POST /api/tripPlans/:key/distinction` | `GetTripDistinction()`, `SetTripDistinction()` | `wanderlog distinction` | `get_trip_distinction` |
| `POST /api/tripPlans/:key/createGuideFromTripPlan` | `CreateGuideFromTripPlan()` | `wanderlog create-guide` | `create_guide_from_trip` (write) |

**Config & Session (Recently Added):**

| Endpoint | Go Implementation | CLI Command | MCP Tool |
|----------|-------------------|-------------|----------|
| `GET /api/config/globalConfig` | `GetGlobalConfig()` | `wanderlog config global` | `get_global_config` |
| `GET/POST /api/sessionStore` | `GetSessionStore()`, `SetSessionStoreValue()` | `wanderlog config session`, `session-set` | - |
| `GET /api/sessionStore/preferences/:locale` | `GetSessionPreferences()` | `wanderlog config preferences` | - |

**Still Out of Scope (intentionally):**
- OAuth login variants (`/api/user/loginFacebookAccessToken`, `/api/user/loginGoogleAuthCode/v2`, `/api/user/loginGoogleIdToken`, `/api/user/loginAppleAuthCode`) — require browser handshake; email login covers CLI needs
- `/api/user/register`, `/api/user/resetPassword`, `/api/user/startResetPassword`, `/api/user/isValidPasswordResetToken`, `/api/user/activate/*`, `/api/user/changeEmail/*` — account-lifecycle flows better handled in the web UI
- `/api/user/leaderboard`, `/api/user/following/*` (beyond `list`), `/api/user/mutuallyFollowing`, `/api/user/followingMultiple`, `/api/user/profilePicture`, `/api/user/fcmToken`, `/api/user/saveFlightDealSettings` — low-value for CLI/LLM workflows
- `/api/tripPlans/landingPage/*` — marketing surface
- `/api/tripPlans/admin/*` — require admin tokens
- `/api/payments/*` — Stripe browser flows
- `/api/analytics/*` — client telemetry, irrelevant for a CLI

---

## Implementation Status Legend

- ✅ **Fully Implemented** - Complete implementation with tests
- ✅ **Partial** - Basic implementation exists but may lack full features
- ❌ **Not Implemented** - Endpoint exists in web app but not in CLI
- 🧰 **Raw API Covered** - Endpoint can be called with `wanderlog api`, but no typed wrapper exists yet

> **Note**: The tables below are a historical inventory of every discovered
> endpoint. Browser-only, payment, telemetry, and admin endpoints may remain
> intentionally unwrapped even when they are available through `wanderlog api`.

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
| `/api/flights/allAirlines` | GET | ✅ `GetAllAirlines()` | `client.go` | `get_all_airlines` |
| `/api/flights/autocompleteAirport` | GET | ✅ `AutocompleteAirport()` | `client.go` | `autocomplete_airports` |
| `/api/flights/autocompleteAirportWithLocation` | GET | ✅ `AutocompleteAirportWithLocation()` | `client.go` | `autocomplete_airports` |
| `/api/flights/flightStopsLista` | GET | ✅ `GetFlightStops()` | `client.go` | `get_flight_stops` |
| `/api/tripPlans/:key/flights` | GET | ✅ `GetTripFlights()` | `write_ops_social.go:346` | `TestIntegration_GetTripFlights` |
| `/api/lodging/searchLodgings` | POST | ✅ `SearchLodgings()` | `client.go:671` | `search_hotels` |
| `/api/lodging/getGooglePriceRates` | POST | ✅ `GetGooglePriceRates()` | `client.go` | `get_hotel_rates` |

> **Note:** The MCP tool `search_hotels` uses the lodging search method. The old `search_flights` tool was removed because it returned placeholder errors; attached trip flights are available through `get_flights`.

### Collaboration

| Endpoint | Method | Implementation | Location | MCP Tool |
|----------|--------|----------------|----------|----------|
| `/api/tripPlans/:key/invite` | POST | ✅ `SendTripInvites()` | `write_ops.go:812` | `send_trip_invites` |
| `/api/tripPlans/:key/invites` | GET | ✅ `ListTripInvites()` | `write_ops.go:855` | `list_trip_invites` |
| `/api/tripPlans/:key/collaborator` | POST | ✅ `AddCollaborator()` | `write_ops_social.go` | `add_collaborator` |
| `/api/tripPlans/:key/collaborator` | DELETE | ✅ `RemoveCollaborator()` | `write_ops_social.go` | `remove_collaborator` |
| `/api/tripPlans/:editKey/shareKey` | POST | ✅ `GetOrCreateShareKey()` | `write_ops_social.go` | `get_or_create_share_key` |

### Social Features

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/:key/like` | POST | ✅ `SetLike()` | `write_ops.go:891` | `TestIntegration_SetLike` |
| `/api/tripPlans/:key/likeCount` | GET | ✅ `GetLikeCount()` | `write_ops.go:934` | `TestIntegration_GetLikeCount` |

### User Management & Authentication

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/user` | GET/POST | ✅ `GetMe()` / `UpdateMe()` | Get/update user profile |
| `/api/user/login` | POST | ✅ Partial | User login (implemented in auth.go) |
| `/api/user/logout` | POST | ✅ `ServerLogout()` | User logout |
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
| `/api/user/isUsernameTaken/:username` | GET | ✅ `IsUsernameTaken()` | Check username availability |
| `/api/user/subscribeBlog` | POST | ❌ | Subscribe to blog |
| `/api/user/keyValue/:key` | GET/POST | ✅ `GetUserKV()` / `SetUserKV()` | Key-value storage |
| `/api/user/notification/settings` | GET/POST | ✅ `GetNotificationSettings()` / `UpdateNotificationSettings()` | Notification settings |
| `/api/user/notifications` | GET | ✅ `GetNotifications()` | Get notifications |
| `/api/user/notifications/markRead` | POST | ✅ `MarkNotificationsRead()` | Mark notifications read |
| `/api/user/emails` | GET | ✅ `GetUserEmails()` | Get user emails |
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
| `/api/user/autocomplete/:search` | GET | ✅ `AutocompleteUsers()` | Autocomplete users |
| `/api/user/byEmail` | POST | ❌ | Find user by email |
| `/api/user/leaderboard` | GET | ❌ | Get leaderboard |
| `/api/user/block` | POST | ❌ | Block user |
| `/api/user/combine/:token` | GET/POST | ❌ | Combine accounts |

### Trip Operations - Feed & Discovery

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans` | GET | ✅ | List user trips |
| `/api/tripPlans/feed` | GET | ✅ `GetFeed()` | Get trip feed |
| `/api/tripPlans/feed/v2` | GET | ✅ `GetFeedV2()` | Get trip feed v2 |
| `/api/tripPlans/feed/mostRecentlyEdited` | GET | ✅ `GetFeedMostRecent()` | Recently edited trips |
| `/api/tripPlans/home` | GET | ✅ `GetFeedHome()` | Home feed |
| `/api/tripPlans/history` | GET | ✅ `GetTripHistory()` | Trip history |
| `/api/tripPlans/friendsPlans` | GET | ✅ `GetFriendsPlans()` | Friends' trips |
| `/api/tripPlans/myProfile/` | GET | ✅ | User's profile trips |
| `/api/tripPlans/profile/:userId` | GET | ✅ `GetUserProfile()` | User profile trips |
| `/api/tripPlans/profile/byUsername/:username` | GET | ✅ `GetUserProfileByUsername()` | Profile by username |
| `/api/tripPlans/profile/sampleMapsByUsernames/:usernames` | GET | ❌ | Sample maps by usernames |

### Trip Operations - Guides & Browse

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans/browse/guides` | GET | ✅ `BrowseGuides()` | Browse guides |
| `/api/tripPlans/browse/guides/:geoId` | GET | ✅ `BrowseGuides()` | Browse guides by location |
| `/api/tripPlans/landingPage/guides` | GET | ❌ | Landing page guides |
| `/api/tripPlans/landingPage/stories` | GET | ❌ | Landing page stories |
| `/api/tripPlans/:key/:geoId/relatedGuides` | GET | ❌ | Related guides |

### Trip Operations - Advanced

| Endpoint | Method | Implementation | Location | Test |
|----------|--------|----------------|----------|------|
| `/api/tripPlans/createExampleTripPlan` | POST | ✅ `CreateExampleTrip()` | `write_ops_trip.go:126` | - |
| `/api/tripPlans/:key/flights` | GET | ✅ `GetTripFlights()` | `write_ops.go:1125` | `TestIntegration_GetTripFlights` |
| `/api/tripPlans/:key/export/v2` | POST | ✅ `ExportTrip()` | `write_ops.go:1155` | `TestIntegration_ExportTrip` |
| `/api/tripPlans/:key/createGuideFromTripPlan` | POST | ✅ `CreateGuideFromTripPlan()` | `journal_ops.go:158` | - |
| `/api/tripPlans/:key/updateTripPlanGeo/:geoId` | POST | ✅ `UpdateTripPlanGeo()` | `journal_ops.go:158` | `TestUpdateTripPlanGeo` |
| `/api/tripPlans/:key/distinction` | GET/POST | ✅ `GetTripDistinction()` / `SetTripDistinction()` | `journal_ops.go:118` | - |
| `/api/tripPlans/:key/registerView` | POST | ✅ `RegisterTripView()` | `journal_ops.go:80` | - |
| `/api/tripPlans/:key/updateRequired` | GET | ✅ `GetTripUpdateRequired()` | `journal_ops.go:94` | - |
| `/api/tripPlans/getIfEdited` | POST | ✅ `GetIfEdited()` | `feed_ops.go:129` | - |
| `/api/tripPlans/:key/sections` | GET | ✅ | `GetTripSections()` | |
| `/api/tripPlans/:key/sections/:sectionId/place` | POST | ✅ | `AddPlace()` | |
| `/api/tripPlans/:key/sections/:sectionId/place/:placeId` | DELETE | ✅ | `RemovePlace()` | |
| `/api/tripPlans/autofillDay` | POST | ✅ `AutofillDay()` | `write_ops.go:1201` | `TestIntegration_AutofillDay` |
| `/api/tripPlans/checklistSection` | POST | ✅ `AddChecklistItems()` | `write_ops.go:1257` | `TestIntegration_AddChecklistItems` |

### Trip Operations - Journal & View Only

| Endpoint | Method | Status | Purpose |
|----------|--------|--------|---------|
| `/api/tripPlans/viewOnlyJournal/:journalKey` | GET | ✅ `GetViewOnlyJournal()` | View-only journal |
| `/api/tripPlans/viewOnlyJournal/mobile/:journalKey` | GET | ❌ | Mobile view-only journal |
| `/api/tripPlans/:key/expensesAsCSV` | GET | ✅ `GetTripExpensesCSV()` | Export expenses as CSV |

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
| `/api/sessionStore` | GET/POST | ✅ `GetSessionStore()` / `SetSessionStoreValue()` | Session storage |
| `/api/sessionStore/preferences/:key` | GET/POST | ✅ `GetSessionPreferences()` | Session preferences |
| `/api/config/globalConfig` | GET | ✅ `GetGlobalConfig()` | Global configuration |

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
