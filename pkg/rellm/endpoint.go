package rellm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Model
// value for models can be found:
//   - For openrouter: https://openrouter.ai/models (curl --request GET --url 'https://openrouter.ai/api/v1/models?limit=10' | jq)
//   - For lmstudio: https://lmstudio.ai/models
//
// names used by openrouter and lmstudio are not interchangeable:
//   - lms: "google/gemma-4-26b-a4b"
//   - openrouter: "google/gemma-4-26b-a4b-it"
type Model string

type ClientHttpDo interface {
	Do(request *http.Request) (*http.Response, error)
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
	ReasoningEffort_None   ReasoningEffort = "none"
	ReasoningEffort_Low    ReasoningEffort = "low"
	ReasoningEffort_High   ReasoningEffort = "high"
	ReasoningEffort_Medium ReasoningEffort = "medium"
	ReasoningEffort_XHigh  ReasoningEffort = "xhigh"
)

func parseResponsesApiResponse(rawBody []byte, inspectResp InspectEachResponse) (*ResponsesApiResp, error) {

	conversationResponse, err := unmarshal[ResponsesApiResp](rawBody)
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

func unmarshal[K any](rawBody []byte) (K, error) {
	var data K
	err := json.Unmarshal(rawBody, &data)
	if err != nil {
		return *new(K), fmt.Errorf("error unmarshaling: %s; unmarshaling type %T; raw: %s", err, data, string(rawBody))
	}

	return data, nil
}

func closeWithError(err *error, c io.Closer) {
	*err = errors.Join(*err, c.Close())
}
