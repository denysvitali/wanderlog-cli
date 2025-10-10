#!/bin/bash
# Integration test runner for Wanderlog CLI
# This script runs integration tests against the real Wanderlog API

set -e

# Check if credentials are set
if [ -z "$WANDERLOG_SESSION_COOKIE" ] || [ -z "$WANDERLOG_XSRF_TOKEN" ]; then
    echo "Error: WANDERLOG_SESSION_COOKIE and WANDERLOG_XSRF_TOKEN must be set"
    echo ""
    echo "To get these credentials:"
    echo "1. Log in to wanderlog.com in your browser"
    echo "2. Open Developer Tools (F12)"
    echo "3. Go to Application/Storage -> Cookies"
    echo "4. Copy the value of 'connect.sid' cookie to WANDERLOG_SESSION_COOKIE"
    echo "5. Look for the X-XSRF-TOKEN in the request headers"
    echo ""
    echo "Example:"
    echo "  export WANDERLOG_SESSION_COOKIE='s%3A...'"
    echo "  export WANDERLOG_XSRF_TOKEN='...'"
    echo "  ./test_integration.sh"
    exit 1
fi

# Set test trip ID (default provided if not set)
if [ -z "$WANDERLOG_TEST_TRIP_ID" ]; then
    echo "Using default test trip ID: vetyiadvqjgikbvx"
    echo "Set WANDERLOG_TEST_TRIP_ID to use a different trip"
fi

# Enable integration tests
export WANDERLOG_INTEGRATION_TEST=1

# Run the tests
echo "Running integration tests..."
go test -v -tags=integration -timeout 30m ./pkg/wanderlog -run TestIntegration

echo ""
echo "Integration tests completed!"
