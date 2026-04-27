package cmd

import (
	"fmt"
	"os"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/spf13/cobra"
)

var (
	updateTitle        string
	updateStartDate    string
	updateEndDate      string
	updatePrivacy      string
	likeValue          bool
	shareCanEdit       bool
	shareCanView       bool
	autofillQuery      string
	inviteEmails       []string
	collaboratorID     int
	moveFromSectionID  int
	moveToSectionID    int
	movePosition       int
	reorderPlaceIDs    string
	checklistItems     []string
	checklistChecked   bool
	hotelCheckIn       string
	hotelCheckOut      string
	hotelGuests        int
	locationLat        float64
	locationLng        float64
	flightStopsAirline string
	flightStopsDate    string
)

var updateTripCmd = &cobra.Command{
	Use:   "update-trip [trip-key]",
	Short: "Update trip title, dates, or privacy",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateDateFlag(updateStartDate, "start")
		validateDateFlag(updateEndDate, "end")
		if updateTitle == "" && updateStartDate == "" && updateEndDate == "" && updatePrivacy == "" {
			logger.Error("At least one of --title, --start, --end, or --privacy is required")
			os.Exit(1)
		}

		client := newClient(true)
		err := client.UpdateTrip(args[0], wanderlog.UpdateTripRequest{
			Title:     updateTitle,
			StartDate: updateStartDate,
			EndDate:   updateEndDate,
			Privacy:   updatePrivacy,
		})
		if err != nil {
			logger.WithError(err).Error("Failed to update trip")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Updated trip %s", args[0]), map[string]string{"tripKey": args[0]})
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore [trip-key]",
	Short: "Restore a deleted trip",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		if err := client.RestoreTrip(args[0]); err != nil {
			logger.WithError(err).Error("Failed to restore trip")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Restored trip %s", args[0]), map[string]string{"tripKey": args[0]})
	},
}

var sectionsCmd = &cobra.Command{
	Use:   "sections [trip-key]",
	Short: "List trip sections",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		sections, err := client.GetTripSections(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch sections")
			os.Exit(1)
		}
		ui.PrintJSON(sections)
	},
}

var tripFlightsCmd = &cobra.Command{
	Use:   "trip-flights [trip-key]",
	Short: "List flights attached to a trip",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		flights, err := client.GetTripFlights(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to fetch trip flights")
			os.Exit(1)
		}
		ui.PrintJSON(flights)
	},
}

var exportTripCmd = &cobra.Command{
	Use:   "export [trip-key]",
	Short: "Export a trip to Google Maps",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.ExportTrip(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to export trip")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var likeCmd = &cobra.Command{
	Use:   "like [trip-key]",
	Short: "Like or unlike a trip",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		if err := client.SetLike(args[0], likeValue); err != nil {
			logger.WithError(err).Error("Failed to update like")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Set like=%t for trip %s", likeValue, args[0]), map[string]interface{}{"tripKey": args[0], "liked": likeValue})
	},
}

var likeCountCmd = &cobra.Command{
	Use:   "like-count [trip-key]",
	Short: "Get trip like count",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		resp, err := client.GetLikeCount(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to get like count")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Manage trip invites",
}

var inviteSendCmd = &cobra.Command{
	Use:   "send [trip-key]",
	Short: "Send trip invites",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(inviteEmails) == 0 {
			logger.Error("At least one --email is required")
			os.Exit(1)
		}
		client := newClient(true)
		if err := client.SendTripInvites(args[0], wanderlog.SendInvitesRequest{Invitees: inviteEmails}); err != nil {
			logger.WithError(err).Error("Failed to send invites")
			os.Exit(1)
		}
		printSuccess(outputFormat, fmt.Sprintf("Sent %d invite(s)", len(inviteEmails)), map[string]interface{}{"tripKey": args[0], "invitees": inviteEmails})
	},
}

var inviteListCmd = &cobra.Command{
	Use:   "list [trip-key]",
	Short: "List trip invites",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		invites, err := client.ListTripInvites(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to list invites")
			os.Exit(1)
		}
		ui.PrintJSON(invites)
	},
}

var collaboratorCmd = &cobra.Command{
	Use:   "collaborator",
	Short: "Manage trip collaborators",
}

var collaboratorAddCmd = &cobra.Command{
	Use:   "add [trip-key]",
	Short: "Add a collaborator by user ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		if err := client.AddCollaborator(args[0], collaboratorID); err != nil {
			logger.WithError(err).Error("Failed to add collaborator")
			os.Exit(1)
		}
		printSuccess(outputFormat, "Added collaborator", map[string]interface{}{"tripKey": args[0], "userId": collaboratorID})
	},
}

var collaboratorRemoveCmd = &cobra.Command{
	Use:   "remove [trip-key]",
	Short: "Remove a collaborator by user ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		if err := client.RemoveCollaborator(args[0], collaboratorID); err != nil {
			logger.WithError(err).Error("Failed to remove collaborator")
			os.Exit(1)
		}
		printSuccess(outputFormat, "Removed collaborator", map[string]interface{}{"tripKey": args[0], "userId": collaboratorID})
	},
}

