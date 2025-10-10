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
	if _, err := os.Stat(configPath); err != nil {
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

	return nil
}

// ConfigAuth represents the auth section of the config file
type ConfigAuth struct {
	Email    string        `yaml:"email,omitempty"`
	Password string        `yaml:"password,omitempty"`
	Session  ConfigSession `yaml:"session,omitempty"`
}

// ConfigSession represents session credentials in the config file
type ConfigSession struct {
	Cookie    string `yaml:"cookie,omitempty"`
	XSRFToken string `yaml:"xsrf_token,omitempty"`
	UserID    string `yaml:"user_id,omitempty"`
}

// Config represents the entire config file structure
type Config struct {
	Auth ConfigAuth `yaml:"auth,omitempty"`
}

// SaveCredentialsToConfig saves authentication credentials to the config file
// If email/password are provided, it saves them along with the session tokens
// This allows automatic re-login when session tokens expire
func SaveCredentialsToConfig(creds *AuthCredentials, email, password string) error {
	if creds == nil {
		return fmt.Errorf("credentials cannot be nil")
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

	// Read existing config or create new one
	var config Config
	if _, err := os.Stat(configPath); err == nil {
		// Config file exists, read it
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("reading config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("parsing config file: %w", err)
		}
	}

	// Update config with new credentials
	config.Auth.Session.Cookie = creds.SessionCookie
	config.Auth.Session.XSRFToken = creds.XSRFToken
	config.Auth.Session.UserID = creds.UserID

	// Save email/password if provided
	if email != "" {
		config.Auth.Email = email
	}
	if password != "" {
		config.Auth.Password = password
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// ClearCredentialsFromConfig removes credentials from the config file
func ClearCredentialsFromConfig() error {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		// No config file, nothing to clear
		return nil
	}

	// Read existing config
	var config Config
	if _, err := os.Stat(configPath); err != nil {
		// Config file doesn't exist
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	// Clear auth credentials
	config.Auth = ConfigAuth{}

	// Marshal to YAML
	data, err = yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
