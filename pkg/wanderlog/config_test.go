//go:build integration
// +build integration

package wanderlog

import (
	"testing"

	"github.com/spf13/viper"
)

func TestConfigLoading(t *testing.T) {
	requireProductionIntegrationOptIn(t)

	// Initialize config
	if err := InitConfig(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Check what viper loaded
	t.Logf("Config file used: %s", viper.ConfigFileUsed())
	t.Logf("auth.session.cookie present: %v", viper.GetString("auth.session.cookie") != "")
	t.Logf("auth.session.xsrf_token present: %v", viper.GetString("auth.session.xsrf_token") != "")
	t.Logf("auth.session.user_id: %s", viper.GetString("auth.session.user_id"))
	t.Logf("auth.email present: %v", viper.GetString("auth.email") != "")
	t.Logf("auth.password: [REDACTED]")

	// Create client and test authentication
	client := NewClient()
	if err := client.EnsureAuthenticated("", ""); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	t.Log("✅ Successfully authenticated using config file!")
}
