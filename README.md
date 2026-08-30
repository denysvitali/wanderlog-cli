# Wanderlog CLI

> **Unofficial** command-line client for [Wanderlog](https://wanderlog.com). Not affiliated with Wanderlog or Travelchime Inc.

A command-line interface for interacting with Wanderlog trip data. Built with Go using Cobra, Viper, Logrus, and Charm's Lipgloss for terminal output.

**Note:** This tool uses Wanderlog's unofficial API, which may change or break at any time without notice.

## Features

- 🌍 **Fetch trip details** from Wanderlog API or local JSON files
- ✍️ **Create and edit trips** - full write functionality with the same API as the Android app
- 📍 **Place management** - add/remove places, manage itineraries
- 🔐 **Authentication** - secure login with session management
- 📋 **Trip management** - list, create, copy, delete trips
- 🖼️ **Images and media** - view trip images and attachments
- 📅 **Trip overview** with dates, duration, and statistics  
- ✈️ **Flight information** with departure/arrival details
- 🗺️ **Day-by-day itinerary** showing destinations and notes
- 🏠 **Places details** with ratings, addresses, and descriptions
- 🎨 **Beautiful terminal output** with colors and emojis
- 📊 **Multiple output formats** - pretty, JSON, and Markdown
- 🤖 **LLM-friendly** Markdown output for AI analysis
- 🔧 **Configurable** logging and output formats
- 📦 **Go package** - usable as both CLI tool and library

## Installation

```bash
go install github.com/denysvitali/wanderlog-cli@latest
```

Or build from source:

```bash
git clone https://github.com/denysvitali/wanderlog-cli.git
cd wanderlog-cli
go build -o wanderlog
```

## Usage

### Reading Trip Data

```bash
# Get trip overview from API
wanderlog trips show abc123xyz

# Get detailed itinerary with flights and destinations
wanderlog trips show abc123xyz --details

# Load trip from local JSON file
wanderlog trips show --file trips/trip1.json

# Show places with details, ratings, and addresses
wanderlog trips places abc123xyz
wanderlog trips places --file trips/trip1.json

# View trip images
wanderlog trips images abc123xyz

# Output as JSON for scripting
wanderlog trips show abc123xyz --output json
wanderlog trips places abc123xyz --output json

# Output as Markdown for LLMs and documentation
wanderlog trips show abc123xyz --output markdown --details
wanderlog trips places abc123xyz --output markdown
```

### Writing and Editing Trips

```bash
# Authenticate with Wanderlog
wanderlog login

# List your trips
wanderlog trips list

# Create a new trip
wanderlog trips create --title "Trip to Japan" --geo-id 1 --start 2026-06-01 --end 2026-06-15

# Copy an existing trip
wanderlog trips copy abc123xyz

# Add a place to a trip
wanderlog trips edit add-place abc123xyz --name "Eiffel Tower" --place-id "ChIJLU7jZClu5kcR4PcOOO6p3I0"

# Add a place with coordinates
wanderlog trips edit add-place abc123xyz --name "Tokyo Station" --lat 35.6812 --lng 139.7671 --start-time 09:30

# Remove a place from a trip  
wanderlog trips edit remove-place abc123xyz 12345

# Set or change a place visit time
wanderlog trips edit set-place-time abc123xyz 12345 --section 100 --start-time 09:30 --end-time 11:00

# Add, edit, or remove a flight reservation (block IDs are shown by `trips flights`)
wanderlog trips flight add abc123xyz --flight-number MU244 --departure-date 2026-06-02 --departure-time 09:30
wanderlog trips flight update abc123xyz 12345 --confirmation ABC123 --arrival-time 14:20
wanderlog trips flight delete abc123xyz 12345

# Add, edit, or remove a lodging reservation
wanderlog trips lodging add abc123xyz --name "Hôtel du Louvre" --place-id ChIJ... --check-in 2026-06-03 --check-out 2026-06-06
wanderlog trips lodging update abc123xyz 23456 --confirmation HOTEL123 --traveler Ada --traveler Grace
wanderlog trips lodging delete abc123xyz 23456

# Add rail/transit. A station Place ID can replace each name/coordinate group.
wanderlog trips train add abc123xyz --carrier "SBB EC 317" --departure-date 2026-06-07 \
  --departure-name "Zürich HB" --departure-lat 47.3782 --departure-lng 8.5402 \
  --arrival-name "Milano Centrale" --arrival-lat 45.4863 --arrival-lng 9.2045
wanderlog trips train delete abc123xyz 34567

# Delete a trip (careful!)
wanderlog trips delete abc123xyz

# Skip the confirmation only in an intentional script
wanderlog trips delete abc123xyz --yes --output json
```

### Authentication

For write operations (creating, editing, deleting trips), you need to authenticate:

```bash
# Interactive login (credentials are securely stored in system keychain)
wanderlog login

# Check authentication status
wanderlog status

# Logout (clear stored credentials)
wanderlog logout

# Or set credentials via environment variables (not recommended for security)
export WANDERLOG_AUTH_SESSION_COOKIE="your-session-cookie"
export WANDERLOG_AUTH_SESSION_XSRF_TOKEN="your-xsrf-token"

# Or pass as flags (not recommended for security)
wanderlog trips create --title "New Trip" --geo-id 1 --session "cookie" --xsrf "token"
```

**Security Features:**
- 🔐 **Secure Storage**: Credentials are stored in your system keychain (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux)
- 🔄 **Automatic Loading**: Once logged in, credentials are automatically used for all write operations
- 🗑️ **Easy Logout**: Clear stored credentials with `wanderlog logout`
- ✅ **Status Check**: Verify authentication status with `wanderlog status`

### Discovery, Feed, and Profile

```bash
# Your current profile
wanderlog user profile

# Another user's profile by id or @username
wanderlog user profile 12345
wanderlog user profile @some-user

# Inbox
wanderlog user notifications
wanderlog user mark-read --id n-123 --id n-456

# Per-user key-value store
wanderlog user kv-get userPrefs
wanderlog user kv-set userPrefs --value '{"theme":"dark"}'

# Notification settings (GET / replace)
wanderlog user settings
wanderlog user settings-set --body '{"notify":true}'

# Search users & relationships
wanderlog user search "alice"
wanderlog user by-email --email someone@example.com
wanderlog user following --user-id 123 --user-id 456
wanderlog user username-taken --username cool-name

# Home feed, history, friends, guides
wanderlog feed home
wanderlog feed recent
wanderlog feed history --offset 20
wanderlog feed friends
wanderlog feed guides --geo-id 1
```

### Journal & advanced trip ops

```bash
# Read a published journal by share key
wanderlog trips journal <journal-key>

# Download expenses for a trip
wanderlog trips expenses <trip-key> > expenses.csv

# Set a budget and add ticket/expense costs
wanderlog trips budget set <trip-key> --amount 2500 --currency USD
wanderlog trips expenses add <trip-key> --description "Museum tickets" --amount 80 --currency USD --category activities
wanderlog trips expenses update <trip-key> <expense-id> --amount 95
wanderlog trips expenses delete <trip-key> <expense-id>

# Register a view, check whether the client needs an upgrade
wanderlog trips register-view <trip-key>
wanderlog trips update-required <trip-key>

# Get / set distinction (badges)
wanderlog trips distinction <trip-key>
wanderlog trips distinction <trip-key> --set community-pick

# Promote a trip into a published guide
wanderlog trips create-guide <trip-key>

# Analyze itinerary density, block mix, metadata consistency, and costs
wanderlog trips analytics <trip-key>
wanderlog trips analytics --file trip.json --output json

# Optimize a route or request destination recommendations
wanderlog trips optimize-route --body '{"mode":"driving","places":[{"id":123},{"id":456}]}'
wanderlog trips recommendations <trip-key> --geo-id 1
```

### Server configuration

```bash
# Pretty-prints /api/config/globalConfig
wanderlog config global

# Authenticated session store
wanderlog config session
wanderlog config session-set somekey --value '"somevalue"'
wanderlog config preferences --locale en
```

### Finding Trip IDs

Trip IDs can be found in Wanderlog URLs:
- URL: `https://wanderlog.com/view/abc123xyz/my-amazing-trip`
- Trip ID: `abc123xyz`

### As a Go Package

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

func main() {
    client := wanderlog.NewClient()
    
    // Read trip data
    trip, err := client.GetTrip("abc123xyz")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Trip: %s\n", trip.TripPlan.Title)
    fmt.Printf("Duration: %s to %s\n", 
        trip.TripPlan.StartDate, 
        trip.TripPlan.EndDate)
    fmt.Printf("Places: %d\n", trip.TripPlan.PlaceCount)
    
    // Authenticate for write operations
    creds, err := client.Login("user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }
    client.SetAuth(creds)
    
    // Create a new trip
    newTrip, err := client.CreateTrip(wanderlog.CreateTripRequest{
        Title: "My New Trip",
        GeoIDs: []int{1},
        InitialMapsPlaceIDs: []int{},
        Type: "plan",
        StartDate: "2024-06-01",
        EndDate: "2024-06-07",
        Privacy: "private",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created trip: %s (ID: %s)\n", 
        newTrip.TripPlan.Title, 
        newTrip.TripPlan.Key)
}
```

## MCP Server (LLM Integration)

Wanderlog CLI includes a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for LLM integration:

```bash
# Start MCP server on stdio (default, for LLM hosts like Claude Code)
wanderlog mcp

# Start authenticated HTTP MCP on localhost (endpoint: /mcp)
export WANDERLOG_MCP_HTTP_TOKEN="replace-with-a-random-secret-of-at-least-16-characters"
wanderlog mcp --http :8080

# A non-loopback bind also requires TLS
wanderlog mcp --http 0.0.0.0:8443 --tls-cert server.crt --tls-key server.key

# Enable write operations (read-only by default)
wanderlog mcp --enable-write

# Set default trip ID for all operations
wanderlog mcp --trip-id abc123xyz
```

HTTP transport always requires `Authorization: Bearer $WANDERLOG_MCP_HTTP_TOKEN`.
An address such as `:8080` is deliberately normalized to `127.0.0.1:8080`;
non-loopback addresses are refused unless both a TLS certificate and key are
configured. HTTP requests are origin-checked, size-limited, and served with
bounded read/header/idle timeouts. Prefer stdio unless remote access is needed.

**Available MCP tools include:**
- **Read-only:** `list_trips`, `get_trip`, `get_trip_plan`, `get_itinerary`, `list_places`, `list_sections`, `get_flights`, `get_trip_sections`, `search_places`, `search_restaurants`, `search_places_wanderlog`, `search_hotels`, `get_place_details`, `get_flight_stops`, `get_like_count`, `list_trip_invites`, `get_me`, `get_user_profile`, `get_notifications`, `get_notification_settings`, `get_user_emails`, `get_user_kv`, `list_following`, `find_user_by_email`, `autocomplete_users`, `is_username_taken`, `get_feed_home`, `get_feed`, `get_feed_v2`, `get_feed_recent`, `get_feed_friends`, `get_trip_history`, `get_if_edited`, `browse_guides`, `search_geos`, `get_view_only_journal`, `get_journal_stop_polylines`, `get_trip_expenses_csv`, `get_trip_distinction`, `get_trip_update_required`, `get_trip_images`, `get_trip_places`, `search_places_in_trips`, `get_all_airlines`, `autocomplete_airports`, `get_hotel_rates`, `get_global_config`, `get_session_store`, `get_session_preferences`
- **Write-gated (need `--enable-write`):** `add_place`, `remove_place`, `move_place`, `reorder_places`, `update_place_notes`, `update_place_visit_time`, `clear_section_blocks`, `delete_section`, `nuke_trip_places`, `add_flight`, `update_flight`, `delete_flight`, `add_lodging`, `update_lodging`, `delete_lodging`, `add_train`, `delete_itinerary_block`, `set_trip_budget`, `add_trip_expense`, `update_trip_expense`, `delete_trip_expense`, `create_trip`, `create_example_trip`, `delete_trip`, `delete_trips`, `restore_trip`, `copy_trip`, `update_trip`, `like_trip`, `send_trip_invites`, `mark_notifications_read`, `set_user_kv`, `update_me`, `update_notification_settings`, `set_utc_offset`, `block_user`, `server_logout`, `set_session_store_value`, `register_trip_view`, `create_guide_from_trip`, `export_trip`, `autofill_day`, `add_checklist_items`, `toggle_checklist_item`, `add_collaborator`, `remove_collaborator`, `get_or_create_share_key`, `set_trip_distinction`, `update_trip_plan_geo`

### Using with Claude Code

Add to your Claude Code MCP config:

```json
{
  "mcpServers": {
    "wanderlog": {
      "command": "wanderlog",
      "args": ["mcp", "--enable-write", "--trip-id", "abc123xyz"]
    }
  }
}
```

## Travel Search

```bash
# List all airlines
wanderlog travel airlines

# Autocomplete airports
wanderlog travel airports "New York"

# Get flight stops for a specific flight
wanderlog travel flight-stops 244 --airline MU --date 2026-05-11

# Search hotels
wanderlog travel hotels Tokyo --check-in 2026-06-01 --check-out 2026-06-07

# Get hotel price rates
wanderlog travel hotel-rates some-prop-id
```

## Raw API Passthrough

For API endpoints that don't have a typed wrapper yet:

```bash
# Call any Wanderlog endpoint directly
wanderlog api /tripPlans/abc123xyz?clientSchemaVersion=2

# POST with JSON body
wanderlog api /user/notifications/markRead -X POST \
  --body '{"notificationIds":["n1"]}'

# With authentication
wanderlog api /user --auth

# Raw output (no formatting)
wanderlog api /config/globalConfig --output raw
```

## Configuration

The CLI supports configuration via:
- Config file: `$XDG_CONFIG_HOME/wanderlog/config.yaml` (normally `~/.config/wanderlog/config.yaml`)
- Environment variables (prefixed with `WANDERLOG_`)
- Command-line flags

Authentication secrets are stored in the system keychain rather than in this
file. Use `--config /path/to/config.yaml` to select a different config file.

## Example Output

```bash
$ wanderlog trips show --file trips/trip1.json --details

🌍 Trip to China

📅 Oct 12, 2025 → Oct 25, 2025 (14 days)

╭───────────────────────────────╮
│                               │
│  📍 8 places  •  👀 15 views  │
│                               │
╰───────────────────────────────╯

✈️  Flights

🛫 EgyptAir 706
   Depart: Oct 12 17:50 from MXP (Milan)
   Arrive: Oct 12 22:30 at CAI (Cairo)

🛫 EgyptAir 955
   Depart: Oct 13 00:50 from CAI (Cairo)
   Arrive: Oct 13 15:20 at PEK (Beijing)

🌍 Destinations

📍 Beijing
   Sunday, Oct 12, 2025
   📝 Arriving in the afternoon

📍 Beijing
   Monday, Oct 13, 2025
   📝 Hotel options: Sunworld Hotel
```

## LLM Integration

The `--output markdown` option produces clean, structured Markdown perfect for feeding to Large Language Models:

```bash
# Generate trip analysis for an LLM
wanderlog trips show abc123xyz --output markdown --details > trip.md

# Get places data for AI processing
wanderlog trips places abc123xyz --output markdown > places.md
```

**Example Markdown output:**
```markdown
# Trip to China

## Trip Information
- **Dates:** October 12, 2025 to October 25, 2025
- **Duration:** 14 days
- **Places:** 8

## Flights
### EgyptAir Flight 706
- **Departure:** October 12, 2025 at 17:50
- **From:** Milan Malpensa Airport (MXP) - Milan
- **Arrival:** October 12, 2025 at 22:30
- **To:** Cairo International Airport (CAI) - Cairo
```

This format allows you to easily:
- 📊 **Analyze trip patterns** with AI
- 💡 **Get travel recommendations** 
- 📋 **Generate travel summaries**
- 🗺️ **Plan optimized itineraries**

## Current Features

**Working:**
- ✅ **Complete CRUD operations** - create, read, update, delete trips
- ✅ **Authentication** - secure login with session management
- ✅ **Trip management** - list, create, copy, delete your trips
- ✅ **Place editing** - add/remove places from itineraries
- ✅ **Trip metadata** - title, dates, duration, statistics
- ✅ **Flight details** - airline, flight numbers, departure/arrival times
- ✅ **Daily itinerary** - destination breakdown with dates
- ✅ **Places information** - ratings, addresses, descriptions, websites
- ✅ **Images and media** - view trip photos and attachments
- ✅ **Notes and text** - travel notes and planning details
- ✅ **Multiple output formats** - pretty terminal, JSON, Markdown
- ✅ **LLM integration** - structured Markdown for AI analysis
- ✅ **Local file loading** - test with offline JSON data
- ✅ **Beautiful formatting** - colorized terminal output with emojis

**Coming Soon:**
- 🔄 **Interactive mode** - explore trips interactively with TUI
- 🔄 **Trip analytics** - distance, duration, cost analysis

**Partially implemented:**
- 🔄 **Budget tracking** - budget and expense writes plus CSV export
- 🔄 **Export features** - Google Maps export via `wanderlog export <trip-key>`

**Already implemented:**
- ✅ **Collaboration** - invite collaborators, manage share keys
- ✅ **Batch operations** - operational transforms for bulk edits
- ✅ **Search and filtering** - Wanderlog place search and trip filtering

## Security

The CLI implements secure credential storage using your system's native keychain:

- **macOS**: Keychain Access
- **Windows**: Windows Credential Manager  
- **Linux**: Secret Service (GNOME Keyring, KDE Wallet, etc.)

Your account password is **never stored**. New session tokens are stored only in
the native keychain. On startup, the CLI removes plaintext passwords left in
config files by older releases and repairs config permissions to `0600`.
`wanderlog status` verifies stored credentials with the server without printing
token fragments, and `wanderlog logout` invalidates the remote session before
clearing local keychain and legacy config data.

Raw `wanderlog api` requests never receive stored credentials unless `--auth` is
explicitly set. Authenticated raw requests must use the configured Wanderlog API
origin, and credential-bearing requests cannot follow cross-origin redirects.

## Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o wanderlog

# Run locally
./wanderlog trips show abc123xyz
```

## Project Structure

```
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command setup
│   ├── mcp.go             # MCP server (LLM integration)
│   ├── mcp_tools.go       # MCP tool definitions
│   ├── helpers.go         # Shared command helpers
│   ├── auth.go            # login/logout/status
│   ├── api.go             # Raw API passthrough
│   ├── trips*.go          # Trip subcommand tree (~10 files)
│   ├── user.go            # User management commands
│   ├── feed.go            # Feed & discovery commands
│   ├── config_cmd.go      # Config & session commands
│   ├── journal.go         # Journal & advanced ops
│   ├── travel.go          # Travel search commands
│   ├── search*.go         # Place search commands
│   └── ...                # Additional command files
├── pkg/
│   ├── wanderlog/         # Core API client
│   │   ├── client.go      # HTTP client & read APIs
│   │   ├── request.go     # Shared request helpers
│   │   ├── auth.go        # Authentication logic
│   │   ├── auth_helper.go # Credential management
│   │   ├── write_ops.go   # Trip write operations
│   │   ├── user_ops.go    # User management APIs
│   │   ├── feed_ops.go    # Feed & discovery APIs
│   │   ├── journal_ops.go # Journal & advanced APIs
│   │   ├── config_ops.go  # Config & session APIs
│   │   ├── visualization.go # Image & stats APIs
│   │   └── models.go      # Generated Go structs
│   └── ui/                # Terminal output formatting
│       ├── trip.go        # Pretty trip output
│       ├── places.go      # Pretty places output
│       ├── markdown.go    # Markdown output
│       └── search.go      # Search results output
├── trips/                 # Example trip data
│   └── trip1.json        # Sample trip for development
├── main.go               # Entry point
├── go.mod                # Go module definition
├── TESTING.md            # Test documentation
└── CLAUDE.md             # Development guidance
```

## Dependencies

- **[Cobra](https://github.com/spf13/cobra)**: CLI framework
- **[Viper](https://github.com/spf13/viper)**: Configuration management  
- **[Logrus](https://github.com/sirupsen/logrus)**: Structured logging
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Disclaimer

This project is **not affiliated with, endorsed by, or associated with** [Wanderlog](https://wanderlog.com) or Travelchime Inc. in any way. It is an independent, unofficial command-line client that interacts with publicly available APIs. "Wanderlog" is a trademark of Travelchime Inc.

## License

MIT License - see LICENSE file for details.
