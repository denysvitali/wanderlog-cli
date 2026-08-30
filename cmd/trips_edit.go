package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
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
  wanderlog trips edit add-place abc123xyz --name "Tokyo Station" --lat 35.6812 --lng 139.7671 --section 123 --start-time 09:30`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tripKey := args[0]

		if tripsEditPlaceName == "" {
			return fmt.Errorf("place name is required (--name)")
		}

		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
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
			Place:     placeInfo,
			Text:      tripsEditPlaceText,
			StartTime: tripsEditStartTime,
			EndTime:   tripsEditEndTime,
		}

		err = client.AddPlace(tripKey, tripsEditSectionID, req)
		if err != nil {
			return fmt.Errorf("add place: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("📍 Successfully added place '%s' to trip %s", tripsEditPlaceName, tripKey)))
		if tripsEditSectionID > 0 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Section ID: %d", tripsEditSectionID)))
		}
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		tripKey := args[0]
		placeIDStr := args[1]

		placeIDInt, err := parseRequiredIntE(placeIDStr, "place ID")
		if err != nil {
			return err
		}

		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		err = client.RemovePlaceContext(cmd.Context(), tripKey, tripsEditSectionID, placeIDInt)
		if err != nil {
			return fmt.Errorf("remove place: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("🗑️  Successfully removed place %d from trip %s", placeIDInt, tripKey)))
		if tripsEditSectionID > 0 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Section ID: %d", tripsEditSectionID)))
		}
		return nil
	},
}

var tripsEditClearSectionCmd = &cobra.Command{
	Use:   "clear-section [trip-key] [section-id]",
	Short: "Clear all blocks from a section",
	Long: `Clear all blocks (places, notes, etc.) from a specific section of a trip.

Examples:
  wanderlog trips edit clear-section abc123xyz 6310036`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tripKey := args[0]
		sectionIDStr := args[1]

		sectionID, err := parseRequiredIntE(sectionIDStr, "section ID")
		if err != nil {
			return err
		}

		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		err = client.ClearSectionBlocksContext(cmd.Context(), tripKey, sectionID)
		if err != nil {
			return fmt.Errorf("clear section: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("🧹 Successfully cleared all blocks from section %d in trip %s", sectionID, tripKey)))
		return nil
	},
}

var tripsEditDeleteSectionCmd = &cobra.Command{
	Use:   "delete-section [trip-key] [section-id]",
	Short: "Delete an entire section from a trip",
	Long: `Delete an entire section from a trip. This removes the section completely.

Examples:
  wanderlog trips edit delete-section abc123xyz 6310036`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tripKey := args[0]
		sectionIDStr := args[1]

		sectionID, err := parseRequiredIntE(sectionIDStr, "section ID")
		if err != nil {
			return err
		}

		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		err = client.DeleteSectionContext(cmd.Context(), tripKey, sectionID)
		if err != nil {
			return fmt.Errorf("delete section: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("🗑️ Successfully deleted section %d from trip %s", sectionID, tripKey)))
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		tripKey := args[0]

		_, _ = fmt.Fprint(cmd.ErrOrStderr(), ui.WarningStyle.Render("⚠️  WARNING: This will remove ALL places from ALL sections of your trip!\n"))
		_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Are you sure you want to continue? (y/N): ")

		var response string
		_, _ = fmt.Fscanln(cmd.InOrStdin(), &response)

		if response != "y" && response != "Y" && response != "yes" {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), ui.InfoStyle.Render("Operation canceled."))
			return err
		}

		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		err = client.NukeTripPlacesContext(cmd.Context(), tripKey)
		if err != nil {
			return fmt.Errorf("nuke trip places: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("💥 Successfully nuked all place data from trip %s", tripKey)))
		fmt.Println(ui.InfoStyle.Render("🔄 Try accessing your trip now - the location error should be fixed."))
		return nil
	},
}

var tripsEditMovePlaceCmd = &cobra.Command{
	Use:   "move-place [trip-key] [place-id]",
	Short: "Move a place between sections",
	Long: `Move a place from one section to another.

Examples:
  wanderlog trips edit move-place abc123xyz 12345 --from-section 100 --to-section 200 --position 0`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		placeID, err := parseRequiredIntE(args[1], "place ID")
		if err != nil {
			return err
		}
		if err := client.MovePlaceContext(cmd.Context(), args[0], placeID, tripsEditMoveFromSection, tripsEditMoveToSection, tripsEditMovePosition); err != nil {
			return fmt.Errorf("move place: %w", err)
		}
		return printSuccess(outputFormat, "Moved place", map[string]interface{}{"tripKey": args[0], "placeId": placeID})
	},
}

var tripsEditReorderPlacesCmd = &cobra.Command{
	Use:   "reorder-places [trip-key] [section-id]",
	Short: "Reorder places in a section",
	Long: `Reorder places within a section by providing the desired order of place IDs.

