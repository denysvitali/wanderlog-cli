package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
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
			_, _ = fmt.Scanln(&email)
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
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
			fmt.Println(ui.WarningStyle.Render("⚠️ Credentials saved in memory only (this session)"))
		} else {
			if keychainErr == nil {
				fmt.Println(ui.SuccessStyle.Render("🔐 Credentials saved to keychain"))
			}
			if configErr == nil {
				fmt.Println(ui.SuccessStyle.Render("📝 Credentials saved to config file"))
			}
		}

		fmt.Println(ui.SuccessStyle.Render("✅ Successfully logged in!"))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Session: %s...", creds.SessionCookie[:20])))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("User ID: %s", creds.UserID)))
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

		fmt.Println(ui.SuccessStyle.Render("✅ Successfully logged out"))
		if keychainErr == nil {
			fmt.Println(ui.InfoStyle.Render("🗑️ Credentials cleared from keychain"))
		}
		if configErr == nil {
			fmt.Println(ui.InfoStyle.Render("🗑️ Credentials cleared from config file"))
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
		var creds *wanderlog.AuthCredentials
		var source string

		if wanderlog.HasStoredCredentials() {
			c, err := wanderlog.LoadCredentials()
			if err != nil {
				logger.WithError(err).Error("Failed to load credentials from keychain")
			} else {
				creds = c
				source = "keychain"
			}
		}

		if creds == nil && wanderlog.HasConfigCredentials() {
			c, err := wanderlog.LoadCredentialsFromConfig()
			if err != nil {
				logger.WithError(err).Error("Failed to load credentials from config file")
			} else {
				creds = c
				source = "config file"
			}
		}

		if creds != nil {
			fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✅ Authenticated (via %s)", source)))
			sessionDisplay := creds.SessionCookie
			if len(sessionDisplay) > 20 {
				sessionDisplay = sessionDisplay[:20]
			}
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Session: %s...", sessionDisplay)))
			if creds.UserID != "" {
				fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("User ID: %s", creds.UserID)))
			}
		} else {
			fmt.Println(ui.ErrorStyle.Render("❌ Not authenticated"))
			fmt.Println(ui.InfoStyle.Render("Run 'wanderlog login' to authenticate"))
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)
	loginCmd.Flags().String("email", "", "Email address for login")
}