var shareKeyCmd = &cobra.Command{
	Use:   "share-key [edit-key]",
	Short: "Create or get a share key for an edit key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		resp, err := client.GetOrCreateShareKey(args[0], wanderlog.ShareKeyPermissions{
			CanEdit: shareCanEdit,
			CanView: shareCanView,
		})
		if err != nil {
			logger.WithError(err).Error("Failed to create share key")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var movePlaceCmd = &cobra.Command{
	Use:   "move-place [trip-key] [place-id]",
	Short: "Move a place between sections",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		placeID := parseRequiredInt(args[1], "place ID")
		if err := client.MovePlace(args[0], placeID, moveFromSectionID, moveToSectionID, movePosition); err != nil {
			logger.WithError(err).Error("Failed to move place")
			os.Exit(1)
		}
		printSuccess(outputFormat, "Moved place", map[string]interface{}{"tripKey": args[0], "placeId": placeID})
	},
}

var reorderPlacesCmd = &cobra.Command{
	Use:   "reorder-places [trip-key] [section-id]",
	Short: "Reorder places in a section",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		sectionID := parseRequiredInt(args[1], "section ID")
		placeIDs := parseIntCSV(reorderPlaceIDs, "place IDs")
		if err := client.ReorderPlaces(args[0], sectionID, placeIDs); err != nil {
			logger.WithError(err).Error("Failed to reorder places")
			os.Exit(1)
		}
		printSuccess(outputFormat, "Reordered places", map[string]interface{}{"tripKey": args[0], "sectionId": sectionID, "placeIds": placeIDs})
	},
}

