package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Wanderlog",
	Long: `Login to Wanderlog to enable trip editing and creation features.

Your credentials are used to obtain a session token which is securely stored 
in the system keychain for future use.

Examples:
  wanderlog login
  wanderlog login --email user@example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		
		if email == "" {
			fmt.Print("Email: ")
			fmt.Scanln(&email)
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			logger.WithError(err).Error("Failed to read password")
			os.Exit(1)
		}
		fmt.Println() // New line after password input

		password := string(passwordBytes)

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		creds, err := client.Login(email, password)
		if err != nil {
			logger.WithError(err).Error("Login failed")
			os.Exit(1)
		}

		// Store credentials in both keychain and config file
		keychainErr := wanderlog.SaveCredentials(creds)
		configErr := wanderlog.SaveCredentialsToConfig(creds, email, password)

		if keychainErr != nil && configErr != nil {
			logger.WithError(keychainErr).Warn("Failed to save credentials to keychain")
			logger.WithError(configErr).Warn("Failed to save credentials to config file")
			fmt.Printf("⚠️ Credentials saved in memory only (this session)\n")
		} else {
			if keychainErr == nil {
				fmt.Printf("🔐 Credentials saved to keychain\n")
			}
			if configErr == nil {
				fmt.Printf("📝 Credentials saved to config file\n")
			}
		}

		fmt.Printf("✅ Successfully logged in!\n")
		fmt.Printf("Session: %s...\n", creds.SessionCookie[:20])
		fmt.Printf("User ID: %s\n", creds.UserID)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored authentication credentials",
	Long: `Remove stored authentication credentials from the system keychain.

This will require you to login again before performing write operations.

Examples:
  wanderlog logout`,
	Run: func(cmd *cobra.Command, args []string) {
		keychainErr := wanderlog.DeleteCredentials()
		configErr := wanderlog.ClearCredentialsFromConfig()

		if keychainErr != nil && configErr != nil {
			logger.WithError(keychainErr).Error("Failed to clear credentials from keychain")
			logger.WithError(configErr).Error("Failed to clear credentials from config")
			os.Exit(1)
		}

		fmt.Printf("✅ Successfully logged out\n")
		if keychainErr == nil {
			fmt.Printf("🗑️ Credentials cleared from keychain\n")
		}
		if configErr == nil {
			fmt.Printf("🗑️ Credentials cleared from config file\n")
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long: `Check if you are currently authenticated and show session information.

Examples:
  wanderlog status`,
	Run: func(cmd *cobra.Command, args []string) {
		if wanderlog.HasStoredCredentials() {
			creds, err := wanderlog.LoadCredentials()
			if err != nil {
				logger.WithError(err).Error("Failed to load credentials")
				os.Exit(1)
			}
			fmt.Printf("✅ Authenticated\n")
			fmt.Printf("Session: %s...\n", creds.SessionCookie[:20])
			fmt.Printf("User ID: %s\n", creds.UserID)
		} else {
			fmt.Printf("❌ Not authenticated\n")
			fmt.Printf("Run 'wanderlog login' to authenticate\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)
	loginCmd.Flags().String("email", "", "Email address for login")
}