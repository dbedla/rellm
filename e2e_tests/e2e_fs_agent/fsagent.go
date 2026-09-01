package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"rellm/internal/examplesutils"
	"rellm/pkg/agentsutils"
	"rellm/pkg/rellm"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

const (
	fsAgentSysPrompt = `You are a helpful assistant, with limited access to the file system.`
)

func buildFSAgent(p rellm.Provider, dirs agentsDirs) (*rellm.Agent, error) {

	agentName := "FSAgent"

	fsToolset, err := buildFSToolset(dirs.readOnlyDir, dirs.outputDir)
	if err != nil {
		return nil, err
	}

	conversationFilePath := path.Join(dirs.workspace, agentName+"_conversation.json")
	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(agentName).
		WithMaxAgentSteps(20).
		WithConversationStorage(rellm.NewFilesystemStorage(conversationFilePath)).
		WithToolset(fsToolset).
		WithSystemMessage(fsAgentSysPrompt).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		Build()
}

func buildLMSProvider() (rellm.Provider, error) {
	return rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
}

func buildOpenRouterProvider() (rellm.Provider, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load .env: %w", err)
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing apikey for OPENROUTER_API_KEY")
	}

	return rellm.NewOpenRouterProvider(apiKey, "google/gemini-3.1-flash-lite")
}

func buildFSToolset(readOnlyDir, outputDir string) (*agentsutils.FSToolset, error) {
	fs, err := agentsutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		return nil, err
	}

	return agentsutils.NewFSToolset(fs), nil
}

func buildFsPath() (agentsDirs, error) {
	workspace, err := agentsutils.CreateDirInSysTmp("agent-log")
	if err != nil {
		return agentsDirs{}, err
	}

	readOnlyDir, err := agentsutils.CreateSubDir(workspace, "readonly_agent_input")
	if err != nil {
		return agentsDirs{}, err
	}

	outputDir, err := agentsutils.CreateSubDir(workspace, "output")
	if err != nil {
		return agentsDirs{}, err
	}

	dirs := agentsDirs{
		readOnlyDir: readOnlyDir,
		outputDir:   outputDir,
		workspace:   workspace,
	}

	return dirs, nil
}

type agentsDirs struct {
	workspace   string
	readOnlyDir string
	outputDir   string
}

func setup(fl flag) (agentsDirs, *rellm.Agent, error) {

	dirs, err := buildFsPath()
	if err != nil {
		return agentsDirs{}, nil, err
	}

	ep, err := providerForFlag(fl)
	if err != nil {
		return agentsDirs{}, nil, err
	}

	fsAgent, err := buildFSAgent(ep, dirs)
	if err != nil {
		return agentsDirs{}, nil, err
	}

	return dirs, fsAgent, nil
}

func providerForFlag(fl flag) (rellm.Provider, error) {
	switch fl {
	case flag_LMS:
		return buildLMSProvider()
	case flag_OpenRouter:
		return buildOpenRouterProvider()
	default:
		return nil, fmt.Errorf("unknown flag provided: %s", fl)
	}

}

func panicWithLog(msg string, err error) {
	logMsg := msg + "\n" + err.Error()
	color.Red(logMsg)
	panic(logMsg)
}

func createFile(locationPath, fname, content string) error {
	filePath := filepath.Join(locationPath, fname)
	return os.WriteFile(filePath, []byte(content), 0644)
}

type flag string

const (
	flag_LMS        flag = "--lms"
	flag_OpenRouter flag = "--openrouter"
	flag_Invalid    flag = "NO_FLAG"
)

func help() {
	color.Yellow("allowed args:")
	color.Yellow("\t %s", flag_LMS)
	color.Yellow("\t %s", flag_OpenRouter)
}

func argsToFlag(args []string) flag {
	if len(args) != 2 {
		return flag_Invalid
	}

	f := flag(args[1])
	if f == flag_LMS || f == flag_OpenRouter {
		return f
	}

	return flag_Invalid
}
