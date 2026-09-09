package main

import (
	"fmt"
	"os"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const structuredOutputSysPrompt = `You are an assistant that extracts structured information from user text and returns only valid JSON matching the requested schema.`

func buildStructuredOutputAgent() (*rellm.Agent, error) {

	agentName := "OpenRouterStructuredOutputAgent"

	// p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	// p, err := newOpenRouterProvider("google/gemini-3.1-flash-lite")
	p, err := newOpenRouterProvider("openai/gpt-5.6-luna")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(structuredOutputSysPrompt).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
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
