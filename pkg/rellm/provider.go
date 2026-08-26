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

	// ID reports the element's provider-assigned id ("" if none). Used for
	// logging and FunctionCall <-> FunctionCallResp pairing.
	ID() string
}

// MessageContent holds shared fields for role-typed messages. Role is the wire
// role field; the concrete type constrains it to a fixed value and drives
// per-type serialization shape. Exported so providers outside this package can
// construct and read message elements.
type MessageContent struct {
	Id      string        `json:"id,omitempty"`
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
func (m *UserMessage) ID() string      { return m.Id }

// AssistantMessage is a model-generated text message. Content serializes as a
// plain string on the wire.
type AssistantMessage struct {
	MessageContent
}

func (*AssistantMessage) Kind() ElementKind { return KindAssistantMessage }
func (m *AssistantMessage) ID() string      { return m.Id }

// SystemMessage is a system instruction. Content serializes as a plain string
// on the wire.
type SystemMessage struct {
	MessageContent
}

func (*SystemMessage) Kind() ElementKind { return KindSystemMessage }
func (m *SystemMessage) ID() string      { return m.Id }

// FunctionCall is a model-requested tool invocation.
type FunctionCall struct {
	Id     string          `json:"id"`
	Name   string          `json:"name"`
	Args   json.RawMessage `json:"arguments"`
	CallId string          `json:"call_id"`
}

func (*FunctionCall) Kind() ElementKind { return KindFunctionCall }
func (f *FunctionCall) ID() string      { return f.Id }

// FunctionCallResponse is the tool's result sent back to the model.
type FunctionCallResp struct {
	Id     string `json:"id,omitempty"`
	Type   string `json:"type"`
	CallId string `json:"call_id"`
	Output string `json:"output"` // raw output (may be JSON-encoded by some providers, plain text by others)
}

func (*FunctionCallResp) Kind() ElementKind { return KindFunctionCallResp }
func (f *FunctionCallResp) ID() string      { return f.Id }

// Reasoning captures model chain-of-thought output.
type Reasoning struct {
	Id        string   `json:"id,omitempty"`
	Status    string   `json:"status,omitempty"`
	Summary   []string `json:"summary,omitempty"`
	Text      string   `json:"text"`
	Signature string   `json:"signature,omitempty"` // OpenRouter-only signing key
}

func (*Reasoning) Kind() ElementKind { return KindReasoning }
func (r *Reasoning) ID() string      { return r.Id }

// ImageGeneration is a generated image with its result note.
type ImageGeneration struct {
	Id     string `json:"id,omitempty"`
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
	Result string `json:"result"` // user-visible identifier returned by the handler
}

func (*ImageGeneration) Kind() ElementKind { return KindImageGeneration }
func (i *ImageGeneration) ID() string      { return i.Id }

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

// parseImageGeneration parses an image_generation_call wire item into an
// ImageGeneration element.
func parseImageGeneration(raw json.RawMessage) (ConversationElement, error) {
	var ig struct {
		Id     string `json:"id"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(raw, &ig); err != nil {
		return nil, errors.Join(ErrImageParsingFailed, err)
	}
	return &ImageGeneration{Id: ig.Id, Status: ig.Status, Result: ig.Result}, nil
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
