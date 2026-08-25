package main

import (
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
)

const (
	proAgentSysPrompt = `You are a helpful assistant with deep weather knowledge.`
)

func buildProAgent() (*rellm.Agent, error) {

	agentName := "ProAgent"

	p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(proAgentSysPrompt).
		WithToolset(&WeatherToolset{}).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		Build()
}
