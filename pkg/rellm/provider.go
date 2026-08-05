package rellm

import (
	"encoding/json"
	"strings"
)

// --- Canonical conversation element types ------------------------------------

// ElementType identifies the kind of a conversation element.
type ElementType string

const (
	ElementTypeMessage          ElementType = "message"
	ElementTypeFunctionCall     ElementType = "function_call"
	ElementTypeFunctionCallResp ElementType = "function_call_response"
	ElementTypeReasoning        ElementType = "reasoning"
	ElementTypeImageGeneration  ElementType = "image_generation"
)

// ConversationElement is a typed item in a conversation. Providers convert their
// raw wire output into these; the agent operates on them without knowing which
// provider produced them.
type ConversationElement interface {
	elementType() ElementType
}

// messageContent holds shared fields for role-typed messages. Role is the wire
// role field; the concrete type constrains it to a fixed value and drives
// per-type serialization shape.
type messageContent struct {
	Id      string        `json:"id,omitempty"`
	Role    string        `json:"role"`
	Status  string        `json:"status,omitempty"`
	Content []MessagePart `json:"content"` // structured parts (text + images etc.)
}

// UserMessage is a user-authored message. Content is always a parts array on
// the wire (supports multimodal: text, images, files).
type UserMessage struct {
	messageContent
}

func (*UserMessage) elementType() ElementType { return ElementTypeMessage }

// AssistantMessage is a model-generated text message. Content serializes as a
// plain string on the wire.
type AssistantMessage struct {
	messageContent
}

func (*AssistantMessage) elementType() ElementType { return ElementTypeMessage }

// SystemMessage is a system instruction. Content serializes as a plain string
// on the wire.
type SystemMessage struct {
	messageContent
}

func (*SystemMessage) elementType() ElementType { return ElementTypeMessage }

// FunctionCall is a model-requested tool invocation.
type FunctionCall struct {
	Id     string          `json:"id"`
	Name   string          `json:"name"`
	Args   json.RawMessage `json:"arguments"`
	CallId string          `json:"call_id"`
}

func (*FunctionCall) elementType() ElementType { return ElementTypeFunctionCall }

// FunctionCallResponse is the tool's result sent back to the model.
type FunctionCallResp struct {
	Id     string `json:"id,omitempty"`
	Type   string `json:"type"`
	CallId string `json:"call_id"`
	Output string `json:"output"` // raw output (may be JSON-encoded by some providers, plain text by others)
}

func (*FunctionCallResp) elementType() ElementType { return ElementTypeFunctionCallResp }

// Reasoning captures model chain-of-thought output.
type Reasoning struct {
	Id        string   `json:"id,omitempty"`
	Status    string   `json:"status,omitempty"`
	Summary   []string `json:"summary,omitempty"`
	Text      string   `json:"text"`
	Signature string   `json:"signature,omitempty"` // OpenRouter-only signing key
}

func (*Reasoning) elementType() ElementType { return ElementTypeReasoning }

// ImageGeneration is a generated image with its result note.
type ImageGeneration struct {
	Id     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Result string `json:"result"` // user-visible identifier returned by the handler
}

func (*ImageGeneration) elementType() ElementType { return ElementTypeImageGeneration }

// isSimpleTextContent returns true if the content array contains only simple text entries.
func isSimpleTextContent(parts []MessagePart) bool {
	for _, p := range parts {
		if p.Type != "text" && p.Type != "input_text" {
			return false
		}
	}
	return true
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

type LLMProvider struct {
	cc       ConversationConverter
	endpoint *Endpoint
}

type ConversationConverter interface {
	ToConversationElements(items []json.RawMessage) (elements []ConversationElement, err error)
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

// parseMessageContent normalizes a wire content field (string, []string, or
// []MessagePart) into []MessagePart.
func parseMessageContent(content json.RawMessage) []MessagePart {
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
func parseImageGeneration(raw json.RawMessage) ConversationElement {
	var ig struct {
		Id     string `json:"id"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal(raw, &ig); err != nil {
		return nil // skip malformed items
	}
	return &ImageGeneration{Id: ig.Id, Status: ig.Status, Result: ig.Result}
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
