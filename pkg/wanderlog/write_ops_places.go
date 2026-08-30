package wanderlog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// RemovePlace removes a place from a trip section
func (c *Client) RemovePlace(tripKey string, sectionID, placeID int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for removing places")
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeID":   placeID,
	}).Debug("Removing place from trip")

	body := map[string]any{"placeIds": []int{placeID}}
	var statusCode int
	var respBody []byte
	if sectionID > 0 {
		resp, err := c.apiJSON(context.Background(), http.MethodDelete, fmt.Sprintf("tripPlans/%s/sections/%d/places", url.PathEscape(tripKey), sectionID), nil, body, true)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode
		respBody = resp.Body
	} else {
		resp, err := c.apiJSON(context.Background(), http.MethodDelete, "tripPlans/"+url.PathEscape(tripKey)+"/sections/places", nil, body, true)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		statusCode = resp.StatusCode
		respBody = resp.Body
	}

	if err := decodeOptionalMutationBody("RemovePlace", statusCode, respBody); err != nil {
		return err
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
	c.logger.WithFields(map[string]interface{}{
		"tripKey":    tripKey,
		"operations": len(ops),
	}).Debug("Applying operations to trip")

	resp, err := c.apiRequest(context.Background(), http.MethodPost, "tripPlans/"+url.PathEscape(tripKey)+"/applyOps", nil, reqBody, true)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":      tripKey,
		"operations":   len(ops),
		"statusCode":   resp.StatusCode,
		"responseBody": string(resp.Body),
	}).Debug("ApplyOperations API response")

	if err := decodeMutationBody("ApplyOperations", resp.StatusCode, resp.Body, nil); err != nil {
		return err
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

	trip, err := c.GetTripRaw(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}
	sectionIdx, section, err := findRawItinerarySection(trip, sectionID)
	if err != nil {
		return err
	}
	oldBlocks, err := rawBlocks(section)
	if err != nil {
		return fmt.Errorf("section %d: %w", sectionID, err)
	}
	if len(oldBlocks) == 0 {
		return nil
	}

	clearOp := ReplaceInObject([]any{"itinerary", "sections", sectionIdx, "blocks"}, oldBlocks, []any{})

	err = c.ApplyOperations(tripKey, []Operation{clearOp})
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

	trip, err := c.GetTripRaw(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}
	sectionIdx, oldSection, err := findRawItinerarySection(trip, sectionID)
	if err != nil {
		return err
	}

	// Create an operation to remove the section
	deleteOp := DeleteFromList(
		[]any{"itinerary", "sections"},
		sectionIdx,
		oldSection,
	)

	err = c.ApplyOperations(tripKey, []Operation{deleteOp})
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

	trip, err := c.GetTripRaw(tripKey)
	if err != nil {
		return fmt.Errorf("failed to fetch trip: %w", err)
	}

	sections, err := rawItinerarySections(trip)
	if err != nil {
		return err
	}
	operations := make([]Operation, 0, len(sections)+1)
	removedPlaces := 0
	for sectionIdx, sectionAny := range sections {
		section, ok := sectionAny.(map[string]any)
		if !ok {
			return fmt.Errorf("itinerary section %d has unexpected type %T", sectionIdx, sectionAny)
		}
		oldBlocks, err := rawBlocks(section)
		if err != nil {
			return fmt.Errorf("itinerary section %d: %w", sectionIdx, err)
		}
		newBlocks := make([]any, 0, len(oldBlocks))
		for _, block := range oldBlocks {
			if rawBlockIsPlace(block) {
				removedPlaces++
				continue
			}
			newBlocks = append(newBlocks, block)
		}
		if len(newBlocks) != len(oldBlocks) {
			operations = append(operations, ReplaceInObject(
				[]any{"itinerary", "sections", sectionIdx, "blocks"}, oldBlocks, newBlocks,
			))
		}
	}

	if resources, ok := trip["resources"].(map[string]any); ok {
		if metadata, exists := resources["placeMetadata"]; exists && !rawContainerEmpty(metadata) {
			emptyMetadata, err := emptyRawContainer(metadata)
			if err != nil {
				return fmt.Errorf("resources.placeMetadata: %w", err)
			}
			operations = append(operations, ReplaceInObject(
				[]any{"resources", "placeMetadata"}, metadata, emptyMetadata,
			))
		}
	}
	if len(operations) == 0 {
		c.logger.WithField("tripKey", tripKey).Info("No place blocks found in trip, nothing to clear")
		return nil
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":       tripKey,
		"placesRemoved": removedPlaces,
	}).Debug("Removing place blocks from trip")

	err = c.ApplyOperations(tripKey, operations)
	if err != nil {
		return fmt.Errorf("failed to nuke trip places: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey": tripKey,
		"places":  removedPlaces,
	}).Info("Successfully nuked all place data from trip")

	return nil
}

