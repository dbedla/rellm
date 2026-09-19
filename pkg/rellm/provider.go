package rellm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Provider is the single abstraction the Agent talks to. One implementation
// per backend (LM Studio, OpenRouter, ...). The Agent holds one Provider and
// never switches on backend anywhere. Each provider owns its transport
// (URL/headers/http client) and its wire-format translation
// (ConversationElement <-> provider JSON).
type Provider interface {
	// Model reports the model this provider sends requests to. The agent
	// stamps it onto each request (ResponsesAPIReq.Model) before the
	// InspectEachRequest hook runs, so inspectors always see a complete
	// request. Model names are provider-specific, so only the provider
	// knows them.
	Model() Model

	// Send one round trip: canonical request in, parsed response out.
	// The agent guarantees req.Model is already set (from Model) before
	// calling. Owns URL, headers, http.Client, and HTTP status handling.
	Send(ctx context.Context, req *ResponsesAPIReq) (*ResponsesAPIResp, error)

	// ToConversationElements parses a backend's raw response output into canonical
	// conversation elements.
	ToConversationElements(items []json.RawMessage) ([]ConversationElement, error)

	// ToProviderRepresentation serializes canonical elements back into this
	// backend's wire format for the next request's Input.
	ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error)
}

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
	KindUnknown          ElementKind = "unknown"
)

type ProviderTag string

const (
	ProviderTagProviderOpenRouter ProviderTag = "openrouter"
	ProviderTagProviderLMStudio   ProviderTag = "lmstudio"
	ProviderTagProviderOpenAI     ProviderTag = "openai"
)

// ConversationElement is a typed item in a conversation. Providers convert their
// raw wire output into these; the agent operates on them without knowing which
// provider produced them. The interface is intentionally NOT sealed: custom
// element types and external providers may implement it.
type ConversationElement interface {
	// Kind reports this element's canonical discriminator.
	Kind() ElementKind

	// Clone returns an independent copy of this element: mutating the
	// original or the clone must not affect the other. Copy every slice,
	// pointer, and map field; a struct copy suffices only while all fields
	// are immutable (strings, numbers, bools).
	Clone() ConversationElement
}

// cloneConversationElements returns independent copies of the elements via
// each element's Clone, so callers cannot alias stored history or report
// snapshots.
func cloneConversationElements(elements []ConversationElement) []ConversationElement {
	clones := make([]ConversationElement, 0, len(elements))
	for _, el := range elements {
		clones = append(clones, el.Clone())
	}
	return clones
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

func (e *UserMessage) Clone() ConversationElement {
	c := *e
	c.Content = cloneMessageParts(e.Content)
	return &c
}

// AssistantMessage is a model-generated text message. Content serializes as a
// plain string on the wire.
type AssistantMessage struct {
	MessageContent
}

func (*AssistantMessage) Kind() ElementKind { return KindAssistantMessage }

func (e *AssistantMessage) Clone() ConversationElement {
	c := *e
	c.Content = cloneMessageParts(e.Content)
	return &c
}

// SystemMessage is a system instruction. Content serializes as a plain string
// on the wire.
type SystemMessage struct {
	MessageContent
}

func (*SystemMessage) Kind() ElementKind { return KindSystemMessage }

func (e *SystemMessage) Clone() ConversationElement {
	c := *e
	c.Content = cloneMessageParts(e.Content)
	return &c
}

// FunctionCall is a model-requested tool invocation.
type FunctionCall struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Args   json.RawMessage `json:"arguments"`
	CallID string          `json:"call_id"`
}

func (*FunctionCall) Kind() ElementKind { return KindFunctionCall }

func (e *FunctionCall) Clone() ConversationElement {
	c := *e
	c.Args = append(json.RawMessage(nil), e.Args...)
	return &c
}

// FunctionCallResponse is the tool's result sent back to the model.
type FunctionCallResp struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"` // raw output (may be JSON-encoded by some providers, plain text by others)
}

func (*FunctionCallResp) Kind() ElementKind { return KindFunctionCallResp }

func (e *FunctionCallResp) Clone() ConversationElement {
	c := *e
	return &c
}

// ReasoningSummaryPart is a text summary supplied by a reasoning model.
type ReasoningSummaryPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Reasoning captures reasoning output and opaque provider continuation state.
type Reasoning struct {
	ID               string                 `json:"id,omitempty"`
	Status           string                 `json:"status,omitempty"`
	Summary          []ReasoningSummaryPart `json:"summary,omitempty"`
	Text             string                 `json:"text"`
	Signature        string                 `json:"signature,omitempty"` // OpenRouter-only signing key
	EncryptedContent string                 `json:"encrypted_content,omitempty"`
	Format           string                 `json:"format,omitempty"`
}

func (*Reasoning) Kind() ElementKind { return KindReasoning }

func (e *Reasoning) Clone() ConversationElement {
	c := *e
	c.Summary = append([]ReasoningSummaryPart(nil), e.Summary...)
	return &c
}

// ImageGeneration is a generated image with its result note.
type ImageGeneration struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type"`
	Status string `json:"status,omitempty"`
	Result string `json:"result"` // user-visible identifier returned by the handler
}

