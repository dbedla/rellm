package main

import (
	"fmt"
	"net/http"
	"os"
	"rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const (
	baseAgentSysPrompt = `You are a helpful assistant.`
)

func buildLBaseAgent(workspace string) (*rellm.Agent, error) {

	agentName := "OpenRouterAgent"

	orEndpoint, err := newOpenRouterEndpoint()
	if err != nil {
		return nil, err
	}
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(orEndpoint).
		WithModel(rellm.Model_OpenRouter_Google_Gemini_3_1_Flash_Lite).
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
		WithSystemMessage(baseAgentSysPrompt).
		Build()
}

func newOpenRouterEndpoint() (rellm.ResponsesApiEndpoint, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load .env: %w", err)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing apikey for OPENROUTER_API_KEY")
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Bearer "+apiKey)

	return rellm.NewUniversalResponsesEndpoint("https://openrouter.ai", "", "/api/v1/responses", header), nil

}
