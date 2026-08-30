package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
)

var (
	sessionSetValue string
	sessionLocale   string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect Wanderlog server configuration and session store",
}

var configGlobalCmd = &cobra.Command{
	Use:   "global",
	Short: "Fetch the server's global configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(false)
		if err != nil {
			return err
		}
		cfg, err := client.GetGlobalConfig()
		if err != nil {
			return fmt.Errorf("fetch global config: %w", err)
		}
		if len(cfg.Raw) > 0 {
			if _, err := cmd.OutOrStdout().Write(cfg.Raw); err != nil {
				return err
			}
			if cfg.Raw[len(cfg.Raw)-1] != '\n' {
				_, err = fmt.Fprintln(cmd.OutOrStdout())
			}
			return err
		}
		return ui.PrintJSON(cfg)
	},
}

var configSessionGetCmd = &cobra.Command{
	Use:   "session",
	Short: "Fetch the authenticated session store",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetSessionStore()
		if err != nil {
			return fmt.Errorf("fetch session store: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var configSessionSetCmd = &cobra.Command{
	Use:   "session-set [key]",
	Short: "Write a value to the authenticated session store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(sessionSetValue) == "" {
			return fmt.Errorf("--value is required")
		}
		var value any
		if err := json.Unmarshal([]byte(sessionSetValue), &value); err != nil {
			value = sessionSetValue
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		if err := client.SetSessionStoreValue(args[0], value); err != nil {
			return fmt.Errorf("write session value: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Wrote session key %s", args[0]), map[string]interface{}{"key": args[0]})
	},
}

var configSessionPreferencesCmd = &cobra.Command{
	Use:   "preferences",
	Short: "Fetch locale-scoped session preferences",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sessionLocale == "" {
			sessionLocale = "en"
		}
		client, err := newClientE(false)
		if err != nil {
			return err
		}
		resp, err := client.GetSessionPreferences(sessionLocale)
		if err != nil {
			return fmt.Errorf("fetch session preferences: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGlobalCmd, configSessionGetCmd, configSessionSetCmd, configSessionPreferencesCmd)

	configSessionSetCmd.Flags().StringVar(&sessionSetValue, "value", "", "JSON value, or a bare string if not valid JSON")
	configSessionPreferencesCmd.Flags().StringVar(&sessionLocale, "locale", "en", "Locale code")

	for _, command := range []*cobra.Command{configGlobalCmd, configSessionGetCmd, configSessionSetCmd, configSessionPreferencesCmd} {
		command.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
