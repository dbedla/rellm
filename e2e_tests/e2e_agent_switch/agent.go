package main

import (
	"fmt"
	"os"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"

	"github.com/joho/godotenv"
)

const switchAgentSysPrompt = `You are a helpful assistant.`

func buildORLunaSwitchAgent(conversation rellm.Conversation) (*rellm.Agent, error) {
	p, err := newOpenRouterProvider("openai/gpt-5.6-luna")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName("or-luna-switch-agent").
		WithMaxAgentSteps(20).
		WithConversation(conversation).
		WithSystemMessage(switchAgentSysPrompt).
		WithToolset(&examplesutils.DataSrcToolset{}).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()
}

func buildLMSSwitchAgent(conversation rellm.Conversation) (*rellm.Agent, error) {
	p, err := rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
	if err != nil {
		return nil, err
	}

	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName("lms-switch-agent").
		WithMaxAgentSteps(20).
		WithConversation(conversation).
		WithSystemMessage(switchAgentSysPrompt).
		WithToolset(&examplesutils.DataSrcToolset{}).
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
