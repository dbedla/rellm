package rellm

import (
	"context"
	"encoding/json"
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
	return StdToConversationElements(items, ProviderTagProviderOpenRouter)
}

func (p *OpenRouterProvider) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
	return StdToProviderRepresentation(elements)
}

var _ Provider = &OpenRouterProvider{}
