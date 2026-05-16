package models

import "encoding/json"

// AssistantTextRequest is the body for POST /api/chat/tripPlanAssistant/getText/v2.
// The endpoint streams NDJSON, so the wrapper accumulates chunks.
type AssistantTextRequest struct {
	ChatID      string `json:"chatId,omitempty"`
	Message     string `json:"message"`
	TripPlanID  int    `json:"tripPlanId,omitempty"`
	TripPlanKey string `json:"tripPlanKey,omitempty"`
	GeoID       int    `json:"geoId,omitempty"`
}

// AssistantStreamEvent is one NDJSON record streamed from the assistant.
// The Type field is one of "chatMetadata", "messageMetadata", "content".
type AssistantStreamEvent struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data,omitempty"`
	Success *bool           `json:"success,omitempty"`
}

// AssistantTextResponse is the accumulated synchronous response returned by
// GetTripPlanAssistantText (collects all stream events and concatenates the
// content fragments).
type AssistantTextResponse struct {
	ChatMetadata    json.RawMessage        `json:"chatMetadata,omitempty"`
	MessageMetadata json.RawMessage        `json:"messageMetadata,omitempty"`
	Content         string                 `json:"content,omitempty"`
	Events          []AssistantStreamEvent `json:"events,omitempty"`
}

// AssistantHighlightsRequest is the body for POST /api/chat/tripPlanAssistant/getHighlights/v2.
type AssistantHighlightsRequest struct {
	AssistantMessage    string `json:"assistantMessage"`
	TripPlanID          int    `json:"tripPlanId"`
	SelectedGeoID       int    `json:"selectedGeoId,omitempty"`
	ChatID              string `json:"chatId,omitempty"`
	AssistantChatItemID string `json:"assistantChatItemId,omitempty"`
}

// AssistantHighlightsResponse is the response wrapper for getHighlights/v2.
type AssistantHighlightsResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// AssistantHistoryResponse mirrors GET /api/chat/tripPlanAssistant/history.
type AssistantHistoryResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// AssistantChatsResponse mirrors GET /api/chat/tripPlanAssistant/chats.
type AssistantChatsResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// AssistantInitialChatResponse mirrors GET /api/chat/tripPlanAssistant/initialChatWithItems.
type AssistantInitialChatResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}
