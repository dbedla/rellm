package main

import (
	"rellm/pkg/rellm"
)

const (
	askLikeProAgentSysPrompt = `You are a helpful assistant with deep weather knowledge.`
)

func buildLAskLikeProAgent(workspace string) *rellm.Agent {

	agentName := "AskLikeProAgent"

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
		WithSystemMessage(askLikeProAgentSysPrompt).
		WithToolset(&WeatherToolset{}).
		Build()

	return baseAgent
}

func SetParameters(req *rellm.ResponsesApiRequest) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: "low"}
	req.Temperature = 0.5
}
