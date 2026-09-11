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

	ctx := context.Background()
	finalReport, err := agent.Ask(ctx, "Hi")
	assert.NoError(t, err, "failed to ask")
	assert.NotEmpty(t, finalReport.Messages)
	assert.Equal(t, "Hello! How can I help you today? \n\nIf you have any questions about the weather, meteorology, climate patterns, or even how certain atmospheric phenomena work, feel free to ask!", finalReport.Messages, "response message should match")
	assert.Equal(t, expectedStepStats(t, goldenRespHi), finalReport.StepsStats)
}

//go:embed testdata/base_api_req_hi_no_sys_msg.json
var goldenReqHiNoSysMsg string

func TestAgentAskNoSysMsg(t *testing.T) {
	agent, httpDo := buildTestAgentNoSysMsg(t)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenReqHiNoSysMsg, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenRespHi)),
		}, nil)

	ctx := context.Background()
	finalReport, err := agent.Ask(ctx, "Hi")
	assert.NoError(t, err, "failed to ask")
	assert.NotEmpty(t, finalReport.Messages)
	assert.Equal(t, "Hello! How can I help you today? \n\nIf you have any questions about the weather, meteorology, climate patterns, or even how certain atmospheric phenomena work, feel free to ask!", finalReport.Messages, "response message should match")
	assert.Equal(t, expectedStepStats(t, goldenRespHi), finalReport.StepsStats)

	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 3)
}

func TestAgentAskContextAlreadyCanceled(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	finalReport, err := agent.Ask(ctx, "Hi")
	assert.Error(t, err)
	assert.ErrorIs(t, context.Canceled, err)
	assert.Empty(t, finalReport.Messages)
}

func TestAgentAsk_HTTP200EmptyBody(t *testing.T) {
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

	ctx := context.Background()
	finalReport, err := agent.Ask(ctx, "Hi")
	assert.Error(t, err)
	assert.Empty(t, finalReport.Messages)
}

func TestAgentAsk_EmptyString(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	ctx := context.Background()
	_, err := agent.Ask(ctx, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrEmptyPrompt)
}

func TestAgentExecute_NilPrompt(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	ctx := context.Background()
	_, err := agent.Execute(ctx, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrEmptyPrompt)
}

func TestPromptBuilder_AllFields(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			b, err := io.ReadAll(args.Get(0).(*http.Request).Body)
			assert.NoError(t, err)

			var req rellm.ResponsesAPIReq
			err = json.Unmarshal(b, &req)
			assert.NoError(t, err)

			assert.Equal(t, float64(0.7), *req.Temperature)
			assert.Equal(t, rellm.ReasoningEffortHigh, req.Reasoning.Effort)
			assert.NotNil(t, req.Text)
			assert.Equal(t, "json_schema", req.Text.Format.Type)
			assert.Equal(t, "person", req.Text.Format.Name)
			assert.True(t, req.Text.Format.Strict)
			assert.Equal(t, 512, req.MaxOutputTokens)
			assert.InDelta(t, 0.9, *req.TopP, 0.001)
			assert.InDelta(t, 1.0, *req.PresencePenalty, 0.001)
			assert.InDelta(t, 0.5, *req.FrequencyPenalty, 0.001)
			assert.Equal(t, 3, req.TopLogprobs)
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenRespHi)),
		}, nil)

	prompt, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithTemperature(0.7).
		WithReasoning(rellm.ReasoningEffortHigh).
		WithTextFormat(&rellm.TextFormat{
			Type:   "json_schema",
			Name:   "person",
			Strict: true,
			Schema: map[string]interface{}{"type": "object"},
		}).
		WithMaxOutputTokens(512).
		WithTopP(0.9).
		WithPresencePenalty(1.0).
		WithFrequencyPenalty(0.5).
		WithTopLogprobs(3).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	report, err := agent.Execute(ctx, prompt)
	assert.NoError(t, err)
	assert.Equal(t, expectedStepStats(t, goldenRespHi), report.StepsStats)
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

	ctx := context.Background()
	finalReport, err := agent.Ask(ctx, "Hi")
	assert.Error(t, err)
	assert.Empty(t, finalReport.Messages)
}

func TestAgentAsk_RetryAfterProviderFailedMessageStaysInConversation(t *testing.T) {
	agent, httpDo := buildTestAgent(t)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(goldenRespHi)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenRespHi)),
		}, nil)

	ctx := context.Background()
	_, err := agent.Ask(ctx, "Hi")
	assert.Error(t, err)

	report, err := agent.Ask(ctx, "Hi")
	assert.NoError(t, err)
	assert.Equal(t, expectedStepStats(t, goldenRespHi), report.StepsStats)

	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 5)
}

func buildTestAgent(t *testing.T) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewLMStudioProviderWithHTTPClient("google/gemma-4-26b-a4b", testBaseUrl, testPort, mockHttp)
	assert.NoError(t, err, "failed to create provider")

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage("You are a helpful assistant with deep weather knowledge.").
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func buildTestAgentNoSysMsg(t *testing.T) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewLMStudioProviderWithHTTPClient("google/gemma-4-26b-a4b", testBaseUrl, testPort, mockHttp)
	assert.NoError(t, err, "failed to create provider")

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}
