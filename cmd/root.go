package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	logger        *logrus.Logger
	verbose       bool
	outputFormat  string
	showDetails   bool
	fromFile      string
	sessionCookie string
	xsrfToken     string
)

var rootCmd = &cobra.Command{
	Use:   "wanderlog",
	Short: "A beautiful CLI for interacting with Wanderlog trip data",
	Long: `Wanderlog CLI is a tool for fetching and displaying trip planning data
from Wanderlog in a beautiful, easy-to-read format.

You can view trip details, itineraries, places, and more directly from your terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		setupLogging()
		return configureOutputFormat(cmd)
	},
	SilenceUsage: true,
}

// configureOutputFormat reads the value from the command's own flag. Many
// commands intentionally have different defaults (for example, `api` defaults
// to raw while the human-facing commands default to pretty), so relying on a
// package-global variable's initialization order produces surprising results.
func configureOutputFormat(cmd *cobra.Command) error {
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		return nil
	}

	format := flag.DefValue
	if flag.Changed {
		var err error
		format, err = cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "md" {
		format = "markdown"
	}

	allowed := map[string]bool{"pretty": true, "json": true}
	if strings.Contains(flag.Usage, "markdown") {
		allowed["markdown"] = true
	}
	if strings.Contains(flag.Usage, "raw") {
		allowed["raw"] = true
	}
	if !allowed[format] {
		values := []string{"pretty", "json"}
		if allowed["markdown"] {
			values = append(values, "markdown")
		}
		if allowed["raw"] {
			values = append(values, "raw")
		}
		return fmt.Errorf("invalid output format %q (valid values: %s)", format, strings.Join(values, ", "))
	}

	outputFormat = format
	return nil
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := ExecuteContext(ctx)
	stop()
	if err != nil {
		os.Exit(1)
	}
}

// ExecuteContext runs the root command with ctx. Commands pass this context to
// client request methods so cancellation interrupts in-flight HTTP calls.
func ExecuteContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("execute: nil context")
	}
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/wanderlog/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Use XDG config directory
		configDir := filepath.Join(xdg.ConfigHome, "wanderlog")

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("WANDERLOG")

	if err := viper.ReadInConfig(); err == nil {
		logrus.WithField("config", viper.ConfigFileUsed()).Debug("Using config file")
		changed, migrationErr := wanderlog.MigrateLegacyConfig()
		if migrationErr != nil {
			logrus.WithError(migrationErr).Warn("Failed to remove legacy plaintext credentials from config")
		} else if changed {
			if reloadErr := viper.ReadInConfig(); reloadErr != nil {
				logrus.WithError(reloadErr).Warn("Failed to reload migrated config")
			} else {
				logrus.Warn("Removed a legacy plaintext password from the Wanderlog config")
			}
		}
		// Never expose a legacy config-file password to authentication code, even
		// if migration could not rewrite the file.
		viper.Set("auth.password", "")
	}
}

func setupLogging() {
	logger = logrus.New()

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})
}