func (*ImageGeneration) Kind() ElementKind { return KindImageGeneration }

func (e *ImageGeneration) Clone() ConversationElement {
	c := *e
	return &c
}

// UnknownElement preserves a provider output item that rellm does not yet
// understand. Raw is the original provider representation; Provider identifies
// the wire format needed to interpret it, while Type and Role are routing hints.
type UnknownElement struct {
	Provider string          `json:"provider"`
	Type     string          `json:"type,omitempty"`
	Role     string          `json:"role,omitempty"`
	Raw      json.RawMessage `json:"raw"`
}

func (*UnknownElement) Kind() ElementKind { return KindUnknown }

func (e *UnknownElement) Clone() ConversationElement {
	c := *e
	c.Raw = append(json.RawMessage(nil), e.Raw...)
	return &c
}

func newUnknownElement(provider, itemType, role string, raw json.RawMessage) *UnknownElement {
	return &UnknownElement{
		Provider: provider,
		Type:     itemType,
		Role:     role,
		Raw:      append(json.RawMessage(nil), raw...),
	}
}

func unknownElementRepresentation(element *UnknownElement, provider string) (json.RawMessage, error) {
	if element == nil {
		return nil, fmt.Errorf("%w: nil unknown element", ErrMalformedUnknownElement)
	}
	if element.Provider != provider {
		return nil, fmt.Errorf(
			"%w: element belongs to %q, target provider is %q",
			ErrUnknownElementProviderMismatch,
			element.Provider,
			provider,
		)
	}
	if !json.Valid(element.Raw) {
		return nil, fmt.Errorf("%w: invalid raw JSON", ErrMalformedUnknownElement)
	}
	return append(json.RawMessage(nil), element.Raw...), nil
}

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

func (u *UnknownElement) MarshalJSON() ([]byte, error) {
	type plain UnknownElement
	return marshalWithKind(KindUnknown, (*plain)(u))
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
	case KindUnknown:
		var e UnknownElement
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, err
		}
		return &e, nil
	default:
		return nil, ErrUnknownTypeForConversationElement
	}
}

// TextFromContent extracts concatenated text from structured parts.
func TextFromContent(parts []MessagePart) string {
	texts := make([]string, 0, len(parts))
	for _, p := range parts {
		texts = append(texts, p.Text)
	}
	return JoinTextParts(texts)
}

// JoinTextParts joins non-empty text parts with spaces.
func JoinTextParts(parts []string) string {
	var sb strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(p)
	}
	return sb.String()
}

type ImageURL struct {
	URL string `json:"url"`
}

type MessagePart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

func cloneMessageParts(parts []MessagePart) []MessagePart {
	clones := make([]MessagePart, len(parts))
	for i, p := range parts {
		clones[i] = p
		if p.ImageURL != nil {
			u := *p.ImageURL
			clones[i].ImageURL = &u
		}
	}
	return clones
}

func StdToConversationElements(items []json.RawMessage, pTag ProviderTag) ([]ConversationElement, error) {
	elements := make([]ConversationElement, 0, len(items))
	for _, raw := range items {
		msg, err := parseWireItem(raw)
		if err != nil {
			return nil, err
		}

		switch msg.Type {
		case "function_call":
			elements = append(elements, parseFunctionCallItem(msg))

		case "function_call_output":
			elements = append(elements, parseFunctionCallRespItem(raw, msg))

		case "message", "": // messages often lack an explicit type field; infer from role
			elements = append(elements, parseMessageWithStatus(raw, string(pTag), msg.ID, msg.Type, msg.Role, msg.Content))

		case "reasoning":
			elm, err := parseReasoningSigned(raw)
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
			elements = append(elements, newUnknownElement(string(pTag), msg.Type, msg.Role, raw))
		}
	}
	return elements, nil
}

func StdToProviderRepresentation(elements []ConversationElement, pTag ProviderTag) ([]json.RawMessage, error) {
	if len(elements) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, 0, len(elements))
	for _, e := range elements {
		b, err := marshalConversationElement(e, pTag)
		if err != nil {
			return nil, err
		}
		if b == nil {
			continue
		}
		raw = append(raw, b)
	}
	return raw, nil
}

func marshalConversationElement(element ConversationElement, pTag ProviderTag) (json.RawMessage, error) {
	switch el := element.(type) {
	case *UserMessage:
		return marshalMessageAsTypedParts(el.MessageContent)
	case *AssistantMessage:
		return marshalMessageAsTypedText(el.MessageContent)
	case *SystemMessage:
		return marshalMessageAsTypedText(el.MessageContent)
	case *FunctionCall:
		return marshalFunctionCall(el)
	case *FunctionCallResp:
		return marshalFunctionCallResp(el)
	case *Reasoning:
		return marshalReasoningWithSignature(el)
	case *ImageGeneration:
		return marshalImageGeneration(el)
	case *UnknownElement:
		return unknownElementRepresentation(el, string(pTag))
	default:
		return nil, errors.Join(ErrOpenAIMarshalingConversationElement, fmt.Errorf("unknown element type: %T", el))
	}
}
