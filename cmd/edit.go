package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

var (
	placeName     string
	placeID       string
	latitude      float64
	longitude     float64
	sectionIDFlag int
	placeText     string
	startTimeFlag string
	endTimeFlag   string
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit trip content",
	Long: `Edit trip content including adding/removing places and managing itinerary.

Requires authentication via 'wanderlog login' or environment variables.`,
}

var addPlaceCmd = &cobra.Command{
	Use:   "add-place [trip-key]",
	Short: "Add a place to a trip",
	Long: `Add a place to a trip section.

Examples:
  wanderlog edit add-place abc123xyz --name "Eiffel Tower" --place-id "ChIJLU7jZClu5kcR4PcOOO6p3I0"
  wanderlog edit add-place abc123xyz --name "Tokyo Station" --lat 35.6812 --lng 139.7671 --section 123
  wanderlog edit add-place abc123xyz --name "Custom Place" --text "Great restaurant!" --start-time 19:00`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tripKey := args[0]

		if placeName == "" {
			return fmt.Errorf("place name is required")
		}

		client, err := newClientContextE(cmd.Context(), true)
		if err != nil {
			return err
		}

		// Build the place info with proper geometry structure
		placeInfo := wanderlog.AddPlaceInfo{
			PlaceID: placeID,
			Name:    placeName,
		}

		// Only add geometry if coordinates are provided
		if latitude != 0 || longitude != 0 {
			placeInfo.Geometry = &models.PlaceGeometry{
				Location: models.PlaceLocation{
					Lat: latitude,
					Lng: longitude,
				},
			}
		}

		req := wanderlog.AddPlaceRequest{
			Place:     placeInfo,
			Text:      placeText,
			StartTime: startTimeFlag,
			EndTime:   endTimeFlag,
		}

		err = client.AddPlace(tripKey, sectionIDFlag, req)
		if err != nil {
			return fmt.Errorf("add place: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("📍 Successfully added place '%s' to trip %s", placeName, tripKey)))
		if sectionIDFlag > 0 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Section ID: %d", sectionIDFlag)))
		}
		return nil
	},
}

var removePlaceCmd = &cobra.Command{
	Use:   "remove-place [trip-key] [place-id]",
	Short: "Remove a place from a trip",
	Long: `Remove a place from a trip section.

Examples:
  wanderlog edit remove-place abc123xyz 12345
  wanderlog edit remove-place abc123xyz 12345 --section 123`,
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

		err = client.RemovePlaceContext(cmd.Context(), tripKey, sectionIDFlag, placeIDInt)
		if err != nil {
			return fmt.Errorf("remove place: %w", err)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("🗑️  Successfully removed place %d from trip %s", placeIDInt, tripKey)))
		if sectionIDFlag > 0 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Section ID: %d", sectionIDFlag)))
		}
		return nil
	},
}

var clearSectionCmd = &cobra.Command{
	Use:   "clear-section [trip-key] [section-id]",
	Short: "Clear all blocks from a section",
	Long: `Clear all blocks (places, notes, etc.) from a specific section of a trip.

Examples:
  wanderlog edit clear-section abc123xyz 6310036`,
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

var deleteSectionCmd = &cobra.Command{
	Use:   "delete-section [trip-key] [section-id]",
	Short: "Delete an entire section from a trip",
	Long: `Delete an entire section from a trip. This removes the section completely.

Examples:
  wanderlog edit delete-section abc123xyz 6310036`,
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

var nukeTrippPlacesCmd = &cobra.Command{
	Use:   "nuke-places [trip-key]",
	Short: "Nuclear option: Clear ALL place data from a trip",
	Long: `Nuclear option to clear all place blocks from all sections in a trip. 
Use this as a last resort to fix corrupted trip data.

WARNING: This will remove ALL places from ALL sections of your trip!

Examples:
  wanderlog edit nuke-places abc123xyz`,
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

func init() {
	// root registration disabled - command moved under `trips edit`
	// rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(addPlaceCmd)
	editCmd.AddCommand(removePlaceCmd)
	editCmd.AddCommand(clearSectionCmd)
	editCmd.AddCommand(deleteSectionCmd)
	editCmd.AddCommand(nukeTrippPlacesCmd)

	// Add place flags
	addPlaceCmd.Flags().StringVarP(&placeName, "name", "n", "", "Place name (required)")
	addPlaceCmd.Flags().StringVar(&placeID, "place-id", "", "Google Place ID")
	addPlaceCmd.Flags().Float64Var(&latitude, "lat", 0, "Latitude")
	addPlaceCmd.Flags().Float64Var(&longitude, "lng", 0, "Longitude")
	addPlaceCmd.Flags().IntVar(&sectionIDFlag, "section", 0, "Section ID")
	addPlaceCmd.Flags().StringVar(&placeText, "text", "", "Additional text/notes")
	addPlaceCmd.Flags().StringVar(&startTimeFlag, "start-time", "", "Visit start time (HH:MM, 24-hour)")
	addPlaceCmd.Flags().StringVar(&endTimeFlag, "end-time", "", "Visit end time (HH:MM, 24-hour)")

	// Remove place flags
	removePlaceCmd.Flags().IntVar(&sectionIDFlag, "section", 0, "Section ID")

	// Auth flags for all edit commands
	for _, cmd := range []*cobra.Command{addPlaceCmd, removePlaceCmd, clearSectionCmd, deleteSectionCmd, nukeTrippPlacesCmd} {
		cmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		cmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
