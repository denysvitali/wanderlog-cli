package wanderlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/openapi"
)

// RemovePlace removes a place from a trip section
func (c *Client) RemovePlace(tripKey string, sectionID, placeID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for removing places")
	}

	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeID":   placeID,
	}).Debug("Removing place from trip")

	body := openapi.RemovePlacesRequest{PlaceIds: []int{placeID}}
	var statusCode int
	var respBody []byte
	if sectionID > 0 {
		resp, err := api.RemovePlacesFromTripPlanWithResponse(context.Background(), tripKey, sectionID, body)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode()
		respBody = resp.Body
	} else {
		resp, err := api.RemovePlacesFromTripPlanWithoutSectionWithResponse(context.Background(), tripKey, body)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode()
		respBody = resp.Body
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("RemovePlace: HTTP %d: %s", statusCode, truncateForLog(string(respBody), 500))
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"placeID": placeID,
	}).Info("Successfully removed place from trip")

	return nil
}

// ApplyOperations applies a batch of operations to a trip (for complex edits)
func (c *Client) ApplyOperations(tripKey string, ops []Operation) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for applying operations")
	}

	opReq := OperationRequest{Ops: ops}
	reqBody, err := json.Marshal(opReq)
	if err != nil {
		return fmt.Errorf("marshaling operations request: %w", err)
	}
	api, err := c.openAPI()
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":    tripKey,
		"operations": len(ops),
	}).Debug("Applying operations to trip")

	resp, err := api.ApplyOpsToTripPlanWithBodyWithResponse(context.Background(), tripKey, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":      tripKey,
		"operations":   len(ops),
		"statusCode":   resp.StatusCode(),
		"responseBody": string(resp.Body),
	}).Debug("ApplyOperations API response")

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("ApplyOperations: HTTP %d: %s", resp.StatusCode(), truncateForLog(string(resp.Body), 500))
	}

	// Try to parse the response to check for API-level errors
	var apiResp map[string]interface{}
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		c.logger.WithField("responseBody", string(resp.Body)).Warn("Could not parse API response as JSON")
	} else {
		// Check if the response indicates success
		if success, ok := apiResp["success"]; ok {
			if successBool, ok := success.(bool); ok && !successBool {
				// API returned success: false
				errorMsg := "unknown error"
				if msg, ok := apiResp["error"]; ok {
					if msgStr, ok := msg.(string); ok {
						errorMsg = msgStr
					}
				}
				// Also check for messages array
				if messages, ok := apiResp["messages"]; ok {
					if msgArray, ok := messages.([]interface{}); ok && len(msgArray) > 0 {
						if firstMsg, ok := msgArray[0].(string); ok {
							errorMsg = firstMsg
						}
					}
				}
				return fmt.Errorf("API request failed: %s", errorMsg)
			}
		}
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":    tripKey,
		"operations": len(ops),
	}).Info("Successfully applied operations to trip")

	return nil
}

// ClearSectionBlocks removes all blocks from a specific section using operational transforms
func (c *Client) ClearSectionBlocks(tripKey string, sectionID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for clearing section blocks")
	}

	// Create an operation to replace the blocks array with an empty array
	clearOp := ReplaceInObject(
		[]any{"itinerary", "sections", sectionID, "blocks"},
		[]any{}, // old value placeholder for ShareDB OD field
		[]any{},
	)

	err := c.ApplyOperations(tripKey, []Operation{clearOp})
	if err != nil {
		return fmt.Errorf("failed to clear section blocks: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
	}).Info("Successfully cleared all blocks from section")

	return nil
}

// DeleteSection removes an entire section from a trip using operational transforms
func (c *Client) DeleteSection(tripKey string, sectionID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for deleting sections")
	}

	// Create an operation to remove the section
	deleteOp := DeleteFromList(
		[]any{"itinerary", "sections"},
		sectionID,
		map[string]interface{}{}, // old value placeholder for ShareDB LD field
	)

	err := c.ApplyOperations(tripKey, []Operation{deleteOp})
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
	}).Info("Successfully deleted section")

	return nil
}

