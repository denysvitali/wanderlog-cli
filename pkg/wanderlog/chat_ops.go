package wanderlog

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/denysvitali/wanderlog-cli/pkg/wanderlog/models"
)

const (
	// MaxAssistantStreamBytes bounds the complete NDJSON response. Assistant
	// streams use the same response budget as non-streaming API operations.
	MaxAssistantStreamBytes = MaxAPIResponseBodyBytes
	// MaxAssistantStreamEventBytes bounds one NDJSON record, preventing a single
	// event from forcing an oversized scanner/reader allocation.
	MaxAssistantStreamEventBytes = 1 << 20
	// MaxAssistantContentBytes bounds the concatenated decoded content.
	MaxAssistantContentBytes = MaxAPIResponseBodyBytes
	// MaxAssistantStreamEvents bounds event slice overhead even for tiny events.
	MaxAssistantStreamEvents = 10_000
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
	return c.GetTripPlanAssistantTextContext(context.Background(), req)
}

// GetTripPlanAssistantTextContext sends a user message to the trip-plan
// assistant and binds the request and streamed response to ctx.
func (c *Client) GetTripPlanAssistantTextContext(ctx context.Context, req AssistantTextRequest) (*AssistantTextResponse, error) {
	const operation = "GetTripPlanAssistantText"
	if req.Message == "" {
		return nil, newAPIError(operation, 0, "message is required", nil, nil)
	}
	if ctx == nil {
		return nil, newAPIError(operation, 0, "creating request: nil context", nil, nil)
	}
	if err := c.requireAuth(operation); err != nil {
		return nil, err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, newAPIError(operation, 0, "marshaling request: "+err.Error(), nil, err)
	}

	apiURL, err := c.buildAPIURL("chat/tripPlanAssistant/getText/v2", nil)
	if err != nil {
		return nil, newAPIError(operation, 0, "building request URL: "+err.Error(), nil, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, newAPIError(operation, 0, "creating request: "+err.Error(), nil, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)
	if err := c.ensureAuthOrigin(httpReq); err != nil {
		return nil, newAPIError(operation, 0, err.Error(), nil, err)
	}
	if err := c.addAuthHeaders(httpReq); err != nil {
		return nil, newAPIError(operation, 0, "adding auth headers: "+err.Error(), nil, err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, newAPIError(operation, 0, "making request: "+err.Error(), nil, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, readErr := readAPIResponseBody(resp.Body)
		if readErr != nil {
			return nil, newAPIError(operation, resp.StatusCode, "reading response body: "+readErr.Error(), raw, readErr)
		}
		return nil, apiHTTPError(operation, resp.StatusCode, raw)
	}

	var result AssistantTextResponse
	var content strings.Builder
	limited := &io.LimitedReader{R: resp.Body, N: MaxAssistantStreamBytes + 1}
	reader := bufio.NewReaderSize(limited, MaxAssistantStreamEventBytes+2)
	storedEventBytes := 0
	metadataBytes := 0
	for {
		rawLine, readErr := reader.ReadSlice('\n')
		if errors.Is(readErr, bufio.ErrBufferFull) {
			return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("assistant stream event exceeds %d bytes", MaxAssistantStreamEventBytes), rawLine, readErr)
		}

		line := bytes.TrimSpace(rawLine)
		if len(line) > MaxAssistantStreamEventBytes {
			return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("assistant stream event exceeds %d bytes", MaxAssistantStreamEventBytes), line, nil)
		}
		if len(line) > 0 {
			if len(result.Events) >= MaxAssistantStreamEvents {
				return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("assistant stream exceeds %d events", MaxAssistantStreamEvents), line, nil)
			}
			storedEventBytes += len(line)
			if storedEventBytes > MaxAssistantStreamBytes {
				return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("stored assistant events exceed %d bytes", MaxAssistantStreamBytes), line, nil)
			}

			var event AssistantStreamEvent
			if err := json.Unmarshal(line, &event); err != nil {
				return nil, newAPIError(operation, resp.StatusCode, "parsing stream event: "+err.Error(), line, err)
			}
			if event.Success != nil && !*event.Success {
				return nil, newAPIError(operation, resp.StatusCode, assistantStreamFailureMessage(event), line, nil)
			}

			switch event.Type {
			case "chatMetadata":
				metadataBytes += len(event.Data)
				if metadataBytes > MaxAssistantStreamBytes {
					return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("assistant metadata exceeds %d bytes", MaxAssistantStreamBytes), line, nil)
				}
				result.ChatMetadata = append([]byte(nil), event.Data...)
			case "messageMetadata":
				metadataBytes += len(event.Data)
				if metadataBytes > MaxAssistantStreamBytes {
					return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("assistant metadata exceeds %d bytes", MaxAssistantStreamBytes), line, nil)
				}
				result.MessageMetadata = append([]byte(nil), event.Data...)
			case "content":
				var chunk string
				data := bytes.TrimSpace(event.Data)
				if len(data) == 0 || bytes.Equal(data, []byte("null")) {
					return nil, newAPIError(operation, resp.StatusCode, "content event data must be a JSON string", line, nil)
				}
				if err := json.Unmarshal(data, &chunk); err != nil {
					return nil, newAPIError(operation, resp.StatusCode, "content event data must be a JSON string: "+err.Error(), line, err)
				}
				if content.Len()+len(chunk) > MaxAssistantContentBytes {
					return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("assistant content exceeds %d bytes", MaxAssistantContentBytes), line, nil)
				}
				_, _ = content.WriteString(chunk)
			}
			result.Events = append(result.Events, event)
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return nil, newAPIError(operation, resp.StatusCode, "reading assistant stream: "+readErr.Error(), rawLine, readErr)
		}
	}
	if limited.N == 0 {
		return nil, newAPIError(operation, resp.StatusCode, fmt.Sprintf("assistant stream exceeds %d bytes", MaxAssistantStreamBytes), nil, nil)
	}
	result.Content = content.String()
	return &result, nil
}

