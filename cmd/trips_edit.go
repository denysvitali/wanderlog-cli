package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

var tripsEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit trip content",
	Long: `Edit trip content including adding/removing places and managing itinerary.

Requires authentication via 'wanderlog login' or environment variables.`,
}

var tripsEditAddPlaceCmd = &cobra.Command{
	Use:   "add-place [trip-key]",
	Short: "Add a place to a trip",
	Long: `Add a place to a trip section.

Examples:
  wanderlog trips edit add-place abc123xyz --name "Eiffel Tower" --place-id "ChIJLU7jZClu5kcR4PcOOO6p3I0"
  wanderlog trips edit add-place abc123xyz --name "Tokyo Station" --lat 35.6812 --lng 139.7671 --section 123`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]

		if tripsEditPlaceName == "" {
			logger.Error("Place name is required (--name)")
			os.Exit(1)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		placeInfo := wanderlog.AddPlaceInfo{
			PlaceID: tripsEditPlaceID,
			Name:    tripsEditPlaceName,
		}

		if tripsEditLatitude != 0 || tripsEditLongitude != 0 {
			placeInfo.Geometry = &models.PlaceGeometry{
				Location: models.PlaceLocation{
					Lat: tripsEditLatitude,
					Lng: tripsEditLongitude,
				},
			}
		}

		req := wanderlog.AddPlaceRequest{
			Place: placeInfo,
			Text:  tripsEditPlaceText,
		}

		err := client.AddPlace(tripKey, tripsEditSectionID, req)
		if err != nil {
			logger.WithError(err).Error("Failed to add place")
			os.Exit(1)
		}

		fmt.Printf("📍 Successfully added place '%s' to trip %s\n", tripsEditPlaceName, tripKey)
		if tripsEditSectionID > 0 {
			fmt.Printf("Section ID: %d\n", tripsEditSectionID)
		}
	},
}

var tripsEditRemovePlaceCmd = &cobra.Command{
	Use:   "remove-place [trip-key] [place-id]",
	Short: "Remove a place from a trip",
	Long: `Remove a place from a trip section.

Examples:
  wanderlog trips edit remove-place abc123xyz 12345
  wanderlog trips edit remove-place abc123xyz 12345 --section 123`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]
		placeIDStr := args[1]

		placeIDInt, err := strconv.Atoi(placeIDStr)
		if err != nil {
			logger.WithError(err).Error("Invalid place ID - must be a number")
			os.Exit(1)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err = client.RemovePlace(tripKey, tripsEditSectionID, placeIDInt)
		if err != nil {
			logger.WithError(err).Error("Failed to remove place")
			os.Exit(1)
		}

		fmt.Printf("🗑️  Successfully removed place %d from trip %s\n", placeIDInt, tripKey)
		if tripsEditSectionID > 0 {
			fmt.Printf("Section ID: %d\n", tripsEditSectionID)
		}
	},
}

var tripsEditClearSectionCmd = &cobra.Command{
	Use:   "clear-section [trip-key] [section-id]",
	Short: "Clear all blocks from a section",
	Long: `Clear all blocks (places, notes, etc.) from a specific section of a trip.

Examples:
  wanderlog trips edit clear-section abc123xyz 6310036`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]
		sectionIDStr := args[1]

		sectionID, err := strconv.Atoi(sectionIDStr)
		if err != nil {
			logger.WithError(err).Error("Invalid section ID")
			os.Exit(1)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err = client.ClearSectionBlocks(tripKey, sectionID)
		if err != nil {
			logger.WithError(err).Error("Failed to clear section")
			os.Exit(1)
		}

		fmt.Printf("🧹 Successfully cleared all blocks from section %d in trip %s\n", sectionID, tripKey)
	},
}

var tripsEditDeleteSectionCmd = &cobra.Command{
	Use:   "delete-section [trip-key] [section-id]",
	Short: "Delete an entire section from a trip",
	Long: `Delete an entire section from a trip. This removes the section completely.

Examples:
  wanderlog trips edit delete-section abc123xyz 6310036`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]
		sectionIDStr := args[1]

		sectionID, err := strconv.Atoi(sectionIDStr)
		if err != nil {
			logger.WithError(err).Error("Invalid section ID")
			os.Exit(1)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err = client.DeleteSection(tripKey, sectionID)
		if err != nil {
			logger.WithError(err).Error("Failed to delete section")
			os.Exit(1)
		}

		fmt.Printf("🗑️ Successfully deleted section %d from trip %s\n", sectionID, tripKey)
	},
}

var tripsEditNukePlacesCmd = &cobra.Command{
	Use:   "nuke-places [trip-key]",
	Short: "Nuclear option: clear ALL place data from a trip",
	Long: `Nuclear option to clear all place blocks from all sections in a trip.
Use this as a last resort to fix corrupted trip data.

WARNING: This will remove ALL places from ALL sections of your trip!

Examples:
  wanderlog trips edit nuke-places abc123xyz`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]

		fmt.Print("⚠️  WARNING: This will remove ALL places from ALL sections of your trip!\n")
		fmt.Print("Are you sure you want to continue? (y/N): ")

		var response string
		_, _ = fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Operation canceled.")
			return
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err := client.NukeTripPlaces(tripKey)
		if err != nil {
			logger.WithError(err).Error("Failed to nuke trip places")
			os.Exit(1)
		}

		fmt.Printf("💥 Successfully nuked all place data from trip %s\n", tripKey)
		fmt.Println("🔄 Try accessing your trip now - the location error should be fixed.")
	},
}

