package wanderlog

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestMain redirects the plaintext credential fallback away from the
// developer's real credentials file: tests that exercise SaveCredentials
// with a working keychain remove the fallback file afterwards, and that
// must never delete real stored session tokens.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "wanderlog-creds-*")
	if err == nil {
		defer os.RemoveAll(dir)
		os.Setenv("WANDERLOG_CREDENTIALS_FILE", filepath.Join(dir, "credentials.json"))
	}
	os.Exit(m.Run())
}

func TestKeychainUnavailableClassification(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"not found", keyring.ErrNotFound, false},
		{"unsupported platform", keyring.ErrUnsupportedPlatform, true},
		{"missing dbus-launch", errors.New(`exec: "dbus-launch": executable file not found in $PATH`), true},
		{"dbus transport failure", errors.New("dbus: connection closed by user"), true},
		{"secret service missing", errors.New("The name org.freedesktop.secrets was not provided by any .service files"), true},
		{"unrelated failure", errors.New("keychain locked"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := keychainUnavailable(tc.err); got != tc.want {
				t.Fatalf("keychainUnavailable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestPlaintextFallbackRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	t.Setenv("WANDERLOG_CREDENTIALS_FILE", path)
	t.Setenv("WANDERLOG_DISABLE_KEYCHAIN", "1")

	if HasStoredCredentials() {
		t.Fatal("expected no stored credentials before save")
	}

	creds := &AuthCredentials{SessionCookie: "session-plain", XSRFToken: "xsrf-plain", UserID: "17"}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("plaintext fallback file missing: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("plaintext fallback permissions = %o, want 600", perm)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading plaintext fallback: %v", err)
	}
	if !strings.Contains(string(data), "session-plain") {
		t.Fatalf("plaintext fallback does not contain the session token: %s", data)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if loaded == nil || *loaded != *creds {
		t.Fatalf("loaded credentials = %+v, want %+v", loaded, creds)
	}
	if !HasStoredCredentials() {
		t.Fatal("HasStoredCredentials returned false after saving")
	}

	if err := DeleteCredentials(); err != nil {
		t.Fatalf("DeleteCredentials: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("plaintext fallback still present after delete: %v", err)
	}
	loaded, err = LoadCredentials()
	if err != nil || loaded != nil {
		t.Fatalf("expected credentials gone, got %+v, %v", loaded, err)
	}
}

func TestUnavailableKeychainFallsBackToPlaintext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	t.Setenv("WANDERLOG_CREDENTIALS_FILE", path)
	t.Setenv("WANDERLOG_DISABLE_KEYCHAIN", "")
	keyring.MockInitWithError(errors.New(`exec: "dbus-launch": executable file not found in $PATH`))
	t.Cleanup(keyring.MockInit)

	creds := &AuthCredentials{SessionCookie: "session-dbus", XSRFToken: "xsrf-dbus"}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials with unavailable keychain: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("plaintext fallback file missing: %v", err)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials with unavailable keychain: %v", err)
	}
	if loaded == nil || *loaded != *creds {
		t.Fatalf("loaded credentials = %+v, want %+v", loaded, creds)
	}

	if err := DeleteCredentials(); err != nil {
		t.Fatalf("DeleteCredentials with unavailable keychain: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("plaintext fallback still present after delete: %v", err)
	}
}

func TestKeychainSaveRemovesStalePlaintextCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	t.Setenv("WANDERLOG_CREDENTIALS_FILE", path)
	t.Setenv("WANDERLOG_DISABLE_KEYCHAIN", "")
	keyring.MockInit()
	t.Cleanup(keyring.MockInit)

	if err := saveCredentialsFile([]byte(`{"SessionCookie":"stale"}`)); err != nil {
		t.Fatalf("seeding stale plaintext credentials: %v", err)
	}
	if err := SaveCredentials(&AuthCredentials{SessionCookie: "fresh", XSRFToken: "fresh-xsrf"}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("stale plaintext credentials survived a successful keychain save")
	}
	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if loaded == nil || loaded.SessionCookie != "fresh" {
		t.Fatalf("loaded credentials = %+v, want the keychain copy", loaded)
	}
}
