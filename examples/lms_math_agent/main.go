package main

import (
	"context"
	"rellm/internal/examplesutils"

	"github.com/fatih/color"
)

func main() {
	agent, err := buildMathAgent()
	if err != nil {
		panic(err)
	}

	for {
		msg, err := examplesutils.ReadConsoleInput()
		if err != nil {
			color.Red("unable to read console input: %s", err)
			return
		}
		if msg == "EXIT" {
			return
		}

		ctx := context.Background()
		llmResp, err := agent.Ask(ctx, msg)
		if err != nil {
			color.Red("unable to ask question: %s", err.Error())
			continue
		}
		color.Blue(*llmResp.Messages)
	}
}
