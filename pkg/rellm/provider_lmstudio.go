package rellm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// LMStudioProvider talks to a local LM Studio Responses API endpoint.
type LMStudioProvider struct {
	model  Model
	url    string
	header http.Header
	client ClientHttpDo
}

// NewLMStudioProvider builds an LM Studio provider. model, host and port are required.
func NewLMStudioProvider(model Model, host, port string) (*LMStudioProvider, error) {
	if model == "" {
		return nil, ErrEndpointMissingModelName
	}
	if host == "" {
		return nil, ErrEndpointMissingHost
	}
	if port == "" {
		return nil, ErrEndpointMissingPort
	}

	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	baseURL := fmt.Sprintf("%s:%s/v1/responses", strings.TrimRight(host, "/"), port)
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid LM Studio URL %q: %w", baseURL, err)
	}

	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	return &LMStudioProvider{
		model:  model,
		url:    baseURL,
		header: h,
		client: &http.Client{},
	}, nil
}

func NewLMStudioProviderWithHTTPClient(model Model, host, port string, c ClientHttpDo) (*LMStudioProvider, error) {
	if c == nil {
		return nil, ErrMissingHttpClientForProvider
	}

	p, err := NewLMStudioProvider(model, host, port)
	if err != nil {
		return nil, err
	}
	p.client = c

	return p, nil
}

func (p *LMStudioProvider) Model() Model        { return p.model }
func (p *LMStudioProvider) URL() string         { return p.url }
func (p *LMStudioProvider) Header() http.Header { return p.header }
func (p *LMStudioProvider) Do(r *http.Request) (*http.Response, error) {
	return p.client.Do(r)
}

func (p *LMStudioProvider) ToConversationElements(items []json.RawMessage) ([]ConversationElement, error) {
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

		case "reasoning":
			elements = append(elements, p.parseReasoning(raw))

		case "image_generation_call":
			elem, err := parseImageGeneration(raw)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)

		case "function_call_output":
			// Extract output field directly from raw JSON
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

		case "message", "": // messages often lack an explicit type field
			parts := normalizeMessageParts(ParseMessageContent(msg.Content), msg.Role)
			switch msg.Role {
			case "assistant":
				elements = append(elements, &AssistantMessage{MessageContent{ID: msg.ID, Role: msg.Role, Content: parts}})
			case "system":
				elements = append(elements, &SystemMessage{MessageContent{ID: msg.ID, Role: msg.Role, Content: parts}})
			default: // "user" or unknown
				elements = append(elements, &UserMessage{MessageContent{ID: msg.ID, Role: msg.Role, Content: parts}})
			}
		}
	}
	return elements, nil
}

// parseReasoning extracts reasoning text and summary from a raw item.
func (p *LMStudioProvider) parseReasoning(raw json.RawMessage) ConversationElement {
	r := &Reasoning{}

	// Extract id, status, summary from top level
	var meta struct {
		ID      string   `json:"id"`
		Status  string   `json:"status"`
		Summary []string `json:"summary"`
	}
	if err := json.Unmarshal(raw, &meta); err == nil {
		r.ID = meta.ID
		r.Status = meta.Status
		if len(meta.Summary) > 0 {
			r.Summary = meta.Summary
		}
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
		r.Text = JoinTextParts(textParts)
	}

	return r
}

// marshalMessage serializes any role-typed message for LM Studio: content is
// always a structured array (LM Studio does not use the "type":"message" field).
func (p *LMStudioProvider) marshalMessage(mc MessageContent) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"role": mc.Role,
	}
	if mc.ID != "" {
		payload["id"] = mc.ID
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

func (p *LMStudioProvider) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
	if len(elements) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, 0, len(elements))
	for _, e := range elements {
		switch el := e.(type) {
		case *UserMessage:
			b, err := p.marshalMessage(el.MessageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *AssistantMessage:
			b, err := p.marshalMessage(el.MessageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *SystemMessage:
			b, err := p.marshalMessage(el.MessageContent)
			if err != nil {
				return nil, err
			}
			raw = append(raw, b)

		case *FunctionCall:
			fc := map[string]interface{}{
				"id":        el.ID,
				"name":      el.Name,
				"arguments": json.RawMessage(el.Args),
				"call_id":   el.CallID,
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
			// Always include summary field (even if empty) for faithful round-trip.
			payload := map[string]interface{}{
				"id":     el.ID,
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

var _ Provider = &LMStudioProvider{}

// JoinTextParts joins text parts with spaces.
func JoinTextParts(parts []string) string {
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
