package wanderlog

import (
	"encoding/json"
	"fmt"
	"time"
)

// PlaceSearchResponse represents the response from the place search API
type PlaceSearchResponse struct {
	Success bool           `json:"success"`
	Places  []SearchResult `json:"places"`
}

// SearchResult represents a single place result from search
type SearchResult struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	PlaceID     string   `json:"place_id"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Rating      float64  `json:"rating"`
	Categories  []string `json:"categories"`
	Description string   `json:"description"`
	Website     string   `json:"website"`
}

type PlacesListsGeo struct {
	Bounds      []float64 `json:"bounds"`
	CountryName any       `json:"countryName"`
	Depth       int       `json:"depth"`
	ID          int       `json:"id,omitzero"`
	Latitude    float64   `json:"latitude,omitzero"`
	Longitude   float64   `json:"longitude,omitzero"`
	Name        string    `json:"name,omitzero"`
	ParentID    any       `json:"parentId"`
	Popularity  int       `json:"popularity"`
	StateName   any       `json:"stateName"`
	Subcategory string    `json:"subcategory,omitzero"`
}

type Place struct {
	ID        int     `json:"id,omitzero"`
	ImageKey  string  `json:"imageKey,omitzero"`
	Latitude  float64 `json:"latitude,omitzero"`
	Longitude float64 `json:"longitude,omitzero"`
	Name      string  `json:"name,omitzero"`
	PlaceID   string  `json:"placeId,omitzero"`
}

type LoadResources struct {
	BlockID             any      `json:"blockId"`
	ClickbaitText       any      `json:"clickbaitText"`
	ExtendedText        string   `json:"extendedText,omitzero"`
	ID                  string   `json:"id,omitzero"`
	MainText            string   `json:"mainText,omitzero"`
	ScrollToBlockOrLink any      `json:"scrollToBlockOrLink"`
	Tags                []string `json:"tags"`
	Title               string   `json:"title,omitzero"`
	Type                string   `json:"type,omitzero"`
}

type Metadata struct {
	Address              string   `json:"address,omitzero"`
	BusinessStatus       string   `json:"businessStatus,omitzero"`
	Categories           []string `json:"categories"`
	Description          *string  `json:"description"`
	GeneratedDescription *string  `json:"generatedDescription"`
	Geo                  any      `json:"geo"`
	HasDetails           bool     `json:"hasDetails,omitzero"`
	ID                   int      `json:"id,omitzero"`
	ImageKeys            []string `json:"imageKeys"`
	Images               []struct {
		Height int    `json:"height,omitzero"`
		Key    string `json:"key,omitzero"`
		Width  int    `json:"width,omitzero"`
	} `json:"images"`
	InternationalPhoneNumber string  `json:"internationalPhoneNumber,omitzero"`
	MaxMinutesSpent          any     `json:"maxMinutesSpent"`
	MinMinutesSpent          any     `json:"minMinutesSpent"`
	Name                     string  `json:"name,omitzero"`
	NumRatings               int     `json:"numRatings,omitzero"`
	OpeningPeriods           any     `json:"openingPeriods"`
	PermanentlyClosed        bool    `json:"permanentlyClosed"`
	PlaceID                  string  `json:"placeId,omitzero"`
	PlacePageType            string  `json:"placePageType,omitzero"`
	PriceLevel               any     `json:"priceLevel"`
	Rating                   float64 `json:"rating,omitzero"`
	RatingDistribution       struct {
		Google struct {
			One   int `json:"1,omitzero"`
			Two   int `json:"2,omitzero"`
			Three int `json:"3,omitzero"`
			Four  int `json:"4,omitzero"`
			Five  int `json:"5,omitzero"`
		} `json:"Google"`
	} `json:"ratingDistribution"`
	Reviews               []any    `json:"reviews"`
	Sources               []any    `json:"sources"`
	TripadvisorNumRatings *int     `json:"tripadvisorNumRatings"`
	TripadvisorRating     *float64 `json:"tripadvisorRating"`
	UtcOffset             int      `json:"utcOffset,omitzero"`
	Website               string   `json:"website,omitzero"`
}

type POI struct {
	WebPlacesListGeos []struct {
		ID int `json:"id,omitzero"`
	} `json:"WebPlacesListGeos"`
	Bounds      []float64 `json:"bounds"`
	CountryName string    `json:"countryName,omitzero"`
	Depth       int       `json:"depth,omitzero"`
	ID          int       `json:"id,omitzero"`
	Latitude    float64   `json:"latitude,omitzero"`
	Longitude   float64   `json:"longitude,omitzero"`
	Name        string    `json:"name,omitzero"`
	ParentID    int       `json:"parentId,omitzero"`
	Popularity  int       `json:"popularity,omitzero"`
	StateName   string    `json:"stateName,omitzero"`
	Subcategory string    `json:"subcategory,omitzero"`
}

type Geo struct {
	Bounds      []float64 `json:"bounds"`
	CountryName any       `json:"countryName"`
	Depth       int       `json:"depth"`
	ID          int       `json:"id,omitzero"`
	Latitude    float64   `json:"latitude,omitzero"`
	Longitude   float64   `json:"longitude,omitzero"`
	Name        string    `json:"name,omitzero"`
	ParentID    any       `json:"parentId"`
	Popularity  int       `json:"popularity"`
	StateName   any       `json:"stateName"`
	Subcategory string    `json:"subcategory,omitzero"`
}

type Author struct {
	CountriesCount      int    `json:"countriesCount"`
	ID                  int    `json:"id,omitzero"`
	IsProUser           bool   `json:"isProUser"`
	Name                string `json:"name,omitzero"`
	ProfilePictureKey   string `json:"profilePictureKey,omitzero"`
	ShowProfileProBadge bool   `json:"showProfileProBadge"`
	Username            string `json:"username,omitzero"`
	VisitGeosCount      int    `json:"visitGeosCount"`
}

type CarouselItemData struct {
	Author        *Author `json:"author"`
	AuthorBlurb   *string `json:"authorBlurb"`
	Distinction   *string `json:"distinction"`
	IconSiteID    any     `json:"iconSiteId"`
	ID            string  `json:"id,omitzero"`
	LikeCount     int     `json:"likeCount,omitempty,omitzero"`
	PlaceCount    int     `json:"placeCount,omitzero"`
	ProviderTitle *string `json:"providerTitle"`
	SourceSite    string  `json:"sourceSite,omitzero"`
	Title         string  `json:"title,omitzero"`
	TopImageKey   string  `json:"topImageKey,omitzero"`
	TopImageType  string  `json:"topImageType,omitzero"`
	TripPlanKey   *string `json:"tripPlanKey"`
	Type          string  `json:"type,omitzero"`
	ViewCount     int     `json:"viewCount,omitempty,omitzero"`
}

type CarouselItems struct {
	Data CarouselItemData `json:"data"`
	Type string           `json:"type,omitzero"`
}

type Resources struct {
	Ancestors                   []any `json:"ancestors"`
	CurrencyRatesUsd            map[string]float64
	DefaultPlacesListsGeo       PlacesListsGeo     `json:"defaultPlacesListsGeo"`
	DistancesBetweenPlaces      map[string]any     `json:"distancesBetweenPlaces"`
	DistancesBetweenPlacesError any                `json:"distancesBetweenPlacesError"`
	ExploreCarouselItems        []CarouselItems    `json:"exploreCarouselItems"`
	FlightUpdates               struct{}           `json:"flightUpdates"`
	Geo                         Geo                `json:"geo"`
	Geos                        []Geo              `json:"geos"`
	HotelDeals                  []any              `json:"hotelDeals"`
	LiveFlightUpdates           struct{}           `json:"liveFlightUpdates"`
	Nearby                      []POI              `json:"nearby"`
	PlaceMetadata               []Metadata         `json:"placeMetadata"`
	SectionRecommendations      map[string][]Place `json:"sectionRecommendations"`
	TipsOnLoadResources         []LoadResources    `json:"tipsOnLoadResources"`
	TopPlace                    Place              `json:"topPlace"`
}

type GuidedResources struct {
	RelatedGuides []any `json:"relatedGuides"`
}

type TripPlanKey struct {
	CreatedAt        time.Time `json:"createdAt,omitzero"`
	ID               int       `json:"id,omitzero"`
	Key              string    `json:"key,omitzero"`
	ShowNotes        bool      `json:"showNotes,omitzero"`
	ShowReservations bool      `json:"showReservations"`
	TripPlanID       int       `json:"tripPlanId,omitzero"`
	UpdatedAt        time.Time `json:"updatedAt,omitzero"`
}

type CurrencyAmount struct {
	Amount       int    `json:"amount"`
	CurrencyCode string `json:"currencyCode,omitzero"`
}

type Budget struct {
	Amount       CurrencyAmount `json:"amount"`
	Expenses     []any          `json:"expenses"`
	Payments     []any          `json:"payments"`
	SimplifyDebt bool           `json:"simplifyDebt"`
}

type By struct {
	Type   string `json:"type,omitzero"`
	UserID int    `json:"userId,omitzero"`
}

type GooglePlace struct {
	AddressComponents []struct {
		LongName  string   `json:"long_name,omitzero"`
		ShortName string   `json:"short_name,omitzero"`
		Types     []string `json:"types"`
	} `json:"address_components"`
	AdrAddress           string `json:"adr_address,omitempty,omitzero"`
	BusinessStatus       string `json:"business_status,omitzero"`
	FormattedAddress     string `json:"formatted_address,omitzero"`
	FormattedPhoneNumber string `json:"formatted_phone_number,omitempty,omitzero"`
	Geometry             struct {
		Location struct {
			Lat float64 `json:"lat,omitzero"`
			Lng float64 `json:"lng,omitzero"`
		} `json:"location"`
		Viewport *struct {
			East  float64 `json:"east,omitzero"`
			North float64 `json:"north,omitzero"`
			South float64 `json:"south,omitzero"`
			West  float64 `json:"west,omitzero"`
		} `json:"viewport,omitempty,omitzero"`
	} `json:"geometry"`
	Icon                     string   `json:"icon,omitempty,omitzero"`
	InternationalPhoneNumber string   `json:"international_phone_number,omitzero"`
	Name                     string   `json:"name,omitzero"`
	PhotoURLs                []string `json:"photo_urls,omitempty"`
	Photos                   []struct {
		Height           int      `json:"height,omitzero"`
		HtmlAttributions []string `json:"html_attributions"`
		PhotoReference   string   `json:"photo_reference,omitzero"`
		Width            int      `json:"width,omitzero"`
	} `json:"photos,omitempty"`
	PlaceID  string `json:"place_id,omitzero"`
	PlusCode *struct {
		CompoundCode string `json:"compound_code,omitzero"`
		GlobalCode   string `json:"global_code,omitzero"`
	} `json:"plus_code,omitempty,omitzero"`
	Rating    float64 `json:"rating,omitzero"`
	Reference string  `json:"reference,omitempty,omitzero"`
	Reviews   []struct {
		AuthorName              string `json:"author_name,omitzero"`
		AuthorURL               string `json:"author_url,omitzero"`
		Language                string `json:"language,omitzero"`
		ProfilePhotoURL         string `json:"profile_photo_url,omitzero"`
		Rating                  int    `json:"rating,omitzero"`
		RelativeTimeDescription string `json:"relative_time_description,omitzero"`
		Text                    string `json:"text,omitzero"`
		Time                    int    `json:"time,omitzero"`
	} `json:"reviews,omitempty"`
	Types            []string `json:"types"`
	URL              string   `json:"url,omitzero"`
	UserRatingsTotal int      `json:"user_ratings_total,omitzero"`
	UtcOffset        int      `json:"utc_offset,omitempty,omitzero"`
	Vicinity         string   `json:"vicinity,omitzero"`
	Website          string   `json:"website,omitzero"`
}

type Station struct {
	Airport struct {
		CityName    string      `json:"cityName,omitzero"`
		GooglePlace GooglePlace `json:"googlePlace"`
		Iata        string      `json:"iata,omitzero"`
		Name        string      `json:"name,omitzero"`
	} `json:"airport"`
	Date  string      `json:"date,omitzero"`
	Place *BlockPlace `json:"place,omitempty"`
	Time  string      `json:"time,omitzero"`
	Type  string      `json:"type,omitzero"`
}

type Flight struct {
	Airline struct {
		Iata          string `json:"iata,omitzero"`
		Icao          string `json:"icao,omitzero"`
		LocalizedName string `json:"localizedName,omitzero"`
		Name          string `json:"name,omitzero"`
	} `json:"airline"`
	Number int `json:"number,omitzero"`
}

type Text struct {
	Ops []struct {
		Attributes *struct {
			Bold bool   `json:"bold,omitempty"`
			Link string `json:"link,omitempty"`
			List string `json:"list,omitzero"`
		} `json:"attributes,omitempty,omitzero"`
		Insert string `json:"insert,omitzero"`
	} `json:"ops"`
}

// FlexibleText can be either a string or a Text object
type FlexibleText struct {
	IsString bool
	String   string
	Text     Text
}

// UnmarshalJSON implements custom unmarshaling for FlexibleText
func (ft *FlexibleText) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		ft.IsString = true
		ft.String = str
		return nil
	}

	// Try to unmarshal as Text object
	var text Text
	if err := json.Unmarshal(data, &text); err == nil {
		ft.IsString = false
		ft.Text = text
		return nil
	}

	return fmt.Errorf("FlexibleText: cannot unmarshal as string or Text object")
}

// MarshalJSON implements custom marshaling for FlexibleText
func (ft FlexibleText) MarshalJSON() ([]byte, error) {
	if ft.IsString {
		return json.Marshal(ft.String)
	}
	return json.Marshal(ft.Text)
}

type ItSections struct {
	Blocks []struct {
		AddedBy            By       `json:"addedBy"`
		Arrive             *Station `json:"arrive,omitempty,omitzero"`
		Attachments        []any    `json:"attachments"`
		Carrier            string   `json:"carrier,omitempty"`
		ConfirmationNumber string   `json:"confirmationNumber"`
		Depart             Station  `json:"depart,omitempty,omitzero"`
		EndTime            string   `json:"endTime,omitempty"`
		FlightInfo         *Flight  `json:"flightInfo,omitempty,omitzero"`
		Hotel              *struct {
			CheckIn            string `json:"checkIn"`
			CheckOut           string `json:"checkOut"`
			TravelerNames      []any  `json:"travelerNames"`
			ConfirmationNumber string `json:"confirmationNumber"`
		} `json:"hotel,omitempty"`
		ID               int          `json:"id,omitzero"`
		ImageKeys        []string     `json:"imageKeys,omitempty"`
		ImageSize        string       `json:"imageSize,omitempty"`
		NoteIcon         string       `json:"noteIcon,omitempty"`
		Place            *BlockPlace  `json:"place,omitempty"`
		SelectedImageKey string       `json:"selectedImageKey,omitempty"`
		StartTime        string       `json:"startTime,omitempty"`
		Text             FlexibleText `json:"text"`
		TravelMode       *string      `json:"travelMode"`
		TravelerNames    []any        `json:"travelerNames"`
		Type             string       `json:"type,omitzero"`
		UpvotedBy        []any        `json:"upvotedBy"`
	} `json:"blocks"`
	Date             *string `json:"date"`
	Heading          string  `json:"heading"`
	DisplayHeading   string  `json:"displayHeading"` // Alternative heading field from sections endpoint
	ID               int     `json:"id,omitzero"`
	Mode             string  `json:"mode,omitzero"`
	PlaceMarkerColor string  `json:"placeMarkerColor,omitzero"`
	PlaceMarkerIcon  string  `json:"placeMarkerIcon,omitzero"`
	Text             Text    `json:"text"`
	Type             string  `json:"type,omitzero"`
}

type Itinerary struct {
	Budget  Budget `json:"budget"`
	Journal struct {
		Stops   []any  `json:"stops"`
		Summary string `json:"summary"`
	} `json:"journal"`
	Options  struct{}     `json:"options"`
	Sections []ItSections `json:"sections"`
}

type Plan struct {
	TripPlanKeys              []TripPlanKey `json:"TripPlanKeys"`
	TripPlanViews             []any         `json:"TripPlanViews"`
	TripPlanJournalViews      []any         `json:"TripPlanJournalViews"`
	AuthorBlurb               string        `json:"authorBlurb"`
	CharacterCount            int           `json:"characterCount,omitzero"`
	Contributors              []Author      `json:"contributors"`
	CreatedAt                 time.Time     `json:"createdAt,omitzero"`
	Days                      int           `json:"days,omitzero"`
	DeletedAt                 any           `json:"deletedAt"`
	Distinction               any           `json:"distinction"`
	EditKey                   string        `json:"editKey,omitzero"`
	EditedAt                  time.Time     `json:"editedAt,omitzero"`
	Editors                   []Author      `json:"editors"`
	EndDate                   string        `json:"endDate,omitzero"`
	GooglePlaceListID         any           `json:"googlePlaceListId"`
	HasSentCreateGuideEmail   bool          `json:"hasSentCreateGuideEmail"`
	HeaderImageID             int           `json:"headerImageId,omitzero"`
	HeaderImageKey            string        `json:"headerImageKey,omitzero"`
	ID                        int           `json:"id,omitzero"`
	ImageCount                any           `json:"imageCount"`
	IsHeaderImageUserSelected bool          `json:"isHeaderImageUserSelected"`
	IsHighQuality             bool          `json:"isHighQuality,omitzero"`
	IsMapEmbed                bool          `json:"isMapEmbed"`
	Itinerary                 Itinerary     `json:"itinerary"`
	JournalKey                *string       `json:"journalKey"`
	JournalStopCount          *int          `json:"journalStopCount"`
	Key                       string        `json:"key,omitzero"`
	LastSentUpdateEmailAt     time.Time     `json:"lastSentUpdateEmailAt,omitzero"`
	LikeCount                 int           `json:"likeCount"`
	OverallVersion            int           `json:"overallVersion,omitzero"`
	PlaceCount                int           `json:"placeCount,omitzero"`
	PlaceListAccountID        any           `json:"placeListAccountId"`
	Privacy                   string        `json:"privacy,omitzero"`
	PublishedAt               any           `json:"publishedAt"`
	RequiresVerification      bool          `json:"requiresVerification"`
	SampleType                string        `json:"sampleType,omitzero"`
	SchemaVersion             int           `json:"schemaVersion,omitzero"`
	StartDate                 string        `json:"startDate,omitzero"`
	SuggestKey                string        `json:"suggestKey,omitzero"`
	Title                     string        `json:"title,omitzero"`
	TopImageKeys              []any         `json:"topImageKeys"`
	Type                      string        `json:"type,omitzero"`
	UpdatedAt                 time.Time     `json:"updatedAt,omitzero"`
	UpdatedDistinctionAt      any           `json:"updatedDistinctionAt"`
	UpdatedDistinctionUserID  any           `json:"updatedDistinctionUserId"`
	UserID                    int           `json:"userId,omitzero"`
	ViewCount                 int           `json:"viewCount,omitzero"`
	ViewKey                   string        `json:"viewKey,omitzero"`
	WebPlacesListID           any           `json:"webPlacesListId"`
}

type BlockPlace struct {
	Name             string  `json:"name,omitempty"`
	PlaceID          string  `json:"place_id,omitempty"`
	Rating           float64 `json:"rating,omitempty"`
	FormattedAddress string  `json:"formatted_address,omitempty"`
	Geometry         *struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry,omitempty"`
	UserRatingsTotal         int    `json:"user_ratings_total,omitempty"`
	Website                  string `json:"website,omitempty"`
	InternationalPhoneNumber string `json:"international_phone_number,omitempty"`
	AddressComponents        []struct {
		LongName  string   `json:"long_name,omitempty"`
		ShortName string   `json:"short_name,omitempty"`
		Types     []string `json:"types"`
	} `json:"address_components,omitempty"`
	OpeningHours *struct {
		WeekdayText []string `json:"weekday_text,omitempty"`
		Periods     []struct {
			Open struct {
				Day  int    `json:"day"`
				Time string `json:"time"`
			} `json:"open"`
			Close struct {
				Day  int    `json:"day"`
				Time string `json:"time"`
			} `json:"close"`
		} `json:"periods,omitempty"`
	} `json:"opening_hours,omitempty"`
	Vicinity       string          `json:"vicinity,omitempty"`
	Types          []string        `json:"types,omitempty"`
	URL            string          `json:"url,omitempty"`
	BusinessStatus string          `json:"business_status,omitempty"`
	PhotoURLs      []string        `json:"photo_urls,omitempty"`
	Amenities      map[string]bool `json:"amenities,omitempty"`
	PriceLevel     *int            `json:"price_level,omitempty"`
}

