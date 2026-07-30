package rellm

import (
	"encoding/json"
	"errors"
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

// TextMessage is a user/assistant/system message with text or multimodal parts.
type TextMessage struct {
	Id        string        `json:"id,omitempty"`
	Role      string        `json:"role"` // "user", "assistant", "system"
	Status    string        `json:"status,omitempty"`
	Content   []MessagePart `json:"content"` // structured parts (text + images etc.)
	hasType   bool          // private: track if original had type field
	isRawText bool          // private: true if content was originally a plain string (not structured)
}

func (*TextMessage) elementType() ElementType { return ElementTypeMessage }

// FunctionCall is a model-requested tool invocation.
type FunctionCall struct {
	Id     string          `json:"id"`
	Name   string          `json:"name"`
	Args   json.RawMessage `json:"arguments"`
	CallId string          `json:"call_id"`
}

func (*FunctionCall) elementType() ElementType { return ElementTypeFunctionCall }

// FunctionCallResponse is the tool's result sent back to the model.
type FunctionCallResponse struct {
	Id     string `json:"id,omitempty"`
	CallId string `json:"call_id"`
	Output string `json:"output"` // raw output (may be JSON-encoded by some providers, plain text by others)
}

func (*FunctionCallResponse) elementType() ElementType { return ElementTypeFunctionCallResp }

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

type ProviderConfig interface {
	// FromWire parses the provider's response items into canonical conversation
	// elements. Each item becomes one or more ConversationElements plus any
	// extracted text for that element.
	FromWire(items []json.RawMessage) (elements []ConversationElement, err error)

	// ToWire serializes canonical conversation elements back into a JSON-encoded
	// list of wire-format messages for this provider's request payload.
	ToWire(elements []ConversationElement) ([]json.RawMessage, error)
}

// --- Provider registry -------------------------------------------------------

var providerConfigs = map[Provider]ProviderConfig{
	Provider_OpenRouter: &openrouterProvider{},
	Provider_LMStudio:   &lmstudioProvider{},
}

func getProviderConfig(p Provider) (ProviderConfig, error) {
	cfg, ok := providerConfigs[p]
	if !ok {
		return nil, errors.New("rellm: unknown provider " + string(p))
	}
	return cfg, nil
}

// messagePartsWithStrings creates MessagePart slice from a string list.
func messagePartsWithStrings(texts []string) []MessagePart {
	parts := make([]MessagePart, 0, len(texts))
	for _, t := range texts {
		parts = append(parts, MessagePart{Type: "input_text", Text: t})
	}
	return parts
}
