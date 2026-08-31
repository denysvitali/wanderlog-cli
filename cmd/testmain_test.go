package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain redirects the plaintext credential fallback away from the
// developer's real credentials file: cmd tests disable the keychain, which
// makes every LoadCredentials call read the fallback file.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "wanderlog-cmd-creds-*")
	if err == nil {
		defer os.RemoveAll(dir)
		os.Setenv("WANDERLOG_CREDENTIALS_FILE", filepath.Join(dir, "credentials.json"))
	}
	os.Exit(m.Run())
}
