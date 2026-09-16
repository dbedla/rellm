package rellm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAgentAsk_ConversationBeforeAndAfter(t *testing.T) {
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

	ctx := context.Background()
	conversationBeforeAsk, err := agent.Conversation().Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversationBeforeAsk, 0)

	finalReport, err := agent.Ask(ctx, "Hi")
	assert.NoError(t, err)
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "Hello! How can I help you today? \n\nIf you have any questions about the weather, meteorology, climate patterns, or even how certain atmospheric phenomena work, feel free to ask!", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenRespHi), finalReport.StepsStats)

	conversationFromGolden, err := buildConversationFromGoldenLms(
		goldenReqHi,
		goldenRespHi)
	assert.NoError(t, err)

	conversationAfterAsk, err := agent.Conversation().Load(ctx)
	assert.NoError(t, err)

	assert.Equal(t, conversationFromGolden, conversationAfterAsk)
}

func TestAgentLMS_ToolsCallWithConversationCheck(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMS(t, testDefaultMaxAgentSteps)
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

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, prompt)
	assert.NoError(t, err, "failed to ask")

	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "The first tool call to `GetStaticData` returned the value `42`. The second tool call to `GetDataFor` with the input \"the meaning of 42\" returned a list containing `[\"abc\", \"def\"]`. Therefore, based on these specific tool outputs, the data associated with the value 42 is \"abc\" and \"def\".", finalReport.Message, "response message should match")
	assert.Equal(t, expectedStepStats(t, goldenProRespA1, goldenProRespA2, goldenProRespA3), finalReport.StepsStats)

	agentConversation, err := agent.Conversation().Load(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(agentConversation))

	conversationFromGolden, err := buildConversationFromGoldenLms(
		goldenProReqA3,
		goldenProRespA3)
	assert.NoError(t, err)

	assert.Equal(t, conversationFromGolden, agentConversation)
}

func TestAgentLMS_ToolsCallWithConversationCheck_SecondRespFail(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMS(t, testDefaultMaxAgentSteps)
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
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, nil)

	q := "call one tool, check output, then call second tool, check output, provide conclusion"
	prompt, err := rellm.NewPromptBuilder().
		WithMessage(q).
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, prompt)
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrEndpointNilBodyInResponse)

	assert.Empty(t, finalReport.Message)
	// Usage of the two successful calls is retained despite the failing third.
	assert.Equal(t, expectedStepStats(t, goldenProRespA1, goldenProRespA2), finalReport.StepsStats)

	agentConversation, err := agent.Conversation().Load(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 8, len(agentConversation))

	conversationFromGolden, err := buildConversationFromGoldenLms(
		goldenProReqA3)
	assert.NoError(t, err)

	assert.Equal(t, conversationFromGolden, agentConversation)
}

func TestAgentOpenRouterGemma_ConversationCheck(t *testing.T) {
	agent, httpDo := buildTestProToolAgentOpenRouter(t, "google/gemma-4-26b-a4b-it", testDefaultMaxAgentSteps)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenOR_Hi_01_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenOR_Hi_02_resp)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenOR_Hi_03_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenOR_Hi_04_resp)),
		}, nil)

	prompt, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, prompt)
	assert.NoError(t, err)
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "Hello! How can I help you today?", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenOR_Hi_02_resp), finalReport.StepsStats)

	secondPrompt, err := rellm.NewPromptBuilder().
		WithMessage("what tool do you see?").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	finalReport, err = agent.Execute(ctx, secondPrompt)
	assert.NoError(t, err)
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "I have access to the following tools:\n\n1.  **`GetDataFor`**: This tool allows me to retrieve specific data based on an input string you provide.\n2.  **`GetStaticData`**: This tool allows me to retrieve pre-defined static data.", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenOR_Hi_04_resp), finalReport.StepsStats)

	agentConversation, err := agent.Conversation().Load(ctx)
	assert.NoError(t, err)

	conversationFromGolden, err := buildConversationFromGoldenOpenRouter(
		goldenOR_Hi_03_req,
		goldenOR_Hi_04_resp,
	)
	assert.NoError(t, err)

	assert.Equal(t, conversationFromGolden, agentConversation)
}

