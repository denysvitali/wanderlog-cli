package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLegacyCommandsReuseCanonicalContracts(t *testing.T) {
	tests := []struct {
		path     string
		flagName string
		badArgs  []string
	}{
		{path: "trip", flagName: "file", badArgs: nil},
		{path: "create", flagName: "title", badArgs: []string{"unexpected"}},
		{path: "delete", flagName: "yes", badArgs: nil},
		{path: "places", flagName: "file", badArgs: nil},
		{path: "search-places", flagName: "lat", badArgs: nil},
		{path: "place-details", flagName: "output", badArgs: nil},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			command, _, err := rootCmd.Find([]string{test.path})
			require.NoError(t, err)
			assert.True(t, command.Hidden)
			assert.NotNil(t, command.Flags().Lookup(test.flagName))
			require.NotNil(t, command.Args)
			assert.Error(t, command.Args(command, test.badArgs))
		})
	}
}

func TestConfigureOutputFormatUsesPerCommandDefault(t *testing.T) {
	original := outputFormat
	t.Cleanup(func() { outputFormat = original })

	command := &cobra.Command{Use: "test"}
	command.Flags().StringVar(&outputFormat, "output", "raw", "Output format (raw, json, pretty)")
	outputFormat = "pretty" // Simulate another command registering its default later.

	require.NoError(t, configureOutputFormat(command))
	assert.Equal(t, "raw", outputFormat)
}

func TestConfigureOutputFormatValidatesAndNormalizes(t *testing.T) {
	original := outputFormat
	t.Cleanup(func() { outputFormat = original })

	command := &cobra.Command{Use: "test"}
	command.Flags().StringVar(&outputFormat, "output", "pretty", "Output format (pretty, json, markdown)")
	require.NoError(t, command.Flags().Set("output", "md"))
	require.NoError(t, configureOutputFormat(command))
	assert.Equal(t, "markdown", outputFormat)

	require.NoError(t, command.Flags().Set("output", "xml"))
	err := configureOutputFormat(command)
	assert.EqualError(t, err, `invalid output format "xml" (valid values: pretty, json, markdown)`)
}

func TestConfirmAction(t *testing.T) {
	original := outputFormat
	t.Cleanup(func() { outputFormat = original })

	command := &cobra.Command{}
	command.SetIn(bytes.NewBufferString("yes\n"))
	var stderr bytes.Buffer
	command.SetErr(&stderr)
	outputFormat = "pretty"

	confirmed, err := confirmAction(command, "Delete it?", false)
	require.NoError(t, err)
	assert.True(t, confirmed)
	assert.Contains(t, stderr.String(), "Delete it?")

	outputFormat = "json"
	confirmed, err = confirmAction(command, "Delete it?", false)
	assert.False(t, confirmed)
	assert.EqualError(t, err, "--yes is required with --output json")

	confirmed, err = confirmAction(command, "Delete it?", true)
	require.NoError(t, err)
	assert.True(t, confirmed)
}

func TestAssignedCommandsReturnValidationErrors(t *testing.T) {
	originalQuery := tripsAutofillQuery
	originalDate := flightStopsDate
	originalAirline := flightStopsAirline
	t.Cleanup(func() {
		tripsAutofillQuery = originalQuery
		flightStopsDate = originalDate
		flightStopsAirline = originalAirline
	})

	tripsAutofillQuery = ""
	err := tripsAutofillCmd.RunE(tripsAutofillCmd, []string{"trip", "123"})
	assert.EqualError(t, err, "--query is required")

	flightStopsAirline = "MU"
	flightStopsDate = "not-a-date"
	err = travelFlightStopsCmd.RunE(travelFlightStopsCmd, []string{"244"})
	assert.EqualError(t, err, `invalid departure date "not-a-date": use YYYY-MM-DD`)
}