var tripsEditMovePlaceCmd = &cobra.Command{
	Use:   "move-place [trip-key] [place-id]",
	Short: "Move a place between sections",
	Long: `Move a place from one section to another.

Examples:
  wanderlog trips edit move-place abc123xyz 12345 --from-section 100 --to-section 200 --position 0`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		placeID := parseRequiredInt(args[1], "place ID")
		if err := client.MovePlace(args[0], placeID, tripsEditMoveFromSection, tripsEditMoveToSection, tripsEditMovePosition); err != nil {
			logger.WithError(err).Error("Failed to move place")
			os.Exit(1)
		}
		printSuccess(outputFormat, "Moved place", map[string]interface{}{"tripKey": args[0], "placeId": placeID})
	},
}

var tripsEditReorderPlacesCmd = &cobra.Command{
	Use:   "reorder-places [trip-key] [section-id]",
	Short: "Reorder places in a section",
	Long: `Reorder places within a section by providing the desired order of place IDs.

Examples:
  wanderlog trips edit reorder-places abc123xyz 123 --place-ids "456,789,012"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := wanderlog.NewClient()
		client.SetLogger(logger)

		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		sectionID := parseRequiredInt(args[1], "section ID")
		placeIDs := parseIntCSV(tripsEditReorderPlaceIDs, "place IDs")
		if err := client.ReorderPlaces(args[0], sectionID, placeIDs); err != nil {
			logger.WithError(err).Error("Failed to reorder places")
			os.Exit(1)
		}
		printSuccess(outputFormat, "Reordered places", map[string]interface{}{"tripKey": args[0], "sectionId": sectionID, "placeIds": placeIDs})
	},
}

var (
	tripsEditPlaceName     string
	tripsEditPlaceID       string
	tripsEditLatitude      float64
	tripsEditLongitude     float64
	tripsEditSectionID     int
	tripsEditPlaceText     string
	tripsEditMoveFromSection int
	tripsEditMoveToSection   int
	tripsEditMovePosition    int
	tripsEditReorderPlaceIDs string
)

func init() {
	tripsCmd.AddCommand(tripsEditCmd)
	tripsEditCmd.AddCommand(
		tripsEditAddPlaceCmd, tripsEditRemovePlaceCmd,
		tripsEditClearSectionCmd, tripsEditDeleteSectionCmd,
		tripsEditNukePlacesCmd, tripsEditMovePlaceCmd, tripsEditReorderPlacesCmd,
	)

	// add-place flags
	tripsEditAddPlaceCmd.Flags().StringVarP(&tripsEditPlaceName, "name", "n", "", "Place name (required)")
	tripsEditAddPlaceCmd.Flags().StringVar(&tripsEditPlaceID, "place-id", "", "Google Place ID")
	tripsEditAddPlaceCmd.Flags().Float64Var(&tripsEditLatitude, "lat", 0, "Latitude")
	tripsEditAddPlaceCmd.Flags().Float64Var(&tripsEditLongitude, "lng", 0, "Longitude")
	tripsEditAddPlaceCmd.Flags().IntVar(&tripsEditSectionID, "section", 0, "Section ID")
	tripsEditAddPlaceCmd.Flags().StringVar(&tripsEditPlaceText, "text", "", "Additional text/notes")

	// remove-place flags
	tripsEditRemovePlaceCmd.Flags().IntVar(&tripsEditSectionID, "section", 0, "Section ID")

	// move-place flags
	tripsEditMovePlaceCmd.Flags().IntVar(&tripsEditMoveFromSection, "from-section", 0, "Source section ID")
	tripsEditMovePlaceCmd.Flags().IntVar(&tripsEditMoveToSection, "to-section", 0, "Destination section ID")
	tripsEditMovePlaceCmd.Flags().IntVar(&tripsEditMovePosition, "position", 0, "Destination position")
	_ = tripsEditMovePlaceCmd.MarkFlagRequired("from-section")
	_ = tripsEditMovePlaceCmd.MarkFlagRequired("to-section")

	// reorder-places flags
	tripsEditReorderPlacesCmd.Flags().StringVar(&tripsEditReorderPlaceIDs, "place-ids", "", "Comma-separated place IDs in the desired order")
	_ = tripsEditReorderPlacesCmd.MarkFlagRequired("place-ids")

	// auth flags
	for _, c := range []*cobra.Command{
		tripsEditAddPlaceCmd, tripsEditRemovePlaceCmd,
		tripsEditClearSectionCmd, tripsEditDeleteSectionCmd,
		tripsEditNukePlacesCmd, tripsEditMovePlaceCmd, tripsEditReorderPlacesCmd,
	} {
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
