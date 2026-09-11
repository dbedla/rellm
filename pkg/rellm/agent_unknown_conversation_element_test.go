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

//go:embed testdata/lms/hi_02_gemma_resp_unknown_el.json
var goldenLMS_Hi_02_resp_unknown string

func TestLMSAgentHiUnknownConversationElInRespDrop(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMSWithUnknownElHandler(t, testDefaultMaxAgentSteps, unknownConversationElementHandlerDrop)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_Hi_01_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMS_Hi_02_resp_unknown)),
		}, nil)

	promptFirst, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, promptFirst)
	assert.NoError(t, err)

	assert.NotEmpty(t, finalReport.Messages)
	assert.Equal(t, "Hello! How can I help you today?", finalReport.Messages)
	assert.Equal(t, expectedStepStats(t, goldenLMS_Hi_02_resp_unknown), finalReport.StepsStats)
	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 4)
}

func TestLMSAgentHiUnknownConversationElInRespKeepInTheLoop(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMSWithUnknownElHandler(t, testDefaultMaxAgentSteps, unknownConversationElementHandlerKeepInTheLoop)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_Hi_01_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMS_Hi_02_resp_unknown)),
		}, nil)

	promptFirst, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, promptFirst)
	assert.NoError(t, err)

	assert.NotEmpty(t, finalReport.Messages)
	assert.Equal(t, "Hello! How can I help you today?", finalReport.Messages)
	assert.Equal(t, expectedStepStats(t, goldenLMS_Hi_02_resp_unknown), finalReport.StepsStats)
	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 5)
}

func TestLMSAgentHiUnknownConversationElHandlerAddAdditionalData(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMSWithUnknownElHandler(t, testDefaultMaxAgentSteps, unknownConversationElementHandlerAddResp)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_Hi_01_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMS_Hi_02_resp_unknown)),
		}, nil)

	promptFirst, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, promptFirst)
	assert.NoError(t, err)

	assert.NotEmpty(t, finalReport.Messages)
	assert.Equal(t, "Hello! How can I help you today?", finalReport.Messages)
	assert.Equal(t, expectedStepStats(t, goldenLMS_Hi_02_resp_unknown), finalReport.StepsStats)
	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 6)
}

//go:embed testdata/lms/hi_02_gemma_resp_only_unknown_el.json
var goldenLMS_Hi_02_resp_only_unknown string

func TestLMSAgentHiUnknownConversationElInRespDropNoNewMessages(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMSWithUnknownElHandler(t, testDefaultMaxAgentSteps, unknownConversationElementHandlerDrop)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_Hi_01_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMS_Hi_02_resp_only_unknown)),
		}, nil)

	promptFirst, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, promptFirst)
	assert.Error(t, err)
	assert.ErrorIs(t, err, rellm.ErrNoNewConversationElementAfterDispatch)

	assert.Empty(t, finalReport.Messages)
	// Usage is recorded even though dispatching produced no conversation elements.
	assert.Equal(t, expectedStepStats(t, goldenLMS_Hi_02_resp_only_unknown), finalReport.StepsStats)
	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 2)
}

func TestLMSAgentHiUnknownConversationElInRespReplace(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMSWithUnknownElHandler(t, testDefaultMaxAgentSteps, unknownConversationElementHandlerReplace)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_Hi_01_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMS_Hi_02_resp_unknown)),
		}, nil)

	promptFirst, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	finalReport, err := agent.Execute(ctx, promptFirst)
	assert.NoError(t, err)

	assert.NotEmpty(t, finalReport.Messages)
	assert.Equal(t, "Hello! How can I help you today?", finalReport.Messages)
	assert.Equal(t, expectedStepStats(t, goldenLMS_Hi_02_resp_unknown), finalReport.StepsStats)
	conversation, err := agent.CurrentConversation(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 5)

	found := false
	for _, el := range conversation {
		if um, ok := el.(*rellm.UserMessage); ok && rellm.TextFromContent(um.Content) == "handled" {
			found = true
		}
	}
	assert.True(t, found, "replacement user message should be persisted")
}

type unknownConversationElementHandlerTestType int

const (
	unknownConversationElementHandlerDrop          unknownConversationElementHandlerTestType = 1
	unknownConversationElementHandlerKeepInTheLoop unknownConversationElementHandlerTestType = 3
	unknownConversationElementHandlerAddResp       unknownConversationElementHandlerTestType = 4
	unknownConversationElementHandlerReplace       unknownConversationElementHandlerTestType = 5
)

func buildTestProToolAgentLMSWithUnknownElHandler(
	t *testing.T, maxAgentSteps uint64,
	uh unknownConversationElementHandlerTestType) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestProAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewLMStudioProviderWithHTTPClient("google/gemma-4-26b-a4b", testBaseUrl, testPort, mockHttp)
	assert.NoError(t, err, "failed to create provider")

	builder := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage("You are a helpful assistant.").
		WithImageGenerationKeepInTheLoop().
		WithToolset(&examplesutils.DataSrcToolset{})

	switch uh {
	case unknownConversationElementHandlerKeepInTheLoop:
		builder.WithUnknownConversationElementKeepInTheLoop()
	case unknownConversationElementHandlerDrop:
		builder.WithUnknownConversationElementDrop()
	case unknownConversationElementHandlerAddResp:
		builder.WithUnknownConversationElementHandler(twoUnknownElements)
	case unknownConversationElementHandlerReplace:
		builder.WithUnknownConversationElementHandler(replaceWithUserMessage)
	}

	ta, err := builder.Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func twoUnknownElements(ctx context.Context, el *rellm.UnknownElement) ([]rellm.ConversationElement, error) {
	extraData := rellm.UnknownElement{
		Provider: "lmstudio",
		Type:     "testType",
		Role:     "testRole",
		Raw:      json.RawMessage(`{"content":{"info":"testInfo"}}`),
	}
	return []rellm.ConversationElement{el, &extraData}, nil
}

func replaceWithUserMessage(_ context.Context, _ *rellm.UnknownElement) ([]rellm.ConversationElement, error) {
	return []rellm.ConversationElement{
		&rellm.UserMessage{MessageContent: rellm.MessageContent{
			Role:    "user",
			Content: []rellm.MessagePart{{Type: "input_text", Text: "handled"}},
		}},
	}, nil
}
