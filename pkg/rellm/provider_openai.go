package rellm

import (
	"context"
	"encoding/json"
	"net/http"
)

// OpenAIProvider talks to the OpenAI Responses API.
type OpenAIProvider struct {
	model  Model
	url    string
	header http.Header
	client HTTPClient
}

const openAIDefaultURL = "https://api.openai.com/v1/responses"

// NewOpenAIProvider builds an OpenAIProvider provider. apiKey and model are required.
func NewOpenAIProvider(apiKey string, model Model) (*OpenAIProvider, error) {
	if model == "" {
		return nil, ErrEndpointMissingModelName
	}
	if apiKey == "" {
		return nil, ErrMissingApiKeyForProvider
	}
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer "+apiKey)
	return &OpenAIProvider{
		model:  model,
		url:    openAIDefaultURL,
		header: h,
		client: &http.Client{},
	}, nil
}

// NewOpenAIProviderWithHTTPClient builds an OpenAI provider with custom HTTP client. apiKey and model are required.
func NewOpenAIProviderWithHTTPClient(apiKey string, model Model, client HTTPClient) (*OpenAIProvider, error) {
	if client == nil {
		return nil, ErrMissingHTTPClientForProvider
	}
	p, err := NewOpenAIProvider(apiKey, model)
	if err != nil {
		return nil, err
	}
	p.client = client
	return p, nil
}

// Model reports the model this provider targets.
func (p *OpenAIProvider) Model() Model { return p.model }

// Send performs one round trip against the OpenAI Responses API
// endpoint. The agent guarantees req.Model is already set.
func (p *OpenAIProvider) Send(ctx context.Context, req *ResponsesAPIReq) (*ResponsesAPIResp, error) {
	return StdSendResponsesAPI(ctx, p.client, p.url, p.header, req)
}

func (p *OpenAIProvider) ToConversationElements(items []json.RawMessage) ([]ConversationElement, error) {
	return StdToConversationElements(items, ProviderTagOpenAI)
}

func (p *OpenAIProvider) ToProviderRepresentation(elements []ConversationElement) ([]json.RawMessage, error) {
	return StdToProviderRepresentation(elements)
}

var _ Provider = &OpenAIProvider{}
