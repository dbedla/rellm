package main

import (
	"fmt"
	"os"
	examplesutils2 "rellm/internal/examplesutils"
	"rellm/pkg/rellm"

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
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(sysprompt).
		WithToolset_X(&examplesutils2.DataSrcToolset_X{}).
		WithInspectEachRequest(examplesutils2.InspectWithReqLog).
		WithInspectEachResponse(examplesutils2.InspectWithRespLog).
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
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(sysprompt).
		WithToolset_X(&examplesutils2.DataSrcToolset_X{}).
		WithInspectEachRequest(examplesutils2.InspectWithReqLog).
		WithInspectEachResponse(examplesutils2.InspectWithRespLog).
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
