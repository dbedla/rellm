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

	agent := buildProAgent(agentLogDir)

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
		llmResp, rawRsp, err := agent.AskLikeAPro(msg, SetParameters)
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}
		color.Blue(llmResp)
		if rawRsp != nil {
			color.Yellow("This conversation total cost in tokens: %d\n", rawRsp.Usage.TotalTokens)
			color.Yellow("input tokens: %d\n", rawRsp.Usage.InputTokens)
			color.Yellow("output tokens: %d\n", rawRsp.Usage.OutputTokens)
		}
	}
}
