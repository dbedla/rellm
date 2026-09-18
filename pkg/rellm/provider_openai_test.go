package rellm_test

import (
	"context"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testReadOnlyDir = "/test/readonly_agent_input"
const testDir = "/test/output"

func disable_TestXXX(t *testing.T) {

	openAIAgent, httpMock, fsMock, err := buildTestOpenAIFSAgent()
	assert.NoError(t, err)
	defer httpMock.AssertExpectations(t)
	assert.NoError(t, err)

	fsMock.On("Definitions").
		Return(examplesutils.NewFSToolset(nil).Definitions()).Once()
	fsMock.On("Dispatch", mock.Anything, "FSToolset_GetFileContentAsString",
		quotedJSONString(`{"path":"`+testReadOnlyDir+`/locations.txt"}`)).
		Return(rellm.ToolCallResult{Value: "London"}, nil).Once()
	fsMock.On("Dispatch", mock.Anything, "FSToolset_GetFileContentAsString",
		quotedJSONString(`{"path":"`+testReadOnlyDir+`/names.txt"}`)).
		Return(rellm.ToolCallResult{Value: "John Smith"}, nil).Once()

	ctx := context.Background()
	report, err := openAIAgent.Ask(ctx, "Hello")
	assert.NoError(t, err)
	assert.Equal(t, "zzzzz", report.Message)
	assert.Empty(t, report.Images)
}

func buildTestOpenAIFSAgent() (*rellm.Agent, *HTTPDoMock, *ToolsetMock, error) {
	openAI, httpMock, err := buildTestOpenAIProvider("gpt-5.6-luna")
	if err != nil {
		return nil, nil, nil, err
	}

	sysPrompt := "You are a helpful assistant, with limited access to the file system."
	fsToolset := new(ToolsetMock)

	openAIAgent, err := rellm.NewAgentBuilder().
		WithProvider(openAI).
		WithAgentName("TestOpenAIAgent").
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithToolset(fsToolset, rellm.ParallelToolCallsEnable).
		WithSystemMessage(sysPrompt).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	if err != nil {
		return nil, nil, nil, err
	}

	return openAIAgent, httpMock, fsToolset, nil
}

func buildTestOpenAIProvider(model rellm.Model) (rellm.Provider, *HTTPDoMock, error) {
	mock := HTTPDoMock{}
	p, err := rellm.NewOpenAIProviderWithHTTPClient("test-api-key-openai", model, &mock)
	if err != nil {
		return nil, nil, err
	}
	return p, &mock, nil
}