func assistantStreamFailureMessage(event AssistantStreamEvent) string {
	if event.Message != "" {
		return event.Message
	}
	for _, raw := range []json.RawMessage{event.Error, event.Data} {
		if len(raw) == 0 {
			continue
		}
		if message := apiErrorMessage(raw); message != "" {
			return message
		}
		var message string
		if err := json.Unmarshal(raw, &message); err == nil && message != "" {
			return message
		}
	}
	return "assistant stream reported success=false"
}

// GetTripPlanAssistantHighlights extracts place highlights from a previous
// assistant message.
func (c *Client) GetTripPlanAssistantHighlights(req AssistantHighlightsRequest) (*AssistantHighlightsResponse, error) {
	return c.GetTripPlanAssistantHighlightsContext(context.Background(), req)
}

func (c *Client) GetTripPlanAssistantHighlightsContext(ctx context.Context, req AssistantHighlightsRequest) (*AssistantHighlightsResponse, error) {
	if req.AssistantMessage == "" {
		return nil, fmt.Errorf("GetTripPlanAssistantHighlights: assistantMessage is required")
	}
	if req.TripPlanID == 0 {
		return nil, fmt.Errorf("GetTripPlanAssistantHighlights: tripPlanId is required")
	}
	if err := c.requireAuth("GetTripPlanAssistantHighlights"); err != nil {
		return nil, err
	}
	resp, err := c.apiJSON(ctx, http.MethodPost, "chat/tripPlanAssistant/getHighlights/v2", nil, req, true)
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
	return c.GetTripPlanAssistantHistoryContext(context.Background(), params)
}

func (c *Client) GetTripPlanAssistantHistoryContext(ctx context.Context, params map[string]string) (*AssistantHistoryResponse, error) {
	if err := c.requireAuth("GetTripPlanAssistantHistory"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(ctx, http.MethodGet, "chat/tripPlanAssistant/history", apiQuery(params), nil, true)
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
	return c.ListTripPlanAssistantChatsContext(context.Background(), tripPlanID, search, lastItemIsBeforeMillis, pageSize)
}

func (c *Client) ListTripPlanAssistantChatsContext(ctx context.Context, tripPlanID int, search string, lastItemIsBeforeMillis int64, pageSize int) (*AssistantChatsResponse, error) {
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
	resp, err := c.apiRequest(ctx, http.MethodGet, "chat/tripPlanAssistant/chats", apiQuery(params), nil, true)
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
	return c.GetTripPlanAssistantInitialChatContext(context.Background(), tripPlanID)
}

func (c *Client) GetTripPlanAssistantInitialChatContext(ctx context.Context, tripPlanID int) (*AssistantInitialChatResponse, error) {
	if tripPlanID == 0 {
		return nil, fmt.Errorf("GetTripPlanAssistantInitialChat: tripPlanId is required")
	}
	if err := c.requireAuth("GetTripPlanAssistantInitialChat"); err != nil {
		return nil, err
	}
	resp, err := c.apiRequest(ctx, http.MethodGet, "chat/tripPlanAssistant/initialChatWithItems", apiQuery(map[string]string{
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
