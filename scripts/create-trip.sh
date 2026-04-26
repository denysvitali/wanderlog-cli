#!/bin/bash
set -e

# Build the binary
echo "=== Building wanderlog binary ==="
cd "$(dirname "$0")/.."
go build -o wanderlog .

# Check authentication status
echo ""
echo "=== Checking authentication status ==="
AUTH_STATUS=$(./wanderlog status --format json 2>/dev/null || echo '{"authenticated": false}')
if echo "$AUTH_STATUS" | jq -e '.authenticated == true' >/dev/null 2>&1; then
    echo "✓ Authenticated with Wanderlog"
else
    echo "✗ Not authenticated with Wanderlog"
    echo ""
    echo "Please run './wanderlog login' first to authenticate."
    echo "Then run this script again."
    exit 1
fi

# Configuration
PORT=3789
TRIP_TITLE="Japan Golden Week Adventure $(date +%Y-%m-%d)"
GEO_ID=86647  # Japan
START_DATE="2026-05-11"
END_DATE="2026-05-20"
MCP_PID=""

# Cleanup function
cleanup() {
    if [[ -n "$MCP_PID" ]]; then
        echo ""
        echo "=== Stopping MCP server (PID: $MCP_PID) ==="
        kill "$MCP_PID" 2>/dev/null || true
        wait "$MCP_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Start MCP server in background
echo ""
echo "=== Starting MCP server on port $PORT (read-write mode) ==="
./wanderlog mcp --enable-write --http ":$PORT" &
MCP_PID=$!

# Wait for server to be ready
echo "=== Waiting for MCP server to be ready ==="
for i in {1..30}; do
    if curl -s "http://localhost:$PORT/mcp" >/dev/null 2>&1; then
        echo "MCP server is ready!"
        break
    fi
    if ! kill -0 "$MCP_PID" 2>/dev/null; then
        echo "MCP server process died!"
        exit 1
    fi
    sleep 0.5
done

echo ""
echo "=== Creating trip via MCP ==="
echo "Title: $TRIP_TITLE"
echo "Geo ID: $GEO_ID (Japan)"
echo "Dates: $START_DATE to $END_DATE"
echo ""

# JSON-RPC request to create the trip
RESPONSE=$(curl -s -X POST "http://localhost:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -d "{
        \"jsonrpc\": \"2.0\",
        \"id\": 1,
        \"method\": \"tools/call\",
        \"params\": {
            \"name\": \"create_trip\",
            \"arguments\": {
                \"title\": \"$TRIP_TITLE\",
                \"geo_id\": $GEO_ID,
                \"start_date\": \"$START_DATE\",
                \"end_date\": \"$END_DATE\"
            }
        }
    }")

echo "=== MCP Response ==="
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Check if trip was created successfully by looking for the trip key in the response
TRIP_KEY=$(echo "$RESPONSE" | jq -r '.result.content[0].text // empty' 2>/dev/null | grep -oE '[a-zA-Z0-9]{10,}' | head -1 || echo "")

if [[ -n "$TRIP_KEY" ]]; then
    echo ""
    echo "=== Trip created successfully! ==="
    echo "Trip Key: $TRIP_KEY"
    echo ""
    echo "=== Verifying trip data via MCP ==="

    # Verify the trip
    VERIFY_RESPONSE=$(curl -s -X POST "http://localhost:$PORT/mcp" \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"id\": 2,
            \"method\": \"tools/call\",
            \"params\": {
                \"name\": \"get_trip\",
                \"arguments\": {
                    \"trip_id\": \"$TRIP_KEY\"
                }
            }
        }")

    echo "$VERIFY_RESPONSE" | jq '.' 2>/dev/null || echo "$VERIFY_RESPONSE"

    echo ""
    echo "=== Trip Sections ==="
    SECTIONS_RESPONSE=$(curl -s -X POST "http://localhost:$PORT/mcp" \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"id\": 3,
            \"method\": \"tools/call\",
            \"params\": {
                \"name\": \"list_sections\",
                \"arguments\": {
                    \"trip_id\": \"$TRIP_KEY\"
                }
            }
        }")
    echo "$SECTIONS_RESPONSE" | jq '.' 2>/dev/null || echo "$SECTIONS_RESPONSE"
else
    echo ""
    echo "=== Trip creation may have failed ==="
    echo "Response: $RESPONSE"
fi

echo ""
echo "=== Script completed ==="
