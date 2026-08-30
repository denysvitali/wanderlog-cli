package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/denysvitali/wanderlog-cli/pkg/ui"
	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
)

type travelWriteClient interface {
	AddFlightReservation(wanderlog.AddFlightReservationRequest) (*wanderlog.TravelMutationResult, error)
	UpdateFlightReservation(wanderlog.UpdateFlightReservationRequest) (*wanderlog.TravelMutationResult, error)
	DeleteFlightReservation(wanderlog.DeleteTravelReservationRequest) (*wanderlog.TravelMutationResult, error)
	AddLodgingReservation(wanderlog.AddLodgingReservationRequest) (*wanderlog.TravelMutationResult, error)
	UpdateLodgingReservation(wanderlog.UpdateLodgingReservationRequest) (*wanderlog.TravelMutationResult, error)
	DeleteLodgingReservation(wanderlog.DeleteTravelReservationRequest) (*wanderlog.TravelMutationResult, error)
	AddTrainReservation(wanderlog.AddTrainReservationRequest) (*wanderlog.TravelMutationResult, error)
	DeleteTrainReservation(wanderlog.DeleteTravelReservationRequest) (*wanderlog.TravelMutationResult, error)
}

type travelWriteClientFactory func() (travelWriteClient, error)

var hhmmPattern = regexp.MustCompile(`^(?:[01][0-9]|2[0-3]):[0-5][0-9]$`)
var airportCodePattern = regexp.MustCompile(`^[A-Za-z]{3}$`)

func newAuthenticatedTravelWriteClient() (travelWriteClient, error) {
	return newClientE(true)
}

func executeTravelWrite(cmd *cobra.Command, output string, result *wanderlog.TravelMutationResult, err error) error {
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("travel mutation returned no result")
	}
	if output == "json" {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			return fmt.Errorf("write JSON output: %w", err)
		}
		return nil
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), ui.SafeText(result.Message))
	return err
}

func travelWriteService(factory travelWriteClientFactory) (travelWriteClient, error) {
	client, err := factory()
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}
	return client, nil
}

func positiveID(value, name string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return id, nil
}

func validateOptionalDate(name, value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("invalid %s %q: use YYYY-MM-DD", name, value)
	}
	return nil
}

func validateOptionalTime(name, value string) error {
	if value != "" && !hhmmPattern.MatchString(value) {
		return fmt.Errorf("invalid %s %q: use HH:MM in 24-hour time", name, value)
	}
	return nil
}

type flightWriteOptions struct {
	output             string
	sectionID          int
	flightNumber       string
	departureDate      string
	departureTime      string
	arrivalDate        string
	arrivalTime        string
	confirmationNumber string
	notes              string
	departureAirport   string
	arrivalAirport     string
}

func newTripsFlightWriteCmd(factory travelWriteClientFactory) *cobra.Command {
	options := &flightWriteOptions{}
	command := &cobra.Command{
		Use:     "flight",
		Aliases: []string{"flights-write"},
		Short:   "Add, update, or delete flight reservations",
	}
	command.PersistentFlags().StringVarP(&options.output, "output", "o", "pretty", "Output format (pretty, json)")
	command.AddCommand(newTripsFlightAddCmd(factory, options), newTripsFlightUpdateCmd(factory, options), newTripsFlightDeleteCmd(factory, options))
	return command
}

func addFlightFlags(command *cobra.Command, options *flightWriteOptions, requireCore bool) {
	command.Flags().StringVar(&options.flightNumber, "flight-number", "", "Flight number with airline code (for example, MU244)")
	command.Flags().StringVar(&options.departureDate, "departure-date", "", "Departure date (YYYY-MM-DD)")
	command.Flags().StringVar(&options.departureTime, "departure-time", "", "Departure time (HH:MM)")
	command.Flags().StringVar(&options.arrivalDate, "arrival-date", "", "Arrival date (YYYY-MM-DD)")
	command.Flags().StringVar(&options.arrivalTime, "arrival-time", "", "Arrival time (HH:MM)")
	command.Flags().StringVar(&options.confirmationNumber, "confirmation", "", "Confirmation number")
	command.Flags().StringVar(&options.notes, "notes", "", "Reservation notes")
	command.Flags().StringVar(&options.departureAirport, "departure-airport", "", "Departure airport IATA code override")
	command.Flags().StringVar(&options.arrivalAirport, "arrival-airport", "", "Arrival airport IATA code override")
	if requireCore {
		_ = command.MarkFlagRequired("flight-number")
		_ = command.MarkFlagRequired("departure-date")
	}
}

