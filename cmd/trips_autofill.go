package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsAutofillCmd = &cobra.Command{
	Use:   "autofill [trip-key] [section-id]",
	Short: "Get itinerary suggestions for a day",
	Long: `Get itinerary suggestions (restaurants, museums, etc.) for a specific section/day.

Examples:
  wanderlog trips autofill abc123xyz 123 --query "restaurants"
  wanderlog trips autofill abc123xyz 123 --query "museums"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sectionID, err := parseRequiredIntE(args[1], "section ID")
		if err != nil {
			return err
		}
		query := strings.TrimSpace(tripsAutofillQuery)
		if query == "" {
			return fmt.Errorf("--query is required")
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		resp, err := client.AutofillDay(args[0], sectionID, query)
		if err != nil {
			return fmt.Errorf("autofill day: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var tripsAutofillQuery string

func init() {
	tripsCmd.AddCommand(tripsAutofillCmd)

	tripsAutofillCmd.Flags().StringVarP(&tripsAutofillQuery, "query", "q", "", "Suggestion query, such as restaurants or museums")

	tripsAutofillCmd.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
	tripsAutofillCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsAutofillCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
}
