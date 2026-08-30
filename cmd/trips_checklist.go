package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

var tripsChecklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Manage checklist sections",
}

var tripsChecklistAddCmd = &cobra.Command{
	Use:   "add [trip-key] [section-id]",
	Short: "Add checklist items",
	Long: `Add checklist items to a section.

Examples:
  wanderlog trips checklist add abc123xyz 123 --item "Pack passport" --item "Book hotel"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sectionID, err := parseRequiredIntE(args[1], "section ID")
		if err != nil {
			return err
		}
		items, err := parseChecklistItemsE(tripsChecklistItems)
		if err != nil {
			return err
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		resp, err := client.AddChecklistItems(args[0], sectionID, items)
		if err != nil {
			return fmt.Errorf("add checklist items: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var tripsChecklistToggleCmd = &cobra.Command{
	Use:   "toggle [trip-key] [section-id] [item-id]",
	Short: "Toggle a checklist item",
	Long: `Toggle a checklist item's checked state.

Examples:
  wanderlog trips checklist toggle abc123xyz 123 456 --checked=true`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		sectionID, err := parseRequiredIntE(args[1], "section ID")
		if err != nil {
			return err
		}
		itemID, err := parseRequiredIntE(args[2], "item ID")
		if err != nil {
			return err
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		resp, err := client.ToggleChecklistItem(args[0], sectionID, itemID, tripsChecklistChecked)
		if err != nil {
			return fmt.Errorf("toggle checklist item: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var (
	tripsChecklistItems   []string
	tripsChecklistChecked bool
)

func init() {
	tripsCmd.AddCommand(tripsChecklistCmd)
	tripsChecklistCmd.AddCommand(tripsChecklistAddCmd, tripsChecklistToggleCmd)

	tripsChecklistAddCmd.Flags().StringArrayVar(&tripsChecklistItems, "item", nil, "Checklist item text; may be supplied multiple times")
	tripsChecklistToggleCmd.Flags().BoolVar(&tripsChecklistChecked, "checked", true, "Checked state")

	for _, c := range []*cobra.Command{tripsChecklistAddCmd, tripsChecklistToggleCmd} {
		c.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
