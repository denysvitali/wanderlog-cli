package cmd

import (
	"fmt"
	"os"
	"strconv"

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
  wanderlog edit add-place abc123xyz --name "Custom Place" --text "Great restaurant!"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]

		if placeName == "" {
			logger.Error("Place name is required")
			os.Exit(1)
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Ensure authentication (from flags, env vars, or keychain)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
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
			Place: placeInfo,
			Text:  placeText,
		}

		err := client.AddPlace(tripKey, sectionIDFlag, req)
		if err != nil {
			logger.WithError(err).Error("Failed to add place")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("📍 Successfully added place '%s' to trip %s", placeName, tripKey)))
		if sectionIDFlag > 0 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Section ID: %d", sectionIDFlag)))
		}
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

		// Ensure authentication (from flags, env vars, or keychain)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err = client.RemovePlace(tripKey, sectionIDFlag, placeIDInt)
		if err != nil {
			logger.WithError(err).Error("Failed to remove place")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("🗑️  Successfully removed place %d from trip %s", placeIDInt, tripKey)))
		if sectionIDFlag > 0 {
			fmt.Println(ui.InfoStyle.Render(fmt.Sprintf("Section ID: %d", sectionIDFlag)))
		}
	},
}

var clearSectionCmd = &cobra.Command{
	Use:   "clear-section [trip-key] [section-id]",
	Short: "Clear all blocks from a section",
	Long: `Clear all blocks (places, notes, etc.) from a specific section of a trip.

Examples:
  wanderlog edit clear-section abc123xyz 6310036`,
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

		// Ensure authentication (from flags, env vars, or keychain)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err = client.ClearSectionBlocks(tripKey, sectionID)
		if err != nil {
			logger.WithError(err).Error("Failed to clear section")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("🧹 Successfully cleared all blocks from section %d in trip %s", sectionID, tripKey)))
	},
}

var deleteSectionCmd = &cobra.Command{
	Use:   "delete-section [trip-key] [section-id]",
	Short: "Delete an entire section from a trip",
	Long: `Delete an entire section from a trip. This removes the section completely.

Examples:
  wanderlog edit delete-section abc123xyz 6310036`,
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

		// Ensure authentication (from flags, env vars, or keychain)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err = client.DeleteSection(tripKey, sectionID)
		if err != nil {
			logger.WithError(err).Error("Failed to delete section")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("🗑️ Successfully deleted section %d from trip %s", sectionID, tripKey)))
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
	Run: func(cmd *cobra.Command, args []string) {
		tripKey := args[0]

		fmt.Print(ui.WarningStyle.Render("⚠️  WARNING: This will remove ALL places from ALL sections of your trip!\n"))
		fmt.Print("Are you sure you want to continue? (y/N): ")

		var response string
		_, _ = fmt.Scanln(&response)

		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println(ui.InfoStyle.Render("Operation canceled."))
			return
		}

		client := wanderlog.NewClient()
		client.SetLogger(logger)

		// Ensure authentication (from flags, env vars, or keychain)
		if err := client.EnsureAuthenticated(sessionCookie, xsrfToken); err != nil {
			logger.WithError(err).Error("Authentication required")
			os.Exit(1)
		}

		err := client.NukeTripPlaces(tripKey)
		if err != nil {
			logger.WithError(err).Error("Failed to nuke trip places")
			os.Exit(1)
		}

		fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("💥 Successfully nuked all place data from trip %s", tripKey)))
		fmt.Println(ui.InfoStyle.Render("🔄 Try accessing your trip now - the location error should be fixed."))
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

	// Remove place flags
	removePlaceCmd.Flags().IntVar(&sectionIDFlag, "section", 0, "Section ID")

	// Auth flags for all edit commands
	for _, cmd := range []*cobra.Command{addPlaceCmd, removePlaceCmd, clearSectionCmd, deleteSectionCmd, nukeTrippPlacesCmd} {
		cmd.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		cmd.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
