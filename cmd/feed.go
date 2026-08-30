package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
)

var (
	feedHistoryOffset int
	feedGuidesGeoID   int
)

var feedCmd = &cobra.Command{
	Use:   "feed",
	Short: "Discover trips and guides",
}

var feedHomeCmd = &cobra.Command{
	Use:   "home",
	Short: "Show the authenticated user's home feed",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetFeedHome()
		if err != nil {
			return fmt.Errorf("fetch home feed: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var feedRecentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Show the most recently edited trip",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetFeedMostRecent()
		if err != nil {
			return fmt.Errorf("fetch most recent feed: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var feedFriendsCmd = &cobra.Command{
	Use:   "friends",
	Short: "Show trip plans from friends",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetFriendsPlans()
		if err != nil {
			return fmt.Errorf("fetch friends plans: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var feedHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show trip edit history",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetTripHistory(feedHistoryOffset)
		if err != nil {
			return fmt.Errorf("fetch trip history: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var feedLegacyCmd = &cobra.Command{
	Use:   "legacy",
	Short: "Show the legacy /tripPlans/feed response",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetFeed()
		if err != nil {
			return fmt.Errorf("fetch feed: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var feedV2Cmd = &cobra.Command{
	Use:   "v2",
	Short: "Show the v2 trip feed",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(true)
		if err != nil {
			return err
		}
		resp, err := client.GetFeedV2()
		if err != nil {
			return fmt.Errorf("fetch feed v2: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

var feedGuidesCmd = &cobra.Command{
	Use:   "guides",
	Short: "Browse curated Wanderlog guides",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientE(false)
		if err != nil {
			return err
		}
		resp, err := client.BrowseGuides(feedGuidesGeoID)
		if err != nil {
			return fmt.Errorf("fetch guides: %w", err)
		}
		return ui.PrintJSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(feedCmd)
	feedCmd.AddCommand(feedHomeCmd, feedRecentCmd, feedFriendsCmd, feedHistoryCmd, feedLegacyCmd, feedV2Cmd, feedGuidesCmd)
	feedHistoryCmd.Flags().IntVar(&feedHistoryOffset, "offset", 0, "Pagination offset")
	feedGuidesCmd.Flags().IntVar(&feedGuidesGeoID, "geo-id", 0, "Limit guides to a specific geo ID")

	for _, command := range []*cobra.Command{feedHomeCmd, feedRecentCmd, feedFriendsCmd, feedHistoryCmd, feedLegacyCmd, feedV2Cmd, feedGuidesCmd} {
		command.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
