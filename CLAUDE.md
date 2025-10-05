# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

```bash
# Build the project
go build -o wanderlog

# Run locally
./wanderlog [command]

# Install dependencies
go mod download

# Run tests
go test ./...

# Run from source
go run . [command]

# Install globally
go install github.com/denysvitali/wanderlog-cli@latest
```

## Architecture Overview

This is a Go CLI application for interacting with the Wanderlog travel planning API. It provides both read and write operations for trip management.

### Core Structure

- **Main Entry**: `main.go` → `cmd.Execute()` - Simple entry point
- **CLI Framework**: Built with Cobra for command structure and Viper for configuration
- **Commands**: Located in `cmd/` - each command is a separate file (trip.go, places.go, auth.go, etc.)
- **Core Logic**: `pkg/wanderlog/` contains the API client and data models
- **UI Formatting**: `pkg/ui/` handles terminal output formatting with Lipgloss styling

### Key Components

**API Client (`pkg/wanderlog/client.go`)**:
- HTTP client with 30s timeout
- Handles authentication via session cookies and XSRF tokens
- Base URL: `https://wanderlog.com/api`
- Client version: "2"
- All API operations go through the `Client` struct

**Authentication System**:
- Uses system keychain for secure credential storage (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Session-based auth with cookies and XSRF tokens
- Credentials stored via `github.com/zalando/go-keyring`
- `EnsureAuthenticated()` in `auth_helper.go` provides automatic credential loading from keychain, env vars, or flags
- Login credentials (email/password) are NEVER stored - only session tokens

**Command Structure**:
- `trip` - View trip details and itineraries
- `places` - View places in a trip
- `auth` (login/logout/status) - Authentication management
- `create` - Create new trips
- `edit` - Modify existing trips (add/remove places)
- `list` - List user's trips
- `search` - Search for places
- `mcp` - Model Context Protocol server for LLM integration

**Output Formats**:
- Pretty (default) - Colorized terminal output with emojis using Lipgloss
- JSON - Machine readable
- Markdown - LLM-friendly structured format

### Data Models

Generated Go structs in `pkg/wanderlog/models.go` represent the Wanderlog API responses:
- `TripResponse` - Complete trip data with itinerary
- `Place` - Location data with coordinates, ratings, descriptions
- `Flight` - Flight information with airline, times, airports
- `Destination` - Daily itinerary destinations with notes
- `Metadata` - Place metadata with detailed information

### Configuration

- Config file: `~/.wanderlog.yaml`
- Environment variables with `WANDERLOG_` prefix
- Command-line flags override config and env vars
- Viper handles the configuration hierarchy

### Security Features

- Never stores login credentials (email/password)
- Only session tokens stored in system keychain
- Automatic token loading for write operations
- Easy logout to clear stored credentials

### MCP Server Implementation

The `mcp` command (`cmd/mcp.go`) implements a Model Context Protocol server:
- **Read-only mode by default** - Use `--enable-write` flag to enable write operations
- **Tools**: list_trips, get_trip, list_places, list_sections, search_places, get_place_details, add_place (write mode), remove_place (write mode)
- **Resources**: Trip details accessible via `wanderlog://trips/{trip_id}` URI
- **Prompts**: analyze_trip prompt for trip analysis
- **Transport**: Supports both stdio (default) and HTTP server (`--http` flag)
- **Default trip ID**: Use `--trip-id` flag to set a default trip ID for all operations
- Context injection via `withTripID()` and `tripIDFromContext()` functions

### Development Patterns

- Use `logrus` for structured logging with debug/info levels
- HTTP client in `pkg/wanderlog/client.go` handles all API communication
- UI formatting separated from business logic in `pkg/ui/`
- Commands follow Cobra patterns with persistent flags
- Error handling returns detailed errors to user
- Write operations in `pkg/wanderlog/write_ops.go` use operational transforms for complex edits

### API Integration Notes

**Authentication Flow**:
1. User calls `wanderlog login` with email/password
2. Client calls `Login()` in `auth.go` which POSTs to `/api/user/login`
3. Response contains session cookie (`connect.sid`) and XSRF token
4. Tokens stored in system keychain via `SaveCredentials()` in `keychain.go`
5. All subsequent write operations use `addAuthHeaders()` to attach cookies and XSRF token

**Write Operations**:
- Located in `write_ops.go`
- All require authentication via `addAuthHeaders()`
- Use operational transforms (`ApplyOperations()`) for complex batch edits
- Common operations: `CreateTrip()`, `DeleteTrip()`, `AddPlace()`, `RemovePlace()`, `CopyTrip()`

**Place Search**:
- Two methods: Google Places API (requires `GOOGLE_PLACES_API_KEY`) and Wanderlog's native autocomplete
- `SearchPlaces()` uses new Google Places API v1 with Text Search
- `SearchPlacesWithWanderllog()` uses Wanderlog's autocomplete API (no API key needed)
- `GetPlaceDetails()` fetches detailed place info from Wanderlog's API

### Notable Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/zalando/go-keyring` - Secure credential storage
- `github.com/mark3labs/mcp-go` - Model Context Protocol implementation
- `github.com/sirupsen/logrus` - Structured logging
