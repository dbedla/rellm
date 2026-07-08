package rellm_test

import (
	"bytes"
	"io"
	"net/http"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestAgentAskBodyIsNil(t *testing.T) {
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
		}, nil)

	respMsg, err := agent.Ask("Hi")
	assert.Error(t, err)
	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "", respMsg, "response message should match")
}

func TestAgentAskStatusInternalServerError(t *testing.T) {
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
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(goldenRespHi)),
		}, nil)

	respMsg, err := agent.Ask("Hi")
	assert.Error(t, err)
	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "", respMsg, "response message should match")
}

func buildTestAgent(t *testing.T) (*rellm.Agent, *HttpDoMock) {

	agentName := "TestAgent"
	workspace := t.TempDir()
	mockHttp := new(HttpDoMock)

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
		WithNoOpLogger().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