// NukeTripPlaces removes all place blocks from all sections in a trip using operational transforms
// This function first fetches the trip to determine which sections exist, then clears them
func (c *Client) NukeTripPlaces(tripKey string) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for nuking trip places")
	}

	// First fetch the trip to see what sections actually exist
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("failed to fetch trip: %w", err)
	}

	if len(trip.TripPlan.Itinerary.Sections) == 0 {
		c.logger.WithField("tripKey", tripKey).Info("No sections found in trip, nothing to clear")
		return nil
	}

	// Build operations only for sections that exist
	operations := []Operation{}
	for i := range trip.TripPlan.Itinerary.Sections {
		operations = append(operations, ReplaceInObject(
			[]any{"itinerary", "sections", i, "blocks"},
			[]any{}, // old value placeholder for ShareDB OD field
			[]any{},
		))
	}

	// Also clear place metadata
	operations = append(operations, ReplaceInObject(
		[]any{"resources", "placeMetadata"},
		[]any{}, // old value placeholder for ShareDB OD field
		[]any{},
	))

	c.logger.WithFields(map[string]interface{}{
		"tripKey":         tripKey,
		"sectionsCleared": len(trip.TripPlan.Itinerary.Sections),
	}).Debug("Clearing sections from trip")

	err = c.ApplyOperations(tripKey, operations)
	if err != nil {
		return fmt.Errorf("failed to nuke trip places: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":  tripKey,
		"sections": len(trip.TripPlan.Itinerary.Sections),
	}).Info("Successfully nuked all place data from trip")

	return nil
}

// MovePlace moves a place from one section to another at a specific position
func (c *Client) MovePlace(tripKey string, placeID, fromSectionID, toSectionID, position int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for moving places")
	}

	// First, get the current trip to find the place data
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	fromIdx := FindSectionIndex(trip.TripPlan.Itinerary.Sections, fromSectionID)
	if fromIdx < 0 {
		return fmt.Errorf("source section %d not found", fromSectionID)
	}

	toIdx := FindSectionIndex(trip.TripPlan.Itinerary.Sections, toSectionID)
	if toIdx < 0 {
		return fmt.Errorf("destination section %d not found", toSectionID)
	}

	// Find the block index of the place in the source section
	blockIdx := -1
	var blockData any
	for i, block := range trip.TripPlan.Itinerary.Sections[fromIdx].Blocks {
		if block.ID == placeID {
			blockIdx = i
			blockData = block
			break
		}
	}
	if blockIdx < 0 {
		return fmt.Errorf("place %d not found in section %d", placeID, fromSectionID)
	}

	// Build operations: delete from source, insert into destination
	ops := []Operation{
		DeleteFromList(
			[]any{"itinerary", "sections", fromIdx, "blocks"},
			blockIdx,
			blockData,
		),
		InsertInList(
			[]any{"itinerary", "sections", toIdx, "blocks"},
			position,
			blockData,
		),
	}

	if err := c.ApplyOperations(tripKey, ops); err != nil {
		return fmt.Errorf("applying move operations: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":       tripKey,
		"placeID":       placeID,
		"fromSectionID": fromSectionID,
		"toSectionID":   toSectionID,
		"position":      position,
	}).Info("Successfully moved place")

	return nil
}

// ReorderPlaces reorders places within a section by replacing the blocks list
func (c *Client) ReorderPlaces(tripKey string, sectionID int, placeIDs []int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for reordering places")
	}

	// First, get the current trip to find the section data
	trip, err := c.GetTrip(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	sectionIdx := FindSectionIndex(trip.TripPlan.Itinerary.Sections, sectionID)
	if sectionIdx < 0 {
		return fmt.Errorf("section %d not found", sectionID)
	}

	section := trip.TripPlan.Itinerary.Sections[sectionIdx]

	// Build a map of block ID -> block data
	blockMap := make(map[int]interface{})
	for _, block := range section.Blocks {
		blockMap[block.ID] = block
	}

	// Build the new blocks list in the requested order
	newBlocks := make([]any, 0, len(placeIDs))
	for _, id := range placeIDs {
		block, ok := blockMap[id]
		if !ok {
			return fmt.Errorf("place %d not found in section %d", id, sectionID)
		}
		newBlocks = append(newBlocks, block)
	}

	// Replace the entire blocks array
	ops := []Operation{
		ReplaceInObject(
			[]any{"itinerary", "sections", sectionIdx, "blocks"},
			section.Blocks,
			newBlocks,
		),
	}

	if err := c.ApplyOperations(tripKey, ops); err != nil {
		return fmt.Errorf("applying reorder operations: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeIDs":  placeIDs,
	}).Info("Successfully reordered places")

	return nil
}
