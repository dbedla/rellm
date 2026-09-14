package main

import (
	"encoding/json"
	"fmt"
	"os"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const structuredOutputSysPrompt = `You are an assistant that extracts structured information from user text and returns only valid JSON matching the requested schema.`

func buildLMSStructuredOutputAgent() (*rellm.Agent, error) {
	p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}
	schema, err := json.Marshal(textFormat.Schema)
	if err != nil {
		return nil, err
	}
	lmsSysPrompt := structuredOutputSysPrompt + string(schema) + "\nONLY parsable json string allowed as result, no additional markdown formatting\n"
	return buildStructuredOutputAgent(p, "lms-agent", lmsSysPrompt)
}

func buildORGlm3flashStructuredOutputAgent() (*rellm.Agent, error) {
	p, err := newOpenRouterProvider("z-ai/glm-5.3-flash")
	if err != nil {
		return nil, err
	}
	return buildStructuredOutputAgent(p, "or-glm-5.3-flash-agent", structuredOutputSysPrompt)
}

func buildORLunaStructuredOutputAgent() (*rellm.Agent, error) {
	p, err := newOpenRouterProvider("openai/gpt-5.6-luna")
	if err != nil {
		return nil, err
	}
	return buildStructuredOutputAgent(p, "or-luna-agent", structuredOutputSysPrompt)
}

func buildStructuredOutputAgent(provider rellm.Provider, name string, sysPrompt string) (*rellm.Agent, error) {

	return rellm.NewAgentBuilder().
		WithAgentName(name).
		WithProvider(provider).
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		WithSystemMessage(sysPrompt).
		WithTextFormat(textFormat).
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
