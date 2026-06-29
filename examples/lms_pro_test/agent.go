package main

import (
	"encoding/json"
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
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
		Build()
}

func SetParametersWithReqLog(req *rellm.ResponsesApiReq) {
	req.Reasoning = &rellm.ReasoningConfig{Effort: "medium"}
	req.Temperature = 0.5

	color.White(" === REQ ===")
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	color.White(string(b))
}

func SniffResp(req *rellm.ResponsesApiResp) {
	color.White(" === RESP ===")
	b, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	color.White(string(b))
}
