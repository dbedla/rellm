package main

import (
	"rellm/pkg/agent"
)

const (
	baseAgentSysPrompt = `You are a helpful assistant.`
)

func buildLBaseAgent(workspace string) *agent.Agent {

	agentName := "BaseAgent"

	params := agent.NewConversationParameterBuilder().
		Build()

	or := agent.NewCustomResponseEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep := agent.NewEndpointBuilder().
		WithResponseApiEndpoint(or).
		WithModel(agent.Model_Gemma_4).
		Build()

	baseAgent := agent.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(workspace).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(false).
		WithConversationParameters(params).
		WithSystemMessage(baseAgentSysPrompt).
		Build()

	return baseAgent
}
