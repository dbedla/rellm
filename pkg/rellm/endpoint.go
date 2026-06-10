package rellm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"rellm/pkg/utils"
)

type ResponseApiEndpoint interface {
	GetUrl() string
	GetHttpHeader() http.Header
}

type Model string

const (
	Model_Gpt5_4                 Model = "openai/gpt-5.4"
	Model_Gemini_3_flash_preview Model = "google/gemini-3-flash-preview"
	Model_Claude_sonnet_4_6      Model = "anthropic/claude-sonnet-4-6"

	Model_Gpt5_4_mini      Model = "openai/gpt-5.4-mini"
	Model_Gemini_2_5_flash Model = "google/gemini-2.5-flash"

	Model_Gemma_4_26b_a4b = "google/gemma-4-26b-a4b"
	Model_gemma_4_12b_qat = "google/gemma-4-12b-qat"
)

type ClientHttpDo interface {
	Do(request *http.Request) (*http.Response, error)
}

type Endpoint struct {
	client ClientHttpDo
	model  Model
	rae    ResponseApiEndpoint
}

type ReasoningEffort string

const (
	ReasoningEffort_Low    ReasoningEffort = "low"
	ReasoningEffort_High   ReasoningEffort = "high"
	ReasoningEffort_Medium ReasoningEffort = "medium"
)

type ConversationParameters struct {
	Reasoning *ReasoningConfig
	tools     []Tool
}

func (e *Endpoint) Post(conversation []json.RawMessage, params *ConversationParameters) (*ConversationResponse, error) {

	apiUrl, err := url.Parse(e.rae.GetUrl())
	if err != nil {
		return nil, err
	}

	body, err := e.buildRequestBody(conversation, params)
	if err != nil {
		return nil, err
	}

	req := http.Request{
		Method: "POST",
		Header: e.rae.GetHttpHeader(),
		URL:    apiUrl,
		Body:   io.NopCloser(bytes.NewBuffer(body)),
	}

	resp, err := e.client.Do(&req)
	if err != nil {
		return nil, err
	}

	defer utils.CloseAndLogIfError_DEFER_ME(resp.Body)
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("error reading body: " + err.Error())
	}
	conversationResponse, err := utils.Unmarshall[ConversationResponse](rawBody)
	if err != nil {
		sb := string(rawBody)
		utils.PrintPrettyJsonFromString(&sb)
		panic(err)
	}

	return &conversationResponse, nil
}

func (e *Endpoint) buildRequestBody(conversation []json.RawMessage, params *ConversationParameters) ([]byte, error) {

	respBody := ResponsesApiRequest{
		Model: string(e.model),
		Input: conversation,
	}

	if params != nil {
		if params.Reasoning != nil {
			respBody.Reasoning = params.Reasoning
		}
		if len(params.tools) > 0 {
			respBody.Tools = params.tools
		}
	}

	marshaled, err := json.Marshal(respBody)
	if err != nil {
		return nil, err
	}

	return marshaled, nil
}
