package rellm

import (
	"encoding/json"
	"errors"
	"net/http"
)

// OpenRouterProvider talks to the OpenRouter Responses API.
type OpenRouterProvider struct {
	model  Model
	url    string
	header http.Header
	client HTTPClient
}

const openRouterDefaultURL = "https://openrouter.ai/api/v1/responses"

// NewOpenRouterProvider builds an OpenRouter provider. apiKey and model are required.
func NewOpenRouterProvider(apiKey string, model Model) (*OpenRouterProvider, error) {
	if model == "" {
		return nil, ErrEndpointMissingModelName
	}
	if apiKey == "" {
		return nil, ErrMissingApiKeyForProvider
	}
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer "+apiKey)
	return &OpenRouterProvider{
		model:  model,
		url:    openRouterDefaultURL,
		header: h,
		client: &http.Client{},
	}, nil
}

func NewOpenRouterProviderWithHTTPClient(apiKey string, model Model, client HTTPClient) (*OpenRouterProvider, error) {
	if client == nil {
		return nil, ErrMissingHTTPClientForProvider
	}
	p, err := NewOpenRouterProvider(apiKey, model)
	if err != nil {
		return nil, err
	}
	p.client = client
	return p, nil
}

func (p *OpenRouterProvider) Model() Model        { return p.model }
func (p *OpenRouterProvider) URL() string         { return p.url }
func (p *OpenRouterProvider) Header() http.Header { return p.header }
func (p *OpenRouterProvider) Do(r *http.Request) (*http.Response, error) {
	return p.client.Do(r)
}

func (p *OpenRouterProvider) ToConversationElements(items []json.RawMessage) ([]ConversationElement, error) {
	elements := make([]ConversationElement, 0, len(items))
	for _, raw := range items {
		var msg struct {
			Type    string          `json:"type"`
			Role    string          `json:"role"`
			ID      string          `json:"id"`
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
				ID:     msg.ID,
				Name:   msg.Name,
				Args:   msg.Args,
				CallID: msg.CallID,
			})

		case "function_call_output":
			// Extract output field directly from raw JSON.
			var out struct {
				Output string `json:"output"`
			}
			if err := json.Unmarshal(raw, &out); err == nil {
				elements = append(elements, &FunctionCallResp{
					ID:     msg.ID,
					Type:   "function_call_output",
					CallID: msg.CallID,
					Output: out.Output,
				})
			} else {
				elements = append(elements, &FunctionCallResp{
					ID:     msg.ID,
					Type:   "function_call_output",
					CallID: msg.CallID,
					Output: string(msg.Content),
				})
			}

		case "message", "": // messages often lack an explicit type field; infer from role
			elements = append(elements, p.parseMessage(raw, msg.ID, msg.Role, msg.Content))

		case "reasoning":
			elm, err := p.parseReasoning(raw)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elm)

		case "image_generation_call":
			elem, err := parseImageGeneration(raw)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)

		default:
			// Unknown types pass through unchanged.
		}
	}
	return elements, nil
}

// parseMessage parses a message item into a role-typed message. Content is
// normalized to []MessagePart; ToProviderRepresentation re-serializes per the type's shape rule.
func (p *OpenRouterProvider) parseMessage(raw json.RawMessage, id, role string, content json.RawMessage) ConversationElement {
	var statusInfo struct {
		Status string `json:"status"`
	}
	status := ""
	if err := json.Unmarshal(raw, &statusInfo); err == nil {
		status = statusInfo.Status
	}
	parts := normalizeMessageParts(ParseMessageContent(content), role)
	switch role {
	case "assistant":
		return &AssistantMessage{MessageContent{ID: id, Role: role, Status: status, Content: parts}}
	case "system":
		return &SystemMessage{MessageContent{ID: id, Role: role, Content: parts}}
	default: // "user" or unknown
		return &UserMessage{MessageContent{ID: id, Role: role, Status: status, Content: parts}}
	}
}

// marshalTextMessage serializes an assistant/system message for OpenRouter:
// text-only content becomes a string (matching the observed wire format,
// which stringifies output_text parts too); multimodal stays an array.
func (p *OpenRouterProvider) marshalTextMessage(mc MessageContent) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"role": mc.Role,
		"type": "message",
	}
	if mc.ID != "" {
		payload["id"] = mc.ID
	}
	if mc.Status != "" {
		payload["status"] = mc.Status
	}
	if len(mc.Content) == 0 {
		payload["content"] = ""
	} else if hasImageParts(mc.Content) {
		payload["content"] = mc.Content
	} else {
		payload["content"] = TextFromContent(mc.Content)
	}
	return json.Marshal(payload)
}

// hasImageParts reports whether any part carries an image (multimodal content
// that must stay a structured array).
func hasImageParts(parts []MessagePart) bool {
	for _, p := range parts {
		if p.ImageURL != nil {
			return true
		}
	}
	return false
}

// parseReasoning extracts reasoning text, summary, and provider continuation state.
func (p *OpenRouterProvider) parseReasoning(raw json.RawMessage) (ConversationElement, error) {
	var item struct {
		ID        string                 `json:"id"`
		Status    string                 `json:"status"`
		Summary   []string               `json:"summary"`
		Content   []ReasoningContentPart `json:"content"`
		Signature string                 `json:"signature"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, errors.Join(ErrReasoningParsingFailed, err) // skip malformed items
	}

	textParts := make([]string, 0, len(item.Content))
	for _, part := range item.Content {
		if part.Type == "reasoning_text" || part.Type == "text" {
			textParts = append(textParts, part.Text)
		}
	}
	if len(item.Summary) == 0 {
		item.Summary = nil
	}

	return &Reasoning{
		ID:        item.ID,
		Status:    item.Status,
		Summary:   item.Summary,
		Text:      JoinTextParts(textParts),
		Signature: item.Signature,
	}, nil
}

func (p *OpenRouterProvider) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
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
			if el.ID != "" {
				payload["id"] = el.ID
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
			b, err := p.marshalTextMessage(el.MessageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *SystemMessage:
			b, err := p.marshalTextMessage(el.MessageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *FunctionCall:
			fc := map[string]interface{}{
				"id":        el.ID,
				"call_id":   el.CallID,
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
				"call_id": el.CallID,
				"type":    "function_call_output",
			}
			if el.ID != "" {
				resp["id"] = el.ID
			}
			// Use Output which may be JSON-encoded or plain text.
			resp["output"] = el.Output
			b, err := json.Marshal(resp)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *Reasoning:
			// A signature-only reasoning item carries provider continuation state
			// and must be replayed even though it has no user-visible text.
			if el.Text == "" && el.Signature == "" && len(el.Summary) == 0 {
				continue
			}
			r := map[string]interface{}{
				"id":     el.ID,
				"status": el.Status,
				"type":   "reasoning",
			}
			// Signed reasoning blocks are provider continuation state; preserve
			// their summary field even when it is empty.
			if len(el.Summary) > 0 {
				r["summary"] = el.Summary
			} else if el.Signature != "" {
				r["summary"] = []string{}
			}
			if el.Text != "" {
				r["content"] = []MessagePart{{Type: "reasoning_text", Text: el.Text}}
			}
			if el.Signature != "" {
				r["signature"] = el.Signature
			}
			b, err := json.Marshal(r)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *ImageGeneration:
			sig := map[string]interface{}{
				"id":     el.ID,
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

var _ Provider = &OpenRouterProvider{}

type ReasoningContentPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
