package main

import (
	"fmt"
	"os"
	"path"
	"github.com/dbedla/rellm/internal/examplesutils"
	"github.com/dbedla/rellm/pkg/rellm"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

const switchAgentSysPrompt = `You are a helpful assistant, with limited access to the file system.`

// setup builds the shared workspace dirs and two agents over the same
// conversation: the OpenRouter (luna) agent runs first, then the LMStudio
// agent takes over. Both share the same FS toolset configuration.
func setup() (agentsDirs, *rellm.Agent, *rellm.Agent, error) {

	dirs, err := buildDirs()
	if err != nil {
		return agentsDirs{}, nil, nil, err
	}

	fsToolset, err := buildFSToolset(dirs.readOnlyDir, dirs.outputDir)
	if err != nil {
		return agentsDirs{}, nil, nil, err
	}

	conversation := rellm.NewInMemoryConversation()

	orAgent, err := buildORLunaSwitchAgent("or-luna-switch-agent", conversation, fsToolset)
	if err != nil {
		return agentsDirs{}, nil, nil, err
	}

	lmsAgent, err := buildLMSSwitchAgent("lms-switch-agent", conversation, fsToolset)
	if err != nil {
		return agentsDirs{}, nil, nil, err
	}

	return dirs, orAgent, lmsAgent, nil
}

func buildORLunaSwitchAgent(name string, conversation rellm.Conversation, fsToolset *examplesutils.FSToolset) (*rellm.Agent, error) {
	p, err := buildOpenRouterProvider("openai/gpt-5.6-luna")
	if err != nil {
		return nil, err
	}

	return buildSwitchAgent(p, name, conversation, fsToolset)
}

func buildLMSSwitchAgent(name string, conversation rellm.Conversation, fsToolset *examplesutils.FSToolset) (*rellm.Agent, error) {
	p, err := buildLMSProvider()
	if err != nil {
		return nil, err
	}

	return buildSwitchAgent(p, name, conversation, fsToolset)
}

func buildSwitchAgent(p rellm.Provider, name string, conversation rellm.Conversation, fsToolset *examplesutils.FSToolset) (*rellm.Agent, error) {
	return rellm.NewAgentBuilder().
		WithProvider(p).
		WithAgentName(name).
		WithMaxAgentSteps(20).
		WithConversation(conversation).
		WithToolset(fsToolset, rellm.ParallelToolCallsEnable).
		WithSystemMessage(switchAgentSysPrompt).
		WithInspectEachRequest(examplesutils.InspectWithReqLog).
		WithInspectEachResponse(examplesutils.InspectWithRespLog).
		WithImageGenerationKeepInTheLoop().
		WithUnknownConversationElementKeepInTheLoop().
		Build()
}

func buildLMSProvider() (rellm.Provider, error) {
	return rellm.NewLMStudioProvider("google/gemma-4-26b-a4b", "http://127.0.0.1", "1234")
}

func buildOpenRouterProvider(model rellm.Model) (rellm.Provider, error) {
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

func buildFSToolset(readOnlyDir, outputDir string) (*examplesutils.FSToolset, error) {
	fs, err := examplesutils.NewLimitedFileSystem([]string{readOnlyDir}, outputDir)
	if err != nil {
		return nil, err
	}

	return examplesutils.NewFSToolset(fs), nil
}

func buildDirs() (agentsDirs, error) {
	workspace, err := examplesutils.CreateDirInSysTmp("agent-log")
	if err != nil {
		return agentsDirs{}, err
	}

	readOnlyDir, err := examplesutils.CreateSubDir(workspace, "readonly_agent_input")
	if err != nil {
		return agentsDirs{}, err
	}

	outputDir, err := examplesutils.CreateSubDir(workspace, "output")
	if err != nil {
		return agentsDirs{}, err
	}

	return agentsDirs{
		readOnlyDir: readOnlyDir,
		outputDir:   outputDir,
		workspace:   workspace,
	}, nil
}

type agentsDirs struct {
	workspace   string
	readOnlyDir string
	outputDir   string
}

func panicWithLog(msg string, err error) {
	logMsg := msg + "\n" + err.Error()
	color.Red(logMsg)
	panic(logMsg)
}

func createFile(locationPath, fname, content string) error {
	filePath := path.Join(locationPath, fname)
	return os.WriteFile(filePath, []byte(content), 0644)
}
