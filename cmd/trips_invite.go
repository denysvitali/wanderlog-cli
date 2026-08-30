package cmd

import (
	"fmt"
	"net/mail"
	"strings"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(tripsInviteEmails) == 0 {
			return fmt.Errorf("at least one --email is required")
		}
		emails := make([]string, 0, len(tripsInviteEmails))
		for _, value := range tripsInviteEmails {
			email := strings.TrimSpace(value)
			address, err := mail.ParseAddress(email)
			if err != nil || address.Address != email {
				return fmt.Errorf("invalid invitee email %q", value)
			}
			emails = append(emails, email)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		if err := client.SendTripInvites(args[0], wanderlog.SendInvitesRequest{Invitees: emails}); err != nil {
			return fmt.Errorf("send invites: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Sent %d invite(s)", len(emails)), map[string]interface{}{"tripKey": args[0], "invitees": emails})
	},
}

var tripsInviteListCmd = &cobra.Command{
	Use:   "list [trip-key]",
	Short: "List trip invites",
	Long: `List pending invites for a trip.

Examples:
  wanderlog trips invite list abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		invites, err := client.ListTripInvites(args[0])
		if err != nil {
			return fmt.Errorf("list invites: %w", err)
		}
		return ui.PrintJSON(invites)
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
