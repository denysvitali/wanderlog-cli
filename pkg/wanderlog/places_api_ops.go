package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

// autocompleteSessionToken returns a session token for placesAPI/autocomplete
// requests. Tests override this to make request URLs deterministic.
var autocompleteSessionToken = func() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

type (
	PlacesAPIEnvelope          = models.PlacesAPIEnvelope
	FindPlaceFromLngLatRequest = models.FindPlaceFromLngLatRequest
	PlacesMetadataResponse     = models.PlacesMetadataResponse
	PlacesCardResponse         = models.PlacesCardResponse
)

// AutocompletePlaces calls the Google-style placesAPI autocomplete (v1).
// Use SearchPlacesWithWanderlog for the v2 / Wanderlog autocomplete that also
// returns Wanderlog suggestions.
func (c *Client) AutocompletePlaces(query string, lat, lng float64) (*WanderlogAutocompleteResponse, error) {
	reqData := WanderlogAutocompleteRequest{
		Input:        query,
		SessionToken: autocompleteSessionToken(),
		Location: struct {
			Longitude float64 `json:"longitude"`
			Latitude  float64 `json:"latitude"`
		}{Longitude: lng, Latitude: lat},
		Radius:   50000,
		Language: "en",
	}
	encoded, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("AutocompletePlaces: marshaling request: %w", err)
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "placesAPI/autocomplete", apiQuery(map[string]string{
		"request": string(encoded),
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result WanderlogAutocompleteResponse
	if err := decodeAPIBody("AutocompletePlaces", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FindPlaceFromLngLat reverse-geocodes a coordinate to a Wanderlog/Google
// place suggestion.
func (c *Client) FindPlaceFromLngLat(lat, lng float64) (*PlacesAPIEnvelope, error) {
	encoded, err := json.Marshal(FindPlaceFromLngLatRequest{Longitude: lng, Latitude: lat})
	if err != nil {
		return nil, fmt.Errorf("FindPlaceFromLngLat: marshaling request: %w", err)
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "placesAPI/findPlaceFromLngLat", apiQuery(map[string]string{
		"request": string(encoded),
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesAPIEnvelope
	if err := decodeAPIBody("FindPlaceFromLngLat", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMapLayerGroups returns places grouped into the map's category layers for
// a viewport / set of categories. The request payload shape is opaque; pass
// the bundle's expected fields (bbox, layerGroupIds, ...).
func (c *Client) GetMapLayerGroups(payload any) (*PlacesAPIEnvelope, error) {
	if payload == nil {
		return nil, fmt.Errorf("GetMapLayerGroups: payload is required")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("GetMapLayerGroups: marshaling payload: %w", err)
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "placesAPI/mapLayerGroups", apiQuery(map[string]string{
		"request": string(encoded),
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesAPIEnvelope
	if err := decodeAPIBody("GetMapLayerGroups", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMultiplePlaceDetails fetches detail records for a batch of Google place
// IDs in one call.
func (c *Client) GetMultiplePlaceDetails(placeIDs []string, language string) (*PlacesAPIEnvelope, error) {
	if len(placeIDs) == 0 {
		return nil, fmt.Errorf("GetMultiplePlaceDetails: at least one placeId is required")
	}
	if language == "" {
		language = "en"
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "placesAPI/getMultiplePlaceDetails", apiQuery(map[string]string{
		"placeIds": strings.Join(placeIDs, ","),
		"language": language,
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesAPIEnvelope
	if err := decodeAPIBody("GetMultiplePlaceDetails", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPlaceDetailsV2 returns the detail-only view for a single place. Compared
// to GetPlaceDetails, this v2 endpoint does not include card data.
func (c *Client) GetPlaceDetailsV2(placeID, language string) (*PlacesAPIEnvelope, error) {
	if placeID == "" {
		return nil, fmt.Errorf("GetPlaceDetailsV2: placeId is required")
	}
	if language == "" {
		language = "en"
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "placesAPI/getPlaceDetails/v2", apiQuery(map[string]string{
		"placeId":  placeID,
		"language": language,
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesAPIEnvelope
	if err := decodeAPIBody("GetPlaceDetailsV2", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SearchPlacesGoogle runs a Google-style place search. Pass an opaque payload
// matching the bundle's expected shape (input/text query, location bias, ...).
func (c *Client) SearchPlacesGoogle(payload any) (*PlacesAPIEnvelope, error) {
	if payload == nil {
		return nil, fmt.Errorf("SearchPlacesGoogle: payload is required")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("SearchPlacesGoogle: marshaling payload: %w", err)
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "placesAPI/searchPlaces", apiQuery(map[string]string{
		"request": string(encoded),
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesAPIEnvelope
	if err := decodeAPIBody("SearchPlacesGoogle", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPlacesMetadata fetches Google place metadata in bulk. Optional query
// parameters scope the call to a list/geo (`geoId`, `listId`, `listType`,
// `getDetails` flag).
func (c *Client) GetPlacesMetadata(placeIDs []string, opts map[string]string) (*PlacesMetadataResponse, error) {
	if len(placeIDs) == 0 {
		return nil, fmt.Errorf("GetPlacesMetadata: at least one placeId is required")
	}
	params := map[string]string{"placeIds": strings.Join(placeIDs, ",")}
	for k, v := range opts {
		if v != "" {
			params[k] = v
		}
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "places/metadata", apiQuery(params), nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesMetadataResponse
	if err := decodeAPIBody("GetPlacesMetadata", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPlaceCards fetches the lightweight place-card data (Wanderlog summaries,
// quick highlights) for a batch of place IDs.
func (c *Client) GetPlaceCards(placeIDs []string) (*PlacesCardResponse, error) {
	if len(placeIDs) == 0 {
		return nil, fmt.Errorf("GetPlaceCards: at least one placeId is required")
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "places/card", apiQuery(map[string]string{
		"placeIds": strings.Join(placeIDs, ","),
	}), nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesCardResponse
	if err := decodeAPIBody("GetPlaceCards", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListPlacePageGeos returns geos that have a Wanderlog "place page" available.
func (c *Client) ListPlacePageGeos() (*PlacesAPIEnvelope, error) {
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "places/placePageGeos", nil, nil, false)
	if err != nil {
		return nil, err
	}
	var result PlacesAPIEnvelope
	if err := decodeAPIBody("ListPlacePageGeos", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
