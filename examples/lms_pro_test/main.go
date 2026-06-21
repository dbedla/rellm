package main

import (
	"rellm/pkg/agentsutills"

	"github.com/fatih/color"
)

func main() {

	agentLogDir, err := agentsutills.CreateDirInSysTmp("agent-log")
	if err != nil {
		panic(err)
	}

	color.Red("Agent log dir: %s", agentLogDir)

	agent, err := buildTestProToolAgent(agentLogDir)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := agent.StoreConversation()
		if err != nil {
			color.Red("unable to store conversation: %s", err)
		}
	}()

	for {
		msg, err := agentsutills.ReadConsoleInput()
		if err != nil {
			color.Red("unable to read console input: %s", err)
			return
		}
		if msg == "EXIT" {
			return
		}
		llmResp, rawRsp, err := agent.AskLikeAPro(msg, SetParametersWithReqLog)
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
