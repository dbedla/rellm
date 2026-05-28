package main

import (
	"os"
	"rellm/pkg/utils"

	"github.com/fatih/color"
)

func CreateDirInSysTmp(prefix string) (string, error) {
	pattern := prefix + "-*"

	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", err
	}

	return tempDir, nil
}

func main() {
	agentLogDir, err := CreateDirInSysTmp("agent-log")
	if err != nil {
		panic(err)
	}

	color.Red("Agent log dir: %s", agentLogDir)

	agent := buildLBaseAgent(agentLogDir)

	defer agent.StoreConversation()

	for {
		msg := utils.ReadConsoleInput()
		if msg == "EXIT" {
			return
		}
		llmResp := agent.Ask(msg)
		color.Blue(llmResp)
	}

}
