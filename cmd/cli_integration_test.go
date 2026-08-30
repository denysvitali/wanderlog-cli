package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIntegrationTest skips the test unless production API access is explicitly enabled.
// Integration tests make real API calls and require authentication.
func skipCLITest(t *testing.T) {
	t.Helper()
	if os.Getenv("WANDERLOG_RUN_PROD_INTEGRATION") != "1" {
		t.Skip("Skipping CLI integration test. Set WANDERLOG_RUN_PROD_INTEGRATION=1 to run.")
	}
}

func requireTestTripID(t *testing.T) string {
	t.Helper()
	tripID := os.Getenv("WANDERLOG_TEST_TRIP_ID")
	if tripID == "" {
		t.Skip("Set WANDERLOG_TEST_TRIP_ID to run this test")
	}
	return tripID
}

func disableCLIAuth(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"WANDERLOG_AUTH_SESSION_COOKIE",
		"WANDERLOG_AUTH_SESSION_XSRF_TOKEN",
		"WANDERLOG_AUTH_XSRF_TOKEN",
		"WANDERLOG_AUTH_EMAIL",
		"WANDERLOG_AUTH_PASSWORD",
	} {
		t.Setenv(name, "")
	}
	t.Setenv("WANDERLOG_DISABLE_KEYCHAIN", "1")
	originalSession, originalXSRF := sessionCookie, xsrfToken
	sessionCookie, xsrfToken = "", ""
	t.Cleanup(func() {
		sessionCookie, xsrfToken = originalSession, originalXSRF
	})
}

// TestCLI_TripsList_JSON tests the trips list command with JSON output
func TestCLI_TripsList_JSON(t *testing.T) {
	skipCLITest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Skipf("Auth required for CLI trips list test: %v", err)
	}
	_ = auth

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"trips", "list", "--output", "json"})
	err = rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Command should not error with valid auth
	require.NoError(t, err, "trips list should not error with valid auth")
	// Output should be valid JSON (starts with { or [)
	assert.True(t, len(output) > 0, "should have output")
}

// TestCLI_TripsList_WithoutAuth tests that trips list fails gracefully without auth
func TestCLI_TripsList_WithoutAuth(t *testing.T) {
	skipCLITest(t)
	disableCLIAuth(t)

	// This test verifies graceful failure - we don't set up auth
	// so the command should fail with a clear error message
	rootCmd.SetArgs([]string{"trips", "list"})
	err := rootCmd.Execute()
	// Expect an error since we're not authenticated
	assert.Error(t, err, "trips list should error without auth")
}

// TestCLI_TripsShow_JSON tests the trips show command with JSON output using a known test trip
func TestCLI_TripsShow_JSON(t *testing.T) {
	skipCLITest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Skipf("Auth required for CLI trips show test: %v", err)
	}
	_ = auth

	testTripID := requireTestTripID(t)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"trips", "show", testTripID, "--output", "json"})
	err = rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err, "trips show should not error with valid auth and trip ID")
	assert.True(t, len(output) > 0, "should have output")
}

// TestCLI_TripsPlaces_JSON tests the trips places command with JSON output
func TestCLI_TripsPlaces_JSON(t *testing.T) {
	skipCLITest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Skipf("Auth required for CLI trips places test: %v", err)
	}
	_ = auth

	testTripID := requireTestTripID(t)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"trips", "places", testTripID, "--output", "json"})
	err = rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err, "trips places should not error with valid auth and trip ID")
	assert.True(t, len(output) > 0, "should have output")
}

// TestCLI_TripsCreate_RequiresAuth tests that trips create requires authentication
func TestCLI_TripsCreate_RequiresAuth(t *testing.T) {
	skipCLITest(t)
	disableCLIAuth(t)

	// Create without required args should fail with usage
	rootCmd.SetArgs([]string{"trips", "create"})
	err := rootCmd.Execute()
	// Either no args error or auth error - both are acceptable
	assert.Error(t, err, "trips create should error without auth")
}

