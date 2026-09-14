package main

import (
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
)

const (
	mathAgentSysPrompt = `You are a helpful assistant with math knowledge. Use the calculator tools for all arithmetic operations.`
)

func buildMathAgent() (*rellm.Agent, error) {

	agentName := "MathAgent"

	p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithSystemMessage(mathAgentSysPrompt).
		WithToolset(NewCalculatorToolset(&Calculator{})).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()
}
