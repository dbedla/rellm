package rellm_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	//go:embed testdata/agentswitch/01_luna_whatisinfile_req.json
	goldenAS01LunaReq string

	//go:embed testdata/agentswitch/02_luna_whatisinfile_resp.json
	goldenAS02LunaResp string

	//go:embed testdata/agentswitch/03_luna_whatisinfile_req.json
	goldenAS03LunaReq string

	//go:embed testdata/agentswitch/04_luna_whatisinfile_resp.json
	goldenAS04LunaResp string

	//go:embed testdata/agentswitch/05_lms_saveinoutput_req.json
	goldenAS05LmsReq string

	//go:embed testdata/agentswitch/06_lms_saveinoutput_resp.json
	goldenAS06LmsResp string

	//go:embed testdata/agentswitch/07_lms_saveinoutput_req.json
	goldenAS07LmsReq string

	//go:embed testdata/agentswitch/08_lms_saveinoutput_resp.json
	goldenAS08LmsResp string
)

const asReadOnlyDir = "/var/folders/nj/bkj2m2nx3292lgx0x2zsh35h0000gn/T/agent-log-2572546028/readonly_agent_input"
const asOutputDir = "/var/folders/nj/bkj2m2nx3292lgx0x2zsh35h0000gn/T/agent-log-2572546028/output"

func TestAgentSwitchProvider(t *testing.T) {
	ctx := context.Background()
	sharedConversation := rellm.NewInMemoryConversation()

	seed, err := seedConversationFromGolden(goldenAS01LunaReq)
	assert.NoError(t, err)
	assert.NoError(t, sharedConversation.Append(ctx, seed))

	// --- Phase 1: OpenRouter agent on the shared conversation ---
	orAgent, orHttp, orTools := buildTestFSAgentOpenRouterWithConversation(t, "openai/gpt-5.6-luna", testDefaultMaxAgentSteps, sharedConversation)
	defer orHttp.AssertExpectations(t)
	defer orTools.AssertExpectations(t)

	orTools.On("Definitions").
		Return(examplesutils.NewFSToolset(nil).Definitions()).Once()
	orTools.On("Dispatch", mock.Anything, "FSToolset_GetFileContentAsString",
		quotedJSONString(`{"path":"`+asReadOnlyDir+`/locations.txt"}`)).
		Return(rellm.ToolCallResult{Value: "London"}, nil).Once()
	orTools.On("Dispatch", mock.Anything, "FSToolset_GetFileContentAsString",
		quotedJSONString(`{"path":"`+asReadOnlyDir+`/names.txt"}`)).
		Return(rellm.ToolCallResult{Value: "John Smith"}, nil).Once()

	expectGoldenReqAndReturnGoldenResp(t, orHttp, goldenAS01LunaReq, goldenAS02LunaResp)
	expectGoldenReqAndReturnGoldenResp(t, orHttp, goldenAS03LunaReq, goldenAS04LunaResp)

	finalReport, err := orAgent.Ask(ctx, "show me what is in files you mention")
	assert.NoError(t, err)
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "`locations.txt`:\n```text\nLondon\n```\n\n`names.txt`:\n```text\nJohn Smith\n```", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenAS02LunaResp, goldenAS04LunaResp), finalReport.StepsStats)

	orConversation, err := buildConversationFromGoldenOpenRouter(goldenAS03LunaReq, goldenAS04LunaResp)
	assert.NoError(t, err)

	sharedConversationAfterOR, err := sharedConversation.Load(ctx)
	assert.NoError(t, err)
	assert.Equal(t, orConversation, sharedConversationAfterOR)

	// --- Phase 2: LMStudio agent on the same conversation ---
	lmsAgent, lmsHttp, lmsTools := buildTestFSAgentLMSWithConversation(t, testDefaultMaxAgentSteps, sharedConversation)
	defer lmsHttp.AssertExpectations(t)
	defer lmsTools.AssertExpectations(t)

	lmsTools.On("Definitions").
		Return(examplesutils.NewFSToolset(nil).Definitions()).Once()
	lmsTools.On("Dispatch", mock.Anything, "FSToolset_WriteStringToFile",
		quotedJSONString(`{"content":"London\nJohn Smith","path":"`+asOutputDir+`/data.txt"}`)).
		Return(rellm.ToolCallResult{Value: "ok"}, nil).Once()

	expectGoldenReqAndReturnGoldenResp(t, lmsHttp, goldenAS05LmsReq, goldenAS06LmsResp)
	expectGoldenReqAndReturnGoldenResp(t, lmsHttp, goldenAS07LmsReq, goldenAS08LmsResp)

	finalReport, err = lmsAgent.Ask(ctx, "create file in your output directory, file name 'data.txt', file should contain content of both files from read only director")
	assert.NoError(t, err, "failed to ask with an OpenRouter-built conversation")
	assert.NotEmpty(t, finalReport.Message)
	assert.Equal(t, "I have created the file `data.txt` in your output directory with the combined content from both files.", finalReport.Message)
	assert.Equal(t, expectedStepStats(t, goldenAS06LmsResp, goldenAS08LmsResp), finalReport.StepsStats)

	finalConversation, err := lmsAgent.Conversation().Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, finalConversation, 26)
}

