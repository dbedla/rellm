package main

import (
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"
)

const (
	fsAgentSysPrompt = `You are a helpful assistant, with limited access to the file system.`
)

func buildFSAgent(workspace string) (*rellm.Agent, error) {

	agentName := "FSAgent"

	readOnlyDir, outputDir, logDir, err := buildFsPath(workspace)
	if err != nil {
		return nil, err
	}

	fsToolset, err := buildFSToolset(readOnlyDir, outputDir)
	if err != nil {
		return nil, err
	}

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
		WithWorkspaceDir(logDir).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(true).
		WithToolset(fsToolset).
		WithSystemMessage(fsAgentSysPrompt).
		WithWorkspaceLogger().
		WithInspectEachRequest(agentsutils.InspectWithReqLog).
		WithInspectEachResponse(agentsutils.InspectWithRespLog).
		Build()
}

func buildFSToolset(readOnlyDir, outputDir string) (*agentsutils.FSToolset, error) {
	fs, err := agentsutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		return nil, err
	}
	fsToolset := agentsutils.NewFSToolset(fs)

	return fsToolset, nil
}

func buildFsPath(workspace string) (string, string, string, error) {
	readOnlyDir, err := agentsutils.CreateSubDir(workspace, "readonly_agent_input")
	if err != nil {
		return "", "", "", err
	}

	outputDir, err := agentsutils.CreateSubDir(workspace, "output")
	if err != nil {
		return "", "", "", err
	}

	logdir, err := agentsutils.CreateSubDir(workspace, "logs")
	if err != nil {
		return "", "", "", err
	}
	return readOnlyDir, outputDir, logdir, nil
}
