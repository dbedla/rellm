package rellm_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//go:embed testdata/pro_api_req_hi.json
var goldenProReqHi string

//go:embed testdata/pro_api_resp_hi.json
var goldenProRespHi string

//go:embed testdata/pro_api_req_hi_reasoning.json
var goldenProReqAfterReasoning string

func TestAgentAskLikeAPro(t *testing.T) {
	agent, httpDo := buildTestProToolAgent(t, TestDefaultMaxToolsIterationWithoutReturnMessage)
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
	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqAfterReasoning, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespHi)),
		}, nil)

	promptFirst, err := rellm.NewPromptBuilder().
		WithMessage("Hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	respMsg, err := agent.Execute(promptFirst)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "Hello! How can I help you today?", respMsg, "response message should match")

	promptReasoning, err := rellm.NewPromptBuilder().
		WithMessage("What did you reason about?").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	_, err = agent.Execute(promptReasoning)
	assert.NoError(t, err, "failed to ask with reasoning in conversation")
}

//go:embed testdata/or_api_req_hi.json
var goldenOpenRouterReqHi string

//go:embed testdata/or_api_req_hi_no_reasoning.json
var goldenOpenRouterReqAfterReasoning string

//go:embed testdata/or_api_resp_hi.json
var goldenOpenRouterRespHi string

//go:embed testdata/or_api_resp_hi_2.json
var goldenOpenRouterRespHi2 string

func TestAgentAskLikeOpenRouterSkipsSignatureOnlyReasoningReplay(t *testing.T) {
	agent, httpDo := buildTestOpenRouterToolAgent(t, TestDefaultMaxToolsIterationWithoutReturnMessage)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenOpenRouterReqHi, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenOpenRouterRespHi)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenOpenRouterReqAfterReasoning, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenOpenRouterRespHi2)),
		}, nil)

	prompt, err := rellm.NewPromptBuilder().
		WithMessage("Hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err, "failed to build prompt")

	respMsg, err := agent.Execute(prompt)
	assert.NoError(t, err, "failed to ask")
	assert.Equal(t, "Hello! How can I help you today?", respMsg)

	secondPrompt, err := rellm.NewPromptBuilder().
		WithMessage("how are you").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err, "failed to build prompt")

	respMsg, err = agent.Execute(secondPrompt)
	assert.NoError(t, err, "failed to ask with reasoning in conversation")
	assert.Equal(t, "I am doing well.", respMsg)
}

//go:embed testdata/pro_api_tool_A1_req.json
var goldenProReqA1 string

//go:embed testdata/pro_api_tool_A1_resp.json
var goldenProRespA1 string

//go:embed testdata/pro_api_tool_A2_req.json
var goldenProReqA2 string

//go:embed testdata/pro_api_tool_A2_resp.json
var goldenProRespA2 string

//go:embed testdata/pro_api_tool_A3_req.json
var goldenProReqA3 string

//go:embed testdata/pro_api_tool_A3_resp.json
var goldenProRespA3 string

func TestAgentAskLikeAProToolsCall(t *testing.T) {
	agent, httpDo := buildTestProToolAgent(t, TestDefaultMaxToolsIterationWithoutReturnMessage)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqA1, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespA1)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqA2, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespA2)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqA3, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespA3)),
		}, nil)

	q := "call one tool, check output, then call second tool, check output, provide conclusion"
	prompt, err := rellm.NewPromptBuilder().
		WithMessage(q).
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	respMsg, err := agent.Execute(prompt)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "The first tool call to `GetStaticData` returned the value `42`. The second tool call to `GetDataFor` with the input \"the meaning of 42\" returned a list containing `[\"abc\", \"def\"]`. Therefore, based on these specific tool outputs, the data associated with the value 42 is \"abc\" and \"def\".", respMsg, "response message should match")
}

//go:embed testdata/pro_api_tool_B1_req_unknown_fn_call.json
var goldenProReqB1_UnknownFnCAll string

//go:embed testdata/pro_api_tool_B1_resp_unknown_fn_call.json
var goldenProRespB1_UnknownFnCAll string

