package wanderlog

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

type (
	AssistantTextRequest         = models.AssistantTextRequest
	AssistantStreamEvent         = models.AssistantStreamEvent
	AssistantTextResponse        = models.AssistantTextResponse
	AssistantHighlightsRequest   = models.AssistantHighlightsRequest
	AssistantHighlightsResponse  = models.AssistantHighlightsResponse
	AssistantHistoryResponse     = models.AssistantHistoryResponse
	AssistantChatsResponse       = models.AssistantChatsResponse
	AssistantInitialChatResponse = models.AssistantInitialChatResponse
)

// GetTripPlanAssistantText sends a user message to the trip-plan assistant
// and consumes the resulting NDJSON stream synchronously, returning the final
// accumulated response. Chat metadata, message metadata, and concatenated
// content fragments are exposed individually.
func (c *Client) GetTripPlanAssistantText(req AssistantTextRequest) (*AssistantTextResponse, error) {
	if req.Message == "" {
		return nil, fmt.Errorf("GetTripPlanAssistantText: message is required")
	}
	if err := c.requireAuth("GetTripPlanAssistantText"); err != nil {
		return nil, err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("GetTripPlanAssistantText: marshaling request: %w", err)
	}

	apiURL, err := buildAPIURL("chat/tripPlanAssistant/getText/v2", nil)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("GetTripPlanAssistantText: creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)
	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, fmt.Errorf("GetTripPlanAssistantText: adding auth headers: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("GetTripPlanAssistantText: making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GetTripPlanAssistantText: HTTP %d: %s", resp.StatusCode, truncateForLog(string(raw), 500))
	}

	var result AssistantTextResponse
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var event AssistantStreamEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, fmt.Errorf("GetTripPlanAssistantText: parsing stream chunk %q: %w", line, err)
		}
		switch event.Type {
		case "chatMetadata":
			result.ChatMetadata = append([]byte(nil), event.Data...)
		case "messageMetadata":
			result.MessageMetadata = append([]byte(nil), event.Data...)
		case "content":
			if len(event.Data) > 0 {
				var chunk string
				if err := json.Unmarshal(event.Data, &chunk); err == nil {
					result.Content += chunk
				}
			}
		}
		result.Events = append(result.Events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("GetTripPlanAssistantText: reading stream: %w", err)
	}
	return &result, nil
}

// GetTripPlanAssistantHighlights extracts place highlights from a previous
// assistant message.
func (c *Client) GetTripPlanAssistantHighlights(req AssistantHighlightsRequest) (*AssistantHighlightsResponse, error) {
	if req.AssistantMessage == "" {
		return nil, fmt.Errorf("GetTripPlanAssistantHighlights: assistantMessage is required")
	}
	if req.TripPlanID == 0 {
		return nil, fmt.Errorf("GetTripPlanAssistantHighlights: tripPlanId is required")
	}
	if err := c.requireAuth("GetTripPlanAssistantHighlights"); err != nil {
		return nil, err
	}
	resp, err := c.apiJSON(context.Background(), http.MethodPost, "chat/tripPlanAssistant/getHighlights/v2", nil, req, true)
	if err != nil {
		return nil, err
	}
	var result AssistantHighlightsResponse
	if err := decodeAPIBody("GetTripPlanAssistantHighlights", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripPlanAssistantHistory returns prior assistant messages for the given
// chat. Pass arbitrary query params (e.g. chatId, pageSize, sentAtBefore) via
// the params map.
func (c *Client) GetTripPlanAssistantHistory(params map[string]string) (*AssistantHistoryResponse, error) {
	if err := c.requireAuth("GetTripPlanAssistantHistory"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "chat/tripPlanAssistant/history", apiQuery(params), nil, true)
	if err != nil {
		return nil, err
	}
	var result AssistantHistoryResponse
	if err := decodeAPIBody("GetTripPlanAssistantHistory", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListTripPlanAssistantChats lists assistant chat threads for a trip plan.
// search/lastItemIsBefore/pageSize are optional.
func (c *Client) ListTripPlanAssistantChats(tripPlanID int, search string, lastItemIsBeforeMillis int64, pageSize int) (*AssistantChatsResponse, error) {
	if tripPlanID == 0 {
		return nil, fmt.Errorf("ListTripPlanAssistantChats: tripPlanId is required")
	}
	if err := c.requireAuth("ListTripPlanAssistantChats"); err != nil {
		return nil, err
	}
	params := map[string]string{"tripPlanId": strconv.Itoa(tripPlanID)}
	if search != "" {
		params["search"] = search
	}
	if lastItemIsBeforeMillis > 0 {
		params["lastItemIsBefore"] = strconv.FormatInt(lastItemIsBeforeMillis, 10)
	}
	if pageSize > 0 {
		params["pageSize"] = strconv.Itoa(pageSize)
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "chat/tripPlanAssistant/chats", apiQuery(params), nil, true)
	if err != nil {
		return nil, err
	}
	var result AssistantChatsResponse
	if err := decodeAPIBody("ListTripPlanAssistantChats", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTripPlanAssistantInitialChat returns the seeded initial chat (with first
// few items) for a trip plan.
func (c *Client) GetTripPlanAssistantInitialChat(tripPlanID int) (*AssistantInitialChatResponse, error) {
	if tripPlanID == 0 {
		return nil, fmt.Errorf("GetTripPlanAssistantInitialChat: tripPlanId is required")
	}
	if err := c.requireAuth("GetTripPlanAssistantInitialChat"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(context.Background(), http.MethodGet, "chat/tripPlanAssistant/initialChatWithItems", apiQuery(map[string]string{
		"tripPlanId": strconv.Itoa(tripPlanID),
	}), nil, true)
	if err != nil {
		return nil, err
	}
	var result AssistantInitialChatResponse
	if err := decodeAPIBody("GetTripPlanAssistantInitialChat", resp.StatusCode, resp.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
