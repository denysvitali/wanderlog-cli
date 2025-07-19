package wanderlog

import (
	"time"
)

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
			_1 int `json:"1,omitzero"`
			_2 int `json:"2,omitzero"`
			_3 int `json:"3,omitzero"`
			_4 int `json:"4,omitzero"`
			_5 int `json:"5,omitzero"`
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
	DefaultPlacesListsGeo       PlacesListsGeo   `json:"defaultPlacesListsGeo"`
	DistancesBetweenPlaces      struct{}         `json:"distancesBetweenPlaces"`
	DistancesBetweenPlacesError any              `json:"distancesBetweenPlacesError"`
	ExploreCarouselItems        []CarouselItems  `json:"exploreCarouselItems"`
	FlightUpdates               struct{}         `json:"flightUpdates"`
	Geo                         Geo              `json:"geo"`
	Geos                        []Geo            `json:"geos"`
	HotelDeals                  []any            `json:"hotelDeals"`
	LiveFlightUpdates           struct{}         `json:"liveFlightUpdates"`
	Nearby                      []POI            `json:"nearby"`
	PlaceMetadata               []Metadata       `json:"placeMetadata"`
	SectionRecommendations      map[string][]Place `json:"sectionRecommendations"`
	TipsOnLoadResources         []LoadResources  `json:"tipsOnLoadResources"`
	TopPlace                    Place            `json:"topPlace"`
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
	Date string `json:"date,omitzero"`
	Time string `json:"time,omitzero"`
	Type string `json:"type,omitzero"`
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
			List string `json:"list,omitzero"`
		} `json:"attributes,omitempty,omitzero"`
		Insert string `json:"insert,omitzero"`
	} `json:"ops"`
}

type ItSections struct {
	Blocks []struct {
		AddedBy            By       `json:"addedBy"`
		Arrive             *Station `json:"arrive,omitempty,omitzero"`
		Attachments        []any    `json:"attachments"`
		ConfirmationNumber string   `json:"confirmationNumber"`
		Depart             Station  `json:"depart,omitempty,omitzero"`
		FlightInfo         *Flight  `json:"flightInfo,omitempty,omitzero"`
		ID                 int      `json:"id,omitzero"`
		Text               Text     `json:"text"`
		TravelerNames      []any    `json:"travelerNames"`
		Type               string   `json:"type,omitzero"`
	} `json:"blocks"`
	Date             *string `json:"date"`
	Heading          string  `json:"heading"`
	ID               int     `json:"id,omitzero"`
	Mode             string  `json:"mode,omitzero"`
	PlaceMarkerColor string  `json:"placeMarkerColor,omitzero"`
	PlaceMarkerIcon  string  `json:"placeMarkerIcon,omitzero"`
	Text             Text    `json:"text"`
	Type             string  `json:"type,omitzero"`
}

type Itinerary struct {
	Budget   Budget       `json:"budget"`
	Options  struct{}     `json:"options"`
	Sections []ItSections `json:"sections"`
}

type Plan struct {
	TripPlanKeys              []TripPlanKey `json:"TripPlanKeys"`
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

type TripResponse struct {
	GuideResources GuidedResources `json:"guideResources"`
	Resources      Resources       `json:"resources"`
	Settings       struct{}        `json:"settings"`
	Success        bool            `json:"success,omitzero"`
	TripPlan       Plan            `json:"tripPlan"`
}
