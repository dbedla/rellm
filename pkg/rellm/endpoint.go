package rellm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

type HTTPStatusError struct {
	StatusCode int
	Body       string
	URL        string
	RequestID  string
}

func (e *HTTPStatusError) Error() string {
	parts := []string{fmt.Sprintf("api request failed with status %d", e.StatusCode), "url: " + e.URL}
	if e.RequestID != "" {
		parts = append(parts, "request_id: "+e.RequestID)
	}
	if e.Body != "" {
		parts = append(parts, "body: "+e.Body)
	}
	return strings.Join(parts, "; ")
}

type ReasoningEffort string

const (
	ReasoningEffort_Low    ReasoningEffort = "low"
	ReasoningEffort_High   ReasoningEffort = "high"
	ReasoningEffort_Medium ReasoningEffort = "medium"
)

// Post sends a fully-built ResponsesApiReq over HTTP. The caller is responsible
// for constructing the request (model, input, inference params). The endpoint
// marshals it to JSON and handles transport.
func (e *Endpoint) Post(req *ResponsesApiReq, inspectReq InspectEachRequest, inspectResp InspectEachResponse) (*ResponsesApiResp, error) {
	apiUrl, err := url.Parse(e.rae.GetUrl())
	if err != nil {
		return nil, err
	}

	if inspectReq != nil {
		inspectReq(req)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq := &http.Request{
		Method: "POST",
		Header: e.rae.GetHttpHeader(),
		URL:    apiUrl,
		Body:   io.NopCloser(bytes.NewReader(body)),
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("nil response")
	}

	if resp.Body == nil {
		return nil, errors.New("empty response body")
	}
	defer closeAndLogIfError_DEFER_ME(resp.Body)
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseResponsesApiResponse(resp, rawBody, apiUrl.String(), inspectResp)
}

func parseResponsesApiResponse(resp *http.Response, rawBody []byte, apiURL string, inspectResp InspectEachResponse) (*ResponsesApiResp, error) {
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newHTTPStatusError(resp, rawBody, apiURL)
	}

	conversationResponse, err := unmarshall[ResponsesApiResp](rawBody)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("%s", string(rawBody)))
	}

	if inspectResp != nil {
		inspectResp(&conversationResponse)
	}

	return &conversationResponse, nil
}

func newHTTPStatusError(resp *http.Response, rawBody []byte, apiURL string) *HTTPStatusError {
	return &HTTPStatusError{
		StatusCode: resp.StatusCode,
		Body:       bodySnippet(rawBody),
		URL:        apiURL,
		RequestID:  responseRequestID(resp.Header),
	}
}

func responseRequestID(header http.Header) string {
	return header.Get("X-Request-ID")
}

func bodySnippet(rawBody []byte) string {
	const maxBodySnippetLength = 1024
	body := string(rawBody)
	if len(body) <= maxBodySnippetLength {
		return body
	}
	return body[:maxBodySnippetLength] + "..."
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
