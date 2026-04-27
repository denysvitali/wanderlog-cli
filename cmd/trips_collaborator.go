package cmd

import (
	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsCollaboratorCmd = &cobra.Command{
	Use:   "collaborator",
	Short: "Manage trip collaborators",
}

var tripsCollaboratorAddCmd = &cobra.Command{
	Use:   "add [trip-key]",
	Short: "Add a collaborator by user ID",
	Long: `Add a collaborator to a trip by their user ID.

Examples:
  wanderlog trips collaborator add abc123xyz --user-id 12345`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			return
		}

		if err := client.AddCollaborator(args[0], tripsCollaboratorID); err != nil {
			logger.WithError(err).Error("Failed to add collaborator")
			return
		}
		printSuccess(outputFormat, "Added collaborator", map[string]interface{}{"tripKey": args[0], "userId": tripsCollaboratorID})
	},
}

var tripsCollaboratorRemoveCmd = &cobra.Command{
	Use:   "remove [trip-key]",
	Short: "Remove a collaborator by user ID",
	Long: `Remove a collaborator from a trip.

Examples:
  wanderlog trips collaborator remove abc123xyz --user-id 12345`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			return
		}

		if err := client.RemoveCollaborator(args[0], tripsCollaboratorID); err != nil {
			logger.WithError(err).Error("Failed to remove collaborator")
			return
		}
		printSuccess(outputFormat, "Removed collaborator", map[string]interface{}{"tripKey": args[0], "userId": tripsCollaboratorID})
	},
}

var tripsCollaboratorID int

func init() {
	tripsCmd.AddCommand(tripsCollaboratorCmd)
	tripsCollaboratorCmd.AddCommand(tripsCollaboratorAddCmd, tripsCollaboratorRemoveCmd)

	tripsCollaboratorAddCmd.Flags().IntVar(&tripsCollaboratorID, "user-id", 0, "User ID to add")
	tripsCollaboratorRemoveCmd.Flags().IntVar(&tripsCollaboratorID, "user-id", 0, "User ID to remove")
	_ = tripsCollaboratorAddCmd.MarkFlagRequired("user-id")
	_ = tripsCollaboratorRemoveCmd.MarkFlagRequired("user-id")

	for _, c := range []*cobra.Command{tripsCollaboratorAddCmd, tripsCollaboratorRemoveCmd} {
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
