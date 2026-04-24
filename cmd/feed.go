package cmd

import (
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetFeedHome()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch home feed")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var feedRecentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Show the most recently edited trip",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetFeedMostRecent()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch most recent feed")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var feedFriendsCmd = &cobra.Command{
	Use:   "friends",
	Short: "Show trip plans from friends",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetFriendsPlans()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch friends plans")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var feedHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show trip edit history",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetTripHistory(feedHistoryOffset)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch trip history")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var feedLegacyCmd = &cobra.Command{
	Use:   "legacy",
	Short: "Show the legacy /tripPlans/feed response",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetFeed()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch feed")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var feedV2Cmd = &cobra.Command{
	Use:   "v2",
	Short: "Show the v2 trip feed",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetFeedV2()
		if err != nil {
			logger.WithError(err).Error("Failed to fetch feed v2")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var feedGuidesCmd = &cobra.Command{
	Use:   "guides",
	Short: "Browse curated Wanderlog guides",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		resp, err := client.BrowseGuides(feedGuidesGeoID)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch guides")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

func init() {
	rootCmd.AddCommand(feedCmd)
	feedCmd.AddCommand(feedHomeCmd, feedRecentCmd, feedFriendsCmd, feedHistoryCmd, feedLegacyCmd, feedV2Cmd, feedGuidesCmd)
	feedHistoryCmd.Flags().IntVar(&feedHistoryOffset, "offset", 0, "Pagination offset")
	feedGuidesCmd.Flags().IntVar(&feedGuidesGeoID, "geo-id", 0, "Limit guides to a specific geo ID")

	for _, command := range []*cobra.Command{feedHomeCmd, feedRecentCmd, feedFriendsCmd, feedHistoryCmd, feedLegacyCmd, feedV2Cmd, feedGuidesCmd} {
		command.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