func newTripsFlightAddCmd(factory travelWriteClientFactory, options *flightWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "add <trip-key>",
		Short: "Add a flight reservation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(options.departureDate) == "" {
				return fmt.Errorf("--departure-date is required")
			}
			if _, _, err := validateFlightNumber(options.flightNumber); err != nil {
				return err
			}
			if err := validateOptionalDate("departure date", options.departureDate); err != nil {
				return err
			}
			if err := validateOptionalDate("arrival date", options.arrivalDate); err != nil {
				return err
			}
			if err := validateOptionalTime("departure time", options.departureTime); err != nil {
				return err
			}
			if err := validateOptionalTime("arrival time", options.arrivalTime); err != nil {
				return err
			}
			for name, code := range map[string]string{"departure airport": options.departureAirport, "arrival airport": options.arrivalAirport} {
				if code != "" && !airportCodePattern.MatchString(code) {
					return fmt.Errorf("invalid %s %q: use a three-letter IATA code", name, code)
				}
			}
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.AddFlightReservation(wanderlog.AddFlightReservationRequest{
				TripKey: strings.TrimSpace(args[0]), FlightNumber: strings.ToUpper(strings.TrimSpace(options.flightNumber)),
				DepartureDate: options.departureDate, DepartureTime: options.departureTime, ArrivalDate: options.arrivalDate, ArrivalTime: options.arrivalTime,
				ConfirmationNumber: options.confirmationNumber, Notes: options.notes, DepartureAirport: strings.ToUpper(options.departureAirport), ArrivalAirport: strings.ToUpper(options.arrivalAirport),
			})
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	addFlightFlags(command, options, true)
	return command
}

func newTripsFlightUpdateCmd(factory travelWriteClientFactory, options *flightWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <trip-key> <block-id>",
		Short: "Update a flight reservation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			blockID, err := positiveID(args[1], "block-id")
			if err != nil {
				return err
			}
			changed := []string{"flight-number", "departure-date", "departure-time", "arrival-date", "arrival-time", "confirmation", "notes", "departure-airport", "arrival-airport"}
			if !anyFlagChanged(cmd, changed...) {
				return fmt.Errorf("at least one flight field flag is required")
			}
			if cmd.Flags().Changed("flight-number") {
				if _, _, err := validateFlightNumber(options.flightNumber); err != nil {
					return err
				}
			}
			if err := validateOptionalDate("departure date", options.departureDate); err != nil {
				return err
			}
			if err := validateOptionalDate("arrival date", options.arrivalDate); err != nil {
				return err
			}
			if err := validateOptionalTime("departure time", options.departureTime); err != nil {
				return err
			}
			if err := validateOptionalTime("arrival time", options.arrivalTime); err != nil {
				return err
			}
			if cmd.Flags().Changed("section") && options.sectionID <= 0 {
				return fmt.Errorf("--section must be a positive integer")
			}
			request := wanderlog.UpdateFlightReservationRequest{TripKey: strings.TrimSpace(args[0]), SectionID: options.sectionID, BlockID: blockID}
			request.FlightNumber = changedString(cmd, "flight-number", strings.ToUpper(strings.TrimSpace(options.flightNumber)))
			request.DepartureDate = changedString(cmd, "departure-date", options.departureDate)
			request.DepartureTime = changedString(cmd, "departure-time", options.departureTime)
			request.ArrivalDate = changedString(cmd, "arrival-date", options.arrivalDate)
			request.ArrivalTime = changedString(cmd, "arrival-time", options.arrivalTime)
			request.ConfirmationNumber = changedString(cmd, "confirmation", options.confirmationNumber)
			request.Notes = changedString(cmd, "notes", options.notes)
			request.DepartureAirport = changedString(cmd, "departure-airport", strings.ToUpper(options.departureAirport))
			request.ArrivalAirport = changedString(cmd, "arrival-airport", strings.ToUpper(options.arrivalAirport))
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.UpdateFlightReservation(request)
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	command.Flags().IntVar(&options.sectionID, "section", 0, "Section ID containing the block (optional)")
	addFlightFlags(command, options, false)
	return command
}

func newTripsFlightDeleteCmd(factory travelWriteClientFactory, options *flightWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <trip-key> <block-id>",
		Short: "Delete a flight reservation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			blockID, err := positiveID(args[1], "block-id")
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("section") && options.sectionID <= 0 {
				return fmt.Errorf("--section must be a positive integer")
			}
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.DeleteFlightReservation(wanderlog.DeleteTravelReservationRequest{TripKey: strings.TrimSpace(args[0]), SectionID: options.sectionID, BlockID: blockID})
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	command.Flags().IntVar(&options.sectionID, "section", 0, "Section ID containing the block (optional)")
	return command
}

