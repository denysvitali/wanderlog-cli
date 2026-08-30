package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TravelMutationResult identifies the itinerary block changed by a travel
// reservation mutation.
type TravelMutationResult struct {
	Success   bool   `json:"success"`
	TripKey   string `json:"trip_key"`
	SectionID int    `json:"section_id,omitempty"`
	BlockID   int    `json:"block_id,omitempty"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
}

type AddFlightReservationRequest struct {
	TripKey            string
	FlightNumber       string
	DepartureDate      string
	DepartureTime      string
	ArrivalDate        string
	ArrivalTime        string
	ConfirmationNumber string
	Notes              string
	DepartureAirport   string
	ArrivalAirport     string
}

type UpdateFlightReservationRequest struct {
	TripKey            string
	SectionID          int
	BlockID            int
	FlightNumber       *string
	DepartureDate      *string
	DepartureTime      *string
	ArrivalDate        *string
	ArrivalTime        *string
	ConfirmationNumber *string
	Notes              *string
	DepartureAirport   *string
	ArrivalAirport     *string
}

type DeleteTravelReservationRequest struct {
	TripKey   string
	SectionID int
	BlockID   int
}

type AddLodgingReservationRequest struct {
	TripKey            string
	Name               string
	PlaceID            string
	Latitude           float64
	Longitude          float64
	CheckIn            string
	CheckOut           string
	ConfirmationNumber string
	TravelerNames      []string
	Notes              string
}

type UpdateLodgingReservationRequest struct {
	TripKey            string
	SectionID          int
	BlockID            int
	Name               *string
	PlaceID            *string
	Latitude           *float64
	Longitude          *float64
	CheckIn            *string
	CheckOut           *string
	ConfirmationNumber *string
	TravelerNames      *[]string
	Notes              *string
}

type TravelStop struct {
	PlaceID   string
	Name      string
	Latitude  float64
	Longitude float64
}

type AddTrainReservationRequest struct {
	TripKey            string
	Carrier            string
	Departure          TravelStop
	DepartureDate      string
	DepartureTime      string
	Arrival            TravelStop
	ArrivalDate        string
	ArrivalTime        string
	ConfirmationNumber string
	Notes              string
}

func (c *Client) AddFlightReservation(req AddFlightReservationRequest) (*TravelMutationResult, error) {
	return c.AddFlightReservationContext(context.Background(), req)
}

func (c *Client) AddFlightReservationContext(ctx context.Context, req AddFlightReservationRequest) (*TravelMutationResult, error) {
	ctx = normalizeTravelContext(ctx)
	if err := validateTravelMutationContext(ctx, req.TripKey); err != nil {
		return nil, err
	}
	airline, number, err := parseTravelFlightNumber(req.FlightNumber)
	if err != nil {
		return nil, err
	}
	if err := validateTravelDate("departure date", req.DepartureDate, true); err != nil {
		return nil, err
	}
	if err := validateTravelDate("arrival date", req.ArrivalDate, false); err != nil {
		return nil, err
	}
	if req.ArrivalDate == "" {
		req.ArrivalDate = req.DepartureDate
	}
	if err := validateTravelDateTimeRange(req.DepartureDate, req.DepartureTime, req.ArrivalDate, req.ArrivalTime); err != nil {
		return nil, err
	}
	departAirport, arriveAirport := map[string]any(nil), map[string]any(nil)
	stops, stopsErr := c.GetFlightStopsContext(ctx, strconv.Itoa(number), airline, req.DepartureDate)
	if stopsErr == nil && stops != nil && len(stops.Data) > 0 {
		first := stops.Data[0]
		if first.Depart.Airport.IATA != "" {
			departAirport = travelAirportMap(first.Depart.Airport.CityName, first.Depart.Airport.IATA, first.Depart.Airport.Name)
		}
		if first.Arrive.Airport.IATA != "" {
			arriveAirport = travelAirportMap(first.Arrive.Airport.CityName, first.Arrive.Airport.IATA, first.Arrive.Airport.Name)
		}
		if req.ArrivalDate == req.DepartureDate && first.Arrive.Date != "" {
			req.ArrivalDate = first.Arrive.Date
		}
		if req.ArrivalTime == "" {
			req.ArrivalTime = first.Arrive.Time
		}
	}
	if err := validateTravelDateTimeRange(req.DepartureDate, req.DepartureTime, req.ArrivalDate, req.ArrivalTime); err != nil {
		return nil, err
	}
	departAirport = applyTravelAirportOverride(departAirport, req.DepartureAirport)
	arriveAirport = applyTravelAirportOverride(arriveAirport, req.ArrivalAirport)
	if departAirport != nil {
		if place := c.travelGooglePlaceForAirport(ctx, departAirport); place != nil {
			departAirport["googlePlace"] = place
		}
	}
	if arriveAirport != nil {
		if place := c.travelGooglePlaceForAirport(ctx, arriveAirport); place != nil {
			arriveAirport["googlePlace"] = place
		}
	}

	depart := map[string]any{"type": "depart", "date": req.DepartureDate, "time": req.DepartureTime}
	arrive := map[string]any{"type": "arrive", "date": req.ArrivalDate, "time": req.ArrivalTime}
	if departAirport != nil {
		depart["airport"] = departAirport
	}
	if arriveAirport != nil {
		arrive["airport"] = arriveAirport
	}
	block := map[string]any{
		"type": "flight", "confirmationNumber": req.ConfirmationNumber,
		"startTime": req.DepartureTime, "endTime": req.ArrivalTime,
		"depart": depart, "arrive": arrive,
		"flightInfo": map[string]any{"airline": map[string]any{"iata": airline}, "number": number},
		"text":       travelQuillText(req.Notes), "travelerNames": []any{},
	}
	sectionID, blockID, err := c.addTravelBlock(ctx, req.TripKey, "flights", block)
	if err != nil {
		return nil, fmt.Errorf("add flight %s: %w", req.FlightNumber, err)
	}
	return &TravelMutationResult{Success: true, TripKey: req.TripKey, SectionID: sectionID, BlockID: blockID, Kind: "flight", Message: fmt.Sprintf("Added flight %s", strings.ToUpper(strings.TrimSpace(req.FlightNumber)))}, nil
}

func (c *Client) UpdateFlightReservation(req UpdateFlightReservationRequest) (*TravelMutationResult, error) {
	if err := validateTravelBlockRequest(req.TripKey, req.BlockID); err != nil {
		return nil, err
	}
	err := c.replaceTravelBlock(req.TripKey, req.SectionID, req.BlockID, func(block map[string]any) error {
		if block["type"] != "flight" {
			return fmt.Errorf("block %d is not a flight block", req.BlockID)
		}
		if req.FlightNumber != nil {
			airline, number, err := parseTravelFlightNumber(*req.FlightNumber)
			if err != nil {
				return err
			}
			flightInfo := travelChildMap(block, "flightInfo")
			travelChildMap(flightInfo, "airline")["iata"] = airline
			flightInfo["number"] = number
		}
		if req.DepartureDate != nil {
			if err := validateTravelDate("departure date", *req.DepartureDate, true); err != nil {
				return err
			}
			travelChildMap(block, "depart")["date"] = *req.DepartureDate
		}
		if req.DepartureTime != nil {
			block["startTime"] = *req.DepartureTime
			travelChildMap(block, "depart")["time"] = *req.DepartureTime
		}
		if req.ArrivalDate != nil {
			if err := validateTravelDate("arrival date", *req.ArrivalDate, true); err != nil {
				return err
			}
			travelChildMap(block, "arrive")["date"] = *req.ArrivalDate
		}
		if req.ArrivalTime != nil {
			block["endTime"] = *req.ArrivalTime
			travelChildMap(block, "arrive")["time"] = *req.ArrivalTime
		}
		if req.ConfirmationNumber != nil {
			block["confirmationNumber"] = *req.ConfirmationNumber
		}
		if req.Notes != nil {
			block["text"] = travelQuillText(*req.Notes)
		}
		if req.DepartureAirport != nil {
			travelSetAirport(block, "depart", *req.DepartureAirport)
		}
		if req.ArrivalAirport != nil {
			travelSetAirport(block, "arrive", *req.ArrivalAirport)
		}
		if req.DepartureDate != nil || req.DepartureTime != nil || req.ArrivalDate != nil || req.ArrivalTime != nil {
			depart := travelChildMap(block, "depart")
			arrive := travelChildMap(block, "arrive")
			if err := validateTravelDateTimeRange(
				travelString(depart["date"]), travelString(depart["time"]),
				travelString(arrive["date"]), travelString(arrive["time"]),
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("update flight block %d: %w", req.BlockID, err)
	}
	return &TravelMutationResult{Success: true, TripKey: req.TripKey, SectionID: req.SectionID, BlockID: req.BlockID, Kind: "flight", Message: fmt.Sprintf("Updated flight block %d", req.BlockID)}, nil
}

func (c *Client) DeleteFlightReservation(req DeleteTravelReservationRequest) (*TravelMutationResult, error) {
	return c.deleteTravelReservation(req, "flight", false, "flight")
}

func (c *Client) AddLodgingReservation(req AddLodgingReservationRequest) (*TravelMutationResult, error) {
	return c.AddLodgingReservationContext(context.Background(), req)
}

func (c *Client) AddLodgingReservationContext(ctx context.Context, req AddLodgingReservationRequest) (*TravelMutationResult, error) {
	ctx = normalizeTravelContext(ctx)
	if err := validateTravelMutationContext(ctx, req.TripKey); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" && strings.TrimSpace(req.PlaceID) == "" {
		return nil, fmt.Errorf("name or place ID is required")
	}
	if err := validateTravelLodgingRange(req.CheckIn, req.CheckOut); err != nil {
		return nil, err
	}
	place, canonicalName, err := c.resolveTravelPlace(ctx, req.Name, req.PlaceID, req.Latitude, req.Longitude, true)
	if err != nil {
		return nil, err
	}
	hotel := map[string]any{"checkIn": req.CheckIn, "checkOut": req.CheckOut, "travelerNames": req.TravelerNames, "confirmationNumber": nil}
	if req.ConfirmationNumber != "" {
		hotel["confirmationNumber"] = req.ConfirmationNumber
	}
	block := map[string]any{"type": "place", "place": place, "hotel": hotel, "text": travelQuillText(req.Notes), "imageSize": "small", "travelMode": nil, "reactions": []any{}}
	sectionID, blockID, err := c.addTravelBlock(ctx, req.TripKey, "hotels", block)
	if err != nil {
		return nil, fmt.Errorf("add lodging %s: %w", canonicalName, err)
	}
	return &TravelMutationResult{Success: true, TripKey: req.TripKey, SectionID: sectionID, BlockID: blockID, Kind: "lodging", Message: fmt.Sprintf("Added lodging %s", canonicalName)}, nil
}

func (c *Client) UpdateLodgingReservation(req UpdateLodgingReservationRequest) (*TravelMutationResult, error) {
	if err := validateTravelBlockRequest(req.TripKey, req.BlockID); err != nil {
		return nil, err
	}
	err := c.replaceTravelBlock(req.TripKey, req.SectionID, req.BlockID, func(block map[string]any) error {
		if block["type"] != "place" || block["hotel"] == nil {
			return fmt.Errorf("block %d is not a lodging block", req.BlockID)
		}
		hotel := travelChildMap(block, "hotel")
		if req.PlaceID != nil {
			name := ""
			if req.Name != nil {
				name = *req.Name
			}
			lat, lng := 0.0, 0.0
			if req.Latitude != nil {
				lat = *req.Latitude
			}
			if req.Longitude != nil {
				lng = *req.Longitude
			}
			place, _, err := c.resolveTravelPlace(context.Background(), name, *req.PlaceID, lat, lng, false)
			if err != nil {
				return err
			}
			block["place"] = place
		} else {
			place := travelChildMap(block, "place")
			if req.Name != nil {
				place["name"] = *req.Name
			}
			if req.Latitude != nil && req.Longitude != nil {
				place["geometry"] = map[string]any{"location": map[string]any{"lat": *req.Latitude, "lng": *req.Longitude}}
			}
		}
		if req.CheckIn != nil {
			if err := validateTravelDate("check-in date", *req.CheckIn, true); err != nil {
				return err
			}
			hotel["checkIn"] = *req.CheckIn
		}
		if req.CheckOut != nil {
			if err := validateTravelDate("check-out date", *req.CheckOut, true); err != nil {
				return err
			}
			hotel["checkOut"] = *req.CheckOut
		}
		if req.CheckIn != nil || req.CheckOut != nil {
			if err := validateTravelLodgingRange(travelString(hotel["checkIn"]), travelString(hotel["checkOut"])); err != nil {
				return err
			}
		}
		if req.ConfirmationNumber != nil {
			hotel["confirmationNumber"] = *req.ConfirmationNumber
		}
		if req.TravelerNames != nil {
			hotel["travelerNames"] = *req.TravelerNames
		}
		if req.Notes != nil {
			block["text"] = travelQuillText(*req.Notes)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("update lodging block %d: %w", req.BlockID, err)
	}
	return &TravelMutationResult{Success: true, TripKey: req.TripKey, SectionID: req.SectionID, BlockID: req.BlockID, Kind: "lodging", Message: fmt.Sprintf("Updated lodging block %d", req.BlockID)}, nil
}

func (c *Client) DeleteLodgingReservation(req DeleteTravelReservationRequest) (*TravelMutationResult, error) {
	return c.deleteTravelReservation(req, "place", true, "lodging")
}

func (c *Client) AddTrainReservation(req AddTrainReservationRequest) (*TravelMutationResult, error) {
	return c.AddTrainReservationContext(context.Background(), req)
}

func (c *Client) AddTrainReservationContext(ctx context.Context, req AddTrainReservationRequest) (*TravelMutationResult, error) {
	ctx = normalizeTravelContext(ctx)
	if err := validateTravelMutationContext(ctx, req.TripKey); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Carrier) == "" {
		return nil, fmt.Errorf("carrier is required")
	}
	if err := validateTravelDate("departure date", req.DepartureDate, true); err != nil {
		return nil, err
	}
	if err := validateTravelDate("arrival date", req.ArrivalDate, false); err != nil {
		return nil, err
	}
	if req.ArrivalDate == "" {
		req.ArrivalDate = req.DepartureDate
	}
	if err := validateTravelDateTimeRange(req.DepartureDate, req.DepartureTime, req.ArrivalDate, req.ArrivalTime); err != nil {
		return nil, err
	}
	depart, _, err := c.resolveTravelPlace(ctx, req.Departure.Name, req.Departure.PlaceID, req.Departure.Latitude, req.Departure.Longitude, false)
	if err != nil {
		return nil, fmt.Errorf("departure stop: %w", err)
	}
	arrive, _, err := c.resolveTravelPlace(ctx, req.Arrival.Name, req.Arrival.PlaceID, req.Arrival.Latitude, req.Arrival.Longitude, false)
	if err != nil {
		return nil, fmt.Errorf("arrival stop: %w", err)
	}
	block := map[string]any{
		"type": "train", "carrier": strings.TrimSpace(req.Carrier), "confirmationNumber": req.ConfirmationNumber,
		"depart": map[string]any{"place": depart, "date": req.DepartureDate, "time": req.DepartureTime},
		"arrive": map[string]any{"place": arrive, "date": req.ArrivalDate, "time": req.ArrivalTime}, "text": travelQuillText(req.Notes),
	}
	sectionID, blockID, err := c.addTravelBlock(ctx, req.TripKey, "transit", block)
	if err != nil {
		return nil, fmt.Errorf("add train %s: %w", req.Carrier, err)
	}
	return &TravelMutationResult{Success: true, TripKey: req.TripKey, SectionID: sectionID, BlockID: blockID, Kind: "train", Message: fmt.Sprintf("Added train %s", strings.TrimSpace(req.Carrier))}, nil
}

func (c *Client) DeleteTrainReservation(req DeleteTravelReservationRequest) (*TravelMutationResult, error) {
	return c.deleteTravelReservation(req, "train", false, "train")
}

func validateTravelMutationContext(ctx context.Context, tripKey string) error {
	if strings.TrimSpace(tripKey) == "" {
		return fmt.Errorf("trip key is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func normalizeTravelContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func validateTravelBlockRequest(tripKey string, blockID int) error {
	if strings.TrimSpace(tripKey) == "" {
		return fmt.Errorf("trip key is required")
	}
	if blockID <= 0 {
		return fmt.Errorf("block ID must be positive")
	}
	return nil
}

func validateTravelDate(name, value string, required bool) error {
	if value == "" {
		if required {
			return fmt.Errorf("%s is required", name)
		}
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("invalid %s %q: use YYYY-MM-DD", name, value)
	}
	return nil
}

func validateTravelTime(name, value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("15:04", value); err != nil {
		return fmt.Errorf("invalid %s %q: use HH:MM 24-hour time", name, value)
	}
	return nil
}

func validateTravelDateTimeRange(departureDate, departureTime, arrivalDate, arrivalTime string) error {
	if err := validateTravelDate("departure date", departureDate, true); err != nil {
		return err
	}
	if err := validateTravelDate("arrival date", arrivalDate, true); err != nil {
		return err
	}
	if err := validateTravelTime("departure time", departureTime); err != nil {
		return err
	}
	if err := validateTravelTime("arrival time", arrivalTime); err != nil {
		return err
	}

	departure, _ := time.Parse("2006-01-02", departureDate)
	arrival, _ := time.Parse("2006-01-02", arrivalDate)
	if arrival.Before(departure) {
		return fmt.Errorf("arrival date must not be before departure date")
	}
	if arrival.Equal(departure) && departureTime != "" && arrivalTime != "" {
		departureClock, _ := time.Parse("15:04", departureTime)
		arrivalClock, _ := time.Parse("15:04", arrivalTime)
		if arrivalClock.Before(departureClock) {
			return fmt.Errorf("arrival time must not be before departure time on the same date")
		}
	}
	return nil
}

func validateTravelLodgingRange(checkIn, checkOut string) error {
	if err := validateTravelDate("check-in date", checkIn, true); err != nil {
		return err
	}
	if err := validateTravelDate("check-out date", checkOut, true); err != nil {
		return err
	}
	start, _ := time.Parse("2006-01-02", checkIn)
	end, _ := time.Parse("2006-01-02", checkOut)
	if end.Before(start) {
		return fmt.Errorf("check-out date must not be before check-in date")
	}
	return nil
}

func travelString(value any) string {
	text, _ := value.(string)
	return text
}

func parseTravelFlightNumber(value string) (string, int, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	i := 0
	for i < len(normalized) && (normalized[i] < '0' || normalized[i] > '9') {
		i++
	}
	if i == 0 || i == len(normalized) {
		return "", 0, fmt.Errorf("flight number must include an airline code and number, e.g. MU244")
	}
	number, err := strconv.Atoi(normalized[i:])
	if err != nil || number <= 0 {
		return "", 0, fmt.Errorf("flight number must include an airline code and number, e.g. MU244")
	}
	return normalized[:i], number, nil
}

func travelQuillText(text string) map[string]any {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return map[string]any{"ops": []any{map[string]any{"insert": text}}}
}

func travelChildMap(parent map[string]any, key string) map[string]any {
	child, _ := parent[key].(map[string]any)
	if child == nil {
		child = map[string]any{}
		parent[key] = child
	}
	return child
}

func travelSetAirport(block map[string]any, endpoint, code string) {
	airport := travelChildMap(travelChildMap(block, endpoint), "airport")
	code = strings.ToUpper(strings.TrimSpace(code))
	airport["iata"] = code
	if airport["name"] == nil || airport["name"] == "" {
		airport["name"] = code
	}
	if airport["cityName"] == nil || airport["cityName"] == "" {
		airport["cityName"] = code
	}
}

func travelAirportMap(city, iata, name string) map[string]any {
	if name == "" {
		name = iata
	}
	if city == "" {
		city = name
	}
	return map[string]any{"cityName": city, "iata": iata, "name": name}
}

func applyTravelAirportOverride(airport map[string]any, code string) map[string]any {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return airport
	}
	if airport == nil {
		airport = map[string]any{}
	}
	airport["iata"] = code
	if airport["name"] == nil || airport["name"] == "" {
		airport["name"] = code
	}
	if airport["cityName"] == nil || airport["cityName"] == "" {
		airport["cityName"] = airport["name"]
	}
	return airport
}

func (c *Client) travelGooglePlaceForAirport(ctx context.Context, airport map[string]any) map[string]any {
	iata, _ := airport["iata"].(string)
	name, _ := airport["name"].(string)
	city, _ := airport["cityName"].(string)
	query := strings.TrimSpace(city + " " + name + " airport")
	results, err := c.SearchPlacesContext(ctx, query, nil, nil)
	if err != nil || results == nil || !results.Success || len(results.Places) == 0 {
		return nil
	}
	match := results.Places[0]
	for _, candidate := range results.Places {
		if strings.Contains(strings.ToUpper(candidate.Name+" "+candidate.Address), iata) {
			match = candidate
			break
		}
	}
	details, err := c.GetPlaceDetailsContext(ctx, match.PlaceID)
	if err != nil || details == nil || !details.Success {
		return nil
	}
	d := details.Data.Details
	return map[string]any{
		"place_id": d.PlaceID, "name": d.Name, "formatted_address": d.FormattedAddress,
		"rating": d.Rating, "user_ratings_total": d.UserRatingsTotal, "types": d.Types,
		"business_status": d.BusinessStatus, "url": "https://maps.google.com/?" + url.Values{"q": {d.PlaceID}}.Encode(),
		"geometry": map[string]any{"location": map[string]float64{"lat": d.Geometry.Location.Lat, "lng": d.Geometry.Location.Lng}},
	}
}

func (c *Client) resolveTravelPlace(ctx context.Context, name, placeID string, latitude, longitude float64, requireCoordinates bool) (map[string]any, string, error) {
	name, placeID = strings.TrimSpace(name), strings.TrimSpace(placeID)
	if name == "" && placeID == "" {
		return nil, "", fmt.Errorf("name or place ID is required")
	}
	place := map[string]any{"name": name}
	if placeID != "" {
		details, err := c.GetPlaceDetailsContext(ctx, placeID)
		if err != nil {
			return nil, "", fmt.Errorf("fetch place details for %s: %w", placeID, err)
		}
		if details == nil || !details.Success {
			return nil, "", fmt.Errorf("place details for %s returned no data", placeID)
		}
		d := details.Data.Details
		if name == "" {
			name = d.Name
		}
		place = map[string]any{
			"name": name, "place_id": d.PlaceID, "placeId": d.PlaceID,
			"formatted_address": d.FormattedAddress, "rating": d.Rating,
			"user_ratings_total": d.UserRatingsTotal, "website": d.Website,
			"international_phone_number": d.InternationalPhoneNumber, "types": d.Types,
			"business_status": d.BusinessStatus,
			"geometry":        map[string]any{"location": map[string]any{"lat": d.Geometry.Location.Lat, "lng": d.Geometry.Location.Lng}},
		}
	} else if latitude != 0 || longitude != 0 {
		place["geometry"] = map[string]any{"location": map[string]any{"lat": latitude, "lng": longitude}}
	}
	if requireCoordinates {
		if _, ok := place["geometry"]; !ok {
			return nil, "", fmt.Errorf("latitude and longitude are required when adding lodging by name only")
		}
	}
	return place, name, nil
}

func travelSectionPresentation(kind string) (heading, icon, color string) {
	heading, icon, color = "Flights", "plane", "#3498db"
	switch kind {
	case "hotels":
		heading, icon, color = "Hotels and lodging", "bed", "#9b59b6"
	case "transit":
		heading, icon, color = "Transit", "subway", "#17b978"
	}
	return heading, icon, color
}

func findTravelSectionIndex(sections []ItSections, kind, heading string) int {
	for index, section := range sections {
		if section.Type == kind || (kind == "hotels" && section.Type == "lodging") {
			return index
		}
	}
	for index, section := range sections {
		if strings.EqualFold(strings.TrimSpace(section.Heading), heading) {
			return index
		}
	}
	return -1
}

func newTravelSection(kind string, sectionID int, blocks []any) map[string]any {
	heading, icon, color := travelSectionPresentation(kind)
	return map[string]any{
		"id":               sectionID,
		"heading":          heading,
		"type":             kind,
		"mode":             "placeList",
		"placeMarkerColor": color,
		"placeMarkerIcon":  icon,
		"text":             travelQuillText(""),
		"blocks":           blocks,
	}
}

func prepareTravelBlock(block map[string]any, blockID int) (map[string]any, error) {
	rebuiltBlock, err := travelCloneMap(block)
	if err != nil {
		return nil, fmt.Errorf("copy travel block: %w", err)
	}
	rebuiltBlock["id"] = blockID
	rebuiltBlock["addedBy"] = map[string]any{"type": "user"}
	rebuiltBlock["attachments"] = []any{}
	rebuiltBlock["upvotedBy"] = []any{}
	return rebuiltBlock, nil
}

// addTravelBlock atomically creates a missing reservation section with its
// first block, or appends to an existing section. Every conflict rebuilds IDs
// and insertion positions from a fresh snapshot.
func (c *Client) addTravelBlock(ctx context.Context, tripKey, kind string, block map[string]any) (int, int, error) {
	heading, _, _ := travelSectionPresentation(kind)
	sectionID, blockID := 0, 0
	err := c.retryJSON0MutationContext(ctx, tripKey, "AddTravelBlock", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripContext(ctx, tripKey)
		if err != nil {
			return nil, fmt.Errorf("get current trip: %w", err)
		}
		sections := trip.TripPlan.Itinerary.Sections
		maxID := travelMaxItineraryID(trip)
		sectionIdx := findTravelSectionIndex(sections, kind, heading)
		if sectionIdx >= 0 {
			sectionID = sections[sectionIdx].ID
			if sectionID <= 0 {
				return nil, fmt.Errorf("%s section is missing a positive ID", heading)
			}
			blockID = maxID + 1
			rebuiltBlock, err := prepareTravelBlock(block, blockID)
			if err != nil {
				return nil, err
			}
			position := len(sections[sectionIdx].Blocks)
			return []Operation{InsertInList([]interface{}{"itinerary", "sections", sectionIdx, "blocks"}, position, rebuiltBlock)}, nil
		}

		sectionID = maxID + 1
		blockID = maxID + 2
		rebuiltBlock, err := prepareTravelBlock(block, blockID)
		if err != nil {
			return nil, err
		}
		section := newTravelSection(kind, sectionID, []any{rebuiltBlock})
		position := travelSectionInsertPosition(kind, sections)
		return []Operation{InsertInList([]interface{}{"itinerary", "sections"}, position, section)}, nil
	})
	if err != nil {
		return 0, 0, err
	}
	return sectionID, blockID, nil
}

func travelMaxItineraryID(trip *TripResponse) int {
	maxID := 0
	for _, section := range trip.TripPlan.Itinerary.Sections {
		if section.ID > maxID {
			maxID = section.ID
		}
		for _, block := range section.Blocks {
			if block.ID > maxID {
				maxID = block.ID
			}
		}
	}
	return maxID
}

func travelSectionInsertPosition(kind string, sections []ItSections) int {
	order := []string{"textOnly", "attachments", "flights", "hotels", "rentalCars", "transit", "cruise", "normal"}
	index := len(order)
	for i, candidate := range order {
		if candidate == kind {
			index = i
			break
		}
	}
	later := map[string]bool{}
	for _, candidate := range order[index+1:] {
		later[candidate] = true
	}
	for i, section := range sections {
		if later[section.Type] {
			return i
		}
	}
	return len(sections)
}

func (c *Client) appendTravelBlock(ctx context.Context, tripKey string, sectionID int, block map[string]any) (int, error) {
	blockID := 0
	err := c.retryJSON0MutationContext(ctx, tripKey, "AppendTravelBlock", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripContext(ctx, tripKey)
		if err != nil {
			return nil, fmt.Errorf("get current trip: %w", err)
		}
		sectionIdx := FindSectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
		if sectionIdx < 0 {
			return nil, fmt.Errorf("section %d not found", sectionID)
		}
		blockID = travelMaxItineraryID(trip) + 1
		rebuiltBlock, err := prepareTravelBlock(block, blockID)
		if err != nil {
			return nil, err
		}
		position := len(trip.TripPlan.Itinerary.Sections[sectionIdx].Blocks)
		return []Operation{InsertInList([]interface{}{"itinerary", "sections", sectionIdx, "blocks"}, position, rebuiltBlock)}, nil
	})
	if err != nil {
		return 0, err
	}
	return blockID, nil
}

func travelFindRawBlock(trip map[string]any, sectionID, blockID int) (int, int, map[string]any, error) {
	plan, _ := trip["tripPlan"].(map[string]any)
	itinerary, _ := plan["itinerary"].(map[string]any)
	sections, _ := itinerary["sections"].([]any)
	for sectionIdx, sectionAny := range sections {
		section, _ := sectionAny.(map[string]any)
		if section == nil || (sectionID > 0 && travelInt(section["id"]) != sectionID) {
			continue
		}
		blocks, _ := section["blocks"].([]any)
		for blockIdx, blockAny := range blocks {
			block, _ := blockAny.(map[string]any)
			if block != nil && travelInt(block["id"]) == blockID {
				return sectionIdx, blockIdx, block, nil
			}
		}
	}
	return 0, 0, nil, fmt.Errorf("block %d not found", blockID)
}

func travelInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	}
	return 0
}

func travelCloneMap(value map[string]any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) replaceTravelBlock(tripKey string, sectionID, blockID int, update func(map[string]any) error) error {
	return c.replaceTravelBlockContext(context.Background(), tripKey, sectionID, blockID, update)
}

func (c *Client) replaceTravelBlockContext(ctx context.Context, tripKey string, sectionID, blockID int, update func(map[string]any) error) error {
	return c.retryJSON0MutationContext(ctx, tripKey, "UpdateTravelBlock", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripRawContext(ctx, tripKey)
		if err != nil {
			return nil, err
		}
		sectionIdx, blockIdx, oldBlock, err := travelFindRawBlock(trip, sectionID, blockID)
		if err != nil {
			return nil, err
		}
		newBlock, err := travelCloneMap(oldBlock)
		if err != nil {
			return nil, err
		}
		if err := update(newBlock); err != nil {
			return nil, err
		}
		return []Operation{ReplaceInList([]interface{}{"itinerary", "sections", sectionIdx, "blocks"}, blockIdx, oldBlock, newBlock)}, nil
	})
}

func (c *Client) deleteTravelReservation(req DeleteTravelReservationRequest, expectedType string, requireHotel bool, kind string) (*TravelMutationResult, error) {
	return c.deleteTravelReservationContext(context.Background(), req, expectedType, requireHotel, kind)
}

func (c *Client) deleteTravelReservationContext(ctx context.Context, req DeleteTravelReservationRequest, expectedType string, requireHotel bool, kind string) (*TravelMutationResult, error) {
	if err := validateTravelBlockRequest(req.TripKey, req.BlockID); err != nil {
		return nil, err
	}
	err := c.retryJSON0MutationContext(ctx, req.TripKey, "DeleteTravelBlock", func(ctx context.Context) ([]Operation, error) {
		trip, err := c.GetTripRawContext(ctx, req.TripKey)
		if err != nil {
			return nil, err
		}
		sectionIdx, blockIdx, oldBlock, err := travelFindRawBlock(trip, req.SectionID, req.BlockID)
		if err != nil {
			return nil, err
		}
		if oldBlock["type"] != expectedType {
			return nil, fmt.Errorf("block %d has type %q, expected %q", req.BlockID, oldBlock["type"], expectedType)
		}
		if requireHotel {
			if _, ok := oldBlock["hotel"]; !ok {
				return nil, fmt.Errorf("block %d is not a lodging block", req.BlockID)
			}
		}
		return []Operation{DeleteFromList([]interface{}{"itinerary", "sections", sectionIdx, "blocks"}, blockIdx, oldBlock)}, nil
	})
	if err != nil {
		return nil, err
	}
	return &TravelMutationResult{Success: true, TripKey: req.TripKey, SectionID: req.SectionID, BlockID: req.BlockID, Kind: kind, Message: fmt.Sprintf("Deleted %s block %d", kind, req.BlockID)}, nil
}
