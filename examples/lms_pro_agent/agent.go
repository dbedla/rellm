package main

import (
	"rellm/pkg/examplesutils"
	"rellm/pkg/rellm"
)

const (
	proAgentSysPrompt = `You are a helpful assistant with deep weather knowledge.`
)

func buildProAgent() (*rellm.Agent, error) {

	agentName := "ProAgent"

	p, err := rellm.NewLMStudioProvider(rellm.Model_LMS_Google_Gemma_4_26B_A4B, "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(proAgentSysPrompt).
		WithToolset(&WeatherToolset{}).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		Build()
}