func anyFlagChanged(cmd *cobra.Command, names ...string) bool {
	for _, name := range names {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func changedString(cmd *cobra.Command, flag, value string) *string {
	if !cmd.Flags().Changed(flag) {
		return nil
	}
	return &value
}

type lodgingWriteOptions struct {
	output             string
	sectionID          int
	name               string
	placeID            string
	latitude           float64
	longitude          float64
	checkIn            string
	checkOut           string
	confirmationNumber string
	travelerNames      []string
	notes              string
}

func newTripsLodgingWriteCmd(factory travelWriteClientFactory) *cobra.Command {
	options := &lodgingWriteOptions{}
	command := &cobra.Command{Use: "lodging", Aliases: []string{"hotel"}, Short: "Add, update, or delete lodging reservations"}
	command.PersistentFlags().StringVarP(&options.output, "output", "o", "pretty", "Output format (pretty, json)")
	command.AddCommand(newTripsLodgingAddCmd(factory, options), newTripsLodgingUpdateCmd(factory, options), newTripsLodgingDeleteCmd(factory, options))
	return command
}

func addLodgingFlags(command *cobra.Command, options *lodgingWriteOptions) {
	command.Flags().StringVar(&options.name, "name", "", "Lodging name")
	command.Flags().StringVar(&options.placeID, "place-id", "", "Google Place ID")
	command.Flags().Float64Var(&options.latitude, "lat", 0, "Latitude")
	command.Flags().Float64Var(&options.longitude, "lng", 0, "Longitude")
	command.Flags().StringVar(&options.checkIn, "check-in", "", "Check-in date (YYYY-MM-DD)")
	command.Flags().StringVar(&options.checkOut, "check-out", "", "Check-out date (YYYY-MM-DD)")
	command.Flags().StringVar(&options.confirmationNumber, "confirmation", "", "Confirmation number")
	command.Flags().StringSliceVar(&options.travelerNames, "traveler", nil, "Traveler name (repeatable)")
	command.Flags().StringVar(&options.notes, "notes", "", "Reservation notes")
}

func validateLodgingOptions(cmd *cobra.Command, options *lodgingWriteOptions) error {
	if err := validateOptionalDate("check-in date", options.checkIn); err != nil {
		return err
	}
	if err := validateOptionalDate("check-out date", options.checkOut); err != nil {
		return err
	}
	if options.checkIn != "" && options.checkOut != "" && options.checkOut < options.checkIn {
		return fmt.Errorf("check-out date must not be before check-in date")
	}
	latSet, lngSet := cmd.Flags().Changed("lat"), cmd.Flags().Changed("lng")
	if latSet != lngSet {
		return fmt.Errorf("--lat and --lng must be provided together")
	}
	if latSet && (options.latitude < -90 || options.latitude > 90 || options.longitude < -180 || options.longitude > 180) {
		return fmt.Errorf("coordinates are outside valid latitude/longitude ranges")
	}
	return nil
}

func newTripsLodgingAddCmd(factory travelWriteClientFactory, options *lodgingWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "add <trip-key>",
		Short: "Add a lodging reservation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(options.name) == "" && strings.TrimSpace(options.placeID) == "" {
				return fmt.Errorf("--name or --place-id is required")
			}
			if options.checkIn == "" || options.checkOut == "" {
				return fmt.Errorf("--check-in and --check-out are required")
			}
			if err := validateLodgingOptions(cmd, options); err != nil {
				return err
			}
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.AddLodgingReservation(wanderlog.AddLodgingReservationRequest{
				TripKey: strings.TrimSpace(args[0]), Name: strings.TrimSpace(options.name), PlaceID: strings.TrimSpace(options.placeID),
				Latitude: options.latitude, Longitude: options.longitude, CheckIn: options.checkIn, CheckOut: options.checkOut,
				ConfirmationNumber: options.confirmationNumber, TravelerNames: options.travelerNames, Notes: options.notes,
			})
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	addLodgingFlags(command, options)
	return command
}

func newTripsLodgingUpdateCmd(factory travelWriteClientFactory, options *lodgingWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <trip-key> <block-id>",
		Short: "Update a lodging reservation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			blockID, err := positiveID(args[1], "block-id")
			if err != nil {
				return err
			}
			if !anyFlagChanged(cmd, "name", "place-id", "lat", "lng", "check-in", "check-out", "confirmation", "traveler", "notes") {
				return fmt.Errorf("at least one lodging field flag is required")
			}
			if err := validateLodgingOptions(cmd, options); err != nil {
				return err
			}
			if cmd.Flags().Changed("section") && options.sectionID <= 0 {
				return fmt.Errorf("--section must be a positive integer")
			}
			request := wanderlog.UpdateLodgingReservationRequest{TripKey: strings.TrimSpace(args[0]), SectionID: options.sectionID, BlockID: blockID}
			request.Name = changedString(cmd, "name", strings.TrimSpace(options.name))
			request.PlaceID = changedString(cmd, "place-id", strings.TrimSpace(options.placeID))
			if cmd.Flags().Changed("lat") {
				request.Latitude, request.Longitude = &options.latitude, &options.longitude
			}
			request.CheckIn = changedString(cmd, "check-in", options.checkIn)
			request.CheckOut = changedString(cmd, "check-out", options.checkOut)
			request.ConfirmationNumber = changedString(cmd, "confirmation", options.confirmationNumber)
			if cmd.Flags().Changed("traveler") {
				request.TravelerNames = &options.travelerNames
			}
			request.Notes = changedString(cmd, "notes", options.notes)
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.UpdateLodgingReservation(request)
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	command.Flags().IntVar(&options.sectionID, "section", 0, "Section ID containing the block (optional)")
	addLodgingFlags(command, options)
	return command
}

func newTripsLodgingDeleteCmd(factory travelWriteClientFactory, options *lodgingWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <trip-key> <block-id>",
		Short: "Delete a lodging reservation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			blockID, err := positiveID(args[1], "block-id")
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("section") && options.sectionID <= 0 {
				return fmt.Errorf("--section must be a positive integer")
			}
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.DeleteLodgingReservation(wanderlog.DeleteTravelReservationRequest{TripKey: strings.TrimSpace(args[0]), SectionID: options.sectionID, BlockID: blockID})
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	command.Flags().IntVar(&options.sectionID, "section", 0, "Section ID containing the block (optional)")
	return command
}

type trainWriteOptions struct {
	output             string
	sectionID          int
	carrier            string
	departurePlaceID   string
	departureName      string
	departureLatitude  float64
	departureLongitude float64
	departureDate      string
	departureTime      string
	arrivalPlaceID     string
	arrivalName        string
	arrivalLatitude    float64
	arrivalLongitude   float64
	arrivalDate        string
	arrivalTime        string
	confirmationNumber string
	notes              string
}

func newTripsTrainWriteCmd(factory travelWriteClientFactory) *cobra.Command {
	options := &trainWriteOptions{}
	command := &cobra.Command{Use: "train", Aliases: []string{"transit"}, Short: "Add train or rail reservations"}
	command.PersistentFlags().StringVarP(&options.output, "output", "o", "pretty", "Output format (pretty, json)")
	command.AddCommand(newTripsTrainAddCmd(factory, options), newTripsTrainDeleteCmd(factory, options))
	return command
}

func newTripsTrainAddCmd(factory travelWriteClientFactory, options *trainWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "add <trip-key>",
		Short: "Add a train reservation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(options.carrier) == "" {
				return fmt.Errorf("--carrier is required")
			}
			if err := validateTrainStop(cmd, "departure", options.departurePlaceID, options.departureName, options.departureLatitude, options.departureLongitude); err != nil {
				return err
			}
			if err := validateTrainStop(cmd, "arrival", options.arrivalPlaceID, options.arrivalName, options.arrivalLatitude, options.arrivalLongitude); err != nil {
				return err
			}
			if err := validateOptionalDate("departure date", options.departureDate); err != nil {
				return err
			}
			if options.departureDate == "" {
				return fmt.Errorf("--departure-date is required")
			}
			if err := validateOptionalDate("arrival date", options.arrivalDate); err != nil {
				return err
			}
			if options.arrivalDate != "" && options.arrivalDate < options.departureDate {
				return fmt.Errorf("arrival date must not be before departure date")
			}
			if err := validateOptionalTime("departure time", options.departureTime); err != nil {
				return err
			}
			if err := validateOptionalTime("arrival time", options.arrivalTime); err != nil {
				return err
			}
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.AddTrainReservation(wanderlog.AddTrainReservationRequest{
				TripKey: strings.TrimSpace(args[0]), Carrier: strings.TrimSpace(options.carrier),
				Departure:     wanderlog.TravelStop{PlaceID: strings.TrimSpace(options.departurePlaceID), Name: strings.TrimSpace(options.departureName), Latitude: options.departureLatitude, Longitude: options.departureLongitude},
				DepartureDate: options.departureDate, DepartureTime: options.departureTime,
				Arrival:     wanderlog.TravelStop{PlaceID: strings.TrimSpace(options.arrivalPlaceID), Name: strings.TrimSpace(options.arrivalName), Latitude: options.arrivalLatitude, Longitude: options.arrivalLongitude},
				ArrivalDate: options.arrivalDate, ArrivalTime: options.arrivalTime, ConfirmationNumber: options.confirmationNumber, Notes: options.notes,
			})
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	command.Flags().StringVar(&options.carrier, "carrier", "", "Carrier or full train designator (for example, SBB EC 317)")
	command.Flags().StringVar(&options.departurePlaceID, "departure-place-id", "", "Google Place ID for the departure station")
	command.Flags().StringVar(&options.departureName, "departure-name", "", "Departure station name")
	command.Flags().Float64Var(&options.departureLatitude, "departure-lat", 0, "Departure latitude")
	command.Flags().Float64Var(&options.departureLongitude, "departure-lng", 0, "Departure longitude")
	command.Flags().StringVar(&options.departureDate, "departure-date", "", "Departure date (YYYY-MM-DD)")
	command.Flags().StringVar(&options.departureTime, "departure-time", "", "Departure time (HH:MM)")
	command.Flags().StringVar(&options.arrivalPlaceID, "arrival-place-id", "", "Google Place ID for the arrival station")
	command.Flags().StringVar(&options.arrivalName, "arrival-name", "", "Arrival station name")
	command.Flags().Float64Var(&options.arrivalLatitude, "arrival-lat", 0, "Arrival latitude")
	command.Flags().Float64Var(&options.arrivalLongitude, "arrival-lng", 0, "Arrival longitude")
	command.Flags().StringVar(&options.arrivalDate, "arrival-date", "", "Arrival date (YYYY-MM-DD; defaults to departure date)")
	command.Flags().StringVar(&options.arrivalTime, "arrival-time", "", "Arrival time (HH:MM)")
	command.Flags().StringVar(&options.confirmationNumber, "confirmation", "", "Confirmation number")
	command.Flags().StringVar(&options.notes, "notes", "", "Reservation notes, coach, or seat")
	_ = command.MarkFlagRequired("carrier")
	_ = command.MarkFlagRequired("departure-date")
	return command
}

