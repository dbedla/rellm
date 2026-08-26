package rellm

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// --- Canonical conversation element types ------------------------------------

// ElementKind discriminates conversation elements at runtime. It is
// canonical (provider-independent): wire "type" values are translated to
// kinds by each provider's parser.
type ElementKind string

const (
	KindUserMessage      ElementKind = "user_message"
	KindAssistantMessage ElementKind = "assistant_message"
	KindSystemMessage    ElementKind = "system_message"
	KindFunctionCall     ElementKind = "function_call"
	KindFunctionCallResp ElementKind = "function_call_response"
	KindReasoning        ElementKind = "reasoning"
	KindImageGeneration  ElementKind = "image_generation"
)

// ConversationElement is a typed item in a conversation. Providers convert their
// raw wire output into these; the agent operates on them without knowing which
// provider produced them. The interface is intentionally NOT sealed: custom
// element types and external providers may implement it.
type ConversationElement interface {
	// Kind reports this element's canonical discriminator.
	Kind() ElementKind
}

// MessageContent holds shared fields for role-typed messages. Role is the wire
// role field; the concrete type constrains it to a fixed value and drives
// per-type serialization shape. Exported so providers outside this package can
// construct and read message elements.
type MessageContent struct {
	ID      string        `json:"id,omitempty"`
	Role    string        `json:"role"`
	Status  string        `json:"status,omitempty"`
	Content []MessagePart `json:"content"` // structured parts (text + images etc.)
}

// UserMessage is a user-authored message. Content is always a parts array on
// the wire (supports multimodal: text, images, files).
type UserMessage struct {
	MessageContent
}

func (*UserMessage) Kind() ElementKind { return KindUserMessage }

// AssistantMessage is a model-generated text message. Content serializes as a
// plain string on the wire.
type AssistantMessage struct {
	MessageContent
}

func (*AssistantMessage) Kind() ElementKind { return KindAssistantMessage }

// SystemMessage is a system instruction. Content serializes as a plain string
// on the wire.
type SystemMessage struct {
	MessageContent
}

func (*SystemMessage) Kind() ElementKind { return KindSystemMessage }

// FunctionCall is a model-requested tool invocation.
type FunctionCall struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Args   json.RawMessage `json:"arguments"`
	CallID string          `json:"call_id"`
}

func (*FunctionCall) Kind() ElementKind { return KindFunctionCall }

// FunctionCallResponse is the tool's result sent back to the model.
type FunctionCallResp struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"` // raw output (may be JSON-encoded by some providers, plain text by others)
}

func (*FunctionCallResp) Kind() ElementKind { return KindFunctionCallResp }

// Reasoning captures model chain-of-thought output.
type Reasoning struct {
	ID        string   `json:"id,omitempty"`
	Status    string   `json:"status,omitempty"`
	Summary   []string `json:"summary,omitempty"`
	Text      string   `json:"text"`
	Signature string   `json:"signature,omitempty"` // OpenRouter-only signing key
}

func (*Reasoning) Kind() ElementKind { return KindReasoning }

// ImageGeneration is a generated image with its result note.
type ImageGeneration struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
	Result string `json:"result"` // user-visible identifier returned by the handler
}

func (*ImageGeneration) Kind() ElementKind { return KindImageGeneration }