// seedConversationFromGolden rebuilds the conversation history preceding the
// captured scenario from a golden request: every input element except the
// last one (the follow-up user question) is converted the same way the agent
// converts stored conversation elements.
func seedConversationFromGolden(goldenReq string) ([]rellm.ConversationElement, error) {
	var g struct {
		Input []json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal([]byte(goldenReq), &g); err != nil {
		return nil, err
	}

	seed := g.Input[:len(g.Input)-1]
	raw, err := json.Marshal(map[string]json.RawMessage{"input": mustMarshal(seed)})
	if err != nil {
		return nil, err
	}

	return buildConversationFromGoldenOpenRouter(string(raw))
}

// quotedJSONString encodes s as a JSON string: Dispatch receives the tool
// call arguments re-marshaled, i.e. double-encoded.
func quotedJSONString(s string) json.RawMessage {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return b
}

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func expectGoldenReqAndReturnGoldenResp(t *testing.T, httpDo *HTTPDoMock, goldenReq, goldenResp string) {
	t.Helper()

	httpDo.On("Do", mock.MatchedBy(baseRequestMatch)).
		Once().
		Run(func(args mock.Arguments) {
			req := args.Get(0).(*http.Request)

			b, err := io.ReadAll(req.Body)
			assert.NoError(t, err, "failed to read request body")

			req.Body = io.NopCloser(bytes.NewBuffer(b))

			assert.JSONEq(t, goldenReq, string(b))
		}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(goldenResp)),
		}, nil)
}

func buildTestFSAgentOpenRouterWithConversation(t *testing.T, model rellm.Model, maxAgentSteps uint64, conversation rellm.Conversation) (*rellm.Agent, *HTTPDoMock, *ToolsetMock) {

	agentName := "TestFSAgent"
	mockHttp := new(HTTPDoMock)
	fsToolset := new(ToolsetMock)

	p, err := rellm.NewOpenRouterProviderWithHTTPClient("test-key", model, mockHttp)
	assert.NoError(t, err)

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversation(conversation).
		WithSystemMessage("You are a helpful assistant, with limited access to the file system.").
		WithToolset(fsToolset, rellm.ParallelToolCallsEnable).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp, fsToolset
}

func buildTestFSAgentLMSWithConversation(t *testing.T, maxAgentSteps uint64, conversation rellm.Conversation) (*rellm.Agent, *HTTPDoMock, *ToolsetMock) {

	agentName := "TestFSAgent"
	mockHttp := new(HTTPDoMock)
	fsToolset := new(ToolsetMock)

	p, err := rellm.NewLMStudioProviderWithHTTPClient("google/gemma-4-26b-a4b", testBaseUrl, testPort, mockHttp)
	assert.NoError(t, err)

	ta, err := rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(maxAgentSteps).
		WithConversation(conversation).
		WithSystemMessage("You are a helpful assistant, with limited access to the file system.").
		WithToolset(fsToolset, rellm.ParallelToolCallsEnable).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return ta, mockHttp, fsToolset
}
