package rellm_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//go:embed testdata/lms/hi_01_gemma_req.json
var goldenLMS_Hi_01_req string

//go:embed testdata/lms/hi_02_gemma_resp.json
var goldenLMS_Hi_02_resp string

//go:embed testdata/lms/hi_03_gemma_req.json
var goldenLMS_Hi_03_req string

//go:embed testdata/lms/hi_04_gemma_resp.json
var goldenLMS_Hi_04_resp string

func TestLMSAgentHiWithToolsNoCall(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMS(t, testDefaultMaxAgentSteps)
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
			Body:       io.NopCloser(strings.NewReader(goldenLMS_Hi_02_resp)),
		}, nil)
	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_Hi_03_req, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMS_Hi_04_resp)),
		}, nil)

	promptFirst, err := rellm.NewPromptBuilder().
		WithMessage("hi").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	respMsg, err := agent.Execute(ctx, promptFirst)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "Hello! How can I help you today?", respMsg)

	promptReasoning, err := rellm.NewPromptBuilder().
		WithMessage("what tools do you see?").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	respMsg, err = agent.Execute(ctx, promptReasoning)
	assert.NoError(t, err, "failed to ask with reasoning in conversation")
	assert.Equal(t, "I have access to the following tools:\n\n1.  **`GetDataFor`**: This tool allows me to retrieve specific data based on an input string you provide.\n2.  **`GetStaticData`**: This tool allows me to retrieve predefined static information.", respMsg)
}

//go:embed testdata/openrouter/hi_01_gemma_req.json
var goldenOR_Hi_01_req string

//go:embed testdata/openrouter/hi_02_gemma_resp.json
var goldenOR_Hi_02_resp string

//go:embed testdata/openrouter/hi_03_gemma_req.json
var goldenOR_Hi_03_req string

//go:embed testdata/openrouter/hi_04_gemma_resp.json
var goldenOR_Hi_04_resp string

func TestOpenRouterAgentHiWithToolsNoCallGemma(t *testing.T) {
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
	assert.NoError(t, err, "failed to ask with reasoning in conversation")
	assert.Equal(t, "I have access to the following tools:\n\n1.  **`GetDataFor`**: This tool allows me to retrieve specific data based on an input string you provide.\n2.  **`GetStaticData`**: This tool allows me to retrieve pre-defined static data.", respMsg)
}

//go:embed testdata/openrouter/hi_gemini_01_req.json
var goldenOR_Hi_gemini_01_req string

//go:embed testdata/openrouter/hi_gemini_02_resp.json
var goldenOR_Hi_gemini_02_resp string

//go:embed testdata/openrouter/hi_gemini_03_req.json
var goldenOR_Hi_gemini_03_req string

//go:embed testdata/openrouter/hi_gemini_04_resp.json
var goldenOR_Hi_gemini_04_resp string

func TestOpenRouterAgentHiWithToolsNoCallGemini(t *testing.T) {
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

func TestAgentLMS_ToolsCall(t *testing.T) {
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
}

//go:embed testdata/lms/unknown_fn_call_01_req.json
var goldenLMS_UnknownFnCall_req_01 string

//go:embed testdata/lms/unknown_fn_call_02_resp.json
var goldenLMs_UnknownFnCall_resp_02 string

//go:embed testdata/lms/unknown_fn_call_03_req.json
var goldenLMS_UnknownFnCall_req_03 string

//go:embed testdata/lms/unknown_fn_call_04_resp.json
var goldenLMS_UnknownFnCall_resp_04 string

func TestAgentLMS_UnknownFnCall(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMS(t, testDefaultMaxAgentSteps)
	defer httpDo.AssertExpectations(t)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_UnknownFnCall_req_01, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMs_UnknownFnCall_resp_02)),
		}, nil)

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenLMS_UnknownFnCall_req_03, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenLMS_UnknownFnCall_resp_04)),
		}, nil)

	q := "call function GetSpecialData"
	prompt, err := rellm.NewPromptBuilder().
		WithMessage(q).
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	respMsg, err := agent.Execute(ctx, prompt)
	assert.Equal(t, respMsg, "")
	assert.Error(t, err, "failed to ask")
	assert.ErrorIs(t, err, rellm.ErrWhileDispatchToolCall)

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

	respMsg, err = agent.Execute(ctx, promptContinue)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "It appears that the attempt to call `GetSpecialData` resulted in an error indicating the function was not found. \n\nHow would you like to proceed? I can try calling one of the other available functions, such as `GetStaticData` or `GetDataFor`, if you provide a specific input.", respMsg, "response message should match")
}

