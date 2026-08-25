package main

import (
	"context"
	"fmt"
	"os"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const (
	baseAgentSysPrompt = `You are a helpful assistant.`
)

func buildLBaseAgent() (*rellm.Agent, error) {

	agentName := "OpenRouterImageAgent"

	p, err := newOpenRouterProvider(rellm.Model("x-ai/grok-imagine-image-quality"))
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversationStorage(rellm.NewInMemoryStorage()).
		WithSystemMessage(baseAgentSysPrompt).
		WithHandleImageGeneration(testHandleImage).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		Build()
}

func testHandleImage(_ context.Context, image *rellm.ImageGeneration) (string, error) {
	return "asd", nil
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
