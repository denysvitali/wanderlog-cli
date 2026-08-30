package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandValidationErrorsPropagate(t *testing.T) {
	originalPlaceName := tripsEditPlaceName
	originalSettingsBody := userSettingsBody
	originalSessionValue := sessionSetValue
	originalAPIBody := apiBody
	originalAPIBodyFile := apiBodyFile
	originalTitle := tripsUpdateTitle
	originalStart := updateStartDate
	originalEnd := updateEndDate
	originalPrivacy := updatePrivacy
	t.Cleanup(func() {
		tripsEditPlaceName = originalPlaceName
		userSettingsBody = originalSettingsBody
		sessionSetValue = originalSessionValue
		apiBody = originalAPIBody
		apiBodyFile = originalAPIBodyFile
		tripsUpdateTitle = originalTitle
		updateStartDate = originalStart
		updateEndDate = originalEnd
		updatePrivacy = originalPrivacy
	})

	tripsEditPlaceName = ""
	assert.EqualError(t, tripsEditAddPlaceCmd.RunE(tripsEditAddPlaceCmd, []string{"trip"}), "place name is required (--name)")

	userSettingsBody = "{"
	assert.ErrorContains(t, userSettingsSetCmd.RunE(userSettingsSetCmd, nil), "invalid --body JSON")

	sessionSetValue = ""
	assert.EqualError(t, configSessionSetCmd.RunE(configSessionSetCmd, []string{"key"}), "--value is required")

	apiBody = "{"
	apiBodyFile = ""
	assert.ErrorContains(t, apiCmd.RunE(apiCmd, []string{"/api/test"}), "invalid JSON body")

	tripsUpdateTitle = ""
	updateStartDate = "tomorrow"
	updateEndDate = ""
	updatePrivacy = ""
	assert.EqualError(t, tripsUpdateCmd.RunE(tripsUpdateCmd, []string{"trip"}), `invalid start date "tomorrow": use YYYY-MM-DD`)
}

func TestErrorReturningParsersRejectInvalidInput(t *testing.T) {
	_, err := parseIntCSVE("1,nope,3", "place IDs")
	assert.EqualError(t, err, `invalid place IDs "nope": must be a number`)

	_, err = parseOptionalIntCSVE("12,-4")
	assert.EqualError(t, err, `invalid user ID "-4" in comma-separated list: must be greater than zero`)

	_, err = parseJSONBody(`{"broken"`)
	assert.ErrorContains(t, err, "invalid JSON body")
}
