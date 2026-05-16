package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var (
	getIfEditedBody  string
	distinctionValue string
)

var journalCmd = &cobra.Command{
	Use:   "journal [journal-key]",
	Short: "Fetch a published view-only journal",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		resp, err := client.GetViewOnlyJournal(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch journal")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var expensesCmd = &cobra.Command{
	Use:   "expenses [trip-key]",
	Short: "Download a trip's expenses as CSV",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		csv, err := client.GetTripExpensesCSV(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch expenses CSV")
			os.Exit(1)
		}
		_, _ = os.Stdout.Write(csv)
		if len(csv) > 0 && csv[len(csv)-1] != '\n' {
			fmt.Println()
		}
	},
}

var registerViewCmd = &cobra.Command{
	Use:   "register-view [trip-key]",
	Short: "Register a view on a shared trip",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		if err := client.RegisterTripView(args[0]); err != nil {
			logger.WithError(err).Error("Failed to register view")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Registered view on %s", args[0]), map[string]interface{}{"tripKey": args[0]})
	},
}

var updateRequiredCmd = &cobra.Command{
	Use:   "update-required [trip-key]",
	Short: "Check whether the client must upgrade for this trip's schema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		resp, err := client.GetTripUpdateRequired(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch updateRequired status")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var distinctionCmd = &cobra.Command{
	Use:   "distinction [trip-key]",
	Short: "Get or set the trip's distinction/badge",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(distinctionValue != "")
		if distinctionValue == "" {
			resp, err := client.GetTripDistinction(args[0])
			if err != nil {
				logger.WithError(err).Error("Failed to fetch distinction")
				os.Exit(1)
			}
			ui.PrintJSON(resp)
			return
		}
		if err := client.SetTripDistinction(args[0], distinctionValue); err != nil {
			logger.WithError(err).Error("Failed to set distinction")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Set distinction to %q", distinctionValue), map[string]interface{}{"tripKey": args[0], "distinction": distinctionValue})
	},
}

var createGuideCmd = &cobra.Command{
	Use:   "create-guide [trip-key]",
	Short: "Promote a trip plan into a published guide",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.CreateGuideFromTripPlan(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to create guide")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var getIfEditedCmd = &cobra.Command{
	Use:   "get-if-edited",
	Short: "Ask the server which trip plans changed since given revisions",
	Run: func(cmd *cobra.Command, args []string) {
		var req wanderlog.GetIfEditedRequest
		if getIfEditedBody == "" {
			logger.Error("--body is required (JSON: {\"tripPlans\":[{\"key\":\"...\",\"lastEditedAt\":\"...\"}]})")
			os.Exit(1)
		}
		if err := json.Unmarshal([]byte(getIfEditedBody), &req); err != nil {
			logger.WithError(err).Error("Invalid --body JSON")
			os.Exit(1)
		}
		client := newClient(true)
		resp, err := client.GetIfEdited(req)
		if err != nil {
			logger.WithError(err).Error("getIfEdited failed")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

func init() {
	// root registrations disabled - commands moved under `trips`
	// rootCmd.AddCommand(journalCmd, expensesCmd, registerViewCmd, updateRequiredCmd, distinctionCmd, createGuideCmd, getIfEditedCmd)

	distinctionCmd.Flags().StringVar(&distinctionValue, "set", "", "Set the distinction to this value (otherwise get)")
	getIfEditedCmd.Flags().StringVar(&getIfEditedBody, "body", "", "JSON request body")

	for _, command := range []*cobra.Command{journalCmd, expensesCmd, registerViewCmd, updateRequiredCmd, distinctionCmd, createGuideCmd, getIfEditedCmd} {
		command.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
