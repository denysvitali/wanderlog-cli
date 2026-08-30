package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	userEmailLookup    string
	userBlockID        string
	userUsername       string
	userKVValue        string
	userUTCOffset      int
	userFollowingIDs   []string
	userNotifOffset    int
	userNotifIDs       []string
	userSettingsBody   string
	userUpdateName     string
	userUpdateUsername string
	userUpdateBio      string
	userUpdateLocation string
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if len(args) == 0 {
			profile, err := client.GetMe()
			if err != nil {
				return fmt.Errorf("get profile: %w", err)
			}
			return ui.PrintJSON(profile)
		}
		target := args[0]
		if strings.HasPrefix(target, "@") {
			resp, err := client.GetUserProfileByUsername(strings.TrimPrefix(target, "@"))
			if err != nil {
				return fmt.Errorf("get profile by username: %w", err)
			}
			return ui.PrintJSON(resp)
		}
		id, err := parseRequiredIntE(target, "user ID")
		if err != nil {
			return err
		}
		resp, err := client.GetUserProfile(id)
		if err != nil {
			return fmt.Errorf("get profile: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the authenticated user's profile",
	Long: `Update one or more profile fields. Passing an empty value explicitly
clears bio or location.

Examples:
  wanderlog user update --name "Ada Lovelace" --location London
  wanderlog user update --bio ""`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("username") &&
			!cmd.Flags().Changed("bio") && !cmd.Flags().Changed("location") {
			return fmt.Errorf("at least one of --name, --username, --bio, or --location is required")
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		profile, err := client.UpdateMe(wanderlog.UpdateUserRequest{
			Name:          userUpdateName,
			Username:      userUpdateUsername,
			Bio:           userUpdateBio,
			Location:      userUpdateLocation,
			ClearBio:      cmd.Flags().Changed("bio") && userUpdateBio == "",
			ClearLocation: cmd.Flags().Changed("location") && userUpdateLocation == "",
		})
		if err != nil {
			return fmt.Errorf("update profile: %w", err)
		}
		return ui.PrintJSON(profile)
	},
}

var userNotificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "List the authenticated user's notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetNotifications(userNotifOffset)
		if err != nil {
			return fmt.Errorf("list notifications: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userNotificationsMarkReadCmd = &cobra.Command{
	Use:   "mark-read",
	Short: "Mark notifications as read",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(userNotifIDs) == 0 {
			return fmt.Errorf("at least one --id is required")
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.MarkNotificationsRead(userNotifIDs); err != nil {
			return fmt.Errorf("mark notifications read: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Marked %d notification(s) read", len(userNotifIDs)), map[string]interface{}{"ids": userNotifIDs})
	},
}

var userSettingsGetCmd = &cobra.Command{
	Use:   "settings",
	Short: "Get the authenticated user's notification settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetNotificationSettings()
		if err != nil {
			return fmt.Errorf("get notification settings: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userSettingsSetCmd = &cobra.Command{
	Use:   "settings-set",
	Short: "Replace the authenticated user's notification settings",
	Long: `Replace notification settings. The --body JSON becomes the value of
"notificationSettings" in the POST payload.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(userSettingsBody) == "" {
			return fmt.Errorf("--body is required")
		}
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(userSettingsBody), &raw); err != nil {
			return fmt.Errorf("invalid --body JSON: %w", err)
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.UpdateNotificationSettings(raw)
		if err != nil {
			return fmt.Errorf("update notification settings: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userKVGetCmd = &cobra.Command{
	Use:   "kv-get [key]",
	Short: "Read a value from the authenticated user's key-value store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		value, err := client.GetKeyValue(args[0])
		if err != nil {
			return fmt.Errorf("get key-value: %w", err)
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(value))
		return err
	},
}

var userKVSetCmd = &cobra.Command{
	Use:   "kv-set [key]",
	Short: "Write a value to the authenticated user's key-value store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(userKVValue) == "" {
			return fmt.Errorf("--value is required")
		}
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(userKVValue), &raw); err != nil {
			raw = json.RawMessage(fmt.Sprintf("%q", userKVValue))
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.SetKeyValue(args[0], raw); err != nil {
			return fmt.Errorf("set key-value: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Wrote %s", args[0]), map[string]interface{}{"key": args[0]})
	},
}

var userUTCOffsetCmd = &cobra.Command{
	Use:   "utc-offset",
	Short: "Persist the authenticated user's UTC offset (minutes)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.SetUTCOffset(userUTCOffset); err != nil {
			return fmt.Errorf("set UTC offset: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Set UTC offset to %d minutes", userUTCOffset), map[string]interface{}{"utcOffset": userUTCOffset})
	},
}

var userFollowingCmd = &cobra.Command{
	Use:   "following",
	Short: "Report whether the authenticated user follows each listed userId",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(userFollowingIDs) == 0 {
			return fmt.Errorf("at least one --user-id is required")
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.ListFollowing(userFollowingIDs)
		if err != nil {
			return fmt.Errorf("list following: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Autocomplete Wanderlog users by name prefix",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.AutocompleteUsers(args[0])
		if err != nil {
			return fmt.Errorf("search users: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userByEmailCmd = &cobra.Command{
	Use:   "by-email",
	Short: "Look up a user by email",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userEmailLookup == "" {
			return fmt.Errorf("--email is required")
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.FindUserByEmail(userEmailLookup)
		if err != nil {
			return fmt.Errorf("find user by email: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block a Wanderlog user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userBlockID == "" {
			return fmt.Errorf("--user-id is required")
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.BlockUser(userBlockID); err != nil {
			return fmt.Errorf("block user: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Blocked user %s", userBlockID), map[string]interface{}{"userId": userBlockID})
	},
}

var userUsernameTakenCmd = &cobra.Command{
	Use:   "username-taken",
	Short: "Check whether a username is already taken",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userUsername == "" {
			return fmt.Errorf("--username is required")
		}
		client, err := newClientE(false)
		if err != nil {
			return err
		}
		taken, err := client.IsUsernameTaken(userUsername)
		if err != nil {
			return fmt.Errorf("check username: %w", err)
		}
		return ui.PrintJSON(map[string]interface{}{"username": userUsername, "taken": taken})
	},
}

var userEmailsCmd = &cobra.Command{
	Use:   "emails",
	Short: "List the authenticated user's registered email addresses",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetUserEmails()
		if err != nil {
			return fmt.Errorf("get emails: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var userLogoutServerCmd = &cobra.Command{
	Use:   "server-logout",
	Short: "Invalidate the current session on the server (keeps local creds unless --clear)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.ServerLogout(); err != nil {
			return fmt.Errorf("server logout: %w", err)
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
		return printSuccess(outputFormat, "Server session invalidated", map[string]interface{}{"cleared": clearLocal})
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(
		userProfileCmd,
		userUpdateCmd,
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
	userUpdateCmd.Flags().StringVar(&userUpdateName, "name", "", "Display name")
	userUpdateCmd.Flags().StringVar(&userUpdateUsername, "username", "", "Username")
	userUpdateCmd.Flags().StringVar(&userUpdateBio, "bio", "", "Profile bio (pass an empty value to clear)")
	userUpdateCmd.Flags().StringVar(&userUpdateLocation, "location", "", "Profile location (pass an empty value to clear)")

	for _, command := range []*cobra.Command{
		userProfileCmd,
		userUpdateCmd,
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
		command.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
