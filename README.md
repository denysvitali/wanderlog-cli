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
wanderlog trip abc123xyz

# Get detailed itinerary with flights and destinations
wanderlog trip abc123xyz --details

# Load trip from local JSON file
wanderlog trip --file trips/trip1.json

# Show places with details, ratings, and addresses
wanderlog places abc123xyz
wanderlog places --file trips/trip1.json

# View trip images
wanderlog images abc123xyz

# Output as JSON for scripting
wanderlog trip abc123xyz --format json
wanderlog places abc123xyz --format json

# Output as Markdown for LLMs and documentation
wanderlog trip abc123xyz --format markdown --details
wanderlog places abc123xyz --format markdown
```

### Writing and Editing Trips

```bash
# Authenticate with Wanderlog
wanderlog login

# List your trips
wanderlog list

# Create a new trip
wanderlog create --title "Trip to Japan" --start 2024-06-01 --end 2024-06-15

# Copy an existing trip
wanderlog copy abc123xyz

# Add a place to a trip
wanderlog edit add-place abc123xyz --name "Eiffel Tower" --place-id "ChIJLU7jZClu5kcR4PcOOO6p3I0"

# Add a place with coordinates
wanderlog edit add-place abc123xyz --name "Tokyo Station" --lat 35.6812 --lng 139.7671

# Remove a place from a trip  
wanderlog edit remove-place abc123xyz 12345

# Delete a trip (careful!)
wanderlog delete abc123xyz
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
export WANDERLOG_SESSION="your-session-cookie"
export WANDERLOG_XSRF="your-xsrf-token"

# Or pass as flags (not recommended for security)
wanderlog create --title "New Trip" --session "cookie" --xsrf "token"
```

**Security Features:**
- 🔐 **Secure Storage**: Credentials are stored in your system keychain (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux)
- 🔄 **Automatic Loading**: Once logged in, credentials are automatically used for all write operations
- 🗑️ **Easy Logout**: Clear stored credentials with `wanderlog logout`
- ✅ **Status Check**: Verify authentication status with `wanderlog status`

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
        StartDate: "2024-06-01",
        EndDate: "2024-06-07",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created trip: %s (ID: %s)\n", 
        newTrip.TripPlan.Title, 
        newTrip.TripPlan.Key)
}
```

## Configuration

The CLI supports configuration via:
- Config file: `~/.wanderlog.yaml`
- Environment variables (prefixed with `WANDERLOG_`)
- Command-line flags

Example config file:
```yaml
verbose: true
format: pretty
```

## Example Output

```bash
$ wanderlog trip --file trips/trip1.json --details

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

📍 Beijing (Pechino)
   Sunday, Oct 12, 2025
   📝 Arriviamo al 13.10 pomeriggio

📍 Pechino
   Monday, Oct 13, 2025
   📝 Opzioni hotel: Sunworld Hotel
```

## LLM Integration

The `--format markdown` option produces clean, structured Markdown perfect for feeding to Large Language Models:

```bash
# Generate trip analysis for an LLM
wanderlog trip abc123xyz --format markdown --details > trip.md

# Get places data for AI processing
wanderlog places abc123xyz --format markdown > places.md
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
- 🔄 **Collaboration** - share trips, invite collaborators
- 🔄 **Budget tracking** - expenses and costs management 
- 🔄 **Batch operations** - bulk editing with operational transforms
- 🔄 **Interactive mode** - explore trips interactively with TUI
- 🔄 **Export features** - PDF, HTML, calendar formats
- 🔄 **Search and filtering** - find specific places or dates
- 🔄 **Trip analytics** - distance, duration, cost analysis

## Security

The CLI implements secure credential storage using your system's native keychain:

- **macOS**: Keychain Access
- **Windows**: Windows Credential Manager  
- **Linux**: Secret Service (GNOME Keyring, KDE Wallet, etc.)

Your login credentials (email/password) are **never stored**. Only session tokens are securely stored for convenience. You can always run `wanderlog logout` to clear stored credentials.

## Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o wanderlog

# Run locally
./wanderlog trip abc123xyz
```

## Project Structure

```
├── cmd/                 # CLI commands (Cobra)
│   ├── root.go         # Root command setup
│   └── trip.go         # Trip command implementation
├── pkg/
│   ├── wanderlog/      # Core API client package
│   │   ├── client.go   # HTTP client for Wanderlog API
│   │   ├── models.go   # Generated Go structs from JSON
│   │   └── wanderlog.go # Package documentation
│   └── ui/             # Terminal UI formatting
│       └── trip.go     # Beautiful trip output formatting
├── trips/              # Example trip data
│   └── trip1.json     # Sample trip for development
├── main.go            # CLI entry point
└── go.mod            # Go module definition
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