package main

import (
	"rellm/pkg/agentsutils"

	"github.com/fatih/color"
)

func main() {
	fsAgentSpace, err := agentsutils.CreateDirInSysTmp("agent-log")
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
		msg, err := agentsutils.ReadConsoleInput()
		if err != nil {
			color.Red("unable to read console input: %s", err)
			return
		}
		if msg == "EXIT" {
			return
		}
		llmResp, err := fsAgent.Prompt(msg).Execute()
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}

		color.Blue(llmResp)
	}
}