func TestAgentLMS_DispatchFailurePersistsPartialToolResults(t *testing.T) {
	agent, httpDo := buildTestProToolAgentLMS(t, testDefaultMaxAgentSteps)
	defer httpDo.AssertExpectations(t)

	const response = `{
		"output": [
			{
				"id": "fc_success",
				"call_id": "call_success",
				"type": "function_call",
				"name": "GetStaticData",
				"arguments": "{}"
			},
			{
				"id": "fc_failure",
				"call_id": "call_failure",
				"type": "function_call",
				"name": "UnknownTool",
				"arguments": "{}"
			}
		],
		"error": null
	}`

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(response)),
		}, nil)

	prompt, err := rellm.NewPromptBuilder().
		WithMessage("call both tools").
		WithReasoning(testReasoningEffort).
		WithTemperature(testTemperature).
		Build()
	assert.NoError(t, err)

	ctx := context.Background()
	respMsg, err := agent.Execute(ctx, prompt)
	assert.Empty(t, respMsg)
	assert.ErrorIs(t, err, rellm.ErrWhileDispatchToolCall)

	conversation, err := agent.CurrentConversation()
	assert.NoError(t, err)

	outputsByCallID := make(map[string]string)
	for _, el := range conversation {
		if resp, ok := el.(*rellm.FunctionCallResp); ok && resp.Type == "function_call_output" {
			outputsByCallID[resp.CallID] = resp.Output
		}
	}

	assert.Equal(t, "42", outputsByCallID["call_success"], "successful tool result should have been persisted despite the dispatch error")
	assert.Equal(t, "invalid function call (function not found) UnknownTool", outputsByCallID["call_failure"])
}

func TestAgentLMSTooManyFunctionCall(t *testing.T) {
	const NotEnoughToolLoopLimit = 2
	agent, httpDo := buildTestProToolAgentLMS(t, NotEnoughToolLoopLimit)
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

	ctx := context.Background()
	_, err = agent.Execute(ctx, prompt)
	assert.Error(t, err, "expected error when tool iteration budget is exceeded")
	assert.True(t, errors.Is(err, rellm.ErrMaxAgentStepsReached))
}

func TestFuncResultToFunctionCallRespSerializesOutputAsString(t *testing.T) {
	resp := rellm.FuncResultToFunctionCallResp("call_123", int64(42))

	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal function call response")
	assert.JSONEq(t, `{"type":"function_call_output","call_id":"call_123","output":"42"}`, string(jsonResp))
}

func buildTestProToolAgentLMS(t *testing.T, maxAgentSteps uint64) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestProAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewLMStudioProviderWithHTTPClient("google/gemma-4-26b-a4b", testBaseUrl, testPort, mockHttp)
	assert.NoError(t, err, "failed to create provider")

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage("You are a helpful assistant.").
		WithToolset_X(&examplesutils.DataSrcToolset_X{}).
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func buildTestProToolAgentOpenRouter(t *testing.T, model rellm.Model, maxAgentSteps uint64) (*rellm.Agent, *HTTPDoMock) {

	agentName := "TestProAgent"
	mockHttp := new(HTTPDoMock)

	p, err := rellm.NewOpenRouterProviderWithHTTPClient("test-key", model, mockHttp)
	assert.NoError(t, err)

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage("You are a helpful assistant.").
		WithToolset_X(&examplesutils.DataSrcToolset_X{}).
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func buildConversationFromGoldenLms(goldens ...string) ([]rellm.ConversationElement, error) {
	rawConversation, err := buildRawConversationFromGolden(goldens...)
	if err != nil {
		return nil, fmt.Errorf("build raw conversation from golden: %w", err)
	}

	lmsProvider := rellm.LMStudioProvider{}
	return lmsProvider.ToConversationElements(rawConversation)
}

func buildConversationFromGoldenOpenRouter(goldens ...string) ([]rellm.ConversationElement, error) {
	rawConversation, err := buildRawConversationFromGolden(goldens...)
	if err != nil {
		return nil, fmt.Errorf("build raw conversation from golden: %w", err)
	}

	orProvider := rellm.OpenRouterProvider{}
	return orProvider.ToConversationElements(rawConversation)
}

func buildRawConversationFromGolden(goldens ...string) ([]json.RawMessage, error) {
	type golden struct {
		Input  []json.RawMessage `json:"input"`
		Output []json.RawMessage `json:"output"`
	}

	var conversation []json.RawMessage

	for _, data := range goldens {
		var g golden
		if err := json.Unmarshal([]byte(data), &g); err != nil {
			return nil, fmt.Errorf("unmarshal golden: %w", err)
		}

		conversation = append(conversation, g.Input...)
		conversation = append(conversation, g.Output...)
	}

	return conversation, nil
}
