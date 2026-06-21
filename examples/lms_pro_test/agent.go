package main

import (
	"encoding/json"
	"fmt"
	"rellm/pkg/agentsutills"
	"rellm/pkg/rellm"
)

const (
	sysprompt = `You are a helpful assistant.`
)

func buildTestProToolAgent(workspace string) (*rellm.Agent, error) {

	agentName := "TestProToolAgent"

	lmsEndpoint := rellm.NewUniversalResponsesEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponsesApiEndpoint(lmsEndpoint).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
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
		WithToolset(&agentsutills.DataSrcToolset{}).
		Build()
}

func SetParametersWithReqLog(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: "low"}
	req.Temperature = 0.5

	fmt.Println(" === REQ ===")
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