// MovePlace moves a place from one section to another at a specific position
func (c *Client) MovePlace(tripKey string, placeID, fromSectionID, toSectionID, position int) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for moving places")
	}

	trip, err := c.GetTripRaw(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	ops, err := movePlaceRawOperations(trip, placeID, fromSectionID, toSectionID, position)
	if err != nil {
		return err
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

	trip, err := c.GetTripRaw(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	ops, err := reorderPlacesRawOperations(trip, sectionID, placeIDs)
	if err != nil {
		return err
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

// UpdatePlaceVisitTime sets the displayed visit time for a place block.
func (c *Client) UpdatePlaceVisitTime(tripKey string, sectionID, placeID int, startTime, endTime string) error {
	if c.auth == nil {
		return fmt.Errorf("authentication required for updating place visit time")
	}
	if startTime == "" && endTime == "" {
		return fmt.Errorf("start time or end time is required")
	}
	if err := ValidateVisitTime(startTime); err != nil {
		return fmt.Errorf("start_time: %w", err)
	}
	if err := ValidateVisitTime(endTime); err != nil {
		return fmt.Errorf("end_time: %w", err)
	}

	trip, err := c.GetTripRaw(tripKey)
	if err != nil {
		return fmt.Errorf("getting current trip: %w", err)
	}

	ops, err := updatePlaceVisitTimeRawOperations(trip, sectionID, placeID, startTime, endTime)
	if err != nil {
		return err
	}
	if err := c.ApplyOperations(tripKey, ops); err != nil {
		return fmt.Errorf("applying visit time operations: %w", err)
	}

	c.logger.WithFields(map[string]interface{}{
		"tripKey":   tripKey,
		"sectionID": sectionID,
		"placeID":   placeID,
		"startTime": startTime,
		"endTime":   endTime,
	}).Info("Successfully updated place visit time")

	return nil
}

func rawInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	default:
		return 0
	}
}

func cloneRawMap(value map[string]any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var cloned map[string]any
	if err := decodeJSONPreserveNumbers(data, &cloned); err != nil {
		return nil, err
	}
	return cloned, nil
}

func rawItinerarySections(trip map[string]any) ([]any, error) {
	tripPlan, ok := trip["tripPlan"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("trip response is missing tripPlan")
	}
	itinerary, ok := tripPlan["itinerary"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("trip response is missing tripPlan.itinerary")
	}
	sectionsValue, exists := itinerary["sections"]
	if !exists || sectionsValue == nil {
		return []any{}, nil
	}
	sections, ok := sectionsValue.([]any)
	if !ok {
		return nil, fmt.Errorf("tripPlan.itinerary.sections has unexpected type %T", sectionsValue)
	}
	return sections, nil
}

func findRawItinerarySection(trip map[string]any, sectionID int) (int, map[string]any, error) {
	sections, err := rawItinerarySections(trip)
	if err != nil {
		return 0, nil, err
	}
	for sectionIdx, sectionAny := range sections {
		section, ok := sectionAny.(map[string]any)
		if !ok {
			continue
		}
		if rawInt(section["id"]) == sectionID {
			return sectionIdx, section, nil
		}
	}
	return 0, nil, fmt.Errorf("section %d not found", sectionID)
}

func rawBlocks(section map[string]any) ([]any, error) {
	blocksValue, exists := section["blocks"]
	if !exists || blocksValue == nil {
		return []any{}, nil
	}
	blocks, ok := blocksValue.([]any)
	if !ok {
		return nil, fmt.Errorf("blocks has unexpected type %T", blocksValue)
	}
	return blocks, nil
}

func rawBlockIsPlace(blockAny any) bool {
	block, ok := blockAny.(map[string]any)
	if !ok {
		return false
	}
	if blockType, _ := block["type"].(string); blockType != "" {
		return blockType == "place"
	}
	_, hasPlace := block["place"]
	return hasPlace
}

func rawContainerEmpty(value any) bool {
	switch container := value.(type) {
	case nil:
		return true
	case []any:
		return len(container) == 0
	case map[string]any:
		return len(container) == 0
	default:
		return false
	}
}

func emptyRawContainer(value any) (any, error) {
	switch value.(type) {
	case []any:
		return []any{}, nil
	case map[string]any:
		return map[string]any{}, nil
	default:
		return nil, fmt.Errorf("expected array or object, got %T", value)
	}
}

func findRawItineraryBlock(trip map[string]any, sectionID, blockID int) (int, int, map[string]any, error) {
	sections, err := rawItinerarySections(trip)
	if err != nil {
		return 0, 0, nil, err
	}
	for sectionIdx, sectionAny := range sections {
		section, _ := sectionAny.(map[string]any)
		if section == nil {
			continue
		}
		if sectionID > 0 && rawInt(section["id"]) != sectionID {
			continue
		}
		blocks, _ := section["blocks"].([]any)
		for blockIdx, blockAny := range blocks {
			block, _ := blockAny.(map[string]any)
			if block == nil || rawInt(block["id"]) != blockID {
				continue
			}
			return sectionIdx, blockIdx, block, nil
		}
	}
	if sectionID > 0 {
		return 0, 0, nil, fmt.Errorf("place %d not found in section %d", blockID, sectionID)
	}
	return 0, 0, nil, fmt.Errorf("place %d not found", blockID)
}

func movePlaceRawOperations(trip map[string]any, placeID, fromSectionID, toSectionID, position int) ([]Operation, error) {
	if position < 0 {
		return nil, fmt.Errorf("position must be >= 0")
	}
	fromIdx, fromSection, err := findRawItinerarySection(trip, fromSectionID)
	if err != nil {
		return nil, fmt.Errorf("source %w", err)
	}
	toIdx, toSection, err := findRawItinerarySection(trip, toSectionID)
	if err != nil {
		return nil, fmt.Errorf("destination %w", err)
	}
	fromBlocks, err := rawBlocks(fromSection)
	if err != nil {
		return nil, fmt.Errorf("source section %d: %w", fromSectionID, err)
	}
	toBlocks, err := rawBlocks(toSection)
	if err != nil {
		return nil, fmt.Errorf("destination section %d: %w", toSectionID, err)
	}

	blockIdx := -1
	var blockData any
	for i, block := range fromBlocks {
		blockMap, ok := block.(map[string]any)
		if ok && rawInt(blockMap["id"]) == placeID && rawBlockIsPlace(block) {
			blockIdx = i
			blockData = block
			break
		}
	}
	if blockIdx < 0 {
		return nil, fmt.Errorf("place %d not found in section %d", placeID, fromSectionID)
	}

	// position is the desired zero-based index in the final destination list.
	// A same-section deletion happens first, so its final list has one fewer item.
	maxPosition := len(toBlocks)
	if fromIdx == toIdx {
		maxPosition--
	}
	if position > maxPosition {
		position = maxPosition
	}

	return []Operation{
		DeleteFromList([]any{"itinerary", "sections", fromIdx, "blocks"}, blockIdx, blockData),
		InsertInList([]any{"itinerary", "sections", toIdx, "blocks"}, position, blockData),
	}, nil
}

func reorderPlacesRawOperations(trip map[string]any, sectionID int, placeIDs []int) ([]Operation, error) {
	if len(placeIDs) == 0 {
		return nil, fmt.Errorf("at least one place ID is required")
	}
	sectionIdx, section, err := findRawItinerarySection(trip, sectionID)
	if err != nil {
		return nil, err
	}
	oldBlocks, err := rawBlocks(section)
	if err != nil {
		return nil, fmt.Errorf("section %d: %w", sectionID, err)
	}

	requested := make(map[int]struct{}, len(placeIDs))
	for _, id := range placeIDs {
		if _, exists := requested[id]; exists {
			return nil, fmt.Errorf("duplicate place %d in requested order", id)
		}
		requested[id] = struct{}{}
	}
	byID := make(map[int]any, len(placeIDs))
	for _, block := range oldBlocks {
		blockMap, ok := block.(map[string]any)
		if !ok || !rawBlockIsPlace(block) {
			continue
		}
		id := rawInt(blockMap["id"])
		if _, wanted := requested[id]; wanted {
			byID[id] = block
		}
	}
	orderedBlocks := make([]any, 0, len(placeIDs))
	for _, id := range placeIDs {
		block, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("place %d not found in section %d", id, sectionID)
		}
		orderedBlocks = append(orderedBlocks, block)
	}

	newBlocks := make([]any, 0, len(oldBlocks))
	orderedIdx := 0
	for _, block := range oldBlocks {
		blockMap, ok := block.(map[string]any)
		if ok {
			if _, replace := requested[rawInt(blockMap["id"])]; replace && rawBlockIsPlace(block) {
				newBlocks = append(newBlocks, orderedBlocks[orderedIdx])
				orderedIdx++
				continue
			}
		}
		newBlocks = append(newBlocks, block)
	}
	return []Operation{ReplaceInObject(
		[]any{"itinerary", "sections", sectionIdx, "blocks"}, oldBlocks, newBlocks,
	)}, nil
}

func updatePlaceVisitTimeRawOperations(trip map[string]any, sectionID, placeID int, startTime, endTime string) ([]Operation, error) {
	sectionIdx, blockIdx, oldBlock, err := findRawItineraryBlock(trip, sectionID, placeID)
	if err != nil {
		return nil, err
	}
	if blockType, ok := oldBlock["type"].(string); ok && blockType != "" && blockType != "place" {
		return nil, fmt.Errorf("block %d has type %q, expected %q", placeID, blockType, "place")
	}
	if _, ok := oldBlock["place"]; !ok {
		return nil, fmt.Errorf("block %d is not a place block", placeID)
	}

	newBlock, err := cloneRawMap(oldBlock)
	if err != nil {
		return nil, fmt.Errorf("copying place block %d: %w", placeID, err)
	}
	if startTime != "" {
		newBlock["startTime"] = startTime
	}
	if endTime != "" {
		newBlock["endTime"] = endTime
	}

	return []Operation{
		ReplaceInList(
			[]any{"itinerary", "sections", sectionIdx, "blocks"},
			blockIdx,
			oldBlock,
			newBlock,
		),
	}, nil
}

func updatePlaceVisitTimeOperations(sections []ItSections, sectionID, placeID int, startTime, endTime string) ([]Operation, error) {
	sectionIdx := FindSectionIndex(sections, sectionID)
	if sectionIdx < 0 {
		return nil, fmt.Errorf("section %d not found", sectionID)
	}

	blockIdx := -1
	var oldBlock any
	var newBlock any
	for i, block := range sections[sectionIdx].Blocks {
		if block.ID == placeID {
			blockIdx = i
			oldBlock = block
			block.StartTime = startTime
			block.EndTime = endTime
			newBlock = block
			break
		}
	}
	if blockIdx < 0 {
		return nil, fmt.Errorf("place %d not found in section %d", placeID, sectionID)
	}

	return []Operation{
		ReplaceInList(
			[]any{"itinerary", "sections", sectionIdx, "blocks"},
			blockIdx,
			oldBlock,
			newBlock,
		),
	}, nil
}
