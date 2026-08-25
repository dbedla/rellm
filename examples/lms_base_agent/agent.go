package main

import (
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
)

const (
	baseAgentSysPrompt = `You are a helpful assistant.`
)

func buildLBaseAgent() (*rellm.Agent, error) {

	agentName := "BaseAgent"

	p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(baseAgentSysPrompt).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		Build()
}
