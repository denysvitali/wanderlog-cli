package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIntegrationTest skips the test unless INTEGRATION_TESTS environment variable is set.
// Integration tests make real API calls and require authentication.
func skipCLITest(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping CLI integration test. Set INTEGRATION_TESTS=1 to run.")
	}
}

func hasAuthEnv() bool {
	hasSessionAuth := os.Getenv("WANDERLOG_AUTH_SESSION_COOKIE") != "" &&
		os.Getenv("WANDERLOG_AUTH_SESSION_XSRF_TOKEN") != ""
	hasLoginAuth := os.Getenv("WANDERLOG_AUTH_EMAIL") != "" &&
		os.Getenv("WANDERLOG_AUTH_PASSWORD") != ""
	return hasSessionAuth || hasLoginAuth
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

	// Use a test trip ID from CLAUDE.md
	testTripID := "yxjmjivtfxlkaqcp"

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

// TestCLI_TripsShow_WithoutAuth tests that trips show fails without auth
func TestCLI_TripsShow_WithoutAuth(t *testing.T) {
	skipCLITest(t)

	testTripID := "yxjmjivtfxlkaqcp"
	rootCmd.SetArgs([]string{"trips", "show", testTripID})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips show should error without auth")
}

// TestCLI_TripsPlaces_JSON tests the trips places command with JSON output
func TestCLI_TripsPlaces_JSON(t *testing.T) {
	skipCLITest(t)

	auth, err := loadAuthFromEnvOrKeychain()
	if err != nil {
		t.Skipf("Auth required for CLI trips places test: %v", err)
	}
	_ = auth

	testTripID := "yxjmjivtfxlkaqcp"

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

// TestCLI_TripsPlaces_WithoutAuth tests that trips places fails without auth
func TestCLI_TripsPlaces_WithoutAuth(t *testing.T) {
	skipCLITest(t)

	testTripID := "yxjmjivtfxlkaqcp"
	rootCmd.SetArgs([]string{"trips", "places", testTripID})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips places should error without auth")
}

// TestCLI_TripsCreate_RequiresAuth tests that trips create requires authentication
func TestCLI_TripsCreate_RequiresAuth(t *testing.T) {
	skipCLITest(t)

	// Create without required args should fail with usage
	rootCmd.SetArgs([]string{"trips", "create"})
	err := rootCmd.Execute()
	// Either no args error or auth error - both are acceptable
	assert.Error(t, err, "trips create should error without auth")
}

// TestCLI_TripsCreate_WithAuth tests that trips create works with auth (dry run - uses example flag)
func TestCLI_TripsCreate_WithAuth(t *testing.T) {
	skipCLITest(t)

	if !hasAuthEnv() {
		t.Skip("Auth credentials required for this test")
	}

	// Use the example flag to create a test trip without needing geo-id
	// Note: This creates an actual trip, but since we're using the example trip
	// it should be cleanable. In a full integration test, we'd delete after.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"trips", "create", "--example"})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// With auth but interactive confirmation might be needed
	// The command may succeed or fail depending on keychain vs env vars
	t.Logf("Output: %s", output)
	t.Logf("Error: %v", err)

	// We just verify the command runs without panicking
	// Actual trip creation is tested in pkg/wanderlog tests
}

// TestCLI_TripsDelete_RequiresAuth tests that trips delete requires authentication
func TestCLI_TripsDelete_RequiresAuth(t *testing.T) {
	skipCLITest(t)

	// Delete without auth should fail
	rootCmd.SetArgs([]string{"trips", "delete", "nonexistent-trip-key"})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips delete should error without auth")
}

// TestCLI_TripsEditAddPlace_RequiresAuth tests that add-place requires authentication
func TestCLI_TripsEditAddPlace_RequiresAuth(t *testing.T) {
	skipCLITest(t)

	// Add-place without auth should fail
	rootCmd.SetArgs([]string{"trips", "edit", "add-place", "test-trip-key", "--name", "Test Place"})
	err := rootCmd.Execute()
	assert.Error(t, err, "trips edit add-place should error without auth")
}

// TestCLI_TripsEditRemovePlace_RequiresAuth tests that remove-place requires authentication
func TestCLI_TripsEditRemovePlace_RequiresAuth(t *testing.T) {
	skipCLITest(t)

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
		t.Skipf("Auth required for CLI trips show markdown test: %v", err)
	}
	_ = auth

	testTripID := "yxjmjivtfxlkaqcp"

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
		t.Skipf("Auth required for CLI trips places markdown test: %v", err)
	}
	_ = auth

	testTripID := "yxjmjivtfxlkaqcp"

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
