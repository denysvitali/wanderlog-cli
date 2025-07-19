# Wanderlog CLI

A beautiful command-line interface for interacting with Wanderlog trip data. Built with Go using Cobra, Viper, Logrus, and Charm's Lipgloss for stunning terminal output.

## Features

- 🌍 Fetch trip details from Wanderlog API
- 📅 Display trip dates, duration, and statistics
- 📍 Show detailed itineraries with places and locations
- ✈️ Flight information display
- 🎨 Beautiful, colorized terminal output
- 📊 JSON output support for scripting
- 🔧 Configurable logging and output formats
- 📦 Usable as both CLI tool and Go package

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

### Command Line

```bash
# Get trip overview
wanderlog trip abc123xyz

# Get detailed trip information
wanderlog trip abc123xyz --details

# Output as JSON
wanderlog trip abc123xyz --format json

# Enable verbose logging
wanderlog trip abc123xyz --verbose
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
    
    trip, err := client.GetTrip("abc123xyz")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Trip: %s\n", trip.TripPlan.Title)
    fmt.Printf("Duration: %s to %s\n", 
        trip.TripPlan.StartDate, 
        trip.TripPlan.EndDate)
    fmt.Printf("Places: %d\n", trip.TripPlan.PlaceCount)
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

## API Coverage

Currently supports:
- ✅ Trip metadata (title, dates, statistics)
- ✅ Daily itineraries with places
- ✅ Flight information
- ✅ Place details with Google Places data
- ✅ Budget information (in data model)

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

## License

MIT License - see LICENSE file for details.