package rellm_test

import (
	"context"
	"github.com/dbedla/rellm/internal/examplesutils"
	"github.com/dbedla/rellm/pkg/rellm"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testReadOnlyDir = "/test/readonly_agent_input"
const testDir = "/test/output"

var (
	//go:embed testdata/openai/fs_01_whatfile_req.json
	goldenOS01Req string

	//go:embed testdata/openai/fs_02_whatfile_resp.json
	goldenOS02Resp string

	//go:embed testdata/openai/fs_03_whatfile_req.json
	goldenOS03Req string

	//go:embed testdata/openai/fs_04_whatfile_resp.json
	goldenOS04Resp string

	//go:embed testdata/openai/fs_05_whatfile_req.json
	goldenOS05Req string

	//go:embed testdata/openai/fs_06_whatfile_resp.json
	goldenOS06Resp string

	//go:embed testdata/openai/fs_07_whatfile_req.json
	goldenOS07Req string

	//go:embed testdata/openai/fs_08_whatfile_resp.json
	goldenOS08Resp string
)

func TestOpenAIAgentFSScenario(t *testing.T) {
	openAIAgent, httpMock, fsMock := buildTestOpenAIFSAgent(t)
	defer httpMock.AssertExpectations(t)
	defer fsMock.AssertExpectations(t)

	fsMock.On("Definitions").
		Return(examplesutils.NewFSToolset(nil).Definitions()).Once()
	fsMock.On("Dispatch", mock.Anything, "FSToolset_GetReadOnlyPaths",
		quotedJSONString(`{}`)).
		Return(rellm.ToolCallResult{Value: []string{testReadOnlyDir}}, nil).Once()
	fsMock.On("Dispatch", mock.Anything, "FSToolset_GetOutputDir",
		quotedJSONString(`{}`)).
		Return(rellm.ToolCallResult{Value: testDir}, nil).Once()
	fsMock.On("Dispatch", mock.Anything, "FSToolset_ListFilesIn",
		quotedJSONString(`{"path":"`+testReadOnlyDir+`"}`)).
		Return(rellm.ToolCallResult{Value: []string{testReadOnlyDir + "/locations.txt", testReadOnlyDir + "/names.txt"}}, nil).Once()
	fsMock.On("Dispatch", mock.Anything, "FSToolset_ListFilesIn",
		quotedJSONString(`{"path":"`+testDir+`"}`)).
		Return(rellm.ToolCallResult{}, nil).Once()

	expectGoldenReqAndReturnGoldenResp(t, httpMock, goldenOS01Req, goldenOS02Resp)
	expectGoldenReqAndReturnGoldenResp(t, httpMock, goldenOS03Req, goldenOS04Resp)
	expectGoldenReqAndReturnGoldenResp(t, httpMock, goldenOS05Req, goldenOS06Resp)
	expectGoldenReqAndReturnGoldenResp(t, httpMock, goldenOS07Req, goldenOS08Resp)

	ctx := context.Background()
	finalReport, err := openAIAgent.Ask(ctx, "what files do you see")
	assert.NoError(t, err)
	assert.Equal(t, "I can see these files:\n\n- `locations.txt`\n- `names.txt`\n\nThe output directory is currently empty.", finalReport.Message)
	assert.Empty(t, finalReport.Images)
	assert.Equal(t, expectedStepStats(t, goldenOS02Resp, goldenOS04Resp, goldenOS06Resp, goldenOS08Resp), finalReport.StepsStats)

	conversation, err := openAIAgent.Conversation().Load(ctx)
	assert.NoError(t, err)
	assert.Len(t, conversation, 13)
}

func buildTestOpenAIFSAgent(t *testing.T) (*rellm.Agent, *HTTPDoMock, *ToolsetMock) {
	t.Helper()

	openAI, httpMock, err := buildTestOpenAIProvider("gpt-5.6-luna")
	assert.NoError(t, err, "failed to create provider")

	sysPrompt := "You are a helpful assistant, with limited access to the file system."
	fsToolset := new(ToolsetMock)

	openAIAgent, err := rellm.NewAgentBuilder().
		WithProvider(openAI).
		WithAgentName("TestOpenAIAgent").
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithToolset(fsToolset, rellm.ParallelToolCallsDefaultForProvider).
		WithSystemMessage(sysPrompt).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	assert.NoError(t, err, "failed to create agent")
	return openAIAgent, httpMock, fsToolset
}

func buildTestOpenAIProvider(model rellm.Model) (rellm.Provider, *HTTPDoMock, error) {
	mock := HTTPDoMock{}
	p, err := rellm.NewOpenAIProviderWithHTTPClient("test-api-key-openai", model, &mock)
	if err != nil {
		return nil, nil, err
	}
	return p, &mock, nil
}
