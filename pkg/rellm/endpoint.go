package rellm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ResponsesApiEndpoint interface {
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
	rae    ResponsesApiEndpoint
}

type ReasoningEffort string

const (
	ReasoningEffort_Low    ReasoningEffort = "low"
	ReasoningEffort_High   ReasoningEffort = "high"
	ReasoningEffort_Medium ReasoningEffort = "medium"
)

func (e *Endpoint) Post(conversation []json.RawMessage, proParameterSet FuncLikeProSet, toolset Toolset) (*ResponsesApiResp, error) {

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

	defer closeAndLogIfError_DEFER_ME(resp.Body)
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	conversationResponse, err := unmarshall[ResponsesApiResp](rawBody)
	if err != nil {
		sb := string(rawBody)
		return nil, errors.Join(err, fmt.Errorf("%s", sb))
	}

	return &conversationResponse, nil
}

func (e *Endpoint) buildRequestBody(conversation []json.RawMessage, proParameterSet FuncLikeProSet, toolset Toolset) ([]byte, error) {

	respBody := ResponsesApiReq{
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

type UniversalResponsesEndpoint struct {
	baseUrl              string
	port                 string
	responsesApiEndpoint string
	httpHeader           http.Header
}

func NewUniversalResponsesEndpoint(baseUrl, port, responsesApiEndpoint string, httpHeader http.Header) *UniversalResponsesEndpoint {
	return &UniversalResponsesEndpoint{
		baseUrl:              baseUrl,
		port:                 port,
		responsesApiEndpoint: responsesApiEndpoint,
		httpHeader:           httpHeader,
	}
}

func (e *UniversalResponsesEndpoint) GetUrl() string {
	port := ""
	if e.port != "" {
		port = ":" + e.port
	}

	return e.baseUrl + port + e.responsesApiEndpoint
}

func (e *UniversalResponsesEndpoint) GetHttpHeader() http.Header {
	if e.httpHeader != nil {
		return e.httpHeader
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return header
}

func unmarshall[K any](rawBody []byte) (K, error) {
	var data K
	err := json.Unmarshal(rawBody, &data)
	if err != nil {
		return *new(K), fmt.Errorf("error unmarshalling: %s; unmarshaling type %T; raw: %s", err, data, string(rawBody))
	}

	return data, nil
}

func closeAndLogIfError_DEFER_ME(closeMe io.Closer) {
	err := closeMe.Close()
	if err != nil {
		fmt.Printf("error closing: %s\n", err)
	}
}