func newTripsTrainDeleteCmd(factory travelWriteClientFactory, options *trainWriteOptions) *cobra.Command {
	command := &cobra.Command{
		Use: "delete <trip-key> <block-id>", Short: "Delete a train reservation", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			blockID, err := positiveID(args[1], "block-id")
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("section") && options.sectionID <= 0 {
				return fmt.Errorf("--section must be a positive integer")
			}
			client, err := travelWriteService(factory)
			if err != nil {
				return err
			}
			result, err := client.DeleteTrainReservation(wanderlog.DeleteTravelReservationRequest{TripKey: strings.TrimSpace(args[0]), SectionID: options.sectionID, BlockID: blockID})
			return executeTravelWrite(cmd, options.output, result, err)
		},
	}
	command.Flags().IntVar(&options.sectionID, "section", 0, "Section ID containing the block (optional)")
	return command
}

func validateTrainStop(cmd *cobra.Command, prefix, placeID, name string, latitude, longitude float64) error {
	if strings.TrimSpace(placeID) == "" && strings.TrimSpace(name) == "" {
		return fmt.Errorf("--%s-place-id or --%s-name is required", prefix, prefix)
	}
	latFlag, lngFlag := prefix+"-lat", prefix+"-lng"
	latSet, lngSet := cmd.Flags().Changed(latFlag), cmd.Flags().Changed(lngFlag)
	if latSet != lngSet {
		return fmt.Errorf("--%s and --%s must be provided together", latFlag, lngFlag)
	}
	if strings.TrimSpace(placeID) == "" && (!latSet || !lngSet) {
		return fmt.Errorf("--%s-lat and --%s-lng are required when using --%s-name without a place ID", prefix, prefix, prefix)
	}
	if latSet && (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) {
		return fmt.Errorf("%s coordinates are outside valid latitude/longitude ranges", prefix)
	}
	return nil
}

func init() {
	tripsCmd.AddCommand(
		newTripsFlightWriteCmd(newAuthenticatedTravelWriteClient),
		newTripsLodgingWriteCmd(newAuthenticatedTravelWriteClient),
		newTripsTrainWriteCmd(newAuthenticatedTravelWriteClient),
	)
}
