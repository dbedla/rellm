package rellm

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenRouterConversationConverter struct{}

func (o *OpenRouterConversationConverter) ToConversationElements(items []json.RawMessage) ([]ConversationElement, error) {
	elements := make([]ConversationElement, 0, len(items))
	for _, raw := range items {
		var msg struct {
			Type    string          `json:"type"`
			Role    string          `json:"role"`
			Id      string          `json:"id"`
			Name    string          `json:"name"`
			CallID  string          `json:"call_id"`
			Args    json.RawMessage `json:"arguments"`
			Content json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil, err
		}

		switch msg.Type {
		case "function_call":
			elements = append(elements, &FunctionCall{
				Id:     msg.Id,
				Name:   msg.Name,
				Args:   msg.Args,
				CallId: msg.CallID,
			})

		case "function_call_output":
			// Extract output field directly from raw JSON.
			var out struct {
				Output string `json:"output"`
			}
			if err := json.Unmarshal(raw, &out); err == nil {
				elements = append(elements, &FunctionCallResp{
					Id:     msg.Id,
					CallId: msg.CallID,
					Output: out.Output,
				})
			} else {
				elements = append(elements, &FunctionCallResp{
					Id:     msg.Id,
					CallId: msg.CallID,
					Output: string(msg.Content),
				})
			}

		case "message", "": // messages often lack an explicit type field; infer from role
			elements = append(elements, o.parseMessage(raw, msg.Id, msg.Role, msg.Content))

		case "reasoning":
			elements = append(elements, o.parseReasoning(raw))

		case "image_generation_call":
			elements = append(elements, parseImageGeneration(raw))

		default:
			// Unknown types pass through unchanged.
		}
	}
	return elements, nil
}

// parseMessage parses a message item into a role-typed message. Content is
// normalized to []MessagePart; ToProviderRepresentation re-serializes per the type's shape rule.
func (o *OpenRouterConversationConverter) parseMessage(raw json.RawMessage, id, role string, content json.RawMessage) ConversationElement {
	var statusInfo struct {
		Status string `json:"status"`
	}
	status := ""
	if err := json.Unmarshal(raw, &statusInfo); err == nil {
		status = statusInfo.Status
	}
	parts := parseMessageContent(content)
	switch role {
	case "assistant":
		return &AssistantMessage{messageContent{Id: id, Role: role, Status: status, Content: parts}}
	case "system":
		return &SystemMessage{messageContent{Id: id, Role: role, Content: parts}}
	default: // "user" or unknown
		return &UserMessage{messageContent{Id: id, Role: role, Status: status, Content: parts}}
	}
}

// marshalTextMessage serializes an assistant/system message for OpenRouter:
// text-only content becomes a string, multimodal stays an array.
func marshalTextMessage(mc messageContent) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"role": mc.Role,
		"type": "message",
	}
	if mc.Id != "" {
		payload["id"] = mc.Id
	}
	if mc.Status != "" {
		payload["status"] = mc.Status
	}
	if len(mc.Content) == 0 {
		payload["content"] = ""
	} else if isSimpleTextContent(mc.Content) {
		payload["content"] = TextFromContent(mc.Content)
	} else {
		payload["content"] = mc.Content
	}
	return json.Marshal(payload)
}

// parseReasoning extracts reasoning text and summary from a raw item.
func (o *OpenRouterConversationConverter) parseReasoning(raw json.RawMessage) ConversationElement {
	var item struct {
		Id      string                 `json:"id"`
		Status  string                 `json:"status"`
		Summary []string               `json:"summary"`
		Content []ReasoningContentPart `json:"content"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil // skip malformed items
	}

	textParts := make([]string, 0, len(item.Content))
	for _, part := range item.Content {
		if part.Type == "reasoning_text" || part.Type == "text" {
			textParts = append(textParts, part.Text)
		}
	}

	return &Reasoning{
		Id:      item.Id,
		Status:  item.Status,
		Summary: item.Summary,
		Text:    joinTextParts(textParts),
		// Signing data is intentionally not retained for replay compatibility.
	}
}

func (o *OpenRouterConversationConverter) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
	if len(elements) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, 0, len(elements))
	for _, e := range elements {
		switch el := e.(type) {
		case *UserMessage:
			payload := map[string]interface{}{
				"role":    el.Role,
				"type":    "message",
				"content": el.Content,
			}
			if el.Id != "" {
				payload["id"] = el.Id
			}
			if el.Status != "" {
				payload["status"] = el.Status
			}
			b, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *AssistantMessage:
			b, err := marshalTextMessage(el.messageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *SystemMessage:
			b, err := marshalTextMessage(el.messageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *FunctionCall:
			fc := map[string]interface{}{
				"id":        el.Id,
				"call_id":   el.CallId,
				"name":      el.Name,
				"arguments": json.RawMessage(el.Args),
				"status":    "completed",
				"type":      "function_call",
			}
			b, err := json.Marshal(fc)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *FunctionCallResp:
			resp := map[string]interface{}{
				"call_id": el.CallId,
				"type":    "function_call_output",
			}
			if el.Id != "" {
				resp["id"] = el.Id
			}
			// Use Output which may be JSON-encoded or plain text.
			resp["output"] = el.Output
			b, err := json.Marshal(resp)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *Reasoning:
			r := map[string]interface{}{
				"id":     el.Id,
				"status": el.Status,
				"type":   "reasoning",
			}
			if el.Summary != nil {
				r["summary"] = el.Summary
			}
			if el.Text != "" {
				r["content"] = []MessagePart{{Type: "reasoning_text", Text: el.Text}}
			}
			b, err := json.Marshal(r)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *ImageGeneration:
			sig := map[string]interface{}{
				"id":     el.Id,
				"type":   "image_generation_call",
				"status": el.Status,
				"result": el.Result,
			}
			b, err := json.Marshal(sig)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		default:
			continue // skip unknown types
		}
	}
	return raw, nil
}

var _ ConversationConverter = &OpenRouterConversationConverter{}

// NewOpenRouterEndpoint returns a configured endpoint for OpenRouter's API. apiKey and model are required.
func NewOpenRouterEndpoint(apiKey string, model Model) (*Endpoint, error) {
	if model == "" {
		return nil, ErrEndpointMissingModelName
	}
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key for OpenRouter endpoint")
	}
	header := make(http.Header)
	header.Set("Authorization", "Bearer "+apiKey)
	return &Endpoint{
		model:    model,
		provider: Provider_OpenRouter,
		rae:      &UniversalResponsesEndpoint{baseUrl: "https://router.openrouter.ai", responsesApiEndpoint: "/v1/responses", httpHeader: header},
		client:   &http.Client{},
	}, nil
}
