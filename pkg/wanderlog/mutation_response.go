package wanderlog

import (
	"bytes"
	"encoding/json"
)

// decodeOptionalMutationBody is for endpoints whose successful response is
// legitimately empty. If a body is present it must still be a valid success
// envelope, so HTTP 200 {"success":false} can never be mistaken for success.
func decodeOptionalMutationBody(opName string, statusCode int, body []byte) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return decodeAPIBody(opName, statusCode, body, nil)
	}
	return decodeMutationBody(opName, statusCode, body, nil)
}

// decodeMutationBody validates the common Wanderlog mutation envelope before
// decoding endpoint-specific fields. A 2xx status alone is not sufficient:
// the API can report failures as {"success":false} with HTTP 200.
func decodeMutationBody(opName string, statusCode int, body []byte, out any) error {
	if statusCode < 200 || statusCode >= 300 {
		return apiHTTPError(opName, statusCode, body)
	}
	if len(body) == 0 {
		return newAPIError(opName, statusCode, "empty mutation response", body, nil)
	}

	var envelope struct {
		Success  *bool  `json:"success"`
		Error    any    `json:"error"`
		Message  string `json:"message"`
		Messages []any  `json:"messages"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return newAPIError(opName, statusCode, "decoding mutation response: "+err.Error(), body, err)
	}
	if envelope.Success == nil {
		return newAPIError(opName, statusCode, "mutation response is missing boolean success", body, nil)
	}
	if !*envelope.Success {
		message := envelope.Message
		if message == "" && len(envelope.Messages) > 0 {
			switch value := envelope.Messages[0].(type) {
			case string:
				message = value
			case map[string]any:
				message, _ = value["message"].(string)
			}
		}
		if message == "" {
			switch value := envelope.Error.(type) {
			case string:
				message = value
			case map[string]any:
				message, _ = value["message"].(string)
			}
		}
		if message == "" {
			message = "API returned success=false"
		}
		return newAPIError(opName, statusCode, message, body, nil)
	}
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return newAPIError(opName, statusCode, "decoding response: "+err.Error(), body, err)
		}
	}
	return nil
}
