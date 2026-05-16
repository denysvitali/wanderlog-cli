package wanderlog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

type GeoEnvelope = models.GeoEnvelope

// ListGeosWithGoodGuides returns curated geos that have high-quality guides.
func (c *Client) ListGeosWithGoodGuides() (*GeoEnvelope, error) {
	return c.geoGet("ListGeosWithGoodGuides", "geo/geosWithGoodGuides", nil)
}

// ListPopularAndNearbyGeos returns popular and nearby geos for the calling user.
func (c *Client) ListPopularAndNearbyGeos() (*GeoEnvelope, error) {
	return c.geoGet("ListPopularAndNearbyGeos", "geo/popularAndNearby", nil)
}

// FindCountryForIP returns the country geo associated with the caller's IP.
func (c *Client) FindCountryForIP() (*GeoEnvelope, error) {
	return c.geoGet("FindCountryForIP", "geo/findCountryForIP", nil)
}

// FindNearestTripadvisorGeo returns the nearest Tripadvisor-mapped geo for a coordinate.
func (c *Client) FindNearestTripadvisorGeo(lat, lng float64) (*GeoEnvelope, error) {
	return c.geoGet("FindNearestTripadvisorGeo", "geo/nearestTripadvisorGeo", apiQuery(map[string]string{
		"latitude":  strconv.FormatFloat(lat, 'f', -1, 64),
		"longitude": strconv.FormatFloat(lng, 'f', -1, 64),
	}))
}

// FindNearestGeosToIP returns the nearest geos for the caller's IP, optionally
// limited by query params (radius, kind) which the server understands as
// `params`.
func (c *Client) FindNearestGeosToIP(extraParams map[string]string) (*GeoEnvelope, error) {
	return c.geoGet("FindNearestGeosToIP", "geo/findNearestGeosToIP", apiQuery(extraParams))
}

// FindNearestKayakCity maps a place coordinate + name to the nearest Kayak
// city, used for hotel-search backends that take a Kayak city id.
func (c *Client) FindNearestKayakCity(lat, lng float64, cityNameToMatch string) (*GeoEnvelope, error) {
	return c.geoGet("FindNearestKayakCity", "geo/nearestKayakCity", apiQuery(map[string]string{
		"latitude":        strconv.FormatFloat(lat, 'f', -1, 64),
		"longitude":       strconv.FormatFloat(lng, 'f', -1, 64),
		"cityNameToMatch": cityNameToMatch,
	}))
}

// GetClientGeos fetches metadata for one or more geos by ID.
func (c *Client) GetClientGeos(geoIDs []int) (*GeoEnvelope, error) {
	if len(geoIDs) == 0 {
		return nil, fmt.Errorf("GetClientGeos: at least one geo id is required")
	}
	parts := make([]string, 0, len(geoIDs))
	for _, id := range geoIDs {
		parts = append(parts, strconv.Itoa(id))
	}
	return c.geoGet("GetClientGeos", "geo/clientGeos", apiQuery(map[string]string{
		"geoIds": strings.Join(parts, ","),
	}))
}

// ListTripPlannerGeos returns geos that have trip-planner content available.
func (c *Client) ListTripPlannerGeos() (*GeoEnvelope, error) {
	return c.geoGet("ListTripPlannerGeos", "geo/tripPlannerGeos", nil)
}

// ListCountries returns the country list for a UI language.
func (c *Client) ListCountries(language string) (*GeoEnvelope, error) {
	return c.geoGet("ListCountries", "geo/countries", apiQuery(map[string]string{
		"language": language,
	}))
}

// ListGeoCategoriesForCategory returns geos that have content for the given
// keyword category (e.g. "best-museums").
func (c *Client) ListGeoCategoriesForCategory(keywordCategoryID int, language string) (*GeoEnvelope, error) {
	return c.geoGet("ListGeoCategoriesForCategory", "geo/listGeoCategoriesForCategory", apiQuery(map[string]string{
		"keywordCategoryId": strconv.Itoa(keywordCategoryID),
		"language":          language,
	}))
}

// ListGeoCategoriesForGeo returns the categories that have content for the
// given geo (e.g. "Tokyo > best museums, best ramen, ...").
func (c *Client) ListGeoCategoriesForGeo(geoID int, source string) (*GeoEnvelope, error) {
	return c.geoGet("ListGeoCategoriesForGeo", "geo/listGeoCategoriesForGeo", apiQuery(map[string]string{
		"geoId":  strconv.Itoa(geoID),
		"source": source,
	}))
}

// ListGeoInMonthGeos returns "best places to visit in <month>" content.
func (c *Client) ListGeoInMonthGeos() (*GeoEnvelope, error) {
	return c.geoGet("ListGeoInMonthGeos", "geo/geoInMonthGeos", nil)
}

// ListKeywordCategories returns the top-level keyword taxonomy
// (e.g. museums, restaurants, beaches).
func (c *Client) ListKeywordCategories(language string) (*GeoEnvelope, error) {
	return c.geoGet("ListKeywordCategories", "geo/listKeywordCategories", apiQuery(map[string]string{
		"language": language,
	}))
}

// SearchGeo runs a free-form geo search. Pass arbitrary query parameters such
// as `q`, `language`, etc.
func (c *Client) SearchGeo(params map[string]string) (*GeoEnvelope, error) {
	return c.geoGet("SearchGeo", "geo/search", apiQuery(params))
}

func (c *Client) geoGet(opName, path string, query url.Values) (*GeoEnvelope, error) {
	resp, err := c.apiRequest(context.Background(), http.MethodGet, path, query, nil, false)
	if err != nil {
		return nil, err
	}
	var result GeoEnvelope
	if err := decodeAPIBody(opName, resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
