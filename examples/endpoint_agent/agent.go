package main

import (
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"
)

const (
	sysprompt = `You are a helpful assistant.`
)

func buildTestProToolAgent(workspace string) (*rellm.Agent, error) {

	agentName := "TestEndpointAgent"

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithProvider(rellm.Provider_LMStudio).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithDefaultHttpClient().
		Build()
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(false).
		WithSystemMessage(sysprompt).
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithWorkspaceLogger().
		WithInspectEachRequest(agentsutils.InspectWithReqLog).
		WithInspectEachResponse(agentsutils.InspectWithRespLog).
		Build()
}
