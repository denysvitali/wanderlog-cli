package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsLikeCmd = &cobra.Command{
	Use:   "like [trip-key]",
	Short: "Like or unlike a trip",
	Long: `Like or unlike a trip.

Examples:
  wanderlog trips like abc123xyz --liked
  wanderlog trips like abc123xyz --liked=false`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		if err := client.SetLike(args[0], tripsLikeValue); err != nil {
			logger.WithError(err).Error("Failed to update like")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Set like=%t for trip %s", tripsLikeValue, args[0]), map[string]interface{}{"tripKey": args[0], "liked": tripsLikeValue})
	},
}

var tripsLikeCountCmd = &cobra.Command{
	Use:   "like-count [trip-key]",
	Short: "Get trip like count",
	Long: `Get the like count and current user's like status for a trip.

Examples:
  wanderlog trips like-count abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetLikeCount(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to get like count")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var tripsShareKeyCmd = &cobra.Command{
	Use:   "share-key [edit-key]",
	Short: "Create or get a share key",
	Long: `Create or get a share key for an edit key.

Examples:
  wanderlog trips share-key abc123xyz --can-edit
  wanderlog trips share-key abc123xyz --can-view`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		resp, err := client.GetOrCreateShareKey(args[0], wanderlog.ShareKeyPermissions{
			CanEdit: shareCanEdit,
			CanView: shareCanView,
		})
		if err != nil {
			logger.WithError(err).Error("Failed to create share key")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var tripsRegisterViewCmd = &cobra.Command{
	Use:   "register-view [trip-key]",
	Short: "Register a view on a shared trip",
	Long: `Register a view on a shared trip.

Examples:
  wanderlog trips register-view abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.RegisterTripView(args[0]); err != nil {
			logger.WithError(err).Error("Failed to register view")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Registered view on %s", args[0]), map[string]interface{}{"tripKey": args[0]})
	},
}

var tripsDistinctionCmd = &cobra.Command{
	Use:   "distinction [trip-key]",
	Short: "Get or set the trip's distinction/badge",
	Long: `Get or set the trip's distinction badge.
Use --set to assign a new distinction.

Examples:
  wanderlog trips distinction abc123xyz
  wanderlog trips distinction abc123xyz --set "Best Trip"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		authenticated := tripsDistinctionValue != ""
		if authenticated {
			if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
				logger.WithError(err).Error("Authentication required")
				os.Exit(1)
			}
		}

		if tripsDistinctionValue == "" {
			resp, err := client.GetTripDistinction(args[0])
			if err != nil {
				logger.WithError(err).Error("Failed to fetch distinction")
				os.Exit(1)
			}
			ui.PrintJSON(resp)
			return
		}

		if err := client.SetTripDistinction(args[0], tripsDistinctionValue); err != nil {
			logger.WithError(err).Error("Failed to set distinction")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Set distinction to %q", tripsDistinctionValue), map[string]interface{}{"tripKey": args[0], "distinction": tripsDistinctionValue})
	},
}

var tripsCreateGuideCmd = &cobra.Command{
	Use:   "create-guide [trip-key]",
	Short: "Promote a trip plan into a published guide",
	Long: `Promote a trip plan into a published guide.

Examples:
  wanderlog trips create-guide abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		resp, err := client.CreateGuideFromTripPlan(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to create guide")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var tripsGetIfEditedCmd = &cobra.Command{
	Use:   "get-if-edited",
	Short: "Ask the server which trip plans changed since given revisions",
	Long: `Ask the server which trip plans changed since given revisions.

Examples:
  wanderlog trips get-if-edited --body '{"tripPlans":[{"key":"abc","lastEditedAt":"..."}]}'`,
	Run: func(cmd *cobra.Command, args []string) {
		var req wanderlog.GetIfEditedRequest
		if tripsGetIfEditedBody == "" {
			logger.Error("--body is required (JSON: {\"tripPlans\":[{\"key\":\"...\",\"lastEditedAt\":\"...\"}]})")
			os.Exit(1)
		}
		if err := json.Unmarshal([]byte(tripsGetIfEditedBody), &req); err != nil {
			logger.WithError(err).Error("Invalid --body JSON")
			os.Exit(1)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		resp, err := client.GetIfEdited(req)
		if err != nil {
			logger.WithError(err).Error("getIfEdited failed")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var tripsUpdateRequiredCmd = &cobra.Command{
	Use:   "update-required [trip-key]",
	Short: "Check whether the client must upgrade for this trip's schema",
	Long: `Check whether the CLI client must upgrade for a given trip's schema version.

Examples:
  wanderlog trips update-required abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetTripUpdateRequired(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch updateRequired status")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var tripsJournalCmd = &cobra.Command{
	Use:   "journal [journal-key]",
	Short: "Fetch a published view-only journal",
	Long: `Fetch a published view-only journal by its journal key.

Examples:
  wanderlog trips journal abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetViewOnlyJournal(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch journal")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var (
	tripsLikeValue        bool
	tripsDistinctionValue string
	tripsGetIfEditedBody  string
)

func init() {
	tripsCmd.AddCommand(
		tripsLikeCmd, tripsLikeCountCmd, tripsShareKeyCmd,
		tripsRegisterViewCmd, tripsDistinctionCmd, tripsCreateGuideCmd,
		tripsGetIfEditedCmd, tripsUpdateRequiredCmd, tripsJournalCmd,
	)

	tripsLikeCmd.Flags().BoolVar(&tripsLikeValue, "liked", true, "Whether the trip should be liked")
	tripsShareKeyCmd.Flags().BoolVar(&shareCanEdit, "can-edit", false, "Allow editing")
	tripsShareKeyCmd.Flags().BoolVar(&shareCanView, "can-view", true, "Allow viewing")
	tripsDistinctionCmd.Flags().StringVar(&tripsDistinctionValue, "set", "", "Set the distinction to this value (otherwise get)")
	tripsGetIfEditedCmd.Flags().StringVar(&tripsGetIfEditedBody, "body", "", "JSON request body")

	for _, c := range []*cobra.Command{
		tripsLikeCmd, tripsLikeCountCmd, tripsShareKeyCmd,
		tripsRegisterViewCmd, tripsDistinctionCmd, tripsCreateGuideCmd,
		tripsGetIfEditedCmd, tripsUpdateRequiredCmd, tripsJournalCmd,
	} {
		c.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
