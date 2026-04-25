package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			return
		}

		sectionID := parseRequiredInt(args[1], "section ID")
		resp, err := client.AutofillDay(args[0], sectionID, tripsAutofillQuery)
		if err != nil {
			logger.WithError(err).Error("Failed to autofill day")
			return
		}
		ui.PrintJSON(resp)
	},
}

var tripsAutofillQuery string

func init() {
	tripsCmd.AddCommand(tripsAutofillCmd)

	tripsAutofillCmd.Flags().StringVarP(&tripsAutofillQuery, "query", "q", "", "Suggestion query, such as restaurants or museums")

	tripsAutofillCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format")
	tripsAutofillCmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
	tripsAutofillCmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
}
