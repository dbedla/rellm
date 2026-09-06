package rellm

import (
	"encoding/json"
	"errors"
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
	client HTTPClient
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

func NewLMStudioProviderWithHTTPClient(model Model, host, port string, c HTTPClient) (*LMStudioProvider, error) {
	if c == nil {
		return nil, ErrMissingHTTPClientForProvider
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
		msg, err := parseWireItem(raw)
		if err != nil {
			return nil, err
		}

		switch msg.Type {
		case "function_call":
			elements = append(elements, parseFunctionCallItem(msg))

		case "reasoning":
			elements = append(elements, parseReasoningBasic(raw))

		case "image_generation_call":
			elem, err := parseImageGeneration(raw)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)

		case "function_call_output":
			elements = append(elements, parseFunctionCallRespItem(raw, msg))

		case "message", "": // messages often lack an explicit type field
			elements = append(elements, parseMessageByRole(raw, ProviderLMStudio, msg.ID, msg.Type, msg.Role, msg.Content))

		default:
			elements = append(elements, newUnknownElement(ProviderLMStudio, msg.Type, msg.Role, raw))
		}
	}
	return elements, nil
}

func (p *LMStudioProvider) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
	if len(elements) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, 0, len(elements))
	for _, e := range elements {
		b, err := p.marshalConversationElement(e)
		if err != nil {
			return nil, err
		}
		raw = append(raw, b)
	}
	return raw, nil
}

func (p *LMStudioProvider) marshalConversationElement(element ConversationElement) (json.RawMessage, error) {
	switch el := element.(type) {
	case *UserMessage:
		return marshalMessageAsUntypedParts(el.MessageContent)
	case *AssistantMessage:
		return marshalMessageAsUntypedParts(el.MessageContent)
	case *SystemMessage:
		return marshalMessageAsUntypedParts(el.MessageContent)
	case *FunctionCall:
		return marshalFunctionCall(el)
	case *FunctionCallResp:
		return marshalFunctionCallResp(el)
	case *Reasoning:
		return marshalReasoningWithSummary(el)
	case *ImageGeneration:
		return marshalImageGeneration(el)
	case *UnknownElement:
		return unknownElementRepresentation(el, ProviderLMStudio)
	default:
		return nil, errors.Join(ErrLMSMarshalingConversationElement, fmt.Errorf("unknown conversation element type: %T", el))
	}
}

var _ Provider = &LMStudioProvider{}
