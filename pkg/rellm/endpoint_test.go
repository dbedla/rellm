package rellm

import (
	"context"
	"encoding/json"
	"net/http"
	"rellm/internal/examplesutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

//func newTestAgent(t *testing.T, statusCode int, body string) (*Agent, *HttpDoMock) {
//	t.Helper()
//
//
//	httpDoMock := HttpDoMock{}
//	p, err := NewOpenRouterProviderWithHTTPClient("test-key", "google/gemma-4-26b-a4b", &httpDoMock)
//	assert.NoError(t, err)
//
//	return , &httpDoMock
//}

type HttpDoMock struct {
	mock.Mock
}

func (h *HttpDoMock) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func buildTestProToolAgentOpenRouter(t *testing.T, model Model, maxAgentSteps uint64) (*Agent, *HttpDoMock) {

	agentName := "TestProAgent"
	mockHttp := new(HttpDoMock)

	p, err := NewOpenRouterProviderWithHTTPClient("test-key", model, mockHttp)
	assert.NoError(t, err)

	ta, err := NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(NewInMemoryStorage()).
		WithSystemMessage("You are a helpful assistant.").
		WithToolset(&examplesutils.DataSrcToolset{}).
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
