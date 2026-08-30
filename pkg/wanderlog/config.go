package wanderlog

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// InitConfig initializes Viper configuration from the standard config file location
// This is primarily for use in tests that need to load configuration
func InitConfig() error {
	// Use XDG config directory
	configDir := filepath.Join(xdg.ConfigHome, "wanderlog")
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, that's okay
		return nil
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("WANDERLOG")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	if _, err := MigrateLegacyConfig(); err != nil {
		return err
	}
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reloading migrated config file: %w", err)
	}

	return nil
}

// ConfigAuth represents the auth section of the config file
type ConfigAuth struct {
	Email    string         `yaml:"email,omitempty"`
	Password string         `yaml:"password,omitempty"`
	Session  ConfigSession  `yaml:"session,omitempty"`
	Extra    map[string]any `yaml:",inline"`
}

// ConfigSession represents session credentials in the config file
type ConfigSession struct {
	Cookie    string         `yaml:"cookie,omitempty"`
	XSRFToken string         `yaml:"xsrf_token,omitempty"`
	UserID    string         `yaml:"user_id,omitempty"`
	Extra     map[string]any `yaml:",inline"`
}

// Config represents the entire config file structure
type Config struct {
	Auth  ConfigAuth     `yaml:"auth,omitempty"`
	Extra map[string]any `yaml:",inline"`
}

// SaveCredentialsToConfig saves a legacy config-file session fallback. The
// password parameter remains for source compatibility but is deliberately
// ignored: account passwords are never persisted.
func SaveCredentialsToConfig(creds *AuthCredentials, email, _ string) error {
	if err := creds.Validate(); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	// Determine config file path
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		// No config file loaded, create one in XDG config directory
		configDir := filepath.Join(xdg.ConfigHome, "wanderlog")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}
		configPath = filepath.Join(configDir, "config.yaml")
	}

	config, err := readConfig(configPath)
	if err != nil {
		return err
	}

	// Update config with new credentials
	config.Auth.Session.Cookie = creds.SessionCookie
	config.Auth.Session.XSRFToken = creds.XSRFToken
	config.Auth.Session.UserID = creds.UserID

	// Email is not secret and is useful for identifying the account. Always
	// remove any password written by older versions while updating the session.
	if email != "" {
		config.Auth.Email = email
	}
	config.Auth.Password = ""

	// Marshal to YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return writeConfigAtomically(configPath, data)
}

// HasConfigCredentials checks if credentials are stored in the config file
func HasConfigCredentials() bool {
	return viper.GetString("auth.session.cookie") != "" && viper.GetString("auth.session.xsrf_token") != ""
}

// LoadCredentialsFromConfig loads credentials from the config file
func LoadCredentialsFromConfig() (*AuthCredentials, error) {
	sessionCookie := viper.GetString("auth.session.cookie")
	xsrfToken := viper.GetString("auth.session.xsrf_token")
	userID := viper.GetString("auth.session.user_id")

	creds := &AuthCredentials{
		SessionCookie: sessionCookie,
		XSRFToken:     xsrfToken,
		UserID:        userID,
	}
	if err := creds.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config credentials: %w", err)
	}
	return creds, nil
}

// ClearCredentialsFromConfig removes credentials from the config file
func ClearCredentialsFromConfig() error {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		// No config file, nothing to clear
		return nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist
		return nil
	}

	config, err := readConfig(configPath)
	if err != nil {
		return err
	}

	// Clear auth credentials
	config.Auth = ConfigAuth{}

	// Marshal to YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return writeConfigAtomically(configPath, data)
}

// MigrateLegacyConfig removes plaintext passwords written by older releases
// and repairs config-file permissions. It returns whether a password was
// removed. Session tokens are retained as a compatibility fallback.
func MigrateLegacyConfig() (bool, error) {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		return false, nil
	}
	info, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stating config file: %w", err)
	}
	config, err := readConfig(configPath)
	if err != nil {
		return false, err
	}
	removedPassword := config.Auth.Password != ""
	config.Auth.Password = ""
	if removedPassword {
		data, marshalErr := yaml.Marshal(&config)
		if marshalErr != nil {
			return false, fmt.Errorf("marshaling config: %w", marshalErr)
		}
		if err := writeConfigAtomically(configPath, data); err != nil {
			return false, err
		}
	} else if info.Mode().Perm() != 0600 {
		if err := os.Chmod(configPath, 0600); err != nil {
			return false, fmt.Errorf("securing config file permissions: %w", err)
		}
	}
	return removedPassword, nil
}

func readConfig(configPath string) (Config, error) {
	var config Config
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return config, fmt.Errorf("reading config file: %w", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("parsing config file: %w", err)
	}
	return config, nil
}

func writeConfigAtomically(configPath string, data []byte) (retErr error) {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	temp, err := os.CreateTemp(configDir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary config file: %w", err)
	}
	tempPath := temp.Name()
	defer func() {
		if retErr != nil {
			_ = os.Remove(tempPath)
		}
	}()
	if err := temp.Chmod(0600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("securing temporary config file: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("writing temporary config file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("syncing temporary config file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("closing temporary config file: %w", err)
	}
	if err := os.Rename(tempPath, configPath); err != nil {
		return fmt.Errorf("replacing config file: %w", err)
	}
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("securing config file permissions: %w", err)
	}
	return nil
}
