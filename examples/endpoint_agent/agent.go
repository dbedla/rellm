package main

import (
	"fmt"
	"os"
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const (
	sysprompt = `You are a helpful assistant.`
)

func buildLMSAgent(workspace string) (*rellm.Agent, error) {

	agentName := "TestEndpointAgent"

	p, err := rellm.NewLMStudioProvider(rellm.Model_LMS_Google_Gemma_4_26B_A4B, "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(sysprompt).
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithInspectEachRequest(agentsutils.InspectWithReqLog).
		WithInspectEachResponse(agentsutils.InspectWithRespLog).
		Build()
}

func buildLOpenRouterAgent(workspace string, model rellm.Model) (*rellm.Agent, error) {

	agentName := "OpenRouterImageAgent"

	p, err := newOpenRouterProvider(model)
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(sysprompt).
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithInspectEachRequest(agentsutils.InspectWithReqLog).
		WithInspectEachResponse(agentsutils.InspectWithRespLog).
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
