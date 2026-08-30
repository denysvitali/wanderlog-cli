package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// compatibilityCommand exposes a canonical command at its former root path.
// Sharing the canonical flag set keeps validation, defaults, arguments, and
// behavior in sync instead of maintaining a second command implementation.
func compatibilityCommand(name string, target *cobra.Command) *cobra.Command {
	use := name
	if _, suffix, ok := strings.Cut(target.Use, " "); ok {
		use += " " + suffix
	}

	path := target.CommandPath()
	path = strings.TrimPrefix(path, "wanderlog ")
	compat := &cobra.Command{
		Use:                use,
		Short:              target.Short + " (deprecated: use 'wanderlog " + path + "')",
		Long:               target.Long,
		Args:               target.Args,
		ArgAliases:         target.ArgAliases,
		ValidArgs:          target.ValidArgs,
		ValidArgsFunction:  target.ValidArgsFunction,
		Hidden:             true,
		DisableSuggestions: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if target.PreRunE != nil {
				if err := target.PreRunE(cmd, args); err != nil {
					return err
				}
			} else if target.PreRun != nil {
				target.PreRun(cmd, args)
			}

			if target.RunE != nil {
				return target.RunE(cmd, args)
			}
			target.Run(cmd, args)
			return nil
		},
	}
	compat.Flags().AddFlagSet(target.Flags())
	return compat
}

func init() {
	rootCmd.AddCommand(
		compatibilityCommand("list", tripsListCmd),
		compatibilityCommand("trip", tripsShowCmd),
		compatibilityCommand("create", tripsCreateCmd),
		compatibilityCommand("delete", tripsDeleteCmd),
		compatibilityCommand("copy", tripsCopyCmd),
		compatibilityCommand("restore", tripsRestoreCmd),
		compatibilityCommand("places", tripsPlacesCmd),
		compatibilityCommand("images", tripsImagesCmd),
		compatibilityCommand("expenses", tripsExpensesCmd),
		compatibilityCommand("sections", tripsSectionsCmd),
		compatibilityCommand("like", tripsLikeCmd),
		compatibilityCommand("like-count", tripsLikeCountCmd),
		compatibilityCommand("search-places", searchWanderlogCmd),
		compatibilityCommand("place-details", searchPlaceDetailsCmd),
	)

	// `search` is now a command group, but it can still accept the former
	// positional query form without shadowing its modern subcommands.
	searchParentCmd.Use = "search [query]"
	searchParentCmd.Args = searchWanderlogCmd.Args
	searchParentCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if searchWanderlogCmd.RunE != nil {
			return searchWanderlogCmd.RunE(cmd, args)
		}
		searchWanderlogCmd.Run(cmd, args)
		return nil
	}
	searchParentCmd.Flags().AddFlagSet(searchWanderlogCmd.Flags())
}
