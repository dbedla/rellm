package main

import (
	"rellm/pkg/utils"

	"github.com/fatih/color"
)

func main() {
	fsAgentSpace, err := utils.CreateDirInSysTmp("agent-log")
	if err != nil {
		panic(err)
	}

	color.Red("Agent space dir: %s", fsAgentSpace)

	fsAgent, err := buildFSAgent(fsAgentSpace)
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
		llmResp, err := fsAgent.Ask(msg)
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}

		color.Blue(llmResp)
	}
}
