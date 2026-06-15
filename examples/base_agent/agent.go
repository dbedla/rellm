package main

import (
	"rellm/pkg/rellm"
)

const (
	baseAgentSysPrompt = `You are a helpful assistant.`
)

func buildLBaseAgent(workspace string) *rellm.Agent {

	agentName := "BaseAgent"

	or := rellm.NewCustomResponseEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep := rellm.NewEndpointBuilder().
		WithResponseApiEndpoint(or).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		Build()

	baseAgent := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(false).
		WithSystemMessage(baseAgentSysPrompt).
		Build()

	return baseAgent
}
