package rellm_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testBaseUrl                                      = "http://127.0.0.1"
	testPort                                         = "1234"
	testResponsesApiEndpoint                         = "/v1/responses"
	testReasoningEffort                              = "low"
	testTemperature                                  = 0.5
	TestDefaultMaxToolsIterationWithoutReturnMessage = 5
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

	respMsg, resp, err := agent.AskLikeAPro("Hi", setTestReasoningAndTemperature, nil)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "Hello! How can I help you today?", respMsg, "response message should match")

	assert.NotNil(t, resp, "response should not be nil")

	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal response")
	assert.JSONEq(t, goldenProRespHi, string(jsonResp))

	_, _, err = agent.AskLikeAPro("What did you reason about?", setTestReasoningAndTemperature, nil)
	assert.NoError(t, err, "failed to ask with reasoning in conversation")
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
	respMsg, resp, err := agent.AskLikeAPro(q, setTestReasoningAndTemperature, nil)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "The first tool call to `GetStaticData` returned the value `42`. The second tool call to `GetDataFor` with the input \"the meaning of 42\" returned a list containing `[\"abc\", \"def\"]`. Therefore, based on these specific tool outputs, the data associated with the value 42 is \"abc\" and \"def\".", respMsg, "response message should match")

	assert.NotNil(t, resp, "response should not be nil")

	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal response")
	assert.JSONEq(t, goldenProRespA3, string(jsonResp))

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
	respMsg, resp, err := agent.AskLikeAPro(q, SetParametersWithReqLog, SniffResp)
	assert.Error(t, err, "failed to ask")
	//assert.Equal(t, err, rellm.ErrUnknownToolCall)
	assert.ErrorIs(t, err, rellm.ErrUnknownToolCallsErrorsWillBePassedToModelInNextReq)

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "", respMsg, "response message should match")

	//since we get err ErrUnknownToolCall we simulate user ask to continue
	q = "continue"
	respMsg, resp, err = agent.AskLikeAPro(q, SetParametersWithReqLog, SniffResp)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "It appears that the attempt to call `GetSpecialData` resulted in an error indicating the function was not found. \n\nHow would you like to proceed? I can try calling one of the other available functions, such as `GetStaticData` or `GetDataFor`, if you provide a specific input.", respMsg, "response message should match")

	assert.NotNil(t, resp, "response should not be nil")
	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal response")
	assert.JSONEq(t, goldenProRespB2_UnknownFnCAll, string(jsonResp))

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
	respMsg, resp, err := agent.AskLikeAPro(q, setTestReasoningAndTemperature, nil)
	assert.NoError(t, err, "failed to ask")

	assert.NotNil(t, respMsg, "response message should not be nil")
	assert.Equal(t, "Warn too many function call iterations without return message", respMsg, "response message should match")

	assert.Nil(t, resp, "response should not be nil")
}

func TestFuncResultToFunctionCallRespSerializesOutputAsString(t *testing.T) {
	resp := rellm.FuncResultToFunctionCallResp("call_123", int64(42))

	jsonResp, err := json.Marshal(resp)
	assert.NoError(t, err, "failed to marshal function call response")
	assert.JSONEq(t, `{"type":"function_call_output","call_id":"call_123","output":"42"}`, string(jsonResp))
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

func buildTestProToolAgent(t *testing.T, maxToolsIterationWithoutReturnMessage uint64) (*rellm.Agent, *HttpDoMock) {

	agentName := "TestProAgent"
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
		WithMaxToolsIterationWithoutReturnMessage(maxToolsIterationWithoutReturnMessage).
		WithContinueConversation(false).
		WithSystemMessage("You are a helpful assistant.").
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithNoOpLogger().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp
}

func baseRequestMatch(req *http.Request) bool {
	return req.URL.String() == testBaseUrl+":"+testPort+testResponsesApiEndpoint &&
		req.Method == "POST"
}

func setTestReasoningAndTemperature(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: testReasoningEffort}
	req.Temperature = testTemperature
}

type HttpDoMock struct {
	mock.Mock
}

func (h *HttpDoMock) Do(req *http.Request) (*http.Response, error) {
	args := h.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func SetParametersWithReqLog(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: "medium"}
	req.Temperature = 0.5

	color.White(" === REQ ===")
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	color.White(string(b))
}

func SniffResp(req *rellm.ResponsesApiResp) {
	color.White(" === RESP ===")
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	color.White(string(b))
}
