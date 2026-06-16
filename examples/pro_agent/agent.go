package main

import (
	"rellm/pkg/rellm"
)

const (
	proAgentSysPrompt = `You are a helpful assistant with deep weather knowledge.`
)

func buildProAgent(workspace string) (*rellm.Agent, error) {

	agentName := "ProAgent"

	or := rellm.NewUniversalResponsesEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponseApiEndpoint(or).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
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
		WithSystemMessage(proAgentSysPrompt).
		WithToolset(&WeatherToolset{}).
		Build()
}

func SetParameters(req *rellm.ResponsesApiRequest) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: "low"}
	req.Temperature = 0.5
}
