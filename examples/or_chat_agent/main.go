package main

import (
	"rellm/pkg/agentsutils"

	"github.com/fatih/color"
)

func main() {
	agentLogDir, err := agentsutils.CreateDirInSysTmp("agent-log")
	if err != nil {
		panic(err)
	}

	color.Red("Agent log dir: %s", agentLogDir)

	agent, err := buildLBaseAgent(agentLogDir)
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
		msg, err := agentsutils.ReadConsoleInput()
		if err != nil {
			color.Red("unable to read console input: %s", err)
			return
		}
		if msg == "EXIT" {
			return
		}

		//llmResp, rawResp, err := agent.AskLikeAPro(msg, agentsutils.InspectWithReqLog, agentsutils.InspectWithRespLog)
		llmResp, err := agent.Ask(msg)
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}
		//if rawResp != nil {
		//	color.Yellow("This conversation total cost in tokens: %d\n", rawResp.Usage.TotalTokens)
		//	if rawResp.Usage.Cost != nil {
		//		color.Yellow("This conversation total cost in USD: %f\n", *rawResp.Usage.Cost)
		//	}
		//}
		color.Blue(llmResp)
	}

}
