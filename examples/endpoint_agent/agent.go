package main

import (
	"fmt"
	"os"
	examplesutils2 "github.com/dbedla/rellm/internal/examplesutils"
	"github.com/dbedla/rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const (
	sysprompt = `You are a helpful assistant.`
)

func buildLMSAgent() (*rellm.Agent, error) {

	agentName := "TestEndpointAgent"

	p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithSystemMessage(sysprompt).
		WithToolset(&examplesutils2.DataSrcToolset{}, rellm.ParallelToolCallsDefaultForProvider).
		WithInspectEachRequest(examplesutils2.InspectWithReqLog).
		WithInspectEachResponse(examplesutils2.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()
}

func buildOpenRouterAgent(model rellm.Model) (*rellm.Agent, error) {

	agentName := "OpenRouterImageAgent"

	p, err := newOpenRouterProvider(model)
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithSystemMessage(sysprompt).
		WithToolset(&examplesutils2.DataSrcToolset{}, rellm.ParallelToolCallsDefaultForProvider).
		WithInspectEachRequest(examplesutils2.InspectWithReqLog).
		WithInspectEachResponse(examplesutils2.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()
}

func newOpenRouterProvider(model rellm.Model) (rellm.Provider, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load .env: %w", err)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing apikey for OPENROUTER_API_KEY")
	}

	return rellm.NewOpenRouterProvider(apiKey, model)
}
