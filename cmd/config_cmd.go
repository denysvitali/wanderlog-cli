package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		cfg, err := client.GetGlobalConfig()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch global config")
			os.Exit(1)
		}
		if len(cfg.Raw) > 0 {
			_, _ = os.Stdout.Write(cfg.Raw)
			if cfg.Raw[len(cfg.Raw)-1] != '\n' {
				fmt.Println()
			}
			return
		}
		ui.PrintJSON(cfg)
	},
}

var configSessionGetCmd = &cobra.Command{
	Use:   "session",
	Short: "Fetch the authenticated session store",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetSessionStore()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch session store")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var configSessionSetCmd = &cobra.Command{
	Use:   "session-set [key]",
	Short: "Write a value to the authenticated session store",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(sessionSetValue) == "" {
			logger.Error("--value is required")
			os.Exit(1)
		}
		var value any
		if err := json.Unmarshal([]byte(sessionSetValue), &value); err != nil {
			value = sessionSetValue
		}
		client := newClient(true)
		if err := client.SetSessionStoreValue(args[0], value); err != nil {
			logger.WithError(err).Error("Failed to write session value")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Wrote session key %s", args[0]), map[string]interface{}{"key": args[0]})
	},
}

var configSessionPreferencesCmd = &cobra.Command{
	Use:   "preferences",
	Short: "Fetch locale-scoped session preferences",
	Run: func(cmd *cobra.Command, args []string) {
		if sessionLocale == "" {
			sessionLocale = "en"
		}
		client := newClient(false)
		resp, err := client.GetSessionPreferences(sessionLocale)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch session preferences")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGlobalCmd, configSessionGetCmd, configSessionSetCmd, configSessionPreferencesCmd)

	configSessionSetCmd.Flags().StringVar(&sessionSetValue, "value", "", "JSON value, or a bare string if not valid JSON")
	configSessionPreferencesCmd.Flags().StringVar(&sessionLocale, "locale", "en", "Locale code")

	for _, command := range []*cobra.Command{configGlobalCmd, configSessionGetCmd, configSessionSetCmd, configSessionPreferencesCmd} {
		command.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
