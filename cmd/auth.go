package cmd

import (
	"errors"
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

Your credentials are used to obtain a session token which is stored in the
system keychain for future use. When no keychain backend is available (e.g.
headless Linux without D-Bus), the session token falls back to a chmod-0600
plaintext file. The account password is never stored anywhere.

Examples:
  wanderlog login
  wanderlog login --email user@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")

		if email == "" {
			fmt.Print("Email: ")
			_, _ = fmt.Scanln(&email)
		}

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("reading password: %w", err)
		}
		fmt.Println() // New line after password input

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		creds, err := client.LoginContext(cmd.Context(), email, string(passwordBytes))
		// Do not retain the password any longer than required for the login call.
		for i := range passwordBytes {
			passwordBytes[i] = 0
		}
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Only session tokens are persisted, in the keychain when one exists and
		// in the plaintext fallback otherwise. The account password is never
		// written anywhere.
		if err := wanderlog.SaveCredentials(creds); err != nil {
			return fmt.Errorf("login succeeded, but securely storing credentials failed: %w", err)
		}
		if err := wanderlog.ClearCredentialsFromConfig(); err != nil {
			logger.WithError(err).Warn("Logged in, but failed to remove legacy config credentials")
		}

		if _, statErr := os.Stat(wanderlog.CredentialsFilePath()); statErr == nil {
			fmt.Println(ui.WarningStyle.Render(fmt.Sprintf("🔐 Credentials saved to plaintext file: %s", wanderlog.CredentialsFilePath())))
		} else {
			fmt.Println(ui.SuccessStyle.Render("🔐 Credentials saved to keychain"))
		}
		fmt.Println(ui.SuccessStyle.Render("✅ Successfully logged in!"))
		fmt.Println(ui.InfoStyle.Render("Session: [redacted]"))
		fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("User ID: %s", creds.UserID)))
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored authentication credentials",
	Long: `Invalidate the current server session and remove stored authentication
credentials from the system keychain, the plaintext fallback file, and any
legacy config-file session.

This will require you to login again before performing write operations.

Examples:
  wanderlog logout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, _, loadErr := loadAuthCredentials()
		var serverErr error
		if creds != nil {
			client := wanderlog.NewClient()
			client.SetLogger(logger)
			client.SetAuth(creds)
			serverErr = client.ServerLogout()
		}

		// Local deletion is attempted even when server invalidation or one storage
		// backend fails; this minimizes the chance of leaving credentials behind.
		keychainErr := wanderlog.DeleteCredentials()
		configErr := wanderlog.ClearCredentialsFromConfig()

		if keychainErr != nil || configErr != nil {
			return errors.Join(loadErr, serverErr, keychainErr, configErr)
		}

		if serverErr != nil {
			fmt.Println(ui.SuccessStyle.Render("🗑️ Local credentials cleared"))
			logger.WithError(serverErr).Warn("Local credentials cleared, but the remote session could not be invalidated")
		} else {
			fmt.Println(ui.SuccessStyle.Render("✅ Successfully logged out"))
		}
		if loadErr != nil {
			logger.WithError(loadErr).Warn("Credentials could not be loaded before local deletion")
		}
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long: `Check if you are currently authenticated and show session information.

Examples:
  wanderlog status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, source, err := loadAuthCredentials()
		if err != nil {
			return err
		}
		if creds == nil {
			fmt.Println(ui.ErrorStyle.Render("❌ Not authenticated"))
			fmt.Println(ui.InfoStyle.Render("Run 'wanderlog login' to authenticate"))
			return fmt.Errorf("not authenticated")
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)
		client.SetAuth(creds)
		profile, err := client.GetMe()
		if err != nil {
			return fmt.Errorf("verifying %s: %w", source, err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("✅ Authenticated (verified via %s)", source)))
		fmt.Println(ui.InfoStyle.Render("Session: [redacted]"))
		if profile.ID != 0 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("User ID: %d", profile.ID)))
		} else if creds.UserID != "" {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("User ID: %s", creds.UserID)))
		}
		return nil
	},
}

func loadAuthCredentials() (*wanderlog.AuthCredentials, string, error) {
	creds, keychainErr := wanderlog.LoadCredentials()
	if creds != nil {
		return creds, "stored credentials", nil
	}
	if wanderlog.HasConfigCredentials() {
		creds, err := wanderlog.LoadCredentialsFromConfig()
		if err != nil {
			return nil, "", fmt.Errorf("loading credentials from config file: %w", err)
		}
		return creds, "legacy config file", nil
	}
	if keychainErr != nil {
		return nil, "", fmt.Errorf("loading credentials from keychain: %w", keychainErr)
	}
	return nil, "", nil
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)
	loginCmd.Flags().String("email", "", "Email address for login")
}