type TripResponse struct {
	GuideResources GuidedResources `json:"guideResources"`
	Resources      Resources       `json:"resources"`
	Settings       struct{}        `json:"settings"`
	Success        bool            `json:"success,omitzero"`
	Error          string          `json:"error,omitempty"`
	TripPlan       Plan            `json:"tripPlan"`
}

// AirlinesResponse represents the response from the all airlines API
type AirlinesResponse struct {
	Success bool      `json:"success"`
	Data    []Airline `json:"data"`
}

// Airline represents an airline with IATA/ICAO codes and names
type Airline struct {
	Iata          string `json:"iata,omitempty"`
	Icao          string `json:"icao,omitempty"`
	Name          string `json:"name,omitempty"`
	LocalizedName string `json:"localizedName,omitempty"`
}

// AirportAutocompleteResponse represents the response from the airport autocomplete API
type AirportAutocompleteResponse struct {
	Success bool                `json:"success"`
	Data    []AirportSuggestion `json:"data"`
}

// AirportSuggestion represents a single airport suggestion
type AirportSuggestion struct {
	IATA     string `json:"iata"`
	Name     string `json:"name"`
	CityName string `json:"cityName"`
}

// FlightStopsResponse represents the response from the flight stops API
type FlightStopsResponse struct {
	Success bool         `json:"success"`
	Data    []FlightStop `json:"data"`
}

