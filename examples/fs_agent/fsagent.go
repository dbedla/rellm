package main

import (
	"rellm/pkg/rellm"
	"rellm/pkg/tools/limited_file_system"
	"rellm/pkg/utils"
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

	params := rellm.NewConversationParameterBuilder().
		WithToolset(fsToolset.BuildTools()).
		Build()

	or := rellm.NewCustomResponseEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep := rellm.NewEndpointBuilder().
		WithResponseApiEndpoint(or).
		WithModel(rellm.Model_Gemma_4_26b_a4b).
		Build()

	baseAgent := rellm.NewAgentBuilder().
		WithEndpoint(ep).
		WithAgentName(agentName).
		WithWorkspaceDir(logDir).
		WithMaxToolsIterationWithoutReturnMessage(20).
		WithContinueConversation(true).
		WithConversationParameters(params).
		WithSystemMessage(fsAgentSysPrompt).
		WithToolsDispatcher(fsToolset.DispatchTools).
		Build()

	return baseAgent, nil
}

func buildFSToolset(readOnlyDir, outputDir string) (*lfs.FSToolset, error) {
	smallFS, err := lfs.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		return nil, err
	}
	fsToolset := lfs.NewFSToolset(smallFS)

	return fsToolset, nil
}

func buildFsPath(workspace string) (string, string, string, error) {
	readOnlyDir, err := utils.CreateSubDir(workspace, "readonly_agent_input")
	if err != nil {
		return "", "", "", err
	}

	outputDir, err := utils.CreateSubDir(workspace, "output")
	if err != nil {
		return "", "", "", err
	}

	logdir, err := utils.CreateSubDir(workspace, "logs")
	if err != nil {
		return "", "", "", err
	}
	return readOnlyDir, outputDir, logdir, nil
}
