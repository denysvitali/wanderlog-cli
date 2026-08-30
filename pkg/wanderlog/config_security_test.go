package wanderlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSaveCredentialsToConfigNeverPersistsPassword(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "wanderlog", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	initial := "theme: dark\nauth:\n  password: legacy-secret\n  custom: keep-me\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	creds := &AuthCredentials{SessionCookie: "session", XSRFToken: "xsrf", UserID: "42"}
	if err := SaveCredentialsToConfig(creds, "user@example.com", "new-secret"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, secret := range []string{"legacy-secret", "new-secret", "password:"} {
		if strings.Contains(text, secret) {
			t.Fatalf("config contains forbidden password material %q: %s", secret, text)
		}
	}
	for _, preserved := range []string{"theme: dark", "custom: keep-me", "cookie: session", "xsrf_token: xsrf"} {
		if !strings.Contains(text, preserved) {
			t.Fatalf("config did not preserve %q: %s", preserved, text)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("config permissions = %o, want 600", got)
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".config-*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary config files left behind: %v", matches)
	}
}

func TestMigrateLegacyConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("auth:\n  email: user@example.com\n  password: legacy-secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	changed, err := MigrateLegacyConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected legacy password migration")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "legacy-secret") || strings.Contains(string(data), "password:") {
		t.Fatalf("password was not removed: %s", data)
	}
}
