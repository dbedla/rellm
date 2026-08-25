package rellm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpointPostHTTPStatusHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{name: "400 JSON error", statusCode: http.StatusBadRequest, body: `{"error":{"message":"bad request"}}`},
		{name: "401 unauthorized", statusCode: http.StatusUnauthorized, body: "unauthorized"},
		{name: "429 rate limit", statusCode: http.StatusTooManyRequests, body: "rate limited"},
		{name: "500 server error", statusCode: http.StatusInternalServerError, body: "server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postWithResponse(t, tt.statusCode, tt.body)
			assertHTTPStatusError(t, err, tt.statusCode, tt.body)
		})
	}
}

func TestEndpointPostOKValidResponse(t *testing.T) {
	resp, err := postWithValidResponse(t)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "resp_test", resp.Id)
}

func TestEndpointPostMalformedJSONResponse(t *testing.T) {
	err := postWithResponse(t, http.StatusOK, `{"id":`)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshaling")
	assert.Contains(t, err.Error(), `{"id":`)
}

func assertHTTPStatusError(t *testing.T, err error, statusCode int, body string) {
	t.Helper()

	var statusErr *HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, statusCode, statusErr.StatusCode)
	assert.Equal(t, body, statusErr.Body)
	assert.Equal(t, testURL, statusErr.URL)
	assert.Equal(t, "req_test", statusErr.RequestID)
	assert.NotContains(t, err.Error(), "test-key")
}

const testURL = "https://api.example.test/responses"

func postWithValidResponse(t *testing.T) (*ResponsesApiResp, error) {
	t.Helper()

	ctx := context.Background()
	req := &ResponsesApiReq{Model: "test-model", Input: []json.RawMessage{}}
	return newTestAgent(t, http.StatusOK, `{"id":"resp_test"}`).post(ctx, req)
}

func postWithResponse(t *testing.T, statusCode int, body string) error {
	t.Helper()

	ctx := context.Background()
	req := &ResponsesApiReq{Model: "test-model", Input: []json.RawMessage{}}
	_, err := newTestAgent(t, statusCode, body).post(ctx, req)
	return err
}

func newTestAgent(t *testing.T, statusCode int, body string) *Agent {
	t.Helper()

	p, err := NewOpenRouterProvider("test-key", "google/gemma-4-26b-a4b")
	require.NoError(t, err)
	p.WithURL(testURL).WithHTTPClient(testEndpointClient{statusCode: statusCode, body: body})

	return &Agent{provider: p}
}

type testEndpointClient struct {
	statusCode int
	body       string
}

func (c testEndpointClient) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") == "" {
		return nil, errors.New("missing authorization header")
	}

	header := make(http.Header)
	header.Set("X-Request-ID", "req_test")
	return &http.Response{StatusCode: c.statusCode, Header: header, Body: io.NopCloser(strings.NewReader(c.body))}, nil
}
