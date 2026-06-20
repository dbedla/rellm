package rellm_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"rellm/pkg/agentsutills"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testBaseUrl              = "http://127.0.0.1"
	testPort                 = "1234"
	testResponsesApiEndpoint = "/v1/responses"
	testReasoningEffort      = "low"
	testTemperature          = 0.5
)

//go:embed testdata/base_api_req_hi.json
var goldenReqHi string

//go:embed testdata/base_api_resp_hi.json
var goldenRespHi string

func TestAgentAsk(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenReqHi, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenRespHi)),
		}, nil)

	respMsg, err := agent.Ask("Hi")
	assert.NoError(t, err, "failed to ask")
	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "Hello! How can I help you today? \n\nIf you have any questions about the weather, meteorology, climate patterns, or even how certain atmospheric phenomena work, feel free to ask!", respMsg, "response message should match")
}

//go:embed testdata/pro_api_req_hi.json
var goldenProReqHi string

//go:embed testdata/pro_api_resp_hi.json
var goldenProRespHi string

func TestAgentAskLikeAPro(t *testing.T) {
	agent, httpDo := buildTestProToolAgent(t)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqHi, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespHi)),
		}, nil)

	respMsg, resp, err := agent.AskLikeAPro("Hi", setTestReasoningAndTemperature)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "Hello! How can I help you today?", respMsg, "response message should match")

	assert.NotNil(t, resp, "response should not be nil")

	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal response")
	assert.JSONEq(t, goldenProRespHi, string(jsonResp))

}

func buildTestAgent(t *testing.T) (*rellm.Agent, *HttpDo) {

	agentName := "TestAgent"
	workspace := inMemoryWorkspace(t)
	mockHttp := new(HttpDo)

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithClientHttpDo(mockHttp).
		Build()
	assert.NoError(t, err, "failed to create endpoint")

	ta, err := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(false).
		WithSystemMessage("You are a helpful assistant with deep weather knowledge.").
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func buildTestProToolAgent(t *testing.T) (*rellm.Agent, *HttpDo) {

	agentName := "TestProAgent"
	workspace := inMemoryWorkspace(t)
	mockHttp := new(HttpDo)

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithClientHttpDo(mockHttp).
		Build()
	assert.NoError(t, err, "failed to create endpoint")

	ta, err := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(false).
		WithSystemMessage("You are a helpful assistant.").
		WithToolset(&agentsutills.DataSrcToolset{}).
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func inMemoryWorkspace(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "rellm-test-*")
	assert.NoError(t, err, "failed to create temp dir")
	return tempDir
}

func baseRequestMatch(req *http.Request) bool {
	return req.URL.String() == testBaseUrl+":"+testPort+testResponsesApiEndpoint &&
		req.Method == "POST"
}

func setTestReasoningAndTemperature(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: testReasoningEffort}
	req.Temperature = testTemperature
}

type HttpDo struct {
	mock.Mock
}

func (h *HttpDo) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