// FlightStop represents a single flight stop (leg)
type FlightStop struct {
	Depart FlightStopEndpoint `json:"depart"`
	Arrive FlightStopEndpoint `json:"arrive"`
}

// FlightStopEndpoint represents the departure or arrival of a flight stop
type FlightStopEndpoint struct {
	Type   string            `json:"type"`
	Date   string            `json:"date"`
	Time   string            `json:"time"`
	Airport FlightStopAirport `json:"airport"`
}

// FlightStopAirport represents airport info in a flight stop
type FlightStopAirport struct {
	IATA     string `json:"iata"`
	Name     string `json:"name"`
	CityName string `json:"cityName"`
}

// LodgingSearchResponse represents the response from the lodging search API
type LodgingSearchResponse struct {
	Success bool              `json:"success"`
	Data    []LodgingProperty `json:"data"`
}

// LodgingProperty represents a single lodging/hotel result
type LodgingProperty struct {
	PropertyID    string  `json:"propertyId"`
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	Country       string  `json:"country"`
	Rating        float64 `json:"rating,omitempty"`
	PricePerNight string  `json:"pricePerNight,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	ImageURL      string  `json:"imageUrl,omitempty"`
	BookerType    string  `json:"bookerType,omitempty"` // e.g., "google", "expedia", "hotels.com"
}

// GooglePriceRatesResponse represents the response from the Google price rates API
type GooglePriceRatesResponse struct {
	Success bool `json:"success"`
	Data    struct {
		PropertyID string `json:"propertyId"`
		Rates      []struct {
			BookerType string `json:"bookerType"`
			Price      string `json:"price"`
			Currency   string `json:"currency"`
			URL        string `json:"url"`
		} `json:"rates"`
	} `json:"data"`
}
