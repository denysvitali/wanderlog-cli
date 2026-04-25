package cmd

import (
	"github.com/spf13/cobra"
)

// Backward compatibility aliases for old command paths.
// These allow existing scripts using old commands to continue working.

var compatListCmd = &cobra.Command{
	Use:                "list",
	Short:              "List your trips (deprecated: use 'wanderlog trips list')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsListCmd.Run(tripsListCmd, args)
	},
}

var compatTripCmd = &cobra.Command{
	Use:                "trip",
	Short:              "Get trip information (deprecated: use 'wanderlog trips show')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsShowCmd.Run(tripsShowCmd, args)
	},
}

var compatCreateCmd = &cobra.Command{
	Use:                "create",
	Short:              "Create a new trip (deprecated: use 'wanderlog trips create')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsCreateCmd.Run(tripsCreateCmd, args)
	},
}

var compatDeleteCmd = &cobra.Command{
	Use:                "delete",
	Short:              "Delete a trip (deprecated: use 'wanderlog trips delete')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsDeleteCmd.Run(tripsDeleteCmd, args)
	},
}

var compatCopyCmd = &cobra.Command{
	Use:                "copy",
	Short:              "Copy an existing trip (deprecated: use 'wanderlog trips copy')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsCopyCmd.Run(tripsCopyCmd, args)
	},
}

var compatRestoreCmd = &cobra.Command{
	Use:                "restore",
	Short:              "Restore a deleted trip (deprecated: use 'wanderlog trips restore')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsRestoreCmd.Run(tripsRestoreCmd, args)
	},
}

var compatPlacesCmd = &cobra.Command{
	Use:                "places",
	Short:              "Show places from a trip (deprecated: use 'wanderlog trips places')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsPlacesCmd.Run(tripsPlacesCmd, args)
	},
}

var compatImagesCmd = &cobra.Command{
	Use:                "images",
	Short:              "Show trip images (deprecated: use 'wanderlog trips images')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsImagesCmd.Run(tripsImagesCmd, args)
	},
}

var compatExpensesCmd = &cobra.Command{
	Use:                "expenses",
	Short:              "Download trip expenses as CSV (deprecated: use 'wanderlog trips expenses')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsExpensesCmd.Run(tripsExpensesCmd, args)
	},
}

var compatSectionsCmd = &cobra.Command{
	Use:                "sections",
	Short:              "List trip sections (deprecated: use 'wanderlog trips sections')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsSectionsCmd.Run(tripsSectionsCmd, args)
	},
}

var compatLikeCmd = &cobra.Command{
	Use:                "like",
	Short:              "Like or unlike a trip (deprecated: use 'wanderlog trips like')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsLikeCmd.Run(tripsLikeCmd, args)
	},
}

var compatLikeCountCmd = &cobra.Command{
	Use:                "like-count",
	Short:              "Get trip like count (deprecated: use 'wanderlog trips like-count')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		tripsLikeCountCmd.Run(tripsLikeCountCmd, args)
	},
}

var compatSearchCmd = &cobra.Command{
	Use:                "search",
	Short:              "Search for places (deprecated: use 'wanderlog search google')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		searchGoogleCmd.Run(searchGoogleCmd, args)
	},
}

var compatSearchPlacesCmd = &cobra.Command{
	Use:                "search-places",
	Short:              "Search places using Wanderlog autocomplete (deprecated: use 'wanderlog search wanderlog')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		searchWanderlogCmd.Run(searchWanderlogCmd, args)
	},
}

var compatPlaceDetailsCmd = &cobra.Command{
	Use:                "place-details",
	Short:              "Get place details (deprecated: use 'wanderlog search place-details')",
	Hidden:             true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		searchPlaceDetailsCmd.Run(searchPlaceDetailsCmd, args)
	},
}

func init() {
	// Auth commands - no aliases needed, these are unchanged
	// login, logout, status stay at root

	// Trip aliases
	rootCmd.AddCommand(
		compatListCmd, compatTripCmd, compatCreateCmd, compatDeleteCmd,
		compatCopyCmd, compatRestoreCmd, compatPlacesCmd, compatImagesCmd,
		compatExpensesCmd, compatSectionsCmd, compatLikeCmd, compatLikeCountCmd,
	)

	// Search aliases
	rootCmd.AddCommand(compatSearchCmd, compatSearchPlacesCmd, compatPlaceDetailsCmd)
}
