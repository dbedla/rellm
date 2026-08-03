package rellm

import (
	"encoding/json"
	"net/http"
	"strings"
)

type LMSConversationConverter struct{}

func (l *LMSConversationConverter) ToConversationElements(items []json.RawMessage) ([]ConversationElement, error) {
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

		case "reasoning":
			elements = append(elements, l.parseReasoning(raw))

		case "image_generation_call":
			elements = append(elements, parseImageGeneration(raw))

		case "function_call_output":
			// Extract output field directly from raw JSON
			var out struct {
				Output string `json:"output"`
			}
			if err := json.Unmarshal(raw, &out); err == nil {
				elements = append(elements, &FunctionCallResponse{
					Id:     msg.Id,
					CallId: msg.CallID,
					Output: out.Output,
				})
			} else {
				elements = append(elements, &FunctionCallResponse{
					Id:     msg.Id,
					CallId: msg.CallID,
					Output: string(msg.Content),
				})
			}

		case "message", "": // messages often lack an explicit type field
			parts := parseMessageContent(msg.Content)
			switch msg.Role {
			case "assistant":
				elements = append(elements, &AssistantMessage{messageContent{Id: msg.Id, Role: msg.Role, Content: parts}})
			case "system":
				elements = append(elements, &SystemMessage{messageContent{Id: msg.Id, Role: msg.Role, Content: parts}})
			default: // "user" or unknown
				elements = append(elements, &UserMessage{messageContent{Id: msg.Id, Role: msg.Role, Content: parts}})
			}
		}
	}
	return elements, nil
}

// parseReasoning extracts reasoning text and summary from a raw item.
func (l *LMSConversationConverter) parseReasoning(raw json.RawMessage) ConversationElement {
	r := &Reasoning{}

	// Extract id, status, summary from top level
	var meta struct {
		Id      string   `json:"id"`
		Status  string   `json:"status"`
		Summary []string `json:"summary"`
	}
	if err := json.Unmarshal(raw, &meta); err == nil {
		r.Id = meta.Id
		r.Status = meta.Status
		r.Summary = meta.Summary
	}

	// Extract text from content array (reasoning_text items)
	var msg struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &msg); err == nil {
		var textParts []string
		for _, c := range msg.Content {
			if c.Type == "reasoning_text" || c.Type == "text" {
				textParts = append(textParts, c.Text)
			}
		}
		r.Text = joinTextParts(textParts)
	}

	return r
}

// joinTextParts joins text parts with spaces.
func joinTextParts(parts []string) string {
	var sb strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i > 0 && sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(p)
	}
	return sb.String()
}

// marshalMessage serializes any role-typed message for LM Studio: content is
// always a structured array (LM Studio does not use the "type":"message" field).
func (l *LMSConversationConverter) marshalMessage(mc messageContent) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"role": mc.Role,
	}
	if mc.Id != "" {
		payload["id"] = mc.Id
	}
	if mc.Status != "" {
		payload["status"] = mc.Status
	}
	if len(mc.Content) == 0 {
		payload["content"] = []MessagePart{}
	} else {
		payload["content"] = mc.Content
	}
	return json.Marshal(payload)
}

func (l *LMSConversationConverter) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
	if len(elements) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, 0, len(elements))
	for _, e := range elements {
		switch el := e.(type) {
		case *UserMessage:
			b, err := l.marshalMessage(el.messageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *AssistantMessage:
			b, err := l.marshalMessage(el.messageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *SystemMessage:
			b, err := l.marshalMessage(el.messageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *FunctionCall:
			fc := map[string]interface{}{
				"id":        el.Id,
				"name":      el.Name,
				"arguments": json.RawMessage(el.Args),
				"call_id":   el.CallId,
				"status":    "completed",
				"type":      "function_call",
			}
			b, err := json.Marshal(fc)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *FunctionCallResponse:
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
			// Always include summary field (even if empty) for faithful round-trip.
			payload := map[string]interface{}{
				"id":     el.Id,
				"status": el.Status,
				"type":   "reasoning",
			}
			payload["summary"] = []string{}
			if len(el.Summary) > 0 {
				payload["summary"] = el.Summary
			}
			if el.Text != "" {
				payload["content"] = []MessagePart{{Type: "reasoning_text", Text: el.Text}}
			}
			b, err := json.Marshal(payload)
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

var _ ConversationConverter = &LMSConversationConverter{}

// NewLMStudioEndpoint returns a configured endpoint for LM Studio. All arguments are required.
func NewLMStudioEndpoint(model Model, host string, port string) (*Endpoint, error) {
	if model == "" {
		return nil, ErrEndpointMissingModelName
	}
	if host == "" {
		return nil, ErrEndpointMissingHost
	}
	if port == "" {
		return nil, ErrEndpointMissingPort
	}
	return &Endpoint{
		model:    model,
		provider: Provider_LMStudio,
		rae:      &UniversalResponsesEndpoint{baseUrl: host, port: port, responsesApiEndpoint: "/v1/responses"},
		client:   &http.Client{},
	}, nil
}
