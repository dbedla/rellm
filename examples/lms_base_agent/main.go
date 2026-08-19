package main

import (
	"rellm/pkg/agentsutils"

	"github.com/fatih/color"
)

func main() {
	agent, err := buildLBaseAgent()
	if err != nil {
		panic(err)
	}

	for {
		msg, err := agentsutils.ReadConsoleInput()
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
