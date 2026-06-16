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

	or := rellm.NewUniversalResponsesEndpoint("http://127.0.0.1", "1234", "/v1/responses", nil)
	ep, err := rellm.NewEndpointBuilder().
		WithResponseApiEndpoint(or).
		WithModel(rellm.Model_LMS_Google_Gemma_4_26B_A4B).
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
		Build()
}

func buildFSToolset(readOnlyDir, outputDir string) (*lfs.FSToolset, error) {
	fs, err := lfs.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		return nil, err
	}
	fsToolset := lfs.NewFSToolset(fs)

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
