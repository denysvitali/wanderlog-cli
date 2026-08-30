#!/usr/bin/env bash

# Runs the production integration suites against wanderlog.com. The explicit
# opt-in and credential preflight prevent accidental writes and false-green CI.
set -euo pipefail

if [[ ${WANDERLOG_RUN_PROD_INTEGRATION:-} != "1" ]]; then
    echo "Refusing to run production tests without WANDERLOG_RUN_PROD_INTEGRATION=1" >&2
    exit 1
fi

session_cookie=${WANDERLOG_AUTH_SESSION_COOKIE:-}
session_xsrf=${WANDERLOG_AUTH_SESSION_XSRF_TOKEN:-}
auth_email=${WANDERLOG_AUTH_EMAIL:-}
auth_password=${WANDERLOG_AUTH_PASSWORD:-}

if [[ -n $session_cookie || -n $session_xsrf ]]; then
    if [[ -z $session_cookie ]]; then
        echo "Session auth requires WANDERLOG_AUTH_SESSION_COOKIE (the XSRF token is optional since Wanderlog stopped issuing it)" >&2
        exit 1
    fi
    has_session_auth=1
else
    has_session_auth=0
fi

if [[ -n $auth_email || -n $auth_password ]]; then
    if [[ -z $auth_email || -z $auth_password ]]; then
        echo "Login auth requires both WANDERLOG_AUTH_EMAIL and WANDERLOG_AUTH_PASSWORD" >&2
        exit 1
    fi
    has_login_auth=1
else
    has_login_auth=0
fi

if [[ $has_session_auth == 0 && $has_login_auth == 0 ]]; then
    echo "Set a complete session-token pair or email/password pair for production tests" >&2
    exit 1
fi

if [[ -z ${WANDERLOG_TEST_TRIP_ID:-} ]]; then
    echo "WANDERLOG_TEST_TRIP_ID must identify a disposable test trip" >&2
    exit 1
fi

pkg_pattern='^TestIntegration_'
cmd_pattern='^(TestCLI_|TestMCPIntegration_|TestIntegration_)'

if ! go test -tags=integration -list "$pkg_pattern" ./pkg/wanderlog | grep -q '^Test'; then
    echo "No pkg/wanderlog production integration tests matched; refusing a false-green run" >&2
    exit 1
fi

if ! go test -list "$cmd_pattern" ./cmd | grep -q '^Test'; then
    echo "No cmd production integration tests matched; refusing a false-green run" >&2
    exit 1
fi

echo "Running pkg/wanderlog production integration tests..."
go test -count=1 -v -tags=integration -timeout 30m ./pkg/wanderlog -run "$pkg_pattern"

echo "Running cmd production integration tests..."
go test -count=1 -v -timeout 30m ./cmd -run "$cmd_pattern"

echo "Production integration tests completed."
