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
		msg, err := agentsutills.ReadConsoleInput()
		if err != nil {
			color.Red("unable to read console input: %s", err)
			return
		}
		if msg == "EXIT" {
			return
		}
		llmResp, err := agent.Ask(msg)
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}
		color.Blue(llmResp)
	}

}
