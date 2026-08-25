package rellm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"rellm/pkg/rellm"
	"strings"
	"testing"

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

	conversationBeforeAskRaw, err := agent.CurrentConversation()
	assert.NoError(t, err)
	assert.Len(t, conversationBeforeAskRaw, 1)

	conversationBeforeAsk, err := json.Marshal(conversationBeforeAskRaw)
	assert.NoError(t, err)

	assert.JSONEq(t, `[{"role":"system","content":[{"type":"input_text","text":"You are a helpful assistant with deep weather knowledge."}]}]`, string(conversationBeforeAsk))

	ctx := context.Background()
	respMsg, err := agent.Ask(ctx, "Hi")
	assert.NoError(t, err)
	assert.NotNil(t, respMsg)
	assert.Equal(t, "Hello! How can I help you today? \n\nIf you have any questions about the weather, meteorology, climate patterns, or even how certain atmospheric phenomena work, feel free to ask!", respMsg)

	conversationFromGolden, err := buildConversationFromGoldenLms(
		goldenReqHi,
		goldenRespHi)
	assert.NoError(t, err)

	expectedAfterAsk, err := json.Marshal(conversationFromGolden)
	assert.NoError(t, err)

	conversationAfterAskRaw, err := agent.CurrentConversation()
	assert.NoError(t, err)

	conversationAfterAsk, err := json.Marshal(conversationAfterAskRaw)
	assert.NoError(t, err)

	assert.JSONEq(t, string(expectedAfterAsk), string(conversationAfterAsk))
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
	respMsg, err := agent.Execute(ctx, prompt)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "The first tool call to `GetStaticData` returned the value `42`. The second tool call to `GetDataFor` with the input \"the meaning of 42\" returned a list containing `[\"abc\", \"def\"]`. Therefore, based on these specific tool outputs, the data associated with the value 42 is \"abc\" and \"def\".", respMsg, "response message should match")

	agentConversation, err := agent.CurrentConversation()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(agentConversation))

	conversationFromGolden, err := buildConversationFromGoldenLms(
		goldenProReqA3,
		goldenProRespA3)
	assert.NoError(t, err)

	expected, err := json.Marshal(conversationFromGolden)
	assert.NoError(t, err)

	actual, err := json.Marshal(agentConversation)
	assert.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
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
	respMsg, err := agent.Execute(ctx, prompt)
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrEndpointNilBodyInResponse)

	assert.Empty(t, respMsg)

	agentConversation, err := agent.CurrentConversation()
	assert.NoError(t, err)
	assert.Equal(t, 8, len(agentConversation))

	conversationFromGolden, err := buildConversationFromGoldenLms(
		goldenProReqA3)
	assert.NoError(t, err)

	expected, err := json.Marshal(conversationFromGolden)
	assert.NoError(t, err)

	actual, err := json.Marshal(agentConversation)
	assert.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
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
	respMsg, err := agent.Execute(ctx, prompt)
	assert.NoError(t, err)
	assert.Equal(t, "Hello! How can I help you today?", respMsg)

	secondPrompt, err := rellm.NewPromptBuilder().
		WithMessage("what tool do you see?").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	respMsg, err = agent.Execute(ctx, secondPrompt)
	assert.NoError(t, err)
	assert.Equal(t, "I have access to the following tools:\n\n1.  **`GetDataFor`**: This tool allows me to retrieve specific data based on an input string you provide.\n2.  **`GetStaticData`**: This tool allows me to retrieve pre-defined static data.", respMsg)

	agentConversation, err := agent.CurrentConversation()
	assert.NoError(t, err)

	conversationFromGolden, err := buildConversationFromGoldenOpenRouter(
		goldenOR_Hi_03_req,
		goldenOR_Hi_04_resp,
	)
	assert.NoError(t, err)

	expected, err := json.Marshal(conversationFromGolden)
	assert.NoError(t, err)

	actual, err := json.Marshal(agentConversation)
	assert.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
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
	respMsg, err := agent.Execute(ctx, prompt)
	assert.NoError(t, err)
	assert.Equal(t, "Hello! How can I help you today?", respMsg)

	secondPrompt, err := rellm.NewPromptBuilder().
		WithMessage("what tools do you see?").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	respMsg, err = agent.Execute(ctx, secondPrompt)
	assert.NoError(t, err, "failed to ask with reasoning in conversation")
	assert.Equal(t, "I have access to the following tools:\n\n*   **`GetDataFor`**: This tool allows me to retrieve specific data based on an input you provide.\n*   **`GetStaticData`**: This tool allows me to retrieve general static information.\n\nHow can I help you use these today?", respMsg)

	agentConversation, err := agent.CurrentConversation()
	assert.NoError(t, err)

	conversationFromGolden, err := buildConversationFromGoldenOpenRouter(
		goldenOR_Hi_gemini_03_req,
		goldenOR_Hi_gemini_04_resp,
	)
	assert.NoError(t, err)

	expected, err := json.Marshal(conversationFromGolden)
	assert.NoError(t, err)

	actual, err := json.Marshal(agentConversation)
	assert.NoError(t, err)

	assert.JSONEq(t, string(expected), string(actual))
}