Examples:
  wanderlog trips edit reorder-places abc123xyz 123 --place-ids "456,789,012"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		sectionID, err := parseRequiredIntE(args[1], "section ID")
		if err != nil {
			return err
		}
		placeIDs, err := parseIntCSVE(tripsEditReorderPlaceIDs, "place IDs")
		if err != nil {
			return err
		}
		if err := client.ReorderPlacesContext(cmd.Context(), args[0], sectionID, placeIDs); err != nil {
			return fmt.Errorf("reorder places: %w", err)
		}
		return printSuccess(outputFormat, "Reordered places", map[string]interface{}{"tripKey": args[0], "sectionId": sectionID, "placeIds": placeIDs})
	},
}

var tripsEditSetPlaceTimeCmd = &cobra.Command{
	Use:   "set-place-time [trip-key] [place-id]",
	Short: "Set a place visit time",
	Long: `Set the visit time shown on a place in an itinerary section.

Examples:
  wanderlog trips edit set-place-time abc123xyz 12345 --section 100 --start-time 09:30
  wanderlog trips edit set-place-time abc123xyz 12345 --section 100 --start-time 09:30 --end-time 11:00`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		placeID, err := parseRequiredIntE(args[1], "place ID")
		if err != nil {
			return err
		}
		if err := client.UpdatePlaceVisitTimeContext(cmd.Context(), args[0], tripsEditSectionID, placeID, tripsEditStartTime, tripsEditEndTime); err != nil {
			return fmt.Errorf("update place visit time: %w", err)
		}
		return printSuccess(outputFormat, "Updated place visit time", map[string]interface{}{
			"tripKey":   args[0],
			"sectionId": tripsEditSectionID,
			"placeId":   placeID,
			"startTime": tripsEditStartTime,
			"endTime":   tripsEditEndTime,
		})
	},
}

var (
	tripsEditPlaceName       string
	tripsEditPlaceID         string
	tripsEditLatitude        float64
	tripsEditLongitude       float64
	tripsEditSectionID       int
	tripsEditPlaceText       string
	tripsEditMoveFromSection int
	tripsEditMoveToSection   int
	tripsEditMovePosition    int
	tripsEditReorderPlaceIDs string
	tripsEditStartTime       string
	tripsEditEndTime         string
)

func init() {
	tripsCmd.AddCommand(tripsEditCmd)
	tripsEditCmd.AddCommand(
		tripsEditAddPlaceCmd, tripsEditRemovePlaceCmd,
		tripsEditClearSectionCmd, tripsEditDeleteSectionCmd,
		tripsEditNukePlacesCmd, tripsEditMovePlaceCmd, tripsEditReorderPlacesCmd,
		tripsEditSetPlaceTimeCmd,
	)

	// add-place flags
	tripsEditAddPlaceCmd.Flags().StringVarP(&tripsEditPlaceName, "name", "n", "", "Place name (required)")
	tripsEditAddPlaceCmd.Flags().StringVar(&tripsEditPlaceID, "place-id", "", "Google Place ID")
	tripsEditAddPlaceCmd.Flags().Float64Var(&tripsEditLatitude, "lat", 0, "Latitude")
	tripsEditAddPlaceCmd.Flags().Float64Var(&tripsEditLongitude, "lng", 0, "Longitude")
	tripsEditAddPlaceCmd.Flags().IntVar(&tripsEditSectionID, "section", 0, "Section ID")
	tripsEditAddPlaceCmd.Flags().StringVar(&tripsEditPlaceText, "text", "", "Additional text/notes")
	tripsEditAddPlaceCmd.Flags().StringVar(&tripsEditStartTime, "start-time", "", "Visit start time (HH:MM, 24-hour)")
	tripsEditAddPlaceCmd.Flags().StringVar(&tripsEditEndTime, "end-time", "", "Visit end time (HH:MM, 24-hour)")

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

	// set-place-time flags
	tripsEditSetPlaceTimeCmd.Flags().IntVar(&tripsEditSectionID, "section", 0, "Section ID")
	tripsEditSetPlaceTimeCmd.Flags().StringVar(&tripsEditStartTime, "start-time", "", "Visit start time (HH:MM, 24-hour)")
	tripsEditSetPlaceTimeCmd.Flags().StringVar(&tripsEditEndTime, "end-time", "", "Visit end time (HH:MM, 24-hour)")
	_ = tripsEditSetPlaceTimeCmd.MarkFlagRequired("section")

	// auth flags
	for _, c := range []*cobra.Command{
		tripsEditAddPlaceCmd, tripsEditRemovePlaceCmd,
		tripsEditClearSectionCmd, tripsEditDeleteSectionCmd,
		tripsEditNukePlacesCmd, tripsEditMovePlaceCmd, tripsEditReorderPlacesCmd,
		tripsEditSetPlaceTimeCmd,
	} {
		c.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		c.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
