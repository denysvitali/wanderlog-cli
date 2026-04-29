package cmd

import (
	"testing"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLodgingBlockShapeMatchesReactNativeBundle(t *testing.T) {
	place := minimalPlaceForBlock("Hotel Test", "ChIJhotel", 48.8566, 2.3522)
	hotel := map[string]any{
		"checkIn":            "2026-06-01",
		"checkOut":           "2026-06-07",
		"travelerNames":      []any{},
		"confirmationNumber": "CONF123",
	}
	block := map[string]any{
		"type":       "place",
		"place":      place,
		"hotel":      hotel,
		"text":       quillTextForString("front desk note"),
		"imageSize":  "small",
		"travelMode": nil,
		"reactions":  []any{},
	}

	assert.Equal(t, "place", block["type"])
	assert.NotContains(t, block, "Hotel")
	assert.NotContains(t, block, "Place")
	assert.NotContains(t, block, "Geometry")

	gotPlace, ok := block["place"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Hotel Test", gotPlace["name"])
	assert.Equal(t, "ChIJhotel", gotPlace["place_id"])
	assert.Contains(t, gotPlace, "geometry")

	gotHotel, ok := block["hotel"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "2026-06-01", gotHotel["checkIn"])
	assert.Equal(t, "2026-06-07", gotHotel["checkOut"])
	assert.Contains(t, gotHotel, "travelerNames")
	assert.Equal(t, "CONF123", gotHotel["confirmationNumber"])

	text, ok := block["text"].(map[string]any)
	require.True(t, ok)
	ops, ok := text["ops"].([]any)
	require.True(t, ok)
	require.Len(t, ops, 1)
	op, ok := ops[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "front desk note\n", op["insert"])
}

func TestCreateLodgingSectionShapeMatchesReactNativeBundle(t *testing.T) {
	section := newLodgingSection(123)

	assert.Equal(t, 123, section["id"])
	assert.Equal(t, "Hotels and lodging", section["heading"])
	assert.Equal(t, "hotels", section["type"])
	assert.Equal(t, "bed", section["placeMarkerIcon"])
	assert.NotEqual(t, "lodging", section["type"])
	assert.NotContains(t, section, "displayHeading")
	assert.NotContains(t, section, "date")

	text, ok := section["text"].(map[string]any)
	require.True(t, ok)
	ops, ok := text["ops"].([]any)
	require.True(t, ok)
	require.Len(t, ops, 1)
}

func TestSectionIndexToAddSectionTypeMatchesReactNativeBundleOrder(t *testing.T) {
	sections := []wanderlog.ItSections{
		{Type: "attachments"},
		{Type: "normal"},
		{Type: "normal"},
	}

	assert.Equal(t, 1, sectionIndexToAddSectionType("flights", sections))
	assert.Equal(t, 1, sectionIndexToAddSectionType("hotels", sections))

	sections = []wanderlog.ItSections{
		{Type: "attachments"},
		{Type: "flights"},
		{Type: "normal"},
	}
	assert.Equal(t, 2, sectionIndexToAddSectionType("hotels", sections))

	sections = []wanderlog.ItSections{
		{Type: "attachments"},
		{Type: "flights"},
		{Type: "hotels"},
		{Type: "normal"},
	}
	assert.Equal(t, 3, sectionIndexToAddSectionType("rentalCars", sections))
}
