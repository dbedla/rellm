package rellm

import (
	"encoding/json"
	"net/http"
	"strings"
)

// --- Wire format: lmstudioStyle -----------------------------------------------

// lmstudioStyle implements FromWire/ToWire for the LM Studio wire format (pass-through).
type lmstudioStyle struct{}

func (l *lmstudioStyle) FromWire(items []json.RawMessage) ([]ConversationElement, error) {
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
			var textParts []string
			if err := json.Unmarshal(msg.Content, &textParts); err == nil {
				elements = append(elements, &TextMessage{Id: msg.Id, Role: msg.Role, Content: messagePartsWithStrings(textParts)})
				continue
			}

			var text string
			if err := json.Unmarshal(msg.Content, &text); err == nil {
				elements = append(elements, &TextMessage{Id: msg.Id, Role: msg.Role, Content: messagePartsWithStrings([]string{text})})
				continue
			}

			var parts []MessagePart
			if err := json.Unmarshal(msg.Content, &parts); err == nil {
				elements = append(elements, &TextMessage{Id: msg.Id, Role: msg.Role, Content: parts})
				continue
			}

			elements = append(elements, &TextMessage{Id: msg.Id, Role: msg.Role, Content: nil})
		}
	}
	return elements, nil
}

// parseReasoning extracts reasoning text and summary from a raw item.
func (l *lmstudioStyle) parseReasoning(raw json.RawMessage) ConversationElement {
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

func (l *lmstudioStyle) ToWire(elements []ConversationElement) ([]json.RawMessage, error) {
	if len(elements) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, 0, len(elements))
	for _, e := range elements {
		switch el := e.(type) {
		case *TextMessage:
			// LM Studio always uses structured content arrays.
			payload := map[string]interface{}{
				"role": el.Role,
			}
			if el.Id != "" {
				payload["id"] = el.Id
			}
			if el.Status != "" {
				payload["status"] = el.Status
			}
			if len(el.Content) == 0 {
				payload["content"] = []MessagePart{}
			} else {
				payload["content"] = el.Content
			}
			b, err := json.Marshal(payload)
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

		default:
			continue // skip unknown types
		}
	}
	return raw, nil
}

// --- Provider -----------------------------------------------------------------

// lmstudioProvider handles LM Studio-specific configuration (URL, port).
type lmstudioProvider struct {
	lmstudioStyle
}

var _ ProviderConfig = &lmstudioProvider{}

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
		rae:      &UniversalResponsesEndpoint{baseUrl: "http://" + host, port: port, responsesApiEndpoint: "/v1/responses"},
		client:   &http.Client{},
	}, nil
}
