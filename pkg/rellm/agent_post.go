package rellm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (a *Agent) post(ctx context.Context, req *ResponsesAPIReq) (*ResponsesAPIResp, error) {
	if a.inspectReq != nil {
		a.inspectReq(req)
	}

	resp, err := a.provider.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	if a.inspectResp != nil {
		a.inspectResp(resp)
	}
	return resp, nil
}

// postResponsesAPI is the shared transport for every provider's Send:
// marshal -> POST -> read body -> status check -> unmarshal.
func postResponsesAPI(ctx context.Context, client HTTPClient, url string,
	header http.Header, req *ResponsesAPIReq,
) (_ *ResponsesAPIResp, finalErr error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header = header

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, ErrEndpointNilResponse
	}

	if resp.Body == nil {
		return nil, errors.Join(ErrEndpointNilBodyInResponse, fmt.Errorf("response status: %s", resp.Status))
	}

	defer closeWithError(&finalErr, resp.Body)
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(err, ErrUnableToReadResponseBody, fmt.Errorf("response status: %s", resp.Status))
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newHTTPStatusError(resp, rawBody, httpReq.URL.String())
	}

	return parseResponsesAPIResponse(rawBody)
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
func newHTTPStatusError(resp *http.Response, rawBody []byte, apiURL string) *HTTPStatusError {
	return &HTTPStatusError{
		StatusCode: resp.StatusCode,
		Body:       bodySnippet(rawBody),
		URL:        apiURL,
		RequestID:  responseRequestID(resp.Header),
	}
}

func bodySnippet(rawBody []byte) string {
	const maxBodySnippetLength = 1024
	body := string(rawBody)
	if len(body) <= maxBodySnippetLength {
		return body
	}
	return body[:maxBodySnippetLength] + "..."
}

func responseRequestID(header http.Header) string {
	return header.Get("X-Request-ID")
}

func parseResponsesAPIResponse(rawBody []byte) (*ResponsesAPIResp, error) {
	conversationResponse, err := unmarshal[ResponsesAPIResp](rawBody)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("%s", string(rawBody)))
	}

	return &conversationResponse, nil
}

func unmarshal[K any](rawBody []byte) (K, error) {
	var data K
	err := json.Unmarshal(rawBody, &data)
	if err != nil {
		return *new(K), fmt.Errorf("error unmarshaling %T: %w (raw body: %s)", data, err, string(rawBody))
	}

	return data, nil
}

func closeWithError(err *error, c io.Closer) {
	if err == nil {
		return
	}
	if c == nil {
		return
	}
	*err = errors.Join(*err, c.Close())
}
