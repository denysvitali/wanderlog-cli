package cmd

import (
	"encoding/json"
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(false)
		if err != nil {
			return err
		}
		resp, err := client.GetViewOnlyJournal(args[0])
		if err != nil {
			return fmt.Errorf("fetch journal: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var expensesCmd = &cobra.Command{
	Use:   "expenses [trip-key]",
	Short: "Download a trip's expenses as CSV",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		csv, err := client.GetTripExpensesCSV(args[0])
		if err != nil {
			return fmt.Errorf("fetch expenses CSV: %w", err)
		}
		if _, err := cmd.OutOrStdout().Write(csv); err != nil {
			return err
		}
		if len(csv) > 0 && csv[len(csv)-1] != '\n' {
			_, err = fmt.Fprintln(cmd.OutOrStdout())
		}
		return err
	},
}

var registerViewCmd = &cobra.Command{
	Use:   "register-view [trip-key]",
	Short: "Register a view on a shared trip",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(false)
		if err != nil {
			return err
		}
		if err := client.RegisterTripView(args[0]); err != nil {
			return fmt.Errorf("register view: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Registered view on %s", args[0]), map[string]interface{}{"tripKey": args[0]})
	},
}

var updateRequiredCmd = &cobra.Command{
	Use:   "update-required [trip-key]",
	Short: "Check whether the client must upgrade for this trip's schema",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(false)
		if err != nil {
			return err
		}
		resp, err := client.GetTripUpdateRequired(args[0])
		if err != nil {
			return fmt.Errorf("fetch updateRequired status: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var distinctionCmd = &cobra.Command{
	Use:   "distinction [trip-key]",
	Short: "Get or set the trip's distinction/badge",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(distinctionValue != "")
		if err != nil {
			return err
		}
		if distinctionValue == "" {
			resp, err := client.GetTripDistinction(args[0])
			if err != nil {
				return fmt.Errorf("fetch distinction: %w", err)
			}
			return ui.PrintJSON(resp)
		}
		if err := client.SetTripDistinction(args[0], distinctionValue); err != nil {
			return fmt.Errorf("set distinction: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Set distinction to %q", distinctionValue), map[string]interface{}{"tripKey": args[0], "distinction": distinctionValue})
	},
}

var createGuideCmd = &cobra.Command{
	Use:   "create-guide [trip-key]",
	Short: "Promote a trip plan into a published guide",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.CreateGuideFromTripPlan(args[0])
		if err != nil {
			return fmt.Errorf("create guide: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var getIfEditedCmd = &cobra.Command{
	Use:   "get-if-edited",
	Short: "Ask the server which trip plans changed since given revisions",
	RunE: func(cmd *cobra.Command, args []string) error {
		var req wanderlog.GetIfEditedRequest
		if getIfEditedBody == "" {
			return fmt.Errorf("--body is required (JSON: {\"tripPlans\":[{\"key\":\"...\",\"lastEditedAt\":\"...\"}]})")
		}
		if err := json.Unmarshal([]byte(getIfEditedBody), &req); err != nil {
			return fmt.Errorf("invalid --body JSON: %w", err)
		}
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetIfEdited(req)
		if err != nil {
			return fmt.Errorf("getIfEdited failed: %w", err)
		}
		return ui.PrintJSON(resp)
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