//go:embed testdata/pro_api_tool_B2_req_unknown_fn_call.json
var goldenProReqB2_UnknownFnCAll string

//go:embed testdata/pro_api_tool_B2_resp_unknown_fn_call.json
var goldenProRespB2_UnknownFnCAll string

func TestAgentAskLikeAProToolsCall_UnknownFnCall(t *testing.T) {
	agent, httpDo := buildTestProToolAgent(t, TestDefaultMaxToolsIterationWithoutReturnMessage)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqB1_UnknownFnCAll, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespB1_UnknownFnCAll)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqB2_UnknownFnCAll, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespB2_UnknownFnCAll)),
		}, nil)

	q := "call function GetSpecialData"
	prompt, err := rellm.NewPromptBuilder().
		WithMessage(q).
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	respMsg, err := agent.Execute(prompt)
	assert.Equal(t, respMsg, "")
	assert.Error(t, err, "failed to ask")
	assert.ErrorIs(t, err, rellm.ErrUnknownToolCall)

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "", respMsg, "response message should match")

	//since we get err ErrUnknownToolCall so we simulate user ask to continue
	q = "continue"
	promptContinue, err := rellm.NewPromptBuilder().
		WithMessage(q).
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	respMsg, err = agent.Execute(promptContinue)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "It appears that the attempt to call `GetSpecialData` resulted in an error indicating the function was not found. \n\nHow would you like to proceed? I can try calling one of the other available functions, such as `GetStaticData` or `GetDataFor`, if you provide a specific input.", respMsg, "response message should match")
}

func TestTooManyFunctionCall(t *testing.T) {
	const NotEnoughToolLoopLimit = 2
	agent, httpDo := buildTestProToolAgent(t, NotEnoughToolLoopLimit)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqA1, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespA1)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenProReqA2, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenProRespA2)),
		}, nil)

	q := "call one tool, check output, then call second tool, check output, provide conclusion"
	prompt, err := rellm.NewPromptBuilder().
		WithMessage(q).
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	_, err = agent.Execute(prompt)
	assert.Error(t, err, "expected error when tool iteration budget is exceeded")
	assert.True(t, errors.Is(err, rellm.ErrMaxToolIterationsReached), "error should wrap ErrMaxToolIterationsReached")
}

func TestFuncResultToFunctionCallRespSerializesOutputAsString(t *testing.T) {
	resp := rellm.FuncResultToFunctionCallResp("call_123", int64(42))

	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal function call response")
	assert.JSONEq(t, `{"type":"function_call_output","call_id":"call_123","output":"42"}`, string(jsonResp))
}

func buildTestProToolAgent(t *testing.T, maxToolsIterationWithoutReturnMessage uint64) (*rellm.Agent, *HttpDoMock) {

	agentName := "TestProAgent"
	workspace := t.TempDir()
	mockHttp := new(HttpDoMock)

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithProvider(rellm.Provider_LMStudio).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithClientHttpDo(mockHttp).
		Build()
	assert.NoError(t, err, "failed to create endpoint")

	ta, err := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(maxToolsIterationWithoutReturnMessage).
		WithContinueConversation(false).
		WithSystemMessage("You are a helpful assistant.").
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithNoOpLogger().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func buildTestOpenRouterToolAgent(t *testing.T, maxToolsIterationWithoutReturnMessage uint64) (*rellm.Agent, *HttpDoMock) {
	agentName := "TestOpenRouterAgent"
	workspace := t.TempDir()
	mockHttp := new(HttpDoMock)

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint(testBaseUrl, testPort, testResponsesApiEndpoint, nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithProvider(rellm.Provider_OpenRouter).
		WithModel(rellm.Model_OpenRouter_Google_Gemini_3_1_Flash_Lite).
		WithClientHttpDo(mockHttp).
		Build()
	assert.NoError(t, err, "failed to create endpoint")

	ta, err := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(maxToolsIterationWithoutReturnMessage).
		WithContinueConversation(false).
		WithSystemMessage("You are a helpful assistant.").
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithNoOpLogger().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
