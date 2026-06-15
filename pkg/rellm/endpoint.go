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

func (e *Endpoint) Post(conversation []json.RawMessage, proParameterSet FuncLikeProSet, toolset Toolset) (*ConversationResponse, error) {

	apiUrl, err := url.Parse(e.rae.GetUrl())
	if err != nil {
		return nil, err
	}

	body, err := e.buildRequestBody(conversation, proParameterSet, toolset)
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

func (e *Endpoint) buildRequestBody(conversation []json.RawMessage, proParameterSet FuncLikeProSet, toolset Toolset) ([]byte, error) {

	respBody := ResponsesApiRequest{
		Model: string(e.model),
		Input: conversation,
	}

	if toolset != nil {
		respBody.Tools = toolset.BuildTools()
	}

	proParameterSet(&respBody)

	marshaled, err := json.Marshal(respBody)
	if err != nil {
		return nil, err
	}

	return marshaled, nil
}