var autofillCmd = &cobra.Command{
	Use:   "autofill-day [trip-key] [section-id]",
	Short: "Get itinerary suggestions for a day",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		sectionID := parseRequiredInt(args[1], "section ID")
		resp, err := client.AutofillDay(args[0], sectionID, autofillQuery)
		if err != nil {
			logger.WithError(err).Error("Failed to autofill day")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var checklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Manage checklist sections",
}

var checklistAddCmd = &cobra.Command{
	Use:   "add [trip-key] [section-id]",
	Short: "Add checklist items",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		sectionID := parseRequiredInt(args[1], "section ID")
		resp, err := client.AddChecklistItems(args[0], sectionID, parseChecklistItems(checklistItems))
		if err != nil {
			logger.WithError(err).Error("Failed to add checklist items")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var checklistToggleCmd = &cobra.Command{
	Use:   "toggle [trip-key] [section-id] [item-id]",
	Short: "Toggle a checklist item",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(true)
		sectionID := parseRequiredInt(args[1], "section ID")
		itemID := parseRequiredInt(args[2], "item ID")
		resp, err := client.ToggleChecklistItem(args[0], sectionID, itemID, checklistChecked)
		if err != nil {
			logger.WithError(err).Error("Failed to toggle checklist item")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var travelCmd = &cobra.Command{
	Use:   "travel",
	Short: "Search flights and lodging helpers",
}

var airlinesCmd = &cobra.Command{
	Use:   "airlines",
	Short: "List all airlines",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		resp, err := client.GetAllAirlines()
		if err != nil {
			logger.WithError(err).Error("Failed to list airlines")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var airportsCmd = &cobra.Command{
	Use:   "airports [query]",
	Short: "Search airports",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		var resp interface{}
		var err error
		if locationLat != 0 || locationLng != 0 {
			resp, err = client.AutocompleteAirportWithLocation(args[0], locationLat, locationLng)
		} else {
			resp, err = client.AutocompleteAirport(args[0])
		}
		if err != nil {
			logger.WithError(err).Error("Failed to search airports")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var flightStopsCmd = &cobra.Command{
	Use:   "flight-stops [flight-number]",
	Short: "Show stops for a flight number",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if flightStopsAirline == "" {
			logger.Error("--airline is required (e.g., UA, BA, LH)")
			os.Exit(1)
		}
		if flightStopsDate == "" {
			logger.Error("--date is required (YYYY-MM-DD)")
			os.Exit(1)
		}
		client := newClient(false)
		resp, err := client.GetFlightStops(args[0], flightStopsAirline, flightStopsDate)
		if err != nil {
			logger.WithError(err).Error("Failed to get flight stops")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var hotelsCmd = &cobra.Command{
	Use:   "hotels [query]",
	Short: "Search hotels/lodging",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateDateFlag(hotelCheckIn, "check-in")
		validateDateFlag(hotelCheckOut, "check-out")
		client := newClient(false)
		resp, err := client.SearchLodgings(args[0], hotelCheckIn, hotelCheckOut, hotelGuests)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Hotel search failed: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success || resp.Data == nil {
			fmt.Fprintf(os.Stderr, "Hotel search returned no results for %q. The lodging API may be unavailable or your session may have expired.\n", args[0])
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

var hotelRatesCmd = &cobra.Command{
	Use:   "hotel-rates [property-id]",
	Short: "Get Google lodging price rates",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(false)
		resp, err := client.GetGooglePriceRates(args[0])
		if err != nil {
			logger.WithError(err).Error("Failed to get hotel rates")
			os.Exit(1)
		}
		ui.PrintJSON(resp)
	},
}

func init() {
	// root registrations disabled - commands moved under `trips` or `travel`
	// rootCmd.AddCommand(restoreCmd, sectionsCmd, tripFlightsCmd, exportTripCmd, likeCmd, likeCountCmd, inviteCmd, collaboratorCmd, shareKeyCmd, autofillCmd, checklistCmd, travelCmd)
	editCmd.AddCommand(updateTripCmd, movePlaceCmd, reorderPlacesCmd)

	updateTripCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "Trip title")
	updateTripCmd.Flags().StringVar(&updateStartDate, "start", "", "Start date (YYYY-MM-DD)")
	updateTripCmd.Flags().StringVar(&updateEndDate, "end", "", "End date (YYYY-MM-DD)")
	updateTripCmd.Flags().StringVar(&updatePrivacy, "privacy", "", "Trip privacy (public, private, unlisted)")
	likeCmd.Flags().BoolVar(&likeValue, "liked", true, "Whether the trip should be liked")
	shareKeyCmd.Flags().BoolVar(&shareCanEdit, "can-edit", false, "Allow editing")
	shareKeyCmd.Flags().BoolVar(&shareCanView, "can-view", true, "Allow viewing")
	autofillCmd.Flags().StringVarP(&autofillQuery, "query", "q", "", "Suggestion query, such as restaurants or museums")

	inviteCmd.AddCommand(inviteSendCmd, inviteListCmd)
	inviteSendCmd.Flags().StringArrayVar(&inviteEmails, "email", nil, "Invitee email; may be supplied multiple times")

	collaboratorCmd.AddCommand(collaboratorAddCmd, collaboratorRemoveCmd)
	collaboratorAddCmd.Flags().IntVar(&collaboratorID, "user-id", 0, "User ID")
	collaboratorRemoveCmd.Flags().IntVar(&collaboratorID, "user-id", 0, "User ID")
	_ = collaboratorAddCmd.MarkFlagRequired("user-id")
	_ = collaboratorRemoveCmd.MarkFlagRequired("user-id")

	movePlaceCmd.Flags().IntVar(&moveFromSectionID, "from-section", 0, "Source section ID")
	movePlaceCmd.Flags().IntVar(&moveToSectionID, "to-section", 0, "Destination section ID")
	movePlaceCmd.Flags().IntVar(&movePosition, "position", 0, "Destination position")
	_ = movePlaceCmd.MarkFlagRequired("from-section")
	_ = movePlaceCmd.MarkFlagRequired("to-section")
	reorderPlacesCmd.Flags().StringVar(&reorderPlaceIDs, "place-ids", "", "Comma-separated place IDs in the desired order")
	_ = reorderPlacesCmd.MarkFlagRequired("place-ids")

	checklistCmd.AddCommand(checklistAddCmd, checklistToggleCmd)
	checklistAddCmd.Flags().StringArrayVar(&checklistItems, "item", nil, "Checklist item text; may be supplied multiple times")
	checklistToggleCmd.Flags().BoolVar(&checklistChecked, "checked", true, "Checked state")

	travelCmd.AddCommand(airlinesCmd, airportsCmd, flightStopsCmd, hotelsCmd, hotelRatesCmd)
	airportsCmd.Flags().Float64Var(&locationLat, "lat", 0, "Latitude for location bias")
	airportsCmd.Flags().Float64Var(&locationLng, "lng", 0, "Longitude for location bias")
	flightStopsCmd.Flags().StringVar(&flightStopsAirline, "airline", "", "Airline IATA code (e.g., UA, BA)")
	flightStopsCmd.Flags().StringVar(&flightStopsDate, "date", "", "Departure date (YYYY-MM-DD)")
	hotelsCmd.Flags().StringVar(&hotelCheckIn, "check-in", "", "Check-in date (YYYY-MM-DD)")
	hotelsCmd.Flags().StringVar(&hotelCheckOut, "check-out", "", "Check-out date (YYYY-MM-DD)")
	hotelsCmd.Flags().IntVar(&hotelGuests, "guests", 1, "Number of guests")

	for _, command := range []*cobra.Command{
		restoreCmd, sectionsCmd, tripFlightsCmd, exportTripCmd, likeCmd, likeCountCmd,
		inviteSendCmd, inviteListCmd, collaboratorAddCmd, collaboratorRemoveCmd,
		shareKeyCmd, updateTripCmd, movePlaceCmd, reorderPlacesCmd, autofillCmd,
		checklistAddCmd, checklistToggleCmd, airlinesCmd, airportsCmd, flightStopsCmd,
		hotelsCmd, hotelRatesCmd,
	} {
		command.Flags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format (pretty, json)")
		command.Flags().StringVar(&sessionCookie, "session", "", "Session cookie for authentication")
		command.Flags().StringVar(&xsrfToken, "xsrf", "", "XSRF token for authentication")
	}
}