// marshalWithKind serializes v and injects the canonical "kind" discriminator.
func marshalWithKind(kind ElementKind, v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	m["kind"], err = json.Marshal(string(kind))
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// Canonical serialization: storage (and anything else) marshals elements to a
// self-describing form keyed on "kind". Provider wire serialization stays in
// ToProviderRepresentation and never goes through these methods.

func (m *UserMessage) MarshalJSON() ([]byte, error) {
	return marshalWithKind(KindUserMessage, &m.MessageContent)
}

func (m *AssistantMessage) MarshalJSON() ([]byte, error) {
	return marshalWithKind(KindAssistantMessage, &m.MessageContent)
}

func (m *SystemMessage) MarshalJSON() ([]byte, error) {
	return marshalWithKind(KindSystemMessage, &m.MessageContent)
}

func (f *FunctionCall) MarshalJSON() ([]byte, error) {
	type plain FunctionCall
	return marshalWithKind(KindFunctionCall, (*plain)(f))
}

func (f *FunctionCallResp) MarshalJSON() ([]byte, error) {
	type plain FunctionCallResp
	return marshalWithKind(KindFunctionCallResp, (*plain)(f))
}

func (r *Reasoning) MarshalJSON() ([]byte, error) {
	type plain Reasoning
	return marshalWithKind(KindReasoning, (*plain)(r))
}

func (i *ImageGeneration) MarshalJSON() ([]byte, error) {
	type plain ImageGeneration
	return marshalWithKind(KindImageGeneration, (*plain)(i))
}

// ParseConversationElement decodes a canonical (kind-tagged) element produced
// by MarshalJSON. The "kind" field itself is ignored by the concrete structs'
// unmarshaling.
func ParseConversationElement(raw json.RawMessage) (ConversationElement, error) {
	var k struct {
		Kind ElementKind `json:"kind"`
	}
	if err := json.Unmarshal(raw, &k); err != nil {
		return nil, err
	}
	switch k.Kind {
	case KindUserMessage:
		var e UserMessage
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case KindAssistantMessage:
		var e AssistantMessage
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case KindSystemMessage:
		var e SystemMessage
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case KindFunctionCall:
		var e FunctionCall
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case KindFunctionCallResp:
		var e FunctionCallResp
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case KindReasoning:
		var e Reasoning
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	case KindImageGeneration:
		var e ImageGeneration
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	default:
		return nil, ErrUnknownConversationElement
	}
}

// TextFromContent extracts concatenated text from structured parts.
// Works on both string-based and MessagePart-based content.
func TextFromContent(parts []MessagePart) string {
	var sb strings.Builder
	for i, p := range parts {
		if p.Text == "" {
			continue
		}
		if i > 0 && sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(p.Text)
	}
	return sb.String()
}

// Provider is the single abstraction the Agent talks to. One implementation
// per backend (LM Studio, OpenRouter, ...). The Agent holds one Provider and
// never switches on backend anywhere. Each provider owns its transport
// (URL/headers/http client) and its wire-format translation
// (ConversationElement <-> provider JSON).
type Provider interface {
	// Model name as configured (fills ResponsesApiReq.Model).
	Model() Model

	// Transport shim: send a pre-built http.Request, return the raw response.
	// Owns only the http.Client. Agent.post builds the request from
	// ResponsesApiReq + URL() + Header() and handles status/unmarshal.
	Do(request *http.Request) (*http.Response, error)

	// URL for the Responses API endpoint this backend talks to.
	URL() string

	// HTTP headers for this backend (Authorization, Content-Type, ...).
	Header() http.Header

	// Parse a backend's raw response output into canonical conversation
	// elements.
	ToConversationElements(items []json.RawMessage) ([]ConversationElement, error)

	// Serialize canonical elements back into this backend's wire format for
	// the next request's Input.
	ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error)
}

// messagePartsWithStrings creates MessagePart slice from a string list.
func messagePartsWithStrings(texts []string) []MessagePart {
	parts := make([]MessagePart, 0, len(texts))
	for _, t := range texts {
		parts = append(parts, MessagePart{Type: "input_text", Text: t})
	}
	return parts
}

// ParseMessageContent normalizes a wire content field (string, []string, or
// []MessagePart) into []MessagePart. Exported for external providers that need
// the same content normalization when parsing wire messages.
func ParseMessageContent(content json.RawMessage) []MessagePart {
	var textStr string
	if json.Unmarshal(content, &textStr) == nil {
		return messagePartsWithStrings([]string{textStr})
	}
	var textParts []string
	if json.Unmarshal(content, &textParts) == nil {
		return messagePartsWithStrings(textParts)
	}
	var parts []MessagePart
	if json.Unmarshal(content, &parts) == nil {
		return parts
	}
	return nil
}

// normalizeMessageParts canonicalizes provider text parts independently of
// the shape used on the wire. Providers may replay text as a plain string or
// return it as an input_text/output_text object; the canonical representation
// is determined by the message role.
func normalizeMessageParts(parts []MessagePart, role string) []MessagePart {
	textType := "input_text"
	if role == "assistant" {
		textType = "output_text"
	}

	for i := range parts {
		if parts[i].Type == "input_text" || parts[i].Type == "output_text" {
			parts[i].Type = textType
		}
		if len(parts[i].Annotations) == 0 {
			parts[i].Annotations = nil
		}
		if len(parts[i].Logprobs) == 0 {
			parts[i].Logprobs = nil
		}
	}

	return parts
}

// parseImageGeneration parses an image_generation_call wire item into an
// ImageGeneration element.
func parseImageGeneration(raw json.RawMessage) (ConversationElement, error) {
	var ig struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(raw, &ig); err != nil {
		return nil, errors.Join(ErrImageParsingFailed, err)
	}
	return &ImageGeneration{ID: ig.ID, Status: ig.Status, Result: ig.Result}, nil
}

type ImageURL struct {
	URL string `json:"url"`
}

type MessagePart struct {
	Type        string        `json:"type"`
	Text        string        `json:"text,omitempty"`
	ImageURL    *ImageURL     `json:"image_url,omitempty"`
	Annotations []interface{} `json:"annotations,omitempty"`
	Logprobs    []interface{} `json:"logprobs,omitempty"`
}
