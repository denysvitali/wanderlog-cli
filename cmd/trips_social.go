package cmd

import (
	"encoding/json"
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}

		if err := client.SetLike(args[0], tripsLikeValue); err != nil {
			return fmt.Errorf("update like: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Set like=%t for trip %s", tripsLikeValue, args[0]), map[string]interface{}{"tripKey": args[0], "liked": tripsLikeValue})
	},
}

var tripsLikeCountCmd = &cobra.Command{
	Use:   "like-count [trip-key]",
	Short: "Get trip like count",
	Long: `Get the like count and current user's like status for a trip.

Examples:
  wanderlog trips like-count abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetLikeCount(args[0])
		if err != nil {
			return fmt.Errorf("get like count: %w", err)
		}
		return ui.PrintJSON(resp)
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}

		resp, err := client.GetOrCreateShareKey(args[0], wanderlog.ShareKeyPermissions{
			CanEdit: shareCanEdit,
			CanView: shareCanView,
		})
		if err != nil {
			return fmt.Errorf("create share key: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var tripsRegisterViewCmd = &cobra.Command{
	Use:   "register-view [trip-key]",
	Short: "Register a view on a shared trip",
	Long: `Register a view on a shared trip.

Examples:
  wanderlog trips register-view abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.RegisterTripView(args[0]); err != nil {
			return fmt.Errorf("register view: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Registered view on %s", args[0]), map[string]interface{}{"tripKey": args[0]})
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(tripsDistinctionValue != "")
		if err != nil {
			return err
		}

		if tripsDistinctionValue == "" {
			resp, err := client.GetTripDistinction(args[0])
			if err != nil {
				return fmt.Errorf("fetch distinction: %w", err)
			}
			return ui.PrintJSON(resp)
		}

		if err := client.SetTripDistinction(args[0], tripsDistinctionValue); err != nil {
			return fmt.Errorf("set distinction: %w", err)
		}
		return printSuccess(outputFormat, fmt.Sprintf("Set distinction to %q", tripsDistinctionValue), map[string]interface{}{"tripKey": args[0], "distinction": tripsDistinctionValue})
	},
}

var tripsCreateGuideCmd = &cobra.Command{
	Use:   "create-guide [trip-key]",
	Short: "Promote a trip plan into a published guide",
	Long: `Promote a trip plan into a published guide.

Examples:
  wanderlog trips create-guide abc123xyz`,
	Args: cobra.ExactArgs(1),
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

var tripsGetIfEditedCmd = &cobra.Command{
	Use:   "get-if-edited",
	Short: "Ask the server which trip plans changed since given revisions",
	Long: `Ask the server which trip plans changed since given revisions.

Examples:
  wanderlog trips get-if-edited --body '{"tripPlans":[{"key":"abc","lastEditedAt":"..."}]}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var req wanderlog.GetIfEditedRequest
		if tripsGetIfEditedBody == "" {
			return fmt.Errorf("--body is required (JSON: {\"tripPlans\":[{\"key\":\"...\",\"lastEditedAt\":\"...\"}]})")
		}
		if err := json.Unmarshal([]byte(tripsGetIfEditedBody), &req); err != nil {
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

var tripsUpdateRequiredCmd = &cobra.Command{
	Use:   "update-required [trip-key]",
	Short: "Check whether the client must upgrade for this trip's schema",
	Long: `Check whether the CLI client must upgrade for a given trip's schema version.

Examples:
  wanderlog trips update-required abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetTripUpdateRequired(args[0])
		if err != nil {
			return fmt.Errorf("fetch updateRequired status: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var tripsJournalCmd = &cobra.Command{
	Use:   "journal [journal-key]",
	Short: "Fetch a published view-only journal",
	Long: `Fetch a published view-only journal by its journal key.

Examples:
  wanderlog trips journal abc123xyz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		resp, err := client.GetViewOnlyJournal(args[0])
		if err != nil {
			return fmt.Errorf("fetch journal: %w", err)
		}
		return ui.PrintJSON(resp)
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
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
