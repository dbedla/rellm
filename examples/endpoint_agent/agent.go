package main

import (
	"fmt"
	"net/http"
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

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithProvider(rellm.Provider_LMStudio).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
		WithDefaultHttpClient().
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
		WithSystemMessage(sysprompt).
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithWorkspaceLogger().
		WithInspectEachRequest(agentsutils.InspectWithReqLog).
		WithInspectEachResponse(agentsutils.InspectWithRespLog).
		Build()
}

func buildLOpenRouterAgent(workspace string, model rellm.Model) (*rellm.Agent, error) {

	agentName := "OpenRouterImageAgent"

	orEndpoint, err := newOpenRouterEndpoint()
	if err != nil {
		return nil, err
	}
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(orEndpoint).
		WithProvider(rellm.Provider_OpenRouter).
		WithModel(model).
		WithDefaultHttpClient().
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
		WithSystemMessage(sysprompt).
		WithWorkspaceLogger().
		WithToolset(&agentsutils.DataSrcToolset{}).
		WithInspectEachRequest(agentsutils.InspectWithReqLog).
		WithInspectEachResponse(agentsutils.InspectWithRespLog).
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