func TestAgentOpenRouterGemini_ConversationCheck(t *testing.T) {
	agent, httpDo := buildTestProToolAgentOpenRouter(t, "google/gemini-3.1-flash-lite", testDefaultMaxAgentSteps)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenOR_Hi_gemini_01_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenOR_Hi_gemini_02_resp)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenOR_Hi_gemini_03_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenOR_Hi_gemini_04_resp)),
		}, nil)

	prompt, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, prompt)
	assert.NoError(t, err)
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "Hello! How can I help you today?", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenOR_Hi_gemini_02_resp), finalReport.StepsStats)

	secondPrompt, err := rellm.NewPromptBuilder().
		WithMessage("what tools do you see?").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	finalReport, err = agent.Execute(ctx, secondPrompt)
	assert.NoError(t, err, "failed to ask with reasoning in conversation")
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "I have access to the following tools:\n\n*   **`GetDataFor`**: This tool allows me to retrieve specific data based on an input you provide.\n*   **`GetStaticData`**: This tool allows me to retrieve general static information.\n\nHow can I help you use these today?", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenOR_Hi_gemini_04_resp), finalReport.StepsStats)

	agentConversation, err := agent.Conversation().Load(ctx)
	assert.NoError(t, err)

	conversationFromGolden, err := buildConversationFromGoldenOpenRouter(
		goldenOR_Hi_gemini_03_req,
		goldenOR_Hi_gemini_04_resp,
	)
	assert.NoError(t, err)

	assert.Equal(t, conversationFromGolden, agentConversation)
}

//go:embed testdata/openrouter/file_summary_01_luna_req.json
var goldenFS01LunaReq string

//go:embed testdata/openrouter/file_summary_02_luna_resp.json
var goldenFS02LunaResp string

//go:embed testdata/openrouter/file_summary_03_luna_req.json
var goldenFS03LunaReq string

//go:embed testdata/openrouter/file_summary_04_luna_resp.json
var goldenFS04LunaResp string

//go:embed testdata/openrouter/file_summary_05_luna_req.json
var goldenFS05LunaReq string

//go:embed testdata/openrouter/file_summary_06_luna_resp.json
var goldenFS06LunaResp string

func TestAgentAskGetSummaryFromOpenAiWithFsToolset(t *testing.T) {
	fsToolset := new(ToolsetMock)
	fsToolset.Test(t)
	defer fsToolset.AssertExpectations(t)

	fsToolset.On("Definitions").
		Return(examplesutils.NewFSToolset(nil).Definitions()).Once()
	fsToolset.On("Dispatch", mock.Anything, "FSToolset_GetReadOnlyPaths", json.RawMessage(`"{}"`)).
		Return(rellm.ToolCallResult{Value: []string{"/input"}}, nil).Once()
	fsToolset.On("Dispatch", mock.Anything, "FSToolset_GetOutputDir", json.RawMessage(`"{}"`)).
		Return(rellm.ToolCallResult{Value: "/output"}, nil).Once()
	fsToolset.On("Dispatch", mock.Anything, "FSToolset_ListFilesIn", json.RawMessage(`"{\"path\":\"/input\"}"`)).
		Return(rellm.ToolCallResult{Value: []string{"/input/locations.txt", "/input/names.txt"}}, nil).Once()
	fsToolset.On("Dispatch", mock.Anything, "FSToolset_ListFilesIn", json.RawMessage(`"{\"path\":\"/output\"}"`)).
		Return(rellm.ToolCallResult{Value: []string(nil)}, nil).Once()

	agent, httpDo := buildTestFileSystemAgentOpenRouter(t, rellm.Model("openai/gpt-5.6-luna"), testDefaultMaxAgentSteps, fsToolset)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenFS01LunaReq, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenFS02LunaResp)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenFS03LunaReq, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenFS04LunaResp)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenFS05LunaReq, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenFS06LunaResp)),
		}, nil)

	ctx := context.Background()
	finalReport, err := agent.Ask(ctx, "what files do you see")
	assert.NoError(t, err)
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "I can see these files:\n\n- `locations.txt`\n- `names.txt`\n\nThe output directory is currently empty.", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenFS02LunaResp, goldenFS04LunaResp, goldenFS06LunaResp), finalReport.StepsStats)
}

func buildTestFileSystemAgentOpenRouter(t *testing.T, model rellm.Model, maxAgentSteps uint64, fst rellm.Toolset) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestProAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewOpenRouterProviderWithHTTPClient("test-key", model, mockHttp)
	assert.NoError(t, err)

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversation(rellm.NewInMemoryConversation()).
		WithSystemMessage("You are a helpful assistant, with limited access to the file system.").
		WithToolset(fst, rellm.ParallelToolCallsDefaultForProvider).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
