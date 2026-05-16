package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommandDefined(t *testing.T) {
	assert.Equal(t, "wanderlog", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
}

func TestRootCommandHasPersistentFlags(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, flag, "expected --config flag")

	flag = rootCmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, flag, "expected --verbose flag")
}

func TestMajorSubcommandsRegistered(t *testing.T) {
	majorCmds := []string{
		"login", "logout", "status",
		"trips", "mcp", "api",
		"user", "feed", "config", "travel",
		"verify-trip",
	}

	for _, name := range majorCmds {
		_, _, err := rootCmd.Find([]string{name})
		assert.NoError(t, err, "expected subcommand %q to be registered", name)
	}
}

func TestMCPServerCommandFlags(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"mcp"})
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, cmd.Flags().Lookup("http"), "expected --http flag")
	assert.NotNil(t, cmd.Flags().Lookup("enable-write"), "expected --enable-write flag")
	assert.NotNil(t, cmd.Flags().Lookup("trip-id"), "expected --trip-id flag")
}

func TestLoginCommandHasFlags(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"login"})
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, cmd.Flags().Lookup("email"), "expected --email flag")
}

func TestTripsSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"trips"})
	if !assert.NoError(t, err) {
		return
	}

	tripSubs := []string{"show", "places", "list", "create", "delete", "copy", "restore",
		"edit", "invite", "collaborator", "share-key", "autofill", "checklist",
		"flights", "export", "like", "like-count", "sections",
		"journal", "expenses", "register-view", "update-required", "distinction", "create-guide"}
	for _, name := range tripSubs {
		sub, _, err := cmd.Find([]string{name})
		assert.NoError(t, err, "expected trips %q subcommand", name)
		assert.NotNil(t, sub)
	}
}

func TestUserSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"user"})
	if !assert.NoError(t, err) {
		return
	}

	userSubs := []string{
		"profile", "notifications", "mark-read",
		"settings", "settings-set",
		"kv-get", "kv-set",
		"utc-offset",
		"search", "by-email", "following", "block",
		"username-taken", "emails",
	}
	for _, name := range userSubs {
		sub, _, err := cmd.Find([]string{name})
		assert.NoError(t, err, "expected user %q subcommand", name)
		assert.NotNil(t, sub)
	}
}

func TestTravelSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"travel"})
	if !assert.NoError(t, err) {
		return
	}

	travelSubs := []string{"airlines", "airports", "flight-stops", "hotels", "hotel-rates"}
	for _, name := range travelSubs {
		sub, _, err := cmd.Find([]string{name})
		assert.NoError(t, err, "expected travel %q subcommand", name)
		assert.NotNil(t, sub)
	}
}

func TestConfigSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"config"})
	if !assert.NoError(t, err) {
		return
	}

	configSubs := []string{"global", "session", "session-set"}
	for _, name := range configSubs {
		sub, _, err := cmd.Find([]string{name})
		assert.NoError(t, err, "expected config %q subcommand", name)
		assert.NotNil(t, sub)
	}
}

func TestFeedSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"feed"})
	if !assert.NoError(t, err) {
		return
	}

	feedSubs := []string{"home", "recent", "friends", "history", "guides", "legacy", "v2"}
	for _, name := range feedSubs {
		sub, _, err := cmd.Find([]string{name})
		assert.NoError(t, err, "expected feed %q subcommand", name)
		assert.NotNil(t, sub)
	}
}

func TestEditSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"trips", "edit"})
	if !assert.NoError(t, err) {
		return
	}

	editSubs := []string{"add-place", "remove-place"}
	for _, name := range editSubs {
		sub, _, err := cmd.Find([]string{name})
		assert.NoError(t, err, "expected trips edit %q subcommand", name)
		assert.NotNil(t, sub)
	}
}

func TestSearchSubcommands(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"search"})
	if !assert.NoError(t, err) {
		return
	}

	searchSubs := []string{"google", "wanderlog", "place-details"}
	for _, name := range searchSubs {
		sub, _, err := cmd.Find([]string{name})
		assert.NoError(t, err, "expected search %q subcommand", name)
		assert.NotNil(t, sub)
	}
}