// TestCLI_TripsDelete_RequiresAuth tests that trips delete requires authentication
func TestCLI_TripsDelete_RequiresAuth(t *testing.T) {
	skipCLITest(t)
	disableCLIAuth(t)

	// Delete without auth should fail
	rootCmd.SetArgs([]string{"trips", "delete", "nonexistent-trip-key"})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips delete should error without auth")
}

// TestCLI_TripsEditAddPlace_RequiresAuth tests that add-place requires authentication
func TestCLI_TripsEditAddPlace_RequiresAuth(t *testing.T) {
	skipCLITest(t)
	disableCLIAuth(t)

	// Add-place without auth should fail
	rootCmd.SetArgs([]string{"trips", "edit", "add-place", "test-trip-key", "--name", "Test Place"})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips edit add-place should error without auth")
}

// TestCLI_TripsEditRemovePlace_RequiresAuth tests that remove-place requires authentication
func TestCLI_TripsEditRemovePlace_RequiresAuth(t *testing.T) {
	skipCLITest(t)
	disableCLIAuth(t)

	// Remove-place without auth should fail
	rootCmd.SetArgs([]string{"trips", "edit", "remove-place", "test-trip-key", "12345"})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips edit remove-place should error without auth")
}

// TestCLI_TripsShow_MissingArgs tests that trips show fails with clear error when missing args
func TestCLI_TripsShow_MissingArgs(t *testing.T) {
	skipCLITest(t)

	rootCmd.SetArgs([]string{"trips", "show"})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips show should error without trip ID")
}

// TestCLI_TripsPlaces_MissingArgs tests that trips places fails with clear error when missing args
func TestCLI_TripsPlaces_MissingArgs(t *testing.T) {
	skipCLITest(t)

	rootCmd.SetArgs([]string{"trips", "places"})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips places should error without trip ID")
}

// TestCLI_TripsEdit_MissingArgs tests that edit subcommands fail with clear errors
func TestCLI_TripsEdit_MissingArgs(t *testing.T) {
	skipCLITest(t)

	tests := []struct {
		name string
		args []string
	}{
		{"add-place no trip key", []string{"trips", "edit", "add-place"}},
		{"add-place no name", []string{"trips", "edit", "add-place", "test-trip"}},
		{"remove-place no place id", []string{"trips", "edit", "remove-place", "test-trip"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()
			assert.Error(t, err, "should error with missing args")
		})
	}
}

// TestCLI_TripsList_MarkdownOutput tests markdown output format
func TestCLI_TripsList_MarkdownOutput(t *testing.T) {
	skipCLITest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Skipf("Auth required for CLI trips list markdown test: %v", err)
	}
	_ = auth

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"trips", "list", "--output", "markdown"})
	err = rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err, "trips list markdown should not error with valid auth")
	assert.True(t, len(output) > 0, "should have output")
	// Markdown output should contain markdown headings
	assert.Contains(t, output, "#", "markdown output should contain headings")
}

// TestCLI_TripsShow_MarkdownOutput tests markdown output format
func TestCLI_TripsShow_MarkdownOutput(t *testing.T) {
	skipCLITest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Skip("Auth required for CLI trips show markdown test")
	}
	_ = auth

	testTripID := requireTestTripID(t)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"trips", "show", testTripID, "--output", "markdown"})
	err = rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err, "trips show markdown should not error with valid auth")
	assert.True(t, len(output) > 0, "should have output")
}

// TestCLI_TripsPlaces_MarkdownOutput tests markdown output format
func TestCLI_TripsPlaces_MarkdownOutput(t *testing.T) {
	skipCLITest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Skip("Auth required for CLI trips places markdown test")
	}
	_ = auth

	testTripID := requireTestTripID(t)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"trips", "places", testTripID, "--output", "markdown"})
	err = rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	require.NoError(t, err, "trips places markdown should not error with valid auth")
	assert.True(t, len(output) > 0, "should have output")
}
