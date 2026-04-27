package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsInviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Manage trip invites",
}

var tripsInviteSendCmd = &cobra.Command{
	Use:   "send [trip-key]",
	Short: "Send trip invites",
	Long: `Send invites to collaborate on a trip.

Examples:
  wanderlog trips invite send abc123xyz --email alice@example.com --email bob@example.com`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(tripsInviteEmails) == 0 {
			logger.Error("At least one --email is required")
			return
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			return
		}

		if err := client.SendTripInvites(args[0], wanderlog.SendInvitesRequest{Invitees: tripsInviteEmails}); err != nil {
			logger.WithError(err).Error("Failed to send invites")
			return
		}
		printSuccess(outputFormat, fmt.Sprintf("Sent %d invite(s)", len(tripsInviteEmails)), map[string]interface{}{"tripKey": args[0], "invitees": tripsInviteEmails})
	},
}

var tripsInviteListCmd = &cobra.Command{
	Use:   "list [trip-key]",
	Short: "List trip invites",
	Long: `List pending invites for a trip.

Examples:
  wanderlog trips invite list abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			return
		}

		invites, err := client.ListTripInvites(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to list invites")
			return
		}
		ui.PrintJSON(invites)
	},
}

var tripsInviteEmails []string

func init() {
	tripsCmd.AddCommand(tripsInviteCmd)
	tripsInviteCmd.AddCommand(tripsInviteSendCmd, tripsInviteListCmd)

	tripsInviteSendCmd.Flags().StringArrayVar(&tripsInviteEmails, "email", nil, "Invitee email; may be supplied multiple times")

	for _, c := range []*cobra.Command{tripsInviteSendCmd, tripsInviteListCmd} {
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
