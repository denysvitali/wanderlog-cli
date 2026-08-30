# Testing Guide

The default test suite is hermetic: it must not contact wanderlog.com or depend
on credentials. Production integration tests have a separate build/run path and
require an explicit opt-in.

## Local quality checks

```bash
make fmt-check
make vet
make test
make test-race
make coverage-check
make integration-compile
make build
```

`make coverage-check` writes `coverage.out` and `coverage.html`, then enforces
the current project-wide statement coverage floor of 35%. Raise
`COVERAGE_MIN` locally when validating a coverage improvement:

```bash
make coverage-check COVERAGE_MIN=35
```

Run `make vulncheck` after installing the pinned scanner used by CI:

```bash
go install golang.org/x/vuln/cmd/govulncheck@v1.7.0
make vulncheck
```

## Production integration tests

These tests call the real Wanderlog service and some tests create, update, or
delete data. Use a dedicated test account and a disposable trip. Never point
`WANDERLOG_TEST_TRIP_ID` at data you care about.

There is one opt-in switch for every production integration suite:

```bash
export WANDERLOG_RUN_PROD_INTEGRATION=1
export WANDERLOG_TEST_TRIP_ID='your-disposable-trip-id'
```

Provide exactly one complete authentication method.

Session credentials:

```bash
export WANDERLOG_AUTH_SESSION_COOKIE='...'
export WANDERLOG_AUTH_SESSION_XSRF_TOKEN='...'
```

Or login credentials:

```bash
export WANDERLOG_AUTH_EMAIL='test-account@example.com'
export WANDERLOG_AUTH_PASSWORD='...'
```

Then run both the client and command integration suites:

```bash
./test_integration.sh
```

The runner fails before testing when the opt-in, trip ID, credential pair, or
expected test names are missing. This prevents a green result caused by every
test silently skipping.

The production runner never mutates the desktop keychain. Keychain storage has
a separate destructive opt-in (`WANDERLOG_RUN_KEYCHAIN_INTEGRATION=1`) and
restores any credential it replaced; do not enable it on shared CI runners.

To compile all files guarded by the `integration` build tag without contacting
the service, run:

```bash
make integration-compile
```

To run an individual tagged test after exporting the same opt-in and
credentials:

```bash
go test -count=1 -v -tags=integration -timeout 10m \
  ./pkg/wanderlog -run '^TestIntegration_CreateAndDeleteTrip$'
```

Snapshot tests are intentionally outside the default runner because they are
long-running scenarios. Invoke them explicitly:

```bash
SAVE_SNAPSHOTS=1 go test -count=1 -v -tags=integration -timeout 30m \
  ./pkg/wanderlog -run '^TestSnapshotTrip$'
```

## CI behavior

Every push and pull request checks formatting, module tidiness, vet, lint, unit
tests, race detection, coverage, known vulnerabilities, integration-suite
compilation, a versioned binary, and the non-root container image.

Live production tests run only on branch pushes when the repository variable
`WANDERLOG_RUN_PROD_INTEGRATION` is set to `1`. The repository must also define
`WANDERLOG_TEST_TRIP_ID` and one complete credential pair as Actions secrets.
If live testing is enabled but its preconditions are incomplete, CI fails
instead of reporting a false success.

Tags beginning with `v` trigger GoReleaser. Releases contain checksummed
Linux, macOS, and Windows archives, per-archive SPDX SBOMs, injected version
metadata, and GitHub build-provenance attestations.
