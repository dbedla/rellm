package main

import (
	"rellm/pkg/utils"

	"github.com/fatih/color"
)

func main() {
	baseAgentWorkspace, err := utils.CreateDirInSysTmp("agent-log")
	if err != nil {
		panic(err)
	}

	color.Red("Agent log dir: %s", baseAgentWorkspace)

	fsAgent, err := buildFSAgent(baseAgentWorkspace)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := fsAgent.StoreConversation()
		if err != nil {
			color.Red("unable to store conversation: %s", err)
		}
	}()

	for {
		msg := utils.ReadConsoleInput()
		if msg == "EXIT" {
			return
		}
		llmResp := fsAgent.Ask(msg)
		color.Blue(llmResp)
	}
}
