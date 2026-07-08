package rellm_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"rellm/pkg/rellm"

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
	assert.Contains(t, err.Error(), "error unmarshalling")
	assert.Contains(t, err.Error(), `{"id":`)
}

func assertHTTPStatusError(t *testing.T, err error, statusCode int, body string) {
	t.Helper()

	var statusErr *rellm.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, statusCode, statusErr.StatusCode)
	assert.Equal(t, body, statusErr.Body)
	assert.Equal(t, "https://api.example.test/responses", statusErr.URL)
	assert.Equal(t, "req_test", statusErr.RequestID)
	assert.NotContains(t, err.Error(), "secret-token")
}

func postWithValidResponse(t *testing.T) (*rellm.ResponsesApiResp, error) {
	t.Helper()

	req := &rellm.ResponsesApiReq{Model: "test-model", Input: []json.RawMessage{}}
	return newTestEndpoint(t, http.StatusOK, `{"id":"resp_test"}`).Post(req, tNopInspectReq, tNopInspectResp)
}

func postWithResponse(t *testing.T, statusCode int, body string) error {
	t.Helper()

	req := &rellm.ResponsesApiReq{Model: "test-model", Input: []json.RawMessage{}}
	_, err := newTestEndpoint(t, statusCode, body).Post(req, tNopInspectReq, tNopInspectResp)
	return err
}

func newTestEndpoint(t *testing.T, statusCode int, body string) *rellm.Endpoint {
	t.Helper()

	endpoint, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(newTestResponsesEndpoint()).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithClientHttpDo(testEndpointClient{statusCode: statusCode, body: body}).
		Build()
	require.NoError(t, err)
	return endpoint
}

type testResponsesEndpoint struct{}

func newTestResponsesEndpoint() testResponsesEndpoint {
	return testResponsesEndpoint{}
}

func (e testResponsesEndpoint) GetUrl() string {
	return "https://api.example.test/responses"
}

func (e testResponsesEndpoint) GetHttpHeader() http.Header {
	header := make(http.Header)
	header.Set("Authorization", "Bearer secret-token")
	return header
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

func tNopInspectReq(*rellm.ResponsesApiReq)   {}
func tNopInspectResp(*rellm.ResponsesApiResp) {}
