#!/bin/bash
set -e

# Build the binary
echo "=== Building wanderlog binary ==="
cd "$(dirname "$0")/.."
go build -o wanderlog .

# Check authentication status
echo ""
echo "=== Checking authentication status ==="
AUTH_OUTPUT=$(./wanderlog status 2>&1)
if echo "$AUTH_OUTPUT" | grep -q "Authenticated"; then
    echo "✓ Authenticated with Wanderlog"
else
    echo "✗ Not authenticated with Wanderlog"
    echo ""
    echo "Please run './wanderlog login' first to authenticate."
    echo "Then run this script again."
    exit 1
fi

# Configuration
TRIP_TITLE="Japan Golden Week Adventure $(date +%Y-%m-%d)"
GEO_ID=86647  # Japan
START_DATE="2026-05-11"
END_DATE="2026-05-20"

echo ""
echo "=== Creating trip via CLI ==="
echo "Title: $TRIP_TITLE"
echo "Geo ID: $GEO_ID (Japan)"
echo "Dates: $START_DATE to $END_DATE"
echo ""

# Create the trip using CLI
TRIP_OUTPUT=$(./wanderlog trips create --title "$TRIP_TITLE" --geo-id $GEO_ID --start $START_DATE --end $END_DATE 2>&1)
echo "$TRIP_OUTPUT"

# Extract trip key - look for "key=" followed by the key (more specific pattern)
TRIP_KEY=$(echo "$TRIP_OUTPUT" | grep -oE 'key[^=]*=[ ]*[a-zA-Z0-9]+' | tail -1 | cut -d= -f2 | tr -d ' ')

if [[ -n "$TRIP_KEY" ]]; then
    echo ""
    echo "=== Trip created successfully! ==="
    echo "Trip Key: $TRIP_KEY"
    echo ""

    # Verify using CLI's verify-trip command
    echo "=== Verifying trip data using verify-trip command ==="
    ./wanderlog verify-trip "$TRIP_KEY" 2>&1 | head -100

    # Also verify via raw API
    echo ""
    echo "=== Raw API response (first 50 lines) ==="
    ./wanderlog api "tripPlans/${TRIP_KEY}?clientSchemaVersion=2" 2>&1 | head -50
else
    echo ""
    echo "=== Trip creation may have failed ==="
    echo "Output: $TRIP_OUTPUT"
fi

echo ""
echo "=== Script completed ==="
