package rellm_test

import (
	"context"
	"fmt"
	"os"
	"rellm/internal/examplesutils"
	"rellm/pkg/rellm"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestXXX(t *testing.T) {
	openAI, err := buildOpenAIProvider(rellm.Model("gpt-5.6-luna"))
	assert.NoError(t, err)

	openAIAgent, err := rellm.NewAgentBuilder().
		WithProvider(openAI).
		WithAgentName("TestOpenAIAgent").
		WithMaxAgentSteps(20).
		WithConversation(rellm.NewInMemoryConversation()).
		//WithToolset(fsToolset, rellm.ParallelToolCallsEnable).
		//WithSystemMessage(switchAgentSysPrompt).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()

	assert.NoError(t, err)

	ctx := context.Background()
	report, err := openAIAgent.Ask(ctx, "Hello")
	assert.NoError(t, err)
	assert.Equal(t, "zzzzz", report.Message)
	assert.Empty(t, report.Images)
}

func buildOpenAIProvider(model rellm.Model) (rellm.Provider, error) {
	err := godotenv.Overload("/Users/dawidbedla/development/rellm/rellm/.env")
	if err != nil {
		return nil, fmt.Errorf("cannot load .env: %w", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing apikey for OPENAI_API_KEY")
	}

	return rellm.NewOpenAIProvider(apiKey, model)
}
