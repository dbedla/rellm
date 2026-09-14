package main

import (
	"fmt"
	"os"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const (
	baseAgentSysPrompt = `You are a helpful assistant.`
)

func buildBaseAgent() (*rellm.Agent, error) {

	agentName := "OpenRouterAgent"

	p, err := newOpenRouterProvider("google/gemini-3.1-flash-lite")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithSystemMessage(baseAgentSysPrompt).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
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
