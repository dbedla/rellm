package rellm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// NewOpenRouterProviderWithHTTPClient builds an OpenRouter provider with custom HTTP client. apiKey and model are required.
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

// Model reports the model this provider targets.
func (p *OpenRouterProvider) Model() Model { return p.model }

// Send performs one round trip against the OpenRouter Responses API
// endpoint. The agent guarantees req.Model is already set.
func (p *OpenRouterProvider) Send(ctx context.Context, req *ResponsesAPIReq) (*ResponsesAPIResp, error) {
	return postResponsesAPI(ctx, p.client, p.url, p.header, req)
}

func (p *OpenRouterProvider) ToConversationElements(items []json.RawMessage) ([]ConversationElement, error) {
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
			elements = append(elements, parseMessageWithStatus(raw, ProviderOpenRouter, msg.ID, msg.Type, msg.Role, msg.Content))

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
			elements = append(elements, newUnknownElement(ProviderOpenRouter, msg.Type, msg.Role, raw))
		}
	}
	return elements, nil
}

func (p *OpenRouterProvider) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
	if len(elements) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, 0, len(elements))
	for _, e := range elements {
		b, err := p.marshalConversationElement(e)
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

func (p *OpenRouterProvider) marshalConversationElement(element ConversationElement) (json.RawMessage, error) {
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
		return unknownElementRepresentation(el, ProviderOpenRouter)
	default:
		return nil, errors.Join(ErrOpenRouterMarshalingConversationElement, fmt.Errorf("unknown element type: %T", el))
	}
}

var _ Provider = &OpenRouterProvider{}
