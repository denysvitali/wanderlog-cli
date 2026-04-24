package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	userEmailLookup  string
	userBlockID      string
	userUsername     string
	userKVValue      string
	userUTCOffset    int
	userFollowingIDs []string
	userNotifOffset  int
	userNotifIDs     []string
	userSettingsBody string
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Query and manage the authenticated user",
}

var userProfileCmd = &cobra.Command{
	Use:   "profile [target]",
	Short: "Show a user's profile (defaults to the authenticated user)",
	Long: `Show a user profile. Target may be omitted to show the current user, a
numeric user ID, or @username.

Examples:
  wanderlog user profile
  wanderlog user profile 12345
  wanderlog user profile @someuser`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		if len(args) == 0 {
			profile, err := client.GetMe()
			if err != nil {
				logger.WithError(err).Error("Failed to get profile")
				os.Exit(1)
			}
			ui.PrintJSON(profile)
			return
		}
		target := args[0]
		if strings.HasPrefix(target, "@") {
			resp, err := client.GetUserProfileByUsername(strings.TrimPrefix(target, "@"))
			if err != nil {
				logger.WithError(err).Error("Failed to get profile by username")
				os.Exit(1)
			}
			ui.PrintJSON(resp)
			return
		}
		id := parseRequiredInt(target, "user ID")
		resp, err := client.GetUserProfile(id)
		if err != nil {
			logger.WithError(err).Error("Failed to get profile")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userNotificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "List the authenticated user's notifications",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetNotifications(userNotifOffset)
		if err != nil {
			logger.WithError(err).Error("Failed to list notifications")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userNotificationsMarkReadCmd = &cobra.Command{
	Use:   "mark-read",
	Short: "Mark notifications as read",
	Run: func(cmd *cobra.Command, args []string) {
		if len(userNotifIDs) == 0 {
			logger.Error("At least one --id is required")
			os.Exit(1)
		}
		client := newClient(true)
		if err := client.MarkNotificationsRead(userNotifIDs); err != nil {
			logger.WithError(err).Error("Failed to mark notifications read")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Marked %d notification(s) read", len(userNotifIDs)), map[string]interface{}{"ids": userNotifIDs})
	},
}

var userSettingsGetCmd = &cobra.Command{
	Use:   "settings",
	Short: "Get the authenticated user's notification settings",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetNotificationSettings()
		if err != nil {
			logger.WithError(err).Error("Failed to get notification settings")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userSettingsSetCmd = &cobra.Command{
	Use:   "settings-set",
	Short: "Replace the authenticated user's notification settings",
	Long: `Replace notification settings. The --body JSON becomes the value of
"notificationSettings" in the POST payload.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(userSettingsBody) == "" {
			logger.Error("--body is required")
			os.Exit(1)
		}
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(userSettingsBody), &raw); err != nil {
			logger.WithError(err).Error("Invalid --body JSON")
			os.Exit(1)
		}
		client := newClient(true)
		resp, err := client.UpdateNotificationSettings(raw)
		if err != nil {
			logger.WithError(err).Error("Failed to update notification settings")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userKVGetCmd = &cobra.Command{
	Use:   "kv-get [key]",
	Short: "Read a value from the authenticated user's key-value store",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		value, err := client.GetKeyValue(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to get key-value")
			os.Exit(1)
		}
		fmt.Println(string(value))
	},
}

var userKVSetCmd = &cobra.Command{
	Use:   "kv-set [key]",
	Short: "Write a value to the authenticated user's key-value store",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(userKVValue) == "" {
			logger.Error("--value is required")
			os.Exit(1)
		}
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(userKVValue), &raw); err != nil {
			raw = json.RawMessage(fmt.Sprintf("%q", userKVValue))
		}
		client := newClient(true)
		if err := client.SetKeyValue(args[0], raw); err != nil {
			logger.WithError(err).Error("Failed to set key-value")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Wrote %s", args[0]), map[string]interface{}{"key": args[0]})
	},
}

var userUTCOffsetCmd = &cobra.Command{
	Use:   "utc-offset",
	Short: "Persist the authenticated user's UTC offset (minutes)",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		if err := client.SetUTCOffset(userUTCOffset); err != nil {
			logger.WithError(err).Error("Failed to set UTC offset")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Set UTC offset to %d minutes", userUTCOffset), map[string]interface{}{"utcOffset": userUTCOffset})
	},
}

var userFollowingCmd = &cobra.Command{
	Use:   "following",
	Short: "Report whether the authenticated user follows each listed userId",
	Run: func(cmd *cobra.Command, args []string) {
		if len(userFollowingIDs) == 0 {
			logger.Error("At least one --user-id is required")
			os.Exit(1)
		}
		client := newClient(true)
		resp, err := client.ListFollowing(userFollowingIDs)
		if err != nil {
			logger.WithError(err).Error("Failed to list following")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Autocomplete Wanderlog users by name prefix",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.AutocompleteUsers(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to search users")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userByEmailCmd = &cobra.Command{
	Use:   "by-email",
	Short: "Look up a user by email",
	Run: func(cmd *cobra.Command, args []string) {
		if userEmailLookup == "" {
			logger.Error("--email is required")
			os.Exit(1)
		}
		client := newClient(true)
		resp, err := client.FindUserByEmail(userEmailLookup)
		if err != nil {
			logger.WithError(err).Error("Failed to find user by email")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block a Wanderlog user",
	Run: func(cmd *cobra.Command, args []string) {
		if userBlockID == "" {
			logger.Error("--user-id is required")
			os.Exit(1)
		}
		client := newClient(true)
		if err := client.BlockUser(userBlockID); err != nil {
			logger.WithError(err).Error("Failed to block user")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Blocked user %s", userBlockID), map[string]interface{}{"userId": userBlockID})
	},
}

var userUsernameTakenCmd = &cobra.Command{
	Use:   "username-taken",
	Short: "Check whether a username is already taken",
	Run: func(cmd *cobra.Command, args []string) {
		if userUsername == "" {
			logger.Error("--username is required")
			os.Exit(1)
		}
		client := newClient(false)
		taken, err := client.IsUsernameTaken(userUsername)
		if err != nil {
			logger.WithError(err).Error("Failed to check username")
			os.Exit(1)
		}
		ui.PrintJSON(map[string]interface{}{"username": userUsername, "taken": taken})
	},
}

var userEmailsCmd = &cobra.Command{
	Use:   "emails",
	Short: "List the authenticated user's registered email addresses",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetUserEmails()
		if err != nil {
			logger.WithError(err).Error("Failed to get emails")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var userLogoutServerCmd = &cobra.Command{
	Use:   "server-logout",
	Short: "Invalidate the current session on the server (keeps local creds unless --clear)",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		if err := client.ServerLogout(); err != nil {
			logger.WithError(err).Error("Server logout failed")
			os.Exit(1)
		}
		clearLocal, _ := cmd.Flags().GetBool("clear")
		if clearLocal {
			if err := wanderlog.DeleteCredentials(); err != nil {
				logger.WithError(err).Warn("Failed to clear keychain credentials")
			}
			if err := wanderlog.ClearCredentialsFromConfig(); err != nil {
				logger.WithError(err).Warn("Failed to clear config credentials")
			}
		}
		printSuccess(outputFormat, "Server session invalidated", map[string]interface{}{"cleared": clearLocal})
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(
		userProfileCmd,
		userNotificationsCmd,
		userNotificationsMarkReadCmd,
		userSettingsGetCmd,
		userSettingsSetCmd,
		userKVGetCmd,
		userKVSetCmd,
		userUTCOffsetCmd,
		userFollowingCmd,
		userSearchCmd,
		userByEmailCmd,
		userBlockCmd,
		userUsernameTakenCmd,
		userEmailsCmd,
		userLogoutServerCmd,
	)

	userNotificationsCmd.Flags().IntVar(&userNotifOffset, "offset", 0, "Pagination offset")
	userNotificationsMarkReadCmd.Flags().StringArrayVar(&userNotifIDs, "id", nil, "Notification ID to mark read; may be repeated")
	userSettingsSetCmd.Flags().StringVar(&userSettingsBody, "body", "", "Raw JSON object for notificationSettings")
	userKVSetCmd.Flags().StringVar(&userKVValue, "value", "", "JSON value, or a bare string if not valid JSON")
	userUTCOffsetCmd.Flags().IntVar(&userUTCOffset, "minutes", 0, "Offset in minutes from UTC")
	userFollowingCmd.Flags().StringArrayVar(&userFollowingIDs, "user-id", nil, "User ID to check; may be repeated")
	userByEmailCmd.Flags().StringVar(&userEmailLookup, "email", "", "Email address to look up")
	userBlockCmd.Flags().StringVar(&userBlockID, "user-id", "", "User ID to block")
	userUsernameTakenCmd.Flags().StringVar(&userUsername, "username", "", "Username to check")
	userLogoutServerCmd.Flags().Bool("clear", false, "Also clear locally stored credentials")

	for _, command := range []*cobra.Command{
		userProfileCmd,
		userNotificationsCmd,
		userNotificationsMarkReadCmd,
		userSettingsGetCmd,
		userSettingsSetCmd,
		userKVGetCmd,
		userKVSetCmd,
		userUTCOffsetCmd,
		userFollowingCmd,
		userSearchCmd,
		userByEmailCmd,
		userBlockCmd,
		userUsernameTakenCmd,
		userEmailsCmd,
		userLogoutServerCmd,
	} {
		command.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
