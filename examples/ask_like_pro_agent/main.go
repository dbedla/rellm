package main

import (
	"rellm/pkg/utils"

	"github.com/fatih/color"
)

func main() {

	agentLogDir, err := utils.CreateDirInSysTmp("agent-log")
	if err != nil {
		panic(err)
	}

	color.Red("Agent log dir: %s", agentLogDir)

	agent := buildLAskLikeProAgent(agentLogDir)

	defer func() {
		err := agent.StoreConversation()
		if err != nil {
			color.Red("unable to store conversation: %s", err)
		}
	}()

	for {
		msg := utils.ReadConsoleInput()
		if msg == "EXIT" {
			return
		}
		llmResp := agent.AskLikePro(msg, SetParameters)
		color.Blue(llmResp)
	}
}
